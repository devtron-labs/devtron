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

package bean

type RoleType string

func (r RoleType) String() string {
	return string(r)
}

const (
	SYSTEM_USER_ID                              = 1
	PROJECT_TYPE                                = "team"
	ENV_TYPE                                    = "environment"
	APP_TYPE                                    = "app"
	WorkflowType                                = "workflow"
	CHART_GROUP_TYPE                            = "chart-group"
	MANAGER_TYPE                       RoleType = "manager"
	ADMIN_TYPE                         RoleType = "admin"
	TRIGGER_TYPE                       RoleType = "trigger"
	VIEW_TYPE                          RoleType = "view"
	ENTITY_ALL_TYPE                    RoleType = "entityAll"
	ENTITY_VIEW_TYPE                   RoleType = "entityView"
	ENTITY_SPECIFIC_TYPE               RoleType = "entitySpecific"
	ENTITY_SPECIFIC_ADMIN_TYPE         RoleType = "entitySpecificAdmin"
	ENTITY_SPECIFIC_VIEW_TYPE          RoleType = "entitySpecificView"
	ROLE_SPECIFIC_TYPE                 RoleType = "roleSpecific"
	ENTITY_CLUSTER_ADMIN_TYPE          RoleType = "clusterAdmin"
	ENTITY_CLUSTER_VIEW_TYPE           RoleType = "clusterView"
	ADMIN_HELM_TYPE                    RoleType = "admin"
	EDIT_HELM_TYPE                     RoleType = "edit"
	VIEW_HELM_TYPE                     RoleType = "view"
	ENTITY_CLUSTER_EDIT_TYPE           RoleType = "clusterEdit"
	SUPER_ADMIN                                 = "super-admin"
	GLOBAL_ENTITY                               = "globalEntity"
	EMPTY_ROLEFILTER_ENTRY_PLACEHOLDER          = "NONE"
	RoleNotFoundStatusPrefix                    = "role not fount for any given filter: "
	EmptyStringIndicatingAll                    = ""
)

// entity

const (
	ENTITY_APPS        = "apps"
	EntityJobs         = "jobs"
	CHART_GROUP_ENTITY = "chart-group"
	CLUSTER_ENTITIY    = "cluster"
)

//access types

const (
	DEVTRON_APP          = "devtron-app"
	APP_ACCESS_TYPE_HELM = "helm-app"
	EmptyAccessType      = ""
)

const (
	VALIDATION_FAILED_ERROR_MSG string = "validation failed: group name with , is not allowed"
)

// messages constants
const (
	NoTokenProvidedMessage    = "no token provided"
	RoleAlreadyExistMessage   = "role already exist"
	PolicyAlreadyExistMessage = "policy already exist"
)

type RbacRoleDto struct {
	Id              int    `json:"id"` // id of the default role
	RoleName        string `json:"roleName"`
	RoleDisplayName string `json:"roleDisplayName"`
	RoleDescription string `json:"roleDescription"`
	*RbacPolicyEntityGroupDto
}

type RbacPolicyEntityGroupDto struct {
	Entity     string `json:"entity" validate:"oneof=apps cluster chart-group jobs"`
	AccessType string `json:"accessType,omitempty"`
}

type SortBy string
type SortOrder string

func (s SortBy) String() string {
	return string(s)
}

func (s SortOrder) String() string {
	return string(s)
}

const (
	Asc  SortOrder = "ASC"
	Desc SortOrder = "DESC"
)

const (
	Email     SortBy = "email_id"
	LastLogin SortBy = "last_login"
	GroupName SortBy = "name"
)

const (
	DefaultSize int = 20
)

const (
	AdminUser  string = "admin"
	SystemUser string = "system"
)

const (
	AdminUserId  = 2 // we have established Admin user as 2 while setting up devtron
	SystemUserId = 1 // we have established System user as 1 while setting up devtron, which are being used for auto-trigger operations
)

const (
	API_TOKEN_USER_EMAIL_PREFIX = "API-TOKEN:"
	ApiTokenTableName           = "api_token"
)

const AnonymousUserEmail string = "anonymous"

type MergingBaseKey string

const (
	ApplicationBasedKey MergingBaseKey = "application"
	EnvironmentBasedKey MergingBaseKey = "environment"
)

type UserMetadata struct {
	UserEmailId      string
	IsUserSuperAdmin bool
	UserId           int32
}
