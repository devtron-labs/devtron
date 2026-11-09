package pipeline

import (
	app "github.com/devtron-labs/devtron/internal/sql/repository/app"
	pc "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	installed "github.com/devtron-labs/devtron/pkg/appStore/installedApp/read"
	installedbean "github.com/devtron-labs/devtron/pkg/appStore/installedApp/read/bean"
	charts "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	environment "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	cluster "github.com/devtron-labs/devtron/pkg/cluster/repository"
	common "github.com/devtron-labs/devtron/pkg/deployment/common"
	bean "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	override "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/bean"
	chartbean "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	chartread "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/read"
	read "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/read"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/require"
	"testing"
)

type placementOverride struct{ read.EnvConfigOverrideService }

func (placementOverride) FindLatestChartForAppByAppIdAndEnvId(*pg.Tx, int, int) (*override.EnvConfigOverride, error) {
	return &override.EnvConfigOverride{Id: 1, IsOverride: true, Chart: &charts.Chart{ChartRefId: 1, ChartVersion: "0.1.0"}}, nil
}

type placementChart struct{ chartread.ChartRefReadService }

func (placementChart) FindById(int) (*chartbean.ChartRefDto, error) {
	return &chartbean.ChartRefDto{Location: "charts/example"}, nil
}

type placementPipelines struct{ pc.PipelineRepository }

func (placementPipelines) GetArgoPipelineByArgoAppName(string) ([]pc.Pipeline, error) {
	return nil, nil
}

type placementDeployment struct{ common.DeploymentConfigService }

func (placementDeployment) FilterPipelinesByApplicationClusterIdAndNamespace([]pc.Pipeline, int, string) (pc.Pipeline, error) {
	return pc.Pipeline{}, nil
}

type placementInstalled struct {
	installed.InstalledAppReadService
}

func (placementInstalled) GetInstalledAppByGitOpsAppName(string) (*installedbean.InstalledAppMin, error) {
	return &installedbean.InstalledAppMin{AppId: 1, EnvironmentId: 1}, nil
}

func TestApplicationPlacementMetadataAndDuplicateLink(t *testing.T) {
	for _, namespace := range []string{"devtroncd", "argo-control"} {
		t.Run(namespace, func(t *testing.T) {
			t.Setenv("ACD_NAMESPACE", namespace)
			t.Setenv("CD_NAMESPACE", "connection-only")
			service := &CdPipelineConfigServiceImpl{envConfigOverrideService: placementOverride{}, chartRefReadService: placementChart{}, pipelineRepository: placementPipelines{}, deploymentConfigService: placementDeployment{}, installedAppReadService: placementInstalled{}}
			t.Run("metadata", func(t *testing.T) {
				config, err := service.parseReleaseConfigForACDApp(&app.App{Id: 1}, &bean.DeploymentConfig{}, &environment.Environment{Id: 1, Namespace: "workload", Cluster: &cluster.Cluster{ServerUrl: "https://kubernetes.default.svc"}})
				require.NoError(t, err)
				require.Equal(t, namespace, config.ArgoCDSpec.Metadata.Namespace)
				require.Equal(t, "workload", config.ArgoCDSpec.Spec.Destination.Namespace)
			})
			t.Run("duplicate-link", func(t *testing.T) {
				require.Error(t, service.ValidateIfAcdAppAlreadyLinked("example", 1, namespace))
				require.NoError(t, service.ValidateIfAcdAppAlreadyLinked("example", 1, "external-apps"))
			})
		})
	}
}
