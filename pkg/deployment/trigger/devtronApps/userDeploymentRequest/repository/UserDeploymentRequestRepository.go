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
	"encoding/json"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.opentelemetry.io/otel"
	"time"
)

type UserDeploymentRequest struct {
	tableName            struct{}                            `sql:"user_deployment_request" pg:",discard_unknown_columns"`
	Id                   int                                 `sql:"id,pk"`
	PipelineId           int                                 `sql:"pipeline_id"`
	CiArtifactId         int                                 `sql:"ci_artifact_id"`
	AdditionalOverride   json.RawMessage                     `sql:"additional_override"`
	ForceTrigger         bool                                `sql:"force_trigger"`
	ForceSyncDeployment  bool                                `sql:"force_sync_deployment"`
	Strategy             string                              `sql:"strategy"`
	DeploymentWithConfig apiBean.DeploymentConfigurationType `sql:"deployment_with_config"`
	SpecificTriggerWfrId int                                 `sql:"specific_trigger_wfr_id"` // target cd_workflow_runner_id for rollback. Used in rollback deployment cases
	CdWorkflowId         int                                 `sql:"cd_workflow_id"`
	DeploymentType       models.DeploymentType               `sql:"deployment_type"`
	TriggeredAt          time.Time                           `sql:"triggered_at"`
	TriggeredBy          int32                               `sql:"triggered_by"`
	Status               bean.UserDeploymentRequestStatus    `sql:"status"`
}

type UserDeploymentRequestWithAdditionalFields struct {
	UserDeploymentRequest
	CdWorkflowRunnerId int `sql:"cd_workflow_runner_id"`
	PipelineOverrideId int `sql:"pipeline_override_id"`
}

type UserDeploymentRequestRepository interface {
	// transaction util funcs
	sql.TransactionWrapper
	Save(ctx context.Context, models ...*UserDeploymentRequest) error
	FindById(id int) (*UserDeploymentRequestWithAdditionalFields, error)
	GetLatestIdForPipeline(deploymentReqId int) (int, error)
	FindByCdWfId(cdWfId int) (*UserDeploymentRequest, error)
	FindByCdWfIds(ctx context.Context, cdWfIds ...int) ([]UserDeploymentRequest, error)
	GetAllInCompleteRequests() ([]UserDeploymentRequestWithAdditionalFields, error)
	MarkAllPreviousSuperseded(ctx context.Context, tx *pg.Tx, pipelineId, previousToReqId int) (int, error)
	UpdateStatusForCdWfIds(ctx context.Context, tx *pg.Tx, status bean.UserDeploymentRequestStatus, cdWfIds ...int) (int, error)
	IsLatestForPipelineId(id, pipelineId int) (bool, error)
	TerminateForPipelineId(tx *pg.Tx, pipelineId int) (int, error)
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

func (impl *UserDeploymentRequestRepositoryImpl) Save(ctx context.Context, models ...*UserDeploymentRequest) error {
	_, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestRepositoryImpl.Save")
	defer span.End()
	return impl.dbConnection.Insert(&models)
}

func (impl *UserDeploymentRequestRepositoryImpl) FindById(id int) (*UserDeploymentRequestWithAdditionalFields, error) {
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

func (impl *UserDeploymentRequestRepositoryImpl) GetLatestIdForPipeline(deploymentReqId int) (int, error) {
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

func (impl *UserDeploymentRequestRepositoryImpl) FindByCdWfIds(ctx context.Context, cdWfIds ...int) ([]UserDeploymentRequest, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestRepositoryImpl.FindByCdWfIds")
	defer span.End()
	if len(cdWfIds) == 0 {
		return nil, pg.ErrNoRows
	}
	var model []UserDeploymentRequest
	err := impl.dbConnection.Model(&model).
		Where("cd_workflow_id IN (?)", pg.In(cdWfIds)).
		Order("id DESC").
		Select()
	return model, err
}

func (impl *UserDeploymentRequestRepositoryImpl) GetAllInCompleteRequests() ([]UserDeploymentRequestWithAdditionalFields, error) {
	var model []UserDeploymentRequestWithAdditionalFields
	query := impl.dbConnection.Model().
		Table("user_deployment_request").
		ColumnExpr("MAX(id)").
		Where("status NOT IN (?)", pg.In([]bean.UserDeploymentRequestStatus{
			bean.DeploymentRequestCompleted, bean.DeploymentRequestSuperseded,
		})).
		Group("pipeline_id")
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
		Where("cdwfr.status NOT IN (?)", pg.In(append(pipelineConfig.WfrTerminalStatusList, pipelineConfig.WorkflowInQueue))).
		Where("user_deployment_request.id IN (?)", query).
		Select(&model)
	return model, err
}

func (impl *UserDeploymentRequestRepositoryImpl) MarkAllPreviousSuperseded(ctx context.Context, tx *pg.Tx, pipelineId, previousToReqId int) (int, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestRepositoryImpl.MarkAllPreviousSuperseded")
	defer span.End()
	var query *orm.Query
	if tx == nil {
		query = impl.dbConnection.Model((*UserDeploymentRequest)(nil))
	} else {
		query = tx.Model((*UserDeploymentRequest)(nil))
	}
	res, err := query.
		Set("status = ?", bean.DeploymentRequestSuperseded).
		Where("pipeline_id = ?", pipelineId).
		Where("id < ?", previousToReqId).
		Where("status NOT IN (?)", pg.In([]bean.UserDeploymentRequestStatus{
			bean.DeploymentRequestCompleted, bean.DeploymentRequestSuperseded,
		})).
		Update()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), err
}

func (impl *UserDeploymentRequestRepositoryImpl) UpdateStatusForCdWfIds(ctx context.Context, tx *pg.Tx, status bean.UserDeploymentRequestStatus, cdWfIds ...int) (int, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestRepositoryImpl.UpdateStatusForCdWfIds")
	defer span.End()
	if len(cdWfIds) == 0 {
		return 0, nil
	}
	var query *orm.Query
	if tx == nil {
		query = impl.dbConnection.Model((*UserDeploymentRequest)(nil))
	} else {
		query = tx.Model((*UserDeploymentRequest)(nil))
	}
	res, err := query.
		Set("status = ?", status).
		Where("cd_workflow_id IN (?)", pg.In(cdWfIds)).
		Update()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), err
}

func (impl *UserDeploymentRequestRepositoryImpl) TerminateForPipelineId(tx *pg.Tx, pipelineId int) (int, error) {
	var query *orm.Query
	if tx == nil {
		query = impl.dbConnection.Model((*UserDeploymentRequest)(nil))
	} else {
		query = tx.Model((*UserDeploymentRequest)(nil))
	}
	res, err := query.
		Set("status = ?", bean.DeploymentRequestTerminated).
		Join("INNER join pipeline p").
		JoinOn("user_deployment_request.pipeline_id = p.id").
		Where("user_deployment_request.pipeline_id = ?", pipelineId).
		Where("p.deleted = ?", true).
		Update()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), err
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
