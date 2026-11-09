package common

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/read"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	util "github.com/devtron-labs/devtron/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

type placementConfigs struct {
	deploymentConfig.Repository
	rows []*deploymentConfig.DeploymentConfig
}

func (p placementConfigs) GetAllConfigsForActiveApps() ([]*deploymentConfig.DeploymentConfig, error) {
	return p.rows, nil
}
func (p placementConfigs) GetAllEnvLevelConfigsWithReleaseMode(mode string) ([]*deploymentConfig.DeploymentConfig, error) {
	var rows []*deploymentConfig.DeploymentConfig
	for _, r := range p.rows {
		if r.ReleaseMode == mode {
			rows = append(rows, r)
		}
	}
	return rows, nil
}

type placementPipelines struct {
	pipelineConfig.PipelineRepository
}

func (placementPipelines) GetAllArgoAppInfoByDeploymentAppNames([]string) ([]*pipelineConfig.PipelineDeploymentConfigObj, error) {
	return []*pipelineConfig.PipelineDeploymentConfigObj{{AppId: 1, EnvironmentId: 1, DeploymentAppName: "created"}, {AppId: 2, EnvironmentId: 1, DeploymentAppName: "linked"}, {AppId: 3, EnvironmentId: 1, DeploymentAppName: "legacy"}}, nil
}

type placementInstalled struct{ read.InstalledAppReadServiceEA }

func (placementInstalled) GetAllArgoAppNamesByDeploymentAppNames([]string) ([]string, error) {
	return []string{"chart-store"}, nil
}
func TestApplicationDiscoveryPreservesRecordedPlacement(t *testing.T) {
	var rows []*deploymentConfig.DeploymentConfig
	for _, c := range []struct {
		id, cluster     int
		namespace, mode string
	}{{1, 1, "recorded-native", "create"}, {2, 7, "recorded-linked", "link"}} {
		raw, err := json.Marshal(&bean.ReleaseConfiguration{Version: bean.Version, ArgoCDSpec: bean.ArgoCDSpec{Metadata: bean.ApplicationMetadata{ClusterId: c.cluster, Namespace: c.namespace}}})
		require.NoError(t, err)
		rows = append(rows, &deploymentConfig.DeploymentConfig{AppId: c.id, EnvironmentId: 1, DeploymentAppType: "argo_cd", ReleaseMode: c.mode, ReleaseConfig: string(raw)})
	}
	rows = append(rows, &deploymentConfig.DeploymentConfig{AppId: 4, DeploymentAppType: "argo_cd", ReleaseConfig: "not-json"}, &deploymentConfig.DeploymentConfig{AppId: 5, EnvironmentId: 1, DeploymentAppType: "helm", ReleaseConfig: "not-json"})
	service := &DeploymentConfigServiceImpl{logger: zap.NewNop().Sugar(), deploymentConfigRepository: placementConfigs{rows: rows}, pipelineRepository: placementPipelines{}, installedAppReadService: placementInstalled{}, acdAuthConfig: &util.ACDAuthConfig{ACDConfigMapNamespace: "argo-control"}}
	actual, err := service.GetAllArgoAppInfosByDeploymentAppNames([]string{"created", "linked", "legacy", "chart-store"})
	require.NoError(t, err)
	require.Equal(t, []*bean.DevtronArgoCdAppInfo{{ArgoCdAppName: "created", ArgoAppClusterId: 1, ArgoAppNamespace: "recorded-native"}, {ArgoCdAppName: "linked", ArgoAppClusterId: 7, ArgoAppNamespace: "recorded-linked"}, {ArgoCdAppName: "legacy", ArgoAppClusterId: 1, ArgoAppNamespace: "devtroncd"}, {ArgoCdAppName: "chart-store", ArgoAppClusterId: 1, ArgoAppNamespace: "argo-control"}}, actual)
}
