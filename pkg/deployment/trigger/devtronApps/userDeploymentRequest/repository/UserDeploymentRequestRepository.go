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
	"encoding/json"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"time"
)

type UserDeploymentRequest struct {
	tableName            struct{}                            `sql:"user_deployment_request" pg:",discard_unknown_columns"`
	Id                   int                                 `sql:"id,pk"`
	PipelineId           int                                 `sql:"pipeline_id"`
	CiArtifactId         int                                 `sql:"ciArtifact_id"`
	AdditionalOverride   json.RawMessage                     `sql:"additional_override"`
	ForceTrigger         bool                                `sql:"force_trigger"`
	ForceSync            bool                                `sql:"force_sync"`
	Strategy             string                              `sql:"strategy"`
	DeploymentWithConfig apiBean.DeploymentConfigurationType `sql:"deployment_with_config"`
	SpecificTriggerWfrId int                                 `sql:"specific_trigger_wfr_id"`
	CdWorkflowId         int                                 `sql:"cd_workflow_id"`
	DeploymentType       models.DeploymentType               `sql:"deployment_type"`
	TriggeredAt          time.Time                           `sql:"triggered_at"`
	TriggeredBy          int32                               `sql:"triggered_by"`
	Status               bean.UserDeploymentRequestStatus    `sql:"status"`
}

type UserDeploymentRequestRepository interface {
	// transaction util funcs
	sql.TransactionWrapper
	Save(tx *pg.Tx, models []*UserDeploymentRequest) error
	FindById(id int) (*UserDeploymentRequest, error)
	FindByCdWfId(cdWfId int) (*UserDeploymentRequest, error)
	FindByCdWfIds(cdWfIds ...int) ([]UserDeploymentRequest, error)
	MarkAllPreviousSuperseded(tx *pg.Tx, pipelineId, previousToReqId int) (int, error)
	UpdateStatusForCdWfIds(tx *pg.Tx, status bean.UserDeploymentRequestStatus, cdWfIds ...int) (int, error)
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

func (impl *UserDeploymentRequestRepositoryImpl) Save(tx *pg.Tx, models []*UserDeploymentRequest) error {
	if tx == nil {
		return impl.dbConnection.Insert(models)
	}
	return tx.Insert(models)
}

func (impl *UserDeploymentRequestRepositoryImpl) FindById(id int) (*UserDeploymentRequest, error) {
	model := &UserDeploymentRequest{}
	err := impl.dbConnection.Model(model).
		Where("id = ?", id).
		Select()
	return model, err
}

func (impl *UserDeploymentRequestRepositoryImpl) FindByCdWfId(cdWfId int) (*UserDeploymentRequest, error) {
	model := &UserDeploymentRequest{}
	err := impl.dbConnection.Model(model).
		Where("cd_workflow_id = ?", cdWfId).
		Select()
	return model, err
}

func (impl *UserDeploymentRequestRepositoryImpl) FindByCdWfIds(cdWfIds ...int) ([]UserDeploymentRequest, error) {
	var model []UserDeploymentRequest
	err := impl.dbConnection.Model(model).
		Where("cd_workflow_id IN (?)", cdWfIds).
		Order("id DESC").
		Select()
	return model, err
}

func (impl *UserDeploymentRequestRepositoryImpl) MarkAllPreviousSuperseded(tx *pg.Tx, pipelineId, previousToReqId int) (int, error) {
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
	return res.RowsAffected(), err
}

func (impl *UserDeploymentRequestRepositoryImpl) UpdateStatusForCdWfIds(tx *pg.Tx, status bean.UserDeploymentRequestStatus, cdWfIds ...int) (int, error) {
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
	return res.RowsAffected(), err
}

func (impl *UserDeploymentRequestRepositoryImpl) IsLatestForPipelineId(id, pipelineId int) (bool, error) {
	model := &UserDeploymentRequest{}
	ifAnySuccessorExists, err := impl.dbConnection.
		Model(model).
		Column("user_deployment_request.*").
		Join("INNER join cd_workflow wf").
		JoinOn("wf.id = user_deployment_request.cd_workflow_id").
		Where("wf.pipeline_id = ?", pipelineId).
		Where("user_deployment_request.id > ?", id).
		Order("user_deployment_request.id DESC").
		Exists()
	return !ifAnySuccessorExists, err
}
