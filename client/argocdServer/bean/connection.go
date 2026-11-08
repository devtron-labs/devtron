package bean

import (
	"github.com/caarlos0/env/v6"
	"k8s.io/client-go/rest"
)

type ArgoGRPCConfig struct {
	ConnectionConfig *Config
	AuthConfig       *AcdAuthConfig
}

type ArgoK8sConfig struct {
	RestConfig       *rest.Config
	AcdNamespace     string
	AcdConfigMapName string
}

type AcdAuthConfig struct {
	ClusterId                 int
	DevtronSecretName         string
	DevtronDexSecretNamespace string
	UserName                  string
	Password                  string
}

type Config struct {
	Host      string `env:"CD_HOST" envDefault:"localhost" description:"Host for the devtron stack"`
	Port      string `env:"CD_PORT" envDefault:"8000" description:"Port for pre/post-cd" `
	Namespace string `env:"CD_NAMESPACE" envDefault:"devtroncd"`
	TokenFile string `env:"CD_TOKEN_FILE" description:"File containing an externally managed Argo CD API token; disables automatic account provisioning"`
	CAFile    string `env:"CD_TLS_CA_FILE" description:"PEM CA bundle for Argo CD TLS; external token mode uses system roots when unset"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	return cfg, err
}
