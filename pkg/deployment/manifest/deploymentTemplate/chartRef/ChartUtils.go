/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package chartRef

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
)

func patchWinterSoldierConfig(override json.RawMessage, newChartType string) (json.RawMessage, error) {
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal(override, &jsonMap); err != nil {
		return override, err
	}
	updatedJson, err := patchWinterSoldierIfExists(newChartType, jsonMap)
	if err != nil {
		return override, err
	}
	updatedOverride, err := json.Marshal(updatedJson)

	if err != nil {
		return override, err
	}
	return updatedOverride, nil
}

func patchWinterSoldierIfExists(newChartType string, jsonMap map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	winterSoldierConfig, found := jsonMap["winterSoldier"]
	if !found {
		return jsonMap, nil
	}
	var winterSoldierUnmarshalled map[string]json.RawMessage
	if err := json.Unmarshal(winterSoldierConfig, &winterSoldierUnmarshalled); err != nil {
		return jsonMap, err
	}

	_, found = winterSoldierUnmarshalled["type"]
	if !found {
		return jsonMap, nil
	}
	switch newChartType {
	case bean.DeploymentChartType:
		winterSoldierUnmarshalled["type"] = json.RawMessage("\"Deployment\"")
	case bean.RolloutChartType:
		winterSoldierUnmarshalled["type"] = json.RawMessage("\"Rollout\"")
	}

	winterSoldierMarshalled, err := json.Marshal(winterSoldierUnmarshalled)
	if err != nil {
		return jsonMap, err
	}
	jsonMap["winterSoldier"] = winterSoldierMarshalled
	return jsonMap, nil
}
