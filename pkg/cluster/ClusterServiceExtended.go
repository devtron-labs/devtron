package cluster

import (
	"context"
	"fmt"
	cluster3 "github.com/argoproj/argo-cd/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	cluster2 "github.com/devtron-labs/devtron/client/argocdServer/cluster"
	"github.com/devtron-labs/devtron/client/grafana"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/appstore"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

//extends ClusterServiceImpl and enhances method of ClusterService with full mode specific errors
type ClusterServiceImplExtended struct {
	environmentRepository  repository.EnvironmentRepository
	grafanaClient          grafana.GrafanaClient
	installedAppRepository appstore.InstalledAppRepository
	clusterServiceCD       cluster2.ServiceClient
	*ClusterServiceImpl
}

func NewClusterServiceImplExtended(repository repository.ClusterRepository, environmentRepository repository.EnvironmentRepository,
	grafanaClient grafana.GrafanaClient, logger *zap.SugaredLogger, installedAppRepository appstore.InstalledAppRepository,
	K8sUtil *util.K8sUtil,
	clusterServiceCD cluster2.ServiceClient) *ClusterServiceImplExtended {
	return &ClusterServiceImplExtended{
		environmentRepository:  environmentRepository,
		grafanaClient:          grafanaClient,
		installedAppRepository: installedAppRepository,
		clusterServiceCD:       clusterServiceCD,
		ClusterServiceImpl: &ClusterServiceImpl{
			clusterRepository: repository,
			logger:            logger,
			K8sUtil:           K8sUtil,
		},
	}
}

func (impl *ClusterServiceImplExtended) FindAll() ([]*ClusterBean, error) {
	beans, err := impl.ClusterServiceImpl.FindAll()
	if err != nil {
		return nil, err
	}
	//devtron full mode logic
	var clusterIds []int
	for _, cluster := range beans {
		clusterIds = append(clusterIds, cluster.Id)
	}
	clusterComponentsMap := make(map[int][]*appstore.InstalledAppVersions)
	charts, err := impl.installedAppRepository.GetInstalledAppVersionByClusterIdsV2(clusterIds)
	if err != nil {
		impl.logger.Errorw("error on fetching installed apps for cluster ids", "err", err, "clusterIds", clusterIds)
		return nil, err
	}
	for _, item := range charts {
		if _, ok := clusterComponentsMap[item.InstalledApp.Environment.ClusterId]; !ok {
			var charts []*appstore.InstalledAppVersions
			charts = append(charts, item)
			clusterComponentsMap[item.InstalledApp.Environment.ClusterId] = charts
		} else {
			charts := clusterComponentsMap[item.InstalledApp.Environment.ClusterId]
			charts = append(charts, item)
			clusterComponentsMap[item.InstalledApp.Environment.ClusterId] = charts
		}
	}

	for _, item := range beans {
		defaultClusterComponents := make([]*DefaultClusterComponent, 0)
		if _, ok := clusterComponentsMap[item.Id]; ok {
			charts := clusterComponentsMap[item.Id]
			failed := false
			chartLen := 0
			chartPass := 0
			if len(charts) > 0 {
				chartLen = len(charts)
			}
			for _, chart := range charts {
				defaultClusterComponent := &DefaultClusterComponent{}
				defaultClusterComponent.AppId = chart.InstalledApp.AppId
				defaultClusterComponent.InstalledAppId = chart.InstalledApp.Id
				defaultClusterComponent.EnvId = chart.InstalledApp.EnvironmentId
				defaultClusterComponent.EnvName = chart.InstalledApp.Environment.Name
				defaultClusterComponent.ComponentName = chart.AppStoreApplicationVersion.AppStore.Name
				defaultClusterComponent.Status = chart.InstalledApp.Status.String()
				defaultClusterComponents = append(defaultClusterComponents, defaultClusterComponent)
				if chart.InstalledApp.Status == appstore.QUE_ERROR || chart.InstalledApp.Status == appstore.TRIGGER_ERROR ||
					chart.InstalledApp.Status == appstore.DEQUE_ERROR || chart.InstalledApp.Status == appstore.GIT_ERROR ||
					chart.InstalledApp.Status == appstore.ACD_ERROR {
					failed = true
				}
				if chart.InstalledApp.Status == appstore.DEPLOY_SUCCESS {
					chartPass = chartPass + 1
				}
			}
			if chartPass == chartLen {
				item.AgentInstallationStage = 2
			} else if failed {
				item.AgentInstallationStage = 3
			} else {
				item.AgentInstallationStage = 1
			}
		}
		if item.Id == 1 {
			item.AgentInstallationStage = -1
		}
		item.DefaultClusterComponent = defaultClusterComponents
	}
	return beans, nil
}

func (impl *ClusterServiceImplExtended) Update(ctx context.Context, bean *ClusterBean, userId int32) (*ClusterBean, error) {
	bean, err := impl.ClusterServiceImpl.Update(ctx, bean, userId)
	if err != nil {
		return nil, err
	}

	envs, err := impl.environmentRepository.FindByClusterId(bean.Id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return nil, err
	}

	// TODO: Can be called in goroutines if performance issue
	for _, env := range envs {
		if len(bean.PrometheusUrl) > 0 && env.GrafanaDatasourceId == 0 {
			grafanaDatasourceId, _ := impl.CreateGrafanaDataSource(bean, env)
			if grafanaDatasourceId == 0 {
				impl.logger.Errorw("unable to create data source for environment which doesn't exists", "env", env)
				continue
			}
			env.GrafanaDatasourceId = grafanaDatasourceId
		}
		promDatasource, err := impl.grafanaClient.GetDatasource(env.GrafanaDatasourceId)
		if err != nil {
			impl.logger.Errorw("error on getting data source", "err", err)
			return nil, err
		}

		updateDatasourceReq := grafana.UpdateDatasourceRequest{
			Id:                env.GrafanaDatasourceId,
			OrgId:             promDatasource.OrgId,
			Name:              promDatasource.Name,
			Type:              promDatasource.Type,
			Url:               bean.PrometheusUrl,
			Access:            promDatasource.Access,
			BasicAuth:         promDatasource.BasicAuth,
			BasicAuthUser:     promDatasource.BasicAuthUser,
			BasicAuthPassword: promDatasource.BasicAuthPassword,
			JsonData:          promDatasource.JsonData,
		}

		if bean.PrometheusAuth != nil {
			secureJsonData := &grafana.SecureJsonData{}
			if len(bean.PrometheusAuth.UserName) > 0 {
				updateDatasourceReq.BasicAuthUser = bean.PrometheusAuth.UserName
				updateDatasourceReq.BasicAuthPassword = bean.PrometheusAuth.Password
				secureJsonData.BasicAuthPassword = bean.PrometheusAuth.Password
			}
			if len(bean.PrometheusAuth.TlsClientCert) > 0 {
				secureJsonData.TlsClientCert = bean.PrometheusAuth.TlsClientCert
				secureJsonData.TlsClientKey = bean.PrometheusAuth.TlsClientKey
				updateDatasourceReq.BasicAuth = false

				jsonData := &grafana.JsonData{
					HttpMethod: http.MethodGet,
					TlsAuth:    true,
				}
				updateDatasourceReq.JsonData = *jsonData
			}
			updateDatasourceReq.SecureJsonData = secureJsonData
		}
		_, err = impl.grafanaClient.UpdateDatasource(updateDatasourceReq, env.GrafanaDatasourceId)
		if err != nil {
			impl.logger.Error(err)
			return nil, err
		}
	}
	configMap := bean.Config
	serverUrl := bean.ServerUrl
	bearerToken := ""
	if configMap["bearer_token"] != "" {
		bearerToken = configMap["bearer_token"]
	}

	tlsConfig := v1alpha1.TLSClientConfig{
		Insecure: true,
	}
	cdClusterConfig := v1alpha1.ClusterConfig{
		BearerToken:     bearerToken,
		TLSClientConfig: tlsConfig,
	}

	cl := &v1alpha1.Cluster{
		Name:   bean.ClusterName,
		Server: serverUrl,
		Config: cdClusterConfig,
	}

	_, err = impl.clusterServiceCD.Update(ctx, &cluster3.ClusterUpdateRequest{Cluster: cl})

	if err != nil {
		impl.logger.Errorw("service err, Update", "error", err, "payload", cl)
		userMsg := "failed to update on cluster via ACD"
		if strings.Contains(err.Error(), "https://kubernetes.default.svc") {
			userMsg = fmt.Sprintf("%s, %s", err.Error(), ", successfully updated in ACD")
		}
		err = &util.ApiError{
			Code:            constants.ClusterUpdateACDFailed,
			InternalMessage: err.Error(),
			UserMessage:     userMsg,
		}
		return nil, err
	}
	return bean, err
}

func (impl *ClusterServiceImplExtended) CreateGrafanaDataSource(clusterBean *ClusterBean, env *repository.Environment) (int, error) {
	grafanaDatasourceId := env.GrafanaDatasourceId
	if grafanaDatasourceId == 0 {
		//starts grafana creation
		createDatasourceReq := grafana.CreateDatasourceRequest{
			Name:      "Prometheus-" + env.Name,
			Type:      "prometheus",
			Url:       clusterBean.PrometheusUrl,
			Access:    "proxy",
			BasicAuth: true,
		}

		if clusterBean.PrometheusAuth != nil {
			secureJsonData := &grafana.SecureJsonData{}
			if len(clusterBean.PrometheusAuth.UserName) > 0 {
				createDatasourceReq.BasicAuthUser = clusterBean.PrometheusAuth.UserName
				createDatasourceReq.BasicAuthPassword = clusterBean.PrometheusAuth.Password
				secureJsonData.BasicAuthPassword = clusterBean.PrometheusAuth.Password
			}
			if len(clusterBean.PrometheusAuth.TlsClientCert) > 0 {
				secureJsonData.TlsClientCert = clusterBean.PrometheusAuth.TlsClientCert
				secureJsonData.TlsClientKey = clusterBean.PrometheusAuth.TlsClientKey

				jsonData := &grafana.JsonData{
					HttpMethod: http.MethodGet,
					TlsAuth:    true,
				}
				createDatasourceReq.JsonData = jsonData
			}
			createDatasourceReq.SecureJsonData = secureJsonData
		}

		grafanaResp, err := impl.grafanaClient.CreateDatasource(createDatasourceReq)
		if err != nil {
			impl.logger.Errorw("error on create grafana datasource", "err", err)
			return 0, err
		}
		//ends grafana creation
		grafanaDatasourceId = grafanaResp.Id
	}
	env.GrafanaDatasourceId = grafanaDatasourceId
	err := impl.environmentRepository.Update(env)
	if err != nil {
		impl.logger.Errorw("error in updating environment", "err", err)
		return 0, err
	}
	return grafanaDatasourceId, nil
}

func (impl *ClusterServiceImplExtended) Save(ctx context.Context, bean *ClusterBean, userId int32) (*ClusterBean, error) {
	clusterBean, err := impl.ClusterServiceImpl.Save(ctx, bean, userId)
	if err != nil {
		return nil, err
	}

	//create it into argo cd as well
	configMap := bean.Config
	serverUrl := bean.ServerUrl
	bearerToken := ""
	if configMap["bearer_token"] != "" {
		bearerToken = configMap["bearer_token"]
	}
	tlsConfig := v1alpha1.TLSClientConfig{
		Insecure: true,
	}
	cdClusterConfig := v1alpha1.ClusterConfig{
		BearerToken:     bearerToken,
		TLSClientConfig: tlsConfig,
	}

	cl := &v1alpha1.Cluster{
		Name:   bean.ClusterName,
		Server: serverUrl,
		Config: cdClusterConfig,
	}

	_, err = impl.clusterServiceCD.Create(ctx, &cluster3.ClusterCreateRequest{Upsert: true, Cluster: cl})
	if err != nil {
		impl.logger.Errorw("service err, Save", "err", err, "payload", cl)
		err1 := impl.ClusterServiceImpl.Delete(bean, userId) //FIXME nishant call local
		if err1 != nil {
			impl.logger.Errorw("service err, Save, delete on rollback", "err", err, "payload", bean)
			err = &util.ApiError{
				Code:            constants.ClusterDBRollbackFailed,
				InternalMessage: err.Error(),
				UserMessage:     "failed to rollback cluster from db as it has failed in registering on ACD",
			}
			return nil, err

		}
		err = &util.ApiError{
			Code:            constants.ClusterCreateACDFailed,
			InternalMessage: err.Error(),
			UserMessage:     "failed to register on ACD, rollback completed from db",
		}
		return nil, err

	}
	return clusterBean, nil
}
