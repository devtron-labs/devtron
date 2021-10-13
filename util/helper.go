/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

type resourceParser struct {
	name        string
	pattern     string
	regex       *regexp.Regexp
	conversions map[string]float64
}

var memoryParser *resourceParser
var cpuParser *resourceParser

func ContainsString(list []string, element string) bool {
	if len(list) == 0 {
		return false
	}
	for _, l := range list {
		if l == element {
			return true
		}
	}
	return false
}

func AppendErrorString(errs []string, err error) []string {
	if err != nil {
		errs = append(errs, err.Error())
	}
	return errs
}

func GetErrorOrNil(errs []string) error {
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

func ExtractChartVersion(chartVersion string) (int, int, error) {
	if len(chartVersion) == 0 {
		return 0, 0, nil
	}
	chartVersions := strings.Split(chartVersion, ".")
	chartMajorVersion, err := strconv.Atoi(chartVersions[0])
	if err != nil {
		return 0, 0, err
	}
	chartMinorVersion, err := strconv.Atoi(chartVersions[1])
	if err != nil {
		return 0, 0, err
	}
	return chartMajorVersion, chartMinorVersion, nil
}

type Closer interface {
	Close() error
}

func Close(c Closer, logger *zap.SugaredLogger) {
	if err := c.Close(); err != nil {
		logger.Warnf("failed to close %v: %v", c, err)
	}
}

var chars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func Generate(size int) string {
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	for i := 0; i < size; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()
	return str
}

func HttpRequest(url string) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	//var client *http.Client
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		var apiRes map[string]interface{}
		err = json.Unmarshal(resBody, &apiRes)
		if err != nil {
			return nil, err
		}
		return apiRes, err
	}
	return nil, err
}

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
		fmt.Printf("expected pattern for %s should match %s, found %s\n", rp.name, rp.pattern, resource)
		return float64(0), fmt.Errorf("expected pattern for %s should match %s, found %s", rp.name, rp.pattern, resource)
	}
	num, err := ParseFloat(matches[0][1])
	if err != nil {
		fmt.Println(err)
		return float64(0), err
	}
	if len(matches[0]) == 3 && matches[0][2] != "" {
		if suffix, ok := rp.conversions[matches[0][2]]; ok {
			return num * suffix, nil
		}
	} else {
		return num, nil
	}
	fmt.Printf("expected pattern for %s should match %s, found %s\n", rp.name, rp.pattern, resource)
	return float64(0), fmt.Errorf("expected pattern for %s should match %s, found %s", rp.name, rp.pattern, resource)
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
