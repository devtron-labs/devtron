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
	"net/http"
)

func (impl AppStoreDeploymentServiceImpl) AppStoreDeployOperationDB(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx, skipAppCreation bool) (*appStoreBean.InstallAppVersionDTO, error) {

	var isInternalUse = impl.deploymentTypeConfig.IsInternalUse

	isGitOpsConfigured, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error while checking IsGitOpsConfigured", "err", err)
		return nil, err
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

	var isOCIRepo bool
	if appStoreAppVersion.AppStore.DockerArtifactStore != nil {
		isOCIRepo = true
	} else {
		isOCIRepo = false
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

	if environment.IsVirtualEnvironment {
		if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) || util.IsHelmApp(installAppVersionRequest.DeploymentAppType) {
			impl.logger.Errorw("deployment type type helm/argocd not supported on virtual cluster")
			err := &util.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				InternalMessage: "deployment type type helm/argocd not supported on virtual cluster",
				UserMessage:     "deployment type type helm/argocd not supported on virtual cluster",
			}
			return nil, err
		}
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

	if !isInternalUse && !environment.IsVirtualEnvironment {
		if isGitOpsConfigured && appInstallationMode == util2.SERVER_MODE_FULL && !isOCIRepo {
			installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
		} else {
			installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
		}
	}
	if installAppVersionRequest.DeploymentAppType == "" {
		if environment.IsVirtualEnvironment {
			installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD
		} else if isGitOpsConfigured && appInstallationMode == util2.SERVER_MODE_FULL && !isOCIRepo {
			installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_ACD
		} else {
			installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
		}
	}

	if util2.IsFullStack() {
		installAppVersionRequest.GitOpsRepoName = impl.gitOpsConfigReadService.GetGitOpsRepoName(installAppVersionRequest.AppName)
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

	adapter.SetGeneratedHelmPackageName(installAppVersionRequest, installedApp.UpdatedOn)

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

func (impl AppStoreDeploymentServiceImpl) GetInstalledAppByClusterNamespaceAndName(clusterId int, namespace string, appName string) (*appStoreBean.InstallAppVersionDTO, error) {
	installedApp, err := impl.installedAppRepository.GetInstalledApplicationByClusterIdAndNamespaceAndAppName(clusterId, namespace, appName)
	if err != nil {
		if err == pg.ErrNoRows {
			impl.logger.Warnw("no installed apps found", "clusterId", clusterId)
			return nil, nil
		} else {
			impl.logger.Errorw("error while fetching installed apps", "clusterId", clusterId, "error", err)
			return nil, err
		}
	}

	if installedApp.Id > 0 {
		installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedApp.Id, installedApp.EnvironmentId)
		if err != nil {
			return nil, err
		}
		return adapter.GenerateInstallAppVersionDTO(installedApp, installedAppVersion), nil
	}

	return nil, nil
}

func (impl AppStoreDeploymentServiceImpl) GetInstalledAppByInstalledAppId(installedAppId int) (*appStoreBean.InstallAppVersionDTO, error) {
	installedAppVersion, err := impl.installedAppRepository.GetActiveInstalledAppVersionByInstalledAppId(installedAppId)
	if err != nil {
		return nil, err
	}
	installedApp := &installedAppVersion.InstalledApp
	return adapter.GenerateInstallAppVersionDTO(installedApp, installedAppVersion), nil
}
