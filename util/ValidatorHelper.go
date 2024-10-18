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
	"errors"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"k8s.io/apimachinery/pkg/api/resource"
	"regexp"
)

const (
	CpuRegex    = "(^\\d*\\.?\\d+e?\\d*)(m?)$"
	MemoryRegex = "(^\\d*\\.?\\d+e?\\d*)(Ei?|Pi?|Ti?|Gi?|Mi?|Ki?|$)$"
)

var (
	CpuUnitChecker, _    = regexp.Compile(CpuRegex)
	MemoryUnitChecker, _ = regexp.Compile(MemoryRegex)
)

type resourceParser struct {
	name        string
	pattern     string
	regex       *regexp.Regexp
	conversions map[string]float64
}

var memoryParser *resourceParser
var cpuParser *resourceParser

func getResourcesLimitsKeys(envoyProxy bool) []string {
	if envoyProxy {
		return []string{"envoyproxy", "resources", "limits"}
	} else {
		return []string{"resources", "limits"}
	}
}
func getResourcesRequestsKeys(envoyProxy bool) []string {
	if envoyProxy {
		return []string{"envoyproxy", "resources", "requests"}
	} else {
		return []string{"resources", "requests"}
	}
}

func validateAndBuildResourcesAssignment(dat map[string]interface{}, validationKeys []string) (validatedMap map[string]interface{}) {
	var test map[string]interface{}
	test = dat
	for _, validationKey := range validationKeys {
		if test[validationKey] != nil {
			test = test[validationKey].(map[string]interface{})
		} else {
			return map[string]interface{}{"cpu": "0", "memory": "0"}
		}
	}
	return test
}

func MemoryToNumber(memory string) (int64, error) {
	quantity, err := resource.ParseQuantity(memory)
	if err != nil {
		return 0, err
	}
	if quantity.Value() < 0 {
		return 0, fmt.Errorf("value cannot be negative")
	}
	return quantity.Value(), nil
}

func CpuToNumber(cpu string) (int64, error) {
	quantity, err := resource.ParseQuantity(cpu)
	if err != nil {
		return 0, err
	}
	if quantity.MilliValue() < 0 {
		return 0, fmt.Errorf("value cannot be negative")
	}
	return quantity.MilliValue(), nil
}

func CompareLimitsRequests(dat map[string]interface{}, chartVersion string) (bool, error) {
	if dat == nil {
		return true, nil
	}
	limit := validateAndBuildResourcesAssignment(dat, getResourcesLimitsKeys(false))
	envoproxyLimit := validateAndBuildResourcesAssignment(dat, getResourcesLimitsKeys(true))
	checkCPUlimit, cpulimitOk := limit["cpu"]
	checkMemorylimit, memoryLimitOk := limit["memory"]
	checkEnvoproxyCPUlimit, ok := envoproxyLimit["cpu"]
	if !ok {
		return false, errors.New("envoyproxy.resources.limits.cpu is required")
	}
	checkEnvoproxyMemorylimit, ok := envoproxyLimit["memory"]
	if !ok {
		return false, errors.New("envoyproxy.resources.limits.memory is required")
	}
	request := validateAndBuildResourcesAssignment(dat, getResourcesRequestsKeys(false))
	envoproxyRequest := validateAndBuildResourcesAssignment(dat, getResourcesRequestsKeys(true))
	checkCPURequests, cpuRequestsOk := request["cpu"]
	checkMemoryRequests, memoryRequestsOk := request["memory"]
	checkEnvoproxyCPURequests, ok := envoproxyRequest["cpu"]
	if !ok {
		return true, nil
	}
	checkEnvoproxyMemoryRequests, ok := envoproxyRequest["memory"]
	if !ok {
		return true, nil
	}
	var cpuLimit int64
	var err error
	if checkCPUlimit != nil && cpulimitOk {
		cpuLimit, err = CpuToNumber(checkCPUlimit.(string))
		if err != nil {
			return false, err
		}
	}
	var memoryLimit int64
	if checkMemorylimit != nil && memoryLimitOk {
		memoryLimit, err = MemoryToNumber(checkMemorylimit.(string))
		if err != nil {
			return false, err
		}
	}
	var cpuRequest int64
	if checkCPURequests != nil && cpuRequestsOk {
		cpuRequest, err = CpuToNumber(checkCPURequests.(string))
		if err != nil {
			return false, err
		}
	}
	var memoryRequest int64
	if checkMemoryRequests != nil && memoryRequestsOk {
		memoryRequest, err = MemoryToNumber(checkMemoryRequests.(string))
		if err != nil {
			return false, err
		}
	}
	envoproxyCPULimit, err := CpuToNumber(checkEnvoproxyCPUlimit.(string))
	if err != nil {
		return false, err
	}
	envoproxyMemoryLimit, err := MemoryToNumber(checkEnvoproxyMemorylimit.(string))
	if err != nil {
		return false, err
	}
	envoproxyCPURequest, err := CpuToNumber(checkEnvoproxyCPURequests.(string))
	if err != nil {
		return false, err
	}
	envoproxyMemoryRequest, err := MemoryToNumber(checkEnvoproxyMemoryRequests.(string))
	if err != nil {
		return false, err
	}
	if envoproxyCPULimit < envoproxyCPURequest && envoproxyCPULimit != 0 {
		return false, errors.New("envoyproxy.resources.limits.cpu must be greater than or equal to envoyproxy.resources.requests.cpu")
	} else if envoproxyMemoryLimit < envoproxyMemoryRequest && envoproxyMemoryLimit != 0 {
		return false, errors.New("envoyproxy.resources.limits.memory must be greater than or equal to envoyproxy.resources.requests.memory")
	} else if cpulimitOk && cpuRequestsOk && cpuLimit < cpuRequest && cpuLimit != 0 {
		return false, errors.New("resources.limits.cpu must be greater than or equal to resources.requests.cpu")
	} else if memoryLimitOk && memoryRequestsOk && memoryLimit < memoryRequest && memoryLimit != 0 {
		return false, errors.New("resources.limits.memory must be greater than or equal to resources.requests.memory")
	}
	return true, nil

}

func AutoScale(dat map[string]interface{}) (bool, error) {
	if dat == nil {
		return true, nil
	}
	kedaAutoScaleEnabled := false
	if dat["kedaAutoscaling"] != nil {
		kedaAutoScale, ok := dat["kedaAutoscaling"].(map[string]interface{})["enabled"]
		if ok {
			kedaAutoScaleEnabled = kedaAutoScale.(bool)
		}
	}
	if dat["autoscaling"] != nil {
		autoScaleEnabled, ok := dat["autoscaling"].(map[string]interface{})["enabled"]
		if !ok {
			return true, nil
		}
		if autoScaleEnabled.(bool) {
			minReplicas, okMin := dat["autoscaling"].(map[string]interface{})["MinReplicas"]
			maxReplicas, okMax := dat["autoscaling"].(map[string]interface{})["MaxReplicas"]
			if !okMin || !okMax {
				return false, errors.New("autoscaling.MinReplicas and autoscaling.MaxReplicas are mandatory fields")
			}
			// see https://pkg.go.dev/encoding/json#Unmarshal for why conversion to float64 and not int
			// Bug fix PR https://github.com/devtron-labs/devtron/pull/884
			if minReplicas.(float64) > maxReplicas.(float64) {
				return false, errors.New("autoscaling.MinReplicas can not be greater than autoscaling.MaxReplicas")
			}
			if kedaAutoScaleEnabled {
				return false, errors.New("autoscaling and kedaAutoscaling can not be enabled at the same time. Use additional scalers in kedaAutoscaling instead.")
			}
		}
	}
	return true, nil
}

func (f CpuChecker) IsFormat(input interface{}) bool {
	if input == nil {
		return false
	}
	asString, ok := input.(string)
	if !ok {
		return false
	}
	quantity, err := resource.ParseQuantity(asString)
	return err == nil && quantity.Value() > 0
}

func (f MemoryChecker) IsFormat(input interface{}) bool {
	if input == nil {
		return false
	}
	asString, ok := input.(string)
	if !ok {
		return false
	}
	quantity, err := resource.ParseQuantity(asString)
	return err == nil && quantity.Value() > 0
}

type (
	CpuChecker    struct{}
	MemoryChecker struct{}
)

type CustomFormatCheckers struct {
}

func (c CustomFormatCheckers) AddCheckers() {
	gojsonschema.FormatCheckers.Add("cpu", CpuChecker{})
	gojsonschema.FormatCheckers.Add("memory", MemoryChecker{})
}

func NewGoJsonSchemaCustomFormatChecker() *CustomFormatCheckers {

	checker := &CustomFormatCheckers{}
	checker.AddCheckers()
	return checker
}
