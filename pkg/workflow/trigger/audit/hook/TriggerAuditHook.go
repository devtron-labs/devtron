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

package hook

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	repository4 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/service"
	"go.uber.org/zap"
)

// TriggerAuditHook provides common hooks for auditing workflow triggers
type TriggerAuditHook interface {
	// AuditCiTrigger audits CI trigger
	AuditCiTrigger(workflowId int, pipeline *pipelineConfig.CiPipeline, workflowRequest *types.WorkflowRequest,
		triggerType string, triggeredBy int32,
		infraConfigTriggerHistoryId int) error

	// AuditPreCdTrigger audits Pre-CD trigger
	AuditPreCdTrigger(workflowRunnerId int, pipeline *pipelineConfig.Pipeline, environment *repository4.Environment,
		workflowRequest *types.WorkflowRequest,
		triggerType string, triggeredBy int32, infraConfigTriggerHistoryId int) error

	// AuditPostCdTrigger audits Post-CD trigger
	AuditPostCdTrigger(workflowRunnerId int, pipeline *pipelineConfig.Pipeline, environment *repository4.Environment,
		workflowRequest *types.WorkflowRequest,
		triggerType string, triggeredBy int32, infraConfigTriggerHistoryId int) error

	// GetRetriggerConfig gets configuration for retrigger
	//GetRetriggerConfig(auditId int) (*bean.RetriggerWorkflowConfig, error)
}

type TriggerAuditHookImpl struct {
	logger                      *zap.SugaredLogger
	workflowTriggerAuditService service.WorkflowTriggerAuditService
}

func NewTriggerAuditHookImpl(logger *zap.SugaredLogger, workflowTriggerAuditService service.WorkflowTriggerAuditService) *TriggerAuditHookImpl {
	return &TriggerAuditHookImpl{
		logger:                      logger,
		workflowTriggerAuditService: workflowTriggerAuditService,
	}
}

func (impl *TriggerAuditHookImpl) AuditCiTrigger(workflowId int, pipeline *pipelineConfig.CiPipeline,
	workflowRequest *types.WorkflowRequest, triggerType string, triggeredBy int32, infraConfigTriggerHistoryId int) error {

	impl.logger.Infow("auditing CI trigger", "workflowId", workflowId, "pipelineId", pipeline.Id, "triggeredBy", triggeredBy)

	request := &bean.CiTriggerAuditRequest{
		WorkflowId: workflowId,
		Pipeline:   pipeline,
		CommonAuditRequest: &bean.CommonAuditRequest{
			WorkflowRequest:             workflowRequest,
			TriggerType:                 triggerType,
			TriggeredBy:                 triggeredBy,
			InfraConfigTriggerHistoryId: infraConfigTriggerHistoryId,
		},
	}

	_, err := impl.workflowTriggerAuditService.SaveCiTriggerAudit(request)
	if err != nil {
		impl.logger.Errorw("error in auditing CI trigger", "err", err, "workflowId", workflowId)
		// Don't fail the trigger if audit fails, just log the error
		return nil
	}

	impl.logger.Infow("successfully audited CI trigger", "workflowId", workflowId, "pipelineId", pipeline.Id)
	return nil
}

func (impl *TriggerAuditHookImpl) AuditPreCdTrigger(workflowRunnerId int, pipeline *pipelineConfig.Pipeline,
	environment *repository4.Environment, workflowRequest *types.WorkflowRequest, triggerType string, triggeredBy int32, infraConfigTriggerHistoryId int) error {

	impl.logger.Infow("auditing Pre-CD trigger", "workflowRunnerId", workflowRunnerId, "pipelineId", pipeline.Id, "triggeredBy", triggeredBy)

	request := &bean.CdTriggerAuditRequest{
		WorkflowRunnerId: workflowRunnerId,
		Pipeline:         pipeline,
		Environment:      environment,
		CommonAuditRequest: &bean.CommonAuditRequest{
			WorkflowRequest:             workflowRequest,
			TriggerType:                 triggerType,
			TriggeredBy:                 triggeredBy,
			InfraConfigTriggerHistoryId: infraConfigTriggerHistoryId,
		},
	}

	_, err := impl.workflowTriggerAuditService.SavePreCdTriggerAudit(request)
	if err != nil {
		impl.logger.Errorw("error in auditing Pre-CD trigger", "err", err, "workflowRunnerId", workflowRunnerId)
		// Don't fail the trigger if audit fails, just log the error
		return nil
	}

	impl.logger.Infow("successfully audited Pre-CD trigger", "workflowRunnerId", workflowRunnerId, "pipelineId", pipeline.Id)
	return nil
}

func (impl *TriggerAuditHookImpl) AuditPostCdTrigger(workflowRunnerId int, pipeline *pipelineConfig.Pipeline,
	environment *repository4.Environment, workflowRequest *types.WorkflowRequest, triggerType string, triggeredBy int32, infraConfigTriggerHistoryId int) error {

	impl.logger.Infow("auditing Post-CD trigger", "workflowRunnerId", workflowRunnerId, "pipelineId", pipeline.Id, "triggeredBy", triggeredBy)

	request := &bean.CdTriggerAuditRequest{
		WorkflowRunnerId: workflowRunnerId,
		Pipeline:         pipeline,
		Environment:      environment,
		CommonAuditRequest: &bean.CommonAuditRequest{
			WorkflowRequest:             workflowRequest,
			TriggerType:                 triggerType,
			TriggeredBy:                 triggeredBy,
			InfraConfigTriggerHistoryId: infraConfigTriggerHistoryId,
		},
	}

	_, err := impl.workflowTriggerAuditService.SavePostCdTriggerAudit(request)
	if err != nil {
		impl.logger.Errorw("error in auditing Post-CD trigger", "err", err, "workflowRunnerId", workflowRunnerId)
		// Don't fail the trigger if audit fails, just log the error
		return nil
	}

	impl.logger.Infow("successfully audited Post-CD trigger", "workflowRunnerId", workflowRunnerId, "pipelineId", pipeline.Id)
	return nil
}

//
//func (impl *TriggerAuditHookImpl) GetRetriggerConfig(auditId int) (*bean.RetriggerWorkflowConfig, error) {
//	impl.logger.Infow("getting retrigger config", "auditId", auditId)
//
//	config, err := impl.workflowTriggerAuditService.GetWorkflowConfigForRetrigger(auditId)
//	if err != nil {
//		impl.logger.Errorw("error in getting retrigger config", "err", err, "auditId", auditId)
//		return nil, err
//	}
//
//	impl.logger.Infow("successfully retrieved retrigger config", "auditId", auditId, "workflowType", config.WorkflowType)
//	return config, nil
//}
