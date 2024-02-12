/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package service

import (
	"context"
	"errors"
	"fmt"
	bean3 "github.com/devtron-labs/devtron/api/helm-app/bean"
	bean4 "github.com/devtron-labs/devtron/api/helm-app/gRPC"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/adapter"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	repository3 "github.com/devtron-labs/devtron/pkg/appStore/chartGroup/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode/deployment"
	bean2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type AppStoreDeploymentService interface {
	InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error)
	UpdateInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error)
	DeleteInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error)
	LinkHelmApplicationToChartStore(ctx context.Context, request *openapi.UpdateReleaseWithChartLinkingRequest, appIdentifier *service.AppIdentifier, userId int32) (*openapi.UpdateReleaseResponse, bool, error)
	UpdateProjectHelmApp(updateAppRequest *appStoreBean.UpdateProjectHelmAppDTO) error
	RollbackApplication(ctx context.Context, request *openapi2.RollbackReleaseRequest, installedApp *appStoreBean.InstallAppVersionDTO, userId int32) (bool, error)
	GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*bean3.DeploymentHistoryAndInstalledAppInfo, error)
	GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, installedAppVersionHistoryId int) (*openapi.HelmAppDeploymentManifestDetail, error)
	InstallAppByHelm(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error)
	UpdatePreviousDeploymentStatusForAppStore(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, triggeredAt time.Time, err error) error
	MarkGitOpsInstalledAppsDeletedIfArgoAppIsDeleted(installedAppId, envId int) error
}

type AppStoreDeploymentServiceImpl struct {
	logger                               *zap.SugaredLogger
	installedAppRepository               repository.InstalledAppRepository
	installedAppService                  EAMode.InstalledAppDBService
	appStoreDeploymentDBService          AppStoreDeploymentDBService
	chartGroupDeploymentRepository       repository3.ChartGroupDeploymentRepository
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	appRepository                        app.AppRepository
	eaModeDeploymentService              EAMode.EAModeDeploymentService
	fullModeDeploymentService            deployment.FullModeDeploymentService
	environmentService                   cluster.EnvironmentService
	helmAppService                       service.HelmAppService
	installedAppRepositoryHistory        repository.InstalledAppVersionHistoryRepository
	deploymentTypeConfig                 *util2.DeploymentServiceTypeConfig
	aCDConfig                            *argocdServer.ACDConfig
	gitOpsConfigReadService              config.GitOpsConfigReadService
}

func NewAppStoreDeploymentServiceImpl(logger *zap.SugaredLogger,
	installedAppRepository repository.InstalledAppRepository,
	installedAppService EAMode.InstalledAppDBService,
	appStoreDeploymentDBService AppStoreDeploymentDBService,
	chartGroupDeploymentRepository repository3.ChartGroupDeploymentRepository,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	appRepository app.AppRepository,
	eaModeDeploymentService EAMode.EAModeDeploymentService,
	fullModeDeploymentService deployment.FullModeDeploymentService,
	environmentService cluster.EnvironmentService,
	helmAppService service.HelmAppService,
	installedAppRepositoryHistory repository.InstalledAppVersionHistoryRepository,
	envVariables *util2.EnvironmentVariables,
	aCDConfig *argocdServer.ACDConfig,
	gitOpsConfigReadService config.GitOpsConfigReadService) *AppStoreDeploymentServiceImpl {
	return &AppStoreDeploymentServiceImpl{
		logger:                               logger,
		installedAppRepository:               installedAppRepository,
		installedAppService:                  installedAppService,
		appStoreDeploymentDBService:          appStoreDeploymentDBService,
		chartGroupDeploymentRepository:       chartGroupDeploymentRepository,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		appRepository:                        appRepository,
		eaModeDeploymentService:              eaModeDeploymentService,
		fullModeDeploymentService:            fullModeDeploymentService,
		environmentService:                   environmentService,
		helmAppService:                       helmAppService,
		installedAppRepositoryHistory:        installedAppRepositoryHistory,
		deploymentTypeConfig:                 envVariables.DeploymentServiceTypeConfig,
		aCDConfig:                            aCDConfig,
		gitOpsConfigReadService:              gitOpsConfigReadService,
	}
}

func (impl *AppStoreDeploymentServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error) {

	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	//step 1 db operation initiated
	installAppVersionRequest, err = impl.appStoreDeploymentDBService.AppStoreDeployOperationDB(installAppVersionRequest, tx)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return nil, err
	}
	installedAppDeploymentAction := adapter.NewInstalledAppDeploymentAction(installAppVersionRequest.DeploymentAppType)

	if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) || util.IsManifestDownload(installAppVersionRequest.DeploymentAppType) {
		_ = impl.fullModeDeploymentService.SaveTimelineForHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED, "Deployment initiated successfully.", time.Now(), tx)
	}

	if util.IsManifestDownload(installAppVersionRequest.DeploymentAppType) {
		_ = impl.fullModeDeploymentService.SaveTimelineForHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_DESCRIPTION_MANIFEST_GENERATED, "Manifest generated successfully.", time.Now(), tx)
	}

	var gitOpsResponse *bean2.AppStoreGitOpsResponse
	if installedAppDeploymentAction.PerformGitOps {
		manifest, err := impl.fullModeDeploymentService.GenerateManifest(installAppVersionRequest)
		if err != nil {
			impl.logger.Errorw("error in performing manifest and git operations", "err", err)
			return nil, err
		}
		gitOpsResponse, err = impl.fullModeDeploymentService.GitOpsOperations(manifest, installAppVersionRequest)
		if err != nil {
			impl.logger.Errorw("error in doing gitops operation", "err", err)
			if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
				_ = impl.fullModeDeploymentService.SaveTimelineForHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED, fmt.Sprintf("Git commit failed - %v", err), time.Now(), tx)
			}
			return nil, err
		}
		if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
			_ = impl.fullModeDeploymentService.SaveTimelineForHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT, "Git commit done successfully.", time.Now(), tx)
			if !impl.aCDConfig.ArgoCDAutoSyncEnabled {
				_ = impl.fullModeDeploymentService.SaveTimelineForHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_INITIATED, "argocd sync initiated.", time.Now(), tx)
			}
		}
		installAppVersionRequest.GitHash = gitOpsResponse.GitHash
		if len(installAppVersionRequest.GitHash) > 0 {
			err = impl.installedAppRepositoryHistory.UpdateGitHash(installAppVersionRequest.InstalledAppVersionHistoryId, gitOpsResponse.GitHash, tx)
			if err != nil {
				impl.logger.Errorw("error in updating git hash ", "err", err)
				return nil, err
			}
		}
	}

	if util2.IsBaseStack() || util2.IsHelmApp(installAppVersionRequest.AppOfferingMode) || util.IsHelmApp(installAppVersionRequest.DeploymentAppType) {
		installAppVersionRequest, err = impl.eaModeDeploymentService.InstallApp(installAppVersionRequest, nil, ctx, tx)
	} else if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
		if gitOpsResponse == nil && gitOpsResponse.ChartGitAttribute != nil {
			return nil, errors.New("service err, Error in git operations")
		}
		installAppVersionRequest, err = impl.fullModeDeploymentService.InstallApp(installAppVersionRequest, gitOpsResponse.ChartGitAttribute, ctx, tx)
	}
	if err != nil {
		return nil, err
	}
	err = tx.Commit()

	err = impl.appStoreDeploymentDBService.InstallAppPostDbOperation(installAppVersionRequest)
	if err != nil {
		return nil, err
	}

	return installAppVersionRequest, nil
}

func (impl *AppStoreDeploymentServiceImpl) DeleteInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error) {
	installAppVersionRequest.InstalledAppDeleteResponse = &appStoreBean.InstalledAppDeleteResponseDTO{
		DeleteInitiated:  false,
		ClusterReachable: true,
	}
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	environment, err := impl.environmentService.GetExtendedEnvBeanById(installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}
	if len(environment.ErrorInConnecting) > 0 {
		installAppVersionRequest.InstalledAppDeleteResponse.ClusterReachable = false
		installAppVersionRequest.InstalledAppDeleteResponse.ClusterName = environment.ClusterName
	}

	app, err := impl.appRepository.FindById(installAppVersionRequest.AppId)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, fmt.Errorf("App not found in database")
		}
		return nil, err
	}
	model, err := impl.installedAppRepository.GetInstalledApp(installAppVersionRequest.InstalledAppId)
	if err != nil {
		impl.logger.Errorw("error in fetching installed app", "id", installAppVersionRequest.InstalledAppId, "err", err)
		return nil, err
	}

	if installAppVersionRequest.AcdPartialDelete == true {
		if !util2.IsBaseStack() && !util2.IsHelmApp(app.AppOfferingMode) && !util.IsHelmApp(model.DeploymentAppType) {
			if !installAppVersionRequest.InstalledAppDeleteResponse.ClusterReachable {
				impl.logger.Errorw("cluster connection error", "err", environment.ErrorInConnecting)
				if !installAppVersionRequest.NonCascadeDelete {
					return installAppVersionRequest, nil
				}
			}
			err = impl.fullModeDeploymentService.DeleteACDAppObject(ctx, app.AppName, environment.Environment, installAppVersionRequest)
		}
		if err != nil {
			impl.logger.Errorw("error on delete installed app", "err", err)
			return nil, err
		}
		model.DeploymentAppDeleteRequest = true
		model.UpdatedBy = installAppVersionRequest.UserId
		model.UpdatedOn = time.Now()
		_, err = impl.installedAppRepository.UpdateInstalledApp(model, tx)
		if err != nil {
			impl.logger.Errorw("error while creating install app", "error", err)
			return nil, err
		}
	} else {
		//soft delete app
		app.Active = false
		app.UpdatedBy = installAppVersionRequest.UserId
		app.UpdatedOn = time.Now()
		err = impl.appRepository.UpdateWithTxn(app, tx)
		if err != nil {
			impl.logger.Errorw("error in update entity ", "entity", app)
			return nil, err
		}

		// soft delete install app
		model.Active = false
		model.UpdatedBy = installAppVersionRequest.UserId
		model.UpdatedOn = time.Now()
		_, err = impl.installedAppRepository.UpdateInstalledApp(model, tx)
		if err != nil {
			impl.logger.Errorw("error while creating install app", "error", err)
			return nil, err
		}
		models, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppId(installAppVersionRequest.InstalledAppId)
		if err != nil {
			impl.logger.Errorw("error while fetching install app versions", "error", err)
			return nil, err
		}

		// soft delete install app versions
		for _, item := range models {
			item.Active = false
			item.UpdatedBy = installAppVersionRequest.UserId
			item.UpdatedOn = time.Now()
			_, err = impl.installedAppRepository.UpdateInstalledAppVersion(item, tx)
			if err != nil {
				impl.logger.Errorw("error while fetching from db", "error", err)
				return nil, err
			}
		}

		// soft delete chart-group deployment
		chartGroupDeployment, err := impl.chartGroupDeploymentRepository.FindByInstalledAppId(model.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error while fetching chart group deployment", "error", err)
			return nil, err
		}
		if chartGroupDeployment.Id != 0 {
			chartGroupDeployment.Deleted = true
			_, err = impl.chartGroupDeploymentRepository.Update(chartGroupDeployment, tx)
			if err != nil {
				impl.logger.Errorw("error while updating chart group deployment", "error", err)
				return nil, err
			}
		}

		if util2.IsBaseStack() || util2.IsHelmApp(app.AppOfferingMode) || util.IsHelmApp(model.DeploymentAppType) {
			// there might be a case if helm release gets uninstalled from helm cli.
			//in this case on deleting the app from API, it should not give error as it should get deleted from db, otherwise due to delete error, db does not get clean
			// so in helm, we need to check first if the release exists or not, if exists then only delete
			err = impl.eaModeDeploymentService.DeleteInstalledApp(ctx, app.AppName, environment.Environment, installAppVersionRequest, model, tx)
		} else {
			err = impl.fullModeDeploymentService.DeleteInstalledApp(ctx, app.AppName, environment.Environment, installAppVersionRequest, model, tx)
		}
		if err != nil {
			impl.logger.Errorw("error on delete installed app", "err", err)
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in commit db transaction on delete", "err", err)
		return nil, err
	}
	installAppVersionRequest.InstalledAppDeleteResponse.DeleteInitiated = true
	return installAppVersionRequest, nil
}

func (impl *AppStoreDeploymentServiceImpl) LinkHelmApplicationToChartStore(ctx context.Context, request *openapi.UpdateReleaseWithChartLinkingRequest,
	appIdentifier *service.AppIdentifier, userId int32) (*openapi.UpdateReleaseResponse, bool, error) {

	impl.logger.Infow("Linking helm application to chart store", "appId", request.GetAppId())

	// check if chart repo is active starts
	isChartRepoActive, err := impl.appStoreDeploymentDBService.IsChartProviderActive(int(request.GetAppStoreApplicationVersionId()))
	if err != nil {
		impl.logger.Errorw("Error in checking if chart repo is active or not", "err", err)
		return nil, isChartRepoActive, err
	}
	if !isChartRepoActive {
		return nil, isChartRepoActive, nil
	}
	// check if chart repo is active ends

	// STEP-1 check if the app is installed or not
	isInstalled, err := impl.helmAppService.IsReleaseInstalled(ctx, appIdentifier)
	if err != nil {
		impl.logger.Errorw("error while checking if the release is installed", "error", err)
		return nil, isChartRepoActive, err
	}
	if !isInstalled {
		return nil, isChartRepoActive, errors.New("release is not installed. so can not be updated")
	}
	// STEP-1 ends

	// Initialise bean
	installAppVersionRequestDto := &appStoreBean.InstallAppVersionDTO{
		AppName:            appIdentifier.ReleaseName,
		UserId:             userId,
		AppOfferingMode:    util2.SERVER_MODE_HYPERION,
		ClusterId:          appIdentifier.ClusterId,
		Namespace:          appIdentifier.Namespace,
		AppStoreVersion:    int(request.GetAppStoreApplicationVersionId()),
		ValuesOverrideYaml: request.GetValuesYaml(),
		ReferenceValueId:   int(request.GetReferenceValueId()),
		ReferenceValueKind: request.GetReferenceValueKind(),
		DeploymentAppType:  util.PIPELINE_DEPLOYMENT_TYPE_HELM,
	}

	// STEP-2 InstallApp with only DB operations
	// STEP-3 update APP with chart info
	res, err := impl.linkHelmApplicationToChartStore(installAppVersionRequestDto, ctx)
	if err != nil {
		impl.logger.Errorw("error while linking helm app with chart store", "error", err)
		return nil, isChartRepoActive, err
	}
	// STEP-2 and STEP-3 ends

	return res, isChartRepoActive, nil
}

func (impl *AppStoreDeploymentServiceImpl) UpdateProjectHelmApp(updateAppRequest *appStoreBean.UpdateProjectHelmAppDTO) error {

	appIdSplitted := strings.Split(updateAppRequest.AppId, "|")

	appName := updateAppRequest.AppName

	if len(appIdSplitted) > 1 {
		// app id is zero for CLI apps
		appIdentifier, _ := impl.helmAppService.DecodeAppId(updateAppRequest.AppId)
		appName = appIdentifier.ReleaseName
	}
	impl.logger.Infow("update helm project request", updateAppRequest)
	err := impl.appStoreDeploymentDBService.UpdateProjectForHelmApp(appName, updateAppRequest.TeamId, updateAppRequest.UserId)
	if err != nil {
		impl.logger.Errorw("error in linking project to helm app", "appName", appName, "err", err)
		return err
	}
	return nil
}

func (impl *AppStoreDeploymentServiceImpl) RollbackApplication(ctx context.Context, request *openapi2.RollbackReleaseRequest,
	installedApp *appStoreBean.InstallAppVersionDTO, userId int32) (bool, error) {
	triggeredAt := time.Now()
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return false, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// Rollback starts
	var success bool
	if util2.IsHelmApp(installedApp.AppOfferingMode) {
		installedApp, success, err = impl.eaModeDeploymentService.RollbackRelease(ctx, installedApp, request.GetVersion(), tx)
		if err != nil {
			impl.logger.Errorw("error while rollback helm release", "error", err)
			return false, err
		}
	} else {
		installedApp, success, err = impl.fullModeDeploymentService.RollbackRelease(ctx, installedApp, request.GetVersion(), tx)
		if err != nil {
			impl.logger.Errorw("error while rollback helm release", "error", err)
			return false, err
		}
	}
	if !success {
		return false, fmt.Errorf("rollback request failed")
	}
	//DB operation
	if installedApp.InstalledAppId > 0 && installedApp.InstalledAppVersionId > 0 {
		installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersionAny(installedApp.InstalledAppVersionId)
		if err != nil {
			impl.logger.Errorw("error while fetching chart installed version", "error", err)
			return false, err
		}
		installedApp.Id = installedAppVersion.Id
		installedAppVersion.Active = true
		installedAppVersion.ValuesYaml = installedApp.ValuesOverrideYaml
		installedAppVersion.UpdatedOn = time.Now()
		installedAppVersion.UpdatedBy = userId
		_, err = impl.installedAppRepository.UpdateInstalledAppVersion(installedAppVersion, tx)
		if err != nil {
			impl.logger.Errorw("error while updating db", "error", err)
			return false, err
		}
	}
	//STEP 8: finish with return response
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while committing transaction to db", "error", err)
		return false, err
	}

	err1 := impl.UpdatePreviousDeploymentStatusForAppStore(installedApp, triggeredAt, err)
	if err1 != nil {
		impl.logger.Errorw("error while update previous installed app version history", "err", err, "installAppVersionRequest", installedApp)
		//if installed app is updated and error is in updating previous deployment status, then don't block user, just show error.
	}

	return success, nil
}

func (impl *AppStoreDeploymentServiceImpl) GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*bean3.DeploymentHistoryAndInstalledAppInfo, error) {
	result := &bean3.DeploymentHistoryAndInstalledAppInfo{}
	var err error
	if util2.IsHelmApp(installedApp.AppOfferingMode) {
		deploymentHistory, err := impl.eaModeDeploymentService.GetDeploymentHistory(ctx, installedApp)
		if err != nil {
			impl.logger.Errorw("error while getting deployment history", "error", err)
			return nil, err
		}
		result.DeploymentHistory = deploymentHistory.GetDeploymentHistory()
	} else {
		deploymentHistory, err := impl.fullModeDeploymentService.GetDeploymentHistory(ctx, installedApp)
		if err != nil {
			impl.logger.Errorw("error while getting deployment history", "error", err)
			return nil, err
		}
		result.DeploymentHistory = deploymentHistory.GetDeploymentHistory()
	}

	if installedApp.InstalledAppId > 0 {
		result.InstalledAppInfo = &bean3.InstalledAppInfo{
			AppId:                 installedApp.AppId,
			EnvironmentName:       installedApp.EnvironmentName,
			AppOfferingMode:       installedApp.AppOfferingMode,
			InstalledAppId:        installedApp.InstalledAppId,
			InstalledAppVersionId: installedApp.InstalledAppVersionId,
			AppStoreChartId:       installedApp.InstallAppVersionChartDTO.AppStoreChartId,
			ClusterId:             installedApp.ClusterId,
			EnvironmentId:         installedApp.EnvironmentId,
			DeploymentType:        installedApp.DeploymentAppType,
		}
	}

	return result, err
}

func (impl *AppStoreDeploymentServiceImpl) GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int) (*openapi.HelmAppDeploymentManifestDetail, error) {
	//var result interface{}
	result := &openapi.HelmAppDeploymentManifestDetail{}
	var err error
	if util2.IsHelmApp(installedApp.AppOfferingMode) {
		_, span := otel.Tracer("orchestrator").Start(ctx, "eaModeDeploymentService.GetDeploymentHistoryInfo")
		result, err = impl.eaModeDeploymentService.GetDeploymentHistoryInfo(ctx, installedApp, int32(version))
		span.End()
		if err != nil {
			impl.logger.Errorw("error while getting deployment history info", "error", err)
			return nil, err
		}
	} else {
		_, span := otel.Tracer("orchestrator").Start(ctx, "fullModeDeploymentService.GetDeploymentHistoryInfo")
		result, err = impl.fullModeDeploymentService.GetDeploymentHistoryInfo(ctx, installedApp, int32(version))
		span.End()
		if err != nil {
			impl.logger.Errorw("error while getting deployment history info", "error", err)
			return nil, err
		}
	}
	return result, err
}

func (impl *AppStoreDeploymentServiceImpl) UpdateInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error) {
	// db operations
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	installedApp, err := impl.installedAppService.GetInstalledAppById(installAppVersionRequest.InstalledAppId)
	if err != nil {
		return nil, err
	}
	installAppVersionRequest.UpdateDeploymentAppType(installedApp.DeploymentAppType)

	installedAppDeploymentAction := adapter.NewInstalledAppDeploymentAction(installedApp.DeploymentAppType)

	var installedAppVersion *repository.InstalledAppVersions

	// mark previous versions of chart as inactive if chart or version is updated
	isChartChanged := false   // flag for keeping track if chart is updated by user or not
	isVersionChanged := false // flag for keeping track if version of chart is upgraded

	//if chart is changed, then installedAppVersion id is sent as 0 from front-end
	if installAppVersionRequest.Id == 0 {
		isChartChanged = true
		err = impl.appStoreDeploymentDBService.MarkInstalledAppVersionsInactiveByInstalledAppId(installAppVersionRequest.InstalledAppId, installAppVersionRequest.UserId, tx)
		if err != nil {
			return nil, err
		}
	} else {
		installedAppVersion, err = impl.installedAppRepository.GetInstalledAppVersion(installAppVersionRequest.Id)
		if err != nil {
			impl.logger.Errorw("error in fetching installedAppVersion by installAppVersionRequest id ", "err", err)
			return nil, fmt.Errorf("The values are outdated. Please make your changes to the latest version and try again.")
		}
		// version is upgraded if appStoreApplication version from request payload is not equal to installed app version saved in DB
		if installedAppVersion.AppStoreApplicationVersionId != installAppVersionRequest.AppStoreVersion {
			isVersionChanged = true
			err = impl.appStoreDeploymentDBService.MarkInstalledAppVersionModelInActive(installedAppVersion, installAppVersionRequest.UserId, tx)
		}
	}

	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}
	// create new entry for installed app version if chart or version is changed
	if isChartChanged || isVersionChanged {
		installedAppVersion, err = impl.installedAppService.CreateInstalledAppVersion(installAppVersionRequest, tx)
		if err != nil {
			impl.logger.Errorw("error in creating installed app version", "err", err)
			return nil, err
		}
	} else if installedAppVersion != nil && installedAppVersion.Id != 0 {
		adapter.UpdateInstalledAppVersionModel(installedAppVersion, installAppVersionRequest)
		installedAppVersion, err = impl.installedAppService.UpdateInstalledAppVersion(installedAppVersion, installAppVersionRequest, tx)
		if err != nil {
			impl.logger.Errorw("error in creating installed app version", "err", err)
			return nil, err
		}
	}
	// populate the related model data into repository.InstalledAppVersions
	// related tables: repository.InstalledApps AND appStoreDiscoverRepository.AppStoreApplicationVersion
	installedAppVersion.AppStoreApplicationVersion = *appStoreAppVersion
	installedAppVersion.InstalledApp = *installedApp

	// populate appStoreBean.InstallAppVersionDTO from the DB models
	installAppVersionRequest.Id = installedAppVersion.Id
	installAppVersionRequest.InstalledAppVersionId = installedAppVersion.Id
	adapter.UpdateInstallAppDetails(installAppVersionRequest, installedApp)
	adapter.UpdateAppDetails(installAppVersionRequest, &installedApp.App)
	environment, err := impl.environmentService.GetExtendedEnvBeanById(installedApp.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching environment error", "envId", installedApp.EnvironmentId, "err", err)
		return nil, err
	}
	adapter.UpdateAdditionalEnvDetails(installAppVersionRequest, environment)

	helmInstallConfigDTO := appStoreBean.HelmReleaseStatusConfig{
		InstallAppVersionHistoryId: 0,
		Message:                    "Install initiated",
		IsReleaseInstalled:         false,
		ErrorInInstallation:        false,
	}
	installedAppVersionHistory, err := adapter.NewInstallAppVersionHistoryModel(installAppVersionRequest, pipelineConfig.WorkflowInProgress, helmInstallConfigDTO)
	_, err = impl.installedAppRepositoryHistory.CreateInstalledAppVersionHistory(installedAppVersionHistory, tx)
	if err != nil {
		impl.logger.Errorw("error while creating installed app version history for updating installed app", "error", err)
		return nil, err
	}
	installAppVersionRequest.InstalledAppVersionHistoryId = installedAppVersionHistory.Id
	_ = impl.fullModeDeploymentService.SaveTimelineForHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED, "Deployment initiated successfully.", time.Now(), tx)

	if util.IsManifestDownload(installAppVersionRequest.DeploymentAppType) {
		_ = impl.fullModeDeploymentService.SaveTimelineForHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_DESCRIPTION_MANIFEST_GENERATED, "Manifest generated successfully.", time.Now(), tx)
	}
	// gitOps operation
	monoRepoMigrationRequired := false
	gitOpsResponse := &bean2.AppStoreGitOpsResponse{}

	if installedAppDeploymentAction.PerformGitOps {
		manifest, err := impl.fullModeDeploymentService.GenerateManifest(installAppVersionRequest)
		if err != nil {
			impl.logger.Errorw("error in generating manifest for helm apps", "err", err)
			_ = impl.appStoreDeploymentDBService.UpdateInstalledAppVersionHistoryStatus(installAppVersionRequest.InstalledAppVersionHistoryId, pipelineConfig.WorkflowFailed)
			return nil, err
		}
		// required if gitOps repo name is changed, gitOps repo name will change if env variable which we use as suffix changes
		gitOpsRepoName := installedApp.GitOpsRepoName
		if len(gitOpsRepoName) == 0 {
			if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
				gitOpsRepoName, err = impl.fullModeDeploymentService.GetAcdAppGitOpsRepoName(installAppVersionRequest.AppName, installAppVersionRequest.EnvironmentName)
				if err != nil {
					return nil, err
				}
			} else {
				gitOpsRepoName = impl.gitOpsConfigReadService.GetGitOpsRepoName(installAppVersionRequest.AppName)
			}
		}
		//here will set new git repo name if required to migrate
		newGitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoName(installedApp.App.AppName)
		//checking weather git repo migration needed or not, if existing git repo and new independent git repo is not same than go ahead with migration
		if newGitOpsRepoName != gitOpsRepoName {
			monoRepoMigrationRequired = true
			installAppVersionRequest.GitOpsRepoName = newGitOpsRepoName
		} else {
			installAppVersionRequest.GitOpsRepoName = gitOpsRepoName
		}
		installedApp.GitOpsRepoName = installAppVersionRequest.GitOpsRepoName
		argocdAppName := installedApp.App.AppName + "-" + installedApp.Environment.Name
		installAppVersionRequest.ACDAppName = argocdAppName

		var gitOpsErr error
		gitOpsResponse, gitOpsErr = impl.fullModeDeploymentService.UpdateAppGitOpsOperations(manifest, installAppVersionRequest, &monoRepoMigrationRequired, isChartChanged || isVersionChanged)
		if gitOpsErr != nil {
			impl.logger.Errorw("error in performing GitOps operation", "err", gitOpsErr)
			_ = impl.fullModeDeploymentService.SaveTimelineForHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED, fmt.Sprintf("Git commit failed - %v", gitOpsErr), time.Now(), tx)
			return nil, gitOpsErr
		}

		installAppVersionRequest.GitHash = gitOpsResponse.GitHash
		_ = impl.fullModeDeploymentService.SaveTimelineForHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT, "Git commit done successfully.", time.Now(), tx)
		if !impl.aCDConfig.ArgoCDAutoSyncEnabled {
			_ = impl.fullModeDeploymentService.SaveTimelineForHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_INITIATED, "Argocd sync initiated", time.Now(), tx)
		}
		installedAppVersionHistory.GitHash = gitOpsResponse.GitHash
		_, err = impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistory(installedAppVersionHistory, tx)
		if err != nil {
			impl.logger.Errorw("error on updating history for chart deployment", "error", err, "installedAppVersion", installedAppVersion)
			return nil, err
		}
	}

	if installedAppDeploymentAction.PerformACDDeployment {
		// refresh update repo details on ArgoCD if repo is changed
		err = impl.fullModeDeploymentService.UpdateAndSyncACDApps(installAppVersionRequest, gitOpsResponse.ChartGitAttribute, monoRepoMigrationRequired, ctx, tx)
		if err != nil {
			impl.logger.Errorw("error in acd patch request", "err", err)
			return nil, err
		}
	} else if installedAppDeploymentAction.PerformHelmDeployment {
		err = impl.eaModeDeploymentService.UpgradeDeployment(installAppVersionRequest, gitOpsResponse.ChartGitAttribute, installAppVersionRequest.InstalledAppVersionHistoryId, ctx)
		if err != nil {
			if err != nil {
				impl.logger.Errorw("error in helm update request", "err", err)
				return nil, err
			}
		}
	}
	installedApp.UpdateStatus(appStoreBean.DEPLOY_SUCCESS)
	installedApp.UpdateAuditLog(installAppVersionRequest.UserId)
	installedApp, err = impl.installedAppRepository.UpdateInstalledApp(installedApp, tx)
	if err != nil {
		impl.logger.Errorw("error in updating installed app", "err", err)
		return nil, err
	}
	//STEP 8: finish with return response
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while committing transaction to db", "error", err)
		return nil, err
	}

	if util.IsManifestDownload(installAppVersionRequest.DeploymentAppType) {
		err = impl.appStoreDeploymentDBService.UpdateInstalledAppVersionHistoryStatus(installAppVersionRequest.InstalledAppVersionHistoryId, pipelineConfig.WorkflowSucceeded)
		if err != nil {
			impl.logger.Errorw("error on creating history for chart deployment", "error", err)
			return nil, err
		}
	} else if util.IsHelmApp(installAppVersionRequest.DeploymentAppType) && !impl.deploymentTypeConfig.HelmInstallASyncMode {
		err = impl.appStoreDeploymentDBService.MarkInstalledAppVersionHistorySucceeded(installAppVersionRequest.InstalledAppVersionHistoryId, installAppVersionRequest.DeploymentAppType)
		if err != nil {
			impl.logger.Errorw("error in updating install app version history on sync", "err", err)
			return nil, err
		}
	}

	return installAppVersionRequest, nil
}

func (impl *AppStoreDeploymentServiceImpl) InstallAppByHelm(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error) {
	installAppVersionRequest, err := impl.eaModeDeploymentService.InstallApp(installAppVersionRequest, nil, ctx, nil)
	if err != nil {
		impl.logger.Errorw("error while installing app via helm", "error", err)
		return installAppVersionRequest, err
	}
	if !impl.deploymentTypeConfig.HelmInstallASyncMode {
		err = impl.appStoreDeploymentDBService.MarkInstalledAppVersionHistorySucceeded(installAppVersionRequest.InstalledAppVersionHistoryId, installAppVersionRequest.DeploymentAppType)
		if err != nil {
			impl.logger.Errorw("error in updating installed app version history with sync", "err", err)
			return installAppVersionRequest, err
		}
	}
	return installAppVersionRequest, nil
}

func (impl *AppStoreDeploymentServiceImpl) UpdatePreviousDeploymentStatusForAppStore(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, triggeredAt time.Time, err error) error {
	//creating pipeline status timeline for deployment failed
	if !util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
		return nil
	}
	err1 := impl.fullModeDeploymentService.UpdateInstalledAppAndPipelineStatusForFailedDeploymentStatus(installAppVersionRequest, triggeredAt, err)
	if err1 != nil {
		impl.logger.Errorw("error in updating previous deployment status for appStore", "err", err1, "installAppVersionRequest", installAppVersionRequest)
		return err1
	}
	return nil
}

func (impl *AppStoreDeploymentServiceImpl) MarkGitOpsInstalledAppsDeletedIfArgoAppIsDeleted(installedAppId, envId int) error {
	apiError := &util.ApiError{}
	installedApp, err := impl.installedAppRepository.GetGitOpsInstalledAppsWhereArgoAppDeletedIsTrue(installedAppId, envId)
	if err != nil {
		impl.logger.Errorw("error in fetching partially deleted argoCd apps from installed app repo", "err", err)
		apiError.HttpStatusCode = http.StatusInternalServerError
		apiError.InternalMessage = "error in fetching partially deleted argoCd apps from installed app repo"
		return apiError
	}
	if !util.IsAcdApp(installedApp.DeploymentAppType) || !installedApp.DeploymentAppDeleteRequest {
		return nil
	}
	// Operates for ArgoCd apps only
	acdAppName := fmt.Sprintf("%s-%s", installedApp.App.AppName, installedApp.Environment.Name)
	isFound, err := impl.fullModeDeploymentService.CheckIfArgoAppExists(acdAppName)
	if err != nil {
		impl.logger.Errorw("error in CheckIfArgoAppExists", "err", err)
		apiError.HttpStatusCode = http.StatusInternalServerError
		apiError.InternalMessage = err.Error()
		return apiError
	}

	if isFound {
		apiError.HttpStatusCode = http.StatusInternalServerError
		apiError.InternalMessage = "App Exist in argo, error in fetching resource tree"
		return apiError
	}

	impl.logger.Warnw("app not found in argo, deleting from db ", "err", err)
	//make call to delete it from pipeline DB
	deleteRequest := &appStoreBean.InstallAppVersionDTO{}
	deleteRequest.ForceDelete = false
	deleteRequest.NonCascadeDelete = false
	deleteRequest.AcdPartialDelete = false
	deleteRequest.InstalledAppId = installedApp.Id
	deleteRequest.AppId = installedApp.AppId
	deleteRequest.AppName = installedApp.App.AppName
	deleteRequest.Namespace = installedApp.Environment.Namespace
	deleteRequest.ClusterId = installedApp.Environment.ClusterId
	deleteRequest.EnvironmentId = installedApp.EnvironmentId
	deleteRequest.AppOfferingMode = installedApp.App.AppOfferingMode
	deleteRequest.UserId = 1
	_, err = impl.DeleteInstalledApp(context.Background(), deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting installed app", "err", err)
		apiError.HttpStatusCode = http.StatusNotFound
		apiError.InternalMessage = "error in deleting installed app"
		return apiError
	}
	apiError.HttpStatusCode = http.StatusNotFound
	return apiError
}

func (impl *AppStoreDeploymentServiceImpl) linkHelmApplicationToChartStore(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*openapi.UpdateReleaseResponse, error) {
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// skipAppCreation flag is set for CLI apps because for CLI Helm apps if project is created first before linking to chart store then app is created during project update time.
	// skipAppCreation - This flag will skip app creation if app already exists.

	//step 1 db operation initiated
	appModel, err := impl.appRepository.FindActiveByName(installAppVersionRequest.AppName)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in getting app", "appName", installAppVersionRequest.AppName)
		return nil, err
	}
	if appModel != nil && appModel.Id > 0 {
		impl.logger.Infow(" app already exists", "name", installAppVersionRequest.AppName)
		installAppVersionRequest.AppId = appModel.Id
	}
	installAppVersionRequest, err = impl.appStoreDeploymentDBService.AppStoreDeployOperationDB(installAppVersionRequest, tx)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return nil, err
	}

	// fetch app store application version from DB
	appStoreApplicationVersionId := installAppVersionRequest.AppStoreVersion
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(appStoreApplicationVersionId)
	if err != nil {
		impl.logger.Errorw("Error in fetching app store application version", "err", err, "appStoreApplicationVersionId", appStoreApplicationVersionId)
		return nil, err
	}

	// STEP-2 update APP with chart info
	chartRepoInfo := appStoreAppVersion.AppStore.ChartRepo
	updateReleaseRequest := &bean3.UpdateApplicationWithChartInfoRequestDto{
		InstallReleaseRequest: &bean4.InstallReleaseRequest{
			ValuesYaml:   installAppVersionRequest.ValuesOverrideYaml,
			ChartName:    appStoreAppVersion.Name,
			ChartVersion: appStoreAppVersion.Version,
			ReleaseIdentifier: &bean4.ReleaseIdentifier{
				ReleaseNamespace: installAppVersionRequest.Namespace,
				ReleaseName:      installAppVersionRequest.AppName,
			},
		},
		SourceAppType: bean3.SOURCE_HELM_APP,
	}
	if chartRepoInfo != nil {
		updateReleaseRequest.ChartRepository = &bean4.ChartRepository{
			Name:     chartRepoInfo.Name,
			Url:      chartRepoInfo.Url,
			Username: chartRepoInfo.UserName,
			Password: chartRepoInfo.Password,
		}
	}
	res, err := impl.helmAppService.UpdateApplicationWithChartInfo(ctx, installAppVersionRequest.ClusterId, updateReleaseRequest)
	if err != nil {
		return nil, err
	}
	// STEP-2 ends

	// tx commit here because next operation will be process after this commit.
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	// STEP-3 install app DB post operations
	installAppVersionRequest.UpdateDeploymentAppType(util.PIPELINE_DEPLOYMENT_TYPE_HELM)
	err = impl.appStoreDeploymentDBService.InstallAppPostDbOperation(installAppVersionRequest)
	if err != nil {
		return nil, err
	}
	// STEP-3 ends

	return res, nil
}
