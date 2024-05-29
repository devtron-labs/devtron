/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
