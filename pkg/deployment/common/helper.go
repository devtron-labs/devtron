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

func GetAppIdToEnvIsMappingFromConfigSelectors(configSelector []*bean.DeploymentConfigSelector) map[int][]int {
	appIdToEnvIdsMap := make(map[int][]int)
	for _, c := range configSelector {
		if _, ok := appIdToEnvIdsMap[c.AppId]; !ok {
			appIdToEnvIdsMap[c.AppId] = make([]int, 0)
		}
		appIdToEnvIdsMap[c.AppId] = append(appIdToEnvIdsMap[c.AppId], c.EnvironmentId)
	}
	return appIdToEnvIdsMap
}
