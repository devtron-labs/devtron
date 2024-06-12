/*
 * Copyright (c) 2024. Devtron Inc.
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

package serverEnvConfig

import (
	"fmt"
	"github.com/caarlos0/env"
)

type ServerEnvConfig struct {
	DevtronInstallationType              string `env:"DEVTRON_INSTALLATION_TYPE"`
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
	DevtronModulesIdentifierInHelmValues string `env:"DEVTRON_MODULES_IDENTIFIER_IN_HELM_VALUES" envDefault:"installer.modules"`
	DevtronBomUrl                        string `env:"DEVTRON_BOM_URL" envDefault:"https://raw.githubusercontent.com/devtron-labs/devtron/%s/charts/devtron/devtron-bom.yaml"`
	AppSyncImage                         string `env:"APP_SYNC_IMAGE" envDefault:"quay.io/devtron/chart-sync:1227622d-132-3775"`
	AppSyncJobResourcesObj               string `env:"APP_SYNC_JOB_RESOURCES_OBJ"`
	ModuleMetaDataApiUrl                 string `env:"MODULE_METADATA_API_URL" envDefault:"https://api.devtron.ai/module?name=%s"`
}

func ParseServerEnvConfig() (*ServerEnvConfig, error) {
	cfg := &ServerEnvConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse server env config: " + err.Error())
		return nil, err
	}
	return cfg, nil
}
