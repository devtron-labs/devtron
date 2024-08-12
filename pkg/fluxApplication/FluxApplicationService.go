package fluxApplication

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/fluxApplication/bean"
	"github.com/gogo/protobuf/proto"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"net/http"
)

type FluxApplicationService interface {
	ListFluxApplications(ctx context.Context, clusterIds []int, w http.ResponseWriter)
	GetFluxAppDetail(ctx context.Context, app *bean.FluxAppIdentifier) (*bean.FluxApplicationDetailDto, error)
	HibernateFluxApplication(ctx context.Context, app *bean.FluxAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
	UnHibernateFluxApplication(ctx context.Context, app *bean.FluxAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
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
func (impl *FluxApplicationServiceImpl) HibernateFluxApplication(ctx context.Context, app *bean.FluxAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	conf, err := impl.helmAppService.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("HibernateFluxApplication", "error in getting the cluster config", err, "clusterId", app.ClusterId, "appName", app.Name)
		return nil, err
	}
	req := service.HibernateReqAdaptor(hibernateRequest)
	req.ClusterConfig = conf
	res, err := impl.helmAppClient.Hibernate(ctx, req)
	if err != nil {
		impl.logger.Errorw("HibernateFluxApplication", "error in hibernating the requested resource", err, "clusterId", app.ClusterId, "appName", app.Name)
		return nil, err
	}
	response := service.HibernateResponseAdaptor(res.Status)
	return response, nil
}

func (impl *FluxApplicationServiceImpl) UnHibernateFluxApplication(ctx context.Context, app *bean.FluxAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {

	conf, err := impl.helmAppService.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("UnHibernateFluxApplication", "error in getting the cluster config", err, "clusterId", app.ClusterId, "appName", app.Name)
		return nil, err
	}
	req := service.HibernateReqAdaptor(hibernateRequest)
	req.ClusterConfig = conf
	res, err := impl.helmAppClient.UnHibernate(ctx, req)
	if err != nil {
		impl.logger.Errorw("UnHibernateFluxApplication", "error in unHibernating the requested resources", err, "clusterId", app.ClusterId, "appName", app.Name)
		return nil, err
	}
	response := service.HibernateResponseAdaptor(res.Status)
	return response, nil
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
func (impl *FluxApplicationServiceImpl) listApplications(ctx context.Context, clusterIds []int) (gRPC.ApplicationService_ListFluxApplicationsClient, error) {
	var err error
	req := &gRPC.AppListRequest{}
	if len(clusterIds) == 0 {
		return nil, nil
	}
	_, span := otel.Tracer("clusterService").Start(ctx, "FindByIds")
	clusters, err := impl.clusterService.FindByIds(clusterIds)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}

	for _, clusterDetail := range clusters {
		config := &gRPC.ClusterConfig{
			ApiServerUrl:          clusterDetail.ServerUrl,
			Token:                 clusterDetail.Config[k8s.BearerToken],
			ClusterId:             int32(clusterDetail.Id),
			ClusterName:           clusterDetail.ClusterName,
			InsecureSkipTLSVerify: clusterDetail.InsecureSkipTLSVerify,
		}
		if clusterDetail.InsecureSkipTLSVerify == false {
			config.KeyData = clusterDetail.Config[k8s.TlsKey]
			config.CertData = clusterDetail.Config[k8s.CertData]
			config.CaData = clusterDetail.Config[k8s.CertificateAuthorityData]
		}
		req.Clusters = append(req.Clusters, config)
	}
	applicationStream, err := impl.helmAppClient.ListFluxApplication(ctx, req)

	return applicationStream, err
}
func (impl *FluxApplicationServiceImpl) appListRespProtoTransformer(deployedApps *gRPC.FluxApplicationList) bean.FluxAppList {
	appList := bean.FluxAppList{ClusterId: &[]int32{deployedApps.ClusterId}}
	if deployedApps.Errored {
		appList.Errored = &deployedApps.Errored
		appList.ErrorMsg = &deployedApps.ErrorMsg
	} else {
		fluxApps := make([]bean.FluxApplication, 0, len(deployedApps.FluxApplication))
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
	}

	return appList
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
