/*
 * Copyright (c) 2024. Devtron Inc.
 */

package service

import (
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository/bean"
	"go.uber.org/zap"
)

type AppStoreValidator interface {
	Validate(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, environment *bean.EnvironmentBean) error
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

func (impl *AppStoreValidatorImpl) Validate(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, environment *bean.EnvironmentBean) error {
	return nil
}
