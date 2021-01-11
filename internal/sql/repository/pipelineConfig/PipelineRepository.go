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

package pipelineConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type PipelineType string
type TriggerType string //HOW pipeline should be triggered
type DeploymentTemplate string

const TRIGGER_TYPE_AUTOMATIC TriggerType = "AUTOMATIC"
const TRIGGER_TYPE_MANUAL TriggerType = "MANUAL"

const DEPLOYMENT_TEMPLATE_BLUE_GREEN DeploymentTemplate = "BLUE-GREEN"
const DEPLOYMENT_TEMPLATE_ROLLING DeploymentTemplate = "ROLLING"
const DEPLOYMENT_TEMPLATE_CANARY DeploymentTemplate = "CANARY"
const DEPLOYMENT_TEMPLATE_RECREATE DeploymentTemplate = "RECREATE"

type Pipeline struct {
	tableName                     struct{} `sql:"pipeline" pg:",discard_unknown_columns"`
	Id                            int      `sql:"id,pk"`
	AppId                         int      `sql:"app_id,notnull"`
	App                           App
	CiPipelineId                  int         `sql:"ci_pipeline_id"`
	TriggerType                   TriggerType `sql:"trigger_type,notnull"` // automatic, manual
	EnvironmentId                 int         `sql:"environment_id"`
	Name                          string      `sql:"pipeline_name,notnull"`
	Deleted                       bool        `sql:"deleted,notnull"`
	PreStageConfig                string      `sql:"pre_stage_config_yaml"`
	PostStageConfig               string      `sql:"post_stage_config_yaml"`
	PreTriggerType                TriggerType `sql:"pre_trigger_type"`                   // automatic, manual
	PostTriggerType               TriggerType `sql:"post_trigger_type"`                  // automatic, manual
	PreStageConfigMapSecretNames  string      `sql:"pre_stage_config_map_secret_names"`  // configmap names
	PostStageConfigMapSecretNames string      `sql:"post_stage_config_map_secret_names"` // secret names
	RunPreStageInEnv              bool        `sql:"run_pre_stage_in_env"`               // secret names
	RunPostStageInEnv             bool        `sql:"run_post_stage_in_env"`              // secret names
	Environment                   cluster.Environment
	models.AuditLog
}

type PipelineRepository interface {
	Save(pipeline []*Pipeline, tx *pg.Tx) error
	Update(pipeline *Pipeline, tx *pg.Tx) error
	FindActiveByAppId(appId int) (pipelines []*Pipeline, err error)
	Delete(id int, tx *pg.Tx) error
	FindByName(pipelineName string) (pipeline *Pipeline, err error)
	PipelineExists(pipelineName string) (bool, error)
	FindById(id int) (pipeline *Pipeline, err error)
	FindByIdsIn(ids []int) ([]*Pipeline, error)
	FindByCiPipelineIdsIn(ciPipelineIds []int) ([]*Pipeline, error)
	FindAutomaticByCiPipelineId(ciPipelineId int) (pipelines []*Pipeline, err error)
	GetByEnvOverrideId(envOverrideId int) ([]Pipeline, error)
	FindActiveByAppIdAndEnvironmentId(appId int, environmentId int) (pipelines []*Pipeline, err error)
	UndoDelete(id int) error
	UniqueAppEnvironmentPipelines() ([]*Pipeline, error)
	FindByCiPipelineId(ciPipelineId int) (pipelines []*Pipeline, err error)
	FindByPipelineTriggerGitHash(gitHash string) (pipeline *Pipeline, err error)
	FindByIdsInAndEnvironment(ids []int, environmentId int) ([]*Pipeline, error)
	FindActiveByAppIdAndEnvironmentIdV2() (pipelines []*Pipeline, err error)
	GetConnection() *pg.DB
}

type CiArtifactDTO struct {
	Id           int    `json:"id"`
	PipelineId   int    `json:"pipelineId"` //id of the ci pipeline from which this webhook was triggered
	Image        string `json:"image"`
	ImageDigest  string `json:"imageDigest"`
	MaterialInfo string `json:"materialInfo"` //git material metadata json array string
	DataSource   string `json:"dataSource"`
	WorkflowId   *int   `json:"workflowId"`
}

type PipelineRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPipelineRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *PipelineRepositoryImpl {
	return &PipelineRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl PipelineRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}

func (impl PipelineRepositoryImpl) FindByIdsIn(ids []int) ([]*Pipeline, error) {
	var pipelines []*Pipeline
	err := impl.dbConnection.Model(&pipelines).
		Where("id in (?)", pg.In(ids)).
		Select()
	if err != nil {
		impl.logger.Errorw("error on fetching pipelines", "ids", ids)
	}
	return pipelines, err
}

func (impl PipelineRepositoryImpl) FindByIdsInAndEnvironment(ids []int, environmentId int) ([]*Pipeline, error) {
	var pipelines []*Pipeline
	err := impl.dbConnection.Model(&pipelines).
		Where("id in (?)", pg.In(ids)).
		Where("environment_id = ?", environmentId).
		Select()
	return pipelines, err
}
func (impl PipelineRepositoryImpl) FindByCiPipelineIdsIn(ciPipelineIds []int) ([]*Pipeline, error) {
	var pipelines []*Pipeline
	err := impl.dbConnection.Model(&pipelines).
		Where("ci_pipeline_id in (?)", pg.In(ciPipelineIds)).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) Save(pipeline []*Pipeline, tx *pg.Tx) error {
	var v []interface{}
	for _, i := range pipeline {
		v = append(v, i)
	}
	_, err := tx.Model(v...).Insert()
	return err
}

func (impl PipelineRepositoryImpl) Update(pipeline *Pipeline, tx *pg.Tx) error {
	err := tx.Update(pipeline)
	return err
}

func (impl PipelineRepositoryImpl) FindAutomaticByCiPipelineId(ciPipelineId int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Where("ci_pipeline_id =?", ciPipelineId).
		Where("trigger_type =?", TRIGGER_TYPE_AUTOMATIC).
		Where("deleted =?", false).
		Select()
	if err != nil && util.IsErrNoRows(err) {
		return make([]*Pipeline, 0), nil
	} else if err != nil {
		return nil, err
	}
	return pipelines, nil
}

func (impl PipelineRepositoryImpl) FindByCiPipelineId(ciPipelineId int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Where("ci_pipeline_id =?", ciPipelineId).
		Where("deleted =?", false).
		Select()
	if err != nil && util.IsErrNoRows(err) {
		return make([]*Pipeline, 0), nil
	} else if err != nil {
		return nil, err
	}
	return pipelines, nil
}

func (impl PipelineRepositoryImpl) FindActiveByAppId(appId int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Where("app_id = ?", appId).
		Where("deleted = ?", false).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) FindActiveByAppIdAndEnvironmentId(appId int, environmentId int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Where("app_id = ?", appId).
		Where("deleted = ?", false).
		Where("environment_id = ? ", environmentId).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) FindActiveByAppIdAndEnvironmentIdV2() (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Where("deleted = ?", false).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) Delete(id int, tx *pg.Tx) error {
	pipeline := &Pipeline{}
	r, err := tx.Model(pipeline).Set("deleted =?", true).Where("id =?", id).Update()
	impl.logger.Debugw("update result", "r-affected", r.RowsAffected(), "r-return", r.RowsReturned(), "model", r.Model())
	return err
}

func (impl PipelineRepositoryImpl) UndoDelete(id int) error {
	pipeline := &Pipeline{}
	_, err := impl.dbConnection.Model(pipeline).Set("deleted =?", false).Where("id =?", id).Update()
	return err
}
func (impl PipelineRepositoryImpl) FindByName(pipelineName string) (pipeline *Pipeline, err error) {
	pipeline = &Pipeline{}
	err = impl.dbConnection.Model(pipeline).
		Where("pipeline_name = ?", pipelineName).
		Select()
	return pipeline, err
}

func (impl PipelineRepositoryImpl) PipelineExists(pipelineName string) (bool, error) {
	pipeline := &Pipeline{}
	exists, err := impl.dbConnection.Model(pipeline).
		Where("pipeline_name = ?", pipelineName).
		Where("deleted =? ", false).
		Exists()
	return exists, err
}

func (impl PipelineRepositoryImpl) FindById(id int) (pipeline *Pipeline, err error) {
	pipeline = &Pipeline{}
	err = impl.dbConnection.
		Model(pipeline).
		Column("pipeline.*", "App", "Environment").
		Join("inner join app a on pipeline.app_id = a.id").
		Where("pipeline.id = ?", id).
		Where("deleted = ?", false).
		Select()
	return pipeline, err
}

//Deprecated:
func (impl PipelineRepositoryImpl) FindByEnvOverrideId(envOverrideId int) (pipeline []Pipeline, err error) {
	var pipelines []Pipeline
	err = impl.dbConnection.
		Model(&pipelines).
		Column("pipeline.*").
		Join("INNER JOIN pipeline_config_override pco on pco.pipeline_id = pipeline.id").
		Where("pco.env_config_override_id = ?", envOverrideId).Group("pipeline.id, pipeline.pipeline_name").
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) GetByEnvOverrideId(envOverrideId int) ([]Pipeline, error) {
	var pipelines []Pipeline
	query := "" +
		" SELECT p.*" +
		" FROM chart_env_config_override ceco" +
		" INNER JOIN chart ch on ch.id = ceco.chart_id" +
		" INNER JOIN environment env on env.id = ceco.target_environment" +
		" INNER JOIN app ap on ap.id = ch.app_id" +
		" INNER JOIN pipeline p on p.app_id = ap.id" +
		" WHERE ceco.id=?;"
	_, err := impl.dbConnection.Query(&pipelines, query, envOverrideId)

	if err != nil {
		return nil, err
	}
	return pipelines, err
}

func (impl PipelineRepositoryImpl) UniqueAppEnvironmentPipelines() ([]*Pipeline, error) {
	var pipelines []*Pipeline

	err := impl.dbConnection.
		Model(&pipelines).
		ColumnExpr("DISTINCT app_id, environment_id").
		Where("deleted = ?", false).
		Select()
	if err != nil {
		return nil, err
	}
	return pipelines, err
}

func (impl PipelineRepositoryImpl) FindByPipelineTriggerGitHash(gitHash string) (pipeline *Pipeline, err error) {
	var pipelines *Pipeline
	err = impl.dbConnection.
		Model(&pipelines).
		Column("pipeline.*").
		Join("INNER JOIN pipeline_config_override pco on pco.pipeline_id = pipeline.id").
		Where("pco.git_hash = ?", gitHash).Order(" ORDER BY pco.created_on DESC").Limit(1).
		Select()
	return pipelines, err
}
