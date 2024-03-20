package bitbucket

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type Issues struct {
	c *Client
}

func (p *Issues) Gets(io *IssuesOptions) (interface{}, error) {
	url, err := url.Parse(p.c.GetApiBaseURL() + "/repositories/" + io.Owner + "/" + io.RepoSlug + "/issues/")
	if err != nil {
		return nil, err
	}

	if io.States != nil && len(io.States) != 0 {
		query := url.Query()
		for _, state := range io.States {
			query.Set("state", state)
		}
		url.RawQuery = query.Encode()
	}

	if io.Query != "" {
		query := url.Query()
		query.Set("q", io.Query)
		url.RawQuery = query.Encode()
	}

	if io.Sort != "" {
		query := url.Query()
		query.Set("sort", io.Sort)
		url.RawQuery = query.Encode()
	}

	return p.c.executePaginated("GET", url.String(), "", nil)
}

func (p *Issues) Get(io *IssuesOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + io.Owner + "/" + io.RepoSlug + "/issues/" + io.ID
	return p.c.execute("GET", urlStr, "")
}

func (p *Issues) Delete(io *IssuesOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + io.Owner + "/" + io.RepoSlug + "/issues/" + io.ID
	return p.c.execute("DELETE", urlStr, "")
}

func (p *Issues) Update(io *IssuesOptions) (interface{}, error) {
	data, err := p.buildIssueBody(io)
	if err != nil {
		return nil, err
	}
	urlStr := p.c.requestUrl("/repositories/%s/%s/issues/%s", io.Owner, io.RepoSlug, io.ID)
	return p.c.execute("PUT", urlStr, data)
}

func (p *Issues) Create(io *IssuesOptions) (interface{}, error) {
	data, err := p.buildIssueBody(io)
	if err != nil {
		return nil, err
	}
	urlStr := p.c.requestUrl("/repositories/%s/%s/issues", io.Owner, io.RepoSlug)
	return p.c.execute("POST", urlStr, data)
}

func (p *Issues) GetVote(io *IssuesOptions) (bool, interface{}, error) {
	// A 404 indicates that the user hasn't voted
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + io.Owner + "/" + io.RepoSlug + "/issues/" + io.ID + "/vote"
	data, err := p.c.execute("GET", urlStr, "")
	if err != nil && strings.HasPrefix(err.Error(), "404") {
		return false, data, nil
	}
	return true, nil, err
}

func (p *Issues) PutVote(io *IssuesOptions) error {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + io.Owner + "/" + io.RepoSlug + "/issues/" + io.ID + "/vote"
	_, err := p.c.execute("PUT", urlStr, "")
	return err
}

func (p *Issues) DeleteVote(io *IssuesOptions) error {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + io.Owner + "/" + io.RepoSlug + "/issues/" + io.ID + "/vote"
	_, err := p.c.execute("DELETE", urlStr, "")
	return err
}

func (p *Issues) GetWatch(io *IssuesOptions) (bool, interface{}, error) {
	// A 404 indicates that the user hasn't watchd
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + io.Owner + "/" + io.RepoSlug + "/issues/" + io.ID + "/watch"
	data, err := p.c.execute("GET", urlStr, "")
	if err != nil && strings.HasPrefix(err.Error(), "404") {
		return false, data, nil
	}
	return true, nil, err
}

func (p *Issues) PutWatch(io *IssuesOptions) error {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + io.Owner + "/" + io.RepoSlug + "/issues/" + io.ID + "/watch"
	_, err := p.c.execute("PUT", urlStr, "")
	return err
}

func (p *Issues) DeleteWatch(io *IssuesOptions) error {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + io.Owner + "/" + io.RepoSlug + "/issues/" + io.ID + "/watch"
	_, err := p.c.execute("DELETE", urlStr, "")
	return err
}

func (p *Issues) buildIssueBody(io *IssuesOptions) (string, error) {
	body := map[string]interface{}{}

	// This feld is required
	if io.Title != "" {
		body["title"] = io.Title
	}

	if io.Content != "" {
		body["content"] = map[string]interface{}{
			"raw": io.Content,
		}
	}

	if io.State != "" {
		body["state"] = io.State
	}

	if io.Kind != "" {
		body["kind"] = io.Kind
	}

	if io.Priority != "" {
		body["priority"] = io.Priority
	}

	if io.Milestone != "" {
		body["milestone"] = map[string]interface{}{
			"name": io.Milestone,
		}
	}

	if io.Component != "" {
		body["component"] = map[string]interface{}{
			"name": io.Component,
		}
	}

	if io.Version != "" {
		body["version"] = map[string]interface{}{
			"name": io.Version,
		}
	}
	if io.Assignee != "" {
		body["assignee"] = map[string]interface{}{
			"uuid": io.Assignee,
		}
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *Issues) GetComments(ico *IssueCommentsOptions) (interface{}, error) {
	url, err := url.Parse(p.c.GetApiBaseURL() + "/repositories/" + ico.Owner + "/" + ico.RepoSlug + "/issues/" + ico.ID + "/comments")
	if err != nil {
		return nil, err
	}

	if ico.Query != "" {
		query := url.Query()
		query.Set("q", ico.Query)
		url.RawQuery = query.Encode()
	}

	if ico.Sort != "" {
		query := url.Query()
		query.Set("sort", ico.Sort)
		url.RawQuery = query.Encode()
	}
	return p.c.execute("GET", url.String(), "")
}

func (p *Issues) CreateComment(ico *IssueCommentsOptions) (interface{}, error) {
	urlStr := p.c.requestUrl("/repositories/%s/%s/issues/%s/comments", ico.Owner, ico.RepoSlug, ico.ID)
	// as the body/map only takes a single value, I do not think it's useful to create a seperate method here

	data, err := p.buildCommentBody(ico)
	if err != nil {
		return nil, err
	}

	return p.c.execute("POST", urlStr, data)
}

func (p *Issues) GetComment(ico *IssueCommentsOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + ico.Owner + "/" + ico.RepoSlug + "/issues/" + ico.ID + "/comments/" + ico.CommentID
	return p.c.execute("GET", urlStr, "")
}

func (p *Issues) UpdateComment(ico *IssueCommentsOptions) (interface{}, error) {
	urlStr := p.c.requestUrl("/repositories/%s/%s/issues/%s/comments/%s", ico.Owner, ico.RepoSlug, ico.ID, ico.CommentID)
	// as the body/map only takes a single value, I do not think it's useful to create a seperate method here

	data, err := p.buildCommentBody(ico)
	if err != nil {
		return nil, err
	}

	return p.c.execute("PUT", urlStr, data)

}

func (p *Issues) DeleteComment(ico *IssueCommentsOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + ico.Owner + "/" + ico.RepoSlug + "/issues/" + ico.ID + "/comments/" + ico.CommentID
	return p.c.execute("DELETE", urlStr, "")
}

func (p *Issues) buildCommentBody(ico *IssueCommentsOptions) (string, error) {
	body := map[string]interface{}{}
	body["content"] = map[string]interface{}{
		"raw": ico.CommentContent,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *Issues) GetChanges(ico *IssueChangesOptions) (interface{}, error) {
	url, err := url.Parse(p.c.GetApiBaseURL() + "/repositories/" + ico.Owner + "/" + ico.RepoSlug + "/issues/" + ico.ID + "/changes")
	if err != nil {
		return nil, err
	}

	if ico.Query != "" {
		query := url.Query()
		query.Set("q", ico.Query)
		url.RawQuery = query.Encode()
	}

	if ico.Sort != "" {
		query := url.Query()
		query.Set("sort", ico.Sort)
		url.RawQuery = query.Encode()
	}

	return p.c.execute("GET", url.String(), "")
}

func (p *Issues) CreateChange(ico *IssueChangesOptions) (interface{}, error) {
	url, err := url.Parse(p.c.GetApiBaseURL() + "/repositories/" + ico.Owner + "/" + ico.RepoSlug + "/issues/" + ico.ID + "/changes")
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{}
	if ico.Message != "" {
		body["message"] = map[string]interface{}{
			"raw": ico.Message,
		}
	}

	changes := map[string]interface{}{}
	for _, change := range ico.Changes {
		changes[change.Type] = map[string]interface{}{
			"new": change.NewValue,
		}
	}
	if len(changes) > 0 {
		body["changes"] = changes
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	fmt.Printf("data %s", data)

	return p.c.execute("POST", url.String(), string(data))
}

func (p *Issues) GetChange(ico *IssueChangesOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + ico.Owner + "/" + ico.RepoSlug + "/issues/" + ico.ID + "/changes/" + ico.ChangeID
	return p.c.execute("GET", urlStr, "")
}
