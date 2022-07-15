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
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/internal/constants"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDeploymentTool "github.com/devtron-labs/devtron/pkg/appStore/deployment/tool"
	appStoreDeploymentGitopsTool "github.com/devtron-labs/devtron/pkg/appStore/deployment/tool/gitops"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type AppStoreDeploymentService interface {
	AppStoreDeployOperationDB(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	AppStoreDeployOperationStatusUpdate(installAppId int, status appStoreBean.AppstoreDeploymentStatus) (bool, error)
	IsChartRepoActive(appStoreVersionId int) (bool, error)
	InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error)
	GetInstalledApp(id int) (*appStoreBean.InstallAppVersionDTO, error)
	GetAllInstalledAppsByAppStoreId(w http.ResponseWriter, r *http.Request, token string, appStoreId int) ([]appStoreBean.InstalledAppsResponse, error)
	DeleteInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error)
	LinkHelmApplicationToChartStore(ctx context.Context, request *openapi.UpdateReleaseWithChartLinkingRequest, appIdentifier *client.AppIdentifier, userId int32) (*openapi.UpdateReleaseResponse, bool, error)
	RollbackApplication(ctx context.Context, request *openapi2.RollbackReleaseRequest, installedApp *appStoreBean.InstallAppVersionDTO, userId int32) (bool, error)
	UpdateInstallAppVersionHistory(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error
	GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*client.DeploymentHistoryAndInstalledAppInfo, error)
	GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, installedAppVersionHistoryId int) (*openapi.HelmAppDeploymentManifestDetail, error)
	UpdateInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error)
	GetInstalledAppVersion(id int, userId int32) (*appStoreBean.InstallAppVersionDTO, error)
	InstallAppByHelm(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error)
}

type AppStoreDeploymentServiceImpl struct {
	logger                               *zap.SugaredLogger
	installedAppRepository               repository.InstalledAppRepository
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	environmentRepository                clusterRepository.EnvironmentRepository
	clusterInstalledAppsRepository       repository.ClusterInstalledAppsRepository
	appRepository                        app.AppRepository
	appStoreDeploymentHelmService        appStoreDeploymentTool.AppStoreDeploymentHelmService
	appStoreDeploymentArgoCdService      appStoreDeploymentGitopsTool.AppStoreDeploymentArgoCdService
	environmentService                   cluster.EnvironmentService
	clusterService                       cluster.ClusterService
	helmAppService                       client.HelmAppService
	appStoreDeploymentCommonService      appStoreDeploymentCommon.AppStoreDeploymentCommonService
	globalEnvVariables                   *util2.GlobalEnvVariables
	installedAppRepositoryHistory        repository.InstalledAppVersionHistoryRepository
	gitOpsRepository                     repository2.GitOpsConfigRepository
}

func NewAppStoreDeploymentServiceImpl(logger *zap.SugaredLogger, installedAppRepository repository.InstalledAppRepository,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository, environmentRepository clusterRepository.EnvironmentRepository,
	clusterInstalledAppsRepository repository.ClusterInstalledAppsRepository, appRepository app.AppRepository,
	appStoreDeploymentHelmService appStoreDeploymentTool.AppStoreDeploymentHelmService,
	appStoreDeploymentArgoCdService appStoreDeploymentGitopsTool.AppStoreDeploymentArgoCdService, environmentService cluster.EnvironmentService,
	clusterService cluster.ClusterService, helmAppService client.HelmAppService, appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	globalEnvVariables *util2.GlobalEnvVariables,
	installedAppRepositoryHistory repository.InstalledAppVersionHistoryRepository, gitOpsRepository repository2.GitOpsConfigRepository) *AppStoreDeploymentServiceImpl {
	return &AppStoreDeploymentServiceImpl{
		logger:                               logger,
		installedAppRepository:               installedAppRepository,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		environmentRepository:                environmentRepository,
		clusterInstalledAppsRepository:       clusterInstalledAppsRepository,
		appRepository:                        appRepository,
		appStoreDeploymentHelmService:        appStoreDeploymentHelmService,
		appStoreDeploymentArgoCdService:      appStoreDeploymentArgoCdService,
		environmentService:                   environmentService,
		clusterService:                       clusterService,
		helmAppService:                       helmAppService,
		appStoreDeploymentCommonService:      appStoreDeploymentCommonService,
		globalEnvVariables:                   globalEnvVariables,
		installedAppRepositoryHistory:        installedAppRepositoryHistory,
		gitOpsRepository:                     gitOpsRepository,
	}
}

func (impl AppStoreDeploymentServiceImpl) AppStoreDeployOperationDB(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
	isGitOpsConfigured := false
	gitOpsConfig, err := impl.gitOpsRepository.GetGitOpsConfigActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GetGitOpsConfigActive, error while getting", "err", err)
		return nil, err
	}
	if gitOpsConfig != nil && gitOpsConfig.Id > 0 {
		isGitOpsConfigured = true
	}

	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}

	var appInstallationMode string
	if util2.GetDevtronVersion().ServerMode == util2.SERVER_MODE_HYPERION || installAppVersionRequest.AppOfferingMode == util2.SERVER_MODE_HYPERION {
		appInstallationMode = util2.SERVER_MODE_HYPERION
	} else {
		appInstallationMode = util2.SERVER_MODE_FULL
	}

	// create env if env not exists for clusterId and namespace for hyperion mode
	if appInstallationMode == util2.SERVER_MODE_HYPERION {
		envId, err := impl.createEnvironmentIfNotExists(installAppVersionRequest)
		if err != nil {
			return nil, err
		}
		installAppVersionRequest.EnvironmentId = envId
	}

	environment, err := impl.environmentRepository.FindById(installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}
	installAppVersionRequest.ClusterId = environment.ClusterId
	appCreateRequest := &bean.CreateAppDTO{
		Id:      installAppVersionRequest.AppId,
		AppName: installAppVersionRequest.AppName,
		TeamId:  installAppVersionRequest.TeamId,
		UserId:  installAppVersionRequest.UserId,
	}

	appCreateRequest, err = impl.createAppForAppStore(appCreateRequest, tx, appInstallationMode)
	if err != nil {
		impl.logger.Errorw("error while creating app", "error", err)
		return nil, err
	}
	installAppVersionRequest.AppId = appCreateRequest.Id

	installedAppModel := &repository.InstalledApps{
		AppId:         appCreateRequest.Id,
		EnvironmentId: environment.Id,
		Status:        appStoreBean.DEPLOY_INIT,
	}
	if isGitOpsConfigured && appInstallationMode == util2.SERVER_MODE_FULL {
		installedAppModel.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
	} else {
		installedAppModel.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
	}
	installAppVersionRequest.DeploymentAppType = installedAppModel.DeploymentAppType
	installedAppModel.CreatedBy = installAppVersionRequest.UserId
	installedAppModel.UpdatedBy = installAppVersionRequest.UserId
	installedAppModel.CreatedOn = time.Now()
	installedAppModel.UpdatedOn = time.Now()
	installedAppModel.Active = true
	if util2.GetDevtronVersion().ServerMode == util2.SERVER_MODE_FULL {
		installedAppModel.GitOpsRepoName = impl.GetGitOpsRepoName(appStoreAppVersion.AppStore.Name)
		installAppVersionRequest.GitOpsRepoName = installedAppModel.GitOpsRepoName
	}
	installedApp, err := impl.installedAppRepository.CreateInstalledApp(installedAppModel, tx)
	if err != nil {
		impl.logger.Errorw("error while creating install app", "error", err)
		return nil, err
	}
	installAppVersionRequest.InstalledAppId = installedApp.Id

	installedAppVersions := &repository.InstalledAppVersions{
		InstalledAppId:               installAppVersionRequest.InstalledAppId,
		AppStoreApplicationVersionId: appStoreAppVersion.Id,
		ValuesYaml:                   installAppVersionRequest.ValuesOverrideYaml,
		//Values:                       "{}",
	}
	installedAppVersions.CreatedBy = installAppVersionRequest.UserId
	installedAppVersions.UpdatedBy = installAppVersionRequest.UserId
	installedAppVersions.CreatedOn = time.Now()
	installedAppVersions.UpdatedOn = time.Now()
	installedAppVersions.Active = true
	installedAppVersions.ReferenceValueId = installAppVersionRequest.ReferenceValueId
	installedAppVersions.ReferenceValueKind = installAppVersionRequest.ReferenceValueKind
	_, err = impl.installedAppRepository.CreateInstalledAppVersion(installedAppVersions, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	installAppVersionRequest.InstalledAppVersionId = installedAppVersions.Id
	installAppVersionRequest.Id = installedAppVersions.Id
	if installAppVersionRequest.DefaultClusterComponent {
		clusterInstalledAppsModel := &repository.ClusterInstalledApps{
			ClusterId:      environment.ClusterId,
			InstalledAppId: installAppVersionRequest.InstalledAppId,
		}
		clusterInstalledAppsModel.CreatedBy = installAppVersionRequest.UserId
		clusterInstalledAppsModel.UpdatedBy = installAppVersionRequest.UserId
		clusterInstalledAppsModel.CreatedOn = time.Now()
		clusterInstalledAppsModel.UpdatedOn = time.Now()
		err = impl.clusterInstalledAppsRepository.Save(clusterInstalledAppsModel, tx)
		if err != nil {
			impl.logger.Errorw("error while creating cluster install app", "error", err)
			return nil, err
		}
	}
	return installAppVersionRequest, nil
}

//TODO - dedupe, move it to one location
func (impl AppStoreDeploymentServiceImpl) GetGitOpsRepoName(appName string) string {
	var repoName string
	if len(impl.globalEnvVariables.GitOpsRepoPrefix) == 0 {
		repoName = appName
	} else {
		repoName = fmt.Sprintf("%s-%s", impl.globalEnvVariables.GitOpsRepoPrefix, appName)
	}
	return repoName
}

func (impl AppStoreDeploymentServiceImpl) AppStoreDeployOperationStatusUpdate(installAppId int, status appStoreBean.AppstoreDeploymentStatus) (bool, error) {
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return false, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	installedApp, err := impl.installedAppRepository.GetInstalledApp(installAppId)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	installedApp.Status = status
	_, err = impl.installedAppRepository.UpdateInstalledApp(installedApp, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while commit db transaction to db", "error", err)
		return false, err
	}
	return true, nil
}

func (impl *AppStoreDeploymentServiceImpl) IsChartRepoActive(appStoreVersionId int) (bool, error) {
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(appStoreVersionId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return false, err
	}
	return appStoreAppVersion.AppStore.ChartRepo.Active, nil
}

func (impl AppStoreDeploymentServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error) {

	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	//step 1 db operation initiated
	installAppVersionRequest, err = impl.AppStoreDeployOperationDB(installAppVersionRequest, tx)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return nil, err
	}

	if util2.GetDevtronVersion().ServerMode == util2.SERVER_MODE_HYPERION || installAppVersionRequest.AppOfferingMode == util2.SERVER_MODE_HYPERION ||
		installAppVersionRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_HELM {
		installAppVersionRequest, err = impl.appStoreDeploymentHelmService.InstallApp(installAppVersionRequest, ctx)
	} else {
		installAppVersionRequest, err = impl.appStoreDeploymentArgoCdService.InstallApp(installAppVersionRequest, ctx)
	}

	if err != nil {
		return nil, err
	}

	// tx commit here because next operation will be process after this commit.
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	err = impl.installAppPostDbOperation(installAppVersionRequest)
	if err != nil {
		return nil, err
	}

	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentServiceImpl) UpdateInstallAppVersionHistory(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
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
	_, err = impl.installedAppRepositoryHistory.CreateInstalledAppVersionHistory(installedAppVersionHistory, tx)
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

func (impl AppStoreDeploymentServiceImpl) createAppForAppStore(createRequest *bean.CreateAppDTO, tx *pg.Tx, appInstallationMode string) (*bean.CreateAppDTO, error) {
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
		return nil, err
	}
	pg := &app.App{
		Active:          true,
		AppName:         createRequest.AppName,
		TeamId:          createRequest.TeamId,
		AppStore:        true,
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
			return nil, err
		}
	}

	createRequest.Id = pg.Id
	return createRequest, nil
}

func (impl AppStoreDeploymentServiceImpl) GetInstalledApp(id int) (*appStoreBean.InstallAppVersionDTO, error) {
	app, err := impl.installedAppRepository.GetInstalledApp(id)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	chartTemplate := impl.chartAdaptor2(app)
	return chartTemplate, nil
}

//converts db object to bean
func (impl AppStoreDeploymentServiceImpl) chartAdaptor2(chart *repository.InstalledApps) *appStoreBean.InstallAppVersionDTO {
	return &appStoreBean.InstallAppVersionDTO{
		EnvironmentId:   chart.EnvironmentId,
		InstalledAppId:  chart.Id,
		AppId:           chart.AppId,
		AppOfferingMode: chart.App.AppOfferingMode,
		ClusterId:       chart.Environment.ClusterId,
		Namespace:       chart.Environment.Namespace,
		AppName:         chart.App.AppName,
		EnvironmentName: chart.Environment.Name,
	}
}

func (impl AppStoreDeploymentServiceImpl) GetAllInstalledAppsByAppStoreId(w http.ResponseWriter, r *http.Request, token string, appStoreId int) ([]appStoreBean.InstalledAppsResponse, error) {
	installedApps, err := impl.installedAppRepository.GetAllIntalledAppsByAppStoreId(appStoreId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return nil, err
	}
	var installedAppsEnvResponse []appStoreBean.InstalledAppsResponse
	for _, a := range installedApps {
		var status string
		if util2.GetDevtronVersion().ServerMode == util2.SERVER_MODE_HYPERION || a.AppOfferingMode == util2.SERVER_MODE_HYPERION ||
			a.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_HELM {
			status, err = impl.appStoreDeploymentHelmService.GetAppStatus(a, w, r, token)
		} else {
			status, err = impl.appStoreDeploymentArgoCdService.GetAppStatus(a, w, r, token)
		}
		if apiErr, ok := err.(*util.ApiError); ok {
			if apiErr.Code == constants.AppDetailResourceTreeNotFound {
				status = "Not Found"
			}
		} else if err != nil {
			impl.logger.Error(err)
			return nil, err
		}
		installedAppRes := appStoreBean.InstalledAppsResponse{
			EnvironmentName:              a.EnvironmentName,
			AppName:                      a.AppName,
			DeployedAt:                   a.UpdatedOn,
			DeployedBy:                   a.EmailId,
			Status:                       status,
			AppStoreApplicationVersionId: a.AppStoreApplicationVersionId,
			InstalledAppVersionId:        a.InstalledAppVersionId,
			InstalledAppsId:              a.InstalledAppId,
			EnvironmentId:                a.EnvironmentId,
			AppOfferingMode:              a.AppOfferingMode,
		}

		// if hyperion mode app, then fill clusterId and namespace
		if a.AppOfferingMode == util2.SERVER_MODE_HYPERION {
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

func (impl AppStoreDeploymentServiceImpl) DeleteInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error) {

	environment, err := impl.environmentRepository.FindById(installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}

	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	app, err := impl.appRepository.FindById(installAppVersionRequest.AppId)
	if err != nil {
		return nil, err
	}
	app.Active = false
	app.UpdatedBy = installAppVersionRequest.UserId
	app.UpdatedOn = time.Now()
	err = impl.appRepository.UpdateWithTxn(app, tx)
	if err != nil {
		impl.logger.Errorw("error in update entity ", "entity", app)
		return nil, err
	}

	model, err := impl.installedAppRepository.GetInstalledApp(installAppVersionRequest.InstalledAppId)
	if err != nil {
		impl.logger.Errorw("error in fetching installed app", "id", installAppVersionRequest.InstalledAppId, "err", err)
		return nil, err
	}
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

	if util2.GetDevtronVersion().ServerMode == util2.SERVER_MODE_HYPERION || app.AppOfferingMode == util2.SERVER_MODE_HYPERION ||
		model.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_HELM {
		// there might be a case if helm release gets uninstalled from helm cli.
		//in this case on deleting the app from API, it should not give error as it should get deleted from db, otherwise due to delete error, db does not get clean
		// so in helm, we need to check first if the release exists or not, if exists then only delete
		err = impl.appStoreDeploymentHelmService.DeleteInstalledApp(ctx, app.AppName, environment.Name, installAppVersionRequest, model, tx)
	} else {
		err = impl.appStoreDeploymentArgoCdService.DeleteInstalledApp(ctx, app.AppName, environment.Name, installAppVersionRequest, model, tx)
	}

	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in commit db transaction on delete", "err", err)
		return nil, err
	}

	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentServiceImpl) LinkHelmApplicationToChartStore(ctx context.Context, request *openapi.UpdateReleaseWithChartLinkingRequest,
	appIdentifier *client.AppIdentifier, userId int32) (*openapi.UpdateReleaseResponse, bool, error) {

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

func (impl AppStoreDeploymentServiceImpl) createEnvironmentIfNotExists(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (int, error) {
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
		Environment: cluster2.BuildEnvironmentIdentifer(cluster.ClusterName, namespace),
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

func (impl AppStoreDeploymentServiceImpl) RollbackApplication(ctx context.Context, request *openapi2.RollbackReleaseRequest,
	installedApp *appStoreBean.InstallAppVersionDTO, userId int32) (bool, error) {
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return false, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// Rollback starts
	var success bool
	if installedApp.AppOfferingMode == util2.SERVER_MODE_HYPERION {
		installedApp, success, err = impl.appStoreDeploymentHelmService.RollbackRelease(ctx, installedApp, request.GetVersion())
		if err != nil {
			impl.logger.Errorw("error while rollback helm release", "error", err)
			return false, err
		}
	} else {
		installedApp, success, err = impl.appStoreDeploymentArgoCdService.RollbackRelease(ctx, installedApp, request.GetVersion())
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

	if installedApp.AppOfferingMode == util2.SERVER_MODE_FULL {
		// create build history for version upgrade, chart upgrade or simple update
		err = impl.UpdateInstallAppVersionHistory(installedApp)
		if err != nil {
			impl.logger.Errorw("error on creating history for chart deployment", "error", err)
			return false, err
		}
	}
	return success, nil
}

func (impl AppStoreDeploymentServiceImpl) linkHelmApplicationToChartStore(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*openapi.UpdateReleaseResponse, error) {
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	//step 1 db operation initiated
	installAppVersionRequest, err = impl.AppStoreDeployOperationDB(installAppVersionRequest, tx)
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
	updateReleaseRequest := &client.InstallReleaseRequest{
		ValuesYaml:   installAppVersionRequest.ValuesOverrideYaml,
		ChartName:    appStoreAppVersion.Name,
		ChartVersion: appStoreAppVersion.Version,
		ReleaseIdentifier: &client.ReleaseIdentifier{
			ReleaseNamespace: installAppVersionRequest.Namespace,
			ReleaseName:      installAppVersionRequest.AppName,
		},
		ChartRepository: &client.ChartRepository{
			Name:     chartRepoInfo.Name,
			Url:      chartRepoInfo.Url,
			Username: chartRepoInfo.UserName,
			Password: chartRepoInfo.Password,
		},
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

func (impl AppStoreDeploymentServiceImpl) installAppPostDbOperation(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	//step 4 db operation status update to deploy success
	_, err := impl.AppStoreDeployOperationStatusUpdate(installAppVersionRequest.InstalledAppId, appStoreBean.DEPLOY_SUCCESS)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return err
	}

	//step 5 create build history first entry for install app version
	if len(installAppVersionRequest.GitHash) > 0 {
		err = impl.UpdateInstallAppVersionHistory(installAppVersionRequest)
		if err != nil {
			impl.logger.Errorw("error on creating history for chart deployment", "error", err)
			return err
		}
	}

	return nil
}

func (impl AppStoreDeploymentServiceImpl) GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*client.DeploymentHistoryAndInstalledAppInfo, error) {
	result := &client.DeploymentHistoryAndInstalledAppInfo{}
	var err error
	if installedApp.AppOfferingMode == util2.SERVER_MODE_HYPERION {
		deploymentHistory, err := impl.appStoreDeploymentHelmService.GetDeploymentHistory(ctx, installedApp)
		if err != nil {
			impl.logger.Errorw("error while getting deployment history", "error", err)
			return nil, err
		}
		result.DeploymentHistory = deploymentHistory.GetDeploymentHistory()
	} else {
		deploymentHistory, err := impl.appStoreDeploymentArgoCdService.GetDeploymentHistory(ctx, installedApp)
		if err != nil {
			impl.logger.Errorw("error while getting deployment history", "error", err)
			return nil, err
		}
		result.DeploymentHistory = deploymentHistory.GetDeploymentHistory()
	}

	if installedApp.InstalledAppId > 0 {
		result.InstalledAppInfo = &client.InstalledAppInfo{
			AppId:                 installedApp.AppId,
			EnvironmentName:       installedApp.EnvironmentName,
			AppOfferingMode:       installedApp.AppOfferingMode,
			InstalledAppId:        installedApp.InstalledAppId,
			InstalledAppVersionId: installedApp.InstalledAppVersionId,
			AppStoreChartId:       installedApp.InstallAppVersionChartDTO.AppStoreChartId,
			ClusterId:             installedApp.ClusterId,
			EnvironmentId:         installedApp.EnvironmentId,
		}
	}

	return result, err
}

func (impl AppStoreDeploymentServiceImpl) GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int) (*openapi.HelmAppDeploymentManifestDetail, error) {
	//var result interface{}
	result := &openapi.HelmAppDeploymentManifestDetail{}
	var err error
	if installedApp.AppOfferingMode == util2.SERVER_MODE_HYPERION {
		result, err = impl.appStoreDeploymentHelmService.GetDeploymentHistoryInfo(ctx, installedApp, int32(version))
		if err != nil {
			impl.logger.Errorw("error while getting deployment history info", "error", err)
			return nil, err
		}
	} else {
		result, err = impl.appStoreDeploymentArgoCdService.GetDeploymentHistoryInfo(ctx, installedApp, int32(version))
		if err != nil {
			impl.logger.Errorw("error while getting deployment history info", "error", err)
			return nil, err
		}
	}
	return result, err
}

func (impl AppStoreDeploymentServiceImpl) UpdateInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error) {
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

	installAppVersionRequest.EnvironmentId = installedApp.EnvironmentId
	installAppVersionRequest.AppName = installedApp.App.AppName
	installAppVersionRequest.EnvironmentName = installedApp.Environment.Name
	environment, err := impl.environmentRepository.FindById(installedApp.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}

	isHelmApp := installedApp.App.AppOfferingMode == util2.SERVER_MODE_HYPERION || installedApp.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_HELM

	// handle gitOps repo name and argoCdAppName for full mode app
	if !isHelmApp {
		gitOpsRepoName := installedApp.GitOpsRepoName
		if len(gitOpsRepoName) == 0 {
			gitOpsRepoName, err = impl.appStoreDeploymentArgoCdService.GetGitOpsRepoName(installAppVersionRequest.AppName, installAppVersionRequest.EnvironmentName)
			if err != nil {
				return nil, err
			}
		}
		installAppVersionRequest.GitOpsRepoName = gitOpsRepoName

		argocdAppName := installedApp.App.AppName + "-" + installedApp.Environment.Name
		installAppVersionRequest.ACDAppName = argocdAppName
	}

	var installedAppVersion *repository.InstalledAppVersions
	if installAppVersionRequest.Id == 0 {
		// upgrade chart to other repo
		_, installedAppVersion, err = impl.upgradeInstalledApp(ctx, installAppVersionRequest, installedApp, tx)
		if err != nil {
			impl.logger.Errorw("error while upgrade the chart", "error", err)
			return nil, err
		}
		installAppVersionRequest.Id = installedAppVersion.Id
	} else {
		// update same chart or upgrade its version only
		installedAppVersionModel, err := impl.installedAppRepository.GetInstalledAppVersion(installAppVersionRequest.Id)
		if err != nil {
			impl.logger.Errorw("error while fetching chart installed version", "error", err)
			return nil, err
		}
		if installedAppVersionModel.AppStoreApplicationVersionId != installAppVersionRequest.AppStoreVersion {
			// upgrade to new version of same chart
			installedAppVersionModel.Active = false
			installedAppVersionModel.UpdatedOn = time.Now()
			installedAppVersionModel.UpdatedBy = installAppVersionRequest.UserId
			_, err = impl.installedAppRepository.UpdateInstalledAppVersion(installedAppVersionModel, tx)
			if err != nil {
				impl.logger.Errorw("error while fetching from db", "error", err)
				return nil, err
			}
			appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
			if err != nil {
				impl.logger.Errorw("fetching error", "err", err)
				return nil, err
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
				return nil, err
			}
			installedAppVersion.AppStoreApplicationVersion = *appStoreAppVersion

			//update requirements yaml in chart for full mode app
			if !isHelmApp {
				err = impl.appStoreDeploymentArgoCdService.UpdateRequirementDependencies(environment, installedAppVersion, installAppVersionRequest, appStoreAppVersion)
				if err != nil {
					impl.logger.Errorw("error while commit required dependencies to git", "error", err)
					return nil, err
				}
			}
			installAppVersionRequest.Id = installedAppVersion.Id
		} else {
			installedAppVersion = installedAppVersionModel
		}

		if isHelmApp {
			installAppVersionRequest, err = impl.appStoreDeploymentHelmService.UpdateInstalledApp(ctx, installAppVersionRequest, environment, installedAppVersion)
		} else {
			installAppVersionRequest, err = impl.appStoreDeploymentArgoCdService.UpdateInstalledApp(ctx, installAppVersionRequest, environment, installedAppVersion)
		}
		if err != nil {
			impl.logger.Errorw("error while updating application", "error", err)
			return nil, err
		}

		//DB operation
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

	//STEP 8: finish with return response
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while committing transaction to db", "error", err)
		return nil, err
	}

	// create build history for version upgrade, chart upgrade or simple update
	err = impl.UpdateInstallAppVersionHistory(installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw("error on creating history for chart deployment", "error", err)
		return nil, err
	}
	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentServiceImpl) GetInstalledAppVersion(id int, userId int32) (*appStoreBean.InstallAppVersionDTO, error) {
	app, err := impl.installedAppRepository.GetInstalledAppVersion(id)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
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
		GitOpsRepoName:     app.InstalledApp.GitOpsRepoName,
		UserId:             userId,
		AppOfferingMode:    app.InstalledApp.App.AppOfferingMode,
		ClusterId:          app.InstalledApp.Environment.ClusterId,
		Namespace:          app.InstalledApp.Environment.Namespace,
		DeploymentAppType:  app.InstalledApp.DeploymentAppType,
	}
	return installAppVersion, err
}

func (impl AppStoreDeploymentServiceImpl) upgradeInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApp *repository.InstalledApps, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, *repository.InstalledAppVersions, error) {
	var installedAppVersion *repository.InstalledAppVersions
	installedAppVersions, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppId(installAppVersionRequest.InstalledAppId)
	if err != nil {
		impl.logger.Errorw("error while fetching installed version", "error", err)
		return installAppVersionRequest, installedAppVersion, err
	}
	for _, installedAppVersionModel := range installedAppVersions {
		installedAppVersionModel.Active = false
		installedAppVersionModel.UpdatedOn = time.Now()
		installedAppVersionModel.UpdatedBy = installAppVersionRequest.UserId
		_, err = impl.installedAppRepository.UpdateInstalledAppVersion(installedAppVersionModel, tx)
		if err != nil {
			impl.logger.Errorw("error while update installed chart", "error", err)
			return installAppVersionRequest, installedAppVersion, err
		}
	}

	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return installAppVersionRequest, installedAppVersion, err
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
		return installAppVersionRequest, installedAppVersion, err
	}
	installedAppVersion.AppStoreApplicationVersion = *appStoreAppVersion
	installAppVersionRequest.InstalledAppVersionId = installedAppVersion.Id

	if installedApp.App.AppOfferingMode == util2.SERVER_MODE_HYPERION || installedApp.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_HELM {
		// update in helm
		installAppVersionRequest, err = impl.appStoreDeploymentHelmService.OnUpdateRepoInInstalledApp(ctx, installAppVersionRequest)
	} else {
		//step 2 git operation pull push
		//step 3 acd operation register, sync
		installAppVersionRequest, err = impl.appStoreDeploymentArgoCdService.OnUpdateRepoInInstalledApp(ctx, installAppVersionRequest)
	}
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return installAppVersionRequest, installedAppVersion, err
	}

	//step 4 db operation status triggered
	installedApp.Status = appStoreBean.DEPLOY_SUCCESS
	_, err = impl.installedAppRepository.UpdateInstalledApp(installedApp, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return installAppVersionRequest, installedAppVersion, err
	}

	return installAppVersionRequest, installedAppVersion, err
}
func (impl AppStoreDeploymentServiceImpl) InstallAppByHelm(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error) {
	installAppVersionRequest, err := impl.appStoreDeploymentHelmService.InstallApp(installAppVersionRequest, ctx)
	if err != nil {
		impl.logger.Errorw("error while installing app via helm", "error", err)
		return installAppVersionRequest, err
	}
	return installAppVersionRequest, nil
}
