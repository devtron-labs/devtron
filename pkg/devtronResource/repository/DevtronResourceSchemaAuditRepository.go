package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DevtronResourceSchemaAuditRepository interface {
	Save(model *DevtronResourceSchemaAudit) error
}

type DevtronResourceSchemaAuditRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewDevtronResourceSchemaAuditRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *DevtronResourceSchemaAuditRepositoryImpl {
	return &DevtronResourceSchemaAuditRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type DevtronResourceSchemaAudit struct {
	tableName               struct{}           `sql:"devtron_resource_schema_audit" pg:",discard_unknown_columns"`
	Id                      int                `sql:"id,pk"`
	DevtronResourceSchemaId int                `sql:"devtron_resource_schema_id"`
	Schema                  string             `sql:"schema"` //json string
	AuditOperation          AuditOperationType `sql:"audit_operation"     `
	sql.AuditLog
}

func (repo *DevtronResourceSchemaAuditRepositoryImpl) Save(model *DevtronResourceSchemaAudit) error {
	err := repo.dbConnection.Insert(model)
	if err != nil {
		repo.logger.Errorw("error in saving devtronResourceSchemaAudit", "err", err, "model", model)
		return err
	}
	return nil
}
