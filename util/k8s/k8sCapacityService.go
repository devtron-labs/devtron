package k8s

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"go.uber.org/zap"
	metav1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"strings"
	"time"
)

const (
	labelNodeRolePrefix = "node-role.kubernetes.io/"
	nodeLabelRole       = "kubernetes.io/role"
)

type K8sCapacityService interface {
	GetClusterCapacityDetailList() ([]*ClusterCapacityDetail, error)
	GetClusterCapacityDetailById(clusterId int, callForList bool) (*ClusterCapacityDetail, error)
	GetNodeCapacityDetailsListByClusterId(clusterId int) ([]*NodeCapacityDetail, error)
}
type K8sCapacityServiceImpl struct {
	logger                *zap.SugaredLogger
	clusterService        cluster.ClusterService
	k8sApplicationService K8sApplicationService
}

func NewK8sCapacityServiceImpl(Logger *zap.SugaredLogger,
	clusterService cluster.ClusterService,
	k8sApplicationService K8sApplicationService) *K8sCapacityServiceImpl {
	return &K8sCapacityServiceImpl{
		logger:                Logger,
		clusterService:        clusterService,
		k8sApplicationService: k8sApplicationService,
	}
}

func (impl *K8sCapacityServiceImpl) GetClusterCapacityDetailList() ([]*ClusterCapacityDetail, error) {
	clusters, err := impl.clusterService.FindAll()
	if err != nil {
		impl.logger.Errorw("error in getting all clusters", "err", err)
		return nil, err
	}
	var clustersDetails []*ClusterCapacityDetail
	for _, cluster := range clusters {
		clusterCapacityDetail, err := impl.GetClusterCapacityDetailById(cluster.Id, true)
		if err != nil {
			impl.logger.Errorw("error in getting cluster capacity details by id", "err", err)
			return nil, err
		}
		clusterCapacityDetail.Id = cluster.Id
		clusterCapacityDetail.Name = cluster.ClusterName
		clustersDetails = append(clustersDetails, clusterCapacityDetail)
	}
	return clustersDetails, nil
}

func (impl *K8sCapacityServiceImpl) GetClusterCapacityDetailById(clusterId int, callForList bool) (*ClusterCapacityDetail, error) {
	//getting rest config by clusterId
	restConfig, err := impl.k8sApplicationService.GetRestConfigByClusterId(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster id", "err", err, "clusterId", clusterId)
		return nil, err
	}
	//getting kubernetes clientSet by rest config
	k8sClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting client set by rest config", "err", err, "restConfig", restConfig)
		return nil, err
	}
	clusterDetail := &ClusterCapacityDetail{}
	nodeList, errorInNodeListing := k8sClientSet.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting node list", "err", err)
		return nil, err
	}
	var clusterCpuCapacity resource.Quantity
	var clusterMemoryCapacity resource.Quantity
	var clusterCpuAllocatable resource.Quantity
	var clusterMemoryAllocatable resource.Quantity
	nodeCount := 0
	nodesK8sVersionMap := make(map[string]bool)
	var nodesK8sVersion []string
	for _, node := range nodeList.Items {
		if callForList {
			nodeCount += 1
			if _, ok := nodesK8sVersionMap[node.Status.NodeInfo.KubeletVersion]; !ok {
				nodesK8sVersionMap[node.Status.NodeInfo.KubeletVersion] = true
				nodesK8sVersion = append(nodesK8sVersion, node.Status.NodeInfo.KubeletVersion)
			}
		}
		clusterCpuCapacity.Add(node.Status.Capacity[metav1.ResourceCPU])
		clusterMemoryCapacity.Add(node.Status.Capacity[metav1.ResourceMemory])
		clusterCpuAllocatable.Add(node.Status.Allocatable[metav1.ResourceCPU])
		clusterMemoryAllocatable.Add(node.Status.Allocatable[metav1.ResourceMemory])
	}
	clusterDetail.Cpu = &ResourceDetailObject{
		Capacity: clusterCpuCapacity.String(),
	}
	clusterDetail.Memory = &ResourceDetailObject{
		Capacity: clusterCpuCapacity.String(),
	}
	if callForList {
		//assigning additional data for cluster listing api call
		errorInNodeListingBool := false
		if errorInNodeListing != nil {
			errorInNodeListingBool = true
		}
		clusterDetail.ErrorInNodeListing = &errorInNodeListingBool
		clusterDetail.NodeK8sVersions = nodesK8sVersion
		clusterDetail.NodeCount = nodeCount
	} else {
		//update data for cluster detail api call
		//getting metrics clientSet by rest config
		metricsClientSet, err := metrics.NewForConfig(restConfig)
		if err != nil {
			impl.logger.Errorw("error in getting metrics client set", "err", err)
			return nil, err
		}
		nmList, err := metricsClientSet.MetricsV1beta1().NodeMetricses().List(context.Background(), v1.ListOptions{})
		if err != nil {
			impl.logger.Errorw("error in getting nodeMetrics list", "err", err)
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
		for _, nm := range nmList.Items {
			clusterCpuUsage.Add(nm.Usage[metav1.ResourceCPU])
			clusterMemoryUsage.Add(nm.Usage[metav1.ResourceMemory])
		}
		for _, pod := range podList.Items {
			requests, limits := resourcehelper.PodRequestsAndLimits(&pod)
			clusterCpuLimits.Add(limits[metav1.ResourceCPU])
			clusterCpuRequests.Add(requests[metav1.ResourceCPU])
			clusterMemoryLimits.Add(limits[metav1.ResourceMemory])
			clusterMemoryRequests.Add(requests[metav1.ResourceMemory])
		}
		clusterDetail.Cpu.RequestPercentage = convertToPercentage(&clusterCpuRequests, &clusterCpuAllocatable)
		clusterDetail.Cpu.LimitPercentage = convertToPercentage(&clusterCpuLimits, &clusterCpuAllocatable)
		clusterDetail.Cpu.UsagePercentage = convertToPercentage(&clusterCpuUsage, &clusterCpuAllocatable)
		clusterDetail.Memory.RequestPercentage = convertToPercentage(&clusterMemoryRequests, &clusterMemoryAllocatable)
		clusterDetail.Memory.LimitPercentage = convertToPercentage(&clusterMemoryLimits, &clusterMemoryAllocatable)
		clusterDetail.Memory.UsagePercentage = convertToPercentage(&clusterMemoryUsage, &clusterMemoryAllocatable)
	}
	return clusterDetail, nil
}

func (impl *K8sCapacityServiceImpl) GetNodeCapacityDetailsListByClusterId(clusterId int) ([]*NodeCapacityDetail, error) {
	//getting rest config by clusterId
	restConfig, err := impl.k8sApplicationService.GetRestConfigByClusterId(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster id", "err", err, "clusterId", clusterId)
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
		return nil, err
	}
	nodeList, err := k8sClientSet.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
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
	for _, nodeMetrics := range nodeMetricsList.Items {
		nodeResourceUsage[nodeMetrics.Name] = nodeMetrics.Usage
	}
	var nodeDetails []*NodeCapacityDetail
	for _, node := range nodeList.Items {
		nodeDetail := impl.GetNodeDetail(node, nodeResourceUsage, podList, true)
		nodeDetails = append(nodeDetails, nodeDetail)
	}
	return nodeDetails, nil
}

func (impl *K8sCapacityServiceImpl) GetNodeDetail(node metav1.Node, nodeResourceUsage map[string]metav1.ResourceList, podList *metav1.PodList, callFromList bool) *NodeCapacityDetail {
	cpuAllocatable := node.Status.Allocatable[metav1.ResourceCPU]
	memoryAllocatable := node.Status.Allocatable[metav1.ResourceMemory]
	podCount := 0
	var nodeRequestsResourceList metav1.ResourceList
	var nodeLimitsResourceList metav1.ResourceList
	var podDetailList []*PodCapacityDetail
	for _, pod := range podList.Items {
		if pod.Spec.NodeName == node.Name {
			if callFromList {
				podCount++
			} else {
				requests, limits := resourcehelper.PodRequestsAndLimits(&pod)
				podDetailList = append(podDetailList, getPodDetail(pod, cpuAllocatable, memoryAllocatable, limits, requests))
				nodeRequestsResourceList = AddTwoResourceList(nodeRequestsResourceList, requests)
				nodeLimitsResourceList = AddTwoResourceList(nodeLimitsResourceList, limits)
			}
		}
	}
	nodeDetail := &NodeCapacityDetail{
		Name:            node.Name,
		K8sVersion:      node.Status.NodeInfo.KubeletVersion,
		Errors:          findNodeErrors(node),
		InternalIp:      getNodeInternalIP(node),
		ExternalIp:      getNodeExternalIP(node),
		Unschedulable:   node.Spec.Unschedulable,
		Roles:           findNodeRoles(node),
		ResourceVersion: node.ResourceVersion,
	}
	nodeUsageResourceList := nodeResourceUsage[node.Name]
	if callFromList {
		// assigning additional data for node listing api call
		nodeDetail.Status = findNodeStatus(node)
		nodeDetail.Age = translateTimestampSince(node.CreationTimestamp)
		nodeDetail.PodCount = podCount
		nodeDetail.TaintCount = len(node.Spec.Taints)
		cpuUsage := nodeUsageResourceList[metav1.ResourceCPU]
		memoryUsage := nodeUsageResourceList[metav1.ResourceMemory]
		nodeDetail.Cpu = &ResourceDetailObject{
			Allocatable:     cpuAllocatable.String(),
			Usage:           cpuUsage.String(),
			UsagePercentage: convertToPercentage(&cpuUsage, &cpuAllocatable),
		}
		nodeDetail.Memory = &ResourceDetailObject{
			Allocatable:     memoryAllocatable.String(),
			Usage:           memoryUsage.String(),
			UsagePercentage: convertToPercentage(&memoryUsage, &memoryAllocatable),
		}

	} else {
		//update data for node detail api call
		updateAdditionalDetailForNode(nodeDetail, node, nodeLimitsResourceList, nodeRequestsResourceList, nodeUsageResourceList, podDetailList)
	}
	return nodeDetail
}

func updateAdditionalDetailForNode(nodeDetail *NodeCapacityDetail, node metav1.Node, nodeLimitsResourceList metav1.ResourceList,
	nodeRequestsResourceList metav1.ResourceList, nodeUsageResourceList metav1.ResourceList, podDetailList []*PodCapacityDetail) {
	nodeDetail.Pods = podDetailList
	nodeDetail.CreatedAt = node.CreationTimestamp.String()
	var labels []*LabelAnnotationTaintObject
	for k, v := range node.Labels {
		labelObj := &LabelAnnotationTaintObject{
			Key:   k,
			Value: v,
		}
		labels = append(labels, labelObj)
	}
	nodeDetail.Labels = labels

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
		annotations = append(annotations, taintObj)
	}
	nodeDetail.Taints = taints
	//Valid conditions to be updated with update at kubernetes end
	NodeAllValidConditionsMap := map[metav1.NodeConditionType]bool{metav1.NodeReady: true}
	var conditions []*NodeConditionObject
	for _, condition := range node.Status.Conditions {
		conditionObj := &NodeConditionObject{
			Type:    string(condition.Type),
			Reason:  condition.Reason,
			Message: condition.Message,
		}
		_, ok := NodeAllValidConditionsMap[condition.Type]
		if (ok && condition.Status == metav1.ConditionTrue) || (!ok && condition.Status == metav1.ConditionFalse) {
			conditionObj.HaveIssue = false
		} else {
			conditionObj.HaveIssue = true
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
			Capacity: capacity.String(),
		}
		if limitsOk {
			r.Limit = limits.String()
			r.Limit = convertToPercentage(&limits, &allocatable)
		}
		if requestsOk {
			r.Request = requests.String()
			r.RequestPercentage = convertToPercentage(&requests, &allocatable)
		}
		if usageOk {
			r.Usage = usage.String()
			r.UsagePercentage = convertToPercentage(&usage, &allocatable)
		}
		nodeDetail.Resources = append(nodeDetail.Resources, r)
	}
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
		Cpu: &ResourceDetailObject{
			Limit:   cpuLimits.String(),
			Request: cpuRequests.String(),
		},
		Memory: &ResourceDetailObject{
			Limit:   memoryLimits.String(),
			Request: memoryRequests.String(),
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
	return fmt.Sprintf("%d%%%%", int64(utilPercent))
}

func translateTimestampSince(timestamp v1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}
	return duration.HumanDuration(time.Since(timestamp.Time))
}

func findNodeRoles(node metav1.Node) []string {
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
		return []string{"<none>"}
	}
}

func findNodeStatus(node metav1.Node) string {
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
				status = "Not " + string(condition.Type)
			}
		}
	}
	if len(status) == 0 {
		status = "Unknown"
	}
	return status
}

func findNodeErrors(node metav1.Node) map[metav1.NodeConditionType]string {
	conditionMap := make(map[metav1.NodeConditionType]*metav1.NodeCondition)
	NodeAllErrorConditions := []metav1.NodeConditionType{metav1.NodeMemoryPressure, metav1.NodeDiskPressure, metav1.NodeNetworkUnavailable, metav1.NodePIDPressure}
	for _, condition := range node.Status.Conditions {
		conditionMap[condition.Type] = &condition
	}
	conditionErrorMap := make(map[metav1.NodeConditionType]string)
	for _, errorCondition := range NodeAllErrorConditions {
		if condition, ok := conditionMap[errorCondition]; ok {
			if condition.Status == metav1.ConditionTrue {
				conditionErrorMap[condition.Type] = fmt.Sprint(condition.Reason + " - " + condition.Message)
			}
		}
	}
	return conditionErrorMap
}

func getNodeExternalIP(node metav1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == metav1.NodeExternalIP {
			return address.Address
		}
	}
	return "<none>"
}

func getNodeInternalIP(node metav1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == metav1.NodeInternalIP {
			return address.Address
		}
	}
	return "<none>"
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
