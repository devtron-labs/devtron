package adapter

import (
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/read/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
)

// GetInstalledAppInternal will return the installed app with environment and cluster details.
//   - input: installedAppModel *repository.InstalledApps
//   - returns: *bean.InstalledAppWithEnvAndClusterDetails
func GetInstalledAppInternal(installedAppModel *repository.InstalledApps) *bean.InstalledAppWithEnvAndClusterDetails {
	if installedAppModel == nil {
		return nil
	}
	installedAppInternal := &bean.InstalledAppWithEnvAndClusterDetails{}
	installedAppInternal.InstalledAppMin = GetInstalledAppMin(installedAppModel)
	// Extra App details
	if !installedAppModel.App.IsEmpty() {
		installedAppInternal.AppName = installedAppModel.App.AppName
		installedAppInternal.AppOfferingMode = installedAppModel.App.AppOfferingMode
		installedAppInternal.TeamId = installedAppModel.App.TeamId
	}
	// Extra Environment details
	if !installedAppModel.Environment.IsEmpty() {
		installedAppInternal.EnvironmentName = installedAppModel.Environment.Name
		installedAppInternal.EnvironmentIdentifier = installedAppModel.Environment.EnvironmentIdentifier
		installedAppInternal.Namespace = installedAppModel.Environment.Namespace
		installedAppInternal.ClusterId = installedAppModel.Environment.ClusterId
	}
	// Cluster details
	if !installedAppModel.Environment.Cluster.IsEmpty() {
		installedAppInternal.ClusterName = installedAppModel.Environment.Cluster.ClusterName
	}
	return installedAppInternal
}

// GetInstalledAppMin will return the installed app minimum details.
//   - input: installedAppModel *repository.InstalledApps
//   - returns: *bean.InstalledAppMin
func GetInstalledAppMin(installedAppModel *repository.InstalledApps) *bean.InstalledAppMin {
	if installedAppModel == nil {
		return nil
	}
	installedAppMin := &bean.InstalledAppMin{
		// Installed App details
		Id:                         installedAppModel.Id,
		Active:                     installedAppModel.Active,
		GitOpsRepoName:             installedAppModel.GitOpsRepoName,
		GitOpsRepoUrl:              installedAppModel.GitOpsRepoUrl,
		IsCustomRepository:         installedAppModel.IsCustomRepository,
		DeploymentAppType:          installedAppModel.DeploymentAppType,
		DeploymentAppDeleteRequest: installedAppModel.DeploymentAppDeleteRequest,
		AppId:                      installedAppModel.AppId,
		EnvironmentId:              installedAppModel.EnvironmentId,
	}
	return installedAppMin
}

// GetInstalledAppDeleteRequest will return the installed app delete request.
//   - input: installedAppModel *repository.InstallAppDeleteRequest
//   - returns: *bean.InstalledAppDeleteRequest
func GetInstalledAppDeleteRequest(installedAppModel *repository.InstallAppDeleteRequest) *bean.InstalledAppDeleteRequest {
	if installedAppModel == nil {
		return nil
	}
	return &bean.InstalledAppDeleteRequest{
		InstalledAppId:  installedAppModel.InstalledAppId,
		AppName:         installedAppModel.AppName,
		AppId:           installedAppModel.AppId,
		EnvironmentId:   installedAppModel.EnvironmentId,
		AppOfferingMode: installedAppModel.AppOfferingMode,
		ClusterId:       installedAppModel.ClusterId,
		Namespace:       installedAppModel.Namespace,
	}
}

// GetInstalledAppVersionMin will return the installed app version minimum details.
//   - input: installedAppVersionModel *repository.InstalledAppVersions
//   - returns: *bean.InstalledAppVersionMin
func GetInstalledAppVersionMin(installedAppVersionModel *repository.InstalledAppVersions) *bean.InstalledAppVersionMin {
	if installedAppVersionModel == nil {
		return nil
	}
	return &bean.InstalledAppVersionMin{
		Id:                           installedAppVersionModel.Id,
		InstalledAppId:               installedAppVersionModel.InstalledAppId,
		AppStoreApplicationVersionId: installedAppVersionModel.AppStoreApplicationVersionId,
		ValuesYaml:                   installedAppVersionModel.ValuesYaml,
		Active:                       installedAppVersionModel.Active,
		ReferenceValueId:             installedAppVersionModel.ReferenceValueId,
		ReferenceValueKind:           installedAppVersionModel.ReferenceValueKind,
	}
}

// GetInstalledAppVersionWithAppStoreDetails will return the installed app version with app store details.
//   - input: installedAppVersionModel *repository.InstalledAppVersions
//   - returns: *bean.InstalledAppVersionWithAppStoreDetails
func GetInstalledAppVersionWithAppStoreDetails(installedAppVersionModel *repository.InstalledAppVersions) *bean.InstalledAppVersionWithAppStoreDetails {
	if installedAppVersionModel == nil {
		return nil
	}
	versionDetails := &bean.InstalledAppVersionWithAppStoreDetails{
		InstalledAppVersionMin: GetInstalledAppVersionMin(installedAppVersionModel),
	}
	// Extra App Store Application Version details
	if !installedAppVersionModel.AppStoreApplicationVersion.IsEmpty() {
		versionDetails.AppStoreVersion = installedAppVersionModel.AppStoreApplicationVersion.Version
	}
	return versionDetails
}
