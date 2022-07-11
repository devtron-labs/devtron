/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package appWorkflow

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type AppWorkflowRepository interface {
	SaveAppWorkflow(wf *AppWorkflow) (*AppWorkflow, error)
	UpdateAppWorkflow(wf *AppWorkflow) (*AppWorkflow, error)
	FindByIdAndAppId(id int, appId int) (*AppWorkflow, error)
	FindByAppId(appId int) (appWorkflow []*AppWorkflow, err error)
	DeleteAppWorkflow(appWorkflow *AppWorkflow, tx *pg.Tx) error

	SaveAppWorkflowMapping(wf *AppWorkflowMapping, tx *pg.Tx) (*AppWorkflowMapping, error)
	FindByWorkflowId(workflowId int) ([]*AppWorkflowMapping, error)

	FindByComponent(id int, componentType string) ([]*AppWorkflowMapping, error)

	FindByNameAndAppId(name string, appId int) (*AppWorkflow, error)
	FindWFCIMappingByWorkflowId(workflowId int) ([]*AppWorkflowMapping, error)
	FindWFAllMappingByWorkflowId(workflowId int) ([]*AppWorkflowMapping, error)
	FindWFCIMappingByCIPipelineId(ciPipelineId int) ([]*AppWorkflowMapping, error)
	FindWFCDMappingByCIPipelineId(ciPipelineId int) ([]*AppWorkflowMapping, error)
	FindWFCDMappingByCDPipelineId(cdPipelineId int) ([]*AppWorkflowMapping, error)
	DeleteAppWorkflowMapping(appWorkflow *AppWorkflowMapping, tx *pg.Tx) error
	FindWFCDMappingByCIPipelineIds(ciPipelineIds []int) ([]*AppWorkflowMapping, error)
	FindWFCDMappingByParentCDPipelineId(cdPipelineId int) ([]*AppWorkflowMapping, error)
}

type AppWorkflowRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewAppWorkflowRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *AppWorkflowRepositoryImpl {
	return &AppWorkflowRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

const (
	CIPIPELINE string = "CI_PIPELINE"
	CDPIPELINE string = "CD_PIPELINE"
)

type AppWorkflow struct {
	TableName   struct{}        `sql:"app_workflow" pg:",discard_unknown_columns"`
	Id          int             `sql:"id,pk"`
	Name        string          `sql:"name,notnull"`
	Active      bool            `sql:"active"`
	WorkflowDAG json.RawMessage `sql:"workflow_dag"`
	AppId       int             `sql:"app_id"`
	sql.AuditLog
}

// TODO: Suraj - This is v1, it has to be evolved later
type WorkflowDAG struct {
	CiPipelines []int `json:"ciPipelines"`
	CdPipelines []int `json:"cdPipelines"`
}

func (impl AppWorkflowRepositoryImpl) SaveAppWorkflow(wf *AppWorkflow) (*AppWorkflow, error) {
	err := impl.dbConnection.Insert(wf)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return wf, err
	}
	return wf, nil
}

func (impl AppWorkflowRepositoryImpl) UpdateAppWorkflow(wf *AppWorkflow) (*AppWorkflow, error) {
	_, err := impl.dbConnection.Model(wf).WherePK().UpdateNotNull()
	return wf, err
}

func (impl AppWorkflowRepositoryImpl) FindByAppId(appId int) (appWorkflow []*AppWorkflow, err error) {
	err = impl.dbConnection.Model(&appWorkflow).
		Where("app_id = ?", appId).
		Where("active = ?", true).
		Select()
	return appWorkflow, err
}

func (impl AppWorkflowRepositoryImpl) FindByIdAndAppId(id int, appId int) (*AppWorkflow, error) {
	appWorkflow := &AppWorkflow{}
	err := impl.dbConnection.Model(appWorkflow).
		Where("id = ?", id).
		Where("app_id = ?", appId).
		Where("active = ?", true).
		Select()
	return appWorkflow, err
}

func (impl AppWorkflowRepositoryImpl) DeleteAppWorkflow(appWorkflow *AppWorkflow, tx *pg.Tx) error {
	appWorkflowMappings, err := impl.FindWFCIMappingByWorkflowId(appWorkflow.Id)
	if err != nil && pg.ErrNoRows != err {
		impl.Logger.Errorw("err", err)
		return err
	}
	if len(appWorkflowMappings) > 0 {
		for _, item := range appWorkflowMappings {
			err = impl.DeleteAppWorkflowMapping(item, tx)
			if err != nil {
				impl.Logger.Errorw("err", err)
				return err
			}
		}
	}

	appWorkflow.Active = false
	err = impl.dbConnection.Update(appWorkflow)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return err
	}
	return nil
}

//---------------------AppWorkflowMapping-----------------------------------

type AppWorkflowMapping struct {
	TableName     struct{} `sql:"app_workflow_mapping" pg:",discard_unknown_columns"`
	Id            int      `sql:"id,pk"`
	ComponentId   int      `sql:"component_id,notnull"`
	AppWorkflowId int      `sql:"app_workflow_id"`
	Type          string   `sql:"type,notnull"`
	ParentId      int      `sql:"parent_id"`
	Active        bool     `sql:"active"`
	ParentType    string   `sql:"parent_type,notnull"`
	AppWorkflow   *AppWorkflow
	sql.AuditLog
}

func (impl AppWorkflowRepositoryImpl) SaveAppWorkflowMapping(wf *AppWorkflowMapping, tx *pg.Tx) (*AppWorkflowMapping, error) {
	err := tx.Insert(wf)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return wf, err
	}
	return wf, nil
}

func (impl AppWorkflowRepositoryImpl) FindByWorkflowId(workflowId int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping

	err := impl.dbConnection.Model(&appWorkflowsMapping).
		Where("app_workflow_id = ?", workflowId).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) FindByComponent(id int, componentType string) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping
	err := impl.dbConnection.Model(&appWorkflowsMapping).
		Where("component_id = ?", id).
		Where("type = ?", componentType).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) FindByNameAndAppId(name string, appId int) (*AppWorkflow, error) {
	appWorkflow := &AppWorkflow{}
	err := impl.dbConnection.Model(appWorkflow).
		Where("name = ?", name).
		Where("app_id = ?", appId).
		Where("active = ?", true).
		Select()
	return appWorkflow, err
}

func (impl AppWorkflowRepositoryImpl) FindWFCIMappingByWorkflowId(workflowId int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping

	err := impl.dbConnection.Model(&appWorkflowsMapping).
		Where("app_workflow_id = ?", workflowId).
		Where("type = ?", CIPIPELINE).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) FindWFAllMappingByWorkflowId(workflowId int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping

	err := impl.dbConnection.Model(&appWorkflowsMapping).
		Where("app_workflow_id = ?", workflowId).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) FindWFCIMappingByCIPipelineId(ciPipelineId int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping

	err := impl.dbConnection.Model(&appWorkflowsMapping).
		Where("component_id = ?", ciPipelineId).
		Where("type = ?", CIPIPELINE).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) FindWFCDMappingByCIPipelineId(ciPipelineId int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping

	err := impl.dbConnection.Model(&appWorkflowsMapping).
		Where("parent_id = ?", ciPipelineId).
		Where("parent_type = ?", CIPIPELINE).
		Where("type = ?", CDPIPELINE).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) FindWFCDMappingByCIPipelineIds(ciPipelineIds []int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping

	err := impl.dbConnection.Model(&appWorkflowsMapping).
		Where("parent_id in (?) ", pg.In(ciPipelineIds)).
		Where("parent_type = ?", CIPIPELINE).
		Where("type = ?", CDPIPELINE).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) FindWFCDMappingByCDPipelineId(cdPipelineId int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping

	err := impl.dbConnection.Model(&appWorkflowsMapping).
		Where("component_id = ?", cdPipelineId).
		Where("type = ?", CDPIPELINE).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) FindWFCDMappingByParentCDPipelineId(cdPipelineId int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping

	err := impl.dbConnection.Model(&appWorkflowsMapping).
		Where("parent_id = ?", cdPipelineId).
		Where("parent_type = ?", CDPIPELINE).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}
func (impl AppWorkflowRepositoryImpl) DeleteAppWorkflowMapping(appWorkflow *AppWorkflowMapping, tx *pg.Tx) error {
	appWorkflow.Active = false
	err := tx.Update(appWorkflow)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return err
	}
	return nil
}
