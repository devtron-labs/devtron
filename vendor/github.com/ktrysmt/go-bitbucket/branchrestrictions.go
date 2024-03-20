package bitbucket

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

type BranchRestrictions struct {
	c *Client

	ID      int
	Pattern string
	Kind    string
	Value   *int
}

func (b *BranchRestrictions) Gets(bo *BranchRestrictionsOptions) (interface{}, error) {
	urlStr := b.c.requestUrl("/repositories/%s/%s/branch-restrictions", bo.Owner, bo.RepoSlug)
	return b.c.executePaginated("GET", urlStr, "", nil)
}

func (b *BranchRestrictions) Create(bo *BranchRestrictionsOptions) (*BranchRestrictions, error) {
	data, err := b.buildBranchRestrictionsBody(bo)
	if err != nil {
		return nil, err
	}
	urlStr := b.c.requestUrl("/repositories/%s/%s/branch-restrictions", bo.Owner, bo.RepoSlug)
	response, err := b.c.execute("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodeBranchRestriction(response)
}

func (b *BranchRestrictions) Get(bo *BranchRestrictionsOptions) (*BranchRestrictions, error) {
	urlStr := b.c.requestUrl("/repositories/%s/%s/branch-restrictions/%s", bo.Owner, bo.RepoSlug, bo.ID)
	response, err := b.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeBranchRestriction(response)
}

func (b *BranchRestrictions) Update(bo *BranchRestrictionsOptions) (interface{}, error) {
	data, err := b.buildBranchRestrictionsBody(bo)
	if err != nil {
		return nil, err
	}
	urlStr := b.c.requestUrl("/repositories/%s/%s/branch-restrictions/%s", bo.Owner, bo.RepoSlug, bo.ID)
	response, err := b.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodeBranchRestriction(response)
}

func (b *BranchRestrictions) Delete(bo *BranchRestrictionsOptions) (interface{}, error) {
	urlStr := b.c.requestUrl("/repositories/%s/%s/branch-restrictions/%s", bo.Owner, bo.RepoSlug, bo.ID)
	return b.c.execute("DELETE", urlStr, "")
}

type branchRestrictionsBody struct {
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
	Links   struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
	Value  interface{}                   `json:"value"`
	ID     int                           `json:"id"`
	Users  []branchRestrictionsBodyUser  `json:"users"`
	Groups []branchRestrictionsBodyGroup `json:"groups"`
}

type branchRestrictionsBodyGroup struct {
	Name  string `json:"name"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Html struct {
			Href string `json:"href"`
		} `json:"html"`
	} `json:"links"`
	FullSlug string `json:"full_slug"`
	Slug     string `json:"slug"`
}

type branchRestrictionsBodyUser struct {
	Username     string `json:"username"`
	Website      string `json:"website"`
	Display_name string `json:"display_name"`
	UUID         string `json:"uuid"`
	Created_on   string `json:"created_on"`
	Links        struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Repositories struct {
			Href string `json:"href"`
		} `json:"repositories"`
		Html struct {
			Href string `json:"href"`
		} `json:"html"`
		Followers struct {
			Href string `json:"href"`
		} `json:"followers"`
		Avatar struct {
			Href string `json:"href"`
		} `json:"avatar"`
		Following struct {
			Href string `json:"href"`
		} `json:"following"`
	} `json:"links"`
}

func (b *BranchRestrictions) buildBranchRestrictionsBody(bo *BranchRestrictionsOptions) (string, error) {
	var users []branchRestrictionsBodyUser
	var groups []branchRestrictionsBodyGroup
	for _, u := range bo.Users {
		user := branchRestrictionsBodyUser{
			Username: u,
		}
		users = append(users, user)
	}
	for _, g := range bo.Groups {
		group := branchRestrictionsBodyGroup{
			Slug: g,
		}
		groups = append(groups, group)
	}

	body := branchRestrictionsBody{
		Kind:    bo.Kind,
		Pattern: bo.Pattern,
		Users:   users,
		Groups:  groups,
		Value:   bo.Value,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func decodeBranchRestriction(branchResponse interface{}) (*BranchRestrictions, error) {
	branchMap := branchResponse.(map[string]interface{})

	if branchMap["type"] == "error" {
		return nil, DecodeError(branchMap)
	}

	var branchRestriction = new(BranchRestrictions)
	err := mapstructure.Decode(branchMap, branchRestriction)
	if err != nil {
		return nil, err
	}
	return branchRestriction, nil
}
