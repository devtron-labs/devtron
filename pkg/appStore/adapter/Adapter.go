package adapter

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/cluster/repository/bean"
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
// GitHash and FinishedOn are not provided by NewInstallAppVersionHistoryModel; since it is used only on update operations;
// Note: Do not use for update operations;
func NewInstallAppVersionHistoryModel(request *appStoreBean.InstallAppVersionDTO, status string, helmInstallConfigDTO appStoreBean.HelmReleaseStatusConfig) (*repository.InstalledAppVersionHistory, error) {
	installedAppVersions := &repository.InstalledAppVersionHistory{
		InstalledAppVersionId: request.InstalledAppVersionId,
		ValuesYamlRaw:         request.ValuesOverrideYaml,
	}
	helmReleaseStatus, err := getHelmReleaseStatusConfig(helmInstallConfigDTO)
	if err != nil {
		return nil, err
	}
	installedAppVersions.HelmReleaseStatusConfig = helmReleaseStatus
	installedAppVersions.SetStartedOn()
	installedAppVersions.SetStatus(status)
	installedAppVersions.CreateAuditLog(request.UserId)
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

// NewInstalledAppVersionModel will generate a new repository.InstalledAppVersions for the given appStoreBean.InstallAppVersionDTO
func NewInstalledAppVersionModel(request *appStoreBean.InstallAppVersionDTO) *repository.InstalledAppVersions {
	installedAppVersion := &repository.InstalledAppVersions{
		InstalledAppId:               request.InstalledAppId,
		AppStoreApplicationVersionId: request.AppStoreVersion,
		ValuesYaml:                   request.ValuesOverrideYaml,
		ReferenceValueId:             request.ReferenceValueId,
		ReferenceValueKind:           request.ReferenceValueKind,
	}
	installedAppVersion.CreateAuditLog(request.UserId)
	installedAppVersion.MarkActive()
	return installedAppVersion
}

// UpdateInstalledAppVersionModel will update the same repository.InstalledAppVersions model object for the given appStoreBean.InstallAppVersionDTO
func UpdateInstalledAppVersionModel(model *repository.InstalledAppVersions, request *appStoreBean.InstallAppVersionDTO) {
	if model == nil || request == nil {
		return
	}
	model.Id = request.Id
	model.ValuesYaml = request.ValuesOverrideYaml
	model.ReferenceValueId = request.ReferenceValueId
	model.ReferenceValueKind = request.ReferenceValueKind
	model.UpdateAuditLog(request.UserId)
	return
}

// UpdateAdditionalEnvDetails update cluster.EnvironmentBean data into the same InstallAppVersionDTO
func UpdateAdditionalEnvDetails(request *appStoreBean.InstallAppVersionDTO, envBean *bean.EnvironmentBean) {
	if request == nil {
		return
	}
	request.Environment = envBean
	request.EnvironmentName = envBean.Environment
	request.ClusterId = envBean.ClusterId
	request.Namespace = envBean.Namespace
	request.UpdateACDAppName()
}

// UpdateAppDetails update app.App data into the same InstallAppVersionDTO
func UpdateAppDetails(request *appStoreBean.InstallAppVersionDTO, app *app.App) {
	if request == nil {
		return
	}
	request.AppId = app.Id
	request.AppName = app.AppName
	request.TeamId = app.TeamId
	request.AppOfferingMode = app.AppOfferingMode
}

// UpdateInstallAppDetails update repository.InstalledApps data into the same InstallAppVersionDTO
func UpdateInstallAppDetails(request *appStoreBean.InstallAppVersionDTO, installedApp *repository.InstalledApps) {
	if request == nil {
		return
	}
	request.AppId = installedApp.AppId
	request.EnvironmentId = installedApp.EnvironmentId
	request.Status = installedApp.Status
	request.GitOpsRepoName = installedApp.GitOpsRepoName
	request.DeploymentAppType = installedApp.DeploymentAppType
}

// UpdateAppStoreApplicationDetails update appStoreDiscoverRepository.AppStoreApplicationVersion data into the same InstallAppVersionDTO
func UpdateAppStoreApplicationDetails(request *appStoreBean.InstallAppVersionDTO, appStoreApplicationVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) {
	if request == nil {
		return
	}
	request.AppStoreId = appStoreApplicationVersion.AppStoreId
	request.AppStoreName = appStoreApplicationVersion.AppStore.Name
	request.Deprecated = appStoreApplicationVersion.Deprecated
	request.Readme = appStoreApplicationVersion.Readme
}

func getHelmReleaseStatusConfig(helmInstallConfigDTO appStoreBean.HelmReleaseStatusConfig) (string, error) {
	helmInstallConfig, err := json.Marshal(helmInstallConfigDTO)
	if err != nil {
		return "", err
	}
	return string(helmInstallConfig), nil
}
