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
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/helper"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/repository"
	"go.uber.org/zap"
)

type WorkflowTriggerAuditService interface {
	// SaveCiTriggerAudit saves audit data for CI trigger
	SaveCiTriggerAudit(request *bean.CiTriggerAuditRequest) (*repository.WorkflowConfigSnapshot, error)

	// SavePreCdTriggerAudit saves audit data for Pre-CD trigger
	SavePreCdTriggerAudit(request *bean.CdTriggerAuditRequest) (*repository.WorkflowConfigSnapshot, error)

	// SavePostCdTriggerAudit saves audit data for Post-CD trigger
	SavePostCdTriggerAudit(request *bean.CdTriggerAuditRequest) (*repository.WorkflowConfigSnapshot, error)

	// GetTriggerAuditByWorkflowId retrieves audit data by workflow ID and type
	GetTriggerAuditByWorkflowId(workflowId int, workflowType repository.WorkflowType) (*bean.WorkflowTriggerAuditResponse, error)

	// GetTriggerAuditHistory retrieves trigger audit history for a pipeline
	GetTriggerAuditHistory(pipelineId int, workflowType repository.WorkflowType, limit int, offset int) ([]*bean.WorkflowTriggerAuditResponse, error)

	// GetWorkflowConfigForRetrigger retrieves workflow configuration for retrigger
	//GetWorkflowConfigForRetrigger(auditId int) (*bean.RetriggerWorkflowConfig, error)

	sql.TransactionWrapper
}

type WorkflowTriggerAuditServiceImpl struct {
	logger                           *zap.SugaredLogger
	workflowConfigSnapshotRepository repository.WorkflowConfigSnapshotRepository
	secretSanitizer                  helper.SecretSanitizer
	*sql.TransactionUtilImpl
}

func NewWorkflowTriggerAuditServiceImpl(
	logger *zap.SugaredLogger,
	workflowConfigSnapshotRepository repository.WorkflowConfigSnapshotRepository,
	secretSanitizer helper.SecretSanitizer,
	transactionUtilImpl *sql.TransactionUtilImpl) *WorkflowTriggerAuditServiceImpl {

	return &WorkflowTriggerAuditServiceImpl{
		logger:                           logger,
		workflowConfigSnapshotRepository: workflowConfigSnapshotRepository,
		secretSanitizer:                  secretSanitizer,
		TransactionUtilImpl:              transactionUtilImpl,
	}
}

func (impl *WorkflowTriggerAuditServiceImpl) SaveCiTriggerAudit(request *bean.CiTriggerAuditRequest) (*repository.WorkflowConfigSnapshot, error) {
	tx, err := impl.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction for CI trigger audit", "err", err)
		return nil, err
	}
	defer impl.RollbackTx(tx)

	// Create consolidated workflow config snapshot
	configSnapshot, err := impl.createWorkflowConfigSnapshot(request.WorkflowRequest, repository.CI_WORKFLOW_TYPE, request.Pipeline, nil, request.WorkflowId, request.TriggerType, request.TriggeredBy, request.InfraConfigTriggerHistoryId, 0)
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

func (impl *WorkflowTriggerAuditServiceImpl) SavePreCdTriggerAudit(request *bean.CdTriggerAuditRequest) (*repository.WorkflowConfigSnapshot, error) {
	return impl.saveCdTriggerAudit(request, repository.PRE_CD_WORKFLOW_TYPE)
}

func (impl *WorkflowTriggerAuditServiceImpl) SavePostCdTriggerAudit(request *bean.CdTriggerAuditRequest) (*repository.WorkflowConfigSnapshot, error) {
	return impl.saveCdTriggerAudit(request, repository.POST_CD_WORKFLOW_TYPE)
}

func (impl *WorkflowTriggerAuditServiceImpl) saveCdTriggerAudit(request *bean.CdTriggerAuditRequest, workflowType repository.WorkflowType) (*repository.WorkflowConfigSnapshot, error) {
	tx, err := impl.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction for CD trigger audit", "err", err, "workflowType", workflowType)
		return nil, err
	}
	defer impl.RollbackTx(tx)

	// Create consolidated workflow config snapshot
	configSnapshot, err := impl.createWorkflowConfigSnapshot(request.WorkflowRequest, workflowType, request.Pipeline, request.Environment, request.WorkflowRunnerId, request.TriggerType, request.TriggeredBy, request.TriggerMetadata, request.InfraConfigTriggerHistoryId, request.ArtifactId)
	if err != nil {
		impl.logger.Errorw("error in creating workflow config snapshot for CD", "err", err, "workflowType", workflowType)
		return nil, err
	}

	savedSnapshot, err := impl.workflowConfigSnapshotRepository.SaveWithTx(tx, configSnapshot)
	if err != nil {
		impl.logger.Errorw("error in saving CD trigger audit", "err", err, "workflowType", workflowType)
		return nil, err
	}

	err = impl.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction for CD trigger audit", "err", err, "workflowType", workflowType)
		return nil, err
	}

	return savedSnapshot, nil
}

func (impl *WorkflowTriggerAuditServiceImpl) createWorkflowConfigSnapshot(workflowRequest *types.WorkflowRequest, workflowType repository.WorkflowType, pipeline interface{}, environment interface{}, workflowId int, triggerType string, triggeredBy int32, triggerMetadata interface{}, infraConfigTriggerHistoryId int, artifactId int) (*repository.WorkflowConfigSnapshot, error) {
	// Sanitize secrets before storing
	sanitizedWorkflowRequest, err := impl.secretSanitizer.SanitizeWorkflowRequest(workflowRequest)
	if err != nil {
		impl.logger.Errorw("error in sanitizing workflow request", "err", err)
		return nil, err
	}

	// Marshal sanitized workflow request
	workflowRequestJson, err := json.Marshal(sanitizedWorkflowRequest)
	if err != nil {
		impl.logger.Errorw("error in marshaling sanitized workflow request", "err", err)
		return nil, err
	}

	var pipelineId, appId int
	var environmentId int

	if workflowType == repository.CI_WORKFLOW_TYPE {
		ciPipeline := pipeline.(*pipelineConfig.CiPipeline)
		pipelineId = ciPipeline.Id
		appId = ciPipeline.AppId
	} else {
		cdPipeline := pipeline.(*pipelineConfig.Pipeline)
		pipelineId = cdPipeline.Id
		appId = cdPipeline.AppId
		environmentId = cdPipeline.EnvironmentId
	}

	configSnapshot := &repository.WorkflowConfigSnapshot{
		WorkflowId:                  workflowId,
		WorkflowType:                workflowType,
		PipelineId:                  pipelineId,
		AppId:                       appId,
		EnvironmentId:               environmentId,
		ArtifactId:                  artifactId,
		TriggerType:                 impl.getTriggerType(triggerType),
		TriggeredBy:                 triggeredBy,
		TriggerMetadata:             impl.marshalTriggerMetadata(triggerMetadata),
		InfraConfigTriggerHistoryId: infraConfigTriggerHistoryId,
		WorkflowRequestJson:         string(workflowRequestJson),
		WorkflowRequestVersion:      "V1", // for backward compatibility
		AuditLog:                    sql.NewDefaultAuditLog(triggeredBy),
	}

	return configSnapshot, nil
}

func (impl *WorkflowTriggerAuditServiceImpl) getTriggerType(triggerType string) repository.TriggerType {
	switch triggerType {
	case "MANUAL":
		return repository.MANUAL_TRIGGER
	case "AUTO":
		return repository.AUTO_TRIGGER
	case "WEBHOOK":
		return repository.WEBHOOK_TRIGGER
	default:
		return repository.MANUAL_TRIGGER
	}
}

func (impl *WorkflowTriggerAuditServiceImpl) marshalTriggerMetadata(metadata interface{}) string {
	if metadata == nil {
		return "{}"
	}
	metadataJson, err := json.Marshal(metadata)
	if err != nil {
		impl.logger.Errorw("error in marshaling trigger metadata", "err", err)
		return "{}"
	}
	return string(metadataJson)
}

func (impl *WorkflowTriggerAuditServiceImpl) GetTriggerAuditByWorkflowId(workflowId int, workflowType repository.WorkflowType) (*bean.WorkflowTriggerAuditResponse, error) {
	snapshot, err := impl.workflowConfigSnapshotRepository.FindByWorkflowIdAndType(workflowId, workflowType)
	if err != nil {
		impl.logger.Errorw("error in finding trigger audit by workflow id and type", "err", err, "workflowId", workflowId, "workflowType", workflowType)
		return nil, err
	}

	return impl.buildAuditResponse(snapshot)
}

func (impl *WorkflowTriggerAuditServiceImpl) GetTriggerAuditHistory(pipelineId int, workflowType repository.WorkflowType, limit int, offset int) ([]*bean.WorkflowTriggerAuditResponse, error) {
	snapshots, err := impl.workflowConfigSnapshotRepository.FindByPipelineId(pipelineId, limit, offset)
	if err != nil {
		impl.logger.Errorw("error in finding trigger audit history", "err", err, "pipelineId", pipelineId)
		return nil, err
	}

	var responses []*bean.WorkflowTriggerAuditResponse
	for _, snapshot := range snapshots {
		if snapshot.WorkflowType == workflowType {
			response, err := impl.buildAuditResponse(snapshot)
			if err != nil {
				impl.logger.Errorw("error in building audit response", "err", err, "snapshotId", snapshot.Id)
				continue
			}
			responses = append(responses, response)
		}
	}

	return responses, nil
}

func (impl *WorkflowTriggerAuditServiceImpl) buildAuditResponse(snapshot *repository.WorkflowConfigSnapshot) (*bean.WorkflowTriggerAuditResponse, error) {
	response := &bean.WorkflowTriggerAuditResponse{
		Id:              snapshot.Id,
		WorkflowId:      snapshot.WorkflowId,
		WorkflowType:    string(snapshot.WorkflowType),
		PipelineId:      snapshot.PipelineId,
		AppId:           snapshot.AppId,
		EnvironmentId:   snapshot.EnvironmentId,
		ArtifactId:      snapshot.ArtifactId,
		TriggerType:     string(snapshot.TriggerType),
		TriggeredBy:     snapshot.TriggeredBy,
		TriggerMetadata: snapshot.TriggerMetadata,
		Status:          "SAVED", // Always saved in simplified approach
		CreatedOn:       snapshot.CreatedOn,
		ConfigSnapshot:  snapshot,
	}

	return response, nil
}

//func (impl *WorkflowTriggerAuditServiceImpl) GetWorkflowConfigForRetrigger(auditId int) (*bean.RetriggerWorkflowConfig, error) {
//	// Get config snapshot
//	configSnapshot, err := impl.workflowConfigSnapshotRepository.FindById(auditId)
//	if err != nil {
//		impl.logger.Errorw("error in finding config snapshot for retrigger", "err", err, "auditId", auditId)
//		return nil, err
//	}
//
//	// Unmarshal sanitized workflow request
//	var sanitizedWorkflowRequest interface{}
//	err = json.Unmarshal([]byte(configSnapshot.WorkflowRequestJson), &sanitizedWorkflowRequest)
//	if err != nil {
//		impl.logger.Errorw("error in unmarshaling sanitized workflow request for retrigger", "err", err)
//		return nil, err
//	}
//
//	// Reconstruct secrets with current values from environment
//	reconstructedWorkflowRequest, err := impl.secretSanitizer.ReconstructSecrets(sanitizedWorkflowRequest, configSnapshot.EnvironmentId, configSnapshot.AppId)
//	if err != nil {
//		impl.logger.Errorw("error in reconstructing secrets for retrigger", "err", err)
//		return nil, err
//	}
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
