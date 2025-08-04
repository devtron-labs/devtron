package util

import (
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/bean/common"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/urlUtil"
)

func IsValidUrlSubPath(subPath string) bool {
	url := "http://127.0.0.1:8080/" + subPath
	return urlUtil.IsValidUrl(url)
}

// in oss, only WorkflowCacheConfigInherit is supported
func GetWorkflowCacheConfig(WorkflowCacheConfig common.WorkflowCacheConfigType, globalValue bool) bean.WorkflowCacheConfig {
	//explicitly return inherit here to handle empty case
	return bean.WorkflowCacheConfig{
		Type:        common.WorkflowCacheConfigInherit,
		Value:       !globalValue,
		GlobalValue: !globalValue,
	}
}

// in oss, only WorkflowCacheConfigInherit is supported
func GetWorkflowCacheConfigWithBackwardCompatibility(WorkflowCacheConfig common.WorkflowCacheConfigType, WorkflowCacheConfigEnv string, globalValue bool, oldGlobalValue bool) bean.WorkflowCacheConfig {
	isEmptyJson, _ := util.IsEmptyJSONForJsonString(WorkflowCacheConfigEnv)
	//TODO: error handling in next phase
	if isEmptyJson {
		//this means new global flag is not configured
		return bean.WorkflowCacheConfig{
			Type:        common.WorkflowCacheConfigInherit,
			Value:       !oldGlobalValue,
			GlobalValue: !oldGlobalValue,
		}
	} else {
		return bean.WorkflowCacheConfig{
			Type:        common.WorkflowCacheConfigInherit,
			Value:       !globalValue,
			GlobalValue: !globalValue,
		}
	}
}

// RemoveDuplicateInts helper function to remove duplicate integers from slice
func RemoveDuplicateInts(slice []int) []int {
	keys := make(map[int]bool)
	var result []int
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}
