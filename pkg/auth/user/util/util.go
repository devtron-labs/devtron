package util

import (
	"fmt"
	"strings"
)

const (
	ApiTokenPrefix = "API-TOKEN:"
)

func CheckValidationForRoleGroupCreation(name string) bool {
	if strings.Contains(name, ",") {
		return false
	}
	return true
}

func GetGroupCasbinName(groups []string) []string {
	groupCasbinNames := make([]string, 0)
	for _, group := range groups {
		toLowerGroup := strings.ToLower(group)
		groupCasbinNames = append(groupCasbinNames, fmt.Sprintf("group:%s", strings.ReplaceAll(toLowerGroup, " ", "_")))
	}
	return groupCasbinNames
}
func CheckIfApiToken(email string) bool {
	return strings.HasPrefix(email, ApiTokenPrefix)
}

func CheckIfAdminOrApiToken(email string) bool {
	if email == "admin" || CheckIfApiToken(email) {
		return true
	}
	return false
}

func FindMin(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func FindStartAndEndForOffsetAndSize(totalLen, offset, size int) (int, int) {
	start := offset
	end := FindMin(offset+size, totalLen)
	return start, end
}
