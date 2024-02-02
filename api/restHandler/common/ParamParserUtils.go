package common

import (
	"net/http"
	"strconv"
)

const TokenHeaderKey = "token"

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

func convertToInt(w http.ResponseWriter, paramValue string) (int, error) {
	paramIntValue, err := strconv.Atoi(paramValue)
	if err != nil {
		WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return 0, err
	}
	return paramIntValue, nil
}
