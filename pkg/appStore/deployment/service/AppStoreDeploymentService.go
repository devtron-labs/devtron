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
	"github.com/caarlos0/env/v6"
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/internal/constants"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDeploymentTool "github.com/devtron-labs/devtron/pkg/appStore/deployment/tool"
	appStoreDeploymentGitopsTool "github.com/devtron-labs/devtron/pkg/appStore/deployment/tool/gitops"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
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
	AppStoreDeployOperationDB(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx, skipAppCreation bool) (*appStoreBean.InstallAppVersionDTO, error)
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
	UpdateInstalledAppVersionHistoryWithGitHash(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error
	GetInstalledAppVersion(id int, userId int32) (*appStoreBean.InstallAppVersionDTO, error)
	InstallAppByHelm(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error)
	UpdateProjectHelmApp(updateAppRequest *appStoreBean.UpdateProjectHelmAppDTO) error
	UpdateNotesForInstalledApp(installAppId int, notes string) (bool, error)
	UpdatePreviousDeploymentStatusForAppStore(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, triggeredAt time.Time, err error) error
}

type DeploymentServiceTypeConfig struct {
	IsInternalUse bool `env:"IS_INTERNAL_USE" envDefault:"false"`
}

func GetDeploymentServiceTypeConfig() (*DeploymentServiceTypeConfig, error) {
	cfg := &DeploymentServiceTypeConfig{}
	err := env.Parse(cfg)
	return cfg, err
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
	deploymentTypeConfig                 *DeploymentServiceTypeConfig
	ChartTemplateService                 util.ChartTemplateService
}

func NewAppStoreDeploymentServiceImpl(logger *zap.SugaredLogger, installedAppRepository repository.InstalledAppRepository,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository, environmentRepository clusterRepository.EnvironmentRepository,
	clusterInstalledAppsRepository repository.ClusterInstalledAppsRepository, appRepository app.AppRepository,
	appStoreDeploymentHelmService appStoreDeploymentTool.AppStoreDeploymentHelmService,
	appStoreDeploymentArgoCdService appStoreDeploymentGitopsTool.AppStoreDeploymentArgoCdService, environmentService cluster.EnvironmentService,
	clusterService cluster.ClusterService, helmAppService client.HelmAppService, appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	globalEnvVariables *util2.GlobalEnvVariables,
	installedAppRepositoryHistory repository.InstalledAppVersionHistoryRepository, gitOpsRepository repository2.GitOpsConfigRepository, attributesService attributes.AttributesService,
	deploymentTypeConfig *DeploymentServiceTypeConfig, ChartTemplateService util.ChartTemplateService) *AppStoreDeploymentServiceImpl {
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
		deploymentTypeConfig:                 deploymentTypeConfig,
		ChartTemplateService:                 ChartTemplateService,
	}
}

func (impl AppStoreDeploymentServiceImpl) AppStoreDeployOperationDB(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx, skipAppCreation bool) (*appStoreBean.InstallAppVersionDTO, error) {

	var isInternalUse = impl.deploymentTypeConfig.IsInternalUse

	isGitOpsConfigured := false
	gitOpsConfig, err := impl.gitOpsRepository.GetGitOpsConfigActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GetGitOpsConfigActive, error while getting", "err", err)
		return nil, err
	}
	if gitOpsConfig != nil && gitOpsConfig.Id > 0 {
		isGitOpsConfigured = true
	}

	if isInternalUse && !isGitOpsConfigured && installAppVersionRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD {
		impl.logger.Errorw("gitops not configured but selected for CD")
		err := &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: "Gitops integration is not installed/configured. Please install/configure gitops or use helm option.",
			UserMessage:     "Gitops integration is not installed/configured. Please install/configure gitops or use helm option.",
		}
		return nil, err
	}

	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}

	var appInstallationMode string
	if util2.IsBaseStack() || util2.IsHelmApp(installAppVersionRequest.AppOfferingMode) {
		appInstallationMode = util2.SERVER_MODE_HYPERION
	} else {
		appInstallationMode = util2.SERVER_MODE_FULL
	}

	// create env if env not exists for clusterId and namespace for hyperion mode
	if util2.IsHelmApp(appInstallationMode) {
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
	installAppVersionRequest.Environment = environment
	installAppVersionRequest.ACDAppName = fmt.Sprintf("%s-%s", installAppVersionRequest.AppName, installAppVersionRequest.Environment.Name)
	installAppVersionRequest.ClusterId = environment.ClusterId
	appCreateRequest := &bean.CreateAppDTO{
		Id:      installAppVersionRequest.AppId,
		AppName: installAppVersionRequest.AppName,
		TeamId:  installAppVersionRequest.TeamId,
		UserId:  installAppVersionRequest.UserId,
	}

	appCreateRequest, err = impl.createAppForAppStore(appCreateRequest, tx, appInstallationMode, skipAppCreation)
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

	//if isGitOpsConfigured && appInstallationMode == util2.SERVER_MODE_FULL && (installAppVersionRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD || !isInternalUse) {
	//	installedAppModel.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
	//} else {
	//	installedAppModel.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
	//}
	//
	//if isInternalUse && installedAppModel.DeploymentAppType == "" {
	//	if isGitOpsConfigured {
	//		installedAppModel.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
	//	} else {
	//		installedAppModel.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
	//	}
	//}

	if installAppVersionRequest.DeploymentAppType == "" {
		if isGitOpsConfigured {
			installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
		} else {
			installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
		}
	}

	installedAppModel.DeploymentAppType = installAppVersionRequest.DeploymentAppType
	installedAppModel.CreatedBy = installAppVersionRequest.UserId
	installedAppModel.UpdatedBy = installAppVersionRequest.UserId
	installedAppModel.CreatedOn = time.Now()
	installedAppModel.UpdatedOn = time.Now()
	installedAppModel.Active = true
	if util2.IsFullStack() {
		installedAppModel.GitOpsRepoName = impl.ChartTemplateService.GetGitOpsRepoName(installAppVersionRequest.AppName)
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

	if installAppVersionRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD || installAppVersionRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD {
		installedAppVersionHistory := &repository.InstalledAppVersionHistory{}
		installedAppVersionHistory.InstalledAppVersionId = installedAppVersions.Id
		installedAppVersionHistory.ValuesYamlRaw = installAppVersionRequest.ValuesOverrideYaml
		installedAppVersionHistory.CreatedBy = installAppVersionRequest.UserId
		installedAppVersionHistory.CreatedOn = time.Now()
		installedAppVersionHistory.UpdatedBy = installAppVersionRequest.UserId
		installedAppVersionHistory.UpdatedOn = time.Now()
		installedAppVersionHistory.StartedOn = time.Now()
		installedAppVersionHistory.Status = pipelineConfig.WorkflowInProgress
		_, err = impl.installedAppRepositoryHistory.CreateInstalledAppVersionHistory(installedAppVersionHistory, tx)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		installAppVersionRequest.InstalledAppVersionHistoryId = installedAppVersionHistory.Id
	}

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

func (impl AppStoreDeploymentServiceImpl) UpdateNotesForInstalledApp(installAppId int, notes string) (bool, error) {
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
	installedApp.Notes = notes
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
	installAppVersionRequest, err = impl.AppStoreDeployOperationDB(installAppVersionRequest, tx, false)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return nil, err
	}
	impl.updateDeploymentParametersInRequest(installAppVersionRequest, installAppVersionRequest.DeploymentAppType)

	if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) || util.IsManifestDownload(installAppVersionRequest.DeploymentAppType) {
		_ = impl.appStoreDeploymentArgoCdService.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED, "Deployment initiated successfully.", tx)
	}

	manifest, err := impl.appStoreDeploymentCommonService.GenerateManifest(installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw("error in performing manifest and git operations", "err", err)
		return nil, err
	}

	if installAppVersionRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD {
		_ = impl.appStoreDeploymentArgoCdService.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_DESCRIPTION_MANIFEST_GENERATED, "Manifest generated successfully.", tx)
	}

	var gitOpsResponse *appStoreDeploymentCommon.AppStoreGitOpsResponse
	if installAppVersionRequest.PerformGitOps {
		gitOpsResponse, err = impl.appStoreDeploymentCommonService.GitOpsOperations(manifest, installAppVersionRequest)
		if err != nil {
			impl.logger.Errorw("error in doing gitops operation", "err", err)
			if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
				_ = impl.appStoreDeploymentArgoCdService.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED, fmt.Sprintf("Git commit failed - %v", err), tx)
			}
			return nil, err
		}
		if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
			_ = impl.appStoreDeploymentArgoCdService.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT, "Git commit done successfully.", tx)
		}
		installAppVersionRequest.GitHash = gitOpsResponse.GitHash
	}

	if util2.IsBaseStack() || util2.IsHelmApp(installAppVersionRequest.AppOfferingMode) || util.IsHelmApp(installAppVersionRequest.DeploymentAppType) {
		installAppVersionRequest, err = impl.appStoreDeploymentHelmService.InstallApp(installAppVersionRequest, nil, ctx, tx)
	} else if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
		if gitOpsResponse == nil && gitOpsResponse.ChartGitAttribute != nil {
			return nil, errors.New("service err, Error in git operations")
		}
		installAppVersionRequest, err = impl.appStoreDeploymentArgoCdService.InstallApp(installAppVersionRequest, gitOpsResponse.ChartGitAttribute, ctx, tx)
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

//func (impl AppStoreDeploymentServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error) {
//
//	dbConnection := impl.installedAppRepository.GetConnection()
//	tx, err := dbConnection.Begin()
//	if err != nil {
//		return nil, err
//	}
//	// Rollback tx on error.
//	defer tx.Rollback()
//
//	//step 1 db operation initiated
//	installAppVersionRequest, err = impl.AppStoreDeployOperationDB(installAppVersionRequest, tx, false)
//	if err != nil {
//		impl.logger.Errorw(" error", "err", err)
//		return nil, err
//	}
//
//	if util2.IsBaseStack() || util2.IsHelmApp(installAppVersionRequest.AppOfferingMode) || util.IsHelmApp(installAppVersionRequest.DeploymentAppType) {
//		installAppVersionRequest, err = impl.appStoreDeploymentHelmService.InstallApp(installAppVersionRequest, ctx)
//	} else {
//		installAppVersionRequest, err = impl.appStoreDeploymentArgoCdService.InstallApp(installAppVersionRequest, ctx)
//	}
//
//	if err != nil {
//		return nil, err
//	}
//
//	// tx commit here because next operation will be process after this commit.
//	err = tx.Commit()
//	if err != nil {
//		return nil, err
//	}
//
//	err = impl.installAppPostDbOperation(installAppVersionRequest)
//	if err != nil {
//		return nil, err
//	}
//
//	return installAppVersionRequest, nil
//}

func (impl AppStoreDeploymentServiceImpl) UpdateInstalledAppVersionHistoryStatus(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, status string) error {
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
func (impl AppStoreDeploymentServiceImpl) UpdateInstalledAppVersionHistoryWithGitHash(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	savedInstalledAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(installAppVersionRequest.InstalledAppVersionHistoryId)
	savedInstalledAppVersionHistory.GitHash = installAppVersionRequest.GitHash
	savedInstalledAppVersionHistory.UpdatedOn = time.Now()
	savedInstalledAppVersionHistory.UpdatedBy = installAppVersionRequest.UserId

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

func (impl AppStoreDeploymentServiceImpl) createAppForAppStore(createRequest *bean.CreateAppDTO, tx *pg.Tx, appInstallationMode string, skipAppCreation bool) (*bean.CreateAppDTO, error) {
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

func (impl AppStoreDeploymentServiceImpl) GetInstalledApp(id int) (*appStoreBean.InstallAppVersionDTO, error) {
	app, err := impl.installedAppRepository.GetInstalledApp(id)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	chartTemplate := impl.chartAdaptor2(app)
	return chartTemplate, nil
}

// converts db object to bean
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
		if util2.IsBaseStack() || util2.IsHelmApp(a.AppOfferingMode) || util.IsHelmApp(a.DeploymentAppType) {
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

func (impl AppStoreDeploymentServiceImpl) DeleteInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error) {
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
	app, err := impl.appRepository.FindById(installAppVersionRequest.AppId)
	if err != nil {
		return nil, err
	}
	model, err := impl.installedAppRepository.GetInstalledApp(installAppVersionRequest.InstalledAppId)
	if err != nil {
		impl.logger.Errorw("error in fetching installed app", "id", installAppVersionRequest.InstalledAppId, "err", err)
		return nil, err
	}

	if installAppVersionRequest.AcdPartialDelete == true {
		if util2.IsBaseStack() || util2.IsHelmApp(app.AppOfferingMode) || util.IsHelmApp(model.DeploymentAppType) {
			err = impl.appStoreDeploymentHelmService.DeleteDeploymentApp(ctx, app.AppName, environment.Name, installAppVersionRequest)
		} else {
			err = impl.appStoreDeploymentArgoCdService.DeleteDeploymentApp(ctx, app.AppName, environment.Name, installAppVersionRequest)
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

		if util2.IsBaseStack() || util2.IsHelmApp(app.AppOfferingMode) || util.IsHelmApp(model.DeploymentAppType) {
			// there might be a case if helm release gets uninstalled from helm cli.
			//in this case on deleting the app from API, it should not give error as it should get deleted from db, otherwise due to delete error, db does not get clean
			// so in helm, we need to check first if the release exists or not, if exists then only delete
			err = impl.appStoreDeploymentHelmService.DeleteInstalledApp(ctx, app.AppName, environment.Name, installAppVersionRequest, model, tx)
		} else {
			err = impl.appStoreDeploymentArgoCdService.DeleteInstalledApp(ctx, app.AppName, environment.Name, installAppVersionRequest, model, tx)
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
		installedApp, success, err = impl.appStoreDeploymentHelmService.RollbackRelease(ctx, installedApp, request.GetVersion(), tx)
		if err != nil {
			impl.logger.Errorw("error while rollback helm release", "error", err)
			return false, err
		}
	} else {
		installedApp, success, err = impl.appStoreDeploymentArgoCdService.RollbackRelease(ctx, installedApp, request.GetVersion(), tx)
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
		err = impl.UpdateInstalledAppVersionHistoryWithGitHash(installedApp)
		if err != nil {
			impl.logger.Errorw("error on creating history for chart deployment", "error", err)
			return false, err
		}
	}
	err1 := impl.UpdatePreviousDeploymentStatusForAppStore(installedApp, triggeredAt, err)
	if err1 != nil {
		impl.logger.Errorw("error while update previous installed app version history", "err", err, "installAppVersionRequest", installedApp)
		//if installed app is updated and error is in updating previous deployment status, then don't block user, just show error.
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

	// skipAppCreation flag is set for CLI apps because for CLI Helm apps if project is created first before linking to chart store then app is created during project update time.
	// skipAppCreation - This flag will skip app creation if app already exists.

	//step 1 db operation initiated
	installAppVersionRequest, err = impl.AppStoreDeployOperationDB(installAppVersionRequest, tx, true)
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
	updateReleaseRequest := &client.UpdateApplicationWithChartInfoRequestDto{
		InstallReleaseRequest: &client.InstallReleaseRequest{
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
		},
		SourceAppType: client.SOURCE_HELM_APP,
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

	//step 5 create build history first entry for install app version for argocd or helm type deployments
	if len(installAppVersionRequest.GitHash) > 0 {
		err = impl.UpdateInstalledAppVersionHistoryWithGitHash(installAppVersionRequest)
		if err != nil {
			impl.logger.Errorw("error in installAppPostDbOperation", "err", err)
			return err
		}
	}

	if installAppVersionRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD {
		err = impl.UpdateInstalledAppVersionHistoryStatus(installAppVersionRequest, pipelineConfig.WorkflowSucceeded)
		if err != nil {
			impl.logger.Errorw("error on creating history for chart deployment", "error", err)
			return err
		}
	}

	if installAppVersionRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_HELM {
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
	if util2.IsHelmApp(installedApp.AppOfferingMode) {
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
			DeploymentType:        installedApp.DeploymentAppType,
		}
	}

	return result, err
}

func (impl AppStoreDeploymentServiceImpl) GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int) (*openapi.HelmAppDeploymentManifestDetail, error) {
	//var result interface{}
	result := &openapi.HelmAppDeploymentManifestDetail{}
	var err error
	if util2.IsHelmApp(installedApp.AppOfferingMode) || installedApp.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_HELM {
		_, span := otel.Tracer("orchestrator").Start(ctx, "appStoreDeploymentHelmService.GetDeploymentHistoryInfo")
		result, err = impl.appStoreDeploymentHelmService.GetDeploymentHistoryInfo(ctx, installedApp, int32(version))
		span.End()
		if err != nil {
			impl.logger.Errorw("error while getting deployment history info", "error", err)
			return nil, err
		}
	} else {
		_, span := otel.Tracer("orchestrator").Start(ctx, "appStoreDeploymentArgoCdService.GetDeploymentHistoryInfo")
		result, err = impl.appStoreDeploymentArgoCdService.GetDeploymentHistoryInfo(ctx, installedApp, int32(version))
		span.End()
		if err != nil {
			impl.logger.Errorw("error while getting deployment history info", "error", err)
			return nil, err
		}
	}
	return result, err
}

func (impl AppStoreDeploymentServiceImpl) updateInstalledAppVersion(installedAppVersion *repository.InstalledAppVersions, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx, installedApp *repository.InstalledApps) (*repository.InstalledAppVersions, *appStoreBean.InstallAppVersionDTO, error) {
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
func (impl *AppStoreDeploymentServiceImpl) CheckIfMonoRepoMigrationRequired(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApp *repository.InstalledApps) bool {

	monoRepoMigrationRequired := false
	var err error
	gitOpsRepoName := installedApp.GitOpsRepoName
	if len(gitOpsRepoName) == 0 {
		if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
			gitOpsRepoName, err = impl.appStoreDeploymentArgoCdService.GetGitOpsRepoName(installAppVersionRequest.AppName, installAppVersionRequest.EnvironmentName)
			if err != nil {
				return false
			}
		} else {
			gitOpsRepoName = impl.ChartTemplateService.GetGitOpsRepoName(installAppVersionRequest.AppName)
		}
	}
	//here will set new git repo name if required to migrate
	newGitOpsRepoName := impl.ChartTemplateService.GetGitOpsRepoName(installedApp.App.AppName)
	//checking weather git repo migration needed or not, if existing git repo and new independent git repo is not same than go ahead with migration
	if newGitOpsRepoName != gitOpsRepoName {
		monoRepoMigrationRequired = true
	}
	return monoRepoMigrationRequired
}

func (impl *AppStoreDeploymentServiceImpl) updateDeploymentParametersInRequest(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, deploymentAppType string) {

	installAppVersionRequest.DeploymentAppType = deploymentAppType

	switch installAppVersionRequest.DeploymentAppType {
	case util.PIPELINE_DEPLOYMENT_TYPE_ACD:
		installAppVersionRequest.PerformGitOps = true
		installAppVersionRequest.PerformACDDeployment = true
		installAppVersionRequest.PerformHelmDeployment = false
	case util.PIPELINE_DEPLOYMENT_TYPE_HELM:
		installAppVersionRequest.PerformGitOps = false
		installAppVersionRequest.PerformACDDeployment = false
		installAppVersionRequest.PerformHelmDeployment = true
	case util.PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD:
		installAppVersionRequest.PerformGitOps = false
		installAppVersionRequest.PerformHelmDeployment = false
		installAppVersionRequest.PerformACDDeployment = false
	}

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

	impl.updateDeploymentParametersInRequest(installAppVersionRequest, installedApp.DeploymentAppType)

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
			return nil, err
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
	} else {
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

	installAppVersionHistoryStatus := "Unknown"
	if util.IsAcdApp(installedApp.DeploymentAppType) || util.IsManifestDownload(installedApp.DeploymentAppType) {
		installAppVersionHistoryStatus = pipelineConfig.WorkflowInProgress
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
		_ = impl.appStoreDeploymentArgoCdService.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED, "Deployment initiated successfully.", tx)
	}

	manifest, err := impl.appStoreDeploymentCommonService.GenerateManifest(installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw("error in generating manifest for helm apps", "err", err)
		_ = impl.UpdateInstalledAppVersionHistoryStatus(installAppVersionRequest, pipelineConfig.WorkflowFailed)
		return nil, err
	}
	if installAppVersionRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD {
		_ = impl.appStoreDeploymentArgoCdService.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_DESCRIPTION_MANIFEST_GENERATED, "Manifest generated successfully.", tx)
	}
	// gitOps operation
	monoRepoMigrationRequired := false
	gitOpsResponse := &appStoreDeploymentCommon.AppStoreGitOpsResponse{}

	if installAppVersionRequest.PerformGitOps {

		// required if gitOps repo name is changed, gitOps repo name will change if env variable which we use as suffix changes
		gitOpsRepoName := installedApp.GitOpsRepoName
		if len(gitOpsRepoName) == 0 {
			if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
				gitOpsRepoName, err = impl.appStoreDeploymentArgoCdService.GetGitOpsRepoName(installAppVersionRequest.AppName, installAppVersionRequest.EnvironmentName)
				if err != nil {
					return nil, err
				}
			} else {
				gitOpsRepoName = impl.ChartTemplateService.GetGitOpsRepoName(installAppVersionRequest.AppName)
			}
		}
		//here will set new git repo name if required to migrate
		newGitOpsRepoName := impl.ChartTemplateService.GetGitOpsRepoName(installedApp.App.AppName)
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

		var createRepoErr, requirementsCommitErr, valuesCommitErr error
		var gitHash string

		if monoRepoMigrationRequired {
			// create new git repo if repo name changed
			gitOpsResponse, createRepoErr = impl.appStoreDeploymentCommonService.GitOpsOperations(manifest, installAppVersionRequest)
			gitHash = gitOpsResponse.GitHash

		} else if isChartChanged || isVersionChanged {
			// update dependency if chart or chart version is changed
			_, requirementsCommitErr = impl.appStoreDeploymentCommonService.CommitConfigToGit(manifest.RequirementsConfig)
			gitHash, valuesCommitErr = impl.appStoreDeploymentCommonService.CommitConfigToGit(manifest.ValuesConfig)

		} else {
			// only values are changed in update, so commit values config
			gitHash, valuesCommitErr = impl.appStoreDeploymentCommonService.CommitConfigToGit(manifest.ValuesConfig)
		}

		if valuesCommitErr != nil || requirementsCommitErr != nil {

			noTargetFoundForValues, _ := impl.appStoreDeploymentCommonService.ParseGitRepoErrorResponse(valuesCommitErr)
			noTargetFoundForRequirements, _ := impl.appStoreDeploymentCommonService.ParseGitRepoErrorResponse(requirementsCommitErr)

			if noTargetFoundForRequirements || noTargetFoundForValues {
				//create repo again and try again  -  auto fix
				monoRepoMigrationRequired = true // since repo is created again, will use this flag to check if ACD patch operation required
				gitOpsResponse, createRepoErr = impl.appStoreDeploymentCommonService.GitOpsOperations(manifest, installAppVersionRequest)
				gitHash = gitOpsResponse.GitHash
			}

		}

		if createRepoErr != nil || requirementsCommitErr != nil || valuesCommitErr != nil {
			impl.logger.Errorw("error in doing gitops operation", "err", err)
			_ = impl.appStoreDeploymentArgoCdService.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED, fmt.Sprintf("Git commit failed - %v", err), tx)
			return nil, err
		}

		installAppVersionRequest.GitHash = gitHash
		_ = impl.appStoreDeploymentArgoCdService.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT, "Git commit done successfully.", tx)

	}

	if installAppVersionRequest.PerformACDDeployment {
		if monoRepoMigrationRequired {
			// update repo details on ArgoCD as repo is changed
			err = impl.appStoreDeploymentArgoCdService.UpdateChartInfo(installAppVersionRequest, gitOpsResponse.ChartGitAttribute, ctx)
			if err != nil {
				impl.logger.Errorw("error in acd patch request", "err", err)
				return nil, err
			}
		}
	} else if installAppVersionRequest.PerformHelmDeployment {
		err = impl.appStoreDeploymentHelmService.UpdateChartInfo(installAppVersionRequest, gitOpsResponse.ChartGitAttribute, ctx)
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

	if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
		err = impl.UpdateInstalledAppVersionHistoryWithGitHash(installAppVersionRequest)
		if err != nil {
			impl.logger.Errorw("error on updating history for chart deployment", "error", err, "installedAppVersion", installedAppVersion)
			return nil, err
		}
	} else if util.IsManifestDownload(installAppVersionRequest.DeploymentAppType) {
		err = impl.UpdateInstalledAppVersionHistoryStatus(installAppVersionRequest, pipelineConfig.WorkflowSucceeded)
		if err != nil {
			impl.logger.Errorw("error on creating history for chart deployment", "error", err)
			return nil, err
		}
	} else {
		// create build history for chart on default component deployed via helm
		err = impl.UpdateInstallAppVersionHistory(installAppVersionRequest)
		if err != nil {
			impl.logger.Errorw("error on creating history for chart deployment", "error", err, "installedAppVersion", installedAppVersion)
			return nil, err
		}
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
		Environment:        &app.InstalledApp.Environment,
		ACDAppName:         fmt.Sprintf("%s-%s", app.InstalledApp.App.AppName, app.InstalledApp.Environment.Name),
	}
	return installAppVersion, err
}

func (impl AppStoreDeploymentServiceImpl) InstallAppByHelm(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error) {
	installAppVersionRequest, err := impl.appStoreDeploymentHelmService.InstallApp(installAppVersionRequest, nil, ctx, nil)
	if err != nil {
		impl.logger.Errorw("error while installing app via helm", "error", err)
		return installAppVersionRequest, err
	}
	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentServiceImpl) UpdateProjectHelmApp(updateAppRequest *appStoreBean.UpdateProjectHelmAppDTO) error {

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

func (impl AppStoreDeploymentServiceImpl) UpdatePreviousDeploymentStatusForAppStore(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, triggeredAt time.Time, err error) error {
	//creating pipeline status timeline for deployment failed
	err1 := impl.appStoreDeploymentArgoCdService.UpdateInstalledAppAndPipelineStatusForFailedDeploymentStatus(installAppVersionRequest, triggeredAt, err)
	if err1 != nil {
		impl.logger.Errorw("error in updating previous deployment status for appStore", "err", err1, "installAppVersionRequest", installAppVersionRequest)
		return err1
	}
	return nil
}
