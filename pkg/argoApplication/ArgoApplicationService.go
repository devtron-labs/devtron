package argoApplication

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
	"github.com/devtron-labs/devtron/util/argo"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
)

type ArgoApplicationService interface {
	ListApplications(clusterIds []int) ([]*bean.ArgoApplicationListDto, error)
	GetAppDetail(resourceName, resourceNamespace string, clusterId int) (*bean.ArgoApplicationDetailDto, error)
}

type ArgoApplicationServiceImpl struct {
	logger            *zap.SugaredLogger
	clusterRepository clusterRepository.ClusterRepository
	k8sUtil           *k8s.K8sUtil
	k8sCommonService  k8s2.K8sCommonService
	argoUserService   argo.ArgoUserService
	argoServiceClient application.ServiceClient
	helmAppService    client.HelmAppService
}

func NewArgoApplicationServiceImpl(logger *zap.SugaredLogger,
	clusterRepository clusterRepository.ClusterRepository,
	k8sUtil *k8s.K8sUtil,
	k8sCommonService k8s2.K8sCommonService,
	argoUserService argo.ArgoUserService,
	argoServiceClient application.ServiceClient,
	helmAppService client.HelmAppService) *ArgoApplicationServiceImpl {
	return &ArgoApplicationServiceImpl{
		logger:            logger,
		clusterRepository: clusterRepository,
		k8sUtil:           k8sUtil,
		k8sCommonService:  k8sCommonService,
		argoUserService:   argoUserService,
		argoServiceClient: argoServiceClient,
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
		if cluster.IsVirtualCluster || len(cluster.ErrorInConnecting) != 0 {
			continue
		}
		clusterBean := cluster2.GetClusterBean(cluster)
		clusterConfig, err := clusterBean.GetClusterConfig()
		restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
		if err != nil {
			impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", cluster.Id)
			return nil, err
		}
		resp, _, err := impl.k8sUtil.GetResourceList(context.Background(), restConfig, bean.GvkForArgoApplication, bean.AllNamespaces)
		if err != nil {
			if errStatus, ok := err.(*errors.StatusError); ok {
				if errStatus.Status().Code == 404 {
					//no argo apps found, not sending error
					impl.logger.Warnw("error in getting external argo app list, no apps found", "err", err, "clusterId", cluster.Id)
					continue
				}
			}
			impl.logger.Errorw("error in getting resource list", "err", err)
			return nil, err
		}
		appLists := getApplicationListDtos(resp.Resources.Object, cluster.ClusterName, cluster.Id)
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
	cluster, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by id", "err", err, "clusterId", clusterId)
		return nil, err
	}
	if cluster != nil {
		appDetail.ClusterName = cluster.ClusterName
	}
	if cluster.IsVirtualCluster {
		return appDetail, nil
	} else if len(cluster.ErrorInConnecting) != 0 {
		return nil, fmt.Errorf("error in connecting to cluster")
	}
	clusterBean := cluster2.GetClusterBean(*cluster)
	clusterConfig, err := clusterBean.GetClusterConfig()
	restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster Id", "err", err, "clusterId", cluster.Id)
		return nil, err
	}
	resp, err := impl.k8sUtil.GetResource(context.Background(), resourceNamespace, resourceName, bean.GvkForArgoApplication, restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting resource list", "err", err)
		return nil, err
	}
	var argoManagedResources []*bean.ArgoManagedResource
	if resp != nil && resp.Manifest.Object != nil {
		appDetail.Manifest = resp.Manifest.Object
		appDetail.HealthStatus, appDetail.SyncStatus, argoManagedResources =
			getHealthSyncStatusAndManagedResourcesForArgoK8sRawObject(resp.Manifest.Object)
	}
	resourceTreeResp, err := impl.getResourceTreeForExternalCluster(clusterId, argoManagedResources)
	if err != nil {
		impl.logger.Errorw("error in getting resource tree response", "err", err)
		return nil, err
	}
	appDetail.ResourceTree = resourceTreeResp
	return appDetail, nil
}

func (impl *ArgoApplicationServiceImpl) getResourceTreeForExternalCluster(clusterId int, argoManagedResources []*bean.ArgoManagedResource) (*client.ResourceTreeResponse, error) {
	var resources []*client.ExternalResourceDetail
	for _, argoManagedResource := range argoManagedResources {
		resources = append(resources, &client.ExternalResourceDetail{
			Group:     argoManagedResource.Group,
			Kind:      argoManagedResource.Kind,
			Version:   argoManagedResource.Version,
			Name:      argoManagedResource.Name,
			Namespace: argoManagedResource.Namespace,
		})
	}
	resourceTreeResp, err := impl.helmAppService.GetResourceTreeForExternalResources(context.Background(), clusterId, resources)
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

func getHealthSyncStatusAndManagedResourcesForArgoK8sRawObject(obj map[string]interface{}) (string,
	string, []*bean.ArgoManagedResource) {
	var healthStatus, syncStatus string
	argoManagedResources := make([]*bean.ArgoManagedResource, 0)
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
	return healthStatus, syncStatus, argoManagedResources
}
