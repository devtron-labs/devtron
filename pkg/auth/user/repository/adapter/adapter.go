/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package adapter

import (
	bean "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

func GetLastLoginTime(model repository.UserModel) time.Time {
	lastLoginTime := time.Time{}
	if model.UserAudit != nil {
		lastLoginTime = model.UserAudit.UpdatedOn
	}
	return lastLoginTime
}

func GetUserModelBasicAdapter(emailId, accessToken, userType string) *repository.UserModel {
	model := &repository.UserModel{
		EmailId:     emailId,
		AccessToken: accessToken,
		UserType:    userType,
	}
	return model
}

func GetUserRoleModelAdapter(userId, userLoggedInId int32, roleId int, twcConfigDto *bean.TimeoutWindowConfigDto) *repository.UserRoleModel {
	return &repository.UserRoleModel{
		UserId:   userId,
		RoleId:   roleId,
		AuditLog: sql.NewDefaultAuditLog(userLoggedInId),
	}
}

func GetRoleGroupRoleMappingModelAdapter(roleGroupId int32, roleId int, userId int32) *repository.RoleGroupRoleMapping {
	return &repository.RoleGroupRoleMapping{
		RoleGroupId: roleGroupId,
		RoleId:      roleId,
		AuditLog:    sql.NewDefaultAuditLog(userId),
	}
}
