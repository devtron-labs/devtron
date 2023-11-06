package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type AuditOperationType string

const (
	AuditOperationTypeCreate  AuditOperationType = "CREATE"
	AuditOperationTypeUpdate  AuditOperationType = "UPDATE"
	AuditOperationTypeDeleted AuditOperationType = "DELETE"
)

type DevtronResourceObjectAuditRepository interface {
	Save(model *DevtronResourceObjectAudit) error
}

type DevtronResourceObjectAuditRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewDevtronResourceObjectAuditRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *DevtronResourceObjectAuditRepositoryImpl {
	return &DevtronResourceObjectAuditRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type DevtronResourceObjectAudit struct {
	tableName               struct{}           `sql:"devtron_resource_object_audit" pg:",discard_unknown_columns"`
	Id                      int                `sql:"id,pk"`
	DevtronResourceObjectId int                `sql:"devtron_resource_object_id"`
	ObjectData              string             `sql:"object_data"` //json string
	AuditOperation          AuditOperationType `sql:"audit_operation"     `
	sql.AuditLog
}

func (repo *DevtronResourceObjectAuditRepositoryImpl) Save(model *DevtronResourceObjectAudit) error {
	err := repo.dbConnection.Insert(model)
	if err != nil {
		repo.logger.Errorw("error in saving devtronResourceObjectAudit", "err", err, "model", model)
		return err
	}
	return nil
}
