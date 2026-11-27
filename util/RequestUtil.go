/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package util

import (
	"net"
	"net/http"
	"strings"
)

const (
	xForwardedForHeaderName = "X-Forwarded-For"
	xRealIPHeaderName       = "X-Real-IP"
)

// GetClientIP gets a request's originating client IP address, in order of trust:
// the nearest proxy's own view of the connecting peer (X-Forwarded-For's
// rightmost entry), X-Real-IP, then the raw connection address.
//
// X-Forwarded-For is a comma-separated hop chain; each proxy appends the peer IP
// it directly observed. The leftmost entry is whatever the client sent and is
// not trustworthy on proxies that append rather than replace (e.g. AWS ALB) -
// the rightmost entry is the nearest proxy's own observation, which is what we
// trust here. All three sources (X-Forwarded-For, X-Real-IP, RemoteAddr) are
// only as trustworthy as whatever sits in front of this process - a request
// with no proxy in front (bare NodePort/LoadBalancer) lets the client set
// either header to anything.
func GetClientIP(r *http.Request) string {
	if xForwardedFor := r.Header.Get(xForwardedForHeaderName); xForwardedFor != "" {
		hops := strings.Split(xForwardedFor, ",")
		for i := len(hops) - 1; i >= 0; i-- {
			if hop := strings.TrimSpace(hops[i]); hop != "" {
				return stripPort(hop)
			}
		}
	}
	if xRealIP := strings.TrimSpace(r.Header.Get(xRealIPHeaderName)); xRealIP != "" {
		return stripPort(xRealIP)
	}
	return stripPort(r.RemoteAddr)
}

// stripPort removes a trailing ":port" from a host[:port] value (IPv4, or IPv6
// bracketed as "[::1]:1234"). A value with no port - including a bare IPv6
// address, which net.SplitHostPort would otherwise misparse - is returned
// unchanged.
func stripPort(hostport string) string {
	if host, _, err := net.SplitHostPort(hostport); err == nil {
		return host
	}
	return hostport
}
