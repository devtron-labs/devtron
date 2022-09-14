package util

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/juju/errors"
)

type GlobalEnvVariables struct {
	GitOpsRepoPrefix string `env:"GITOPS_REPO_PREFIX" envDefault:""`
}

func GetGlobalEnvVariables() (*GlobalEnvVariables, error) {
	cfg := &GlobalEnvVariables{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

type DevtronSecretConfig struct {
	DevtronSecretName string `env:"DEVTRON_SECRET_NAME" envDefault:"devtron-secret"`
}

func GetDevtronSecretName() (*DevtronSecretConfig, error) {
	secretConfig := &DevtronSecretConfig{}
	err := env.Parse(secretConfig)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not get devtron secret name from environment : %v", err))
	}
	return secretConfig, err
}
