package dashboard

import "github.com/caarlos0/env"

type Config struct {
	Host      string `env:"DASHBOARD_HOST" envDefault:"localhost"`
	Port      string `env:"DASHBOARD_PORT" envDefault:"3000"`
	Namespace string `env:"DASHBOARD_NAMESPACE" envDefault:"devtroncd"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}
