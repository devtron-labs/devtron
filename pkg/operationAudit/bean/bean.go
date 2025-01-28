package bean

import "github.com/devtron-labs/devtron/api/bean"

type OperationType string

const (
	CreateOperationType OperationType = "CREATE"
	UpdateOperationType OperationType = "UPDATE"
	DeleteOperationType OperationType = "DELETE"
)

type EntityType string

const (
	UserEntity      EntityType = "user"
	RoleGroupEntity EntityType = "role-group" // this is similar to permissions group
)

type SchemaFor string

const (
	UserSchema      SchemaFor = "user/v1"
	RoleGroupSchema SchemaFor = "role-group/v1"
)

var MapOfAuditSchemaForVsSchema = map[SchemaFor]interface{}{
	UserSchema:      bean.UserPermissionsAuditDto{},
	RoleGroupSchema: bean.GroupPermissionsAuditDto{},
}
