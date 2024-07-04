/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package appStoreBean

import (
	"encoding/json"
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/pkg/cluster/repository/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/util/gitUtil"
	"time"
)

// v1
type InstallAppVersionHistoryDto struct {
	InstalledAppInfo *InstalledAppDto `json:"installedAppInfo"`
	IAVHistory       []*IAVHistory    `json:"deploymentHistory"`
}
type IAVHistory struct {
	ChartMetaData         IAVHistoryChartMetaData `json:"chartMetadata"`
	DeployedAt            IAVHistoryDeployedAt    `json:"deployedAt"`
	DockerImages          []string                `json:"dockerImages"`
	Version               int                     `json:"version"`
	InstalledAppVersionId int                     `json:"installedAppVersionId"`
}
type IAVHistoryChartMetaData struct {
	ChartName    string   `json:"chartName"`
	ChartVersion string   `json:"chartVersion"`
	Description  string   `json:"description"`
	Home         string   `json:"home"`
	Sources      []string `json:"sources"`
}

type IAVHistoryDeployedAt struct {
	Nanos   int   `json:"nanos,omitempty"`
	Seconds int64 `json:"seconds,omitempty"`
}

type IAVHistoryValues struct {
	Manifest   string `json:"manifest"`
	ValuesYaml string `json:"valuesYaml"`
}

type InstalledAppDto struct {
	AppId           int    `json:"appId"`
	InstalledAppId  int    `json:"installedAppId"`
	EnvironmentName string `json:"environmentName"`
	AppOfferingMode string `json:"appOfferingMode"`
	ClusterId       int    `json:"clusterId"`
	EnvironmentId   int    `json:"environmentId"`
}

type InstallAppVersionRequestType int

const (
	INSTALL_APP_REQUEST InstallAppVersionRequestType = iota
	BULK_DEPLOY_REQUEST
	DEFAULT_COMPONENT_DEPLOYMENT_REQUEST
)

type InstallAppVersionDTO struct {
	Id                           int                            `json:"id,omitempty"` // TODO: redundant data; refers to InstalledAppVersionId
	AppId                        int                            `json:"appId,omitempty"`
	AppName                      string                         `json:"appName,omitempty"` // AppName can be display_name in case of external-apps (which is not unique in that case)
	TeamId                       int                            `json:"teamId,omitempty"`
	TeamName                     string                         `json:"teamName,omitempty"`
	EnvironmentId                int                            `json:"environmentId,omitempty"`
	InstalledAppId               int                            `json:"installedAppId,omitempty,notnull"`
	InstalledAppVersionId        int                            `json:"installedAppVersionId,omitempty,notnull"`
	InstalledAppVersionHistoryId int                            `json:"installedAppVersionHistoryId,omitempty"`
	AppStoreVersion              int                            `json:"appStoreVersion,omitempty,notnull"`
	ValuesOverrideYaml           string                         `json:"valuesOverrideYaml,omitempty"`
	Readme                       string                         `json:"readme,omitempty"`
	ReferenceValueId             int                            `json:"referenceValueId, omitempty" validate:"required,number"`                            // TODO: ineffective usage of omitempty; can be removed
	ReferenceValueKind           string                         `json:"referenceValueKind, omitempty" validate:"oneof=DEFAULT TEMPLATE DEPLOYED EXISTING"` // TODO: ineffective usage of omitempty; can be removed
	AppStoreId                   int                            `json:"appStoreId"`
	AppStoreName                 string                         `json:"appStoreName"`
	Deprecated                   bool                           `json:"deprecated"`
	ClusterId                    int                            `json:"clusterId"` // needed for hyperion mode
	Namespace                    string                         `json:"namespace"` // needed for hyperion mode
	AppOfferingMode              string                         `json:"appOfferingMode"`
	GitOpsPath                   string                         `json:"gitOpsPath"`
	GitHash                      string                         `json:"gitHash"`
	DeploymentAppType            string                         `json:"deploymentAppType"` // TODO: instead of string, use enum
	AcdPartialDelete             bool                           `json:"acdPartialDelete"`
	InstalledAppDeleteResponse   *InstalledAppDeleteResponseDTO `json:"deleteResponse,omitempty"`
	UpdatedOn                    time.Time                      `json:"updatedOn"`
	IsVirtualEnvironment         bool                           `json:"isVirtualEnvironment"`
	HelmPackageName              string                         `json:"helmPackageName"`
	GitOpsRepoURL                string                         `json:"gitRepoURL"`
	IsCustomRepository           bool                           `json:"-"`
	IsNewGitOpsRepo              bool                           `json:"-"`
	ACDAppName                   string                         `json:"-"`
	Environment                  *bean.EnvironmentBean          `json:"-"`
	ChartGroupEntryId            int                            `json:"-"`
	DefaultClusterComponent      bool                           `json:"-"`
	Status                       AppstoreDeploymentStatus       `json:"-"`
	UserId                       int32                          `json:"-"`
	ForceDelete                  bool                           `json:"-"`
	NonCascadeDelete             bool                           `json:"-"`
	EnvironmentName              string                         `json:"-"`
	InstallAppVersionChartDTO    *InstallAppVersionChartDTO     `json:"-"`
	AppStoreApplicationVersionId int
	DisplayName                  string `json:"-"` // used only for external apps
}

// UpdateDeploymentAppType updates deploymentAppType to InstallAppVersionDTO
func (chart *InstallAppVersionDTO) UpdateDeploymentAppType(deploymentAppType string) {
	if chart == nil {
		return
	}
	chart.DeploymentAppType = deploymentAppType
}

// UpdateACDAppName updates ArgoCd app object name to InstallAppVersionDTO
func (chart *InstallAppVersionDTO) UpdateACDAppName() {
	if chart == nil {
		return
	}
	chart.ACDAppName = fmt.Sprintf("%s-%s", chart.AppName, chart.EnvironmentName)
}

func (chart *InstallAppVersionDTO) UpdateCustomGitOpsRepoUrl(allowCustomRepository bool, installAppVersionRequestType InstallAppVersionRequestType) {
	// Handling for chart-group deployment request
	if allowCustomRepository && len(chart.GitOpsRepoURL) == 0 &&
		(installAppVersionRequestType == BULK_DEPLOY_REQUEST || installAppVersionRequestType == DEFAULT_COMPONENT_DEPLOYMENT_REQUEST) {
		chart.GitOpsRepoURL = apiBean.GIT_REPO_DEFAULT
	}
}

// InstalledAppDeploymentAction is an internal struct for Helm App deployment; used to decide the deployment steps to be performed
type InstalledAppDeploymentAction struct {
	PerformGitOpsForHelmApp bool
	PerformGitOps           bool
	PerformACDDeployment    bool
	PerformHelmDeployment   bool
}

type InstalledAppDeleteResponseDTO struct {
	DeleteInitiated  bool   `json:"deleteInitiated"`
	ClusterReachable bool   `json:"clusterReachable"`
	ClusterName      string `json:"clusterName"`
}

type InstallAppVersionChartDTO struct {
	AppStoreChartId               int                            `json:"-"`
	ChartName                     string                         `json:"-"`
	ChartVersion                  string                         `json:"-"`
	InstallAppVersionChartRepoDTO *InstallAppVersionChartRepoDTO `json:"-"`
}

type InstallAppVersionChartRepoDTO struct {
	RepoName string `json:"-"`
	RepoUrl  string `json:"-"`
	UserName string `json:"-"`
	Password string `json:"-"`
}

func (chart *InstallAppVersionDTO) GetDeploymentConfig() *bean2.DeploymentConfig {
	var configType string
	if chart.IsCustomRepository {
		configType = bean2.CUSTOM.String()
	} else {
		configType = bean2.SYSTEM_GENERATED.String()
	}
	return &bean2.DeploymentConfig{
		AppId:             chart.AppId,
		EnvironmentId:     chart.EnvironmentId,
		ConfigType:        configType,
		DeploymentAppType: chart.DeploymentAppType,
		RepoURL:           chart.GitOpsRepoURL,
		RepoName:          gitUtil.GetGitRepoNameFromGitRepoUrl(chart.GitOpsRepoURL),
		Active:            true,
	}
}

// /
type RefChartProxyDir string

const (
	RefChartProxyDirPath = "scripts/devtron-reference-helm-charts"
)

var CHART_PROXY_TEMPLATE = "reference-chart-proxy"
var REQUIREMENTS_YAML_FILE = "requirements.yaml"
var VALUES_YAML_FILE = "values.yaml"

type InstalledAppsResponse struct {
	AppStoreApplicationName      string    `json:"appStoreApplicationName"`
	ChartName                    string    `json:"chartName"`
	Icon                         string    `json:"icon"`
	Status                       string    `json:"status"`
	AppName                      string    `json:"appName"`
	InstalledAppVersionId        int       `json:"installedAppVersionId"`
	AppStoreApplicationVersionId int       `json:"appStoreApplicationVersionId"`
	EnvironmentName              string    `json:"environmentName"`
	DeployedAt                   time.Time `json:"deployedAt"`
	DeployedBy                   string    `json:"deployedBy"`
	DeploymentAppType            string    `json:"deploymentAppType,omitempty"` // TODO: instead of string, use enum
	InstalledAppsId              int       `json:"installedAppId"`
	Readme                       string    `json:"readme"`
	EnvironmentId                int       `json:"environmentId"`
	Deprecated                   bool      `json:"deprecated"`
	AppOfferingMode              string    `json:"appOfferingMode" validate:"oneof=EA_ONLY FULL"`
	ClusterId                    int       `json:"clusterId"` // needed for hyperion app
	Namespace                    string    `json:"namespace"` // needed for hyperion app
}

type AppNames struct {
	Name          string `json:"name,omitempty"`
	Exists        bool   `json:"exists"`
	SuggestedName string `json:"suggestedName,omitempty"`
}

type Dependencies struct {
	Dependencies []Dependency `json:"dependencies"`
}
type Dependency struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Repository string `json:"repository"`
}

const REFERENCE_TYPE_DEFAULT string = "DEFAULT"
const REFERENCE_TYPE_TEMPLATE string = "TEMPLATE"
const REFERENCE_TYPE_DEPLOYED string = "DEPLOYED"
const REFERENCE_TYPE_EXISTING string = "EXISTING"

type AppStoreVersionValuesDTO struct {
	Id                 int       `json:"id,omitempty"`
	AppStoreVersionId  int       `json:"appStoreVersionId,omitempty,notnull"`
	Name               string    `json:"name,omitempty"`
	Values             string    `json:"values,omitempty"` //yaml format user value
	ChartVersion       string    `json:"chartVersion,omitempty"`
	EnvironmentName    string    `json:"environmentName,omitempty"`
	Description        string    `json:"description,omitempty"`
	UpdatedByUserEmail string    `json:"updatedBy,omitempty"`
	UpdatedByUserId    int32     `json:"-"`
	UpdatedOn          time.Time `json:"updatedOn"`
	UserId             int32     `json:"-"`
}

type AppStoreVersionValuesCategoryWiseDTO struct {
	Values []*AppStoreVersionValuesDTO `json:"values"`
	Kind   string                      `json:"kind"`
}

type AppSotoreVersionDTOWrapper struct {
	Values []*AppStoreVersionValuesCategoryWiseDTO `json:"values"`
}

type ValuesListCategory struct {
	Id                int             `json:"id,omitempty"`
	AppStoreVersionId int             `json:"appStoreVersionId,omitempty,notnull"`
	ReferenceId       int             `json:"referenceId,omitempty,notnull"`
	Name              string          `json:"name,omitempty"`
	ValuesOverride    json.RawMessage `json:"valuesOverride,omitempty"` //json format user value
}

type ValuesCategoryResponse struct {
	ReferenceType      json.RawMessage      `json:"referenceType,omitempty"` //json format user value
	ValuesListCategory []ValuesListCategory `json:"valuesListCategory,omitempty"`
}

type AppStoreApplication struct {
	Id                          int                                   `json:"id"`
	Name                        string                                `json:"name"`
	ChartRepoId                 int                                   `json:"chartRepoId"`
	Active                      bool                                  `json:"active"`
	ChartGitLocation            string                                `json:"chartGitLocation"`
	CreatedOn                   time.Time                             `json:"createdOn"`
	UpdatedOn                   time.Time                             `json:"updatedOn"`
	AppStoreApplicationVersions []*AppStoreApplicationVersionResponse `json:"appStoreApplicationVersions"`
}

type AppStoreApplicationVersionResponse struct {
	Id                      int       `json:"id"`
	Version                 string    `json:"version"`
	AppVersion              string    `json:"appVersion"`
	Created                 time.Time `json:"created"`
	Deprecated              bool      `json:"deprecated"`
	Description             string    `json:"description"`
	Digest                  string    `json:"digest"`
	Icon                    string    `json:"icon"`
	Name                    string    `json:"name"`
	ChartName               string    `json:"chartName"`
	AppStoreApplicationName string    `json:"appStoreApplicationName"`
	Home                    string    `json:"home"`
	Source                  string    `json:"source"`
	ValuesYaml              string    `json:"valuesYaml"`
	ChartYaml               string    `json:"chartYaml"`
	AppStoreId              int       `json:"appStoreId"`
	Latest                  bool      `json:"latest"`
	CreatedOn               time.Time `json:"createdOn"`
	RawValues               string    `json:"rawValues"`
	Readme                  string    `json:"readme"`
	ValuesSchemaJson        string    `json:"valuesSchemaJson"`
	Notes                   string    `json:"notes"`
	UpdatedOn               time.Time `json:"updatedOn"`
	IsChartRepoActive       bool      `json:"isChartRepoActive"`
	IsOCICompliantChart     bool      `json:"isOCICompliantChart"`
}

type AppStoreVersionsResponse struct {
	Version string `json:"version"`
	Id      int    `json:"id"`
}

type ChartInfoRes struct {
	AppStoreApplicationVersionId int    `json:"appStoreApplicationVersionId"`
	Readme                       string `json:"readme"`
	ValuesSchemaJson             string `json:"valuesSchemaJson"`
	Notes                        string `json:"notes"`
}

type AppStoreWithVersion struct {
	Id                           int       `json:"id"`
	AppStoreApplicationVersionId int       `json:"appStoreApplicationVersionId"`
	Name                         string    `json:"name"`
	ChartRepoId                  int       `json:"chart_repo_id"`
	DockerArtifactStoreId        string    `json:"docker_artifact_store_id"`
	ChartName                    string    `json:"chart_name"`
	Icon                         string    `json:"icon"`
	Active                       bool      `json:"active"`
	ChartGitLocation             string    `json:"chart_git_location"`
	CreatedOn                    time.Time `json:"created_on"`
	UpdatedOn                    time.Time `json:"updated_on"`
	Version                      string    `json:"version"`
	Deprecated                   bool      `json:"deprecated"`
	Description                  string    `json:"description"`
}

type AppStoreFilter struct {
	ChartRepoId       []int    `json:"chartRepoId"`
	RegistryId        []string `json:"registryId"`
	AppStoreName      string   `json:"appStoreName"`
	AppName           string   `json:"appName"`
	IncludeDeprecated bool     `json:"includeDeprecated"`
	Offset            int      `json:"offset"`
	Size              int      `json:"size"`
	EnvIds            []int    `json:"envIds"`
	OnlyDeprecated    bool     `json:"onlyDeprecated"`
	ClusterIds        []int    `json:"clusterIds"`
	AppStatuses       []string `json:"appStatuses"`
}

type ChartRepoSearch struct {
	AppStoreApplicationVersionId int    `json:"appStoreApplicationVersionId"`
	ChartId                      int    `json:"chartId"`
	ChartName                    string `json:"chartName"`
	ChartRepoId                  int    `json:"chartRepoId"`
	ChartRepoName                string `json:"chartRepoName"`
	Version                      string `json:"version"`
	Deprecated                   bool   `json:"deprecated"`
}

type UpdateProjectHelmAppDTO struct {
	AppId          string `json:"appId"`
	InstalledAppId int    `json:"installedAppId"`
	AppName        string `json:"appName"`
	TeamId         int    `json:"teamId"`
	UserId         int32  `json:"userId"`
}

type AppstoreDeploymentStatus int

const (
	WF_UNKNOWN AppstoreDeploymentStatus = iota
	REQUEST_ACCEPTED
	ENQUEUED
	QUE_ERROR
	DEQUE_ERROR
	TRIGGER_ERROR
	DEPLOY_SUCCESS
	DEPLOY_INIT
	GIT_ERROR
	GIT_SUCCESS
	ACD_ERROR
	ACD_SUCCESS
	HELM_ERROR
	HELM_SUCCESS
)

func (a AppstoreDeploymentStatus) String() string {
	return [...]string{"WF_UNKNOWN", "REQUEST_ACCEPTED", "ENQUEUED", "QUE_ERROR", "DEQUE_ERROR", "TRIGGER_ERROR", "DEPLOY_SUCCESS", "DEPLOY_INIT", "GIT_ERROR", "GIT_SUCCESS", "ACD_ERROR", "ACD_SUCCESS", "HELM_ERROR",
		"HELM_SUCCESS"}[a]
}

type HelmReleaseStatusConfig struct {
	InstallAppVersionHistoryId int
	Message                    string
	IsReleaseInstalled         bool
	ErrorInInstallation        bool
}

type ChartComponents struct {
	ChartComponent []*ChartComponent `json:"charts"`
}

type ChartComponent struct {
	Name   string `json:"name"`
	Values string `json:"values"`
}

const (
	DEFAULT_CLUSTER_ID                          = 1
	DEFAULT_NAMESPACE                           = "default"
	DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT = "devtron"
	CLUSTER_COMPONENT_DIR_PATH                  = "/cluster/component"
	HELM_RELEASE_STATUS_FAILED                  = "Failed"
	HELM_RELEASE_STATUS_PROGRESSING             = "Progressing"
	HELM_RELEASE_STATUS_UNKNOWN                 = "Unknown"
	FAILED_TO_REGISTER_IN_ACD_ERROR             = "failed to register app on ACD with error: "
	FAILED_TO_DELETE_APP_PREFIX_ERROR           = "error deleting app with error: "
	COULD_NOT_FETCH_APP_NAME_AND_ENV_NAME_ERR   = "could not fetch app name or environment name"
	APP_NOT_DELETED_YET_ERROR                   = "App Not Yet Deleted."
)

type EnvironmentDetails struct {
	EnvironmentName *string `json:"environmentName,omitempty"`
	// id in which app is deployed
	EnvironmentId *int32 `json:"environmentId,omitempty"`
	// namespace corresponding to the environemnt
	Namespace *string `json:"namespace,omitempty"`
	// if given environemnt is marked as production or not, nullable
	IsPrduction *bool `json:"isPrduction,omitempty"`
	// cluster corresponding to the environemt where application is deployed
	ClusterName *string `json:"clusterName,omitempty"`
	// clusterId corresponding to the environemt where application is deployed
	ClusterId *int32 `json:"clusterId,omitempty"`

	IsVirtualEnvironment *bool `json:"isVirtualEnvironment"`
}

type HelmAppDetails struct {
	// time when this application was last deployed/updated
	LastDeployedAt *time.Time `json:"lastDeployedAt,omitempty"`
	// name of the helm application/helm release name
	AppName *string `json:"appName,omitempty"`
	// unique identifier for app
	AppId *string `json:"appId,omitempty"`
	// name of the chart
	ChartName *string `json:"chartName,omitempty"`
	// url/location of the chart icon
	ChartAvatar *string `json:"chartAvatar,omitempty"`
	// unique identifier for the project, APP with no project will have id `0`
	ProjectId *int32 `json:"projectId,omitempty"`
	// chart version
	ChartVersion      *string             `json:"chartVersion,omitempty"`
	EnvironmentDetail *EnvironmentDetails `json:"environmentDetail,omitempty"`
	AppStatus         *string             `json:"appStatus,omitempty"`
}

type AppListDetail struct {
	// clusters to which result corresponds
	ClusterIds *[]int32 `json:"clusterIds,omitempty"`
	// application type inside the array
	ApplicationType *string `json:"applicationType,omitempty"`
	// if data fetch for that cluster produced error
	Errored *bool `json:"errored,omitempty"`
	// error msg if client failed to fetch
	ErrorMsg *string `json:"errorMsg,omitempty"`
	// all helm app list, EA+ devtronapp
	HelmApps *[]HelmAppDetails `json:"helmApps,omitempty"`
	// all helm app list, EA+ devtronapp
	DevtronApps *[]openapi.DevtronApp `json:"devtronApps,omitempty"`
}
