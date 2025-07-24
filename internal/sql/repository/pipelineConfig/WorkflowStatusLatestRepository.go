/*
 * Copyright (c) 2024. Devtron Inc.
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

package pipelineConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type WorkflowStatusLatestRepository interface {
	// CI Workflow Status Latest methods
	SaveCiWorkflowStatusLatest(tx *pg.Tx, model *CiWorkflowStatusLatest) error
	GetCiWorkflowStatusLatestByPipelineIds(pipelineIds []int) ([]*CiWorkflowStatusLatest, error)

	// CD Workflow Status Latest methods
	SaveCdWorkflowStatusLatest(tx *pg.Tx, model *CdWorkflowStatusLatest) error
	GetCdWorkflowStatusLatestByPipelineIds(pipelineIds []int) ([]*CdWorkflowStatusLatest, error)
}

type WorkflowStatusLatestRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewWorkflowStatusLatestRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *WorkflowStatusLatestRepositoryImpl {
	return &WorkflowStatusLatestRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

// CI Workflow Status Latest model
type CiWorkflowStatusLatest struct {
	tableName    struct{} `sql:"ci_workflow_status_latest" pg:",discard_unknown_columns"`
	Id           int      `sql:"id,pk"`
	PipelineId   int      `sql:"pipeline_id"`
	AppId        int      `sql:"app_id"`
	CiWorkflowId int      `sql:"ci_workflow_id"`
	sql.AuditLog
}

// CD Workflow Status Latest model
type CdWorkflowStatusLatest struct {
	tableName        struct{} `sql:"cd_workflow_status_latest" pg:",discard_unknown_columns"`
	Id               int      `sql:"id,pk"`
	PipelineId       int      `sql:"pipeline_id"`
	AppId            int      `sql:"app_id"`
	EnvironmentId    int      `sql:"environment_id"`
	WorkflowType     string   `sql:"workflow_type"`
	WorkflowRunnerId int      `sql:"workflow_runner_id"`
	sql.AuditLog
}

// CI Workflow Status Latest methods implementation
func (impl *WorkflowStatusLatestRepositoryImpl) SaveCiWorkflowStatusLatest(tx *pg.Tx, model *CiWorkflowStatusLatest) error {
	var connection orm.DB
	if tx != nil {
		connection = tx
	} else {
		connection = impl.dbConnection
	}
	err := connection.Insert(model)
	if err != nil {
		impl.logger.Errorw("error in saving ci workflow status latest", "err", err, "model", model)
		return err
	}
	return nil
}

func (impl *WorkflowStatusLatestRepositoryImpl) GetCiWorkflowStatusLatestByPipelineIds(pipelineIds []int) ([]*CiWorkflowStatusLatest, error) {
	if len(pipelineIds) == 0 {
		return []*CiWorkflowStatusLatest{}, nil
	}

	var models []*CiWorkflowStatusLatest
	err := impl.dbConnection.Model(&models).
		Where("pipeline_id IN (?)", pg.In(pipelineIds)).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting ci workflow status latest by pipeline ids", "err", err, "pipelineIds", pipelineIds)
		return nil, err
	}
	return models, nil
}

// CD Workflow Status Latest methods implementation
func (impl *WorkflowStatusLatestRepositoryImpl) SaveCdWorkflowStatusLatest(tx *pg.Tx, model *CdWorkflowStatusLatest) error {
	var connection orm.DB
	if tx != nil {
		connection = tx
	} else {
		connection = impl.dbConnection
	}
	err := connection.Insert(model)
	if err != nil {
		impl.logger.Errorw("error in saving cd workflow status latest", "err", err, "model", model)
		return err
	}
	return nil
}

func (impl *WorkflowStatusLatestRepositoryImpl) GetCdWorkflowStatusLatestByPipelineIds(pipelineIds []int) ([]*CdWorkflowStatusLatest, error) {
	var models []*CdWorkflowStatusLatest
	err := impl.dbConnection.Model(&models).
		Where("pipeline_id IN (?)", pg.In(pipelineIds)).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting cd workflow status latest by pipeline ids", "err", err, "pipelineIds", pipelineIds)
		return nil, err
	}
	return models, nil
}
