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
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/service"
	"go.uber.org/zap"
)

// TriggerAuditHook provides common hooks for auditing workflow triggers
type TriggerAuditHook interface {
	// AuditCiTrigger audits CI trigger
	AuditCiTrigger(ciTriggerAuditRequest *bean.CiTriggerAuditRequest) error

	// AuditPrePostCdTrigger audits Pre-CD trigger
	AuditPrePostCdTrigger(cdTriggerAuditRequest *bean.CdTriggerAuditRequest) error

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

func (impl *TriggerAuditHookImpl) AuditCiTrigger(ciTriggerAuditRequest *bean.CiTriggerAuditRequest) error {

	impl.logger.Infow("auditing CI trigger", "workflowId", ciTriggerAuditRequest.WorkflowId, "pipelineId", ciTriggerAuditRequest.PipelineId, "triggeredBy", ciTriggerAuditRequest.TriggeredBy)

	//request := &bean.CiTriggerAuditRequest{
	//	WorkflowId: workflowId,
	//	Pipeline:   pipeline,
	//	CommonAuditRequest: &bean.CommonAuditRequest{
	//		WorkflowRequest: workflowRequest,
	//		TriggerType:     triggerType,
	//		TriggeredBy:     triggeredBy,
	//	},
	//}

	_, err := impl.workflowTriggerAuditService.SaveCiTriggerAudit(ciTriggerAuditRequest)
	if err != nil {
		impl.logger.Errorw("error in auditing CI trigger", "workflowId", ciTriggerAuditRequest.WorkflowId, "err", err)
		// Don't fail/return the trigger if audit fails, just log the error
	}

	impl.logger.Infow("successfully audited CI trigger", "workflowId", ciTriggerAuditRequest.WorkflowId, "pipelineId", ciTriggerAuditRequest.PipelineId)
	return nil
}

func (impl *TriggerAuditHookImpl) AuditPrePostCdTrigger(cdTriggerAuditRequest *bean.CdTriggerAuditRequest) error {

	impl.logger.Infow("auditing Pre/Post-CD trigger", "workflowRunnerId", cdTriggerAuditRequest.WorkflowRunnerId, "pipelineId", cdTriggerAuditRequest.PipelineId, "workflowType", cdTriggerAuditRequest.WorkflowType, "triggeredBy", cdTriggerAuditRequest.TriggeredBy)

	//request := &bean.CdTriggerAuditRequest{
	//	WorkflowRunnerId: workflowRunnerId,
	//	Pipeline:         pipeline,
	//	Environment:      environment,
	//	WorkflowType:     workflowType,
	//	CommonAuditRequest: &bean.CommonAuditRequest{
	//		WorkflowRequest: workflowRequest,
	//		TriggerType:     triggerType,
	//		TriggeredBy:     triggeredBy,
	//	},
	//}

	_, err := impl.workflowTriggerAuditService.SaveCdTriggerAudit(cdTriggerAuditRequest)
	if err != nil {
		impl.logger.Errorw("error in auditing Pre/Post-CD trigger", "workflowRunnerId", cdTriggerAuditRequest.WorkflowRunnerId, "err", err)
		// Don't fail/return the trigger if audit fails, just log the error
		return nil
	}

	impl.logger.Infow("successfully audited Pre/Post-CD trigger", "workflowRunnerId", cdTriggerAuditRequest.WorkflowRunnerId, "pipelineId", cdTriggerAuditRequest.PipelineId)
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
