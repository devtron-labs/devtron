package repository

import (
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type WorkflowStageRepository interface {
	SaveWorkflowStages(workflowStage []*WorkflowExecutionStage, tx *pg.Tx) ([]*WorkflowExecutionStage, error)
	UpdateWorkflowStages(workflowStage []*WorkflowExecutionStage, tx *pg.Tx) ([]*WorkflowExecutionStage, error)
	GetWorkflowStagesByWorkflowIdAndType(workflowId int, workflowType string) ([]*WorkflowExecutionStage, error)
	GetWorkflowStagesByWorkflowIdAndWtype(wfId int, wfType string) ([]*WorkflowExecutionStage, error)
	GetWorkflowStagesByWorkflowIdsAndWtype(wfIds []int, wfType string) ([]*WorkflowExecutionStage, error)
}

type WorkflowStageRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

type WorkflowExecutionStage struct {
	tableName    struct{}                    `sql:"workflow_execution_stage" pg:",discard_unknown_columns"`
	Id           int                         `sql:"id,pk"`
	StageName    bean.WorkflowStageName      `sql:"stage_name,notnull"` // same as app name
	Status       bean.WorkflowStageStatus    `sql:"status"`
	StatusFor    bean.WorkflowStageStatusFor `sql:"status_type"`
	Message      string                      `sql:"message"`
	Metadata     string                      `sql:"metadata"`
	WorkflowId   int                         `sql:"workflow_id,notnull"`
	WorkflowType string                      `sql:"workflow_type,notnull"`
	StartTime    string                      `sql:"start_time"`
	EndTime      string                      `sql:"end_time"`

	sql.AuditLog
}

func NewWorkflowStageRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *WorkflowStageRepositoryImpl {
	return &WorkflowStageRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

func (impl *WorkflowStageRepositoryImpl) SaveWorkflowStages(workflowStages []*WorkflowExecutionStage, tx *pg.Tx) ([]*WorkflowExecutionStage, error) {
	err := tx.Insert(&workflowStages)
	return workflowStages, err
}

func (impl *WorkflowStageRepositoryImpl) UpdateWorkflowStages(workflowStages []*WorkflowExecutionStage, tx *pg.Tx) ([]*WorkflowExecutionStage, error) {
	if len(workflowStages) == 0 {
		return workflowStages, nil
	}
	//todo optimise below for bulk update
	for _, stage := range workflowStages {
		_, err := tx.Model(stage).WherePK().Update()
		if err != nil {
			return workflowStages, err
		}
	}
	//_, err := .WherePK().UpdateNotNull()
	return workflowStages, nil
}

func (impl *WorkflowStageRepositoryImpl) GetWorkflowStagesByWorkflowIdAndType(workflowId int, workflowType string) ([]*WorkflowExecutionStage, error) {
	var workflowStages []*WorkflowExecutionStage
	err := impl.dbConnection.Model(&workflowStages).Where("workflow_id = ?", workflowId).Where("workflow_type = ?", workflowType).Order("id ASC").Select()
	return workflowStages, err
}

func (impl *WorkflowStageRepositoryImpl) GetWorkflowStagesByWorkflowIdAndWtype(wfId int, wfType string) ([]*WorkflowExecutionStage, error) {
	var workflowStages []*WorkflowExecutionStage
	err := impl.dbConnection.Model(&workflowStages).Where("workflow_id = ?", wfId).Where("workflow_type = ?", wfType).Order("id ASC").Select()
	if err != nil {
		impl.logger.Errorw("error in fetching ci workflow stages", "err", err)
		return workflowStages, err
	}
	return workflowStages, err
}

func (impl *WorkflowStageRepositoryImpl) GetWorkflowStagesByWorkflowIdsAndWtype(wfIds []int, wfType string) ([]*WorkflowExecutionStage, error) {
	var workflowStages []*WorkflowExecutionStage
	if len(wfIds) == 0 {
		return []*WorkflowExecutionStage{}, nil
	}
	err := impl.dbConnection.Model(&workflowStages).Where("workflow_id in (?)", pg.In(wfIds)).Where("workflow_type = ?", wfType).Order("id ASC").Select()
	if err != nil {
		impl.logger.Errorw("error in fetching ci workflow stages", "err", err)
		return workflowStages, err
	}
	return workflowStages, err
}
