package bitbucket

import (
	"encoding/json"
	"net/url"
)

type Commits struct {
	c *Client
}

func (cm *Commits) GetCommits(cmo *CommitsOptions) (interface{}, error) {
	urlStr := cm.c.requestUrl("/repositories/%s/%s/commits/%s", cmo.Owner, cmo.RepoSlug, cmo.Branchortag)
	urlStr += cm.buildCommitsQuery(cmo.Include, cmo.Exclude)
	return cm.c.executePaginated("GET", urlStr, "", cmo.Page)
}

func (cm *Commits) GetCommit(cmo *CommitsOptions) (interface{}, error) {
	urlStr := cm.c.requestUrl("/repositories/%s/%s/commit/%s", cmo.Owner, cmo.RepoSlug, cmo.Revision)
	return cm.c.execute("GET", urlStr, "")
}

func (cm *Commits) GetCommitComments(cmo *CommitsOptions) (interface{}, error) {
	urlStr := cm.c.requestUrl("/repositories/%s/%s/commit/%s/comments", cmo.Owner, cmo.RepoSlug, cmo.Revision)
	return cm.c.executePaginated("GET", urlStr, "", nil)
}

func (cm *Commits) GetCommitComment(cmo *CommitsOptions) (interface{}, error) {
	urlStr := cm.c.requestUrl("/repositories/%s/%s/commit/%s/comments/%s", cmo.Owner, cmo.RepoSlug, cmo.Revision, cmo.CommentID)
	return cm.c.execute("GET", urlStr, "")
}

func (cm *Commits) GetCommitStatuses(cmo *CommitsOptions) (interface{}, error) {
	urlStr := cm.c.requestUrl("/repositories/%s/%s/commit/%s/statuses", cmo.Owner, cmo.RepoSlug, cmo.Revision)
	return cm.c.executePaginated("GET", urlStr, "", nil)
}

func (cm *Commits) GetCommitStatus(cmo *CommitsOptions, commitStatusKey string) (interface{}, error) {
	urlStr := cm.c.requestUrl("/repositories/%s/%s/commit/%s/statuses/build/%s", cmo.Owner, cmo.RepoSlug, cmo.Revision, commitStatusKey)
	return cm.c.execute("GET", urlStr, "")
}

func (cm *Commits) GiveApprove(cmo *CommitsOptions) (interface{}, error) {
	urlStr := cm.c.requestUrl("/repositories/%s/%s/commit/%s/approve", cmo.Owner, cmo.RepoSlug, cmo.Revision)
	return cm.c.execute("POST", urlStr, "")
}

func (cm *Commits) RemoveApprove(cmo *CommitsOptions) (interface{}, error) {
	urlStr := cm.c.requestUrl("/repositories/%s/%s/commit/%s/approve", cmo.Owner, cmo.RepoSlug, cmo.Revision)
	return cm.c.execute("DELETE", urlStr, "")
}

func (cm *Commits) CreateCommitStatus(cmo *CommitsOptions, cso *CommitStatusOptions) (interface{}, error) {
	urlStr := cm.c.requestUrl("/repositories/%s/%s/commit/%s/statuses/build", cmo.Owner, cmo.RepoSlug, cmo.Revision)
	data, err := json.Marshal(cso)
	if err != nil {
		return nil, err
	}
	return cm.c.execute("POST", urlStr, string(data))
}

func (cm *Commits) buildCommitsQuery(include, exclude string) string {

	p := url.Values{}

	if include != "" {
		p.Add("include", include)
	}
	if exclude != "" {
		p.Add("exclude", exclude)
	}

	if res := p.Encode(); len(res) > 0 {
		return "?" + res
	}
	return ""
}
