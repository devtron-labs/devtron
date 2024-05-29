/*
 * Copyright (c) 2024. Devtron Inc.
 */

package util

func IsGroupsPresent(groups []string) bool {
	if len(groups) > 0 {
		return true
	}
	return false
}
