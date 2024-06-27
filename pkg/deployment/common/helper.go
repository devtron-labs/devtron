package common

import "github.com/devtron-labs/devtron/pkg/deployment/common/bean"

func GetDeploymentConfigType(isCustomGitOpsRepo bool) string {
	if isCustomGitOpsRepo {
		return string(bean.CUSTOM)
	}
	return string(bean.SYSTEM_GENERATED)
}

func IsCustomGitOpsRepo(deploymentConfigType string) bool {
	return deploymentConfigType == bean.CUSTOM.String()
}
