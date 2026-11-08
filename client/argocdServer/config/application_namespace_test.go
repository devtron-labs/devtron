package config

import (
	"testing"

	commonk8s "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	clusterBean "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/read"
	"github.com/devtron-labs/devtron/pkg/util"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

type targetClusterReader struct{ read.ClusterReadService }

func (targetClusterReader) FindOne(string) (*clusterBean.ClusterBean, error) {
	return &clusterBean.ClusterBean{ServerUrl: "https://cluster.example.test"}, nil
}

type targetKubernetesService struct{ commonk8s.K8sService }

func (targetKubernetesService) GetRestConfigByCluster(cluster *commonk8s.ClusterConfig, _ ...commonk8s.K8sServiceOpts) (*rest.Config, error) {
	return &rest.Config{Host: cluster.Host}, nil
}

func TestRepositoryNamespaceUsesArgoConfiguration(t *testing.T) {
	getter := &ArgoCDConfigGetterImpl{
		config:             &bean.Config{Namespace: "argo-control"},
		ACDAuthConfig:      &util.ACDAuthConfig{ACDConfigMapNamespace: "installation", ACDConfigMapName: "argocd-cm"},
		clusterReadService: targetClusterReader{}, K8sService: targetKubernetesService{},
	}
	config, err := getter.GetK8sConfig()
	require.NoError(t, err)
	require.Equal(t, "argo-control", config.AcdNamespace)
	require.Equal(t, "argocd-cm", config.AcdConfigMapName)
	require.Equal(t, "https://cluster.example.test", config.RestConfig.Host)
	require.Equal(t, "installation", getter.ACDAuthConfig.ACDConfigMapNamespace)
}
