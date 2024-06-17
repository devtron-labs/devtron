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
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	cron2 "github.com/devtron-labs/devtron/util/cron"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"strconv"
)

type CiStatusUpdateCron interface {
	UpdateCiWorkflowStatusFailedCron()
}

type CiStatusUpdateCronImpl struct {
	logger                       *zap.SugaredLogger
	cron                         *cron.Cron
	appService                   app.AppService
	ciWorkflowStatusUpdateConfig *CiWorkflowStatusUpdateConfig
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	ciHandler                    pipeline.CiHandler
}

func NewCiStatusUpdateCronImpl(logger *zap.SugaredLogger, appService app.AppService,
	ciWorkflowStatusUpdateConfig *CiWorkflowStatusUpdateConfig, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	ciHandler pipeline.CiHandler, cronLogger *cron2.CronLoggerImpl) *CiStatusUpdateCronImpl {
	cron := cron.New(
		cron.WithChain(cron.Recover(cronLogger)))
	cron.Start()
	impl := &CiStatusUpdateCronImpl{
		logger:                       logger,
		cron:                         cron,
		appService:                   appService,
		ciWorkflowStatusUpdateConfig: ciWorkflowStatusUpdateConfig,
		ciPipelineRepository:         ciPipelineRepository,
		ciHandler:                    ciHandler,
	}

	// execute periodically, update ci workflow status for failed process
	_, err := cron.AddFunc(ciWorkflowStatusUpdateConfig.CiWorkflowStatusUpdateCron, impl.UpdateCiWorkflowStatusFailedCron)
	if err != nil {
		logger.Errorw("error while configure cron job for ci workflow status update", "err", err)
		return impl
	}
	return impl
}

type CiWorkflowStatusUpdateConfig struct {
	CiWorkflowStatusUpdateCron string `env:"CI_WORKFLOW_STATUS_UPDATE_CRON" envDefault:"*/5 * * * *"`
	TimeoutForFailedCiBuild    string `env:"TIMEOUT_FOR_FAILED_CI_BUILD" envDefault:"15"` //in minutes
}

func GetCiWorkflowStatusUpdateConfig() (*CiWorkflowStatusUpdateConfig, error) {
	cfg := &CiWorkflowStatusUpdateConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse ci workflow status update config: " + err.Error())
		return nil, err
	}
	return cfg, nil
}

// UpdateCiWorkflowStatusFailedCron this function will execute periodically
func (impl *CiStatusUpdateCronImpl) UpdateCiWorkflowStatusFailedCron() {
	timeoutForFailureCiBuild, err := strconv.Atoi(impl.ciWorkflowStatusUpdateConfig.TimeoutForFailedCiBuild)
	if err != nil {
		impl.logger.Errorw("error in converting string to int", "err", err)
		return
	}
	err = impl.ciHandler.UpdateCiWorkflowStatusFailure(timeoutForFailureCiBuild)
	if err != nil {
		impl.logger.Errorw("error in updating ci workflow status for failed workflows", "err", err)
		return
	}
	return
}
