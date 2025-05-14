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
