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
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/cdWorkflow"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type CiWorkflowRepository interface {
	SaveWorkFlow(wf *CiWorkflow) error
	FindLastTriggeredWorkflow(pipelineId int) (*CiWorkflow, error)
	UpdateWorkFlow(wf *CiWorkflow) error
	UpdateArtifactUploaded(id int, isUploaded bool) error
	FindByStatusesIn(activeStatuses []string) ([]*CiWorkflow, error)
	FindByPipelineId(pipelineId int, offset int, size int) ([]WorkflowWithArtifact, error)
	FindById(id int) (*CiWorkflow, error)
	FindRetriedWorkflowCountByReferenceId(id int) (int, error)
	FindCiWorkflowGitTriggersById(id int) (workflow *CiWorkflow, err error)
	FindCiWorkflowGitTriggersByIds(ids []int) ([]*CiWorkflow, error)
	FindByName(name string) (*CiWorkflow, error)

	FindLastTriggeredWorkflowByCiIds(pipelineId []int) (ciWorkflow []*CiWorkflow, err error)
	FindLastTriggeredWorkflowByArtifactId(ciArtifactId int) (ciWorkflow *CiWorkflow, err error)
	FindAllLastTriggeredWorkflowByArtifactId(ciArtifactId []int) (ciWorkflow []*CiWorkflow, err error)
	FindAllTriggeredWorkflowCountInLast24Hour() (ciWorkflowCount int, err error)
	FindLastTriggeredWorkflowGitTriggersByArtifactId(ciArtifactId int) (ciWorkflow *CiWorkflow, err error)
	FindLastTriggeredWorkflowGitTriggersByArtifactIds(ciArtifactIds []int) ([]*WorkflowWithArtifact, error)
	ExistsByStatus(status string) (bool, error)
	FindBuildTypeAndStatusDataOfLast1Day() []*BuildTypeCount
	FIndCiWorkflowStatusesByAppId(appId int) ([]*CiWorkflowStatus, error)
}

type CiWorkflowRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

type CiWorkflow struct {
	tableName               struct{}                        `sql:"ci_workflow" pg:",discard_unknown_columns"`
	Id                      int                             `sql:"id,pk"`
	Name                    string                          `sql:"name"`
	Status                  string                          `sql:"status"`
	PodStatus               string                          `sql:"pod_status"`
	Message                 string                          `sql:"message"`
	StartedOn               time.Time                       `sql:"started_on"`
	FinishedOn              time.Time                       `sql:"finished_on"`
	CiPipelineId            int                             `sql:"ci_pipeline_id"`
	Namespace               string                          `sql:"namespace"`
	BlobStorageEnabled      bool                            `sql:"blob_storage_enabled,notnull"`
	LogLocation             string                          `sql:"log_file_path"`
	GitTriggers             map[int]GitCommit               `sql:"git_triggers"`
	TriggeredBy             int32                           `sql:"triggered_by"`
	CiArtifactLocation      string                          `sql:"ci_artifact_location"`
	IsArtifactUploaded      *bool                           `sql:"is_artifact_uploaded"`
	PodName                 string                          `sql:"pod_name"`
	CiBuildType             string                          `sql:"ci_build_type"`
	EnvironmentId           int                             `sql:"environment_id"`
	ReferenceCiWorkflowId   int                             `sql:"ref_ci_workflow_id"`
	ParentCiWorkFlowId      int                             `sql:"parent_ci_workflow_id"`
	ExecutorType            cdWorkflow.WorkflowExecutorType `sql:"executor_type"` //awf, system
	ImagePathReservationId  int                             `sql:"image_path_reservation_id"`
	ImagePathReservationIds []int                           `sql:"image_path_reservation_ids" pg:",array"`
	CiPipeline              *CiPipeline
}

func (ciWorkflow *CiWorkflow) InProgress() bool {
	return ciWorkflow.Status == "Running" || ciWorkflow.Status == "Starting"
}

func (ciWorkflow *CiWorkflow) IsExternalRunInJobType() bool {
	return ciWorkflow.EnvironmentId != 0
}

type WorkflowWithArtifact struct {
	Id                      int                             `psql:"id"`
	Name                    string                          `psql:"name"`
	PodName                 string                          `psql:"pod_name"`
	Status                  string                          `psql:"status"`
	PodStatus               string                          `psql:"pod_status"`
	Message                 string                          `psql:"message"`
	StartedOn               time.Time                       `psql:"started_on"`
	FinishedOn              time.Time                       `psql:"finished_on"`
	CiPipelineId            int                             `psql:"ci_pipeline_id"`
	Namespace               string                          `psql:"namespace"`
	LogFilePath             string                          `psql:"log_file_path"`
	GitTriggers             map[int]GitCommit               `psql:"git_triggers"`
	TriggeredBy             int32                           `psql:"triggered_by"`
	EmailId                 string                          `psql:"email_id"`
	Image                   string                          `psql:"image"`
	CiArtifactLocation      string                          `psql:"ci_artifact_location"`
	CiArtifactId            int                             `psql:"ci_artifact_id"`
	BlobStorageEnabled      bool                            `psql:"blob_storage_enabled"`
	CiBuildType             string                          `psql:"ci_build_type"`
	IsArtifactUploadedV2    *bool                           `psql:"is_artifact_uploaded"`     // IsArtifactUploadedV2 is the new column from ci_workflow table, IsArtifactUploaded is Deprecated and will be removed in future
	IsArtifactUploaded      bool                            `psql:"old_is_artifact_uploaded"` // Deprecated; Use IsArtifactUploadedV2 instead. IsArtifactUploaded is the column from ci_artifact table
	EnvironmentId           int                             `psql:"environment_id"`
	EnvironmentName         string                          `psql:"environment_name"`
	RefCiWorkflowId         int                             `psql:"ref_ci_workflow_id"`
	ParentCiWorkflowId      int                             `psql:"parent_ci_workflow_id"`
	ExecutorType            cdWorkflow.WorkflowExecutorType `psql:"executor_type"` //awf, system
	ImagePathReservationId  int                             `psql:"image_path_reservation_id"`
	ImagePathReservationIds []int                           `psql:"image_path_reservation_ids" pg:",array"`
}

type GitCommit struct {
	Commit                 string //git hash
	Author                 string
	Date                   time.Time
	Message                string
	Changes                []string
	WebhookData            WebhookData
	CiConfigureSourceValue string
	GitRepoUrl             string
	GitRepoName            string
	CiConfigureSourceType  SourceType
}

type WebhookData struct {
	Id              int               `json:"id"`
	EventActionType string            `json:"eventActionType"`
	Data            map[string]string `json:"data"`
}

func NewCiWorkflowRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CiWorkflowRepositoryImpl {
	return &CiWorkflowRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl *CiWorkflowRepositoryImpl) FindLastTriggeredWorkflow(pipelineId int) (ciWorkflow *CiWorkflow, err error) {
	workflow := &CiWorkflow{}
	err = impl.dbConnection.Model(workflow).
		Column("ci_workflow.*", "CiPipeline").
		Where("ci_workflow.ci_pipeline_id = ? ", pipelineId).
		Order("ci_workflow.started_on Desc").
		Limit(1).
		Select()
	return workflow, err
}

func (impl *CiWorkflowRepositoryImpl) FindByStatusesIn(activeStatuses []string) ([]*CiWorkflow, error) {
	var ciWorkFlows []*CiWorkflow
	err := impl.dbConnection.Model(&ciWorkFlows).
		Column("ci_workflow.*").
		Where("ci_workflow.status in (?)", pg.In(activeStatuses)).
		Select()
	return ciWorkFlows, err
}

// FindByPipelineId gets only those workflowWithArtifact whose parent_ci_workflow_id is null, this is done to accommodate multiple ci_artifacts through a single workflow(parent), making child workflows for other ci_artifacts (this has been done due to design understanding and db constraint) single workflow single ci-artifact
func (impl *CiWorkflowRepositoryImpl) FindByPipelineId(pipelineId int, offset int, limit int) ([]WorkflowWithArtifact, error) {
	var wfs []WorkflowWithArtifact
	queryTemp := "select cia.id as ci_artifact_id, env.environment_name, cia.image, cia.is_artifact_uploaded as old_is_artifact_uploaded, wf.*, u.email_id from ci_workflow wf left join users u on u.id = wf.triggered_by left join ci_artifact cia on wf.id = cia.ci_workflow_id left join environment env on env.id = wf.environment_id where wf.ci_pipeline_id = ? and parent_ci_workflow_id is null order by wf.started_on desc offset ? limit ?;"
	_, err := impl.dbConnection.Query(&wfs, queryTemp, pipelineId, offset, limit)
	if err != nil {
		return nil, err
	}
	return wfs, err
}

func (impl *CiWorkflowRepositoryImpl) FindByName(name string) (*CiWorkflow, error) {
	var ciWorkFlow *CiWorkflow
	err := impl.dbConnection.Model(&ciWorkFlow).
		Column("ci_workflow.*").
		Where("ci_workflow.name = ?", name).
		Select()
	return ciWorkFlow, err
}

func (impl *CiWorkflowRepositoryImpl) FindById(id int) (*CiWorkflow, error) {
	workflow := &CiWorkflow{}
	err := impl.dbConnection.Model(workflow).
		Column("ci_workflow.*", "CiPipeline", "CiPipeline.App", "CiPipeline.CiTemplate", "CiPipeline.CiTemplate.DockerRegistry").
		Where("ci_workflow.id = ? ", id).
		Select()
	return workflow, err
}

func (impl *CiWorkflowRepositoryImpl) FindRetriedWorkflowCountByReferenceId(id int) (int, error) {
	retryCount := 0
	query := fmt.Sprintf("select count(*) "+
		"from ci_workflow where ref_ci_workflow_id = %v", id)

	_, err := impl.dbConnection.Query(&retryCount, query)
	return retryCount, err
}

func (impl *CiWorkflowRepositoryImpl) FindCiWorkflowGitTriggersById(id int) (ciWorkflow *CiWorkflow, err error) {
	workflow := &CiWorkflow{}
	err = impl.dbConnection.Model(workflow).
		Column("ci_workflow.git_triggers").
		Where("ci_workflow.id = ? ", id).
		Select()

	return workflow, err
}

func (impl *CiWorkflowRepositoryImpl) FindCiWorkflowGitTriggersByIds(ids []int) ([]*CiWorkflow, error) {
	workflows := make([]*CiWorkflow, 0)
	if len(ids) == 0 {
		return workflows, nil
	}
	err := impl.dbConnection.Model(&workflows).
		Column("ci_workflow.git_triggers").
		Where("ci_workflow.id IN (?)", pg.In(ids)).
		Select()

	return workflows, err
}

func (impl *CiWorkflowRepositoryImpl) SaveWorkFlow(wf *CiWorkflow) error {
	err := impl.dbConnection.Insert(wf)
	return err
}

func (impl *CiWorkflowRepositoryImpl) UpdateWorkFlow(wf *CiWorkflow) error {
	err := impl.dbConnection.Update(wf)
	return err
}

func (impl *CiWorkflowRepositoryImpl) UpdateArtifactUploaded(id int, isUploaded bool) error {
	_, err := impl.dbConnection.Model(&CiWorkflow{}).
		Set("is_artifact_uploaded = ?", isUploaded).
		Where("id = ?", id).
		Update()
	return err
}

func (impl *CiWorkflowRepositoryImpl) FindLastTriggeredWorkflowByCiIds(pipelineId []int) (ciWorkflow []*CiWorkflow, err error) {
	err = impl.dbConnection.Model(&ciWorkflow).
		Column("ci_workflow.*", "CiPipeline").
		Where("ci_workflow.ci_pipeline_id in (?) ", pg.In(pipelineId)).
		Order("ci_workflow.started_on Desc").
		Select()
	return ciWorkflow, err
}

func (impl *CiWorkflowRepositoryImpl) FindLastTriggeredWorkflowByArtifactId(ciArtifactId int) (ciWorkflow *CiWorkflow, err error) {
	workflow := &CiWorkflow{}
	err = impl.dbConnection.Model(workflow).
		Column("ci_workflow.*", "CiPipeline").
		Join("inner join ci_artifact cia on cia.ci_workflow_id = ci_workflow.id").
		Where("cia.id = ? ", ciArtifactId).
		Select()
	return workflow, err
}
func (impl *CiWorkflowRepositoryImpl) FindAllTriggeredWorkflowCountInLast24Hour() (ciWorkflowCount int, err error) {
	cnt, err := impl.dbConnection.Model(&CiWorkflow{}).
		ColumnExpr("DISTINCT ci_pipeline_id").
		Where("started_on > ? ", time.Now().AddDate(0, 0, -1)).
		Count()
	if err != nil {
		impl.logger.Errorw("error occurred while fetching ci workflow", "err", err)
	}
	return cnt, err
}
func (impl *CiWorkflowRepositoryImpl) FindAllLastTriggeredWorkflowByArtifactId(ciArtifactIds []int) (ciWorkflows []*CiWorkflow, err error) {
	err = impl.dbConnection.Model(&ciWorkflows).
		Column("ci_workflow.git_triggers", "ci_workflow.ci_pipeline_id", "CiPipeline", "cia.id").
		Join("inner join ci_artifact cia on cia.ci_workflow_id = ci_workflow.id").
		Where("cia.id in (?) ", pg.In(ciArtifactIds)).
		Select()
	return ciWorkflows, err
}

func (impl *CiWorkflowRepositoryImpl) FindLastTriggeredWorkflowGitTriggersByArtifactId(ciArtifactId int) (ciWorkflow *CiWorkflow, err error) {
	workflow := &CiWorkflow{}
	err = impl.dbConnection.Model(workflow).
		Column("ci_workflow.git_triggers").
		Join("inner join ci_artifact cia on cia.ci_workflow_id = ci_workflow.id").
		Where("cia.id = ? ", ciArtifactId).
		Select()

	return workflow, err
}

func (impl *CiWorkflowRepositoryImpl) FindLastTriggeredWorkflowGitTriggersByArtifactIds(ciArtifactIds []int) ([]*WorkflowWithArtifact, error) {
	workflows := make([]*WorkflowWithArtifact, 0)
	if len(ciArtifactIds) == 0 {
		return workflows, nil
	}
	query := "SELECT cw.git_triggers,cw.id,cw.triggered_by,cw.ci_pipeline_id,cia.id as ci_artifact_id" +
		" FROM ci_workflow cw INNER JOIN ci_artifact cia on cia.ci_workflow_id = cw.id " +
		" WHERE cia.id IN (%s)"
	query = fmt.Sprintf(query, helper.GetCommaSepratedString(ciArtifactIds))
	_, err := impl.dbConnection.Query(&workflows, query)
	return workflows, err
}

func (impl *CiWorkflowRepositoryImpl) ExistsByStatus(status string) (bool, error) {
	exists, err := impl.dbConnection.Model(&CiWorkflow{}).
		Where("status =?", status).
		Exists()
	return exists, err
}

func (impl *CiWorkflowRepositoryImpl) FindBuildTypeAndStatusDataOfLast1Day() []*BuildTypeCount {
	var buildTypeCounts []*BuildTypeCount
	query := "select status,ci_build_type as type, count(*) from ci_workflow where status in ('Succeeded','Failed') and started_on > ? group by (ci_build_type, status)"
	_, err := impl.dbConnection.Query(&buildTypeCounts, query, time.Now().AddDate(0, 0, -1))
	if err != nil {
		impl.logger.Errorw("error occurred while fetching build type vs status vs count data", "err", err)
	}
	return buildTypeCounts
}

func (impl *CiWorkflowRepositoryImpl) FIndCiWorkflowStatusesByAppId(appId int) ([]*CiWorkflowStatus, error) {

	ciworkflowStatuses := make([]*CiWorkflowStatus, 0)

	query := "SELECT cw1.ci_pipeline_id,cw1.status AS ci_status,cw1.blob_storage_enabled AS storage_configured " +
		" FROM ci_workflow cw1 " +
		" INNER JOIN " +
		" (WITH cp AS (SELECT id, parent_ci_pipeline FROM ci_pipeline WHERE app_id = ? AND deleted=false ) " +
		" SELECT  cw.ci_pipeline_id, max(cw.id) " +
		" FROM ci_workflow cw WHERE cw.ci_pipeline_id IN (SELECT cp.id FROM cp) OR cw.ci_pipeline_id IN (SELECT cp.parent_ci_pipeline FROM cp) " +
		" GROUP BY ci_pipeline_id) cw2 " +
		" ON cw1.id = cw2.max;"
	_, err := impl.dbConnection.Query(&ciworkflowStatuses, query, appId) //, pg.In(ciPipelineIds))
	if err != nil {
		impl.logger.Errorw("error occurred while fetching build type vs status vs count data", "err", err)
	}
	return ciworkflowStatuses, err
}
