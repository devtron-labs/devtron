package service

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/adapter"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/bean"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
)

func (impl *AppStoreDeploymentServiceImpl) AppStoreDeployOperationDB(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
	err := impl.setEnvironmentForInstallAppRequest(installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw("error in getting environment for install helm chart", "envId", installAppVersionRequest.EnvironmentId, "err", err)
		return nil, err
	}
	// Stage 1:  Create App in tx (Only if AppId is not set already)
	if installAppVersionRequest.AppId <= 0 {
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
	err = impl.validateAndOverrideDeploymentAppType(installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw("error in validating deployment app type", "error", err)
		return nil, err
	}
	// Stage 2: ends

	// Stage 3: save installed_apps model
	if util2.IsFullStack() && util.IsAcdApp(installAppVersionRequest.DeploymentAppType) {
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

func (impl *AppStoreDeploymentServiceImpl) validateAndOverrideDeploymentAppType(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	isGitOpsConfigured, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error while checking IsGitOpsConfigured", "err", err)
		return err
	}
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.logger.Errorw("error in fetching app store application version", "err", err)
		return err
	}
	isOCIRepo := appStoreAppVersion.AppStore.DockerArtifactStore != nil
	if isOCIRepo || getAppInstallationMode(installAppVersionRequest.AppOfferingMode) == util2.SERVER_MODE_HYPERION {
		installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
	}
	err = impl.deploymentTypeOverrideService.SetAndValidateDeploymentAppType(&installAppVersionRequest.DeploymentAppType, isGitOpsConfigured, installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("validation error for the used deployment type", "appName", installAppVersionRequest.AppName, "err", err)
		return err
	}
	return nil
}

func (impl *AppStoreDeploymentServiceImpl) setEnvironmentForInstallAppRequest(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {

	// create env if env not exists for clusterId and namespace for hyperion mode
	if util2.IsHelmApp(getAppInstallationMode(installAppVersionRequest.AppOfferingMode)) {
		// TODO refactoring: should it be in transaction
		envId, err := impl.createEnvironmentIfNotExists(installAppVersionRequest)
		if err != nil {
			return err
		}
		installAppVersionRequest.EnvironmentId = envId
	}

	environment, err := impl.environmentRepository.FindById(installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("fetching environment error", "envId", installAppVersionRequest.EnvironmentId, "err", err)
		return err
	}
	// setting additional env data required in appStoreBean.InstallAppVersionDTO
	installAppVersionRequest.Environment = environment
	installAppVersionRequest.ACDAppName = fmt.Sprintf("%s-%s", installAppVersionRequest.AppName, installAppVersionRequest.Environment.Name)
	installAppVersionRequest.ClusterId = environment.ClusterId
	return nil
}

func getAppInstallationMode(appOfferingMode string) string {
	appInstallationMode := util2.SERVER_MODE_FULL
	if util2.IsBaseStack() || util2.IsHelmApp(appOfferingMode) {
		appInstallationMode = util2.SERVER_MODE_HYPERION
	}
	return appInstallationMode
}
