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
	"github.com/devtron-labs/devtron/client/k8s/informer"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type EnvironmentBean struct {
	Id                    int    `json:"id,omitempty" validate:"number"`
	Environment           string `json:"environment_name,omitempty" validate:"required,max=50"`
	ClusterId             int    `json:"cluster_id,omitempty" validate:"number,required"`
	ClusterName           string `json:"cluster_name,omitempty"`
	Active                bool   `json:"active"`
	Default               bool   `json:"default"`
	PrometheusEndpoint    string `json:"prometheus_endpoint,omitempty"`
	Namespace             string `json:"namespace,omitempty" validate:"max=50"`
	CdArgoSetup           bool   `json:"isClusterCdActive"`
	EnvironmentIdentifier string `json:"environmentIdentifier"`
}

type EnvDto struct {
	EnvironmentId         int    `json:"environmentId" validate:"number"`
	EnvironmentName       string `json:"environmentName,omitempty" validate:"max=50"`
	Namespace             string `json:"namespace,omitempty" validate:"max=50"`
	EnvironmentIdentifier string `json:"environmentIdentifier,omitempty"`
}

type ClusterEnvDto struct {
	ClusterId    int       `json:"clusterId"`
	ClusterName  string    `json:"clusterName,omitempty"`
	Environments []*EnvDto `json:"environments,omitempty"`
}

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
	FindByNamespaceAndClusterName(namespaces string, clusterName string) (*repository.Environment, error)
	GetByClusterId(id int) ([]*EnvironmentBean, error)
	GetCombinedEnvironmentListForDropDown(token string, auth func(token string, object string) bool) ([]*ClusterEnvDto, error)
}

type EnvironmentServiceImpl struct {
	environmentRepository repository.EnvironmentRepository
	logger                *zap.SugaredLogger
	clusterService        ClusterService
	K8sUtil               *util.K8sUtil
	k8sInformerFactory    informer.K8sInformerFactory
	//propertiesConfigService pipeline.PropertiesConfigService
}

func NewEnvironmentServiceImpl(environmentRepository repository.EnvironmentRepository,
	clusterService ClusterService, logger *zap.SugaredLogger,
	K8sUtil *util.K8sUtil, k8sInformerFactory informer.K8sInformerFactory,
//	propertiesConfigService pipeline.PropertiesConfigService,
) *EnvironmentServiceImpl {
	return &EnvironmentServiceImpl{
		environmentRepository: environmentRepository,
		logger:                logger,
		clusterService:        clusterService,
		K8sUtil:               K8sUtil,
		k8sInformerFactory:    k8sInformerFactory,
		//propertiesConfigService: propertiesConfigService,
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

	identifier := clusterBean.ClusterName + "__" + mappings.Namespace

	model, err := impl.environmentRepository.FindByNameOrIdentifier(mappings.Environment, identifier)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in finding environment for update", "err", err)
		return mappings, err
	}
	if model.Id > 0 {
		impl.logger.Warnw("environment already exists for this cluster and namespace", "model", model)
		return mappings, fmt.Errorf("environment already exists")
	}

	model = &repository.Environment{
		Name:                  mappings.Environment,
		ClusterId:             mappings.ClusterId,
		Active:                mappings.Active,
		Namespace:             mappings.Namespace,
		Default:               mappings.Default,
		EnvironmentIdentifier: identifier,
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
		cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
		if err != nil {
			return nil, err
		}
		if err := impl.K8sUtil.CreateNsIfNotExists(model.Namespace, cfg); err != nil {
			impl.logger.Errorw("error in creating ns", "ns", model.Namespace, "err", err)
		}

	}

	//ignore grafana if no prometheus url found
	if len(clusterBean.PrometheusUrl) > 0 {
		_, err = impl.clusterService.CreateGrafanaDataSource(clusterBean, model)
		if err != nil {
			impl.logger.Errorw("unable to create grafana data source", "env", model)
		}
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
		Id:                    model.Id,
		Environment:           model.Name,
		ClusterId:             model.Cluster.Id,
		Active:                model.Active,
		PrometheusEndpoint:    model.Cluster.PrometheusEndpoint,
		Namespace:             model.Namespace,
		Default:               model.Default,
		EnvironmentIdentifier: model.EnvironmentIdentifier,
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
			Id:                    model.Id,
			Environment:           model.Name,
			ClusterId:             model.Cluster.Id,
			ClusterName:           model.Cluster.ClusterName,
			Active:                model.Active,
			PrometheusEndpoint:    model.Cluster.PrometheusEndpoint,
			Namespace:             model.Namespace,
			Default:               model.Default,
			CdArgoSetup:           model.Cluster.CdArgoSetup,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
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
			Id:                    model.Id,
			Environment:           model.Name,
			ClusterId:             model.Cluster.Id,
			ClusterName:           model.Cluster.ClusterName,
			Active:                model.Active,
			PrometheusEndpoint:    model.Cluster.PrometheusEndpoint,
			Namespace:             model.Namespace,
			Default:               model.Default,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
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
		Id:                    model.Id,
		Environment:           model.Name,
		ClusterId:             model.Cluster.Id,
		Active:                model.Active,
		PrometheusEndpoint:    model.Cluster.PrometheusEndpoint,
		Namespace:             model.Namespace,
		Default:               model.Default,
		EnvironmentIdentifier: model.EnvironmentIdentifier,
	}

	/*clusterBean := &ClusterBean{
		id:model.Cluster.id,
		ClusterName: model.Cluster.ClusterName,
		Active:model.Cluster.Active,
	}*/
	return bean, nil
}

func (impl EnvironmentServiceImpl) Update(mappings *EnvironmentBean, userId int32) (*EnvironmentBean, error) {
	model, err := impl.environmentRepository.FindById(mappings.Id)
	if err != nil {
		impl.logger.Errorw("error in finding environment for update", "err", err)
		return mappings, err
	}
	/*isNamespaceChange := false
	if model.Namespace != mappings.Namespace {
		isNamespaceChange = true
	}*/

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
		cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
		if err != nil {
			return nil, err
		}
		if err := impl.K8sUtil.CreateNsIfNotExists(model.Namespace, cfg); err != nil {
			impl.logger.Errorw("error in creating ns", "ns", model.Namespace, "err", err)
		}
	}
	//namespace changed, update it on chart env override config as well
	/*if isNamespaceChange == true {
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
	}*/
	grafanaDatasourceId := model.GrafanaDatasourceId
	//grafana datasource create if not exist
	if len(clusterBean.PrometheusUrl) > 0 && grafanaDatasourceId == 0 {
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
			Id:                    model.Id,
			Environment:           model.Name,
			Namespace:             model.Namespace,
			CdArgoSetup:           model.Cluster.CdArgoSetup,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
		})
	}
	return beans, nil
}

func (impl EnvironmentServiceImpl) validateNamespaces(namespace string, envs []*repository.Environment) error {
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
			Id:                    model.Id,
			Environment:           model.Name,
			Active:                model.Active,
			Namespace:             model.Namespace,
			Default:               model.Default,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
		})
	}
	return beans, nil
}

func (impl EnvironmentServiceImpl) FindByNamespaceAndClusterName(namespaces string, clusterName string) (*repository.Environment, error) {
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
			Id:                    model.Id,
			Environment:           model.Name,
			Namespace:             model.Namespace,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
		})
	}
	return beans, nil
}

func (impl EnvironmentServiceImpl) GetCombinedEnvironmentListForDropDown(token string, auth func(token string, object string) bool) ([]*ClusterEnvDto, error) {
	var namespaceGroupByClusterResponse []*ClusterEnvDto
	clusterModels, err := impl.clusterService.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching clusters", "err", err)
		return namespaceGroupByClusterResponse, err
	}
	clusterMap := make(map[string]int)
	for _, item := range clusterModels {
		clusterMap[item.ClusterName] = item.Id
	}
	models, err := impl.environmentRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching environments", "err", err)
		return namespaceGroupByClusterResponse, err
	}
	uniqueComboMap := make(map[string]bool)
	grantedEnvironmentMap := make(map[string][]*EnvDto)
	for _, model := range models {
		// auth enforcer applied here
		isValidAuth := auth(token, model.EnvironmentIdentifier)
		if !isValidAuth {
			impl.logger.Debugw("authentication for env failed", "object", model.EnvironmentIdentifier)
			continue
		}
		key := fmt.Sprintf("%s__%s", model.Cluster.ClusterName, model.Namespace)
		groupKey := fmt.Sprintf("%s__%d", model.Cluster.ClusterName, model.ClusterId)
		uniqueComboMap[key] = true
		grantedEnvironmentMap[groupKey] = append(grantedEnvironmentMap[groupKey], &EnvDto{
			EnvironmentId:         model.Id,
			EnvironmentName:       model.Name,
			Namespace:             model.Namespace,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
		})
	}

	namespaceListGroupByClusters := impl.k8sInformerFactory.GetLatestNamespaceListGroupByCLuster()
	for clusterName, namespaces := range namespaceListGroupByClusters {
		clusterId := clusterMap[clusterName]
		for namespace := range namespaces {
			environmentIdentifier := fmt.Sprintf("%s__%s", clusterName, namespace)
			// auth enforcer applied here
			isValidAuth := auth(token, environmentIdentifier)
			if !isValidAuth {
				impl.logger.Debugw("authentication for env failed", "object", environmentIdentifier)
				continue
			}
			//deduplication for cluster and namespace combination
			groupKey := fmt.Sprintf("%s__%d", clusterName, clusterId)
			if _, ok := uniqueComboMap[groupKey]; !ok {
				grantedEnvironmentMap[groupKey] = append(grantedEnvironmentMap[groupKey], &EnvDto{
					EnvironmentName:       environmentIdentifier,
					Namespace:             namespace,
					EnvironmentIdentifier: environmentIdentifier,
				})
			}
		}
	}

	//final result builds here, namespace group by clusters
	for k, v := range grantedEnvironmentMap {
		clusterInfo := strings.Split(k, "__")
		clusterId, err := strconv.Atoi(clusterInfo[1])
		if err != nil {
			clusterId = 0
		}
		namespaceGroupByClusterResponse = append(namespaceGroupByClusterResponse, &ClusterEnvDto{
			ClusterName:  clusterInfo[0],
			ClusterId:    clusterId,
			Environments: v,
		})
	}
	return namespaceGroupByClusterResponse, nil
}

/*
deprecated
*/
func (impl EnvironmentServiceImpl) getAllClusterNamespaceCombination() ([]*EnvironmentBean, error) {
	var beans []*EnvironmentBean
	models, err := impl.clusterService.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching clusters", "err", err)
	}

	for _, clusterBean := range models {
		var client *v1.CoreV1Client
		if clusterBean.ClusterName == DefaultClusterName {
			client, err = impl.K8sUtil.GetClientForInCluster()
			if err != nil {
				continue
				//return nil, err
			}
		} else {
			client, err = impl.K8sUtil.GetClientByToken(clusterBean.ServerUrl, clusterBean.Config)
			if err != nil {
				continue
				//return nil, err
			}
		}
		namespaceList, err := impl.K8sUtil.ListNamespaces(client)
		statusError, _ := err.(*errors2.StatusError)
		if err != nil && statusError != nil && statusError.Status().Code != http.StatusNotFound {
			impl.logger.Errorw("namespaces not found", "err", err)
			continue
			//return nil, err
		}

		for _, namespace := range namespaceList.Items {
			beans = append(beans, &EnvironmentBean{
				Environment:           fmt.Sprintf("%s__%s", clusterBean.ClusterName, namespace.ObjectMeta.Name),
				Namespace:             namespace.ObjectMeta.Name,
				ClusterName:           clusterBean.ClusterName,
				ClusterId:             clusterBean.Id,
				EnvironmentIdentifier: fmt.Sprintf("%s__%s", clusterBean.ClusterName, namespace.ObjectMeta.Name),
			})
		}
	}

	return beans, nil
}
