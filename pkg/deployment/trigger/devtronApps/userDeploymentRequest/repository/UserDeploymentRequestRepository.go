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

package repository

import (
	"context"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"time"
)

type UserDeploymentRequest struct {
	tableName            struct{}                            `sql:"user_deployment_request" pg:",discard_unknown_columns"`
	Id                   int                                 `sql:"id,pk"`
	PipelineId           int                                 `sql:"pipeline_id"`
	CiArtifactId         int                                 `sql:"ci_artifact_id"`
	AdditionalOverride   []byte                              `sql:"additional_override"`
	ForceTrigger         bool                                `sql:"force_trigger"`
	ForceSyncDeployment  bool                                `sql:"force_sync_deployment"`
	Strategy             string                              `sql:"strategy"`
	DeploymentWithConfig apiBean.DeploymentConfigurationType `sql:"deployment_with_config"`
	SpecificTriggerWfrId int                                 `sql:"specific_trigger_wfr_id"` // target cd_workflow_runner_id for rollback. Used in rollback deployment cases
	CdWorkflowId         int                                 `sql:"cd_workflow_id"`
	DeploymentType       models.DeploymentType               `sql:"deployment_type"`
	TriggeredAt          time.Time                           `sql:"triggered_at"`
	TriggeredBy          int32                               `sql:"triggered_by"`
}

type UserDeploymentRequestWithAdditionalFields struct {
	UserDeploymentRequest
	CdWorkflowRunnerId int `sql:"cd_workflow_runner_id"`
	PipelineOverrideId int `sql:"pipeline_override_id"`
}

type UserDeploymentRequestRepository interface {
	// transaction util funcs
	sql.TransactionWrapper
	Save(ctx context.Context, tx *pg.Tx, models ...*UserDeploymentRequest) error
	FindById(ctx context.Context, id int) (*UserDeploymentRequestWithAdditionalFields, error)
	GetLatestIdForPipeline(ctx context.Context, deploymentReqId int) (int, error)
	FindByCdWfId(cdWfId int) (*UserDeploymentRequest, error)
	GetAllInCompleteRequests(ctx context.Context) ([]UserDeploymentRequestWithAdditionalFields, error)
	IsLatestForPipelineId(id, pipelineId int) (bool, error)
}

func NewUserDeploymentRequestRepositoryImpl(dbConnection *pg.DB, transactionUtilImpl *sql.TransactionUtilImpl) *UserDeploymentRequestRepositoryImpl {
	return &UserDeploymentRequestRepositoryImpl{
		dbConnection:        dbConnection,
		TransactionUtilImpl: transactionUtilImpl,
	}
}

type UserDeploymentRequestRepositoryImpl struct {
	*sql.TransactionUtilImpl
	dbConnection *pg.DB
}

func (impl *UserDeploymentRequestRepositoryImpl) Save(ctx context.Context, tx *pg.Tx, models ...*UserDeploymentRequest) error {
	_, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestRepositoryImpl.Save")
	defer span.End()
	if tx != nil {
		return tx.Insert(&models)
	}
	return impl.dbConnection.Insert(&models)
}

func (impl *UserDeploymentRequestRepositoryImpl) FindById(ctx context.Context, id int) (*UserDeploymentRequestWithAdditionalFields, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestRepositoryImpl.FindById")
	defer span.End()
	model := &UserDeploymentRequestWithAdditionalFields{}
	err := impl.dbConnection.Model().
		Table("user_deployment_request").
		Column("user_deployment_request.*").
		ColumnExpr("cdwfr.id AS cd_workflow_runner_id").
		ColumnExpr("pco.id AS pipeline_override_id").
		Join("INNER JOIN cd_workflow_runner cdwfr").
		JoinOn("user_deployment_request.cd_workflow_id = cdwfr.cd_workflow_id").
		JoinOn("cdwfr.workflow_type = ?", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
		Join("LEFT JOIN pipeline_config_override pco").
		JoinOn("user_deployment_request.cd_workflow_id = pco.cd_workflow_id").
		Where("user_deployment_request.id = ?", id).
		Select(model)
	return model, err
}

func (impl *UserDeploymentRequestRepositoryImpl) FindByCdWfId(cdWfId int) (*UserDeploymentRequest, error) {
	model := &UserDeploymentRequest{}
	err := impl.dbConnection.Model(model).
		Where("cd_workflow_id = ?", cdWfId).
		Select()
	return model, err
}

func (impl *UserDeploymentRequestRepositoryImpl) GetLatestIdForPipeline(ctx context.Context, deploymentReqId int) (int, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestRepositoryImpl.GetLatestIdForPipeline")
	defer span.End()
	var latestId int
	query := impl.dbConnection.Model().
		Table("user_deployment_request").
		Column("pipeline_id").
		Where("id = ?", deploymentReqId)
	err := impl.dbConnection.Model().
		Table("user_deployment_request").
		ColumnExpr("MAX(id) AS id").
		Where("pipeline_id IN (?)", query).
		Select(&latestId)
	return latestId, err
}

// TODO Asutosh: analyse query execution time
func (impl *UserDeploymentRequestRepositoryImpl) GetAllInCompleteRequests(ctx context.Context) ([]UserDeploymentRequestWithAdditionalFields, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestRepositoryImpl.GetAllInCompleteRequests")
	defer span.End()
	var model []UserDeploymentRequestWithAdditionalFields
	subQuery := impl.dbConnection.Model().
		Table("pipeline_status_timeline").
		ColumnExpr("1").
		Where("pipeline_status_timeline.cd_workflow_runner_id = cdwfr.id").
		Where("pipeline_status_timeline.status = ?", pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_COMPLETED)
	latestRequestQuery := impl.dbConnection.Model().
		Table("user_deployment_request").
		ColumnExpr("MAX(user_deployment_request.id)").
		Join("INNER JOIN cd_workflow cdwf").
		JoinOn("user_deployment_request.cd_workflow_id = cdwf.id").
		Join("INNER JOIN cd_workflow_runner cdwfr").
		JoinOn("cdwf.id = cdwfr.cd_workflow_id").
		Join("LEFT JOIN pipeline_status_timeline pst").
		JoinOn("cdwfr.id = pst.cd_workflow_runner_id").
		Where("cdwfr.workflow_type = ?", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
		Where("cdwfr.status NOT IN (?)", pg.In(append(pipelineConfig.WfrTerminalStatusList, pipelineConfig.WorkflowInQueue))).
		Where("pst.status = ?", pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_REQUEST_VALIDATED).
		Where("NOT EXISTS (?)", subQuery).
		Group("pipeline_id")
	err := impl.dbConnection.Model().
		Table("user_deployment_request").
		Column("user_deployment_request.*").
		ColumnExpr("cdwfr.id AS cd_workflow_runner_id").
		ColumnExpr("pco.id AS pipeline_override_id").
		Join("INNER JOIN cd_workflow_runner cdwfr").
		JoinOn("user_deployment_request.cd_workflow_id = cdwfr.cd_workflow_id").
		Join("LEFT JOIN pipeline_config_override pco").
		JoinOn("user_deployment_request.cd_workflow_id = pco.cd_workflow_id").
		Where("cdwfr.workflow_type = ?", apiBean.CD_WORKFLOW_TYPE_DEPLOY).
		Where("user_deployment_request.id IN (?)", latestRequestQuery).
		Select(&model)
	return model, err
}

func (impl *UserDeploymentRequestRepositoryImpl) IsLatestForPipelineId(id, pipelineId int) (bool, error) {
	model := &UserDeploymentRequest{}
	ifAnySuccessorExists, err := impl.dbConnection.
		Model(model).
		Column("user_deployment_request.*").
		Join("INNER join cd_workflow wf").
		JoinOn("user_deployment_request.cd_workflow_id = wf.id").
		Where("wf.pipeline_id = ?", pipelineId).
		Where("user_deployment_request.id > ?", id).
		Order("user_deployment_request.id DESC").
		Exists()
	return !ifAnySuccessorExists, err
}
