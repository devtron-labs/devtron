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
	"go.uber.org/zap"
)

type WorkflowStatusLatestRepository interface {
	// CI Workflow Status Latest methods
	SaveCiWorkflowStatusLatest(model *CiWorkflowStatusLatest) error
	UpdateCiWorkflowStatusLatest(model *CiWorkflowStatusLatest) error
	GetCiWorkflowStatusLatestByPipelineId(pipelineId int) (*CiWorkflowStatusLatest, error)
	GetCiWorkflowStatusLatestByAppId(appId int) ([]*CiWorkflowStatusLatest, error)
	GetCachedPipelineIds(pipelineIds []int) ([]int, error)
	GetCiWorkflowStatusLatestByPipelineIds(pipelineIds []int) ([]*CiWorkflowStatusLatest, error)
	DeleteCiWorkflowStatusLatestByPipelineId(pipelineId int) error

	// CD Workflow Status Latest methods
	SaveCdWorkflowStatusLatest(model *CdWorkflowStatusLatest) error
	UpdateCdWorkflowStatusLatest(model *CdWorkflowStatusLatest) error
	GetCdWorkflowStatusLatestByPipelineIdAndWorkflowType(pipelineId int, workflowType string) (*CdWorkflowStatusLatest, error)
	GetCdWorkflowStatusLatestByAppId(appId int) ([]*CdWorkflowStatusLatest, error)
	GetCdWorkflowStatusLatestByPipelineId(pipelineId int) ([]*CdWorkflowStatusLatest, error)
	DeleteCdWorkflowStatusLatestByPipelineId(pipelineId int) error
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
func (impl *WorkflowStatusLatestRepositoryImpl) SaveCiWorkflowStatusLatest(model *CiWorkflowStatusLatest) error {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.logger.Errorw("error in saving ci workflow status latest", "err", err, "model", model)
		return err
	}
	return nil
}

func (impl *WorkflowStatusLatestRepositoryImpl) UpdateCiWorkflowStatusLatest(model *CiWorkflowStatusLatest) error {
	_, err := impl.dbConnection.Model(model).WherePK().UpdateNotNull()
	if err != nil {
		impl.logger.Errorw("error in updating ci workflow status latest", "err", err, "model", model)
		return err
	}
	return nil
}

func (impl *WorkflowStatusLatestRepositoryImpl) GetCiWorkflowStatusLatestByPipelineId(pipelineId int) (*CiWorkflowStatusLatest, error) {
	model := &CiWorkflowStatusLatest{}
	err := impl.dbConnection.Model(model).
		Where("pipeline_id = ?", pipelineId).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting ci workflow status latest by pipeline id", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	return model, nil
}

func (impl *WorkflowStatusLatestRepositoryImpl) GetCiWorkflowStatusLatestByAppId(appId int) ([]*CiWorkflowStatusLatest, error) {
	var models []*CiWorkflowStatusLatest
	err := impl.dbConnection.Model(&models).
		Where("app_id = ?", appId).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting ci workflow status latest by app id", "err", err, "appId", appId)
		return nil, err
	}
	return models, nil
}

func (impl *WorkflowStatusLatestRepositoryImpl) DeleteCiWorkflowStatusLatestByPipelineId(pipelineId int) error {
	_, err := impl.dbConnection.Model(&CiWorkflowStatusLatest{}).
		Where("pipeline_id = ?", pipelineId).
		Delete()
	if err != nil {
		impl.logger.Errorw("error in deleting ci workflow status latest by pipeline id", "err", err, "pipelineId", pipelineId)
		return err
	}
	return nil
}

func (impl *WorkflowStatusLatestRepositoryImpl) GetCachedPipelineIds(pipelineIds []int) ([]int, error) {
	if len(pipelineIds) == 0 {
		return []int{}, nil
	}

	var cachedPipelineIds []int
	err := impl.dbConnection.Model(&CiWorkflowStatusLatest{}).
		Column("pipeline_id").
		Where("pipeline_id IN (?)", pg.In(pipelineIds)).
		Select(&cachedPipelineIds)
	if err != nil {
		impl.logger.Errorw("error in getting cached pipeline ids", "err", err, "pipelineIds", pipelineIds)
		return nil, err
	}
	return cachedPipelineIds, nil
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
func (impl *WorkflowStatusLatestRepositoryImpl) SaveCdWorkflowStatusLatest(model *CdWorkflowStatusLatest) error {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.logger.Errorw("error in saving cd workflow status latest", "err", err, "model", model)
		return err
	}
	return nil
}

func (impl *WorkflowStatusLatestRepositoryImpl) UpdateCdWorkflowStatusLatest(model *CdWorkflowStatusLatest) error {
	_, err := impl.dbConnection.Model(model).WherePK().UpdateNotNull()
	if err != nil {
		impl.logger.Errorw("error in updating cd workflow status latest", "err", err, "model", model)
		return err
	}
	return nil
}

func (impl *WorkflowStatusLatestRepositoryImpl) GetCdWorkflowStatusLatestByPipelineIdAndWorkflowType(pipelineId int, workflowType string) (*CdWorkflowStatusLatest, error) {
	model := &CdWorkflowStatusLatest{}
	err := impl.dbConnection.Model(model).
		Where("pipeline_id = ?", pipelineId).
		Where("workflow_type = ?", workflowType).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting cd workflow status latest by pipeline id and workflow type", "err", err, "pipelineId", pipelineId, "workflowType", workflowType)
		return nil, err
	}
	return model, nil
}

func (impl *WorkflowStatusLatestRepositoryImpl) GetCdWorkflowStatusLatestByAppId(appId int) ([]*CdWorkflowStatusLatest, error) {
	var models []*CdWorkflowStatusLatest
	err := impl.dbConnection.Model(&models).
		Where("app_id = ?", appId).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting cd workflow status latest by app id", "err", err, "appId", appId)
		return nil, err
	}
	return models, nil
}

func (impl *WorkflowStatusLatestRepositoryImpl) GetCdWorkflowStatusLatestByPipelineId(pipelineId int) ([]*CdWorkflowStatusLatest, error) {
	var models []*CdWorkflowStatusLatest
	err := impl.dbConnection.Model(&models).
		Where("pipeline_id = ?", pipelineId).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting cd workflow status latest by pipeline id", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	return models, nil
}

func (impl *WorkflowStatusLatestRepositoryImpl) DeleteCdWorkflowStatusLatestByPipelineId(pipelineId int) error {
	_, err := impl.dbConnection.Model(&CdWorkflowStatusLatest{}).
		Where("pipeline_id = ?", pipelineId).
		Delete()
	if err != nil {
		impl.logger.Errorw("error in deleting cd workflow status latest by pipeline id", "err", err, "pipelineId", pipelineId)
		return err
	}
	return nil
}
