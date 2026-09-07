package fluxApplication

import (
	"context"
	"fmt"
	"io"
	"net/http"

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
)

type FluxApplicationService interface {
	ListFluxApplications(ctx context.Context, clusterIds []int, noStream bool, w http.ResponseWriter,
		token string, fluxAuth func(token string, clusterName string, namespace string, appName string) bool)
	GetFluxAppDetail(ctx context.Context, app *bean.FluxAppIdentifier) (*bean.FluxApplicationDetailDto, error)
	HibernateFluxApplication(ctx context.Context, app *bean.FluxAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
	UnHibernateFluxApplication(ctx context.Context, app *bean.FluxAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
	GetFluxApplicationList(ctx context.Context, clusterIds []int) ([]bean.FluxApplication, error)
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

func (impl *FluxApplicationServiceImpl) GetFluxApplicationList(ctx context.Context, clusterIds []int) ([]bean.FluxApplication, error) {
	appStream, err := impl.listApplications(ctx, clusterIds)
	if err != nil {
		return nil, err
	}
	cdPipelineMap, installedAppMap, err := impl.getDevtronManagedMaps(clusterIds)
	if err != nil {
		return nil, err
	}

	apps := make([]bean.FluxApplication, 0)
	for {
		appDetail, err := appStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if appDetail.Errored {
			impl.logger.Errorw("error in listing flux applications for cluster, skipping it",
				"clusterId", appDetail.ClusterId, "errorMsg", appDetail.ErrorMsg)
			continue
		}
		for _, d := range appDetail.FluxApplication {
			key := fmt.Sprintf("%v-%s", d.EnvironmentDetail.ClusterId, d.EnvironmentDetail.Namespace)
			if _, ok := cdPipelineMap[key][d.Name]; ok {
				continue
			}
			if _, ok := installedAppMap[key][d.Name]; ok {
				continue
			}
			apps = append(apps, toFluxApplication(d))
		}
	}
	return apps, nil
}

func (impl *FluxApplicationServiceImpl) getDevtronManagedMaps(clusterIds []int) (map[string]map[string]bool, map[string]map[string]bool, error) {
	fluxCdPipelines, err := impl.pipelineRepository.GetAppAndEnvDetailsForDeploymentAppTypePipeline(util.PIPELINE_DEPLOYMENT_TYPE_FLUX, clusterIds)
	if err != nil {
		impl.logger.Errorw("error in fetching helm app list from DB created using cd_pipelines", "clusters", clusterIds, "err", err)
		return nil, nil, err
	}

	installedHelmApps, err := impl.installedAppRepository.GetAppAndEnvDetailsForDeploymentAppTypeInstalledApps(util.PIPELINE_DEPLOYMENT_TYPE_FLUX, clusterIds)
	if err != nil {
		impl.logger.Errorw("error in fetching helm app list from DB created from app store", "clusters", clusterIds, "err", err)
		return nil, nil, err
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

	return cdPipelineMap, installedAppMap, nil
}

func toFluxApplication(app *gRPC.FluxApplication) bean.FluxApplication {
	fluxApp := bean.FluxApplication{
		Name:                  app.Name,
		HealthStatus:          app.HealthStatus,
		SyncStatus:            app.SyncStatus,
		ClusterId:             int(app.EnvironmentDetail.ClusterId),
		ClusterName:           app.EnvironmentDetail.ClusterName,
		Namespace:             app.EnvironmentDetail.Namespace,
		FluxAppDeploymentType: app.FluxAppDeploymentType,
	}

	return fluxApp
}

func (impl *FluxApplicationServiceImpl) ListFluxApplications(ctx context.Context, clusterIds []int, noStream bool, w http.ResponseWriter,
	token string, fluxAuth func(token string, clusterName string, namespace string, appName string) bool) {

	if !noStream {
		appStream, err := impl.listApplications(ctx, clusterIds)
		if err != nil {
			impl.logger.Errorw("error in listing flux applications", "clusterIds", clusterIds, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		cdPipelineMap, installedAppMap, err := impl.getDevtronManagedMaps(clusterIds)
		if err != nil {
			impl.logger.Errorw("error in getting devtron managed flux apps", "clusterIds", clusterIds, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		impl.pump.StartStreamWithTransformer(w, func() (proto.Message, error) {
			return appStream.Recv()
		}, err,
			func(message interface{}) interface{} {
				return impl.appListRespProtoTransformer(message.(*gRPC.FluxApplicationList), cdPipelineMap, installedAppMap, token, fluxAuth)
			})
	} else {
		fluxApps, err := impl.GetFluxApplicationList(ctx, clusterIds)
		if err != nil {
			impl.logger.Errorw("error in getting flux application list", "clusterIds", clusterIds, "err", err)
			errored := true
			errMsg := err.Error()
			appList := bean.FluxAppList{
				Errored:  &errored,
				ErrorMsg: &errMsg,
			}
			common.WriteJsonResp(w, nil, appList, http.StatusOK)
			return
		}

		if fluxAuth != nil {
			authorised := make([]bean.FluxApplication, 0, len(fluxApps))
			for _, app := range fluxApps {
				if fluxAuth(token, app.ClusterName, app.Namespace, app.Name) {
					authorised = append(authorised, app)
				}
			}
			fluxApps = authorised
		}
		//RBAC enforcer Ends
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
		if clusterDetail.IsVirtualCluster || len(clusterDetail.ErrorInConnecting) != 0 {
			impl.logger.Debugw("skipping cluster for flux app listing", "clusterId", clusterDetail.Id,
				"isVirtualCluster", clusterDetail.IsVirtualCluster, "errorInConnecting", clusterDetail.ErrorInConnecting)
			continue
		}
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
func (impl *FluxApplicationServiceImpl) appListRespProtoTransformer(deployedApps *gRPC.FluxApplicationList, fluxCdPipelines map[string]map[string]bool, fluxInstalledApps map[string]map[string]bool,
	token string, fluxAuth func(token string, clusterName string, namespace string, appName string) bool) bean.FluxAppList {

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
			if fluxAuth != nil && !fluxAuth(token, deployedApp.EnvironmentDetail.ClusterName,
				deployedApp.EnvironmentDetail.Namespace, deployedApp.Name) {
				continue
			}
			//RBAC enforcer Ends
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
