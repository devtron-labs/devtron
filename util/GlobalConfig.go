package util

import (
	"github.com/caarlos0/env"
)

type EnvironmentVariables struct {
	GlobalEnvVariables          *GlobalEnvVariables
	DevtronSecretConfig         *DevtronSecretConfig
	DeploymentServiceTypeConfig *DeploymentServiceTypeConfig
}

type DeploymentServiceTypeConfig struct {
	ExternallyManagedDeploymentType bool `env:"IS_INTERNAL_USE" envDefault:"false"`
	HelmInstallASyncMode            bool `env:"RUN_HELM_INSTALL_IN_ASYNC_MODE_HELM_APPS" envDefault:"false"`
}

type GlobalEnvVariables struct {
	GitOpsRepoPrefix               string `env:"GITOPS_REPO_PREFIX" envDefault:""`
	EnableAsyncInstallDevtronChart bool   `env:"ENABLE_ASYNC_INSTALL_DEVTRON_CHART" envDefault:"false"`
	ExposeCiMetrics                bool   `env:"EXPOSE_CI_METRICS" envDefault:"false"`
	DeploymentTriggerTimeout       int    `env:"DEPLOYMENT_TRIGGER_TIMEOUT" envDefault:"3"`
	HelmAppInstallTimeout          int    `env:"HELM_APP_INSTALL_TIMEOUT" envDefault:"3"`
}

type DevtronSecretConfig struct {
	DevtronSecretName         string `env:"DEVTRON_SECRET_NAME" envDefault:"devtron-secret"`
	DevtronDexSecretNamespace string `env:"DEVTRON_DEX_SECRET_NAMESPACE" envDefault:"devtroncd"`
}

func GetEnvironmentVariables() (*EnvironmentVariables, error) {
	cfg := &EnvironmentVariables{
		GlobalEnvVariables:          &GlobalEnvVariables{},
		DevtronSecretConfig:         &DevtronSecretConfig{},
		DeploymentServiceTypeConfig: &DeploymentServiceTypeConfig{},
	}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}
