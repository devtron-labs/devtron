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

type GlobalAuthorisationConfigType string

const (
	DevtronSystemManaged       GlobalAuthorisationConfigType = "devtron-system-managed"
	DevtronSelfRegisteredGroup GlobalAuthorisationConfigType = "devtron-self-registered-group"
	GroupClaims                GlobalAuthorisationConfigType = "group-claims" // GroupClaims are currently used for Active directory and LDAP
)

type GlobalAuthorisationConfig struct {
	ConfigTypes []string `json:"configTypes" validate:"required"`
	UserId      int32    `json:"userId"` //for Internal Use
}

type GlobalAuthorisationConfigResponse struct {
	Id         int    `json:"id"`
	ConfigType string `json:"configType"`
	Active     bool   `json:"active"`
}
