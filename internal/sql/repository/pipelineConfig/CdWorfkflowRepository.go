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
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type CdWorkflowRepository interface {
	SaveWorkFlow(wf *CdWorkflow) error
	UpdateWorkFlow(wf *CdWorkflow) error
	FindById(wfId int) (*CdWorkflow, error)
	FindCdWorkflowMetaByEnvironmentId(appId int, environmentId int, offset int, size int) ([]CdWorkflowRunner, error)
	FindCdWorkflowMetaByPipelineId(pipelineId int, offset int, size int) ([]CdWorkflowRunner, error)
	FindArtifactByPipelineIdAndRunnerType(pipelineId int, runnerType bean.WorkflowType, limit int) ([]CdWorkflowRunner, error)

	SaveWorkFlowRunner(wfr *CdWorkflowRunner) (*CdWorkflowRunner, error)
	UpdateWorkFlowRunner(wfr *CdWorkflowRunner) error
	UpdateWorkFlowRunnersWithTxn(wfrs []CdWorkflowRunner, tx *pg.Tx) error
	UpdateWorkFlowRunners(wfr []*CdWorkflowRunner) error
	FindWorkflowRunnerByCdWorkflowId(wfIds []int) ([]*CdWorkflowRunner, error)
	FindPreviousCdWfRunnerByStatus(pipelineId int, currentWFRunnerId int, status []string) ([]*CdWorkflowRunner, error)
	FindConfigByPipelineId(pipelineId int) (*CdWorkflowConfig, error)
	FindWorkflowRunnerById(wfrId int) (*CdWorkflowRunner, error)
	FindLatestWfrByAppIdAndEnvironmentId(appId int, environmentId int) (CdWorkflowRunner, error)
	FindCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId int, environmentId int, runnerType bean.WorkflowType) (CdWorkflowRunner, error)

	GetConnection() *pg.DB

	FindLastPreOrPostTriggeredByPipelineId(pipelineId int) (CdWorkflowRunner, error)
	FindLastPreOrPostTriggeredByEnvironmentId(appId int, environmentId int) (CdWorkflowRunner, error)

	FindByWorkflowIdAndRunnerType(wfId int, runnerType bean.WorkflowType) (CdWorkflowRunner, error)
	FindLastStatusByPipelineIdAndRunnerType(pipelineId int, runnerType bean.WorkflowType) (CdWorkflowRunner, error)
	SaveWorkFlows(wfs ...*CdWorkflow) error
	IsLatestWf(pipelineId int, wfId int) (bool, error)
	FindLatestCdWorkflowByPipelineId(pipelineIds []int) (*CdWorkflow, error)
	FindLatestCdWorkflowByPipelineIdV2(pipelineIds []int) ([]*CdWorkflow, error)
	FetchAllCdStagesLatestEntity(pipelineIds []int) ([]*CdWorkflowStatus, error)
	FetchAllCdStagesLatestEntityStatus(wfrIds []int) ([]*CdWorkflowRunner, error)
	ExistsByStatus(status string) (bool, error)
}

type CdWorkflowRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

type WorkflowStatus int

const (
	WF_UNKNOWN WorkflowStatus = iota
	REQUEST_ACCEPTED
	ENQUEUED
	QUE_ERROR
	WF_STARTED
	DROPPED_STALE
	DEQUE_ERROR
	TRIGGER_ERROR
)

func (a WorkflowStatus) String() string {
	return [...]string{"WF_UNKNOWN", "REQUEST_ACCEPTED", "ENQUEUED", "QUE_ERROR", "WF_STARTED", "DROPPED_STALE", "DEQUE_ERROR", "TRIGGER_ERROR"}[a]
}

type CdWorkflow struct {
	tableName        struct{}       `sql:"cd_workflow" pg:",discard_unknown_columns"`
	Id               int            `sql:"id,pk"`
	CiArtifactId     int            `sql:"ci_artifact_id"`
	PipelineId       int            `sql:"pipeline_id"`
	WorkflowStatus   WorkflowStatus `sql:"workflow_status,notnull"`
	Pipeline         *Pipeline
	CiArtifact       *repository.CiArtifact
	CdWorkflowRunner []CdWorkflowRunner
	sql.AuditLog
}

type CdWorkflowConfig struct {
	tableName                struct{} `sql:"cd_workflow_config" pg:",discard_unknown_columns"`
	Id                       int      `sql:"id,pk"`
	CdTimeout                int64    `sql:"cd_timeout"`
	MinCpu                   string   `sql:"min_cpu"`
	MaxCpu                   string   `sql:"max_cpu"`
	MinMem                   string   `sql:"min_mem"`
	MaxMem                   string   `sql:"max_mem"`
	MinStorage               string   `sql:"min_storage"`
	MaxStorage               string   `sql:"max_storage"`
	MinEphStorage            string   `sql:"min_eph_storage"`
	MaxEphStorage            string   `sql:"max_eph_storage"`
	CdCacheBucket            string   `sql:"cd_cache_bucket"`
	CdCacheRegion            string   `sql:"cd_cache_region"`
	CdImage                  string   `sql:"cd_image"`
	Namespace                string   `sql:"wf_namespace"`
	CdPipelineId             int      `sql:"cd_pipeline_id"`
	LogsBucket               string   `sql:"logs_bucket"`
	CdArtifactLocationFormat string   `sql:"cd_artifact_location_format"`
}

type WorkflowExecutorType string

const WORKFLOW_EXECUTOR_TYPE_AWF = "AWF"
const WORKFLOW_EXECUTOR_TYPE_SYSTEM = "SYSTEM"

type CdWorkflowRunner struct {
	tableName          struct{}             `sql:"cd_workflow_runner" pg:",discard_unknown_columns"`
	Id                 int                  `sql:"id,pk"`
	Name               string               `sql:"name"`
	WorkflowType       bean.WorkflowType    `sql:"workflow_type"` //pre,post,deploy
	ExecutorType       WorkflowExecutorType `sql:"executor_type"` //awf, system
	Status             string               `sql:"status"`
	PodStatus          string               `sql:"pod_status"`
	Message            string               `sql:"message"`
	StartedOn          time.Time            `sql:"started_on"`
	FinishedOn         time.Time            `sql:"finished_on"`
	Namespace          string               `sql:"namespace"`
	BlobStorageEnabled bool                 `sql:"blob_storage_enabled,notnull"`
	LogLocation        string               `sql:"log_file_path"`
	TriggeredBy        int32                `sql:"triggered_by"`
	CdWorkflowId       int                  `sql:"cd_workflow_id"`
	CdWorkflow         *CdWorkflow
}

type CdWorkflowWithArtifact struct {
	Id                 int       `json:"id"`
	CdWorkflowId       int       `json:"cd_workflow_id"`
	Name               string    `json:"name"`
	Status             string    `json:"status"`
	PodStatus          string    `json:"pod_status"`
	Message            string    `json:"message"`
	StartedOn          time.Time `json:"started_on"`
	FinishedOn         time.Time `json:"finished_on"`
	PipelineId         int       `json:"pipeline_id"`
	Namespace          string    `json:"namespace"`
	LogFilePath        string    `json:"log_file_path"`
	TriggeredBy        int32     `json:"triggered_by"`
	EmailId            string    `json:"email_id"`
	Image              string    `json:"image"`
	MaterialInfo       string    `json:"material_info,omitempty"`
	DataSource         string    `json:"data_source,omitempty"`
	CiArtifactId       int       `json:"ci_artifact_id,omitempty"`
	WorkflowType       string    `json:"workflow_type,omitempty"`
	ExecutorType       string    `json:"executor_type,omitempty"`
	BlobStorageEnabled bool      `json:"blobStorageEnabled"`
}

type TriggerWorkflowStatus struct {
	CdWorkflowStatus []*CdWorkflowStatus `json:"cdWorkflowStatus"`
	CiWorkflowStatus []*CiWorkflowStatus `json:"ciWorkflowStatus"`
}

type CdWorkflowStatus struct {
	CiPipelineId int    `json:"ci_pipeline_id"`
	PipelineId   int    `json:"pipeline_id"`
	PipelineName string `json:"pipeline_name,omitempty"`
	DeployStatus string `json:"deploy_status"`
	PreStatus    string `json:"pre_status"`
	PostStatus   string `json:"post_status"`
	WorkflowType string `json:"workflow_type,omitempty"`
	WfrId        int    `json:"wfr_id,omitempty"`
}

type CiWorkflowStatus struct {
	CiPipelineId      int    `json:"ciPipelineId"`
	CiPipelineName    string `json:"ciPipelineName,omitempty"`
	CiStatus          string `json:"ciStatus"`
	StorageConfigured bool   `json:"storageConfigured"`
}

func NewCdWorkflowRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CdWorkflowRepositoryImpl {
	return &CdWorkflowRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl *CdWorkflowRepositoryImpl) FindPreviousCdWfRunnerByStatus(pipelineId int, currentWFRunnerId int, status []string) ([]*CdWorkflowRunner, error) {
	var runner []*CdWorkflowRunner
	err := impl.dbConnection.
		Model(&runner).
		Column("cd_workflow_runner.*", "CdWorkflow").
		Where("cd_workflow.pipeline_id = ?", pipelineId).
		Where("cd_workflow_runner.id < ?", currentWFRunnerId).
		Where("workflow_type = ? ", bean.CD_WORKFLOW_TYPE_DEPLOY).
		Where("cd_workflow_runner.status not in (?) ", pg.In(status)).
		Order("cd_workflow_runner.id DESC").
		Select()
	return runner, err
}

func (impl *CdWorkflowRepositoryImpl) SaveWorkFlow(wf *CdWorkflow) error {
	err := impl.dbConnection.Insert(wf)
	return err
}
func (impl *CdWorkflowRepositoryImpl) SaveWorkFlows(wfs ...*CdWorkflow) error {
	err := impl.dbConnection.Insert(&wfs)
	return err
}

func (impl *CdWorkflowRepositoryImpl) UpdateWorkFlow(wf *CdWorkflow) error {
	_, err := impl.dbConnection.Model(wf).WherePK().UpdateNotNull()
	return err
}

func (impl *CdWorkflowRepositoryImpl) FindById(wfId int) (*CdWorkflow, error) {
	ddWorkflow := &CdWorkflow{}
	err := impl.dbConnection.Model(ddWorkflow).
		Column("cd_workflow.*, CdWorkflowRunner").Where("id = ?", wfId).Select()
	return ddWorkflow, err
}

func (impl *CdWorkflowRepositoryImpl) FindConfigByPipelineId(pipelineId int) (*CdWorkflowConfig, error) {
	cdWorkflowConfig := &CdWorkflowConfig{}
	err := impl.dbConnection.Model(cdWorkflowConfig).Where("cd_pipeline_id = ?", pipelineId).Select()
	return cdWorkflowConfig, err
}

func (impl *CdWorkflowRepositoryImpl) FindLatestCdWorkflowByPipelineId(pipelineIds []int) (*CdWorkflow, error) {
	cdWorkflow := &CdWorkflow{}
	err := impl.dbConnection.Model(cdWorkflow).Where("pipeline_id in (?)", pg.In(pipelineIds)).Order("id DESC").Limit(1).Select()
	return cdWorkflow, err
}

func (impl *CdWorkflowRepositoryImpl) FindLatestCdWorkflowByPipelineIdV2(pipelineIds []int) ([]*CdWorkflow, error) {
	var cdWorkflow []*CdWorkflow
	//err := impl.dbConnection.Model(&cdWorkflow).Where("pipeline_id in (?)", pg.In(pipelineIds)).Order("id DESC").Select()
	query := "SELECT cdw.pipeline_id, cdw.workflow_status, MAX(id) as id from cd_workflow cdw" +
		" WHERE cdw.pipeline_id in(?)" +
		" GROUP by cdw.pipeline_id, cdw.workflow_status ORDER by id desc;"
	_, err := impl.dbConnection.Query(&cdWorkflow, query, pg.In(pipelineIds))
	if err != nil {
		return cdWorkflow, err
	}
	//TODO - Group By Environment And Pipeline will get latest pipeline from top
	return cdWorkflow, err
}

func (impl *CdWorkflowRepositoryImpl) FindCdWorkflowMetaByEnvironmentId(appId int, environmentId int, offset int, limit int) ([]CdWorkflowRunner, error) {
	var wfrList []CdWorkflowRunner
	err := impl.dbConnection.
		Model(&wfrList).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("p.environment_id = ?", environmentId).
		Where("p.app_id = ?", appId).
		Order("cd_workflow_runner.id DESC").
		Join("inner join cd_workflow wf on wf.id = cd_workflow_runner.cd_workflow_id").
		Join("inner join ci_artifact cia on cia.id = wf.ci_artifact_id").
		Join("inner join pipeline p on p.id = wf.pipeline_id").
		//Join("left join users u on u.id = wfr.triggered_by").
		Offset(offset).Limit(limit).
		Select()
	if err != nil {
		return nil, err
	}
	return wfrList, err
}

func (impl *CdWorkflowRepositoryImpl) FindCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId int, environmentId int, runnerType bean.WorkflowType) (CdWorkflowRunner, error) {
	var wfr CdWorkflowRunner
	err := impl.dbConnection.
		Model(&wfr).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("p.environment_id = ?", environmentId).
		Where("p.app_id = ?", appId).
		Where("cd_workflow_runner.workflow_type = ?", runnerType).
		Order("cd_workflow_runner.id DESC").
		Join("inner join cd_workflow wf on wf.id = cd_workflow_runner.cd_workflow_id").
		Join("inner join ci_artifact cia on cia.id = wf.ci_artifact_id").
		Join("inner join pipeline p on p.id = wf.pipeline_id").
		Order("cd_workflow_runner.id DESC").Limit(1).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting cdWfr by appId, envId and runner type", "appId", appId, "envId", environmentId, "runnerType", runnerType)
		return wfr, err
	}
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}

func (impl *CdWorkflowRepositoryImpl) FindCdWorkflowMetaByPipelineId(pipelineId int, offset int, limit int) ([]CdWorkflowRunner, error) {
	var wfrList []CdWorkflowRunner
	err := impl.dbConnection.
		Model(&wfrList).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("cd_workflow.pipeline_id = ?", pipelineId).
		Order("cd_workflow_runner.id DESC").
		//Join("inner join cd_workflow wf on wf.id = cd_workflow_runner.cd_workflow_id").
		//Join("inner join ci_artifact cia on cia.id = wf.ci_artifact_id").
		//Join("inner join pipeline p on p.id = wf.pipeline_id").
		//Join("left join users u on u.id = wfr.triggered_by").
		//Order("ORDER BY cd_workflow_runner.started_on DESC").
		Offset(offset).Limit(limit).
		Select()

	if err != nil {
		return nil, err
	}
	return wfrList, err
}

func (impl *CdWorkflowRepositoryImpl) FindArtifactByPipelineIdAndRunnerType(pipelineId int, runnerType bean.WorkflowType, limit int) ([]CdWorkflowRunner, error) {
	var wfrList []CdWorkflowRunner
	err := impl.dbConnection.
		Model(&wfrList).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("cd_workflow.pipeline_id = ?", pipelineId).
		Where("cd_workflow_runner.workflow_type = ?", runnerType).
		Order("cd_workflow_runner.id DESC").
		//Join("inner join cd_workflow wf on wf.id = cd_workflow_runner.cd_workflow_id").
		//Join("inner join ci_artifact cia on cia.id = wf.ci_artifact_id").
		//Join("inner join pipeline p on p.id = wf.pipeline_id").
		//Join("left join users u on u.id = wfr.triggered_by").
		//Order("ORDER BY cd_workflow_runner.started_on DESC").
		Limit(limit).
		Select()
	if err != nil {
		return nil, err
	}
	return wfrList, err
}

func (impl *CdWorkflowRepositoryImpl) FindLastPreOrPostTriggeredByPipelineId(pipelineId int) (CdWorkflowRunner, error) {
	wfr := CdWorkflowRunner{}
	err := impl.dbConnection.
		Model(&wfr).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("cd_workflow.pipeline_id = ?", pipelineId).
		Where("cd_workflow_runner.workflow_type != ?", bean.CD_WORKFLOW_TYPE_DEPLOY).
		Order("cd_workflow_runner.id DESC").
		Limit(1).
		Select()
	if err != nil {
		return wfr, err
	}
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) FindLatestWfrByAppIdAndEnvironmentId(appId int, environmentId int) (CdWorkflowRunner, error) {
	wfr := CdWorkflowRunner{}
	err := impl.dbConnection.
		Model(&wfr).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("p.environment_id = ?", environmentId).
		Where("p.app_id = ?", appId).
		Where("cd_workflow_runner.workflow_type = ?", bean.CD_WORKFLOW_TYPE_DEPLOY).
		Order("cd_workflow_runner.id DESC").
		Join("inner join cd_workflow wf on wf.id = cd_workflow_runner.cd_workflow_id").
		Join("inner join ci_artifact cia on cia.id = wf.ci_artifact_id").
		Join("inner join pipeline p on p.id = wf.pipeline_id").
		Limit(1).
		Select()
	if err != nil {
		return wfr, err
	}
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) FindLastPreOrPostTriggeredByEnvironmentId(appId int, environmentId int) (CdWorkflowRunner, error) {
	wfr := CdWorkflowRunner{}
	err := impl.dbConnection.
		Model(&wfr).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("p.environment_id = ?", environmentId).
		Where("p.app_id = ?", appId).
		Where("cd_workflow_runner.workflow_type != ?", bean.CD_WORKFLOW_TYPE_DEPLOY).
		Order("cd_workflow_runner.id DESC").
		Join("inner join cd_workflow wf on wf.id = cd_workflow_runner.cd_workflow_id").
		Join("inner join ci_artifact cia on cia.id = wf.ci_artifact_id").
		Join("inner join pipeline p on p.id = wf.pipeline_id").
		Limit(1).
		Select()
	if err != nil {
		return wfr, err
	}
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) SaveWorkFlowRunner(wfr *CdWorkflowRunner) (*CdWorkflowRunner, error) {
	err := impl.dbConnection.Insert(wfr)
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) UpdateWorkFlowRunner(wfr *CdWorkflowRunner) error {
	err := impl.dbConnection.Update(wfr)
	return err
}

func (impl *CdWorkflowRepositoryImpl) UpdateWorkFlowRunnersWithTxn(wfrs []CdWorkflowRunner, tx *pg.Tx) error {
	_, err := tx.Model(&wfrs).Update()
	return err
}

func (impl *CdWorkflowRepositoryImpl) UpdateWorkFlowRunners(wfr []*CdWorkflowRunner) error {
	_, err := impl.dbConnection.Model(&wfr).Update()
	return err
}
func (impl *CdWorkflowRepositoryImpl) FindWorkflowRunnerByCdWorkflowId(wfIds []int) ([]*CdWorkflowRunner, error) {
	var wfr []*CdWorkflowRunner
	err := impl.dbConnection.Model(&wfr).Where("cd_workflow_id in (?)", pg.In(wfIds)).Select()
	if err != nil {
		return nil, err
	}
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) FindWorkflowRunnerById(wfrId int) (*CdWorkflowRunner, error) {
	wfr := &CdWorkflowRunner{}
	err := impl.dbConnection.Model(wfr).Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("cd_workflow_runner.id = ?", wfrId).Select()
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) FindByWorkflowIdAndRunnerType(wfId int, runnerType bean.WorkflowType) (CdWorkflowRunner, error) {
	var wfr CdWorkflowRunner
	err := impl.dbConnection.
		Model(&wfr).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("cd_workflow.id = ?", wfId).
		Where("cd_workflow_runner.workflow_type = ?", runnerType).
		Select()
	if err != nil {
		return wfr, err
	}
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) FindLastStatusByPipelineIdAndRunnerType(pipelineId int, runnerType bean.WorkflowType) (CdWorkflowRunner, error) {
	wfr := CdWorkflowRunner{}
	err := impl.dbConnection.
		Model(&wfr).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("cd_workflow.pipeline_id = ?", pipelineId).
		Where("cd_workflow_runner.workflow_type = ?", runnerType).
		Order("cd_workflow_runner.id DESC").
		Limit(1).
		Select()
	if err != nil {
		return wfr, err
	}
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) IsLatestWf(pipelineId int, wfId int) (bool, error) {
	exists, err := impl.dbConnection.Model(&CdWorkflow{}).
		Where("pipeline_id =?", pipelineId).
		Where("id > ?", wfId).
		Exists()
	return !exists, err
}

func (impl *CdWorkflowRepositoryImpl) FetchAllCdStagesLatestEntity(pipelineIds []int) ([]*CdWorkflowStatus, error) {
	var cdWorkflowStatus []*CdWorkflowStatus
	if len(pipelineIds) == 0 {
		return cdWorkflowStatus, nil
	}
	query := "select p.ci_pipeline_id, wf.pipeline_id, wfr.workflow_type, max(wfr.id) as wfr_id from cd_workflow_runner wfr" +
		" inner join cd_workflow wf on wf.id=wfr.cd_workflow_id" +
		" inner join pipeline p on p.id = wf.pipeline_id" +
		" where wf.pipeline_id in (" + sqlIntSeq(pipelineIds) + ")" +
		" group by p.ci_pipeline_id, wf.pipeline_id, wfr.workflow_type order by wfr_id desc;"
	_, err := impl.dbConnection.Query(&cdWorkflowStatus, query)
	if err != nil {
		impl.logger.Error("err", err)
		return cdWorkflowStatus, err
	}
	return cdWorkflowStatus, nil
}

func (impl *CdWorkflowRepositoryImpl) FetchAllCdStagesLatestEntityStatus(wfrIds []int) ([]*CdWorkflowRunner, error) {
	var wfrList []*CdWorkflowRunner
	err := impl.dbConnection.Model(&wfrList).Column("cd_workflow_runner.*").
		Where("cd_workflow_runner.id in (?)", pg.In(wfrIds)).Select()
	return wfrList, err
}

func (impl *CdWorkflowRepositoryImpl) ExistsByStatus(status string) (bool, error) {
	exists, err := impl.dbConnection.Model(&CdWorkflowRunner{}).
		Where("status =?", status).
		Exists()
	return exists, err
}
