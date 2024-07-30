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

package pipelineConfig

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/ciPipeline"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	ciPipelineBean "github.com/devtron-labs/devtron/pkg/pipeline/bean/CiPipeline"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util/response/pagination"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type CiPipeline struct {
	tableName                struct{} `sql:"ci_pipeline" pg:",discard_unknown_columns"`
	Id                       int      `sql:"id,pk"`
	AppId                    int      `sql:"app_id"`
	App                      *app.App
	CiTemplateId             int    `sql:"ci_template_id"`
	DockerArgs               string `sql:"docker_args"`
	Name                     string `sql:"name"`
	Version                  string `sql:"version"`
	Active                   bool   `sql:"active,notnull"`
	Deleted                  bool   `sql:"deleted,notnull"`
	IsManual                 bool   `sql:"manual,notnull"`
	IsExternal               bool   `sql:"external,notnull"`
	ParentCiPipeline         int    `sql:"parent_ci_pipeline"`
	ScanEnabled              bool   `sql:"scan_enabled,notnull"`
	IsDockerConfigOverridden bool   `sql:"is_docker_config_overridden, notnull"`
	PipelineType             string `sql:"ci_pipeline_type"`
	sql.AuditLog
	CiPipelineMaterials []*CiPipelineMaterial
	CiTemplate          *CiTemplate
}

type CiEnvMapping struct {
	tableName     struct{} `sql:"ci_env_mapping" pg:",discard_unknown_columns"`
	Id            int      `sql:"id,pk"`
	EnvironmentId int      `sql:"environment_id"`
	CiPipelineId  int      `sql:"ci_pipeline_id"`
	Deleted       bool     `sql:"deleted,notnull"`
	CiPipeline    CiPipeline
	Environment   repository.Environment
	sql.AuditLog
}

type ExternalCiPipeline struct {
	tableName   struct{} `sql:"external_ci_pipeline" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	AppId       int      `sql:"app_id"`
	Active      bool     `sql:"active,notnull"`
	AccessToken string   `sql:"access_token"`
	sql.AuditLog
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
	sql.TransactionWrapper
	Save(pipeline *CiPipeline, tx *pg.Tx) error
	SaveCiEnvMapping(cienvmapping *CiEnvMapping, tx *pg.Tx) error
	SaveExternalCi(pipeline *ExternalCiPipeline, tx *pg.Tx) (*ExternalCiPipeline, error)
	UpdateExternalCi(pipeline *ExternalCiPipeline, tx *pg.Tx) (*ExternalCiPipeline, error)
	FindExternalCiByCiPipelineId(ciPipelineId int) (*ExternalCiPipeline, error)
	FindExternalCiById(id int) (*ExternalCiPipeline, error)
	FindExternalCiByAppId(appId int) ([]*ExternalCiPipeline, error)
	FindExternalCiByAppIds(appIds []int) ([]*ExternalCiPipeline, error)
	FindCiScriptsByCiPipelineId(ciPipelineId int) ([]*CiPipelineScript, error)
	FindCiScriptsByCiPipelineIds(ciPipelineId []int) ([]*CiPipelineScript, error)
	SaveCiPipelineScript(ciPipelineScript *CiPipelineScript, tx *pg.Tx) error
	UpdateCiPipelineScript(script *CiPipelineScript, tx *pg.Tx) error
	MarkCiPipelineScriptsInactiveByCiPipelineId(ciPipelineId int, tx *pg.Tx) error
	FindByAppId(appId int) (pipelines []*CiPipeline, err error)
	FindCiPipelineByAppIdAndEnvIds(appId int, envIds []int) ([]*CiPipeline, error)
	FindByAppIds(appIds []int) (pipelines []*CiPipeline, err error)
	//find any pipeline by id, includes soft deleted as well
	FindByIdIncludingInActive(id int) (pipeline *CiPipeline, err error)
	//find non deleted pipeline
	FindById(id int) (pipeline *CiPipeline, err error)
	// FindOneWithAppData is to be used for fetching minimum data (including app.App) for CiPipeline for the given CiPipeline.Id
	FindOneWithAppData(id int) (pipeline *CiPipeline, err error)
	FindCiEnvMappingByCiPipelineId(ciPipelineId int) (*CiEnvMapping, error)
	FindParentCiPipelineMapByAppId(appId int) ([]*CiPipeline, []int, error)
	FindByCiAndAppDetailsById(pipelineId int) (pipeline *CiPipeline, err error)
	FindByIdsIn(ids []int) ([]*CiPipeline, error)
	Update(pipeline *CiPipeline, tx *pg.Tx) error
	UpdateCiEnvMapping(cienvmapping *CiEnvMapping, tx *pg.Tx) error
	PipelineExistsByName(names []string) (found []string, err error)
	FindByName(pipelineName string) (pipeline *CiPipeline, err error)
	CheckIfPipelineExistsByNameAndAppId(pipelineName string, appId int) (bool, error)
	FindByLinkedCiCount(parentCiPipelineId int) (int, error)
	FindByParentCiPipelineId(parentCiPipelineId int) ([]*CiPipeline, error)
	FindByParentIdAndType(parentCiPipelineId int, pipelineType string) ([]*CiPipeline, error)

	FetchParentCiPipelinesForDG() ([]*ciPipeline.CiPipelinesMap, error)
	FetchCiPipelinesForDG(parentId int, childCiPipelineIds []int) (*CiPipeline, int, error)
	FinDByParentCiPipelineAndAppId(parentCiPipeline int, appIds []int) ([]*CiPipeline, error)
	FindAllPipelineCreatedCountInLast24Hour() (pipelineCount int, err error)
	FindAllDeletedPipelineCountInLast24Hour() (pipelineCount int, err error)
	FindNumberOfAppsWithCiPipeline(appIds []int) (count int, err error)
	FindAppAndProjectByCiPipelineIds(ciPipelineIds []int) ([]*CiPipeline, error)
	FindCiPipelineConfigsByIds(ids []int) ([]*CiPipeline, error)
	FindByParentCiPipelineIds(parentCiPipelineIds []int) ([]*CiPipeline, error)
	FindWithMinDataByCiPipelineId(id int) (pipeline *CiPipeline, err error)
	FindAppIdsForCiPipelineIds(pipelineIds []int) (map[int]int, error)
	GetCiPipelineByArtifactId(artifactId int) (*CiPipeline, error)
	GetExternalCiPipelineByArtifactId(artifactId int) (*ExternalCiPipeline, error)
	FindLinkedCiCount(ciPipelineId int) (int, error)
	GetLinkedCiPipelines(ctx context.Context, ciPipelineId int) ([]*CiPipeline, error)
	GetDownStreamInfo(ctx context.Context, sourceCiPipelineId int,
		appNameMatch, envNameMatch string, req *pagination.RepositoryRequest) ([]ciPipeline.LinkedCIDetails, int, error)
}

type CiPipelineRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
	*sql.TransactionUtilImpl
}

func NewCiPipelineRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger, TransactionUtilImpl *sql.TransactionUtilImpl) *CiPipelineRepositoryImpl {
	return &CiPipelineRepositoryImpl{
		dbConnection:        dbConnection,
		logger:              logger,
		TransactionUtilImpl: TransactionUtilImpl,
	}
}

func (impl *CiPipelineRepositoryImpl) FindByLinkedCiCount(parentCiPipelineId int) (int, error) {
	return impl.dbConnection.Model((*CiPipeline)(nil)).
		Where("parent_ci_pipeline = ?", parentCiPipelineId).
		Where("active = ?", true).
		Count()
}

func (impl *CiPipelineRepositoryImpl) FindByParentCiPipelineId(parentCiPipelineId int) ([]*CiPipeline, error) {
	var ciPipelines []*CiPipeline
	err := impl.dbConnection.Model(&ciPipelines).
		Where("parent_ci_pipeline = ?", parentCiPipelineId).
		Where("active = ?", true).
		Select()
	return ciPipelines, err
}

func (impl *CiPipelineRepositoryImpl) FindByParentIdAndType(parentCiPipelineId int, pipelineType string) ([]*CiPipeline, error) {
	var ciPipelines []*CiPipeline
	err := impl.dbConnection.Model(&ciPipelines).
		Where("parent_ci_pipeline = ?", parentCiPipelineId).
		Where("ci_pipeline_type = ?", pipelineType).
		Where("active = ?", true).
		Select()
	return ciPipelines, err
}

func (impl *CiPipelineRepositoryImpl) FindByIdsIn(ids []int) ([]*CiPipeline, error) {
	var ciPipelines []*CiPipeline
	err := impl.dbConnection.Model(&ciPipelines).
		Where("id in (?)", pg.In(ids)).
		Select()
	return ciPipelines, err
}

func (impl *CiPipelineRepositoryImpl) SaveExternalCi(pipeline *ExternalCiPipeline, tx *pg.Tx) (*ExternalCiPipeline, error) {
	err := tx.Insert(pipeline)
	return pipeline, err
}

func (impl *CiPipelineRepositoryImpl) UpdateExternalCi(pipeline *ExternalCiPipeline, tx *pg.Tx) (*ExternalCiPipeline, error) {
	err := tx.Update(pipeline)
	return pipeline, err
}

func (impl *CiPipelineRepositoryImpl) Save(pipeline *CiPipeline, tx *pg.Tx) error {
	return tx.Insert(pipeline)
}

func (impl *CiPipelineRepositoryImpl) SaveCiEnvMapping(cienvmapping *CiEnvMapping, tx *pg.Tx) error {
	return tx.Insert(cienvmapping)
}

func (impl *CiPipelineRepositoryImpl) UpdateCiEnvMapping(cienvmapping *CiEnvMapping, tx *pg.Tx) error {
	return tx.Update(cienvmapping)
}

func (impl *CiPipelineRepositoryImpl) Update(pipeline *CiPipeline, tx *pg.Tx) error {
	r, err := tx.Model(pipeline).WherePK().UpdateNotNull()
	impl.logger.Debugf("total rows saved %d", r.RowsAffected())
	return err
}

func (impl *CiPipelineRepositoryImpl) UpdateCiPipelineScript(script *CiPipelineScript, tx *pg.Tx) error {
	r, err := tx.Model(script).WherePK().UpdateNotNull()
	impl.logger.Debugf("total rows saved %d", r.RowsAffected())
	return err
}

func (impl *CiPipelineRepositoryImpl) MarkCiPipelineScriptsInactiveByCiPipelineId(ciPipelineId int, tx *pg.Tx) error {
	var script CiPipelineScript
	_, err := tx.Model(&script).Set("active = ?", false).
		Where("ci_pipeline_id = ?", ciPipelineId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking ciPipelineScript inactive by ciPipelineId", "err", err, "ciPipelineId", ciPipelineId)
		return err

	}
	return nil
}

func (impl *CiPipelineRepositoryImpl) FindByAppId(appId int) (pipelines []*CiPipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Column("ci_pipeline.*", "CiPipelineMaterials", "CiPipelineMaterials.GitMaterial").
		Where("ci_pipeline.app_id =?", appId).
		Where("deleted =? ", false).
		Select()
	return pipelines, err
}

func (impl *CiPipelineRepositoryImpl) FindByAppIds(appIds []int) (pipelines []*CiPipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Column("ci_pipeline.*", "App", "CiPipelineMaterials", "CiPipelineMaterials.GitMaterial").
		Where("ci_pipeline.app_id in (?)", pg.In(appIds)).
		Where("deleted =? ", false).
		Select()
	return pipelines, err
}

func (impl *CiPipelineRepositoryImpl) FindExternalCiByCiPipelineId(ciPipelineId int) (*ExternalCiPipeline, error) {
	externalCiPipeline := &ExternalCiPipeline{}
	err := impl.dbConnection.Model(externalCiPipeline).
		Column("external_ci_pipeline.*", "CiPipeline").
		Where("external_ci_pipeline.ci_pipeline_id = ?", ciPipelineId).
		Where("external_ci_pipeline.active =? ", true).
		Select()
	return externalCiPipeline, err
}

func (impl *CiPipelineRepositoryImpl) FindExternalCiById(id int) (*ExternalCiPipeline, error) {
	externalCiPipeline := &ExternalCiPipeline{}
	err := impl.dbConnection.Model(externalCiPipeline).
		Column("external_ci_pipeline.*").
		Where("id = ?", id).
		Where("active =? ", true).
		Select()
	return externalCiPipeline, err
}

func (impl *CiPipelineRepositoryImpl) FindExternalCiByAppId(appId int) ([]*ExternalCiPipeline, error) {
	var externalCiPipeline []*ExternalCiPipeline
	err := impl.dbConnection.Model(&externalCiPipeline).
		Column("external_ci_pipeline.*").
		Where("app_id = ?", appId).
		Where("active =? ", true).
		Select()
	return externalCiPipeline, err
}

func (impl *CiPipelineRepositoryImpl) FindExternalCiByAppIds(appIds []int) ([]*ExternalCiPipeline, error) {
	var externalCiPipeline []*ExternalCiPipeline
	err := impl.dbConnection.Model(&externalCiPipeline).
		Column("external_ci_pipeline.*").
		Where("app_id in (?)", pg.In(appIds)).
		Where("active =? ", true).
		Select()
	return externalCiPipeline, err
}

func (impl *CiPipelineRepositoryImpl) FindCiScriptsByCiPipelineId(ciPipelineId int) ([]*CiPipelineScript, error) {
	var ciPipelineScripts []*CiPipelineScript
	err := impl.dbConnection.Model(&ciPipelineScripts).
		Where("ci_pipeline_id = ?", ciPipelineId).
		Where("active = ?", true).
		Order("index ASC").
		Select()
	return ciPipelineScripts, err
}

func (impl *CiPipelineRepositoryImpl) FindCiScriptsByCiPipelineIds(ciPipelineIds []int) ([]*CiPipelineScript, error) {
	var ciPipelineScripts []*CiPipelineScript
	err := impl.dbConnection.Model(&ciPipelineScripts).
		Where("ci_pipeline_id in (?)", ciPipelineIds).
		Where("active = ?", true).
		Order("index ASC").
		Select()
	return ciPipelineScripts, err
}

func (impl *CiPipelineRepositoryImpl) SaveCiPipelineScript(ciPipelineScript *CiPipelineScript, tx *pg.Tx) error {
	ciPipelineScript.Active = true
	return tx.Insert(ciPipelineScript)
}

func (impl *CiPipelineRepositoryImpl) FindByIdIncludingInActive(id int) (pipeline *CiPipeline, err error) {
	pipeline = &CiPipeline{Id: id}
	err = impl.dbConnection.Model(pipeline).
		Column("ci_pipeline.*", "App", "CiPipelineMaterials", "CiTemplate", "CiTemplate.DockerRegistry", "CiPipelineMaterials.GitMaterial").
		Relation("CiPipelineMaterials", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("(ci_pipeline_material.active=true)"), nil
		}).
		Where("ci_pipeline.id= ?", id).
		Select()

	return pipeline, err
}

func (impl *CiPipelineRepositoryImpl) FindById(id int) (pipeline *CiPipeline, err error) {
	pipeline = &CiPipeline{Id: id}
	err = impl.dbConnection.Model(pipeline).
		Column("ci_pipeline.*", "App", "CiPipelineMaterials", "CiTemplate", "CiTemplate.DockerRegistry", "CiPipelineMaterials.GitMaterial").
		Relation("CiPipelineMaterials", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("(ci_pipeline_material.active=true)"), nil
		}).
		Where("ci_pipeline.id= ?", id).
		Where("ci_pipeline.deleted =? ", false).
		Select()

	return pipeline, err
}

// FindOneWithAppData is to be used for fetching minimum data (including app.App) for CiPipeline for the given CiPipeline.Id
func (impl *CiPipelineRepositoryImpl) FindOneWithAppData(id int) (pipeline *CiPipeline, err error) {
	pipeline = &CiPipeline{}
	err = impl.dbConnection.Model(pipeline).
		Column("ci_pipeline.*", "App").
		Where("ci_pipeline.id= ?", id).
		Where("ci_pipeline.deleted =? ", false).
		Select()

	return pipeline, err
}

func (impl *CiPipelineRepositoryImpl) FindCiEnvMappingByCiPipelineId(ciPipelineId int) (*CiEnvMapping, error) {
	ciEnvMapping := &CiEnvMapping{}
	err := impl.dbConnection.Model(ciEnvMapping).
		Where("ci_pipeline_id= ?", ciPipelineId).
		Where("deleted =? ", false).
		Select()

	return ciEnvMapping, err
}

func (impl *CiPipelineRepositoryImpl) FindWithMinDataByCiPipelineId(id int) (pipeline *CiPipeline, err error) {
	pipeline = &CiPipeline{Id: id}
	err = impl.dbConnection.Model(pipeline).
		Column("ci_pipeline.*", "CiTemplate").
		Where("ci_pipeline.id= ?", id).
		Where("ci_pipeline.deleted =? ", false).
		Select()

	return pipeline, err
}

func (impl *CiPipelineRepositoryImpl) FindParentCiPipelineMapByAppId(appId int) ([]*CiPipeline, []int, error) {
	var parentCiPipelines []*CiPipeline
	var linkedCiPipelineIds []int
	queryLinked := `select * from ci_pipeline where id in (select parent_ci_pipeline from ci_pipeline where app_id=? and deleted=? and parent_ci_pipeline is not null) order by id asc;`
	_, err := impl.dbConnection.Query(&parentCiPipelines, queryLinked, appId, false)
	if err != nil {
		impl.logger.Error("error in fetching linked ci pipelines", "error", err)
		return nil, nil, err
	}
	queryParent := `select id from ci_pipeline where app_id=? and deleted=? and parent_ci_pipeline is not null order by parent_ci_pipeline asc;`
	_, err = impl.dbConnection.Query(&linkedCiPipelineIds, queryParent, appId, false)
	if err != nil {
		impl.logger.Error("error in fetching parent ci pipelines", "error", err)
		return nil, nil, err
	}

	return parentCiPipelines, linkedCiPipelineIds, nil
}

func (impl *CiPipelineRepositoryImpl) PipelineExistsByName(names []string) (found []string, err error) {
	var name []string
	err = impl.dbConnection.Model((*CiPipeline)(nil)).
		Where("name in (?)", pg.In(names)).
		Where("deleted =? ", false).
		Column("name").
		Select(&name)
	return name, err

}

func (impl *CiPipelineRepositoryImpl) FindByCiAndAppDetailsById(pipelineId int) (pipeline *CiPipeline, err error) {
	pipeline = &CiPipeline{}
	err = impl.dbConnection.Model(pipeline).
		Column("ci_pipeline.*", "App").
		Join("inner join app a on ci_pipeline.app_id = a.id").
		Where("ci_pipeline.id = ?", pipelineId).
		Where("ci_pipeline.deleted =? ", false).
		Select()
	return pipeline, err
}

func (impl *CiPipelineRepositoryImpl) FindByName(pipelineName string) (pipeline *CiPipeline, err error) {
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

func (impl *CiPipelineRepositoryImpl) CheckIfPipelineExistsByNameAndAppId(pipelineName string, appId int) (bool, error) {
	pipeline := &CiPipeline{}
	found, err := impl.dbConnection.Model(pipeline).
		Column("ci_pipeline.*").
		Where("name = ?", pipelineName).
		Where("app_id = ?", appId).
		Where("deleted =? ", false).
		Exists()

	return found, err
}

func (impl *CiPipelineRepositoryImpl) FetchParentCiPipelinesForDG() ([]*ciPipeline.CiPipelinesMap, error) {
	var ciPipelinesMap []*ciPipeline.CiPipelinesMap
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

func (impl *CiPipelineRepositoryImpl) FetchCiPipelinesForDG(parentId int, childCiPipelineIds []int) (*CiPipeline, int, error) {
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

// TODO remove this util and use pg.In()
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

func (impl *CiPipelineRepositoryImpl) FindAllPipelineCreatedCountInLast24Hour() (pipelineCount int, err error) {
	pipelineCount, err = impl.dbConnection.Model(&CiPipeline{}).
		Where("created_on > ?", time.Now().AddDate(0, 0, -1)).
		Count()
	return pipelineCount, err
}
func (impl *CiPipelineRepositoryImpl) FindAllDeletedPipelineCountInLast24Hour() (pipelineCount int, err error) {
	pipelineCount, err = impl.dbConnection.Model(&CiPipeline{}).
		Where("created_on > ? and deleted=?", time.Now().AddDate(0, 0, -1), true).
		Count()
	return pipelineCount, err
}

func (impl *CiPipelineRepositoryImpl) FindNumberOfAppsWithCiPipeline(appIds []int) (count int, err error) {
	var ciPipelines []*CiPipeline
	count, err = impl.dbConnection.
		Model(&ciPipelines).
		ColumnExpr("DISTINCT app_id").
		Where("app_id in (?)", pg.In(appIds)).
		Count()

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (impl *CiPipelineRepositoryImpl) FindAppAndProjectByCiPipelineIds(ciPipelineIds []int) ([]*CiPipeline, error) {
	var ciPipelines []*CiPipeline
	err := impl.dbConnection.Model(&ciPipelines).Column("ci_pipeline.*", "App", "App.Team").
		Where("ci_pipeline.id in(?)", pg.In(ciPipelineIds)).
		Where("ci_pipeline.deleted = ?", false).
		Select()
	return ciPipelines, err
}

func (impl *CiPipelineRepositoryImpl) FindCiPipelineConfigsByIds(ids []int) ([]*CiPipeline, error) {
	var ciPipelines []*CiPipeline
	err := impl.dbConnection.Model(&ciPipelines).
		Column("ci_pipeline.*", "App", "CiPipelineMaterials", "CiTemplate", "CiTemplate.DockerRegistry", "CiPipelineMaterials.GitMaterial").
		Where("ci_pipeline.id in (?)", pg.In(ids)).
		Where("ci_pipeline.deleted =? ", false).
		Select()
	return ciPipelines, err
}

func (impl *CiPipelineRepositoryImpl) FindByParentCiPipelineIds(parentCiPipelineIds []int) ([]*CiPipeline, error) {
	var ciPipelines []*CiPipeline
	err := impl.dbConnection.Model(&ciPipelines).
		Where("parent_ci_pipeline in (?)", pg.In(parentCiPipelineIds)).
		Where("active = ?", true).
		Select()
	return ciPipelines, err
}

func (impl *CiPipelineRepositoryImpl) FindAppIdsForCiPipelineIds(pipelineIds []int) (map[int]int, error) {
	ciPipelineIdVsAppId := make(map[int]int, 0)
	if len(pipelineIds) == 0 {
		return ciPipelineIdVsAppId, nil
	}

	pipelineResponse := []CiPipeline{}
	query := "select ci_pipeline.id, ci_pipeline.app_id from ci_pipeline where id in (?) and active = ?"

	_, err := impl.dbConnection.Query(&pipelineResponse, query, pg.In(pipelineIds), true)

	if err != nil && err != pg.ErrNoRows {
		return ciPipelineIdVsAppId, err
	}
	for _, ciPipeline := range pipelineResponse {
		ciPipelineIdVsAppId[ciPipeline.Id] = ciPipeline.AppId
	}

	return ciPipelineIdVsAppId, nil
}

func (impl *CiPipelineRepositoryImpl) GetCiPipelineByArtifactId(artifactId int) (*CiPipeline, error) {
	ciPipeline := &CiPipeline{}
	err := impl.dbConnection.Model(ciPipeline).
		Column("ci_pipeline.*").
		Join("INNER JOIN ci_artifact cia on cia.pipeline_id = ci_pipeline.id").
		//Where("ci_pipeline.deleted=?", false).
		Where("cia.id = ?", artifactId).
		Select()
	return ciPipeline, err
}

func (impl *CiPipelineRepositoryImpl) GetExternalCiPipelineByArtifactId(artifactId int) (*ExternalCiPipeline, error) {
	ciPipeline := &ExternalCiPipeline{}
	query := "SELECT ecp.* " +
		" FROM external_ci_pipeline ecp " +
		" INNER JOIN ci_artifact cia ON cia.external_ci_pipeline_id=ecp.id " +
		" WHERE ecp.active=true AND cia.id=?"
	_, err := impl.dbConnection.Query(ciPipeline, query, artifactId)
	return ciPipeline, err
}

func (impl *CiPipelineRepositoryImpl) FindCiPipelineByAppIdAndEnvIds(appId int, envIds []int) ([]*CiPipeline, error) {
	var pipelines []*CiPipeline
	query := `SELECT DISTINCT ci_pipeline.* FROM ci_pipeline INNER JOIN pipeline ON pipeline.ci_pipeline_id = ci_pipeline.id WHERE ci_pipeline.app_id = ? 
              AND pipeline.environment_id IN (?) AND ci_pipeline.deleted = false AND pipeline.deleted = false;`
	_, err := impl.dbConnection.Query(&pipelines, query, appId, pg.In(envIds))
	return pipelines, err
}

func (impl *CiPipelineRepositoryImpl) FindLinkedCiCount(ciPipelineId int) (int, error) {
	pipeline := &CiPipeline{}
	cnt, err := impl.dbConnection.Model(pipeline).
		Where("parent_ci_pipeline = ?", ciPipelineId).
		Where("ci_pipeline_type != ?", ciPipelineBean.LINKED_CD).
		Where("deleted = ?", false).
		Count()
	if err == pg.ErrNoRows {
		return 0, nil
	}
	return cnt, err
}

func (impl *CiPipelineRepositoryImpl) GetLinkedCiPipelines(ctx context.Context, ciPipelineId int) ([]*CiPipeline, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "GetLinkedCiPipelines")
	defer span.End()
	var linkedCIPipelines []*CiPipeline
	err := impl.dbConnection.Model(&linkedCIPipelines).
		Where("parent_ci_pipeline = ?", ciPipelineId).
		Where("ci_pipeline_type != ?", ciPipelineBean.LINKED_CD).
		Where("deleted = ?", false).
		Select()
	if err != nil {
		return nil, err
	}
	return linkedCIPipelines, nil
}

func (impl *CiPipelineRepositoryImpl) GetDownStreamInfo(ctx context.Context, sourceCiPipelineId int,
	appNameMatch, envNameMatch string, req *pagination.RepositoryRequest) ([]ciPipeline.LinkedCIDetails, int, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "GetDownStreamInfo")
	defer span.End()
	linkedCIDetails := make([]ciPipeline.LinkedCIDetails, 0)
	query := impl.dbConnection.Model().
		Table("ci_pipeline").
		// added columns that has no duplicated reference across joined tables
		Column("ci_pipeline.app_id").
		// added columns that has duplicated reference across joined tables and assign alias name
		ColumnExpr("a.app_name as app_name").
		ColumnExpr("e.environment_name as environment_name").
		ColumnExpr("p.id as pipeline_id").
		ColumnExpr("p.trigger_type as trigger_mode").
		ColumnExpr("p.environment_id as environment_id").
		// join app table
		Join("INNER JOIN app a").
		JoinOn("a.id = ci_pipeline.app_id").
		JoinOn("a.active = ?", true).
		// join pipeline table
		Join("LEFT JOIN pipeline p").
		JoinOn("p.ci_pipeline_id = ci_pipeline.id").
		JoinOn("p.deleted = ?", false).
		// join environment table
		Join("LEFT JOIN environment e").
		JoinOn("e.id = p.environment_id").
		JoinOn("e.active = ?", true).
		// constrains
		Where("ci_pipeline.parent_ci_pipeline = ?", sourceCiPipelineId).
		Where("ci_pipeline.ci_pipeline_type != ?", ciPipelineBean.LINKED_CD).
		Where("ci_pipeline.deleted = ?", false)
	// app name filtering with lower case
	if len(appNameMatch) != 0 {
		query = query.Where("LOWER(a.app_name) LIKE ?", "%"+appNameMatch+"%")
	}
	// env name filtering
	if len(envNameMatch) != 0 {
		query = query.Where("e.environment_name = ?", envNameMatch)
	}
	// get total response count
	totalCount, err := query.Count()
	if err != nil {
		return nil, 0, err
	}
	// query execution
	if req != nil {
		if len(req.SortBy) != 0 && len(req.Order) != 0 {
			query = query.Order(fmt.Sprintf("%s %s", req.SortBy, string(req.Order)))
		}
		if req.Limit != 0 {
			query = query.Limit(req.Limit).
				Offset(req.Offset)
		}
	}

	err = query.Select(&linkedCIDetails)
	if err != nil {
		return nil, 0, err
	}
	return linkedCIDetails, totalCount, err
}
