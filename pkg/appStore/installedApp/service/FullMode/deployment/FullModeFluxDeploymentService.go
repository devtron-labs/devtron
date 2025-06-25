/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package deployment

import (
	"context"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/client/fluxcd"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/common"
	"github.com/devtron-labs/devtron/pkg/cluster"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

// FullModeDeploymentService TODO refactoring: Use extended binding over EAMode.EAModeDeploymentService
// Currently creating duplicate methods in EAMode.EAModeDeploymentService
type FullModeFluxDeploymentService interface {
	// FluxCd Services

	// InstallApp will create git repo and helm release crd objects
	InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	// DeleteInstalledApp will delete entry from appStatus.AppStatusDto table, from repository.ChartGroupDeployment table (if exists) and delete crd objects
	DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error
	// RollbackRelease will rollback to a previous deployment for the given installedAppVersionHistoryId; returns - valuesYamlStr, success, error
	RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32) (*appStoreBean.InstallAppVersionDTO, bool, error)
	// GetDeploymentHistory will return gRPC.HelmAppDeploymentHistory for the given installedAppDto.InstalledAppId
	GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*gRPC.HelmAppDeploymentHistory, error)
	// GetDeploymentHistoryInfo will return openapi.HelmAppDeploymentManifestDetail for the given appStoreBean.InstallAppVersionDTO
	GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error)
}

type FullModeFluxDeploymentServiceImpl struct {
	logger                          *zap.SugaredLogger
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService
	fluxCdDeploymentService         fluxcd.DeploymentService
	clusterService                  cluster.ClusterService
}

func NewFullModeFluxDeploymentServiceImpl(logger *zap.SugaredLogger,
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	fluxCdDeploymentService fluxcd.DeploymentService, clusterService cluster.ClusterService) *FullModeFluxDeploymentServiceImpl {
	return &FullModeFluxDeploymentServiceImpl{
		logger:                          logger,
		appStoreDeploymentCommonService: appStoreDeploymentCommonService,
		fluxCdDeploymentService:         fluxCdDeploymentService,
		clusterService:                  clusterService,
	}
}

func (impl *FullModeFluxDeploymentServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute,
	ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
	clusterConfig, err := impl.clusterService.GetClusterConfigByClusterId(installAppVersionRequest.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster", "clusterId", installAppVersionRequest.ClusterId, "error", err)
		return nil, err
	}
	//deploy app
	err = impl.fluxCdDeploymentService.DeployFluxCdApp(ctx, &fluxcd.DeploymentRequest{
		ClusterConfig:    clusterConfig,
		DeploymentConfig: installAppVersionRequest.GetDeploymentConfig(),
		IsAppCreated:     false,
	})
	if err != nil {
		impl.logger.Errorw("error in deploy Flux Cd App", "error", err)
		return nil, err
	}
	return installAppVersionRequest, nil
}

func (impl *FullModeFluxDeploymentServiceImpl) DeleteInstalledApp(ctx context.Context, appName, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO,
	installedApps *repository.InstalledApps, tx *pg.Tx) error {
	return nil
}

func (impl *FullModeFluxDeploymentServiceImpl) RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32) (*appStoreBean.InstallAppVersionDTO, bool, error) {
	return nil, false, nil
}

func (impl *FullModeFluxDeploymentServiceImpl) GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*gRPC.HelmAppDeploymentHistory, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "FullModeFluxDeploymentServiceImpl.GetDeploymentHistory")
	defer span.End()
	return impl.appStoreDeploymentCommonService.GetDeploymentHistoryFromDB(newCtx, installedApp)
}

func (impl *FullModeFluxDeploymentServiceImpl) GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error) {
	return impl.appStoreDeploymentCommonService.GetDeploymentHistoryInfoFromDB(ctx, installedApp, version)
}
