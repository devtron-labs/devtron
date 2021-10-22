package bitbucket

import (
	"encoding/json"
	"os"

	"github.com/k0kubun/pp"
	"github.com/mitchellh/mapstructure"
)

type Project struct {
	c *Client

	Uuid        string
	Key         string
	Name        string
	Description string
	Is_private  bool
}

type ProjectOptions struct {
	Uuid        string `json:"uuid"`
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

func (t *Workspace) GetProject(opt *ProjectOptions) (*Project, error) {
	urlStr := t.c.requestUrl("/workspaces/%s/projects/%s", opt.Owner, opt.Key)
	response, err := t.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeProject(response)
}

func (t *Workspace) CreateProject(opt *ProjectOptions) (*Project, error) {
	data := t.buildProjectBody(opt)
	urlStr := t.c.requestUrl("/workspaces/%s/projects", opt.Owner)
	response, err := t.c.execute("POST", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodeProject(response)
}

func (t *Workspace) DeleteProject(opt *ProjectOptions) (interface{}, error) {
	urlStr := t.c.requestUrl("/workspaces/%s/projects/%s", opt.Owner, opt.Key)
	return t.c.execute("DELETE", urlStr, "")
}

func (t *Workspace) UpdateProject(opt *ProjectOptions) (*Project, error) {
	data := t.buildProjectBody(opt)
	urlStr := t.c.requestUrl("/workspaces/%s/projects/%s", opt.Owner, opt.Key)
	response, err := t.c.execute("PUT", urlStr, data)
	if err != nil {
		return nil, err
	}

	return decodeProject(response)
}

func (t *Workspace) buildJsonBody(body map[string]interface{}) string {
	data, err := json.Marshal(body)
	if err != nil {
		pp.Println(err)
		os.Exit(9)
	}

	return string(data)
}

func (t *Workspace) buildProjectBody(opts *ProjectOptions) string {
	body := map[string]interface{}{}

	if opts.Description != "" {
		body["description"] = opts.Description
	}

	if opts.Name != "" {
		body["name"] = opts.Name
	}

	if opts.Key != "" {
		body["key"] = opts.Key
	}

	body["is_private"] = opts.IsPrivate

	return t.buildJsonBody(body)
}

func decodeProject(project interface{}) (*Project, error) {
	var projectEntry Project
	projectResponseMap := project.(map[string]interface{})

	if projectResponseMap["type"] != nil && projectResponseMap["type"] == "error" {
		return nil, DecodeError(projectResponseMap)
	}

	err := mapstructure.Decode(project, &projectEntry)
	return &projectEntry, err
}
