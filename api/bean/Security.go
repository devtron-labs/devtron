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

// CreateVulnerabilityPolicyRequest defines model for CreateVulnerabilityPolicyRequest.
type CreateVulnerabilityPolicyRequest struct {
	// actions which can be taken on vulnerabilities
	Action    *VulnerabilityAction `json:"action,omitempty"`
	AppId     int                  `json:"appId,omitempty"`
	ClusterId int                  `json:"clusterId,omitempty"`
	CveId     string               `json:"cveId,omitempty"`
	EnvId     int                  `json:"envId,omitempty"`
	Severity  string               `json:"severity,omitempty"`
}

// CreateVulnerabilityPolicyResponse defines model for CreateVulnerabilityPolicyResponse.
type CreateVulnerabilityPolicyResponse struct {
	// Error object
	Error  *Error                       `json:"error,omitempty"`
	Result *IdVulnerabilityPolicyResult `json:"result,omitempty"`
}

// CvePolicy defines model for CvePolicy.
type CvePolicy struct {
	// Embedded struct due to allOf(#/components/schemas/SeverityPolicy)
	SeverityPolicy
	// Embedded fields due to inline allOf schema

	// In case of CVE policy this is same as cve name else it is blank
	Name string `json:"name,omitempty"`
}

// DeleteVulnerabilityPolicyResponse defines model for DeleteVulnerabilityPolicyResponse.
type DeleteVulnerabilityPolicyResponse struct {
	// Error object
	Error  *Error                       `json:"error,omitempty"`
	Result *IdVulnerabilityPolicyResult `json:"result,omitempty"`
}

// Error defines model for Error.
type Error struct {
	// Error code
	Code int32 `json:"code"`

	// Error message
	Message string `json:"message"`
}

// GetVulnerabilityPolicyResponse defines model for GetVulnerabilityPolicyResponse.
type GetVulnerabilityPolicyResponse struct {
	// Error object
	Error  *Error                        `json:"error,omitempty"`
	Result *GetVulnerabilityPolicyResult `json:"result,omitempty"`
}

// GetVulnerabilityPolicyResult defines model for GetVulnerabilityPolicyResult.
type GetVulnerabilityPolicyResult struct {
	// Resource Level can be one of global, cluster, environment, application
	Level    ResourceLevel          `json:"level"`
	Policies []*VulnerabilityPolicy `json:"policies"`
}

// IdVulnerabilityPolicyResult defines model for IdVulnerabilityPolicyResult.
type IdVulnerabilityPolicyResult struct {
	Id int `json:"id"`
}

// ResourceLevel defines model for ResourceLevel.
type ResourceLevel string

// SeverityPolicy defines model for SeverityPolicy.
type SeverityPolicy struct {
	Id int `json:"id"`

	// Whether vulnerability is allowed or blocked and is it inherited or is it overriden
	Policy       *VulnerabilityPermission `json:"policy"`
	PolicyOrigin string                   `json:"policyOrigin"`
	Severity     string                   `json:"severity"`
}

// UpdateVulnerabilityPolicyResponse defines model for UpdateVulnerabilityPolicyResponse.
type UpdateVulnerabilityPolicyResponse struct {
	// Error object
	Error  *Error                       `json:"error,omitempty"`
	Result *IdVulnerabilityPolicyResult `json:"result,omitempty"`
}

// VulnerabilityAction defines model for VulnerabilityAction.
type VulnerabilityAction string

// VulnerabilityPermission defines model for VulnerabilityPermission.
type VulnerabilityPermission struct {
	// actions which can be taken on vulnerabilities
	Action      VulnerabilityAction `json:"action"`
	Inherited   bool                `json:"inherited"`
	IsOverriden bool                `json:"isOverriden"`
}

// VulnerabilityPolicy defines model for VulnerabilityPolicy.
type VulnerabilityPolicy struct {
	Cves []*CvePolicy `json:"cves"`

	// environment id in case of application
	EnvId int `json:"envId,omitempty"`

	// Is name of cluster or environment or application/environment
	Name       string            `json:"name,omitempty"`
	Severities []*SeverityPolicy `json:"severities"`
	AppId      int               `json:"-"`
	ClusterId  int               `json:"-"`
}

// DeletePolicyParams defines parameters for DeletePolicy.
type DeletePolicyParams struct {
	Id int `json:"id"`
}

// FetchPolicyParams defines parameters for FetchPolicy.
type FetchPolicyParams struct {
	Level ResourceLevel `json:"level"`
	Id    int           `json:"id,omitempty"`
}

// UpdatePolicyParams defines parameters for UpdatePolicy.
type UpdatePolicyParams struct {
	Id     int    `json:"id"`
	Action string `json:"action"`
}
