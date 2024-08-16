//
// Copyright 2023, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"fmt"
	"net/http"
)

// GroupProtectedEnvironmentsService handles communication with the group-level
// protected environment methods of the GitLab API.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html
type GroupProtectedEnvironmentsService struct {
	client *Client
}

// GroupProtectedEnvironment represents a group-level protected environment.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html
type GroupProtectedEnvironment struct {
	Name                  string                               `json:"name"`
	DeployAccessLevels    []*GroupEnvironmentAccessDescription `json:"deploy_access_levels"`
	RequiredApprovalCount int                                  `json:"required_approval_count"`
	ApprovalRules         []*GroupEnvironmentApprovalRule      `json:"approval_rules"`
}

// GroupEnvironmentAccessDescription represents the access decription for a
// group-level protected environment.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html
type GroupEnvironmentAccessDescription struct {
	ID                     int              `json:"id"`
	AccessLevel            AccessLevelValue `json:"access_level"`
	AccessLevelDescription string           `json:"access_level_description"`
	UserID                 int              `json:"user_id"`
	GroupID                int              `json:"group_id"`
	GroupInheritanceType   int              `json:"group_inheritance_type"`
}

// GroupEnvironmentApprovalRule represents the approval rules for a group-level
// protected environment.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#protect-a-single-environment
type GroupEnvironmentApprovalRule struct {
	ID                     int              `json:"id"`
	UserID                 int              `json:"user_id"`
	GroupID                int              `json:"group_id"`
	AccessLevel            AccessLevelValue `json:"access_level"`
	AccessLevelDescription string           `json:"access_level_description"`
	RequiredApprovalCount  int              `json:"required_approvals"`
	GroupInheritanceType   int              `json:"group_inheritance_type"`
}

// ListGroupProtectedEnvironmentsOptions represents the available
// ListGroupProtectedEnvironments() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#list-group-level-protected-environments
type ListGroupProtectedEnvironmentsOptions ListOptions

// ListGroupProtectedEnvironments returns a list of protected environments from
// a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#list-group-level-protected-environments
func (s *GroupProtectedEnvironmentsService) ListGroupProtectedEnvironments(gid interface{}, opt *ListGroupProtectedEnvironmentsOptions, options ...RequestOptionFunc) ([]*GroupProtectedEnvironment, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/protected_environments", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var pes []*GroupProtectedEnvironment
	resp, err := s.client.Do(req, &pes)
	if err != nil {
		return nil, resp, err
	}

	return pes, resp, nil
}

// GetGroupProtectedEnvironment returns a single group-level protected
// environment.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#get-a-single-protected-environment
func (s *GroupProtectedEnvironmentsService) GetGroupProtectedEnvironment(gid interface{}, environment string, options ...RequestOptionFunc) (*GroupProtectedEnvironment, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/protected_environments/%s", PathEscape(group), environment)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	pe := new(GroupProtectedEnvironment)
	resp, err := s.client.Do(req, pe)
	if err != nil {
		return nil, resp, err
	}

	return pe, resp, nil
}

// ProtectGroupEnvironmentOptions represents the available
// ProtectGroupEnvironment() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#protect-a-single-environment
type ProtectGroupEnvironmentOptions struct {
	Name                  *string                                 `url:"name,omitempty" json:"name,omitempty"`
	DeployAccessLevels    *[]*GroupEnvironmentAccessOptions       `url:"deploy_access_levels,omitempty" json:"deploy_access_levels,omitempty"`
	RequiredApprovalCount *int                                    `url:"required_approval_count,omitempty" json:"required_approval_count,omitempty"`
	ApprovalRules         *[]*GroupEnvironmentApprovalRuleOptions `url:"approval_rules,omitempty" json:"approval_rules,omitempty"`
}

// GroupEnvironmentAccessOptions represents the options for an access decription
// for a group-level protected environment.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#protect-a-single-environment
type GroupEnvironmentAccessOptions struct {
	AccessLevel          *AccessLevelValue `url:"access_level,omitempty" json:"access_level,omitempty"`
	UserID               *int              `url:"user_id,omitempty" json:"user_id,omitempty"`
	GroupID              *int              `url:"group_id,omitempty" json:"group_id,omitempty"`
	GroupInheritanceType *int              `url:"group_inheritance_type,omitempty" json:"group_inheritance_type,omitempty"`
}

// GroupEnvironmentApprovalRuleOptions represents the approval rules for a
// group-level protected environment.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#protect-a-single-environment
type GroupEnvironmentApprovalRuleOptions struct {
	UserID                 *int              `url:"user_id,omitempty" json:"user_id,omitempty"`
	GroupID                *int              `url:"group_id,omitempty" json:"group_id,omitempty"`
	AccessLevel            *AccessLevelValue `url:"access_level,omitempty" json:"access_level,omitempty"`
	AccessLevelDescription *string           `url:"access_level_description,omitempty" json:"access_level_description,omitempty"`
	RequiredApprovalCount  *int              `url:"required_approvals,omitempty" json:"required_approvals,omitempty"`
	GroupInheritanceType   *int              `url:"group_inheritance_type,omitempty" json:"group_inheritance_type,omitempty"`
}

// ProtectGroupEnvironment protects a single group-level environment.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#protect-a-single-environment
func (s *GroupProtectedEnvironmentsService) ProtectGroupEnvironment(gid interface{}, opt *ProtectGroupEnvironmentOptions, options ...RequestOptionFunc) (*GroupProtectedEnvironment, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/protected_environments", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	pe := new(GroupProtectedEnvironment)
	resp, err := s.client.Do(req, pe)
	if err != nil {
		return nil, resp, err
	}

	return pe, resp, nil
}

// UpdateGroupProtectedEnvironmentOptions represents the available
// UpdateGroupProtectedEnvironment() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#update-a-protected-environment
type UpdateGroupProtectedEnvironmentOptions struct {
	Name                  *string                                       `url:"name,omitempty" json:"name,omitempty"`
	DeployAccessLevels    *[]*UpdateGroupEnvironmentAccessOptions       `url:"deploy_access_levels,omitempty" json:"deploy_access_levels,omitempty"`
	RequiredApprovalCount *int                                          `url:"required_approval_count,omitempty" json:"required_approval_count,omitempty"`
	ApprovalRules         *[]*UpdateGroupEnvironmentApprovalRuleOptions `url:"approval_rules,omitempty" json:"approval_rules,omitempty"`
}

// UpdateGroupEnvironmentAccessOptions represents the options for updates to the
// access decription for a group-level protected environment.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#update-a-protected-environment
type UpdateGroupEnvironmentAccessOptions struct {
	AccessLevel          *AccessLevelValue `url:"access_level,omitempty" json:"access_level,omitempty"`
	ID                   *int              `url:"id,omitempty" json:"id,omitempty"`
	UserID               *int              `url:"user_id,omitempty" json:"user_id,omitempty"`
	GroupID              *int              `url:"group_id,omitempty" json:"group_id,omitempty"`
	GroupInheritanceType *int              `url:"group_inheritance_type,omitempty" json:"group_inheritance_type,omitempty"`
	Destroy              *bool             `url:"_destroy,omitempty" json:"_destroy,omitempty"`
}

// UpdateGroupEnvironmentApprovalRuleOptions represents the updates to the
// approval rules for a group-level protected environment.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#update-a-protected-environment
type UpdateGroupEnvironmentApprovalRuleOptions struct {
	ID                     *int              `url:"id,omitempty" json:"id,omitempty"`
	UserID                 *int              `url:"user_id,omitempty" json:"user_id,omitempty"`
	GroupID                *int              `url:"group_id,omitempty" json:"group_id,omitempty"`
	AccessLevel            *AccessLevelValue `url:"access_level,omitempty" json:"access_level,omitempty"`
	AccessLevelDescription *string           `url:"access_level_description,omitempty" json:"access_level_description,omitempty"`
	RequiredApprovalCount  *int              `url:"required_approvals,omitempty" json:"required_approvals,omitempty"`
	GroupInheritanceType   *int              `url:"group_inheritance_type,omitempty" json:"group_inheritance_type,omitempty"`
	Destroy                *bool             `url:"_destroy,omitempty" json:"_destroy,omitempty"`
}

// UpdateGroupProtectedEnvironment updates a single group-level protected
// environment.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#update-a-protected-environment
func (s *GroupProtectedEnvironmentsService) UpdateGroupProtectedEnvironment(gid interface{}, environment string, opt *UpdateGroupProtectedEnvironmentOptions, options ...RequestOptionFunc) (*GroupProtectedEnvironment, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/protected_environments/%s", PathEscape(group), environment)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	pe := new(GroupProtectedEnvironment)
	resp, err := s.client.Do(req, pe)
	if err != nil {
		return nil, resp, err
	}

	return pe, resp, nil
}

// UnprotectGroupEnvironment unprotects the given protected group-level
// environment.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_protected_environments.html#unprotect-a-single-environment
func (s *GroupProtectedEnvironmentsService) UnprotectGroupEnvironment(gid interface{}, environment string, options ...RequestOptionFunc) (*Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("groups/%s/protected_environments/%s", PathEscape(group), environment)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
