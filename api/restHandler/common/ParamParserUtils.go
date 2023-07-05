package common

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
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

func ExtractIntQueryParam(w http.ResponseWriter, r *http.Request, paramName string) (int, error) {
	queryParams := r.URL.Query()
	paramValue := queryParams.Get(paramName)
	paramIntValue, err := convertToInt(w, paramValue)
	if err != nil {
		return 0, err
	}
	return paramIntValue, nil
}
