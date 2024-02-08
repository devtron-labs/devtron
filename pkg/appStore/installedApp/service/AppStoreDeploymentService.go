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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/api/bean/gitOps"
	bean3 "github.com/devtron-labs/devtron/api/helm-app/bean"
	bean4 "github.com/devtron-labs/devtron/api/helm-app/gRPC"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
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
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/common"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type AppStoreDeploymentService interface {
	AppStoreDeployOperationDB(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx, skipAppCreation bool, installAppVersionRequestType appStoreBean.InstallAppVersionRequestType) (*appStoreBean.InstallAppVersionDTO, error)
	AppStoreDeployOperationStatusUpdate(installAppId int, status appStoreBean.AppstoreDeploymentStatus) (bool, error)
	IsChartRepoActive(appStoreVersionId int) (bool, error)
	InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error)
	GetInstalledApp(id int) (*appStoreBean.InstallAppVersionDTO, error)
	GetAllInstalledAppsByAppStoreId(w http.ResponseWriter, r *http.Request, token string, appStoreId int) ([]appStoreBean.InstalledAppsResponse, error)
	DeleteInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error)
	LinkHelmApplicationToChartStore(ctx context.Context, request *openapi.UpdateReleaseWithChartLinkingRequest, appIdentifier *service.AppIdentifier, userId int32) (*openapi.UpdateReleaseResponse, bool, error)
	RollbackApplication(ctx context.Context, request *openapi2.RollbackReleaseRequest, installedApp *appStoreBean.InstallAppVersionDTO, userId int32) (bool, error)
	UpdateInstallAppVersionHistory(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*repository.InstalledAppVersionHistory, error)
	GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*bean3.DeploymentHistoryAndInstalledAppInfo, error)
	GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, installedAppVersionHistoryId int) (*openapi.HelmAppDeploymentManifestDetail, error)
	UpdateInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error)
	UpdateInstalledAppVersionHistoryWithGitHash(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) error
	GetInstalledAppVersion(id int, userId int32) (*appStoreBean.InstallAppVersionDTO, error)
	InstallAppByHelm(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error)
	UpdateProjectHelmApp(updateAppRequest *appStoreBean.UpdateProjectHelmAppDTO) error
	UpdatePreviousDeploymentStatusForAppStore(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, triggeredAt time.Time, err error) error
	UpdateInstallAppVersionHistoryStatus(installedAppVersionHistoryId int, status string) error
	MarkGitOpsInstalledAppsDeletedIfArgoAppIsDeleted(installedAppId, envId int) error
}

type DeploymentServiceTypeConfig struct {
	IsInternalUse        bool `env:"IS_INTERNAL_USE" envDefault:"false"`
	HelmInstallASyncMode bool `env:"RUN_HELM_INSTALL_IN_ASYNC_MODE_HELM_APPS" envDefault:"false"`
}

func GetDeploymentServiceTypeConfig() (*DeploymentServiceTypeConfig, error) {
	cfg := &DeploymentServiceTypeConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

type AppStoreDeploymentServiceImpl struct {
	logger                               *zap.SugaredLogger
	installedAppRepository               repository.InstalledAppRepository
	chartGroupDeploymentRepository       repository3.ChartGroupDeploymentRepository
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	environmentRepository                clusterRepository.EnvironmentRepository
	clusterInstalledAppsRepository       repository.ClusterInstalledAppsRepository
	appRepository                        app.AppRepository
	eaModeDeploymentService              EAMode.EAModeDeploymentService
	fullModeDeploymentService            deployment.FullModeDeploymentService
	environmentService                   cluster.EnvironmentService
	clusterService                       cluster.ClusterService
	helmAppService                       service.HelmAppService
	appStoreDeploymentCommonService      appStoreDeploymentCommon.AppStoreDeploymentCommonService
	installedAppRepositoryHistory        repository.InstalledAppVersionHistoryRepository
	deploymentTypeConfig                 *DeploymentServiceTypeConfig
	aCDConfig                            *argocdServer.ACDConfig
	gitOpsConfigReadService              config.GitOpsConfigReadService
}

func NewAppStoreDeploymentServiceImpl(logger *zap.SugaredLogger, installedAppRepository repository.InstalledAppRepository,
	chartGroupDeploymentRepository repository3.ChartGroupDeploymentRepository, appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository, environmentRepository clusterRepository.EnvironmentRepository,
	clusterInstalledAppsRepository repository.ClusterInstalledAppsRepository, appRepository app.AppRepository,
	eaModeDeploymentService EAMode.EAModeDeploymentService,
	fullModeDeploymentService deployment.FullModeDeploymentService, environmentService cluster.EnvironmentService,
	clusterService cluster.ClusterService, helmAppService service.HelmAppService, appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	installedAppRepositoryHistory repository.InstalledAppVersionHistoryRepository,
	deploymentTypeConfig *DeploymentServiceTypeConfig, aCDConfig *argocdServer.ACDConfig,
	gitOpsConfigReadService config.GitOpsConfigReadService) *AppStoreDeploymentServiceImpl {
	return &AppStoreDeploymentServiceImpl{
		logger:                               logger,
		installedAppRepository:               installedAppRepository,
		chartGroupDeploymentRepository:       chartGroupDeploymentRepository,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		environmentRepository:                environmentRepository,
		clusterInstalledAppsRepository:       clusterInstalledAppsRepository,
		appRepository:                        appRepository,
		eaModeDeploymentService:              eaModeDeploymentService,
		fullModeDeploymentService:            fullModeDeploymentService,
		environmentService:                   environmentService,
		clusterService:                       clusterService,
		helmAppService:                       helmAppService,
		appStoreDeploymentCommonService:      appStoreDeploymentCommonService,
		installedAppRepositoryHistory:        installedAppRepositoryHistory,
		deploymentTypeConfig:                 deploymentTypeConfig,
		aCDConfig:                            aCDConfig,
		gitOpsConfigReadService:              gitOpsConfigReadService,
	}
}

func (impl *AppStoreDeploymentServiceImpl) IsChartRepoActive(appStoreVersionId int) (bool, error) {
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(appStoreVersionId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return false, err
	}
	if appStoreAppVersion.AppStore.ChartRepo != nil {
		return appStoreAppVersion.AppStore.ChartRepo.Active, nil
	} else if appStoreAppVersion.AppStore.DockerArtifactStore.OCIRegistryConfig != nil {
		return appStoreAppVersion.AppStore.DockerArtifactStore.OCIRegistryConfig[0].IsChartPullActive, err
	}
	return false, nil
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
	installAppVersionRequest, err = impl.AppStoreDeployOperationDB(installAppVersionRequest, tx, false, appStoreBean.INSTALL_APP_REQUEST)
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

	err = impl.installAppPostDbOperation(installAppVersionRequest)
	if err != nil {
		return nil, err
	}

	return installAppVersionRequest, nil
}

func (impl *AppStoreDeploymentServiceImpl) UpdateInstalledAppVersionHistoryStatus(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, status string) error {
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	savedInstalledAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(installAppVersionRequest.InstalledAppVersionHistoryId)
	savedInstalledAppVersionHistory.Status = status

	_, err = impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistory(savedInstalledAppVersionHistory, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while committing transaction to db", "error", err)
		return err
	}
	return nil
}

func (impl *AppStoreDeploymentServiceImpl) UpdateInstallAppVersionHistory(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*repository.InstalledAppVersionHistory, error) {
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	installedAppVersionHistory := &repository.InstalledAppVersionHistory{
		InstalledAppVersionId: installAppVersionRequest.Id,
	}
	installedAppVersionHistory.ValuesYamlRaw = installAppVersionRequest.ValuesOverrideYaml
	installedAppVersionHistory.CreatedBy = installAppVersionRequest.UserId
	installedAppVersionHistory.CreatedOn = time.Now()
	installedAppVersionHistory.UpdatedBy = installAppVersionRequest.UserId
	installedAppVersionHistory.UpdatedOn = time.Now()
	installedAppVersionHistory.GitHash = installAppVersionRequest.GitHash
	installedAppVersionHistory.Status = "Unknown"
	installedAppVersionHistory, err = impl.installedAppRepositoryHistory.CreateInstalledAppVersionHistory(installedAppVersionHistory, tx)
	if err != nil {
		impl.logger.Errorw("error while creating installed app version history", "error", err)
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while committing transaction to db", "error", err)
		return nil, err
	}
	return installedAppVersionHistory, nil
}

func (impl *AppStoreDeploymentServiceImpl) UpdateInstalledAppVersionHistoryWithGitHash(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) error {
	savedInstalledAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(installAppVersionRequest.InstalledAppVersionHistoryId)
	savedInstalledAppVersionHistory.GitHash = installAppVersionRequest.GitHash
	savedInstalledAppVersionHistory.UpdatedOn = time.Now()
	savedInstalledAppVersionHistory.UpdatedBy = installAppVersionRequest.UserId
	_, err = impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistory(savedInstalledAppVersionHistory, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return err
	}
	return nil
}

func (impl *AppStoreDeploymentServiceImpl) createAppForAppStore(createRequest *bean.CreateAppDTO, tx *pg.Tx, appInstallationMode string, skipAppCreation bool) (*bean.CreateAppDTO, error) {
	app1, err := impl.appRepository.FindActiveByName(createRequest.AppName)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if app1 != nil && app1.Id > 0 {
		impl.logger.Infow(" app already exists", "name", createRequest.AppName)
		err = &util.ApiError{
			Code:            constants.AppAlreadyExists.Code,
			InternalMessage: "app already exists",
			UserMessage:     fmt.Sprintf("app already exists with name %s", createRequest.AppName),
		}

		if !skipAppCreation {
			return nil, err
		} else {
			createRequest.Id = app1.Id
			return createRequest, nil
		}
	}
	pg := &app.App{
		Active:          true,
		AppName:         createRequest.AppName,
		TeamId:          createRequest.TeamId,
		AppType:         helper.ChartStoreApp,
		AppOfferingMode: appInstallationMode,
		AuditLog:        sql.AuditLog{UpdatedBy: createRequest.UserId, CreatedBy: createRequest.UserId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
	}
	err = impl.appRepository.SaveWithTxn(pg, tx)
	if err != nil {
		impl.logger.Errorw("error in saving entity ", "entity", pg)
		return nil, err
	}

	// if found more than 1 application, soft delete all except first item
	apps, err := impl.appRepository.FindActiveListByName(createRequest.AppName)
	if err != nil {
		return nil, err
	}
	appLen := len(apps)
	if appLen > 1 {
		firstElement := apps[0]
		if firstElement.Id != pg.Id {
			pg.Active = false
			err = impl.appRepository.UpdateWithTxn(pg, tx)
			if err != nil {
				impl.logger.Errorw("error in saving entity ", "entity", pg)
				return nil, err
			}
			err = &util.ApiError{
				Code:            constants.AppAlreadyExists.Code,
				InternalMessage: "app already exists",
				UserMessage:     fmt.Sprintf("app already exists with name %s", createRequest.AppName),
			}

			if !skipAppCreation {
				return nil, err
			}
		}
	}

	createRequest.Id = pg.Id
	return createRequest, nil
}

func (impl *AppStoreDeploymentServiceImpl) GetInstalledApp(id int) (*appStoreBean.InstallAppVersionDTO, error) {
	app, err := impl.installedAppRepository.GetInstalledApp(id)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	chartTemplate := adapter.GenerateInstallAppVersionMinDTO(app)
	return chartTemplate, nil
}

func (impl *AppStoreDeploymentServiceImpl) GetAllInstalledAppsByAppStoreId(w http.ResponseWriter, r *http.Request, token string, appStoreId int) ([]appStoreBean.InstalledAppsResponse, error) {
	installedApps, err := impl.installedAppRepository.GetAllInstalledAppsByAppStoreId(appStoreId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return nil, err
	}
	var installedAppsEnvResponse []appStoreBean.InstalledAppsResponse
	for _, a := range installedApps {
		installedAppRes := appStoreBean.InstalledAppsResponse{
			EnvironmentName:              a.EnvironmentName,
			AppName:                      a.AppName,
			DeployedAt:                   a.UpdatedOn,
			DeployedBy:                   a.EmailId,
			Status:                       a.AppStatus,
			AppStoreApplicationVersionId: a.AppStoreApplicationVersionId,
			InstalledAppVersionId:        a.InstalledAppVersionId,
			InstalledAppsId:              a.InstalledAppId,
			EnvironmentId:                a.EnvironmentId,
			AppOfferingMode:              a.AppOfferingMode,
			DeploymentAppType:            a.DeploymentAppType,
		}

		// if hyperion mode app, then fill clusterId and namespace
		if util2.IsHelmApp(a.AppOfferingMode) {
			environment, err := impl.environmentRepository.FindById(a.EnvironmentId)
			if err != nil {
				impl.logger.Errorw("fetching environment error", "err", err)
				return nil, err
			}
			installedAppRes.ClusterId = environment.ClusterId
			installedAppRes.Namespace = environment.Namespace
		}

		installedAppsEnvResponse = append(installedAppsEnvResponse, installedAppRes)
	}
	return installedAppsEnvResponse, nil
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

	environment, err := impl.environmentRepository.FindById(installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}
	if len(environment.Cluster.ErrorInConnecting) > 0 {
		installAppVersionRequest.InstalledAppDeleteResponse.ClusterReachable = false
		installAppVersionRequest.InstalledAppDeleteResponse.ClusterName = environment.Cluster.ClusterName
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
				impl.logger.Errorw("cluster connection error", "err", environment.Cluster.ErrorInConnecting)
				if !installAppVersionRequest.NonCascadeDelete {
					return installAppVersionRequest, nil
				}
			}
			err = impl.fullModeDeploymentService.DeleteACDAppObject(ctx, app.AppName, environment.Name, installAppVersionRequest)
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
			err = impl.eaModeDeploymentService.DeleteInstalledApp(ctx, app.AppName, environment.Name, installAppVersionRequest, model, tx)
		} else {
			err = impl.fullModeDeploymentService.DeleteInstalledApp(ctx, app.AppName, environment.Name, installAppVersionRequest, model, tx)
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
	isChartRepoActive, err := impl.IsChartRepoActive(int(request.GetAppStoreApplicationVersionId()))
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

func (impl *AppStoreDeploymentServiceImpl) createEnvironmentIfNotExists(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (int, error) {
	clusterId := installAppVersionRequest.ClusterId
	namespace := installAppVersionRequest.Namespace
	env, err := impl.environmentRepository.FindOneByNamespaceAndClusterId(namespace, clusterId)

	if err == nil {
		return env.Id, nil
	}

	if err != pg.ErrNoRows {
		return 0, err
	}

	// create env
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		return 0, err
	}

	environmentBean := &cluster2.EnvironmentBean{
		Environment: cluster2.BuildEnvironmentName(cluster.ClusterName, namespace),
		ClusterId:   clusterId,
		Namespace:   namespace,
		Default:     false,
		Active:      true,
	}
	envCreateRes, err := impl.environmentService.Create(environmentBean, installAppVersionRequest.UserId)
	if err != nil {
		return 0, err
	}

	return envCreateRes.Id, nil
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
	if installedApp.InstalledAppId > 0 {
		installedAppModel, err := impl.installedAppRepository.GetInstalledApp(installedApp.InstalledAppId)
		if err != nil {
			impl.logger.Errorw("error while fetching installed app", "error", err)
			return false, err
		}
		installedApp.GitOpsRepoURL = installedAppModel.GitOpsRepoUrl
		// migrate installedApp.GitOpsRepoName to installedApp.GitOpsRepoUrl
		if util.IsAcdApp(installedAppModel.DeploymentAppType) &&
			len(installedAppModel.GitOpsRepoName) != 0 &&
			len(installedAppModel.GitOpsRepoUrl) == 0 {
			//as the installedApp.GitOpsRepoName is not an empty string; migrate installedApp.GitOpsRepoName to installedApp.GitOpsRepoUrl
			migrationErr := impl.handleGitOpsRepoUrlMigration(tx, installedAppModel, userId)
			if migrationErr != nil {
				impl.logger.Errorw("error in GitOps repository url migration", "err", migrationErr)
				return false, err
			}
			installedApp.GitOpsRepoURL = installedAppModel.GitOpsRepoUrl
		}
		// migration ends
	}
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
	installAppVersionRequest, err = impl.AppStoreDeployOperationDB(installAppVersionRequest, tx, true, appStoreBean.INSTALL_APP_REQUEST)
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
	installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
	err = impl.installAppPostDbOperation(installAppVersionRequest)
	if err != nil {
		return nil, err
	}
	// STEP-3 ends

	return res, nil
}

func (impl *AppStoreDeploymentServiceImpl) installAppPostDbOperation(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	//step 4 db operation status update to deploy success
	_, err := impl.AppStoreDeployOperationStatusUpdate(installAppVersionRequest.InstalledAppId, appStoreBean.DEPLOY_SUCCESS)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return err
	}

	//step 5 create build history first entry for install app version for argocd or helm type deployments
	if !impl.deploymentTypeConfig.HelmInstallASyncMode {
		err = impl.updateInstalledAppVersionHistoryWithSync(installAppVersionRequest)
		if err != nil {
			impl.logger.Errorw("error in updating installedApp History with sync ", "err", err)
			return err
		}
	}
	return nil
}

func (impl *AppStoreDeploymentServiceImpl) updateInstalledAppVersionHistoryWithSync(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	if installAppVersionRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD {
		err := impl.UpdateInstalledAppVersionHistoryStatus(installAppVersionRequest, pipelineConfig.WorkflowSucceeded)
		if err != nil {
			impl.logger.Errorw("error on creating history for chart deployment", "error", err)
			return err
		}
	}

	if installAppVersionRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_HELM {
		installedAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(installAppVersionRequest.InstalledAppVersionHistoryId)
		if err != nil {
			impl.logger.Errorw("error in fetching installed app by installed app id in subscribe helm status callback", "err", err)
			return err
		}
		installedAppVersionHistory.Status = pipelineConfig.WorkflowSucceeded
		helmInstallStatus := &appStoreBean.HelmReleaseStatusConfig{
			InstallAppVersionHistoryId: installedAppVersionHistory.Id,
			Message:                    "Release Installed",
			IsReleaseInstalled:         true,
			ErrorInInstallation:        false,
		}
		data, err := json.Marshal(helmInstallStatus)
		if err != nil {
			impl.logger.Errorw("error in marshalling helmInstallStatus message")
			return err
		}
		installedAppVersionHistory.HelmReleaseStatusConfig = string(data)
		_, err = impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistory(installedAppVersionHistory, nil)
		if err != nil {
			impl.logger.Errorw("error in updating helm release status data in installedAppVersionHistoryRepository", "err", err)
			return err
		}
	}
	return nil
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

func (impl *AppStoreDeploymentServiceImpl) updateInstalledAppVersion(installedAppVersion *repository.InstalledAppVersions, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx, installedApp *repository.InstalledApps) (*repository.InstalledAppVersions, *appStoreBean.InstallAppVersionDTO, error) {
	var err error
	if installAppVersionRequest.Id == 0 {
		installedAppVersions, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppId(installAppVersionRequest.InstalledAppId)
		if err != nil {
			impl.logger.Errorw("error while fetching installed version", "error", err)
			return installedAppVersion, installAppVersionRequest, err
		}
		for _, installedAppVersionModel := range installedAppVersions {
			installedAppVersionModel.Active = false
			installedAppVersionModel.UpdatedOn = time.Now()
			installedAppVersionModel.UpdatedBy = installAppVersionRequest.UserId
			_, err = impl.installedAppRepository.UpdateInstalledAppVersion(installedAppVersionModel, tx)
			if err != nil {
				impl.logger.Errorw("error while update installed chart", "error", err)
				return installedAppVersion, installAppVersionRequest, err
			}
		}

		appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
		if err != nil {
			impl.logger.Errorw("fetching error", "err", err)
			return installedAppVersion, installAppVersionRequest, err
		}
		installedAppVersion = &repository.InstalledAppVersions{
			InstalledAppId:               installAppVersionRequest.InstalledAppId,
			AppStoreApplicationVersionId: installAppVersionRequest.AppStoreVersion,
			ValuesYaml:                   installAppVersionRequest.ValuesOverrideYaml,
		}
		installedAppVersion.CreatedBy = installAppVersionRequest.UserId
		installedAppVersion.UpdatedBy = installAppVersionRequest.UserId
		installedAppVersion.CreatedOn = time.Now()
		installedAppVersion.UpdatedOn = time.Now()
		installedAppVersion.Active = true
		installedAppVersion.ReferenceValueId = installAppVersionRequest.ReferenceValueId
		installedAppVersion.ReferenceValueKind = installAppVersionRequest.ReferenceValueKind
		_, err = impl.installedAppRepository.CreateInstalledAppVersion(installedAppVersion, tx)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return installedAppVersion, installAppVersionRequest, err
		}
		installedAppVersion.AppStoreApplicationVersion = *appStoreAppVersion
		installedAppVersion.InstalledApp = *installedApp
		installAppVersionRequest.InstalledAppVersionId = installedAppVersion.Id
	} else {
		installedAppVersionModel, err := impl.installedAppRepository.GetInstalledAppVersion(installAppVersionRequest.Id)
		if err != nil {
			impl.logger.Errorw("error while fetching chart installed version", "error", err)
			return installedAppVersion, installAppVersionRequest, err
		}
		if installedAppVersionModel.AppStoreApplicationVersionId != installAppVersionRequest.AppStoreVersion {
			// upgrade to new version of same chart
			installedAppVersionModel.Active = false
			installedAppVersionModel.UpdatedOn = time.Now()
			installedAppVersionModel.UpdatedBy = installAppVersionRequest.UserId
			_, err = impl.installedAppRepository.UpdateInstalledAppVersion(installedAppVersionModel, tx)
			if err != nil {
				impl.logger.Errorw("error while fetching from db", "error", err)
				return installedAppVersion, installAppVersionRequest, err
			}
			appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
			if err != nil {
				impl.logger.Errorw("fetching error", "err", err)
				return installedAppVersion, installAppVersionRequest, err
			}
			installedAppVersion = &repository.InstalledAppVersions{
				InstalledAppId:               installAppVersionRequest.InstalledAppId,
				AppStoreApplicationVersionId: installAppVersionRequest.AppStoreVersion,
				ValuesYaml:                   installAppVersionRequest.ValuesOverrideYaml,
			}
			installedAppVersion.CreatedBy = installAppVersionRequest.UserId
			installedAppVersion.UpdatedBy = installAppVersionRequest.UserId
			installedAppVersion.CreatedOn = time.Now()
			installedAppVersion.UpdatedOn = time.Now()
			installedAppVersion.Active = true
			installedAppVersion.ReferenceValueId = installAppVersionRequest.ReferenceValueId
			installedAppVersion.ReferenceValueKind = installAppVersionRequest.ReferenceValueKind
			_, err = impl.installedAppRepository.CreateInstalledAppVersion(installedAppVersion, tx)
			if err != nil {
				impl.logger.Errorw("error while fetching from db", "error", err)
				return installedAppVersion, installAppVersionRequest, err
			}
			installedAppVersion.AppStoreApplicationVersion = *appStoreAppVersion
		} else {
			installedAppVersion = installedAppVersionModel
		}

	}
	return installedAppVersion, installAppVersionRequest, err
}

func (impl *AppStoreDeploymentServiceImpl) MarkInstalledAppVersionsInactiveByInstalledAppId(installedAppId int, UserId int32, tx *pg.Tx) error {
	installedAppVersions, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppId(installedAppId)
	if err != nil {
		impl.logger.Errorw("error while fetching installed version", "error", err)
		return err
	}
	for _, installedAppVersionModel := range installedAppVersions {
		installedAppVersionModel.Active = false
		installedAppVersionModel.UpdatedOn = time.Now()
		installedAppVersionModel.UpdatedBy = UserId
		_, err = impl.installedAppRepository.UpdateInstalledAppVersion(installedAppVersionModel, tx)
		if err != nil {
			impl.logger.Errorw("error while update installed chart", "error", err)
			return err
		}
	}
	return nil
}

func (impl *AppStoreDeploymentServiceImpl) MarkInstalledAppVersionModelInActive(installedAppVersionModel *repository.InstalledAppVersions, UserId int32, tx *pg.Tx) error {
	installedAppVersionModel.Active = false
	installedAppVersionModel.UpdatedOn = time.Now()
	installedAppVersionModel.UpdatedBy = UserId
	_, err := impl.installedAppRepository.UpdateInstalledAppVersion(installedAppVersionModel, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return err
	}
	return nil
}

func (impl *AppStoreDeploymentServiceImpl) CreateInstalledAppVersion(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*repository.InstalledAppVersions, error) {
	// TODO fix me next
	// TODO refactoring: move this to adapter
	installedAppVersion := &repository.InstalledAppVersions{
		InstalledAppId:               installAppVersionRequest.InstalledAppId,
		AppStoreApplicationVersionId: installAppVersionRequest.AppStoreVersion,
		ValuesYaml:                   installAppVersionRequest.ValuesOverrideYaml,
	}
	installedAppVersion.CreatedBy = installAppVersionRequest.UserId
	installedAppVersion.UpdatedBy = installAppVersionRequest.UserId
	installedAppVersion.CreatedOn = time.Now()
	installedAppVersion.UpdatedOn = time.Now()
	installedAppVersion.Active = true
	installedAppVersion.ReferenceValueId = installAppVersionRequest.ReferenceValueId
	installedAppVersion.ReferenceValueKind = installAppVersionRequest.ReferenceValueKind
	_, err := impl.installedAppRepository.CreateInstalledAppVersion(installedAppVersion, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	return installedAppVersion, nil
}

// CheckIfMonoRepoMigrationRequired checks if gitOps repo name is changed
func (impl *AppStoreDeploymentServiceImpl) CheckIfMonoRepoMigrationRequired(installedApp *repository.InstalledApps) bool {
	monoRepoMigrationRequired := false
	if !util.IsAcdApp(installedApp.DeploymentAppType) || gitOps.IsGitOpsRepoNotConfigured(installedApp.GitOpsRepoUrl) || installedApp.IsCustomRepository {
		return false
	}
	var err error
	gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(installedApp.GitOpsRepoUrl)
	if len(gitOpsRepoName) == 0 {
		gitOpsRepoName, err = impl.fullModeDeploymentService.GetAcdAppGitOpsRepoName(installedApp.App.AppName, installedApp.Environment.Name)
		if err != nil || gitOpsRepoName == "" {
			return false
		}
	}
	//here will set new git repo name if required to migrate
	newGitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoName(installedApp.App.AppName)
	//checking weather git repo migration needed or not, if existing git repo and new independent git repo is not same than go ahead with migration
	if newGitOpsRepoName != gitOpsRepoName {
		monoRepoMigrationRequired = true
	}
	return monoRepoMigrationRequired
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

	installedApp, err := impl.installedAppRepository.GetInstalledApp(installAppVersionRequest.InstalledAppId)
	if err != nil {
		return nil, err
	}
	installAppVersionRequest.UpdateDeploymentAppType(installedApp.DeploymentAppType)

	installedAppDeploymentAction := adapter.NewInstalledAppDeploymentAction(installedApp.DeploymentAppType)
	// migrate installedApp.GitOpsRepoName to installedApp.GitOpsRepoUrl
	if util.IsAcdApp(installedApp.DeploymentAppType) &&
		len(installedApp.GitOpsRepoName) != 0 &&
		len(installedApp.GitOpsRepoUrl) == 0 {
		//as the installedApp.GitOpsRepoName is not an empty string; migrate installedApp.GitOpsRepoName to installedApp.GitOpsRepoUrl
		gitRepoUrl, err := impl.fullModeDeploymentService.GetGitRepoUrl(installedApp.GitOpsRepoName)
		if err != nil {
			impl.logger.Errorw("error in GitOps repository url migration", "err", err)
			return nil, err
		}
		installedApp.GitOpsRepoUrl = gitRepoUrl
	}
	// migration ends
	var installedAppVersion *repository.InstalledAppVersions

	// mark previous versions of chart as inactive if chart or version is updated

	isChartChanged := false   // flag for keeping track if chart is updated by user or not
	isVersionChanged := false // flag for keeping track if version of chart is upgraded

	//if chart is changed, then installedAppVersion id is sent as 0 from front-end
	if installAppVersionRequest.Id == 0 {
		isChartChanged = true
		err = impl.MarkInstalledAppVersionsInactiveByInstalledAppId(installAppVersionRequest.InstalledAppId, installAppVersionRequest.UserId, tx)
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
			err = impl.MarkInstalledAppVersionModelInActive(installedAppVersion, installAppVersionRequest.UserId, tx)
		}
	}

	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}
	// create new entry for installed app version if chart or version is changed
	if isChartChanged || isVersionChanged {
		installedAppVersion, err = impl.CreateInstalledAppVersion(installAppVersionRequest, tx)
		if err != nil {
			return nil, err
		}
		installedAppVersion.Id = installedAppVersion.Id
	} else {
		// TODO fix me next
		// TODO refactoring: move this to adapter
		installedAppVersion.ValuesYaml = installAppVersionRequest.ValuesOverrideYaml
		installedAppVersion.UpdatedOn = time.Now()
		installedAppVersion.UpdatedBy = installAppVersionRequest.UserId
		installedAppVersion.ReferenceValueId = installAppVersionRequest.ReferenceValueId
		installedAppVersion.ReferenceValueKind = installAppVersionRequest.ReferenceValueKind
		_, err = impl.installedAppRepository.UpdateInstalledAppVersion(installedAppVersion, tx)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
	}

	installedAppVersion.AppStoreApplicationVersion = *appStoreAppVersion
	installedAppVersion.InstalledApp = *installedApp

	installAppVersionRequest.InstalledAppVersionId = installedAppVersion.Id
	installAppVersionRequest.Id = installedAppVersion.Id
	installAppVersionRequest.EnvironmentId = installedApp.EnvironmentId
	installAppVersionRequest.AppName = installedApp.App.AppName
	installAppVersionRequest.EnvironmentName = installedApp.Environment.Name
	installAppVersionRequest.Environment = &installedApp.Environment

	installAppVersionHistoryStatus := pipelineConfig.WorkflowInProgress
	// TODO fix me next
	// TODO refactoring: move this to adapter
	installedAppVersionHistory := &repository.InstalledAppVersionHistory{
		InstalledAppVersionId: installedAppVersion.Id,
		ValuesYamlRaw:         installAppVersionRequest.ValuesOverrideYaml,
		StartedOn:             time.Now(),
		Status:                installAppVersionHistoryStatus,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: installAppVersionRequest.UserId,
			UpdatedOn: time.Now(),
			UpdatedBy: installAppVersionRequest.UserId,
		},
	}
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
			_ = impl.UpdateInstalledAppVersionHistoryStatus(installAppVersionRequest, pipelineConfig.WorkflowFailed)
			return nil, err
		}
		// required if gitOps repo name is changed, gitOps repo name will change if env variable which we use as suffix changes
		monoRepoMigrationRequired = impl.CheckIfMonoRepoMigrationRequired(installedApp)
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
	installedApp.Status = appStoreBean.DEPLOY_SUCCESS
	installedApp.UpdatedOn = time.Now()
	installedAppVersion.UpdatedBy = installAppVersionRequest.UserId
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
		err = impl.UpdateInstalledAppVersionHistoryStatus(installAppVersionRequest, pipelineConfig.WorkflowSucceeded)
		if err != nil {
			impl.logger.Errorw("error on creating history for chart deployment", "error", err)
			return nil, err
		}
	} else if util.IsHelmApp(installAppVersionRequest.DeploymentAppType) && !impl.deploymentTypeConfig.HelmInstallASyncMode {
		err = impl.updateInstalledAppVersionHistoryWithSync(installAppVersionRequest)
		if err != nil {
			impl.logger.Errorw("error in updating install app version history on sync", "err", err)
			return nil, err
		}
	}

	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentServiceImpl) handleGitOpsRepoUrlMigration(tx *pg.Tx, installedApp *repository.InstalledApps, userId int32) error {
	var (
		localTx *pg.Tx
		err     error
	)

	if tx == nil {
		dbConnection := impl.installedAppRepository.GetConnection()
		localTx, err = dbConnection.Begin()
		if err != nil {
			return err
		}
		// Rollback tx on error.
		defer localTx.Rollback()
	}

	gitRepoUrl, err := impl.fullModeDeploymentService.GetGitRepoUrl(installedApp.GitOpsRepoName)
	if err != nil {
		impl.logger.Errorw("error in GitOps repository url migration", "err", err)
		return err
	}
	installedApp.GitOpsRepoUrl = gitRepoUrl
	installedApp.UpdatedOn = time.Now()
	installedApp.UpdatedBy = userId
	if localTx != nil {
		_, err = impl.installedAppRepository.UpdateInstalledApp(installedApp, localTx)
	} else {
		_, err = impl.installedAppRepository.UpdateInstalledApp(installedApp, tx)
	}
	if err != nil {
		impl.logger.Errorw("error in updating installed app model", "err", err)
		return err
	}
	if localTx != nil {
		err = localTx.Commit()
		if err != nil {
			impl.logger.Errorw("error while committing transaction to db", "error", err)
			return err
		}
	}
	return err
}
func (impl *AppStoreDeploymentServiceImpl) GetInstalledAppVersion(id int, userId int32) (*appStoreBean.InstallAppVersionDTO, error) {
	app, err := impl.installedAppRepository.GetInstalledAppVersion(id)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, &util.ApiError{HttpStatusCode: http.StatusBadRequest, Code: "400", UserMessage: "values are outdated. please fetch the latest version and try again", InternalMessage: err.Error()}
		}
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	// migrate installedApp.GitOpsRepoName to installedApp.GitOpsRepoUrl
	if util.IsAcdApp(app.InstalledApp.DeploymentAppType) &&
		len(app.InstalledApp.GitOpsRepoName) != 0 &&
		len(app.InstalledApp.GitOpsRepoUrl) == 0 {
		//as the installedApp.GitOpsRepoName is not an empty string; migrate installedApp.GitOpsRepoName to installedApp.GitOpsRepoUrl
		// db operations
		installedAppModel := &app.InstalledApp
		migrationErr := impl.handleGitOpsRepoUrlMigration(nil, installedAppModel, userId)
		if migrationErr != nil {
			impl.logger.Errorw("error in GitOps repository url migration", "err", migrationErr)
			return nil, err
		}
	}
	// migration ends
	installAppVersion := &appStoreBean.InstallAppVersionDTO{
		InstalledAppId:     app.InstalledAppId,
		AppName:            app.InstalledApp.App.AppName,
		AppId:              app.InstalledApp.App.Id,
		Id:                 app.Id,
		TeamId:             app.InstalledApp.App.TeamId,
		EnvironmentId:      app.InstalledApp.EnvironmentId,
		ValuesOverrideYaml: app.ValuesYaml,
		Readme:             app.AppStoreApplicationVersion.Readme,
		ReferenceValueKind: app.ReferenceValueKind,
		ReferenceValueId:   app.ReferenceValueId,
		AppStoreVersion:    app.AppStoreApplicationVersionId, //check viki
		Status:             app.InstalledApp.Status,
		AppStoreId:         app.AppStoreApplicationVersion.AppStoreId,
		AppStoreName:       app.AppStoreApplicationVersion.AppStore.Name,
		Deprecated:         app.AppStoreApplicationVersion.Deprecated,
		GitOpsRepoURL:      app.InstalledApp.GitOpsRepoUrl,
		UserId:             userId,
		AppOfferingMode:    app.InstalledApp.App.AppOfferingMode,
		ClusterId:          app.InstalledApp.Environment.ClusterId,
		Namespace:          app.InstalledApp.Environment.Namespace,
		DeploymentAppType:  app.InstalledApp.DeploymentAppType,
		Environment:        &app.InstalledApp.Environment,
		ACDAppName:         fmt.Sprintf("%s-%s", app.InstalledApp.App.AppName, app.InstalledApp.Environment.Name),
	}
	return installAppVersion, err
}

func (impl *AppStoreDeploymentServiceImpl) InstallAppByHelm(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error) {
	installAppVersionRequest, err := impl.eaModeDeploymentService.InstallApp(installAppVersionRequest, nil, ctx, nil)
	if err != nil {
		impl.logger.Errorw("error while installing app via helm", "error", err)
		return installAppVersionRequest, err
	}
	if !impl.deploymentTypeConfig.HelmInstallASyncMode {
		err = impl.updateInstalledAppVersionHistoryWithSync(installAppVersionRequest)
		if err != nil {
			impl.logger.Errorw("error in updating installed app version history with sync", "err", err)
			return installAppVersionRequest, err
		}
	}
	return installAppVersionRequest, nil
}

func (impl *AppStoreDeploymentServiceImpl) UpdateProjectHelmApp(updateAppRequest *appStoreBean.UpdateProjectHelmAppDTO) error {

	appIdSplitted := strings.Split(updateAppRequest.AppId, "|")

	appName := updateAppRequest.AppName

	if len(appIdSplitted) > 1 {
		// app id is zero for CLI apps
		appIdentifier, _ := impl.helmAppService.DecodeAppId(updateAppRequest.AppId)
		appName = appIdentifier.ReleaseName
	}

	app, err := impl.appRepository.FindActiveByName(appName)

	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching app", "err", err)
		return err
	}

	var appInstallationMode string

	dbConnection := impl.appRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	impl.logger.Infow("update helm project request", updateAppRequest)
	if app.Id == 0 {
		// for cli Helm app, if app is not yet created
		if util2.IsBaseStack() {
			appInstallationMode = util2.SERVER_MODE_HYPERION
		} else {
			appInstallationMode = util2.SERVER_MODE_FULL
		}
		createAppRequest := bean.CreateAppDTO{
			AppName: appName,
			UserId:  updateAppRequest.UserId,
			TeamId:  updateAppRequest.TeamId,
		}
		_, err := impl.createAppForAppStore(&createAppRequest, tx, appInstallationMode, false)
		if err != nil {
			impl.logger.Errorw("error while creating app", "error", err)
			return err
		}
	} else {
		// update team id if app exist
		app.TeamId = updateAppRequest.TeamId
		app.UpdatedOn = time.Now()
		app.UpdatedBy = updateAppRequest.UserId
		err = impl.appRepository.UpdateWithTxn(app, tx)

		if err != nil {
			impl.logger.Errorw("error in updating project", "err", err)
			return err
		}
	}
	tx.Commit()
	return nil

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

func (impl *AppStoreDeploymentServiceImpl) UpdateInstallAppVersionHistoryStatus(installedAppVersionHistoryId int, status string) error {
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	savedInstalledAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(installedAppVersionHistoryId)
	savedInstalledAppVersionHistory.Status = status
	_, err = impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistory(savedInstalledAppVersionHistory, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while committing transaction to db", "error", err)
		return err
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
