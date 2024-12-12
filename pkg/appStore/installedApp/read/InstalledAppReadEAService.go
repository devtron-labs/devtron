package read

import (
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/read/adapter"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/read/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
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
