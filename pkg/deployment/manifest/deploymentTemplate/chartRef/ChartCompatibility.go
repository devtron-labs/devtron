/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
