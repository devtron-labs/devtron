package util

import "strings"

func CheckValidationForRoleGroupCreation(name string) bool {
	if strings.Contains(name, ",") {
		return false
	}
	return true
}
