package capacity

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/k8s"
	application2 "github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/k8s/capacity/bean"
	k8s2 "github.com/devtron-labs/devtron/util/k8s"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"net/http"
	"strings"
	"time"
)

type K8sCapacityService interface {
	GetClusterCapacityDetailList(ctx context.Context, clusters []*cluster.ClusterBean) ([]*bean.ClusterCapacityDetail, error)
	GetClusterCapacityDetail(ctx context.Context, cluster *cluster.ClusterBean, callForList bool) (*bean.ClusterCapacityDetail, error)
	GetNodeCapacityDetailsListByCluster(ctx context.Context, cluster *cluster.ClusterBean) ([]*bean.NodeCapacityDetail, error)
	GetNodeCapacityDetailByNameAndCluster(ctx context.Context, cluster *cluster.ClusterBean, name string) (*bean.NodeCapacityDetail, error)
	UpdateNodeManifest(ctx context.Context, request *bean.NodeUpdateRequestDto) (*k8s2.ManifestResponse, error)
	DeleteNode(ctx context.Context, request *bean.NodeUpdateRequestDto) (*k8s2.ManifestResponse, error)
	CordonOrUnCordonNode(ctx context.Context, request *bean.NodeUpdateRequestDto) (string, error)
	DrainNode(ctx context.Context, request *bean.NodeUpdateRequestDto) (string, error)
	EditNodeTaints(ctx context.Context, request *bean.NodeUpdateRequestDto) (string, error)
	GetNode(ctx context.Context, clusterId int, nodeName string) (*corev1.Node, error)
}
type K8sCapacityServiceImpl struct {
	logger                *zap.SugaredLogger
	clusterService        cluster.ClusterService
	k8sApplicationService application2.K8sApplicationService
	K8sUtil               *k8s2.K8sUtil
	k8sCommonService      k8s.K8sCommonService
}

func NewK8sCapacityServiceImpl(Logger *zap.SugaredLogger, clusterService cluster.ClusterService, k8sApplicationService application2.K8sApplicationService, K8sUtil *k8s2.K8sUtil, k8sCommonService k8s.K8sCommonService) *K8sCapacityServiceImpl {
	return &K8sCapacityServiceImpl{
		logger:                Logger,
		clusterService:        clusterService,
		k8sApplicationService: k8sApplicationService,
		K8sUtil:               K8sUtil,
		k8sCommonService:      k8sCommonService,
	}
}

func (impl *K8sCapacityServiceImpl) GetClusterCapacityDetailList(ctx context.Context, clusters []*cluster.ClusterBean) ([]*bean.ClusterCapacityDetail, error) {
	var clustersDetails []*bean.ClusterCapacityDetail
	for _, cluster := range clusters {
		clusterCapacityDetail := &bean.ClusterCapacityDetail{}
		var err error
		if cluster.IsVirtualCluster {
			clusterCapacityDetail.IsVirtualCluster = cluster.IsVirtualCluster
		} else if len(cluster.ErrorInConnecting) > 0 {
			clusterCapacityDetail.ErrorInConnection = cluster.ErrorInConnecting
		} else {
			clusterCapacityDetail, err = impl.GetClusterCapacityDetail(ctx, cluster, true)
			if err != nil {
				impl.logger.Errorw("error in getting cluster capacity details by id", "err", err)
				clusterCapacityDetail = &bean.ClusterCapacityDetail{
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

func (impl *K8sCapacityServiceImpl) GetClusterCapacityDetail(ctx context.Context, cluster *cluster.ClusterBean, callForList bool) (*bean.ClusterCapacityDetail, error) {
	//getting kubernetes clientSet by rest config
	restConfig, k8sHttpClient, k8sClientSet, err := impl.getK8sConfigAndClients(ctx, cluster)
	if err != nil {
		return nil, err
	}
	clusterDetail := &bean.ClusterCapacityDetail{}
	nodeList, err := impl.K8sUtil.GetNodesList(ctx, k8sClientSet)
	if err != nil {
		impl.logger.Errorw("error in getting node list", "err", err, "clusterId", cluster.Id)
		return nil, err
	}
	clusterCpuAllocatable, clusterMemoryAllocatable, nodeCount := impl.setBasicClusterDetails(nodeList, clusterDetail)
	if callForList {
		//assigning additional data for cluster listing api call
		clusterDetail.NodeCount = nodeCount
		//getting serverVersion
		serverVersion, err := impl.K8sUtil.GetServerVersionFromDiscoveryClient(k8sClientSet)
		if err != nil {
			impl.logger.Errorw("error in getting server version", "err", err, "clusterId", cluster.Id)
			return nil, err
		}
		clusterDetail.ServerVersion = serverVersion.GitVersion
	} else {
		metricsClientSet, err := impl.K8sUtil.GetMetricsClientSet(restConfig, k8sHttpClient)
		if err != nil {
			impl.logger.Errorw("error in getting metrics client set", "err", err)
			return nil, err
		}
		err = impl.updateMetricsData(ctx, metricsClientSet, k8sClientSet, clusterDetail, clusterCpuAllocatable, clusterMemoryAllocatable)
		if err != nil {
			return nil, err
		}
	}
	return clusterDetail, nil
}

func (impl *K8sCapacityServiceImpl) setBasicClusterDetails(nodeList *corev1.NodeList, clusterDetail *bean.ClusterCapacityDetail) (resource.Quantity, resource.Quantity, int) {
	var clusterCpuCapacity resource.Quantity
	var clusterMemoryCapacity resource.Quantity
	var clusterCpuAllocatable resource.Quantity
	var clusterMemoryAllocatable resource.Quantity
	nodeCount := 0
	clusterNodeDetails := make([]bean.NodeDetails, 0)
	nodesK8sVersionMap := make(map[string]bool)
	//map of node condition and name of all nodes that condition is true on
	nodeErrors := make(map[corev1.NodeConditionType][]string)
	var nodesK8sVersion []string
	for _, node := range nodeList.Items {
		nodeGroup, taints := impl.getNodeGroupAndTaints(&node)
		nodeNameGroupName := bean.NodeDetails{
			NodeName:  node.Name,
			NodeGroup: nodeGroup,
			Taints:    taints,
		}
		clusterNodeDetails = append(clusterNodeDetails, nodeNameGroupName)
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
		clusterCpuCapacity.Add(node.Status.Capacity[corev1.ResourceCPU])
		clusterMemoryCapacity.Add(node.Status.Capacity[corev1.ResourceMemory])
		clusterCpuAllocatable.Add(node.Status.Allocatable[corev1.ResourceCPU])
		clusterMemoryAllocatable.Add(node.Status.Allocatable[corev1.ResourceMemory])
	}
	clusterDetail.NodeErrors = nodeErrors
	clusterDetail.NodeK8sVersions = nodesK8sVersion
	clusterDetail.NodeDetails = clusterNodeDetails
	clusterDetail.Cpu = &bean.ResourceDetailObject{
		Capacity: getResourceString(clusterCpuCapacity, corev1.ResourceCPU),
	}
	clusterDetail.Memory = &bean.ResourceDetailObject{
		Capacity: getResourceString(clusterMemoryCapacity, corev1.ResourceMemory),
	}
	return clusterCpuAllocatable, clusterMemoryAllocatable, nodeCount
}

func (impl *K8sCapacityServiceImpl) updateMetricsData(ctx context.Context, metricsClientSet *metrics.Clientset, k8sClientSet *kubernetes.Clientset, clusterDetail *bean.ClusterCapacityDetail, clusterCpuAllocatable resource.Quantity, clusterMemoryAllocatable resource.Quantity) error {
	//update data for cluster detail api call
	//getting metrics clientSet by rest config

	//empty namespace: get pods for all namespaces
	podList, err := impl.K8sUtil.GetPodsListForAllNamespaces(ctx, k8sClientSet)
	if err != nil {
		impl.logger.Errorw("error in getting pod list", "err", err)
		return err
	}
	var clusterCpuUsage resource.Quantity
	var clusterMemoryUsage resource.Quantity
	var clusterCpuLimits resource.Quantity
	var clusterCpuRequests resource.Quantity
	var clusterMemoryLimits resource.Quantity
	var clusterMemoryRequests resource.Quantity
	nmList, err := impl.K8sUtil.GetNmList(ctx, metricsClientSet)
	if err != nil {
		impl.logger.Errorw("error in getting nodeMetrics list", "err", err)
	} else if nmList != nil {
		for _, nm := range nmList.Items {
			clusterCpuUsage.Add(nm.Usage[corev1.ResourceCPU])
			clusterMemoryUsage.Add(nm.Usage[corev1.ResourceMemory])
		}
		clusterDetail.Cpu.UsagePercentage = convertToPercentage(&clusterCpuUsage, &clusterCpuAllocatable)
		clusterDetail.Memory.UsagePercentage = convertToPercentage(&clusterMemoryUsage, &clusterMemoryAllocatable)
	}
	for _, pod := range podList.Items {
		if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
			requests, limits := resourcehelper.PodRequestsAndLimits(&pod)
			clusterCpuLimits.Add(limits[corev1.ResourceCPU])
			clusterCpuRequests.Add(requests[corev1.ResourceCPU])
			clusterMemoryLimits.Add(limits[corev1.ResourceMemory])
			clusterMemoryRequests.Add(requests[corev1.ResourceMemory])
		}
	}
	clusterDetail.Cpu.RequestPercentage = convertToPercentage(&clusterCpuRequests, &clusterCpuAllocatable)
	clusterDetail.Cpu.LimitPercentage = convertToPercentage(&clusterCpuLimits, &clusterCpuAllocatable)
	clusterDetail.Memory.RequestPercentage = convertToPercentage(&clusterMemoryRequests, &clusterMemoryAllocatable)
	clusterDetail.Memory.LimitPercentage = convertToPercentage(&clusterMemoryLimits, &clusterMemoryAllocatable)
	return nil
}

func (impl *K8sCapacityServiceImpl) GetNodeCapacityDetailsListByCluster(ctx context.Context, cluster *cluster.ClusterBean) ([]*bean.NodeCapacityDetail, error) {
	//getting kubernetes clientSet by cluster config
	restConfig, k8sHttpClient, k8sClientSet, err := impl.getK8sConfigAndClients(ctx, cluster)
	if err != nil {
		return nil, err
	}
	//getting metrics clientSet by rest config
	metricsClientSet, err := impl.K8sUtil.GetMetricsClientSet(restConfig, k8sHttpClient)
	if err != nil {
		impl.logger.Errorw("error in getting metrics client set", "err", err)
		return nil, err
	}
	nodeMetricsList, err := impl.K8sUtil.GetNmList(ctx, metricsClientSet)
	if err != nil {
		impl.logger.Errorw("error in getting node metrics", "err", err)
	}
	nodeList, err := impl.K8sUtil.GetNodesList(ctx, k8sClientSet)
	if err != nil {
		impl.logger.Errorw("error in getting node list", "err", err, "clusterId", cluster.Id)
		return nil, err
	}
	//empty namespace: get pods for all namespaces
	podList, err := impl.K8sUtil.GetPodsListForAllNamespaces(ctx, k8sClientSet)
	if err != nil {
		impl.logger.Errorw("error in getting pod list", "err", err)
		return nil, err
	}
	nodeResourceUsage := make(map[string]corev1.ResourceList)
	if nodeMetricsList != nil {
		for _, nodeMetrics := range nodeMetricsList.Items {
			nodeResourceUsage[nodeMetrics.Name] = nodeMetrics.Usage
		}
	}
	var nodeDetails []*bean.NodeCapacityDetail
	for _, node := range nodeList.Items {
		nodeDetail, err := impl.getNodeDetail(ctx, &node, nodeResourceUsage, podList, true, cluster)
		if err != nil {
			impl.logger.Errorw("error in getting node detail for list", "err", err)
			return nil, err
		}
		nodeDetails = append(nodeDetails, nodeDetail)
	}
	return nodeDetails, nil
}

func (impl *K8sCapacityServiceImpl) GetNodeCapacityDetailByNameAndCluster(ctx context.Context, cluster *cluster.ClusterBean, name string) (*bean.NodeCapacityDetail, error) {

	//getting kubernetes clientSet by rest config
	restConfig, k8sHttpClient, k8sClientSet, err := impl.getK8sConfigAndClients(ctx, cluster)
	if err != nil {
		return nil, err
	}
	//getting metrics clientSet by rest config
	metricsClientSet, err := impl.K8sUtil.GetMetricsClientSet(restConfig, k8sHttpClient)
	if err != nil {
		impl.logger.Errorw("error in getting metrics client set", "err", err)
		return nil, err
	}
	nodeMetrics, err := impl.K8sUtil.GetNmByName(ctx, metricsClientSet, name)
	if err != nil {
		impl.logger.Errorw("error in getting node metrics", "err", err)
	}
	node, err := impl.K8sUtil.GetNodeByName(ctx, k8sClientSet, name)
	if err != nil {
		impl.logger.Errorw("error in getting node list", "err", err)
		return nil, err
	}
	//empty namespace: get pods for all namespaces
	podList, err := impl.K8sUtil.GetPodsListForAllNamespaces(ctx, k8sClientSet)
	if err != nil {
		impl.logger.Errorw("error in getting pod list", "err", err)
		return nil, err
	}
	nodeResourceUsage := make(map[string]corev1.ResourceList)
	if nodeMetrics != nil {
		nodeResourceUsage[nodeMetrics.Name] = nodeMetrics.Usage
	}
	nodeDetail, err := impl.getNodeDetail(ctx, node, nodeResourceUsage, podList, false, cluster)
	if err != nil {
		impl.logger.Errorw("error in getting node detail", "err", err)
		return nil, err
	}
	//updating cluster name
	nodeDetail.ClusterName = cluster.ClusterName
	return nodeDetail, nil
}

func (impl *K8sCapacityServiceImpl) getK8sConfigAndClients(ctx context.Context, cluster *cluster.ClusterBean) (*rest.Config, *http.Client, *kubernetes.Clientset, error) {
	clusterConfig := cluster.GetClusterConfig()
	return impl.K8sUtil.GetK8sConfigAndClients(&clusterConfig)
}
func (impl *K8sCapacityServiceImpl) getNodeGroupAndTaints(node *corev1.Node) (string, []*bean.LabelAnnotationTaintObject) {

	nodeGroup := impl.getNodeGroup(node)
	taints := impl.getTaints(node)
	return nodeGroup, taints
}

func (impl *K8sCapacityServiceImpl) getNodeGroup(node *corev1.Node) string {
	var nodeGroup = ""
	//different cloud providers have their own node group label
	for _, label := range bean.NodeGroupLabels {
		if ng, ok := node.Labels[label]; ok {
			nodeGroup = ng
		}
	}
	return nodeGroup
}

func (impl *K8sCapacityServiceImpl) getNodeDetail(ctx context.Context, node *corev1.Node, nodeResourceUsage map[string]corev1.ResourceList, podList *corev1.PodList, callForList bool, cluster *cluster.ClusterBean) (*bean.NodeCapacityDetail, error) {
	cpuAllocatable := node.Status.Allocatable[corev1.ResourceCPU]
	memoryAllocatable := node.Status.Allocatable[corev1.ResourceMemory]
	podCount := 0
	nodeRequestsResourceList := make(corev1.ResourceList)
	nodeLimitsResourceList := make(corev1.ResourceList)
	var podDetailList []*bean.PodCapacityDetail
	for _, pod := range podList.Items {
		if pod.Spec.NodeName == node.Name && pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
			if callForList {
				podCount++
			} else {
				var requests, limits corev1.ResourceList
				requests, limits = resourcehelper.PodRequestsAndLimits(&pod)
				nodeRequestsResourceList = AddTwoResourceList(nodeRequestsResourceList, requests)
				nodeLimitsResourceList = AddTwoResourceList(nodeLimitsResourceList, limits)
				podDetailList = append(podDetailList, getPodDetail(pod, cpuAllocatable, memoryAllocatable, limits, requests))
			}
		}
	}

	labels, taints := impl.getNodeLabelsAndTaints(node)
	nodeGroup := impl.getNodeGroup(node)
	nodeDetail := &bean.NodeCapacityDetail{
		Name:          node.Name,
		K8sVersion:    node.Status.NodeInfo.KubeletVersion,
		Errors:        findNodeErrors(node),
		InternalIp:    getNodeInternalIP(node),
		ExternalIp:    getNodeExternalIP(node),
		Unschedulable: node.Spec.Unschedulable,
		Roles:         findNodeRoles(node),
		Labels:        labels,
		Status:        findNodeStatus(node),
		CreatedAt:     node.CreationTimestamp.String(),
		ClusterName:   cluster.ClusterName,
		NodeGroup:     nodeGroup,
	}
	nodeDetail.Version = "v1"
	nodeDetail.Kind = "Node"
	nodeDetail.Taints = taints
	nodeUsageResourceList := nodeResourceUsage[node.Name]
	if callForList {
		// assigning additional data for node listing api call
		impl.updateBasicDetailsForNode(nodeDetail, node, podCount, nodeUsageResourceList, cpuAllocatable, memoryAllocatable)
	} else {
		//update data for node detail api call
		err := impl.updateAdditionalDetailForNode(nodeDetail, node, nodeLimitsResourceList, nodeRequestsResourceList, nodeUsageResourceList, podDetailList)
		if err != nil {
			impl.logger.Errorw("error in getting updating data for node detail", "err", err)
			return nil, err
		}
		err = impl.updateManifestData(ctx, nodeDetail, node, cluster.Id)
		if err != nil {
			return nil, err
		}
	}
	return nodeDetail, nil
}

func (impl *K8sCapacityServiceImpl) getNodeLabelsAndTaints(node *corev1.Node) ([]*bean.LabelAnnotationTaintObject, []*bean.LabelAnnotationTaintObject) {

	var labels []*bean.LabelAnnotationTaintObject
	taints := impl.getTaints(node)
	for k, v := range node.Labels {
		labelObj := &bean.LabelAnnotationTaintObject{
			Key:   k,
			Value: v,
		}
		labels = append(labels, labelObj)
	}
	return labels, taints
}

func (impl *K8sCapacityServiceImpl) getTaints(node *corev1.Node) []*bean.LabelAnnotationTaintObject {
	var taints []*bean.LabelAnnotationTaintObject
	for _, taint := range node.Spec.Taints {
		taintObj := &bean.LabelAnnotationTaintObject{
			Key:    taint.Key,
			Value:  taint.Value,
			Effect: string(taint.Effect),
		}
		taints = append(taints, taintObj)
	}
	return taints
}

func (impl *K8sCapacityServiceImpl) updateBasicDetailsForNode(nodeDetail *bean.NodeCapacityDetail, node *corev1.Node, podCount int, nodeUsageResourceList corev1.ResourceList, cpuAllocatable resource.Quantity, memoryAllocatable resource.Quantity) {
	nodeDetail.Age = translateTimestampSince(node.CreationTimestamp)
	nodeDetail.PodCount = podCount
	cpuUsage, cpuUsageOk := nodeUsageResourceList[corev1.ResourceCPU]
	memoryUsage, memoryUsageOk := nodeUsageResourceList[corev1.ResourceMemory]
	nodeDetail.Cpu = &bean.ResourceDetailObject{
		Allocatable:        getResourceString(cpuAllocatable, corev1.ResourceCPU),
		AllocatableInBytes: cpuAllocatable.Value(),
	}
	nodeDetail.Memory = &bean.ResourceDetailObject{
		Allocatable:        getResourceString(memoryAllocatable, corev1.ResourceMemory),
		AllocatableInBytes: memoryAllocatable.Value(),
	}
	if cpuUsageOk {
		nodeDetail.Cpu.Usage = getResourceString(cpuUsage, corev1.ResourceCPU)
		nodeDetail.Cpu.UsageInBytes = cpuUsage.Value()
		nodeDetail.Cpu.UsagePercentage = convertToPercentage(&cpuUsage, &cpuAllocatable)
	}
	if memoryUsageOk {
		nodeDetail.Memory.Usage = getResourceString(memoryUsage, corev1.ResourceMemory)
		nodeDetail.Memory.UsageInBytes = memoryUsage.Value()
		nodeDetail.Memory.UsagePercentage = convertToPercentage(&memoryUsage, &memoryAllocatable)
	}
}

func (impl *K8sCapacityServiceImpl) updateAdditionalDetailForNode(nodeDetail *bean.NodeCapacityDetail, node *corev1.Node,
	nodeLimitsResourceList corev1.ResourceList, nodeRequestsResourceList corev1.ResourceList,
	nodeUsageResourceList corev1.ResourceList, podDetailList []*bean.PodCapacityDetail) error {
	nodeDetail.Pods = podDetailList
	var annotations []*bean.LabelAnnotationTaintObject
	for k, v := range node.Annotations {
		annotationObj := &bean.LabelAnnotationTaintObject{
			Key:   k,
			Value: v,
		}
		annotations = append(annotations, annotationObj)
	}
	nodeDetail.Annotations = annotations
	impl.updateNodeConditions(node, nodeDetail)
	impl.updateNodeResources(node, nodeLimitsResourceList, nodeRequestsResourceList, nodeUsageResourceList, nodeDetail)
	return nil
}

func (impl *K8sCapacityServiceImpl) updateManifestData(ctx context.Context, nodeDetail *bean.NodeCapacityDetail, node *corev1.Node, clusterId int) error {
	//getting manifest
	manifestRequest := &k8s2.K8sRequestBean{
		ResourceIdentifier: k8s2.ResourceIdentifier{
			Name: node.Name,
			GroupVersionKind: schema.GroupVersionKind{
				Version: nodeDetail.Version,
				Kind:    nodeDetail.Kind,
			},
		},
	}
	request := &k8s.ResourceRequestBean{
		K8sRequest: manifestRequest,
		ClusterId:  clusterId,
	}
	//manifestResponse, err := impl.k8sClientService.GetResource(ctx, restConfig, manifestRequest)
	manifestResponse, err := impl.k8sCommonService.GetResource(ctx, request)
	if err != nil {
		impl.logger.Errorw("error in getting node manifest", "err", err)
		return err
	}
	nodeDetail.Manifest = manifestResponse.Manifest
	return nil
}

func (impl *K8sCapacityServiceImpl) updateNodeConditions(node *corev1.Node, nodeDetail *bean.NodeCapacityDetail) {
	//map of {conditionType : isErrorCondition }, Valid/Non-error conditions to be updated with update at kubernetes end
	NodeAllConditionsMap := map[corev1.NodeConditionType]bool{corev1.NodeReady: false, corev1.NodeMemoryPressure: true,
		corev1.NodeDiskPressure: true, corev1.NodeNetworkUnavailable: true, corev1.NodePIDPressure: true}
	var conditions []*bean.NodeConditionObject
	for _, condition := range node.Status.Conditions {
		if isErrorCondition, ok := NodeAllConditionsMap[condition.Type]; ok {
			conditionObj := &bean.NodeConditionObject{
				Type:    string(condition.Type),
				Reason:  condition.Reason,
				Message: condition.Message,
			}
			if (!isErrorCondition && condition.Status == corev1.ConditionTrue) || (isErrorCondition && condition.Status == corev1.ConditionFalse) {
				conditionObj.HaveIssue = false
			} else {
				conditionObj.HaveIssue = true
			}
			conditions = append(conditions, conditionObj)
		}
	}
	nodeDetail.Conditions = conditions
}

func (impl *K8sCapacityServiceImpl) updateNodeResources(node *corev1.Node, nodeLimitsResourceList corev1.ResourceList, nodeRequestsResourceList corev1.ResourceList, nodeUsageResourceList corev1.ResourceList, nodeDetail *bean.NodeCapacityDetail) {
	nodeCapacityResourceList := node.Status.Capacity
	nodeAllocatableResourceList := node.Status.Allocatable
	for resourceName, allocatable := range nodeAllocatableResourceList {
		limits, limitsOk := nodeLimitsResourceList[resourceName]
		requests, requestsOk := nodeRequestsResourceList[resourceName]
		usage, usageOk := nodeUsageResourceList[resourceName]
		capacity := nodeCapacityResourceList[resourceName]
		r := &bean.ResourceDetailObject{
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
}

func (impl *K8sCapacityServiceImpl) UpdateNodeManifest(ctx context.Context, request *bean.NodeUpdateRequestDto) (*k8s2.ManifestResponse, error) {
	manifestUpdateReq := &k8s2.K8sRequestBean{
		ResourceIdentifier: k8s2.ResourceIdentifier{
			Name: request.Name,
			GroupVersionKind: schema.GroupVersionKind{
				Group:   "",
				Version: request.Version,
				Kind:    request.Kind,
			},
		},
		Patch: request.ManifestPatch,
	}
	requestResourceBean := &k8s.ResourceRequestBean{K8sRequest: manifestUpdateReq, ClusterId: request.ClusterId}
	manifestResponse, err := impl.k8sCommonService.UpdateResource(ctx, requestResourceBean)
	if err != nil {
		impl.logger.Errorw("error in updating node manifest", "err", err)
		return nil, err
	}
	return manifestResponse, nil
}

func (impl *K8sCapacityServiceImpl) DeleteNode(ctx context.Context, request *bean.NodeUpdateRequestDto) (*k8s2.ManifestResponse, error) {
	deleteReq := &k8s2.K8sRequestBean{
		ResourceIdentifier: k8s2.ResourceIdentifier{
			Name: request.Name,
			GroupVersionKind: schema.GroupVersionKind{
				Group:   "",
				Version: request.Version,
				Kind:    request.Kind,
			},
		},
	}
	resourceRequest := &k8s.ResourceRequestBean{K8sRequest: deleteReq, ClusterId: request.ClusterId}
	// Here Sending userId as 0 as it appIdentifier is being sent nil so user id is not used in method. Update userid if appIdentifier is used
	manifestResponse, err := impl.k8sCommonService.DeleteResource(ctx, resourceRequest, 0)
	if err != nil {
		impl.logger.Errorw("error in deleting node", "err", err)
		return nil, err
	}
	return manifestResponse, nil
}

func (impl *K8sCapacityServiceImpl) CordonOrUnCordonNode(ctx context.Context, request *bean.NodeUpdateRequestDto) (string, error) {
	respMessage := ""
	cluster, err := impl.getClusterBean(request.ClusterId)
	if err != nil {
		return respMessage, err
	}
	//getting kubernetes clientSet by rest config
	_, _, k8sClientSet, err := impl.getK8sConfigAndClients(ctx, cluster)
	if err != nil {
		return respMessage, err
	}
	//get node
	node, err := impl.K8sUtil.GetNodeByName(ctx, k8sClientSet, request.Name)
	if err != nil {
		impl.logger.Errorw("error in getting node", "err", err)
		return respMessage, err
	}
	if node.Spec.Unschedulable == request.NodeCordonHelper.UnschedulableDesired {
		return respMessage, getErrorForCordonUpdateReq(request.NodeCordonHelper.UnschedulableDesired)
	}
	//updating node with desired cordon value
	node, err = k8s2.UpdateNodeUnschedulableProperty(request.NodeCordonHelper.UnschedulableDesired, node, k8sClientSet)
	if err != nil {
		impl.logger.Errorw("error in updating node", "err", err)
		return respMessage, err
	}

	if request.NodeCordonHelper.UnschedulableDesired {
		respMessage = fmt.Sprintf("Node successfully Cordoned.")
	} else {
		respMessage = fmt.Sprintf("Node successfully UnCordoned.")
	}
	return respMessage, nil
}

func (impl *K8sCapacityServiceImpl) DrainNode(ctx context.Context, request *bean.NodeUpdateRequestDto) (string, error) {
	impl.logger.Infow("received node drain request", "request", request)
	respMessage := ""
	cluster, err := impl.getClusterBean(request.ClusterId)
	if err != nil {
		return respMessage, err
	}
	//getting kubernetes clientSet by rest config
	_, _, k8sClientSet, err := impl.getK8sConfigAndClients(ctx, cluster)
	if err != nil {
		return respMessage, err
	}
	//get node
	node, err := impl.K8sUtil.GetNodeByName(context.Background(), k8sClientSet, request.Name)
	if err != nil {
		impl.logger.Errorw("error in getting node", "err", err)
		return respMessage, err
	}
	//checking if node is unschedulable or not, if not then need to unschedule before draining
	if !node.Spec.Unschedulable {
		node, err = k8s2.UpdateNodeUnschedulableProperty(true, node, k8sClientSet)
		if err != nil {
			impl.logger.Errorw("error in making node unschedulable", "err", err)
			return respMessage, err
		}
	}
	request.NodeDrainHelper.K8sClientSet = k8sClientSet
	err = impl.deleteOrEvictPods(request.Name, request.NodeDrainHelper)
	if err != nil {
		impl.logger.Errorw("error in deleting/evicting pods", "err", err, "nodeName", request.Name)
		return respMessage, err
	}
	respMessage = "Node Drained Successfully."
	return respMessage, nil
}

func (impl *K8sCapacityServiceImpl) getClusterBean(clusterId int) (*cluster.ClusterBean, error) {
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by ID", "err", err, "clusterId", clusterId)
		return nil, err
	}
	return cluster, err
}

func (impl *K8sCapacityServiceImpl) EditNodeTaints(ctx context.Context, request *bean.NodeUpdateRequestDto) (string, error) {
	respMessage := ""
	cluster, err := impl.getClusterBean(request.ClusterId)
	if err != nil {
		return respMessage, err
	}
	//getting kubernetes clientSet by rest config
	_, _, k8sClientSet, err := impl.getK8sConfigAndClients(ctx, cluster)
	if err != nil {
		return respMessage, err
	}
	err = validateTaintEditRequest(request.Taints)
	if err != nil {
		impl.logger.Errorw("error in validating taint edit request", "err", err, "requestTaints", request.Taints)
		return respMessage, err
	}
	//get node
	node, err := impl.K8sUtil.GetNodeByName(context.Background(), k8sClientSet, request.Name)
	if err != nil {
		impl.logger.Errorw("error in getting node", "err", err)
		return respMessage, err
	}
	node.Spec.Taints = request.Taints
	node, err = k8sClientSet.CoreV1().Nodes().Update(context.Background(), node, v1.UpdateOptions{})
	if err != nil {
		impl.logger.Errorw("error in updating taints in node", "err", err)
		return respMessage, err
	}
	respMessage = "Taints edited Successfully."
	return respMessage, nil
}

func (impl *K8sCapacityServiceImpl) GetNode(ctx context.Context, clusterId int, nodeName string) (*corev1.Node, error) {
	cluster, err := impl.getClusterBean(clusterId)
	if err != nil {
		return nil, err
	}
	//getting kubernetes clientSet by rest config
	_, _, k8sClientSet, err := impl.getK8sConfigAndClients(ctx, cluster)
	if err != nil {
		return nil, err
	}
	return impl.K8sUtil.GetNodeByName(context.Background(), k8sClientSet, nodeName)
}

func validateTaintEditRequest(reqTaints []corev1.Taint) error {
	if len(reqTaints) == 0 {
		return nil
	}
	var errs []error
	uniqueTaints := map[corev1.TaintEffect]sets.String{}
	for _, taint := range reqTaints {
		parseErr := parseTaint(taint)
		if parseErr != nil {
			errs = append(errs, parseErr)
		}
		// validate if taint is unique by <key, effect>
		if len(uniqueTaints[taint.Effect]) > 0 && uniqueTaints[taint.Effect].Has(taint.Key) {
			errs = append(errs, fmt.Errorf("duplicated taints with the same key and effect: %v", taint))
		}
		// add taint to existingTaints for uniqueness check
		if len(uniqueTaints[taint.Effect]) == 0 {
			uniqueTaints[taint.Effect] = sets.String{}
		}
		uniqueTaints[taint.Effect].Insert(taint.Key)
	}
	return utilerrors.NewAggregate(errs)
}

// parseTaint parses a taint from a string, whose form must be either
// '<key>=<value>:<effect>', '<key>:<effect>', or '<key>'.
func parseTaint(taint corev1.Taint) error {
	var key string
	var value string
	var effect corev1.TaintEffect
	var errs []error
	effect = taint.Effect
	if err := validateTaintEffect(effect); err != nil {
		errs = append(errs, err)
	}
	value = taint.Value
	if len(value) > 0 {
		if errStrs := validation.IsValidLabelValue(value); len(errStrs) > 0 {
			for _, errStr := range errStrs {
				errs = append(errs, fmt.Errorf(errStr))
			}
		}
	}
	key = taint.Key
	if errStrs := validation.IsQualifiedName(key); len(errStrs) > 0 {
		for _, errStr := range errStrs {
			errs = append(errs, fmt.Errorf(errStr))
		}
	}

	return utilerrors.NewAggregate(errs)
}

func validateTaintEffect(effect corev1.TaintEffect) error {
	if effect != corev1.TaintEffectNoSchedule && effect != corev1.TaintEffectPreferNoSchedule && effect != corev1.TaintEffectNoExecute {
		return fmt.Errorf("invalid taint effect: %v, unsupported taint effect", effect)
	}
	return nil
}

func (impl *K8sCapacityServiceImpl) deleteOrEvictPods(nodeName string, nodeDrainHelper *bean.NodeDrainHelper) error {
	impl.logger.Infow("received node drain - deleteOrEvictPods request", "nodeName", nodeName, "nodeDrainHelper", nodeDrainHelper)
	list, errs := GetPodsByNodeNameForDeletion(nodeName, nodeDrainHelper)
	if errs != nil {
		return utilerrors.NewAggregate(errs)
	}
	impl.logger.Infow("received pod list, deleteOrEvictPods", "podList", list)
	pods := list.Pods()
	if len(pods) == 0 {
		return nil
	}
	deleteOptions := v1.DeleteOptions{}
	if nodeDrainHelper.GracePeriodSeconds >= 0 {
		gracePeriodSecConverted := int64(nodeDrainHelper.GracePeriodSeconds)
		deleteOptions.GracePeriodSeconds = &gracePeriodSecConverted
	}
	if nodeDrainHelper.DisableEviction {
		//delete instead of eviction
		return impl.deletePods(pods, nodeDrainHelper.K8sClientSet, deleteOptions)
	} else {
		evictionGroupVersion, err := k8s2.CheckEvictionSupport(nodeDrainHelper.K8sClientSet)
		if err != nil {
			return err
		}
		if !evictionGroupVersion.Empty() {
			return impl.evictPods(pods, nodeDrainHelper.K8sClientSet, evictionGroupVersion, deleteOptions)
		}
	}
	return nil
}

func (impl *K8sCapacityServiceImpl) evictPods(pods []corev1.Pod, k8sClientSet *kubernetes.Clientset, evictionGroupVersion schema.GroupVersion, deleteOptions v1.DeleteOptions) error {
	impl.logger.Infow("receive pod eviction request", "pods", pods)
	returnCh := make(chan error, 1)
	for _, pod := range pods {
		impl.logger.Infow("evicting pod", "pod", pod)
		go func(pod corev1.Pod, returnCh chan error) {
			// Create a temporary pod, so we don't mutate the pod in the loop.
			activePod := pod
			err := k8s2.EvictPod(activePod, k8sClientSet, evictionGroupVersion, deleteOptions)
			if err == nil {
				returnCh <- nil
				return
			} else if apierrors.IsNotFound(err) {
				returnCh <- nil
				return
			} else if apierrors.IsTooManyRequests(err) {
				time.Sleep(5 * time.Second)
			} else {
				returnCh <- fmt.Errorf("error when evicting pods/%q -n %q: %v", activePod.Name, activePod.Namespace, err)
				return
			}
		}(pod, returnCh)
	}
	doneCount := 0
	var errors []error
	numPods := len(pods)
	for doneCount < numPods {
		select {
		case err := <-returnCh:
			doneCount++
			if err != nil {
				impl.logger.Errorw("error in pod eviction", "err", err)
				errors = append(errors, err)
			}
		}
	}
	return utilerrors.NewAggregate(errors)
}

func (impl *K8sCapacityServiceImpl) deletePods(pods []corev1.Pod, k8sClientSet *kubernetes.Clientset, deleteOptions v1.DeleteOptions) error {
	impl.logger.Infow("received pod deletion request", "pods", pods)
	var podDeletionErrors []error
	for _, pod := range pods {
		impl.logger.Infow("deleting pod", "pod", pod)
		err := k8s2.DeletePod(pod, k8sClientSet, deleteOptions)
		if err != nil && !apierrors.IsNotFound(err) {
			podDeletionErrors = append(podDeletionErrors, err)
		}
	}
	if len(podDeletionErrors) > 0 {
		return utilerrors.NewAggregate(podDeletionErrors)
	}
	return nil
}

func getErrorForCordonUpdateReq(desired bool) error {
	if desired {
		return fmt.Errorf("node already cordoned")
	}
	return fmt.Errorf("node already uncordoned")
}

func GetPodsByNodeNameForDeletion(nodeName string, nodeDrainHelper *bean.NodeDrainHelper) (*bean.PodDeleteList, []error) {
	initialOpts := v1.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": nodeName}).String(),
	}
	podList, err := nodeDrainHelper.K8sClientSet.CoreV1().Pods(corev1.NamespaceAll).List(context.Background(), initialOpts)
	if err != nil {
		return nil, []error{err}
	}

	list := bean.FilterPods(podList, nodeDrainHelper.MakeFilters())
	if errs := list.Errors(); len(errs) > 0 {
		return list, errs
	}

	return list, nil
}

func getPodDetail(pod corev1.Pod, cpuAllocatable resource.Quantity, memoryAllocatable resource.Quantity, limits corev1.ResourceList, requests corev1.ResourceList) *bean.PodCapacityDetail {
	cpuLimits, cpuLimitsOk := limits[corev1.ResourceCPU]
	cpuRequests, cpuRequestsOk := requests[corev1.ResourceCPU]
	memoryLimits, memoryLimitsOk := limits[corev1.ResourceMemory]
	memoryRequests, memoryRequestsOk := requests[corev1.ResourceMemory]
	podDetail := &bean.PodCapacityDetail{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Age:       translateTimestampSince(pod.CreationTimestamp),
		CreatedAt: pod.CreationTimestamp.String(),
		Cpu: &bean.ResourceDetailObject{
			Limit:   getResourceString(cpuLimits, corev1.ResourceCPU),
			Request: getResourceString(cpuRequests, corev1.ResourceCPU),
		},
		Memory: &bean.ResourceDetailObject{
			Limit:   getResourceString(memoryLimits, corev1.ResourceMemory),
			Request: getResourceString(memoryRequests, corev1.ResourceMemory),
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

func getResourceString(quantity resource.Quantity, resourceName corev1.ResourceName) string {
	standardResources := map[corev1.ResourceName]bool{corev1.ResourceCPU: true, corev1.ResourceMemory: true, corev1.ResourceStorage: true, corev1.ResourceEphemeralStorage: true}

	if _, ok := standardResources[resourceName]; !ok {
		//not a standard resource, we do not know if conversion would be valid or not
		//for example - pods: "250", this is not in bytes but an integer so conversion is invalid
		return quantity.String()
	} else {
		var quantityStr string
		value := quantity.Value()
		valueGi := value / bean.Gibibyte
		//allowing remainder 0 only, because for Gi rounding off will be highly erroneous
		if valueGi > 1 && value%bean.Gibibyte == 0 {
			quantityStr = fmt.Sprintf("%dGi", valueGi)
		} else {
			valueMi := value / bean.Mebibyte
			if valueMi > 10 {
				if value%bean.Mebibyte != 0 {
					valueMi++
				}
				quantityStr = fmt.Sprintf("%dMi", valueMi)
			} else if value > 1000 {
				valueKi := value / bean.Kibibyte
				if value%bean.Kibibyte != 0 {
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

func findNodeRoles(node *corev1.Node) []string {
	roles := sets.NewString()
	for k, v := range node.Labels {
		switch {
		case strings.HasPrefix(k, bean.LabelNodeRolePrefix):
			if role := strings.TrimPrefix(k, bean.LabelNodeRolePrefix); len(role) > 0 {
				roles.Insert(role)
			}
		case k == bean.NodeLabelRole && v != "":
			roles.Insert(v)
		}
	}
	if roles.Len() > 0 {
		return roles.List()
	} else {
		return []string{"none"}
	}
}

func findNodeStatus(node *corev1.Node) string {
	conditionMap := make(map[corev1.NodeConditionType]*corev1.NodeCondition)
	//Valid conditions to be updated with update at kubernetes end
	NodeAllValidConditions := []corev1.NodeConditionType{corev1.NodeReady}
	for _, condition := range node.Status.Conditions {
		conditionMap[condition.Type] = &condition
	}
	var status string
	for _, validCondition := range NodeAllValidConditions {
		if condition, ok := conditionMap[validCondition]; ok {
			if condition.Status == corev1.ConditionTrue {
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

func findNodeErrors(node *corev1.Node) map[corev1.NodeConditionType]string {
	conditionMap := make(map[corev1.NodeConditionType]corev1.NodeCondition)
	NodeAllErrorConditions := []corev1.NodeConditionType{corev1.NodeMemoryPressure, corev1.NodeDiskPressure, corev1.NodeNetworkUnavailable, corev1.NodePIDPressure}
	for _, condition := range node.Status.Conditions {
		conditionMap[condition.Type] = condition
	}
	conditionErrorMap := make(map[corev1.NodeConditionType]string)
	for _, errorCondition := range NodeAllErrorConditions {
		if condition, ok := conditionMap[errorCondition]; ok {
			if condition.Status == corev1.ConditionTrue {
				conditionErrorMap[condition.Type] = condition.Message
			}
		}
	}
	return conditionErrorMap
}

func getNodeExternalIP(node *corev1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeExternalIP {
			return address.Address
		}
	}
	return "none"
}

func getNodeInternalIP(node *corev1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			return address.Address
		}
	}
	return "none"
}

func AddTwoResourceList(oldResourceList corev1.ResourceList, newResourceList corev1.ResourceList) corev1.ResourceList {
	for res, quantity := range newResourceList {
		if oldQuantity, ok1 := oldResourceList[res]; ok1 {
			quantity.Add(oldQuantity)
		}
		oldResourceList[res] = quantity
	}
	return oldResourceList
}
