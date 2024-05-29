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

func convertToBool(w http.ResponseWriter, paramValue string) (bool, error) {
	paramBoolValue, err := strconv.ParseBool(paramValue)
	if err != nil {
		WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return false, err
	}
	return paramBoolValue, nil
}

func convertToIntArray(w http.ResponseWriter, paramValue string) ([]int, error) {
	var paramValues []int
	splittedParamValues := strings.Split(paramValue, ",")
	for _, splittedParamValue := range splittedParamValues {
		paramIntValue, err := strconv.Atoi(splittedParamValue)
		if err != nil {
			WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return paramValues, err
		}
		paramValues = append(paramValues, paramIntValue)
	}
	return paramValues, nil
}

func ExtractIntQueryParam(w http.ResponseWriter, r *http.Request, paramName string, defaultVal *int) (int, error) {
	queryParams := r.URL.Query()
	paramValue := queryParams.Get(paramName)
	if len(paramValue) == 0 && defaultVal != nil {
		return *defaultVal, nil
	}
	paramIntValue, err := convertToInt(w, paramValue)
	if err != nil {
		return 0, err
	}
	return paramIntValue, nil
}

func ExtractBooleanQueryParam(w http.ResponseWriter, r *http.Request, paramName string, defaultVal bool) (bool, error) {
	queryParams := r.URL.Query()
	paramValue := queryParams.Get(paramName)
	if len(paramValue) == 0 {
		return defaultVal, nil
	}
	paramBooleanValue, err := convertToBool(w, paramValue)
	if err != nil {
		return false, err
	}
	return paramBooleanValue, nil
}

func ExtractIntArrayQueryParam(w http.ResponseWriter, r *http.Request, paramName string) ([]int, error) {
	queryParams := r.URL.Query()
	paramValue := queryParams.Get(paramName)
	paramIntValues, err := convertToIntArray(w, paramValue)
	return paramIntValues, err
}
