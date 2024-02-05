package common

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

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
	if len(paramValue) == 0 {
		return *defaultVal, nil
	}
	paramIntValue, err := convertToInt(w, paramValue)
	if err != nil {
		return 0, err
	}
	return paramIntValue, nil
}

func ExtractIntArrayQueryParam(w http.ResponseWriter, r *http.Request, paramName string) ([]int, error) {
	queryParams := r.URL.Query()
	paramValue := queryParams.Get(paramName)
	paramIntValues, err := convertToIntArray(w, paramValue)
	return paramIntValues, err
}
