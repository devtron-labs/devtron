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
	"fmt"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/adapter"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/repository"
	"go.uber.org/zap"
)

type WorkflowTriggerAuditService interface {
	// SaveTriggerAudit saves audit data for CI, Pre-cd and Post-cd trigger
	SaveTriggerAudit(workflowRequest *types.WorkflowRequest) (*repository.WorkflowConfigSnapshot, error)

	// GetWorkflowRequestFromSnapshotForRetrigger fetches workflow request by workflowId and workflowType from snapshot for retrigger
	GetWorkflowRequestFromSnapshotForRetrigger(workflowId int, workflowType types.WorkflowType) (*types.WorkflowRequest, error)
}

type WorkflowTriggerAuditServiceImpl struct {
	logger                           *zap.SugaredLogger
	workflowConfigSnapshotRepository repository.WorkflowConfigSnapshotRepository
	config                           *types.CiCdConfig
	*sql.TransactionUtilImpl
}

func NewWorkflowTriggerAuditServiceImpl(
	logger *zap.SugaredLogger,
	workflowConfigSnapshotRepository repository.WorkflowConfigSnapshotRepository,
	config *types.CiCdConfig,
	transactionUtilImpl *sql.TransactionUtilImpl) *WorkflowTriggerAuditServiceImpl {

	return &WorkflowTriggerAuditServiceImpl{
		logger:                           logger,
		workflowConfigSnapshotRepository: workflowConfigSnapshotRepository,
		config:                           config,
		TransactionUtilImpl:              transactionUtilImpl,
	}
}

func (impl *WorkflowTriggerAuditServiceImpl) SaveTriggerAudit(workflowRequest *types.WorkflowRequest) (*repository.WorkflowConfigSnapshot, error) {
	tx, err := impl.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction for CI trigger audit", "err", err)
		return nil, err
	}
	defer impl.RollbackTx(tx)

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
	compressedWorkflowJson, err := sanitizedWorkflowRequest.CompressWorkflowRequest()
	if err != nil {
		impl.logger.Errorw("error in compressing sanitized workflow request", "err", err)
		return nil, err
	}
	workflowType, pipelineId, workflowId := types.PRE_CD_WORKFLOW_TYPE, workflowRequest.CdPipelineId, workflowRequest.WorkflowRunnerId
	if workflowRequest.IsCdStageTypePost() {
		workflowType = types.POST_CD_WORKFLOW_TYPE
	} else if workflowRequest.IsCiTypeWorkflowRequest() {
		workflowType, pipelineId, workflowId = types.CI_WORKFLOW_TYPE, workflowRequest.PipelineId, workflowRequest.WorkflowId
	}
	configSnapshot := adapter.GetWorkflowConfigSnapshot(workflowId, workflowType, pipelineId, compressedWorkflowJson, types.TriggerAuditSchemaVersionV1, workflowRequest.TriggeredBy)
	return configSnapshot, nil
}

// GetWorkflowRequestFromSnapshotForRetrigger retrieves workflow request for retrigger scenarios
func (impl *WorkflowTriggerAuditServiceImpl) GetWorkflowRequestFromSnapshotForRetrigger(workflowId int, workflowType types.WorkflowType) (*types.WorkflowRequest, error) {
	// For retrigger, we want to get the original failed workflow's snapshot
	snapshot, err := impl.workflowConfigSnapshotRepository.FindLatestFailedWorkflowSnapshot(workflowId, workflowType)
	if err != nil {
		impl.logger.Errorw("error in finding failed workflow config snapshot for retrigger", "err", err, "workflowId", workflowId, "workflowType", workflowType)
		return nil, err
	}

	// Decompress and unmarshal the workflow request
	var workflowRequest types.WorkflowRequest
	err = workflowRequest.DecompressWorkflowRequest(snapshot.WorkflowRequestJson)
	if err != nil {
		impl.logger.Errorw("error in decompressing workflow request from snapshot for retrigger", "err", err, "snapshotId", snapshot.Id)
		return nil, err
	}

	// Restore secrets from current environment variables
	err = impl.restoreSecretsInWorkflowRequest(&workflowRequest)
	if err != nil {
		impl.logger.Errorw("error in restoring secrets in workflow request", "err", err, "workflowId", workflowId)
		return nil, err
	}

	impl.logger.Infow("successfully retrieved workflow request from snapshot for retrigger", "workflowId", workflowId, "workflowType", workflowType, "snapshotId", snapshot.Id)
	return &workflowRequest, nil
}

// RestoreSecretsInWorkflowRequest restores secrets that were sanitized during storage
func (impl *WorkflowTriggerAuditServiceImpl) restoreSecretsInWorkflowRequest(workflowRequest *types.WorkflowRequest) error {
	impl.logger.Debugw("restoring secrets in workflow request", "workflowId", workflowRequest.WorkflowId)

	// Restore secrets in blob storage config
	switch workflowRequest.CloudProvider {
	case types.BLOB_STORAGE_S3:
		if workflowRequest.BlobStorageS3Config != nil {
			workflowRequest.BlobStorageS3Config.AccessKey = impl.config.BlobStorageS3AccessKey
			workflowRequest.BlobStorageS3Config.Passkey = impl.config.BlobStorageS3SecretKey
		}
	case types.BLOB_STORAGE_GCP:
		if workflowRequest.GcpBlobConfig != nil {
			workflowRequest.GcpBlobConfig.CredentialFileJsonData = impl.config.BlobStorageGcpCredentialJson
		}
	case types.BLOB_STORAGE_AZURE:
		if workflowRequest.AzureBlobConfig != nil {
			workflowRequest.AzureBlobConfig.AccountKey = impl.config.AzureAccountKey
		}
		if workflowRequest.BlobStorageS3Config != nil {
			workflowRequest.BlobStorageS3Config.AccessKey = impl.config.AzureAccountName
		}
	default:
		if impl.config.BlobStorageEnabled {
			return fmt.Errorf("blob storage %s not supported", workflowRequest.CloudProvider)
		}
	}

	impl.logger.Debugw("completed secret restoration in workflow request", "workflowId", workflowRequest.WorkflowId)
	return nil
}
