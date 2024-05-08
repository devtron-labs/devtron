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

package util

import (
	"net/http"
)

const xForwardedForHeaderName = "X-Forwarded-For"

// GetClientIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func GetClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get(xForwardedForHeaderName)
	if len(xForwardedFor) > 0 {
		return xForwardedFor
	}
	return r.RemoteAddr
}
