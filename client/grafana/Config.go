package grafana

import "github.com/caarlos0/env"

type Config struct {
	Host      string `env:"GRAFANA_HOST" envDefault:"localhost"`
	Port      string `env:"GRAFANA_PORT" envDefault:"8090"`
	Namespace string `env:"GRAFANA_NAMESPACE" envDefault:"devtroncd"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}
