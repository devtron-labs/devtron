package git_sensor_client

import (
	"context"
	"fmt"
	"github.com/caarlos0/env"
	pb "github.com/devtron-labs/protos/git-sensor"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

const (
	ContextTimeoutInSeconds = 10
	MaxMsgSizeBytes         = 20 * 1024 * 1024
)

type GitSensorGrpcClient interface {
	SaveGitProvider(ctx context.Context, provider *pb.GitProvider) (res *pb.Empty, err error)
	AddRepo(ctx context.Context, req *pb.AddRepoRequest) (*pb.Empty, error)
	UpdateRepo(ctx context.Context, req *pb.GitMaterial) (*pb.Empty, error)
	SavePipelineMaterial(ctx context.Context, req *pb.SavePipelineMaterialRequest) (*pb.Empty, error)
	FetchChanges(ctx context.Context, req *pb.FetchScmChangesRequest) (*pb.MaterialChangeResponse, error)
	GetHeadForPipelineMaterials(ctx context.Context, req *pb.HeadRequest) (*pb.GetHeadForPipelineMaterialsResponse, error)
	GetCommitMetadata(ctx context.Context, req *pb.CommitMetadataRequest) (*pb.GitCommit, error)
	GetCommitMetadataForPipelineMaterial(ctx context.Context, req *pb.CommitMetadataRequest) (*pb.GitCommit, error)
	GetCommitInfoForTag(ctx context.Context, req *pb.CommitMetadataRequest) (*pb.GitCommit, error)
	RefreshGitMaterial(ctx context.Context, req *pb.RefreshGitMaterialRequest) (*pb.RefreshGitMaterialResponse, error)
	ReloadAllMaterial(ctx context.Context, req *pb.Empty) (*pb.Empty, error)
	ReloadMaterial(ctx context.Context, req *pb.ReloadMaterialRequest) (*pb.GenericResponse, error)
	GetChangesInRelease(ctx context.Context, req *pb.ReleaseChangeRequest) (*pb.GitChanges, error)
	GetWebhookData(ctx context.Context, req *pb.WebhookDataRequest) (*pb.WebhookAndCiData, error)
	GetAllWebhookEventConfigForHost(ctx context.Context, req *pb.WebhookEventConfigRequest) (*pb.WebhookEventConfigResponse, error)
	GetWebhookEventConfig(ctx context.Context, req *pb.WebhookEventConfigRequest) (*pb.WebhookEventConfig, error)
	GetWebhookPayloadDataForPipelineMaterialId(ctx context.Context, req *pb.WebhookPayloadDataRequest) (*pb.WebhookPayloadDataResponse, error)
	GetWebhookPayloadFilterDataForPipelineMaterialId(ctx context.Context, req *pb.WebhookPayloadFilterDataRequest) (*pb.WebhookPayloadFilterDataResponse, error)
}

type GitSensorGrpcClientImpl struct {
	logger        *zap.SugaredLogger
	config        *GitSensorGrpcClientConfig
	serviceClient pb.GitSensorServiceClient
}

func NewGitSensorGrpcClientImpl(logger *zap.SugaredLogger, config *GitSensorGrpcClientConfig) *GitSensorGrpcClientImpl {
	return &GitSensorGrpcClientImpl{
		logger: logger,
		config: config,
	}
}

// getGitSensorServiceClient initializes and returns gRPC GitSensorService client
func (client *GitSensorGrpcClientImpl) getGitSensorServiceClient() (pb.GitSensorServiceClient, error) {
	if client.serviceClient == nil {
		conn, err := client.getConnection()
		if err != nil {
			return nil, err
		}
		client.serviceClient = pb.NewGitSensorServiceClient(conn)
	}
	return client.serviceClient, nil
}

// getConnection initializes and returns a grpc client connection
func (client *GitSensorGrpcClientImpl) getConnection() (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeoutInSeconds*time.Second)
	defer cancel()

	// Configure gRPC dial options
	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(MaxMsgSizeBytes),
		),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	endpoint := fmt.Sprintf("dns:///%s", client.config.Url)

	// initialize connection
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		client.logger.Errorw("error while initializing grpc connection",
			"endpoint", endpoint,
			"err", err)
		return nil, err
	}
	return conn, nil
}

type GitSensorGrpcClientConfig struct {
	Url string `env:"GIT_SENSOR_HOST" envDefault:"127.0.0.1:7070"`
}

// GetConfig parses and returns GitSensor gRPC client configuration
func GetConfig() (*GitSensorGrpcClientConfig, error) {
	cfg := &GitSensorGrpcClientConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

// SaveGitProvider saves Git provider
func (client *GitSensorGrpcClientImpl) SaveGitProvider(ctx context.Context, provider *pb.GitProvider) (res *pb.Empty, err error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.SaveGitProvider(ctx, provider)
}

// AddRepo adds git materials
func (client *GitSensorGrpcClientImpl) AddRepo(ctx context.Context, req *pb.AddRepoRequest) (*pb.Empty, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.AddRepo(ctx, req)
}

// UpdateRepo updates the git materail
func (client *GitSensorGrpcClientImpl) UpdateRepo(ctx context.Context, req *pb.GitMaterial) (*pb.Empty, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.UpdateRepo(ctx, req)
}

// SavePipelineMaterial saves ci pipeline material info
func (client *GitSensorGrpcClientImpl) SavePipelineMaterial(ctx context.Context, req *pb.SavePipelineMaterialRequest) (
	*pb.Empty, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.SavePipelineMaterial(ctx, req)
}

func (client *GitSensorGrpcClientImpl) FetchChanges(ctx context.Context, req *pb.FetchScmChangesRequest) (
	*pb.MaterialChangeResponse, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.FetchChanges(ctx, req)
}

func (client *GitSensorGrpcClientImpl) GetHeadForPipelineMaterials(ctx context.Context, req *pb.HeadRequest) (
	*pb.GetHeadForPipelineMaterialsResponse, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.GetHeadForPipelineMaterials(ctx, req)
}

func (client *GitSensorGrpcClientImpl) GetCommitMetadata(ctx context.Context, req *pb.CommitMetadataRequest) (
	*pb.GitCommit, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.GetCommitMetadata(ctx, req)
}

func (client *GitSensorGrpcClientImpl) GetCommitMetadataForPipelineMaterial(ctx context.Context, req *pb.CommitMetadataRequest) (
	*pb.GitCommit, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.GetCommitMetadataForPipelineMaterial(ctx, req)
}

func (client *GitSensorGrpcClientImpl) GetCommitInfoForTag(ctx context.Context, req *pb.CommitMetadataRequest) (
	*pb.GitCommit, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.GetCommitInfoForTag(ctx, req)
}

func (client *GitSensorGrpcClientImpl) RefreshGitMaterial(ctx context.Context, req *pb.RefreshGitMaterialRequest) (
	*pb.RefreshGitMaterialResponse, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.RefreshGitMaterial(ctx, req)
}

func (client *GitSensorGrpcClientImpl) ReloadAllMaterial(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.ReloadAllMaterial(ctx, req)
}

func (client *GitSensorGrpcClientImpl) ReloadMaterial(ctx context.Context, req *pb.ReloadMaterialRequest) (
	*pb.GenericResponse, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.ReloadMaterial(ctx, req)
}

func (client *GitSensorGrpcClientImpl) GetChangesInRelease(ctx context.Context, req *pb.ReleaseChangeRequest) (
	*pb.GitChanges, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.GetChangesInRelease(ctx, req)
}

func (client *GitSensorGrpcClientImpl) GetWebhookData(ctx context.Context, req *pb.WebhookDataRequest) (
	*pb.WebhookAndCiData, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.GetWebhookData(ctx, req)
}

func (client *GitSensorGrpcClientImpl) GetAllWebhookEventConfigForHost(ctx context.Context, req *pb.WebhookEventConfigRequest) (
	*pb.WebhookEventConfigResponse, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.GetAllWebhookEventConfigForHost(ctx, req)
}

func (client *GitSensorGrpcClientImpl) GetWebhookEventConfig(ctx context.Context, req *pb.WebhookEventConfigRequest) (
	*pb.WebhookEventConfig, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.GetWebhookEventConfig(ctx, req)
}

func (client *GitSensorGrpcClientImpl) GetWebhookPayloadDataForPipelineMaterialId(ctx context.Context,
	req *pb.WebhookPayloadDataRequest) (*pb.WebhookPayloadDataResponse, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.GetWebhookPayloadDataForPipelineMaterialId(ctx, req)
}

func (client *GitSensorGrpcClientImpl) GetWebhookPayloadFilterDataForPipelineMaterialId(ctx context.Context,
	req *pb.WebhookPayloadFilterDataRequest) (*pb.WebhookPayloadFilterDataResponse, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, nil
	}
	return serviceClient.GetWebhookPayloadFilterDataForPipelineMaterialId(ctx, req)
}
