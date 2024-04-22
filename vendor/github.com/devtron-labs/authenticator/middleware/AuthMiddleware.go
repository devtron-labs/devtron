/*
 * Copyright (c) 2021 Devtron Labs
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

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/authenticator/oidc"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

const ApiTokenHeaderKey = "api-token"
const tokenHeaderKey = "token"
const argocdTokenHeaderKey = "argocd.token"

// Authorizer is a middleware for authorization
func Authorizer(sessionManager *SessionManager, whitelistChecker func(url string) bool, userStatusCheckInDb func(token string) (bool, int32, error)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			token := ""
			apiToken := r.Header.Get(ApiTokenHeaderKey)
			if len(apiToken) > 0 {
				// for external ci webhook request, will be authorize by api-token
				token = apiToken
			} else {
				authCookieToken, _ := oidc.JoinCookies(oidc.AuthCookieName, r.Cookies())
				if authCookieToken != "" {
					token = authCookieToken
					r.Header.Set(tokenHeaderKey, token)
				}
				if token == "" && authCookieToken == "" {
					token = r.Header.Get(tokenHeaderKey)
				}
			}

			// users = append(users, "anonymous")
			authEnabled := true
			pass := false
			config := GetConfig()
			authEnabled = config.AuthEnabled

			if token != "" && authEnabled && !whitelistChecker(r.URL.Path) {
				_, err := sessionManager.VerifyToken(token)
				if err != nil {
					log.Printf("Error verifying token: %+v\n", err)
					if len(apiToken) == 0 {
						http.SetCookie(w, &http.Cookie{Name: argocdTokenHeaderKey, Value: token, Path: "/", MaxAge: -1})
					}
					writeResponse(http.StatusUnauthorized, "Unauthorized", w, err)
					return
				}
				pass = true

				// this function only supplied in case of enterprise build. handled here for all other case.
				if userStatusCheckInDb != nil {

					// checking user status in db
					isInactive, userId, err := userStatusCheckInDb(token)
					if err != nil {
						writeResponse(http.StatusUnauthorized, "Invalid User", w, err)
						return
					} else if isInactive {
						writeResponse(http.StatusUnauthorized, "Inactive User", w, fmt.Errorf("inactive User"))
						return
					}

					// setting user id in context
					context.WithValue(r.Context(), "userId", userId)
				}
			}
			if pass {
				next.ServeHTTP(w, r)
			} else if whitelistChecker(r.URL.Path) {
				next.ServeHTTP(w, r)
			} else if token == "" {
				writeResponse(http.StatusUnauthorized, "UN-AUTHENTICATED", w, fmt.Errorf("unauthenticated"))
				return
			} else {
				writeResponse(http.StatusForbidden, "FORBIDDEN", w, fmt.Errorf("unauthorized"))
				return
			}
		}

		return http.HandlerFunc(fn)
	}
}

func WhitelistChecker(url string) bool {
	urls := []string{
		"/auth/login",
		"/auth/callback",
		"/user/login",
		"/",
	}
	for _, a := range urls {
		if a == url {
			return true
		}
	}
	prefixUrls := []string{
		"/api/dex/",
	}
	for _, a := range prefixUrls {
		if strings.Contains(url, a) {
			return true
		}
	}
	return false
}

func writeResponse(status int, message string, w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	type Response struct {
		Code   int         `json:"code,omitempty"`
		Status string      `json:"status,omitempty"`
		Result interface{} `json:"result,omitempty"`
	}
	response := Response{}
	response.Code = status
	response.Result = message
	b, err := json.Marshal(response)
	if err != nil {
		b = []byte("OK")
		log.Error("Unexpected error in apiError", "err", err)
	}
	_, err = w.Write(b)
	if err != nil {
		log.Error("error", "err", err)
	}
}
