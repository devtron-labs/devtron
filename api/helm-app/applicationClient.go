package client

import (
	"context"
	"fmt"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

type HelmAppClient interface {
	ListApplication(req *AppListRequest) (ApplicationService_ListApplicationsClient, error)
	GetAppDetail(ctx context.Context, in *AppDetailRequest) (*AppDetail, error)
	Hibernate(ctx context.Context, in *HibernateRequest) (*HibernateResponse, error)
	UnHibernate(ctx context.Context, in *HibernateRequest) (*HibernateResponse, error)
	GetDeploymentHistory(ctx context.Context, in *AppDetailRequest) (*HelmAppDeploymentHistory, error)
	GetValuesYaml(ctx context.Context, in *AppDetailRequest) (*ReleaseInfo, error)
	GetDesiredManifest(ctx context.Context, in *ObjectRequest) (*DesiredManifestResponse, error)
	DeleteApplication(ctx context.Context, in *ReleaseIdentifier) (*UninstallReleaseResponse, error)
	UpdateApplication(ctx context.Context, in *UpgradeReleaseRequest) (*UpgradeReleaseResponse, error)
	GetDeploymentDetail(ctx context.Context, in *DeploymentDetailRequest) (*DeploymentDetailResponse, error)
	InstallRelease(ctx context.Context, in *InstallReleaseRequest) (*InstallReleaseResponse, error)
	UpdateApplicationWithChartInfo(ctx context.Context, in *InstallReleaseRequest) (*UpgradeReleaseResponse, error)
}

type HelmAppClientImpl struct {
	logger                   *zap.SugaredLogger
	helmClientConfig         *HelmClientConfig
	applicationServiceClient ApplicationServiceClient
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

func (impl *HelmAppClientImpl) getApplicationClient() (ApplicationServiceClient, error) {
	if impl.applicationServiceClient == nil {
		connection, err := impl.getConnection()
		if err != nil {
			return nil, err
		}
		impl.applicationServiceClient = NewApplicationServiceClient(connection)
	}
	return impl.applicationServiceClient, nil
}

func (impl *HelmAppClientImpl) getConnection() (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	opts = append(opts,
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(20*1024*1024),
		),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	endpoint := fmt.Sprintf("dns:///%s", impl.helmClientConfig.Url)
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (impl *HelmAppClientImpl) ListApplication(req *AppListRequest) (ApplicationService_ListApplicationsClient, error) {
	applicationClient, err := impl.getApplicationClient()
	if err != nil {
		return nil, err
	}
	stream, err := applicationClient.ListApplications(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

///	GetAppDetail(ctx context.Context, in *AppDetailRequest, opts ...grpc.CallOption) (*AppDetail, error)

func (impl *HelmAppClientImpl) GetAppDetail(ctx context.Context, in *AppDetailRequest) (*AppDetail, error) {
	applicationClient, err := impl.getApplicationClient()
	if err != nil {
		return nil, err
	}
	detail, err := applicationClient.GetAppDetail(ctx, in)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func (impl *HelmAppClientImpl) Hibernate(ctx context.Context, in *HibernateRequest) (*HibernateResponse, error) {
	applicationClient, err := impl.getApplicationClient()
	if err != nil {
		return nil, err
	}
	detail, err := applicationClient.Hibernate(ctx, in)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func (impl *HelmAppClientImpl) UnHibernate(ctx context.Context, in *HibernateRequest) (*HibernateResponse, error) {
	applicationClient, err := impl.getApplicationClient()
	if err != nil {
		return nil, err
	}
	detail, err := applicationClient.UnHibernate(ctx, in)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func (impl *HelmAppClientImpl) GetDeploymentHistory(ctx context.Context, in *AppDetailRequest) (*HelmAppDeploymentHistory, error) {
	applicationClient, err := impl.getApplicationClient()
	if err != nil {
		return nil, err
	}
	history, err := applicationClient.GetDeploymentHistory(ctx, in)
	if err != nil {
		return nil, err
	}
	return history, nil
}

func (impl *HelmAppClientImpl) GetValuesYaml(ctx context.Context, in *AppDetailRequest) (*ReleaseInfo, error) {
	applicationClient, err := impl.getApplicationClient()
	if err != nil {
		return nil, err
	}
	values, err := applicationClient.GetValuesYaml(ctx, in)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func (impl *HelmAppClientImpl) GetDesiredManifest(ctx context.Context, in *ObjectRequest) (*DesiredManifestResponse, error) {
	applicationClient, err := impl.getApplicationClient()
	if err != nil {
		return nil, err
	}
	manifest, err := applicationClient.GetDesiredManifest(ctx, in)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

func (impl *HelmAppClientImpl) DeleteApplication(ctx context.Context, in *ReleaseIdentifier) (*UninstallReleaseResponse, error) {
	applicationClient, err := impl.getApplicationClient()
	if err != nil {
		return nil, err
	}
	manifest, err := applicationClient.UninstallRelease(ctx, in)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

func (impl *HelmAppClientImpl) UpdateApplication(ctx context.Context, in *UpgradeReleaseRequest) (*UpgradeReleaseResponse, error) {
	applicationClient, err := impl.getApplicationClient()
	if err != nil {
		return nil, err
	}
	manifest, err := applicationClient.UpgradeRelease(ctx, in)
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

func (impl *HelmAppClientImpl) GetDeploymentDetail(ctx context.Context, in *DeploymentDetailRequest) (*DeploymentDetailResponse, error) {
	applicationClient, err := impl.getApplicationClient()
	if err != nil {
		return nil, err
	}
	deploymentDetail, err := applicationClient.GetDeploymentDetail(ctx, in)
	if err != nil {
		return nil, err
	}
	return deploymentDetail, nil
}

func (impl *HelmAppClientImpl) InstallRelease(ctx context.Context, in *InstallReleaseRequest) (*InstallReleaseResponse, error) {
	applicationClient, err := impl.getApplicationClient()
	if err != nil {
		return nil, err
	}
	installReleaseResponse, err := applicationClient.InstallRelease(ctx, in)
	if err != nil {
		return nil, err
	}
	return installReleaseResponse, nil
}

func (impl *HelmAppClientImpl) UpdateApplicationWithChartInfo(ctx context.Context, in *InstallReleaseRequest) (*UpgradeReleaseResponse, error) {
	applicationClient, err := impl.getApplicationClient()
	if err != nil {
		return nil, err
	}
	updateReleaseResponse, err := applicationClient.UpgradeReleaseWithChartInfo(ctx, in)
	if err != nil {
		return nil, err
	}
	return updateReleaseResponse, nil
}
