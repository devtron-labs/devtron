package deployment

import (
	"context"
	"testing"

	"github.com/devtron-labs/devtron/client/argocdServer"
	environmentBean "github.com/devtron-labs/devtron/pkg/cluster/environment/bean"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/stretchr/testify/require"
)

type targetApplicationCreator struct {
	argocdServer.ArgoK8sClient
	request *argocdServer.AppTemplate
}

func (client *targetApplicationCreator) CreateAcdApp(_ context.Context, request *argocdServer.AppTemplate, _ string) (string, error) {
	client.request = request
	return request.ApplicationName, nil
}

func TestChartApplicationUsesConfiguredArgoTarget(t *testing.T) {
	creator := &targetApplicationCreator{}
	service := &FullModeDeploymentServiceImpl{argoK8sClient: creator,
		acdConfig: &argocdServer.ACDConfig{ApplicationNamespace: "argo-control", ApplicationProject: "applications", ArgoCDAutoSyncEnabled: true}}
	err := service.createInArgo(context.Background(), &commonBean.ChartGitAttribute{RepoUrl: "https://git.example.test/configs.git", ChartLocation: "chart"},
		environmentBean.EnvironmentBean{Namespace: "application-dev", ClusterServerUrl: "https://cluster.example.test"}, "sample-app")
	require.NoError(t, err)
	require.Equal(t, "argo-control", creator.request.Namespace)
	require.Equal(t, "applications", creator.request.Project)
	require.Equal(t, "application-dev", creator.request.TargetNamespace)
	require.Equal(t, "https://cluster.example.test", creator.request.TargetServer)
	require.Equal(t, "https://git.example.test/configs.git", creator.request.RepoUrl)
	require.True(t, creator.request.AutoSyncEnabled)
}
