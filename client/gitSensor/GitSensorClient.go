package gitSensor

import (
	"context"
	"errors"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

type Client interface {
	SaveGitProvider(ctx context.Context, provider *GitProvider) error
	AddRepo(ctx context.Context, materials []*GitMaterial) error
	UpdateRepo(ctx context.Context, material *GitMaterial) error
	SavePipelineMaterial(ctx context.Context, ciPipelineMaterials []*CiPipelineMaterial) error

	FetchChanges(ctx context.Context, req *FetchScmChangesRequest) (*MaterialChangeResp, error)
	GetHeadForPipelineMaterials(ctx context.Context, req *HeadRequest) ([]*CiPipelineMaterial, error)
	GetCommitMetadata(ctx context.Context, req *CommitMetadataRequest) (*GitCommit, error)
	GetCommitMetadataForPipelineMaterial(ctx context.Context, req *CommitMetadataRequest) (*GitCommit, error)
	RefreshGitMaterial(ctx context.Context, req *RefreshGitMaterialRequest) (*RefreshGitMaterialResponse, error)

	GetWebhookData(ctx context.Context, req *WebhookDataRequest) (*WebhookAndCiData, error)
	GetAllWebhookEventConfigForHost(ctx context.Context, req *WebhookEventConfigRequest) ([]*WebhookEventConfig, error)
	GetWebhookEventConfig(ctx context.Context, req *WebhookEventConfigRequest) (*WebhookEventConfig, error)
	GetWebhookPayloadDataForPipelineMaterialId(ctx context.Context, req *WebhookPayloadDataRequest) (*WebhookPayloadDataResponse, error)
	GetWebhookPayloadFilterDataForPipelineMaterialId(ctx context.Context, req *WebhookPayloadFilterDataRequest) (*WebhookPayloadFilterDataResponse, error)
}

type ClientImpl struct {
	logger    *zap.SugaredLogger
	config    *ClientConfig
	apiClient ApiClient
}

func NewGitSensorClient(logger *zap.SugaredLogger, config *ClientConfig) (*ClientImpl, error) {
	client := &ClientImpl{
		logger: logger,
		config: config,
	}

	var apiClient ApiClient
	var err error
	if config.Protocol == "REST" {
		logger.Infow("using REST api client for git sensor")
		apiClient, err = NewGitSensorSession(config, logger)

	} else if config.Protocol == "GRPC" {
		logger.Infow("using gRPC api client for git sensor")
		apiClient, err = NewGitSensorGrpcClientImpl(logger, config)

	} else {
		err = errors.New("unknown protocol configured for git sensor client")
		logger.Errorw(err.Error())
		return nil, err
	}

	if err != nil {
		return nil, err
	} else {
		client.apiClient = apiClient
	}
	return client, nil
}

type ClientConfig struct {
	Url                       string `env:"GIT_SENSOR_URL" envDefault:"127.0.0.1:7070"`
	Protocol                  string `env:"GIT_SENSOR_PROTOCOL" envDefault:"REST"`
	Timeout                   int    `env:"GIT_SENSOR_TIMEOUT" envDefault:"0"` // in seconds
	MaxSizeOfDataTransferInMb int    `env:"GIT_SENSOR_GRPC_DATA_TRANSFER_MAX_SIZE" envDefault:"4"`
}

func GetConfig() (*ClientConfig, error) {
	cfg := &ClientConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func (c *ClientImpl) SaveGitProvider(ctx context.Context, provider *GitProvider) error {
	return c.apiClient.SaveGitProvider(ctx, provider)
}

func (c *ClientImpl) AddRepo(ctx context.Context, materials []*GitMaterial) error {
	return c.apiClient.AddRepo(ctx, materials)
}

func (c *ClientImpl) UpdateRepo(ctx context.Context, material *GitMaterial) error {
	return c.apiClient.UpdateRepo(ctx, material)
}

func (c *ClientImpl) SavePipelineMaterial(ctx context.Context, ciPipelineMaterials []*CiPipelineMaterial) error {
	return c.apiClient.SavePipelineMaterial(ctx, ciPipelineMaterials)
}

func (c *ClientImpl) FetchChanges(ctx context.Context, req *FetchScmChangesRequest) (*MaterialChangeResp, error) {
	return c.apiClient.FetchChanges(ctx, req)
}

func (c *ClientImpl) GetHeadForPipelineMaterials(ctx context.Context, req *HeadRequest) ([]*CiPipelineMaterial, error) {
	return c.apiClient.GetHeadForPipelineMaterials(ctx, req)
}

func (c *ClientImpl) GetCommitMetadata(ctx context.Context, req *CommitMetadataRequest) (*GitCommit, error) {
	return c.apiClient.GetCommitMetadata(ctx, req)
}

func (c *ClientImpl) GetCommitMetadataForPipelineMaterial(ctx context.Context, req *CommitMetadataRequest) (*GitCommit, error) {
	return c.apiClient.GetCommitMetadataForPipelineMaterial(ctx, req)
}

func (c *ClientImpl) RefreshGitMaterial(ctx context.Context, req *RefreshGitMaterialRequest) (*RefreshGitMaterialResponse, error) {
	return c.apiClient.RefreshGitMaterial(ctx, req)
}

func (c *ClientImpl) GetWebhookData(ctx context.Context, req *WebhookDataRequest) (*WebhookAndCiData, error) {
	return c.apiClient.GetWebhookData(ctx, req)
}

func (c *ClientImpl) GetAllWebhookEventConfigForHost(ctx context.Context, req *WebhookEventConfigRequest) ([]*WebhookEventConfig, error) {
	return c.apiClient.GetAllWebhookEventConfigForHost(ctx, req)
}

func (c *ClientImpl) GetWebhookEventConfig(ctx context.Context, req *WebhookEventConfigRequest) (*WebhookEventConfig, error) {
	return c.apiClient.GetWebhookEventConfig(ctx, req)
}

func (c *ClientImpl) GetWebhookPayloadDataForPipelineMaterialId(ctx context.Context, req *WebhookPayloadDataRequest) (*WebhookPayloadDataResponse, error) {
	return c.apiClient.GetWebhookPayloadDataForPipelineMaterialId(ctx, req)
}

func (c *ClientImpl) GetWebhookPayloadFilterDataForPipelineMaterialId(ctx context.Context, req *WebhookPayloadFilterDataRequest) (*WebhookPayloadFilterDataResponse, error) {
	return c.apiClient.GetWebhookPayloadFilterDataForPipelineMaterialId(ctx, req)
}
