package bean

type RoleType string

const (
	MANAGER_TYPE               RoleType = "manager"
	ADMIN_TYPE                 RoleType = "admin"
	TRIGGER_TYPE               RoleType = "trigger"
	VIEW_TYPE                  RoleType = "view"
	ENTITY_ALL_TYPE            RoleType = "entityAll"
	ENTITY_VIEW_TYPE           RoleType = "entityView"
	ENTITY_SPECIFIC_TYPE       RoleType = "entitySpecific"
	ENTITY_SPECIFIC_ADMIN_TYPE RoleType = "entitySpecificAdmin"
	ENTITY_SPECIFIC_VIEW_TYPE  RoleType = "entitySpecificView"
	ROLE_SPECIFIC_TYPE         RoleType = "roleSpecific"
	ENTITY_CLUSTER_ADMIN_TYPE  RoleType = "clusterAdmin"
	ENTITY_CLUSTER_VIEW_TYPE   RoleType = "clusterView"
	ADMIN_HELM_TYPE            RoleType = "admin"
	EDIT_HELM_TYPE             RoleType = "edit"
	VIEW_HELM_TYPE             RoleType = "view"
	ENTITY_CLUSTER_EDIT_TYPE   RoleType = "clusterEdit"
)

const (
	PROJECT_TYPE                       = "team"
	ENV_TYPE                           = "environment"
	APP_TYPE                           = "app"
	WorkflowType                       = "workflow"
	CHART_GROUP_TYPE                   = "chart-group"
	SUPER_ADMIN                        = "super-admin"
	GLOBAL_ENTITY                      = "globalEntity"
	EMPTY_ROLEFILTER_ENTRY_PLACEHOLDER = "NONE"
)

// Entity Constants
const (
	ENTITY_APPS = "apps"
	EntityJobs  = "jobs"
	CLUSTER     = "cluster"
)

// AccessType Constants
const (
	DEVTRON_APP = "devtron-app"
)

const (
	VALIDATION_FAILED_ERROR_MSG string = "validation failed: group name with , is not allowed"
	RoleNotFoundStatusPrefix           = "role not fount for any given filter: "
)

type RbacRoleDto struct {
	Id              int    `json:"id"` // id of the default role
	RoleName        string `json:"roleName"`
	RoleDisplayName string `json:"roleDisplayName"`
	RoleDescription string `json:"roleDescription"`
	*RbacPolicyEntityGroupDto
}

type RbacPolicyEntityGroupDto struct {
	Entity     string `json:"entity" validate:"oneof=apps cluster chart-group jobs"`
	AccessType string `json:"accessType,omitempty"`
}
