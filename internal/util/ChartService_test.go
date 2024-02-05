package util

import (
	"context"
	"github.com/devtron-labs/devtron/util"
	"github.com/stretchr/testify/assert"
	"k8s.io/helm/pkg/chartutil"
	chart2 "k8s.io/helm/pkg/proto/hapi/chart"
	"math/rand"
	"testing"
)

func TestChartTemplateService(t *testing.T) {

	t.Run("getValues", func(t *testing.T) {
		logger, err := NewSugardLogger()
		assert.Nil(t, err)
		impl := ChartTemplateServiceImpl{
			logger: logger,
		}
		directory := "/scripts/devtron-reference-helm-charts/reference-chart_3-11-0"
		pipelineStrategyPath := "pipeline-values.yaml"
		values, err := impl.getValues(directory, pipelineStrategyPath)
		assert.Nil(t, err)
		assert.NotNil(t, values)
	})

	t.Run("buildChart", func(t *testing.T) {
		logger, err := NewSugardLogger()
		assert.Nil(t, err)
		impl := ChartTemplateServiceImpl{
			logger:          logger,
			randSource:      rand.NewSource(0),
			chartWorkingDir: "/tmp/charts/",
		}
		chartMetaData := &chart2.Metadata{
			Name:    "sample-app",
			Version: "1.0.0",
		}
		refChartDir := "/scripts/devtron-reference-helm-charts/reference-chart_3-11-0"

		builtChartPath, err := impl.BuildChart(context.Background(), chartMetaData, refChartDir)
		assert.Nil(t, err)
		assert.DirExists(t, builtChartPath)

		isValidChart, err := chartutil.IsChartDir(builtChartPath)
		assert.Nil(t, err)
		assert.Equal(t, isValidChart, true)
	})

	t.Run("LoadChartInBytesWithDeleteFalse", func(t *testing.T) {
		logger, err := NewSugardLogger()
		assert.Nil(t, err)
		impl := ChartTemplateServiceImpl{
			logger:          logger,
			randSource:      rand.NewSource(0),
			chartWorkingDir: "/tmp/charts/",
		}
		chartMetaData := &chart2.Metadata{
			Name:    "sample-app",
			Version: "1.0.0",
		}
		refChartDir := "/scripts/devtron-reference-helm-charts/reference-chart_3-11-0"

		builtChartPath, err := impl.BuildChart(context.Background(), chartMetaData, refChartDir)

		chartBytes, err := impl.LoadChartInBytes(builtChartPath, false, "", "")
		assert.Nil(t, err)

		chartBytesLen := len(chartBytes)
		assert.NotEqual(t, chartBytesLen, 0)

	})

	t.Run("LoadChartInBytesWithDeleteTrue", func(t *testing.T) {
		logger, err := NewSugardLogger()
		assert.Nil(t, err)
		impl := ChartTemplateServiceImpl{
			logger:          logger,
			randSource:      rand.NewSource(0),
			chartWorkingDir: "/tmp/charts/",
		}
		chartMetaData := &chart2.Metadata{
			Name:    "sample-app",
			Version: "1.0.0",
		}
		refChartDir := "/scripts/devtron-reference-helm-charts/reference-chart_3-11-0"

		builtChartPath, err := impl.BuildChart(context.Background(), chartMetaData, refChartDir)

		chartBytes, err := impl.LoadChartInBytes(builtChartPath, true, "", "")
		assert.Nil(t, err)

		assert.NoDirExists(t, builtChartPath)

		chartBytesLen := len(chartBytes)
		assert.NotEqual(t, chartBytesLen, 0)

	})

	t.Run("GetGitOpsRepoName", func(t *testing.T) {
		logger, err := NewSugardLogger()
		assert.Nil(t, err)
		globalEnvVariables, err := util.GetGlobalEnvVariables()
		globalEnvVariables.GitOpsRepoPrefix = "devtron"
		impl := ChartTemplateServiceImpl{
			logger:             logger,
			globalEnvVariables: globalEnvVariables,
		}
		appName := "sample"
		wantedRepoName := "devtron-sample"
		gitOpsRepoName := impl.GetGitOpsRepoName(appName)
		assert.Equal(t, wantedRepoName, gitOpsRepoName)
	})

}
