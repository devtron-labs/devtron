//
// Copyright 2023, Nick Westbury
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

// GroupRepositoryStorageMoveService handles communication with the
// group repositories related methods of the GitLab API.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_repository_storage_moves.html
type GroupRepositoryStorageMoveService struct {
	client *Client
}

// GroupRepositoryStorageMove represents the status of a repository move.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_repository_storage_moves.html
type GroupRepositoryStorageMove struct {
	ID                     int              `json:"id"`
	CreatedAt              *time.Time       `json:"created_at"`
	State                  string           `json:"state"`
	SourceStorageName      string           `json:"source_storage_name"`
	DestinationStorageName string           `json:"destination_storage_name"`
	Group                  *RepositoryGroup `json:"group"`
}

type RepositoryGroup struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	WebURL string `json:"web_url"`
}

// RetrieveAllGroupStorageMovesOptions represents the available
// RetrieveAllStorageMoves() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_repository_storage_moves.html#retrieve-all-group-repository-storage-moves
type RetrieveAllGroupStorageMovesOptions ListOptions

// RetrieveAllStorageMoves retrieves all group repository storage moves
// accessible by the authenticated user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_repository_storage_moves.html#retrieve-all-group-repository-storage-moves
func (g GroupRepositoryStorageMoveService) RetrieveAllStorageMoves(opts RetrieveAllGroupStorageMovesOptions, options ...RequestOptionFunc) ([]*GroupRepositoryStorageMove, *Response, error) {
	req, err := g.client.NewRequest(http.MethodGet, "group_repository_storage_moves", opts, options)
	if err != nil {
		return nil, nil, err
	}

	var gsms []*GroupRepositoryStorageMove
	resp, err := g.client.Do(req, &gsms)
	if err != nil {
		return nil, resp, err
	}

	return gsms, resp, err
}

// RetrieveAllStorageMovesForGroup retrieves all repository storage moves for
// a single group accessible by the authenticated user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_repository_storage_moves.html#retrieve-all-repository-storage-moves-for-a-single-group
func (g GroupRepositoryStorageMoveService) RetrieveAllStorageMovesForGroup(group int, opts RetrieveAllGroupStorageMovesOptions, options ...RequestOptionFunc) ([]*GroupRepositoryStorageMove, *Response, error) {
	u := fmt.Sprintf("groups/%d/repository_storage_moves", group)

	req, err := g.client.NewRequest(http.MethodGet, u, opts, options)
	if err != nil {
		return nil, nil, err
	}

	var gsms []*GroupRepositoryStorageMove
	resp, err := g.client.Do(req, &gsms)
	if err != nil {
		return nil, resp, err
	}

	return gsms, resp, err
}

// GetStorageMove gets a single group repository storage move.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_repository_storage_moves.html#get-a-single-group-repository-storage-move
func (g GroupRepositoryStorageMoveService) GetStorageMove(repositoryStorage int, options ...RequestOptionFunc) (*GroupRepositoryStorageMove, *Response, error) {
	u := fmt.Sprintf("group_repository_storage_moves/%d", repositoryStorage)

	req, err := g.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	gsm := new(GroupRepositoryStorageMove)
	resp, err := g.client.Do(req, gsm)
	if err != nil {
		return nil, resp, err
	}

	return gsm, resp, err
}

// GetStorageMoveForGroup gets a single repository storage move for a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_repository_storage_moves.html#get-a-single-repository-storage-move-for-a-group
func (g GroupRepositoryStorageMoveService) GetStorageMoveForGroup(group int, repositoryStorage int, options ...RequestOptionFunc) (*GroupRepositoryStorageMove, *Response, error) {
	u := fmt.Sprintf("groups/%d/repository_storage_moves/%d", group, repositoryStorage)

	req, err := g.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	gsm := new(GroupRepositoryStorageMove)
	resp, err := g.client.Do(req, gsm)
	if err != nil {
		return nil, resp, err
	}

	return gsm, resp, err
}

// ScheduleStorageMoveForGroupOptions represents the available
// ScheduleStorageMoveForGroup() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_repository_storage_moves.html#schedule-a-repository-storage-move-for-a-group
type ScheduleStorageMoveForGroupOptions struct {
	DestinationStorageName *string `url:"destination_storage_name,omitempty" json:"destination_storage_name,omitempty"`
}

// ScheduleStorageMoveForGroup schedule a repository to be moved for a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_repository_storage_moves.html#schedule-a-repository-storage-move-for-a-group
func (g GroupRepositoryStorageMoveService) ScheduleStorageMoveForGroup(group int, opts ScheduleStorageMoveForGroupOptions, options ...RequestOptionFunc) (*GroupRepositoryStorageMove, *Response, error) {
	u := fmt.Sprintf("groups/%d/repository_storage_moves", group)

	req, err := g.client.NewRequest(http.MethodPost, u, opts, options)
	if err != nil {
		return nil, nil, err
	}

	gsm := new(GroupRepositoryStorageMove)
	resp, err := g.client.Do(req, gsm)
	if err != nil {
		return nil, resp, err
	}

	return gsm, resp, err
}

// ScheduleAllGroupStorageMovesOptions represents the available
// ScheduleAllStorageMoves() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_repository_storage_moves.html#schedule-repository-storage-moves-for-all-groups-on-a-storage-shard
type ScheduleAllGroupStorageMovesOptions struct {
	SourceStorageName      *string `url:"source_storage_name,omitempty" json:"source_storage_name,omitempty"`
	DestinationStorageName *string `url:"destination_storage_name,omitempty" json:"destination_storage_name,omitempty"`
}

// ScheduleAllStorageMoves schedules all group repositories to be moved.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_repository_storage_moves.html#schedule-repository-storage-moves-for-all-groups-on-a-storage-shard
func (g GroupRepositoryStorageMoveService) ScheduleAllStorageMoves(opts ScheduleAllGroupStorageMovesOptions, options ...RequestOptionFunc) (*Response, error) {
	req, err := g.client.NewRequest(http.MethodPost, "group_repository_storage_moves", opts, options)
	if err != nil {
		return nil, err
	}

	return g.client.Do(req, nil)
}
