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

package ci

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/service"
	"github.com/devtron-labs/devtron/pkg/infraConfig/service/audit"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders/infraGetters"
	"go.uber.org/zap"
)

// InfraGetter gets infra config for ci workflows
type InfraGetter struct {
	logger                  *zap.SugaredLogger
	infraConfigService      service.InfraConfigService
	infraConfigAuditService audit.InfraConfigAuditService
}

func NewCiInfraGetter(logger *zap.SugaredLogger,
	infraConfigService service.InfraConfigService,
	infraConfigAuditService audit.InfraConfigAuditService) *InfraGetter {
	return &InfraGetter{
		logger:                  logger,
		infraConfigService:      infraConfigService,
		infraConfigAuditService: infraConfigAuditService,
	}
}

// GetConfigurationsByScopeAndTargetPlatforms gets infra config for ci workflows using the scope
func (ciInfraGetter *InfraGetter) GetConfigurationsByScopeAndTargetPlatforms(request *infraGetters.InfraRequest) (map[string]*v1.InfraConfig, error) {
	return ciInfraGetter.infraConfigService.GetConfigurationsByScopeAndTargetPlatforms(request.GetWorkflowScope(), request.GetTargetPlatforms())
}

func (ciInfraGetter *InfraGetter) SaveInfraConfigHistorySnapshot(workflowId int, triggeredBy int32, infraConfigs map[string]*v1.InfraConfig) error {
	tx, err := ciInfraGetter.infraConfigAuditService.StartTx()
	if err != nil {
		ciInfraGetter.logger.Errorw("error in starting the transaction", "err", err)
		return err
	}
	defer ciInfraGetter.infraConfigAuditService.RollbackTx(tx)
	err = ciInfraGetter.infraConfigAuditService.SaveCiInfraConfigHistorySnapshot(tx, workflowId, triggeredBy, infraConfigs)
	if err != nil {
		ciInfraGetter.logger.Errorw("error in creating ci infra trigger snapshot", "infraConfigs", infraConfigs, "err", err)
		return err
	}
	err = ciInfraGetter.infraConfigService.HandleInfraConfigTriggerAudit(workflowId, triggeredBy, infraConfigs)
	if err != nil {
		ciInfraGetter.logger.Errorw("error in handling infra config trigger audit", "infraConfigs", infraConfigs, "err", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		ciInfraGetter.logger.Errorw("err in committing transaction", "err", err)
		return err
	}
	return nil
}
