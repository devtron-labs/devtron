package grpc

import "github.com/caarlos0/env"

// CATEGORY=INFRA_SETUP
type Configuration struct {
	KubelinkMaxRecvMsgSize    int    `env:"KUBELINK_GRPC_MAX_RECEIVE_MSG_SIZE" envDefault:"20"` // In mb
	KubelinkMaxSendMsgSize    int    `env:"KUBELINK_GRPC_MAX_SEND_MSG_SIZE" envDefault:"4"`     // In mb
	KubelinkGRPCServiceConfig string `env:"KUBELINK_GRPC_SERVICE_CONFIG" envDefault:"{\"loadBalancingPolicy\":\"round_robin\"}" description:"kubelink grpc service config"`
}

func GetConfiguration() (*Configuration, error) {
	cfg := &Configuration{}
	err := env.Parse(cfg)
	return cfg, err
}
