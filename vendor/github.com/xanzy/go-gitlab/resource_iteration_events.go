//
// Copyright 2023, Hakki Ceylan, Yavuz Turk
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

// ResourceIterationEventsService handles communication with the event related
// methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/resource_iteration_events.html
type ResourceIterationEventsService struct {
	client *Client
}

// IterationEvent represents a resource iteration event.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/resource_iteration_events.html
type IterationEvent struct {
	ID           int        `json:"id"`
	User         *BasicUser `json:"user"`
	CreatedAt    *time.Time `json:"created_at"`
	ResourceType string     `json:"resource_type"`
	ResourceID   int        `json:"resource_id"`
	Iteration    *Iteration `json:"iteration"`
	Action       string     `json:"action"`
}

// Iteration represents a project issue iteration.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/resource_iteration_events.html
type Iteration struct {
	ID          int        `json:"id"`
	IID         int        `json:"iid"`
	Sequence    int        `json:"sequence"`
	GroupId     int        `json:"group_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	State       int        `json:"state"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	DueDate     *ISOTime   `json:"due_date"`
	StartDate   *ISOTime   `json:"start_date"`
	WebURL      string     `json:"web_url"`
}

// ListIterationEventsOptions represents the options for all resource state
// events list methods.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_iteration_events.html#list-project-issue-iteration-events
type ListIterationEventsOptions struct {
	ListOptions
}

// ListIssueIterationEvents retrieves resource iteration events for the
// specified project and issue.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_iteration_events.html#list-project-issue-iteration-events
func (s *ResourceIterationEventsService) ListIssueIterationEvents(pid interface{}, issue int, opt *ListIterationEventsOptions, options ...RequestOptionFunc) ([]*IterationEvent, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/issues/%d/resource_iteration_events", PathEscape(project), issue)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ies []*IterationEvent
	resp, err := s.client.Do(req, &ies)
	if err != nil {
		return nil, resp, err
	}

	return ies, resp, nil
}

// GetIssueIterationEvent gets a single issue iteration event.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_iteration_events.html#get-single-issue-iteration-event
func (s *ResourceIterationEventsService) GetIssueIterationEvent(pid interface{}, issue int, event int, options ...RequestOptionFunc) (*IterationEvent, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/issues/%d/resource_iteration_events/%d", PathEscape(project), issue, event)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	ie := new(IterationEvent)
	resp, err := s.client.Do(req, ie)
	if err != nil {
		return nil, resp, err
	}

	return ie, resp, nil
}
