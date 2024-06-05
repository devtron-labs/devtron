/*
 * Copyright (c) 2024. Devtron Inc.
 */

package lockConfiguration

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"testing"
)

func setup() *LockConfigurationServiceImpl {
	sugaredLogger, _ := util.NewSugardLogger()
	sqlConfig, _ := sql.GetConfig()
	db, _ := sql.NewDbConnection(sqlConfig, sugaredLogger)
	repositoryImpl := NewRepositoryImpl(db)
	lockConfigurationServiceImpl := NewLockConfigurationServiceImpl(sugaredLogger, repositoryImpl, nil, util.MergeUtil{})
	return lockConfigurationServiceImpl
}

func TestArrayDiff(t *testing.T) {
	impl := setup()
	savedConfig := ""

	currentConfig := ""
	var savedConfigMap map[string]interface{}
	var currentConfigMap map[string]interface{}

	err := json.Unmarshal([]byte(savedConfig), &savedConfigMap)
	if err != nil {
		//impl.logger.Errorw("Error in umMarshal data", "err", err, "savedConfig", savedConfig)
	}

	err = json.Unmarshal([]byte(currentConfig), &currentConfigMap)
	if err != nil {
		//impl.logger.Errorw("Error in umMarshal data", "err", err, "currentConfig", currentConfig)

	}
	fmt.Println("initcontainer saved confimap length", len(savedConfigMap["EnvVariables"].([]interface{})))
	fmt.Println("initcontainer current confimap length", len(currentConfigMap["EnvVariables"].([]interface{})))
	allChanges, deletedMap, addedMap, modifiedMap, containChangesInArray, deletedPaths := impl.getDiffJson(savedConfigMap, currentConfigMap, "")
	fmt.Println(allChanges)
	fmt.Println(deletedMap)
	fmt.Println(addedMap)
	fmt.Println(modifiedMap)
	fmt.Println(containChangesInArray)
	fmt.Println(deletedPaths)

}
