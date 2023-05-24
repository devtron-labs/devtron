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

package module

import (
	"fmt"
	"github.com/caarlos0/env"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ModuleInfoDto struct {
	Name                  string                     `json:"name,notnull"`
	Status                string                     `json:"status,notnull" validate:"oneof=notInstalled installed installing installFailed timeout"`
	ModuleResourcesStatus []*ModuleResourceStatusDto `json:"moduleResourcesStatus"`
}

type ModuleConfigDto struct {
	Enabled bool `json:"enabled"`
}

type BlobStorageConfig struct {
	Enabled bool `env:"BLOB_STORAGE_ENABLED" envDefault:"false"`
}

type ModuleActionRequestDto struct {
	Action  string `json:"action,notnull" validate:"oneof=install"`
	Version string `json:"version,notnull"`
}

type ActionResponse struct {
	Success bool `json:"success"`
}

type ModuleEnvConfig struct {
	ModuleStatusHandlingCronDurationInMin int `env:"MODULE_STATUS_HANDLING_CRON_DURATION_MIN" envDefault:"3"` // default 3 minutes
}

type ModuleResourceStatusDto struct {
	Group         string `json:"group"`
	Version       string `json:"version"`
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	HealthStatus  string `json:"healthStatus"`
	HealthMessage string `json:"healthMessage"`
}

func ParseModuleEnvConfig() (*ModuleEnvConfig, error) {
	cfg := &ModuleEnvConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse module env config: " + err.Error())
		return nil, err
	}

	return cfg, nil
}

type ModuleStatus = string
type ModuleName = string

const BlobStorage = "blob-storage"
const INSTALLER_MODULES_HELM_KEY = "installer.modules"

const (
	ModuleStatusNotInstalled  ModuleStatus = "notInstalled"
	ModuleStatusInstalled     ModuleStatus = "installed"
	ModuleStatusInstalling    ModuleStatus = "installing"
	ModuleStatusInstallFailed ModuleStatus = "installFailed"
	ModuleStatusTimeout       ModuleStatus = "timeout"
)

const (
	ModuleNameCicd              ModuleName = "cicd"
	ModuleNameArgoCd            ModuleName = "argo-cd"
	ModuleNameSecurityClair     ModuleName = "security.clair"
	ModuleNameNotification      ModuleName = "notifier"
	ModuleNameMonitoringGrafana ModuleName = "monitoring.grafana"
)

var SupportedModuleNamesListFirstReleaseExcludingCicd = []string{ModuleNameArgoCd, ModuleNameSecurityClair, ModuleNameNotification, ModuleNameMonitoringGrafana}

type ResourceFilter struct {
	GlobalFilter    *ResourceIdentifier `json:"globalFilter,omitempty"`
	GvkLevelFilters []*GvkLevelFilter   `json:"gvkLevelFilters,omitempty"`
}

type GvkLevelFilter struct {
	Gvk                *schema.GroupVersionKind `json:"gvk"`
	ResourceIdentifier *ResourceIdentifier      `json:"filter"`
}

type ResourceIdentifier struct {
	Labels map[string]string `json:"labels"`
}
