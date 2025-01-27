package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type PermissionsAuditRepository interface {
	SaveAudit(audit *PermissionsAudit) error
}
type PermissionsAuditRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPermissionsAuditRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *PermissionsAuditRepositoryImpl {
	return &PermissionsAuditRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

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

type PermissionsAudit struct {
	TableName       struct{}      `sql:"permissions_audit" pg:",discard_unknown_columns"`
	Id              int           `sql:"id,pk"`
	EntityId        int32         `sql:"entity_id,notnull"`        // User Id or Role Group Id
	EntityType      EntityType    `sql:"entity_type,notnull"`      // user or role-group
	OperationType   OperationType `sql:"operation_type,notnull"`   // create,update,delete
	PermissionsJson string        `sql:"permissions_json,notnull"` // create - permissions to be created with user, update - we will keep final updated permissions and delete will have operation as delete with existing permissions captured
	sql.AuditLog
}

func (repo *PermissionsAuditRepositoryImpl) SaveAudit(audit *PermissionsAudit) error {
	err := repo.dbConnection.Insert(audit)
	if err != nil {
		repo.logger.Errorw("error in saving audit", "audit", audit, "err", err)
	}
	return err
}
