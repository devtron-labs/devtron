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
	"github.com/devtron-labs/devtron/client/grafana"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"time"
)

type EnvironmentBean struct {
	Id                 int    `json:"id,omitempty" validate:"number"`
	Environment        string `json:"environment_name,omitempty" validate:"required,max=50"`
	ClusterId          int    `json:"cluster_id,omitempty" validate:"number,required"`
	ClusterName        string `json:"cluster_name,omitempty"`
	Active             bool   `json:"active"`
	Default            bool   `json:"default"`
	PrometheusEndpoint string `json:"prometheus_endpoint,omitempty"`
	Namespace          string `json:"namespace,omitempty" validate:"max=50"`
	CdArgoSetup        bool   `json:"isClusterCdActive"`
}

const ClusterName = "default_cluster"
const TokenFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

type EnvironmentService interface {
	FindOne(environment string) (*EnvironmentBean, error)
	Create(mappings *EnvironmentBean, userId int32) (*EnvironmentBean, error)
	GetAll() ([]EnvironmentBean, error)
	GetAllActive() ([]EnvironmentBean, error)

	FindById(id int) (*EnvironmentBean, error)
	Update(mappings *EnvironmentBean, userId int32) (*EnvironmentBean, error)
	FindClusterByEnvId(id int) (*ClusterBean, error)
	GetEnvironmentListForAutocomplete() ([]EnvironmentBean, error)
	FindByIds(ids []*int) ([]*EnvironmentBean, error)
	FindByNamespaceAndClusterName(namespaces string, clusterName string) (*cluster.Environment, error)
	GetByClusterId(id int) ([]*EnvironmentBean, error)
}

type EnvironmentServiceImpl struct {
	environmentRepository   cluster.EnvironmentRepository
	logger                  *zap.SugaredLogger
	clusterService          ClusterService
	K8sUtil                 *util.K8sUtil
	propertiesConfigService pipeline.PropertiesConfigService
	grafanaClient           grafana.GrafanaClient
}

func NewEnvironmentServiceImpl(environmentRepository cluster.EnvironmentRepository,
	clusterService ClusterService, logger *zap.SugaredLogger,
	K8sUtil *util.K8sUtil,
	propertiesConfigService pipeline.PropertiesConfigService,
	grafanaClient grafana.GrafanaClient,
) *EnvironmentServiceImpl {
	return &EnvironmentServiceImpl{
		environmentRepository:   environmentRepository,
		logger:                  logger,
		clusterService:          clusterService,
		K8sUtil:                 K8sUtil,
		propertiesConfigService: propertiesConfigService,
		grafanaClient:           grafanaClient,
	}
}

func (impl EnvironmentServiceImpl) Create(mappings *EnvironmentBean, userId int32) (*EnvironmentBean, error) {
	existingEnvs, err := impl.environmentRepository.FindByClusterId(mappings.ClusterId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetch", "err", err)
		return nil, err
	}
	err = impl.validateNamespaces(mappings.Namespace, existingEnvs)
	if err != nil {
		return nil, err
	}

	clusterBean, err := impl.clusterService.FindById(mappings.ClusterId)
	if err != nil {
		return nil, err
	}

	model := &cluster.Environment{
		Name:      mappings.Environment,
		ClusterId: mappings.ClusterId,
		Active:    mappings.Active,
		Namespace: mappings.Namespace,
		Default:   mappings.Default,
	}
	model.CreatedBy = userId
	model.UpdatedBy = userId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	err = impl.environmentRepository.Create(model)
	if err != nil {
		impl.logger.Errorw("error in saving environment", "err", err)
		return mappings, err
	}
	if len(model.Namespace) > 0 {
		cfg, err := impl.getClusterConfig(clusterBean)
		if err != nil {
			return nil, err
		}
		if err := impl.K8sUtil.CreateNsIfNotExists(model.Namespace, cfg); err != nil {
			impl.logger.Errorw("error in creating ns", "ns", model.Namespace, "err", err)
		}

	}

	_, err = impl.clusterService.CreateGrafanaDataSource(clusterBean, model)
	if err != nil {
		impl.logger.Errorw("unable to create grafana data source", "env", model)
	}

	mappings.Id = model.Id
	return mappings, nil
}

func (impl EnvironmentServiceImpl) FindOne(environment string) (*EnvironmentBean, error) {
	model, err := impl.environmentRepository.FindOne(environment)
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
		return nil, err
	}
	bean := &EnvironmentBean{
		Id:                 model.Id,
		Environment:        model.Name,
		ClusterId:          model.Cluster.Id,
		Active:             model.Active,
		PrometheusEndpoint: model.Cluster.PrometheusEndpoint,
		Namespace:          model.Namespace,
		Default:            model.Default,
	}
	return bean, nil
}

func (impl EnvironmentServiceImpl) GetAll() ([]EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []EnvironmentBean
	for _, model := range models {
		beans = append(beans, EnvironmentBean{
			Id:                 model.Id,
			Environment:        model.Name,
			ClusterId:          model.Cluster.Id,
			ClusterName:        model.Cluster.ClusterName,
			Active:             model.Active,
			PrometheusEndpoint: model.Cluster.PrometheusEndpoint,
			Namespace:          model.Namespace,
			Default:            model.Default,
			CdArgoSetup:        model.Cluster.CdArgoSetup,
		})
	}
	return beans, nil
}

func (impl EnvironmentServiceImpl) GetAllActive() ([]EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []EnvironmentBean
	for _, model := range models {
		beans = append(beans, EnvironmentBean{
			Id:                 model.Id,
			Environment:        model.Name,
			ClusterId:          model.Cluster.Id,
			Active:             model.Active,
			PrometheusEndpoint: model.Cluster.PrometheusEndpoint,
			Namespace:          model.Namespace,
			Default:            model.Default,
		})
	}
	return beans, nil
}

func (impl EnvironmentServiceImpl) FindById(id int) (*EnvironmentBean, error) {
	model, err := impl.environmentRepository.FindById(id)
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
		return nil, err
	}
	bean := &EnvironmentBean{
		Id:                 model.Id,
		Environment:        model.Name,
		ClusterId:          model.Cluster.Id,
		Active:             model.Active,
		PrometheusEndpoint: model.Cluster.PrometheusEndpoint,
		Namespace:          model.Namespace,
		Default:            model.Default,
	}

	/*clusterBean := &ClusterBean{
		id:model.Cluster.id,
		ClusterName: model.Cluster.ClusterName,
		Active:model.Cluster.Active,
	}*/
	return bean, nil
}

func (impl EnvironmentServiceImpl) getClusterConfig(cluster *ClusterBean) (*util.ClusterConfig, error) {
	host := cluster.ServerUrl
	configMap := cluster.Config
	bearerToken := configMap["bearer_token"]
	if cluster.Id == 1 && cluster.ClusterName == ClusterName {
		if _, err := os.Stat(TokenFilePath); os.IsNotExist(err) {
			impl.logger.Errorw("no directory or file exists", "TOKEN_FILE_PATH", TokenFilePath, "err", err)
			return nil, err
		} else {
			content, err := ioutil.ReadFile(TokenFilePath)
			if err != nil {
				impl.logger.Errorw("error on reading file", "err", err)
				return nil, err
			}
			bearerToken = string(content)
		}
	}
	clusterCfg := &util.ClusterConfig{Host: host, BearerToken: bearerToken}
	return clusterCfg, nil
}

func (impl EnvironmentServiceImpl) Update(mappings *EnvironmentBean, userId int32) (*EnvironmentBean, error) {
	model, err := impl.environmentRepository.FindById(mappings.Id)
	if err != nil {
		impl.logger.Errorw("error in finding environment for update", "err", err)
		return mappings, err
	}
	isNamespaceChange := false
	if model.Namespace != mappings.Namespace {
		isNamespaceChange = true
	}

	clusterBean, err := impl.clusterService.FindById(mappings.ClusterId)
	if err != nil {
		return nil, err
	}

	model.Name = mappings.Environment
	model.Active = mappings.Active
	model.Namespace = mappings.Namespace
	model.Default = mappings.Default
	model.UpdatedBy = userId
	model.UpdatedOn = time.Now()

	//namespace create if not exist
	if len(model.Namespace) > 0 {
		cfg, err := impl.getClusterConfig(clusterBean)
		if err != nil {
			return nil, err
		}
		if err := impl.K8sUtil.CreateNsIfNotExists(model.Namespace, cfg); err != nil {
			impl.logger.Errorw("error in creating ns", "ns", model.Namespace, "err", err)
		}
	}
	//namespace changed, update it on chart env override config as well
	if isNamespaceChange == true {
		impl.logger.Debug("namespace has modified in request, it will update related config")
		envPropertiesList, err := impl.propertiesConfigService.GetEnvironmentPropertiesById(mappings.Id)
		if err != nil {
			impl.logger.Error("failed to fetch chart environment override config", "err", err)
			//TODO - atomic operation breaks, throw internal error codes

		} else {
			for _, envProperties := range envPropertiesList {
				_, err := impl.propertiesConfigService.UpdateEnvironmentProperties(0, &pipeline.EnvironmentProperties{Id: envProperties.Id, Namespace: mappings.Namespace}, userId)
				if err != nil {
					impl.logger.Error("failed to update chart environment override config", "err", err)
					//TODO - atomic operation breaks, throw internal error codes
				}
			}
		}
	}
	grafanaDatasourceId := model.GrafanaDatasourceId
	//grafana datasource create if not exist
	if grafanaDatasourceId == 0 {
		grafanaDatasourceId, err = impl.clusterService.CreateGrafanaDataSource(clusterBean, model)
		if err != nil {
			impl.logger.Errorw("unable to create grafana data source for missing env", "env", model)
		}
	}
	model.GrafanaDatasourceId = grafanaDatasourceId
	err = impl.environmentRepository.Update(model)
	if err != nil {
		impl.logger.Errorw("error in updating environment", "err", err)
		return mappings, err
	}

	mappings.Id = model.Id
	return mappings, nil
}

func (impl EnvironmentServiceImpl) FindClusterByEnvId(id int) (*ClusterBean, error) {
	model, err := impl.environmentRepository.FindById(id)
	if err != nil {
		impl.logger.Errorw("fetch cluster by environment id", "err", err)
		return nil, err
	}

	clusterBean := &ClusterBean{
		Id:          model.Cluster.Id,
		ClusterName: model.Cluster.ClusterName,
		Active:      model.Cluster.Active,
		ServerUrl:   model.Cluster.ServerUrl,
		Config:      model.Cluster.Config,
	}
	return clusterBean, nil
}

func (impl EnvironmentServiceImpl) GetEnvironmentListForAutocomplete() ([]EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []EnvironmentBean
	for _, model := range models {
		beans = append(beans, EnvironmentBean{
			Id:          model.Id,
			Environment: model.Name,
			Namespace:   model.Namespace,
			CdArgoSetup: model.Cluster.CdArgoSetup,
		})
	}
	return beans, nil
}

func (impl EnvironmentServiceImpl) validateNamespaces(namespace string, envs []*cluster.Environment) error {
	if len(envs) >= 1 {
		if namespace == "" {
			impl.logger.Errorw("namespace cannot be empty")
			return errors.New("namespace cannot be empty")
		}
		for _, n := range envs {
			if n.Namespace == "" {
				impl.logger.Errorw("cannot create env as existing old envs have no namespace")
				return errors.New("cannot create env as existing old envs have no namespace")
			}
		}
	}
	return nil
}

func (impl EnvironmentServiceImpl) FindByIds(ids []*int) ([]*EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindByIds(ids)
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []*EnvironmentBean
	for _, model := range models {
		beans = append(beans, &EnvironmentBean{
			Id:          model.Id,
			Environment: model.Name,
			//ClusterId:          model.Cluster.Id,
			//ClusterName:        model.Cluster.ClusterName,
			Active: model.Active,
			//PrometheusEndpoint: model.Cluster.PrometheusEndpoint,
			Namespace: model.Namespace,
			Default:   model.Default,
			//CdArgoSetup:        model.Cluster.CdArgoSetup,
		})
	}
	return beans, nil
}

func (impl EnvironmentServiceImpl) FindByNamespaceAndClusterName(namespaces string, clusterName string) (*cluster.Environment, error) {
	env, err := impl.environmentRepository.FindByNamespaceAndClusterName(namespaces, clusterName)
	return env, err
}

func (impl EnvironmentServiceImpl) GetByClusterId(id int) ([]*EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindByClusterId(id)
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []*EnvironmentBean
	for _, model := range models {
		beans = append(beans, &EnvironmentBean{
			Id:          model.Id,
			Environment: model.Name,
			Namespace:   model.Namespace,
		})
	}
	return beans, nil
}
