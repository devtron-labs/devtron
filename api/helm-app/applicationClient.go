package client

import (
	"context"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type HelmAppClient interface {
	ListApplication(req *AppListRequest) (ApplicationService_ListApplicationsClient, error)
}

type HelmAppClientImpl struct {
	logger           *zap.SugaredLogger
	helmClientConfig *HelmClientConfig
}

func NewHelmAppClientImpl(logger *zap.SugaredLogger, helmClientConfig *HelmClientConfig) *HelmAppClientImpl {
	return &HelmAppClientImpl{
		logger:           logger,
		helmClientConfig: helmClientConfig,
	}
}

type HelmClientConfig struct {
	Url string `env:"HELM_CLIENT_URL" envDefault:"127.0.0.1:50051"`
}

func GetConfig() (*HelmClientConfig, error) {
	cfg := &HelmClientConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func (impl *HelmAppClientImpl) getConnection() (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(impl.helmClientConfig.Url, opts...)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (impl *HelmAppClientImpl) ListApplication(req *AppListRequest) (ApplicationService_ListApplicationsClient, error) {
	conn, err := impl.getConnection()
	//defer util.Close(conn, impl.logger)
	if err != nil {
		return nil, err
	}
	applicationClient := NewApplicationServiceClient(conn)
	stream, err := applicationClient.ListApplications(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return stream, nil
}
