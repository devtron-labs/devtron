/*
 * Copyright (c) 2024. Devtron Inc.
 */

package job

import (
	"github.com/caarlos0/env"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/config/read"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/service/audit"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders/infraGetters"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"go.uber.org/zap"
)

// InfraGetter gets the infra config for job workflows
type InfraGetter struct {
	logger                  *zap.SugaredLogger
	jobInfra                v1.InfraConfig
	configMapService        read.ConfigReadService
	infraConfigAuditService audit.InfraConfigAuditService
}

func NewJobInfraGetter(logger *zap.SugaredLogger,
	configMapService read.ConfigReadService,
	infraConfigAuditService audit.InfraConfigAuditService) (*InfraGetter, error) {
	infra := v1.InfraConfig{}
	err := env.Parse(&infra)
	return &InfraGetter{
		logger:                  logger,
		jobInfra:                infra,
		configMapService:        configMapService,
		infraConfigAuditService: infraConfigAuditService,
	}, err
}

// GetConfigurationsByScopeAndTargetPlatforms gets infra config for ci workflows using the Scope
func (jobInfraGetter *InfraGetter) GetConfigurationsByScopeAndTargetPlatforms(request *infraGetters.InfraRequest) (map[string]*v1.InfraConfig, error) {
	response := make(map[string]*v1.InfraConfig)
	infra := jobInfraGetter.jobInfra
	configMaps, secrets, err := jobInfraGetter.getCmCsForPrePostStageTrigger(request.GetWorkflowScope(), request.GetAppId(), request.GetEnvId())
	if err != nil {
		jobInfraGetter.logger.Errorw("error getting cm/cs for job", "request", request, "error", err)
		return response, err
	}
	infra.ConfigMaps = configMaps.Maps
	infra.Secrets = sliceUtil.GetDeReferencedSlice(secrets.Secrets)
	response[v1.RUNNER_PLATFORM] = &infra
	return response, nil
}

func (jobInfraGetter *InfraGetter) SaveInfraConfigHistorySnapshot(workflowId int, triggeredBy int32, infraConfigs map[string]*v1.InfraConfig) error {
	tx, err := jobInfraGetter.infraConfigAuditService.StartTx()
	if err != nil {
		jobInfraGetter.logger.Errorw("error in starting the transaction", "err", err)
		return err
	}
	defer jobInfraGetter.infraConfigAuditService.RollbackTx(tx)
	err = jobInfraGetter.infraConfigAuditService.SaveCiInfraConfigHistorySnapshot(tx, workflowId, triggeredBy, infraConfigs)
	if err != nil {
		jobInfraGetter.logger.Errorw("error in creating ci infra trigger snapshot", "infraConfigs", infraConfigs, "err", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		jobInfraGetter.logger.Errorw("err in committing transaction", "err", err)
		return err
	}
	return nil
}

func (jobInfraGetter *InfraGetter) getCmCsForPrePostStageTrigger(scope resourceQualifiers.Scope, appId int, envId int) (*apiBean.ConfigMapJson, *apiBean.ConfigSecretJson, error) {
	return jobInfraGetter.configMapService.GetCmCsForPrePostStageTrigger(scope, appId, envId, true)
}
