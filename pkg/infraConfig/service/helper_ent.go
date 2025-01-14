package service

import (
	"github.com/caarlos0/env"
	v1 "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
)

func getDefaultInfraConfigFromEnv(envConfig *types.CiConfig) (*v1.InfraConfig, error) {
	infraConfiguration := &v1.InfraConfig{}
	err := env.Parse(infraConfiguration)
	if err != nil {
		return infraConfiguration, err
	}
	return infraConfiguration, nil
}
