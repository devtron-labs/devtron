package devtronApps

import (
	"context"
	"testing"

	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	environment "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	cluster "github.com/devtron-labs/devtron/pkg/cluster/repository"
	deployment "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	manifest "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/bean"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type targetEnvironmentReader struct {
	environment.EnvironmentRepository
}

func (targetEnvironmentReader) FindById(id int) (*environment.Environment, error) {
	return &environment.Environment{Id: id, Cluster: &cluster.Cluster{ServerUrl: "https://cluster.example.test"}}, nil
}

type targetGitOperation struct{ git.GitOperationService }

func (targetGitOperation) GetRepoUrlWithUserName(url string) (string, error) { return url, nil }

type targetPipelineRepository struct {
	pipelineConfig.PipelineRepository
	created bool
}

func (repository *targetPipelineRepository) SetDeploymentAppCreatedInPipeline(created bool, _ int, _ int32) error {
	repository.created = created
	return nil
}

type targetApplicationCreator struct {
	argocdServer.ArgoK8sClient
	request *argocdServer.AppTemplate
}

func (client *targetApplicationCreator) CreateAcdApp(_ context.Context, request *argocdServer.AppTemplate, _ string) (string, error) {
	client.request = request
	return request.ApplicationName, nil
}

func TestPipelineApplicationUsesConfiguredArgoTarget(t *testing.T) {
	creator := &targetApplicationCreator{}
	pipelines := &targetPipelineRepository{}
	handler := &HandlerServiceImpl{logger: zap.NewNop().Sugar(), envRepository: targetEnvironmentReader{},
		gitOperationService: targetGitOperation{}, pipelineRepository: pipelines, argoK8sClient: creator,
		ACDConfig: &argocdServer.ACDConfig{ApplicationNamespace: "argo-control", ApplicationProject: "applications", ArgoCDAutoSyncEnabled: true}}
	name, err := handler.createArgoApplicationIfRequired(context.Background(),
		&manifest.EnvConfigOverride{TargetEnvironment: 42, Namespace: "application-dev"},
		&pipelineConfig.Pipeline{Id: 123, DeploymentAppName: "sample-app"},
		&deployment.DeploymentConfig{DeploymentAppType: util.PIPELINE_DEPLOYMENT_TYPE_ACD, RepoURL: "https://git.example.test/configs.git"}, 1)
	require.NoError(t, err)
	require.Equal(t, "sample-app", name)
	require.Equal(t, "argo-control", creator.request.Namespace)
	require.Equal(t, "applications", creator.request.Project)
	require.Equal(t, "application-dev", creator.request.TargetNamespace)
	require.Equal(t, "https://cluster.example.test", creator.request.TargetServer)
	require.True(t, creator.request.AutoSyncEnabled)
	require.True(t, pipelines.created)
	creator.request = nil
	pipelines.created = false
	handler.ACDConfig.ApplicationProject = "another-project"
	name, err = handler.createArgoApplicationIfRequired(context.Background(),
		&manifest.EnvConfigOverride{TargetEnvironment: 42, Namespace: "application-dev"},
		&pipelineConfig.Pipeline{Id: 123, DeploymentAppName: "sample-app", DeploymentAppCreated: true},
		&deployment.DeploymentConfig{DeploymentAppType: util.PIPELINE_DEPLOYMENT_TYPE_ACD}, 1)
	require.NoError(t, err)
	require.Equal(t, "sample-app", name)
	require.Nil(t, creator.request, "configuration changes must not recreate an existing application")
	require.False(t, pipelines.created)
}
