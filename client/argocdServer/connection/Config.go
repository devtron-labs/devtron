/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package connection

import (
	"github.com/caarlos0/env"
)

type Config struct {
	Host      string `env:"CD_HOST" envDefault:"localhost"`
	Port      string `env:"CD_PORT" envDefault:"8000"`
	Namespace string `env:"CD_NAMESPACE" envDefault:"devtroncd"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}
