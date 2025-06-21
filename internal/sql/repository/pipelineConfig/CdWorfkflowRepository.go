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
	"errors"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/constants"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"time"
)

type CdWorkflowRepository interface {
	CheckWorkflowRunnerByReferenceId(referenceId string) (bool, error)
	SaveWorkFlow(ctx context.Context, wf *CdWorkflow) error
	UpdateWorkFlow(wf *CdWorkflow) error
	FindById(wfId int) (*CdWorkflow, error)
	FindCdWorkflowMetaByEnvironmentId(appId int, environmentId int, offset int, size int) ([]CdWorkflowRunner, error)
	FindCdWorkflowMetaByPipelineId(pipelineId int, offset int, size int) ([]CdWorkflowRunner, error)
	FindArtifactByPipelineIdAndRunnerType(pipelineId int, runnerType apiBean.WorkflowType, limit int, runnerStatuses []string) ([]CdWorkflowRunner, error)
	SaveWorkFlowRunnerWithTx(wfr *CdWorkflowRunner, tx *pg.Tx) error
	UpdateWorkFlowRunnerWithTx(wfr *CdWorkflowRunner, tx *pg.Tx) error
	UpdateIsArtifactUploaded(wfrId int, isArtifactUploaded workflow.ArtifactUploadedType) error
	GetPreviousQueuedRunners(cdWfrId, pipelineId int) ([]*CdWorkflowRunner, error)
	UpdateRunnerStatusToFailedForIds(errMsg string, triggeredBy int32, cdWfrIds ...int) error
	UpdateWorkFlowRunnersWithTxn(wfrs []*CdWorkflowRunner, tx *pg.Tx) error
	UpdateWorkFlowRunners(wfr []*CdWorkflowRunner) error
	FindWorkflowRunnerByCdWorkflowId(wfIds []int) ([]*CdWorkflowRunner, error)
	FindPreviousCdWfRunnerByStatus(pipelineId int, currentWFRunnerId int, status []string) ([]*CdWorkflowRunner, error)
	FindWorkflowRunnerById(wfrId int) (*CdWorkflowRunner, error)
	FindPreOrPostCdWorkflowRunnerById(wfrId int) (*CdWorkflowRunner, error)
	FindBasicWorkflowRunnerById(wfrId int) (*CdWorkflowRunner, error)
	FindRetriedWorkflowCountByReferenceId(wfrId int) (int, error)
	FindLatestWfrByAppIdAndEnvironmentId(appId int, environmentId int) (*CdWorkflowRunner, error)
	FindLastUnFailedProcessedRunner(appId int, environmentId int) (*CdWorkflowRunner, error)
	IsLatestCDWfr(pipelineId, wfrId int) (bool, error)
	FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId int, environmentId int, runnerType apiBean.WorkflowType) (CdWorkflowRunner, error)
	FindLatestCdWorkflowRunnerArtifactMetadataForAppAndEnvIds(appVsEnvIdMap map[int][]int, runnerType apiBean.WorkflowType) ([]*cdWorkflow.CdWorkflowRunnerArtifactMetadata, error)
	FindAllTriggeredWorkflowCountInLast24Hour() (cdWorkflowCount int, err error)
	GetConnection() *pg.DB

	FindLastPreOrPostTriggeredByPipelineId(pipelineId int) (CdWorkflowRunner, error)
	FindLastPreOrPostTriggeredByEnvironmentId(appId int, environmentId int) (CdWorkflowRunner, error)

	FindByWorkflowIdAndRunnerType(ctx context.Context, wfId int, runnerType apiBean.WorkflowType) (CdWorkflowRunner, error)
	FindLatestByPipelineIdAndRunnerType(pipelineId int, runnerType apiBean.WorkflowType) (CdWorkflowRunner, error)
	SaveWorkFlows(wfs ...*CdWorkflow) error
	IsLatestWf(pipelineId int, wfId int) (bool, error)
	FindLatestCdWorkflowByPipelineId(pipelineIds []int) (*CdWorkflow, error)
	FindLatestCdWorkflowByPipelineIdV2(pipelineIds []int) ([]*CdWorkflow, error)
	FetchAllCdStagesLatestEntity(pipelineIds []int) ([]*CdWorkflowStatus, error)
	FetchAllCdStagesLatestEntityStatus(wfrIds []int) ([]*CdWorkflowRunner, error)
	ExistsByStatus(status string) (bool, error)
	FetchEnvAllCdStagesLatestEntityStatus(wfrIds []int, envID int) ([]*CdWorkflowRunner, error)
	FetchArtifactsByCdPipelineId(pipelineId int, runnerType apiBean.WorkflowType, offset, limit int, searchString string) ([]CdWorkflowRunner, error)
	GetLatestTriggersOfHelmPipelinesStuckInNonTerminalStatuses(getPipelineDeployedWithinHours int) ([]*CdWorkflowRunner, error)
	FindLatestRunnerByPipelineIdsAndRunnerType(ctx context.Context, pipelineIds []int, runnerType apiBean.WorkflowType) ([]CdWorkflowRunner, error)

	MigrateIsArtifactUploaded(wfrId int, isArtifactUploaded bool)
	MigrateCdArtifactLocation(wfrId int, cdArtifactLocation string)
	FindDeployedCdWorkflowRunnersByPipelineId(pipelineId int) ([]*CdWorkflowRunner, error)
}

type CdWorkflowRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

type CdWorkflow struct {
	tableName        struct{}                  `sql:"cd_workflow" pg:",discard_unknown_columns"`
	Id               int                       `sql:"id,pk"`
	CiArtifactId     int                       `sql:"ci_artifact_id"`
	PipelineId       int                       `sql:"pipeline_id"`
	WorkflowStatus   cdWorkflow.WorkflowStatus `sql:"workflow_status,notnull"`
	Pipeline         *Pipeline
	CiArtifact       *repository.CiArtifact
	CdWorkflowRunner []CdWorkflowRunner
	sql.AuditLog
}

type CdWorkflowRunnerWithExtraFields struct {
	CdWorkflowRunner
	TotalCount int
}

type CdWorkflowRunner struct {
	tableName               struct{}                            `sql:"cd_workflow_runner" pg:",discard_unknown_columns"`
	Id                      int                                 `sql:"id,pk"`
	Name                    string                              `sql:"name"`
	WorkflowType            apiBean.WorkflowType                `sql:"workflow_type"` // pre,post,deploy
	ExecutorType            cdWorkflow.WorkflowExecutorType     `sql:"executor_type"` // awf, system
	Status                  string                              `sql:"status"`
	PodStatus               string                              `sql:"pod_status"`
	Message                 string                              `sql:"message"`
	StartedOn               time.Time                           `sql:"started_on"`
	FinishedOn              time.Time                           `sql:"finished_on"`
	Namespace               string                              `sql:"namespace"`
	LogLocation             string                              `sql:"log_file_path"`
	CdArtifactLocation      string                              `sql:"cd_artifact_location"`
	IsArtifactUploaded      workflow.ArtifactUploadedType       `sql:"is_artifact_uploaded"`
	TriggeredBy             int32                               `sql:"triggered_by"`
	CdWorkflowId            int                                 `sql:"cd_workflow_id"`
	PodName                 string                              `sql:"pod_name"`
	BlobStorageEnabled      bool                                `sql:"blob_storage_enabled,notnull"`
	RefCdWorkflowRunnerId   int                                 `sql:"ref_cd_workflow_runner_id,notnull"`
	ImagePathReservationIds []int                               `sql:"image_path_reservation_ids" pg:",array,notnull"`
	ReferenceId             *string                             `sql:"reference_id"`
	ImageState              constants.ImageStateWhileDeployment `sql:"image_state"` // image_state currently not utilized in oss
	CdWorkflow              *CdWorkflow
	sql.AuditLog
}

func (c *CdWorkflowRunner) GetIsArtifactUploaded() (isArtifactUploaded bool, isMigrationRequired bool) {
	return workflow.IsArtifactUploaded(c.IsArtifactUploaded)
}

func (c *CdWorkflowRunner) WithIsArtifactUploaded(isArtifactUploaded bool) *CdWorkflowRunner {
	c.IsArtifactUploaded = workflow.GetArtifactUploadedType(isArtifactUploaded)
	return c
}

func (c *CdWorkflowRunner) IsExternalRun() bool {
	var isExtCluster bool
	if c.WorkflowType == cdWorkflow.WorkflowTypePre {
		isExtCluster = c.CdWorkflow.Pipeline.RunPreStageInEnv
	} else if c.WorkflowType == cdWorkflow.WorkflowTypePost {
		isExtCluster = c.CdWorkflow.Pipeline.RunPostStageInEnv
	}
	return isExtCluster
}

type TriggerWorkflowStatus struct {
	CdWorkflowStatus []*CdWorkflowStatus `json:"cdWorkflowStatus"`
	CiWorkflowStatus []*CiWorkflowStatus `json:"ciWorkflowStatus"`
}

type CdWorkflowStatus struct {
	CiPipelineId               int    `json:"ci_pipeline_id"`
	PipelineId                 int    `json:"pipeline_id"`
	PipelineName               string `json:"pipeline_name,omitempty"`
	DeployStatus               string `json:"deploy_status"`
	PreStatus                  string `json:"pre_status"`
	PostStatus                 string `json:"post_status"`
	WorkflowType               string `json:"workflow_type,omitempty"`
	WfrId                      int    `json:"wfr_id,omitempty"`
	DeploymentAppDeleteRequest bool   `json:"deploymentAppDeleteRequest"`
}

type CiWorkflowStatus struct {
	CiPipelineId      int    `json:"ciPipelineId"`
	CiPipelineName    string `json:"ciPipelineName,omitempty"`
	CiStatus          string `json:"ciStatus"`
	StorageConfigured bool   `json:"storageConfigured"`
	CiWorkflowId      int    `json:"ciWorkflowId,omitempty"`
}

type AppDeploymentStatus struct {
	AppId        int    `json:"appId"`
	PipelineId   int    `json:"pipelineId"`
	DeployStatus string `json:"deployStatus"`
	WfrId        int    `json:"wfrId,omitempty"`
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
		Where("workflow_type = ? ", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
		Where("cd_workflow_runner.status not in (?) ", pg.In(status)).
		Order("cd_workflow_runner.id DESC").
		Select()
	return runner, err
}

func (impl *CdWorkflowRepositoryImpl) SaveWorkFlow(ctx context.Context, wf *CdWorkflow) error {
	_, span := otel.Tracer("orchestrator").Start(ctx, "cdWorkflowRepository.SaveWorkFlow")
	defer span.End()
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

func (impl *CdWorkflowRepositoryImpl) FindLatestCdWorkflowByPipelineId(pipelineIds []int) (*CdWorkflow, error) {
	cdWorkflow := &CdWorkflow{}
	err := impl.dbConnection.Model(cdWorkflow).Where("pipeline_id in (?)", pg.In(pipelineIds)).Order("id DESC").Limit(1).Select()
	return cdWorkflow, err
}

func (impl *CdWorkflowRepositoryImpl) FindLatestCdWorkflowByPipelineIdV2(pipelineIds []int) ([]*CdWorkflow, error) {
	var cdWorkflow []*CdWorkflow
	// err := impl.dbConnection.Model(&cdWorkflow).Where("pipeline_id in (?)", pg.In(pipelineIds)).Order("id DESC").Select()
	query := "SELECT cdw.pipeline_id, cdw.workflow_status, MAX(id) as id from cd_workflow cdw" +
		" WHERE cdw.pipeline_id in(?)" +
		" GROUP by cdw.pipeline_id, cdw.workflow_status ORDER by id desc;"
	_, err := impl.dbConnection.Query(&cdWorkflow, query, pg.In(pipelineIds))
	if err != nil {
		return cdWorkflow, err
	}
	// TODO - Group By Environment And Pipeline will get latest pipeline from top
	return cdWorkflow, err
}
func (impl *CdWorkflowRepositoryImpl) FindAllTriggeredWorkflowCountInLast24Hour() (cdWorkflowCount int, err error) {
	cnt, err := impl.dbConnection.
		Model(&CdWorkflow{}).
		ColumnExpr("DISTINCT pipeline_id").
		Join("JOIN cd_workflow_runner ON cd_workflow.id = cd_workflow_runner.cd_workflow_id").
		Where("cd_workflow_runner.workflow_type = ? AND cd_workflow_runner.started_on > ?", apiBean.CD_WORKFLOW_TYPE_DEPLOY, time.Now().AddDate(0, 0, -1)).
		Group("cd_workflow.pipeline_id").
		Count()
	if err != nil {
		impl.logger.Errorw("error occurred while fetching cd workflow", "err", err)
	}
	return cnt, err
}
func (impl *CdWorkflowRepositoryImpl) FindCdWorkflowMetaByEnvironmentId(appId int, environmentId int, offset int, limit int) ([]CdWorkflowRunner, error) {
	var wfrList []CdWorkflowRunner
	err := impl.dbConnection.
		Model(&wfrList).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("p.environment_id = ?", environmentId).
		Where("p.app_id = ?", appId).
		Where("p.deleted = ?", false).
		Order("cd_workflow_runner.id DESC").
		Join("inner join cd_workflow wf on wf.id = cd_workflow_runner.cd_workflow_id").
		Join("inner join ci_artifact cia on cia.id = wf.ci_artifact_id").
		Join("inner join pipeline p on p.id = wf.pipeline_id").
		// Join("left join users u on u.id = wfr.triggered_by").
		Offset(offset).Limit(limit).
		Select()
	if err != nil {
		return nil, err
	}
	return wfrList, err
}

func (impl *CdWorkflowRepositoryImpl) FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId int, environmentId int, runnerType apiBean.WorkflowType) (CdWorkflowRunner, error) {
	var wfr CdWorkflowRunner
	err := impl.dbConnection.
		Model(&wfr).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("p.environment_id = ?", environmentId).
		Where("p.app_id = ?", appId).
		Where("cd_workflow_runner.workflow_type = ?", runnerType).
		Join("inner join cd_workflow wf on wf.id = cd_workflow_runner.cd_workflow_id").
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
		// Join("inner join cd_workflow wf on wf.id = cd_workflow_runner.cd_workflow_id").
		// Join("inner join ci_artifact cia on cia.id = wf.ci_artifact_id").
		// Join("inner join pipeline p on p.id = wf.pipeline_id").
		// Join("left join users u on u.id = wfr.triggered_by").
		// Order("ORDER BY cd_workflow_runner.started_on DESC").
		Offset(offset).Limit(limit).
		Select()

	if err != nil {
		return nil, err
	}
	return wfrList, err
}

func (impl *CdWorkflowRepositoryImpl) FindArtifactByPipelineIdAndRunnerType(pipelineId int, runnerType apiBean.WorkflowType, limit int, runnerStatuses []string) ([]CdWorkflowRunner, error) {
	var wfrList []CdWorkflowRunner
	query := impl.dbConnection.
		Model(&wfrList).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("cd_workflow.pipeline_id = ?", pipelineId).
		Where("cd_workflow_runner.workflow_type = ?", runnerType)
	if len(runnerStatuses) > 0 {
		query.Where("cd_workflow_runner.status IN (?)", pg.In(runnerStatuses))
	}
	err := query.
		Order("cd_workflow_runner.id DESC").
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
		Where("cd_workflow_runner.workflow_type != ?", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
		Order("cd_workflow_runner.id DESC").
		Limit(1).
		Select()
	if err != nil {
		return wfr, err
	}
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) FindLatestWfrByAppIdAndEnvironmentId(appId int, environmentId int) (*CdWorkflowRunner, error) {
	wfr := &CdWorkflowRunner{}
	err := impl.dbConnection.
		Model(wfr).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("p.environment_id = ?", environmentId).
		Where("p.app_id = ?", appId).
		Where("cd_workflow_runner.workflow_type = ?", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
		Order("cd_workflow_runner.id DESC").
		Join("inner join cd_workflow wf on wf.id = cd_workflow_runner.cd_workflow_id").
		Join("inner join ci_artifact cia on cia.id = wf.ci_artifact_id").
		Join("inner join pipeline p on p.id = wf.pipeline_id").
		Limit(1).
		Select()
	if err != nil {
		return wfr, err
	}
	return wfr, nil
}

func (impl *CdWorkflowRepositoryImpl) FindLastUnFailedProcessedRunner(appId int, environmentId int) (*CdWorkflowRunner, error) {
	wfr := &CdWorkflowRunner{}
	err := impl.dbConnection.
		Model(wfr).
		Column("cd_workflow_runner.*", "CdWorkflow.Pipeline.id", "CdWorkflow.Pipeline.deployment_app_delete_request").
		Where("p.environment_id = ?", environmentId).
		Where("p.app_id = ?", appId).
		Where("cd_workflow_runner.workflow_type = ?", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
		Where("cd_workflow_runner.status NOT IN (?)", pg.In([]string{cdWorkflow.WorkflowInitiated, cdWorkflow.WorkflowInQueue, cdWorkflow.WorkflowFailed})).
		Order("cd_workflow_runner.id DESC").
		Join("inner join cd_workflow wf on wf.id = cd_workflow_runner.cd_workflow_id").
		Join("inner join pipeline p on p.id = wf.pipeline_id").
		Limit(1).
		Select()
	if err != nil {
		return wfr, err
	}
	return wfr, nil

}

func (impl *CdWorkflowRepositoryImpl) IsLatestCDWfr(pipelineId, wfrId int) (bool, error) {
	wfr := &CdWorkflowRunner{}
	ifAnySuccessorWfrExists, err := impl.dbConnection.
		Model(wfr).
		Column("cd_workflow_runner.*", "CdWorkflow").
		Where("wf.pipeline_id = ?", pipelineId).
		Where("cd_workflow_runner.workflow_type = ?", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
		Order("cd_workflow_runner.id DESC").
		Join("inner join cd_workflow wf on wf.id = cd_workflow_runner.cd_workflow_id").
		Where("cd_workflow_runner.id > ?", wfrId).
		Exists()
	return !ifAnySuccessorWfrExists, err
}

func (impl *CdWorkflowRepositoryImpl) FindLastPreOrPostTriggeredByEnvironmentId(appId int, environmentId int) (CdWorkflowRunner, error) {
	wfr := CdWorkflowRunner{}
	err := impl.dbConnection.
		Model(&wfr).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("p.environment_id = ?", environmentId).
		Where("p.app_id = ?", appId).
		Where("cd_workflow_runner.workflow_type != ?", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
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

func (impl *CdWorkflowRepositoryImpl) SaveWorkFlowRunnerWithTx(wfr *CdWorkflowRunner, tx *pg.Tx) error {
	if tx == nil {
		return impl.dbConnection.Insert(wfr)
	}
	return tx.Insert(wfr)
}

func (impl *CdWorkflowRepositoryImpl) UpdateWorkFlowRunnerWithTx(wfr *CdWorkflowRunner, tx *pg.Tx) error {
	err := tx.Update(wfr)
	return err
}

func (impl *CdWorkflowRepositoryImpl) UpdateIsArtifactUploaded(wfrId int, isArtifactUploaded workflow.ArtifactUploadedType) error {
	_, err := impl.dbConnection.Model((*CdWorkflowRunner)(nil)).
		Set("is_artifact_uploaded = ?", isArtifactUploaded).
		Where("id = ?", wfrId).
		Update()
	return err
}

func (impl *CdWorkflowRepositoryImpl) GetPreviousQueuedRunners(cdWfrId, pipelineId int) ([]*CdWorkflowRunner, error) {
	var cdWfrs []*CdWorkflowRunner
	err := impl.dbConnection.Model(&cdWfrs).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline").
		Where("workflow_type = ?", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
		Where("cd_workflow.pipeline_id = ?", pipelineId).
		Where("cd_workflow_runner.id < ?", cdWfrId).
		Where("cd_workflow_runner.status = ?", cdWorkflow.WorkflowInQueue).
		Select()
	return cdWfrs, err
}

func (impl *CdWorkflowRepositoryImpl) UpdateRunnerStatusToFailedForIds(errMsg string, triggeredBy int32, cdWfrIds ...int) error {
	if len(cdWfrIds) == 0 {
		return nil
	}
	_, err := impl.dbConnection.Model((*CdWorkflowRunner)(nil)).
		Set("status = ?", cdWorkflow.WorkflowFailed).
		Set("finished_on = ?", time.Now()).
		Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", triggeredBy).
		Set("message = ?", errMsg).
		Where("id IN (?)", pg.In(cdWfrIds)).
		Update()
	return err
}

func (impl *CdWorkflowRepositoryImpl) UpdateWorkFlowRunnersWithTxn(wfrs []*CdWorkflowRunner, tx *pg.Tx) error {
	_, err := tx.Model(&wfrs).Update()
	return err
}

func (impl *CdWorkflowRepositoryImpl) UpdateWorkFlowRunners(wfrs []*CdWorkflowRunner) error {
	for _, wfr := range wfrs {
		err := impl.dbConnection.Update(wfr)
		if err != nil {
			impl.logger.Errorw("error in updating wfr", "err", err)
			return err
		}
	}
	return nil
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
	err := impl.dbConnection.Model(wfr).Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact", "CdWorkflow.Pipeline.Environment").
		Where("cd_workflow_runner.id = ?", wfrId).Select()
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) FindPreOrPostCdWorkflowRunnerById(wfrId int) (*CdWorkflowRunner, error) {
	wfr := &CdWorkflowRunner{}
	err := impl.dbConnection.Model(wfr).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact", "CdWorkflow.Pipeline.Environment").
		Where("cd_workflow_runner.id = ?", wfrId).
		Where("cd_workflow_runner.workflow_type != ?", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
		Select()
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) FindBasicWorkflowRunnerById(wfrId int) (*CdWorkflowRunner, error) {
	wfr := &CdWorkflowRunner{}
	err := impl.dbConnection.Model(wfr).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline").
		Where("cd_workflow_runner.id = ?", wfrId).Select()
	return wfr, err
}

func (impl *CdWorkflowRepositoryImpl) FindRetriedWorkflowCountByReferenceId(wfrId int) (int, error) {
	retryCount := 0
	query := "select count(id) from cd_workflow_runner where ref_cd_workflow_runner_id = ?"

	_, err := impl.dbConnection.Query(&retryCount, query, wfrId)
	return retryCount, err
}

func (impl *CdWorkflowRepositoryImpl) FindByWorkflowIdAndRunnerType(ctx context.Context, wfId int, runnerType apiBean.WorkflowType) (CdWorkflowRunner, error) {
	var wfr CdWorkflowRunner
	_, span := otel.Tracer("orchestrator").Start(ctx, "cdWorkflowRepository.FindByWorkflowIdAndRunnerType")
	defer span.End()
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

func (impl *CdWorkflowRepositoryImpl) FindLatestByPipelineIdAndRunnerType(pipelineId int, runnerType apiBean.WorkflowType) (CdWorkflowRunner, error) {
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
	err := impl.dbConnection.Model(&wfrList).
		Column("cd_workflow_runner.id", "cd_workflow_runner.status").
		Where("cd_workflow_runner.id in (?)", pg.In(wfrIds)).
		Select()
	return wfrList, err
}

func (impl *CdWorkflowRepositoryImpl) FetchEnvAllCdStagesLatestEntityStatus(wfrIds []int, envID int) ([]*CdWorkflowRunner, error) {
	var wfrList []*CdWorkflowRunner
	query := `
		select wfr.id, wfr.status 
		from cd_workflow_runner wfr
		inner join cd_workflow wf on wf.id = wfr.cd_workflow_id
		inner join pipeline p on p.id = wf.pipeline_id
		where p.environment_id = ? and wfr.id in (?)
	`
	_, err := impl.dbConnection.Query(&wfrList, query, envID, pg.In(wfrIds))
	return wfrList, err
}

func (impl *CdWorkflowRepositoryImpl) ExistsByStatus(status string) (bool, error) {
	exists, err := impl.dbConnection.Model(&CdWorkflowRunner{}).
		Where("status =?", status).
		Exists()
	return exists, err
}

func (impl *CdWorkflowRepositoryImpl) FetchArtifactsByCdPipelineId(pipelineId int, runnerType apiBean.WorkflowType, offset, limit int, searchString string) ([]CdWorkflowRunner, error) {
	var wfrList []CdWorkflowRunner
	searchStringFinal := "%" + searchString + "%"
	err := impl.dbConnection.
		Model(&wfrList).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("cd_workflow.pipeline_id = ?", pipelineId).
		Where("cd_workflow_runner.workflow_type = ?", runnerType).
		Where("cd_workflow__ci_artifact.image LIKE ?", searchStringFinal).
		Order("cd_workflow_runner.id DESC").
		Limit(limit).Offset(offset).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting Wfrs and ci artifacts by pipelineId", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	return wfrList, err
}

func (impl *CdWorkflowRepositoryImpl) FetchArtifactsByCdPipelineIdV2(listingFilterOptions apiBean.ArtifactsListFilterOptions) ([]CdWorkflowRunner, int, error) {
	var wfrList []CdWorkflowRunner
	query := impl.dbConnection.
		Model(&wfrList).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline", "CdWorkflow.CiArtifact").
		Where("cd_workflow.pipeline_id = ?", listingFilterOptions.PipelineId).
		Where("cd_workflow_runner.workflow_type = ?", listingFilterOptions.StageType).
		Where("cd_workflow__ci_artifact.image LIKE ?", listingFilterOptions.SearchString)

	if len(listingFilterOptions.ExcludeArtifactIds) > 0 {
		query = query.Where("cd_workflow__ci_artifact.id NOT IN (?)", pg.In(listingFilterOptions.ExcludeArtifactIds))
	}
	totalCount, err := query.Count()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting Wfrs count and ci artifacts by pipelineId", "err", err, "pipelineId", listingFilterOptions.PipelineId)
		return nil, totalCount, err
	}

	query = query.Order("cd_workflow_runner.id DESC").
		Limit(listingFilterOptions.Limit).
		Offset(listingFilterOptions.Offset)

	err = query.Select()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting Wfrs and ci artifacts by pipelineId", "err", err, "pipelineId", listingFilterOptions.PipelineId)
		return nil, totalCount, err
	}
	return wfrList, totalCount, nil
}

func (impl *CdWorkflowRepositoryImpl) GetLatestTriggersOfHelmPipelinesStuckInNonTerminalStatuses(getPipelineDeployedWithinHours int) ([]*CdWorkflowRunner, error) {
	var wfrList []*CdWorkflowRunner
	excludedStatusList := cdWorkflow.WfrTerminalStatusList
	excludedStatusList = append(excludedStatusList, cdWorkflow.WorkflowInitiated, cdWorkflow.WorkflowInQueue, cdWorkflow.WorkflowStarting)
	err := impl.dbConnection.
		Model(&wfrList).
		Column("cd_workflow_runner.*", "CdWorkflow.id", "CdWorkflow.pipeline_id", "CdWorkflow.Pipeline.id", "CdWorkflow.Pipeline.app_id", "CdWorkflow.Pipeline.environment_id", "CdWorkflow.Pipeline.deployment_app_name", "CdWorkflow.Pipeline.deleted", "CdWorkflow.Pipeline.Environment").
		Join("LEFT JOIN deployment_config dc on dc.active=true and dc.app_id = cd_workflow__pipeline.app_id and dc.environment_id=cd_workflow__pipeline.environment_id").
		Where("cd_workflow_runner.workflow_type=?", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
		Where("cd_workflow_runner.status not in (?)", pg.In(excludedStatusList)).
		Where("cd_workflow_runner.cd_workflow_id in"+
			" (SELECT max(cd_workflow.id) as id from cd_workflow"+
			" INNER JOIN cd_workflow_runner on cd_workflow.id = cd_workflow_runner.cd_workflow_id"+
			" WHERE cd_workflow_runner.workflow_type = ? "+
			" AND cd_workflow_runner.status != ?"+
			" GROUP BY cd_workflow.pipeline_id"+
			" ORDER BY cd_workflow.pipeline_id desc)", apiBean.CD_WORKFLOW_TYPE_DEPLOY, cdWorkflow.WorkflowInQueue).
		Where("(cd_workflow__pipeline.deployment_app_type=? or dc.deployment_app_type=?)", util.PIPELINE_DEPLOYMENT_TYPE_HELM, util.PIPELINE_DEPLOYMENT_TYPE_HELM).
		Where("cd_workflow_runner.started_on > NOW() - INTERVAL '? hours'", getPipelineDeployedWithinHours).
		Where("cd_workflow__pipeline.deleted=?", false).
		Order("cd_workflow_runner.id DESC").
		Select()
	if err != nil {
		impl.logger.Errorw("error,GetLatestTriggersOfHelmPipelinesStuckInNonTerminalStatuses ", "err", err)
		return nil, err
	}
	return wfrList, err
}

func (impl *CdWorkflowRepositoryImpl) CheckWorkflowRunnerByReferenceId(referenceId string) (bool, error) {
	exists, err := impl.dbConnection.Model((*CdWorkflowRunner)(nil)).
		Where("cd_workflow_runner.reference_id = ?", referenceId).
		Exists()
	if errors.Is(err, pg.ErrNoRows) {
		return false, nil
	}
	return exists, err
}

func (impl *CdWorkflowRepositoryImpl) FindLatestRunnerByPipelineIdsAndRunnerType(ctx context.Context, pipelineIds []int, runnerType apiBean.WorkflowType) ([]CdWorkflowRunner, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "FindLatestRunnerByPipelineIdsAndRunnerType")
	defer span.End()
	if pipelineIds == nil || len(pipelineIds) == 0 {
		return nil, pg.ErrNoRows
	}
	var latestWfrs []CdWorkflowRunner
	err := impl.dbConnection.
		Model(&latestWfrs).
		Column("cd_workflow_runner.*", "CdWorkflow", "CdWorkflow.Pipeline").
		ColumnExpr("MAX(cd_workflow_runner.id)").
		Where("cd_workflow.pipeline_id IN (?)", pg.In(pipelineIds)).
		Where("cd_workflow_runner.workflow_type = ?", runnerType).
		Where("cd_workflow__pipeline.deleted = ?", false).
		Group("cd_workflow_runner.id", "cd_workflow.id", "cd_workflow__pipeline.id").
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting cdWfr by appId, envId and runner type", "pipelineIds", pipelineIds, "runnerType", runnerType)
		return nil, err
	}
	return latestWfrs, err
}

func (impl *CdWorkflowRepositoryImpl) MigrateIsArtifactUploaded(wfrId int, isArtifactUploaded bool) {
	_, err := impl.dbConnection.Model((*CdWorkflowRunner)(nil)).
		Set("is_artifact_uploaded = ?", workflow.GetArtifactUploadedType(isArtifactUploaded)).
		Where("id = ?", wfrId).
		Update()
	if err != nil {
		impl.logger.Errorw("error in updating is artifact uploaded", "wfrId", wfrId, "err", err)
	}
}

func (impl *CdWorkflowRepositoryImpl) MigrateCdArtifactLocation(wfrId int, cdArtifactLocation string) {
	_, err := impl.dbConnection.Model((*CdWorkflowRunner)(nil)).
		Set("cd_artifact_location = ?", cdArtifactLocation).
		Where("id = ?", wfrId).
		Update()
	if err != nil {
		impl.logger.Errorw("error in updating cd artifact location", "wfrId", wfrId, "err", err)
	}
}

func (impl *CdWorkflowRepositoryImpl) FindDeployedCdWorkflowRunnersByPipelineId(pipelineId int) ([]*CdWorkflowRunner, error) {
	var runners []*CdWorkflowRunner
	err := impl.dbConnection.
		Model(&runners).
		Column("cd_workflow_runner.*", "CdWorkflow").
		Where("cd_workflow.pipeline_id = ?", pipelineId).
		Where("workflow_type = ? ", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
		Order("cd_workflow_runner.id").
		Select()
	if err != nil {
		impl.logger.Errorw("error in finding previous co workflow runners by pipeline id ", "pipelineId", pipelineId, "err", err)
		return nil, err
	}
	return runners, nil
}

func (impl *CdWorkflowRepositoryImpl) FindLatestCdWorkflowRunnerArtifactMetadataForAppAndEnvIds(appVsEnvIdMap map[int][]int, runnerType apiBean.WorkflowType) ([]*cdWorkflow.CdWorkflowRunnerArtifactMetadata, error) {
	var runners []*cdWorkflow.CdWorkflowRunnerArtifactMetadata

	// Prepare the (app_id, env_id) tuple list for the query
	tupleList := make([]interface{}, 0, len(appVsEnvIdMap))
	for appId, envIds := range appVsEnvIdMap {
		for _, envId := range envIds {
			tupleList = append(tupleList, []interface{}{appId, envId})
		}
	}
	if len(tupleList) == 0 {
		return nil, nil
	}

	query := `
WITH RankedData AS (
    SELECT 
        p.app_id AS "app_id",
        p.environment_id AS "env_id",
        p.deleted AS "deleted",
        wf.ci_artifact_id AS "ci_artifact_id",
        ci_artifact.parent_ci_artifact AS "parent_ci_artifact",
        ci_artifact.scanned AS "scanned",
        ROW_NUMBER() OVER (PARTITION BY p.app_id, p.environment_id ORDER BY cd_workflow_runner.id DESC) AS rn
    FROM cd_workflow_runner INNER JOIN cd_workflow wf ON wf.id = cd_workflow_runner.cd_workflow_id
    INNER JOIN pipeline p ON p.id = wf.pipeline_id
    INNER JOIN ci_artifact ON ci_artifact.id = wf.ci_artifact_id
    WHERE cd_workflow_runner.workflow_type = ? 
      AND (p.app_id, p.environment_id) IN ( ? )
)
SELECT "app_id","env_id","ci_artifact_id","parent_ci_artifact","scanned" FROM RankedData WHERE rn = 1 and deleted= false;
`
	_, err := impl.dbConnection.Query(&runners, query, runnerType, pg.In(tupleList))
	if err != nil {
		impl.logger.Errorw("error in getting cdWfrs by appId and envIds and runner type", "appVsEnvIdMap", appVsEnvIdMap, "err", err)
		return nil, err
	}
	return runners, nil
}
