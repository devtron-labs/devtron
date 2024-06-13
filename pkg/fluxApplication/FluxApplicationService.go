package fluxApplication

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
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
	var clusters []*cluster.ClusterBean
	var err error
	req := &gRPC.AppListRequest{}
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

	} else {
		clusters, err = impl.clusterService.FindAll()
		if err != nil {
			impl.logger.Errorw("error in getting all active clusters", "err", err)
			return nil, err
		}
		var configs []*gRPC.ClusterConfig
		configs, err = convertClusterBeanToClusterConfig(clusters)
		if err != nil {
			impl.logger.Errorw("error in getting all active clusters", "err", err)
			return nil, err
		}
		req.Clusters = configs

	}
	applicatonStream, err := impl.helmAppClient.ListFluxApplication(ctx, req)

	return applicatonStream, err
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

	for _, deployedapp := range deployedApps.FluxApplicationDetail {

		fluxApp := bean.FluxApplication{
			Name:           deployedapp.Name,
			HealthStatus:   deployedapp.HealthStatus,
			SyncStatus:     deployedapp.SyncStatus,
			ClusterId:      int(deployedapp.EnvironmentDetail.ClusterId),
			ClusterName:    deployedapp.EnvironmentDetail.ClusterName,
			Namespace:      deployedapp.EnvironmentDetail.Namespace,
			IsKustomizeApp: deployedapp.IsKustomizeApp,
		}
		fluxApps = append(fluxApps, fluxApp)
	}
	appList.FluxApps = &fluxApps
	return appList
}

func convertClusterBeanToClusterConfig(clusters []*cluster.ClusterBean) ([]*gRPC.ClusterConfig, error) {
	clusterConfigs := make([]*gRPC.ClusterConfig, 0)
	if len(clusters) > 0 {
		for _, cluster := range clusters {

			config := &gRPC.ClusterConfig{
				ApiServerUrl:          cluster.ServerUrl,
				Token:                 cluster.Config[k8s.BearerToken],
				ClusterId:             int32(cluster.Id),
				ClusterName:           cluster.ClusterName,
				InsecureSkipTLSVerify: cluster.InsecureSkipTLSVerify,
			}
			if cluster.InsecureSkipTLSVerify == false {
				config.KeyData = cluster.Config[k8s.TlsKey]
				config.CertData = cluster.Config[k8s.CertData]
				config.CaData = cluster.Config[k8s.CertificateAuthorityData]
			}

			clusterConfigs = append(clusterConfigs, config)
		}
	}
	return clusterConfigs, nil
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
			Name:           app.Name,
			Namespace:      app.Namespace,
			ClusterId:      app.ClusterId,
			IsKustomizeApp: app.IsKustomizeApp,
		},
		FluxAppStatusDetail: &bean.FluxAppStatusDetail{
			Status:  fluxDetailResponse.FluxAppStatusDetail.GetStatus(),
			Reason:  fluxDetailResponse.FluxAppStatusDetail.GetReason(),
			Message: fluxDetailResponse.FluxAppStatusDetail.GetMessage(),
		},
		ResourceTreeArray: fluxDetailResponse.ResourceTreeResponse,
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
