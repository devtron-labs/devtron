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

package util

import (
	"github.com/caarlos0/env"
)

type EnvironmentVariables struct {
	GlobalEnvVariables          *GlobalEnvVariables
	DevtronSecretConfig         *DevtronSecretConfig
	DeploymentServiceTypeConfig *DeploymentServiceTypeConfig
	TerminalEnvVariables        *TerminalEnvVariables
	GlobalClusterConfig         *GlobalClusterConfig
	InternalEnvVariables        *InternalEnvVariables
}

// CATEGORY=CD
type DeploymentServiceTypeConfig struct {
	ExternallyManagedDeploymentType       bool `env:"IS_INTERNAL_USE" envDefault:"false"`
	HelmInstallASyncMode                  bool `env:"RUN_HELM_INSTALL_IN_ASYNC_MODE_HELM_APPS" envDefault:"false"`
	UseDeploymentConfigData               bool `env:"USE_DEPLOYMENT_CONFIG_DATA" envDefault:"false" description:"use deployment config data from deployment_config table" deprecated:"true"`
	MigrateDeploymentConfigData           bool `env:"MIGRATE_DEPLOYMENT_CONFIG_DATA" envDefault:"false" description:"migrate deployment config data from charts table to deployment_config table" deprecated:"false"`
	FeatureMigrateArgoCdApplicationEnable bool `env:"FEATURE_MIGRATE_ARGOCD_APPLICATION_ENABLE" envDefault:"false" description:"enable migration of external argocd application to devtron pipeline" deprecated:"false"`
	ShouldCheckNamespaceOnClone           bool `env:"SHOULD_CHECK_NAMESPACE_ON_CLONE" envDefault:"false"  description:"should we check if namespace exists or not while cloning app" deprecated:"false"`
}

func (d *DeploymentServiceTypeConfig) IsFeatureMigrateArgoCdApplicationEnable() bool {
	return true
}

type GlobalEnvVariables struct {
	GitOpsRepoPrefix                     string `env:"GITOPS_REPO_PREFIX" envDefault:""`
	EnableAsyncHelmInstallDevtronChart   bool   `env:"ENABLE_ASYNC_INSTALL_DEVTRON_CHART" envDefault:"false"`
	EnableAsyncArgoCdInstallDevtronChart bool   `env:"ENABLE_ASYNC_ARGO_CD_INSTALL_DEVTRON_CHART" envDefault:"false"`
	ArgoGitCommitRetryCountOnConflict    int    `env:"ARGO_GIT_COMMIT_RETRY_COUNT_ON_CONFLICT" envDefault:"3"`
	ArgoGitCommitRetryDelayOnConflict    int    `env:"ARGO_GIT_COMMIT_RETRY_DELAY_ON_CONFLICT" envDefault:"1"`
	ExposeCiMetrics                      bool   `env:"EXPOSE_CI_METRICS" envDefault:"false"`
	ExecuteWireNilChecker                bool   `env:"EXECUTE_WIRE_NIL_CHECKER" envDefault:"false"`
}

type GlobalClusterConfig struct {
	ClusterStatusCronTime int `env:"CLUSTER_STATUS_CRON_TIME" envDefault:"15"`
}

type DevtronSecretConfig struct {
	DevtronSecretName         string `env:"DEVTRON_SECRET_NAME" envDefault:"devtron-secret"`
	DevtronDexSecretNamespace string `env:"DEVTRON_DEX_SECRET_NAMESPACE" envDefault:"devtroncd"`
}

type TerminalEnvVariables struct {
	RestrictTerminalAccessForNonSuperUser bool `env:"RESTRICT_TERMINAL_ACCESS_FOR_NON_SUPER_USER" envDefault:"false"`
}

type InternalEnvVariables struct {
	// GoRuntimeEnv specifies the runtime environment of the application,
	//	- enum:
	//		"development"
	//		"production"
	//	- default: "production"
	//	- use cases: test cases to set the runtime environment
	GoRuntimeEnv string `env:"GO_RUNTIME_ENV" envDefault:"production"`
}

func (i *InternalEnvVariables) IsDevelopmentEnv() bool {
	if i == nil {
		return false
	}
	return i.GoRuntimeEnv == "development"
}

func (i *InternalEnvVariables) SetDevelopmentEnv() *InternalEnvVariables {
	i.GoRuntimeEnv = "development"
	return i
}

func GetEnvironmentVariables() (*EnvironmentVariables, error) {
	cfg := &EnvironmentVariables{
		GlobalEnvVariables:          &GlobalEnvVariables{},
		DevtronSecretConfig:         &DevtronSecretConfig{},
		DeploymentServiceTypeConfig: &DeploymentServiceTypeConfig{},
		TerminalEnvVariables:        &TerminalEnvVariables{},
		GlobalClusterConfig:         &GlobalClusterConfig{},
		InternalEnvVariables:        &InternalEnvVariables{},
	}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}
