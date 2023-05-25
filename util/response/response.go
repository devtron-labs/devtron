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
