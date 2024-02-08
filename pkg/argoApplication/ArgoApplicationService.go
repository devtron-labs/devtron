package argoApplication

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/api/helm-app/service"

	"github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/util/argo"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
)

type ArgoApplicationService interface {
	ListApplications(clusterIds []int) ([]*bean.ArgoApplicationListDto, error)
	GetAppDetail(resourceName, resourceNamespace string, clusterId int) (*bean.ArgoApplicationDetailDto, error)
	GetServerConfigIfClusterIsNotAddedOnDevtron(resourceResp *k8s.ManifestResponse, restConfig *rest.Config,
		clusterWithApplicationObject clusterRepository.Cluster, clusterServerUrlIdMap map[string]int) (*rest.Config, error)
	GetClusterConfigFromAllClusters(clusterId int) (*k8s.ClusterConfig, clusterRepository.Cluster, map[string]int, error)
	GetRestConfigForExternalArgo(ctx context.Context, clusterId int, externalArgoApplicationName string) (*rest.Config, error)
}

type ArgoApplicationServiceImpl struct {
	logger            *zap.SugaredLogger
	clusterRepository clusterRepository.ClusterRepository
	k8sUtil           *k8s.K8sServiceImpl
	argoUserService   argo.ArgoUserService
	helmAppService    service.HelmAppService
}

func NewArgoApplicationServiceImpl(logger *zap.SugaredLogger,
	clusterRepository clusterRepository.ClusterRepository,
	k8sUtil *k8s.K8sServiceImpl,
	argoUserService argo.ArgoUserService,
	helmAppService service.HelmAppService) *ArgoApplicationServiceImpl {
	return &ArgoApplicationServiceImpl{
		logger:            logger,
		clusterRepository: clusterRepository,
		k8sUtil:           k8sUtil,
		argoUserService:   argoUserService,
		helmAppService:    helmAppService,
	}

}

func (impl *ArgoApplicationServiceImpl) ListApplications(clusterIds []int) ([]*bean.ArgoApplicationListDto, error) {
	var clusters []clusterRepository.Cluster
	var err error
	if len(clusterIds) > 0 {
		//getting cluster details by ids
		clusters, err = impl.clusterRepository.FindByIds(clusterIds)
		if err != nil {
			impl.logger.Errorw("error in getting clusters by ids", "err", err, "clusterIds", clusterIds)
			return nil, err
		}
	} else {
		clusters, err = impl.clusterRepository.FindAllActive()
		if err != nil {
			impl.logger.Errorw("error in getting all active clusters", "err", err)
			return nil, err
		}
	}

	//TODO: make goroutine and channel for optimization
	appListFinal := make([]*bean.ArgoApplicationListDto, 0)
	for _, cluster := range clusters {
		clusterObj := cluster
		if clusterObj.IsVirtualCluster || len(clusterObj.ErrorInConnecting) != 0 {
			continue
		}
		clusterBean := cluster2.GetClusterBean(clusterObj)
		clusterConfig, err := clusterBean.GetClusterConfig()
		restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
		if err != nil {
			impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", clusterObj.Id)
			return nil, err
		}
		resp, _, err := impl.k8sUtil.GetResourceList(context.Background(), restConfig, bean.GvkForArgoApplication, bean.AllNamespaces)
		if err != nil {
			if errStatus, ok := err.(*errors.StatusError); ok {
				if errStatus.Status().Code == 404 {
					//no argo apps found, not sending error
					impl.logger.Warnw("error in getting external argo app list, no apps found", "err", err, "clusterId", clusterObj.Id)
					continue
				}
			}
			impl.logger.Errorw("error in getting resource list", "err", err)
			return nil, err
		}
		appLists := getApplicationListDtos(resp.Resources.Object, clusterObj.ClusterName, clusterObj.Id)
		appListFinal = append(appListFinal, appLists...)
	}
	return appListFinal, nil
}

func (impl *ArgoApplicationServiceImpl) GetAppDetail(resourceName, resourceNamespace string, clusterId int) (*bean.ArgoApplicationDetailDto, error) {
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
	clusterConfig, err := clusterBean.GetClusterConfig()
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
		//cluster is not added on devtron, need to get server config from secret which argo-cd saved
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

func (impl *ArgoApplicationServiceImpl) getResourceTreeForExternalCluster(clusterId int, destinationServer string,
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

func getApplicationListDtos(manifestObj map[string]interface{}, clusterName string, clusterId int) []*bean.ArgoApplicationListDto {
	appLists := make([]*bean.ArgoApplicationListDto, 0)
	//map of keys and index in row cells, initially set as 0 will be updated by object
	keysToBeFetchedFromColumnDefinitions := map[string]int{k8sCommonBean.K8sResourceColumnDefinitionName: 0,
		k8sCommonBean.K8sResourceColumnDefinitionHealthStatus: 0, k8sCommonBean.K8sResourceColumnDefinitionSyncStatus: 0}
	keysToBeFetchedFromRawObject := []string{k8sCommonBean.K8sClusterResourceNamespaceKey}

	columnsDataRaw := manifestObj[k8sCommonBean.K8sClusterResourceColumnDefinitionKey]
	if columnsDataRaw != nil {
		columnsData := columnsDataRaw.([]interface{})
		for i, columnData := range columnsData {
			columnDataMap := columnData.(map[string]interface{})
			for key := range keysToBeFetchedFromColumnDefinitions {
				if columnDataMap[k8sCommonBean.K8sClusterResourceNameKey] == key {
					keysToBeFetchedFromColumnDefinitions[key] = i
				}
			}
		}
	}

	rowsDataRaw := manifestObj[k8sCommonBean.K8sClusterResourceRowsKey]
	if rowsDataRaw != nil {
		rowsData := rowsDataRaw.([]interface{})
		for _, rowData := range rowsData {
			appListDto := &bean.ArgoApplicationListDto{
				ClusterId:   clusterId,
				ClusterName: clusterName,
			}
			rowDataMap := rowData.(map[string]interface{})
			rowCells := rowDataMap[k8sCommonBean.K8sClusterResourceCellKey].([]interface{})
			for key, value := range keysToBeFetchedFromColumnDefinitions {
				resolvedValueFromRowCell := rowCells[value].(string)
				switch key {
				case k8sCommonBean.K8sResourceColumnDefinitionName:
					appListDto.Name = resolvedValueFromRowCell
				case k8sCommonBean.K8sResourceColumnDefinitionSyncStatus:
					appListDto.SyncStatus = resolvedValueFromRowCell
				case k8sCommonBean.K8sResourceColumnDefinitionHealthStatus:
					appListDto.HealthStatus = resolvedValueFromRowCell
				}
			}
			rowObject := rowDataMap[k8sCommonBean.K8sClusterResourceObjectKey].(map[string]interface{})
			for _, key := range keysToBeFetchedFromRawObject {
				switch key {
				case k8sCommonBean.K8sClusterResourceNamespaceKey:
					metadata := rowObject[k8sCommonBean.K8sClusterResourceMetadataKey].(map[string]interface{})
					appListDto.Namespace = metadata[k8sCommonBean.K8sClusterResourceNamespaceKey].(string)
				}
			}

			appLists = append(appLists, appListDto)
		}
	}
	return appLists
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

func (impl *ArgoApplicationServiceImpl) GetServerConfigIfClusterIsNotAddedOnDevtron(resourceResp *k8s.ManifestResponse, restConfig *rest.Config,
	clusterWithApplicationObject clusterRepository.Cluster, clusterServerUrlIdMap map[string]int) (*rest.Config, error) {
	impl.logger.Infow("request GetServerConfigIfClusterIsNotAddedOnDevtron", "resourceResp", resourceResp, "restConfig", restConfig,
		"clusterWithApplicationObject", clusterWithApplicationObject, "clusterServerUrlIdMap", clusterServerUrlIdMap)
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
		//cluster is not added on devtron, need to get server config from secret which argo-cd saved
		coreV1Client, err := impl.k8sUtil.GetCoreV1ClientByRestConfig(restConfig)
		secrets, err := coreV1Client.Secrets(bean.AllNamespaces).List(context.Background(), v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labels.Set{"argocd.argoproj.io/secret-type": "cluster"}).String(),
		})
		if err != nil {
			impl.logger.Errorw("error in getting resource list, secrets", "err", err)
			return nil, err
		}
		impl.logger.Infow("secrets got by argocd app", "err", err, "secrets", secrets)
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
		impl.logger.Infow("configOfClusterWhereAppIsDeployed", configOfClusterWhereAppIsDeployed)
		if configOfClusterWhereAppIsDeployed != nil {
			restConfig.Host = destinationServer
			restConfig.TLSClientConfig = rest.TLSClientConfig{
				Insecure: configOfClusterWhereAppIsDeployed.TlsClientConfig.Insecure,
				KeyFile:  configOfClusterWhereAppIsDeployed.TlsClientConfig.KeyData,
				CAFile:   configOfClusterWhereAppIsDeployed.TlsClientConfig.CaData,
				CertFile: configOfClusterWhereAppIsDeployed.TlsClientConfig.CertData,
			}
			restConfig.BearerToken = configOfClusterWhereAppIsDeployed.BearerToken
		}
	}
	return restConfig, nil
}

func (impl *ArgoApplicationServiceImpl) GetClusterConfigFromAllClusters(clusterId int) (*k8s.ClusterConfig, clusterRepository.Cluster, map[string]int, error) {
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
	clusterConfig, err := clusterBean.GetClusterConfig()
	return clusterConfig, clusterWithApplicationObject, clusterServerUrlIdMap, err
}

func (impl *ArgoApplicationServiceImpl) GetRestConfigForExternalArgo(ctx context.Context, clusterId int, externalArgoApplicationName string) (*rest.Config, error) {
	clusterConfig, clusterWithApplicationObject, clusterServerUrlIdMap, err := impl.GetClusterConfigFromAllClusters(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster config", "err", err, "clusterId", clusterId)
		return nil, err
	}
	impl.logger.Infow("GetRestConfigForExternalArgo, GetClusterConfigFromAllClusters", "clusterConfig", clusterConfig, "clusterWithApplicationObject", clusterWithApplicationObject, "clusterServerUrlIdMap", clusterServerUrlIdMap)
	restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config", "err", err, "clusterId", clusterId)
		return nil, err
	}
	impl.logger.Infow("GetRestConfigByCluster", "restConfig", restConfig)
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
	impl.logger.Infow("GetServerConfigIfClusterIsNotAddedOnDevtron", "restConfig", restConfig)
	return restConfig, nil
}
