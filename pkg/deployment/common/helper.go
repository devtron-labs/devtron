package common

import (
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
)

func GetDeploymentConfigType(isCustomGitOpsRepo bool) string {
	if isCustomGitOpsRepo {
		return string(bean.CUSTOM)
	}
	return string(bean.SYSTEM_GENERATED)
}

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
