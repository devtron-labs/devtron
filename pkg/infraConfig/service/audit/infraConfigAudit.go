/*
 * Copyright (c) 2024. Devtron Inc.
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

package audit

import (
	audit2 "github.com/devtron-labs/devtron/pkg/infraConfig/adapter/audit"
	infraBean "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository/audit"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strconv"
)

type InfraConfigAuditService interface {
	SaveCiInfraConfigHistorySnapshot(tx *pg.Tx, workflowId int, triggeredBy int32, infraConfigs map[string]*infraBean.InfraConfig) error
	GetInfraConfigByWorkflowId(workflowId int, workflowType string) (*infraBean.InfraConfig, error)
	sql.TransactionWrapper
}

type InfraConfigAuditServiceImpl struct {
	logger                     *zap.SugaredLogger
	infraConfigAuditRepository audit.InfraConfigAuditRepository
	*sql.TransactionUtilImpl
}

func NewInfraConfigAuditServiceImpl(logger *zap.SugaredLogger,
	infraConfigAuditRepository audit.InfraConfigAuditRepository,
	transactionUtilImpl *sql.TransactionUtilImpl) *InfraConfigAuditServiceImpl {
	return &InfraConfigAuditServiceImpl{
		logger:                     logger,
		infraConfigAuditRepository: infraConfigAuditRepository,
		TransactionUtilImpl:        transactionUtilImpl,
	}
}

func (impl *InfraConfigAuditServiceImpl) SaveCiInfraConfigHistorySnapshot(tx *pg.Tx,
	workflowId int, triggeredBy int32, infraConfigs map[string]*infraBean.InfraConfig) error {
	infraConfigTriggerAudits := make([]*audit.InfraConfigTriggerHistory, 0)
	for platform, infraConfig := range infraConfigs {
		infraConfigTriggerHistories, err := audit2.GetInfraConfigTriggerAudit(infraConfig)
		if err != nil {
			impl.logger.Errorw("failed to get infra config trigger audit", "error", err, "infraConfig", infraConfig)
			return err
		}
		for _, infraConfigTriggerHistory := range infraConfigTriggerHistories {
			infraConfigTriggerHistory = infraConfigTriggerHistory.
				WithPlatform(platform).WithWorkflowId(workflowId).
				WithWorkflowType(audit.CIWorkflowType).WithAuditLog(triggeredBy)
		}
		infraConfigTriggerAudits = append(infraConfigTriggerAudits, infraConfigTriggerHistories...)
	}
	impl.logger.Debugw("saving infra config history snapshot", "workflowId", workflowId,
		"infraConfigs", infraConfigs, "infraConfigTriggerAudits", infraConfigTriggerAudits)
	err := impl.infraConfigAuditRepository.SaveInfraConfigHistorySnapshot(tx, infraConfigTriggerAudits)
	if err != nil {
		impl.logger.Errorw("failed to save infra config history snapshot", "error", err, "infraConfigTriggerAudits", infraConfigTriggerAudits)
		return err
	}
	return nil
}

func (impl *InfraConfigAuditServiceImpl) GetInfraConfigByWorkflowId(workflowId int, workflowType string) (*infraBean.InfraConfig, error) {
	workflowTypeEnum := audit.WorkflowType(workflowType)
	infraConfigHistories, err := impl.infraConfigAuditRepository.GetInfraConfigHistoryByWorkflowId(workflowId, workflowTypeEnum)
	if err != nil {
		impl.logger.Errorw("failed to get infra config history by workflow id", "error", err, "workflowId", workflowId, "workflowType", workflowType)
		return nil, err
	}
	infraConfig, err := impl.fetchInfraConfigFromHistory(infraConfigHistories)
	if err != nil {
		impl.logger.Errorw("failed to fetch infra config from history", "infraConfigHistories", infraConfigHistories, "err", err)
		return nil, err
	}
	return infraConfig, nil
}

func (impl *InfraConfigAuditServiceImpl) fetchInfraConfigFromHistory(infraConfigHistories []*audit.InfraConfigTriggerHistory) (*infraBean.InfraConfig, error) {
	infraConfig := &infraBean.InfraConfig{}
	for _, history := range infraConfigHistories {
		switch history.Key {
		case infraBean.CPULimitKey:
			infraConfig.CiLimitCpu = history.ValueString
		case infraBean.CPURequestKey:
			infraConfig.CiReqCpu = history.ValueString
		case infraBean.MemoryLimitKey:
			infraConfig.CiLimitMem = history.ValueString
		case infraBean.MemoryRequestKey:
			infraConfig.CiReqMem = history.ValueString
		case infraBean.TimeOutKey:
			// Convert string back to float64
			if timeout, parseErr := strconv.ParseFloat(history.ValueString, 64); parseErr == nil {
				infraConfig.CiDefaultTimeout = timeout
			} else {
				impl.logger.Errorw("failed to parse timeout value", "valueString", history.ValueString, "parseErr", parseErr)
				return nil, parseErr
			}
		}
	}

	return infraConfig, nil
}
