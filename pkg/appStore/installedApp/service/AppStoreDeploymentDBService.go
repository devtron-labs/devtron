package service

import (
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/adapter"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/bean"
	gitOpsBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/config/bean"
	validationBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation/bean"
	globalUtil "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/ChartsUtil"
	"github.com/go-pg/pg"
	"net/http"
)

func (impl *AppStoreDeploymentServiceImpl) validateCustomGitOpsRepoURL(gitOpsConfigurationStatus *gitOpsBean.GitOpsConfigurationStatus, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (string, bool, error) {
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

func (impl *AppStoreDeploymentServiceImpl) AppStoreDeployOperationDB(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx,
	skipAppCreation bool, installAppVersionRequestType appStoreBean.InstallAppVersionRequestType) (*appStoreBean.InstallAppVersionDTO, error) {

	var isInternalUse = impl.deploymentTypeConfig.IsInternalUse

	gitOpsConfigurationStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error while checking IsGitOpsConfigured", "err", err)
		return nil, err
	}

	if isInternalUse && !gitOpsConfigurationStatus.IsGitOpsConfigured && util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
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

	isOCIRepo := false
	if appStoreAppVersion.AppStore.DockerArtifactStore != nil {
		isOCIRepo = true
	}

	var appInstallationMode string
	if globalUtil.IsBaseStack() || globalUtil.IsHelmApp(installAppVersionRequest.AppOfferingMode) {
		appInstallationMode = globalUtil.SERVER_MODE_HYPERION
	} else {
		appInstallationMode = globalUtil.SERVER_MODE_FULL
	}

	// create env if env not exists for clusterId and namespace for hyperion mode
	if globalUtil.IsHelmApp(appInstallationMode) {
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

	if !isInternalUse {
		installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
		if gitOpsConfigurationStatus.IsGitOpsConfigured && appInstallationMode == globalUtil.SERVER_MODE_FULL && !isOCIRepo {
			installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
			// Handling for chart-group deployment request
			if (installAppVersionRequestType == appStoreBean.BULK_DEPLOY_REQUEST ||
				installAppVersionRequestType == appStoreBean.DEFAULT_COMPONENT_DEPLOYMENT_REQUEST) &&
				gitOpsConfigurationStatus.AllowCustomRepository &&
				len(installAppVersionRequest.GitOpsRepoURL) == 0 {
				installAppVersionRequest.GitOpsRepoURL = apiBean.GIT_REPO_DEFAULT
			}
		}
	}
	if installAppVersionRequest.DeploymentAppType == "" {
		installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
		if gitOpsConfigurationStatus.IsGitOpsConfigured && !isOCIRepo {
			installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
			// Handling for chart-group deployment request
			if (installAppVersionRequestType == appStoreBean.BULK_DEPLOY_REQUEST ||
				installAppVersionRequestType == appStoreBean.DEFAULT_COMPONENT_DEPLOYMENT_REQUEST) &&
				gitOpsConfigurationStatus.AllowCustomRepository &&
				len(installAppVersionRequest.GitOpsRepoURL) == 0 {
				installAppVersionRequest.GitOpsRepoURL = apiBean.GIT_REPO_DEFAULT
			}
		}
	}

	if globalUtil.IsFullStack() && util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
		if !gitOpsConfigurationStatus.AllowCustomRepository && (len(installAppVersionRequest.GitOpsRepoURL) != 0 && installAppVersionRequest.GitOpsRepoURL != apiBean.GIT_REPO_DEFAULT) {
			impl.logger.Errorw("invalid installAppVersionRequest", "error", "custom repo url is not valid, as the global configuration is updated")
			err = &util.ApiError{
				HttpStatusCode:  http.StatusConflict,
				UserMessage:     "Invalid request! Please configure GitOps with 'Allow changing git repository for application'.",
				InternalMessage: "Invalid request! Custom repository is not valid, as the global configuration is updated",
			}
			return nil, err
		}
		if gitOpsConfigurationStatus.AllowCustomRepository && len(installAppVersionRequest.GitOpsRepoURL) == 0 {
			impl.logger.Errorw("invalid installAppVersionRequest", "error", "gitRepoURL is required")
			err = &util.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				Code:            constants.GitOpsConfigValidationConflict,
				InternalMessage: "Invalid request payload! gitRepoURL key is required.",
				UserMessage:     "Invalid request payload! gitRepoURL key is required.",
			}
			return nil, err
		}
		// This should be set before to validateCustomGitOpsRepoURL,
		// as validateCustomGitOpsRepoURL will override installAppVersionRequest.GitOpsRepoURL
		if !ChartsUtil.IsGitOpsRepoNotConfigured(installAppVersionRequest.GitOpsRepoURL) &&
			installAppVersionRequest.GitOpsRepoURL != apiBean.GIT_REPO_DEFAULT {
			// If GitOps repo is configured and not configured to bean.GIT_REPO_DEFAULT
			installAppVersionRequest.IsCustomRepository = true
		}
		// validating the git repository configured for GitOps deployments
		gitopsRepoURL, isNew, gitRepoErr := impl.validateCustomGitOpsRepoURL(gitOpsConfigurationStatus, installAppVersionRequest)
		if gitRepoErr != nil {
			// Found validation err
			impl.logger.Errorw("validation failed for GitOps repository", "repo url", installAppVersionRequest.GitOpsRepoURL, "err", gitRepoErr)
			return nil, gitRepoErr
		}
		// ValidateCustomGitRepoURL returns sanitized repo url after validation
		installAppVersionRequest.GitOpsRepoURL = gitopsRepoURL
		installAppVersionRequest.IsNewGitOpsRepo = isNew
	}
	installedAppModel := adapter.NewInstallAppModel(installAppVersionRequest, appStoreBean.DEPLOY_INIT)
	installedApp, err := impl.installedAppRepository.CreateInstalledApp(installedAppModel, tx)
	if err != nil {
		impl.logger.Errorw("error while creating install app", "error", err)
		return nil, err
	}
	installAppVersionRequest.InstalledAppId = installedApp.Id

	installedAppVersions := adapter.NewInstallAppVersionsModel(installAppVersionRequest)
	_, err = impl.installedAppRepository.CreateInstalledAppVersion(installedAppVersions, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}

	installAppVersionRequest.InstalledAppVersionId = installedAppVersions.Id
	installAppVersionRequest.Id = installedAppVersions.Id

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
	if installAppVersionRequest.DefaultClusterComponent {
		clusterInstalledAppsModel := adapter.NewClusterInstalledAppsModel(installAppVersionRequest, environment.ClusterId)
		err = impl.clusterInstalledAppsRepository.Save(clusterInstalledAppsModel, tx)
		if err != nil {
			impl.logger.Errorw("error while creating cluster install app", "error", err)
			return nil, err
		}
	}
	return installAppVersionRequest, nil
}

func (impl *AppStoreDeploymentServiceImpl) AppStoreDeployOperationStatusUpdate(installAppId int, status appStoreBean.AppstoreDeploymentStatus) (bool, error) {
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
