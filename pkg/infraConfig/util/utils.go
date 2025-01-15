/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	unitsBean "github.com/devtron-labs/devtron/pkg/infraConfig/units/bean"
	"net/http"
	"slices"
	"strings"
)

func CreateConfigKeyPlatformMap(configs []*repository.InfraProfileConfigurationEntity) map[v1.ConfigKeyPlatformKey]bool {
	configMap := make(map[v1.ConfigKeyPlatformKey]bool, len(configs))
	for _, config := range configs {
		platform := config.ProfilePlatformMapping.Platform
		if platform == "" {
			platform = v1.RUNNER_PLATFORM
		}
		configMap[v1.ConfigKeyPlatformKey{Key: config.Key, Platform: platform}] = true
	}
	return configMap
}

// GetUnitSuffix loosely typed method to get the unit suffix using the unitKey type
func GetUnitSuffix(unitKey v1.ConfigKeyStr, unitStr string) unitsBean.UnitType {
	switch unitKey {
	case v1.CPU_LIMIT, v1.CPU_REQUEST:
		return unitsBean.CPUUnitStr(unitStr).GetUnitSuffix()
	case v1.MEMORY_LIMIT, v1.MEMORY_REQUEST:
		return unitsBean.MemoryUnitStr(unitStr).GetUnitSuffix()
	case v1.TIME_OUT:
		return unitsBean.TimeUnitStr(unitStr).GetUnitSuffix()
	case v1.TOLERATIONS, v1.NODE_SELECTOR, v1.SECRET, v1.CONFIG_MAP:
		return unitsBean.NoUnitStr(unitStr).GetUnitSuffix()
	default:
		return unitsBean.NoUnitStr(unitStr).GetUnitSuffix()
	}
}

// GetUnitSuffixStr loosely typed method to get the unit suffix using the unitKey type
func GetUnitSuffixStr(unitKey v1.ConfigKey, unit unitsBean.UnitType) string {
	switch unitKey {
	case v1.CPULimitKey, v1.CPURequestKey:
		return unit.GetCPUUnitStr().String()
	case v1.MemoryLimitKey, v1.MemoryRequestKey:
		return unit.GetMemoryUnitStr().String()
	case v1.TimeOutKey:
		return unit.GetTimeUnitStr().String()
	case v1.TolerationsKey, v1.NodeSelectorKey, v1.SecretKey, v1.ConfigMapKey:
		return unit.GetNoUnitStr().String()
	}
	return unit.GetNoUnitStr().String()
}

// GetDefaultConfigKeysMapV0 returns a map of default config keys
func GetDefaultConfigKeysMapV0() map[v1.ConfigKeyStr]bool {
	return map[v1.ConfigKeyStr]bool{
		v1.CPU_LIMIT:      true,
		v1.CPU_REQUEST:    true,
		v1.MEMORY_LIMIT:   true,
		v1.MEMORY_REQUEST: true,
		v1.TIME_OUT:       true,
		// v1.NODE_SELECTOR is added in V1, but maintained for backward compatibility
		v1.NODE_SELECTOR: true,
		// v1.TOLERATIONS is added in V1, but maintained for backward compatibility
		v1.TOLERATIONS: true,
	}
}

// GetConfigKeysMapForPlatform returns a map of config keys supported for a given platform
func GetConfigKeysMapForPlatform(platform string) v1.InfraConfigKeys {
	defaultConfigKeys := map[v1.ConfigKeyStr]bool{
		v1.CPU_LIMIT:      true,
		v1.CPU_REQUEST:    true,
		v1.MEMORY_LIMIT:   true,
		v1.MEMORY_REQUEST: true,
	}
	if platform == v1.RUNNER_PLATFORM {
		defaultConfigKeys[v1.TIME_OUT] = true
	}
	return getConfigKeysMapForPlatformEnt(defaultConfigKeys, platform)
}

func GetMandatoryConfigKeys(profileName, platformName string) []v1.ConfigKeyStr {
	if profileName == v1.GLOBAL_PROFILE_NAME {
		return GetConfigKeysMapForPlatform(platformName).GetAllSupportedKeys()
	}
	return make([]v1.ConfigKeyStr, 0)
}

func IsAnyRequiredConfigMissing(profileName, platformName string, configuredKeys v1.InfraConfigKeys) bool {
	mandatoryKeys := GetMandatoryConfigKeys(profileName, platformName)
	for _, missingKey := range configuredKeys.GetUnConfiguredKeys() {
		if slices.Contains(mandatoryKeys, missingKey) {
			return true
		}
	}
	return false
}

func GetConfigCompositeKey(config *repository.InfraProfileConfigurationEntity) string {
	return fmt.Sprintf("%s|%s", GetConfigKeyStr(config.Key), config.UniqueId)
}

func GetConfigKeyStr(configKey v1.ConfigKey) v1.ConfigKeyStr {
	switch configKey {
	case v1.CPULimitKey:
		return v1.CPU_LIMIT
	case v1.CPURequestKey:
		return v1.CPU_REQUEST
	case v1.MemoryLimitKey:
		return v1.MEMORY_LIMIT
	case v1.MemoryRequestKey:
		return v1.MEMORY_REQUEST
	case v1.TimeOutKey:
		return v1.TIME_OUT
	}
	return getEntConfigKeyStr(configKey)
}

func GetConfigKey(configKeyStr v1.ConfigKeyStr) v1.ConfigKey {
	switch configKeyStr {
	case v1.CPU_LIMIT:
		return v1.CPULimitKey
	case v1.CPU_REQUEST:
		return v1.CPURequestKey
	case v1.MEMORY_LIMIT:
		return v1.MemoryLimitKey
	case v1.MEMORY_REQUEST:
		return v1.MemoryRequestKey
	case v1.TIME_OUT:
		return v1.TimeOutKey
	}
	return getEntConfigKey(configKeyStr)
}

// ValidatePayloadConfig - validates the payload configuration
func ValidatePayloadConfig(profileToUpdate *v1.ProfileBeanDto) error {
	if len(profileToUpdate.GetName()) == 0 {
		errMsg := "profile name is required"
		return globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	err := validateProfileAttributes(profileToUpdate.ProfileBeanAbstract)
	if err != nil {
		return err
	}
	for platform, config := range profileToUpdate.GetConfigurations() {
		supportedConfigKeys := GetConfigKeysMapForPlatform(platform)
		err = validatePlatformName(platform, profileToUpdate.GetBuildxDriverType())
		if err != nil {
			return err
		}
		err = validateConfigItems(config, supportedConfigKeys)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateConfigItems(propertyConfigs []*v1.ConfigurationBean, supportedConfigKeys v1.InfraConfigKeys) error {
	var validationErrors []string
	for _, config := range propertyConfigs {
		if !supportedConfigKeys.IsSupported(config.Key) {
			validationErrors = append(validationErrors, fmt.Sprintf("invalid configuration property %q", config.Key))
			continue
		}
	}
	// If any validation errors were found, return them as a single error
	if len(validationErrors) > 0 {
		errMsg := fmt.Sprintf("validation errors: %s", strings.Join(validationErrors, "; "))
		return globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	return nil
}

func validateProfileAttributes(profileAbstract v1.ProfileBeanAbstract) error {
	if len(profileAbstract.GetName()) > v1.QualifiedProfileMaxLength {
		errMsg := "profile name is too long"
		return globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	if len(profileAbstract.GetName()) == 0 {
		errMsg := "profile name is empty"
		return globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	if len(profileAbstract.GetDescription()) > v1.QualifiedDescriptionMaxLength {
		errMsg := "profile description is too long"
		return globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	if !profileAbstract.GetBuildxDriverType().IsValid() {
		errMsg := fmt.Sprintf("invalid buildx driver type: %q", profileAbstract.GetBuildxDriverType())
		return globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	return nil
}
