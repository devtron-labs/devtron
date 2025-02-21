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

package adapter

import (
	"github.com/devtron-labs/devtron/api/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"strings"
	"time"
)

func GetLastLoginTime(model repository.UserModel) time.Time {
	lastLoginTime := time.Time{}
	if model.UserAudit != nil {
		lastLoginTime = model.UserAudit.UpdatedOn
	}
	return lastLoginTime
}

func CreateRestrictedGroup(roleGroupName string, hasSuperAdminPermission bool) bean.RestrictedGroup {
	trimmedGroup := strings.TrimPrefix(roleGroupName, "group:")
	return bean.RestrictedGroup{
		Group:                   trimmedGroup,
		HasSuperAdminPermission: hasSuperAdminPermission,
	}
}

func BuildUserInfoResponseAdapter(requestUserInfo *bean.UserInfo, emailId string) *bean.UserInfo {
	return &bean.UserInfo{
		Id:            requestUserInfo.Id,
		EmailId:       emailId,
		Groups:        requestUserInfo.Groups,
		RoleFilters:   requestUserInfo.RoleFilters,
		SuperAdmin:    requestUserInfo.SuperAdmin,
		UserRoleGroup: requestUserInfo.UserRoleGroup,
	}
}

func BuildSelfRegisterDto(userInfo *bean.UserInfo) *bean2.SelfRegisterDto {
	return &bean2.SelfRegisterDto{
		UserInfo: userInfo,
	}
}
