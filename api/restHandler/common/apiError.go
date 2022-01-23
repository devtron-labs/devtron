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

package common

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/hashicorp/go-multierror"
	"github.com/juju/errors"
	"gopkg.in/go-playground/validator.v9"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"strconv"
)

//use of writeJsonRespStructured is preferable. it api exists due to historical reason
// err.message is used as internal message for ApiError object in resp
func WriteJsonResp(w http.ResponseWriter, err error, respBody interface{}, status int) {
	response := Response{}
	if err == nil {
		response.Result = respBody
	} else if apiErr, ok := err.(*util.ApiError); ok {
		response.Errors = []*util.ApiError{apiErr}
		if apiErr.HttpStatusCode != 0 {
			status = apiErr.HttpStatusCode
		}
	} else if validationErrs, ok := err.(validator.ValidationErrors); ok {
		var valErrors []*util.ApiError
		for _, validationErr := range validationErrs {
			//validationErr
			valErr := &util.ApiError{
				UserMessage:     fmt.Sprint(validationErr),
				InternalMessage: fmt.Sprint(validationErr),
			}
			valErrors = append(valErrors, valErr)
		}
		response.Errors = valErrors
	} else if util.IsErrNoRows(err) {
		status = http.StatusNotFound
		apiErr := &util.ApiError{}
		apiErr.Code = "000" // 000=unknown
		apiErr.InternalMessage = errors.Details(err)
		if respBody != nil {
			apiErr.UserMessage = respBody
		} else {
			apiErr.UserMessage = err.Error()
		}
		response.Errors = []*util.ApiError{apiErr}
	} else if multiErr, ok := err.(*multierror.Error); ok {
		var errorsResp []*util.ApiError
		for _, e := range multiErr.Errors {
			errorResp := &util.ApiError{
				UserMessage:     e.Error(),
				InternalMessage: e.Error(),
			}
			errorsResp = append(errorsResp, errorResp)
		}
		response.Errors = errorsResp
	} else if errStatus, ok := err.(*errors2.StatusError);ok{
		apiErr := &util.ApiError{}
		apiErr.Code = strconv.Itoa(int(errStatus.ErrStatus.Code))
		apiErr.InternalMessage = errStatus.Error()
		apiErr.UserMessage = errStatus.Error()
		response.Errors = []*util.ApiError{apiErr}
	} else{
		apiErr := &util.ApiError{}
		apiErr.Code = "000" // 000=unknown
		apiErr.InternalMessage = errors.Details(err)
		if respBody != nil {
			apiErr.UserMessage = respBody
		} else {
			apiErr.UserMessage = err.Error()
		}
		response.Errors = []*util.ApiError{apiErr}
	}
	response.Code = status //TODO : discuss with prashant about http status header
	response.Status = http.StatusText(status)

	b, err := json.Marshal(response)
	if err != nil {
		util.GetLogger().Errorw("error in marshaling err object", "err", err)
		status = 500

		response := Response{}
		apiErr := &util.ApiError{}
		apiErr.Code = "0000" // 000=unknown
		apiErr.InternalMessage = errors.Details(err)
		apiErr.UserMessage = "response marshaling error"
		response.Errors = []*util.ApiError{apiErr}
		b, err = json.Marshal(response)
		if err != nil {
			b = []byte("response marshaling error")
			util.GetLogger().Errorw("Unexpected error in apiError", "err", err)
		}
	}
	if status > 299 || err != nil {
		util.GetLogger().Infow("ERROR RES", "TYPE", "API-ERROR", "RES", response.Code, "ERROR-MSG", response.Errors, "err", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(b)
	if err != nil {
		util.GetLogger().Error(err)
	}

}

//use this method when we have specific api error to be conveyed to api User
func writeJsonRespStructured(w http.ResponseWriter, err error, respBody interface{}, status int, apiErrors []*util.ApiError) {
	response := Response{}
	response.Code = status
	response.Status = http.StatusText(status)
	if err == nil {
		response.Result = respBody
	} else {
		response.Errors = apiErrors
	}
	b, err := json.Marshal(response)
	if err != nil {
		util.GetLogger().Error("error in marshaling err object", err)
		status = 500
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(b)
	if err != nil {
		util.GetLogger().Error(err)
	}
}

//global response body used across api
type Response struct {
	Code   int              `json:"code,omitempty"`
	Status string           `json:"status,omitempty"`
	Result interface{}      `json:"result,omitempty"`
	Errors []*util.ApiError `json:"errors,omitempty"`
}

func contains(s []*string, e *string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
