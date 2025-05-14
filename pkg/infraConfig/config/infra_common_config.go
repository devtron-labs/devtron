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

package config

import (
	internalUtil "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/infraConfig/util"
	globalUtil "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"k8s.io/apimachinery/pkg/api/resource"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type nativeValueKind interface {
	float64
}

func validLimitRequestForCPUorMem(lim, limFactor, req, reqFactor float64) bool {
	// this condition should be true for the valid case => (lim/req)*(lf/rf) >= 1
	limitToReqRatio := lim / req
	convFactor := limFactor / reqFactor
	return limitToReqRatio*convFactor >= 1
}

// parseCPUorMemoryValue parses the quantity that has number values string and returns the value and unitType
// returns error if parsing fails
func parseCPUorMemoryValue[T units.UnitStrService](quantity string) (float64, T, error) {
	var unitType T
	positive, _, num, denom, suffix, err := parseQuantityString(quantity)
	if err != nil {
		return 0, unitType, err
	}
	unitType = T(suffix)
	if !positive {
		errMsg := "negative value not allowed for cpu limits"
		return 0, unitType, internalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	valStr := num
	if denom != "" {
		valStr = num + "." + denom
	}

	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return val, unitType, err
	}
	// currently we are not supporting exponential values upto 2 decimals
	val = globalUtil.TruncateFloat(val, 2)
	return val, unitType, nil
}

// parseQuantityString is a fast scanner for quantity values.
// this parsing is only for cpu and mem resources
func parseQuantityString(str string) (positive bool, value, num, denom, suffix string, err error) {
	positive = true
	pos := 0
	end := len(str)

	// handle leading sign
	if pos < end {
		switch str[0] {
		case '-':
			positive = false
			pos++
		case '+':
			pos++
		}
	}

	// strip leading zeros
Zeroes:
	for i := pos; ; i++ {
		if i >= end {
			num = "0"
			value = num
			return
		}
		switch str[i] {
		case '0':
			pos++
		default:
			break Zeroes
		}
	}

	// extract the numerator
Num:
	for i := pos; ; i++ {
		if i >= end {
			num = str[pos:end]
			value = str[0:end]
			return
		}
		switch str[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			num = str[pos:i]
			pos = i
			break Num
		}
	}

	// if we stripped all numerator positions, always return 0
	if len(num) == 0 {
		num = "0"
	}

	// handle a denominator
	if pos < end && str[pos] == '.' {
		pos++
	Denom:
		for i := pos; ; i++ {
			if i >= end {
				denom = str[pos:end]
				value = str[0:end]
				return
			}
			switch str[i] {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			default:
				denom = str[pos:i]
				pos = i
				break Denom
			}
		}
		// TODO: we currently allow 1.G, but we may not want to in the future.
		// if len(denom) == 0 {
		// 	err = ErrFormatWrong
		// 	return
		// }
	}
	value = str[0:pos]

	// grab the elements of the suffix
	suffixStart := pos
	for i := pos; ; i++ {
		if i >= end {
			suffix = str[suffixStart:end]
			return
		}
		if !strings.ContainsAny(str[i:i+1], "eEinumkKMGTP") {
			pos = i
			break
		}
	}
	if pos < end {
		switch str[pos] {
		case '-', '+':
			pos++
		}
	}
Suffix:
	for i := pos; ; i++ {
		if i >= end {
			suffix = str[suffixStart:end]
			return
		}
		switch str[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			break Suffix
		}
	}
	// we encountered a non decimal in the Suffix loop, but the last character
	// was not a valid exponent
	err = resource.ErrFormatWrong
	return
}

func filterSupportedConfigurations(platformName string, defaultConfigurations []*v1.ConfigurationBean) []*v1.ConfigurationBean {
	// If not, fallback to the default platform configurations
	filteredConfigurationBean := make([]*v1.ConfigurationBean, 0)
	if len(defaultConfigurations) == 0 {
		return filteredConfigurationBean
	}
	// Get the supported configuration keys for the platform
	supportedConfigKeys := util.GetConfigKeysMapForPlatform(platformName)
	for _, configuration := range defaultConfigurations {
		if supportedConfigKeys.IsSupported(configuration.Key) {
			filteredConfigurationBean = append(filteredConfigurationBean, configuration)
		}
		// Skip the not supported configurations for the platform
	}
	return filteredConfigurationBean
}

func getInheritedConfigurations[T nativeValueKind](defaultConfigBeanAbstract v1.ConfigurationBeanAbstract, profileData, defaultData T,
	profileConfigBean, defaultConfigBean *v1.ConfigurationBean) (*v1.ConfigurationBean, error) {
	if !reflect.ValueOf(profileData).IsZero() {
		// CASE 1: If data is found in profileBean, but the globalProfile does not have any data,
		// then the merged data configuration will be the profile configuration.
		return getAppliedConfigurationBean(profileConfigBean, profileData,
			v1.OVERRIDDEN, sliceUtil.GetSliceOf(profileConfigBean.Id)), nil
	} else if !reflect.ValueOf(defaultData).IsZero() {
		// CASE 2: If data is not found in profileBean,
		// then the merged data configuration will be the global profile.
		return getAppliedConfigurationBean(defaultConfigBean, defaultData,
			v1.INHERITING_GLOBAL_PROFILE, sliceUtil.GetSliceOf(defaultConfigBean.Id)), nil
	} else {
		// CASE 3: If data is not found in profileBean as well as in global profile,
		// then the merged data configuration will be the profile configuration.
		if profileConfigBean != nil && profileConfigBean.Active {
			// OVERRIDDEN is used, for a missing configuration case if the profile configuration is active.
			//	- the configuration is active in the profile.
			//	- whenever the default gets updated, it will have no impact on the profile.
			//	- only when the profile configuration is updated, it will change.
			return getMissingConfigurationBean[T](defaultConfigBeanAbstract, profileConfigBean,
				v1.OVERRIDDEN, sliceUtil.GetSliceOf(profileConfigBean.Id)), nil
		} else {
			// INHERITING_GLOBAL_PROFILE is used, for a missing configuration case if the profile configuration is not active.
			//	- the configuration is not active in the profile.
			//	- whenever the default gets updated, it will be inherited by the profile.
			var appliedConfigIds []int
			if defaultConfigBean != nil {
				appliedConfigIds = sliceUtil.GetSliceOf(defaultConfigBean.Id)
			}
			return getMissingConfigurationBean[T](defaultConfigBeanAbstract, defaultConfigBean,
				v1.INHERITING_GLOBAL_PROFILE, appliedConfigIds), nil
		}
	}
}

func getAppliedConfigurationBean[T nativeValueKind](configBean *v1.ConfigurationBean, value T,
	stateType v1.ConfigStateType, appliedConfigIds []int) *v1.ConfigurationBean {
	// make a copy of the data and update the value
	newConfigBean := configBean.DeepCopy()
	newConfigBean.Value = value
	if reflect.ValueOf(value).IsZero() {
		newConfigBean.Count = 0
	} else {
		newConfigBean.Count = 1
	}
	newConfigBean.ConfigState = stateType
	newConfigBean.AppliedConfigIds = sliceUtil.GetUniqueElements(appliedConfigIds)
	return newConfigBean
}

func getMissingConfigurationBean[T nativeValueKind](defaultConfigBeanAbstract v1.ConfigurationBeanAbstract,
	defaultConfigBean *v1.ConfigurationBean, configState v1.ConfigStateType, appliedConfigIds []int) *v1.ConfigurationBean {
	if defaultConfigBean == nil {
		defaultConfigBean = &v1.ConfigurationBean{}
		defaultConfigBean.ConfigurationBeanAbstract = defaultConfigBeanAbstract
	}
	var emptyValue T
	newConfigBean := defaultConfigBean.DeepCopy()
	newConfigBean.Value = emptyValue
	newConfigBean.Count = 0
	newConfigBean.ConfigState = configState
	newConfigBean.AppliedConfigIds = sliceUtil.GetUniqueElements(appliedConfigIds)
	return newConfigBean
}
