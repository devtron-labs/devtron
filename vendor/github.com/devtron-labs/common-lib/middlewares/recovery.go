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

package middlewares

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/common-lib/constants"
	"github.com/devtron-labs/common-lib/pubsub-lib/metrics"
	"log"
	"net/http"
	"runtime/debug"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			err := recover()
			if err != nil {
				if errors.Is(err.(error), http.ErrAbortHandler) {
					// suppress logging for http.ErrAbortHandler panic
					// separate metric for reverse proxy panic recovery
					metrics.IncReverseProxyPanicRecoveryCount("proxy", r.Host, r.Method, r.RequestURI)
				} else {
					// log and increment panic recovery count
					metrics.IncPanicRecoveryCount("handler", r.Host, r.Method, r.RequestURI)
					log.Print(constants.PanicLogIdentifier, "recovered from panic", "err", err, "stack", string(debug.Stack()))
				}
				jsonBody, _ := json.Marshal(map[string]string{
					"error": "internal server error",
				})
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(jsonBody)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
