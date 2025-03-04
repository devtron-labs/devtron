package read

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/read/adapter"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/read/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/bean"
	"go.uber.org/zap"
)

type InstalledAppReadServiceEA interface {
	// GetDeploymentSuccessfulStatusCountForTelemetry will return the count of successful deployments for telemetry
	// Total successful deployments are calculated till now.
	GetDeploymentSuccessfulStatusCountForTelemetry() (int, error)
	// GetInstalledAppByAppName will return the installed app with environment details by app name.
	// Additional details like environment details are also fetched.
	// Refer bean.InstalledAppWithEnvDetails for more details.
	GetInstalledAppByAppName(appName string) (*bean.InstalledAppWithEnvDetails, error)
	// GetInstalledAppsByAppId will return the installed app by app id
	// Only the minimum details are fetched.
	// Refer bean.InstalledAppMin for more details.
	GetInstalledAppsByAppId(appId int) (*bean.InstalledAppMin, error)
	// GetInstalledAppByAppIdAndDeploymentType will return the installed app by app id and deployment type.
	// Only the minimum details are fetched.
	// Refer bean.InstalledAppMin for more details.
	GetInstalledAppByAppIdAndDeploymentType(appId int, deploymentAppType string) (*bean.InstalledAppMin, error)
	// GetInstalledAppByInstalledAppVersionId will return the installed app by installed app version id.
	// Only the minimum details are fetched.
	// Refer bean.InstalledAppMin for more details.
	GetInstalledAppByInstalledAppVersionId(installedAppVersionId int) (*bean.InstalledAppMin, error)
	// GetInstalledAppVersionIncludingDeleted will return the installed app version by installed app version id.
	// Both active and deleted installed app versions are fetched.
	// Additional details like app store details are also fetched.
	// Refer bean.InstalledAppVersionWithAppStoreDetails for more details.
	GetInstalledAppVersionIncludingDeleted(installedAppVersionId int) (*bean.InstalledAppVersionWithAppStoreDetails, error)
	GetAllArgoAppNamesByCluster(clusterId []int) ([]bean2.DeployedInstalledAppInfo, error)
	// IsChartStoreAppManagedByArgoCd returns if a chart store app is deployed via argo-cd or not
	IsChartStoreAppManagedByArgoCd(appId int) (bool, error)
}

type InstalledAppReadServiceEAImpl struct {
	logger                 *zap.SugaredLogger
	installedAppRepository repository.InstalledAppRepository
}

func NewInstalledAppReadServiceEAImpl(
	logger *zap.SugaredLogger,
	installedAppRepository repository.InstalledAppRepository,
) *InstalledAppReadServiceEAImpl {
	return &InstalledAppReadServiceEAImpl{
		logger:                 logger,
		installedAppRepository: installedAppRepository,
	}
}

func (impl *InstalledAppReadServiceEAImpl) GetDeploymentSuccessfulStatusCountForTelemetry() (int, error) {
	return impl.installedAppRepository.GetDeploymentSuccessfulStatusCountForTelemetry()
}

func (impl *InstalledAppReadServiceEAImpl) GetInstalledAppByAppName(appName string) (*bean.InstalledAppWithEnvDetails, error) {
	installedAppModel, err := impl.installedAppRepository.GetInstalledAppByAppName(appName)
	if err != nil {
		impl.logger.Errorw("error while fetching installed app by app name", "appName", appName, "error", err)
		return nil, err
	}
	return adapter.GetInstalledAppInternal(installedAppModel).GetInstalledAppWithEnvDetails(), nil
}

func (impl *InstalledAppReadServiceEAImpl) GetInstalledAppsByAppId(appId int) (*bean.InstalledAppMin, error) {
	installedAppModel, err := impl.installedAppRepository.GetInstalledAppsMinByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error while fetching installed apps by app id", "appId", appId, "error", err)
		return nil, err
	}
	return adapter.GetInstalledAppInternal(installedAppModel).GetInstalledAppMin(), nil
}

func (impl *InstalledAppReadServiceEAImpl) GetInstalledAppByAppIdAndDeploymentType(appId int, deploymentAppType string) (*bean.InstalledAppMin, error) {
	installedAppModel, err := impl.installedAppRepository.GetInstalledAppByAppIdAndDeploymentType(appId, deploymentAppType)
	if err != nil {
		impl.logger.Errorw("error while fetching installed app by app id and deployment type", "appId", appId, "deploymentAppType", deploymentAppType, "error", err)
		return nil, err
	}
	return adapter.GetInstalledAppInternal(installedAppModel).GetInstalledAppMin(), nil
}

func (impl *InstalledAppReadServiceEAImpl) GetInstalledAppByInstalledAppVersionId(installedAppVersionId int) (*bean.InstalledAppMin, error) {
	installedAppModel, err := impl.installedAppRepository.GetInstalledAppByInstalledAppVersionId(installedAppVersionId)
	if err != nil {
		impl.logger.Errorw("error while fetching installed app by installed app version id", "installedAppVersionId", installedAppVersionId, "error", err)
		return nil, err
	}
	return adapter.GetInstalledAppInternal(installedAppModel).GetInstalledAppMin(), nil
}

func (impl *InstalledAppReadServiceEAImpl) GetInstalledAppVersionIncludingDeleted(installedAppVersionId int) (*bean.InstalledAppVersionWithAppStoreDetails, error) {
	installedAppVersionModel, err := impl.installedAppRepository.GetInstalledAppVersionIncludingDeleted(installedAppVersionId)
	if err != nil {
		impl.logger.Errorw("error while fetching installed app version by installed app version id", "installedAppVersionId", installedAppVersionId, "error", err)
		return nil, err
	}
	return adapter.GetInstalledAppVersionWithAppStoreDetails(installedAppVersionModel), nil
}

func (impl *InstalledAppReadServiceEAImpl) GetAllArgoAppNamesByCluster(clusterId []int) ([]bean2.DeployedInstalledAppInfo, error) {
	return impl.installedAppRepository.GetAllAppsByClusterAndDeploymentAppType(clusterId, util.PIPELINE_DEPLOYMENT_TYPE_ACD)
}

func (impl *InstalledAppReadServiceEAImpl) IsChartStoreAppManagedByArgoCd(appId int) (bool, error) {
	installedAppModel, err := impl.installedAppRepository.GetInstalledAppsMinByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error while fetching installed apps by app id", "appId", appId, "error", err)
		return false, err
	}
	return util.IsAcdApp(installedAppModel.DeploymentAppType), nil
}
