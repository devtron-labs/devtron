package chartRef

import "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"

type stringSet map[string]struct{}

var chartCompatibilityMatrix = map[string]stringSet{
	bean.DeploymentChartType: {bean.RolloutChartType: {}, bean.DeploymentChartType: {}},
	bean.RolloutChartType:    {bean.DeploymentChartType: {}, bean.RolloutChartType: {}},
}

func CheckCompatibility(oldChartType, newChartType string) bool {
	compatibilityOfOld, found := chartCompatibilityMatrix[oldChartType]
	if !found {
		return false
	}
	_, found = compatibilityOfOld[newChartType]
	if !found {
		return false
	}
	return true
}

func CompatibleChartsWith(chartType string) []string {
	resultSet, found := chartCompatibilityMatrix[chartType]
	if !found {
		return []string{}
	}
	var result []string
	for k, _ := range resultSet {
		result = append(result, k)
	}
	return result
}
