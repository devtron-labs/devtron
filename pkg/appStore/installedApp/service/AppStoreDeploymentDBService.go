package service

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/adapter"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	discoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/bean"
	clusterService "github.com/devtron-labs/devtron/pkg/cluster"
	clutserBean "github.com/devtron-labs/devtron/pkg/cluster/repository/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/providerConfig"
	globalUtil "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type AppStoreDeploymentDBService interface {
	// AppStoreDeployOperationDB is used to perform Pre-Install DB operations in App Store deployments
	AppStoreDeployOperationDB(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	// AppStoreDeployOperationStatusUpdate updates the bulk deployment status in repository.InstalledApps
	AppStoreDeployOperationStatusUpdate(installAppId int, status appStoreBean.AppstoreDeploymentStatus) (bool, error)
	// IsChartProviderActive validates if the chart provider for Helm App is active
	IsChartProviderActive(appStoreVersionId int) (bool, error)
	// GetInstalledApp returns - appStoreBean.InstallAppVersionDTO for the given InstalledAppId
	GetInstalledApp(id int) (*appStoreBean.InstallAppVersionDTO, error)
	// GetAllInstalledAppsByAppStoreId returns - []appStoreBean.InstalledAppsResponse for the given AppStoreId
	GetAllInstalledAppsByAppStoreId(appStoreId int) ([]appStoreBean.InstalledAppsResponse, error)
	// UpdateInstalledAppVersionHistoryWithGitHash updates GitHash in the repository.InstalledAppVersionHistory
	UpdateInstalledAppVersionHistoryWithGitHash(versionHistoryId int, gitHash string, userId int32) error
	// UpdateProjectForHelmApp updates TeamId in the app.App
	UpdateProjectForHelmApp(appName string, teamId int, userId int32) error
	// InstallAppPostDbOperation is used to perform Post-Install DB operations in App Store deployments
	InstallAppPostDbOperation(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error
	// MarkInstalledAppVersionsInactiveByInstalledAppId will mark the repository.InstalledAppVersions inactive for the given InstalledAppId
	MarkInstalledAppVersionsInactiveByInstalledAppId(installedAppId int, UserId int32, tx *pg.Tx) error
	// MarkInstalledAppVersionModelInActive will mark the given repository.InstalledAppVersions inactive
	MarkInstalledAppVersionModelInActive(installedAppVersionModel *repository.InstalledAppVersions, UserId int32, tx *pg.Tx) error
	// MarkInstalledAppVersionHistorySucceeded will mark the repository.InstalledAppVersionHistory Status - Succeeded
	MarkInstalledAppVersionHistorySucceeded(versionHistoryId int, deploymentAppType string) error
	// UpdateInstalledAppVersionHistoryStatus will update the Status in the repository.InstalledAppVersionHistory
	UpdateInstalledAppVersionHistoryStatus(versionHistoryId int, status string) error
}

type AppStoreDeploymentDBServiceImpl struct {
	logger                               *zap.SugaredLogger
	installedAppRepository               repository.InstalledAppRepository
	appStoreApplicationVersionRepository discoverRepository.AppStoreApplicationVersionRepository
	appRepository                        app.AppRepository
	environmentService                   clusterService.EnvironmentService
	clusterService                       clusterService.ClusterService
	installedAppRepositoryHistory        repository.InstalledAppVersionHistoryRepository
	deploymentTypeConfig                 *globalUtil.DeploymentServiceTypeConfig
	gitOpsConfigReadService              config.GitOpsConfigReadService
	deploymentTypeOverrideService        providerConfig.DeploymentTypeOverrideService
}

func NewAppStoreDeploymentDBServiceImpl(logger *zap.SugaredLogger,
	installedAppRepository repository.InstalledAppRepository,
	appStoreApplicationVersionRepository discoverRepository.AppStoreApplicationVersionRepository,
	appRepository app.AppRepository,
	environmentService clusterService.EnvironmentService,
	clusterService clusterService.ClusterService,
	installedAppRepositoryHistory repository.InstalledAppVersionHistoryRepository,
	envVariables *globalUtil.EnvironmentVariables,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	deploymentTypeOverrideService providerConfig.DeploymentTypeOverrideService) *AppStoreDeploymentDBServiceImpl {
	return &AppStoreDeploymentDBServiceImpl{
		logger:                               logger,
		installedAppRepository:               installedAppRepository,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		appRepository:                        appRepository,
		environmentService:                   environmentService,
		clusterService:                       clusterService,
		installedAppRepositoryHistory:        installedAppRepositoryHistory,
		deploymentTypeConfig:                 envVariables.DeploymentServiceTypeConfig,
		gitOpsConfigReadService:              gitOpsConfigReadService,
		deploymentTypeOverrideService:        deploymentTypeOverrideService,
	}
}

func (impl *AppStoreDeploymentDBServiceImpl) AppStoreDeployOperationDB(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
	environment, err := impl.getEnvironmentForInstallAppRequest(installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw("error in getting environment for install helm chart", "envId", installAppVersionRequest.EnvironmentId, "err", err)
		return nil, err
	}
	// setting additional env data required in appStoreBean.InstallAppVersionDTO
	adapter.UpdateAdditionalEnvDetails(installAppVersionRequest, environment)

	// Stage 1:  Create App in tx (Only if AppId is not set already)
	if installAppVersionRequest.AppId == 0 {
		appCreateRequest := &bean.CreateAppDTO{
			AppName: installAppVersionRequest.AppName,
			TeamId:  installAppVersionRequest.TeamId,
			UserId:  installAppVersionRequest.UserId,
		}
		appCreateRequest, err = impl.createAppForAppStore(appCreateRequest, tx, getAppInstallationMode(installAppVersionRequest.AppOfferingMode))
		if err != nil {
			impl.logger.Errorw("error while creating app", "error", err)
			return nil, err
		}
		installAppVersionRequest.AppId = appCreateRequest.Id
	}
	// Stage 1: ends

	// Stage 2:  validate deployment app type and override if ExternallyManagedDeploymentType
	overrideDeploymentType, err := impl.validateAndGetOverrideDeploymentAppType(installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw("error in validating deployment app type", "error", err)
		return nil, err
	}
	installAppVersionRequest.UpdateDeploymentAppType(overrideDeploymentType)
	// Stage 2: ends

	// Stage 3: save installed_apps model
	if globalUtil.IsFullStack() && util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
		installAppVersionRequest.GitOpsRepoName = impl.gitOpsConfigReadService.GetGitOpsRepoName(installAppVersionRequest.AppName)
	}
	installedAppModel := adapter.NewInstallAppModel(installAppVersionRequest, appStoreBean.DEPLOY_INIT)
	installedApp, err := impl.installedAppRepository.CreateInstalledApp(installedAppModel, tx)
	if err != nil {
		impl.logger.Errorw("error while creating install app", "error", err)
		return nil, err
	}
	installAppVersionRequest.InstalledAppId = installedApp.Id
	// Stage 3: ends

	// Stage 4: save installed_app_versions model
	installedAppVersions := adapter.NewInstallAppVersionsModel(installAppVersionRequest)
	_, err = impl.installedAppRepository.CreateInstalledAppVersion(installedAppVersions, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	installAppVersionRequest.InstalledAppVersionId = installedAppVersions.Id
	installAppVersionRequest.Id = installedAppVersions.Id
	// Stage 4: ends

	// Stage 5: save installed_app_version_history model
	helmInstallConfigDTO := appStoreBean.HelmReleaseStatusConfig{
		InstallAppVersionHistoryId: 0,
		Message:                    "Install initiated",
		IsReleaseInstalled:         false,
		ErrorInInstallation:        false,
	}
	installedAppVersionHistory, err := adapter.NewInstallAppVersionHistoryModel(installAppVersionRequest, pipelineConfig.WorkflowInProgress, helmInstallConfigDTO)
	if err != nil {
		impl.logger.Errorw("error in helm install config marshal", "err", err)
	}
	_, err = impl.installedAppRepositoryHistory.CreateInstalledAppVersionHistory(installedAppVersionHistory, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	installAppVersionRequest.InstalledAppVersionHistoryId = installedAppVersionHistory.Id
	// Stage 5: ends
	return installAppVersionRequest, nil
}

func (impl *AppStoreDeploymentDBServiceImpl) AppStoreDeployOperationStatusUpdate(installAppId int, status appStoreBean.AppstoreDeploymentStatus) (bool, error) {
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

func (impl *AppStoreDeploymentDBServiceImpl) IsChartProviderActive(appStoreVersionId int) (bool, error) {
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

func (impl *AppStoreDeploymentDBServiceImpl) GetInstalledApp(id int) (*appStoreBean.InstallAppVersionDTO, error) {
	app, err := impl.installedAppRepository.GetInstalledApp(id)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	chartTemplate := adapter.GenerateInstallAppVersionMinDTO(app)
	return chartTemplate, nil
}

func (impl *AppStoreDeploymentDBServiceImpl) GetAllInstalledAppsByAppStoreId(appStoreId int) ([]appStoreBean.InstalledAppsResponse, error) {
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
		if globalUtil.IsHelmApp(a.AppOfferingMode) {
			environment, err := impl.environmentService.FindById(a.EnvironmentId)
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

func (impl *AppStoreDeploymentDBServiceImpl) UpdateInstalledAppVersionHistoryWithGitHash(versionHistoryId int, gitHash string, userId int32) error {
	savedInstalledAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(versionHistoryId)
	savedInstalledAppVersionHistory.GitHash = gitHash
	savedInstalledAppVersionHistory.UpdateAuditLog(userId)
	_, err = impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistory(savedInstalledAppVersionHistory, nil)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return err
	}
	return nil
}

func (impl *AppStoreDeploymentDBServiceImpl) UpdateProjectForHelmApp(appName string, teamId int, userId int32) error {
	appModel, err := impl.appRepository.FindActiveByName(appName)
	if err != nil && util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching appModel", "err", err)
		return err
	}

	var appInstallationMode string
	dbConnection := impl.appRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if appModel == nil || appModel.Id == 0 {
		// for cli Helm appModel, if appModel is not yet created
		if globalUtil.IsBaseStack() {
			appInstallationMode = globalUtil.SERVER_MODE_HYPERION
		} else {
			appInstallationMode = globalUtil.SERVER_MODE_FULL
		}
		createAppRequest := bean.CreateAppDTO{
			AppName: appName,
			UserId:  userId,
			TeamId:  teamId,
		}
		_, err = impl.createAppForAppStore(&createAppRequest, tx, appInstallationMode)
		if err != nil {
			impl.logger.Errorw("error while creating appModel", "error", err)
			return err
		}
	} else {
		// update team id if appModel exist
		appModel.TeamId = teamId
		appModel.UpdateAuditLog(userId)
		err = impl.appRepository.UpdateWithTxn(appModel, tx)
		if err != nil {
			impl.logger.Errorw("error in updating project", "err", err)
			return err
		}
	}
	tx.Commit()
	return nil
}

func (impl *AppStoreDeploymentDBServiceImpl) InstallAppPostDbOperation(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	//step 4 db operation status update to deploy success
	_, err := impl.AppStoreDeployOperationStatusUpdate(installAppVersionRequest.InstalledAppId, appStoreBean.DEPLOY_SUCCESS)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return err
	}

	//step 5 create build history first entry for install app version for argocd or helm type deployments
	if !impl.deploymentTypeConfig.HelmInstallASyncMode {
		err = impl.MarkInstalledAppVersionHistorySucceeded(installAppVersionRequest.InstalledAppVersionHistoryId, installAppVersionRequest.DeploymentAppType)
		if err != nil {
			impl.logger.Errorw("error in updating installedApp History with sync ", "err", err)
			return err
		}
	}
	return nil
}

func (impl *AppStoreDeploymentDBServiceImpl) MarkInstalledAppVersionsInactiveByInstalledAppId(installedAppId int, UserId int32, tx *pg.Tx) error {
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

func (impl *AppStoreDeploymentDBServiceImpl) MarkInstalledAppVersionModelInActive(installedAppVersionModel *repository.InstalledAppVersions, UserId int32, tx *pg.Tx) error {
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

func (impl *AppStoreDeploymentDBServiceImpl) UpdateInstalledAppVersionHistoryStatus(versionHistoryId int, status string) error {
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	savedInstalledAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(versionHistoryId)
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

func (impl *AppStoreDeploymentDBServiceImpl) MarkInstalledAppVersionHistorySucceeded(versionHistoryId int, deploymentAppType string) error {
	if util.IsManifestDownload(deploymentAppType) {
		err := impl.UpdateInstalledAppVersionHistoryStatus(versionHistoryId, pipelineConfig.WorkflowSucceeded)
		if err != nil {
			impl.logger.Errorw("error on creating history for chart deployment", "error", err)
			return err
		}
	}

	if util.IsHelmApp(deploymentAppType) {
		installedAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(versionHistoryId)
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

// createAppForAppStore is an internal function used in App Store deployments; It creates an App for the InstallAppRequest
func (impl *AppStoreDeploymentDBServiceImpl) createAppForAppStore(createRequest *bean.CreateAppDTO, tx *pg.Tx, appInstallationMode string) (*bean.CreateAppDTO, error) {
	// TODO refactoring: Handling for concurrent requests with same AppName
	activeApp, err := impl.appRepository.FindActiveByName(createRequest.AppName)
	if err != nil && util.IsErrNoRows(err) {
		return nil, err
	}
	if activeApp != nil && activeApp.Id > 0 {
		impl.logger.Infow(" app already exists", "name", createRequest.AppName)
		err = &util.ApiError{
			Code:            constants.AppAlreadyExists.Code,
			InternalMessage: "app already exists",
			UserMessage:     fmt.Sprintf("app already exists with name %s", createRequest.AppName),
		}
		return nil, err
	}
	appModel := &app.App{
		Active:          true,
		AppName:         createRequest.AppName,
		TeamId:          createRequest.TeamId,
		AppType:         helper.ChartStoreApp,
		AppOfferingMode: appInstallationMode,
	}
	appModel.CreateAuditLog(createRequest.UserId)
	err = impl.appRepository.SaveWithTxn(appModel, tx)
	if err != nil {
		impl.logger.Errorw("error in saving entity ", "entity", appModel)
		return nil, err
	}
	createRequest.Id = appModel.Id
	return createRequest, nil
}

func (impl *AppStoreDeploymentDBServiceImpl) validateAndGetOverrideDeploymentAppType(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (overrideDeploymentType string, err error) {
	// initialise OverrideDeploymentType to the given DeploymentType
	overrideDeploymentType = installAppVersionRequest.DeploymentAppType

	isGitOpsConfigured, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error while checking IsGitOpsConfigured", "err", err)
		return overrideDeploymentType, err
	}
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("error in fetching app store application version", "err", err)
		return overrideDeploymentType, err
	}
	isOCIRepo := appStoreAppVersion.AppStore.DockerArtifactStore != nil
	if isOCIRepo || getAppInstallationMode(installAppVersionRequest.AppOfferingMode) == globalUtil.SERVER_MODE_HYPERION {
		overrideDeploymentType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
	}
	overrideDeploymentType, err = impl.deploymentTypeOverrideService.ValidateAndOverrideDeploymentAppType(overrideDeploymentType, isGitOpsConfigured, installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("validation error for the used deployment type", "appName", installAppVersionRequest.AppName, "err", err)
		return overrideDeploymentType, err
	}
	return overrideDeploymentType, nil
}

func (impl *AppStoreDeploymentDBServiceImpl) getEnvironmentForInstallAppRequest(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*clutserBean.EnvironmentBean, error) {

	// create env if env not exists for clusterId and namespace for hyperion mode
	if globalUtil.IsHelmApp(getAppInstallationMode(installAppVersionRequest.AppOfferingMode)) {
		envBean, err := impl.createEnvironmentIfNotExists(installAppVersionRequest)
		if err != nil {
			return nil, err
		}
		installAppVersionRequest.EnvironmentId = envBean.Id
		return envBean, nil
	}

	environment, err := impl.environmentService.GetExtendedEnvBeanById(installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching environment error", "envId", installAppVersionRequest.EnvironmentId, "err", err)
		return nil, err
	}
	return environment, nil
}

func (impl *AppStoreDeploymentDBServiceImpl) createEnvironmentIfNotExists(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*clutserBean.EnvironmentBean, error) {
	clusterId := installAppVersionRequest.ClusterId
	namespace := installAppVersionRequest.Namespace
	env, err := impl.environmentService.FindOneByNamespaceAndClusterId(namespace, clusterId)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	if env.Id != 0 {
		return env, nil
	}
	// create env
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster details", "clusterId", clusterId)
		return nil, &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: err.Error(),
			UserMessage:     "Invalid cluster details!",
		}
	}

	environmentBean := &clutserBean.EnvironmentBean{
		Environment: clusterService.BuildEnvironmentName(cluster.ClusterName, namespace),
		ClusterId:   clusterId,
		Namespace:   namespace,
		Default:     false,
		Active:      true,
	}
	envCreateRes, err := impl.environmentService.Create(environmentBean, installAppVersionRequest.UserId)
	if err != nil {
		return nil, err
	}

	return envCreateRes, nil
}

func getAppInstallationMode(appOfferingMode string) string {
	appInstallationMode := globalUtil.SERVER_MODE_FULL
	if globalUtil.IsBaseStack() || globalUtil.IsHelmApp(appOfferingMode) {
		appInstallationMode = globalUtil.SERVER_MODE_HYPERION
	}
	return appInstallationMode
}
