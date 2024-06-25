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

package proxy

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const Dashboard = "dashboard"
const Proxy = "proxy"

func NewDashboardHTTPReverseProxy(serverAddr string, transport http.RoundTripper) (func(writer http.ResponseWriter, request *http.Request), error) {
	proxy, err := GetProxyServer(serverAddr, transport, Dashboard)
	if err != nil {
		return nil, err
	}
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}, nil
}

func GetProxyServer(serverAddr string, transport http.RoundTripper, pathToExclude string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(serverAddr)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport
	proxy.Director = func(request *http.Request) {
		path := request.URL.Path
		request.URL.Host = target.Host
		request.URL.Scheme = target.Scheme
		request.URL.Path = rewriteRequestUrl(path, pathToExclude)
		fmt.Printf("%s\n", request.URL.Path)
	}
	return proxy, nil
}

func rewriteRequestUrl(path string, pathToExclude string) string {
	parts := strings.Split(path, "/")
	var finalParts []string
	for _, part := range parts {
		if part == pathToExclude {
			continue
		}
		finalParts = append(finalParts, part)
	}
	return strings.Join(finalParts, "/")
}

func NewHTTPReverseProxy(serverAddr string, transport http.RoundTripper, enforcer casbin.Enforcer) (func(writer http.ResponseWriter, request *http.Request), error) {
	proxy, err := GetProxyServer(serverAddr, transport, Proxy)
	if err != nil {
		return nil, err
	}
	return func(w http.ResponseWriter, r *http.Request) {

		token := r.Header.Get("token")
		if ok := enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, nil, "Unauthorized User", http.StatusForbidden)
			return
		}
		proxy.ServeHTTP(w, r)
	}, nil
}
