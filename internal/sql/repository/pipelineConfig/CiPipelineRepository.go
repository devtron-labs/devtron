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
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type CiPipeline struct {
	tableName        struct{} `sql:"ci_pipeline" pg:",discard_unknown_columns"`
	Id               int      `sql:"id,pk"`
	AppId            int      `sql:"app_id"`
	App              *app.App
	CiTemplateId     int    `sql:"ci_template_id"`
	DockerArgs       string `sql:"docker_args"`
	Name             string `sql:"name"`
	Version          string `sql:"version"`
	Active           bool   `sql:"active,notnull"`
	Deleted          bool   `sql:"deleted,notnull"`
	IsManual         bool   `sql:"manual,notnull"`
	IsExternal       bool   `sql:"external,notnull"`
	ParentCiPipeline int    `sql:"parent_ci_pipeline"`
	ScanEnabled      bool   `sql:"scan_enabled,notnull"`
	sql.AuditLog
	CiPipelineMaterials []*CiPipelineMaterial
	CiTemplate          *CiTemplate
	ExternalCiPipeline  *ExternalCiPipeline
}

type ExternalCiPipeline struct {
	tableName    struct{} `sql:"external_ci_pipeline" pg:",discard_unknown_columns"`
	Id           int      `sql:"id,pk"`
	CiPipelineId int      `sql:"ci_pipeline_id"`
	Active       bool     `sql:"active,notnull"`
	AccessToken  string   `sql:"access_token,notnull"`
	sql.AuditLog
	CiPipeline *CiPipeline
}

type CiPipelineScript struct {
	tableName      struct{} `sql:"ci_pipeline_scripts" pg:",discard_unknown_columns"`
	Id             int      `sql:"id,pk"`
	Name           string   `sql:"name"`
	Index          int      `sql:"index"`
	CiPipelineId   int      `sql:"ci_pipeline_id"`
	Script         string   `sql:"script"`
	Stage          string   `sql:"stage"`
	OutputLocation string   `sql:"output_location"`
	Active         bool     `sql:"active,notnull"`
	sql.AuditLog
}

type CiPipelineRepository interface {
	Save(pipeline *CiPipeline, tx *pg.Tx) error
	SaveExternalCi(pipeline *ExternalCiPipeline, tx *pg.Tx) (*ExternalCiPipeline, error)
	UpdateExternalCi(pipeline *ExternalCiPipeline, tx *pg.Tx) (*ExternalCiPipeline, int, error)
	FindExternalCiByCiPipelineId(ciPipelineId int) (*ExternalCiPipeline, error)
	FindCiScriptsByCiPipelineId(ciPipelineId int) ([]*CiPipelineScript, error)
	SaveCiPipelineScript(ciPipelineScript *CiPipelineScript, tx *pg.Tx) error
	UpdateCiPipelineScript(script *CiPipelineScript, tx *pg.Tx) error
	MarkCiPipelineScriptsInactiveByCiPipelineId(ciPipelineId int, tx *pg.Tx) error
	FindByAppId(appId int) (pipelines []*CiPipeline, err error)
	//find non deleted pipeline
	FindById(id int) (pipeline *CiPipeline, err error)
	FindByCiAndAppDetailsById(pipelineId int) (pipeline *CiPipeline, err error)
	FindByIdsIn(ids []int) ([]*CiPipeline, error)
	Update(pipeline *CiPipeline, tx *pg.Tx) error
	PipelineExistsByName(names []string) (found []string, err error)
	FindByName(pipelineName string) (pipeline *CiPipeline, err error)
	FindByParentCiPipelineId(parentCiPipelineId int) ([]*CiPipeline, error)

	FetchParentCiPipelinesForDG() ([]*CiPipelinesMap, error)
	FetchCiPipelinesForDG(parentId int, childCiPipelineIds []int) (*CiPipeline, int, error)
	FinDByParentCiPipelineAndAppId(parentCiPipeline int, appIds []int) ([]*CiPipeline, error)
	FindAllPipelineInLast24Hour() (pipelines []*CiPipeline, err error)
	Exists() (exist bool, err error)
}
type CiPipelineRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCiPipelineRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CiPipelineRepositoryImpl {
	return &CiPipelineRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl CiPipelineRepositoryImpl) FindByParentCiPipelineId(parentCiPipelineId int) ([]*CiPipeline, error) {
	var ciPipelines []*CiPipeline
	err := impl.dbConnection.Model(&ciPipelines).
		Where("parent_ci_pipeline = ?", parentCiPipelineId).
		Where("active = ?", true).
		Select()
	return ciPipelines, err
}

func (impl CiPipelineRepositoryImpl) FindByIdsIn(ids []int) ([]*CiPipeline, error) {
	var ciPipelines []*CiPipeline
	err := impl.dbConnection.Model(&ciPipelines).
		Where("id in (?)", pg.In(ids)).
		Select()
	return ciPipelines, err
}

func (impl CiPipelineRepositoryImpl) SaveExternalCi(pipeline *ExternalCiPipeline, tx *pg.Tx) (*ExternalCiPipeline, error) {
	err := tx.Insert(pipeline)
	return pipeline, err
}

func (impl CiPipelineRepositoryImpl) UpdateExternalCi(pipeline *ExternalCiPipeline, tx *pg.Tx) (*ExternalCiPipeline, int, error) {
	r, err := tx.Model(pipeline).Where("ci_pipeline_id= ?", pipeline.CiPipelineId).UpdateNotNull()
	rowsUpdated := r.RowsAffected()
	impl.logger.Infof("total rows updated %d", rowsUpdated)
	return pipeline, rowsUpdated, err
}

func (impl CiPipelineRepositoryImpl) Save(pipeline *CiPipeline, tx *pg.Tx) error {
	return tx.Insert(pipeline)
}
func (impl CiPipelineRepositoryImpl) Update(pipeline *CiPipeline, tx *pg.Tx) error {
	r, err := tx.Model(pipeline).WherePK().UpdateNotNull()
	impl.logger.Debugf("total rows saved %d", r.RowsAffected())
	return err
}

func (impl CiPipelineRepositoryImpl) UpdateCiPipelineScript(script *CiPipelineScript, tx *pg.Tx) error {
	r, err := tx.Model(script).WherePK().UpdateNotNull()
	impl.logger.Debugf("total rows saved %d", r.RowsAffected())
	return err
}

func (impl CiPipelineRepositoryImpl) MarkCiPipelineScriptsInactiveByCiPipelineId(ciPipelineId int, tx *pg.Tx) error {
	var script CiPipelineScript
	_, err := tx.Model(&script).Set("active = ?", false).
		Where("ci_pipeline_id = ?", ciPipelineId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking ciPipelineScript inactive by ciPipelineId", "err", err, "ciPipelineId", ciPipelineId)
		return err

	}
	return nil
}

func (impl CiPipelineRepositoryImpl) FindByAppId(appId int) (pipelines []*CiPipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Column("ci_pipeline.*", "CiPipelineMaterials", "ExternalCiPipeline", "CiPipelineMaterials.GitMaterial").
		Where("app_id =?", appId).
		Where("deleted =? ", false).
		Select()
	return pipelines, err
}

func (impl CiPipelineRepositoryImpl) FindExternalCiByCiPipelineId(ciPipelineId int) (*ExternalCiPipeline, error) {
	externalCiPipeline := &ExternalCiPipeline{}
	err := impl.dbConnection.Model(externalCiPipeline).
		Column("external_ci_pipeline.*", "CiPipeline").
		Where("external_ci_pipeline.ci_pipeline_id = ?", ciPipelineId).
		Where("external_ci_pipeline.active =? ", true).
		Select()
	return externalCiPipeline, err
}

func (impl CiPipelineRepositoryImpl) FindCiScriptsByCiPipelineId(ciPipelineId int) ([]*CiPipelineScript, error) {
	var ciPipelineScripts []*CiPipelineScript
	err := impl.dbConnection.Model(&ciPipelineScripts).
		Where("ci_pipeline_id = ?", ciPipelineId).
		Where("active = ?", true).
		Order("index ASC").
		Select()
	return ciPipelineScripts, err
}

func (impl CiPipelineRepositoryImpl) SaveCiPipelineScript(ciPipelineScript *CiPipelineScript, tx *pg.Tx) error {
	ciPipelineScript.Active = true
	return tx.Insert(ciPipelineScript)
}

func (impl CiPipelineRepositoryImpl) FindById(id int) (pipeline *CiPipeline, err error) {
	pipeline = &CiPipeline{Id: id}
	err = impl.dbConnection.Model(pipeline).
		Column("ci_pipeline.*", "App", "CiPipelineMaterials", "CiTemplate", "CiTemplate.DockerRegistry", "CiPipelineMaterials.GitMaterial", "ExternalCiPipeline").
		Where("ci_pipeline.id= ?", id).
		Where("ci_pipeline.deleted =? ", false).
		Select()

	return pipeline, err
}

func (impl CiPipelineRepositoryImpl) PipelineExistsByName(names []string) (found []string, err error) {
	var name []string
	err = impl.dbConnection.Model((*CiPipeline)(nil)).
		Where("name in (?)", pg.In(names)).
		Where("deleted =? ", false).
		Column("name").
		Select(&name)
	return name, err

}

func (impl CiPipelineRepositoryImpl) FindByCiAndAppDetailsById(pipelineId int) (pipeline *CiPipeline, err error) {
	pipeline = &CiPipeline{}
	err = impl.dbConnection.Model(pipeline).
		Column("ci_pipeline.*", "App").
		Join("inner join app a on ci_pipeline.app_id = a.id").
		Where("ci_pipeline.id = ?", pipelineId).
		Where("ci_pipeline.deleted =? ", false).
		Select()
	return pipeline, err
}

func (impl CiPipelineRepositoryImpl) FindByName(pipelineName string) (pipeline *CiPipeline, err error) {
	pipeline = &CiPipeline{}
	err = impl.dbConnection.Model(pipeline).
		Column("ci_pipeline.*", "App").
		Join("inner join app a on ci_pipeline.app_id = a.id").
		Where("name = ?", pipelineName).
		Where("deleted =? ", false).
		Limit(1).
		Select()

	return pipeline, err
}

func (impl CiPipelineRepositoryImpl) FetchParentCiPipelinesForDG() ([]*CiPipelinesMap, error) {
	var ciPipelinesMap []*CiPipelinesMap
	query := "SELECT cip.id, cip.parent_ci_pipeline" +
		" FROM ci_pipeline cip" +
		" WHERE cip.external = TRUE and cip.parent_ci_pipeline > 0 and cip.parent_ci_pipeline IS NOT NULL and cip.deleted = FALSE"
	impl.logger.Debugw("query:", query)
	_, err := impl.dbConnection.Query(&ciPipelinesMap, query)
	if err != nil {
		impl.logger.Error("error in fetching other environment", "error", err)
	}
	return ciPipelinesMap, err
}

type CiPipelinesMap struct {
	Id               int `json:"id"`
	ParentCiPipeline int `json:"parentCiPipeline"`
}
type ConnectedPipelinesMap struct {
	Id    int `json:"id"`
	Count int `json:"count"`
}

func (impl CiPipelineRepositoryImpl) FetchCiPipelinesForDG(parentId int, childCiPipelineIds []int) (*CiPipeline, int, error) {
	pipeline := &CiPipeline{}
	count := 0
	if len(childCiPipelineIds) > 0 {
		query := "SELECT count(p.*) as count FROM pipeline p" +
			" WHERE p.ci_pipeline_id in (" + sqlIntSeq(childCiPipelineIds) + ") and p.deleted = FALSE"
		impl.logger.Debugw("query:", query)
		_, err := impl.dbConnection.Query(&count, query)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Error("error in fetching other environment", "error", err)
			return nil, 0, err
		}
	}
	err := impl.dbConnection.Model(pipeline).
		Column("ci_pipeline.*", "CiPipelineMaterials", "CiPipelineMaterials.GitMaterial").
		Where("ci_pipeline.id = ?", parentId).
		Where("ci_pipeline.deleted = ? ", false).
		Select()
	return pipeline, count, err
}

func sqlIntSeq(ns []int) string {
	if len(ns) == 0 {
		return ""
	}
	estimate := len(ns) * 4
	b := make([]byte, 0, estimate)
	for _, n := range ns {
		b = strconv.AppendInt(b, int64(n), 10)
		b = append(b, ',')
	}
	b = b[:len(b)-1]
	return string(b)
}

func (impl *CiPipelineRepositoryImpl) FinDByParentCiPipelineAndAppId(parentCiPipeline int, appIds []int) ([]*CiPipeline, error) {
	var ciPipelines []*CiPipeline
	err := impl.dbConnection.
		Model(&ciPipelines).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q = q.WhereOr("parent_ci_pipeline =?", parentCiPipeline).
				WhereOr("id = ?", parentCiPipeline)
			return q, nil
		}).
		Where("app_id in (?)", pg.In(appIds)).
		Select()
	return ciPipelines, err
}

func (impl CiPipelineRepositoryImpl) FindAllPipelineInLast24Hour() (pipelines []*CiPipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Column("ci_pipeline.*").
		Where("created_on > ?", time.Now().AddDate(0, 0, -1)).
		Select()
	return pipelines, err
}

func (impl CiPipelineRepositoryImpl) Exists() (exist bool, err error) {
	var pipelines []*CiPipeline
	exist, err = impl.dbConnection.Model(&pipelines).Exists()
	return exist, err
}
