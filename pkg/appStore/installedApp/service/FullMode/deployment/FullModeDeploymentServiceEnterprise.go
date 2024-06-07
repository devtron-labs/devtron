/*
 * Copyright (c) 2024. Devtron Inc.
 */

package deployment

import (
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/common"
	"go.uber.org/zap"
)

type FullModeDeploymentServiceEnterprise interface {
	FullModeDeploymentService
	InstalledAppVirtualDeploymentService
}

type FullModeDeploymentServiceEnterpriseImpl struct {
	FullModeDeploymentService
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonServiceEnterprise
	Logger                          *zap.SugaredLogger
	installedAppRepository          repository.InstalledAppRepository
	installedAppRepositoryHistory   repository.InstalledAppVersionHistoryRepository
}

func NewFullModeDeploymentServiceEnterpriseImpl(fullModeDeploymentService FullModeDeploymentService,
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonServiceEnterprise,
	Logger *zap.SugaredLogger, installedAppRepository repository.InstalledAppRepository,
	installedAppRepositoryHistory repository.InstalledAppVersionHistoryRepository) *FullModeDeploymentServiceEnterpriseImpl {
	return &FullModeDeploymentServiceEnterpriseImpl{
		FullModeDeploymentService:       fullModeDeploymentService,
		appStoreDeploymentCommonService: appStoreDeploymentCommonService,
		Logger:                          Logger,
		installedAppRepository:          installedAppRepository,
		installedAppRepositoryHistory:   installedAppRepositoryHistory,
	}
}
