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
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/helper"
)

func IsGroupsPresent(groups []string) bool {
	if len(groups) > 0 {
		return true
	}
	return false
}

func FilterRoleGroupIfAlreadyPresent(roleGroups []bean.UserRoleGroup, mapOfExistingUserRoleGroup map[string]bool) []bean.UserRoleGroup {
	finalRoleGroups := make([]bean.UserRoleGroup, 0, len(roleGroups))
	for _, roleGrp := range roleGroups {
		if _, ok := mapOfExistingUserRoleGroup[helper.GetCasbinNameFromRoleGroupName(roleGrp.RoleGroup.Name)]; !ok {
			finalRoleGroups = append(finalRoleGroups, roleGrp)
		}
	}
	return finalRoleGroups

}
