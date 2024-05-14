package repository

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DevtronResourceTaskRunRepository interface {
	GetByRunSourceAndTaskTypes(rsIdentifier string, rsdIdentifier []string, taskTypes []bean.TaskType, excludedTaskRunIds []int) ([]DevtronResourceTaskRun, error)
	BulkCreate(tx *pg.Tx, taskRuns []*DevtronResourceTaskRun) error
	sql.TransactionWrapper
}

type DevtronResourceTaskRunRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func NewDevtronResourceTaskRunRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *DevtronResourceTaskRunRepositoryImpl {
	return &DevtronResourceTaskRunRepositoryImpl{
		logger:              logger,
		dbConnection:        dbConnection,
		TransactionUtilImpl: sql.NewTransactionUtilImpl(dbConnection),
	}
}

type DevtronResourceTaskRun struct {
	tableName                     struct{}      `sql:"devtron_resource_task_run" pg:",discard_unknown_columns"`
	Id                            int           `sql:"id,pk"`
	TaskJson                      string        `sql:"task_json"` //json string
	RunSourceIdentifier           string        `sql:"run_source_identifier"`
	RunSourceDependencyIdentifier string        `sql:"run_source_dependency_identifier"`
	RunTargetIdentifier           string        `sql:"run_target_identifier"`
	TaskType                      bean.TaskType `sql:"task_type"`
	TaskTypeIdentifier            int           `sql:"task_type_identifier"` // for pre, post, deploy it refers to cdWorkflowRunnerId
	//DevtronResourceSchemaId       int           `sql:"devtron_resource_schema_id"`// will be introduced in future
	sql.AuditLog
}

func (repo *DevtronResourceTaskRunRepositoryImpl) BulkCreate(tx *pg.Tx, taskRuns []*DevtronResourceTaskRun) error {
	if len(taskRuns) == 0 {
		return nil
	}
	err := tx.Insert(&taskRuns)
	return err
}

func (repo *DevtronResourceTaskRunRepositoryImpl) GetByRunSourceAndTaskTypes(rsIdentifier string, rsdIdentifier []string,
	taskTypes []bean.TaskType, excludedTaskRunIds []int) ([]DevtronResourceTaskRun, error) {
	var dtResourceTaskRun []DevtronResourceTaskRun
	query := repo.dbConnection.
		Model(&dtResourceTaskRun).
		Column("task_type", "task_type_identifier").
		Where("run_source_identifier = ?", rsIdentifier).
		Where("run_source_dependency_identifier IN (?)", pg.In(rsdIdentifier)).
		Where("task_type IN (?)", pg.In(taskTypes))
	if len(excludedTaskRunIds) > 0 {
		query = query.Where("id NOT IN (?)", pg.In(excludedTaskRunIds))
	}
	err := query.Select()
	return dtResourceTaskRun, err
}
