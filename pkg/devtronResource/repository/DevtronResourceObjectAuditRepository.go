/*
 * Copyright (c) 2024. Devtron Inc.
 */

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
	AuditOperationTypePatch   AuditOperationType = "PATCH"
	AuditOperationTypeDeleted AuditOperationType = "DELETE"
	AuditOperationTypeClone   AuditOperationType = "CLONE"
)

type DevtronResourceObjectAuditRepository interface {
	Save(model *DevtronResourceObjectAudit) error
	FindLatestAuditByOpPath(resourceObjectId int, opPath string) (*DevtronResourceObjectAudit, error)
	FindLatestAuditByOpType(resourceObjectId int, opType AuditOperationType) (*DevtronResourceObjectAudit, error)
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
	AuditOperation          AuditOperationType `sql:"audit_operation"`
	AuditOperationPath      []string           `sql:"audit_operation_path" pg:",array"` //path in object at which the audit operation is performed
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

func (repo *DevtronResourceObjectAuditRepositoryImpl) FindLatestAuditByOpPath(resourceObjectId int, opPath string) (*DevtronResourceObjectAudit, error) {
	var model DevtronResourceObjectAudit
	err := repo.dbConnection.Model(&model).Where("devtron_resource_object_id = ?", resourceObjectId).
		Where("? = ANY(audit_operation_path)", opPath).Order("updated_on desc").Limit(1).Select()
	return &model, err
}

func (repo *DevtronResourceObjectAuditRepositoryImpl) FindLatestAuditByOpType(resourceObjectId int, opType AuditOperationType) (*DevtronResourceObjectAudit, error) {
	var model DevtronResourceObjectAudit
	err := repo.dbConnection.Model(&model).Where("devtron_resource_object_id = ?", resourceObjectId).
		Where("audit_operation = ?", opType).Order("updated_on desc").Limit(1).Select()
	return &model, err
}
