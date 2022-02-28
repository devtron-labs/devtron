package util

import "github.com/caarlos0/env"

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
