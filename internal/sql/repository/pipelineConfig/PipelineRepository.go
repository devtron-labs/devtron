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
	"encoding/json"
	"github.com/devtron-labs/common-lib-private/utils/k8s/health"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type PipelineType string
type TriggerType string //HOW pipeline should be triggered

const TRIGGER_TYPE_AUTOMATIC TriggerType = "AUTOMATIC"
const TRIGGER_TYPE_MANUAL TriggerType = "MANUAL"

type Pipeline struct {
	tableName                     struct{} `sql:"pipeline" pg:",discard_unknown_columns"`
	Id                            int      `sql:"id,pk"`
	AppId                         int      `sql:"app_id,notnull"`
	App                           app.App
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
	DeploymentAppCreated          bool        `sql:"deployment_app_created,notnull"`
	DeploymentAppType             string      `sql:"deployment_app_type,notnull"` //helm, acd
	DeploymentAppName             string      `sql:"deployment_app_name"`
	DeploymentAppDeleteRequest    bool        `sql:"deployment_app_delete_request,notnull"`
	UserApprovalConfig            string      `sql:"user_approval_config"`
	Environment                   repository.Environment
	sql.AuditLog
}

type UserApprovalConfig struct {
	RequiredCount int    `json:"requiredCount" validate:"number,required"`
	Description   string `json:"description,omitempty"`
}

func (pipeline Pipeline) ApprovalNodeConfigured() bool {
	if len(pipeline.UserApprovalConfig) > 0 {
		approvalConfig, err := pipeline.GetApprovalConfig()
		if err != nil {
			return false
		}
		return approvalConfig.RequiredCount > 0
	}
	return false
}

func (pipeline Pipeline) GetApprovalConfig() (UserApprovalConfig, error) {
	approvalConfig := UserApprovalConfig{}
	err := json.Unmarshal([]byte(pipeline.UserApprovalConfig), &approvalConfig)
	return approvalConfig, err
}

func (pipeline Pipeline) IsManualTrigger() bool {
	return pipeline.TriggerType == TRIGGER_TYPE_MANUAL
}

type PipelineRepository interface {
	Save(pipeline []*Pipeline, tx *pg.Tx) error
	Update(pipeline *Pipeline, tx *pg.Tx) error
	FindActiveByAppId(appId int) (pipelines []*Pipeline, err error)
	Delete(id int, userId int32, tx *pg.Tx) error
	FindByName(pipelineName string) (pipeline *Pipeline, err error)
	PipelineExists(pipelineName string) (bool, error)
	FindById(id int) (pipeline *Pipeline, err error)
	GetPostStageConfigById(id int) (pipeline *Pipeline, err error)
	FindAppAndEnvDetailsByPipelineId(id int) (pipeline *Pipeline, err error)
	FindActiveByEnvIdAndDeploymentType(environmentId int, deploymentAppType string, exclusionList []int, includeApps []int) ([]*Pipeline, error)
	FindByIdsIn(ids []int) ([]*Pipeline, error)
	FindByCiPipelineIdsIn(ciPipelineIds []int) ([]*Pipeline, error)
	FindAutomaticByCiPipelineId(ciPipelineId int) (pipelines []*Pipeline, err error)
	GetByEnvOverrideId(envOverrideId int) ([]Pipeline, error)
	GetByEnvOverrideIdAndEnvId(envOverrideId, envId int) (Pipeline, error)
	FindActiveByAppIdAndEnvironmentId(appId int, environmentId int) (pipelines []*Pipeline, err error)
	UndoDelete(id int) error
	UniqueAppEnvironmentPipelines() ([]*Pipeline, error)
	FindByCiPipelineId(ciPipelineId int) (pipelines []*Pipeline, err error)
	FindByParentCiPipelineId(ciPipelineId int) (pipelines []*Pipeline, err error)
	FindByPipelineTriggerGitHash(gitHash string) (pipeline *Pipeline, err error)
	FindByIdsInAndEnvironment(ids []int, environmentId int) ([]*Pipeline, error)
	FindActiveByAppIdAndEnvironmentIdV2() (pipelines []*Pipeline, err error)
	GetConnection() *pg.DB
	FindAllPipelineInLast24Hour() (pipelines []*Pipeline, err error)
	FindActiveByEnvId(envId int) (pipelines []*Pipeline, err error)
	FindActiveByEnvIds(envId []int) (pipelines []*Pipeline, err error)
	FindActiveByInFilter(envId int, appIdIncludes []int) (pipelines []*Pipeline, err error)
	FindActiveByNotFilter(envId int, appIdExcludes []int) (pipelines []*Pipeline, err error)
	FindAllPipelinesByChartsOverrideAndAppIdAndChartId(chartOverridden bool, appId int, chartId int) (pipelines []*Pipeline, err error)
	FindActiveByAppIdAndPipelineId(appId int, pipelineId int) ([]*Pipeline, error)
	SetDeploymentAppCreatedInPipeline(deploymentAppCreated bool, pipelineId int, userId int32) error
	UpdateCdPipelineDeploymentAppInFilter(deploymentAppType string, cdPipelineIdIncludes []int, userId int32, deploymentAppCreated bool, delete bool) error
	UpdateCdPipelineAfterDeployment(deploymentAppType string, cdPipelineIdIncludes []int, userId int32, delete bool) error
	FindNumberOfAppsWithCdPipeline(appIds []int) (count int, err error)
	GetAppAndEnvDetailsForDeploymentAppTypePipeline(deploymentAppType string, clusterIds []int) ([]*Pipeline, error)
	GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelines(pendingSinceSeconds int, timeForDegradation int) ([]*Pipeline, error)
	GetArgoPipelinesHavingLatestTriggerStuckInNonTerminalStatuses(deployedBeforeMinutes int, getPipelineDeployedWithinHours int) ([]*Pipeline, error)
	FindIdsByAppIdsAndEnvironmentIds(appIds, environmentIds []int) (ids []int, err error)
	FindIdsByProjectIdsAndEnvironmentIds(projectIds, environmentIds []int) ([]int, error)

	GetArgoPipelineByArgoAppName(argoAppName string) (Pipeline, error)
	GetPartiallyDeletedPipelineByStatus(appId int, envId int) (Pipeline, error)
	FindActiveByAppIds(appIds []int) (pipelines []*Pipeline, err error)
	FindAppAndEnvironmentAndProjectByPipelineIds(pipelineIds []int) (pipelines []*Pipeline, err error)
	FilterDeploymentDeleteRequestedPipelineIds(cdPipelineIds []int) (map[int]bool, error)
	FindDeploymentTypeByPipelineIds(cdPipelineIds []int) (map[int]DeploymentObject, error)
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

type DeploymentObject struct {
	DeploymentType models.DeploymentType `sql:"deployment_type"`
	PipelineId     int                   `sql:"pipeline_id"`
	Status         string                `sql:"status"`
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
		Column("pipeline.*", "App", "Environment", "Environment.Cluster").
		Join("inner join app a on pipeline.app_id = a.id").
		Join("inner join environment e on pipeline.environment_id = e.id").
		Join("inner join cluster c on c.id = e.cluster_id").
		Where("pipeline.id in (?)", pg.In(ids)).
		Where("pipeline.deleted = false").
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
func (impl PipelineRepositoryImpl) FindByParentCiPipelineId(ciPipelineId int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Column("pipeline.*").
		Join("INNER JOIN app_workflow_mapping awm on awm.component_id = pipeline.id").
		Where("pipeline.ci_pipeline_id =?", ciPipelineId).
		Where("awm.parent_type =?", appWorkflow.CIPIPELINE).
		Where("pipeline.deleted =?", false).
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
		Column("pipeline.*", "Environment", "App").
		Where("app_id = ?", appId).
		Where("deleted = ?", false).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) FindActiveByAppIdAndEnvironmentId(appId int, environmentId int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Column("pipeline.*", "Environment", "App").
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

func (impl PipelineRepositoryImpl) Delete(id int, userId int32, tx *pg.Tx) error {
	pipeline := &Pipeline{}
	r, err := tx.Model(pipeline).Set("deleted =?", true).Set("deployment_app_created =?", false).
		Set("updated_on = ?", time.Now()).Set("updated_by = ?", userId).Where("id =?", id).Update()
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

func (impl PipelineRepositoryImpl) GetPostStageConfigById(id int) (pipeline *Pipeline, err error) {
	pipeline = &Pipeline{}
	err = impl.dbConnection.
		Model(pipeline).
		Column("pipeline.post_stage_config_yaml").
		Where("pipeline.id = ?", id).
		Where("deleted = ?", false).
		Select()
	return pipeline, err
}

func (impl PipelineRepositoryImpl) FindAppAndEnvDetailsByPipelineId(id int) (pipeline *Pipeline, err error) {
	pipeline = &Pipeline{}
	err = impl.dbConnection.
		Model(pipeline).
		Column("App.id", "App.app_name", "App.app_type", "Environment.id", "Environment.cluster_id").
		Join("inner join app a on pipeline.app_id = a.id").
		Join("inner join environment e on pipeline.environment_id = e.id").
		Where("pipeline.id = ?", id).
		Where("deleted = ?", false).
		Select()
	return pipeline, err
}

// FindActiveByEnvIdAndDeploymentType takes in environment id and current deployment app type
// and fetches and returns a list of pipelines matching the same excluding given app ids.
func (impl PipelineRepositoryImpl) FindActiveByEnvIdAndDeploymentType(environmentId int,
	deploymentAppType string, exclusionList []int, includeApps []int) ([]*Pipeline, error) {

	// NOTE: PG query throws error with slice of integer
	exclusionListString := []string{}
	for _, appId := range exclusionList {
		exclusionListString = append(exclusionListString, strconv.Itoa(appId))
	}

	inclusionListString := []string{}
	for _, appId := range includeApps {
		inclusionListString = append(inclusionListString, strconv.Itoa(appId))
	}

	var pipelines []*Pipeline

	query := impl.dbConnection.
		Model(&pipelines).
		Column("pipeline.*", "App", "Environment").
		Join("inner join app a on pipeline.app_id = a.id").
		Where("pipeline.environment_id = ?", environmentId).
		Where("pipeline.deployment_app_type = ?", deploymentAppType).
		Where("pipeline.deleted = ?", false)

	if len(exclusionListString) > 0 {
		query.Where("pipeline.app_id not in (?)", pg.In(exclusionListString))
	}

	if len(inclusionListString) > 0 {
		query.Where("pipeline.app_id in (?)", pg.In(inclusionListString))
	}

	err := query.Select()
	return pipelines, err
}

// Deprecated:
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
		" INNER JOIN charts ch on ch.id = ceco.chart_id" +
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

func (impl PipelineRepositoryImpl) GetByEnvOverrideIdAndEnvId(envOverrideId, envId int) (Pipeline, error) {
	var pipeline Pipeline
	query := "" +
		" SELECT p.*" +
		" FROM chart_env_config_override ceco" +
		" INNER JOIN charts ch on ch.id = ceco.chart_id" +
		" INNER JOIN environment env on env.id = ceco.target_environment" +
		" INNER JOIN app ap on ap.id = ch.app_id" +
		" INNER JOIN pipeline p on p.app_id = ap.id" +
		" WHERE ceco.id=? and p.environment_id=?;"
	_, err := impl.dbConnection.Query(&pipeline, query, envOverrideId, envId)

	if err != nil {
		return pipeline, err
	}
	return pipeline, err
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

func (impl PipelineRepositoryImpl) FindAllPipelineInLast24Hour() (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Column("pipeline.*").
		Where("created_on > ?", time.Now().AddDate(0, 0, -1)).
		Select()
	return pipelines, err
}
func (impl PipelineRepositoryImpl) FindActiveByEnvId(envId int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).Column("pipeline.*", "App", "Environment").
		Where("environment_id = ?", envId).
		Where("deleted = ?", false).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) FindActiveByEnvIds(envIds []int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).Column("pipeline.*").
		Where("environment_id in (?)", pg.In(envIds)).
		Where("deleted = ?", false).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) FindActiveByInFilter(envId int, appIdIncludes []int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).Column("pipeline.*", "App", "Environment").
		Where("environment_id = ?", envId).
		Where("app_id in (?)", pg.In(appIdIncludes)).
		Where("deleted = ?", false).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) FindActiveByNotFilter(envId int, appIdExcludes []int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).Column("pipeline.*", "App", "Environment").
		Where("environment_id = ?", envId).
		Where("app_id not in (?)", pg.In(appIdExcludes)).
		Where("deleted = ?", false).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) FindAllPipelinesByChartsOverrideAndAppIdAndChartId(hasConfigOverridden bool, appId int, chartId int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Column("pipeline.*").
		Join("inner join charts on pipeline.app_id = charts.app_id").
		Join("inner join chart_env_config_override ceco on charts.id = ceco.chart_id").
		Where("pipeline.app_id = ?", appId).
		Where("charts.id = ?", chartId).
		Where("ceco.is_override = ?", hasConfigOverridden).
		Where("pipeline.deleted = ?", false).
		Where("ceco.active = ?", true).
		Where("charts.active = ?", true).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) FindActiveByAppIdAndPipelineId(appId int, pipelineId int) ([]*Pipeline, error) {
	var pipelines []*Pipeline
	err := impl.dbConnection.Model(&pipelines).
		Where("app_id = ?", appId).
		Where("ci_pipeline_id = ?", pipelineId).
		Where("deleted = ?", false).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) SetDeploymentAppCreatedInPipeline(deploymentAppCreated bool, pipelineId int, userId int32) error {
	query := "update pipeline set deployment_app_created=?, updated_on=?, updated_by=? where id=?;"
	var pipeline *Pipeline
	_, err := impl.dbConnection.Query(pipeline, query, deploymentAppCreated, time.Now(), userId, pipelineId)
	return err
}

// UpdateCdPipelineDeploymentAppInFilter takes in deployment app type and list of cd pipeline ids and
// updates the deployment_app_type and sets deployment_app_created to false in the table for given ids.
func (impl PipelineRepositoryImpl) UpdateCdPipelineDeploymentAppInFilter(deploymentAppType string,
	cdPipelineIdIncludes []int, userId int32, deploymentAppCreated bool, isDeleted bool) error {
	query := "update pipeline set deployment_app_created = ?, deployment_app_type = ?, " +
		"updated_by = ?, updated_on = ?, deployment_app_delete_request = ? where id in (?);"
	var pipeline *Pipeline
	_, err := impl.dbConnection.Query(pipeline, query, deploymentAppCreated, deploymentAppType, userId, time.Now(), isDeleted, pg.In(cdPipelineIdIncludes))

	return err
}

func (impl PipelineRepositoryImpl) UpdateCdPipelineAfterDeployment(deploymentAppType string,
	cdPipelineIdIncludes []int, userId int32, isDeleted bool) error {
	query := "update pipeline set deployment_app_type = ?, " +
		"updated_by = ?, updated_on = ?, deployment_app_delete_request = ? where id in (?);"
	var pipeline *Pipeline
	_, err := impl.dbConnection.Query(pipeline, query, deploymentAppType, userId, time.Now(), isDeleted, pg.In(cdPipelineIdIncludes))

	return err
}

func (impl PipelineRepositoryImpl) FindNumberOfAppsWithCdPipeline(appIds []int) (count int, err error) {
	var pipelines []*Pipeline
	count, err = impl.dbConnection.
		Model(&pipelines).
		ColumnExpr("DISTINCT app_id").
		Where("app_id in (?)", pg.In(appIds)).
		Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (impl PipelineRepositoryImpl) GetAppAndEnvDetailsForDeploymentAppTypePipeline(deploymentAppType string, clusterIds []int) ([]*Pipeline, error) {
	var pipelines []*Pipeline
	err := impl.dbConnection.
		Model(&pipelines).
		Column("pipeline.id", "App.app_name", "Environment.cluster_id", "Environment.namespace", "Environment.environment_name").
		Join("inner join app a on pipeline.app_id = a.id").
		Join("inner join environment e on pipeline.environment_id = e.id").
		Where("e.cluster_id in (?)", pg.In(clusterIds)).
		Where("a.active = ?", true).
		Where("pipeline.deleted = ?", false).
		Where("pipeline.deployment_app_type = ?", deploymentAppType).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelines(pendingSinceSeconds int, timeForDegradation int) ([]*Pipeline, error) {
	var pipelines []*Pipeline
	queryString := `select p.* from pipeline p inner join cd_workflow cw on cw.pipeline_id = p.id  
    inner join cd_workflow_runner cwr on cwr.cd_workflow_id=cw.id  
    where cwr.id in (select cd_workflow_runner_id from pipeline_status_timeline  
					where id in  
						(select DISTINCT ON (cd_workflow_runner_id) max(id) as id from pipeline_status_timeline 
							group by cd_workflow_runner_id, id order by cd_workflow_runner_id,id desc)  
					and status in (?) and status_time < NOW() - INTERVAL '? seconds')  
    and cwr.started_on > NOW() - INTERVAL '? minutes' and p.deployment_app_type=? and p.deleted=?;`
	_, err := impl.dbConnection.Query(&pipelines, queryString,
		pg.In([]TimelineStatus{TIMELINE_STATUS_KUBECTL_APPLY_SYNCED,
			TIMELINE_STATUS_FETCH_TIMED_OUT, TIMELINE_STATUS_UNABLE_TO_FETCH_STATUS}),
		pendingSinceSeconds, timeForDegradation, util.PIPELINE_DEPLOYMENT_TYPE_ACD, false)
	if err != nil {
		impl.logger.Errorw("error in GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelines", "err", err)
		return nil, err
	}
	return pipelines, nil
}

func (impl PipelineRepositoryImpl) GetArgoPipelinesHavingLatestTriggerStuckInNonTerminalStatuses(getPipelineDeployedBeforeMinutes int, getPipelineDeployedWithinHours int) ([]*Pipeline, error) {
	var pipelines []*Pipeline
	queryString := `select p.id from pipeline p inner join cd_workflow cw on cw.pipeline_id = p.id  
    inner join cd_workflow_runner cwr on cwr.cd_workflow_id=cw.id  
    where cwr.id in (select id from cd_workflow_runner 
                     	where started_on < NOW() - INTERVAL '? minutes' and started_on > NOW() - INTERVAL '? hours' and status not in (?) 
                     	and workflow_type=? and cd_workflow_id in (select DISTINCT ON (pipeline_id) max(id) as id from cd_workflow
                     	  group by pipeline_id, id order by pipeline_id, id desc))
    and p.deployment_app_type=? and p.deleted=?;`
	_, err := impl.dbConnection.Query(&pipelines, queryString, getPipelineDeployedBeforeMinutes, getPipelineDeployedWithinHours,
		pg.In([]string{WorkflowAborted, WorkflowFailed, WorkflowSucceeded, string(health.HealthStatusHealthy), string(health.HealthStatusDegraded)}),
		bean.CD_WORKFLOW_TYPE_DEPLOY, util.PIPELINE_DEPLOYMENT_TYPE_ACD, false)
	if err != nil {
		impl.logger.Errorw("error in GetArgoPipelinesHavingLatestTriggerStuckInNonTerminalStatuses", "err", err)
		return nil, err
	}
	return pipelines, nil
}

func (impl PipelineRepositoryImpl) FindIdsByAppIdsAndEnvironmentIds(appIds, environmentIds []int) ([]int, error) {
	var pipelineIds []int
	query := "select id from pipeline where app_id in (?) and environment_id in (?) and deleted = ?;"
	_, err := impl.dbConnection.Query(&pipelineIds, query, pg.In(appIds), pg.In(environmentIds), false)
	if err != nil {
		impl.logger.Errorw("error in getting pipelineIds by appIds and envIds", "err", err, "appIds", appIds, "envIds", environmentIds)
		return pipelineIds, err
	}
	return pipelineIds, err
}

func (impl PipelineRepositoryImpl) FindIdsByProjectIdsAndEnvironmentIds(projectIds, environmentIds []int) ([]int, error) {
	var pipelineIds []int
	query := "select p.id from pipeline p inner join app a on a.id=p.app_id where a.team_id in (?) and p.environment_id in (?) and p.deleted = ? and a.active = ?;"
	_, err := impl.dbConnection.Query(&pipelineIds, query, pg.In(projectIds), pg.In(environmentIds), false, true)
	if err != nil {
		impl.logger.Errorw("error in getting pipelineIds by projectIds and envIds", "err", err, "projectIds", projectIds, "envIds", environmentIds)
		return pipelineIds, err
	}
	return pipelineIds, err
}

func (impl PipelineRepositoryImpl) GetArgoPipelineByArgoAppName(argoAppName string) (Pipeline, error) {
	var pipeline Pipeline
	err := impl.dbConnection.Model(&pipeline).
		Where("deployment_app_name = ?", argoAppName).
		Where("deployment_app_type = ?", util.PIPELINE_DEPLOYMENT_TYPE_ACD).
		Where("deleted = ?", false).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting pipeline by argoAppName", "err", err, "argoAppName", argoAppName)
		return pipeline, err
	}
	return pipeline, nil
}

func (impl PipelineRepositoryImpl) GetPartiallyDeletedPipelineByStatus(appId int, envId int) (Pipeline, error) {
	var pipeline Pipeline
	err := impl.dbConnection.Model(&pipeline).
		Column("pipeline.*", "App.app_name", "Environment.namespace").
		Where("app_id = ?", appId).
		Where("environment_id = ?", envId).
		Where("deployment_app_delete_request = ?", true).
		Where("deployment_app_type = ?", util.PIPELINE_DEPLOYMENT_TYPE_ACD).
		Where("deleted = ?", false).Select()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in updating argo pipeline delete status")
	}
	return pipeline, err
}

func (impl PipelineRepositoryImpl) FindActiveByAppIds(appIds []int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).
		Column("pipeline.*", "App", "Environment").
		Where("app_id in(?)", pg.In(appIds)).
		Where("deleted = ?", false).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) FindAppAndEnvironmentAndProjectByPipelineIds(pipelineIds []int) (pipelines []*Pipeline, err error) {
	err = impl.dbConnection.Model(&pipelines).Column("pipeline.*", "App", "Environment", "App.Team").
		Where("pipeline.id in(?)", pg.In(pipelineIds)).
		Where("pipeline.deleted = ?", false).
		Select()
	return pipelines, err
}

func (impl PipelineRepositoryImpl) FilterDeploymentDeleteRequestedPipelineIds(cdPipelineIds []int) (map[int]bool, error) {
	var pipelineIds []int
	pipelineIdsMap := make(map[int]bool)
	query := "select pipeline.id from pipeline where pipeline.id in (?) and pipeline.deployment_app_delete_request = ?;"
	_, err := impl.dbConnection.Query(&pipelineIds, query, pg.In(cdPipelineIds), true)
	if err != nil {
		return pipelineIdsMap, err
	}
	for _, pipelineId := range pipelineIds {
		pipelineIdsMap[pipelineId] = true
	}
	return pipelineIdsMap, nil
}

func (impl PipelineRepositoryImpl) FindDeploymentTypeByPipelineIds(cdPipelineIds []int) (map[int]DeploymentObject, error) {

	pipelineIdsMap := make(map[int]DeploymentObject)

	var deploymentType []DeploymentObject
	query := "with pcos as(select max(id) as id from pipeline_config_override where pipeline_id in (?) " +
		"group by pipeline_id) select pco.deployment_type,pco.pipeline_id, aps.status from pipeline_config_override " +
		"pco inner join pcos on pcos.id=pco.id" +
		" inner join pipeline p on p.id=pco.pipeline_id left join app_status aps on aps.app_id=p.app_id " +
		"and aps.env_id=p.environment_id;"

	_, err := impl.dbConnection.Query(&deploymentType, query, pg.In(cdPipelineIds), true)
	if err != nil {
		return pipelineIdsMap, err
	}

	for _, v := range deploymentType {
		pipelineIdsMap[v.PipelineId] = v
	}

	return pipelineIdsMap, nil
}
