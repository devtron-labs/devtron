package fluxApplication

import (
	"context"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/fluxApplication/bean"
	"go.uber.org/zap"
)

type FluxApplicationService interface {
	ListApplications(ctx context.Context, clusterIds []int) ([]*bean.FluxApplicationListDto, error)
}

type FluxApplicationServiceImpl struct {
	logger         *zap.SugaredLogger
	helmAppService service.HelmAppService
	clusterService cluster.ClusterService
	helmAppClient  gRPC.HelmAppClient
}

func NewFluxApplicationServiceImpl(logger *zap.SugaredLogger,
	helmAppService service.HelmAppService,
	clusterService cluster.ClusterService,
	helmAppClient gRPC.HelmAppClient) *FluxApplicationServiceImpl {
	return &FluxApplicationServiceImpl{
		logger:         logger,
		helmAppService: helmAppService,
		clusterService: clusterService,
		helmAppClient:  helmAppClient,
	}

}

func (impl *FluxApplicationServiceImpl) ListApplications(ctx context.Context, clusterIds []int) ([]*bean.FluxApplicationListDto, error) {
	var clusters []*cluster.ClusterBean
	var err error
	appListCluster := make([]*bean.FluxApplicationListDto, 0)
	req := &gRPC.AppListRequest{}
	if len(clusterIds) > 0 {
		for _, clusterId := range clusterIds {
			clusterConfig, err := impl.helmAppService.GetClusterConf(clusterId)
			if err != nil {
				impl.logger.Errorw("error in getting clusters by ids", "err", err, "clusterIds", clusterIds)
				return nil, err
			}
			req.Clusters = append(req.Clusters, clusterConfig)
		}

	} else {
		clusters, err = impl.clusterService.FindAll()
		if err != nil {
			impl.logger.Errorw("error in getting all active clusters", "err", err)
			return nil, err
		}

		configs, err1 := convertClusterBeanToClusterConfig(clusters)
		if err1 != nil {
			impl.logger.Errorw("error in getting all active clusters", "err", err1)
			return nil, err1
		}
		req.Clusters = configs

	}

	applicationStream, err := impl.helmAppClient.ListFluxApplication(ctx, req)
	if err != nil {
		impl.logger.Errorw("error while fetching list application from kubelink", "err", err)

	} else {
		var fluxApplicationList *gRPC.FluxApplicationList
		fluxApplicationList, err = applicationStream.Recv()
		if err != nil {
			impl.logger.Errorw("error in list Flux applications streams recv", "err", err)
		} else {
			appLists := fluxApplicationList.GetFluxApplicationDetail()

			for _, app := range appLists {
				fluxAppListDto := make([]*bean.FluxApplicationDto, 0)
				fluxAppListDto = append(fluxAppListDto, &bean.FluxApplicationDto{
					Name:         app.Name,
					SyncStatus:   app.SyncStatus,
					HealthStatus: app.HealthStatus,
					EnvironmentDetails: &bean.EnvironmentDetail{
						Namespace:   app.EnvironmentDetail.Namespace,
						ClusterId:   int(app.EnvironmentDetail.ClusterId),
						ClusterName: app.EnvironmentDetail.ClusterName,
					},
				})

				appListCluster = append(appListCluster, &bean.FluxApplicationListDto{
					ClusterId:  int(fluxApplicationList.ClusterId),
					FluxAppDto: fluxAppListDto,
				})
			}

		}
	}
	return appListCluster, nil
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
