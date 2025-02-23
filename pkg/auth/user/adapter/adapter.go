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
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"strings"
)

func GetCasbinGroupPolicyForEmailAndRoleOnly(emailId string, role string) bean.Policy {
	return bean.Policy{
		Type: "g",
		Sub:  bean.Subject(emailId),
		Obj:  bean.Object(role),
	}
}

func CreateRestrictedGroup(roleGroupName string, hasSuperAdminPermission bool) bean2.RestrictedGroup {
	trimmedGroup := strings.TrimPrefix(roleGroupName, "group:")
	return bean2.RestrictedGroup{
		Group:                   trimmedGroup,
		HasSuperAdminPermission: hasSuperAdminPermission,
	}
}

func BuildUserInfoResponseAdapter(requestUserInfo *bean2.UserInfo, emailId string) *bean2.UserInfo {
	return &bean2.UserInfo{
		Id:            requestUserInfo.Id,
		EmailId:       emailId,
		Groups:        requestUserInfo.Groups,
		RoleFilters:   requestUserInfo.RoleFilters,
		SuperAdmin:    requestUserInfo.SuperAdmin,
		UserRoleGroup: requestUserInfo.UserRoleGroup,
	}
}

func BuildSelfRegisterDto(userInfo *bean2.UserInfo) *bean2.SelfRegisterDto {
	return &bean2.SelfRegisterDto{
		UserInfo: userInfo,
	}
}

func BuildRoleFilterFromRoleF(roleF bean2.RoleFilter) bean2.RoleFilter {
	return bean2.RoleFilter{
		Entity:      roleF.Entity,
		Team:        roleF.Team,
		Environment: roleF.Environment,
		EntityName:  roleF.EntityName,
		Action:      roleF.Action,
		AccessType:  roleF.AccessType,
		Cluster:     roleF.Cluster,
		Namespace:   roleF.Namespace,
		Group:       roleF.Group,
		Kind:        roleF.Kind,
		Resource:    roleF.Resource,
		Workflow:    roleF.Workflow,
	}
}

func BuildUserInfo(id int32, emailId string, userGroupMapDto *bean2.UserGroupMapDto) bean2.UserInfo {
	return bean2.UserInfo{
		Id:            id,
		EmailId:       emailId,
		RoleFilters:   make([]bean2.RoleFilter, 0),
		Groups:        make([]string, 0),
		UserRoleGroup: make([]bean2.UserRoleGroup, 0),
	}
}
