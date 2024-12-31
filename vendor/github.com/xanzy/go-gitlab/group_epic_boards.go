//
// Copyright 2021, Patrick Webster
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

// GroupEpicBoardsService handles communication with the group epic board
// related methods of the GitLab API.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_epic_boards.html
type GroupEpicBoardsService struct {
	client *Client
}

// GroupEpicBoard represents a GitLab group epic board.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_epic_boards.html
type GroupEpicBoard struct {
	ID     int             `json:"id"`
	Name   string          `json:"name"`
	Group  *Group          `json:"group"`
	Labels []*LabelDetails `json:"labels"`
	Lists  []*BoardList    `json:"lists"`
}

func (b GroupEpicBoard) String() string {
	return Stringify(b)
}

// ListGroupEpicBoardsOptions represents the available
// ListGroupEpicBoards() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_epic_boards.html#list-all-epic-boards-in-a-group
type ListGroupEpicBoardsOptions ListOptions

// ListGroupEpicBoards gets a list of all epic boards in a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_epic_boards.html#list-all-epic-boards-in-a-group
func (s *GroupEpicBoardsService) ListGroupEpicBoards(gid interface{}, opt *ListGroupEpicBoardsOptions, options ...RequestOptionFunc) ([]*GroupEpicBoard, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/epic_boards", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var gs []*GroupEpicBoard
	resp, err := s.client.Do(req, &gs)
	if err != nil {
		return nil, resp, err
	}

	return gs, resp, nil
}

// GetGroupEpicBoard gets a single epic board of a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_epic_boards.html#single-group-epic-board
func (s *GroupEpicBoardsService) GetGroupEpicBoard(gid interface{}, board int, options ...RequestOptionFunc) (*GroupEpicBoard, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/epic_boards/%d", PathEscape(group), board)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	gib := new(GroupEpicBoard)
	resp, err := s.client.Do(req, gib)
	if err != nil {
		return nil, resp, err
	}

	return gib, resp, nil
}
