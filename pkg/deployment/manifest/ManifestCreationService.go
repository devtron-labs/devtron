package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	application3 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	util5 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/bean"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/manifest/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	"github.com/devtron-labs/devtron/pkg/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/imageDigestPolicy"
	"github.com/devtron-labs/devtron/pkg/k8s"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	repository5 "github.com/devtron-labs/devtron/pkg/variables/repository"
	util4 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	errors2 "github.com/juju/errors"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strconv"
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
	deploymentTemplateHistoryRepository repository3.DeploymentTemplateHistoryRepository) *ManifestCreationServiceImpl {
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
	if overrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
		overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
	}
	if len(overrideRequest.DeploymentWithConfig) == 0 {
		overrideRequest.DeploymentWithConfig = bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED
	}
	valuesOverrideResponse := &app.ValuesOverrideResponse{}
	isPipelineOverrideCreated := overrideRequest.PipelineOverrideId > 0
	pipeline, err := impl.pipelineRepository.FindById(overrideRequest.PipelineId)
	valuesOverrideResponse.Pipeline = pipeline
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline by pipeline id", "err", err, "pipeline-id-", overrideRequest.PipelineId)
		return valuesOverrideResponse, err
	}

	_, span := otel.Tracer("orchestrator").Start(ctx, "ciArtifactRepository.Get")
	artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
	valuesOverrideResponse.Artifact = artifact
	span.End()
	if err != nil {
		return valuesOverrideResponse, err
	}
	overrideRequest.Image = artifact.Image

	strategy, err := impl.getDeploymentStrategyByTriggerType(overrideRequest, ctx)
	valuesOverrideResponse.PipelineStrategy = strategy
	if err != nil {
		impl.logger.Errorw("error in getting strategy by trigger type", "err", err)
		return valuesOverrideResponse, err
	}

	envOverride, err := impl.getEnvOverrideByTriggerType(overrideRequest, triggeredAt, ctx)
	valuesOverrideResponse.EnvOverride = envOverride
	if err != nil {
		impl.logger.Errorw("error in getting env override by trigger type", "err", err)
		return valuesOverrideResponse, err
	}
	appMetrics, err := impl.getAppMetricsByTriggerType(overrideRequest, ctx)
	valuesOverrideResponse.AppMetrics = appMetrics
	if err != nil {
		impl.logger.Errorw("error in getting app metrics by trigger type", "err", err)
		return valuesOverrideResponse, err
	}
	var (
		pipelineOverride                *chartConfig.PipelineOverride
		configMapJson, appLabelJsonByte []byte
	)

	// Conditional Block based on PipelineOverrideCreated --> start
	if !isPipelineOverrideCreated {
		_, span = otel.Tracer("orchestrator").Start(ctx, "savePipelineOverride")
		pipelineOverride, err = impl.savePipelineOverride(overrideRequest, envOverride.Id, triggeredAt)
		span.End()
		if err != nil {
			return valuesOverrideResponse, err
		}
		overrideRequest.PipelineOverrideId = pipelineOverride.Id
	} else {
		pipelineOverride, err = impl.pipelineOverrideRepository.FindById(overrideRequest.PipelineOverrideId)
		if err != nil {
			impl.logger.Errorw("error in getting pipelineOverride for valuesOverrideResponse", "PipelineOverrideId", overrideRequest.PipelineOverrideId)
			return nil, err
		}
	}
	// Conditional Block based on PipelineOverrideCreated --> end
	valuesOverrideResponse.PipelineOverride = pipelineOverride

	//TODO: check status and apply lock
	releaseOverrideJson, err := impl.getReleaseOverride(envOverride, overrideRequest, artifact, pipelineOverride, strategy, &appMetrics)
	valuesOverrideResponse.ReleaseOverrideJSON = releaseOverrideJson
	if err != nil {
		return valuesOverrideResponse, err
	}

	// Conditional Block based on PipelineOverrideCreated --> start
	if !isPipelineOverrideCreated {
		chartVersion := envOverride.Chart.ChartVersion
		_, span = otel.Tracer("orchestrator").Start(ctx, "getConfigMapAndSecretJsonV2")
		scope := getScopeForVariables(overrideRequest, envOverride)
		request := createConfigMapAndSecretJsonRequest(overrideRequest, envOverride, chartVersion, scope)

		configMapJson, err = impl.getConfigMapAndSecretJsonV2(request, envOverride)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in fetching config map n secret ", "err", err)
			configMapJson = nil
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "appCrudOperationService.GetLabelsByAppIdForDeployment")
		appLabelJsonByte, err = impl.appCrudOperationService.GetLabelsByAppIdForDeployment(overrideRequest.AppId)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in fetching app labels for gitOps commit", "err", err)
			appLabelJsonByte = nil
		}

		mergedValues, err := impl.mergeOverrideValues(envOverride, releaseOverrideJson, configMapJson, appLabelJsonByte, strategy)
		appName := fmt.Sprintf("%s-%s", overrideRequest.AppName, envOverride.Environment.Name)
		mergedValues = impl.autoscalingCheckBeforeTrigger(ctx, appName, envOverride.Namespace, mergedValues, overrideRequest)

		_, span = otel.Tracer("orchestrator").Start(ctx, "dockerRegistryIpsConfigService.HandleImagePullSecretOnApplicationDeployment")
		// handle image pull secret if access given
		mergedValues, err = impl.dockerRegistryIpsConfigService.HandleImagePullSecretOnApplicationDeployment(envOverride.Environment, artifact, pipeline.CiPipelineId, mergedValues)
		span.End()
		if err != nil {
			return valuesOverrideResponse, err
		}

		pipelineOverride.PipelineMergedValues = string(mergedValues)
		valuesOverrideResponse.MergedValues = string(mergedValues)
		err = impl.pipelineOverrideRepository.Update(pipelineOverride)
		if err != nil {
			return valuesOverrideResponse, err
		}
		valuesOverrideResponse.PipelineOverride = pipelineOverride
	} else {
		valuesOverrideResponse.MergedValues = pipelineOverride.PipelineMergedValues
	}
	// Conditional Block based on PipelineOverrideCreated --> end
	return valuesOverrideResponse, err
}

func (impl *ManifestCreationServiceImpl) getDeploymentStrategyByTriggerType(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context) (*chartConfig.PipelineStrategy, error) {

	strategy := &chartConfig.PipelineStrategy{}
	var err error
	if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
		_, span := otel.Tracer("orchestrator").Start(ctx, "strategyHistoryRepository.GetHistoryByPipelineIdAndWfrId")
		strategyHistory, err := impl.strategyHistoryRepository.GetHistoryByPipelineIdAndWfrId(overrideRequest.PipelineId, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting deployed strategy history by pipelineId and wfrId", "err", err, "pipelineId", overrideRequest.PipelineId, "wfrId", overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
			return nil, err
		}
		strategy.Strategy = strategyHistory.Strategy
		strategy.Config = strategyHistory.Config
		strategy.PipelineId = overrideRequest.PipelineId
	} else if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
		if overrideRequest.ForceTrigger {
			_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineConfigRepository.GetDefaultStrategyByPipelineId")
			strategy, err = impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(overrideRequest.PipelineId)
			span.End()
		} else {
			var deploymentTemplate chartRepoRepository.DeploymentStrategy
			if overrideRequest.DeploymentTemplate == "ROLLING" {
				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_ROLLING
			} else if overrideRequest.DeploymentTemplate == "BLUE-GREEN" {
				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_BLUE_GREEN
			} else if overrideRequest.DeploymentTemplate == "CANARY" {
				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_CANARY
			} else if overrideRequest.DeploymentTemplate == "RECREATE" {
				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_RECREATE
			}

			if len(deploymentTemplate) > 0 {
				_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineConfigRepository.FindByStrategyAndPipelineId")
				strategy, err = impl.pipelineConfigRepository.FindByStrategyAndPipelineId(deploymentTemplate, overrideRequest.PipelineId)
				span.End()
			} else {
				_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineConfigRepository.GetDefaultStrategyByPipelineId")
				strategy, err = impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(overrideRequest.PipelineId)
				span.End()
			}
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

		_, span = otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
		env, err := impl.envRepository.FindById(envOverride.TargetEnvironment)
		span.End()
		if err != nil {
			impl.logger.Errorw("unable to find env", "err", err)
			return nil, err
		}
		envOverride.Environment = env

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
			return envOverride, err
		}
	} else if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
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

		_, span = otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
		env, err := impl.envRepository.FindById(envOverride.TargetEnvironment)
		span.End()
		if err != nil {
			impl.logger.Errorw("unable to find env", "err", err)
			return nil, err
		}
		envOverride.Environment = env
		scope := getScopeForVariables(overrideRequest, envOverride)
		if envOverride.IsOverride {

			entity := repository5.GetEntity(envOverride.Id, repository5.EntityTypeDeploymentTemplateEnvLevel)
			resolvedTemplate, variableMap, err := impl.scopedVariableManager.GetMappedVariablesAndResolveTemplate(envOverride.EnvOverrideValues, scope, entity, true)
			envOverride.ResolvedEnvOverrideValues = resolvedTemplate
			envOverride.VariableSnapshot = variableMap
			if err != nil {
				return envOverride, err
			}

		} else {
			entity := repository5.GetEntity(chart.Id, repository5.EntityTypeDeploymentTemplateAppLevel)
			resolvedTemplate, variableMap, err := impl.scopedVariableManager.GetMappedVariablesAndResolveTemplate(chart.GlobalOverride, scope, entity, true)
			envOverride.Chart.ResolvedGlobalOverride = resolvedTemplate
			envOverride.VariableSnapshot = variableMap
			if err != nil {
				return envOverride, err
			}

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

func (impl *ManifestCreationServiceImpl) mergeOverrideValues(envOverride *chartConfig.EnvConfigOverride,
	releaseOverrideJson string,
	configMapJson []byte,
	appLabelJsonByte []byte,
	strategy *chartConfig.PipelineStrategy,
) (mergedValues []byte, err error) {

	//merge three values on the fly
	//ordering is important here
	//global < environment < db< release
	var merged []byte
	if !envOverride.IsOverride {
		merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), []byte(envOverride.Chart.ResolvedGlobalOverride))
		if err != nil {
			return nil, err
		}
	} else {
		merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), []byte(envOverride.ResolvedEnvOverrideValues))
		if err != nil {
			return nil, err
		}
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

func (impl *ManifestCreationServiceImpl) getConfigMapAndSecretJsonV2(request bean3.ConfigMapAndSecretJsonV2, envOverride *chartConfig.EnvConfigOverride) ([]byte, error) {

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
	artifact *repository.CiArtifact, pipelineOverride *chartConfig.PipelineOverride, strategy *chartConfig.PipelineStrategy, appMetrics *bool) (releaseOverride string, err error) {

	artifactImage := artifact.Image
	imageTag := strings.Split(artifactImage, ":")

	imageTagLen := len(imageTag)

	imageName := ""

	for i := 0; i < imageTagLen-1; i++ {
		if i != imageTagLen-2 {
			imageName = imageName + imageTag[i] + ":"
		} else {
			imageName = imageName + imageTag[i]
		}
	}

	appId := strconv.Itoa(overrideRequest.AppId)
	envId := strconv.Itoa(overrideRequest.EnvId)

	deploymentStrategy := ""
	if strategy != nil {
		deploymentStrategy = string(strategy.Strategy)
	}

	digestConfigurationRequest := imageDigestPolicy.DigestPolicyConfigurationRequest{
		PipelineId:    overrideRequest.PipelineId,
		EnvironmentId: envOverride.TargetEnvironment,
		ClusterId:     envOverride.Environment.ClusterId,
	}
	digestPolicyConfigurations, err := impl.imageDigestPolicyService.GetDigestPolicyConfigurations(digestConfigurationRequest)
	if err != nil {
		impl.logger.Errorw("error in checking if isImageDigestPolicyConfiguredForPipeline", "err", err, "clusterId", envOverride.Environment.ClusterId, "envId", envOverride.TargetEnvironment, "pipelineId", overrideRequest.PipelineId)
		return "", err
	}

	if digestPolicyConfigurations.UseDigestForTrigger() {
		imageTag[imageTagLen-1] = fmt.Sprintf("%s@%s", imageTag[imageTagLen-1], artifact.ImageDigest)
	}

	releaseAttribute := app.ReleaseAttributes{
		Name:           imageName,
		Tag:            imageTag[imageTagLen-1],
		PipelineName:   overrideRequest.PipelineName,
		ReleaseVersion: strconv.Itoa(pipelineOverride.PipelineReleaseCounter),
		DeploymentType: deploymentStrategy,
		App:            appId,
		Env:            envId,
		AppMetrics:     appMetrics,
	}
	override, err := util4.Tprintf(envOverride.Chart.ImageDescriptorTemplate, releaseAttribute)
	if err != nil {
		return "", &util.ApiError{InternalMessage: "unable to render ImageDescriptorTemplate"}
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

func (impl *ManifestCreationServiceImpl) savePipelineOverride(overrideRequest *bean.ValuesOverrideRequest, envOverrideId int, triggeredAt time.Time) (override *chartConfig.PipelineOverride, err error) {
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

func (impl *ManifestCreationServiceImpl) autoscalingCheckBeforeTrigger(ctx context.Context, appName string, namespace string, merged []byte, overrideRequest *bean.ValuesOverrideRequest) []byte {
	var appId = overrideRequest.AppId
	pipelineId := overrideRequest.PipelineId
	var appDeploymentType = overrideRequest.DeploymentAppType
	var clusterId = overrideRequest.ClusterId
	deploymentType := overrideRequest.DeploymentType
	templateMap := make(map[string]interface{})
	err := json.Unmarshal(merged, &templateMap)
	if err != nil {
		return merged
	}

	hpaResourceRequest := getAutoScalingReplicaCount(templateMap, appName)
	impl.logger.Debugw("autoscalingCheckBeforeTrigger", "hpaResourceRequest", hpaResourceRequest)
	if hpaResourceRequest.IsEnable {
		resourceManifest := make(map[string]interface{})
		if util.IsAcdApp(appDeploymentType) {
			query := &application3.ApplicationResourceRequest{
				Name:         &appName,
				Version:      &hpaResourceRequest.Version,
				Group:        &hpaResourceRequest.Group,
				Kind:         &hpaResourceRequest.Kind,
				ResourceName: &hpaResourceRequest.ResourceName,
				Namespace:    &namespace,
			}
			recv, err := impl.acdClient.GetResource(ctx, query)
			impl.logger.Debugw("resource manifest get replica count", "response", recv)
			if err != nil {
				impl.logger.Errorw("ACD Get Resource API Failed", "err", err)
				middleware.AcdGetResourceCounter.WithLabelValues(strconv.Itoa(appId), namespace, appName).Inc()
				return merged
			}
			if recv != nil && len(*recv.Manifest) > 0 {
				err := json.Unmarshal([]byte(*recv.Manifest), &resourceManifest)
				if err != nil {
					impl.logger.Errorw("unmarshal failed for hpa check", "err", err)
					return merged
				}
			}
		} else {
			version := "v2beta2"
			k8sResource, err := impl.k8sCommonService.GetResource(ctx, &k8s.ResourceRequestBean{ClusterId: clusterId,
				K8sRequest: &util5.K8sRequestBean{ResourceIdentifier: util5.ResourceIdentifier{Name: hpaResourceRequest.ResourceName,
					Namespace: namespace, GroupVersionKind: schema.GroupVersionKind{Group: hpaResourceRequest.Group, Kind: hpaResourceRequest.Kind, Version: version}}}})
			if err != nil {
				impl.logger.Errorw("error occurred while fetching resource for app", "resourceName", hpaResourceRequest.ResourceName, "err", err)
				return merged
			}
			resourceManifest = k8sResource.ManifestResponse.Manifest.Object
		}
		if len(resourceManifest) > 0 {
			statusMap := resourceManifest["status"].(map[string]interface{})
			currentReplicaVal := statusMap["currentReplicas"]
			currentReplicaCount, err := util4.ParseFloatNumber(currentReplicaVal)
			if err != nil {
				impl.logger.Errorw("error occurred while parsing replica count", "currentReplicas", currentReplicaVal, "err", err)
				return merged
			}

			reqReplicaCount := fetchRequiredReplicaCount(currentReplicaCount, hpaResourceRequest.ReqMaxReplicas, hpaResourceRequest.ReqMinReplicas)
			templateMap["replicaCount"] = reqReplicaCount
			merged, err = json.Marshal(&templateMap)
			if err != nil {
				impl.logger.Errorw("marshaling failed for hpa check", "err", err)
				return merged
			}
		}
	} else {
		impl.logger.Errorw("autoscaling is not enabled", "pipelineId", pipelineId)
	}

	//check for custom chart support
	if autoscalingEnabledPath, ok := templateMap[bean2.CustomAutoScalingEnabledPathKey]; ok {
		if deploymentType == models.DEPLOYMENTTYPE_STOP {
			merged, err = setScalingValues(templateMap, bean2.CustomAutoScalingEnabledPathKey, merged, false)
			if err != nil {
				impl.logger.Errorw("error occurred while setting autoscaling key", "templateMap", templateMap, "err", err)
				return merged
			}
			merged, err = setScalingValues(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged, 0)
			if err != nil {
				impl.logger.Errorw("error occurred while setting autoscaling key", "templateMap", templateMap, "err", err)
				return merged
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
					return merged
				}
				merged, err = setScalingValues(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged, replicaCount)
				if err != nil {
					impl.logger.Errorw("error occurred while setting autoscaling key", "templateMap", templateMap, "err", err)
					return merged
				}
			}
		}
	}

	return merged
}

func (impl *ManifestCreationServiceImpl) getReplicaCountFromCustomChart(templateMap map[string]interface{}, merged []byte) (float64, error) {
	autoscalingMinVal, err := extractParamValue(templateMap, bean2.CustomAutoscalingMinPathKey, merged)
	if err != nil {
		impl.logger.Errorw("error occurred while parsing float number", "key", bean2.CustomAutoscalingMinPathKey, "err", err)
		return 0, err
	}
	autoscalingMaxVal, err := extractParamValue(templateMap, bean2.CustomAutoscalingMaxPathKey, merged)
	if err != nil {
		impl.logger.Errorw("error occurred while parsing float number", "key", bean2.CustomAutoscalingMaxPathKey, "err", err)
		return 0, err
	}
	autoscalingReplicaCountVal, err := extractParamValue(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged)
	if err != nil {
		impl.logger.Errorw("error occurred while parsing float number", "key", bean2.CustomAutoscalingReplicaCountPathKey, "err", err)
		return 0, err
	}
	return fetchRequiredReplicaCount(autoscalingReplicaCountVal, autoscalingMaxVal, autoscalingMinVal), nil
}

func extractParamValue(inputMap map[string]interface{}, key string, merged []byte) (float64, error) {
	if _, ok := inputMap[key]; !ok {
		return 0, errors.New("empty-val-err")
	}
	return util4.ParseFloatNumber(gjson.Get(string(merged), inputMap[key].(string)).Value())
}

func setScalingValues(templateMap map[string]interface{}, customScalingKey string, merged []byte, value interface{}) ([]byte, error) {
	autoscalingJsonPath := templateMap[customScalingKey]
	autoscalingJsonPathKey := autoscalingJsonPath.(string)
	mergedRes, err := sjson.Set(string(merged), autoscalingJsonPathKey, value)
	if err != nil {
		return []byte{}, err
	}
	return []byte(mergedRes), nil
}

func fetchRequiredReplicaCount(currentReplicaCount float64, reqMaxReplicas float64, reqMinReplicas float64) float64 {
	var reqReplicaCount float64
	if currentReplicaCount <= reqMaxReplicas && currentReplicaCount >= reqMinReplicas {
		reqReplicaCount = currentReplicaCount
	} else if currentReplicaCount > reqMaxReplicas {
		reqReplicaCount = reqMaxReplicas
	} else if currentReplicaCount < reqMinReplicas {
		reqReplicaCount = reqMinReplicas
	}
	return reqReplicaCount
}

func getAutoScalingReplicaCount(templateMap map[string]interface{}, appName string) *util4.HpaResourceRequest {
	hasOverride := false
	if _, ok := templateMap[bean3.FullnameOverride]; ok {
		appNameOverride := templateMap[bean3.FullnameOverride].(string)
		if len(appNameOverride) > 0 {
			appName = appNameOverride
			hasOverride = true
		}
	}
	if !hasOverride {
		if _, ok := templateMap[bean3.NameOverride]; ok {
			nameOverride := templateMap[bean3.NameOverride].(string)
			if len(nameOverride) > 0 {
				appName = fmt.Sprintf("%s-%s", appName, nameOverride)
			}
		}
	}
	hpaResourceRequest := &util4.HpaResourceRequest{}
	hpaResourceRequest.Version = ""
	hpaResourceRequest.Group = autoscaling.ServiceName
	hpaResourceRequest.Kind = bean3.HorizontalPodAutoscaler
	if _, ok := templateMap[bean3.KedaAutoscaling]; ok {
		as := templateMap[bean3.KedaAutoscaling]
		asd := as.(map[string]interface{})
		if _, ok := asd[bean3.Enabled]; ok {
			enable := asd[bean3.Enabled].(bool)
			if enable {
				hpaResourceRequest.IsEnable = enable
				hpaResourceRequest.ReqReplicaCount = templateMap[bean3.ReplicaCount].(float64)
				hpaResourceRequest.ReqMaxReplicas = asd["maxReplicaCount"].(float64)
				hpaResourceRequest.ReqMinReplicas = asd["minReplicaCount"].(float64)
				hpaResourceRequest.ResourceName = fmt.Sprintf("%s-%s-%s", "keda-hpa", appName, "keda")
				return hpaResourceRequest
			}
		}
	}

	if _, ok := templateMap[autoscaling.ServiceName]; ok {
		as := templateMap[autoscaling.ServiceName]
		asd := as.(map[string]interface{})
		if _, ok := asd[bean3.Enabled]; ok {
			enable := asd[bean3.Enabled].(bool)
			if enable {
				hpaResourceRequest.IsEnable = asd[bean3.Enabled].(bool)
				hpaResourceRequest.ReqReplicaCount = templateMap[bean3.ReplicaCount].(float64)
				hpaResourceRequest.ReqMaxReplicas = asd["MaxReplicas"].(float64)
				hpaResourceRequest.ReqMinReplicas = asd["MinReplicas"].(float64)
				hpaResourceRequest.ResourceName = fmt.Sprintf("%s-%s", appName, "hpa")
				return hpaResourceRequest
			}
		}
	}
	return hpaResourceRequest

}

func createConfigMapAndSecretJsonRequest(overrideRequest *bean.ValuesOverrideRequest, envOverride *chartConfig.EnvConfigOverride, chartVersion string, scope resourceQualifiers.Scope) bean3.ConfigMapAndSecretJsonV2 {
	request := bean3.ConfigMapAndSecretJsonV2{
		AppId:                                 overrideRequest.AppId,
		EnvId:                                 envOverride.TargetEnvironment,
		PipeLineId:                            overrideRequest.PipelineId,
		ChartVersion:                          chartVersion,
		DeploymentWithConfig:                  overrideRequest.DeploymentWithConfig,
		WfrIdForDeploymentWithSpecificTrigger: overrideRequest.WfrIdForDeploymentWithSpecificTrigger,
		Scope:                                 scope,
	}
	return request
}

func getScopeForVariables(overrideRequest *bean.ValuesOverrideRequest, envOverride *chartConfig.EnvConfigOverride) resourceQualifiers.Scope {
	scope := resourceQualifiers.Scope{
		AppId:     overrideRequest.AppId,
		EnvId:     envOverride.TargetEnvironment,
		ClusterId: envOverride.Environment.Id,
		SystemMetadata: &resourceQualifiers.SystemMetadata{
			EnvironmentName: envOverride.Environment.Name,
			ClusterName:     envOverride.Environment.Cluster.ClusterName,
			Namespace:       envOverride.Environment.Namespace,
			ImageTag:        util3.GetImageTagFromImage(overrideRequest.Image),
			AppName:         overrideRequest.AppName,
			Image:           overrideRequest.Image,
		},
	}
	return scope
}
