package service

import (
	"encoding/json"
	"fmt"
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/adapter"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	discoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode/deployment"
	util4 "github.com/devtron-labs/devtron/pkg/appStore/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	clusterService "github.com/devtron-labs/devtron/pkg/cluster"
	clutserBean "github.com/devtron-labs/devtron/pkg/cluster/repository/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	gitOpsBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/config/bean"
	validationBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/providerConfig"
	globalUtil "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type AppStoreDeploymentDBService interface {
	// AppStoreDeployOperationDB is used to perform Pre-Install DB operations in App Store deployments
	AppStoreDeployOperationDB(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx, requestType appStoreBean.InstallAppVersionRequestType) (*appStoreBean.InstallAppVersionDTO, error)
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
	UpdateProjectForHelmApp(appName, displayName string, teamId int, userId int32) error
	// InstallAppPostDbOperation is used to perform Post-Install DB operations in App Store deployments
	InstallAppPostDbOperation(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error
	// MarkInstalledAppVersionsInactiveByInstalledAppId will mark the repository.InstalledAppVersions inactive for the given InstalledAppId
	MarkInstalledAppVersionsInactiveByInstalledAppId(installedAppId int, UserId int32, tx *pg.Tx) error
	// MarkInstalledAppVersionModelInActive will mark the given repository.InstalledAppVersions inactive
	MarkInstalledAppVersionModelInActive(installedAppVersionModel *repository.InstalledAppVersions, UserId int32, tx *pg.Tx) error
	// MarkHelmInstalledAppDeploymentSucceeded will mark the helm installed repository.InstalledAppVersionHistory Status - Succeeded
	MarkHelmInstalledAppDeploymentSucceeded(versionHistoryId int) error
	// UpdateInstalledAppVersionHistoryStatus will update the Status in the repository.InstalledAppVersionHistory
	UpdateInstalledAppVersionHistoryStatus(versionHistoryId int, status string) error
	// GetActiveAppForAppIdentifierOrReleaseName returns app db model for an app unique identifier or from display_name if either exists else it throws pg.ErrNoRows
	GetActiveAppForAppIdentifierOrReleaseName(appNameUniqueIdentifier, releaseName string) (*app.App, error)
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
	fullModeDeploymentService            deployment.FullModeDeploymentService
	appStoreValidator                    AppStoreValidator
	installedAppDbService                EAMode.InstalledAppDBService
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
	deploymentTypeOverrideService providerConfig.DeploymentTypeOverrideService,
	fullModeDeploymentService deployment.FullModeDeploymentService, appStoreValidator AppStoreValidator,
	installedAppDbService EAMode.InstalledAppDBService) *AppStoreDeploymentDBServiceImpl {
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
		fullModeDeploymentService:            fullModeDeploymentService,
		appStoreValidator:                    appStoreValidator,
		installedAppDbService:                installedAppDbService,
	}
}

func (impl *AppStoreDeploymentDBServiceImpl) AppStoreDeployOperationDB(installRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx, requestType appStoreBean.InstallAppVersionRequestType) (*appStoreBean.InstallAppVersionDTO, error) {
	environment, err := impl.getEnvironmentForInstallAppRequest(installRequest)
	if err != nil {
		impl.logger.Errorw("error in getting environment for install helm chart", "envId", installRequest.EnvironmentId, "err", err)
		return nil, err
	}
	// setting additional env data required in appStoreBean.InstallAppVersionDTO
	adapter.UpdateAdditionalEnvDetails(installRequest, environment)

	impl.appStoreValidator.Validate(installRequest, environment)

	// Stage 1:  Create App in tx (Only if AppId is not set already)
	if installRequest.AppId == 0 {
		appCreateRequest := &bean.CreateAppDTO{
			AppName: installRequest.AppName,
			TeamId:  installRequest.TeamId,
			UserId:  installRequest.UserId,
		}
		if util4.IsExternalChartStoreApp(installRequest.DisplayName) {
			//this is the case of linking external helm app to devtron chart store
			appCreateRequest.AppType = helper.ExternalChartStoreApp
			appCreateRequest.DisplayName = installRequest.DisplayName
		}
		appCreateRequest, err = impl.createAppForAppStore(appCreateRequest, tx, getAppInstallationMode(installRequest.AppOfferingMode))
		if err != nil {
			impl.logger.Errorw("error while creating app", "error", err)
			return nil, err
		}
		installRequest.AppId = appCreateRequest.Id
	}
	// Stage 1: ends

	gitOpsConfigStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error while checking IsGitOpsConfigured", "err", err)
		return nil, err
	}

	// Stage 2:  validate deployment app type and override if ExternallyManagedDeploymentType
	overrideDeploymentType, err := impl.validateAndGetOverrideDeploymentAppType(installRequest, gitOpsConfigStatus.IsGitOpsConfigured)
	if err != nil {
		impl.logger.Errorw("error in validating deployment app type", "error", err)
		return nil, err
	}
	installRequest.UpdateDeploymentAppType(overrideDeploymentType)
	// Stage 2: ends

	// Stage 3: save installed_apps model
	if globalUtil.IsFullStack() && util.IsAcdApp(installRequest.DeploymentAppType) {
		installRequest.UpdateCustomGitOpsRepoUrl(gitOpsConfigStatus.AllowCustomRepository, requestType)
		// validate GitOps request
		validationErr := impl.validateGitOpsRequest(gitOpsConfigStatus.AllowCustomRepository, installRequest.GitOpsRepoURL)
		if validationErr != nil {
			impl.logger.Errorw("GitOps request validation error", "allowCustomRepository", gitOpsConfigStatus.AllowCustomRepository, "gitOpsRepoURL", installRequest.GitOpsRepoURL, "err", validationErr)
			return nil, validationErr
		}
		// This should be set before to validateCustomGitOpsRepoURL,
		// as validateCustomGitOpsRepoURL will override installRequest.GitOpsRepoURL
		if !apiGitOpsBean.IsGitOpsRepoNotConfigured(installRequest.GitOpsRepoURL) &&
			installRequest.GitOpsRepoURL != apiGitOpsBean.GIT_REPO_DEFAULT {
			// If GitOps repo is configured and not configured to bean.GIT_REPO_DEFAULT
			installRequest.IsCustomRepository = true
		}
		// validating the git repository configured for GitOps deployments
		gitOpsRepoURL, isNew, gitRepoErr := impl.validateCustomGitOpsRepoURL(gitOpsConfigStatus, installRequest)
		if gitRepoErr != nil {
			// Found validation err
			impl.logger.Errorw("validation failed for GitOps repository", "repo url", installRequest.GitOpsRepoURL, "err", gitRepoErr)
			return nil, gitRepoErr
		}
		// validateCustomGitOpsRepoURL returns sanitized repo url after validation
		installRequest.GitOpsRepoURL = gitOpsRepoURL
		installRequest.IsNewGitOpsRepo = isNew
	} else {
		// overriding gitOps repository url -> empty (for Helm installation)
		installRequest.GitOpsRepoURL = ""
	}
	installedAppModel := adapter.NewInstallAppModel(installRequest, appStoreBean.DEPLOY_INIT)
	installedApp, err := impl.installedAppRepository.CreateInstalledApp(installedAppModel, tx)
	if err != nil {
		impl.logger.Errorw("error while creating install app", "error", err)
		return nil, err
	}
	installRequest.InstalledAppId = installedApp.Id
	// Stage 3: ends

	// Stage 4: save installed_app_versions model
	installedAppVersions := adapter.NewInstallAppVersionsModel(installRequest)
	_, err = impl.installedAppRepository.CreateInstalledAppVersion(installedAppVersions, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	installRequest.InstalledAppVersionId = installedAppVersions.Id
	installRequest.Id = installedAppVersions.Id
	// Stage 4: ends

	// populate HelmPackageName; It's used in case of virtual deployments
	installRequest.HelmPackageName = adapter.GetGeneratedHelmPackageName(
		installRequest.AppName,
		installRequest.EnvironmentName,
		installedApp.UpdatedOn)

	// Stage 5: save installed_app_version_history model
	helmInstallConfigDTO := appStoreBean.HelmReleaseStatusConfig{
		InstallAppVersionHistoryId: 0,
		Message:                    "Install initiated",
		IsReleaseInstalled:         false,
		ErrorInInstallation:        false,
	}
	installedAppVersionHistory, err := adapter.NewInstallAppVersionHistoryModel(installRequest, pipelineConfig.WorkflowInProgress, helmInstallConfigDTO)
	if err != nil {
		impl.logger.Errorw("error in helm install config marshal", "err", err)
	}
	_, err = impl.installedAppRepositoryHistory.CreateInstalledAppVersionHistory(installedAppVersionHistory, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	installRequest.InstalledAppVersionHistoryId = installedAppVersionHistory.Id
	// Stage 5: ends
	return installRequest, nil
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

func (impl *AppStoreDeploymentDBServiceImpl) GetActiveAppForAppIdentifierOrReleaseName(appNameUniqueIdentifier, releaseName string) (*app.App, error) {
	app, err := impl.appRepository.FindActiveByName(appNameUniqueIdentifier)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching app meta data by unique app identifier", "appNameUniqueIdentifier", appNameUniqueIdentifier, "err", err)
		return nil, err
	} else if util.IsErrNoRows(err) {
		//find app by displayName/releaseName if not found by unique identifier
		app, err = impl.appRepository.FindActiveByName(releaseName)
		if err != nil {
			impl.logger.Errorw("error in fetching app meta data by display name", "displayName", releaseName, "err", err)
			return nil, err
		}
	}
	return app, nil
}

func (impl *AppStoreDeploymentDBServiceImpl) UpdateProjectForHelmApp(appName, displayName string, teamId int, userId int32) error {
	appModel, err := impl.GetActiveAppForAppIdentifierOrReleaseName(appName, displayName)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching appModel by appName", "appName", appName, "err", err)
		return err
	}

	// only external app will have a display name, so checking the following case only for external apps
	if appModel != nil && appModel.Id > 0 && len(displayName) > 0 {
		/*
				1. now we will check if for that appModel, installed_app entries are present or not i.e. linked to devtron or not,
			    2. if not, then let the normal flow continue as we can change the app_name with app unique identifier.
				3. if exists then we will create new app entries with uniqueAppNameIdentifier for all installed apps.
		*/
		isLinkedToDevtron, installedApps, err := impl.installedAppDbService.IsExternalAppLinkedToChartStore(appModel.Id)
		if err != nil {
			impl.logger.Errorw("UpdateProjectForHelmApp, error in checking IsExternalAppLinkedToChartStore", "appId", appModel.Id, "err", err)
			return err
		}
		if isLinkedToDevtron {
			err := impl.installedAppDbService.CreateNewAppEntryForAllInstalledApps(installedApps)
			if err != nil {
				impl.logger.Errorw("UpdateProjectForHelmApp, error in CreateNewAppEntryForAllInstalledApps", "appName", displayName, "err", err)
				//not returning from here, project update req is yet to be processed for requested ext-app
			}
		}
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
		if util4.IsExternalChartStoreApp(displayName) {
			createAppRequest.AppType = helper.ExternalChartStoreApp
			createAppRequest.DisplayName = displayName
		}
		_, err = impl.createAppForAppStore(&createAppRequest, tx, appInstallationMode)
		if err != nil {
			impl.logger.Errorw("error while creating appModel", "error", err)
			return err
		}
	} else {
		if util4.IsExternalChartStoreApp(displayName) {
			//handling the case when ext-helm app is already assigned to a project and an entry already exist in app table but
			//not yet migrated, then this will override app_name with unique identifier app name and update display_name also
			appModel.AppName = appName
			appModel.DisplayName = displayName
		}
		// update team id if appModel exist
		appModel.TeamId = teamId
		appModel.AppOfferingMode = globalUtil.SERVER_MODE_FULL
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
	// step 1  update installed_apps.status to deploy success
	_, err := impl.AppStoreDeployOperationStatusUpdate(installAppVersionRequest.InstalledAppId, appStoreBean.DEPLOY_SUCCESS)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return err
	}

	// step 2 mark deployment succeeded for manifest download type helm charts
	if util.IsManifestDownload(installAppVersionRequest.DeploymentAppType) {
		err := impl.UpdateInstalledAppVersionHistoryStatus(installAppVersionRequest.InstalledAppVersionHistoryId, pipelineConfig.WorkflowSucceeded)
		if err != nil {
			impl.logger.Errorw("error on updating deployment status to history for chart store deployment", "versionHistoryId", installAppVersionRequest.InstalledAppVersionHistoryId, "error", err)
			return err
		}
	}

	// step 3 mark deployment succeeded for helm installed helm charts
	if util.IsHelmApp(installAppVersionRequest.DeploymentAppType) && !impl.deploymentTypeConfig.HelmInstallASyncMode {
		err = impl.MarkHelmInstalledAppDeploymentSucceeded(installAppVersionRequest.InstalledAppVersionHistoryId)
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

func (impl *AppStoreDeploymentDBServiceImpl) MarkHelmInstalledAppDeploymentSucceeded(versionHistoryId int) error {
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
	return nil
}

// createAppForAppStore is an internal function used in App Store deployments; It creates an App for the InstallAppRequest
func (impl *AppStoreDeploymentDBServiceImpl) createAppForAppStore(createRequest *bean.CreateAppDTO, tx *pg.Tx, appInstallationMode string) (*bean.CreateAppDTO, error) {
	// TODO refactoring: Handling for concurrent requests with same AppName
	activeApp, err := impl.appRepository.FindActiveByName(createRequest.AppName)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	if activeApp != nil && activeApp.Id > 0 {
		impl.logger.Infow("app already exists", "name", createRequest.AppName)
		err = &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
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
	if createRequest.AppType == helper.ExternalChartStoreApp {
		//when linking ext helm app to chart store, there can be a case that two (or more) external apps can have same name, in diff namespaces or diff
		//clusters, so now we are storing display_name also to get rid of multiple installed apps pointing to the same app, which caused unwarranted
		//behaviours. appName in this case will be displayName-namespace-clusterId
		appModel.DisplayName = createRequest.DisplayName
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

func (impl *AppStoreDeploymentDBServiceImpl) validateAndGetOverrideDeploymentAppType(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, isGitOpsConfigured bool) (overrideDeploymentType string, err error) {
	// initialise OverrideDeploymentType to the given DeploymentType
	overrideDeploymentType = installAppVersionRequest.DeploymentAppType
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("error in fetching app store application version", "err", err)
		return overrideDeploymentType, err
	}

	// virtual environments only supports Manifest Download
	if installAppVersionRequest.Environment.IsVirtualEnvironment && util.IsManifestPush(installAppVersionRequest.DeploymentAppType) {
		impl.logger.Errorw("invalid deployment type for a virtual environment", "deploymentType", installAppVersionRequest.DeploymentAppType)
		err = &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: fmt.Sprintf("Deployment type '%s' is not supported on virtual cluster", installAppVersionRequest.DeploymentAppType),
			UserMessage:     fmt.Sprintf("Deployment type '%s' is not supported on virtual cluster", installAppVersionRequest.DeploymentAppType),
		}
		return overrideDeploymentType, err
	}

	// OCI chart currently supports HELM installation only
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
	} else if env != nil && env.Id != 0 {
		return env, nil
	} else {
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
}

func getAppInstallationMode(appOfferingMode string) string {
	appInstallationMode := globalUtil.SERVER_MODE_FULL
	if globalUtil.IsBaseStack() || globalUtil.IsHelmApp(appOfferingMode) {
		appInstallationMode = globalUtil.SERVER_MODE_HYPERION
	}
	return appInstallationMode
}

func (impl *AppStoreDeploymentDBServiceImpl) validateGitOpsRequest(allowCustomRepository bool, gitOpsRepoURL string) (err error) {
	if !allowCustomRepository && (len(gitOpsRepoURL) != 0 && gitOpsRepoURL != apiGitOpsBean.GIT_REPO_DEFAULT) {
		impl.logger.Errorw("invalid installRequest", "error", "custom repo url is not valid, as the global configuration is updated")
		err = &util.ApiError{
			HttpStatusCode:  http.StatusConflict,
			UserMessage:     "Invalid request! Please configure GitOps with 'Allow changing git repository for application'.",
			InternalMessage: "Invalid request! Custom repository is not valid, as the global configuration is updated",
		}
		return err
	}
	if allowCustomRepository && len(gitOpsRepoURL) == 0 {
		impl.logger.Errorw("invalid installRequest", "error", "gitRepoURL is required")
		err = &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			Code:            constants.GitOpsConfigValidationConflict,
			InternalMessage: "Invalid request payload! gitRepoURL key is required.",
			UserMessage:     "Invalid request payload! gitRepoURL key is required.",
		}
		return err
	}
	return nil
}

func (impl *AppStoreDeploymentDBServiceImpl) validateCustomGitOpsRepoURL(gitOpsConfigurationStatus *gitOpsBean.GitOpsConfigurationStatus, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (string, bool, error) {
	validateCustomGitRepoURLRequest := validationBean.ValidateCustomGitRepoURLRequest{
		GitRepoURL:     installAppVersionRequest.GitOpsRepoURL,
		AppName:        installAppVersionRequest.AppName,
		UserId:         installAppVersionRequest.UserId,
		GitOpsProvider: gitOpsConfigurationStatus.Provider,
	}
	gitopsRepoURL, isNew, gitRepoErr := impl.fullModeDeploymentService.ValidateCustomGitRepoURL(validateCustomGitRepoURLRequest)
	if gitRepoErr != nil {
		// Found validation err
		impl.logger.Errorw("found validation error in custom GitOps repo", "repo url", installAppVersionRequest.GitOpsRepoURL, "err", gitRepoErr)
		apiErr := &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			UserMessage:     gitRepoErr.Error(),
			InternalMessage: gitRepoErr.Error(),
		}
		return gitopsRepoURL, isNew, apiErr
	}
	return gitopsRepoURL, isNew, nil
}
