package k8s

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"go.uber.org/zap"
	metav1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"strings"
	"time"
)

const (
	labelNodeRolePrefix = "node-role.kubernetes.io/"
	nodeLabelRole       = "kubernetes.io/role"
	Kibibyte            = 1024
	Mebibyte            = 1024 * 1024
	Gibibyte            = 1024 * 1024 * 1024
	kilobyte            = 1000
	Megabyte            = 1000 * 1000
	Gigabyte            = 1000 * 1000 * 1000
)

type K8sCapacityService interface {
	GetClusterCapacityDetailList(clusters []*cluster.ClusterBean) ([]*ClusterCapacityDetail, error)
	GetClusterCapacityDetail(cluster *cluster.ClusterBean, callForList bool) (*ClusterCapacityDetail, error)
	GetNodeCapacityDetailsListByCluster(cluster *cluster.ClusterBean) ([]*NodeCapacityDetail, error)
	GetNodeCapacityDetailByNameAndCluster(cluster *cluster.ClusterBean, name string) (*NodeCapacityDetail, error)
	UpdateNodeManifest(request *NodeManifestUpdateDto) (*application.ManifestResponse, error)
}
type K8sCapacityServiceImpl struct {
	logger                *zap.SugaredLogger
	clusterService        cluster.ClusterService
	k8sApplicationService K8sApplicationService
	k8sClientService      application.K8sClientService
	clusterCronService    ClusterCronService
}

func NewK8sCapacityServiceImpl(Logger *zap.SugaredLogger,
	clusterService cluster.ClusterService,
	k8sApplicationService K8sApplicationService,
	k8sClientService application.K8sClientService,
	clusterCronService ClusterCronService) *K8sCapacityServiceImpl {
	return &K8sCapacityServiceImpl{
		logger:                Logger,
		clusterService:        clusterService,
		k8sApplicationService: k8sApplicationService,
		k8sClientService:      k8sClientService,
		clusterCronService:    clusterCronService,
	}
}

func (impl *K8sCapacityServiceImpl) GetClusterCapacityDetailList(clusters []*cluster.ClusterBean) ([]*ClusterCapacityDetail, error) {
	var clustersDetails []*ClusterCapacityDetail
	for _, cluster := range clusters {
		clusterCapacityDetail := &ClusterCapacityDetail{}
		var err error
		if len(cluster.ErrorInConnecting) > 0 {
			clusterCapacityDetail.ErrorInConnection = cluster.ErrorInConnecting
		} else {
			clusterCapacityDetail, err = impl.GetClusterCapacityDetail(cluster, true)
			if err != nil {
				impl.logger.Errorw("error in getting cluster capacity details by id", "err", err)
				clusterCapacityDetail = &ClusterCapacityDetail{
					ErrorInConnection: err.Error(),
				}
			}
		}
		clusterCapacityDetail.Id = cluster.Id
		clusterCapacityDetail.Name = cluster.ClusterName
		clustersDetails = append(clustersDetails, clusterCapacityDetail)
	}
	return clustersDetails, nil
}

func (impl *K8sCapacityServiceImpl) GetClusterCapacityDetail(cluster *cluster.ClusterBean, callForList bool) (*ClusterCapacityDetail, error) {
	//getting rest config by clusterId
	restConfig, err := impl.k8sApplicationService.GetRestConfigByCluster(cluster)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster", "err", err, "clusterId", cluster.Id)
		return nil, err
	}
	//getting kubernetes clientSet by rest config
	k8sClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting client set by rest config", "err", err, "restConfig", restConfig)
		return nil, err
	}
	clusterDetail := &ClusterCapacityDetail{}
	nodeList, err := k8sClientSet.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting node list", "err", err, "clusterId", cluster.Id)
		return nil, err
	}
	var clusterCpuCapacity resource.Quantity
	var clusterMemoryCapacity resource.Quantity
	var clusterCpuAllocatable resource.Quantity
	var clusterMemoryAllocatable resource.Quantity
	nodeCount := 0
	nodesK8sVersionMap := make(map[string]bool)
	//map of node condition and name of all nodes that condition is true on
	nodeErrors := make(map[metav1.NodeConditionType][]string)
	var nodesK8sVersion []string
	for _, node := range nodeList.Items {
		errorsInNode := findNodeErrors(&node)
		for conditionName := range errorsInNode {
			if nodeNames, ok := nodeErrors[conditionName]; ok {
				nodeNames = append(nodeNames, node.Name)
				nodeErrors[conditionName] = nodeNames
			} else {
				nodeErrors[conditionName] = []string{node.Name}
			}
		}
		nodeCount += 1
		if _, ok := nodesK8sVersionMap[node.Status.NodeInfo.KubeletVersion]; !ok {
			nodesK8sVersionMap[node.Status.NodeInfo.KubeletVersion] = true
			nodesK8sVersion = append(nodesK8sVersion, node.Status.NodeInfo.KubeletVersion)
		}
		clusterCpuCapacity.Add(node.Status.Capacity[metav1.ResourceCPU])
		clusterMemoryCapacity.Add(node.Status.Capacity[metav1.ResourceMemory])
		clusterCpuAllocatable.Add(node.Status.Allocatable[metav1.ResourceCPU])
		clusterMemoryAllocatable.Add(node.Status.Allocatable[metav1.ResourceMemory])
	}
	clusterDetail.NodeErrors = nodeErrors
	clusterDetail.NodeK8sVersions = nodesK8sVersion
	clusterDetail.Cpu = &ResourceDetailObject{
		Capacity: getResourceString(clusterCpuCapacity, metav1.ResourceCPU),
	}
	clusterDetail.Memory = &ResourceDetailObject{
		Capacity: getResourceString(clusterMemoryCapacity, metav1.ResourceMemory),
	}
	if callForList {
		//assigning additional data for cluster listing api call
		clusterDetail.NodeCount = nodeCount
		//getting serverVersion
		serverVersion, err := k8sClientSet.DiscoveryClient.ServerVersion()
		if err != nil {
			impl.logger.Errorw("error in getting server version", "err", err, "clusterId", cluster.Id)
			return nil, err
		}
		clusterDetail.ServerVersion = serverVersion.GitVersion
	} else {
		//update data for cluster detail api call
		//getting metrics clientSet by rest config
		metricsClientSet, err := metrics.NewForConfig(restConfig)
		if err != nil {
			impl.logger.Errorw("error in getting metrics client set", "err", err)
			return nil, err
		}
		//empty namespace: get pods for all namespaces
		podList, err := k8sClientSet.CoreV1().Pods("").List(context.Background(), v1.ListOptions{})
		if err != nil {
			impl.logger.Errorw("error in getting pod list", "err", err)
			return nil, err
		}
		var clusterCpuUsage resource.Quantity
		var clusterMemoryUsage resource.Quantity
		var clusterCpuLimits resource.Quantity
		var clusterCpuRequests resource.Quantity
		var clusterMemoryLimits resource.Quantity
		var clusterMemoryRequests resource.Quantity
		nmList, err := metricsClientSet.MetricsV1beta1().NodeMetricses().List(context.Background(), v1.ListOptions{})
		if err != nil {
			impl.logger.Errorw("error in getting nodeMetrics list", "err", err)
		} else if nmList != nil {
			for _, nm := range nmList.Items {
				clusterCpuUsage.Add(nm.Usage[metav1.ResourceCPU])
				clusterMemoryUsage.Add(nm.Usage[metav1.ResourceMemory])
			}
			clusterDetail.Cpu.UsagePercentage = convertToPercentage(&clusterCpuUsage, &clusterCpuAllocatable)
			clusterDetail.Memory.UsagePercentage = convertToPercentage(&clusterMemoryUsage, &clusterMemoryAllocatable)
		}
		for _, pod := range podList.Items {
			if pod.Status.Phase != metav1.PodSucceeded && pod.Status.Phase != metav1.PodFailed {
				requests, limits := resourcehelper.PodRequestsAndLimits(&pod)
				clusterCpuLimits.Add(limits[metav1.ResourceCPU])
				clusterCpuRequests.Add(requests[metav1.ResourceCPU])
				clusterMemoryLimits.Add(limits[metav1.ResourceMemory])
				clusterMemoryRequests.Add(requests[metav1.ResourceMemory])
			}
		}
		clusterDetail.Cpu.RequestPercentage = convertToPercentage(&clusterCpuRequests, &clusterCpuAllocatable)
		clusterDetail.Cpu.LimitPercentage = convertToPercentage(&clusterCpuLimits, &clusterCpuAllocatable)
		clusterDetail.Memory.RequestPercentage = convertToPercentage(&clusterMemoryRequests, &clusterMemoryAllocatable)
		clusterDetail.Memory.LimitPercentage = convertToPercentage(&clusterMemoryLimits, &clusterMemoryAllocatable)
	}
	return clusterDetail, nil
}

func (impl *K8sCapacityServiceImpl) GetNodeCapacityDetailsListByCluster(cluster *cluster.ClusterBean) ([]*NodeCapacityDetail, error) {
	//getting rest config by clusterId
	restConfig, err := impl.k8sApplicationService.GetRestConfigByCluster(cluster)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster", "err", err, "clusterId", cluster.Id)
		return nil, err
	}
	//getting kubernetes clientSet by rest config
	k8sClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting client set by rest config", "err", err, "restConfig", restConfig)
		return nil, err
	}
	//getting metrics clientSet by rest config
	metricsClientSet, err := metrics.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting metrics client set", "err", err)
		return nil, err
	}
	nodeMetricsList, err := metricsClientSet.MetricsV1beta1().NodeMetricses().List(context.Background(), v1.ListOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting node metrics", "err", err)
	}
	nodeList, err := k8sClientSet.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting node list", "err", err, "clusterId", cluster.Id)
		return nil, err
	}
	//empty namespace: get pods for all namespaces
	podList, err := k8sClientSet.CoreV1().Pods("").List(context.Background(), v1.ListOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting pod list", "err", err)
		return nil, err
	}
	nodeResourceUsage := make(map[string]metav1.ResourceList)
	if nodeMetricsList != nil {
		for _, nodeMetrics := range nodeMetricsList.Items {
			nodeResourceUsage[nodeMetrics.Name] = nodeMetrics.Usage
		}
	}
	var nodeDetails []*NodeCapacityDetail
	for _, node := range nodeList.Items {
		nodeDetail, err := impl.getNodeDetail(&node, nodeResourceUsage, podList, true, restConfig)
		if err != nil {
			impl.logger.Errorw("error in getting node detail for list", "err", err)
			return nil, err
		}
		nodeDetails = append(nodeDetails, nodeDetail)
	}
	return nodeDetails, nil
}

func (impl *K8sCapacityServiceImpl) GetNodeCapacityDetailByNameAndCluster(cluster *cluster.ClusterBean, name string) (*NodeCapacityDetail, error) {
	//getting rest config by clusterId
	restConfig, err := impl.k8sApplicationService.GetRestConfigByCluster(cluster)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster", "err", err, "clusterId", cluster.Id)
		return nil, err
	}
	//getting kubernetes clientSet by rest config
	k8sClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting client set by rest config", "err", err, "restConfig", restConfig)
		return nil, err
	}
	//getting metrics clientSet by rest config
	metricsClientSet, err := metrics.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting metrics client set", "err", err)
		return nil, err
	}
	nodeMetrics, err := metricsClientSet.MetricsV1beta1().NodeMetricses().Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting node metrics", "err", err)
	}
	node, err := k8sClientSet.CoreV1().Nodes().Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting node list", "err", err)
		return nil, err
	}
	//empty namespace: get pods for all namespaces
	podList, err := k8sClientSet.CoreV1().Pods("").List(context.Background(), v1.ListOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting pod list", "err", err)
		return nil, err
	}
	nodeResourceUsage := make(map[string]metav1.ResourceList)
	if nodeMetrics != nil {
		nodeResourceUsage[nodeMetrics.Name] = nodeMetrics.Usage
	}
	nodeDetail, err := impl.getNodeDetail(node, nodeResourceUsage, podList, false, restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting node detail", "err", err)
		return nil, err
	}
	//updating cluster name
	nodeDetail.ClusterName = cluster.ClusterName
	return nodeDetail, nil
}
func (impl *K8sCapacityServiceImpl) getNodeDetail(node *metav1.Node, nodeResourceUsage map[string]metav1.ResourceList, podList *metav1.PodList, callForList bool, restConfig *rest.Config) (*NodeCapacityDetail, error) {
	cpuAllocatable := node.Status.Allocatable[metav1.ResourceCPU]
	memoryAllocatable := node.Status.Allocatable[metav1.ResourceMemory]
	podCount := 0
	nodeRequestsResourceList := make(metav1.ResourceList)
	nodeLimitsResourceList := make(metav1.ResourceList)
	var podDetailList []*PodCapacityDetail
	for _, pod := range podList.Items {
		if pod.Spec.NodeName == node.Name {
			if callForList {
				podCount++
			} else {
				var requests, limits metav1.ResourceList
				if pod.Status.Phase != metav1.PodSucceeded && pod.Status.Phase != metav1.PodFailed {
					requests, limits = resourcehelper.PodRequestsAndLimits(&pod)
					nodeRequestsResourceList = AddTwoResourceList(nodeRequestsResourceList, requests)
					nodeLimitsResourceList = AddTwoResourceList(nodeLimitsResourceList, limits)
				}
				podDetailList = append(podDetailList, getPodDetail(pod, cpuAllocatable, memoryAllocatable, limits, requests))
			}
		}
	}
	var labels []*LabelAnnotationTaintObject
	for k, v := range node.Labels {
		labelObj := &LabelAnnotationTaintObject{
			Key:   k,
			Value: v,
		}
		labels = append(labels, labelObj)
	}
	nodeDetail := &NodeCapacityDetail{
		Name:          node.Name,
		K8sVersion:    node.Status.NodeInfo.KubeletVersion,
		Errors:        findNodeErrors(node),
		InternalIp:    getNodeInternalIP(node),
		ExternalIp:    getNodeExternalIP(node),
		Unschedulable: node.Spec.Unschedulable,
		Roles:         findNodeRoles(node),
		Labels:        labels,
		Status:        findNodeStatus(node),
		TaintCount:    len(node.Spec.Taints),
		CreatedAt:     node.CreationTimestamp.String(),
	}
	nodeUsageResourceList := nodeResourceUsage[node.Name]
	if callForList {
		// assigning additional data for node listing api call
		nodeDetail.Age = translateTimestampSince(node.CreationTimestamp)
		nodeDetail.PodCount = podCount
		cpuUsage, cpuUsageOk := nodeUsageResourceList[metav1.ResourceCPU]
		memoryUsage, memoryUsageOk := nodeUsageResourceList[metav1.ResourceMemory]
		nodeDetail.Cpu = &ResourceDetailObject{
			Allocatable:        getResourceString(cpuAllocatable, metav1.ResourceCPU),
			AllocatableInBytes: cpuAllocatable.Value(),
		}
		nodeDetail.Memory = &ResourceDetailObject{
			Allocatable:        getResourceString(memoryAllocatable, metav1.ResourceMemory),
			AllocatableInBytes: memoryAllocatable.Value(),
		}
		if cpuUsageOk {
			nodeDetail.Cpu.Usage = getResourceString(cpuUsage, metav1.ResourceCPU)
			nodeDetail.Cpu.UsageInBytes = cpuUsage.Value()
			nodeDetail.Cpu.UsagePercentage = convertToPercentage(&cpuUsage, &cpuAllocatable)
		}
		if memoryUsageOk {
			nodeDetail.Memory.Usage = getResourceString(memoryUsage, metav1.ResourceMemory)
			nodeDetail.Memory.UsageInBytes = memoryUsage.Value()
			nodeDetail.Memory.UsagePercentage = convertToPercentage(&memoryUsage, &memoryAllocatable)
		}
	} else {
		//update data for node detail api call
		err := impl.updateAdditionalDetailForNode(nodeDetail, node, nodeLimitsResourceList, nodeRequestsResourceList, nodeUsageResourceList, podDetailList, restConfig)
		if err != nil {
			impl.logger.Errorw("error in getting updating data for node detail", "err", err)
			return nil, err
		}
	}
	return nodeDetail, nil
}

func (impl *K8sCapacityServiceImpl) updateAdditionalDetailForNode(nodeDetail *NodeCapacityDetail, node *metav1.Node,
	nodeLimitsResourceList metav1.ResourceList, nodeRequestsResourceList metav1.ResourceList,
	nodeUsageResourceList metav1.ResourceList, podDetailList []*PodCapacityDetail, restConfig *rest.Config) error {
	nodeDetail.Version = "v1"
	nodeDetail.Kind = "Node"
	nodeDetail.Pods = podDetailList
	var annotations []*LabelAnnotationTaintObject
	for k, v := range node.Annotations {
		annotationObj := &LabelAnnotationTaintObject{
			Key:   k,
			Value: v,
		}
		annotations = append(annotations, annotationObj)
	}
	nodeDetail.Annotations = annotations

	var taints []*LabelAnnotationTaintObject
	for _, taint := range node.Spec.Taints {
		taintObj := &LabelAnnotationTaintObject{
			Key:    taint.Key,
			Value:  taint.Value,
			Effect: string(taint.Effect),
		}
		taints = append(taints, taintObj)
	}
	nodeDetail.Taints = taints
	//map of {conditionType : isErrorCondition }, Valid/Non-error conditions to be updated with update at kubernetes end
	NodeAllConditionsMap := map[metav1.NodeConditionType]bool{metav1.NodeReady: false, metav1.NodeMemoryPressure: true,
		metav1.NodeDiskPressure: true, metav1.NodeNetworkUnavailable: true, metav1.NodePIDPressure: true}
	var conditions []*NodeConditionObject
	for _, condition := range node.Status.Conditions {
		if isErrorCondition, ok := NodeAllConditionsMap[condition.Type]; ok {
			conditionObj := &NodeConditionObject{
				Type:    string(condition.Type),
				Reason:  condition.Reason,
				Message: condition.Message,
			}
			if (!isErrorCondition && condition.Status == metav1.ConditionTrue) || (isErrorCondition && condition.Status == metav1.ConditionFalse) {
				conditionObj.HaveIssue = false
			} else {
				conditionObj.HaveIssue = true
			}
			conditions = append(conditions, conditionObj)
		}
	}
	nodeDetail.Conditions = conditions

	nodeCapacityResourceList := node.Status.Capacity
	nodeAllocatableResourceList := node.Status.Allocatable
	for resourceName, allocatable := range nodeAllocatableResourceList {
		limits, limitsOk := nodeLimitsResourceList[resourceName]
		requests, requestsOk := nodeRequestsResourceList[resourceName]
		usage, usageOk := nodeUsageResourceList[resourceName]
		capacity := nodeCapacityResourceList[resourceName]
		r := &ResourceDetailObject{
			ResourceName: string(resourceName),
			Allocatable:  getResourceString(allocatable, resourceName),
			Capacity:     getResourceString(capacity, resourceName),
		}
		if limitsOk {
			r.Limit = getResourceString(limits, resourceName)
			r.LimitPercentage = convertToPercentage(&limits, &allocatable)
		}
		if requestsOk {
			r.Request = getResourceString(requests, resourceName)
			r.RequestPercentage = convertToPercentage(&requests, &allocatable)
		}
		if usageOk {
			r.Usage = getResourceString(usage, resourceName)
			r.UsagePercentage = convertToPercentage(&usage, &allocatable)
		}
		nodeDetail.Resources = append(nodeDetail.Resources, r)
	}
	//getting manifest
	manifestRequest := &application.K8sRequestBean{
		ResourceIdentifier: application.ResourceIdentifier{
			Name: node.Name,
			GroupVersionKind: schema.GroupVersionKind{
				Version: nodeDetail.Version,
				Kind:    nodeDetail.Kind,
			},
		},
	}
	manifestResponse, err := impl.k8sClientService.GetResource(restConfig, manifestRequest)
	if err != nil {
		impl.logger.Errorw("error in getting node manifest", "err", err)
		return err
	}
	nodeDetail.Manifest = manifestResponse.Manifest
	return nil
}

func (impl *K8sCapacityServiceImpl) UpdateNodeManifest(request *NodeManifestUpdateDto) (*application.ManifestResponse, error) {
	//getting rest config by clusterId
	restConfig, err := impl.k8sApplicationService.GetRestConfigByClusterId(request.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster id", "err", err, "clusterId", request.ClusterId)
		return nil, err
	}
	manifestUpdateReq := &application.K8sRequestBean{
		ResourceIdentifier: application.ResourceIdentifier{
			Name: request.Name,
			GroupVersionKind: schema.GroupVersionKind{
				Group:   "",
				Version: request.Version,
				Kind:    request.Kind,
			},
		},
		Patch: request.ManifestPatch,
	}
	manifestResponse, err := impl.k8sClientService.UpdateResource(restConfig, manifestUpdateReq)
	if err != nil {
		impl.logger.Errorw("error in updating node manifest", "err", err)
		return nil, err
	}
	return manifestResponse, nil
}
func getPodDetail(pod metav1.Pod, cpuAllocatable resource.Quantity, memoryAllocatable resource.Quantity, limits metav1.ResourceList, requests metav1.ResourceList) *PodCapacityDetail {
	cpuLimits, cpuLimitsOk := limits[metav1.ResourceCPU]
	cpuRequests, cpuRequestsOk := requests[metav1.ResourceCPU]
	memoryLimits, memoryLimitsOk := limits[metav1.ResourceMemory]
	memoryRequests, memoryRequestsOk := requests[metav1.ResourceMemory]
	podDetail := &PodCapacityDetail{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Age:       translateTimestampSince(pod.CreationTimestamp),
		CreatedAt: pod.CreationTimestamp.String(),
		Cpu: &ResourceDetailObject{
			Limit:   getResourceString(cpuLimits, metav1.ResourceCPU),
			Request: getResourceString(cpuRequests, metav1.ResourceCPU),
		},
		Memory: &ResourceDetailObject{
			Limit:   getResourceString(memoryLimits, metav1.ResourceMemory),
			Request: getResourceString(memoryRequests, metav1.ResourceMemory),
		},
	}
	if cpuLimitsOk {
		podDetail.Cpu.LimitPercentage = convertToPercentage(&cpuLimits, &cpuAllocatable)
	}
	if cpuRequestsOk {
		podDetail.Cpu.RequestPercentage = convertToPercentage(&cpuRequests, &cpuAllocatable)
	}
	if memoryLimitsOk {
		podDetail.Memory.LimitPercentage = convertToPercentage(&memoryLimits, &memoryAllocatable)
	}
	if memoryRequestsOk {
		podDetail.Memory.RequestPercentage = convertToPercentage(&memoryRequests, &memoryAllocatable)
	}
	return podDetail
}
func convertToPercentage(actual, allocatable *resource.Quantity) string {
	if actual == nil || allocatable == nil {
		return "<nil>"
	}
	utilPercent := float64(0)
	if allocatable.MilliValue() > 0 {
		utilPercent = float64(actual.MilliValue()) / float64(allocatable.MilliValue()) * 100
	}
	return fmt.Sprintf("%d%%", int64(utilPercent))
}

func getResourceString(quantity resource.Quantity, resourceName metav1.ResourceName) string {
	standardResources := map[metav1.ResourceName]bool{metav1.ResourceCPU: true, metav1.ResourceMemory: true, metav1.ResourceStorage: true, metav1.ResourceEphemeralStorage: true}

	if _, ok := standardResources[resourceName]; !ok {
		//not a standard resource, we do not know if conversion would be valid or not
		//for example - pods: "250", this is not in bytes but an integer so conversion is invalid
		return quantity.String()
	} else {
		var quantityStr string
		value := quantity.Value()
		valueGi := value / Gibibyte
		//allowing remainder 0 only, because for Gi rounding off will be highly erroneous
		if valueGi > 1 && value%Gibibyte == 0 {
			quantityStr = fmt.Sprintf("%dGi", valueGi)
		} else {
			valueMi := value / Mebibyte
			if valueMi > 10 {
				if value%Mebibyte != 0 {
					valueMi++
				}
				quantityStr = fmt.Sprintf("%dMi", valueMi)
			} else if value > 1000 {
				valueKi := value / Kibibyte
				if value%Kibibyte != 0 {
					valueKi++
				}
				quantityStr = fmt.Sprintf("%dKi", valueKi)
			} else {
				quantityStr = fmt.Sprintf("%dm", quantity.MilliValue())
			}
		}
		return quantityStr
	}
}

func translateTimestampSince(timestamp v1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}
	return duration.HumanDuration(time.Since(timestamp.Time))
}

func findNodeRoles(node *metav1.Node) []string {
	roles := sets.NewString()
	for k, v := range node.Labels {
		switch {
		case strings.HasPrefix(k, labelNodeRolePrefix):
			if role := strings.TrimPrefix(k, labelNodeRolePrefix); len(role) > 0 {
				roles.Insert(role)
			}
		case k == nodeLabelRole && v != "":
			roles.Insert(v)
		}
	}
	if roles.Len() > 0 {
		return roles.List()
	} else {
		return []string{"none"}
	}
}

func findNodeStatus(node *metav1.Node) string {
	conditionMap := make(map[metav1.NodeConditionType]*metav1.NodeCondition)
	//Valid conditions to be updated with update at kubernetes end
	NodeAllValidConditions := []metav1.NodeConditionType{metav1.NodeReady}
	for _, condition := range node.Status.Conditions {
		conditionMap[condition.Type] = &condition
	}
	var status string
	for _, validCondition := range NodeAllValidConditions {
		if condition, ok := conditionMap[validCondition]; ok {
			if condition.Status == metav1.ConditionTrue {
				status = string(condition.Type)
			} else {
				status = fmt.Sprintf("Not %s", string(condition.Type))
			}
		}
	}
	if len(status) == 0 {
		status = "Unknown"
	}
	return status
}

func findNodeErrors(node *metav1.Node) map[metav1.NodeConditionType]string {
	conditionMap := make(map[metav1.NodeConditionType]metav1.NodeCondition)
	NodeAllErrorConditions := []metav1.NodeConditionType{metav1.NodeMemoryPressure, metav1.NodeDiskPressure, metav1.NodeNetworkUnavailable, metav1.NodePIDPressure}
	for _, condition := range node.Status.Conditions {
		conditionMap[condition.Type] = condition
	}
	conditionErrorMap := make(map[metav1.NodeConditionType]string)
	for _, errorCondition := range NodeAllErrorConditions {
		if condition, ok := conditionMap[errorCondition]; ok {
			if condition.Status == metav1.ConditionTrue {
				conditionErrorMap[condition.Type] = condition.Message
			}
		}
	}
	return conditionErrorMap
}

func getNodeExternalIP(node *metav1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == metav1.NodeExternalIP {
			return address.Address
		}
	}
	return "none"
}

func getNodeInternalIP(node *metav1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == metav1.NodeInternalIP {
			return address.Address
		}
	}
	return "none"
}

func AddTwoResourceList(oldResourceList metav1.ResourceList, newResourceList metav1.ResourceList) metav1.ResourceList {
	for res, quantity := range newResourceList {
		if oldQuantity, ok1 := oldResourceList[res]; ok1 {
			quantity.Add(oldQuantity)
		}
		oldResourceList[res] = quantity
	}
	return oldResourceList
}
