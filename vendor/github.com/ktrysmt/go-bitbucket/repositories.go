package bitbucket

import (
	"errors"
  "fmt"

	"github.com/mitchellh/mapstructure"
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
	repositories
}

type RepositoriesRes struct {
	Page     int32
	Pagelen  int32
	MaxDepth int32
	Size     int32
	Items    []Repository
}

func (r *Repositories) ListForAccount(ro *RepositoriesOptions) (*RepositoriesRes, error) {
	url := "/repositories"
	if ro.Owner != "" {
		url += fmt.Sprintf("/%s", ro.Owner)
	}
	urlStr := r.c.requestUrl(url)
	if ro.Role != "" {
		urlStr += "?role=" + ro.Role
	}
	repos, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	return decodeRepositorys(repos)
}

func (r *Repositories) ListForTeam(ro *RepositoriesOptions) (*RepositoriesRes, error) {
	urlStr := r.c.requestUrl("/repositories/%s", ro.Owner)
	if ro.Role != "" {
		urlStr += "?role=" + ro.Role
	}
	repos, err := r.c.executeRaw("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	return decodeRepositorys(repos)
}

func (r *Repositories) ListPublic() (*RepositoriesRes, error) {
	urlStr := r.c.requestUrl("/repositories/")
	repos, err := r.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}
	return decodeRepositorys(repos)
}

func decodeRepositorys(reposResponse interface{}) (*RepositoriesRes, error) {
	reposResponseMap, ok := reposResponse.(map[string]interface{})
	if !ok {
		return nil, errors.New("Not a valid format")
	}

	repoArray := reposResponseMap["values"].([]interface{})
	var repos []Repository
	for _, repoEntry := range repoArray {
		var repo Repository
		err := mapstructure.Decode(repoEntry, &repo)
		if err == nil {
			repos = append(repos, repo)
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
	max_depth, ok := reposResponseMap["max_width"].(float64)
	if !ok {
		max_depth = 0
	}
	size, ok := reposResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	repositories := RepositoriesRes{
		Page:     int32(page),
		Pagelen:  int32(pagelen),
		MaxDepth: int32(max_depth),
		Size:     int32(size),
		Items:    repos,
	}
	return &repositories, nil
}
