package util

import (
	"fmt"
	globalUtil "github.com/devtron-labs/devtron/internal/util"
	v1 "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"net/http"
)

func IsValidProfileNameRequested(profileName, payloadProfileName string) bool {
	if len(payloadProfileName) == 0 || len(profileName) == 0 {
		return false
	}
	if profileName != v1.GLOBAL_PROFILE_NAME || payloadProfileName != v1.GLOBAL_PROFILE_NAME {
		return false
	}
	return true
}

func IsValidProfileNameRequestedV0(profileName, payloadProfileName string) bool {
	if len(payloadProfileName) == 0 || len(profileName) == 0 {
		return false
	}
	if profileName != v1.GLOBAL_PROFILE_NAME || payloadProfileName != v1.GLOBAL_PROFILE_NAME {
		return false
	}
	return true
}

func validatePlatformName(platform string, buildxDriverType v1.BuildxDriver) error {
	if platform != v1.RUNNER_PLATFORM {
		errMsg := fmt.Sprintf("platform %q is not supported", platform)
		return globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	return nil
}
