/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cluster

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"net/http"
	"strings"
	"time"

	cluster3 "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/devtron-labs/devtron/client/grafana"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
)

// extends ClusterServiceImpl and enhances method of ClusterService with full mode specific errors
type ClusterServiceImplExtended struct {
	environmentRepository   repository.EnvironmentRepository
	grafanaClient           grafana.GrafanaClient
	installedAppRepository  repository2.InstalledAppRepository
	argoCDClientWrapper     argocdServer.ArgoClientWrapperService
	gitOpsConfigReadService config.GitOpsConfigReadService
	*ClusterServiceImpl
}

func NewClusterServiceImplExtended(environmentRepository repository.EnvironmentRepository,
	grafanaClient grafana.GrafanaClient, installedAppRepository repository2.InstalledAppRepository,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	clusterServiceImpl *ClusterServiceImpl,
	argoCDClientWrapper argocdServer.ArgoClientWrapperService) *ClusterServiceImplExtended {
	clusterServiceExt := &ClusterServiceImplExtended{
		environmentRepository:   environmentRepository,
		grafanaClient:           grafanaClient,
		installedAppRepository:  installedAppRepository,
		argoCDClientWrapper:     argoCDClientWrapper,
		gitOpsConfigReadService: gitOpsConfigReadService,
		ClusterServiceImpl:      clusterServiceImpl,
	}
	return clusterServiceExt
}

func (impl *ClusterServiceImplExtended) FindAllWithoutConfig() ([]*bean.ClusterBean, error) {
	beans, err := impl.FindAll()
	if err != nil {
		return nil, err
	}
	for _, bean := range beans {
		bean.Config = map[string]string{commonBean.BearerToken: ""}
	}
	return beans, nil
}

func (impl *ClusterServiceImplExtended) GetClusterFullModeDTO(beans []*bean.ClusterBean) ([]*bean.ClusterBean, error) {
	//devtron full mode logic
	var clusterIds []int
	for _, cluster := range beans {
		clusterIds = append(clusterIds, cluster.Id)
	}
	clusterComponentsMap := make(map[int][]*repository2.InstalledAppVersions)
	charts, err := impl.installedAppRepository.GetInstalledAppVersionByClusterIdsV2(clusterIds)
	if err != nil {
		impl.logger.Errorw("error on fetching installed apps for cluster ids", "err", err, "clusterIds", clusterIds)
		return nil, err
	}
	for _, item := range charts {
		if _, ok := clusterComponentsMap[item.InstalledApp.Environment.ClusterId]; !ok {
			var charts []*repository2.InstalledAppVersions
			charts = append(charts, item)
			clusterComponentsMap[item.InstalledApp.Environment.ClusterId] = charts
		} else {
			charts := clusterComponentsMap[item.InstalledApp.Environment.ClusterId]
			charts = append(charts, item)
			clusterComponentsMap[item.InstalledApp.Environment.ClusterId] = charts
		}
	}

	for _, item := range beans {
		defaultClusterComponents := make([]*bean.DefaultClusterComponent, 0)
		if _, ok := clusterComponentsMap[item.Id]; ok {
			charts := clusterComponentsMap[item.Id]
			failed := false
			chartLen := 0
			chartPass := 0
			if len(charts) > 0 {
				chartLen = len(charts)
			}
			for _, chart := range charts {
				defaultClusterComponent := &bean.DefaultClusterComponent{}
				defaultClusterComponent.AppId = chart.InstalledApp.AppId
				defaultClusterComponent.InstalledAppId = chart.InstalledApp.Id
				defaultClusterComponent.EnvId = chart.InstalledApp.EnvironmentId
				defaultClusterComponent.EnvName = chart.InstalledApp.Environment.Name
				defaultClusterComponent.ComponentName = chart.AppStoreApplicationVersion.AppStore.Name
				defaultClusterComponent.Status = chart.InstalledApp.Status.String()
				defaultClusterComponents = append(defaultClusterComponents, defaultClusterComponent)
				if chart.InstalledApp.Status == appStoreBean.QUE_ERROR || chart.InstalledApp.Status == appStoreBean.TRIGGER_ERROR ||
					chart.InstalledApp.Status == appStoreBean.DEQUE_ERROR || chart.InstalledApp.Status == appStoreBean.GIT_ERROR ||
					chart.InstalledApp.Status == appStoreBean.ACD_ERROR {
					failed = true
				}
				if chart.InstalledApp.Status == appStoreBean.DEPLOY_SUCCESS {
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

func (impl *ClusterServiceImplExtended) FindAll() ([]*bean.ClusterBean, error) {
	beans, err := impl.ClusterServiceImpl.FindAll()
	if err != nil {
		return nil, err
	}
	return impl.GetClusterFullModeDTO(beans)
}

func (impl *ClusterServiceImplExtended) FindAllExceptVirtual() ([]*bean.ClusterBean, error) {
	beans, err := impl.ClusterServiceImpl.FindAll()
	if err != nil {
		return nil, err
	}
	return impl.GetClusterFullModeDTO(beans)
}

func (impl *ClusterServiceImplExtended) Update(ctx context.Context, bean *bean.ClusterBean, userId int32) (*bean.ClusterBean, error) {
	gitOpsConfigurationStatus, err1 := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err1 != nil {
		return nil, err1
	}

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
		//if the request doesn't have a non empty prometheus url and we don't have a GrafanaDataSourceId defined yet, no point in
		//going to grafana client and trying to get data source
		if bean.PrometheusUrl != "" && env.GrafanaDatasourceId != 0 {
			promDatasource, err := impl.grafanaClient.GetDatasource(env.GrafanaDatasourceId)
			if err.Error() == grafana.DoesNotExist {
				// datasource might have been deleted manually, have to create a new datasource
				grafanaDatasourceId, _ := impl.CreateGrafanaDataSource(bean, env)
				env.GrafanaDatasourceId = grafanaDatasourceId
			} else {
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
				impl.logger.Errorw("Error while updating the datasource", "Datasource id : ", env.GrafanaDatasourceId, "error", err)
				return nil, err
			}
		}
	}

	// if git-ops configured and ArgoCD module is installed, then only update cluster in ACD, otherwise ignore
	if gitOpsConfigurationStatus.IsGitOpsConfiguredAndArgoCdInstalled() {

		cl := impl.ConvertClusterBeanObjectToCluster(bean)
		_, err = impl.argoCDClientWrapper.CreateCluster(ctx, &cluster3.ClusterCreateRequest{Upsert: true, Cluster: cl})
		if err != nil {
			impl.logger.Errorw("service err, Update", "error", err, "payload", cl)
			userMsg := "failed to update on cluster via ACD"
			if strings.Contains(err.Error(), commonBean.DefaultClusterUrl) {
				userMsg = fmt.Sprintf("%s, %s", err.Error(), ", successfully updated in ACD")
			}
			err = &util.ApiError{
				Code:            constants.ClusterUpdateACDFailed,
				InternalMessage: err.Error(),
				UserMessage:     userMsg,
			}
			return nil, err
		}
	}

	if bean.HasConfigOrUrlChanged {
		impl.ClusterServiceImpl.SyncNsInformer(bean)
	}

	return bean, err
}

func (impl *ClusterServiceImplExtended) CreateGrafanaDataSource(clusterBean *bean.ClusterBean, env *repository.Environment) (int, error) {
	grafanaDatasourceId := env.GrafanaDatasourceId
	if grafanaDatasourceId == 0 {
		//starts grafana creation
		// appending envId to ensure unique datasource name for each environment (ex- env got deleted and created with same name)
		DataSourceName := "Prometheus-" + env.Name + "-EnvId-" + fmt.Sprint(env.Id)
		createDatasourceReq := grafana.CreateDatasourceRequest{
			Name:      DataSourceName,
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

func (impl *ClusterServiceImplExtended) Save(ctx context.Context, bean *bean.ClusterBean, userId int32) (*bean.ClusterBean, error) {
	gitOpsConfigurationStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		return nil, err
	}

	clusterBean, err := impl.ClusterServiceImpl.Save(ctx, bean, userId)
	if err != nil {
		return nil, err
	}

	// if git-ops configured and ArgoCD module is installed, then only add cluster in ACD, otherwise ignore
	if gitOpsConfigurationStatus.IsGitOpsConfiguredAndArgoCdInstalled() {
		//create it into argo cd as well
		cl := impl.ConvertClusterBeanObjectToCluster(bean)

		_, err = impl.argoCDClientWrapper.CreateCluster(ctx, &cluster3.ClusterCreateRequest{Upsert: true, Cluster: cl})
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
	}

	//on successful creation of new cluster, update informer cache for namespace group by cluster
	impl.SyncNsInformer(bean)
	return clusterBean, nil
}

func (impl ClusterServiceImplExtended) DeleteFromDb(bean *bean.ClusterBean, userId int32) (string, error) {
	existingCluster, err := impl.clusterRepository.FindById(bean.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete.", "id", bean.Id)
		return "", err
	}
	deleteReq := existingCluster
	deleteReq.UpdatedOn = time.Now()
	deleteReq.UpdatedBy = userId
	err = impl.clusterRepository.MarkClusterDeleted(deleteReq)
	if err != nil {
		impl.logger.Errorw("error in deleting cluster", "id", bean.Id, "err", err)
		return "", err
	}
	return existingCluster.ClusterName, nil
}
