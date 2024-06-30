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

package helper

import (
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"golang.org/x/exp/slices"
	"strings"
)

func IsSystemOrAdminUser(userId int32) bool {
	if userId == bean.SystemUserId || userId == bean.AdminUserId {
		return true
	}
	return false
}

func IsSystemOrAdminUserByEmail(email string) bool {
	if email == bean.AdminUser || email == bean.SystemUser {
		return true
	}
	return false
}

func CheckValidationForAdminAndSystemUserId(userIds []int32) error {
	validated := CheckIfUserDevtronManagedOnly(userIds)
	if !validated {
		err := &util.ApiError{Code: "406", HttpStatusCode: 406, UserMessage: "cannot update status for system or admin user"}
		return err
	}
	return nil
}

func CheckIfUserDevtronManagedOnly(userIds []int32) bool {
	if slices.Contains(userIds, bean.AdminUserId) || slices.Contains(userIds, bean.SystemUserId) {
		return false
	}
	return true
}

func CheckIfUserIdsExists(userIds []int32) error {
	var err error
	if len(userIds) == 0 {
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "no user ids provided"}
		return err
	}
	return nil
}

func ExtractTokenNameFromEmail(email string) string {
	return strings.Split(email, ":")[1]
}

func CreateErrorMessageForUserRoleGroups(restrictedGroups []bean2.RestrictedGroup) (string, string) {
	var restrictedGroupsWithSuperAdminPermission string
	var restrictedGroupsWithoutSuperAdminPermission string
	var errorMessageForGroupsWithoutSuperAdmin string
	var errorMessageForGroupsWithSuperAdmin string
	for _, group := range restrictedGroups {
		if group.HasSuperAdminPermission {
			restrictedGroupsWithSuperAdminPermission += fmt.Sprintf("%s,", group.Group)
		} else {
			restrictedGroupsWithoutSuperAdminPermission += fmt.Sprintf("%s,", group.Group)
		}
	}

	if len(restrictedGroupsWithoutSuperAdminPermission) > 0 {
		// if any group was appended, remove the comma from the end
		restrictedGroupsWithoutSuperAdminPermission = restrictedGroupsWithoutSuperAdminPermission[:len(restrictedGroupsWithoutSuperAdminPermission)-1]
		errorMessageForGroupsWithoutSuperAdmin = fmt.Sprintf("You do not have manager permission for some or all projects in group(s): %v.", restrictedGroupsWithoutSuperAdminPermission)
	}
	if len(restrictedGroupsWithSuperAdminPermission) > 0 {
		// if any group was appended, remove the comma from the end
		restrictedGroupsWithSuperAdminPermission = restrictedGroupsWithSuperAdminPermission[:len(restrictedGroupsWithSuperAdminPermission)-1]
		errorMessageForGroupsWithSuperAdmin = fmt.Sprintf("Only super admins can assign groups with super admin permission: %v.", restrictedGroupsWithSuperAdminPermission)
	}
	return errorMessageForGroupsWithoutSuperAdmin, errorMessageForGroupsWithSuperAdmin
}
