package util

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

type resourceParser struct {
	name        string
	pattern     string
	regex       *regexp.Regexp
	conversions map[string]float64
}

var memoryParser *resourceParser
var cpuParser *resourceParser

func MemoryToNumber(memory string) (float64, error) {
	if memoryParser == nil {
		pattern := "(\\d*e?\\d*)(Ei?|Pi?|Ti?|Gi?|Mi?|Ki?|$)"
		re, _ := regexp.Compile(pattern)
		memoryParser = &resourceParser{
			name:    "memory",
			pattern: pattern,
			regex:   re,
			conversions: map[string]float64{
				"E":  float64(1000000000000000000),
				"P":  float64(1000000000000000),
				"T":  float64(1000000000000),
				"G":  float64(1000000000),
				"M":  float64(1000000),
				"K":  float64(1000),
				"Ei": float64(1152921504606846976),
				"Pi": float64(1125899906842624),
				"Ti": float64(1099511627776),
				"Gi": float64(1073741824),
				"Mi": float64(1048576),
				"Ki": float64(1024),
			},
		}
	}
	return convertResource(memoryParser, memory)
}
func CpuToNumber(cpu string) (float64, error) {
	demo := NoCpuUnitChecker.MatchString(cpu)
	if demo {
		return strconv.ParseFloat(cpu, 64)
	}
	if cpuParser == nil {
		pattern := "(\\d*e?\\d*)(m?)"
		re, _ := regexp.Compile(pattern)
		cpuParser = &resourceParser{
			name:    "cpu",
			pattern: pattern,
			regex:   re,
			conversions: map[string]float64{
				"m": .001,
			},
		}
	}
	return convertResource(cpuParser, cpu)
}
func convertResource(rp *resourceParser, resource string) (float64, error) {
	matches := rp.regex.FindAllStringSubmatch(resource, -1)
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
	limit, ok := dat["resources"].(map[string]interface{})["limits"].(map[string]interface{})
	if !ok {
		return false, errors.New("resources.limits is required")
	}
	envoproxyLimit, ok := dat["envoyproxy"].(map[string]interface{})["resources"].(map[string]interface{})["limits"].(map[string]interface{})
	if !ok {
		return false, errors.New("envoproxy.resources.limits is required")
	}
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
	request, ok := dat["resources"].(map[string]interface{})["requests"].(map[string]interface{})
	if !ok {
		return true, nil
	}
	envoproxyRequest, ok := dat["envoyproxy"].(map[string]interface{})["resources"].(map[string]interface{})["requests"].(map[string]interface{})
	if !ok {
		return true, nil
	}
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

	if (envoproxyCPULimit < envoproxyCPURequest){
		return false, errors.New("envoyproxy.resources.limits.cpu must be greater than or equal to envoyproxy.resources.requests.cpu")
	}else if (envoproxyMemoryLimit < envoproxyMemoryRequest){
		return false, errors.New("envoyproxy.resources.limits.memory must be greater than or equal to envoyproxy.resources.requests.memory")
	}else if (cpuLimit < cpuRequest) {
		return false, errors.New("resources.limits.cpu must be greater than or equal to resources.requests.cpu")
	}else if (memoryLimit < memoryRequest) {
		return false, errors.New("resources.limits.memory must be greater than or equal to resources.requests.memory")
	}
	return true, nil

}

func AutoScale(dat map[string]interface{}) (bool, error) {
	autoscaleEnabled, ok := dat["autoscaling"].(map[string]interface{})["enabled"]
	if !ok {
		return true, nil
	}
	if autoscaleEnabled.(bool) {
		limit, ok := dat["resources"].(map[string]interface{})["limits"].(map[string]interface{})
		if !ok {
			return false, errors.New("resources.limits is required")
		}
		envoproxyLimit, ok := dat["envoyproxy"].(map[string]interface{})["resources"].(map[string]interface{})["limits"].(map[string]interface{})
		if !ok {
			return false, errors.New("envoproxy.resources.limits is required")
		}
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
		request, ok := dat["resources"].(map[string]interface{})["requests"].(map[string]interface{})
		if !ok {
			return false, errors.New("resources.requests is required")
		}
		envoproxyRequest, ok := dat["envoyproxy"].(map[string]interface{})["resources"].(map[string]interface{})["requests"].(map[string]interface{})
		if !ok {
			return false, errors.New("envoyproxy.resources.requests is required")
		}
		checkCPURequests, ok := request["cpu"]
		if !ok {
			return false, errors.New("resources.requests.cpu is required")
		}
		checkMemoryRequests, ok := request["memory"]
		if !ok {
			return false, errors.New("resources.requests.memory is required")
		}
		checkEnvoproxyCPURequests, ok := envoproxyRequest["cpu"]
		if !ok {
			return false, errors.New("envoyproxy.resources.requests.cpu is required")
		}
		checkEnvoproxyMemoryRequests, ok := envoproxyRequest["memory"]
		if !ok {
			return false, errors.New("envoyproxy.resources.requests.memory is required")
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

		if (envoproxyCPULimit < envoproxyCPURequest){
			return false, errors.New("envoyproxy.resources.limits.cpu must be greater than or equal to envoyproxy.resources.requests.cpu")
		}else if (envoproxyMemoryLimit < envoproxyMemoryRequest){
			return false, errors.New("envoyproxy.resources.limits.memory must be greater than or equal to envoyproxy.resources.requests.memory")
		}else if (cpuLimit < cpuRequest) {
			return false, errors.New("resources.limits.cpu must be greater than or equal to resources.requests.cpu")
		}else if (memoryLimit < memoryRequest) {
			return false, errors.New("resources.limits.memory must be greater than or equal to resources.requests.memory")
		}else {
			return true, nil
		}
	} else {
		return true, nil
	}
}

var (
	CpuUnitChecker, _   = regexp.Compile("^([0-9.]+)m$")
	NoCpuUnitChecker, _ = regexp.Compile("^([0-9.]+)$")
	MiChecker, _        = regexp.Compile("^[0-9]+Mi$")
	GiChecker, _        = regexp.Compile("^[0-9]+Gi$")
	TiChecker, _        = regexp.Compile("^[0-9]+Ti$")
	PiChecker, _        = regexp.Compile("^[0-9]+Pi$")
	KiChecker, _        = regexp.Compile("^[0-9]+Ki$")
)

func (f CpuChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	if CpuUnitChecker.MatchString(asString) {
		return true
	} else if NoCpuUnitChecker.MatchString(asString) {
		return true
	} else {
		return false
	}
}

func (f MemoryChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	if MiChecker.MatchString(asString) {
		return true
	} else if GiChecker.MatchString(asString) {
		return true
	} else if TiChecker.MatchString(asString) {
		return true
	} else if PiChecker.MatchString(asString) {
		return true
	} else if KiChecker.MatchString(asString) {
		return true
	} else {
		return false
	}
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
