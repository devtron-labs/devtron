//
// Copyright 2021, Sander van Harmelen
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
	"time"
)

// ResourceGroupService handles communication with the resource
// group related methods of the GitLab API.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_groups.html
type ResourceGroupService struct {
	client *Client
}

// ResourceGrouop represents a GitLab Project Resource Group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_groups.html
type ResourceGroup struct {
	ID          int        `json:"id"`
	Key         string     `json:"key"`
	ProcessMode string     `json:"process_mode"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// Gets a string representation of a ResourceGroup
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_groups.html
func (rg ResourceGroup) String() string {
	return Stringify(rg)
}

// GetAllResourceGroupsForAProject allows you to get all resource
// groups associated with a given project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_groups.html#get-all-resource-groups-for-a-project
func (s *ResourceGroupService) GetAllResourceGroupsForAProject(pid interface{}, options ...RequestOptionFunc) ([]*ResourceGroup, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/resource_groups", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var rgs []*ResourceGroup
	resp, err := s.client.Do(req, &rgs)
	if err != nil {
		return nil, resp, err
	}

	return rgs, resp, nil
}

// GetASpecificResourceGroup allows you to get a specific
// resource group for a given project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_groups.html#get-a-specific-resource-group
func (s *ResourceGroupService) GetASpecificResourceGroup(pid interface{}, key string, options ...RequestOptionFunc) (*ResourceGroup, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/resource_groups/%s", PathEscape(project), key)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	rg := new(ResourceGroup)
	resp, err := s.client.Do(req, rg)
	if err != nil {
		return nil, resp, err
	}

	return rg, resp, nil
}

// ListUpcomingJobsForASpecificResourceGroup allows you to get all
// upcoming jobs for a specific resource group for a given project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_groups.html#list-upcoming-jobs-for-a-specific-resource-group
func (s *ResourceGroupService) ListUpcomingJobsForASpecificResourceGroup(pid interface{}, key string, options ...RequestOptionFunc) ([]*Job, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/resource_groups/%s/upcoming_jobs", PathEscape(project), key)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var js []*Job
	resp, err := s.client.Do(req, &js)
	if err != nil {
		return nil, resp, err
	}

	return js, resp, nil
}

// EditAnExistingResourceGroupOptions represents the available
// EditAnExistingResourceGroup options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_groups.html#edit-an-existing-resource-group
type EditAnExistingResourceGroupOptions struct {
	ProcessMode *ResourceGroupProcessMode `url:"process_mode,omitempty" json:"process_mode,omitempty"`
}

// EditAnExistingResourceGroup allows you to edit a specific
// resource group for a given project
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_groups.html#edit-an-existing-resource-group
func (s *ResourceGroupService) EditAnExistingResourceGroup(pid interface{}, key string, opts *EditAnExistingResourceGroupOptions, options ...RequestOptionFunc) (*ResourceGroup, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/resource_groups/%s", PathEscape(project), key)

	req, err := s.client.NewRequest(http.MethodPut, u, opts, options)
	if err != nil {
		return nil, nil, err
	}

	rg := new(ResourceGroup)
	resp, err := s.client.Do(req, rg)
	if err != nil {
		return nil, resp, err
	}

	return rg, resp, nil
}
