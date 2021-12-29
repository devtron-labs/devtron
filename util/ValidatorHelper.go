package util

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/xeipuuv/gojsonschema"
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

func convertResource(rp *resourceParser, resource string) (float64, error) {
	matches := rp.regex.FindAllStringSubmatch(resource, -1)
	if len(matches) == 0 {
		return float64(0), errors.New("expected pattern for" + rp.name + "should match" + rp.pattern + ", found " + resource)
	}
	if len(matches[0]) < 2 {
		return float64(0), errors.New("expected pattern for" + rp.name + "should match" + rp.pattern + ", found " + resource)
	}
	num, err := ParseFloat(matches[0][1])
	if err != nil {
		return float64(0), err
	}
	if len(matches[0]) == 3 && matches[0][2] != "" {
		if suffix, ok := rp.conversions[matches[0][2]]; ok {
			return num * suffix, nil
		}
	} else {
		return num, nil
	}
	return float64(0), errors.New("expected pattern for" + rp.name + "should match" + rp.pattern + ", found " + resource)
}

func ParseFloat(str string) (float64, error) {
	val, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return val, nil
	}

	//Some number may be seperated by comma, for example, 23,120,123, so remove the comma firstly
	str = strings.Replace(str, ",", "", -1)

	//Some number is specifed in scientific notation
	pos := strings.IndexAny(str, "eE")
	if pos < 0 {
		return strconv.ParseFloat(str, 64)
	}

	var baseVal float64
	var expVal int64

	baseStr := str[0:pos]
	baseVal, err = strconv.ParseFloat(baseStr, 64)
	if err != nil {

		return 0, err
	}

	expStr := str[(pos + 1):]
	expVal, err = strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return baseVal * math.Pow10(int(expVal)), nil
}

func CompareLimitsRequests(dat map[string]interface{}) (bool, error) {
	if dat == nil {
		return true, nil
	}
	limit := validateAndBuildResourcesAssignment(dat, getResourcesLimitsKeys(false))
	envoproxyLimit := validateAndBuildResourcesAssignment(dat, getResourcesLimitsKeys(true))
	checkCPUlimit, ok := limit["cpu"]
	if !ok {
		return false, errors.New("resources.limits.cpu is required")
	}
	checkMemorylimit, ok := limit["memory"]
	if !ok {
		return false, errors.New("resources.limits.memory is required")
	}
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
	checkCPURequests, ok := request["cpu"]
	if !ok {
		return true, nil
	}
	checkMemoryRequests, ok := request["memory"]
	if !ok {
		return true, nil
	}
	checkEnvoproxyCPURequests, ok := envoproxyRequest["cpu"]
	if !ok {
		return true, nil
	}
	checkEnvoproxyMemoryRequests, ok := envoproxyRequest["memory"]
	if !ok {
		return true, nil
	}

	cpuLimit, err := CpuToNumber(checkCPUlimit.(string))
	if err != nil {
		return false, err
	}
	memoryLimit, err := MemoryToNumber(checkMemorylimit.(string))
	if err != nil {
		return false, err
	}
	cpuRequest, err := CpuToNumber(checkCPURequests.(string))
	if err != nil {
		return false, err
	}
	memoryRequest, err := MemoryToNumber(checkMemoryRequests.(string))
	if err != nil {
		return false, err
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
	} else if cpuLimit < cpuRequest && cpuLimit != 0 {
		return false, errors.New("resources.limits.cpu must be greater than or equal to resources.requests.cpu")
	} else if memoryLimit < memoryRequest && memoryLimit != 0 {
		return false, errors.New("resources.limits.memory must be greater than or equal to resources.requests.memory")
	}
	return true, nil

}

func AutoScale(dat map[string]interface{}) (bool, error) {
	if dat == nil {
		return true, nil
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
