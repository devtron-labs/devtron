package app

import (
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/app/bean"
	"go.uber.org/zap"
)

type AppConfigService interface {
	FindById(appId int) (*bean.AppBean, error)
}

type AppConfigServiceImpl struct {
	logger        *zap.SugaredLogger
	appRepository appRepository.AppRepository
}

func NewAppConfigServiceImpl(logger *zap.SugaredLogger, appRepository appRepository.AppRepository) *AppConfigServiceImpl {
	return &AppConfigServiceImpl{
		logger:        logger,
		appRepository: appRepository,
	}
}

func (impl *AppConfigServiceImpl) FindById(appId int) (*bean.AppBean, error) {
	appEntity, err := impl.appRepository.FindById(appId)
	if err != nil {
		impl.logger.Errorw("error occurred while finding app", "appId", appId, "err", err)
		return nil, err
	}
	appBean := bean.InitFromAppEntity(appEntity)
	return appBean, nil
}
