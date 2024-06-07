/*
 * Copyright (c) 2024. Devtron Inc.
 */

package service

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"go.uber.org/zap"
)

type DeletePostProcessor interface {
	Process(app *app.App, installAppVersionRequest *appStoreBean.InstallAppVersionDTO)
}

type DeletePostProcessorImpl struct {
	logger *zap.SugaredLogger
}

func NewDeletePostProcessorImpl(
	logger *zap.SugaredLogger,
) *DeletePostProcessorImpl {
	return &DeletePostProcessorImpl{
		logger: logger,
	}
}

func (impl *DeletePostProcessorImpl) Process(app *app.App, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) {
}
