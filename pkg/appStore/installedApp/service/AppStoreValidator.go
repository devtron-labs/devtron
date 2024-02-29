package service

import (
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/repository/bean"
)

type AppStoreValidator interface {
	Validate(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, environment *bean2.EnvironmentBean) error
}

type AppStoreValidatorImpl struct {
}

func NewAppAppStoreValidatorImpl() *AppStoreValidatorImpl {
	return &AppStoreValidatorImpl{}
}

func (impl *AppStoreValidatorImpl) Validate(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, environment *bean2.EnvironmentBean) error {
	return nil
}
