package bitbucket

import (
	"errors"

	"github.com/mitchellh/mapstructure"
)

type Workspace struct {
	c *Client

	Repositories *Repositories
	Permissions  *Permission

	UUID       string
	Type       string
	Slug       string
	Is_Private bool
	Name       string
	workspace
}

type WorkspaceList struct {
	Page       int
	Pagelen    int
	MaxDepth   int
	Size       int
	Next       string
	Workspaces []Workspace
}

type Permission struct {
	c *Client

	Type string
}

type ProjectsRes struct {
	Page     int32
	Pagelen  int32
	MaxDepth int32
	Size     int32
	Items    []Project
}

type WorkspaceMembers struct {
	Page    int
	Pagelen int
	Size    int
	Members []User
}

func (t *Permission) GetUserPermissions(organization, member string) (*Permission, error) {
	urlStr := t.c.requestUrl("/workspaces/%s/permissions?q=user.nickname=\"%s\"", organization, member)
	response, err := t.c.executePaginated("GET", urlStr, "", nil)
	if err != nil {
		return nil, err
	}

	return decodePermission(response), err
}

func (t *Permission) GetUserPermissionsByUuid(organization, member string) (*Permission, error) {
	urlStr := t.c.requestUrl("/workspaces/%s/permissions?q=user.uuid=\"%s\"", organization, member)
	response, err := t.c.executePaginated("GET", urlStr, "", nil)
	if err != nil {
		return nil, err
	}

	return decodePermission(response), err
}

func (t *Workspace) List() (*WorkspaceList, error) {
	urlStr := t.c.requestUrl("/workspaces")
	response, err := t.c.executePaginated("GET", urlStr, "", nil)
	if err != nil {
		return nil, err
	}

	return decodeWorkspaceList(response)
}

func (t *Workspace) Get(workspace string) (*Workspace, error) {
	urlStr := t.c.requestUrl("/workspaces/%s", workspace)
	response, err := t.c.execute("GET", urlStr, "")
	if err != nil {
		return nil, err
	}

	return decodeWorkspace(response)
}

func (w *Workspace) Members(teamname string) (*WorkspaceMembers, error) {
	urlStr := w.c.requestUrl("/workspaces/%s/members", teamname)
	response, err := w.c.executePaginated("GET", urlStr, "", nil)
	if err != nil {
		return nil, err
	}

	return decodeMembers(response)
}

func (w *Workspace) Projects(teamname string) (*ProjectsRes, error) {
	urlStr := w.c.requestUrl("/workspaces/%s/projects/", teamname)
	response, err := w.c.executePaginated("GET", urlStr, "", nil)
	if err != nil {
		return nil, err
	}

	return decodeProjects(response)
}

func decodePermission(permission interface{}) *Permission {
	permissionResponseMap := permission.(map[string]interface{})
	if permissionResponseMap["size"].(float64) == 0 {
		return nil
	}

	permissionValues := permissionResponseMap["values"].([]interface{})
	if len(permissionValues) == 0 {
		return nil
	}

	permissionValue := permissionValues[0].(map[string]interface{})
	return &Permission{
		Type: permissionValue["permission"].(string),
	}
}

func decodeWorkspace(workspace interface{}) (*Workspace, error) {
	var workspaceEntry Workspace
	workspaceResponseMap := workspace.(map[string]interface{})

	if workspaceResponseMap["type"] != nil && workspaceResponseMap["type"] == "error" {
		return nil, DecodeError(workspaceResponseMap)
	}

	err := mapstructure.Decode(workspace, &workspaceEntry)
	return &workspaceEntry, err
}

func decodeWorkspaceList(workspaceResponse interface{}) (*WorkspaceList, error) {
	workspaceResponseMap := workspaceResponse.(map[string]interface{})
	workspaceMapList := workspaceResponseMap["values"].([]interface{})

	var workspaces []Workspace
	for _, workspaceMap := range workspaceMapList {
		workspaceEntry, err := decodeWorkspace(workspaceMap)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, *workspaceEntry)
	}

	page, ok := workspaceResponseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := workspaceResponseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	max_depth, ok := workspaceResponseMap["max_depth"].(float64)
	if !ok {
		max_depth = 0
	}
	size, ok := workspaceResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	next, ok := workspaceResponseMap["next"].(string)
	if !ok {
		next = ""
	}

	workspacesList := WorkspaceList{
		Page:       int(page),
		Pagelen:    int(pagelen),
		MaxDepth:   int(max_depth),
		Size:       int(size),
		Next:       next,
		Workspaces: workspaces,
	}

	return &workspacesList, nil
}

func decodeProjects(projectResponse interface{}) (*ProjectsRes, error) {
	projectsResponseMap, ok := projectResponse.(map[string]interface{})
	if !ok {
		return nil, errors.New("Not a valid format")
	}

	var projects []Project
	projectArray := projectsResponseMap["values"].([]interface{})
	for _, projectEntry := range projectArray {
		var project Project
		if err := mapstructure.Decode(projectEntry, &project); err == nil {
			projects = append(projects, project)
		}
	}

	page, ok := projectsResponseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := projectsResponseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	max_depth, ok := projectsResponseMap["max_width"].(float64)
	if !ok {
		max_depth = 0
	}
	size, ok := projectsResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	res := ProjectsRes{
		Page:     int32(page),
		Pagelen:  int32(pagelen),
		MaxDepth: int32(max_depth),
		Size:     int32(size),
		Items:    projects,
	}
	return &res, nil
}

func decodeMembers(membersResponse interface{}) (*WorkspaceMembers, error) {
	responseMap, ok := membersResponse.(map[string]interface{})
	if !ok {
		return nil, errors.New("not a valid format")
	}

	var members []User
	userArray := responseMap["values"].([]interface{})
	for _, userEntry := range userArray {
		userEntryMap := userEntry.(map[string]interface{})

		member, err := decodeUser(userEntryMap["user"])
		if err != nil {
			return nil, err
		}
		members = append(members, *member)
	}

	page, ok := responseMap["page"].(int)
	if !ok {
		page = 0
	}
	pagelen, ok := responseMap["pagelen"].(int)
	if !ok {
		pagelen = 0
	}
	size, ok := responseMap["size"].(int)
	if !ok {
		size = 0
	}

	workspaceMembers := WorkspaceMembers{
		Page:    page,
		Pagelen: pagelen,
		Size:    size,
		Members: members,
	}
	return &workspaceMembers, nil
}
