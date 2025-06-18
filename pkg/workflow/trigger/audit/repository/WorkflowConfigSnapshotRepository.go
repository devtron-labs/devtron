/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package repository

import (
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type WorkflowConfigSnapshot struct {
	tableName                    struct{}           `sql:"workflow_config_snapshot" pg:",discard_unknown_columns"`
	Id                           int                `sql:"id,pk"`
	WorkflowId                   int                `sql:"workflow_id,notnull"`
	WorkflowType                 types.WorkflowType `sql:"workflow_type,notnull"`
	PipelineId                   int                `sql:"pipeline_id,notnull"`
	WorkflowRequestJson          string             `sql:"workflow_request_json,notnull"`
	WorkflowRequestSchemaVersion string             `sql:"workflow_request_schema_version"`
	sql.AuditLog
}

type WorkflowConfigSnapshotRepository interface {
	SaveWithTx(tx *pg.Tx, snapshot *WorkflowConfigSnapshot) (*WorkflowConfigSnapshot, error)
	// New methods for retrigger functionality
	FindLatestFailedWorkflowSnapshot(workflowId int, workflowType types.WorkflowType) (*WorkflowConfigSnapshot, error)
	sql.TransactionWrapper
}

type WorkflowConfigSnapshotRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
	*sql.TransactionUtilImpl
}

func NewWorkflowConfigSnapshotRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger, transactionUtilImpl *sql.TransactionUtilImpl) *WorkflowConfigSnapshotRepositoryImpl {
	return &WorkflowConfigSnapshotRepositoryImpl{
		dbConnection:        dbConnection,
		logger:              logger,
		TransactionUtilImpl: transactionUtilImpl,
	}
}

func (impl *WorkflowConfigSnapshotRepositoryImpl) SaveWithTx(tx *pg.Tx, snapshot *WorkflowConfigSnapshot) (*WorkflowConfigSnapshot, error) {
	err := tx.Insert(snapshot)
	if err != nil {
		impl.logger.Errorw("error in saving workflow config snapshot with tx", "err", err, "snapshot", snapshot)
		return snapshot, err
	}
	return snapshot, nil
}

// FindLatestFailedWorkflowSnapshot finds the latest failed workflow snapshot for retrigger
// This method looks for the original workflow that failed, not the retrigger attempts
func (impl *WorkflowConfigSnapshotRepositoryImpl) FindLatestFailedWorkflowSnapshot(workflowId int, workflowType types.WorkflowType) (*WorkflowConfigSnapshot, error) {
	snapshot := &WorkflowConfigSnapshot{}
	err := impl.dbConnection.Model(snapshot).
		Where("workflow_id = ?", workflowId).
		Where("workflow_type = ?", workflowType).
		Select()
	if err != nil {
		impl.logger.Errorw("error in finding latest failed workflow config snapshot", "err", err, "workflowId", workflowId, "workflowType", workflowType)
		return snapshot, err
	}
	return snapshot, nil
}
