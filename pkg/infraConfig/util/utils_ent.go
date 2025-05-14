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
	"fmt"
	globalUtil "github.com/devtron-labs/devtron/internal/util"
	v1 "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"net/http"
)

func getEntConfigKeyStr(configKey v1.ConfigKey) v1.ConfigKeyStr {
	return ""
}

func getEntConfigKey(configKeyStr v1.ConfigKeyStr) v1.ConfigKey {
	return 0
}

func getConfigKeysMapForPlatformEnt(defaultConfigKeys v1.InfraConfigKeys, platform string) v1.InfraConfigKeys {
	return defaultConfigKeys
}

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
