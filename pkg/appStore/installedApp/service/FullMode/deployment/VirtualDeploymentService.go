package deployment

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/bean"
)

type InstalledAppVirtualDeploymentService interface {
	GetChartBytesForLatestDeployment(installedAppId int, installedAppVersionId int) ([]byte, error)
	GetChartBytesForParticularDeployment(installedAppId int, installedAppVersionId int, installedAppVersionHistoryId int) ([]byte, error)
}

func (impl *FullModeDeploymentServiceImpl) GetChartBytesForLatestDeployment(installedAppId int, installedAppVersionId int) ([]byte, error) {

	chartBytes := make([]byte, 0)

	installedApp, err := impl.installedAppRepository.GetInstalledApp(installedAppId)
	if err != nil {
		impl.Logger.Errorw("error in fetching installed app", "err", err, "installed_app_id", installedAppId)
		return chartBytes, err
	}
	installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersion(installedAppVersionId)
	if err != nil {
		impl.Logger.Errorw("Service err, BuildChartWithValuesAndRequirementsConfig", err, "installed_app_version_id", installedAppVersionId)
		return chartBytes, err
	}

	valuesString, err := impl.appStoreDeploymentCommonService.GetValuesString(installedAppVersion.AppStoreApplicationVersion.AppStore.Name, installedAppVersion.ValuesYaml)
	if err != nil {
		return chartBytes, err
	}
	requirementsString, err := impl.appStoreDeploymentCommonService.GetRequirementsString(installedAppVersion.AppStoreApplicationVersionId)
	if err != nil {
		return chartBytes, err
	}

	updateTime := installedApp.UpdatedOn
	timeStampTag := updateTime.Format(bean.LayoutDDMMYY_HHMM12hr)
	chartName := fmt.Sprintf("%s-%s-%s (GMT)", installedApp.App.AppName, installedApp.Environment.Name, timeStampTag)
	chartBytes, err = impl.appStoreDeploymentCommonService.BuildChartWithValuesAndRequirementsConfig(installedApp.App.AppName, valuesString, requirementsString, chartName, fmt.Sprint(installedApp.Id))

	if err != nil {
		return chartBytes, err
	}
	return chartBytes, nil
}

func (impl *FullModeDeploymentServiceImpl) GetChartBytesForParticularDeployment(installedAppId int, installedAppVersionId int, installedAppVersionHistoryId int) ([]byte, error) {

	chartBytes := make([]byte, 0)

	installedApp, err := impl.installedAppRepository.GetInstalledApp(installedAppId)
	if err != nil {
		impl.Logger.Errorw("error in fetching installed app", "err", err, "installed_app_id", installedAppId)
		return chartBytes, err
	}
	installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersionAny(installedAppVersionId)
	if err != nil {
		impl.Logger.Errorw("Service err, BuildChartWithValuesAndRequirementsConfig", err, "installed_app_version_id", installedAppVersionId)
		return chartBytes, err
	}
	installedAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(installedAppVersionHistoryId)

	valuesString, err := impl.appStoreDeploymentCommonService.GetValuesString(installedAppVersion.AppStoreApplicationVersion.AppStore.Name, installedAppVersionHistory.ValuesYamlRaw)
	if err != nil {
		return chartBytes, err
	}
	requirementsString, err := impl.appStoreDeploymentCommonService.GetRequirementsString(installedAppVersion.AppStoreApplicationVersionId)
	if err != nil {
		return chartBytes, err
	}

	updateTime := installedApp.UpdatedOn
	timeStampTag := updateTime.Format(bean.LayoutDDMMYY_HHMM12hr)
	chartName := fmt.Sprintf("%s-%s-%s (GMT)", installedApp.App.AppName, installedApp.Environment.Name, timeStampTag)

	chartBytes, err = impl.appStoreDeploymentCommonService.BuildChartWithValuesAndRequirementsConfig(installedApp.App.AppName, valuesString, requirementsString, chartName, fmt.Sprint(installedApp.Id))
	if err != nil {
		return chartBytes, err
	}
	return chartBytes, nil
}
