package util

import "strings"

func RemoveTrailingAndLeadingSpaces(value string) string {
	return strings.TrimSpace(value)
}
