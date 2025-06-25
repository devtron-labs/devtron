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
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/pkg/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/common"
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
}

func NewFullModeFluxDeploymentServiceImpl(logger *zap.SugaredLogger,
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService) *FullModeFluxDeploymentServiceImpl {
	return &FullModeFluxDeploymentServiceImpl{
		logger:                          logger,
		appStoreDeploymentCommonService: appStoreDeploymentCommonService,
	}
}

func (impl *FullModeFluxDeploymentServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute,
	ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {

}

func (impl *FullModeFluxDeploymentServiceImpl) DeleteInstalledApp(ctx context.Context, appName, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO,
	installedApps *repository.InstalledApps, tx *pg.Tx) error {

}

func (impl *FullModeFluxDeploymentServiceImpl) RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32) (*appStoreBean.InstallAppVersionDTO, bool, error) {

}

func (impl *FullModeFluxDeploymentServiceImpl) GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*gRPC.HelmAppDeploymentHistory, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "FullModeFluxDeploymentServiceImpl.GetDeploymentHistory")
	defer span.End()
	return impl.appStoreDeploymentCommonService.GetDeploymentHistoryFromDB(newCtx, installedApp)
}

func (impl *FullModeFluxDeploymentServiceImpl) GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error) {
	return impl.appStoreDeploymentCommonService.GetDeploymentHistoryInfoFromDB(ctx, installedApp, version)
}

func (impl *HandlerServiceImpl) deployFluxCdApp(ctx context.Context, overrideRequest *bean3.ValuesOverrideRequest,
	valuesOverrideResponse *app.ValuesOverrideResponse) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "HandlerServiceImpl.deployFluxCdApp")
	defer span.End()
	clusterConfig, err := impl.clusterService.GetClusterConfigByClusterId(overrideRequest.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster", "clusterId", overrideRequest.ClusterId, "error", err)
		return err
	}
	gitOpsSecret, err := impl.upsertGitRepoSecret(newCtx,
		valuesOverrideResponse.DeploymentConfig.ReleaseConfiguration.FluxCDSpec.GitOpsSecretName,
		valuesOverrideResponse.ManifestPushTemplate.RepoUrl,
		overrideRequest.Namespace,
		clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in creating git repo secret", "clusterId", overrideRequest.ClusterId, "err", err)
		return err
	}
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config", "clusterId", overrideRequest.ClusterId, "err", err)
		return err
	}
	apiClient, err := getClient(restConfig)
	if err != nil {
		impl.logger.Errorw("error in creating k8s client", "clusterId", overrideRequest.ClusterId, "err", err)
		return err
	}
	//create/update gitOps secret
	if valuesOverrideResponse.Pipeline == nil || !valuesOverrideResponse.Pipeline.DeploymentAppCreated {
		err := impl.createFluxCdApp(newCtx, valuesOverrideResponse, gitOpsSecret.GetName(), apiClient, overrideRequest.UserId)
		if err != nil {
			impl.logger.Errorw("error in creating flux-cd application", "err", err)
			return err
		}
	} else {
		err := impl.updateFluxCdApp(newCtx, valuesOverrideResponse, gitOpsSecret.GetName(), apiClient)
		if err != nil {
			impl.logger.Errorw("error in updating flux-cd application", "err", err)
			return err
		}
	}
	return nil
}
