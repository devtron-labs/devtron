/*
 * Copyright (c) 2024. Devtron Inc.
 */

package service

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/repository/bean"
	"go.uber.org/zap"
	"net/http"
)

type AppStoreValidatorEnterpriseImpl struct {
	*AppStoreValidatorImpl
	logger *zap.SugaredLogger
}

func NewAppStoreValidatorEnterpriseImpl(
	logger *zap.SugaredLogger,
) *AppStoreValidatorEnterpriseImpl {
	return &AppStoreValidatorEnterpriseImpl{
		logger:                logger,
		AppStoreValidatorImpl: NewAppAppStoreValidatorImpl(logger),
	}
}

func (impl *AppStoreValidatorEnterpriseImpl) Validate(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, environment *bean2.EnvironmentBean) error {

	if environment.IsVirtualEnvironment && util.IsManifestPush(installAppVersionRequest.DeploymentAppType) {
		impl.logger.Errorw("invalid deployment type for a virtual environment", "deploymentType", installAppVersionRequest.DeploymentAppType)
		err := &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: fmt.Sprintf("Deployment type '%s' is not supported on virtual cluster", installAppVersionRequest.DeploymentAppType),
			UserMessage:     fmt.Sprintf("Deployment type '%s' is not supported on virtual cluster", installAppVersionRequest.DeploymentAppType),
		}
		return err
	}

	return nil
}
