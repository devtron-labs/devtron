package service

import (
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
)

type AppStoreValidator interface {
	Validate(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, environment *repository.Environment) error
}

type AppStoreValidatorImpl struct {
	logger *zap.SugaredLogger
}

func NewAppAppStoreValidatorImpl(
	logger *zap.SugaredLogger,
) *AppStoreValidatorImpl {
	return &AppStoreValidatorImpl{
		logger: logger,
	}
}

func (impl *AppStoreValidatorImpl) Validate(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, environment *repository.Environment) error {
	return nil
}
