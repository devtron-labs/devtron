package service

import (
	"fmt"
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/adapter"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/bean"
	gitOpsBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/config/bean"
	validationBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation/bean"
	globalUtil "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"net/http"
)

func (impl *AppStoreDeploymentServiceImpl) AppStoreDeployOperationDB(installRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx,
	skipAppCreation bool, requestType appStoreBean.InstallAppVersionRequestType) (*appStoreBean.InstallAppVersionDTO, error) {

	var isInternalUse = impl.deploymentTypeConfig.IsInternalUse

	gitOpsConfigStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error while checking IsGitOpsConfigured", "err", err)
		return nil, err
	}

	if isInternalUse && !gitOpsConfigStatus.IsGitOpsConfigured && util.IsAcdApp(installRequest.DeploymentAppType) {
		impl.logger.Errorw("gitops not configured but selected for CD")
		err := &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: "GitOps integration is not installed/configured. Please install/configure gitops or use helm option.",
			UserMessage:     "GitOps integration is not installed/configured. Please install/configure gitops or use helm option.",
		}
		return nil, err
	}

	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}

	isOCIRepo := false
	if appStoreAppVersion.AppStore.DockerArtifactStore != nil {
		isOCIRepo = true
	}

	var appInstallationMode string
	if globalUtil.IsBaseStack() || globalUtil.IsHelmApp(installRequest.AppOfferingMode) {
		appInstallationMode = globalUtil.SERVER_MODE_HYPERION
	} else {
		appInstallationMode = globalUtil.SERVER_MODE_FULL
	}

	// create env if env not exists for clusterId and namespace for hyperion mode
	if globalUtil.IsHelmApp(appInstallationMode) {
		envId, err := impl.createEnvironmentIfNotExists(installRequest)
		if err != nil {
			return nil, err
		}
		installRequest.EnvironmentId = envId
	}

	environment, err := impl.environmentRepository.FindById(installRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return nil, err
	}
	installRequest.Environment = environment
	installRequest.ACDAppName = fmt.Sprintf("%s-%s", installRequest.AppName, installRequest.Environment.Name)
	installRequest.ClusterId = environment.ClusterId
	appCreateRequest := &bean.CreateAppDTO{
		Id:      installRequest.AppId,
		AppName: installRequest.AppName,
		TeamId:  installRequest.TeamId,
		UserId:  installRequest.UserId,
	}
	appCreateRequest, err = impl.createAppForAppStore(appCreateRequest, tx, appInstallationMode, skipAppCreation)
	if err != nil {
		impl.logger.Errorw("error while creating app", "error", err)
		return nil, err
	}
	installRequest.AppId = appCreateRequest.Id

	if !isInternalUse {
		installRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
		if gitOpsConfigStatus.IsGitOpsConfigured && appInstallationMode == globalUtil.SERVER_MODE_FULL && !isOCIRepo {
			installRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
		}
	}
	if installRequest.DeploymentAppType == "" {
		installRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
		if gitOpsConfigStatus.IsGitOpsConfigured && !isOCIRepo {
			installRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
		}
	}
	if globalUtil.IsFullStack() && util.IsAcdApp(installRequest.DeploymentAppType) {
		installRequest.UpdateCustomGitOpsRepoUrl(gitOpsConfigStatus.AllowCustomRepository, requestType)
		if !gitOpsConfigStatus.AllowCustomRepository && (len(installRequest.GitOpsRepoURL) != 0 && installRequest.GitOpsRepoURL != apiGitOpsBean.GIT_REPO_DEFAULT) {
			impl.logger.Errorw("invalid installRequest", "error", "custom repo url is not valid, as the global configuration is updated")
			err = &util.ApiError{
				HttpStatusCode:  http.StatusConflict,
				UserMessage:     "Invalid request! Please configure GitOps with 'Allow changing git repository for application'.",
				InternalMessage: "Invalid request! Custom repository is not valid, as the global configuration is updated",
			}
			return nil, err
		}
		if gitOpsConfigStatus.AllowCustomRepository && len(installRequest.GitOpsRepoURL) == 0 {
			impl.logger.Errorw("invalid installRequest", "error", "gitRepoURL is required")
			err = &util.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				Code:            constants.GitOpsConfigValidationConflict,
				InternalMessage: "Invalid request payload! gitRepoURL key is required.",
				UserMessage:     "Invalid request payload! gitRepoURL key is required.",
			}
			return nil, err
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
	}
	installedAppModel := adapter.NewInstallAppModel(installRequest, appStoreBean.DEPLOY_INIT)
	installedApp, err := impl.installedAppRepository.CreateInstalledApp(installedAppModel, tx)
	if err != nil {
		impl.logger.Errorw("error while creating install app", "error", err)
		return nil, err
	}
	installRequest.InstalledAppId = installedApp.Id

	installedAppVersions := adapter.NewInstallAppVersionsModel(installRequest)
	_, err = impl.installedAppRepository.CreateInstalledAppVersion(installedAppVersions, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}

	installRequest.InstalledAppVersionId = installedAppVersions.Id
	installRequest.Id = installedAppVersions.Id

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
	if installRequest.DefaultClusterComponent {
		clusterInstalledAppsModel := adapter.NewClusterInstalledAppsModel(installRequest, environment.ClusterId)
		err = impl.clusterInstalledAppsRepository.Save(clusterInstalledAppsModel, tx)
		if err != nil {
			impl.logger.Errorw("error while creating cluster install app", "error", err)
			return nil, err
		}
	}
	return installRequest, nil
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
