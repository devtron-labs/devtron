package util

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
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

func GetUniqueKeyForRoleFilter(roleFilter bean.RoleFilter) string {
	key := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%t-%s-%s-%s-%s-%s-%s", roleFilter.Entity, roleFilter.Team, roleFilter.Environment,
		roleFilter.EntityName, roleFilter.Action, roleFilter.AccessType, roleFilter.Approver, roleFilter.Cluster, roleFilter.Namespace, roleFilter.Group, roleFilter.Kind, roleFilter.Resource, roleFilter.Workflow)
	return key
}
