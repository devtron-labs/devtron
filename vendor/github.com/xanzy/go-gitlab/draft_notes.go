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
)

type DraftNote struct {
	ID                int           `json:"id"`
	AuthorID          int           `json:"author_id"`
	MergeRequestID    int           `json:"merge_request_id"`
	ResolveDiscussion bool          `json:"resolve_discussion"`
	DiscussionID      string        `json:"discussion_id"`
	Note              string        `json:"note"`
	CommitID          string        `json:"commit_id"`
	LineCode          string        `json:"line_code"`
	Position          *NotePosition `json:"position"`
}

// DraftNotesService handles communication with the draft notes related methods
// of the GitLab API.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/draft_notes.html#list-all-merge-request-draft-notes
type DraftNotesService struct {
	client *Client
}

// ListDraftNotesOptions represents the available ListDraftNotes()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/draft_notes.html#list-all-merge-request-draft-notes
type ListDraftNotesOptions struct {
	ListOptions
	OrderBy *string `url:"order_by,omitempty" json:"order_by,omitempty"`
	Sort    *string `url:"sort,omitempty" json:"sort,omitempty"`
}

// ListDraftNotes gets a list of all draft notes for a merge request.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/draft_notes.html#list-all-merge-request-draft-notes
func (s *DraftNotesService) ListDraftNotes(pid interface{}, mergeRequest int, opt *ListDraftNotesOptions, options ...RequestOptionFunc) ([]*DraftNote, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/draft_notes", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var n []*DraftNote
	resp, err := s.client.Do(req, &n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// GetDraftNote gets a single draft note for a merge request.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/draft_notes.html#get-a-single-draft-note
func (s *DraftNotesService) GetDraftNote(pid interface{}, mergeRequest int, note int, options ...RequestOptionFunc) (*DraftNote, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/draft_notes/%d", PathEscape(project), mergeRequest, note)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(DraftNote)
	resp, err := s.client.Do(req, &n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// CreateDraftNoteOptions represents the available CreateDraftNote()
// options.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/draft_notes.html#create-a-draft-note
type CreateDraftNoteOptions struct {
	Note                  *string          `url:"note" json:"note"`
	CommitID              *string          `url:"commit_id,omitempty" json:"commit_id,omitempty"`
	InReplyToDiscussionID *string          `url:"in_reply_to_discussion_id,omitempty" json:"in_reply_to_discussion_id,omitempty"`
	ResolveDiscussion     *bool            `url:"resolve_discussion,omitempty" json:"resolve_discussion,omitempty"`
	Position              *PositionOptions `url:"position,omitempty" json:"position,omitempty"`
}

// CreateDraftNote creates a draft note for a merge request.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/draft_notes.html#create-a-draft-note
func (s *DraftNotesService) CreateDraftNote(pid interface{}, mergeRequest int, opt *CreateDraftNoteOptions, options ...RequestOptionFunc) (*DraftNote, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/draft_notes", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(DraftNote)
	resp, err := s.client.Do(req, &n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// UpdateDraftNoteOptions represents the available UpdateDraftNote()
// options.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/draft_notes.html#create-a-draft-note
type UpdateDraftNoteOptions struct {
	Note     *string          `url:"note,omitempty" json:"note,omitempty"`
	Position *PositionOptions `url:"position,omitempty" json:"position,omitempty"`
}

// UpdateDraftNote updates a draft note for a merge request.
//
// Gitlab API docs: https://docs.gitlab.com/ee/api/draft_notes.html#create-a-draft-note
func (s *DraftNotesService) UpdateDraftNote(pid interface{}, mergeRequest int, note int, opt *UpdateDraftNoteOptions, options ...RequestOptionFunc) (*DraftNote, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/draft_notes/%d", PathEscape(project), mergeRequest, note)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(DraftNote)
	resp, err := s.client.Do(req, &n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// DeleteDraftNote deletes a single draft note for a merge request.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/draft_notes.html#delete-a-draft-note
func (s *DraftNotesService) DeleteDraftNote(pid interface{}, mergeRequest int, note int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/draft_notes/%d", PathEscape(project), mergeRequest, note)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// PublishDraftNote publishes a single draft note for a merge request.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/draft_notes.html#publish-a-draft-note
func (s *DraftNotesService) PublishDraftNote(pid interface{}, mergeRequest int, note int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/draft_notes/%d/publish", PathEscape(project), mergeRequest, note)

	req, err := s.client.NewRequest(http.MethodPut, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// PublishAllDraftNotes publishes all draft notes for a merge request that belong to the user.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/draft_notes.html#publish-a-draft-note
func (s *DraftNotesService) PublishAllDraftNotes(pid interface{}, mergeRequest int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/draft_notes/bulk_publish", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
