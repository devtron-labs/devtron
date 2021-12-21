package client

import (
	"github.com/devtron-labs/devtron/api/connector"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
	"net/http"
)

type HelmAppService interface {
	ListHelmApplications(clusterIds []int, w http.ResponseWriter)
}
type HelmAppServiceImpl struct {
	Logger         *zap.SugaredLogger
	clusterService cluster.ClusterService
	helmAppClient  HelmAppClient
	pump           connector.Pump
}

func NewHelmAppServiceImpl(Logger *zap.SugaredLogger,
	clusterService cluster.ClusterService,
	helmAppClient HelmAppClient,
	pump connector.Pump) *HelmAppServiceImpl {
	return &HelmAppServiceImpl{
		Logger:         Logger,
		clusterService: clusterService,
		helmAppClient:  helmAppClient,
		pump:           pump,
	}
}

func (impl *HelmAppServiceImpl) listApplications(clusterIds []int) (ApplicationService_ListApplicationsClient, error) {
	if len(clusterIds) == 0 {
		return nil, nil
	}
	clusters, err := impl.clusterService.FindByIds(clusterIds)
	if err != nil {
		impl.Logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
	req := &AppListRequest{}
	for _, clusterDetail := range clusters {
		config := &ClusterConfig{
			ApiServerUrl: clusterDetail.ServerUrl,
			Token:        clusterDetail.Config["bearer_token"],
			ClusterId:    int32(clusterDetail.Id),
			ClusterName:  clusterDetail.ClusterName,
		}
		req.Clusters = append(req.Clusters, config)
	}
	applicatonStream, err := impl.helmAppClient.ListApplication(req)
	if err != nil {
		return nil, err
	}

	return applicatonStream, err
}

func (impl *HelmAppServiceImpl) ListHelmApplications(clusterIds []int, w http.ResponseWriter) {
	appStream, err := impl.listApplications(clusterIds)
	if err != nil {
		impl.Logger.Errorw("error in fetching app list", "clusters", clusterIds, "err", err)
	}
	impl.pump.StartStreamWithTransformer(w, func() (proto.Message, error) {
		return appStream.Recv()
	}, err,
		func(message interface{}) interface{} {
			return impl.appListRespProtoTransformer(message.(*DeployedAppList))
		})
}

func (impl *HelmAppServiceImpl) appListRespProtoTransformer(deployedApps *DeployedAppList) openapi.AppList {
	applicationType := "HELM-APP"
	appList := openapi.AppList{ClusterIds: &[]int32{deployedApps.ClusterId}, ApplicationType: &applicationType}
	if deployedApps.Errored {
		appList.Errored = &deployedApps.Errored
		appList.ErrorMsg = &deployedApps.ErrorMsg
	} else {
		var HelmApps []openapi.HelmApp
		projectId := int32(0) //TODO pick from db
		for _, deployedapp := range deployedApps.DeployedAppDetail {
			lastDeployed := deployedapp.LastDeployed.AsTime()
			helmApp := openapi.HelmApp{
				AppName:        &deployedapp.AppName,
				AppId:          &deployedapp.AppId,
				ChartName:      &deployedapp.ChartName,
				ChartAvatar:    &deployedapp.ChartAvatar,
				LastDeployedAt: &lastDeployed,
				ProjectId:      &projectId,
				EnvironmentDetail: &openapi.AppEnvironmentDetail{
					Namespace:   &deployedapp.Environment.Namespace,
					ClusterName: &deployedapp.Environment.ClusterName,
					ClusterId:   &deployedapp.Environment.ClusterId,
				},
			}
			HelmApps = append(HelmApps, helmApp)
		}
		appList.HelmApps = &HelmApps

	}
	return appList
}
