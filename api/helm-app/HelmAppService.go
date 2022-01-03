package client

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/api/connector"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

const DEFAULT_CLUSTER = "default_cluster"

type HelmAppService interface {
	ListHelmApplications(clusterIds []int, w http.ResponseWriter)
	GetApplicationDetail(ctx context.Context, app *AppIdentifier) (*AppDetail, error)
	HibernateApplication(ctx context.Context, app *AppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
	UnHibernateApplication(ctx context.Context, app *AppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
	DecodeAppId(appId string) (*AppIdentifier, error)
	GetDeploymentHistory(ctx context.Context, app *AppIdentifier) (*HelmAppDeploymentHistory, error)
	GetValuesYaml(ctx context.Context, app *AppIdentifier) (*ReleaseInfo, error)
}
type HelmAppServiceImpl struct {
	logger         *zap.SugaredLogger
	clusterService cluster.ClusterService
	helmAppClient  HelmAppClient
	pump           connector.Pump
}

func NewHelmAppServiceImpl(Logger *zap.SugaredLogger,
	clusterService cluster.ClusterService,
	helmAppClient HelmAppClient,
	pump connector.Pump) *HelmAppServiceImpl {
	return &HelmAppServiceImpl{
		logger:         Logger,
		clusterService: clusterService,
		helmAppClient:  helmAppClient,
		pump:           pump,
	}
}

type ResourceRequestBean struct {
	AppId      string                     `json:"appId"`
	K8sRequest application.K8sRequestBean `json:"k8sRequest"`
}

func (impl *HelmAppServiceImpl) listApplications(clusterIds []int) (ApplicationService_ListApplicationsClient, error) {
	if len(clusterIds) == 0 {
		return nil, nil
	}
	clusters, err := impl.clusterService.FindByIds(clusterIds)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
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
		impl.logger.Errorw("error in fetching app list", "clusters", clusterIds, "err", err)
	}
	impl.pump.StartStreamWithTransformer(w, func() (proto.Message, error) {
		return appStream.Recv()
	}, err,
		func(message interface{}) interface{} {
			return impl.appListRespProtoTransformer(message.(*DeployedAppList))
		})
}

func (impl *HelmAppServiceImpl) hibernateReqAdaptor(hibernateRequest *openapi.HibernateRequest) *HibernateRequest {
	req := &HibernateRequest{}
	for _, reqObject := range hibernateRequest.GetResources() {
		obj := &ObjectIdentifier{
			Group:     *reqObject.Group,
			Kind:      *reqObject.Kind,
			Version:   *reqObject.Version,
			Name:      *reqObject.Name,
			Namespace: *reqObject.Namespace,
		}
		req.ObjectIdentifier = append(req.ObjectIdentifier, obj)
	}
	return req
}
func (impl *HelmAppServiceImpl) hibernateResponseAdaptor(in []*HibernateStatus) []*openapi.HibernateStatus {
	var resStatus []*openapi.HibernateStatus
	for _, status := range in {
		resObj := &openapi.HibernateStatus{
			Success:      &status.Success,
			ErrorMessage: &status.ErrorMsg,
			TargetObject: &openapi.HibernateTargetObject{
				Group:     &status.TargetObject.Group,
				Kind:      &status.TargetObject.Kind,
				Version:   &status.TargetObject.Version,
				Name:      &status.TargetObject.Name,
				Namespace: &status.TargetObject.Namespace,
			},
		}
		resStatus = append(resStatus, resObj)
	}
	return resStatus
}
func (impl *HelmAppServiceImpl) HibernateApplication(ctx context.Context, app *AppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	conf, err := impl.getClusterConf(app.ClusterId)
	if err != nil {
		return nil, err
	}
	req := impl.hibernateReqAdaptor(hibernateRequest)
	req.ClusterConfig = conf
	res, err := impl.helmAppClient.Hibernate(ctx, req)
	if err != nil {
		return nil, err
	}
	response := impl.hibernateResponseAdaptor(res.Status)
	return response, nil
}

func (impl *HelmAppServiceImpl) UnHibernateApplication(ctx context.Context, app *AppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {

	conf, err := impl.getClusterConf(app.ClusterId)
	if err != nil {
		return nil, err
	}
	req := impl.hibernateReqAdaptor(hibernateRequest)
	req.ClusterConfig = conf
	res, err := impl.helmAppClient.UnHibernate(ctx, req)
	if err != nil {
		return nil, err
	}
	response := impl.hibernateResponseAdaptor(res.Status)
	return response, nil
}

func (impl *HelmAppServiceImpl) getClusterConf(clusterId int) (*ClusterConfig, error) {
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
	config := &ClusterConfig{
		ApiServerUrl: cluster.ServerUrl,
		Token:        cluster.Config["bearer_token"],
		ClusterId:    int32(cluster.Id),
		ClusterName:  cluster.ClusterName,
	}
	return config, nil
}
func (impl *HelmAppServiceImpl) GetApplicationDetail(ctx context.Context, app *AppIdentifier) (*AppDetail, error) {
	config, err := impl.getClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
	req := &AppDetailRequest{
		ClusterConfig: config,
		Namespace:     app.Namespace,
		ReleaseName:   app.ReleaseName,
	}
	appdetail, err := impl.helmAppClient.GetAppDetail(ctx, req)
	return appdetail, err

}

func (impl *HelmAppServiceImpl) GetDeploymentHistory(ctx context.Context, app *AppIdentifier) (*HelmAppDeploymentHistory, error) {
	config, err := impl.getClusterConf(app.ClusterId)
	if err != nil {
		impl.Logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
	req := &AppDetailRequest{
		ClusterConfig: config,
		Namespace:     app.Namespace,
		ReleaseName:   app.ReleaseName,
	}
	history, err := impl.helmAppClient.GetDeploymentHistory(ctx, req)
	return history, err
}

func (impl *HelmAppServiceImpl) GetValuesYaml(ctx context.Context, app *AppIdentifier) (*ReleaseInfo, error) {
	config, err := impl.getClusterConf(app.ClusterId)
	if err != nil {
		impl.Logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
	req := &AppDetailRequest{
		ClusterConfig: config,
		Namespace:     app.Namespace,
		ReleaseName:   app.ReleaseName,
	}
	history, err := impl.helmAppClient.GetValuesYaml(ctx, req)
	return history, err
}

type AppIdentifier struct {
	ClusterId   int    `json:"clusterId"`
	Namespace   string `json:"namespace"`
	ReleaseName string `json:"releaseName"`
}

func (impl *HelmAppServiceImpl) DecodeAppId(appId string) (*AppIdentifier, error) {
	component := strings.Split(appId, "|")
	if len(component) != 3 {
		return nil, fmt.Errorf("malformed app id %s", appId)
	}
	clustewrId, err := strconv.Atoi(component[0])
	if err != nil {
		return nil, err
	}
	return &AppIdentifier{
		ClusterId:   clustewrId,
		Namespace:   component[1],
		ReleaseName: component[2],
	}, nil
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
					Namespace:   &deployedapp.EnvironmentDetail.Namespace,
					ClusterName: &deployedapp.EnvironmentDetail.ClusterName,
					ClusterId:   &deployedapp.EnvironmentDetail.ClusterId,
				},
			}
			HelmApps = append(HelmApps, helmApp)
		}
		appList.HelmApps = &HelmApps

	}
	return appList
}
