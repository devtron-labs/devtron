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
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/auth/user/repository/bean"
)

func GetCasbinGroupPolicy(emailId string, role string, twcDto *bean2.TimeoutWindowConfigDto) bean.Policy {
	return bean.Policy{
		Type: "g",
		Sub:  bean.Subject(emailId),
		Obj:  bean.Object(role),
	}
}

func BuildClusterRoleFieldsDto(entity, accessType, cluster, namespace, group, kind, resource, actionType, subAction string) *bean3.RoleModelFieldsDto {
	return &bean3.RoleModelFieldsDto{
		Entity:     entity,
		AccessType: accessType,
		Cluster:    cluster,
		Namespace:  namespace,
		Group:      group,
		Kind:       kind,
		Resource:   resource,
		Action:     actionType,
	}
}

func BuildSuperAdminRoleFieldsDto() *bean3.RoleModelFieldsDto {
	return &bean3.RoleModelFieldsDto{
		Action: bean2.SUPER_ADMIN,
	}
}

func BuildOtherRoleFieldsDto(entity, team, entityName, environment, actionType, accessType string, OldValues bool, subAction string, approver bool) *bean3.RoleModelFieldsDto {
	return &bean3.RoleModelFieldsDto{
		Entity:     entity,
		Team:       team,
		App:        entityName,
		Env:        environment,
		Action:     actionType,
		AccessType: accessType,
		OldValues:  OldValues,
	}
}
func BuildJobsRoleFieldsDto(entity, team, entityName, environment, actionType, accessType, workflow, subAction string) *bean3.RoleModelFieldsDto {
	return &bean3.RoleModelFieldsDto{
		Entity:     entity,
		Team:       team,
		App:        entityName,
		Env:        environment,
		Action:     actionType,
		AccessType: accessType,
		Workflow:   workflow,
	}
}
