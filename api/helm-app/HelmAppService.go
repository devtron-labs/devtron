package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/connector"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const DEFAULT_CLUSTER = "default_cluster"

type HelmAppService interface {
	ListHelmApplications(clusterIds []int, w http.ResponseWriter, token string, helmAuth func(token string, object string) bool)
	GetApplicationDetail(ctx context.Context, app *AppIdentifier) (*AppDetail, error)
	HibernateApplication(ctx context.Context, app *AppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
	UnHibernateApplication(ctx context.Context, app *AppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
	DecodeAppId(appId string) (*AppIdentifier, error)
	GetDeploymentHistory(ctx context.Context, app *AppIdentifier) (*HelmAppDeploymentHistory, error)
	GetValuesYaml(ctx context.Context, app *AppIdentifier) (*ReleaseInfo, error)
	GetDesiredManifest(ctx context.Context, app *AppIdentifier, resource *openapi.ResourceIdentifier) (*openapi.DesiredManifestResponse, error)
	DeleteApplication(ctx context.Context, app *AppIdentifier) (*openapi.UninstallReleaseResponse, error)
	UpdateApplication(ctx context.Context, app *AppIdentifier, request *openapi.UpdateReleaseRequest) (*openapi.UpdateReleaseResponse, error)
	GetDeploymentDetail(ctx context.Context, app *AppIdentifier, version int32) (*openapi.HelmAppDeploymentManifestDetail, error)
	InstallRelease(ctx context.Context, clusterId int, installReleaseRequest *InstallReleaseRequest) (*InstallReleaseResponse, error)
	UpdateApplicationWithChartInfo(ctx context.Context, clusterId int, updateReleaseRequest *InstallReleaseRequest) (*openapi.UpdateReleaseResponse, error)
	IsReleaseInstalled(ctx context.Context, app *AppIdentifier) (bool, error)
	RollbackRelease(ctx context.Context, app *AppIdentifier, version int32) (bool, error)
	GetClusterConf(clusterId int) (*ClusterConfig, error)
	GetDevtronHelmAppIdentifier() *AppIdentifier
	UpdateApplicationWithChartInfoWithExtraValues(ctx context.Context, appIdentifier *AppIdentifier, chartRepository *ChartRepository, extraValues map[string]interface{}, extraValuesYamlUrl string, useLatestChartVersion bool) (*openapi.UpdateReleaseResponse, error)
	TemplateChart(ctx context.Context, templateChartRequest *openapi2.TemplateChartRequest) (*openapi2.TemplateChartResponse, error)
}

type HelmAppServiceImpl struct {
	logger                               *zap.SugaredLogger
	clusterService                       cluster.ClusterService
	helmAppClient                        HelmAppClient
	pump                                 connector.Pump
	enforcerUtil                         rbac.EnforcerUtilHelm
	serverDataStore                      *serverDataStore.ServerDataStore
	serverEnvConfig                      *serverEnvConfig.ServerEnvConfig
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	environmentService                   cluster.EnvironmentService
	pipelineRepository                   pipelineConfig.PipelineRepository
}

func NewHelmAppServiceImpl(Logger *zap.SugaredLogger,
	clusterService cluster.ClusterService,
	helmAppClient HelmAppClient,
	pump connector.Pump, enforcerUtil rbac.EnforcerUtilHelm, serverDataStore *serverDataStore.ServerDataStore,
	serverEnvConfig *serverEnvConfig.ServerEnvConfig, appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	environmentService cluster.EnvironmentService, pipelineRepository pipelineConfig.PipelineRepository) *HelmAppServiceImpl {
	return &HelmAppServiceImpl{
		logger:                               Logger,
		clusterService:                       clusterService,
		helmAppClient:                        helmAppClient,
		pump:                                 pump,
		enforcerUtil:                         enforcerUtil,
		serverDataStore:                      serverDataStore,
		serverEnvConfig:                      serverEnvConfig,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		environmentService:                   environmentService,
		pipelineRepository:                   pipelineRepository,
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

func (impl *HelmAppServiceImpl) ListHelmApplications(clusterIds []int, w http.ResponseWriter, token string, helmAuth func(token string, object string) bool) {
	var helmCdPipelines []*pipelineConfig.Pipeline
	appStream, err := impl.listApplications(clusterIds)
	if err != nil {
		impl.logger.Errorw("error in fetching app list", "clusters", clusterIds, "err", err)
	}
	if err == nil && len(clusterIds) > 0 {
		// get helm apps which are created using cd_pipelines
		helmCdPipelines, err = impl.pipelineRepository.GetAppAndEnvDetailsForDeploymentAppTypePipeline(util.PIPELINE_DEPLOYMENT_TYPE_HELM, clusterIds)
		if err != nil {
			impl.logger.Errorw("error in fetching helm app list from DB created using cd_pipelines", "clusters", clusterIds, "err", err)
		}
	}
	impl.pump.StartStreamWithTransformer(w, func() (proto.Message, error) {
		return appStream.Recv()
	}, err,
		func(message interface{}) interface{} {
			return impl.appListRespProtoTransformer(message.(*DeployedAppList), token, helmAuth, helmCdPipelines)
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
	conf, err := impl.GetClusterConf(app.ClusterId)
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

	conf, err := impl.GetClusterConf(app.ClusterId)
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

func (impl *HelmAppServiceImpl) GetClusterConf(clusterId int) (*ClusterConfig, error) {
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
	config, err := impl.GetClusterConf(app.ClusterId)
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
	if err != nil {
		impl.logger.Errorw("error in fetching app detail", "err", err)
		return nil, err
	}

	// if application is devtron app helm release,
	// then for FULL (installer object exists), then status is combination of helm app status and installer object status -
	// if installer status is not applied then check for timeout and progressing
	devtronHelmAppIdentifier := impl.GetDevtronHelmAppIdentifier()
	if app.ClusterId == devtronHelmAppIdentifier.ClusterId && app.Namespace == devtronHelmAppIdentifier.Namespace && app.ReleaseName == devtronHelmAppIdentifier.ReleaseName &&
		impl.serverDataStore.InstallerCrdObjectExists {
		if impl.serverDataStore.InstallerCrdObjectStatus != serverBean.InstallerCrdObjectStatusApplied {
			// if timeout
			if time.Now().After(appdetail.GetLastDeployed().AsTime().Add(1 * time.Hour)) {
				appdetail.ApplicationStatus = serverBean.AppHealthStatusDegraded
			} else {
				appdetail.ApplicationStatus = serverBean.AppHealthStatusProgressing
			}
		}
	}
	return appdetail, err

}

func (impl *HelmAppServiceImpl) GetDeploymentHistory(ctx context.Context, app *AppIdentifier) (*HelmAppDeploymentHistory, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
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
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
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

func (impl *HelmAppServiceImpl) GetDesiredManifest(ctx context.Context, app *AppIdentifier, resource *openapi.ResourceIdentifier) (*openapi.DesiredManifestResponse, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", app.ClusterId, "err", err)
		return nil, err
	}

	req := &ObjectRequest{
		ClusterConfig:    config,
		ReleaseName:      app.ReleaseName,
		ReleaseNamespace: app.Namespace,
		ObjectIdentifier: &ObjectIdentifier{
			Group:     resource.GetGroup(),
			Kind:      resource.GetKind(),
			Version:   resource.GetVersion(),
			Name:      resource.GetName(),
			Namespace: resource.GetNamespace(),
		},
	}

	desiredManifestResponse, err := impl.helmAppClient.GetDesiredManifest(ctx, req)
	if err != nil {
		impl.logger.Errorw("error in fetching desired manifest", "err", err)
		return nil, err
	}

	response := &openapi.DesiredManifestResponse{
		Manifest: &desiredManifestResponse.Manifest,
	}
	return response, nil
}

func (impl *HelmAppServiceImpl) DeleteApplication(ctx context.Context, app *AppIdentifier) (*openapi.UninstallReleaseResponse, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", app.ClusterId, "err", err)
		return nil, err
	}

	req := &ReleaseIdentifier{
		ClusterConfig:    config,
		ReleaseName:      app.ReleaseName,
		ReleaseNamespace: app.Namespace,
	}

	deleteApplicationResponse, err := impl.helmAppClient.DeleteApplication(ctx, req)
	if err != nil {
		impl.logger.Errorw("error in deleting helm application", "err", err)
		return nil, err
	}

	response := &openapi.UninstallReleaseResponse{
		Success: &deleteApplicationResponse.Success,
	}
	return response, nil
}

func (impl *HelmAppServiceImpl) UpdateApplication(ctx context.Context, app *AppIdentifier, request *openapi.UpdateReleaseRequest) (*openapi.UpdateReleaseResponse, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", app.ClusterId, "err", err)
		return nil, err
	}

	req := &UpgradeReleaseRequest{
		ReleaseIdentifier: &ReleaseIdentifier{
			ClusterConfig:    config,
			ReleaseName:      app.ReleaseName,
			ReleaseNamespace: app.Namespace,
		},
		ValuesYaml: request.GetValuesYaml(),
	}

	updateApplicationResponse, err := impl.helmAppClient.UpdateApplication(ctx, req)
	if err != nil {
		impl.logger.Errorw("error in updating helm application", "err", err)
		return nil, err
	}

	response := &openapi.UpdateReleaseResponse{
		Success: &updateApplicationResponse.Success,
	}
	return response, nil
}

func (impl *HelmAppServiceImpl) GetDeploymentDetail(ctx context.Context, app *AppIdentifier, version int32) (*openapi.HelmAppDeploymentManifestDetail, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", app.ClusterId, "err", err)
		return nil, err
	}

	req := &DeploymentDetailRequest{
		ReleaseIdentifier: &ReleaseIdentifier{
			ClusterConfig:    config,
			ReleaseName:      app.ReleaseName,
			ReleaseNamespace: app.Namespace,
		},
		DeploymentVersion: version,
	}

	deploymentDetail, err := impl.helmAppClient.GetDeploymentDetail(ctx, req)
	if err != nil {
		impl.logger.Errorw("error in getting deployment detail", "err", err)
		return nil, err
	}

	response := &openapi.HelmAppDeploymentManifestDetail{
		Manifest:   &deploymentDetail.Manifest,
		ValuesYaml: &deploymentDetail.ValuesYaml,
	}

	return response, nil
}

func (impl *HelmAppServiceImpl) InstallRelease(ctx context.Context, clusterId int, installReleaseRequest *InstallReleaseRequest) (*InstallReleaseResponse, error) {
	config, err := impl.GetClusterConf(clusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", clusterId, "err", err)
		return nil, err
	}

	installReleaseRequest.ReleaseIdentifier.ClusterConfig = config

	installReleaseResponse, err := impl.helmAppClient.InstallRelease(ctx, installReleaseRequest)
	if err != nil {
		impl.logger.Errorw("error in installing release", "err", err)
		return nil, err
	}

	return installReleaseResponse, nil
}

func (impl *HelmAppServiceImpl) UpdateApplicationWithChartInfo(ctx context.Context, clusterId int, updateReleaseRequest *InstallReleaseRequest) (*openapi.UpdateReleaseResponse, error) {
	config, err := impl.GetClusterConf(clusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", clusterId, "err", err)
		return nil, err
	}

	updateReleaseRequest.ReleaseIdentifier.ClusterConfig = config

	updateReleaseResponse, err := impl.helmAppClient.UpdateApplicationWithChartInfo(ctx, updateReleaseRequest)
	if err != nil {
		impl.logger.Errorw("error in installing release", "err", err)
		return nil, err
	}

	response := &openapi.UpdateReleaseResponse{
		Success: &updateReleaseResponse.Success,
	}

	return response, nil
}

func (impl *HelmAppServiceImpl) IsReleaseInstalled(ctx context.Context, app *AppIdentifier) (bool, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", app.ClusterId, "err", err)
		return false, err
	}

	req := &ReleaseIdentifier{
		ClusterConfig:    config,
		ReleaseName:      app.ReleaseName,
		ReleaseNamespace: app.Namespace,
	}

	apiResponse, err := impl.helmAppClient.IsReleaseInstalled(ctx, req)
	if err != nil {
		impl.logger.Errorw("error in checking if helm release is installed", "err", err)
		return false, err
	}

	return apiResponse.Result, nil
}

func (impl *HelmAppServiceImpl) RollbackRelease(ctx context.Context, app *AppIdentifier, version int32) (bool, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", app.ClusterId, "err", err)
		return false, err
	}

	req := &RollbackReleaseRequest{
		ReleaseIdentifier: &ReleaseIdentifier{
			ClusterConfig:    config,
			ReleaseName:      app.ReleaseName,
			ReleaseNamespace: app.Namespace,
		},
		Version: version,
	}

	apiResponse, err := impl.helmAppClient.RollbackRelease(ctx, req)
	if err != nil {
		impl.logger.Errorw("error in rollback release", "err", err)
		return false, err
	}

	return apiResponse.Result, nil
}

func (impl *HelmAppServiceImpl) GetDevtronHelmAppIdentifier() *AppIdentifier {
	return &AppIdentifier{
		ClusterId:   1,
		Namespace:   impl.serverEnvConfig.DevtronHelmReleaseNamespace,
		ReleaseName: impl.serverEnvConfig.DevtronHelmReleaseName,
	}
}

func (impl *HelmAppServiceImpl) UpdateApplicationWithChartInfoWithExtraValues(ctx context.Context, appIdentifier *AppIdentifier,
	chartRepository *ChartRepository, extraValues map[string]interface{}, extraValuesYamlUrl string, useLatestChartVersion bool) (*openapi.UpdateReleaseResponse, error) {

	// get release info
	releaseInfo, err := impl.GetValuesYaml(context.Background(), appIdentifier)
	if err != nil {
		impl.logger.Errorw("error in fetching helm release info", "err", err)
		return nil, err
	}

	// initialise object with original values
	jsonString := releaseInfo.MergedValues

	// handle extra values
	// special handling for array
	if len(extraValues) > 0 {
		for k, v := range extraValues {
			var valueI interface{}
			if reflect.TypeOf(v).Kind() == reflect.Slice {
				currentValue := gjson.Get(jsonString, k).Value()
				value := make([]interface{}, 0)
				if currentValue != nil {
					value = currentValue.([]interface{})
				}
				for _, singleNewVal := range v.([]interface{}) {
					value = append(value, singleNewVal)
				}
				valueI = value
			} else {
				valueI = v
			}
			jsonString, err = sjson.Set(jsonString, k, valueI)
			if err != nil {
				impl.logger.Errorw("error in handing extra values", "err", err)
				return nil, err
			}
		}
	}

	// convert to byte array
	mergedValuesJsonByteArr := []byte(jsonString)

	// handle extra values from url
	if len(extraValuesYamlUrl) > 0 {
		extraValuesUrlYamlByteArr, err := util2.ReadFromUrlWithRetry(extraValuesYamlUrl)
		if err != nil {
			impl.logger.Errorw("error in reading content", "extraValuesYamlUrl", extraValuesYamlUrl, "err", err)
			return nil, err
		} else if extraValuesUrlYamlByteArr == nil {
			impl.logger.Errorw("response is empty from url", "extraValuesYamlUrl", extraValuesYamlUrl)
			return nil, errors.New("response is empty from values url")
		}

		extraValuesUrlJsonByteArr, err := yaml.YAMLToJSON(extraValuesUrlYamlByteArr)
		if err != nil {
			impl.logger.Errorw("error in converting json to yaml", "err", err)
			return nil, err
		}

		mergedValuesJsonByteArr, err = jsonpatch.MergePatch(mergedValuesJsonByteArr, extraValuesUrlJsonByteArr)
		if err != nil {
			impl.logger.Errorw("error in json patch of extra values from url", "err", err)
			return nil, err
		}
	}

	// convert JSON to yaml byte array
	mergedValuesYamlByteArr, err := yaml.JSONToYAML(mergedValuesJsonByteArr)
	if err != nil {
		impl.logger.Errorw("error in converting json to yaml", "err", err)
		return nil, err
	}

	// update in helm
	updateReleaseRequest := &InstallReleaseRequest{
		ReleaseIdentifier: &ReleaseIdentifier{
			ReleaseName:      appIdentifier.ReleaseName,
			ReleaseNamespace: appIdentifier.Namespace,
		},
		ChartName:       releaseInfo.DeployedAppDetail.ChartName,
		ValuesYaml:      string(mergedValuesYamlByteArr),
		ChartRepository: chartRepository,
	}
	if !useLatestChartVersion {
		updateReleaseRequest.ChartVersion = releaseInfo.DeployedAppDetail.ChartVersion
	}

	updateResponse, err := impl.UpdateApplicationWithChartInfo(ctx, appIdentifier.ClusterId, updateReleaseRequest)
	if err != nil {
		impl.logger.Errorw("error in upgrading release", "err", err)
		return nil, err
	}
	// update in helm ends

	response := &openapi.UpdateReleaseResponse{
		Success: updateResponse.Success,
	}

	return response, nil
}

func (impl *HelmAppServiceImpl) TemplateChart(ctx context.Context, templateChartRequest *openapi2.TemplateChartRequest) (*openapi2.TemplateChartResponse, error) {
	appStoreApplicationVersionId := int(*templateChartRequest.AppStoreApplicationVersionId)
	environmentId := int(*templateChartRequest.EnvironmentId)
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(appStoreApplicationVersionId)
	if err != nil {
		impl.logger.Errorw("Error in fetching app-store application version", "appStoreApplicationVersionId", appStoreApplicationVersionId, "err", err)
		return nil, err
	}

	if environmentId > 0 {
		environment, err := impl.environmentService.FindById(environmentId)
		if err != nil {
			impl.logger.Errorw("Error in fetching environment", "environmentId", environmentId, "err", err)
			return nil, err
		}
		templateChartRequest.Namespace = &environment.Namespace
		clusterIdI32 := int32(environment.ClusterId)
		templateChartRequest.ClusterId = &clusterIdI32
	}

	clusterId := int(*templateChartRequest.ClusterId)

	installReleaseRequest := &InstallReleaseRequest{
		ChartName:    appStoreAppVersion.Name,
		ChartVersion: appStoreAppVersion.Version,
		ValuesYaml:   *templateChartRequest.ValuesYaml,
		ChartRepository: &ChartRepository{
			Name:     appStoreAppVersion.AppStore.ChartRepo.Name,
			Url:      appStoreAppVersion.AppStore.ChartRepo.Url,
			Username: appStoreAppVersion.AppStore.ChartRepo.UserName,
			Password: appStoreAppVersion.AppStore.ChartRepo.Password,
		},
		ReleaseIdentifier: &ReleaseIdentifier{
			ReleaseNamespace: *templateChartRequest.Namespace,
			ReleaseName:      *templateChartRequest.ReleaseName,
		},
	}

	config, err := impl.GetClusterConf(clusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", clusterId, "err", err)
		return nil, err
	}

	installReleaseRequest.ReleaseIdentifier.ClusterConfig = config

	templateChartResponse, err := impl.helmAppClient.TemplateChart(ctx, installReleaseRequest)
	if err != nil {
		impl.logger.Errorw("error in templating chart", "err", err)
		return nil, err
	}

	response := &openapi2.TemplateChartResponse{
		Manifest: &templateChartResponse.GeneratedManifest,
	}

	return response, nil
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

func (impl *HelmAppServiceImpl) appListRespProtoTransformer(deployedApps *DeployedAppList, token string, helmAuth func(token string, object string) bool, helmCdPipelines []*pipelineConfig.Pipeline) openapi.AppList {
	applicationType := "HELM-APP"
	appList := openapi.AppList{ClusterIds: &[]int32{deployedApps.ClusterId}, ApplicationType: &applicationType}
	if deployedApps.Errored {
		appList.Errored = &deployedApps.Errored
		appList.ErrorMsg = &deployedApps.ErrorMsg
	} else {
		var HelmApps []openapi.HelmApp
		projectId := int32(0) //TODO pick from db
		for _, deployedapp := range deployedApps.DeployedAppDetail {

			// do not add app in the list which are created using cd_pipelines (check combination of clusterId, namespace, releaseName)
			var toExcludeFromList bool
			for _, helmCdPipeline := range helmCdPipelines {
				helmAppReleaseName := util2.BuildDeployedAppName(helmCdPipeline.App.AppName, helmCdPipeline.Environment.Name)
				if deployedapp.AppName == helmAppReleaseName && int(deployedapp.EnvironmentDetail.ClusterId) == helmCdPipeline.Environment.ClusterId && deployedapp.EnvironmentDetail.Namespace == helmCdPipeline.Environment.Namespace {
					toExcludeFromList = true
					break
				}
			}
			if toExcludeFromList {
				continue
			}

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
			rbacObject := impl.enforcerUtil.GetHelmObjectByClusterId(int(deployedapp.EnvironmentDetail.ClusterId), deployedapp.EnvironmentDetail.Namespace, deployedapp.AppName)
			isValidAuth := helmAuth(token, rbacObject)
			if isValidAuth {
				HelmApps = append(HelmApps, helmApp)
			}
		}
		appList.HelmApps = &HelmApps

	}
	return appList
}
