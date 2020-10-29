/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cluster

import (
	"fmt"
	"github.com/devtron-labs/devtron/client/grafana"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/appstore"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/util"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ClusterBean struct {
	Id                      int                        `json:"id,omitempty" validate:"number"`
	ClusterName             string                     `json:"cluster_name,omitempty" validate:"required"`
	ServerUrl               string                     `json:"server_url,omitempty" validate:"url,required"`
	PrometheusUrl           string                     `json:"prometheus_url,omitempty" validate:"url,required"`
	Active                  bool                       `json:"active"`
	Config                  map[string]string          `json:"config,omitempty" validate:"required"`
	PrometheusAuth          *PrometheusAuth            `json:"prometheusAuth,omitempty"`
	DefaultClusterComponent []*DefaultClusterComponent `json:"defaultClusterComponent"`
	AgentInstallationStage  int                        `json:"agentInstallationStage,notnull"`
}

type PrometheusAuth struct {
	UserName      string `json:"userName,omitempty"`
	Password      string `json:"password,omitempty"`
	TlsClientCert string `json:"tlsClientCert,omitempty"`
	TlsClientKey  string `json:"tlsClientKey,omitempty"`
}

type DefaultClusterComponent struct {
	ComponentName  string `json:"name"`
	AppId          int    `json:"appId"`
	InstalledAppId int    `json:"installedAppId,omitempty"`
	EnvId          int    `json:"envId"`
	EnvName        string `json:"envName"`
	Status         string `json:"status"`
}

type ClusterService interface {
	Save(bean *ClusterBean, userId int32) (*ClusterBean, error)
	FindOne(clusterName string) (*ClusterBean, error)
	FindOneActive(clusterName string) (*ClusterBean, error)
	FindAll() ([]*ClusterBean, error)
	FindAllActive() ([]ClusterBean, error)

	FindById(id int) (*ClusterBean, error)
	FindByIds(id []int) ([]ClusterBean, error)
	Update(bean *ClusterBean, userId int32) (*ClusterBean, error)
	Delete(bean *ClusterBean, userId int32) error

	FindAllForAutoComplete() ([]ClusterBean, error)
	CreateGrafanaDataSource(clusterBean *ClusterBean, env *cluster.Environment) (int, error)
}

type ClusterServiceImpl struct {
	clusterRepository              cluster.ClusterRepository
	environmentRepository          cluster.EnvironmentRepository
	grafanaClient                  grafana.GrafanaClient
	logger                         *zap.SugaredLogger
	installedAppRepository         appstore.InstalledAppRepository
	clusterInstalledAppsRepository appstore.ClusterInstalledAppsRepository
}

func NewClusterServiceImpl(repository cluster.ClusterRepository, environmentRepository cluster.EnvironmentRepository,
	grafanaClient grafana.GrafanaClient, logger *zap.SugaredLogger, installedAppRepository appstore.InstalledAppRepository,
	clusterInstalledAppsRepository appstore.ClusterInstalledAppsRepository) *ClusterServiceImpl {
	return &ClusterServiceImpl{
		clusterRepository:              repository,
		logger:                         logger,
		environmentRepository:          environmentRepository,
		grafanaClient:                  grafanaClient,
		installedAppRepository:         installedAppRepository,
		clusterInstalledAppsRepository: clusterInstalledAppsRepository,
	}
}

func (impl ClusterServiceImpl) Save(bean *ClusterBean, userId int32) (*ClusterBean, error) {
	model := &cluster.Cluster{
		ClusterName:        bean.ClusterName,
		Active:             bean.Active,
		ServerUrl:          bean.ServerUrl,
		Config:             bean.Config,
		PrometheusEndpoint: bean.PrometheusUrl,
	}

	if bean.PrometheusAuth != nil {
		model.PUserName = bean.PrometheusAuth.UserName
		model.PPassword = bean.PrometheusAuth.Password
		model.PTlsClientCert = bean.PrometheusAuth.TlsClientCert
		model.PTlsClientKey = bean.PrometheusAuth.TlsClientKey
	}

	model.CreatedBy = userId
	model.UpdatedBy = userId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	err := impl.clusterRepository.Save(model)
	if err != nil {
		impl.logger.Errorw("error in saving cluster in db", "err", err)
		err = &util.ApiError{
			Code:            constants.ClusterCreateDBFailed,
			InternalMessage: "cluster creation failed in db",
			UserMessage:     fmt.Sprintf("requested by %d", userId),
		}
	}
	bean.Id = model.Id
	return bean, err
}

func (impl ClusterServiceImpl) FindOne(clusterName string) (*ClusterBean, error) {
	model, err := impl.clusterRepository.FindOne(clusterName)
	if err != nil {
		return nil, err
	}
	bean := &ClusterBean{
		Id:            model.Id,
		ClusterName:   model.ClusterName,
		ServerUrl:     model.ServerUrl,
		PrometheusUrl: model.PrometheusEndpoint,
		Active:        model.Active,
		Config:        model.Config,
	}
	return bean, nil
}

func (impl ClusterServiceImpl) FindOneActive(clusterName string) (*ClusterBean, error) {
	model, err := impl.clusterRepository.FindOneActive(clusterName)
	if err != nil {
		return nil, err
	}
	bean := &ClusterBean{
		Id:            model.Id,
		ClusterName:   model.ClusterName,
		ServerUrl:     model.ServerUrl,
		PrometheusUrl: model.PrometheusEndpoint,
		Active:        model.Active,
		Config:        model.Config,
	}
	return bean, nil
}

func (impl ClusterServiceImpl) FindAll() ([]*ClusterBean, error) {
	model, err := impl.clusterRepository.FindAll()
	if err != nil {
		return nil, err
	}
	var clusterIds []int
	var beans []*ClusterBean
	for _, m := range model {
		beans = append(beans, &ClusterBean{
			Id:            m.Id,
			ClusterName:   m.ClusterName,
			PrometheusUrl: m.PrometheusEndpoint,
			ServerUrl:     m.ServerUrl,
			Active:        m.Active,
		})
		clusterIds = append(clusterIds, m.Id)
	}

	clusterComponentsMap := make(map[int][]*appstore.InstalledAppVersions)
	charts, err := impl.installedAppRepository.GetInstalledAppVersionByClusterIdsV2(clusterIds)
	if err != nil {
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
		if _, ok := clusterComponentsMap[item.Id]; ok {
			charts := clusterComponentsMap[item.Id]
			var defaultClusterComponents []*DefaultClusterComponent
			failed := false
			chartLen := 0
			chartPass := 0
			if charts != nil && len(charts) > 0 {
				chartLen = len(charts)
			}
			for _, chart := range charts {
				defaultClusterComponent := &DefaultClusterComponent{}
				defaultClusterComponent.AppId = chart.InstalledApp.AppId
				defaultClusterComponent.InstalledAppId = chart.Id
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
			item.DefaultClusterComponent = defaultClusterComponents
			if chartPass == chartLen {
				item.AgentInstallationStage = 2
			} else if failed {
				item.AgentInstallationStage = 3
			} else {
				item.AgentInstallationStage = 1
			}
		}
	}
	return beans, nil
}

func (impl ClusterServiceImpl) FindAllActive() ([]ClusterBean, error) {
	model, err := impl.clusterRepository.FindAllActive()
	if err != nil {
		return nil, err
	}
	var beans []ClusterBean
	for _, m := range model {
		beans = append(beans, ClusterBean{
			Id:            m.Id,
			ClusterName:   m.ClusterName,
			ServerUrl:     m.ServerUrl,
			Active:        m.Active,
			PrometheusUrl: m.PrometheusEndpoint,
			Config:        m.Config,
		})
	}
	return beans, nil
}

func (impl ClusterServiceImpl) FindById(id int) (*ClusterBean, error) {
	model, err := impl.clusterRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	bean := &ClusterBean{
		Id:            model.Id,
		ClusterName:   model.ClusterName,
		ServerUrl:     model.ServerUrl,
		PrometheusUrl: model.PrometheusEndpoint,
		Active:        model.Active,
		Config:        model.Config,
	}
	prometheusAuth := &PrometheusAuth{
		UserName:      model.PUserName,
		Password:      model.PPassword,
		TlsClientCert: model.PTlsClientCert,
		TlsClientKey:  model.PTlsClientKey,
	}
	bean.PrometheusAuth = prometheusAuth
	return bean, nil
}

func (impl ClusterServiceImpl) FindByIds(ids []int) ([]ClusterBean, error) {
	models, err := impl.clusterRepository.FindByIds(ids)
	if err != nil {
		return nil, err
	}
	var beans []ClusterBean

	for _, model := range models {
		beans = append(beans, ClusterBean{
			Id:            model.Id,
			ClusterName:   model.ClusterName,
			ServerUrl:     model.ServerUrl,
			PrometheusUrl: model.PrometheusEndpoint,
			Active:        model.Active,
			Config:        model.Config,
		})
	}
	return beans, nil
}

func (impl ClusterServiceImpl) Update(bean *ClusterBean, userId int32) (*ClusterBean, error) {
	model, err := impl.clusterRepository.FindById(bean.Id)
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	model.ClusterName = bean.ClusterName
	model.ServerUrl = bean.ServerUrl
	model.PrometheusEndpoint = bean.PrometheusUrl

	if bean.PrometheusAuth != nil {
		model.PUserName = bean.PrometheusAuth.UserName
		model.PPassword = bean.PrometheusAuth.Password
		model.PTlsClientCert = bean.PrometheusAuth.TlsClientCert
		model.PTlsClientKey = bean.PrometheusAuth.TlsClientKey
	}

	model.Active = bean.Active
	model.Config = bean.Config
	model.UpdatedBy = userId
	model.UpdatedOn = time.Now()
	err = impl.clusterRepository.Update(model)
	if err != nil {
		err = &util.ApiError{
			Code:            constants.ClusterUpdateDBFailed,
			InternalMessage: "cluster update failed in db",
			UserMessage:     fmt.Sprintf("requested by %d", userId),
		}
		return bean, err
	}

	envs, err := impl.environmentRepository.FindByClusterId(bean.Id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return nil, err
	}

	// TODO: Can be called in goroutines if performance issue
	for _, env := range envs {
		if env.GrafanaDatasourceId == 0 {
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

	bean.Id = model.Id
	return bean, err
}

func (impl ClusterServiceImpl) Delete(bean *ClusterBean, userId int32) error {
	model, err := impl.clusterRepository.FindById(bean.Id)
	if err != nil {
		return err
	}
	return impl.clusterRepository.Delete(model)
}

func (impl ClusterServiceImpl) FindAllForAutoComplete() ([]ClusterBean, error) {
	model, err := impl.clusterRepository.FindAll()
	if err != nil {
		return nil, err
	}
	var beans []ClusterBean
	for _, m := range model {
		beans = append(beans, ClusterBean{
			Id:          m.Id,
			ClusterName: m.ClusterName,
		})
	}
	return beans, nil
}

func (impl ClusterServiceImpl) CreateGrafanaDataSource(clusterBean *ClusterBean, env *cluster.Environment) (int, error) {
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


