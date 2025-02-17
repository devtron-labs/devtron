/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

type UserRole struct {
	Id      int32  `json:"id" validate:"number"`
	EmailId string `json:"email_id" validate:"email"`
	Role    string `json:"role"`
}

type UserInfo struct {
	Id            int32           `json:"id" validate:"number,not-system-admin-userid"`
	EmailId       string          `json:"email_id" validate:"required,not-system-admin-user"` // TODO : have to migrate json key to emailId and also handle backward compatibility
	Roles         []string        `json:"roles,omitempty"`
	AccessToken   string          `json:"access_token,omitempty"`
	RoleFilters   []RoleFilter    `json:"roleFilters"`
	Status        string          `json:"status,omitempty"`
	Groups        []string        `json:"groups"`         // this will be deprecated in future do not use
	UserRoleGroup []UserRoleGroup `json:"userRoleGroups"` // role group with metadata
	SuperAdmin    bool            `json:"superAdmin,notnull"`
	LastLoginTime time.Time       `json:"lastLoginTime"`
	UserType      string          `json:"-"`
	LastUsedAt    time.Time       `json:"-"`
	LastUsedByIp  string          `json:"-"`
	Exist         bool            `json:"-"`
	UserId        int32           `json:"-"` // created or modified user id
}

type RoleGroup struct {
	Id          int32        `json:"id" validate:"number"`
	Name        string       `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	RoleFilters []RoleFilter `json:"roleFilters"`
	Status      string       `json:"status,omitempty"`
	SuperAdmin  bool         `json:"superAdmin"`
	UserId      int32        `json:"-"` // created or modified user id
}

type RoleFilter struct {
	Entity      string `json:"entity"`
	Team        string `json:"team"`
	EntityName  string `json:"entityName"`
	Environment string `json:"environment"`
	Action      string `json:"action"`
	AccessType  string `json:"accessType"`

	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Group     string `json:"group"`
	Kind      string `json:"kind"`
	Resource  string `json:"resource"`
	Workflow  string `json:"workflow"`
}

func (rf RoleFilter) GetTeam() string        { return rf.Team }
func (rf RoleFilter) GetEntity() string      { return rf.Entity }
func (rf RoleFilter) GetAction() string      { return rf.Action }
func (rf RoleFilter) GetAccessType() string  { return rf.AccessType }
func (rf RoleFilter) GetEnvironment() string { return rf.Environment }
func (rf RoleFilter) GetCluster() string     { return rf.Cluster }
func (rf RoleFilter) GetGroup() string       { return rf.Group }
func (rf RoleFilter) GetKind() string        { return rf.Kind }
func (rf RoleFilter) GetEntityName() string  { return rf.EntityName }
func (rf RoleFilter) GetResource() string    { return rf.Resource }
func (rf RoleFilter) GetWorkflow() string    { return rf.Workflow }
func (rf RoleFilter) GetNamespace() string   { return rf.Namespace }

type Role struct {
	Id   int    `json:"id" validate:"number"`
	Role string `json:"role" validate:"required"`
}

type RoleData struct {
	Id          int    `json:"id" validate:"number"`
	Role        string `json:"role" validate:"required"`
	Entity      string `json:"entity"`
	Team        string `json:"team"`
	EntityName  string `json:"entityName"`
	Environment string `json:"environment"`
	Action      string `json:"action"`
	AccessType  string `json:"accessType"`

	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Group     string `json:"group"`
	Kind      string `json:"kind"`
	Resource  string `json:"resource"`
}

type SSOLoginDto struct {
	Id     int32           `json:"id"`
	Name   string          `json:"name,omitempty"`
	Label  string          `json:"label,omitempty"`
	Url    string          `json:"url,omitempty"`
	Config json.RawMessage `json:"config,omitempty"`
	Active bool            `json:"active"`
	UserId int32           `json:"-"`
}

const (
	NOCHARTEXIST string = "NOCHARTEXIST"
)

type PolicyType int

const (
	POLICY_DIRECT        PolicyType = 1
	POLICY_GROUP         PolicyType = 1
	SUPERADMIN                      = "role:super-admin___"
	APP_ACCESS_TYPE_HELM            = "helm-app"
	USER_TYPE_API_TOKEN             = "apiToken"
	CHART_GROUP_ENTITY              = "chart-group"
	CLUSTER_ENTITIY                 = "cluster"
	ACTION_SUPERADMIN               = "super-admin"
)

type UserListingResponse struct {
	Users      []UserInfo `json:"users"`
	TotalCount int        `json:"totalCount"`
}

type RoleGroupListingResponse struct {
	RoleGroups []*RoleGroup `json:"roleGroups"`
	TotalCount int          `json:"totalCount"`
}

type RestrictedGroup struct {
	Group                   string
	HasSuperAdminPermission bool
}

type ListingRequest struct {
	SearchKey  string         `json:"searchKey"`
	SortOrder  bean.SortOrder `json:"sortOrder"`
	SortBy     bean.SortBy    `json:"sortBy"`
	Offset     int            `json:"offset"`
	Size       int            `json:"size"`
	ShowAll    bool           `json:"showAll"`
	CountCheck bool           `json:"-"`
}

type BulkDeleteRequest struct {
	Ids            []int32         `json:"ids"`
	ListingRequest *ListingRequest `json:"listingRequest,omitempty"`
	LoggedInUserId int32           `json:"-"`
}

type UserRoleGroup struct {
	RoleGroup *RoleGroup `json:"roleGroup"`
}

type GroupPermissionsAuditDto struct {
	RoleGroupInfo *RoleGroup   `json:"roleGroupInfo,omitempty"`
	EntityAudit   sql.AuditLog `json:"entityAudit,omitempty"`
}

func NewGroupPermissionsAuditDto() *GroupPermissionsAuditDto {
	return &GroupPermissionsAuditDto{}
}

func (pa *GroupPermissionsAuditDto) WithRoleGroupInfo(roleGroupInfo *RoleGroup) *GroupPermissionsAuditDto {
	pa.RoleGroupInfo = roleGroupInfo
	return pa
}
func (pa *GroupPermissionsAuditDto) WithEntityAudit(entityAudit sql.AuditLog) *GroupPermissionsAuditDto {
	pa.EntityAudit = entityAudit
	return pa
}

type UserPermissionsAuditDto struct {
	UserInfo    *UserInfo    `json:"userInfo,omitempty"`
	EntityAudit sql.AuditLog `json:"entityAudit,omitempty"`
}

func NewUserPermissionsAuditDto() *UserPermissionsAuditDto {
	return &UserPermissionsAuditDto{}
}

func (pa *UserPermissionsAuditDto) WithUserInfo(userInfo *UserInfo) *UserPermissionsAuditDto {
	pa.UserInfo = userInfo
	return pa
}

func (pa *UserPermissionsAuditDto) WithEntityAudit(entityAudit sql.AuditLog) *UserPermissionsAuditDto {
	pa.EntityAudit = entityAudit
	return pa
}
