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

package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	application3 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	k8sUtil "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/bean"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/manifest/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/helper"
	"github.com/devtron-labs/devtron/pkg/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/imageDigestPolicy"
	"github.com/devtron-labs/devtron/pkg/k8s"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	repository5 "github.com/devtron-labs/devtron/pkg/variables/repository"
	util4 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	errors2 "github.com/juju/errors"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type ManifestCreationService interface {
	BuildManifestForTrigger(overrideRequest *bean.ValuesOverrideRequest, triggeredAt time.Time,
		ctx context.Context) (valuesOverrideResponse *app.ValuesOverrideResponse, builtChartPath string, err error)

	//TODO: remove below method
	GetValuesOverrideForTrigger(overrideRequest *bean.ValuesOverrideRequest, triggeredAt time.Time, ctx context.Context) (*app.ValuesOverrideResponse, error)
}

type ManifestCreationServiceImpl struct {
	logger                         *zap.SugaredLogger
	dockerRegistryIpsConfigService dockerRegistry.DockerRegistryIpsConfigService
	chartRefService                chartRef.ChartRefService
	scopedVariableManager          variables.ScopedVariableCMCSManager
	k8sCommonService               k8s.K8sCommonService
	deployedAppMetricsService      deployedAppMetrics.DeployedAppMetricsService
	imageDigestPolicyService       imageDigestPolicy.ImageDigestPolicyService
	mergeUtil                      *util.MergeUtil
	appCrudOperationService        app.AppCrudOperationService
	deploymentTemplateService      deploymentTemplate.DeploymentTemplateService

	acdClient application2.ServiceClient //TODO: replace with argoClientWrapperService

	configMapHistoryRepository          repository3.ConfigMapHistoryRepository
	configMapRepository                 chartConfig.ConfigMapRepository
	chartRepository                     chartRepoRepository.ChartRepository
	environmentConfigRepository         chartConfig.EnvConfigOverrideRepository
	envRepository                       repository2.EnvironmentRepository
	pipelineRepository                  pipelineConfig.PipelineRepository
	ciArtifactRepository                repository.CiArtifactRepository
	pipelineOverrideRepository          chartConfig.PipelineOverrideRepository
	strategyHistoryRepository           repository3.PipelineStrategyHistoryRepository
	pipelineConfigRepository            chartConfig.PipelineConfigRepository
	deploymentTemplateHistoryRepository repository3.DeploymentTemplateHistoryRepository
	deploymentConfigService             common.DeploymentConfigService
}

func NewManifestCreationServiceImpl(logger *zap.SugaredLogger,
	dockerRegistryIpsConfigService dockerRegistry.DockerRegistryIpsConfigService,
	chartRefService chartRef.ChartRefService,
	scopedVariableManager variables.ScopedVariableCMCSManager,
	k8sCommonService k8s.K8sCommonService,
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService,
	imageDigestPolicyService imageDigestPolicy.ImageDigestPolicyService,
	mergeUtil *util.MergeUtil,
	appCrudOperationService app.AppCrudOperationService,
	deploymentTemplateService deploymentTemplate.DeploymentTemplateService,
	acdClient application2.ServiceClient,
	configMapHistoryRepository repository3.ConfigMapHistoryRepository,
	configMapRepository chartConfig.ConfigMapRepository,
	chartRepository chartRepoRepository.ChartRepository,
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository,
	envRepository repository2.EnvironmentRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	ciArtifactRepository repository.CiArtifactRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	strategyHistoryRepository repository3.PipelineStrategyHistoryRepository,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	deploymentTemplateHistoryRepository repository3.DeploymentTemplateHistoryRepository,
	deploymentConfigService common.DeploymentConfigService) *ManifestCreationServiceImpl {
	return &ManifestCreationServiceImpl{
		logger:                              logger,
		dockerRegistryIpsConfigService:      dockerRegistryIpsConfigService,
		chartRefService:                     chartRefService,
		scopedVariableManager:               scopedVariableManager,
		k8sCommonService:                    k8sCommonService,
		deployedAppMetricsService:           deployedAppMetricsService,
		imageDigestPolicyService:            imageDigestPolicyService,
		mergeUtil:                           mergeUtil,
		appCrudOperationService:             appCrudOperationService,
		deploymentTemplateService:           deploymentTemplateService,
		configMapRepository:                 configMapRepository,
		acdClient:                           acdClient,
		configMapHistoryRepository:          configMapHistoryRepository,
		chartRepository:                     chartRepository,
		environmentConfigRepository:         environmentConfigRepository,
		envRepository:                       envRepository,
		pipelineRepository:                  pipelineRepository,
		ciArtifactRepository:                ciArtifactRepository,
		pipelineOverrideRepository:          pipelineOverrideRepository,
		strategyHistoryRepository:           strategyHistoryRepository,
		pipelineConfigRepository:            pipelineConfigRepository,
		deploymentTemplateHistoryRepository: deploymentTemplateHistoryRepository,
		deploymentConfigService:             deploymentConfigService,
	}
}

func (impl *ManifestCreationServiceImpl) BuildManifestForTrigger(overrideRequest *bean.ValuesOverrideRequest, triggeredAt time.Time,
	ctx context.Context) (valuesOverrideResponse *app.ValuesOverrideResponse, builtChartPath string, err error) {
	valuesOverrideResponse = &app.ValuesOverrideResponse{}
	valuesOverrideResponse, err = impl.GetValuesOverrideForTrigger(overrideRequest, triggeredAt, ctx)
	if err != nil {
		impl.logger.Errorw("error in fetching values for trigger", "err", err)
		return valuesOverrideResponse, "", err
	}
	builtChartPath, err = impl.deploymentTemplateService.BuildChartAndGetPath(overrideRequest.AppName, valuesOverrideResponse.EnvOverride, ctx)
	if err != nil {
		impl.logger.Errorw("error in parsing reference chart", "err", err)
		return valuesOverrideResponse, "", err
	}
	return valuesOverrideResponse, builtChartPath, err
}

func (impl *ManifestCreationServiceImpl) GetValuesOverrideForTrigger(overrideRequest *bean.ValuesOverrideRequest, triggeredAt time.Time, ctx context.Context) (*app.ValuesOverrideResponse, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "ManifestCreationServiceImpl.GetValuesOverrideForTrigger")
	defer span.End()
	helper.ResolveDeploymentTypeAndUpdate(overrideRequest)
	valuesOverrideResponse := &app.ValuesOverrideResponse{}
	isPipelineOverrideCreated := overrideRequest.PipelineOverrideId > 0
	pipeline, err := impl.pipelineRepository.FindById(overrideRequest.PipelineId)
	valuesOverrideResponse.Pipeline = pipeline
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline by pipeline id", "err", err, "pipeline-id-", overrideRequest.PipelineId)
		return valuesOverrideResponse, err
	}
	// TODO: refactor the tracer
	_, span = otel.Tracer("orchestrator").Start(newCtx, "ciArtifactRepository.Get")
	artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
	valuesOverrideResponse.Artifact = artifact
	span.End()
	if err != nil {
		return valuesOverrideResponse, err
	}
	overrideRequest.Image = artifact.Image
	// Currently strategy is used only for history creation; hence it's not required in PipelineConfigOverride
	strategy, err := impl.getDeploymentStrategyByTriggerType(overrideRequest, newCtx)
	valuesOverrideResponse.PipelineStrategy = strategy
	if err != nil {
		impl.logger.Errorw("error in getting strategy by trigger type", "err", err)
		return valuesOverrideResponse, err
	}

	var (
		pipelineOverride                *chartConfig.PipelineOverride
		configMapJson, appLabelJsonByte []byte
		envOverride                     *chartConfig.EnvConfigOverride
	)
	if isPipelineOverrideCreated {
		pipelineOverride, err = impl.pipelineOverrideRepository.FindById(overrideRequest.PipelineOverrideId)
		if err != nil {
			impl.logger.Errorw("error in getting pipelineOverride for valuesOverrideResponse", "PipelineOverrideId", overrideRequest.PipelineOverrideId)
			return nil, err
		}
		envOverride, err = impl.environmentConfigRepository.GetByIdIncludingInactive(pipelineOverride.EnvConfigOverrideId)
		if err != nil {
			impl.logger.Errorw("error in getting env override by id", "id", pipelineOverride.EnvConfigOverrideId, "err", err)
			return valuesOverrideResponse, err
		}
		err = impl.setEnvironmentModelInEnvOverride(ctx, envOverride)
		if err != nil {
			impl.logger.Errorw("error while setting environment data in envOverride", "env", envOverride.TargetEnvironment, "err", err)
			return nil, err
		}
	} else {
		envOverride, err = impl.getEnvOverrideByTriggerType(overrideRequest, triggeredAt, newCtx)
		if err != nil {
			impl.logger.Errorw("error in getting env override by trigger type", "err", err)
			return valuesOverrideResponse, err
		}
	}
	valuesOverrideResponse.EnvOverride = envOverride

	// Conditional Block based on PipelineOverrideCreated --> start
	if !isPipelineOverrideCreated {
		pipelineOverride, err = impl.savePipelineOverride(newCtx, overrideRequest, envOverride.Id, triggeredAt)
		if err != nil {
			return valuesOverrideResponse, err
		}
		overrideRequest.PipelineOverrideId = pipelineOverride.Id
		pipelineOverride.Pipeline = pipeline
		pipelineOverride.CiArtifact = artifact
	}
	// Conditional Block based on PipelineOverrideCreated --> end
	valuesOverrideResponse.PipelineOverride = pipelineOverride

	appMetrics, err := impl.getAppMetricsByTriggerType(overrideRequest, newCtx)
	if err != nil {
		impl.logger.Errorw("error in getting app metrics by trigger type", "err", err)
		return valuesOverrideResponse, err
	}
	//TODO: check status and apply lock
	releaseOverrideJson, err := impl.getReleaseOverride(envOverride, overrideRequest, artifact, pipelineOverride.PipelineReleaseCounter, strategy, &appMetrics)
	valuesOverrideResponse.ReleaseOverrideJSON = releaseOverrideJson
	if err != nil {
		return valuesOverrideResponse, err
	}

	// Conditional Block based on PipelineOverrideCreated --> start
	if !isPipelineOverrideCreated {
		chartVersion := envOverride.Chart.ChartVersion
		scope := helper.GetScopeForVariables(overrideRequest, envOverride)
		request := helper.CreateConfigMapAndSecretJsonRequest(overrideRequest, envOverride, chartVersion, scope)

		configMapJson, err = impl.getConfigMapAndSecretJsonV2(newCtx, request, envOverride)
		if err != nil {
			impl.logger.Errorw("error in fetching config map n secret ", "err", err)
			configMapJson = nil
		}
		appLabelJsonByte, err = impl.appCrudOperationService.GetAppLabelsForDeployment(newCtx, overrideRequest.AppId, overrideRequest.AppName, overrideRequest.EnvName)
		if err != nil {
			impl.logger.Errorw("error in fetching app labels for gitOps commit", "err", err)
			appLabelJsonByte = nil
		}
		mergedValues, err := impl.mergeOverrideValues(envOverride, releaseOverrideJson, configMapJson, appLabelJsonByte, strategy)
		appName := fmt.Sprintf("%s-%s", overrideRequest.AppName, envOverride.Environment.Name)
		mergedValues, err = impl.autoscalingCheckBeforeTrigger(newCtx, appName, envOverride.Namespace, mergedValues, overrideRequest)
		if err != nil {
			impl.logger.Errorw("error in autoscaling check before trigger", "pipelineId", overrideRequest.PipelineId, "err", err)
			return valuesOverrideResponse, err
		}

		// handle image pull secret if access given
		mergedValues, err = impl.dockerRegistryIpsConfigService.HandleImagePullSecretOnApplicationDeployment(newCtx, envOverride.Environment, artifact, pipeline.CiPipelineId, mergedValues)
		if err != nil {
			return valuesOverrideResponse, err
		}

		valuesOverrideResponse.MergedValues = string(mergedValues)
		err = impl.pipelineOverrideRepository.UpdatePipelineMergedValues(newCtx, nil, pipelineOverride.Id, string(mergedValues), overrideRequest.UserId)
		if err != nil {
			return valuesOverrideResponse, err
		}
		pipelineOverride.PipelineMergedValues = string(mergedValues)
		valuesOverrideResponse.PipelineOverride = pipelineOverride
	} else {
		valuesOverrideResponse.MergedValues = pipelineOverride.PipelineMergedValues
	}
	// Conditional Block based on PipelineOverrideCreated --> end
	return valuesOverrideResponse, err
}

func (impl *ManifestCreationServiceImpl) getDeploymentStrategyByTriggerType(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context) (*chartConfig.PipelineStrategy, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "ManifestCreationServiceImpl.getDeploymentStrategyByTriggerType")
	defer span.End()
	strategy := &chartConfig.PipelineStrategy{}
	var err error
	if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
		strategyHistory, err := impl.strategyHistoryRepository.GetHistoryByPipelineIdAndWfrId(newCtx, overrideRequest.PipelineId, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
		if err != nil {
			impl.logger.Errorw("error in getting deployed strategy history by pipelineId and wfrId", "err", err, "pipelineId", overrideRequest.PipelineId, "wfrId", overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
			return nil, err
		}
		strategy.Strategy = strategyHistory.Strategy
		strategy.Config = strategyHistory.Config
		strategy.PipelineId = overrideRequest.PipelineId
	} else if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
		deploymentTemplateType := helper.GetDeploymentTemplateType(overrideRequest)
		if overrideRequest.ForceTrigger || len(deploymentTemplateType) == 0 {
			_, span := otel.Tracer("orchestrator").Start(newCtx, "pipelineConfigRepository.GetDefaultStrategyByPipelineId")
			strategy, err = impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(overrideRequest.PipelineId)
			span.End()
		} else {
			_, span := otel.Tracer("orchestrator").Start(newCtx, "pipelineConfigRepository.FindByStrategyAndPipelineId")
			strategy, err = impl.pipelineConfigRepository.FindByStrategyAndPipelineId(deploymentTemplateType, overrideRequest.PipelineId)
			span.End()
		}
		if err != nil && errors2.IsNotFound(err) == false {
			impl.logger.Errorf("invalid state", "err", err, "req", strategy)
			return nil, err
		}
	}
	return strategy, nil
}

func (impl *ManifestCreationServiceImpl) getEnvOverrideByTriggerType(overrideRequest *bean.ValuesOverrideRequest, triggeredAt time.Time, ctx context.Context) (*chartConfig.EnvConfigOverride, error) {
	envOverride := &chartConfig.EnvConfigOverride{}
	var err error
	if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
		envOverride, err = impl.getEnvOverrideForSpecificConfigTrigger(overrideRequest, ctx)
		if err != nil {
			impl.logger.Errorw("error, getEnvOverrideForSpecificConfigTrigger", "err", err, "overrideRequest", overrideRequest)
			return nil, err
		}
	} else if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
		envOverride, err = impl.getEnvOverrideForLastSavedConfigTrigger(overrideRequest, triggeredAt, ctx)
		if err != nil {
			impl.logger.Errorw("error, getEnvOverrideForLastSavedConfigTrigger", "err", err, "overrideRequest", overrideRequest)
			return nil, err
		}
	}
	return envOverride, nil
}

func (impl *ManifestCreationServiceImpl) getEnvOverrideForSpecificConfigTrigger(overrideRequest *bean.ValuesOverrideRequest,
	ctx context.Context) (*chartConfig.EnvConfigOverride, error) {
	envOverride := &chartConfig.EnvConfigOverride{}
	var err error
	_, span := otel.Tracer("orchestrator").Start(ctx, "deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId")
	deploymentTemplateHistory, err := impl.deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId(overrideRequest.PipelineId, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
	//VARIABLE_SNAPSHOT_GET and resolve
	span.End()
	if err != nil {
		impl.logger.Errorw("error in getting deployed deployment template history by pipelineId and wfrId", "err", err, "pipelineId", &overrideRequest, "wfrId", overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
		return nil, err
	}
	templateName := deploymentTemplateHistory.TemplateName
	templateVersion := deploymentTemplateHistory.TemplateVersion
	if templateName == "Rollout Deployment" {
		templateName = ""
	}
	//getting chart_ref by id
	_, span = otel.Tracer("orchestrator").Start(ctx, "chartRefRepository.FindByVersionAndName")
	chartRefDto, err := impl.chartRefService.FindByVersionAndName(templateVersion, templateName)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in getting chartRef by version and name", "err", err, "version", templateVersion, "name", templateName)
		return nil, err
	}
	//assuming that if a chartVersion is deployed then it's envConfigOverride will be available
	_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.GetByAppIdEnvIdAndChartRefId")
	envOverride, err = impl.environmentConfigRepository.GetByAppIdEnvIdAndChartRefId(overrideRequest.AppId, overrideRequest.EnvId, chartRefDto.Id)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in getting envConfigOverride for pipeline for specific chartVersion", "err", err, "appId", overrideRequest.AppId, "envId", overrideRequest.EnvId, "chartRefId", chartRefDto.Id)
		return nil, err
	}

	err = impl.setEnvironmentModelInEnvOverride(ctx, envOverride)
	if err != nil {
		impl.logger.Errorw("error while setting environment data in envOverride", "env", envOverride.TargetEnvironment, "err", err)
		return nil, err
	}
	//updating historical data in envConfigOverride and appMetrics flag
	envOverride.IsOverride = true
	envOverride.EnvOverrideValues = deploymentTemplateHistory.Template
	reference := repository5.HistoryReference{
		HistoryReferenceId:   deploymentTemplateHistory.Id,
		HistoryReferenceType: repository5.HistoryReferenceTypeDeploymentTemplate,
	}
	variableMap, resolvedTemplate, err := impl.scopedVariableManager.GetVariableSnapshotAndResolveTemplate(envOverride.EnvOverrideValues, parsers.JsonVariableTemplate, reference, true, false)
	envOverride.ResolvedEnvOverrideValues = resolvedTemplate
	envOverride.VariableSnapshot = variableMap
	if err != nil {
		impl.logger.Errorw("error, GetVariableSnapshotAndResolveTemplate", "err", err, "envOverride", envOverride)
		return envOverride, err
	}
	return envOverride, nil
}

func (impl *ManifestCreationServiceImpl) getEnvOverrideForLastSavedConfigTrigger(overrideRequest *bean.ValuesOverrideRequest,
	triggeredAt time.Time, ctx context.Context) (*chartConfig.EnvConfigOverride, error) {
	envOverride := &chartConfig.EnvConfigOverride{}
	var err error
	_, span := otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.ActiveEnvConfigOverride")
	envOverride, err = impl.environmentConfigRepository.ActiveEnvConfigOverride(overrideRequest.AppId, overrideRequest.EnvId)
	var chart *chartRepoRepository.Chart
	span.End()
	if err != nil {
		impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
		return nil, err
	}
	if envOverride.Id == 0 {
		_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindLatestChartForAppByAppId")
		chart, err = impl.chartRepository.FindLatestChartForAppByAppId(overrideRequest.AppId)
		span.End()
		if err != nil {
			impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
			return nil, err
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.FindChartByAppIdAndEnvIdAndChartRefId")
		envOverride, err = impl.environmentConfigRepository.FindChartByAppIdAndEnvIdAndChartRefId(overrideRequest.AppId, overrideRequest.EnvId, chart.ChartRefId)
		span.End()
		if err != nil && !errors2.IsNotFound(err) {
			impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
			return nil, err
		}

		//creating new env override config
		if errors2.IsNotFound(err) || envOverride == nil {
			_, span = otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
			environment, err := impl.envRepository.FindById(overrideRequest.EnvId)
			span.End()
			if err != nil && !util.IsErrNoRows(err) {
				return nil, err
			}
			envOverride = &chartConfig.EnvConfigOverride{
				Active:            true,
				ManualReviewed:    true,
				Status:            models.CHARTSTATUS_SUCCESS,
				TargetEnvironment: overrideRequest.EnvId,
				ChartId:           chart.Id,
				AuditLog:          sql.AuditLog{UpdatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId},
				Namespace:         environment.Namespace,
				IsOverride:        false,
				EnvOverrideValues: "{}",
				Latest:            false,
				IsBasicViewLocked: chart.IsBasicViewLocked,
				CurrentViewEditor: chart.CurrentViewEditor,
			}
			_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.Save")
			err = impl.environmentConfigRepository.Save(envOverride)
			span.End()
			if err != nil {
				impl.logger.Errorw("error in creating envConfig", "data", envOverride, "error", err)
				return nil, err
			}
		}
		envOverride.Chart = chart
	} else if envOverride.Id > 0 && !envOverride.IsOverride {
		_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindLatestChartForAppByAppId")
		chart, err = impl.chartRepository.FindLatestChartForAppByAppId(overrideRequest.AppId)
		span.End()
		if err != nil {
			impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
			return nil, err
		}
		envOverride.Chart = chart
	}

	err = impl.setEnvironmentModelInEnvOverride(ctx, envOverride)
	if err != nil {
		impl.logger.Errorw("error while setting environment data in envOverride", "env", envOverride.TargetEnvironment, "err", err)
		return nil, err
	}
	scope := helper.GetScopeForVariables(overrideRequest, envOverride)
	if envOverride.IsOverride {
		entity := repository5.GetEntity(envOverride.Id, repository5.EntityTypeDeploymentTemplateEnvLevel)
		resolvedTemplate, variableMap, err := impl.scopedVariableManager.GetMappedVariablesAndResolveTemplate(envOverride.EnvOverrideValues, scope, entity, true)
		envOverride.ResolvedEnvOverrideValues = resolvedTemplate
		envOverride.VariableSnapshot = variableMap
		if err != nil {
			impl.logger.Errorw("error,  GetMappedVariablesAndResolveTemplate env override level template", "err", err, "envOverride", envOverride)
			return envOverride, err
		}
	} else {
		entity := repository5.GetEntity(chart.Id, repository5.EntityTypeDeploymentTemplateAppLevel)
		resolvedTemplate, variableMap, err := impl.scopedVariableManager.GetMappedVariablesAndResolveTemplate(chart.GlobalOverride, scope, entity, true)
		envOverride.Chart.ResolvedGlobalOverride = resolvedTemplate
		envOverride.VariableSnapshot = variableMap
		if err != nil {
			impl.logger.Errorw("error,  GetMappedVariablesAndResolveTemplate app level template", "err", err, "chart", chart)
			return envOverride, err
		}
	}
	return envOverride, nil
}

func (impl *ManifestCreationServiceImpl) getAppMetricsByTriggerType(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context) (bool, error) {
	var appMetrics bool
	if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
		_, span := otel.Tracer("orchestrator").Start(ctx, "deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId")
		deploymentTemplateHistory, err := impl.deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId(overrideRequest.PipelineId, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting deployed deployment template history by pipelineId and wfrId", "err", err, "pipelineId", &overrideRequest, "wfrId", overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
			return appMetrics, err
		}
		appMetrics = deploymentTemplateHistory.IsAppMetricsEnabled
	} else if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
		_, span := otel.Tracer("orchestrator").Start(ctx, "deployedAppMetricsService.GetMetricsFlagForAPipelineByAppIdAndEnvId")
		isAppMetricsEnabled, err := impl.deployedAppMetricsService.GetMetricsFlagForAPipelineByAppIdAndEnvId(overrideRequest.AppId, overrideRequest.EnvId)
		if err != nil {
			impl.logger.Errorw("error, GetMetricsFlagForAPipelineByAppIdAndEnvId", "err", err, "appId", overrideRequest.AppId, "envId", overrideRequest.EnvId)
			return appMetrics, err
		}
		span.End()
		appMetrics = isAppMetricsEnabled
	}
	return appMetrics, nil
}

func (impl *ManifestCreationServiceImpl) mergeOverrideValues(envOverride *chartConfig.EnvConfigOverride, releaseOverrideJson string,
	configMapJson []byte, appLabelJsonByte []byte, strategy *chartConfig.PipelineStrategy) (mergedValues []byte, err error) {
	//merge three values on the fly
	//ordering is important here
	//global < environment < db< release
	var merged []byte
	var templateOverrideValuesByte []byte
	if !envOverride.IsOverride {
		templateOverrideValuesByte = []byte(envOverride.Chart.ResolvedGlobalOverride)
	} else {
		templateOverrideValuesByte = []byte(envOverride.ResolvedEnvOverrideValues)
	}
	merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), templateOverrideValuesByte)
	if err != nil {
		impl.logger.Errorw("error in merging deployment template override values", "err", err, "overrideValues", templateOverrideValuesByte)
		return nil, err
	}
	if strategy != nil && len(strategy.Config) > 0 {
		merged, err = impl.mergeUtil.JsonPatch(merged, []byte(strategy.Config))
		if err != nil {
			return nil, err
		}
	}
	merged, err = impl.mergeUtil.JsonPatch(merged, []byte(releaseOverrideJson))
	if err != nil {
		return nil, err
	}
	if configMapJson != nil {
		merged, err = impl.mergeUtil.JsonPatch(merged, configMapJson)
		if err != nil {
			return nil, err
		}
	}
	if appLabelJsonByte != nil {
		merged, err = impl.mergeUtil.JsonPatch(merged, appLabelJsonByte)
		if err != nil {
			return nil, err
		}
	}
	return merged, nil
}

func (impl *ManifestCreationServiceImpl) getConfigMapAndSecretJsonV2(ctx context.Context, request bean3.ConfigMapAndSecretJsonV2, envOverride *chartConfig.EnvConfigOverride) ([]byte, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "ManifestCreationServiceImpl.getConfigMapAndSecretJsonV2")
	defer span.End()
	var configMapJson, secretDataJson, configMapJsonApp, secretDataJsonApp, configMapJsonEnv, secretDataJsonEnv string

	var err error
	configMapA := &chartConfig.ConfigMapAppModel{}
	configMapE := &chartConfig.ConfigMapEnvModel{}
	configMapHistory, secretHistory := &repository3.ConfigmapAndSecretHistory{}, &repository3.ConfigmapAndSecretHistory{}

	merged := []byte("{}")
	if request.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
		configMapA, err = impl.configMapRepository.GetByAppIdAppLevel(request.AppId)
		if err != nil && pg.ErrNoRows != err {
			return []byte("{}"), err
		}
		if configMapA != nil && configMapA.Id > 0 {
			configMapJsonApp = configMapA.ConfigMapData
			secretDataJsonApp = configMapA.SecretData
		}

		configMapE, err = impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(request.AppId, request.EnvId)
		if err != nil && pg.ErrNoRows != err {
			return []byte("{}"), err
		}
		if configMapE != nil && configMapE.Id > 0 {
			configMapJsonEnv = configMapE.ConfigMapData
			secretDataJsonEnv = configMapE.SecretData
		}

	} else if request.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {

		//fetching history and setting envLevelConfig and not appLevelConfig because history already contains merged appLevel and envLevel configs
		configMapHistory, err = impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(request.PipeLineId, request.WfrIdForDeploymentWithSpecificTrigger, repository3.CONFIGMAP_TYPE)
		if err != nil {
			impl.logger.Errorw("error in getting config map history config by pipelineId and wfrId ", "err", err, "pipelineId", request.PipeLineId, "wfrId", request.WfrIdForDeploymentWithSpecificTrigger)
			return []byte("{}"), err
		}
		configMapJsonEnv = configMapHistory.Data

		secretHistory, err = impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(request.PipeLineId, request.WfrIdForDeploymentWithSpecificTrigger, repository3.SECRET_TYPE)
		if err != nil {
			impl.logger.Errorw("error in getting config map history config by pipelineId and wfrId ", "err", err, "pipelineId", request.PipeLineId, "wfrId", request.WfrIdForDeploymentWithSpecificTrigger)
			return []byte("{}"), err
		}
		secretDataJsonEnv = secretHistory.Data
	}
	configMapJson, err = impl.mergeUtil.ConfigMapMerge(configMapJsonApp, configMapJsonEnv)
	if err != nil {
		return []byte("{}"), err
	}
	chartMajorVersion, chartMinorVersion, err := util4.ExtractChartVersion(request.ChartVersion)
	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return []byte("{}"), err
	}
	secretDataJson, err = impl.mergeUtil.ConfigSecretMerge(secretDataJsonApp, secretDataJsonEnv, chartMajorVersion, chartMinorVersion, false)
	if err != nil {
		return []byte("{}"), err
	}
	configResponseR := bean.ConfigMapRootJson{}
	configResponse := bean.ConfigMapJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(configMapJson), &configResponse)
		if err != nil {
			return []byte("{}"), err
		}
	}
	configResponseR.ConfigMapJson = configResponse
	secretResponseR := bean.ConfigSecretRootJson{}
	secretResponse := bean.ConfigSecretJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(secretDataJson), &secretResponse)
		if err != nil {
			return []byte("{}"), err
		}
	}
	secretResponseR.ConfigSecretJson = secretResponse

	configMapByte, err := json.Marshal(configResponseR)
	if err != nil {
		return []byte("{}"), err
	}
	secretDataByte, err := json.Marshal(secretResponseR)
	if err != nil {
		return []byte("{}"), err

	}
	resolvedCM, resolvedCS, snapshotCM, snapshotCS, err := impl.scopedVariableManager.ResolveCMCSTrigger(request.DeploymentWithConfig, request.Scope, configMapA.Id, configMapE.Id, configMapByte, secretDataByte, configMapHistory.Id, secretHistory.Id)
	if err != nil {
		return []byte("{}"), err
	}
	envOverride.VariableSnapshotForCM = snapshotCM
	envOverride.VariableSnapshotForCS = snapshotCS

	merged, err = impl.mergeUtil.JsonPatch([]byte(resolvedCM), []byte(resolvedCS))

	if err != nil {
		return []byte("{}"), err
	}

	return merged, nil
}

func (impl *ManifestCreationServiceImpl) getReleaseOverride(envOverride *chartConfig.EnvConfigOverride, overrideRequest *bean.ValuesOverrideRequest,
	artifact *repository.CiArtifact, pipelineReleaseCounter int, strategy *chartConfig.PipelineStrategy, appMetrics *bool) (releaseOverride string, err error) {

	deploymentStrategy := ""
	if strategy != nil {
		deploymentStrategy = string(strategy.Strategy)
	}

	imageName := ""
	tag := ""
	if artifact != nil {
		artifactImage := artifact.Image
		imageTag := strings.Split(artifactImage, ":")

		imageTagLen := len(imageTag)

		for i := 0; i < imageTagLen-1; i++ {
			if i != imageTagLen-2 {
				imageName = imageName + imageTag[i] + ":"
			} else {
				imageName = imageName + imageTag[i]
			}
		}

		digestConfigurationRequest := imageDigestPolicy.DigestPolicyConfigurationRequest{PipelineId: overrideRequest.PipelineId}
		digestPolicyConfigurations, err := impl.imageDigestPolicyService.GetDigestPolicyConfigurations(digestConfigurationRequest)
		if err != nil {
			impl.logger.Errorw("error in checking if isImageDigestPolicyConfiguredForPipeline", "err", err, "clusterId", envOverride.Environment.ClusterId, "envId", envOverride.TargetEnvironment, "pipelineId", overrideRequest.PipelineId)
			return "", err
		}

		if digestPolicyConfigurations.UseDigestForTrigger() {
			imageTag[imageTagLen-1] = fmt.Sprintf("%s@%s", imageTag[imageTagLen-1], artifact.ImageDigest)
		}

		tag = imageTag[imageTagLen-1]
	}

	override, err := app.NewReleaseAttributes(imageName, tag, overrideRequest.PipelineName, deploymentStrategy,
		overrideRequest.AppId, overrideRequest.EnvId, pipelineReleaseCounter, appMetrics).RenderJson(envOverride.Chart.ImageDescriptorTemplate)
	if err != nil {
		return "", &util.ApiError{HttpStatusCode: http.StatusUnprocessableEntity, UserMessage: "unable to render ImageDescriptorTemplate", InternalMessage: err.Error()}
	}

	if overrideRequest.AdditionalOverride != nil {
		userOverride, err := overrideRequest.AdditionalOverride.MarshalJSON()
		if err != nil {
			return "", err
		}
		data, err := impl.mergeUtil.JsonPatch(userOverride, []byte(override))
		if err != nil {
			return "", err
		}
		override = string(data)
	}
	return override, nil
}

func (impl *ManifestCreationServiceImpl) savePipelineOverride(ctx context.Context, overrideRequest *bean.ValuesOverrideRequest, envOverrideId int, triggeredAt time.Time) (override *chartConfig.PipelineOverride, err error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "ManifestCreationServiceImpl.savePipelineOverride")
	defer span.End()
	currentReleaseNo, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(overrideRequest.PipelineId)
	if err != nil {
		return nil, err
	}
	po := &chartConfig.PipelineOverride{
		EnvConfigOverrideId:    envOverrideId,
		Status:                 models.CHARTSTATUS_NEW,
		PipelineId:             overrideRequest.PipelineId,
		CiArtifactId:           overrideRequest.CiArtifactId,
		PipelineReleaseCounter: currentReleaseNo + 1,
		CdWorkflowId:           overrideRequest.CdWorkflowId,
		AuditLog:               sql.AuditLog{CreatedBy: overrideRequest.UserId, CreatedOn: triggeredAt, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
		DeploymentType:         overrideRequest.DeploymentType,
	}

	err = impl.pipelineOverrideRepository.Save(po)
	if err != nil {
		return nil, err
	}
	err = impl.checkAndFixDuplicateReleaseNo(po)
	if err != nil {
		impl.logger.Errorw("error in checking release no duplicacy", "pipeline", po, "err", err)
		return nil, err
	}
	return po, nil
}

func (impl *ManifestCreationServiceImpl) checkAndFixDuplicateReleaseNo(override *chartConfig.PipelineOverride) error {

	uniqueVerified := false
	retryCount := 0

	for !uniqueVerified && retryCount < 5 {
		retryCount = retryCount + 1
		overrides, err := impl.pipelineOverrideRepository.GetByPipelineIdAndReleaseNo(override.PipelineId, override.PipelineReleaseCounter)
		if err != nil {
			return err
		}
		if overrides[0].Id == override.Id {
			uniqueVerified = true
		} else {
			//duplicate might be due to concurrency, lets fix it
			currentReleaseNo, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(override.PipelineId)
			if err != nil {
				return err
			}
			override.PipelineReleaseCounter = currentReleaseNo + 1
			err = impl.pipelineOverrideRepository.Save(override)
			if err != nil {
				return err
			}
		}
	}
	if !uniqueVerified {
		return fmt.Errorf("duplicate verification retry count exide max overrideId: %d ,count: %d", override.Id, retryCount)
	}
	return nil
}

func (impl *ManifestCreationServiceImpl) getK8sHPAResourceManifest(ctx context.Context, clusterId int, namespace string, hpaResourceRequest *util4.HpaResourceRequest) (map[string]interface{}, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "ManifestCreationServiceImpl.getK8sHPAResourceManifest")
	defer span.End()
	resourceManifest := make(map[string]interface{})
	version, err := impl.k8sCommonService.GetPreferredVersionForAPIGroup(ctx, clusterId, hpaResourceRequest.Group)
	if err != nil && !k8sUtil.IsNotFoundError(err) {
		return resourceManifest, util.NewApiError().
			WithHttpStatusCode(http.StatusPreconditionFailed).
			WithInternalMessage(err.Error()).
			WithUserDetailMessage("unable to find preferred version for hpa resource")
	} else if k8sUtil.IsNotFoundError(err) {
		return resourceManifest, util.NewApiError().
			WithHttpStatusCode(http.StatusPreconditionFailed).
			WithInternalMessage("unable to find preferred version for hpa resource").
			WithUserDetailMessage("unable to find preferred version for hpa resource")
	}
	k8sReq := &k8s.ResourceRequestBean{
		ClusterId: clusterId,
		K8sRequest: k8sUtil.NewK8sRequestBean().
			WithResourceIdentifier(
				k8sUtil.NewResourceIdentifier().
					WithName(hpaResourceRequest.ResourceName).
					WithNameSpace(namespace).
					WithGroup(hpaResourceRequest.Group).
					WithKind(hpaResourceRequest.Kind).
					WithVersion(version),
			),
	}
	k8sResource, err := impl.k8sCommonService.GetResource(newCtx, k8sReq)
	if err != nil {
		if k8s.IsResourceNotFoundErr(err) {
			// this is a valid case for hibernated applications, so returning nil
			// for hibernated applications, we don't have any hpa resource manifest
			return resourceManifest, nil
		} else if k8s.IsBadRequestErr(err) {
			impl.logger.Errorw("bad request error occurred while fetching hpa resource for app", "resourceName", hpaResourceRequest.ResourceName, "err", err)
			return resourceManifest, util.NewApiError().
				WithHttpStatusCode(http.StatusPreconditionFailed).
				WithInternalMessage(err.Error()).
				WithUserDetailMessage(err.Error())
		} else if k8s.IsServerTimeoutErr(err) {
			impl.logger.Errorw("targeted hpa resource could not be served", "resourceName", hpaResourceRequest.ResourceName, "err", err)
			return resourceManifest, util.NewApiError().
				WithHttpStatusCode(http.StatusRequestTimeout).
				WithInternalMessage(err.Error()).
				WithUserDetailMessage("taking longer than expected, please try again later")
		}
		impl.logger.Errorw("error occurred while fetching resource for app", "resourceName", hpaResourceRequest.ResourceName, "err", err)
		return resourceManifest, err
	}
	return k8sResource.ManifestResponse.Manifest.Object, err
}

func (impl *ManifestCreationServiceImpl) getArgoCdHPAResourceManifest(ctx context.Context, appName, namespace string, hpaResourceRequest *util4.HpaResourceRequest) (map[string]interface{}, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "ManifestCreationServiceImpl.getArgoCdHPAResourceManifest")
	defer span.End()
	resourceManifest := make(map[string]interface{})
	query := &application3.ApplicationResourceRequest{
		Name:         &appName,
		Version:      &hpaResourceRequest.Version,
		Group:        &hpaResourceRequest.Group,
		Kind:         &hpaResourceRequest.Kind,
		ResourceName: &hpaResourceRequest.ResourceName,
		Namespace:    &namespace,
	}

	recv, argoErr := impl.acdClient.GetResource(newCtx, query)
	if argoErr != nil {
		grpcCode, errMsg := util.GetClientDetailedError(argoErr)
		if grpcCode.IsInvalidArgumentCode() {
			// this is a valid case for hibernated applications, so returning nil
			// for hibernated applications, we don't have any hpa resource manifest
			return resourceManifest, nil
		} else if grpcCode.IsUnavailableCode() {
			impl.logger.Errorw("ArgoCd server unavailable", "resourceName", hpaResourceRequest.ResourceName, "err", errMsg)
			return resourceManifest, fmt.Errorf("ArgoCd server error: %s", errMsg)
		} else if grpcCode.IsDeadlineExceededCode() {
			impl.logger.Errorw("ArgoCd resource request timeout", "resourceName", hpaResourceRequest.ResourceName, "err", errMsg)
			return resourceManifest, util.NewApiError().
				WithHttpStatusCode(http.StatusRequestTimeout).
				WithInternalMessage(errMsg).
				WithUserDetailMessage("ArgoCd server is not responding, please try again later")
		}
		impl.logger.Errorw("ArgoCd Get Resource API Failed", "resourceName", hpaResourceRequest.ResourceName, "err", errMsg)
		return resourceManifest, fmt.Errorf("ArgoCd client error: %s", errMsg)
	}
	if recv != nil && len(*recv.Manifest) > 0 {
		err := json.Unmarshal([]byte(*recv.Manifest), &resourceManifest)
		if err != nil {
			impl.logger.Errorw("failed to unmarshal hpa resource manifest", "manifest", recv.Manifest, "err", err)
			return resourceManifest, err
		}
	} else {
		impl.logger.Debugw("invalid resource manifest received from ArgoCd", "response", recv)
	}
	return resourceManifest, nil
}

func (impl *ManifestCreationServiceImpl) autoscalingCheckBeforeTrigger(ctx context.Context, appName string, namespace string, merged []byte, overrideRequest *bean.ValuesOverrideRequest) ([]byte, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "ManifestCreationServiceImpl.autoscalingCheckBeforeTrigger")
	defer span.End()
	pipelineId := overrideRequest.PipelineId
	var appDeploymentType = overrideRequest.DeploymentAppType
	var clusterId = overrideRequest.ClusterId
	deploymentType := overrideRequest.DeploymentType
	templateMap := make(map[string]interface{})
	err := json.Unmarshal(merged, &templateMap)
	if err != nil {
		impl.logger.Errorw("unmarshal failed for hpa check", "pipelineId", pipelineId, "err", err)
		return merged, err
	}

	hpaResourceRequest := helper.GetAutoScalingReplicaCount(templateMap, appName)
	impl.logger.Debugw("autoscalingCheckBeforeTrigger", "pipelineId", pipelineId, "hpaResourceRequest", hpaResourceRequest)
	if hpaResourceRequest.IsEnable {
		var resourceManifest map[string]interface{}
		if util.IsAcdApp(appDeploymentType) {
			resourceManifest, err = impl.getArgoCdHPAResourceManifest(newCtx, appName, namespace, hpaResourceRequest)
			if err != nil {
				return merged, err
			}
		} else {
			resourceManifest, err = impl.getK8sHPAResourceManifest(newCtx, clusterId, namespace, hpaResourceRequest)
			if err != nil {
				return merged, err
			}
		}
		if len(resourceManifest) > 0 {
			statusMap := resourceManifest["status"].(map[string]interface{})
			currentReplicaVal := statusMap["currentReplicas"]
			currentReplicaCount, err := util4.ParseFloatNumber(currentReplicaVal)
			if err != nil {
				impl.logger.Errorw("error occurred while parsing replica count", "currentReplicas", currentReplicaVal, "err", err)
				return merged, err
			}

			reqReplicaCount := helper.FetchRequiredReplicaCount(currentReplicaCount, hpaResourceRequest.ReqMaxReplicas, hpaResourceRequest.ReqMinReplicas)
			templateMap["replicaCount"] = reqReplicaCount
			merged, err = json.Marshal(&templateMap)
			if err != nil {
				impl.logger.Errorw("marshaling failed for hpa check", "reqReplicaCount", reqReplicaCount, "err", err)
				return merged, err
			}
		}
	} else {
		impl.logger.Debugw("autoscaling is not enabled", "pipelineId", pipelineId)
	}

	//check for custom chart support
	if autoscalingEnabledPath, ok := templateMap[bean2.CustomAutoScalingEnabledPathKey]; ok {
		if deploymentType == models.DEPLOYMENTTYPE_STOP {
			merged, err = helper.SetScalingValues(templateMap, bean2.CustomAutoScalingEnabledPathKey, merged, false)
			if err != nil {
				impl.logger.Errorw("error occurred while setting autoscaling enabled key", "templateMap", templateMap, "err", err)
				return merged, err
			}
			merged, err = helper.SetScalingValues(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged, 0)
			if err != nil {
				impl.logger.Errorw("error occurred while setting autoscaling replica count key", "templateMap", templateMap, "err", err)
				return merged, err
			}

			merged, err = helper.SetScalingValues(templateMap, bean2.CustomAutoscalingMinPathKey, merged, 0)
			if err != nil {
				impl.logger.Errorw("error occurred while setting autoscaling min key", "templateMap", templateMap, "err", err)
				return merged, err
			}

			merged, err = helper.SetScalingValues(templateMap, bean2.CustomAutoscalingMaxPathKey, merged, 0)
			if err != nil {
				impl.logger.Errorw("error occurred while setting autoscaling max key", "templateMap", templateMap, "err", err)
				return merged, err
			}
		} else {
			autoscalingEnabled := false
			autoscalingEnabledValue := gjson.Get(string(merged), autoscalingEnabledPath.(string)).Value()
			if val, ok := autoscalingEnabledValue.(bool); ok {
				autoscalingEnabled = val
			}
			if autoscalingEnabled {
				// extract replica count, min, max and check for required value
				replicaCount, err := impl.getReplicaCountFromCustomChart(templateMap, merged)
				if err != nil {
					return merged, err
				}
				merged, err = helper.SetScalingValues(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged, replicaCount)
				if err != nil {
					impl.logger.Errorw("error occurred while setting autoscaling key", "templateMap", templateMap, "err", err)
					return merged, err
				}
			}
		}
	}

	return merged, nil
}

func (impl *ManifestCreationServiceImpl) getReplicaCountFromCustomChart(templateMap map[string]interface{}, merged []byte) (float64, error) {
	autoscalingMinVal, err := helper.ExtractParamValue(templateMap, bean2.CustomAutoscalingMinPathKey, merged)
	if helper.IsNotFoundErr(err) {
		return 0, util.NewApiError().
			WithHttpStatusCode(http.StatusPreconditionFailed).
			WithInternalMessage(helper.KeyNotFoundError).
			WithUserDetailMessage(fmt.Sprintf("empty value for key [%s]", bean2.CustomAutoscalingMinPathKey))
	} else if err != nil {
		impl.logger.Errorw("error occurred while parsing float number", "key", bean2.CustomAutoscalingMinPathKey, "err", err)
		return 0, err
	}
	autoscalingMaxVal, err := helper.ExtractParamValue(templateMap, bean2.CustomAutoscalingMaxPathKey, merged)
	if helper.IsNotFoundErr(err) {
		return 0, util.NewApiError().
			WithHttpStatusCode(http.StatusPreconditionFailed).
			WithInternalMessage(helper.KeyNotFoundError).
			WithUserDetailMessage(fmt.Sprintf("empty value for key [%s]", bean2.CustomAutoscalingMaxPathKey))
	} else if err != nil {
		impl.logger.Errorw("error occurred while parsing float number", "key", bean2.CustomAutoscalingMaxPathKey, "err", err)
		return 0, err
	}
	autoscalingReplicaCountVal, err := helper.ExtractParamValue(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged)
	if helper.IsNotFoundErr(err) {
		return 0, util.NewApiError().
			WithHttpStatusCode(http.StatusPreconditionFailed).
			WithInternalMessage(helper.KeyNotFoundError).
			WithUserDetailMessage(fmt.Sprintf("empty value for key [%s]", bean2.CustomAutoscalingReplicaCountPathKey))
	} else if err != nil {
		impl.logger.Errorw("error occurred while parsing float number", "key", bean2.CustomAutoscalingReplicaCountPathKey, "err", err)
		return 0, err
	}
	return helper.FetchRequiredReplicaCount(autoscalingReplicaCountVal, autoscalingMaxVal, autoscalingMinVal), nil
}

func (impl *ManifestCreationServiceImpl) setEnvironmentModelInEnvOverride(ctx context.Context, envOverride *chartConfig.EnvConfigOverride) error {
	_, span := otel.Tracer("orchestrator").Start(ctx, "ManifestCreationServiceImpl.setEnvironmentModelInEnvOverride")
	defer span.End()
	env, err := impl.envRepository.FindById(envOverride.TargetEnvironment)
	if err != nil {
		impl.logger.Errorw("unable to find env", "err", err, "env", envOverride.TargetEnvironment)
		return err
	}
	envOverride.Environment = env
	return nil
}
