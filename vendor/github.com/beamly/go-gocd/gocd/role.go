package gocd

import (
	"context"
	"fmt"
)

// RoleService describes Actions which can be performed on roles
type RoleService service

// Role represents a type of agent/actor who can access resources perform operations
type Role struct {
	Name       string              `json:"name"`
	Type       string              `json:"type"`
	Attributes *RoleAttributesGoCD `json:"attributes"`
	Version    string              `json:"version"`
	Links      *HALLinks           `json:"_links,omitempty"`
}

// RoleAttributesGoCD are attributes describing a role, in this cae, which users are present in the role.
type RoleAttributesGoCD struct {
	Users        []string                   `json:"users,omitempty"`
	AuthConfigID *string                    `json:"auth_config_id,omitempty"`
	Properties   []*RoleAttributeProperties `json:"properties,omitempty"`
}

// RoleAttributeProperties describes properties attached to a role
type RoleAttributeProperties struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RoleListWrapper describes a container for the result of a role list operation
type RoleListWrapper struct {
	Embedded struct {
		Roles []*Role `json:"roles"`
	} `json:"_embedded"`
}

// Create a role
func (rs *RoleService) Create(ctx context.Context, role *Role) (r *Role, resp *APIResponse, err error) {
	apiVersion, err := rs.client.getAPIVersion(ctx, "admin/security/roles")
	if err != nil {
		return nil, nil, err
	}

	r = &Role{}
	_, resp, err = rs.client.postAction(ctx, &APIClientRequest{
		APIVersion:   apiVersion,
		Path:         "admin/security/roles",
		RequestBody:  role,
		ResponseBody: r,
	})

	return
}

// List all roles
func (rs *RoleService) List(ctx context.Context) (r []*Role, resp *APIResponse, err error) {
	apiVersion, err := rs.client.getAPIVersion(ctx, "admin/security/roles")
	if err != nil {
		return nil, nil, err
	}

	wrapper := RoleListWrapper{}
	_, resp, err = rs.client.getAction(ctx, &APIClientRequest{
		APIVersion:   apiVersion,
		Path:         "admin/security/roles",
		ResponseBody: &wrapper,
	})

	return wrapper.Embedded.Roles, resp, err
}

// Get a single role by name
func (rs *RoleService) Get(ctx context.Context, roleName string) (r *Role, resp *APIResponse, err error) {
	apiVersion, err := rs.client.getAPIVersion(ctx, "admin/security/roles/:role_name")
	if err != nil {
		return nil, nil, err
	}
	r = &Role{}
	_, resp, err = rs.client.getAction(ctx, &APIClientRequest{
		APIVersion:   apiVersion,
		Path:         fmt.Sprintf("admin/security/roles/%s", roleName),
		ResponseBody: r,
	})

	return
}

// Delete a role by name
func (rs *RoleService) Delete(ctx context.Context, roleName string) (result string, resp *APIResponse, err error) {
	apiVersion, err := rs.client.getAPIVersion(ctx, "admin/security/roles/:role_name")
	if err != nil {
		return "", nil, err
	}
	return rs.client.deleteAction(
		ctx,
		fmt.Sprintf("admin/security/roles/%s", roleName),
		apiVersion,
	)

}

// Update a role by name
func (rs *RoleService) Update(ctx context.Context, roleName string, role *Role) (
	r *Role, resp *APIResponse, err error) {

	apiVersion, err := rs.client.getAPIVersion(ctx, "admin/security/roles/:role_name")
	if err != nil {
		return nil, nil, err
	}
	r = &Role{}
	_, resp, err = rs.client.putAction(ctx, &APIClientRequest{
		APIVersion:   apiVersion,
		Path:         fmt.Sprintf("admin/security/roles/%s", roleName),
		ResponseBody: r,
		RequestBody:  role,
	})

	return
}
