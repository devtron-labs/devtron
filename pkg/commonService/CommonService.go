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

package commonService

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	bean2 "github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	"github.com/devtron-labs/devtron/pkg/attributes"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/team"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
)

type CommonService interface {
	FetchLatestChart(appId int, envId int) (*chartRepoRepository.Chart, error)
	GlobalChecklist() (*GlobalChecklist, error)
	GetServerConfigIfClusterIsNotAddedOnDevtron(resourceResp *k8s.ManifestResponse, restConfig *rest.Config,
		clusterWithApplicationObject repository3.Cluster, clusterServerUrlIdMap map[string]int) (*rest.Config, error)
	GetClusterConfigFromAllClusters(clusterId int) (*k8s.ClusterConfig, repository3.Cluster, map[string]int, error)
}

type CommonServiceImpl struct {
	logger                      *zap.SugaredLogger
	chartRepository             chartRepoRepository.ChartRepository
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository
	gitOpsRepository            repository.GitOpsConfigRepository
	dockerReg                   dockerRegistryRepository.DockerArtifactStoreRepository
	attributeRepo               repository.AttributesRepository
	gitProviderRepository       repository.GitProviderRepository
	environmentRepository       repository3.EnvironmentRepository
	teamRepository              repository2.TeamRepository
	appRepository               app.AppRepository
	k8sUtil                     *k8s.K8sUtil
	clusterRepository           repository3.ClusterRepository
}

func NewCommonServiceImpl(logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository,
	gitOpsRepository repository.GitOpsConfigRepository,
	dockerReg dockerRegistryRepository.DockerArtifactStoreRepository,
	attributeRepo repository.AttributesRepository,
	gitProviderRepository repository.GitProviderRepository,
	environmentRepository repository3.EnvironmentRepository, teamRepository repository2.TeamRepository,
	appRepository app.AppRepository,
	k8sUtil *k8s.K8sUtil,
	clusterRepository repository3.ClusterRepository) *CommonServiceImpl {
	serviceImpl := &CommonServiceImpl{
		logger:                      logger,
		chartRepository:             chartRepository,
		environmentConfigRepository: environmentConfigRepository,
		gitOpsRepository:            gitOpsRepository,
		dockerReg:                   dockerReg,
		attributeRepo:               attributeRepo,
		gitProviderRepository:       gitProviderRepository,
		environmentRepository:       environmentRepository,
		teamRepository:              teamRepository,
		appRepository:               appRepository,
		k8sUtil:                     k8sUtil,
		clusterRepository:           clusterRepository,
	}
	return serviceImpl
}

type GlobalChecklist struct {
	AppChecklist   *AppChecklist   `json:"appChecklist"`
	ChartChecklist *ChartChecklist `json:"chartChecklist"`
	IsAppCreated   bool            `json:"isAppCreated"`
	UserId         int32           `json:"-"`
}

type ChartChecklist struct {
	GitOps      int `json:"gitOps,omitempty"`
	Project     int `json:"project"`
	Environment int `json:"environment"`
}

type AppChecklist struct {
	GitOps      int `json:"gitOps,omitempty"`
	Project     int `json:"project"`
	Git         int `json:"git"`
	Environment int `json:"environment"`
	Docker      int `json:"docker"`
	HostUrl     int `json:"hostUrl"`
	//ChartChecklist *ChartChecklist `json:",inline"`
}

const (
	Server      = "server"
	Destination = "destination"
	Config      = "config"
)

func (impl *CommonServiceImpl) FetchLatestChart(appId int, envId int) (*chartRepoRepository.Chart, error) {
	var chart *chartRepoRepository.Chart
	if appId > 0 && envId > 0 {
		envOverride, err := impl.environmentConfigRepository.ActiveEnvConfigOverride(appId, envId)
		if err != nil {
			return nil, err
		}
		//if chart is overrides in env, and not mark as overrides in db, it means it was not completed and refer to latest to the app.
		if (envOverride.Id == 0) || (envOverride.Id > 0 && !envOverride.IsOverride) {
			chart, err = impl.chartRepository.FindLatestChartForAppByAppId(appId)
			if err != nil {
				return nil, err
			}
		} else {
			//if chart is overrides in env, it means it may have different version than app level.
			chart = envOverride.Chart
		}
	} else if appId > 0 {
		chartG, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
		if err != nil {
			return nil, err
		}
		chart = chartG

		//TODO - note if secret create/update from global with property (new style).
		// there may be older chart version in env overrides (and in that case it will be ignore, property and isBinary)
	}
	return chart, nil
}

func (impl *CommonServiceImpl) GlobalChecklist() (*GlobalChecklist, error) {

	dockerReg, err := impl.dockerReg.FindActiveDefaultStore()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	attribute, err := impl.attributeRepo.FindByKey(attributes.HostUrlKey)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	env, err := impl.environmentRepository.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	git, err := impl.gitProviderRepository.FindAllActiveForAutocomplete()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	project, err := impl.teamRepository.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	chartChecklist := &ChartChecklist{
		Project:     1,
		Environment: 1,
	}
	appChecklist := &AppChecklist{
		Project:     1,
		Git:         1,
		Environment: 1,
	}
	if len(env) > 0 {
		chartChecklist.Environment = 1
		appChecklist.Environment = 1
	}

	if len(git) > 0 {
		appChecklist.Git = 1
	}

	if len(project) > 0 {
		chartChecklist.Project = 1
		appChecklist.Project = 1
	}

	if len(dockerReg.Id) > 0 {
		appChecklist.Docker = 1
	}
	if attribute.Id > 0 {
		appChecklist.HostUrl = 1
	}
	config := &GlobalChecklist{
		AppChecklist:   appChecklist,
		ChartChecklist: chartChecklist,
	}

	apps, err := impl.appRepository.FindAllActiveAppsWithTeam()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}
	if len(apps) > 0 {
		config.IsAppCreated = true
	}
	return config, err
}

func (impl *CommonServiceImpl) GetServerConfigIfClusterIsNotAddedOnDevtron(resourceResp *k8s.ManifestResponse, restConfig *rest.Config,
	clusterWithApplicationObject repository3.Cluster, clusterServerUrlIdMap map[string]int) (*rest.Config, error) {
	var destinationServer string
	if resourceResp != nil && resourceResp.Manifest.Object != nil {
		_, _, destinationServer, _ =
			getHealthSyncStatusDestinationServerAndManagedResourcesForArgoK8sRawObject(resourceResp.Manifest.Object)
	}
	appDeployedOnClusterId := 0
	if destinationServer == k8s.DefaultClusterUrl {
		appDeployedOnClusterId = clusterWithApplicationObject.Id
	} else if clusterIdFromMap, ok := clusterServerUrlIdMap[destinationServer]; ok {
		appDeployedOnClusterId = clusterIdFromMap
	}
	var configOfClusterWhereAppIsDeployed bean2.ArgoClusterConfigObj
	if appDeployedOnClusterId < 1 {
		//cluster is not added on devtron, need to get server config from secret which argo-cd saved
		coreV1Client, err := impl.k8sUtil.GetCoreV1ClientByRestConfig(restConfig)
		secrets, err := coreV1Client.Secrets(bean2.AllNamespaces).List(context.Background(), metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labels.Set{"argocd.argoproj.io/secret-type": "cluster"}).String(),
		})
		if err != nil {
			impl.logger.Errorw("error in getting resource list, secrets", "err", err)
			return nil, err
		}
		for _, secret := range secrets.Items {
			if secret.Data != nil {
				if val, ok := secret.Data[Server]; ok {
					if string(val) == destinationServer {
						if config, ok := secret.Data[Config]; ok {
							err = json.Unmarshal(config, &configOfClusterWhereAppIsDeployed)
							if err != nil {
								impl.logger.Errorw("error in unmarshaling", "err", err)
								return nil, err
							}
							break
						}
					}
				}
			}
		}
	}
	restConfig.Host = destinationServer
	restConfig.TLSClientConfig.Insecure = configOfClusterWhereAppIsDeployed.TlsClientConfig.Insecure
	restConfig.BearerToken = configOfClusterWhereAppIsDeployed.BearerToken
	return restConfig, nil
}

func getHealthSyncStatusDestinationServerAndManagedResourcesForArgoK8sRawObject(obj map[string]interface{}) (string,
	string, string, []*bean2.ArgoManagedResource) {
	var healthStatus, syncStatus, destinationServer string
	argoManagedResources := make([]*bean2.ArgoManagedResource, 0)
	if specObjRaw, ok := obj[k8sCommonBean.Spec]; ok {
		specObj := specObjRaw.(map[string]interface{})
		if destinationObjRaw, ok2 := specObj[Destination]; ok2 {
			destinationObj := destinationObjRaw.(map[string]interface{})
			if destinationServerIf, ok3 := destinationObj[Server]; ok3 {
				destinationServer = destinationServerIf.(string)
			}
		}
	}
	if statusObjRaw, ok := obj[k8sCommonBean.K8sClusterResourceStatusKey]; ok {
		statusObj := statusObjRaw.(map[string]interface{})
		if healthObjRaw, ok2 := statusObj[k8sCommonBean.K8sClusterResourceHealthKey]; ok2 {
			healthObj := healthObjRaw.(map[string]interface{})
			if healthStatusIf, ok3 := healthObj[k8sCommonBean.K8sClusterResourceStatusKey]; ok3 {
				healthStatus = healthStatusIf.(string)
			}
		}
		if syncObjRaw, ok2 := statusObj[k8sCommonBean.K8sClusterResourceSyncKey]; ok2 {
			syncObj := syncObjRaw.(map[string]interface{})
			if syncStatusIf, ok3 := syncObj[k8sCommonBean.K8sClusterResourceStatusKey]; ok3 {
				syncStatus = syncStatusIf.(string)
			}
		}
		if resourceObjsRaw, ok2 := statusObj[k8sCommonBean.K8sClusterResourceResourcesKey]; ok2 {
			resourceObjs := resourceObjsRaw.([]interface{})
			argoManagedResources = make([]*bean2.ArgoManagedResource, 0, len(resourceObjs))
			for _, resourceObjRaw := range resourceObjs {
				argoManagedResource := &bean2.ArgoManagedResource{}
				resourceObj := resourceObjRaw.(map[string]interface{})
				if groupRaw, ok := resourceObj[k8sCommonBean.K8sClusterResourceGroupKey]; ok {
					argoManagedResource.Group = groupRaw.(string)
				}
				if kindRaw, ok := resourceObj[k8sCommonBean.K8sClusterResourceKindKey]; ok {
					argoManagedResource.Kind = kindRaw.(string)
				}
				if versionRaw, ok := resourceObj[k8sCommonBean.K8sClusterResourceVersionKey]; ok {
					argoManagedResource.Version = versionRaw.(string)
				}
				if nameRaw, ok := resourceObj[k8sCommonBean.K8sClusterResourceMetadataNameKey]; ok {
					argoManagedResource.Name = nameRaw.(string)
				}
				if namespaceRaw, ok := resourceObj[k8sCommonBean.K8sClusterResourceNamespaceKey]; ok {
					argoManagedResource.Namespace = namespaceRaw.(string)
				}
				argoManagedResources = append(argoManagedResources, argoManagedResource)
			}
		}
	}
	return healthStatus, syncStatus, destinationServer, argoManagedResources
}

func (impl *CommonServiceImpl) GetClusterConfigFromAllClusters(clusterId int) (*k8s.ClusterConfig, repository3.Cluster, map[string]int, error) {
	clusters, err := impl.clusterRepository.FindAllActive()
	var clusterWithApplicationObject repository3.Cluster
	if err != nil {
		impl.logger.Errorw("error in getting all active clusters", "err", err)
		return nil, clusterWithApplicationObject, nil, err
	}
	clusterServerUrlIdMap := make(map[string]int, len(clusters))
	for _, cluster := range clusters {
		if cluster.Id == clusterId {
			clusterWithApplicationObject = cluster
		}
		clusterServerUrlIdMap[cluster.ServerUrl] = cluster.Id
	}
	if len(clusterWithApplicationObject.ErrorInConnecting) != 0 {
		return nil, clusterWithApplicationObject, nil, fmt.Errorf("error in connecting to cluster")
	}
	clusterBean := cluster.GetClusterBean(clusterWithApplicationObject)
	clusterConfig, err := clusterBean.GetClusterConfig()
	return clusterConfig, clusterWithApplicationObject, clusterServerUrlIdMap, err
}
