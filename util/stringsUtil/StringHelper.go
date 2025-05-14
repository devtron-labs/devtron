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

package stringsUtil

import (
	"fmt"
	"strconv"
	"strings"
)

func GetCommaSeparatedStringsFromIntArray(vals []int) string {
	res := ""
	for i, val := range vals {
		if i == 0 {
			res = fmt.Sprintf("%d", val)
		} else {
			res = fmt.Sprintf("%s,%d", res, val)
		}
	}
	return res
}

func ParseBool(value string) (bool, error) {
	// remove leading/ trailing quotes if present
	value = strings.Trim(value, "\"")
	boolValue, parseErr := strconv.ParseBool(value)
	if parseErr != nil {
		return false, parseErr
	}
	return boolValue, nil
}

func GetSpaceTrimmedUniqueString(sliceList []string) []string {
	if len(sliceList) == 0 {
		return sliceList
	}
	allKeys := make(map[string]bool)
	var list []string
	for _, item := range sliceList {
		if item == "" {
			continue
		}
		item = strings.TrimSpace(item)
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// SplitCommaSeparatedIntValues splits a comma separated string into an array of integers
func SplitCommaSeparatedIntValues(input string) ([]int, error) {
	items := make([]int, 0)
	itemSlices := strings.Split(input, ",")
	for _, envId := range itemSlices {
		id, err := strconv.Atoi(envId)
		if err != nil {
			return items, err
		}
		items = append(items, id)
	}
	return items, nil
}
