/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package appStoreBean

import (
	"encoding/json"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"time"
)

//v1
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

type InstallAppVersionDTO struct {
	Id                        int                        `json:"id,omitempty"`
	AppId                     int                        `json:"appId,omitempty"`
	AppName                   string                     `json:"appName,omitempty"`
	TeamId                    int                        `json:"teamId,omitempty"`
	EnvironmentId             int                        `json:"environmentId,omitempty"`
	InstalledAppId            int                        `json:"installedAppId,omitempty,notnull"`
	InstalledAppVersionId     int                        `json:"installedAppVersionId,omitempty,notnull"`
	AppStoreVersion           int                        `json:"appStoreVersion,omitempty,notnull"`
	ValuesOverrideYaml        string                     `json:"valuesOverrideYaml,omitempty"`
	Readme                    string                     `json:"readme,omitempty"`
	UserId                    int32                      `json:"-"`
	ReferenceValueId          int                        `json:"referenceValueId, omitempty" validate:"required,number"`
	ReferenceValueKind        string                     `json:"referenceValueKind, omitempty" validate:"oneof=DEFAULT TEMPLATE DEPLOYED EXISTING"`
	ACDAppName                string                     `json:"-"`
	Environment               *repository2.Environment   `json:"-"`
	ChartGroupEntryId         int                        `json:"-"`
	DefaultClusterComponent   bool                       `json:"-"`
	Status                    AppstoreDeploymentStatus   `json:"-"`
	AppStoreId                int                        `json:"appStoreId"`
	AppStoreName              string                     `json:"appStoreName"`
	Deprecated                bool                       `json:"deprecated"`
	ForceDelete               bool                       `json:"-"`
	ClusterId                 int                        `json:"clusterId"` // needed for hyperion mode
	Namespace                 string                     `json:"namespace"` // needed for hyperion mode
	AppOfferingMode           string                     `json:"appOfferingMode"`
	GitOpsRepoName            string                     `json:"gitOpsRepoName"`
	GitOpsPath                string                     `json:"gitOpsPath"`
	GitHash                   string                     `json:"gitHash"`
	EnvironmentName           string                     `json:"-"`
	InstallAppVersionChartDTO *InstallAppVersionChartDTO `json:"-"`
	DeploymentAppType         string                     `json:"-"`
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

/// bean for v2
type ChartGroupInstallRequest struct {
	ProjectId                     int                              `json:"projectId"  validate:"required,number"`
	ChartGroupInstallChartRequest []*ChartGroupInstallChartRequest `json:"charts" validate:"dive,required"`
	ChartGroupId                  int                              `json:"chartGroupId"` //optional
	UserId                        int32                            `json:"-"`
}

type ChartGroupInstallChartRequest struct {
	AppName                 string `json:"appName,omitempty"  validate:"name-component,max=100" `
	EnvironmentId           int    `json:"environmentId,omitempty" validate:"required,number" `
	AppStoreVersion         int    `json:"appStoreVersion,omitempty,notnull" validate:"required,number" `
	ValuesOverrideYaml      string `json:"valuesOverrideYaml,omitempty"` //optional
	ReferenceValueId        int    `json:"referenceValueId, omitempty" validate:"required,number"`
	ReferenceValueKind      string `json:"referenceValueKind, omitempty" validate:"oneof=DEFAULT TEMPLATE DEPLOYED"`
	ChartGroupEntryId       int    `json:"chartGroupEntryId"` //optional
	DefaultClusterComponent bool   `json:"-"`
}
type ChartGroupInstallAppRes struct {
}

///
type RefChartProxyDir string

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

const BULK_APPSTORE_DEPLOY_TOPIC = "ORCHESTRATOR.APP-STORE.BULK-DEPLOY"
const BULK_APPSTORE_DEPLOY_GROUP = "ORCHESTRATOR.APP-STORE.BULK-DEPLOY-GROUP-1"

const BULK_APPSTORE_DEPLOY_DURABLE = "ORCHESTRATOR.APP-STORE.BULK-DEPLOY.DURABLE-1"

type DeployPayload struct {
	InstalledAppVersionId int
}

const REFERENCE_TYPE_DEFAULT string = "DEFAULT"
const REFERENCE_TYPE_TEMPLATE string = "TEMPLATE"
const REFERENCE_TYPE_DEPLOYED string = "DEPLOYED"
const REFERENCE_TYPE_EXISTING string = "EXISTING"

type AppStoreVersionValuesDTO struct {
	Id                int    `json:"id,omitempty"`
	AppStoreVersionId int    `json:"appStoreVersionId,omitempty,notnull"`
	Name              string `json:"name,omitempty"`
	Values            string `json:"values,omitempty"` //yaml format user value
	ChartVersion      string `json:"chartVersion,omitempty"`
	EnvironmentName   string `json:"environmentName,omitempty"`
	UserId            int32  `json:"-"`
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
	ChartName                    string    `json:"chart_name"`
	Icon                         string    `json:"icon"`
	Active                       bool      `json:"active"`
	ChartGitLocation             string    `json:"chart_git_location"`
	CreatedOn                    time.Time `json:"created_on"`
	UpdatedOn                    time.Time `json:"updated_on"`
	Version                      string    `json:"version"`
	Deprecated                   bool      `json:"deprecated"`
}

type AppStoreFilter struct {
	ChartRepoId       []int  `json:"chartRepoId"`
	AppStoreName      string `json:"appStoreName"`
	AppName           string `json:"appName"`
	IncludeDeprecated bool   `json:"includeDeprecated"`
	Offset            int    `json:"offset"`
	Size              int    `json:"size"`
	EnvIds            []int  `json:"envIds"`
	OnlyDeprecated    bool   `json:"onlyDeprecated"`
	ClusterIds        []int  `json:"clusterIds"`
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
