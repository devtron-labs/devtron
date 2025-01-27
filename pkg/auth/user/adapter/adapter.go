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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
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
func BuildPermissionAuditModel(entityId int32, entityType repository.EntityType,
	operationType repository.OperationType, permissionsAuditDto *bean.PermissionsAuditDto, userIdForAuditLog int32) (*repository.PermissionsAudit, error) {
	permissionsJson, err := json.Marshal(permissionsAuditDto)
	if err != nil {
		errToReturn := fmt.Sprintf("error in marshalling permissions audit dto :%s", err.Error())
		return nil, errors.New(errToReturn)
	}
	return &repository.PermissionsAudit{
		EntityId:        entityId,
		EntityType:      entityType,
		OperationType:   operationType,
		PermissionsJson: string(permissionsJson),
		AuditLog:        sql.NewDefaultAuditLog(userIdForAuditLog),
	}, nil

}
