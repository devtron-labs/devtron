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

package cron

import (
	"fmt"
	"github.com/caarlos0/env"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/build/pipeline/bean/common"
	"github.com/devtron-labs/devtron/pkg/build/trigger"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	repository3 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	cron2 "github.com/devtron-labs/devtron/util/cron"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type CiTriggerCron interface {
	TriggerCiCron()
}

type CiTriggerCronImpl struct {
	logger                  *zap.SugaredLogger
	cron                    *cron.Cron
	cfg                     *CiTriggerCronConfig
	pipelineStageRepository repository.PipelineStageRepository
	ciArtifactRepository    repository2.CiArtifactRepository
	globalPluginRepository  repository3.GlobalPluginRepository
	ciHandlerService        trigger.HandlerService
}

func NewCiTriggerCronImpl(logger *zap.SugaredLogger, cfg *CiTriggerCronConfig, pipelineStageRepository repository.PipelineStageRepository,
	ciArtifactRepository repository2.CiArtifactRepository, globalPluginRepository repository3.GlobalPluginRepository, cronLogger *cron2.CronLoggerImpl,
	ciHandlerService trigger.HandlerService) *CiTriggerCronImpl {
	cron := cron.New(
		cron.WithChain(cron.Recover(cronLogger)))
	cron.Start()
	impl := &CiTriggerCronImpl{
		logger:                  logger,
		cron:                    cron,
		pipelineStageRepository: pipelineStageRepository,
		cfg:                     cfg,
		ciArtifactRepository:    ciArtifactRepository,
		globalPluginRepository:  globalPluginRepository,
		ciHandlerService:        ciHandlerService,
	}

	_, err := cron.AddFunc(fmt.Sprintf("@every %dm", cfg.SourceControllerCronTime), impl.TriggerCiCron)
	if err != nil {
		logger.Errorw("error while configure cron job for ci workflow status update", "err", err)
		return impl
	}
	return impl
}

type CiTriggerCronConfig struct {
	SourceControllerCronTime int    `env:"CI_TRIGGER_CRON_TIME" envDefault:"2" description:"For image poll plugin"`
	PluginName               string `env:"PLUGIN_NAME"  envDefault:"Pull images from container repository" description:"Handles image retrieval from a container repository and triggers subsequent CI processes upon detecting new images.Current default plugin name: Pull Images from Container Repository."`
}

func GetCiTriggerCronConfig() (*CiTriggerCronConfig, error) {
	cfg := &CiTriggerCronConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse ci trigger cron config: " + err.Error())
		return nil, err
	}
	return cfg, nil
}

// UpdateCiWorkflowStatusFailedCron this function will execute periodically
func (impl *CiTriggerCronImpl) TriggerCiCron() {
	plugin, err := impl.globalPluginRepository.GetPluginByName(impl.cfg.PluginName)
	if err != nil || len(plugin) == 0 {
		return
	}

	ciPipelineIds, err := impl.pipelineStageRepository.GetAllCiPipelineIdsByPluginIdAndStageType(plugin[0].Id, string(repository.PIPELINE_STAGE_TYPE_PRE_CI))
	if err != nil {
		return
	}
	for _, ciPipelineId := range ciPipelineIds {
		var ciPipelineMaterials []bean.CiPipelineMaterial
		ciTriggerRequest := bean.CiTriggerRequest{
			PipelineId:         ciPipelineId,
			CiPipelineMaterial: ciPipelineMaterials,
			TriggeredBy:        bean2.SYSTEM_USER_ID,
			InvalidateCache:    false,
			PipelineType:       string(common.CI_JOB),
		}
		_, err = impl.ciHandlerService.HandleCIManual(ciTriggerRequest)
		if err != nil {
			return
		}
	}
	return
}
