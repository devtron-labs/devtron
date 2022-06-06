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
	GetNodeCapacityDetailsListByClusterId(clusterId int) ([]*NodeCapacityDetails, error)
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
	nodeList, err := k8sClientSet.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
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
		clusterCpuCapacity.Add(node.Status.Capacity["cpu"])
		clusterMemoryCapacity.Add(node.Status.Capacity["memory"])
		clusterCpuAllocatable.Add(node.Status.Allocatable["cpu"])
		clusterMemoryAllocatable.Add(node.Status.Allocatable["memory"])
	}
	clusterDetail.Cpu = &ResourceDetailObject{
		Capacity: clusterCpuCapacity.String(),
	}
	clusterDetail.Memory = &ResourceDetailObject{
		Capacity: clusterCpuCapacity.String(),
	}
	if callForList {
		//todo: add cluster connection status
		// assigning additional data for cluster listing api call
		clusterDetail.NodeK8sVersions = nodesK8sVersion
		clusterDetail.NodeCount = nodeCount
	} else {
		//load and assign data for cluster detail api call
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
			clusterCpuUsage.Add(nm.Usage["cpu"])
			clusterMemoryUsage.Add(nm.Usage["memory"])
		}
		for _, pod := range podList.Items {
			requests, limits := resourcehelper.PodRequestsAndLimits(&pod)
			clusterCpuLimits.Add(limits["cpu"])
			clusterCpuRequests.Add(requests["cpu"])
			clusterMemoryLimits.Add(limits["memory"])
			clusterMemoryRequests.Add(requests["memory"])
		}
		clusterDetail.Cpu.RequestPercentage = convertToPercentage(clusterCpuRequests, clusterCpuAllocatable)
		clusterDetail.Cpu.LimitPercentage = convertToPercentage(clusterCpuLimits, clusterCpuAllocatable)
		clusterDetail.Cpu.UsagePercentage = convertToPercentage(clusterCpuUsage, clusterCpuAllocatable)
		clusterDetail.Memory.RequestPercentage = convertToPercentage(clusterMemoryRequests, clusterMemoryAllocatable)
		clusterDetail.Memory.LimitPercentage = convertToPercentage(clusterMemoryLimits, clusterMemoryAllocatable)
		clusterDetail.Memory.UsagePercentage = convertToPercentage(clusterMemoryUsage, clusterMemoryAllocatable)
	}
	return clusterDetail, nil
}

func (impl *K8sCapacityServiceImpl) GetNodeCapacityDetailsListByClusterId(clusterId int) ([]*NodeCapacityDetails, error) {
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
	var nodeCpuUsage map[string]resource.Quantity
	var nodeMemoryUsage map[string]resource.Quantity
	nodeMetricsList, err := metricsClientSet.MetricsV1beta1().NodeMetricses().List(context.Background(), v1.ListOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting node metrics", "err", err)
		return nil, err
	}
	for _, nodeMetrics := range nodeMetricsList.Items {
		nodeCpuUsage[nodeMetrics.Name] = nodeMetrics.Usage["cpu"]
		nodeMemoryUsage[nodeMetrics.Name] = nodeMetrics.Usage["memory"]
	}
	var nodeDetails []*NodeCapacityDetails
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
	for _, node := range nodeList.Items {
		tmpPodCount := 0
		for _, pod := range podList.Items {
			if pod.Spec.NodeName == node.Name {
				tmpPodCount++
			}
		}
		cpuAllocatable := node.Status.Allocatable["cpu"]
		memoryAllocatable := node.Status.Allocatable["memory"]
		//TODO: add node status, errors
		nodeDetail := &NodeCapacityDetails{
			Name:       node.Name,
			PodCount:   tmpPodCount,
			TaintCount: len(node.Spec.Taints),
			K8sVersion: node.Status.NodeInfo.KubeletVersion,
			Cpu: &ResourceDetailObject{
				Allocatable: cpuAllocatable.String(),
			},
			Memory: &ResourceDetailObject{
				Allocatable: memoryAllocatable.String(),
			},
			Age: translateTimestampSince(node.CreationTimestamp),
		}
		roles := findNodeRoles(node)
		if len(roles) == 0 {
			roles = []string{"<none>"}
		}
		nodeDetail.Roles = roles
		if cpuUsage, ok := nodeCpuUsage[node.Name]; ok {
			nodeDetail.Cpu.Usage = cpuUsage.String()
			nodeDetail.Cpu.UsagePercentage = convertToPercentage(cpuUsage, cpuAllocatable)
		}
		if memoryUsage, ok := nodeMemoryUsage[node.Name]; ok {
			nodeDetail.Memory.Usage = memoryUsage.String()
			nodeDetail.Memory.UsagePercentage = convertToPercentage(memoryUsage, memoryAllocatable)
		}
		nodeDetails = append(nodeDetails, nodeDetail)
	}
	return nodeDetails, nil
}

func convertToPercentage(actual, allocatable resource.Quantity) string {
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
	return roles.List()
}

func findNodeStatus(node metav1.Node) string {
	conditionMap := make(map[metav1.NodeConditionType]*metav1.NodeCondition)
	NodeAllValidConditions := []metav1.NodeConditionType{metav1.NodeReady}
	for _, condition := range node.Status.Conditions {
		conditionMap[condition.Type] = &condition
	}
	var status string
	for _, validCondition := range NodeAllValidConditions {
		if condition, ok := conditionMap[validCondition]; ok {
			if condition.Status == metav1.ConditionTrue {
				status = string(condition.Type)
			}
		}
	}
	if len(status) == 0 {
		status = "Unknown"
	}
	return status
}
