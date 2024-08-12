// Copyright 2021, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package gitlab

import (
	"fmt"
	"net/http"
)

// JobTokenScopeService handles communication with project CI settings
// such as token permissions.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/project_job_token_scopes.html
type JobTokenScopeService struct {
	client *Client
}

// JobTokenAccessSettings represents job token access attributes for this project.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/project_job_token_scopes.html
type JobTokenAccessSettings struct {
	InboundEnabled  bool `json:"inbound_enabled"`
	OutboundEnabled bool `json:"outbound_enabled"`
}

// GetProjectJobTokenAccessSettings fetch the CI/CD job token access settings (job token scope) of a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#get-a-projects-cicd-job-token-access-settings
func (j *JobTokenScopeService) GetProjectJobTokenAccessSettings(pid interface{}, options ...RequestOptionFunc) (*JobTokenAccessSettings, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf(`projects/%s/job_token_scope`, PathEscape(project))

	req, err := j.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	jt := new(JobTokenAccessSettings)
	resp, err := j.client.Do(req, jt)
	if err != nil {
		return nil, resp, err
	}

	return jt, resp, err
}

// PatchProjectJobTokenAccessSettingsOptions represents the available
// PatchProjectJobTokenAccessSettings() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#patch-a-projects-cicd-job-token-access-settings
type PatchProjectJobTokenAccessSettingsOptions struct {
	Enabled bool `json:"enabled"`
}

// PatchProjectJobTokenAccessSettings patch the Limit access to this project setting (job token scope) of a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#patch-a-projects-cicd-job-token-access-settings
func (j *JobTokenScopeService) PatchProjectJobTokenAccessSettings(pid interface{}, opt *PatchProjectJobTokenAccessSettingsOptions, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf(`projects/%s/job_token_scope`, PathEscape(project))

	req, err := j.client.NewRequest(http.MethodPatch, u, opt, options)
	if err != nil {
		return nil, err
	}

	return j.client.Do(req, nil)
}

// JobTokenInboundAllowItem represents a single job token inbound allowlist item.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/project_job_token_scopes.html
type JobTokenInboundAllowItem struct {
	SourceProjectID int `json:"source_project_id"`
	TargetProjectID int `json:"target_project_id"`
}

// GetJobTokenInboundAllowListOptions represents the available
// GetJobTokenInboundAllowList() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#get-a-projects-cicd-job-token-inbound-allowlist
type GetJobTokenInboundAllowListOptions struct {
	ListOptions
}

// GetProjectJobTokenInboundAllowList fetches the CI/CD job token inbound
// allowlist (job token scope) of a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#get-a-projects-cicd-job-token-inbound-allowlist
func (j *JobTokenScopeService) GetProjectJobTokenInboundAllowList(pid interface{}, opt *GetJobTokenInboundAllowListOptions, options ...RequestOptionFunc) ([]*Project, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf(`projects/%s/job_token_scope/allowlist`, PathEscape(project))

	req, err := j.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ps []*Project
	resp, err := j.client.Do(req, &ps)
	if err != nil {
		return nil, resp, err
	}

	return ps, resp, nil
}

// AddProjectToJobScopeAllowListOptions represents the available
// AddProjectToJobScopeAllowList() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#create-a-new-project-to-a-projects-cicd-job-token-inbound-allowlist
type JobTokenInboundAllowOptions struct {
	TargetProjectID *int `url:"target_project_id,omitempty" json:"target_project_id,omitempty"`
}

// AddProjectToJobScopeAllowList adds a new project to a project's job token
// inbound allow list.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#create-a-new-project-to-a-projects-cicd-job-token-inbound-allowlist
func (j *JobTokenScopeService) AddProjectToJobScopeAllowList(pid interface{}, opt *JobTokenInboundAllowOptions, options ...RequestOptionFunc) (*JobTokenInboundAllowItem, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf(`projects/%s/job_token_scope/allowlist`, PathEscape(project))

	req, err := j.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	jt := new(JobTokenInboundAllowItem)
	resp, err := j.client.Do(req, jt)
	if err != nil {
		return nil, resp, err
	}

	return jt, resp, nil
}

// RemoveProjectFromJobScopeAllowList removes a project from a project's job
// token inbound allow list.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#remove-a-project-from-a-projects-cicd-job-token-inbound-allowlist
func (j *JobTokenScopeService) RemoveProjectFromJobScopeAllowList(pid interface{}, targetProject int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf(`projects/%s/job_token_scope/allowlist/%d`, PathEscape(project), targetProject)

	req, err := j.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return j.client.Do(req, nil)
}

// JobTokenAllowlistItem represents a single job token allowlist item.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/project_job_token_scopes.html
type JobTokenAllowlistItem struct {
	SourceProjectID int `json:"source_project_id"`
	TargetGroupID   int `json:"target_group_id"`
}

// GetJobTokenAllowlistGroupsOptions represents the available
// GetJobTokenAllowlistGroups() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#get-a-projects-cicd-job-token-allowlist-of-groups
type GetJobTokenAllowlistGroupsOptions struct {
	ListOptions
}

// GetJobTokenAllowListGroups fetches the CI/CD job token allowlist groups
// (job token scopes) of a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#get-a-projects-cicd-job-token-allowlist-of-groups
func (j *JobTokenScopeService) GetJobTokenAllowlistGroups(pid interface{}, opt *GetJobTokenAllowlistGroupsOptions, options ...RequestOptionFunc) ([]*Group, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf(`projects/%s/job_token_scope/groups_allowlist`, PathEscape(project))

	req, err := j.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ps []*Group
	resp, err := j.client.Do(req, &ps)
	if err != nil {
		return nil, resp, err
	}

	return ps, resp, nil
}

// AddGroupToJobTokenAllowlistOptions represents the available
// AddGroupToJobTokenAllowlist() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#add-a-group-to-a-cicd-job-token-allowlist
type AddGroupToJobTokenAllowlistOptions struct {
	TargetGroupID *int `url:"target_group_id,omitempty" json:"target_group_id,omitempty"`
}

// AddProjectToJobScopeGroupsAllowList adds a new group to a project's job token
// inbound groups allow list.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#add-a-group-to-a-cicd-job-token-allowlist
func (j *JobTokenScopeService) AddGroupToJobTokenAllowlist(pid interface{}, opt *AddGroupToJobTokenAllowlistOptions, options ...RequestOptionFunc) (*JobTokenAllowlistItem, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf(`projects/%s/job_token_scope/groups_allowlist`, PathEscape(project))

	req, err := j.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	jt := new(JobTokenAllowlistItem)
	resp, err := j.client.Do(req, jt)
	if err != nil {
		return nil, resp, err
	}

	return jt, resp, nil
}

// RemoveGroupFromJopTokenAllowlist removes a group from a project's job
// token inbound groups allow list.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_job_token_scopes.html#remove-a-group-from-a-cicd-job-token-allowlist
func (j *JobTokenScopeService) RemoveGroupFromJobTokenAllowlist(pid interface{}, targetGroup int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf(`projects/%s/job_token_scope/groups_allowlist/%d`, PathEscape(project), targetGroup)

	req, err := j.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return j.client.Do(req, nil)
}
