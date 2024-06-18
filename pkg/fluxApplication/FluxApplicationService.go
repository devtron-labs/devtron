package fluxApplication

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/fluxApplication/bean"
	"go.uber.org/zap"
)

type FluxApplicationService interface {
	DecodeFluxAppId(appId string) (*bean.FluxAppIdentifier, error)
	GetFluxAppDetail(ctx context.Context, app *bean.FluxAppIdentifier) (*bean.FluxApplicationDetailDto, error)
}

type FluxApplicationServiceImpl struct {
	logger         *zap.SugaredLogger
	helmAppService service.HelmAppService
	clusterService cluster.ClusterService
	helmAppClient  gRPC.HelmAppClient
	pump           connector.Pump
}

func NewFluxApplicationServiceImpl(logger *zap.SugaredLogger,
	helmAppService service.HelmAppService,
	helmAppClient gRPC.HelmAppClient) *FluxApplicationServiceImpl {
	return &FluxApplicationServiceImpl{
		logger:         logger,
		helmAppService: helmAppService,
		helmAppClient:  helmAppClient,
	}

}
func (impl *FluxApplicationServiceImpl) DecodeFluxAppId(appId string) (*bean.FluxAppIdentifier, error) {
	return DecodeFluxExternalAppAppId(appId)
}
func (impl *FluxApplicationServiceImpl) GetFluxAppDetail(ctx context.Context, app *bean.FluxAppIdentifier) (*bean.FluxApplicationDetailDto, error) {
	config, err := impl.helmAppService.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster config", "appIdentifier", app, "err", err)
		return nil, fmt.Errorf("failed to get cluster config for app %s in namespace %s: %w", app.Name, app.Namespace, err)
	}
	fluxDetailResponse, err := impl.getFluxAppDetailTree(ctx, config, app)
	if err != nil {
		impl.logger.Errorw("error in getting Flux app detail tree", "appIdentifier", app, "err", err)
		return nil, fmt.Errorf("failed to get Flux app detail tree for app %s in namespace %s: %w", app.Name, app.Namespace, err)
	}

	appDetail := &bean.FluxApplicationDetailDto{
		FluxApplication: &bean.FluxApplication{
			Name:                  app.Name,
			SyncStatus:            fluxDetailResponse.FluxApplication.SyncStatus,
			HealthStatus:          fluxDetailResponse.FluxApplication.HealthStatus,
			Namespace:             app.Namespace,
			ClusterId:             app.ClusterId,
			FluxAppDeploymentType: fluxDetailResponse.FluxApplication.FluxAppDeploymentType,
			ClusterName:           fluxDetailResponse.FluxApplication.EnvironmentDetail.GetClusterName(),
		},
		FluxAppStatusDetail: &bean.FluxAppStatusDetail{
			Status:  fluxDetailResponse.FluxAppStatusDetail.GetStatus(),
			Reason:  fluxDetailResponse.FluxAppStatusDetail.GetReason(),
			Message: fluxDetailResponse.FluxAppStatusDetail.GetMessage(),
		},
		ResourceTreeResponse: fluxDetailResponse.ResourceTreeResponse,
	}

	return appDetail, nil
}
func (impl *FluxApplicationServiceImpl) getFluxAppDetailTree(ctx context.Context, config *gRPC.ClusterConfig, appIdentifier *bean.FluxAppIdentifier) (*gRPC.FluxAppDetail, error) {
	req := &gRPC.FluxAppDetailRequest{
		ClusterConfig:  config,
		Namespace:      appIdentifier.Namespace,
		Name:           appIdentifier.Name,
		IsKustomizeApp: appIdentifier.IsKustomizeApp,
	}
	return impl.helmAppClient.GetExternalFluxAppDetail(ctx, req)
}
