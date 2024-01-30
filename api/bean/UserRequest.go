/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package bean

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"time"
)

type UserRole struct {
	Id      int32  `json:"id" validate:"number"`
	EmailId string `json:"email_id" validate:"email"`
	Role    string `json:"role"`
}

type UserInfo struct {
	Id            int32        `json:"id" validate:"number,not-system-admin-userid"`
	EmailId       string       `json:"emailId" validate:"required,not-system-admin-user-email"`
	Roles         []string     `json:"roles,omitempty"`
	AccessToken   string       `json:"access_token,omitempty"`
	UserType      string       `json:"-"`
	LastUsedAt    time.Time    `json:"-"`
	LastUsedByIp  string       `json:"-"`
	Exist         bool         `json:"-"`
	UserId        int32        `json:"-"` // created or modified user id
	RoleFilters   []RoleFilter `json:"roleFilters"`
	Status        string       `json:"status,omitempty"`
	Groups        []string     `json:"groups"` // this will be deprecated in future do not use
	SuperAdmin    bool         `json:"superAdmin,notnull"`
	RoleGroups    []RoleGroup  `json:"roleGroups,omitempty"` // role group with metadata, currently using for group claims
	LastLoginTime time.Time    `json:"lastLoginTime"`
	TimeToLive    time.Time    `json:"timeToLive"`
	UserStatus    Status       `json:"userStatus"`
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
	Approver    bool   `json:"approver"`
	AccessType  string `json:"accessType"`

	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Group     string `json:"group"`
	Kind      string `json:"kind"`
	Resource  string `json:"resource"`
	Workflow  string `json:"workflow"`
}

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
	Approver    bool   `json:"approver"`
	AccessType  string `json:"accessType"`

	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Group     string `json:"group"`
	Kind      string `json:"kind"`
	Resource  string `json:"resource"`
}

type SSOLoginDto struct {
	Id                   int32           `json:"id"`
	Name                 string          `json:"name,omitempty"`
	Label                string          `json:"label,omitempty"`
	Url                  string          `json:"url,omitempty"`
	Config               json.RawMessage `json:"config,omitempty"`
	Active               bool            `json:"active"`
	GlobalAuthConfigType string          `json:"globalAuthConfigType"`
	UserId               int32           `json:"-"`
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
)

type UserListingResponse struct {
	Users      []UserInfo `json:"users"`
	TotalCount int        `json:"totalCount"`
}

type RoleGroupListingResponse struct {
	RoleGroups []*RoleGroup `json:"roleGroups"`
	TotalCount int          `json:"totalCount"`
}

type Status string

const (
	Active          Status = "active"
	Inactive        Status = "inactive"
	TemporaryAccess Status = "temporaryAccess"
	Unknown         Status = "unknown"
)

type BulkStatusUpdateRequest struct {
	UserIds    []int32   `json:"userIds",validate:"required"`
	Status     Status    `json:"status",validate:"required"'`
	TimeToLive time.Time `json:"timeToLive"`
}

type ActionResponse struct {
	Suceess bool `json:"suceess"`
}

type FetchListingRequest struct {
	Status      Status         `json:"status"`
	SearchKey   string         `json:"searchKey"`
	SortOrder   bean.SortOrder `json:"sortOrder"`
	SortBy      bean.SortBy    `json:"sortBy"`
	Offset      int            `json:"offset"`
	Size        int            `json:"size"`
	ShowAll     bool           `json:"showAll"`
	CurrentTime time.Time      `json:"-"` // for Internal Use
}
