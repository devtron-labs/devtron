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
	SaveAppWorkflowWithTx(wf *AppWorkflow, tx *pg.Tx) (*AppWorkflow, error)
	UpdateAppWorkflow(wf *AppWorkflow) (*AppWorkflow, error)
	FindByIdAndAppId(id int, appId int) (*AppWorkflow, error)
	FindById(id int) (*AppWorkflow, error)
	FindByIds(ids []int) (*AppWorkflow, error)
	FindByAppId(appId int) (appWorkflow []*AppWorkflow, err error)
	FindByAppIds(appIds []int) (appWorkflow []*AppWorkflow, err error)
	DeleteAppWorkflow(appWorkflow *AppWorkflow, tx *pg.Tx) error

	SaveAppWorkflowMapping(wf *AppWorkflowMapping, tx *pg.Tx) (*AppWorkflowMapping, error)
	FindByWorkflowId(workflowId int) ([]*AppWorkflowMapping, error)

	FindByComponent(id int, componentType string) ([]*AppWorkflowMapping, error)

	FindByNameAndAppId(name string, appId int) (*AppWorkflow, error)
	FindWFCIMappingByWorkflowId(workflowId int) ([]*AppWorkflowMapping, error)
	FindWFAllMappingByWorkflowId(workflowId int) ([]*AppWorkflowMapping, error)
	FindWFCIMappingByCIPipelineId(ciPipelineId int) ([]*AppWorkflowMapping, error)
	FindWFCDMappingByCIPipelineId(ciPipelineId int) ([]*AppWorkflowMapping, error)
	FindWFCDMappingByCDPipelineId(cdPipelineId int) (*AppWorkflowMapping, error)
	GetParentDetailsByPipelineId(pipelineId int) (*AppWorkflowMapping, error)
	DeleteAppWorkflowMapping(appWorkflow *AppWorkflowMapping, tx *pg.Tx) error
	DeleteAppWorkflowMappingsByCdPipelineId(pipelineId int, tx *pg.Tx) error
	FindWFCDMappingByCIPipelineIds(ciPipelineIds []int) ([]*AppWorkflowMapping, error)
	FindWFCDMappingByParentCDPipelineId(cdPipelineId int) ([]*AppWorkflowMapping, error)
	FindAllWFMappingsByAppId(appId int) ([]*AppWorkflowMapping, error)
	FindWFCDMappingByExternalCiId(externalCiId int) ([]*AppWorkflowMapping, error)
	FindWFCDMappingByExternalCiIdByIdsIn(externalCiId []int) ([]*AppWorkflowMapping, error)
	FindByTypeAndComponentId(wfId int, componentId int, componentType string) (*AppWorkflowMapping, error)
	FindAllWfsHavingCdPipelinesFromSpecificEnvsOnly(envIds []int, appIds []int) ([]*AppWorkflowMapping, error)
	FindCiPipelineIdsFromAppWfIds(appWfIds []int) ([]int, error)
	FindChildCDIdsByParentCDPipelineId(cdPipelineId int) ([]int, error)
	FindByCDPipelineIds(cdPipelineIds []int) ([]*AppWorkflowMapping, error)
	FindByWorkflowIds(workflowIds []int) ([]*AppWorkflowMapping, error)
	FindMappingByAppIds(appIds []int) ([]*AppWorkflowMapping, error)
	UpdateParentComponentDetails(tx *pg.Tx, oldComponentId int, oldComponentType string, newComponentId int, newComponentType string) error
	FindWFMappingByComponent(componentType string, componentId int) (*AppWorkflowMapping, error)
	FindByComponentId(componentId int) ([]*AppWorkflowMapping, error)
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
	WEBHOOK    string = "WEBHOOK"
)

type AppWorkflow struct {
	TableName   struct{}        `sql:"app_workflow" pg:",discard_unknown_columns"`
	Id          int             `sql:"id,pk"`
	Name        string          `sql:"name,notnull"`
	Active      bool            `sql:"active,notnull"`
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

func (impl AppWorkflowRepositoryImpl) SaveAppWorkflowWithTx(wf *AppWorkflow, tx *pg.Tx) (*AppWorkflow, error) {
	err := tx.Insert(wf)
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

func (impl AppWorkflowRepositoryImpl) FindByAppIds(appIds []int) (appWorkflow []*AppWorkflow, err error) {
	err = impl.dbConnection.Model(&appWorkflow).
		Where("app_id in (?)", pg.In(appIds)).
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

func (impl AppWorkflowRepositoryImpl) FindById(id int) (*AppWorkflow, error) {
	appWorkflow := &AppWorkflow{}
	err := impl.dbConnection.Model(appWorkflow).
		Where("id = ?", id).
		Where("active = ?", true).
		Select()
	return appWorkflow, err
}

func (impl AppWorkflowRepositoryImpl) FindByIds(ids []int) (*AppWorkflow, error) {
	appWorkflow := &AppWorkflow{}
	err := impl.dbConnection.Model(appWorkflow).
		Where("id in (?)", pg.In(ids)).
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
	Active        bool     `sql:"active,notnull"`
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

func (impl AppWorkflowRepositoryImpl) FindWFCDMappingByCDPipelineId(cdPipelineId int) (*AppWorkflowMapping, error) {
	appWorkflowsMapping := &AppWorkflowMapping{}
	err := impl.dbConnection.Model(appWorkflowsMapping).
		Where("component_id = ?", cdPipelineId).
		Where("type = ?", CDPIPELINE).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

// GetParentDetailsByPipelineId returns app workflow which contains only the parent id and parent type for the
// given pipeline component id
func (impl AppWorkflowRepositoryImpl) GetParentDetailsByPipelineId(pipelineId int) (*AppWorkflowMapping, error) {
	appWorkflowsMapping := &AppWorkflowMapping{}
	err := impl.dbConnection.Model(appWorkflowsMapping).
		Column("app_workflow_mapping.parent_id", "app_workflow_mapping.parent_type").
		Where("component_id = ?", pipelineId).
		Where("type = ?", CDPIPELINE).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) FindByTypeAndComponentId(wfId int, componentId int, componentType string) (*AppWorkflowMapping, error) {
	appWorkflowsMapping := &AppWorkflowMapping{}
	err := impl.dbConnection.Model(appWorkflowsMapping).
		Where("app_workflow_id = ?", wfId).
		Where("component_id = ?", componentId).
		Where("type = ?", componentType).
		Where("active = ?", true).
		Limit(1).
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

func (impl AppWorkflowRepositoryImpl) FindAllWFMappingsByAppId(appId int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping
	err := impl.dbConnection.Model(&appWorkflowsMapping).
		Join("INNER JOIN app_workflow aw on aw.id=app_workflow_mapping.app_workflow_id").
		Where("aw.app_id = ?", appId).
		Where("aw.active = ?", true).
		Where("app_workflow_mapping.active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) DeleteAppWorkflowMappingsByCdPipelineId(pipelineId int, tx *pg.Tx) error {
	var model AppWorkflowMapping
	_, err := tx.Model(&model).Set("active = ?", false).
		Where("component_id = ?", pipelineId).
		Where("type = ?", CDPIPELINE).
		Update()
	if err != nil {
		impl.Logger.Errorw("error in deleting appWorkflowMapping by cdPipelineId", "err", err, "cdPipelineId", pipelineId)
		return err
	}
	return nil
}

func (impl AppWorkflowRepositoryImpl) FindAllWfsHavingCdPipelinesFromSpecificEnvsOnly(envIds []int, appIds []int) ([]*AppWorkflowMapping, error) {
	var models []*AppWorkflowMapping
	query := `select * from app_workflow_mapping awm inner join app_workflow aw on aw.id=awm.app_workflow_id 
				where awm.type = ? and aw.app_id in (?) and awm.active = ? and awm.app_workflow_id not in  
					(select app_workflow_id from app_workflow_mapping awm inner join pipeline p on p.id=awm.component_id  
					and awm.type = ? and p.environment_id not in (?) and p.app_id in (?) and p.deleted = ? and awm.active = ?); `
	_, err := impl.dbConnection.Query(&models, query, CDPIPELINE, pg.In(appIds), true, CDPIPELINE, pg.In(envIds), pg.In(appIds), false, true)
	if err != nil {
		impl.Logger.Errorw("error, FindAllWfsHavingCdPipelinesFromSpecificEnvsOnly", "err", err)
		return nil, err
	}
	return models, nil
}

func (impl AppWorkflowRepositoryImpl) FindCiPipelineIdsFromAppWfIds(appWfIds []int) ([]int, error) {
	var ciPipelineIds []int
	query := `select DISTINCT component_id from app_workflow_mapping 
				where type = ? and app_workflow_id in (?) and active = ?; `
	_, err := impl.dbConnection.Query(&ciPipelineIds, query, CIPIPELINE, pg.In(appWfIds), true)
	if err != nil {
		impl.Logger.Errorw("error, FindCiPipelineIdsFromAppWfIds", "err", err)
		return nil, err
	}
	return ciPipelineIds, nil
}

func (impl AppWorkflowRepositoryImpl) FindWFCDMappingByExternalCiId(externalCiId int) ([]*AppWorkflowMapping, error) {
	var models []*AppWorkflowMapping
	err := impl.dbConnection.Model(&models).
		Where("parent_id = ?", externalCiId).
		Where("parent_type = ?", WEBHOOK).
		Where("type = ?", CDPIPELINE).
		Where("active = ?", true).
		Select()
	return models, err
}
func (impl AppWorkflowRepositoryImpl) FindWFMappingByComponent(componentType string, componentId int) (*AppWorkflowMapping, error) {
	model := AppWorkflowMapping{}
	err := impl.dbConnection.Model(&model).
		Where("type = ?", componentType).
		Where("component_id = ?", componentId).
		Where("active = ?", true).
		Select()
	return &model, err
}

func (impl AppWorkflowRepositoryImpl) FindWFCDMappingByExternalCiIdByIdsIn(externalCiId []int) ([]*AppWorkflowMapping, error) {
	var models []*AppWorkflowMapping
	err := impl.dbConnection.Model(&models).
		Where("parent_id in (?)", pg.In(externalCiId)).
		Where("parent_type = ?", WEBHOOK).
		Where("type = ?", CDPIPELINE).
		Where("active = ?", true).
		Select()
	return models, err
}

func (impl AppWorkflowRepositoryImpl) FindChildCDIdsByParentCDPipelineId(cdPipelineId int) ([]int, error) {
	var ids []int
	query := `select component_id from app_workflow_mapping where parent_id=? and parent_type=? and type=? and active=?;`
	_, err := impl.dbConnection.Query(&ids, query, cdPipelineId, CDPIPELINE, CDPIPELINE, true)
	return ids, err
}

func (impl AppWorkflowRepositoryImpl) FindByCDPipelineIds(cdPipelineIds []int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping
	err := impl.dbConnection.Model(&appWorkflowsMapping).
		Where("component_id in (?)", pg.In(cdPipelineIds)).
		Where("type = ?", CDPIPELINE).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) FindByWorkflowIds(workflowIds []int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping
	if len(workflowIds) == 0 {
		return appWorkflowsMapping, nil
	}
	err := impl.dbConnection.Model(&appWorkflowsMapping).
		Where("app_workflow_id in (?)", pg.In(workflowIds)).
		Where("active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) FindMappingByAppIds(appIds []int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping
	err := impl.dbConnection.Model(&appWorkflowsMapping).Column("app_workflow_mapping.*", "AppWorkflow").
		Where("app_workflow.app_id in (?)", pg.In(appIds)).
		Where("app_workflow.active = ?", true).
		Where("app_workflow_mapping.active = ?", true).
		Select()
	return appWorkflowsMapping, err
}

func (impl AppWorkflowRepositoryImpl) UpdateParentComponentDetails(tx *pg.Tx, oldParentId int, oldParentType string, newParentId int, newParentType string) error {

	/*updateQuery := fmt.Sprintf(" UPDATE app_workflow_mapping "+
		" SET parent_type = (select type from new_app_workflow_mapping),parent_id = (select id from new_app_workflow_mapping) where parent_id = %v and parent_type='%v' and active = true", oldComponentId, oldComponentType)

	finalQuery := withQuery + updateQuery*/
	_, err := tx.Model((*AppWorkflowMapping)(nil)).
		Set("parent_type = ?", newParentType).
		Set("parent_id = ?", newParentId).
		Where("parent_type = ?", oldParentType).
		Where("parent_id = ?", oldParentId).
		Where("active = true").
		Update()
	return err
}

func (impl AppWorkflowRepositoryImpl) FindByComponentId(componentId int) ([]*AppWorkflowMapping, error) {
	var appWorkflowsMapping []*AppWorkflowMapping
	err := impl.dbConnection.Model(&appWorkflowsMapping).Column("app_workflow_mapping.*", "AppWorkflow").
		Where("app_workflow_mapping.component_id= ?", componentId).
		Where("app_workflow.active = ?", true).
		Where("app_workflow_mapping.active = ?", true).
		Select()
	return appWorkflowsMapping, err
}
