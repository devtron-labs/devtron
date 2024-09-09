package read

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	clientErrors "github.com/devtron-labs/devtron/pkg/errors"
	"github.com/devtron-labs/devtron/util/argo"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type ArgoApplicationReadService interface {
	GetRestConfigForExternalArgo(ctx context.Context, clusterId int, externalArgoApplicationName string) (*rest.Config, error)
	GetClusterConfigFromAllClusters(clusterId int) (*k8s.ClusterConfig, clusterRepository.Cluster, map[string]int, error)
	ValidateArgoResourceRequest(ctx context.Context, appIdentifier *bean.ArgoAppIdentifier, request *k8s.K8sRequestBean) (bool, error)
	GetAppDetail(resourceName, resourceNamespace string, clusterId int) (*bean.ArgoApplicationDetailDto, error)
}

type ArgoApplicationReadServiceImpl struct {
	logger            *zap.SugaredLogger
	clusterRepository clusterRepository.ClusterRepository
	k8sUtil           *k8s.K8sServiceImpl
	argoUserService   argo.ArgoUserService
	helmAppClient     gRPC.HelmAppClient
	helmAppService    service.HelmAppService
}

func NewArgoApplicationReadServiceImpl(logger *zap.SugaredLogger,
	clusterRepository clusterRepository.ClusterRepository,
	k8sUtil *k8s.K8sServiceImpl,
	argoUserService argo.ArgoUserService, helmAppClient gRPC.HelmAppClient,
	helmAppService service.HelmAppService) *ArgoApplicationReadServiceImpl {
	return &ArgoApplicationReadServiceImpl{
		logger:            logger,
		clusterRepository: clusterRepository,
		k8sUtil:           k8sUtil,
		argoUserService:   argoUserService,
		helmAppService:    helmAppService,
		helmAppClient:     helmAppClient,
	}

}

func (impl *ArgoApplicationReadServiceImpl) GetRestConfigForExternalArgo(ctx context.Context, clusterId int, externalArgoApplicationName string) (*rest.Config, error) {
	clusterConfig, clusterWithApplicationObject, clusterServerUrlIdMap, err := impl.GetClusterConfigFromAllClusters(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster config", "err", err, "clusterId", clusterId)
		return nil, err
	}
	restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config", "err", err, "clusterId", clusterId)
		return nil, err
	}
	resourceResp, err := impl.k8sUtil.GetResource(ctx, bean.DevtronCDNamespae, externalArgoApplicationName, bean.GvkForArgoApplication, restConfig)
	if err != nil {
		impl.logger.Errorw("not on external cluster", "err", err, "externalArgoApplicationName", externalArgoApplicationName)
		return nil, err
	}
	restConfig, err = impl.GetServerConfigIfClusterIsNotAddedOnDevtron(resourceResp, restConfig, clusterWithApplicationObject, clusterServerUrlIdMap)
	if err != nil {
		impl.logger.Errorw("error in getting server config", "err", err, "cluster with application object", clusterWithApplicationObject)
		return nil, err
	}
	return restConfig, nil
}

func (impl *ArgoApplicationReadServiceImpl) GetServerConfigIfClusterIsNotAddedOnDevtron(resourceResp *k8s.ManifestResponse, restConfig *rest.Config,
	clusterWithApplicationObject clusterRepository.Cluster, clusterServerUrlIdMap map[string]int) (*rest.Config, error) {
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
	var configOfClusterWhereAppIsDeployed *bean.ArgoClusterConfigObj
	if appDeployedOnClusterId < 1 {
		// cluster is not added on devtron, need to get server config from secret which argo-cd saved
		coreV1Client, err := impl.k8sUtil.GetCoreV1ClientByRestConfig(restConfig)
		secrets, err := coreV1Client.Secrets(bean.AllNamespaces).List(context.Background(), v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labels.Set{"argocd.argoproj.io/secret-type": "cluster"}).String(),
		})
		if err != nil {
			impl.logger.Errorw("error in getting resource list, secrets", "err", err)
			return nil, err
		}
		for _, secret := range secrets.Items {
			if secret.Data != nil {
				if val, ok := secret.Data[bean.Server]; ok {
					if string(val) == destinationServer {
						if config, ok := secret.Data[bean.Config]; ok {
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
		if configOfClusterWhereAppIsDeployed != nil {
			restConfig, err = impl.k8sUtil.GetRestConfigByCluster(&k8s.ClusterConfig{
				Host:                  destinationServer,
				BearerToken:           configOfClusterWhereAppIsDeployed.BearerToken,
				InsecureSkipTLSVerify: configOfClusterWhereAppIsDeployed.TlsClientConfig.Insecure,
				KeyData:               configOfClusterWhereAppIsDeployed.TlsClientConfig.KeyData,
				CAData:                configOfClusterWhereAppIsDeployed.TlsClientConfig.CaData,
				CertData:              configOfClusterWhereAppIsDeployed.TlsClientConfig.CertData,
			})
			if err != nil {
				impl.logger.Errorw("error in GetRestConfigByCluster, GetServerConfigIfClusterIsNotAddedOnDevtron", "err", err, "serverUrl", destinationServer)
				return nil, err
			}
		}
	}
	return restConfig, nil
}

func (impl *ArgoApplicationReadServiceImpl) GetClusterConfigFromAllClusters(clusterId int) (*k8s.ClusterConfig, clusterRepository.Cluster, map[string]int, error) {
	clusters, err := impl.clusterRepository.FindAllActive()
	var clusterWithApplicationObject clusterRepository.Cluster
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
	clusterBean := cluster2.GetClusterBean(clusterWithApplicationObject)
	clusterConfig := clusterBean.GetClusterConfig()
	return clusterConfig, clusterWithApplicationObject, clusterServerUrlIdMap, err
}

func (impl *ArgoApplicationReadServiceImpl) GetAppDetail(resourceName, resourceNamespace string, clusterId int) (*bean.ArgoApplicationDetailDto, error) {
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
	clusterBean := cluster2.GetClusterBean(clusterWithApplicationObject)
	clusterConfig := clusterBean.GetClusterConfig()
	restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", clusterWithApplicationObject.Id)
		return nil, err
	}
	resp, err := impl.k8sUtil.GetResource(context.Background(), resourceNamespace, resourceName, bean.GvkForArgoApplication, restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting resource list", "err", err)
		return nil, err
	}
	var destinationServer string
	var argoManagedResources []*bean.ArgoManagedResource
	if resp != nil && resp.Manifest.Object != nil {
		appDetail.Manifest = resp.Manifest.Object
		appDetail.HealthStatus, appDetail.SyncStatus, destinationServer, argoManagedResources =
			getHealthSyncStatusDestinationServerAndManagedResourcesForArgoK8sRawObject(resp.Manifest.Object)
	}
	appDeployedOnClusterId := 0
	if destinationServer == k8s.DefaultClusterUrl {
		appDeployedOnClusterId = clusterWithApplicationObject.Id
	} else if clusterIdFromMap, ok := clusterServerUrlIdMap[destinationServer]; ok {
		appDeployedOnClusterId = clusterIdFromMap
	}
	var configOfClusterWhereAppIsDeployed bean.ArgoClusterConfigObj
	if appDeployedOnClusterId < 1 {
		// cluster is not added on devtron, need to get server config from secret which argo-cd saved
		coreV1Client, err := impl.k8sUtil.GetCoreV1ClientByRestConfig(restConfig)
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
					if string(val) == destinationServer {
						if config, ok := secret.Data["config"]; ok {
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
	resourceTreeResp, err := impl.getResourceTreeForExternalCluster(appDeployedOnClusterId, destinationServer, configOfClusterWhereAppIsDeployed, argoManagedResources)
	if err != nil {
		impl.logger.Errorw("error in getting resource tree response", "err", err)
		return nil, err
	}
	appDetail.ResourceTree = resourceTreeResp
	return appDetail, nil
}

func (impl *ArgoApplicationReadServiceImpl) getResourceTreeForExternalCluster(clusterId int, destinationServer string,
	configOfClusterWhereAppIsDeployed bean.ArgoClusterConfigObj, argoManagedResources []*bean.ArgoManagedResource) (*gRPC.ResourceTreeResponse, error) {
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
	app, err := impl.GetAppDetail(appIdentifier.AppName, appIdentifier.Namespace, appIdentifier.ClusterId)
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
	return validateContainerNameIfReqd(valid, request, appDetail), nil
}

func validateContainerNameIfReqd(valid bool, request *k8s.K8sRequestBean, app *gRPC.AppDetail) bool {
	if !valid {
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
	}
	return valid
}

func getHealthSyncStatusDestinationServerAndManagedResourcesForArgoK8sRawObject(obj map[string]interface{}) (string,
	string, string, []*bean.ArgoManagedResource) {
	var healthStatus, syncStatus, destinationServer string
	argoManagedResources := make([]*bean.ArgoManagedResource, 0)
	if specObjRaw, ok := obj[k8sCommonBean.Spec]; ok {
		specObj := specObjRaw.(map[string]interface{})
		if destinationObjRaw, ok2 := specObj[bean.Destination]; ok2 {
			destinationObj := destinationObjRaw.(map[string]interface{})
			if destinationServerIf, ok3 := destinationObj[bean.Server]; ok3 {
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
			argoManagedResources = make([]*bean.ArgoManagedResource, 0, len(resourceObjs))
			for _, resourceObjRaw := range resourceObjs {
				argoManagedResource := &bean.ArgoManagedResource{}
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
