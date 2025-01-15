package util

import v1 "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"

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
