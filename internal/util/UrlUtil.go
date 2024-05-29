/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package util

import (
	"net/url"
)

func IsValidUrl(input string) bool {
	_, err := url.ParseRequestURI(input)
	if err != nil {
		return false
	}

	u, err := url.Parse(input)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}
