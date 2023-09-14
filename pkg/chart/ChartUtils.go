package chart

import (
	"encoding/json"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"strings"
)

func PatchWinterSoldierConfig(override json.RawMessage, newChartType string) (json.RawMessage, error) {
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(override), &jsonMap); err != nil {
		return override, err
	}
	updatedJson, err := PatchWinterSoldierIfExists(newChartType, jsonMap)
	if err != nil {
		return override, err
	}
	updatedOverride, err := json.Marshal(updatedJson)

	if err != nil {
		return override, err
	}
	return updatedOverride, nil
}

func PatchWinterSoldierIfExists(newChartType string, jsonMap map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	winterSoldierConfig, found := jsonMap["winterSoldier"]
	if !found {
		return jsonMap, nil
	}
	var winterSoldierUnmarshalled map[string]json.RawMessage
	if err := json.Unmarshal([]byte(winterSoldierConfig), &winterSoldierUnmarshalled); err != nil {
		return jsonMap, err
	}

	_, found = winterSoldierUnmarshalled["type"]
	if !found {
		return jsonMap, nil
	}
	switch newChartType {
	case DeploymentChartType:
		winterSoldierUnmarshalled["type"] = json.RawMessage("\"Deployment\"")
	case RolloutChartType:
		winterSoldierUnmarshalled["type"] = json.RawMessage("\"Rollout\"")
	}

	winterSoldierMarshalled, err := json.Marshal(winterSoldierUnmarshalled)
	if err != nil {
		return jsonMap, err
	}
	jsonMap["winterSoldier"] = winterSoldierMarshalled
	return jsonMap, nil
}

func SetReservedChartList(devtronChartList []*chartRepoRepository.ChartRef) {
	reservedChartRefNamesList := []ReservedChartList{
		{Name: strings.ToLower(RolloutChartType), LocationPrefix: ""},
		{Name: "", LocationPrefix: ReferenceChart},
	}
	for _, devtronChart := range devtronChartList {
		reservedChartRefNamesList = append(reservedChartRefNamesList, ReservedChartList{
			Name:           strings.ToLower(devtronChart.Name),
			LocationPrefix: strings.Split(devtronChart.Location, "_")[0],
		})
	}
	ReservedChartRefNamesList = &reservedChartRefNamesList
}

//func IsFlaggerCanaryEnabled(override json.RawMessage) (bool, error) {
//
//}
