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

package appstore

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/appstore"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"time"
)

//v1
type InstallAppVersionDTO struct {
	Id                      int                               `json:"id,omitempty"`
	AppId                   int                               `json:"appId,omitempty"`
	AppName                 string                            `json:"appName,omitempty"`
	TeamId                  int                               `json:"teamId,omitempty"`
	EnvironmentId           int                               `json:"environmentId,omitempty"`
	InstalledAppId          int                               `json:"installedAppId,omitempty,notnull"`
	InstalledAppVersionId   int                               `json:"installedAppVersionId,omitempty,notnull"`
	AppStoreVersion         int                               `json:"appStoreVersion,omitempty,notnull"`
	ValuesOverride          json.RawMessage                   `json:"valuesOverride,omitempty"` //json format user value
	ValuesOverrideYaml      string                            `json:"valuesOverrideYaml,omitempty"`
	Readme                  string                            `json:"readme,omitempty"`
	UserId                  int32                             `json:"-"`
	ReferenceValueId        int                               `json:"referenceValueId, omitempty" validate:"required,number"`
	ReferenceValueKind      string                            `json:"referenceValueKind, omitempty" validate:"oneof=DEFAULT TEMPLATE DEPLOYED"`
	ACDAppName              string                            `json:"-"`
	Environment             *cluster.Environment              `json:"-"`
	ChartGroupEntryId       int                               `json:"-"`
	DefaultClusterComponent bool                              `json:"-"`
	Status                  appstore.AppstoreDeploymentStatus `json:"-"`
	AppStoreId              int                               `json:"appStoreId"`
	AppStoreName            string                            `json:"appStoreName"`
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
