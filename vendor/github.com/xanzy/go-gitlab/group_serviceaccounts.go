//
// Copyright 2023, James Hong
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

// GroupServiceAccount represents a GitLab service account user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#create-service-account-user
type GroupServiceAccount struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	UserName string `json:"username"`
}

// ListServiceAccountsOptions represents the available ListServiceAccounts() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/group_service_accounts.html#list-service-account-users
type ListServiceAccountsOptions struct {
	ListOptions
	OrderBy *string `url:"order_by,omitempty" json:"order_by,omitempty"`
	Sort    *string `url:"sort,omitempty" json:"sort,omitempty"`
}

// ListServiceAccounts gets a list of service acxcounts.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/group_service_accounts.html#list-service-account-users
func (s *GroupsService) ListServiceAccounts(gid interface{}, opt *ListServiceAccountsOptions, options ...RequestOptionFunc) ([]*GroupServiceAccount, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/service_accounts", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var sa []*GroupServiceAccount
	resp, err := s.client.Do(req, &sa)
	if err != nil {
		return nil, resp, err
	}

	return sa, resp, nil
}

// CreateServiceAccountOptions represents the available CreateServiceAccount() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/group_service_accounts.html#create-a-service-account-user
type CreateServiceAccountOptions struct {
	Name     *string `url:"name,omitempty" json:"name,omitempty"`
	Username *string `url:"username,omitempty" json:"username,omitempty"`
}

// Creates a service account user.
//
// This API endpoint works on top-level groups only. It does not work on subgroups.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#create-service-account-user
func (s *GroupsService) CreateServiceAccount(gid interface{}, opt *CreateServiceAccountOptions, options ...RequestOptionFunc) (*GroupServiceAccount, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/service_accounts", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	sa := new(GroupServiceAccount)
	resp, err := s.client.Do(req, sa)
	if err != nil {
		return nil, resp, err
	}

	return sa, resp, nil
}

// CreateServiceAccountPersonalAccessTokenOptions represents the available
// CreateServiceAccountPersonalAccessToken() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_service_accounts.html#create-a-personal-access-token-for-a-service-account-user
type CreateServiceAccountPersonalAccessTokenOptions struct {
	Scopes    *[]string `url:"scopes,omitempty" json:"scopes,omitempty"`
	Name      *string   `url:"name,omitempty" json:"name,omitempty"`
	ExpiresAt *ISOTime  `url:"expires_at,omitempty" json:"expires_at,omitempty"`
}

// CreateServiceAccountPersonalAccessToken add a new Personal Access Token for a
// service account user for a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_service_accounts.html#create-a-personal-access-token-for-a-service-account-user
func (s *GroupsService) CreateServiceAccountPersonalAccessToken(gid interface{}, serviceAccount int, opt *CreateServiceAccountPersonalAccessTokenOptions, options ...RequestOptionFunc) (*PersonalAccessToken, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/service_accounts/%d/personal_access_tokens", PathEscape(group), serviceAccount)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	pat := new(PersonalAccessToken)
	resp, err := s.client.Do(req, pat)
	if err != nil {
		return nil, resp, err
	}

	return pat, resp, nil
}

// RotateServiceAccountPersonalAccessToken rotates a Personal Access Token for a
// service account user for a group.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#create-personal-access-token-for-service-account-user
func (s *GroupsService) RotateServiceAccountPersonalAccessToken(gid interface{}, serviceAccount, token int, options ...RequestOptionFunc) (*PersonalAccessToken, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/service_accounts/%d/personal_access_tokens/%d/rotate", PathEscape(group), serviceAccount, token)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	pat := new(PersonalAccessToken)
	resp, err := s.client.Do(req, pat)
	if err != nil {
		return nil, resp, err
	}

	return pat, resp, nil
}

// DeleteServiceAccount Deletes a service account user.
//
// This API endpoint works on top-level groups only. It does not work on subgroups.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/group_service_accounts.html#delete-a-service-account-user
func (s *GroupsService) DeleteServiceAccount(gid interface{}, serviceAccount int, options ...RequestOptionFunc) (*Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("groups/%s/service_accounts/%d", PathEscape(group), serviceAccount)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
