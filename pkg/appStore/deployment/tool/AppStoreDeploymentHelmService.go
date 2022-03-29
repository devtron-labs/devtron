package appStoreDeploymentTool

import (
	"context"
	"errors"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	appStoreRepository "github.com/devtron-labs/devtron/pkg/appStore/repository"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/ghodss/yaml"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type AppStoreDeploymentHelmService interface {
	InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error)
	GetAppStatus(installedAppAndEnvDetails appStoreRepository.InstalledAppAndEnvDetails, w http.ResponseWriter, r *http.Request, token string) (string, error)
	DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *appStoreRepository.InstalledApps, dbTransaction *pg.Tx) error
	RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32) (*appStoreBean.InstallAppVersionDTO, bool, error)
}

type AppStoreDeploymentHelmServiceImpl struct {
	Logger                               *zap.SugaredLogger
	helmAppService                       client.HelmAppService
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	environmentRepository                clusterRepository.EnvironmentRepository
}

func NewAppStoreDeploymentHelmServiceImpl(logger *zap.SugaredLogger, helmAppService client.HelmAppService, appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	environmentRepository clusterRepository.EnvironmentRepository) *AppStoreDeploymentHelmServiceImpl {
	return &AppStoreDeploymentHelmServiceImpl{
		Logger:                               logger,
		helmAppService:                       helmAppService,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		environmentRepository:                environmentRepository,
	}
}

func (impl AppStoreDeploymentHelmServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error) {
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.Logger.Errorw("fetching error", "err", err)
		return installAppVersionRequest, err
	}

	installReleaseRequest := &client.InstallReleaseRequest{
		ChartName:    appStoreAppVersion.Name,
		ChartVersion: appStoreAppVersion.Version,
		ValuesYaml:   installAppVersionRequest.ValuesOverrideYaml,
		ChartRepository: &client.ChartRepository{
			Name:     appStoreAppVersion.AppStore.ChartRepo.Name,
			Url:      appStoreAppVersion.AppStore.ChartRepo.Url,
			Username: appStoreAppVersion.AppStore.ChartRepo.UserName,
			Password: appStoreAppVersion.AppStore.ChartRepo.Password,
		},
		ReleaseIdentifier: &client.ReleaseIdentifier{
			ReleaseNamespace: installAppVersionRequest.Namespace,
			ReleaseName:      installAppVersionRequest.AppName,
		},
	}

	_, err = impl.helmAppService.InstallRelease(ctx, installAppVersionRequest.ClusterId, installReleaseRequest)
	if err != nil {
		return installAppVersionRequest, err
	}

	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentHelmServiceImpl) GetAppStatus(installedAppAndEnvDetails appStoreRepository.InstalledAppAndEnvDetails, w http.ResponseWriter, r *http.Request, token string) (string, error) {

	environment, err := impl.environmentRepository.FindById(installedAppAndEnvDetails.EnvironmentId)
	if err != nil {
		impl.Logger.Errorw("Error in getting environment", "err", err)
		return "", err
	}

	appIdentifier := &client.AppIdentifier{
		ClusterId:   environment.ClusterId,
		Namespace:   environment.Namespace,
		ReleaseName: installedAppAndEnvDetails.AppName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	appDetail, err := impl.helmAppService.GetApplicationDetail(ctx, appIdentifier)
	if err != nil {
		// handling like argocd
		impl.Logger.Errorw("error fetching helm app resource tree", "error", err, "appIdentifier", appIdentifier)
		err = &util.ApiError{
			Code:            constants.AppDetailResourceTreeNotFound,
			InternalMessage: "Failed to get resource tree from helm",
			UserMessage:     "Failed to get resource tree from helm",
		}
		return "", err
	}

	return appDetail.ApplicationStatus, nil
}

func (impl AppStoreDeploymentHelmServiceImpl) DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *appStoreRepository.InstalledApps, dbTransaction *pg.Tx) error {

	appIdentifier := &client.AppIdentifier{
		ClusterId:   installAppVersionRequest.ClusterId,
		ReleaseName: installAppVersionRequest.AppName,
		Namespace:   installAppVersionRequest.Namespace,
	}

	isInstalled, err := impl.helmAppService.IsReleaseInstalled(ctx, appIdentifier)
	if err != nil {
		impl.Logger.Errorw("error in checking if helm release is installed or not", "error", err, "appIdentifier", appIdentifier)
		return err
	}

	if isInstalled {
		deleteResponse, err := impl.helmAppService.DeleteApplication(ctx, appIdentifier)
		if err != nil {
			impl.Logger.Errorw("error in deleting helm application", "error", err, "appIdentifier", appIdentifier)
			return err
		}
		if deleteResponse == nil || !deleteResponse.GetSuccess() {
			return errors.New("delete application response unsuccessful")
		}
	}

	return nil
}

// returns - valuesYamlStr, success, error
func (impl AppStoreDeploymentHelmServiceImpl) RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32) (*appStoreBean.InstallAppVersionDTO, bool, error) {

	// TODO : fetch values yaml from DB instead of fetching from helm cli
	// TODO Dependency : on updating helm APP, DB is not being updated. values yaml is sent directly to helm cli. After DB updatation development, we can fetch values yaml from DB, not from CLI.

	helmAppIdeltifier := &client.AppIdentifier{
		ClusterId:   installedApp.ClusterId,
		Namespace:   installedApp.Namespace,
		ReleaseName: installedApp.AppName,
	}

	helmAppDeploymentDetail, err := impl.helmAppService.GetDeploymentDetail(ctx, helmAppIdeltifier, deploymentVersion)
	if err != nil {
		impl.Logger.Errorw("Error in getting helm application deployment detail", "err", err)
		return installedApp, false, err
	}
	valuesYamlJson := helmAppDeploymentDetail.GetValuesYaml()
	valuesYamlByteArr, err := yaml.JSONToYAML([]byte(valuesYamlJson))
	if err != nil {
		impl.Logger.Errorw("Error in converting json to yaml", "err", err)
		return installedApp, false, err
	}

	// send to helm
	success, err := impl.helmAppService.RollbackRelease(ctx, helmAppIdeltifier, deploymentVersion)
	if err != nil {
		impl.Logger.Errorw("Error in helm rollback release", "err", err)
		return installedApp, false, err
	}
	installedApp.ValuesOverrideYaml = string(valuesYamlByteArr)
	return installedApp, success, nil
}
