package adapter

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
)

// NewInstallAppModel is used to generate new repository.InstalledApps model to be saved;
// Note: Do not use for update operations
func NewInstallAppModel(chart *appStoreBean.InstallAppVersionDTO, status appStoreBean.AppstoreDeploymentStatus) *repository.InstalledApps {
	installAppModel := &repository.InstalledApps{
		AppId:             chart.AppId,
		EnvironmentId:     chart.EnvironmentId,
		DeploymentAppType: chart.DeploymentAppType,
	}
	if status != appStoreBean.WF_UNKNOWN {
		installAppModel.UpdateStatus(status)
	}
	installAppModel.CreateAuditLog(chart.UserId)
	installAppModel.UpdateGitOpsRepoName(chart.GitOpsRepoName)
	installAppModel.MarkActive()
	return installAppModel
}

// NewInstallAppVersionsModel is used to generate new repository.InstalledAppVersions model to be saved;
// Note: Do not use for update operations
func NewInstallAppVersionsModel(chart *appStoreBean.InstallAppVersionDTO) *repository.InstalledAppVersions {
	installedAppVersions := &repository.InstalledAppVersions{
		InstalledAppId:               chart.InstalledAppId,
		AppStoreApplicationVersionId: chart.AppStoreVersion,
		ValuesYaml:                   chart.ValuesOverrideYaml,
		ReferenceValueId:             chart.ReferenceValueId,
		ReferenceValueKind:           chart.ReferenceValueKind,
	}
	installedAppVersions.CreateAuditLog(chart.UserId)
	installedAppVersions.MarkActive()
	return installedAppVersions
}

// NewInstallAppVersionHistoryModel is used to generate new repository.InstalledAppVersionHistory model to be saved;
// Note: Do not use for update operations
func NewInstallAppVersionHistoryModel(chart *appStoreBean.InstallAppVersionDTO, status string, helmInstallConfigDTO appStoreBean.HelmReleaseStatusConfig) (*repository.InstalledAppVersionHistory, error) {
	installedAppVersions := &repository.InstalledAppVersionHistory{
		InstalledAppVersionId: chart.InstalledAppVersionId,
		ValuesYamlRaw:         chart.ValuesOverrideYaml,
	}
	helmInstallConfig, err := json.Marshal(helmInstallConfigDTO)
	if err != nil {
		return nil, err
	}
	installedAppVersions.HelmReleaseStatusConfig = string(helmInstallConfig)
	installedAppVersions.SetStartedOn()
	installedAppVersions.SetStatus(status)
	installedAppVersions.CreateAuditLog(chart.UserId)
	return installedAppVersions, nil
}

// NewClusterInstalledAppsModel is used to generate new repository.ClusterInstalledApps model to be saved;
// Note: Do not use for update operations
func NewClusterInstalledAppsModel(chart *appStoreBean.InstallAppVersionDTO, clusterId int) *repository.ClusterInstalledApps {
	clusterInstalledAppsModel := &repository.ClusterInstalledApps{
		ClusterId:      clusterId,
		InstalledAppId: chart.InstalledAppId,
	}
	clusterInstalledAppsModel.CreateAuditLog(chart.UserId)
	return clusterInstalledAppsModel
}

// NewInstalledAppDeploymentAction is used to generate appStoreBean.InstalledAppDeploymentAction from deploymentAppType
func NewInstalledAppDeploymentAction(deploymentAppType string) *appStoreBean.InstalledAppDeploymentAction {
	installedAppDeploymentAction := &appStoreBean.InstalledAppDeploymentAction{}
	switch deploymentAppType {
	case util.PIPELINE_DEPLOYMENT_TYPE_ACD:
		installedAppDeploymentAction.PerformGitOps = true
		installedAppDeploymentAction.PerformACDDeployment = true
		installedAppDeploymentAction.PerformHelmDeployment = false
	case util.PIPELINE_DEPLOYMENT_TYPE_HELM:
		installedAppDeploymentAction.PerformGitOps = false
		installedAppDeploymentAction.PerformACDDeployment = false
		installedAppDeploymentAction.PerformHelmDeployment = true
	case util.PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD:
		installedAppDeploymentAction.PerformGitOps = false
		installedAppDeploymentAction.PerformHelmDeployment = false
		installedAppDeploymentAction.PerformACDDeployment = false
	}
	return installedAppDeploymentAction
}

// GenerateInstallAppVersionDTO converts repository.InstalledApps and repository.InstalledAppVersions db object to appStoreBean.InstallAppVersionDTO bean
func GenerateInstallAppVersionDTO(chart *repository.InstalledApps, installedAppVersion *repository.InstalledAppVersions) *appStoreBean.InstallAppVersionDTO {
	chartVersionApp := installedAppVersion.AppStoreApplicationVersion

	var chartRepoName, chartRepoUrl, Username, Password string
	if chartVersionApp.AppStore.ChartRepoId != 0 {
		chartRepo := chartVersionApp.AppStore.ChartRepo
		chartRepoName = chartRepo.Name
		chartRepoUrl = chartRepo.Url
		Username = chartRepo.UserName
		Password = chartRepo.Password
	} else {
		chartRepo := chartVersionApp.AppStore.DockerArtifactStore
		chartRepoName = chartRepo.Id
		chartRepoUrl = fmt.Sprintf("%s://%s/%s",
			"oci",
			chartVersionApp.AppStore.DockerArtifactStore.RegistryURL,
			chartVersionApp.AppStore.Name)
		Username = chartVersionApp.AppStore.DockerArtifactStore.Username
		Password = chartVersionApp.AppStore.DockerArtifactStore.Password
	}

	return &appStoreBean.InstallAppVersionDTO{
		EnvironmentId:     chart.EnvironmentId,
		AppId:             chart.AppId,
		TeamId:            chart.App.TeamId,
		TeamName:          chart.App.Team.Name,
		AppOfferingMode:   chart.App.AppOfferingMode,
		ClusterId:         chart.Environment.ClusterId,
		Namespace:         chart.Environment.Namespace,
		AppName:           chart.App.AppName,
		EnvironmentName:   chart.Environment.Name,
		InstalledAppId:    chart.Id,
		DeploymentAppType: chart.DeploymentAppType,

		Id:                    installedAppVersion.Id,
		InstalledAppVersionId: installedAppVersion.Id,
		InstallAppVersionChartDTO: &appStoreBean.InstallAppVersionChartDTO{
			AppStoreChartId: chartVersionApp.AppStore.Id,
			ChartName:       chartVersionApp.Name,
			ChartVersion:    chartVersionApp.Version,
			InstallAppVersionChartRepoDTO: &appStoreBean.InstallAppVersionChartRepoDTO{
				RepoName: chartRepoName,
				RepoUrl:  chartRepoUrl,
				UserName: Username,
				Password: Password,
			},
		},
		AppStoreApplicationVersionId: installedAppVersion.AppStoreApplicationVersionId,
	}
}

// GenerateInstallAppVersionMinDTO converts repository.InstalledApps db object to appStoreBean.InstallAppVersionDTO bean;
// Note: It only generates a minimal DTO and doesn't include repository.InstalledAppVersions data
func GenerateInstallAppVersionMinDTO(chart *repository.InstalledApps) *appStoreBean.InstallAppVersionDTO {
	return &appStoreBean.InstallAppVersionDTO{
		EnvironmentId:     chart.EnvironmentId,
		InstalledAppId:    chart.Id,
		AppId:             chart.AppId,
		AppOfferingMode:   chart.App.AppOfferingMode,
		ClusterId:         chart.Environment.ClusterId,
		Namespace:         chart.Environment.Namespace,
		AppName:           chart.App.AppName,
		EnvironmentName:   chart.Environment.Name,
		TeamId:            chart.App.TeamId,
		TeamName:          chart.App.Team.Name,
		DeploymentAppType: chart.DeploymentAppType,
	}
}
