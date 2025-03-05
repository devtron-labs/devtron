/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	client "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/timelineStatus"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/common"
	repository5 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"time"

	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/chartGroup/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

// FullModeDeploymentService TODO refactoring: Use extended binding over EAMode.EAModeDeploymentService
// Currently creating duplicate methods in EAMode.EAModeDeploymentService
type FullModeDeploymentService interface {
	// ArgoCd Services ---------------------------------

	// InstallApp will register git repo in Argo, create and sync the Argo App and finally update deployment status
	InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	// DeleteInstalledApp will delete entry from appStatus.AppStatusDto table and from repository.ChartGroupDeployment table (if exists)
	DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error
	// RollbackRelease will rollback to a previous deployment for the given installedAppVersionHistoryId; returns - valuesYamlStr, success, error
	RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32) (*appStoreBean.InstallAppVersionDTO, bool, error)
	// GetDeploymentHistory will return gRPC.HelmAppDeploymentHistory for the given installedAppDto.InstalledAppId
	GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*gRPC.HelmAppDeploymentHistory, error)
	// GetDeploymentHistoryInfo will return openapi.HelmAppDeploymentManifestDetail for the given appStoreBean.InstallAppVersionDTO
	GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error)

	InstalledAppArgoCdService
	DeploymentStatusService
	InstalledAppGitOpsService
}

type FullModeDeploymentServiceImpl struct {
	Logger                               *zap.SugaredLogger
	argoK8sClient                        argocdServer.ArgoK8sClient
	aCDAuthConfig                        *util2.ACDAuthConfig
	chartGroupDeploymentRepository       repository2.ChartGroupDeploymentRepository
	installedAppRepository               repository.InstalledAppRepository
	installedAppRepositoryHistory        repository.InstalledAppVersionHistoryRepository
	appStoreDeploymentCommonService      appStoreDeploymentCommon.AppStoreDeploymentCommonService
	helmAppService                       client.HelmAppService
	appStatusService                     appStatus.AppStatusService
	pipelineStatusTimelineService        status.PipelineStatusTimelineService
	pipelineStatusTimelineRepository     pipelineConfig.PipelineStatusTimelineRepository
	userService                          user.UserService
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	argoClientWrapperService             argocdServer.ArgoClientWrapperService
	acdConfig                            *argocdServer.ACDConfig
	gitOperationService                  git.GitOperationService
	gitOpsConfigReadService              config.GitOpsConfigReadService
	gitOpsValidationService              validation.GitOpsValidationService
	environmentRepository                repository5.EnvironmentRepository
	deploymentConfigService              common.DeploymentConfigService
	chartTemplateService                 util.ChartTemplateService
}

func NewFullModeDeploymentServiceImpl(
	logger *zap.SugaredLogger,
	argoK8sClient argocdServer.ArgoK8sClient,
	aCDAuthConfig *util2.ACDAuthConfig,
	chartGroupDeploymentRepository repository2.ChartGroupDeploymentRepository,
	installedAppRepository repository.InstalledAppRepository,
	installedAppRepositoryHistory repository.InstalledAppVersionHistoryRepository,
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	helmAppService client.HelmAppService,
	appStatusService appStatus.AppStatusService,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	userService user.UserService,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	argoClientWrapperService argocdServer.ArgoClientWrapperService,
	acdConfig *argocdServer.ACDConfig,
	gitOperationService git.GitOperationService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	gitOpsValidationService validation.GitOpsValidationService,
	environmentRepository repository5.EnvironmentRepository,
	deploymentConfigService common.DeploymentConfigService,
	chartTemplateService util.ChartTemplateService) *FullModeDeploymentServiceImpl {
	return &FullModeDeploymentServiceImpl{
		Logger:                               logger,
		argoK8sClient:                        argoK8sClient,
		aCDAuthConfig:                        aCDAuthConfig,
		chartGroupDeploymentRepository:       chartGroupDeploymentRepository,
		installedAppRepository:               installedAppRepository,
		installedAppRepositoryHistory:        installedAppRepositoryHistory,
		appStoreDeploymentCommonService:      appStoreDeploymentCommonService,
		helmAppService:                       helmAppService,
		appStatusService:                     appStatusService,
		pipelineStatusTimelineService:        pipelineStatusTimelineService,
		pipelineStatusTimelineRepository:     pipelineStatusTimelineRepository,
		userService:                          userService,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		argoClientWrapperService:             argoClientWrapperService,
		acdConfig:                            acdConfig,
		gitOperationService:                  gitOperationService,
		gitOpsConfigReadService:              gitOpsConfigReadService,
		gitOpsValidationService:              gitOpsValidationService,
		environmentRepository:                environmentRepository,
		deploymentConfigService:              deploymentConfigService,
		chartTemplateService:                 chartTemplateService,
	}
}

func (impl *FullModeDeploymentServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	//STEP 4: registerInArgo
	err := impl.argoClientWrapperService.RegisterGitOpsRepoInArgoWithRetry(ctx, chartGitAttr.RepoUrl, chartGitAttr.TargetRevision, installAppVersionRequest.UserId)
	if err != nil {
		impl.Logger.Errorw("error in argo registry", "err", err)
		return nil, err
	}
	//STEP 5: createInArgo
	err = impl.createInArgo(ctx, chartGitAttr, *installAppVersionRequest.Environment, installAppVersionRequest.ACDAppName)
	if err != nil {
		impl.Logger.Errorw("error in create in argo", "err", err)
		return nil, err
	}
	//STEP 6: Force Sync ACD - works like trigger deployment
	//impl.SyncACD(installAppVersionRequest.ACDAppName, ctx)

	//STEP 7: normal refresh ACD - update for step 6 to avoid delay
	syncTime := time.Now()
	err = impl.argoClientWrapperService.SyncArgoCDApplicationIfNeededAndRefresh(ctx, installAppVersionRequest.ACDAppName, chartGitAttr.TargetRevision)
	if err != nil {
		impl.Logger.Errorw("error in getting the argo application with normal refresh", "err", err)
		return nil, err
	}
	if impl.acdConfig.IsManualSyncEnabled() {
		timeline := &pipelineConfig.PipelineStatusTimeline{
			InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
			Status:                       timelineStatus.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED,
			StatusDetail:                 timelineStatus.TIMELINE_DESCRIPTION_ARGOCD_SYNC_COMPLETED,
			StatusTime:                   syncTime,
			AuditLog: sql.AuditLog{
				CreatedBy: installAppVersionRequest.UserId,
				CreatedOn: time.Now(),
				UpdatedBy: installAppVersionRequest.UserId,
				UpdatedOn: time.Now(),
			},
		}
		err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, tx)
		if err != nil {
			impl.Logger.Errorw("error in creating timeline for argocd sync", "err", err, "timeline", timeline)
		}
	}

	return installAppVersionRequest, nil
}

func (impl *FullModeDeploymentServiceImpl) DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error {

	err := impl.appStatusService.DeleteWithAppIdEnvId(dbTransaction, installedApps.AppId, installedApps.EnvironmentId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in deleting app_status", "appId", installedApps.AppId, "envId", installedApps.EnvironmentId, "err", err)
		return err
	} else if err == pg.ErrNoRows {
		impl.Logger.Warnw("App status not present, skipping app status delete ")
	}

	deployment, err := impl.chartGroupDeploymentRepository.FindByInstalledAppId(installedApps.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in fetching chartGroupMapping", "id", installedApps.Id, "err", err)
		return err
	} else if err == pg.ErrNoRows {
		impl.Logger.Infow("not a chart group deployment skipping chartGroupMapping delete", "id", installedApps.Id)
	} else {
		deployment.Deleted = true
		deployment.UpdatedOn = time.Now()
		deployment.UpdatedBy = installAppVersionRequest.UserId
		_, err := impl.chartGroupDeploymentRepository.Update(deployment, dbTransaction)
		if err != nil {
			impl.Logger.Errorw("error in mapping delete", "err", err)
			return err
		}
	}
	return nil
}

func (impl *FullModeDeploymentServiceImpl) RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, installedAppVersionHistoryId int32) (*appStoreBean.InstallAppVersionDTO, bool, error) {
	return nil, false, errors.New("this is not implemented")
}

func (impl *FullModeDeploymentServiceImpl) GetDeploymentHistory(ctx context.Context, installedAppDto *appStoreBean.InstallAppVersionDTO) (*gRPC.HelmAppDeploymentHistory, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "FullModeDeploymentServiceImp.GetDeploymentHistory")
	defer span.End()
	return impl.appStoreDeploymentCommonService.GetDeploymentHistoryFromDB(newCtx, installedAppDto)
}

func (impl *FullModeDeploymentServiceImpl) migrateDeploymentHistoryMessage(ctx context.Context, updateHistory *repository.InstalledAppVersionHistory) (helmInstallStatusMsg string) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "FullModeDeploymentServiceImp.migrateDeploymentHistoryMessage")
	defer span.End()
	helmInstallStatusMsg = updateHistory.Message
	helmInstallStatus := &appStoreBean.HelmReleaseStatusConfig{}
	jsonErr := json.Unmarshal([]byte(updateHistory.HelmReleaseStatusConfig), helmInstallStatus)
	if jsonErr != nil {
		impl.Logger.Errorw("error while unmarshal helm release status config", "helmReleaseStatusConfig", updateHistory.HelmReleaseStatusConfig, "error", jsonErr)
		return helmInstallStatusMsg
	} else if helmInstallStatus.ErrorInInstallation {
		helmInstallStatusMsg = fmt.Sprintf("Deployment failed: %v", helmInstallStatus.Message)
		dbErr := impl.installedAppRepositoryHistory.UpdateDeploymentHistoryMessage(updateHistory.Id, helmInstallStatusMsg)
		if dbErr != nil {
			impl.Logger.Errorw("error while updating deployment history helmInstallStatusMsg", "error", dbErr)
		}
		return helmInstallStatusMsg
	}
	return helmInstallStatusMsg
}

// GetDeploymentHistoryInfo TODO refactoring: use InstalledAppVersionHistoryId from appStoreBean.InstallAppVersionDTO instead of version int32
func (impl *FullModeDeploymentServiceImpl) GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error) {
	return impl.appStoreDeploymentCommonService.GetDeploymentHistoryInfoFromDB(ctx, installedApp, version)
}

func (impl *FullModeDeploymentServiceImpl) getSourcesFromManifest(chartYaml string) ([]string, error) {
	var b map[string]interface{}
	var sources []string
	err := json.Unmarshal([]byte(chartYaml), &b)
	if err != nil {
		impl.Logger.Errorw("error while unmarshal chart yaml", "error", err)
		return sources, err
	}
	if b != nil && b["sources"] != nil {
		slice := b["sources"].([]interface{})
		for _, item := range slice {
			sources = append(sources, item.(string))
		}
	}
	return sources, nil
}
