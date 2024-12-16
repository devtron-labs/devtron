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

package deploymentTemplate

import (
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeploymentTemplateHistoryService interface {
	CreateDeploymentTemplateHistoryFromGlobalTemplate(chart *chartRepoRepository.Chart, tx *pg.Tx, IsAppMetricsEnabled bool) error
	CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(envOverride *bean.EnvConfigOverride, tx *pg.Tx, IsAppMetricsEnabled bool, pipelineId int) error
	CreateDeploymentTemplateHistoryForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, envOverride *bean.EnvConfigOverride, renderedImageTemplate string, deployedOn time.Time, deployedBy int32) (*repository.DeploymentTemplateHistory, error)
}

type DeploymentTemplateHistoryServiceImpl struct {
	logger                              *zap.SugaredLogger
	deploymentTemplateHistoryRepository repository.DeploymentTemplateHistoryRepository
	pipelineRepository                  pipelineConfig.PipelineRepository
	chartRepository                     chartRepoRepository.ChartRepository
	userService                         user.UserService
	cdWorkflowRepository                pipelineConfig.CdWorkflowRepository
	scopedVariableManager               variables.ScopedVariableManager
	deployedAppMetricsService           deployedAppMetrics.DeployedAppMetricsService
	chartRefService                     chartRef.ChartRefService
}

func NewDeploymentTemplateHistoryServiceImpl(logger *zap.SugaredLogger, deploymentTemplateHistoryRepository repository.DeploymentTemplateHistoryRepository,
	pipelineRepository pipelineConfig.PipelineRepository, chartRepository chartRepoRepository.ChartRepository,
	userService user.UserService, cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	scopedVariableManager variables.ScopedVariableManager, deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService,
	chartRefService chartRef.ChartRefService) *DeploymentTemplateHistoryServiceImpl {
	return &DeploymentTemplateHistoryServiceImpl{
		logger:                              logger,
		deploymentTemplateHistoryRepository: deploymentTemplateHistoryRepository,
		pipelineRepository:                  pipelineRepository,
		chartRepository:                     chartRepository,
		userService:                         userService,
		cdWorkflowRepository:                cdWorkflowRepository,
		scopedVariableManager:               scopedVariableManager,
		deployedAppMetricsService:           deployedAppMetricsService,
		chartRefService:                     chartRefService,
	}
}

func (impl DeploymentTemplateHistoryServiceImpl) CreateDeploymentTemplateHistoryFromGlobalTemplate(chart *chartRepoRepository.Chart, tx *pg.Tx, IsAppMetricsEnabled bool) (err error) {
	//getting all pipelines without overridden charts
	pipelines, err := impl.pipelineRepository.FindAllPipelinesByChartsOverrideAndAppIdAndChartId(false, chart.AppId, chart.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting pipelines, CreateDeploymentTemplateHistoryFromGlobalTemplate", "err", err, "chart", chart)
		return err
	}
	chartRefDto, err := impl.chartRefService.FindById(chart.ChartRefId)
	if err != nil {
		impl.logger.Errorw("err in getting chartRef, CreateDeploymentTemplateHistoryFromGlobalTemplate", "err", err, "chart", chart)
		return err
	}
	//creating history without pipeline id
	historyModel := &repository.DeploymentTemplateHistory{
		AppId:                   chart.AppId,
		ImageDescriptorTemplate: chart.ImageDescriptorTemplate,
		Template:                chart.GlobalOverride,
		Deployed:                false,
		TemplateName:            chartRefDto.Name,
		TemplateVersion:         chartRefDto.Version,
		IsAppMetricsEnabled:     IsAppMetricsEnabled,
		AuditLog: sql.AuditLog{
			CreatedOn: chart.CreatedOn,
			CreatedBy: chart.CreatedBy,
			UpdatedOn: chart.UpdatedOn,
			UpdatedBy: chart.UpdatedBy,
		},
	}
	//creating new entry
	if tx != nil {
		_, err = impl.deploymentTemplateHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
	} else {
		_, err = impl.deploymentTemplateHistoryRepository.CreateHistory(historyModel)
	}
	if err != nil {
		impl.logger.Errorw("err in creating history entry for deployment template", "err", err, "history", historyModel)
		return err
	}
	for _, pipeline := range pipelines {
		historyModel := &repository.DeploymentTemplateHistory{
			AppId:                   chart.AppId,
			PipelineId:              pipeline.Id,
			ImageDescriptorTemplate: chart.ImageDescriptorTemplate,
			Template:                chart.GlobalOverride,
			Deployed:                false,
			TemplateName:            chartRefDto.Name,
			TemplateVersion:         chartRefDto.Version,
			IsAppMetricsEnabled:     IsAppMetricsEnabled,
			AuditLog: sql.AuditLog{
				CreatedOn: chart.CreatedOn,
				CreatedBy: chart.CreatedBy,
				UpdatedOn: chart.UpdatedOn,
				UpdatedBy: chart.UpdatedBy,
			},
		}
		//creating new entry
		if tx != nil {
			_, err = impl.deploymentTemplateHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
		} else {
			_, err = impl.deploymentTemplateHistoryRepository.CreateHistory(historyModel)
		}
		if err != nil {
			impl.logger.Errorw("err in creating history entry for deployment template", "err", err, "history", historyModel)
			return err
		}
	}
	return err
}

func (impl DeploymentTemplateHistoryServiceImpl) CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(envOverride *bean.EnvConfigOverride, tx *pg.Tx, IsAppMetricsEnabled bool, pipelineId int) (err error) {
	chart, err := impl.chartRepository.FindById(envOverride.ChartId)
	if err != nil {
		impl.logger.Errorw("err in getting global deployment template", "err", err, "chart", chart)
		return err
	}
	chartRefDto, err := impl.chartRefService.FindById(chart.ChartRefId)
	if err != nil {
		impl.logger.Errorw("err in getting chartRef, CreateDeploymentTemplateHistoryFromGlobalTemplate", "err", err, "chartRef", chartRefDto)
		return err
	}
	if pipelineId == 0 {
		pipeline, err := impl.pipelineRepository.GetByEnvOverrideIdAndEnvId(envOverride.Id, envOverride.TargetEnvironment)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("err in getting pipelines, CreateDeploymentTemplateHistoryFromEnvOverrideTemplate", "err", err, "envOverrideId", envOverride.Id)
			return err
		}
		pipelineId = pipeline.Id
	}
	historyModel := &repository.DeploymentTemplateHistory{
		AppId:                   chart.AppId,
		PipelineId:              pipelineId,
		ImageDescriptorTemplate: chart.ImageDescriptorTemplate,
		TargetEnvironment:       envOverride.TargetEnvironment,
		Deployed:                false,
		TemplateName:            chartRefDto.Name,
		TemplateVersion:         chartRefDto.Version,
		IsAppMetricsEnabled:     IsAppMetricsEnabled,
		AuditLog: sql.AuditLog{
			CreatedOn: envOverride.CreatedOn,
			CreatedBy: envOverride.CreatedBy,
			UpdatedOn: envOverride.UpdatedOn,
			UpdatedBy: envOverride.UpdatedBy,
		},
	}
	if envOverride.IsOverride {
		historyModel.Template = envOverride.EnvOverrideValues
	} else {
		//this is for the case when env override is created for new cd pipelines with template = "{}"
		historyModel.Template = chart.GlobalOverride
	}
	//creating new entry
	if tx != nil {
		_, err = impl.deploymentTemplateHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
	} else {
		_, err = impl.deploymentTemplateHistoryRepository.CreateHistory(historyModel)
	}
	if err != nil {
		impl.logger.Errorw("err in creating history entry for deployment template", "err", err, "history", historyModel)
		return err
	}
	return nil
}

func (impl DeploymentTemplateHistoryServiceImpl) CreateDeploymentTemplateHistoryForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, envOverride *bean.EnvConfigOverride, renderedImageTemplate string, deployedOn time.Time, deployedBy int32) (*repository.DeploymentTemplateHistory, error) {
	chartRefDto, err := impl.chartRefService.FindById(envOverride.Chart.ChartRefId)
	if err != nil {
		impl.logger.Errorw("err in getting chartRef, CreateDeploymentTemplateHistoryFromGlobalTemplate", "err", err, "chartRef", chartRefDto)
		return nil, err
	}
	isAppMetricsEnabled, err := impl.deployedAppMetricsService.GetMetricsFlagForAPipelineByAppIdAndEnvId(pipeline.AppId, pipeline.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error, GetMetricsFlagForAPipelineByAppIdAndEnvId", "err", err, "appId", pipeline.AppId, "envId", pipeline.EnvironmentId)
		return nil, err
	}
	historyModel := &repository.DeploymentTemplateHistory{
		AppId:                   pipeline.AppId,
		PipelineId:              pipeline.Id,
		TargetEnvironment:       pipeline.EnvironmentId,
		ImageDescriptorTemplate: renderedImageTemplate,
		Deployed:                true,
		DeployedBy:              deployedBy,
		DeployedOn:              deployedOn,
		TemplateName:            chartRefDto.Name,
		TemplateVersion:         chartRefDto.Version,
		IsAppMetricsEnabled:     isAppMetricsEnabled,
		MergeStrategy:           string(envOverride.MergeStrategy),
		AuditLog: sql.AuditLog{
			CreatedOn: deployedOn,
			CreatedBy: deployedBy,
			UpdatedOn: deployedOn,
			UpdatedBy: deployedBy,
		},
	}
	if envOverride.IsOverride {
		historyModel.Template = envOverride.EnvOverrideValues
	} else {
		historyModel.Template = envOverride.Chart.GlobalOverride
	}
	//creating new entry
	history, err := impl.deploymentTemplateHistoryRepository.CreateHistory(historyModel)
	if err != nil {
		impl.logger.Errorw("err in creating history entry for deployment template", "err", err, "history", historyModel)
		return nil, err
	}
	return history, nil
}
