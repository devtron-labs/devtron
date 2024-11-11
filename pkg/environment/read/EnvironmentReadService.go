package read

import (
	bean2 "github.com/devtron-labs/devtron/pkg/environment/bean"
	"github.com/devtron-labs/devtron/pkg/environment/repository"
	"go.uber.org/zap"
)

type EnvironmentReadService interface {
	GetAll() ([]bean2.EnvironmentBean, error)
}

type EnvironmentReadServiceImpl struct {
	logger                *zap.SugaredLogger
	environmentRepository repository.EnvironmentRepository
}

func NewEnvironmentReadServiceImpl(logger *zap.SugaredLogger,
	environmentRepository repository.EnvironmentRepository) *EnvironmentReadServiceImpl {
	return &EnvironmentReadServiceImpl{
		logger:                logger,
		environmentRepository: environmentRepository,
	}
}
func (impl *EnvironmentReadServiceImpl) GetAll() ([]bean2.EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []bean2.EnvironmentBean
	for _, model := range models {
		beans = append(beans, bean2.EnvironmentBean{
			Id:                    model.Id,
			Environment:           model.Name,
			ClusterId:             model.Cluster.Id,
			ClusterName:           model.Cluster.ClusterName,
			Active:                model.Active,
			PrometheusEndpoint:    model.Cluster.PrometheusEndpoint,
			Namespace:             model.Namespace,
			Default:               model.Default,
			CdArgoSetup:           model.Cluster.CdArgoSetup,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
			Description:           model.Description,
			IsVirtualEnvironment:  model.IsVirtualEnvironment,
		})
	}
	return beans, nil
}
