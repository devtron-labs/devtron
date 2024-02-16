package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/helm-app/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/api/helm-app/models"
	"github.com/devtron-labs/devtron/internal/constants"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/go-pg/pg"
	"google.golang.org/grpc/codes"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/api/connector"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gogo/protobuf/proto"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

type HelmAppService interface {
	ListHelmApplications(ctx context.Context, clusterIds []int, w http.ResponseWriter, token string, helmAuth func(token string, object string) bool)
	GetApplicationDetail(ctx context.Context, app *AppIdentifier) (*gRPC.AppDetail, error)
	GetApplicationAndReleaseStatus(ctx context.Context, app *AppIdentifier) (*gRPC.AppStatus, error)
	GetApplicationDetailWithFilter(ctx context.Context, app *AppIdentifier, resourceTreeFilter *gRPC.ResourceTreeFilter) (*gRPC.AppDetail, error)
	HibernateApplication(ctx context.Context, app *AppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
	UnHibernateApplication(ctx context.Context, app *AppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error)
	DecodeAppId(appId string) (*AppIdentifier, error)
	EncodeAppId(appIdentifier *AppIdentifier) string
	GetDeploymentHistory(ctx context.Context, app *AppIdentifier) (*gRPC.HelmAppDeploymentHistory, error)
	GetValuesYaml(ctx context.Context, app *AppIdentifier) (*gRPC.ReleaseInfo, error)
	GetDesiredManifest(ctx context.Context, app *AppIdentifier, resource *openapi.ResourceIdentifier) (*openapi.DesiredManifestResponse, error)
	DeleteApplication(ctx context.Context, app *AppIdentifier) (*openapi.UninstallReleaseResponse, error)
	DeleteDBLinkedHelmApplication(ctx context.Context, app *AppIdentifier, useId int32) (*openapi.UninstallReleaseResponse, error)
	UpdateApplication(ctx context.Context, app *AppIdentifier, request *bean.UpdateApplicationRequestDto) (*openapi.UpdateReleaseResponse, error)
	GetDeploymentDetail(ctx context.Context, app *AppIdentifier, version int32) (*openapi.HelmAppDeploymentManifestDetail, error)
	InstallRelease(ctx context.Context, clusterId int, installReleaseRequest *gRPC.InstallReleaseRequest) (*gRPC.InstallReleaseResponse, error)
	UpdateApplicationWithChartInfo(ctx context.Context, clusterId int, request *bean.UpdateApplicationWithChartInfoRequestDto) (*openapi.UpdateReleaseResponse, error)
	IsReleaseInstalled(ctx context.Context, app *AppIdentifier) (bool, error)
	RollbackRelease(ctx context.Context, app *AppIdentifier, version int32) (bool, error)
	GetClusterConf(clusterId int) (*gRPC.ClusterConfig, error)
	GetDevtronHelmAppIdentifier() *AppIdentifier
	UpdateApplicationWithChartInfoWithExtraValues(ctx context.Context, appIdentifier *AppIdentifier, chartRepository *gRPC.ChartRepository, extraValues map[string]interface{}, extraValuesYamlUrl string, useLatestChartVersion bool) (*openapi.UpdateReleaseResponse, error)
	TemplateChart(ctx context.Context, templateChartRequest *openapi2.TemplateChartRequest) (*openapi2.TemplateChartResponse, error)
	GetNotes(ctx context.Context, request *gRPC.InstallReleaseRequest) (string, error)
	ValidateOCIRegistry(ctx context.Context, OCIRegistryRequest *gRPC.RegistryCredential) bool
	GetRevisionHistoryMaxValue(appType bean.SourceAppType) int32
	GetResourceTreeForExternalResources(ctx context.Context, clusterId int, clusterConfig *gRPC.ClusterConfig, resources []*gRPC.ExternalResourceDetail) (*gRPC.ResourceTreeResponse, error)
}

type HelmAppServiceImpl struct {
	logger                               *zap.SugaredLogger
	clusterService                       cluster.ClusterService
	helmAppClient                        gRPC.HelmAppClient
	pump                                 connector.Pump
	enforcerUtil                         rbac.EnforcerUtilHelm
	serverDataStore                      *serverDataStore.ServerDataStore
	serverEnvConfig                      *serverEnvConfig.ServerEnvConfig
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	environmentService                   cluster.EnvironmentService
	pipelineRepository                   pipelineConfig.PipelineRepository
	installedAppRepository               repository.InstalledAppRepository
	appRepository                        app.AppRepository
	clusterRepository                    clusterRepository.ClusterRepository
	K8sUtil                              *k8s.K8sServiceImpl
	helmReleaseConfig                    *HelmReleaseConfig
}

func NewHelmAppServiceImpl(Logger *zap.SugaredLogger, clusterService cluster.ClusterService,
	helmAppClient gRPC.HelmAppClient, pump connector.Pump, enforcerUtil rbac.EnforcerUtilHelm,
	serverDataStore *serverDataStore.ServerDataStore, serverEnvConfig *serverEnvConfig.ServerEnvConfig,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	environmentService cluster.EnvironmentService, pipelineRepository pipelineConfig.PipelineRepository,
	installedAppRepository repository.InstalledAppRepository, appRepository app.AppRepository,
	clusterRepository clusterRepository.ClusterRepository, K8sUtil *k8s.K8sServiceImpl,
	helmReleaseConfig *HelmReleaseConfig) *HelmAppServiceImpl {
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
		installedAppRepository:               installedAppRepository,
		appRepository:                        appRepository,
		clusterRepository:                    clusterRepository,
		K8sUtil:                              K8sUtil,
		helmReleaseConfig:                    helmReleaseConfig,
	}
}

type HelmReleaseConfig struct {
	RevisionHistoryLimitDevtronApp      int `env:"REVISION_HISTORY_LIMIT_DEVTRON_APP" envDefault:"0"`
	RevisionHistoryLimitHelmApp         int `env:"REVISION_HISTORY_LIMIT_HELM_APP" envDefault:"0"`
	RevisionHistoryLimitExternalHelmApp int `env:"REVISION_HISTORY_LIMIT_EXTERNAL_HELM_APP" envDefault:"0"`
}

func GetHelmReleaseConfig() (*HelmReleaseConfig, error) {
	cfg := &HelmReleaseConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func (impl *HelmAppServiceImpl) listApplications(ctx context.Context, clusterIds []int) (gRPC.ApplicationService_ListApplicationsClient, error) {
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
	req := &gRPC.AppListRequest{}
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
	applicatonStream, err := impl.helmAppClient.ListApplication(ctx, req)
	if err != nil {
		return nil, err
	}

	return applicatonStream, err
}

func (impl *HelmAppServiceImpl) ListHelmApplications(ctx context.Context, clusterIds []int, w http.ResponseWriter, token string, helmAuth func(token string, object string) bool) {
	var helmCdPipelines []*pipelineConfig.Pipeline
	var installedHelmApps []*repository.InstalledApps
	start := time.Now()
	appStream, err := impl.listApplications(ctx, clusterIds)
	middleware.AppListingDuration.WithLabelValues("listApplications", "helm").Observe(time.Since(start).Seconds())
	if err != nil {
		impl.logger.Errorw("error in fetching app list", "clusters", clusterIds, "err", err)
	}
	if err == nil && len(clusterIds) > 0 {
		// get helm apps which are created using cd_pipelines
		newCtx, span := otel.Tracer("pipelineRepository").Start(ctx, "GetAppAndEnvDetailsForDeploymentAppTypePipeline")
		start = time.Now()
		helmCdPipelines, err = impl.pipelineRepository.GetAppAndEnvDetailsForDeploymentAppTypePipeline(util.PIPELINE_DEPLOYMENT_TYPE_HELM, clusterIds)
		middleware.AppListingDuration.WithLabelValues("getAppAndEnvDetailsForDeploymentAppTypePipeline", "helm").Observe(time.Since(start).Seconds())
		span.End()
		if err != nil {
			impl.logger.Errorw("error in fetching helm app list from DB created using cd_pipelines", "clusters", clusterIds, "err", err)
		}

		// if not hyperion mode, then fetch from installed_apps whose deployment_app_type is helm (as in hyperion mode, these apps should be treated as external-apps)
		if !util2.IsBaseStack() {
			newCtx, span = otel.Tracer("pipelineRepository").Start(newCtx, "GetAppAndEnvDetailsForDeploymentAppTypePipeline")
			start = time.Now()
			installedHelmApps, err = impl.installedAppRepository.GetAppAndEnvDetailsForDeploymentAppTypeInstalledApps(util.PIPELINE_DEPLOYMENT_TYPE_HELM, clusterIds)
			middleware.AppListingDuration.WithLabelValues("getAppAndEnvDetailsForDeploymentAppTypeInstalledApps", "helm").Observe(time.Since(start).Seconds())
			span.End()
			if err != nil {
				impl.logger.Errorw("error in fetching helm app list from DB created from app store", "clusters", clusterIds, "err", err)
			}
		}
	}
	impl.pump.StartStreamWithTransformer(w, func() (proto.Message, error) {
		return appStream.Recv()
	}, err,
		func(message interface{}) interface{} {
			return impl.appListRespProtoTransformer(message.(*gRPC.DeployedAppList), token, helmAuth, helmCdPipelines, installedHelmApps)
		})
}

func (impl *HelmAppServiceImpl) hibernateReqAdaptor(hibernateRequest *openapi.HibernateRequest) *gRPC.HibernateRequest {
	req := &gRPC.HibernateRequest{}
	for _, reqObject := range hibernateRequest.GetResources() {
		obj := &gRPC.ObjectIdentifier{
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
func (impl *HelmAppServiceImpl) hibernateResponseAdaptor(in []*gRPC.HibernateStatus) []*openapi.HibernateStatus {
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

func (impl *HelmAppServiceImpl) GetClusterConf(clusterId int) (*gRPC.ClusterConfig, error) {
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
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
	return config, nil
}

func (impl *HelmAppServiceImpl) GetApplicationDetail(ctx context.Context, app *AppIdentifier) (*gRPC.AppDetail, error) {
	return impl.getApplicationDetail(ctx, app, nil)
}

func (impl *HelmAppServiceImpl) GetApplicationAndReleaseStatus(ctx context.Context, app *AppIdentifier) (*gRPC.AppStatus, error) {
	return impl.getApplicationAndReleaseStatus(ctx, app)
}

func (impl *HelmAppServiceImpl) GetApplicationDetailWithFilter(ctx context.Context, app *AppIdentifier, resourceTreeFilter *gRPC.ResourceTreeFilter) (*gRPC.AppDetail, error) {
	return impl.getApplicationDetail(ctx, app, resourceTreeFilter)
}

func (impl *HelmAppServiceImpl) getApplicationDetail(ctx context.Context, app *AppIdentifier, resourceTreeFilter *gRPC.ResourceTreeFilter) (*gRPC.AppDetail, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
	req := &gRPC.AppDetailRequest{
		ClusterConfig:      config,
		Namespace:          app.Namespace,
		ReleaseName:        app.ReleaseName,
		ResourceTreeFilter: resourceTreeFilter,
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

func (impl *HelmAppServiceImpl) GetResourceTreeForExternalResources(ctx context.Context, clusterId int,
	clusterConfig *gRPC.ClusterConfig, resources []*gRPC.ExternalResourceDetail) (*gRPC.ResourceTreeResponse, error) {
	var config *gRPC.ClusterConfig
	var err error
	if clusterId > 0 {
		config, err = impl.GetClusterConf(clusterId)
		if err != nil {
			impl.logger.Errorw("error in fetching cluster detail", "err", err)
			return nil, err
		}
	} else {
		config = clusterConfig
	}
	req := &gRPC.ExternalResourceTreeRequest{
		ClusterConfig:          config,
		ExternalResourceDetail: resources,
	}
	return impl.helmAppClient.GetResourceTreeForExternalResources(ctx, req)
}

func (impl *HelmAppServiceImpl) getApplicationAndReleaseStatus(ctx context.Context, app *AppIdentifier) (*gRPC.AppStatus, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
	req := &gRPC.AppDetailRequest{
		ClusterConfig: config,
		Namespace:     app.Namespace,
		ReleaseName:   app.ReleaseName,
	}
	appStatus, err := impl.helmAppClient.GetAppStatus(ctx, req)
	if err != nil {
		impl.logger.Errorw("error in fetching app status", "err", err)
		return nil, err
	}
	return appStatus, err
}

func (impl *HelmAppServiceImpl) GetDeploymentHistory(ctx context.Context, app *AppIdentifier) (*gRPC.HelmAppDeploymentHistory, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
	req := &gRPC.AppDetailRequest{
		ClusterConfig: config,
		Namespace:     app.Namespace,
		ReleaseName:   app.ReleaseName,
	}
	history, err := impl.helmAppClient.GetDeploymentHistory(ctx, req)
	if util.GetGRPCErrorDetailedMessage(err) == bean.ErrReleaseNotFound {
		err = &util.ApiError{
			Code:            constants.HelmReleaseNotFound,
			InternalMessage: bean.ErrReleaseNotFound,
			UserMessage:     fmt.Sprintf("no release found with release name '%s'", req.ReleaseName),
		}
	}
	return history, err
}

func (impl *HelmAppServiceImpl) GetValuesYaml(ctx context.Context, app *AppIdentifier) (*gRPC.ReleaseInfo, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
	req := &gRPC.AppDetailRequest{
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

	req := &gRPC.ObjectRequest{
		ClusterConfig:    config,
		ReleaseName:      app.ReleaseName,
		ReleaseNamespace: app.Namespace,
		ObjectIdentifier: &gRPC.ObjectIdentifier{
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

func (impl *HelmAppServiceImpl) DeleteDBLinkedHelmApplication(ctx context.Context, app *AppIdentifier, userId int32) (*openapi.UninstallReleaseResponse, error) {
	dbConnection := impl.appRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in beginning transaction", "err", err)
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// For Helm deployments ReleaseName is App.Name
	model, err := impl.installedAppRepository.GetInstalledAppByAppName(app.ReleaseName)
	if err != nil {
		impl.logger.Errorw("error in fetching installed app", "appName", app.ReleaseName, "err", err)
		return nil, err
	}

	// If there are two releases with same name but in different namespace (eg: test -n demo-1 {Hyperion App}, test -n demo-2 {Externally Installed});
	// And if the request is received for the externally installed app, the below condition will handle
	if model.Environment.Namespace != app.Namespace {
		return nil, pg.ErrNoRows
	}

	// App Delete --> Start
	//soft delete app
	appModel := &model.App
	appModel.Active = false
	appModel.UpdatedBy = userId
	appModel.UpdatedOn = time.Now()
	err = impl.appRepository.UpdateWithTxn(appModel, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting appModel", "app", appModel)
		return nil, err
	}
	// App Delete --> End

	// InstalledApp Delete --> Start
	// soft delete install app
	model.Active = false
	model.UpdatedBy = userId
	model.UpdatedOn = time.Now()
	_, err = impl.installedAppRepository.UpdateInstalledApp(model, tx)
	if err != nil {
		impl.logger.Errorw("error while deleting install app", "error", err)
		return nil, err
	}
	// InstalledApp Delete --> End

	// InstalledAppVersions Delete --> Start
	models, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppId(model.Id)
	if err != nil {
		impl.logger.Errorw("error while fetching install app versions", "error", err)
		return nil, err
	}

	// soft delete install app versions
	for _, item := range models {
		item.Active = false
		item.UpdatedBy = userId
		item.UpdatedOn = time.Now()
		_, err = impl.installedAppRepository.UpdateInstalledAppVersion(item, tx)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
	}
	// InstalledAppVersions Delete --> End
	res, err := impl.DeleteApplication(ctx, app)
	if err != nil {
		impl.logger.Errorw("error in deleting helm application", "error", err, "appIdentifier", app)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing data in db", "err", err)
	}
	return res, nil
}

func (impl *HelmAppServiceImpl) DeleteApplication(ctx context.Context, app *AppIdentifier) (*openapi.UninstallReleaseResponse, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", app.ClusterId, "err", err)
		return nil, err
	}
	//handles the case when a user deletes namespace using kubectl but created it using devtron dashboard in
	//that case DeleteApplication returned with grpc error and the user was not able to delete the
	//cd-pipeline after helm app is created in that namespace.
	exists, err := impl.checkIfNsExists(app)
	if err != nil {
		impl.logger.Errorw("error in checking if namespace exists or not", "err", err, "clusterId", app.ClusterId)
		return nil, err
	}
	if !exists {
		return nil, models.NamespaceNotExistError{Err: fmt.Errorf("namespace %s does not exist", app.Namespace)}
	}

	req := &gRPC.ReleaseIdentifier{
		ClusterConfig:    config,
		ReleaseName:      app.ReleaseName,
		ReleaseNamespace: app.Namespace,
	}

	deleteApplicationResponse, err := impl.helmAppClient.DeleteApplication(ctx, req)
	if err != nil {
		code, message := util.GetGRPCDetailedError(err)
		if code == codes.NotFound {
			return nil, &util.ApiError{
				Code:           "404",
				HttpStatusCode: 200,
				UserMessage:    message,
			}
		}
		impl.logger.Errorw("error in deleting helm application", "err", err)
		return nil, errors.New(util.GetGRPCErrorDetailedMessage(err))
	}

	response := &openapi.UninstallReleaseResponse{
		Success: &deleteApplicationResponse.Success,
	}
	return response, nil
}

func (impl *HelmAppServiceImpl) checkIfNsExists(app *AppIdentifier) (bool, error) {
	clusterBean, err := impl.clusterService.FindById(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster bean", "error", err, "clusterId", app.ClusterId)
		return false, err
	}
	config, err := clusterBean.GetClusterConfig()
	if err != nil {
		impl.logger.Errorw("error in getting cluster config", "error", err, "clusterId", app.ClusterId)
		return false, err
	}
	v12Client, err := impl.K8sUtil.GetCoreV1Client(config)
	if err != nil {
		impl.logger.Errorw("error in getting k8s client", "err", err, "clusterHost", config.Host)
		return false, err
	}
	exists, err := impl.K8sUtil.CheckIfNsExists(app.Namespace, v12Client)
	if err != nil {
		impl.logger.Errorw("error in checking if namespace exists or not", "error", err, "clusterConfig", config)
		return false, err
	}

	return exists, nil
}

func (impl *HelmAppServiceImpl) UpdateApplication(ctx context.Context, app *AppIdentifier, request *bean.UpdateApplicationRequestDto) (*openapi.UpdateReleaseResponse, error) {
	config, err := impl.GetClusterConf(app.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", app.ClusterId, "err", err)
		return nil, err
	}

	req := &gRPC.UpgradeReleaseRequest{
		ReleaseIdentifier: &gRPC.ReleaseIdentifier{
			ClusterConfig:    config,
			ReleaseName:      app.ReleaseName,
			ReleaseNamespace: app.Namespace,
		},
		ValuesYaml: request.GetValuesYaml(),
		HistoryMax: impl.GetRevisionHistoryMaxValue(request.SourceAppType),
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

	req := &gRPC.DeploymentDetailRequest{
		ReleaseIdentifier: &gRPC.ReleaseIdentifier{
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

func (impl *HelmAppServiceImpl) InstallRelease(ctx context.Context, clusterId int, installReleaseRequest *gRPC.InstallReleaseRequest) (*gRPC.InstallReleaseResponse, error) {
	config, err := impl.GetClusterConf(clusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", clusterId, "err", err)
		return nil, err
	}

	installReleaseRequest.ReleaseIdentifier.ClusterConfig = config
	impl.logger.Debugw("helm install final request", "request", installReleaseRequest)
	installReleaseResponse, err := impl.helmAppClient.InstallRelease(ctx, installReleaseRequest)
	if err != nil {
		impl.logger.Errorw("error in installing release", "err", err)
		return nil, err
	}

	return installReleaseResponse, nil
}

func (impl *HelmAppServiceImpl) UpdateApplicationWithChartInfo(ctx context.Context, clusterId int,
	request *bean.UpdateApplicationWithChartInfoRequestDto) (*openapi.UpdateReleaseResponse, error) {
	config, err := impl.GetClusterConf(clusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", clusterId, "err", err)
		return nil, err
	}
	request.HistoryMax = impl.GetRevisionHistoryMaxValue(request.SourceAppType)
	request.ReleaseIdentifier.ClusterConfig = config

	updateReleaseResponse, err := impl.helmAppClient.UpdateApplicationWithChartInfo(ctx, request.InstallReleaseRequest)
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

	req := &gRPC.ReleaseIdentifier{
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

	req := &gRPC.RollbackReleaseRequest{
		ReleaseIdentifier: &gRPC.ReleaseIdentifier{
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
	chartRepository *gRPC.ChartRepository, extraValues map[string]interface{}, extraValuesYamlUrl string, useLatestChartVersion bool) (*openapi.UpdateReleaseResponse, error) {

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

	updateReleaseRequest := &bean.UpdateApplicationWithChartInfoRequestDto{
		InstallReleaseRequest: &gRPC.InstallReleaseRequest{
			ReleaseIdentifier: &gRPC.ReleaseIdentifier{
				ReleaseName:      appIdentifier.ReleaseName,
				ReleaseNamespace: appIdentifier.Namespace,
			},
			ChartName:       releaseInfo.DeployedAppDetail.ChartName,
			ValuesYaml:      string(mergedValuesYamlByteArr),
			ChartRepository: chartRepository,
		},
		SourceAppType: bean.SOURCE_UNKNOWN,
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
	clusterDetail, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by id", "err", err, "clusterId", clusterId)
		return nil, err
	}
	if len(clusterDetail.ErrorInConnecting) > 0 || clusterDetail.Active == false {
		clusterNotFoundErr := &util.ApiError{
			HttpStatusCode:    http.StatusInternalServerError,
			Code:              "",
			UserMessage:       fmt.Sprintf("Could not generate manifest output as the Kubernetes cluster %s is unreachable.", clusterDetail.ClusterName),
			UserDetailMessage: "",
		}
		return nil, clusterNotFoundErr
	}
	k8sServerVersion, err := impl.K8sUtil.GetKubeVersion()
	if err != nil {
		impl.logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
		return nil, err
	}
	var IsOCIRepo bool
	var registryCredential *gRPC.RegistryCredential
	var chartRepository *gRPC.ChartRepository
	dockerRegistryId := appStoreAppVersion.AppStore.DockerArtifactStoreId
	if dockerRegistryId != "" {
		ociRegistryConfigs := appStoreAppVersion.AppStore.DockerArtifactStore.OCIRegistryConfig
		if err != nil {
			impl.logger.Errorw("error in fetching oci registry config", "err", err)
			return nil, err
		}
		var ociRegistryConfig *repository2.OCIRegistryConfig
		for _, config := range ociRegistryConfigs {
			if config.RepositoryAction == repository2.STORAGE_ACTION_TYPE_PULL || config.RepositoryAction == repository2.STORAGE_ACTION_TYPE_PULL_AND_PUSH {
				ociRegistryConfig = config
				break
			}
		}
		IsOCIRepo = true
		registryCredential = &gRPC.RegistryCredential{
			RegistryUrl:  appStoreAppVersion.AppStore.DockerArtifactStore.RegistryURL,
			Username:     appStoreAppVersion.AppStore.DockerArtifactStore.Username,
			Password:     appStoreAppVersion.AppStore.DockerArtifactStore.Password,
			AwsRegion:    appStoreAppVersion.AppStore.DockerArtifactStore.AWSRegion,
			AccessKey:    appStoreAppVersion.AppStore.DockerArtifactStore.AWSAccessKeyId,
			SecretKey:    appStoreAppVersion.AppStore.DockerArtifactStore.AWSSecretAccessKey,
			RegistryType: string(appStoreAppVersion.AppStore.DockerArtifactStore.RegistryType),
			RepoName:     appStoreAppVersion.AppStore.Name,
			IsPublic:     ociRegistryConfig.IsPublic,
		}
	} else {
		chartRepository = &gRPC.ChartRepository{
			Name:     appStoreAppVersion.AppStore.ChartRepo.Name,
			Url:      appStoreAppVersion.AppStore.ChartRepo.Url,
			Username: appStoreAppVersion.AppStore.ChartRepo.UserName,
			Password: appStoreAppVersion.AppStore.ChartRepo.Password,
		}
	}

	installReleaseRequest := &gRPC.InstallReleaseRequest{
		ChartName:       appStoreAppVersion.Name,
		ChartVersion:    appStoreAppVersion.Version,
		ValuesYaml:      *templateChartRequest.ValuesYaml,
		K8SVersion:      k8sServerVersion.String(),
		ChartRepository: chartRepository,
		ReleaseIdentifier: &gRPC.ReleaseIdentifier{
			ReleaseNamespace: *templateChartRequest.Namespace,
			ReleaseName:      *templateChartRequest.ReleaseName,
		},
		IsOCIRepo:          IsOCIRepo,
		RegistryCredential: registryCredential,
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
func (impl *HelmAppServiceImpl) GetNotes(ctx context.Context, request *gRPC.InstallReleaseRequest) (string, error) {
	var notesTxt string
	response, err := impl.helmAppClient.GetNotes(ctx, request)
	if err != nil {
		impl.logger.Errorw("error in fetching chart", "err", err)
		return notesTxt, err
	}
	notesTxt = response.Notes
	return notesTxt, err
}

func (impl *HelmAppServiceImpl) ValidateOCIRegistry(ctx context.Context, OCIRegistryRequest *gRPC.RegistryCredential) bool {
	response, err := impl.helmAppClient.ValidateOCIRegistry(ctx, OCIRegistryRequest)
	if err != nil {
		impl.logger.Errorw("error in fetching chart", "err", err)
		return false
	}
	return response.IsLoggedIn
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
	clusterId, err := strconv.Atoi(component[0])
	if err != nil {
		return nil, err
	}
	if clusterId <= 0 {
		return nil, fmt.Errorf("target cluster is not provided")
	}
	return &AppIdentifier{
		ClusterId:   clusterId,
		Namespace:   component[1],
		ReleaseName: component[2],
	}, nil
}

func (impl *HelmAppServiceImpl) EncodeAppId(appIdentifier *AppIdentifier) string {
	return fmt.Sprintf("%d|%s|%s", appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
}

func (impl *HelmAppServiceImpl) appListRespProtoTransformer(deployedApps *gRPC.DeployedAppList, token string, helmAuth func(token string, object string) bool, helmCdPipelines []*pipelineConfig.Pipeline, installedHelmApps []*repository.InstalledApps) openapi.AppList {
	applicationType := "HELM-APP"
	appList := openapi.AppList{ClusterIds: &[]int32{deployedApps.ClusterId}, ApplicationType: &applicationType}
	if deployedApps.Errored {
		appList.Errored = &deployedApps.Errored
		appList.ErrorMsg = &deployedApps.ErrorMsg
	} else {
		var HelmApps []openapi.HelmApp
		//projectId := int32(0) //TODO pick from db
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
			// end

			// do not add helm apps in the list which are created using app_store
			for _, installedHelmApp := range installedHelmApps {
				if deployedapp.AppName == installedHelmApp.App.AppName && int(deployedapp.EnvironmentDetail.ClusterId) == installedHelmApp.Environment.ClusterId && deployedapp.EnvironmentDetail.Namespace == installedHelmApp.Environment.Namespace {
					toExcludeFromList = true
					break
				}
			}
			if toExcludeFromList {
				continue
			}
			// end
			lastDeployed := deployedapp.LastDeployed.AsTime()
			appDetails, appFetchErr := impl.appRepository.FindActiveByName(deployedapp.AppName)
			projectId := int32(0)
			if appFetchErr == nil {
				projectId = int32(appDetails.TeamId)
			} else {
				impl.logger.Debugw("error in fetching Project Id from app repo", "err", appFetchErr)
			}
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
			rbacObject, rbacObject2 := impl.enforcerUtil.GetHelmObjectByClusterIdNamespaceAndAppName(int(deployedapp.EnvironmentDetail.ClusterId), deployedapp.EnvironmentDetail.Namespace, deployedapp.AppName)
			isValidAuth := helmAuth(token, rbacObject) || helmAuth(token, rbacObject2)
			if isValidAuth {
				HelmApps = append(HelmApps, helmApp)
			}
		}
		appList.HelmApps = &HelmApps

	}
	return appList
}

func (impl *HelmAppServiceImpl) GetRevisionHistoryMaxValue(appType bean.SourceAppType) int32 {
	switch appType {
	case bean.SOURCE_DEVTRON_APP:
		return int32(impl.helmReleaseConfig.RevisionHistoryLimitDevtronApp)
	case bean.SOURCE_HELM_APP:
		return int32(impl.helmReleaseConfig.RevisionHistoryLimitHelmApp)
	case bean.SOURCE_EXTERNAL_HELM_APP:
		return int32(impl.helmReleaseConfig.RevisionHistoryLimitExternalHelmApp)
	default:
		return 0
	}
}
