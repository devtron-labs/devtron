package config

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	"github.com/devtron-labs/devtron/pkg/argoApplication/helper"
	"github.com/devtron-labs/devtron/pkg/cluster/adapter"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
)

type ArgoApplicationConfigServiceImpl struct {
	logger            *zap.SugaredLogger
	k8sUtil           *k8s.K8sServiceImpl
	clusterRepository clusterRepository.ClusterRepository
}

type ArgoApplicationConfigService interface {
	GetRestConfigForExternalArgo(ctx context.Context, externalArgoAppIdentifier *bean.ArgoAppIdentifier) (*rest.Config, error)
	GetClusterConfigFromAllClusters(clusterId int) (*k8s.ClusterConfig, clusterRepository.Cluster, map[string]int, error)
}

func NewArgoApplicationConfigServiceImpl(logger *zap.SugaredLogger,
	k8sUtil *k8s.K8sServiceImpl,
	clusterRepository clusterRepository.ClusterRepository) *ArgoApplicationConfigServiceImpl {
	return &ArgoApplicationConfigServiceImpl{
		logger:            logger,
		k8sUtil:           k8sUtil,
		clusterRepository: clusterRepository,
	}
}

func (impl *ArgoApplicationConfigServiceImpl) GetClusterConfigFromAllClusters(clusterId int) (*k8s.ClusterConfig, clusterRepository.Cluster, map[string]int, error) {
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
	clusterBean := adapter.GetClusterBean(clusterWithApplicationObject)
	clusterConfig := clusterBean.GetClusterConfig()
	return clusterConfig, clusterWithApplicationObject, clusterServerUrlIdMap, err
}

func (impl *ArgoApplicationConfigServiceImpl) GetRestConfigForExternalArgo(ctx context.Context, externalArgoAppIdentifier *bean.ArgoAppIdentifier) (*rest.Config, error) {
	clusterConfig, clusterWithApplicationObject, clusterServerUrlIdMap, err := impl.GetClusterConfigFromAllClusters(externalArgoAppIdentifier.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster config", "err", err, "clusterId", externalArgoAppIdentifier.ClusterId)
		return nil, err
	}
	restConfig, err := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config", "err", err, "clusterId", externalArgoAppIdentifier.ClusterId)
		return nil, err
	}
	resourceResp, err := impl.k8sUtil.GetResource(ctx, externalArgoAppIdentifier.Namespace, externalArgoAppIdentifier.AppName, bean.GvkForArgoApplication, restConfig)
	if err != nil {
		impl.logger.Errorw("not on external cluster", "err", err, "externalArgoApplicationName", externalArgoAppIdentifier.AppName, "externalArgoApplicationNamespace", externalArgoAppIdentifier.Namespace)
		return nil, err
	}
	restConfig, err = impl.GetServerConfigIfClusterIsNotAddedOnDevtron(resourceResp, restConfig, clusterWithApplicationObject, clusterServerUrlIdMap)
	if err != nil {
		impl.logger.Errorw("error in getting server config", "err", err, "cluster with application object", clusterWithApplicationObject)
		return nil, err
	}
	return restConfig, nil
}

func (impl *ArgoApplicationConfigServiceImpl) GetServerConfigIfClusterIsNotAddedOnDevtron(resourceResp *k8s.ManifestResponse, restConfig *rest.Config,
	clusterWithApplicationObject clusterRepository.Cluster, clusterServerUrlIdMap map[string]int) (*rest.Config, error) {
	var destinationServer string
	if resourceResp != nil && resourceResp.Manifest.Object != nil {
		_, _, destinationServer, _ =
			helper.GetHealthSyncStatusDestinationServerAndManagedResourcesForArgoK8sRawObject(resourceResp.Manifest.Object)
	}
	appDeployedOnClusterId := 0
	if destinationServer == k8sCommonBean.DefaultClusterUrl {
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
