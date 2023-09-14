package chart

type stringSet map[string]struct{}

const (
	DeploymentChartType = "Deployment"
	RolloutChartType    = "Rollout Deployment"
	ReferenceChart      = "reference-chart"
)

var chartCompatibilityMatrix = map[string]stringSet{
	DeploymentChartType: {RolloutChartType: {}, DeploymentChartType: {}},
	RolloutChartType:    {DeploymentChartType: {}, RolloutChartType: {}},
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
