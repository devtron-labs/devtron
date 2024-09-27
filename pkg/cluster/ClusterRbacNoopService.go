package cluster

type ClusterRbacNoopServiceImpl struct {
}

func NewClusterRbacNoopServiceImpl() *ClusterRbacNoopServiceImpl {
	return &ClusterRbacNoopServiceImpl{}
}

func (impl ClusterRbacNoopServiceImpl) CheckAuthorization(clusterName string, clusterId int, token string, userId int32, rbacForClusterMappingsAlso bool) (bool, error) {
	return true, nil
}

