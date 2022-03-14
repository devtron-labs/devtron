package k8s

import (
	"github.com/devtron-labs/devtron/pkg/cluster"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type K8sCapacityService interface {
	GetClusterCapacityDetailsForAllClusters() ([]*ClusterCapacityDetails, error)
	GetClusterCapacityDetailsByClusterId(clusterId int) (*ClusterCapacityDetails, error)
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

type ClusterCapacityDetails struct {
	Cluster            *cluster.ClusterBean `json:"cluster"`
	NodeCount          int                  `json:"nodeCount"`
	NodesK8sVersion    []string             `json:"nodesK8sVersion"`
	TotalClusterCpu    string               `json:"totalClusterCpu"`
	TotalClusterMemory string               `json:"totalClusterMemory"`
}

type NodeCapacityDetails struct {
	Name            string            `json:"name"`
	StatusReasonMap map[string]string `json:"statusReasonMap"`
	PodCount        int               `json:"podCount"`
	TaintCount      int               `json:"taintCount"`
	CpuCapacity     string            `json:"cpuCapacity"`
	MemoryCapacity  string            `json:"memoryCapacity"`
}

type ResourceMetric struct {
	ResourceType string `json:"resourceType"`
	Allocatable  string `json:"allocatable"`
	Utilization  string `json:"utilization"`
	Request      string `json:"request"`
	Limit        string `json:"limit"`
}

func (impl *K8sCapacityServiceImpl) GetClusterCapacityDetailsForAllClusters() ([]*ClusterCapacityDetails, error) {
	clusters, err := impl.clusterService.FindAll()
	if err != nil {
		impl.logger.Errorw("error in getting all clusters", "err", err)
		return nil, err
	}
	var clustersDetails []*ClusterCapacityDetails
	for _, cluster := range clusters {
		clusterCapacityDetails, err := impl.GetClusterCapacityDetailsByClusterId(cluster.Id)
		if err != nil {
			impl.logger.Errorw("error in getting cluster capacity details by id", "err", err)
			return nil, err
		}
		clusterCapacityDetails.Cluster = cluster
		clustersDetails = append(clustersDetails, clusterCapacityDetails)
	}
	return clustersDetails, nil
}

func (impl *K8sCapacityServiceImpl) GetClusterCapacityDetailsByClusterId(clusterId int) (*ClusterCapacityDetails, error) {
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
	clusterDetails := &ClusterCapacityDetails{}
	nodeList, err := k8sClientSet.CoreV1().Nodes().List(v1.ListOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting node list", "err", err)
		return nil, err
	}
	//TODO: add node status
	var clusterCpuCapacity resource.Quantity
	var clusterMemoryCapacity resource.Quantity
	nodeCount := 0
	for _, node := range nodeList.Items {
		nodeCount += 1
		clusterDetails.NodesK8sVersion = append(clusterDetails.NodesK8sVersion, node.ResourceVersion)
		clusterCpuCapacity.Add(node.Status.Capacity["cpu"])
		clusterMemoryCapacity.Add(node.Status.Capacity["memory"])
	}
	clusterDetails.NodeCount = nodeCount
	clusterDetails.TotalClusterCpu = clusterCpuCapacity.String()
	clusterDetails.TotalClusterMemory = clusterMemoryCapacity.String()
	return clusterDetails, nil
}

func (impl *K8sCapacityServiceImpl) GetNodeCapacityDetailsListByClusterId(clusterId int) error {
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
	var nodeDetails []*NodeCapacityDetails
	nodeList, err := k8sClientSet.CoreV1().Nodes().List(v1.ListOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting node list", "err", err)
		return nil, err
	}
}
