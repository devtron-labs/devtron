/*
 * Copyright (c) 2024. Devtron Inc.
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
