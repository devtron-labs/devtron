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

package read

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/pkg/cluster/adapter"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/coreEntities/argoApplication/bean"
	"github.com/devtron-labs/devtron/pkg/coreEntities/argoApplication/helper"
	clientErrors "github.com/devtron-labs/devtron/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ArgoApplicationReadService interface {
	ValidateArgoResourceRequest(ctx context.Context, appIdentifier *bean.ArgoAppIdentifier, request *k8s.K8sRequestBean) (bool, error)
	GetAppDetailEA(ctx context.Context, resourceName, resourceNamespace string, clusterId int) (*bean.ArgoApplicationDetailDto, error)
	GetArgoManagedResources(resourceName, resourceNamespace string, clusterConfig *k8s.ClusterConfig) (*bean.ArgoManagedResourceResponse, error)
	GetArgoAppResourceTree(clusterConfig *k8s.ClusterConfig, targetClusterId int, resp *bean.ArgoManagedResourceResponse) (*gRPC.ResourceTreeResponse, error)
}

type ArgoApplicationReadServiceImpl struct {
	logger            *zap.SugaredLogger
	clusterRepository clusterRepository.ClusterRepository
	k8sUtil           *k8s.K8sServiceImpl
	helmAppClient     gRPC.HelmAppClient
	helmAppService    service.HelmAppService
}

func NewArgoApplicationReadServiceImpl(logger *zap.SugaredLogger,
	clusterRepository clusterRepository.ClusterRepository,
	k8sUtil *k8s.K8sServiceImpl,
	helmAppClient gRPC.HelmAppClient,
	helmAppService service.HelmAppService) *ArgoApplicationReadServiceImpl {
	return &ArgoApplicationReadServiceImpl{
		logger:            logger,
		clusterRepository: clusterRepository,
		k8sUtil:           k8sUtil,
		helmAppService:    helmAppService,
		helmAppClient:     helmAppClient,
	}

}

func (impl *ArgoApplicationReadServiceImpl) GetAppDetailEA(ctx context.Context, resourceName, resourceNamespace string, clusterId int) (*bean.ArgoApplicationDetailDto, error) {
	appDetail := &bean.ArgoApplicationDetailDto{
		ArgoApplicationListDto: &bean.ArgoApplicationListDto{
			Name:      resourceName,
			Namespace: resourceNamespace,
			ClusterId: clusterId,
		},
	}
	clusters, err := impl.clusterRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in getting all active clusters", "err", err)
		return nil, err
	}
	var clusterWithApplicationObject clusterRepository.Cluster
	clusterServerUrlIdMap := make(map[string]int, len(clusters))
	for _, cluster := range clusters {
		if cluster.Id == clusterId {
			clusterWithApplicationObject = cluster
		}
		clusterServerUrlIdMap[cluster.ServerUrl] = cluster.Id
	}
	if clusterWithApplicationObject.Id > 0 {
		appDetail.ClusterName = clusterWithApplicationObject.ClusterName
	}
	if clusterWithApplicationObject.IsVirtualCluster {
		return appDetail, nil
	} else if len(clusterWithApplicationObject.ErrorInConnecting) != 0 {
		return nil, fmt.Errorf("error in connecting to cluster")
	}
	clusterBean := adapter.GetClusterBean(clusterWithApplicationObject)
	clusterConfig := clusterBean.GetClusterConfig()
	resp, err := impl.GetArgoManagedResources(resourceName, resourceNamespace, clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting argo managed resources", "err", err)
		return nil, err
	}
	targetClusterId := 0
	if len(resp.DestinationServer) != 0 {
		if resp.DestinationServer == k8sCommonBean.DefaultClusterUrl {
			targetClusterId = clusterWithApplicationObject.Id
		} else if clusterIdFromMap, ok := clusterServerUrlIdMap[resp.DestinationServer]; ok {
			targetClusterId = clusterIdFromMap
		}
	}
	resourceTree, err := impl.GetArgoAppResourceTree(clusterConfig, targetClusterId, resp)
	if err != nil {
		impl.logger.Errorw("error in getting argo app resource tree", "err", err)
		return nil, err
	}
	appDetail.ResourceTree = resourceTree
	if resp.ManifestResponse != nil {
		appDetail.Manifest = resp.ManifestResponse.Manifest.Object
	}
	appDetail.HealthStatus = resp.HealthStatus
	appDetail.SyncStatus = resp.SyncStatus
	return appDetail, nil
}

func (impl *ArgoApplicationReadServiceImpl) GetArgoAppResourceTree(clusterConfig *k8s.ClusterConfig, targetClusterId int, resp *bean.ArgoManagedResourceResponse) (*gRPC.ResourceTreeResponse, error) {
	if resp.ManifestResponse == nil || resp.ManifestResponse.Manifest.Object == nil {
		return nil, fmt.Errorf("error in getting argo managed resources")
	}
	var targetClusterConfig bean.ArgoClusterConfigObj
	if targetClusterId < 1 {
		restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
		if err != nil {
			impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterHostUrl", clusterConfig.Host)
			return nil, err
		}
		// cluster is not added on devtron, need to get server config from secret which argo-cd saved
		coreV1Client, err := impl.k8sUtil.GetCoreV1ClientByRestConfig(restConfig)
		if err != nil {
			impl.logger.Errorw("error in getting core v1 client", "err", err, "clusterHostUrl", clusterConfig.Host)
			return nil, err
		}
		secrets, err := coreV1Client.Secrets(bean.AllNamespaces).List(context.Background(), v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labels.Set{"argocd.argoproj.io/secret-type": "cluster"}).String(),
		})
		if err != nil {
			impl.logger.Errorw("error in getting resource list, secrets", "err", err)
			return nil, err
		}
		for _, secret := range secrets.Items {
			if secret.Data != nil {
				if val, ok := secret.Data["server"]; ok {
					if string(val) == resp.DestinationServer {
						if config, ok := secret.Data["config"]; ok {
							err = json.Unmarshal(config, &targetClusterConfig)
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
	resourceTreeResp, err := impl.getResourceTreeForExternalCluster(targetClusterId, targetClusterConfig, resp.DestinationServer, resp.ArgoManagedResources)
	if err != nil {
		impl.logger.Errorw("error in getting resource tree response", "err", err)
		return nil, err
	}
	return resourceTreeResp, nil
}

func (impl *ArgoApplicationReadServiceImpl) GetArgoManagedResources(resourceName, resourceNamespace string, clusterConfig *k8s.ClusterConfig) (*bean.ArgoManagedResourceResponse, error) {
	restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", clusterConfig.ClusterId)
		return nil, err
	}
	resp, err := impl.k8sUtil.GetResource(context.Background(), resourceNamespace, resourceName, bean.GvkForArgoApplication, restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting resource list", "err", err)
		return nil, err
	}
	impl.logger.Infow("resp for argo external", "resp", resp)

	if resp != nil && resp.Manifest.Object != nil {
		healthStatus, syncStatus, destinationServer, argoManagedResources :=
			helper.GetHealthSyncStatusDestinationServerAndManagedResourcesForArgoK8sRawObject(resp.Manifest.Object)
		return &bean.ArgoManagedResourceResponse{
			ManifestResponse:     resp,
			HealthStatus:         healthStatus,
			SyncStatus:           syncStatus,
			DestinationServer:    destinationServer,
			ArgoManagedResources: argoManagedResources,
		}, nil
	} else {
		return &bean.ArgoManagedResourceResponse{}, nil
	}
}

func (impl *ArgoApplicationReadServiceImpl) getResourceTreeForExternalCluster(clusterId int, configOfClusterWhereAppIsDeployed bean.ArgoClusterConfigObj, destinationServer string, argoManagedResources []*bean.ArgoManagedResource) (*gRPC.ResourceTreeResponse, error) {
	var resources []*gRPC.ExternalResourceDetail
	for _, argoManagedResource := range argoManagedResources {
		resources = append(resources, &gRPC.ExternalResourceDetail{
			Group:     argoManagedResource.Group,
			Kind:      argoManagedResource.Kind,
			Version:   argoManagedResource.Version,
			Name:      argoManagedResource.Name,
			Namespace: argoManagedResource.Namespace,
		})
	}
	var clusterConfigOfClusterWhereAppIsDeployed *gRPC.ClusterConfig
	if len(configOfClusterWhereAppIsDeployed.BearerToken) > 0 {
		clusterConfigOfClusterWhereAppIsDeployed = &gRPC.ClusterConfig{
			ApiServerUrl:          destinationServer,
			Token:                 configOfClusterWhereAppIsDeployed.BearerToken,
			InsecureSkipTLSVerify: configOfClusterWhereAppIsDeployed.TlsClientConfig.Insecure,
			KeyData:               configOfClusterWhereAppIsDeployed.TlsClientConfig.KeyData,
			CaData:                configOfClusterWhereAppIsDeployed.TlsClientConfig.CaData,
			CertData:              configOfClusterWhereAppIsDeployed.TlsClientConfig.CertData,
		}
	}
	resourceTreeResp, err := impl.helmAppService.GetResourceTreeForExternalResources(context.Background(), clusterId, clusterConfigOfClusterWhereAppIsDeployed, resources)
	if err != nil {
		impl.logger.Errorw("error in getting resource tree for external resources", "err", err)
		return nil, err
	}
	return resourceTreeResp, nil
}

func (impl *ArgoApplicationReadServiceImpl) ValidateArgoResourceRequest(ctx context.Context, appIdentifier *bean.ArgoAppIdentifier, request *k8s.K8sRequestBean) (bool, error) {
	app, err := impl.GetAppDetailEA(ctx, appIdentifier.AppName, appIdentifier.Namespace, appIdentifier.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting app detail", "err", err, "appDetails", appIdentifier)
		apiError := clientErrors.ConvertToApiError(err)
		if apiError != nil {
			err = apiError
		}
		return false, err
	}

	valid := false

	for _, node := range app.ResourceTree.Nodes {
		nodeDetails := k8s.ResourceIdentifier{
			Name:      node.Name,
			Namespace: node.Namespace,
			GroupVersionKind: schema.GroupVersionKind{
				Group:   node.Group,
				Version: node.Version,
				Kind:    node.Kind,
			},
		}
		if nodeDetails == request.ResourceIdentifier {
			valid = true
			break
		}
	}
	appDetail := &gRPC.AppDetail{
		ResourceTreeResponse: app.ResourceTree,
	}
	if !valid {
		valid = validateContainerName(request, appDetail)
	}
	return valid, nil
}

func validateContainerName(request *k8s.K8sRequestBean, app *gRPC.AppDetail) bool {
	requestContainerName := request.PodLogsRequest.ContainerName
	podName := request.ResourceIdentifier.Name
	for _, pod := range app.ResourceTreeResponse.PodMetadata {
		if pod.Name == podName {

			// finding the container name in main Containers
			for _, container := range pod.Containers {
				if container == requestContainerName {
					return true
				}
			}

			// finding the container name in init containers
			for _, initContainer := range pod.InitContainers {
				if initContainer == requestContainerName {
					return true
				}
			}

			// finding the container name in ephemeral containers
			for _, ephemeralContainer := range pod.EphemeralContainers {
				if ephemeralContainer.Name == requestContainerName {
					return true
				}
			}

		}
	}
	return false
}
