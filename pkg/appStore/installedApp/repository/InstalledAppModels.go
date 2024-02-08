package repository

import (
	"time"
)

// TODO: Remove dependencies on native queries and the structs used in it

// GitOpsAppDetails is used to operate on native query; This should be avoided.
// Usage func: InstalledAppRepository.GetAllGitOpsAppNameAndInstalledAppMapping
type GitOpsAppDetails struct {
	GitOpsAppName  string `sql:"git_ops_app_name"`
	InstalledAppId int    `sql:"installed_app_id"`
}

// InstalledAppsWithChartDetails is used to operate on native query; This should be avoided.
// Usage func: InstalledAppRepository.GetAllInstalledApps
type InstalledAppsWithChartDetails struct {
	AppStoreApplicationName      string    `json:"app_store_application_name"`
	ChartRepoName                string    `json:"chart_repo_name"`
	DockerArtifactStoreId        string    `json:"docker_artifact_store_id"`
	AppName                      string    `json:"app_name"`
	EnvironmentName              string    `json:"environment_name"`
	InstalledAppVersionId        int       `json:"installed_app_version_id"`
	AppStoreApplicationVersionId int       `json:"app_store_application_version_id"`
	Icon                         string    `json:"icon"`
	Readme                       string    `json:"readme"`
	CreatedOn                    time.Time `json:"created_on"`
	UpdatedOn                    time.Time `json:"updated_on"`
	Id                           int       `json:"id"`
	EnvironmentId                int       `json:"environment_id"`
	Deprecated                   bool      `json:"deprecated"`
	ClusterName                  string    `json:"clusterName"`
	Namespace                    string    `json:"namespace"`
	TeamId                       int       `json:"teamId"`
	ClusterId                    int       `json:"clusterId"`
	AppOfferingMode              string    `json:"app_offering_mode"`
	AppStatus                    string    `json:"app_status"`
	DeploymentAppDeleteRequest   bool      `json:"deploymentAppDeleteRequest"`
}

// InstalledAppAndEnvDetails is used to operate on native query; This should be avoided.
// Usage functions: InstalledAppRepository.GetAllInstalledAppsByChartRepoId and InstalledAppRepository.GetAllInstalledAppsByAppStoreId
type InstalledAppAndEnvDetails struct {
	EnvironmentName              string    `json:"environment_name"`
	EnvironmentId                int       `json:"environment_id"`
	AppName                      string    `json:"app_name"`
	AppOfferingMode              string    `json:"appOfferingMode"`
	UpdatedOn                    time.Time `json:"updated_on"`
	EmailId                      string    `json:"email_id"`
	InstalledAppVersionId        int       `json:"installed_app_version_id"`
	AppId                        int       `json:"app_id"`
	InstalledAppId               int       `json:"installed_app_id"`
	AppStoreApplicationVersionId int       `json:"app_store_application_version_id"`
	AppStatus                    string    `json:"app_status"`
	DeploymentAppType            string    `json:"-"`
}

// InstallAppDeleteRequest is used to operate on native query; This should be avoided.
// Usage func: InstalledAppRepository.GetInstalledAppByGitHash
type InstallAppDeleteRequest struct {
	InstalledAppId  int    `json:"installed_app_id,omitempty,notnull"`
	AppName         string `json:"app_name,omitempty"`
	AppId           int    `json:"app_id,omitempty"`
	EnvironmentId   int    `json:"environment_id,omitempty"`
	AppOfferingMode string `json:"app_offering_mode"`
	ClusterId       int    `json:"cluster_id"`
	Namespace       string `json:"namespace"`
}
