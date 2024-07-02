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

package common

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

const TokenHeaderKey = "token"

func ExtractIntPathParam(w http.ResponseWriter, r *http.Request, paramName string) (int, error) {
	vars := mux.Vars(r)
	paramValue := vars[paramName]
	paramIntValue, err := convertToInt(w, paramValue)
	if err != nil {
		return 0, err
	}
	return paramIntValue, nil
}

func convertToInt(w http.ResponseWriter, paramValue string) (int, error) {
	paramIntValue, err := strconv.Atoi(paramValue)
	if err != nil {
		WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return 0, err
	}
	return paramIntValue, nil
}

func convertToIntArray(paramValue string) ([]int, error) {
	var paramValues []int
	splittedParamValues := strings.Split(paramValue, ",")
	for _, splittedParamValue := range splittedParamValues {
		paramIntValue, err := strconv.Atoi(splittedParamValue)
		if err != nil {
			return paramValues, err
		}
		paramValues = append(paramValues, paramIntValue)
	}
	return paramValues, nil
}

func ExtractIntQueryParam(w http.ResponseWriter, r *http.Request, paramName string, defaultValue int) (int, error) {
	queryParams := r.URL.Query()
	paramValue := queryParams.Get(paramName)
	if len(paramValue) == 0 {
		return defaultValue, nil
	}
	paramIntValue, err := convertToInt(w, paramValue)
	if err != nil {
		return defaultValue, err
	}
	return paramIntValue, nil
}

// ExtractIntArrayQueryParam don't use this func, doesn't handle all cases to capture query params
// use ExtractIntArrayFromQueryParam over this func to capture int array from query param.
func ExtractIntArrayQueryParam(w http.ResponseWriter, r *http.Request, paramName string) ([]int, error) {
	queryParams := r.URL.Query()
	paramValue := queryParams.Get(paramName)
	paramIntValues, err := convertToIntArray(paramValue)
	return paramIntValues, err
}

func ExtractBoolQueryParam(r *http.Request, paramName string) (bool, error) {
	queryParams := r.URL.Query()
	paramValue := queryParams.Get(paramName)
	var boolValue bool
	var err error
	if len(paramValue) > 0 {
		boolValue, err = strconv.ParseBool(paramValue)
		if err != nil {
			return boolValue, err
		}
	}

	return boolValue, nil
}

// ExtractIntArrayFromQueryParam returns list of all ids in []int extracted from query param
// use this method over ExtractIntArrayQueryParam if there is list of query params
func ExtractIntArrayFromQueryParam(r *http.Request, paramName string) ([]int, error) {
	queryParams := r.URL.Query()
	paramValue := queryParams[paramName]
	paramIntValues := make([]int, 0)
	var err error
	if paramValue != nil && len(paramValue) > 0 {
		if strings.Contains(paramValue[0], ",") {
			paramIntValues, err = convertToIntArray(paramValue[0])
		} else {
			paramIntValues, err = convertStringArrayToIntArray(paramValue)
		}
	}

	return paramIntValues, err
}

func convertStringArrayToIntArray(strArr []string) ([]int, error) {
	var paramValues []int
	for _, item := range strArr {
		paramIntValue, err := strconv.Atoi(item)
		if err != nil {
			return paramValues, err
		}
		paramValues = append(paramValues, paramIntValue)
	}
	return paramValues, nil
}
