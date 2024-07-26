/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
	"encoding/json"
	"fmt"
	bean3 "github.com/devtron-labs/devtron/pkg/attributes/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/adapter"
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/repository/bean"
	"strconv"
	"strings"
	"time"

	util2 "github.com/devtron-labs/common-lib/utils/k8s"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/k8s/informer"

	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type EnvironmentService interface {
	FindOne(environment string) (*bean2.EnvironmentBean, error)
	Create(mappings *bean2.EnvironmentBean, userId int32) (*bean2.EnvironmentBean, error)
	Update(mappings *bean2.EnvironmentBean, userId int32) (*bean2.EnvironmentBean, error)
	GetAll() ([]bean2.EnvironmentBean, error)
	GetAllActive() ([]bean2.EnvironmentBean, error)
	GetAllActiveEnvironmentCount() (int, error)
	Delete(deleteReq *bean2.EnvironmentBean, userId int32) error
	FindClusterByEnvId(id int) (*ClusterBean, error)
	// FindById provides an exposed struct of bean.EnvironmentBean;
	// Usage - can be used for API responses;
	FindById(id int) (*bean2.EnvironmentBean, error)
	// GetExtendedEnvBeanById is used internally to get combined env and cluster data in bean.EnvironmentBean;
	// Usage - used for intra-service communications only;
	// For exposed view use FindById instead
	GetExtendedEnvBeanById(id int) (*bean2.EnvironmentBean, error)
	GetEnvironmentListForAutocomplete(isDeploymentTypeParam bool) ([]bean2.EnvironmentBean, error)
	GetEnvironmentOnlyListForAutocomplete() ([]bean2.EnvironmentBean, error)
	FindByIds(ids []*int) ([]*bean2.EnvironmentBean, error)
	FindByNamespaceAndClusterName(namespaces string, clusterName string) (*repository.Environment, error)
	FindOneByNamespaceAndClusterId(namespaces string, clusterId int) (*bean2.EnvironmentBean, error)
	GetByClusterId(id int) ([]*bean2.EnvironmentBean, error)
	GetCombinedEnvironmentListForDropDown(token string, isActionUserSuperAdmin bool, auth func(email string, object []string) map[string]bool) ([]*bean2.ClusterEnvDto, error)
	GetCombinedEnvironmentListForDropDownByClusterIds(token string, clusterIds []int, auth func(token string, object string) bool) ([]*bean2.ClusterEnvDto, error)
	HandleErrorInClusterConnections(clusters []*ClusterBean, respMap map[int]error, clusterExistInDb bool)
	GetDetailsById(envId int) (*repository.Environment, error)
}

type EnvironmentServiceImpl struct {
	environmentRepository repository.EnvironmentRepository
	logger                *zap.SugaredLogger
	clusterService        ClusterService
	K8sUtil               *util2.K8sServiceImpl
	k8sInformerFactory    informer.K8sInformerFactory
	//propertiesConfigService pipeline.PropertiesConfigService
	userAuthService      user.UserAuthService
	attributesRepository repository2.AttributesRepository
}

func NewEnvironmentServiceImpl(environmentRepository repository.EnvironmentRepository,
	clusterService ClusterService, logger *zap.SugaredLogger,
	K8sUtil *util2.K8sServiceImpl, k8sInformerFactory informer.K8sInformerFactory,
	//  propertiesConfigService pipeline.PropertiesConfigService,
	userAuthService user.UserAuthService, attributesRepository repository2.AttributesRepository) *EnvironmentServiceImpl {
	return &EnvironmentServiceImpl{
		environmentRepository: environmentRepository,
		logger:                logger,
		clusterService:        clusterService,
		K8sUtil:               K8sUtil,
		k8sInformerFactory:    k8sInformerFactory,
		//propertiesConfigService: propertiesConfigService,
		userAuthService:      userAuthService,
		attributesRepository: attributesRepository,
	}
}

func (impl EnvironmentServiceImpl) Create(mappings *bean2.EnvironmentBean, userId int32) (*bean2.EnvironmentBean, error) {
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

	model, err := impl.environmentRepository.FindByEnvNameOrIdentifierOrNamespace(mappings.ClusterId, mappings.Environment, identifier, mappings.Namespace)
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
		Description:           mappings.Description,
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
		cfg := clusterBean.GetClusterConfig()
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
	mappings.ClusterServerUrl = clusterBean.ServerUrl
	return mappings, nil
}

func (impl EnvironmentServiceImpl) FindOne(environment string) (*bean2.EnvironmentBean, error) {
	model, err := impl.environmentRepository.FindOne(environment)
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
		return nil, err
	}
	bean := &bean2.EnvironmentBean{
		Id:                    model.Id,
		Environment:           model.Name,
		ClusterId:             model.Cluster.Id,
		Active:                model.Active,
		PrometheusEndpoint:    model.Cluster.PrometheusEndpoint,
		Namespace:             model.Namespace,
		Default:               model.Default,
		EnvironmentIdentifier: model.EnvironmentIdentifier,
		Description:           model.Description,
	}
	return bean, nil
}

func (impl EnvironmentServiceImpl) GetAll() ([]bean2.EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []bean2.EnvironmentBean
	for _, model := range models {
		beans = append(beans, bean2.EnvironmentBean{
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
			Description:           model.Description,
			IsVirtualEnvironment:  model.IsVirtualEnvironment,
		})
	}
	return beans, nil
}

func (impl EnvironmentServiceImpl) GetAllActive() ([]bean2.EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []bean2.EnvironmentBean
	for _, model := range models {
		beans = append(beans, bean2.EnvironmentBean{
			Id:                    model.Id,
			Environment:           model.Name,
			ClusterId:             model.Cluster.Id,
			ClusterName:           model.Cluster.ClusterName,
			Active:                model.Active,
			PrometheusEndpoint:    model.Cluster.PrometheusEndpoint,
			Namespace:             model.Namespace,
			Default:               model.Default,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
			Description:           model.Description,
			IsVirtualEnvironment:  model.IsVirtualEnvironment,
		})
	}
	return beans, nil
}
func (impl EnvironmentServiceImpl) GetAllActiveEnvironmentCount() (int, error) {
	cnt, err := impl.environmentRepository.FindAllActiveEnvironmentCount()
	return cnt, err
}
func (impl EnvironmentServiceImpl) FindById(id int) (*bean2.EnvironmentBean, error) {
	model, err := impl.environmentRepository.FindById(id)
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
		return nil, err
	}
	bean := &bean2.EnvironmentBean{
		Id:                    model.Id,
		Environment:           model.Name,
		ClusterId:             model.Cluster.Id,
		Active:                model.Active,
		PrometheusEndpoint:    model.Cluster.PrometheusEndpoint,
		Namespace:             model.Namespace,
		Default:               model.Default,
		EnvironmentIdentifier: model.EnvironmentIdentifier,
		Description:           model.Description,
		IsVirtualEnvironment:  model.IsVirtualEnvironment,
	}
	return bean, nil
}

func (impl EnvironmentServiceImpl) Update(mappings *bean2.EnvironmentBean, userId int32) (*bean2.EnvironmentBean, error) {
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
	model.Description = mappings.Description

	//namespace create if not exist
	if len(model.Namespace) > 0 {
		cfg := clusterBean.GetClusterConfig()
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

func (impl EnvironmentServiceImpl) GetExtendedEnvBeanById(id int) (*bean2.EnvironmentBean, error) {
	model, err := impl.environmentRepository.FindById(id)
	if err != nil {
		impl.logger.Errorw("error in fetch environment by id", "err", err)
		return nil, err
	}
	return adapter.NewEnvironmentBean(model), nil
}

func (impl EnvironmentServiceImpl) FindClusterByEnvId(id int) (*ClusterBean, error) {
	model, err := impl.environmentRepository.FindById(id)
	if err != nil {
		impl.logger.Errorw("fetch cluster by environment id", "err", err)
		return nil, err
	}
	clusterBean := &ClusterBean{}
	clusterBean.Id = model.Cluster.Id
	clusterBean.ClusterName = model.Cluster.ClusterName
	clusterBean.Active = model.Cluster.Active
	clusterBean.ServerUrl = model.Cluster.ServerUrl
	clusterBean.Config = model.Cluster.Config
	return clusterBean, nil
}

var permittedDeploymentConfigString = []string{bean2.PIPELINE_DEPLOYMENT_TYPE_HELM, bean2.PIPELINE_DEPLOYMENT_TYPE_ACD}

func (impl EnvironmentServiceImpl) GetEnvironmentListForAutocomplete(isDeploymentTypeParam bool) ([]bean2.EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []bean2.EnvironmentBean
	//Fetching deployment app type config values along with autocomplete api while creating CD pipeline
	if isDeploymentTypeParam {
		for _, model := range models {
			var (
				allowedDeploymentConfigString []string
			)
			deploymentConfig := make(map[string]map[string]bool)
			deploymentConfigEnvLevel := make(map[string]bool)
			deploymentConfigValues, _ := impl.attributesRepository.FindByKey(bean3.ENFORCE_DEPLOYMENT_TYPE_CONFIG)
			//if empty config received(doesn't exist in table) which can't be parsed
			if deploymentConfigValues.Value != "" {
				if err = json.Unmarshal([]byte(deploymentConfigValues.Value), &deploymentConfig); err != nil {
					return nil, err
				}
				deploymentConfigEnvLevel, _ = deploymentConfig[fmt.Sprintf("%d", model.Id)]
			}

			// if real config along with absurd values exist in table {"argo_cd": true, "helm": false, "absurd": false}",
			if ok, filteredDeploymentConfig := impl.IsReceivedDeploymentTypeValid(deploymentConfigEnvLevel); ok {
				allowedDeploymentConfigString = filteredDeploymentConfig
			} else {
				allowedDeploymentConfigString = permittedDeploymentConfigString
			}
			beans = append(beans, bean2.EnvironmentBean{
				Id:                     model.Id,
				Environment:            model.Name,
				Namespace:              model.Namespace,
				CdArgoSetup:            model.Cluster.CdArgoSetup,
				EnvironmentIdentifier:  model.EnvironmentIdentifier,
				ClusterName:            model.Cluster.ClusterName,
				Description:            model.Description,
				IsVirtualEnvironment:   model.IsVirtualEnvironment,
				AllowedDeploymentTypes: allowedDeploymentConfigString,
				ClusterId:              model.ClusterId,
			})
		}
	} else {
		for _, model := range models {
			beans = append(beans, bean2.EnvironmentBean{
				Id:                    model.Id,
				Environment:           model.Name,
				Namespace:             model.Namespace,
				CdArgoSetup:           model.Cluster.CdArgoSetup,
				EnvironmentIdentifier: model.EnvironmentIdentifier,
				ClusterName:           model.Cluster.ClusterName,
				Description:           model.Description,
				IsVirtualEnvironment:  model.IsVirtualEnvironment,
				ClusterId:             model.ClusterId,
			})
		}
	}
	return beans, nil
}

func (impl EnvironmentServiceImpl) IsReceivedDeploymentTypeValid(deploymentConfig map[string]bool) (bool, []string) {
	var (
		filteredDeploymentConfig []string
		flag                     bool
	)

	for key, value := range deploymentConfig {
		for _, permitted := range permittedDeploymentConfigString {
			if key == permitted {
				//filtering only those deployment app types which are in permitted zone and are marked true
				if value {
					flag = true
					filteredDeploymentConfig = append(filteredDeploymentConfig, key)
				}
				break
			}
		}
	}
	if !flag {
		return false, nil
	}
	return true, filteredDeploymentConfig
}

func (impl EnvironmentServiceImpl) GetEnvironmentOnlyListForAutocomplete() ([]bean2.EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindAllActiveEnvOnlyDetails()
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []bean2.EnvironmentBean
	for _, model := range models {
		beans = append(beans, bean2.EnvironmentBean{
			Id:                    model.Id,
			Environment:           model.Name,
			Namespace:             model.Namespace,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
			ClusterId:             model.ClusterId,
			IsVirtualEnvironment:  model.IsVirtualEnvironment,
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

func (impl EnvironmentServiceImpl) FindByIds(ids []*int) ([]*bean2.EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindByIds(ids)
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []*bean2.EnvironmentBean
	for _, model := range models {
		beans = append(beans, &bean2.EnvironmentBean{
			Id:                    model.Id,
			Environment:           model.Name,
			Active:                model.Active,
			Namespace:             model.Namespace,
			Default:               model.Default,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
			ClusterId:             model.ClusterId,
			Description:           model.Description,
			IsVirtualEnvironment:  model.IsVirtualEnvironment,
		})
	}
	return beans, nil
}

func (impl EnvironmentServiceImpl) FindByNamespaceAndClusterName(namespaces string, clusterName string) (*repository.Environment, error) {
	env, err := impl.environmentRepository.FindByNamespaceAndClusterName(namespaces, clusterName)
	return env, err
}

func (impl EnvironmentServiceImpl) FindOneByNamespaceAndClusterId(namespaces string, clusterId int) (*bean2.EnvironmentBean, error) {
	env, err := impl.environmentRepository.FindOneByNamespaceAndClusterId(namespaces, clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting environment by namespace and cluster id", "namespaces", namespaces, "clusterId", clusterId)
		return nil, err
	}
	return adapter.NewEnvironmentBean(env), err
}

func (impl EnvironmentServiceImpl) GetByClusterId(id int) ([]*bean2.EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindByClusterId(id)
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []*bean2.EnvironmentBean
	for _, model := range models {
		beans = append(beans, &bean2.EnvironmentBean{
			Id:                    model.Id,
			Environment:           model.Name,
			Namespace:             model.Namespace,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
			Description:           model.Description,
		})
	}
	return beans, nil
}

func (impl EnvironmentServiceImpl) GetCombinedEnvironmentListForDropDown(token string, isActionUserSuperAdmin bool, auth func(email string, object []string) map[string]bool) ([]*bean2.ClusterEnvDto, error) {
	var namespaceGroupByClusterResponse []*bean2.ClusterEnvDto
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
	rbacObject := make([]string, 0)
	for _, model := range models {
		rbacObject = append(rbacObject, model.EnvironmentIdentifier)
	}
	// auth enforcer applied here in batch
	rbacObjectResult := auth(token, rbacObject)

	uniqueComboMap := make(map[string]bool)
	grantedEnvironmentMap := make(map[string][]*bean2.EnvDto)
	for _, model := range models {
		// isActionUserSuperAdmin tell that user is super admin or not. auth check skip for admin
		if !isActionUserSuperAdmin {
			isValidAuth := rbacObjectResult[model.EnvironmentIdentifier]
			if !isValidAuth {
				impl.logger.Debugw("authentication for env failed", "object", model.EnvironmentIdentifier)
				continue
			}
		}
		key := fmt.Sprintf("%s__%s", model.Cluster.ClusterName, model.Namespace)
		groupKey := fmt.Sprintf("%s__%d", model.Cluster.ClusterName, model.ClusterId)
		uniqueComboMap[key] = true
		grantedEnvironmentMap[groupKey] = append(grantedEnvironmentMap[groupKey], &bean2.EnvDto{
			EnvironmentId:         model.Id,
			EnvironmentName:       model.Name,
			Namespace:             model.Namespace,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
			Description:           model.Description,
		})
	}

	namespaceListGroupByClusters := impl.k8sInformerFactory.GetLatestNamespaceListGroupByCLuster()
	rbacObject2 := make([]string, 0)
	for clusterName, namespaces := range namespaceListGroupByClusters {
		for namespace := range namespaces {
			environmentIdentifier := fmt.Sprintf("%s__%s", clusterName, namespace)
			rbacObject2 = append(rbacObject2, environmentIdentifier)
		}
	}
	// auth enforcer applied here in batch
	rbacObjectResult2 := auth(token, rbacObject)

	for clusterName, namespaces := range namespaceListGroupByClusters {
		clusterId := clusterMap[clusterName]
		for namespace := range namespaces {
			//deduplication for cluster and namespace combination
			key := fmt.Sprintf("%s__%s", clusterName, namespace)
			groupKey := fmt.Sprintf("%s__%d", clusterName, clusterId)
			if _, ok := uniqueComboMap[key]; !ok {
				environmentIdentifier := fmt.Sprintf("%s__%s", clusterName, namespace)
				// isActionUserSuperAdmin tell that user is super admin or not. auth check skip for admin
				if !isActionUserSuperAdmin {
					isValidAuth := rbacObjectResult2[environmentIdentifier]
					if !isValidAuth {
						impl.logger.Debugw("authentication for env failed", "object", environmentIdentifier)
						continue
					}
				}
				grantedEnvironmentMap[groupKey] = append(grantedEnvironmentMap[groupKey], &bean2.EnvDto{
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
		namespaceGroupByClusterResponse = append(namespaceGroupByClusterResponse, &bean2.ClusterEnvDto{
			ClusterName:  clusterInfo[0],
			ClusterId:    clusterId,
			Environments: v,
		})
	}
	return namespaceGroupByClusterResponse, nil
}

func (impl EnvironmentServiceImpl) GetCombinedEnvironmentListForDropDownByClusterIds(token string, clusterIds []int, auth func(token string, object string) bool) ([]*bean2.ClusterEnvDto, error) {
	var namespaceGroupByClusterResponse []*bean2.ClusterEnvDto
	clusterModels, err := impl.clusterService.FindByIds(clusterIds)
	if err != nil {
		impl.logger.Errorw("error in fetching clusters", "err", err)
		return namespaceGroupByClusterResponse, err
	}
	clusterMap := make(map[string]int)
	for _, item := range clusterModels {
		clusterMap[item.ClusterName] = item.Id
	}
	models, err := impl.environmentRepository.FindByClusterIds(clusterIds)
	if err != nil {
		impl.logger.Errorw("error in fetching environments", "err", err)
		return namespaceGroupByClusterResponse, err
	}
	uniqueComboMap := make(map[string]bool)
	grantedEnvironmentMap := make(map[string][]*bean2.EnvDto)
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
		grantedEnvironmentMap[groupKey] = append(grantedEnvironmentMap[groupKey], &bean2.EnvDto{
			EnvironmentId:         model.Id,
			EnvironmentName:       model.Name,
			Namespace:             model.Namespace,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
			Description:           model.Description,
		})
	}

	namespaceListGroupByClusters := impl.k8sInformerFactory.GetLatestNamespaceListGroupByCLuster()
	for clusterName, namespaces := range namespaceListGroupByClusters {
		clusterId := clusterMap[clusterName]
		if clusterId == 0 {
			continue
		}
		for namespace := range namespaces {
			//deduplication for cluster and namespace combination
			key := fmt.Sprintf("%s__%s", clusterName, namespace)
			groupKey := fmt.Sprintf("%s__%d", clusterName, clusterId)
			if _, ok := uniqueComboMap[key]; !ok {
				environmentIdentifier := fmt.Sprintf("%s__%s", clusterName, namespace)
				// auth enforcer applied here
				isValidAuth := auth(token, environmentIdentifier)
				if !isValidAuth {
					impl.logger.Debugw("authentication for env failed", "object", environmentIdentifier)
					continue
				}
				grantedEnvironmentMap[groupKey] = append(grantedEnvironmentMap[groupKey], &bean2.EnvDto{
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
		namespaceGroupByClusterResponse = append(namespaceGroupByClusterResponse, &bean2.ClusterEnvDto{
			ClusterName:  clusterInfo[0],
			ClusterId:    clusterId,
			Environments: v,
		})
	}
	return namespaceGroupByClusterResponse, nil
}

func (impl EnvironmentServiceImpl) Delete(deleteReq *bean2.EnvironmentBean, userId int32) error {
	existingEnv, err := impl.environmentRepository.FindById(deleteReq.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete.", "id", deleteReq.Id)
		return err
	}
	dbConnection := impl.environmentRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in establishing connection", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	deleteRequest := existingEnv
	deleteRequest.UpdatedOn = time.Now()
	deleteRequest.UpdatedBy = userId
	err = impl.environmentRepository.MarkEnvironmentDeleted(deleteRequest, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting environment", "envId", deleteReq.Id, "envName", deleteReq.Environment)
		return err
	}
	//deleting auth roles entries for this environment
	err = impl.userAuthService.DeleteRoles(bean.ENV_TYPE, deleteRequest.Name, tx, existingEnv.EnvironmentIdentifier, "")
	if err != nil {
		impl.logger.Errorw("error in deleting auth roles", "err", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (impl EnvironmentServiceImpl) HandleErrorInClusterConnections(clusters []*ClusterBean, respMap map[int]error, clusterExistInDb bool) {
	impl.clusterService.HandleErrorInClusterConnections(clusters, respMap, clusterExistInDb)
}

func (impl EnvironmentServiceImpl) GetDetailsById(envId int) (*repository.Environment, error) {
	envDetails, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.logger.Errorw("error encountered in GetDetailsById", "envId", envId, "err", err)
		return nil, err
	}
	return envDetails, nil
}
