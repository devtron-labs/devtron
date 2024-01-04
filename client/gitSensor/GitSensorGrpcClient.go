package gitSensor

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/bean"
	pb "github.com/devtron-labs/protos/gitSensor"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

const (
	ContextTimeoutInSeconds = 10
)

type ApiClient interface {
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

type GrpcApiClientImpl struct {
	logger        *zap.SugaredLogger
	config        *ClientConfig
	serviceClient pb.GitSensorServiceClient
}

func NewGitSensorGrpcClientImpl(logger *zap.SugaredLogger, config *ClientConfig) (*GrpcApiClientImpl, error) {

	return &GrpcApiClientImpl{
		logger: logger,
		config: config,
	}, nil
}

// getGitSensorServiceClient initializes and returns gRPC GitSensorService client
func (client *GrpcApiClientImpl) getGitSensorServiceClient() (pb.GitSensorServiceClient, error) {
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
func (client *GrpcApiClientImpl) getConnection() (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeoutInSeconds*time.Second)
	defer cancel()

	// Configure gRPC dial options
	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithChainUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor, otelgrpc.UnaryClientInterceptor()),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
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

// SaveGitProvider saves Git provider
func (client *GrpcApiClientImpl) SaveGitProvider(ctx context.Context, provider *GitProvider) error {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return err
	}
	// map req
	req := &pb.GitProvider{
		Id:            int64(provider.Id),
		Name:          provider.Name,
		Url:           provider.Url,
		UserName:      provider.UserName,
		Password:      provider.Password,
		AccessToken:   provider.AccessToken,
		SshPrivateKey: provider.SshPrivateKey,
		AuthMode:      string(provider.AuthMode),
		Active:        provider.Active,
	}

	// fetch
	_, err = serviceClient.SaveGitProvider(ctx, req)
	return err
}

// AddRepo adds git materials
func (client *GrpcApiClientImpl) AddRepo(ctx context.Context, materials []*GitMaterial) error {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return err
	}

	// Mapping req to proto type
	var gitMaterials []*pb.GitMaterial
	if materials != nil {
		gitMaterials = make([]*pb.GitMaterial, 0, len(materials))
		for _, item := range materials {

			gitMaterials = append(gitMaterials, &pb.GitMaterial{
				Id:               int64(item.Id),
				GitProviderId:    int64(item.GitProviderId),
				Url:              item.Url,
				FetchSubmodules:  item.FetchSubmodules,
				Name:             item.Name,
				CheckoutLocation: item.CheckoutLocation,
				CheckoutStatus:   item.CheckoutStatus,
				CheckoutMsgAny:   item.CheckoutMsgAny,
				Deleted:          item.Deleted,
				FilterPattern:    item.FilterPattern,
			})
		}
	}

	_, err = serviceClient.AddRepo(ctx, &pb.AddRepoRequest{
		GitMaterialList: gitMaterials,
	})
	return err
}

// UpdateRepo updates the git material
func (client *GrpcApiClientImpl) UpdateRepo(ctx context.Context, material *GitMaterial) error {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return err
	}

	// mapping req
	mappedMaterial := &pb.GitMaterial{
		Id:               int64(material.Id),
		GitProviderId:    int64(material.GitProviderId),
		Url:              material.Url,
		FetchSubmodules:  material.FetchSubmodules,
		Name:             material.Name,
		CheckoutLocation: material.CheckoutLocation,
		CheckoutStatus:   material.CheckoutStatus,
		CheckoutMsgAny:   material.CheckoutMsgAny,
		Deleted:          material.Deleted,
		FilterPattern:    material.FilterPattern,
	}

	_, err = serviceClient.UpdateRepo(ctx, mappedMaterial)
	return err
}

// SavePipelineMaterial saves ci pipeline material info
func (client *GrpcApiClientImpl) SavePipelineMaterial(ctx context.Context, ciPipelineMaterials []*CiPipelineMaterial) error {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return err
	}

	// Mapping request
	var mappedCiPipelineMaterials []*pb.CiPipelineMaterial
	if ciPipelineMaterials != nil {
		mappedCiPipelineMaterials = make([]*pb.CiPipelineMaterial, 0, len(ciPipelineMaterials))
		for _, item := range ciPipelineMaterials {

			mappedCiPipelineMaterials = append(mappedCiPipelineMaterials, &pb.CiPipelineMaterial{
				Id:            int64(item.Id),
				GitMaterialId: int64(item.GitMaterialId),
				Type:          string(item.Type),
				Value:         item.Value,
				Active:        item.Active,
			})
		}
	}

	_, err = serviceClient.SavePipelineMaterial(ctx, &pb.SavePipelineMaterialRequest{
		CiPipelineMaterials: mappedCiPipelineMaterials,
	})
	return err
}

func (client *GrpcApiClientImpl) FetchChanges(ctx context.Context, req *FetchScmChangesRequest) (
	*MaterialChangeResp, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}

	res, err := serviceClient.FetchChanges(ctx, &pb.FetchScmChangesRequest{
		PipelineMaterialId: int64(req.PipelineMaterialId),
		From:               req.From,
		To:                 req.To,
		ShowAll:            req.ShowAll,
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	// Mapping res
	var commits []*GitCommit
	if res.Commits != nil {
		commits = make([]*GitCommit, 0, len(res.Commits))
		for _, item := range res.Commits {
			commit := client.mapGitCommitToLocalType(item)
			commits = append(commits, &commit)
		}
	}

	mappedRes := &MaterialChangeResp{
		Commits:        commits,
		IsRepoError:    res.IsRepoError,
		RepoErrorMsg:   res.RepoErrorMsg,
		IsBranchError:  res.IsBranchError,
		BranchErrorMsg: res.BranchErrorMsg,
	}
	if res.LastFetchTime != nil {
		mappedRes.LastFetchTime = res.LastFetchTime.AsTime()
	}
	return mappedRes, nil
}

func (client *GrpcApiClientImpl) GetHeadForPipelineMaterials(ctx context.Context, req *HeadRequest) (
	[]*CiPipelineMaterial, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}

	// mapping req
	var materialIds []int64
	if req.MaterialIds != nil {
		materialIds = make([]int64, 0, len(req.MaterialIds))
		for _, item := range req.MaterialIds {
			materialIds = append(materialIds, int64(item))
		}
	}

	res, err := serviceClient.GetHeadForPipelineMaterials(ctx, &pb.HeadRequest{MaterialIds: materialIds})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	// Mapping res
	var materials []*CiPipelineMaterial
	if res.Materials != nil {
		materials = make([]*CiPipelineMaterial, 0, len(res.Materials))
		for _, item := range res.Materials {

			materials = append(materials, &CiPipelineMaterial{
				Id:                        int(item.Id),
				GitMaterialId:             int(item.GitMaterialId),
				Type:                      SourceType(item.Type),
				Value:                     item.Value,
				Active:                    item.Active,
				GitCommit:                 client.mapGitCommitToLocalType(item.GitCommit),
				ExtraEnvironmentVariables: item.ExtraEnvironmentVariables,
			})
		}
	}
	return materials, nil
}

func (client *GrpcApiClientImpl) GetCommitMetadata(ctx context.Context, req *CommitMetadataRequest) (
	*GitCommit, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}

	// fetch
	res, err := serviceClient.GetCommitMetadata(ctx, &pb.CommitMetadataRequest{
		PipelineMaterialId: int64(req.PipelineMaterialId),
		GitHash:            req.GitHash,
		GitTag:             req.GitTag,
		BranchName:         req.BranchName,
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	// mapping res
	commit := client.mapGitCommitToLocalType(res)
	return &commit, nil
}

func (client *GrpcApiClientImpl) GetCommitMetadataForPipelineMaterial(ctx context.Context, req *CommitMetadataRequest) (
	*GitCommit, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}

	// fetch
	res, err := serviceClient.GetCommitMetadataForPipelineMaterial(ctx, &pb.CommitMetadataRequest{
		PipelineMaterialId: int64(req.PipelineMaterialId),
		GitHash:            req.GitHash,
		GitTag:             req.GitTag,
		BranchName:         req.BranchName,
	})
	if err != nil && err.Error() == bean.ErrNoCommitFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// mapping res
	commit := client.mapGitCommitToLocalType(res)
	return &commit, nil
}

func (client *GrpcApiClientImpl) GetCommitInfoForTag(ctx context.Context, req *CommitMetadataRequest) (
	*GitCommit, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}
	res, err := serviceClient.GetCommitInfoForTag(ctx, &pb.CommitMetadataRequest{
		PipelineMaterialId: int64(req.PipelineMaterialId),
		GitHash:            req.GitHash,
		GitTag:             req.GitTag,
		BranchName:         req.BranchName,
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	// mapping res
	commit := client.mapGitCommitToLocalType(res)
	return &commit, nil
}

func (client *GrpcApiClientImpl) RefreshGitMaterial(ctx context.Context, req *RefreshGitMaterialRequest) (
	*RefreshGitMaterialResponse, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}
	res, err := serviceClient.RefreshGitMaterial(ctx, &pb.RefreshGitMaterialRequest{
		GitMaterialId: int64(req.GitMaterialId),
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	// mapping res
	mappedRes := &RefreshGitMaterialResponse{
		Message:  res.Message,
		ErrorMsg: res.ErrorMsg,
	}
	if res.LastFetchTime != nil {
		mappedRes.LastFetchTime = res.LastFetchTime.AsTime()
	}
	return mappedRes, nil
}

func (client *GrpcApiClientImpl) ReloadAllMaterial(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}
	return serviceClient.ReloadAllMaterial(ctx, req)
}

func (client *GrpcApiClientImpl) ReloadMaterial(ctx context.Context, materialId int64) (
	*pb.GenericResponse, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}
	return serviceClient.ReloadMaterial(ctx, &pb.ReloadMaterialRequest{
		MaterialId: materialId,
	})
}

func (client *GrpcApiClientImpl) GetChangesInRelease(ctx context.Context, req *pb.ReleaseChangeRequest) (
	*pb.GitChanges, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}
	return serviceClient.GetChangesInRelease(ctx, req)
}

func (client *GrpcApiClientImpl) GetWebhookData(ctx context.Context, req *WebhookDataRequest) (
	*WebhookAndCiData, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}

	res, err := serviceClient.GetWebhookData(ctx, &pb.WebhookDataRequest{
		Id:                   int64(req.Id),
		CiPipelineMaterialId: int64(req.CiPipelineMaterialId),
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	// mapping res
	mappedRes := &WebhookAndCiData{
		ExtraEnvironmentVariables: res.ExtraEnvironmentVariables,
	}
	if res.WebhookData != nil {
		mappedRes.WebhookData = &WebhookData{
			Id:              int(res.WebhookData.Id),
			EventActionType: res.WebhookData.EventActionType,
			Data:            res.WebhookData.Data,
		}
	}
	return mappedRes, nil
}

func (client *GrpcApiClientImpl) GetAllWebhookEventConfigForHost(ctx context.Context, req *WebhookEventConfigRequest) (
	[]*WebhookEventConfig, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}

	res, err := serviceClient.GetAllWebhookEventConfigForHost(ctx, &pb.WebhookEventConfigRequest{
		GitHostId: int64(req.GitHostId),
		EventId:   int64(req.EventId),
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	// mapping res
	var mappedRes []*WebhookEventConfig
	if res.WebhookEventConfig != nil {
		mappedRes = make([]*WebhookEventConfig, 0, len(res.WebhookEventConfig))
		for _, item := range res.WebhookEventConfig {
			mappedRes = append(mappedRes, client.mapWebhookEventConfigToLocalType(item))
		}
	}
	return mappedRes, nil
}

func (client *GrpcApiClientImpl) GetWebhookEventConfig(ctx context.Context, req *WebhookEventConfigRequest) (
	*WebhookEventConfig, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}
	res, err := serviceClient.GetWebhookEventConfig(ctx, &pb.WebhookEventConfigRequest{
		GitHostId: int64(req.GitHostId),
		EventId:   int64(req.EventId),
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return client.mapWebhookEventConfigToLocalType(res), nil
}

func (client *GrpcApiClientImpl) GetWebhookPayloadDataForPipelineMaterialId(ctx context.Context,
	req *WebhookPayloadDataRequest) (*WebhookPayloadDataResponse, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}

	res, err := serviceClient.GetWebhookPayloadDataForPipelineMaterialId(ctx, &pb.WebhookPayloadDataRequest{
		CiPipelineMaterialId: int64(req.CiPipelineMaterialId),
		Limit:                int64(req.Limit),
		Offset:               int64(req.Offset),
		EventTimeSortOrder:   req.EventTimeSortOrder,
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	// mapping res
	var payloads []*WebhookPayloadDataPayloadsResponse
	if res.Payloads != nil {
		payloads = make([]*WebhookPayloadDataPayloadsResponse, 0, len(res.Payloads))
		for _, item := range res.Payloads {

			payload := &WebhookPayloadDataPayloadsResponse{
				ParsedDataId:        int(item.ParsedDataId),
				MatchedFiltersCount: int(item.MatchedFiltersCount),
				FailedFiltersCount:  int(item.FailedFiltersCount),
				MatchedFilters:      item.MatchedFilters,
			}
			if item.EventTime != nil {
				payload.EventTime = item.EventTime.AsTime()
			}
			payloads = append(payloads, payload)
		}
	}

	return &WebhookPayloadDataResponse{
		Filters:       res.Filters,
		RepositoryUrl: res.RepositoryUrl,
		Payloads:      payloads,
	}, nil
}

func (client *GrpcApiClientImpl) GetWebhookPayloadFilterDataForPipelineMaterialId(ctx context.Context,
	req *WebhookPayloadFilterDataRequest) (*WebhookPayloadFilterDataResponse, error) {

	serviceClient, err := client.getGitSensorServiceClient()
	if err != nil {
		return nil, err
	}

	res, err := serviceClient.GetWebhookPayloadFilterDataForPipelineMaterialId(ctx, &pb.WebhookPayloadFilterDataRequest{
		CiPipelineMaterialId: int64(req.CiPipelineMaterialId),
		ParsedDataId:         int64(req.ParsedDataId),
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	// mapping res
	var selectors []*WebhookPayloadFilterDataSelectorResponse
	if res.SelectorsData != nil {
		selectors = make([]*WebhookPayloadFilterDataSelectorResponse, 0, len(res.SelectorsData))
		for _, item := range res.SelectorsData {

			selectors = append(selectors, &WebhookPayloadFilterDataSelectorResponse{
				SelectorName:      item.SelectorName,
				SelectorCondition: item.SelectorCondition,
				SelectorValue:     item.SelectorValue,
				Match:             item.Match,
			})
		}
	}

	mappedRes := &WebhookPayloadFilterDataResponse{
		PayloadId:     int(res.PayloadId),
		PayloadJson:   res.PayloadJson,
		SelectorsData: selectors,
	}
	return mappedRes, nil
}

func (client *GrpcApiClientImpl) mapWebhookEventConfigToLocalType(config *pb.WebhookEventConfig) *WebhookEventConfig {

	var selectors []*WebhookEventSelectors
	if config.Selectors != nil {
		selectors = make([]*WebhookEventSelectors, 0, len(config.Selectors))
		for _, item := range config.Selectors {

			selector := &WebhookEventSelectors{
				Id:               int(item.Id),
				EventId:          int(item.EventId),
				Name:             item.Name,
				ToShow:           item.ToShow,
				ToShowInCiFilter: item.ToShowInCiFilter,
				FixValue:         item.FixValue,
				PossibleValues:   item.PossibleValues,
				IsActive:         item.IsActive,
			}
			if item.CreatedOn != nil {
				selector.CreatedOn = item.CreatedOn.AsTime()
			}
			if item.UpdatedOn != nil {
				selector.UpdatedOn = item.UpdatedOn.AsTime()
			}
			selectors = append(selectors, selector)
		}
	}

	mappedConfig := &WebhookEventConfig{
		Id:            int(config.Id),
		GitHostId:     int(config.GitHostId),
		Name:          config.Name,
		EventTypesCsv: config.EventTypesCsv,
		ActionType:    config.ActionType,
		IsActive:      config.IsActive,
		Selectors:     selectors,
	}
	if config.CreatedOn != nil {
		mappedConfig.CreatedOn = config.CreatedOn.AsTime()
	}
	if config.UpdatedOn != nil {
		mappedConfig.UpdatedOn = config.UpdatedOn.AsTime()
	}
	return mappedConfig
}

// mapGitCommitToLocalType maps the protobuf specified GitCommit to local specified golang based struct
func (client *GrpcApiClientImpl) mapGitCommitToLocalType(commit *pb.GitCommit) GitCommit {

	mappedCommit := GitCommit{
		Commit:   commit.Commit,
		Author:   commit.Author,
		Message:  commit.Message,
		Changes:  commit.Changes,
		Excluded: commit.Excluded,
	}
	if commit.Date != nil {
		mappedCommit.Date = commit.Date.AsTime()
	}
	if commit.WebhookData != nil {
		mappedCommit.WebhookData = &WebhookData{
			Id:              int(commit.WebhookData.Id),
			EventActionType: commit.WebhookData.EventActionType,
			Data:            commit.WebhookData.Data,
		}
	}
	return mappedCommit
}

func (client *GrpcApiClientImpl) mapWebhookEventConfigToProtoType(config *WebhookEventConfig) *pb.WebhookEventConfig {

	var selectors []*pb.WebhookEventSelectors
	if config.Selectors != nil {
		selectors = make([]*pb.WebhookEventSelectors, 0, len(config.Selectors))
		for _, item := range config.Selectors {

			selector := &pb.WebhookEventSelectors{
				Id:               int64(item.Id),
				EventId:          int64(item.EventId),
				Name:             item.Name,
				ToShow:           item.ToShow,
				ToShowInCiFilter: item.ToShowInCiFilter,
				FixValue:         item.FixValue,
				PossibleValues:   item.PossibleValues,
				IsActive:         item.IsActive,
			}
			if !item.CreatedOn.IsZero() {
				selector.CreatedOn = timestamppb.New(item.CreatedOn)
			}
			if !item.UpdatedOn.IsZero() {
				selector.UpdatedOn = timestamppb.New(item.UpdatedOn)
			}
			selectors = append(selectors, selector)
		}
	}

	mappedConfig := &pb.WebhookEventConfig{
		Id:            int64(config.Id),
		GitHostId:     int64(config.GitHostId),
		Name:          config.Name,
		EventTypesCsv: config.EventTypesCsv,
		ActionType:    config.ActionType,
		IsActive:      config.IsActive,
		Selectors:     selectors,
	}
	if !config.CreatedOn.IsZero() {
		mappedConfig.CreatedOn = timestamppb.New(config.CreatedOn)
	}
	if !config.UpdatedOn.IsZero() {
		mappedConfig.UpdatedOn = timestamppb.New(config.UpdatedOn)
	}
	return mappedConfig
}

func (client *GrpcApiClientImpl) mapGitCommitToProtoType(commit *GitCommit) (*pb.GitCommit, error) {

	// Mapping GitCommit
	mappedRes := &pb.GitCommit{
		Commit:  commit.Commit,
		Author:  commit.Author,
		Message: commit.Message,
		Changes: commit.Changes,
	}

	if commit.WebhookData != nil {
		mappedRes.WebhookData = &pb.WebhookData{
			Id:              int64(commit.WebhookData.Id),
			EventActionType: commit.WebhookData.EventActionType,
			Data:            commit.WebhookData.Data,
		}
	}
	if !commit.Date.IsZero() {
		mappedRes.Date = timestamppb.New(commit.Date)
	}
	return mappedRes, nil
}
