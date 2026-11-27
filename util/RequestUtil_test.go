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
	"net/http"
	"testing"
)

func newClientIPTestRequest(xForwardedFor, xRealIP, remoteAddr string) *http.Request {
	r := &http.Request{Header: http.Header{}, RemoteAddr: remoteAddr}
	if xForwardedFor != "" {
		r.Header.Set(xForwardedForHeaderName, xForwardedFor)
	}
	if xRealIP != "" {
		r.Header.Set(xRealIPHeaderName, xRealIP)
	}
	return r
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name          string
		xForwardedFor string
		xRealIP       string
		remoteAddr    string
		want          string
	}{
		{
			name:          "single XFF entry, no port - the common single-proxy case",
			xForwardedFor: "203.0.113.5",
			remoteAddr:    "10.48.1.1:443",
			want:          "203.0.113.5",
		},
		{
			name:          "multi-hop XFF takes the rightmost entry, not the client-claimed leftmost",
			xForwardedFor: "1.2.3.4, 203.0.113.5",
			remoteAddr:    "10.48.1.1:443",
			want:          "203.0.113.5",
		},
		{
			name:          "whitespace around entries is trimmed",
			xForwardedFor: "1.2.3.4 ,  203.0.113.5  ",
			want:          "203.0.113.5",
		},
		{
			name:          "trailing empty entry is skipped in favor of the last real one",
			xForwardedFor: "1.2.3.4, 203.0.113.5,",
			want:          "203.0.113.5",
		},
		{
			name:          "leading empty entry is skipped when scanning from the right",
			xForwardedFor: ", 1.2.3.4",
			want:          "1.2.3.4",
		},
		{
			name:          "Azure App Gateway style ip:port entry has its port stripped",
			xForwardedFor: "1.2.3.4:5000, 203.0.113.5:6000",
			want:          "203.0.113.5",
		},
		{
			name:          "bracketed IPv6 entry with port has its port stripped",
			xForwardedFor: "[2001:db8::1]:6000",
			want:          "2001:db8::1",
		},
		{
			name:       "no XFF - falls back to X-Real-IP",
			xRealIP:    "203.0.113.9",
			remoteAddr: "10.48.1.1:443",
			want:       "203.0.113.9",
		},
		{
			name:       "no XFF, no X-Real-IP - falls back to RemoteAddr with port stripped (IPv4)",
			remoteAddr: "10.0.0.85:51234",
			want:       "10.0.0.85",
		},
		{
			name:       "RemoteAddr with bracketed IPv6 and port",
			remoteAddr: "[2001:db8::1]:51234",
			want:       "2001:db8::1",
		},
		{
			name:       "RemoteAddr with no port is returned as-is",
			remoteAddr: "10.0.0.85",
			want:       "10.0.0.85",
		},
		{
			name:       "unparseable RemoteAddr is returned as-is rather than dropped",
			remoteAddr: "not-an-address",
			want:       "not-an-address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newClientIPTestRequest(tt.xForwardedFor, tt.xRealIP, tt.remoteAddr)
			if got := GetClientIP(r); got != tt.want {
				t.Errorf("GetClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}
