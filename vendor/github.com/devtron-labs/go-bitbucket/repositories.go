package bitbucket

import (
	"errors"
	"fmt"
)

//"github.com/k0kubun/pp"

type Repositories struct {
	c                  *Client
	PullRequests       *PullRequests
	Issues             *Issues
	Pipelines          *Pipelines
	Repository         *Repository
	Commits            *Commits
	Diff               *Diff
	BranchRestrictions *BranchRestrictions
	Webhooks           *Webhooks
	Downloads          *Downloads
	DeployKeys         *DeployKeys
	repositories
}

type RepositoriesRes struct {
	Page    int32
	Pagelen int32
	Size    int32
	Items   []Repository
}

func (r *Repositories) ListForAccount(ro *RepositoriesOptions) (*RepositoriesRes, error) {
	urlPath := "/repositories"
	if ro.Owner != "" {
		urlPath += fmt.Sprintf("/%s", ro.Owner)
	}
	urlStr := r.c.requestUrl(urlPath)
	if ro.Role != "" {
		urlStr += "?role=" + ro.Role
	}
	if ro.Keyword != nil && *ro.Keyword != "" {
		if ro.Role == "" {
			urlStr += "?"
		}
		// https://developer.atlassian.com/cloud/bitbucket/rest/intro/#operators
		urlStr += fmt.Sprintf("q=full_name ~ \"%s\"", *ro.Keyword)
	}
	repos, err := r.c.executePaginated("GET", urlStr, "", ro.Page)
	if err != nil {
		return nil, err
	}
	return decodeRepositories(repos)
}

// Deprecated: Use ListForAccount instead
func (r *Repositories) ListForTeam(ro *RepositoriesOptions) (*RepositoriesRes, error) {
	return r.ListForAccount(ro)
}

func (r *Repositories) ListPublic() (*RepositoriesRes, error) {
	urlStr := r.c.requestUrl("/repositories/")
	repos, err := r.c.executePaginated("GET", urlStr, "", nil)
	if err != nil {
		return nil, err
	}
	return decodeRepositories(repos)
}

func decodeRepositories(reposResponse interface{}) (*RepositoriesRes, error) {
	reposResponseMap, ok := reposResponse.(map[string]interface{})
	if !ok {
		return nil, errors.New("Not a valid format")
	}

	repoArray := reposResponseMap["values"].([]interface{})
	var repos []Repository
	for _, repoEntry := range repoArray {
		repo, err := decodeRepository(repoEntry)
		if err == nil {
			repos = append(repos, *repo)
		}
	}

	page, ok := reposResponseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := reposResponseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	size, ok := reposResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	repositories := RepositoriesRes{
		Page:    int32(page),
		Pagelen: int32(pagelen),
		Size:    int32(size),
		Items:   repos,
	}
	return &repositories, nil
}
