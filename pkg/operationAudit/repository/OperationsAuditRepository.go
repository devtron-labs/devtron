package repository

import (
	"github.com/devtron-labs/devtron/pkg/operationAudit/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type OperationAuditRepository interface {
	SaveAudit(audit *OperationAudit) error
}
type OperationAuditRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewOperationAuditRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *OperationAuditRepositoryImpl {
	return &OperationAuditRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

type OperationAudit struct {
	TableName       struct{}           `sql:"operation_audit" pg:",discard_unknown_columns"`
	Id              int                `sql:"id,pk"`
	EntityId        int32              `sql:"entity_id,notnull"`         // User Id or Role Group Id or any entity id
	EntityType      bean.EntityType    `sql:"entity_type,notnull"`       // user or role-group or etc
	OperationType   bean.OperationType `sql:"operation_type,notnull"`    // create,update,delete
	EntityValueJson string             `sql:"entity_value_json,notnull"` // create - permissions to be created with user, update - we will keep final updated permissions and delete will have operation as delete with existing permissions captured
	SchemaFor       bean.SchemaFor     `sql:"schema_for,notnull"`        // refer SchemaFor
	sql.AuditLog
}

func (repo *OperationAuditRepositoryImpl) SaveAudit(audit *OperationAudit) error {
	err := repo.dbConnection.Insert(audit)
	if err != nil {
		repo.logger.Errorw("error in saving audit", "audit", audit, "err", err)
	}
	return err
}
