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

// SnippetRepositoryStorageMoveService handles communication with the
// snippets related methods of the GitLab API.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/snippet_repository_storage_moves.html
type SnippetRepositoryStorageMoveService struct {
	client *Client
}

// SnippetRepositoryStorageMove represents the status of a repository move.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/snippet_repository_storage_moves.html
type SnippetRepositoryStorageMove struct {
	ID                     int                `json:"id"`
	CreatedAt              *time.Time         `json:"created_at"`
	State                  string             `json:"state"`
	SourceStorageName      string             `json:"source_storage_name"`
	DestinationStorageName string             `json:"destination_storage_name"`
	Snippet                *RepositorySnippet `json:"snippet"`
}

type RepositorySnippet struct {
	ID            int             `json:"id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	Visibility    VisibilityValue `json:"visibility"`
	UpdatedAt     *time.Time      `json:"updated_at"`
	CreatedAt     *time.Time      `json:"created_at"`
	ProjectID     int             `json:"project_id"`
	WebURL        string          `json:"web_url"`
	RawURL        string          `json:"raw_url"`
	SSHURLToRepo  string          `json:"ssh_url_to_repo"`
	HTTPURLToRepo string          `json:"http_url_to_repo"`
}

// RetrieveAllSnippetStorageMovesOptions represents the available
// RetrieveAllStorageMoves() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/snippet_repository_storage_moves.html#retrieve-all-repository-storage-moves-for-a-snippet
type RetrieveAllSnippetStorageMovesOptions ListOptions

// RetrieveAllStorageMoves retrieves all snippet repository storage moves
// accessible by the authenticated user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/snippet_repository_storage_moves.html#retrieve-all-repository-storage-moves-for-a-snippet
func (s SnippetRepositoryStorageMoveService) RetrieveAllStorageMoves(opts RetrieveAllSnippetStorageMovesOptions, options ...RequestOptionFunc) ([]*SnippetRepositoryStorageMove, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "snippet_repository_storage_moves", opts, options)
	if err != nil {
		return nil, nil, err
	}

	var ssms []*SnippetRepositoryStorageMove
	resp, err := s.client.Do(req, &ssms)
	if err != nil {
		return nil, resp, err
	}

	return ssms, resp, err
}

// RetrieveAllStorageMovesForSnippet retrieves all repository storage moves for
// a single snippet accessible by the authenticated user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/snippet_repository_storage_moves.html#retrieve-all-repository-storage-moves-for-a-snippet
func (s SnippetRepositoryStorageMoveService) RetrieveAllStorageMovesForSnippet(snippet int, opts RetrieveAllSnippetStorageMovesOptions, options ...RequestOptionFunc) ([]*SnippetRepositoryStorageMove, *Response, error) {
	u := fmt.Sprintf("snippets/%d/repository_storage_moves", snippet)

	req, err := s.client.NewRequest(http.MethodGet, u, opts, options)
	if err != nil {
		return nil, nil, err
	}

	var ssms []*SnippetRepositoryStorageMove
	resp, err := s.client.Do(req, &ssms)
	if err != nil {
		return nil, resp, err
	}

	return ssms, resp, err
}

// GetStorageMove gets a single snippet repository storage move.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/snippet_repository_storage_moves.html#get-a-single-snippet-repository-storage-move
func (s SnippetRepositoryStorageMoveService) GetStorageMove(repositoryStorage int, options ...RequestOptionFunc) (*SnippetRepositoryStorageMove, *Response, error) {
	u := fmt.Sprintf("snippet_repository_storage_moves/%d", repositoryStorage)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	ssm := new(SnippetRepositoryStorageMove)
	resp, err := s.client.Do(req, ssm)
	if err != nil {
		return nil, resp, err
	}

	return ssm, resp, err
}

// GetStorageMoveForSnippet gets a single repository storage move for a snippet.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/snippet_repository_storage_moves.html#get-a-single-repository-storage-move-for-a-snippet
func (s SnippetRepositoryStorageMoveService) GetStorageMoveForSnippet(snippet int, repositoryStorage int, options ...RequestOptionFunc) (*SnippetRepositoryStorageMove, *Response, error) {
	u := fmt.Sprintf("snippets/%d/repository_storage_moves/%d", snippet, repositoryStorage)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	ssm := new(SnippetRepositoryStorageMove)
	resp, err := s.client.Do(req, ssm)
	if err != nil {
		return nil, resp, err
	}

	return ssm, resp, err
}

// ScheduleStorageMoveForSnippetOptions represents the available
// ScheduleStorageMoveForSnippet() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/snippet_repository_storage_moves.html#schedule-a-repository-storage-move-for-a-snippet
type ScheduleStorageMoveForSnippetOptions struct {
	DestinationStorageName *string `url:"destination_storage_name,omitempty" json:"destination_storage_name,omitempty"`
}

// ScheduleStorageMoveForSnippet schedule a repository to be moved for a snippet.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/snippet_repository_storage_moves.html#schedule-a-repository-storage-move-for-a-snippet
func (s SnippetRepositoryStorageMoveService) ScheduleStorageMoveForSnippet(snippet int, opts ScheduleStorageMoveForSnippetOptions, options ...RequestOptionFunc) (*SnippetRepositoryStorageMove, *Response, error) {
	u := fmt.Sprintf("snippets/%d/repository_storage_moves", snippet)

	req, err := s.client.NewRequest(http.MethodPost, u, opts, options)
	if err != nil {
		return nil, nil, err
	}

	ssm := new(SnippetRepositoryStorageMove)
	resp, err := s.client.Do(req, ssm)
	if err != nil {
		return nil, resp, err
	}

	return ssm, resp, err
}

// ScheduleAllSnippetStorageMovesOptions represents the available
// ScheduleAllStorageMoves() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/snippet_repository_storage_moves.html#schedule-repository-storage-moves-for-all-snippets-on-a-storage-shard
type ScheduleAllSnippetStorageMovesOptions struct {
	SourceStorageName      *string `url:"source_storage_name,omitempty" json:"source_storage_name,omitempty"`
	DestinationStorageName *string `url:"destination_storage_name,omitempty" json:"destination_storage_name,omitempty"`
}

// ScheduleAllStorageMoves schedules all snippet repositories to be moved.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/snippet_repository_storage_moves.html#schedule-repository-storage-moves-for-all-snippets-on-a-storage-shard
func (s SnippetRepositoryStorageMoveService) ScheduleAllStorageMoves(opts ScheduleAllSnippetStorageMovesOptions, options ...RequestOptionFunc) (*Response, error) {
	req, err := s.client.NewRequest(http.MethodPost, "snippet_repository_storage_moves", opts, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
