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

package service

import (
	"encoding/json"
	//"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	//"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/repository"
	"go.uber.org/zap"
)

type WorkflowTriggerAuditService interface {
	// SaveCiTriggerAudit saves audit data for CI trigger
	SaveCiTriggerAudit(workflowRequest *types.WorkflowRequest) (*repository.WorkflowConfigSnapshot, error)

	// SaveCdTriggerAudit saves audit data for Pre-CD trigger
	SaveCdTriggerAudit(workflowRequest *types.WorkflowRequest) (*repository.WorkflowConfigSnapshot, error)

	// GetTriggerAuditByWorkflowId retrieves audit data by workflow ID and type

	// GetTriggerAuditHistory retrieves trigger audit history for a pipeline
	//GetTriggerAuditHistory(pipelineId int, workflowType types.WorkflowType, limit int, offset int) ([]*bean.WorkflowTriggerAuditResponse, error)

	// GetWorkflowConfigForRetrigger retrieves workflow configuration for retrigger
	//GetWorkflowConfigForRetrigger(auditId int) (*bean.RetriggerWorkflowConfig, error)

	//sql.TransactionWrapper
}

type WorkflowTriggerAuditServiceImpl struct {
	logger                           *zap.SugaredLogger
	workflowConfigSnapshotRepository repository.WorkflowConfigSnapshotRepository
	*sql.TransactionUtilImpl
}

func NewWorkflowTriggerAuditServiceImpl(
	logger *zap.SugaredLogger,
	workflowConfigSnapshotRepository repository.WorkflowConfigSnapshotRepository,
	transactionUtilImpl *sql.TransactionUtilImpl) *WorkflowTriggerAuditServiceImpl {

	return &WorkflowTriggerAuditServiceImpl{
		logger:                           logger,
		workflowConfigSnapshotRepository: workflowConfigSnapshotRepository,
		TransactionUtilImpl:              transactionUtilImpl,
	}
}

func (impl *WorkflowTriggerAuditServiceImpl) SaveCiTriggerAudit(workflowRequest *types.WorkflowRequest) (*repository.WorkflowConfigSnapshot, error) {
	tx, err := impl.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction for CI trigger audit", "err", err)
		return nil, err
	}
	defer impl.RollbackTx(tx)

	// Create consolidated workflow config snapshot
	configSnapshot, err := impl.createWorkflowConfigSnapshot(workflowRequest)
	if err != nil {
		impl.logger.Errorw("error in creating workflow config snapshot for CI", "err", err)
		return nil, err
	}

	savedSnapshot, err := impl.workflowConfigSnapshotRepository.SaveWithTx(tx, configSnapshot)
	if err != nil {
		impl.logger.Errorw("error in saving CI trigger audit", "err", err)
		return nil, err
	}

	err = impl.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction for CI trigger audit", "err", err)
		return nil, err
	}

	return savedSnapshot, nil
}

func (impl *WorkflowTriggerAuditServiceImpl) SaveCdTriggerAudit(workflowRequest *types.WorkflowRequest) (*repository.WorkflowConfigSnapshot, error) {
	tx, err := impl.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction for CD trigger audit", "workflowRunnerId", workflowRequest.WorkflowRunnerId, "stageType", workflowRequest.StageType, "err", err)
		return nil, err
	}
	defer impl.RollbackTx(tx)

	// Create consolidated workflow config snapshot
	configSnapshot, err := impl.createWorkflowConfigSnapshot(workflowRequest)
	if err != nil {
		impl.logger.Errorw("error in creating workflow config snapshot for CD", "workflowRunnerId", workflowRequest.WorkflowRunnerId, "stageType", workflowRequest.StageType, "err", err)
		return nil, err
	}

	savedSnapshot, err := impl.workflowConfigSnapshotRepository.SaveWithTx(tx, configSnapshot)
	if err != nil {
		impl.logger.Errorw("error in saving CD trigger audit", "workflowRunnerId", workflowRequest.WorkflowRunnerId, "stageType", workflowRequest.StageType, "err", err)
		return nil, err
	}

	err = impl.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction for CD trigger audit", "workflowRunnerId", workflowRequest.WorkflowRunnerId, "stageType", workflowRequest.StageType, "err", err)
		return nil, err
	}

	return savedSnapshot, nil
}

func (impl *WorkflowTriggerAuditServiceImpl) maskSecretsInWorkflowRequest(workflowRequest *types.WorkflowRequest) *types.WorkflowRequest {
	if workflowRequest.BlobStorageS3Config != nil {
		workflowRequest.BlobStorageS3Config.AccessKey = ""
		workflowRequest.BlobStorageS3Config.Passkey = ""
	}
	if workflowRequest.AzureBlobConfig != nil {
		workflowRequest.AzureBlobConfig.AccountKey = ""
	}
	if workflowRequest.GcpBlobConfig != nil {
		workflowRequest.GcpBlobConfig.CredentialFileJsonData = ""
	}
	return workflowRequest
}

func (impl *WorkflowTriggerAuditServiceImpl) createWorkflowConfigSnapshot(workflowRequest *types.WorkflowRequest) (*repository.WorkflowConfigSnapshot, error) {
	// sanitize secrets before storing
	sanitizedWorkflowRequest := impl.maskSecretsInWorkflowRequest(workflowRequest)
	workflowRequestJson, err := json.Marshal(sanitizedWorkflowRequest)
	if err != nil {
		impl.logger.Errorw("error in marshaling sanitized workflow request", "err", err)
		return nil, err
	}

	workflowType, pipelineId, workflowId := types.PRE_CD_WORKFLOW_TYPE, workflowRequest.CdPipelineId, workflowRequest.WorkflowRunnerId
	if workflowRequest.IsCdStageTypePost() {
		workflowType = types.POST_CD_WORKFLOW_TYPE
	} else {
		workflowType, pipelineId, workflowId = types.CI_WORKFLOW_TYPE, workflowRequest.PipelineId, workflowRequest.WorkflowId
	}

	configSnapshot := &repository.WorkflowConfigSnapshot{
		WorkflowId:                   workflowId,
		WorkflowType:                 workflowType,
		PipelineId:                   pipelineId,
		WorkflowRequestJson:          string(workflowRequestJson),
		WorkflowRequestSchemaVersion: types.TriggerAuditSchemaVersionV1,
		AuditLog:                     sql.NewDefaultAuditLog(workflowRequest.TriggeredBy),
	}

	return configSnapshot, nil
}

//
//	// Convert back to WorkflowRequest type
//	reconstructedJson, err := json.Marshal(reconstructedWorkflowRequest)
//	if err != nil {
//		impl.logger.Errorw("error in marshaling reconstructed workflow request", "err", err)
//		return nil, err
//	}
//
//	var workflowRequest types.WorkflowRequest
//	err = json.Unmarshal(reconstructedJson, &workflowRequest)
//	if err != nil {
//		impl.logger.Errorw("error in unmarshaling reconstructed workflow request", "err", err)
//		return nil, err
//	}
//
//	retriggerConfig := &bean.RetriggerWorkflowConfig{
//		AuditId:             configSnapshot.Id,
//		WorkflowType:        string(configSnapshot.WorkflowType),
//		PipelineId:          configSnapshot.PipelineId,
//		AppId:               configSnapshot.AppId,
//		EnvironmentId:       configSnapshot.EnvironmentId,
//		ArtifactId:          configSnapshot.ArtifactId,
//		WorkflowRequest:     &workflowRequest,
//		ConfigSnapshot:      configSnapshot,
//		OriginalTriggeredBy: configSnapshot.TriggeredBy,
//		OriginalTriggerTime: configSnapshot.CreatedOn,
//	}
//
//	return retriggerConfig, nil
//}
