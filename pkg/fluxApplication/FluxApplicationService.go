package fluxApplication

import (
	"context"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/fluxApplication/bean"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type FluxApplicationService interface {
	ListFluxApplications(ctx context.Context, clusterIds []int, w http.ResponseWriter)
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
	clusterService cluster.ClusterService,
	helmAppClient gRPC.HelmAppClient, pump connector.Pump) *FluxApplicationServiceImpl {
	return &FluxApplicationServiceImpl{
		logger:         logger,
		helmAppService: helmAppService,
		clusterService: clusterService,
		helmAppClient:  helmAppClient,
		pump:           pump,
	}

}

func (impl *FluxApplicationServiceImpl) listApplications(ctx context.Context, clusterIds []int) (gRPC.ApplicationService_ListFluxApplicationsClient, error) {
	var err error
	req := &gRPC.AppListRequest{}

	if len(clusterIds) <= 0 {
		err = errors.New("no clusterIds provided")
		impl.logger.Errorw("please provide any valid clusterIds", "err", err)
		return nil, err
	}
	if len(clusterIds) > 0 {
		for _, clusterId := range clusterIds {
			clusterConfig, err := impl.helmAppService.GetClusterConf(clusterId)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("error in getting clusters by ids", "err", err, "clusterIds", clusterIds)
				return nil, err
			} else if err != nil && util.IsErrNoRows(err) {
				errMsg := fmt.Sprintf("cluster id %d not found", clusterId)
				return nil, util.NewApiError().WithHttpStatusCode(http.StatusNotFound).WithInternalMessage(errMsg).WithUserMessage(errMsg).WithCode(strconv.Itoa(http.StatusNotFound))
			}
			req.Clusters = append(req.Clusters, clusterConfig)
		}
	}
	applicationStream, err := impl.helmAppClient.ListFluxApplication(ctx, req)

	return applicationStream, err
}

func (impl *FluxApplicationServiceImpl) ListFluxApplications(ctx context.Context, clusterIds []int, w http.ResponseWriter) {
	appStream, err := impl.listApplications(ctx, clusterIds)
	impl.pump.StartStreamWithTransformer(w, func() (proto.Message, error) {
		return appStream.Recv()
	}, err,
		func(message interface{}) interface{} {
			return impl.appListRespProtoTransformer(message.(*gRPC.FluxApplicationList))
		})
}

func (impl *FluxApplicationServiceImpl) appListRespProtoTransformer(deployedApps *gRPC.FluxApplicationList) bean.FluxAppList {
	appList := bean.FluxAppList{ClusterId: &deployedApps.ClusterId}

	fluxApps := make([]bean.FluxApplication, 0)

	for _, deployedapp := range deployedApps.FluxApplication {

		fluxApp := bean.FluxApplication{
			Name:                  deployedapp.Name,
			HealthStatus:          deployedapp.HealthStatus,
			SyncStatus:            deployedapp.SyncStatus,
			ClusterId:             int(deployedapp.EnvironmentDetail.ClusterId),
			ClusterName:           deployedapp.EnvironmentDetail.ClusterName,
			Namespace:             deployedapp.EnvironmentDetail.Namespace,
			FluxAppDeploymentType: deployedapp.FluxAppDeploymentType,
		}
		fluxApps = append(fluxApps, fluxApp)
	}
	appList.FluxApps = &fluxApps
	return appList
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
