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
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type WorkflowConfigSnapshotRepository interface {
	Save(snapshot *WorkflowConfigSnapshot) (*WorkflowConfigSnapshot, error)
	SaveWithTx(tx *pg.Tx, snapshot *WorkflowConfigSnapshot) (*WorkflowConfigSnapshot, error)
	Update(snapshot *WorkflowConfigSnapshot) error
	UpdateWithTx(tx *pg.Tx, snapshot *WorkflowConfigSnapshot) error
	FindById(id int) (*WorkflowConfigSnapshot, error)
	FindByWorkflowIdAndType(workflowId int, workflowType WorkflowType) (*WorkflowConfigSnapshot, error)
	FindByPipelineId(pipelineId int, limit int, offset int) ([]*WorkflowConfigSnapshot, error)
	FindByAppId(appId int, limit int, offset int) ([]*WorkflowConfigSnapshot, error)
	FindLatestByPipelineIdAndType(pipelineId int, workflowType WorkflowType) (*WorkflowConfigSnapshot, error)
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

type WorkflowType string

const (
	CI_WORKFLOW_TYPE      WorkflowType = "CI"
	PRE_CD_WORKFLOW_TYPE  WorkflowType = "PRE_CD"
	POST_CD_WORKFLOW_TYPE WorkflowType = "POST_CD"
)

type TriggerType string

const (
	MANUAL_TRIGGER  TriggerType = "MANUAL"
	AUTO_TRIGGER    TriggerType = "AUTO"
	WEBHOOK_TRIGGER TriggerType = "WEBHOOK"
)

type WorkflowConfigSnapshot struct {
	tableName                   struct{}     `sql:"workflow_config_snapshot" pg:",discard_unknown_columns"`
	Id                          int          `sql:"id,pk"`
	WorkflowId                  int          `sql:"workflow_id,notnull"`
	WorkflowType                WorkflowType `sql:"workflow_type,notnull"`
	PipelineId                  int          `sql:"pipeline_id,notnull"`
	ArtifactId                  int          `sql:"artifact_id"`
	TriggerType                 TriggerType  `sql:"trigger_type,notnull"`
	TriggeredBy                 int32        `sql:"triggered_by,notnull"`
	TriggerMetadata             string       `sql:"trigger_metadata"`
	InfraConfigTriggerHistoryId int          `sql:"infra_config_trigger_history_id"`
	WorkflowRequestJson         string       `sql:"workflow_request_json,notnull"`
	WorkflowRequestVersion      string       `sql:"workflow_request_version"`
	sql.AuditLog
}

func (impl *WorkflowConfigSnapshotRepositoryImpl) Save(snapshot *WorkflowConfigSnapshot) (*WorkflowConfigSnapshot, error) {
	err := impl.dbConnection.Insert(snapshot)
	if err != nil {
		impl.logger.Errorw("error in saving workflow config snapshot", "err", err, "snapshot", snapshot)
		return snapshot, err
	}
	return snapshot, nil
}

func (impl *WorkflowConfigSnapshotRepositoryImpl) SaveWithTx(tx *pg.Tx, snapshot *WorkflowConfigSnapshot) (*WorkflowConfigSnapshot, error) {
	err := tx.Insert(snapshot)
	if err != nil {
		impl.logger.Errorw("error in saving workflow config snapshot with tx", "err", err, "snapshot", snapshot)
		return snapshot, err
	}
	return snapshot, nil
}

func (impl *WorkflowConfigSnapshotRepositoryImpl) Update(snapshot *WorkflowConfigSnapshot) error {
	err := impl.dbConnection.Update(snapshot)
	if err != nil {
		impl.logger.Errorw("error in updating workflow config snapshot", "err", err, "snapshot", snapshot)
		return err
	}
	return nil
}

func (impl *WorkflowConfigSnapshotRepositoryImpl) UpdateWithTx(tx *pg.Tx, snapshot *WorkflowConfigSnapshot) error {
	err := tx.Update(snapshot)
	if err != nil {
		impl.logger.Errorw("error in updating workflow config snapshot with tx", "err", err, "snapshot", snapshot)
		return err
	}
	return nil
}

func (impl *WorkflowConfigSnapshotRepositoryImpl) FindById(id int) (*WorkflowConfigSnapshot, error) {
	snapshot := &WorkflowConfigSnapshot{}
	err := impl.dbConnection.Model(snapshot).
		Where("id = ?", id).
		Select()
	if err != nil {
		impl.logger.Errorw("error in finding workflow config snapshot by id", "err", err, "id", id)
		return snapshot, err
	}
	return snapshot, nil
}

func (impl *WorkflowConfigSnapshotRepositoryImpl) FindByWorkflowIdAndType(workflowId int, workflowType WorkflowType) (*WorkflowConfigSnapshot, error) {
	snapshot := &WorkflowConfigSnapshot{}
	err := impl.dbConnection.Model(snapshot).
		Where("workflow_id = ?", workflowId).
		Where("workflow_type = ?", workflowType).
		Select()
	if err != nil {
		impl.logger.Errorw("error in finding workflow config snapshot by workflow id and type", "err", err, "workflowId", workflowId, "workflowType", workflowType)
		return snapshot, err
	}
	return snapshot, nil
}

func (impl *WorkflowConfigSnapshotRepositoryImpl) FindByPipelineId(pipelineId int, limit int, offset int) ([]*WorkflowConfigSnapshot, error) {
	var snapshots []*WorkflowConfigSnapshot
	err := impl.dbConnection.Model(&snapshots).
		Where("pipeline_id = ?", pipelineId).
		Order("created_on DESC").
		Limit(limit).
		Offset(offset).
		Select()
	if err != nil {
		impl.logger.Errorw("error in finding workflow config snapshots by pipeline id", "err", err, "pipelineId", pipelineId)
		return snapshots, err
	}
	return snapshots, nil
}

func (impl *WorkflowConfigSnapshotRepositoryImpl) FindByAppId(appId int, limit int, offset int) ([]*WorkflowConfigSnapshot, error) {
	var snapshots []*WorkflowConfigSnapshot
	err := impl.dbConnection.Model(&snapshots).
		Where("app_id = ?", appId).
		Order("created_on DESC").
		Limit(limit).
		Offset(offset).
		Select()
	if err != nil {
		impl.logger.Errorw("error in finding workflow config snapshots by app id", "err", err, "appId", appId)
		return snapshots, err
	}
	return snapshots, nil
}

func (impl *WorkflowConfigSnapshotRepositoryImpl) FindLatestByPipelineIdAndType(pipelineId int, workflowType WorkflowType) (*WorkflowConfigSnapshot, error) {
	snapshot := &WorkflowConfigSnapshot{}
	err := impl.dbConnection.Model(snapshot).
		Where("pipeline_id = ?", pipelineId).
		Where("workflow_type = ?", workflowType).
		Order("created_on DESC").
		Limit(1).
		Select()
	if err != nil {
		impl.logger.Errorw("error in finding latest workflow config snapshot by pipeline id and type", "err", err, "pipelineId", pipelineId, "workflowType", workflowType)
		return snapshot, err
	}
	return snapshot, nil
}
