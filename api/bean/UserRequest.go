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
	"time"
)

type UserRole struct {
	Id      int32  `json:"id" validate:"number"`
	EmailId string `json:"email_id" validate:"email"`
	Role    string `json:"role"`
}

type UserInfo struct {
	Id           int32        `json:"id" validate:"number"`
	EmailId      string       `json:"email_id" validate:"required"`
	Roles        []string     `json:"roles,omitempty"`
	AccessToken  string       `json:"access_token,omitempty"`
	UserType     string       `json:"-"`
	LastUsedAt   time.Time    `json:"-"`
	LastUsedByIp string       `json:"-"`
	Exist        bool         `json:"-"`
	UserId       int32        `json:"-"` // created or modified user id
	RoleFilters  []RoleFilter `json:"roleFilters"`
	Status       string       `json:"status,omitempty"`
	Groups       []string     `json:"groups"`
	SuperAdmin   bool         `json:"superAdmin,notnull"`
}

type RoleGroup struct {
	Id          int32        `json:"id" validate:"number"`
	Name        string       `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	RoleFilters []RoleFilter `json:"roleFilters"`
	Status      string       `json:"status,omitempty"`
	UserId      int32        `json:"-"` // created or modified user id
}

type RoleFilter struct {
	Entity      string `json:"entity"`
	Team        string `json:"team"`
	EntityName  string `json:"entityName"`
	Environment string `json:"environment"`
	Action      string `json:"action"`
	AccessType  string `json:"accessType"`
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
	AccessType  string `json:"accessType"`
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
	POLICY_DIRECT PolicyType = 1
	POLICY_GROUP  PolicyType = 1
)

const SUPERADMIN = "role:super-admin___"
const APP_ACCESS_TYPE_HELM = "helm-app"

const USER_TYPE_API_TOKEN = "apiToken"
