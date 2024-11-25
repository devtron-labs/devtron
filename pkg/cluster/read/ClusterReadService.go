package read

import (
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
)

type ClusterReadService interface {
	IsClusterReachable(clusterId int) (bool, error)
}

type ClusterReadServiceImpl struct {
	logger            *zap.SugaredLogger
	clusterRepository repository.ClusterRepository
}

func NewClusterReadServiceImpl(logger *zap.SugaredLogger,
	clusterRepository repository.ClusterRepository) *ClusterReadServiceImpl {
	return &ClusterReadServiceImpl{
		logger:            logger,
		clusterRepository: clusterRepository,
	}
}

func (impl *ClusterReadServiceImpl) IsClusterReachable(clusterId int) (bool, error) {
	cluster, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in finding cluster from clusterId", "envId", clusterId)
		return false, err
	}
	if len(cluster.ErrorInConnecting) > 0 {
		return false, nil
	}
	return true, nil

}
