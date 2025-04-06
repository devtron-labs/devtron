package grpc

import "github.com/caarlos0/env"

type Configuration struct {
	KubelinkMaxRecvMsgSize int `env:"KUBELINK_GRPC_MAX_RECEIVE_MSG_SIZE" envDefault:"20" description: "Message size to receive from kubelink"` // In mb
	KubelinkMaxSendMsgSize int `env:"KUBELINK_GRPC_MAX_SEND_MSG_SIZE" envDefault:"4" description: "Message size to send to kubelink"`     // In mb
}

func GetConfiguration() (*Configuration, error) {
	cfg := &Configuration{}
	err := env.Parse(cfg)
	return cfg, err
}
