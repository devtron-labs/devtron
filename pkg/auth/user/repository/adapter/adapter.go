/*
 * Copyright (c) 2024. Devtron Inc.
 */

package adapter

import (
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
)

func GetUserModelBasicAdapter(emailId, accessToken, userType string) *repository.UserModel {
	model := &repository.UserModel{
		EmailId:     emailId,
		AccessToken: accessToken,
		UserType:    userType,
	}
	return model
}

func GetUserRoleModelAdapter(userId, userLoggedInId int32, roleId int) *repository.UserRoleModel {
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
