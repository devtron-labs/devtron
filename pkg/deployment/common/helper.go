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

package common

import (
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
)

func IsCustomGitOpsRepo(deploymentConfigType string) bool {
	return deploymentConfigType == bean.CUSTOM.String()
}

func GetAppIdToEnvIsMapping[T any](configArr []T, appIdSelector func(config T) int, envIdSelector func(config T) int) map[int][]int {

	appIdToEnvIdsMap := make(map[int][]int)
	for _, c := range configArr {
		appId := appIdSelector(c)
		envId := envIdSelector(c)
		if _, ok := appIdToEnvIdsMap[appId]; !ok {
			appIdToEnvIdsMap[appId] = make([]int, 0)
		}
		appIdToEnvIdsMap[appId] = append(appIdToEnvIdsMap[appId], envId)
	}
	return appIdToEnvIdsMap
}
