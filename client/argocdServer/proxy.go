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

package argocdServer

import (
	"bytes"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var messageRe = regexp.MustCompile(`<p>(.*)([\s\S]*?)<\/p>`)

func NewCDHTTPReverseProxy(serverAddr string, transport http.RoundTripper, userVerifier func(token string) (int32, error)) func(writer http.ResponseWriter, request *http.Request) {
	target, err := url.Parse(serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport
	proxy.ModifyResponse = func(resp *http.Response) error {
		log.Printf("reverse proxy called for %s\n", resp.Request.URL.Path)
		log.Printf("reverse proxy called for %s\n", resp.Status)
		if resp.Request.URL.Path == "/auth/callback" {
			cookies := resp.Cookies()
			for _, cookie := range cookies {
				if cookie.Name == "argocd.token" {
					userId, err := userVerifier(cookie.Value)
					if err != nil || userId == 0 {
						//no user found remove
						resp.Header.Set("Set-Cookie", "")
						resp.Header.Set("Location", "/dashboard/login?err=NO_USER")
					} else {
						flags := []string{"path=/"}
						components := []string{
							fmt.Sprintf("%s=%s", "argocd.token", cookie.Value),
						}
						components = append(components, flags...)
						header := strings.Join(components, "; ")
						resp.Header.Set("Set-Cookie", header)
					}
				}
			}
		}
		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

// NewDexHTTPReverseProxy returns a reverse proxy to the Dex server. Dex is assumed to be configured
// with the external issuer URL muxed to the same path configured in server.go. In other words, if
// Argo CD API server wants to proxy requests at /api/dex, then the dex config yaml issuer URL should
// also be /api/dex (e.g. issuer: https://argocd.example.com/api/dex)
func NewDexHTTPReverseProxy(serverAddr string, transport http.RoundTripper) func(writer http.ResponseWriter, request *http.Request) {
	target, err := url.Parse(serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport
	proxy.ModifyResponse = func(resp *http.Response) error {
		log.Printf("reverse proxy called for %s\n", resp.Request.URL.Path)
		log.Printf("reverse proxy called for %s\n", resp.Status)
		if resp.StatusCode == 500 {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			err = resp.Body.Close()
			if err != nil {
				return err
			}
			var message string
			matches := messageRe.FindSubmatch(b)
			if len(matches) > 1 {
				message = html.UnescapeString(string(matches[1]))
			} else {
				message = "Unknown error"
			}
			resp.ContentLength = 0
			resp.Header.Set("Content-Length", strconv.Itoa(0))
			resp.Header.Set("Location", fmt.Sprintf("/dashboard/login?sso_error=%s", url.QueryEscape(message)))
			resp.StatusCode = http.StatusSeeOther
			resp.Body = ioutil.NopCloser(bytes.NewReader(make([]byte, 0)))
			return nil
		}
		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}
