package client

import (
	"context"
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	MegaBytes = 1024 * 1024
)

type CasbinClient interface {
	AddPolicy(ctx context.Context, in *MultiPolicyObj) (*AddPolicyResp, error)
	LoadPolicy(ctx context.Context, in *EmptyObj) (*EmptyObj, error)
	RemovePolicy(ctx context.Context, in *MultiPolicyObj) (*MultiPolicyObj, error)
	GetAllSubjects(ctx context.Context, in *EmptyObj) (*GetAllSubjectsResp, error)
	DeleteRoleForUser(ctx context.Context, in *DeleteRoleForUserRequest) (*DeleteRoleForUserResp, error)
	GetRolesForUser(ctx context.Context, in *GetRolesForUserRequest) (*GetRolesForUserResp, error)
	GetUserByRole(ctx context.Context, in *GetUserByRoleRequest) (*GetUserByRoleResp, error)
	RemovePoliciesByRole(ctx context.Context, in *RemovePoliciesByRoleRequest) (*RemovePoliciesByRoleResp, error)
	RemovePoliciesByRoles(ctx context.Context, in *RemovePoliciesByRolesRequest) (*RemovePoliciesByRolesResp, error)
}

type CasbinClientImpl struct {
	logger              *zap.SugaredLogger
	casbinClientConfig  *CasbinClientConfig
	casbinServiceClient CasbinServiceClient
}

type CasbinClientConfig struct {
	Url                       string `env:"CASBIN_CLIENT_URL" envDefault:"127.0.0.1:9000"`
	MaxSizeOfDataTransferInMb int    `env:"CASBIN_GRPC_DATA_TRANSFER_MAX_SIZE" envDefault:"30"`
}

func NewCasbinClientImpl(logger *zap.SugaredLogger,
	casbinClientConfig *CasbinClientConfig) *CasbinClientImpl {
	return &CasbinClientImpl{
		logger:             logger,
		casbinClientConfig: casbinClientConfig,
	}
}

func GetConfig() (*CasbinClientConfig, error) {
	cfg := &CasbinClientConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func (impl *CasbinClientImpl) getCasbinClient() (CasbinServiceClient, error) {
	if impl.casbinServiceClient == nil {
		connection, err := impl.getConnection()
		if err != nil {
			return nil, err
		}
		impl.casbinServiceClient = NewCasbinServiceClient(connection)
	}
	return impl.casbinServiceClient, nil
}

func (impl *CasbinClientImpl) getConnection() (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	opts = append(opts,
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(impl.casbinClientConfig.MaxSizeOfDataTransferInMb*MegaBytes),
		),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	endpoint := fmt.Sprintf("dns:///%s", impl.casbinClientConfig.Url)
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (impl *CasbinClientImpl) AddPolicy(ctx context.Context, in *MultiPolicyObj) (*AddPolicyResp, error) {
	casbinClient, err := impl.getCasbinClient()
	if err != nil {
		return nil, err
	}
	resp, err := casbinClient.AddPolicy(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (impl *CasbinClientImpl) LoadPolicy(ctx context.Context, in *EmptyObj) (*EmptyObj, error) {
	casbinClient, err := impl.getCasbinClient()
	if err != nil {
		return nil, err
	}
	resp, err := casbinClient.LoadPolicy(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (impl *CasbinClientImpl) RemovePolicy(ctx context.Context, in *MultiPolicyObj) (*MultiPolicyObj, error) {
	casbinClient, err := impl.getCasbinClient()
	if err != nil {
		return nil, err
	}
	resp, err := casbinClient.RemovePolicy(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (impl *CasbinClientImpl) GetAllSubjects(ctx context.Context, in *EmptyObj) (*GetAllSubjectsResp, error) {
	casbinClient, err := impl.getCasbinClient()
	if err != nil {
		return nil, err
	}
	resp, err := casbinClient.GetAllSubjects(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (impl *CasbinClientImpl) DeleteRoleForUser(ctx context.Context, in *DeleteRoleForUserRequest) (*DeleteRoleForUserResp, error) {
	casbinClient, err := impl.getCasbinClient()
	if err != nil {
		return nil, err
	}
	resp, err := casbinClient.DeleteRoleForUser(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (impl *CasbinClientImpl) GetRolesForUser(ctx context.Context, in *GetRolesForUserRequest) (*GetRolesForUserResp, error) {
	casbinClient, err := impl.getCasbinClient()
	if err != nil {
		return nil, err
	}
	resp, err := casbinClient.GetRolesForUser(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (impl *CasbinClientImpl) GetUserByRole(ctx context.Context, in *GetUserByRoleRequest) (*GetUserByRoleResp, error) {
	casbinClient, err := impl.getCasbinClient()
	if err != nil {
		return nil, err
	}
	resp, err := casbinClient.GetUserByRole(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (impl *CasbinClientImpl) RemovePoliciesByRole(ctx context.Context, in *RemovePoliciesByRoleRequest) (*RemovePoliciesByRoleResp, error) {
	casbinClient, err := impl.getCasbinClient()
	if err != nil {
		return nil, err
	}
	resp, err := casbinClient.RemovePoliciesByRole(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (impl *CasbinClientImpl) RemovePoliciesByRoles(ctx context.Context, in *RemovePoliciesByRolesRequest) (*RemovePoliciesByRolesResp, error) {
	casbinClient, err := impl.getCasbinClient()
	if err != nil {
		return nil, err
	}
	resp, err := casbinClient.RemovePoliciesByRoles(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
