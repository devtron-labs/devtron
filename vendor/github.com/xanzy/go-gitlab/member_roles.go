package gitlab

import (
	"fmt"
	"net/http"
)

// MemberRolesService handles communication with the member roles related
// methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/member_roles.html
type MemberRolesService struct {
	client *Client
}

// MemberRole represents a GitLab member role.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/member_roles.html
type MemberRole struct {
	ID                       int              `json:"id"`
	Name                     string           `json:"name"`
	Description              string           `json:"description,omitempty"`
	GroupId                  int              `json:"group_id"`
	BaseAccessLevel          AccessLevelValue `json:"base_access_level"`
	AdminCICDVariables       bool             `json:"admin_cicd_variables,omitempty"`
	AdminMergeRequests       bool             `json:"admin_merge_request,omitempty"`
	AdminTerraformState      bool             `json:"admin_terraform_state,omitempty"`
	AdminVulnerability       bool             `json:"admin_vulnerability,omitempty"`
	ReadCode                 bool             `json:"read_code,omitempty"`
	ReadDependency           bool             `json:"read_dependency,omitempty"`
	ReadVulnerability        bool             `json:"read_vulnerability,omitempty"`
	AdminGroupMembers        bool             `json:"admin_group_member,omitempty"`
	ManageProjectAccessToken bool             `json:"manage_project_access_tokens,omitempty"`
	ArchiveProject           bool             `json:"archive_project,omitempty"`
	RemoveProject            bool             `json:"remove_project,omitempty"`
	ManageGroupAccesToken    bool             `json:"manage_group_access_tokens,omitempty"`
}

// ListMemberRoles gets a list of member roles for a specified group.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/member_roles.html#list-all-member-roles-of-a-group
func (s *MemberRolesService) ListMemberRoles(gid interface{}, options ...RequestOptionFunc) ([]*MemberRole, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/member_roles", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var mrs []*MemberRole
	resp, err := s.client.Do(req, &mrs)
	if err != nil {
		return nil, resp, err
	}

	return mrs, resp, nil
}

// CreateMemberRoleOptions represents the available CreateMemberRole() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/member_roles.html#add-a-member-role-to-a-group
type CreateMemberRoleOptions struct {
	Name               *string           `url:"name,omitempty" json:"name,omitempty"`
	BaseAccessLevel    *AccessLevelValue `url:"base_access_level,omitempty" json:"base_access_level,omitempty"`
	Description        *string           `url:"description,omitempty" json:"description,omitempty"`
	AdminMergeRequest  *bool             `url:"admin_merge_request,omitempty" json:"admin_merge_request,omitempty"`
	AdminVulnerability *bool             `url:"admin_vulnerability,omitempty" json:"admin_vulnerability,omitempty"`
	ReadCode           *bool             `url:"read_code,omitempty" json:"read_code,omitempty"`
	ReadDependency     *bool             `url:"read_dependency,omitempty" json:"read_dependency,omitempty"`
	ReadVulnerability  *bool             `url:"read_vulnerability,omitempty" json:"read_vulnerability,omitempty"`
}

// CreateMemberRole creates a new member role for a specified group.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/member_roles.html#add-a-member-role-to-a-group
func (s *MemberRolesService) CreateMemberRole(gid interface{}, opt *CreateMemberRoleOptions, options ...RequestOptionFunc) (*MemberRole, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/member_roles", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	mr := new(MemberRole)
	resp, err := s.client.Do(req, mr)
	if err != nil {
		return nil, resp, err
	}

	return mr, resp, nil
}

// DeleteMemberRole deletes a member role from a specified group.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/member_roles.html#remove-member-role-of-a-group
func (s *MemberRolesService) DeleteMemberRole(gid interface{}, memberRole int, options ...RequestOptionFunc) (*Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("groups/%s/member_roles/%d", PathEscape(group), memberRole)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
