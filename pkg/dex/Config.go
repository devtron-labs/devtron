/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package dex

import "github.com/caarlos0/env"

type Config struct {
	Host string `env:"DEX_HOST" envDefault:"http://localhost"`
	Port string `env:"DEX_PORT" envDefault:"5556"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}
