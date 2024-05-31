/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package response

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/util"
	"net/http"
)

func WriteResponse(status int, message string, w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	type Response struct {
		Code   int              `json:"code,omitempty"`
		Status string           `json:"status,omitempty"`
		Result interface{}      `json:"result,omitempty"`
		Errors []*util.ApiError `json:"errors,omitempty"`
	}
	response := Response{}
	response.Code = status
	response.Result = message
	b, err := json.Marshal(response)
	if err != nil {
		b = []byte("OK")
		util.GetLogger().Errorw("Unexpected error in apiError", "err", err)
	}
	_, err = w.Write(b)
	if err != nil {
		util.GetLogger().Error(err)
	}
}
