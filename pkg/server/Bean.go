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

package server

import "github.com/caarlos0/env"

type ServerInfoDto struct {
	CurrentVersion string `json:"currentVersion,notnull"`
	Status         string `json:"status,notnull" validate:"oneof=available installing installFailed timeout"`
	ReleaseName    string `json:"releaseName,notnull"`
}

type ServerActionRequestDto struct {
	Action  string `json:"action,notnull" validate:"oneof=upgrade"`
	Version string `json:"version,notnull"`
}

type ActionResponse struct {
	Success bool `json:"success"`
}

type ServerEnvConfig struct {
	CanServerUpdate                      bool   `env:"CAN_SERVER_UPDATE" envDefault:"true"` // default true
	InstallerCrdObjectGroupName          string `env:"INSTALLER_CRD_OBJECT_GROUP_NAME" envDefault:"installer.devtron.ai"`
	InstallerCrdObjectVersion            string `env:"INSTALLER_CRD_OBJECT_VERSION" envDefault:"v1alpha1"`
	InstallerCrdObjectResource           string `env:"INSTALLER_CRD_OBJECT_RESOURCE" envDefault:"installers"`
	InstallerCrdNamespace                string `env:"INSTALLER_CRD_NAMESPACE" envDefault:"devtroncd"`
	DevtronHelmRepoName                  string `env:"DEVTRON_HELM_REPO_NAME" envDefault:"devtron"`
	DevtronHelmRepoUrl                   string `env:"DEVTRON_HELM_REPO_URL" envDefault:"https://helm.devtron.ai"`
	DevtronHelmReleaseName               string `env:"DEVTRON_HELM_RELEASE_NAME" envDefault:"devtron"`
	DevtronHelmReleaseNamespace          string `env:"DEVTRON_HELM_RELEASE_NAMESPACE" envDefault:"devtroncd"`
	DevtronHelmReleaseChartName          string `env:"DEVTRON_HELM_RELEASE_CHART_NAME" envDefault:"devtron-operator"`
	DevtronVersionIdentifierInHelmValues string `env:"DEVTRON_VERSION_IDENTIFIER_IN_HELM_VALUES" envDefault:"installer.release"`
}

func ParseServerEnvConfig() *ServerEnvConfig {
	cfg := &ServerEnvConfig{}
	err := env.Parse(cfg)
	if err != nil {
		panic("failed to parse server env config: " + err.Error())
	}
	return cfg
}

type ServerStatus = string
type InstallerCrdObjectStatus = string

const (
	ServerStatusAvailable     ServerStatus = "available"
	ServerStatusInstalling    ServerStatus = "installing"
	ServerStatusInstallFailed ServerStatus = "installFailed"
	ServerStatusTimeout       ServerStatus = "timeout"

	InstallerCrdObjectStatusBlank      InstallerCrdObjectStatus = ""
	InstallerCrdObjectStatusDownloaded InstallerCrdObjectStatus = "Downloaded"
	InstallerCrdObjectStatusApplied    InstallerCrdObjectStatus = "Applied"
)
