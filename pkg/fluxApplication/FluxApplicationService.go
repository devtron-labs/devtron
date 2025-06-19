package fluxApplication

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/api/helm-app/service/read"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/fluxApplication/bean"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"github.com/gogo/protobuf/proto"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type FluxApplicationService interface {
	ListFluxApplications(ctx context.Context, clusterIds []int, noStream bool, w http.ResponseWriter)
	GetFluxAppDetail(ctx context.Context, app *bean.FluxAppIdentifier) (*bean.FluxApplicationDetailDto, error)
	HibernateFluxApplication(ctx context.Context, app *bean.FluxAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
	UnHibernateFluxApplication(ctx context.Context, app *bean.FluxAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
}

type FluxApplicationServiceImpl struct {
	logger                 *zap.SugaredLogger
	helmAppReadService     read.HelmAppReadService
	clusterService         cluster.ClusterService
	helmAppClient          gRPC.HelmAppClient
	pump                   connector.Pump
	pipelineRepository     pipelineConfig.PipelineRepository
	installedAppRepository repository.InstalledAppRepository
}

func NewFluxApplicationServiceImpl(logger *zap.SugaredLogger,
	helmAppReadService read.HelmAppReadService,
	clusterService cluster.ClusterService,
	helmAppClient gRPC.HelmAppClient, pump connector.Pump,
	pipelineRepository pipelineConfig.PipelineRepository,
	installedAppRepository repository.InstalledAppRepository) *FluxApplicationServiceImpl {
	return &FluxApplicationServiceImpl{
		logger:                 logger,
		helmAppReadService:     helmAppReadService,
		clusterService:         clusterService,
		helmAppClient:          helmAppClient,
		pump:                   pump,
		pipelineRepository:     pipelineRepository,
		installedAppRepository: installedAppRepository,
	}

}
func (impl *FluxApplicationServiceImpl) HibernateFluxApplication(ctx context.Context, app *bean.FluxAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	conf, err := impl.helmAppReadService.GetClusterConf(app.ClusterId)
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

	conf, err := impl.helmAppReadService.GetClusterConf(app.ClusterId)
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

func (impl *FluxApplicationServiceImpl) ListFluxApplications(ctx context.Context, clusterIds []int, noStream bool, w http.ResponseWriter) {
	appStream, err := impl.listApplications(ctx, clusterIds)
	if err != nil {
		impl.logger.Errorw("error in listing flux applications", "clusterIds", clusterIds, "err", err)
		return
	}

	fluxCdPipelines, err := impl.pipelineRepository.GetAppAndEnvDetailsForDeploymentAppTypePipeline(util.PIPELINE_DEPLOYMENT_TYPE_FLUX, clusterIds)
	if err != nil {
		impl.logger.Errorw("error in fetching helm app list from DB created using cd_pipelines", "clusters", clusterIds, "err", err)
		return
	}

	installedHelmApps, err := impl.installedAppRepository.GetAppAndEnvDetailsForDeploymentAppTypeInstalledApps(util.PIPELINE_DEPLOYMENT_TYPE_FLUX, clusterIds)
	if err != nil {
		impl.logger.Errorw("error in fetching helm app list from DB created from app store", "clusters", clusterIds, "err", err)
		return
	}

	cdPipelineMap := make(map[string]map[string]bool) // map of clusterId-namespace, deploymentAppName
	for _, p := range fluxCdPipelines {
		key := fmt.Sprintf("%v-%s", p.Environment.ClusterId, p.Environment.Namespace)
		if _, ok := cdPipelineMap[key]; !ok {
			cdPipelineMap[key] = make(map[string]bool)
		}
		cdPipelineMap[key][p.DeploymentAppName] = true
	}

	installedAppMap := make(map[string]map[string]bool)
	for _, i := range installedHelmApps {
		key := fmt.Sprintf("%v-%s", i.Environment.ClusterId, i.Environment.Namespace)
		if _, ok := installedAppMap[key]; !ok {
			installedAppMap[key] = make(map[string]bool)
		}
		deploymentAppName := fmt.Sprintf("%s-%s", i.App.AppName, i.Environment.Namespace)
		installedAppMap[key][deploymentAppName] = true
	}
	if !noStream {
		impl.pump.StartStreamWithTransformer(w, func() (proto.Message, error) {
			return appStream.Recv()
		}, err,
			func(message interface{}) interface{} {
				return impl.appListRespProtoTransformer(message.(*gRPC.FluxApplicationList), cdPipelineMap, installedAppMap)
			})
	} else {
		fluxApps := make([]bean.FluxApplication, 0)
		for {
			appDetail, err := appStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}
			if appDetail.Errored {
				appList := bean.FluxAppList{
					Errored:  &appDetail.Errored,
					ErrorMsg: &appDetail.ErrorMsg,
				}
				common.WriteJsonResp(w, nil, appList, http.StatusOK)
				return
			} else {
				for _, deployedApp := range appDetail.FluxApplication {
					key := fmt.Sprintf("%v-%s", deployedApp.EnvironmentDetail.ClusterId, deployedApp.EnvironmentDetail.Namespace)
					if _, ok := cdPipelineMap[key][deployedApp.Name]; ok {
						continue
					}
					if _, ok := installedAppMap[key][deployedApp.Name]; ok {
						continue
					}
					fluxApp := bean.FluxApplication{
						Name:                  deployedApp.Name,
						HealthStatus:          deployedApp.HealthStatus,
						SyncStatus:            deployedApp.SyncStatus,
						ClusterId:             int(deployedApp.EnvironmentDetail.ClusterId),
						ClusterName:           deployedApp.EnvironmentDetail.ClusterName,
						Namespace:             deployedApp.EnvironmentDetail.Namespace,
						FluxAppDeploymentType: deployedApp.FluxAppDeploymentType,
					}
					fluxApps = append(fluxApps, fluxApp)
				}
			}
		}
		clusterIdsInt32 := sliceUtil.NewSliceFromFuncExec(clusterIds, func(clusterId int) int32 {
			return int32(clusterId)
		})
		appList := bean.FluxAppList{
			ClusterId: &clusterIdsInt32,
			FluxApps:  &fluxApps,
		}
		common.WriteJsonResp(w, nil, appList, http.StatusOK)
	}
}
func (impl *FluxApplicationServiceImpl) GetFluxAppDetail(ctx context.Context, app *bean.FluxAppIdentifier) (*bean.FluxApplicationDetailDto, error) {
	config, err := impl.helmAppReadService.GetClusterConf(app.ClusterId)
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
		AppHealthStatus:      fluxDetailResponse.ApplicationStatus,
		LastObservedVersion:  fluxDetailResponse.GetLastObservedGeneration(),
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
			Token:                 clusterDetail.Config[commonBean.BearerToken],
			ClusterId:             int32(clusterDetail.Id),
			ClusterName:           clusterDetail.ClusterName,
			InsecureSkipTLSVerify: clusterDetail.InsecureSkipTLSVerify,
		}
		if clusterDetail.InsecureSkipTLSVerify == false {
			config.KeyData = clusterDetail.Config[commonBean.TlsKey]
			config.CertData = clusterDetail.Config[commonBean.CertData]
			config.CaData = clusterDetail.Config[commonBean.CertificateAuthorityData]
		}
		req.Clusters = append(req.Clusters, config)
	}
	applicationStream, err := impl.helmAppClient.ListFluxApplication(ctx, req)

	return applicationStream, err
}
func (impl *FluxApplicationServiceImpl) appListRespProtoTransformer(deployedApps *gRPC.FluxApplicationList, fluxCdPipelines map[string]map[string]bool, fluxInstalledApps map[string]map[string]bool) bean.FluxAppList {

	appList := bean.FluxAppList{ClusterId: &[]int32{deployedApps.ClusterId}}
	if deployedApps.Errored {
		appList.Errored = &deployedApps.Errored
		appList.ErrorMsg = &deployedApps.ErrorMsg
	} else {
		fluxApps := make([]bean.FluxApplication, 0, len(deployedApps.FluxApplication))
		for _, deployedApp := range deployedApps.FluxApplication {
			key := fmt.Sprintf("%v-%s", deployedApp.EnvironmentDetail.ClusterId, deployedApp.EnvironmentDetail.Namespace)
			if _, ok := fluxCdPipelines[key][deployedApp.Name]; ok {
				continue
			}
			if _, ok := fluxInstalledApps[key][deployedApp.Name]; ok {
				continue
			}
			fluxApp := bean.FluxApplication{
				Name:                  deployedApp.Name,
				HealthStatus:          deployedApp.HealthStatus,
				SyncStatus:            deployedApp.SyncStatus,
				ClusterId:             int(deployedApp.EnvironmentDetail.ClusterId),
				ClusterName:           deployedApp.EnvironmentDetail.ClusterName,
				Namespace:             deployedApp.EnvironmentDetail.Namespace,
				FluxAppDeploymentType: deployedApp.FluxAppDeploymentType,
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
