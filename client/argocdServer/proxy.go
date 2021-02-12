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
	proxy.Director = func(request *http.Request) {
		fmt.Printf("path:%s\n", request.URL.Path)
		fmt.Printf("scheme:%s\n", request.URL.Scheme)
		fmt.Printf("host:%s\n", request.URL.Host)
		path := request.URL.Path
		request.URL.Host = target.Host
		request.URL.Scheme = target.Scheme
		request.URL.Path = rewriteRequestUrl(path)
		fmt.Printf("path:%s\n", request.URL.Path)
		fmt.Printf("scheme:%s\n", request.URL.Scheme)
		fmt.Printf("host:%s\n", request.URL.Host)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		log.Printf("reverse proxy called for %s\n", resp.Request.URL.Path)
		log.Printf("reverse proxy called for %s\n", resp.Status)
		if strings.Contains(resp.Request.URL.Path, "/auth/callback") {
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
						redirectUrl := resp.Header.Get("Location")
						if strings.Contains(redirectUrl, "dashboard") {
							strings.ReplaceAll(redirectUrl, "/orchestrator", "")
						}
					}
				}
			}
		} else if strings.Contains(resp.Request.URL.Path, "/auth/login") {
			location := resp.Header.Get("Location")
			if len(location) > 0 {
				newLocation, err := modifyLocation(location)
				if err == nil {
					log.Printf("error parsing url %s, err: %v\n", location, err)
					resp.Header.Set("Location", newLocation)
				}
			}
		}
		log.Printf("response header Location:%s\n", resp.Header.Get("Location"))
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
	//proxy.Director = func(request *http.Request) {
	//	path := request.URL.Path
	//	request.URL.Path = rewriteRequestUrl(path)
	//}
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

func rewriteRequestUrl(path string) string {
	parts := strings.Split(path, "/")
	var finalParts []string
	for _, part := range parts {
		if part == "orchestrator" {
			continue
		}
		finalParts = append(finalParts, part)
	}
	return strings.Join(finalParts, "/")
}

func modifyLocation(location string) (string, error) {
	parsedLocation, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	values, err := url.ParseQuery(parsedLocation.RawQuery)
	if err != nil {
		return "", err
	}
	var redirectUrl string
	for key, value := range values {
		if key == "redirect_uri" {
			redirect, err := url.Parse(value[0])
			if err != nil {
				return "", err
			}
			path := redirect.Path
			if !strings.Contains(path, "/orchestrator") {
				path = "/orchestrator" + path
			}
			redirect.Path = path
			redirectUrl = redirect.String()
			fmt.Printf("redirect url %s\n", redirect.String())
		}
	}
	//values.Del("redirect_uri")
	values.Set("redirect_uri", redirectUrl)
	parsedLocation.RawQuery = values.Encode()
	fmt.Printf("return url: %s\n", parsedLocation.String())
	return parsedLocation.String(), nil
}
