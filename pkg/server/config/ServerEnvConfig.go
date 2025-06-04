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
	DevtronInstallationType                     string `env:"DEVTRON_INSTALLATION_TYPE" description:"Devtron Installation type(EA/Full)"`
	InstallerCrdObjectGroupName                 string `env:"INSTALLER_CRD_OBJECT_GROUP_NAME" envDefault:"installer.devtron.ai" description:"Devtron installer CRD group name, partially deprecated."`
	InstallerCrdObjectVersion                   string `env:"INSTALLER_CRD_OBJECT_VERSION" envDefault:"v1alpha1" description:"version of the CRDs. default is v1alpha1"`
	InstallerCrdObjectResource                  string `env:"INSTALLER_CRD_OBJECT_RESOURCE" envDefault:"installers" description:"Devtron installer CRD resource name, partially deprecated"`
	InstallerCrdNamespace                       string `env:"INSTALLER_CRD_NAMESPACE" envDefault:"devtroncd" description:"namespace where Custom Resource Definitions get installed"`
	DevtronHelmRepoName                         string `env:"DEVTRON_HELM_REPO_NAME" envDefault:"devtron" description:"Is used to install modules (stack manager)"`
	DevtronHelmRepoUrl                          string `env:"DEVTRON_HELM_REPO_URL" envDefault:"https://helm.devtron.ai" description:"Is used to install modules (stack manager)"`
	DevtronHelmReleaseName                      string `env:"DEVTRON_HELM_RELEASE_NAME" envDefault:"devtron" description:"Name of the Devtron Helm release. "`
	DevtronHelmReleaseNamespace                 string `env:"DEVTRON_HELM_RELEASE_NAMESPACE" envDefault:"devtroncd" description:"Namespace of the Devtron Helm release"`
	DevtronHelmReleaseChartName                 string `env:"DEVTRON_HELM_RELEASE_CHART_NAME" envDefault:"devtron-operator" description:""`
	DevtronVersionIdentifierInHelmValues        string `env:"DEVTRON_VERSION_IDENTIFIER_IN_HELM_VALUES" envDefault:"installer.release" description:"devtron operator version identifier in helm values yaml"`
	DevtronModulesIdentifierInHelmValues        string `env:"DEVTRON_MODULES_IDENTIFIER_IN_HELM_VALUES" envDefault:"installer.modules" `
	DevtronBomUrl                               string `env:"DEVTRON_BOM_URL" envDefault:"https://raw.githubusercontent.com/devtron-labs/devtron/%s/charts/devtron/devtron-bom.yaml" description:"Path to devtron-bom.yaml of devtron charts, used for module installation and devtron upgrade"`
	AppSyncImage                                string `env:"APP_SYNC_IMAGE" envDefault:"quay.io/devtron/chart-sync:1227622d-132-3775" description:"For the app sync image, this image will be used in app-manual sync job"`
	AppSyncServiceAccount                       string `env:"APP_SYNC_SERVICE_ACCOUNT" envDefault:"chart-sync" description:"Service account to be used in app sync Job"`
	AppSyncJobResourcesObj                      string `env:"APP_SYNC_JOB_RESOURCES_OBJ" description:"To pass the resource of app sync"`
	ModuleMetaDataApiUrl                        string `env:"MODULE_METADATA_API_URL" envDefault:"https://api.devtron.ai/module?name=%s" description:"Modules list and meta info will be fetched from this server, that is central api server of devtron."`
	ParallelismLimitForTagProcessing            int    `env:"PARALLELISM_LIMIT_FOR_TAG_PROCESSING" description:"App manual sync job parallel tag processing count."`
	AppSyncJobShutDownWaitDuration              int    `env:"APP_SYNC_SHUTDOWN_WAIT_DURATION" envDefault:"120"`
	DevtronOperatorBasePath                     string `env:"DEVTRON_OPERATOR_BASE_PATH" envDefault:"" description:"Base path for devtron operator, used to find the helm charts and values files"`
	DevtronInstallerModulesPath                 string `env:"DEVTRON_INSTALLER_MODULES_PATH" envDefault:"installer.modules" description:"Path to devtron installer modules, used to find the helm charts and values files"`
	DevtronInstallerReleasePath                 string `env:"DEVTRON_INSTALLER_RELEASE_PATH" envDefault:"installer.release" description:"Path to devtron installer release, used to find the helm charts and values files"`
	ErrorEncounteredOnGettingDevtronHelmRelease error
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
