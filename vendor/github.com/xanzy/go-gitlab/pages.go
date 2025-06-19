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

type PagesService struct {
	client *Client
}

// Pages represents the Pages of a project.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/pages.html
type Pages struct {
	URL                   string             `json:"url"`
	IsUniqueDomainEnabled bool               `json:"is_unique_domain_enabled"`
	ForceHTTPS            bool               `json:"force_https"`
	Deployments           []*PagesDeployment `json:"deployments"`
}

// PagesDeployment represents a Pages deployment.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/pages.html
type PagesDeployment struct {
	CreatedAt     time.Time `json:"created_at"`
	URL           string    `json:"url"`
	PathPrefix    string    `json:"path_prefix"`
	RootDirectory string    `json:"root_directory"`
}

// UnpublishPages unpublished pages. The user must have admin privileges.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/pages.html#unpublish-pages
func (s *PagesService) UnpublishPages(gid interface{}, options ...RequestOptionFunc) (*Response, error) {
	page, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/pages", PathEscape(page))

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// GetPages lists Pages settings for a project. The user must have at least
// maintainer privileges.
//
// GitLab API Docs:
// https://docs.gitlab.com/ee/api/pages.html#get-pages-settings-for-a-project
func (s *PagesService) GetPages(gid interface{}, options ...RequestOptionFunc) (*Pages, *Response, error) {
	project, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/pages", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(Pages)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// UpdatePages represents the available UpdatePages() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/pages.html#update-pages-settings-for-a-project
type UpdatePagesOptions struct {
	PagesUniqueDomainEnabled *bool `url:"pages_unique_domain_enabled,omitempty" json:"pages_unique_domain_enabled,omitempty"`
	PagesHTTPSOnly           *bool `url:"pages_https_only,omitempty" json:"pages_https_only,omitempty"`
}

// UpdatePages updates Pages settings for a project. The user must have
// administrator privileges.
//
// GitLab API Docs:
// https://docs.gitlab.com/ee/api/pages.html#update-pages-settings-for-a-project
func (s *PagesService) UpdatePages(pid interface{}, opt UpdatePagesOptions, options ...RequestOptionFunc) (*Pages, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/pages", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPatch, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(Pages)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}
