package util

import "strings"

func GetImageTagFromImage(image string) string {
	parts := strings.Split(image, ":")

	if len(parts) < 1 {
		return ""
	}
	return parts[len(parts)-1]
}
