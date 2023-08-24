package app

//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
//	"github.com/aws/aws-sdk-go/service/autoscaling"
//	"github.com/devtron-labs/devtron/api/bean"
//	"github.com/devtron-labs/devtron/client/argocdServer/application"
//	application3 "github.com/devtron-labs/devtron/client/k8s/application"
//	"github.com/devtron-labs/devtron/internal/middleware"
//	"github.com/devtron-labs/devtron/internal/sql/models"
//	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
//	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
//	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
//	"github.com/devtron-labs/devtron/internal/util"
//	bean2 "github.com/devtron-labs/devtron/pkg/bean"
//	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
//	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
//	"github.com/devtron-labs/devtron/pkg/dockerRegistry"
//	repository "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
//	"github.com/devtron-labs/devtron/pkg/sql"
//	util2 "github.com/devtron-labs/devtron/util"
//	"github.com/devtron-labs/devtron/util/k8s"
//	"github.com/go-pg/pg"
//	"github.com/juju/errors"
//	"github.com/tidwall/gjson"
//	"github.com/tidwall/sjson"
//	"go.opentelemetry.io/otel"
//	"go.uber.org/zap"
//	errors2 "k8s.io/apimachinery/pkg/api/errors"
//	"k8s.io/apimachinery/pkg/runtime/schema"
//	"strconv"
//	"strings"
//	"time"
//)
//
//type AppManifestService interface {
//	GetValuesOverrideForTrigger(overrideRequest *bean.ValuesOverrideRequest)
//}
//
//type AppManifestServiceImpl struct {
//	pipelineRepository                  pipelineConfig.PipelineRepository
//	deploymentTemplateHistoryRepository repository.DeploymentTemplateHistoryRepository
//	chartRefRepository                  chartRepoRepository.ChartRefRepository
//	chartRepository                     chartRepoRepository.ChartRepository
//	environmentConfigRepository         chartConfig.EnvConfigOverrideRepository
//	envRepository                       repository2.EnvironmentRepository
//	configMapHistoryRepository          repository.ConfigMapHistoryRepository
//	strategyHistoryRepository           repository.PipelineStrategyHistoryRepository
//	appLevelMetricsRepository           repository3.AppLevelMetricsRepository
//	envLevelMetricsRepository           repository3.EnvLevelAppMetricsRepository
//	pipelineConfigRepository            chartConfig.PipelineConfigRepository
//	ciArtifactRepository                repository3.CiArtifactRepository
//	dbMigrationConfigRepository         pipelineConfig.DbMigrationConfigRepository
//	configMapRepository                 chartConfig.ConfigMapRepository
//	mergeUtil                           util.MergeUtil
//	pipelineOverrideRepository          chartConfig.PipelineOverrideRepository
//	appCrudOperationService             AppCrudOperationService
//	logger                              *zap.SugaredLogger
//	acdClient                           application.ServiceClient
//	k8sApplicationService               k8s.K8sApplicationService
//	dockerRegistryIpsConfigService      dockerRegistry.DockerRegistryIpsConfigService
//}
//
//func NewAppManifestServiceImpl(
//	logger *zap.SugaredLogger,
//	pipelineRepository pipelineConfig.PipelineRepository,
//	deploymentTemplateHistoryRepository repository.DeploymentTemplateHistoryRepository,
//	chartRefRepository chartRepoRepository.ChartRefRepository,
//	chartRepository chartRepoRepository.ChartRepository,
//	environmentConfigRepository chartConfig.EnvConfigOverrideRepository,
//	envRepository repository2.EnvironmentRepository,
//	configMapHistoryRepository repository.ConfigMapHistoryRepository,
//	strategyHistoryRepository repository.PipelineStrategyHistoryRepository,
//	appLevelMetricsRepository repository3.AppLevelMetricsRepository,
//	envLevelMetricsRepository repository3.EnvLevelAppMetricsRepository,
//	pipelineConfigRepository chartConfig.PipelineConfigRepository,
//	ciArtifactRepository repository3.CiArtifactRepository,
//	dbMigrationConfigRepository pipelineConfig.DbMigrationConfigRepository,
//	configMapRepository chartConfig.ConfigMapRepository,
//	mergeUtil util.MergeUtil,
//	appCrudOperationService AppCrudOperationService,
//	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
//	acdClient application.ServiceClient,
//	k8sApplicationService k8s.K8sApplicationService,
//	dockerRegistryIpsConfigService dockerRegistry.DockerRegistryIpsConfigService,
//) *AppManifestServiceImpl {
//	return &AppManifestServiceImpl{
//		logger:                              logger,
//		pipelineRepository:                  pipelineRepository,
//		deploymentTemplateHistoryRepository: deploymentTemplateHistoryRepository,
//		chartRefRepository:                  chartRefRepository,
//		chartRepository:                     chartRepository,
//		environmentConfigRepository:         environmentConfigRepository,
//		envRepository:                       envRepository,
//		configMapHistoryRepository:          configMapHistoryRepository,
//		strategyHistoryRepository:           strategyHistoryRepository,
//		appLevelMetricsRepository:           appLevelMetricsRepository,
//		envLevelMetricsRepository:           envLevelMetricsRepository,
//		pipelineConfigRepository:            pipelineConfigRepository,
//		ciArtifactRepository:                ciArtifactRepository,
//		dbMigrationConfigRepository:         dbMigrationConfigRepository,
//		configMapRepository:                 configMapRepository,
//		mergeUtil:                           mergeUtil,
//		appCrudOperationService:             appCrudOperationService,
//		pipelineOverrideRepository:          pipelineOverrideRepository,
//		acdClient:                           acdClient,
//		k8sApplicationService:               k8sApplicationService,
//		dockerRegistryIpsConfigService:      dockerRegistryIpsConfigService,
//	}
//}
//
//func (impl *AppManifestServiceImpl) getReleaseOverride(envOverride *chartConfig.EnvConfigOverride,
//	overrideRequest *bean.ValuesOverrideRequest,
//	artifact *repository3.CiArtifact,
//	pipeline *pipelineConfig.Pipeline,
//	pipelineOverride *chartConfig.PipelineOverride, strategy *chartConfig.PipelineStrategy, appMetrics *bool) (releaseOverride string, err error) {
//
//	artifactImage := artifact.Image
//	imageTag := strings.Split(artifactImage, ":")
//
//	imageTagLen := len(imageTag)
//
//	imageName := ""
//
//	for i := 0; i < imageTagLen-1; i++ {
//		if i != imageTagLen-2 {
//			imageName = imageName + imageTag[i] + ":"
//		} else {
//			imageName = imageName + imageTag[i]
//		}
//	}
//
//	appId := strconv.Itoa(pipeline.App.Id)
//	envId := strconv.Itoa(pipeline.EnvironmentId)
//
//	deploymentStrategy := ""
//	if strategy != nil {
//		deploymentStrategy = string(strategy.Strategy)
//	}
//	releaseAttribute := ReleaseAttributes{
//		Name:           imageName,
//		Tag:            imageTag[imageTagLen-1],
//		PipelineName:   pipeline.Name,
//		ReleaseVersion: strconv.Itoa(pipelineOverride.PipelineReleaseCounter),
//		DeploymentType: deploymentStrategy,
//		App:            appId,
//		Env:            envId,
//		AppMetrics:     appMetrics,
//	}
//	override, err := util2.Tprintf(envOverride.Chart.ImageDescriptorTemplate, releaseAttribute)
//	if err != nil {
//		return "", &util.ApiError{InternalMessage: "unable to render ImageDescriptorTemplate"}
//	}
//	if overrideRequest.AdditionalOverride != nil {
//		userOverride, err := overrideRequest.AdditionalOverride.MarshalJSON()
//		if err != nil {
//			return "", err
//		}
//		data, err := impl.mergeUtil.JsonPatch(userOverride, []byte(override))
//		if err != nil {
//			return "", err
//		}
//		override = string(data)
//	}
//	return override, nil
//}
//
//func (impl *AppManifestServiceImpl) checkAndFixDuplicateReleaseNo(override *chartConfig.PipelineOverride) error {
//
//	uniqueVerified := false
//	retryCount := 0
//
//	for !uniqueVerified && retryCount < 5 {
//		retryCount = retryCount + 1
//		overrides, err := impl.pipelineOverrideRepository.GetByPipelineIdAndReleaseNo(override.PipelineId, override.PipelineReleaseCounter)
//		if err != nil {
//			return err
//		}
//		if overrides[0].Id == override.Id {
//			uniqueVerified = true
//		} else {
//			//duplicate might be due to concurrency, lets fix it
//			currentReleaseNo, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(override.PipelineId)
//			if err != nil {
//				return err
//			}
//			override.PipelineReleaseCounter = currentReleaseNo + 1
//			err = impl.pipelineOverrideRepository.Save(override)
//			if err != nil {
//				return err
//			}
//		}
//	}
//	if !uniqueVerified {
//		return fmt.Errorf("duplicate verification retry count exide max overrideId: %d ,count: %d", override.Id, retryCount)
//	}
//	return nil
//}
//
//func (impl *AppManifestServiceImpl) savePipelineOverride(overrideRequest *bean.ValuesOverrideRequest, envOverrideId int, triggeredAt time.Time) (override *chartConfig.PipelineOverride, err error) {
//	currentReleaseNo, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(overrideRequest.PipelineId)
//	if err != nil {
//		return nil, err
//	}
//	po := &chartConfig.PipelineOverride{
//		EnvConfigOverrideId:    envOverrideId,
//		Status:                 models.CHARTSTATUS_NEW,
//		PipelineId:             overrideRequest.PipelineId,
//		CiArtifactId:           overrideRequest.CiArtifactId,
//		PipelineReleaseCounter: currentReleaseNo + 1,
//		CdWorkflowId:           overrideRequest.CdWorkflowId,
//		AuditLog:               sql.AuditLog{CreatedBy: overrideRequest.UserId, CreatedOn: triggeredAt, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
//		DeploymentType:         overrideRequest.DeploymentType,
//	}
//
//	err = impl.pipelineOverrideRepository.Save(po)
//	if err != nil {
//		return nil, err
//	}
//	err = impl.checkAndFixDuplicateReleaseNo(po)
//	if err != nil {
//		impl.logger.Errorw("error in checking release no duplicacy", "pipeline", po, "err", err)
//		return nil, err
//	}
//	return po, nil
//}
//
//func (impl *AppManifestServiceImpl) getConfigMapAndSecretJsonV2(appId int, envId int, pipelineId int, chartVersion string, deploymentWithConfig bean.DeploymentConfigurationType, wfrIdForDeploymentWithSpecificTrigger int) ([]byte, error) {
//
//	var configMapJson string
//	var secretDataJson string
//	var configMapJsonApp string
//	var secretDataJsonApp string
//	var configMapJsonEnv string
//	var secretDataJsonEnv string
//	var err error
//	//var configMapJsonPipeline string
//	//var secretDataJsonPipeline string
//
//	merged := []byte("{}")
//	if deploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
//		configMapA, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
//		if err != nil && pg.ErrNoRows != err {
//			return []byte("{}"), err
//		}
//		if configMapA != nil && configMapA.Id > 0 {
//			configMapJsonApp = configMapA.ConfigMapData
//			secretDataJsonApp = configMapA.SecretData
//		}
//		configMapE, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
//		if err != nil && pg.ErrNoRows != err {
//			return []byte("{}"), err
//		}
//		if configMapE != nil && configMapE.Id > 0 {
//			configMapJsonEnv = configMapE.ConfigMapData
//			secretDataJsonEnv = configMapE.SecretData
//		}
//	} else if deploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
//		//fetching history and setting envLevelConfig and not appLevelConfig because history already contains merged appLevel and envLevel configs
//		configMapHistory, err := impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrIdForDeploymentWithSpecificTrigger, repository.CONFIGMAP_TYPE)
//		if err != nil {
//			impl.logger.Errorw("error in getting config map history config by pipelineId and wfrId ", "err", err, "pipelineId", pipelineId, "wfrid", wfrIdForDeploymentWithSpecificTrigger)
//			return []byte("{}"), err
//		}
//		configMapJsonEnv = configMapHistory.Data
//		secretHistory, err := impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrIdForDeploymentWithSpecificTrigger, repository.SECRET_TYPE)
//		if err != nil {
//			impl.logger.Errorw("error in getting config map history config by pipelineId and wfrId ", "err", err, "pipelineId", pipelineId, "wfrid", wfrIdForDeploymentWithSpecificTrigger)
//			return []byte("{}"), err
//		}
//		secretDataJsonEnv = secretHistory.Data
//	}
//	configMapJson, err = impl.mergeUtil.ConfigMapMerge(configMapJsonApp, configMapJsonEnv)
//	if err != nil {
//		return []byte("{}"), err
//	}
//	chartMajorVersion, chartMinorVersion, err := util2.ExtractChartVersion(chartVersion)
//	if err != nil {
//		impl.logger.Errorw("chart version parsing", "err", err)
//		return []byte("{}"), err
//	}
//	secretDataJson, err = impl.mergeUtil.ConfigSecretMerge(secretDataJsonApp, secretDataJsonEnv, chartMajorVersion, chartMinorVersion)
//	if err != nil {
//		return []byte("{}"), err
//	}
//	configResponseR := bean.ConfigMapRootJson{}
//	configResponse := bean.ConfigMapJson{}
//	if configMapJson != "" {
//		err = json.Unmarshal([]byte(configMapJson), &configResponse)
//		if err != nil {
//			return []byte("{}"), err
//		}
//	}
//	configResponseR.ConfigMapJson = configResponse
//	secretResponseR := bean.ConfigSecretRootJson{}
//	secretResponse := bean.ConfigSecretJson{}
//	if configMapJson != "" {
//		err = json.Unmarshal([]byte(secretDataJson), &secretResponse)
//		if err != nil {
//			return []byte("{}"), err
//		}
//	}
//	secretResponseR.ConfigSecretJson = secretResponse
//
//	configMapByte, err := json.Marshal(configResponseR)
//	if err != nil {
//		return []byte("{}"), err
//	}
//	secretDataByte, err := json.Marshal(secretResponseR)
//	if err != nil {
//		return []byte("{}"), err
//	}
//
//	merged, err = impl.mergeUtil.JsonPatch(configMapByte, secretDataByte)
//	if err != nil {
//		return []byte("{}"), err
//	}
//	return merged, nil
//}
//
//func (impl *AppManifestServiceImpl) GetEnvOverrideByTriggerType(overrideRequest *bean.ValuesOverrideRequest, envId int, triggeredAt time.Time, ctx context.Context) (*chartConfig.EnvConfigOverride, error) {
//
//	envOverride := &chartConfig.EnvConfigOverride{}
//	var err error
//	if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
//		_, span := otel.Tracer("orchestrator").Start(ctx, "deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId")
//		deploymentTemplateHistory, err := impl.deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId(overrideRequest.PipelineId, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
//		span.End()
//		if err != nil {
//			impl.logger.Errorw("error in getting deployed deployment template history by pipelineId and wfrId", "err", err, "pipelineId", &overrideRequest, "wfrId", overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
//			return nil, err
//		}
//		templateName := deploymentTemplateHistory.TemplateName
//		templateVersion := deploymentTemplateHistory.TemplateVersion
//		if templateName == "Rollout Deployment" {
//			templateName = ""
//		}
//		//getting chart_ref by id
//		_, span = otel.Tracer("orchestrator").Start(ctx, "chartRefRepository.FindByVersionAndName")
//		chartRef, err := impl.chartRefRepository.FindByVersionAndName(templateName, templateVersion)
//		span.End()
//		if err != nil {
//			impl.logger.Errorw("error in getting chartRef by version and name", "err", err, "version", templateVersion, "name", templateName)
//			return nil, err
//		}
//		//assuming that if a chartVersion is deployed then it's envConfigOverride will be available
//		_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.GetByAppIdEnvIdAndChartRefId")
//		envOverride, err = impl.environmentConfigRepository.GetByAppIdEnvIdAndChartRefId(overrideRequest.AppId, envId, chartRef.Id)
//		span.End()
//		if err != nil {
//			impl.logger.Errorw("error in getting envConfigOverride for pipeline for specific chartVersion", "err", err, "appId", overrideRequest.AppId, "envId", envId, "chartRefId", chartRef.Id)
//			return nil, err
//		}
//		//updating historical data in envConfigOverride and appMetrics flag
//		envOverride.IsOverride = true
//		envOverride.EnvOverrideValues = deploymentTemplateHistory.Template
//
//	} else if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
//		_, span := otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.ActiveEnvConfigOverride")
//		envOverride, err = impl.environmentConfigRepository.ActiveEnvConfigOverride(overrideRequest.AppId, envId)
//		span.End()
//		if err != nil {
//			impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
//			return nil, err
//		}
//		if envOverride.Id == 0 {
//			_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindLatestChartForAppByAppId")
//			chart, err := impl.chartRepository.FindLatestChartForAppByAppId(overrideRequest.AppId)
//			span.End()
//			if err != nil {
//				impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
//				return nil, err
//			}
//			_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.FindChartByAppIdAndEnvIdAndChartRefId")
//			envOverride, err = impl.environmentConfigRepository.FindChartByAppIdAndEnvIdAndChartRefId(overrideRequest.AppId, envId, chart.ChartRefId)
//			span.End()
//			if err != nil && !errors2.IsNotFound(err) {
//				impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
//				return nil, err
//			}
//
//			//creating new env override config
//			if errors2.IsNotFound(err) || envOverride == nil {
//				_, span = otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
//				environment, err := impl.envRepository.FindById(envId)
//				span.End()
//				if err != nil && !util.IsErrNoRows(err) {
//					return nil, err
//				}
//				envOverride = &chartConfig.EnvConfigOverride{
//					Active:            true,
//					ManualReviewed:    true,
//					Status:            models.CHARTSTATUS_SUCCESS,
//					TargetEnvironment: envId,
//					ChartId:           chart.Id,
//					AuditLog:          sql.AuditLog{UpdatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId},
//					Namespace:         environment.Namespace,
//					IsOverride:        false,
//					EnvOverrideValues: "{}",
//					Latest:            false,
//					IsBasicViewLocked: chart.IsBasicViewLocked,
//					CurrentViewEditor: chart.CurrentViewEditor,
//				}
//				_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.Save")
//				err = impl.environmentConfigRepository.Save(envOverride)
//				span.End()
//				if err != nil {
//					impl.logger.Errorw("error in creating envconfig", "data", envOverride, "error", err)
//					return nil, err
//				}
//			}
//			envOverride.Chart = chart
//		} else if envOverride.Id > 0 && !envOverride.IsOverride {
//			_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindLatestChartForAppByAppId")
//			chart, err := impl.chartRepository.FindLatestChartForAppByAppId(overrideRequest.AppId)
//			span.End()
//			if err != nil {
//				impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
//				return nil, err
//			}
//			envOverride.Chart = chart
//		}
//	}
//	_, span := otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
//	env, err := impl.envRepository.FindById(envOverride.TargetEnvironment)
//	span.End()
//	if err != nil {
//		impl.logger.Errorw("unable to find env", "err", err)
//		return nil, err
//	}
//	envOverride.Environment = env
//	return envOverride, nil
//}
//
//func (impl *AppManifestServiceImpl) GetAppMetricsByTriggerType(overrideRequest *bean.ValuesOverrideRequest, envId int, ctx context.Context) (bool, error) {
//
//	var appMetrics bool
//	if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
//		_, span := otel.Tracer("orchestrator").Start(ctx, "deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId")
//		deploymentTemplateHistory, err := impl.deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId(overrideRequest.PipelineId, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
//		span.End()
//		if err != nil {
//			impl.logger.Errorw("error in getting deployed deployment template history by pipelineId and wfrId", "err", err, "pipelineId", &overrideRequest, "wfrId", overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
//			return appMetrics, err
//		}
//		appMetrics = deploymentTemplateHistory.IsAppMetricsEnabled
//
//	} else if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
//		_, span := otel.Tracer("orchestrator").Start(ctx, "appLevelMetricsRepository.FindByAppId")
//		appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(overrideRequest.AppId)
//		span.End()
//		if err != nil && !util.IsErrNoRows(err) {
//			impl.logger.Errorw("err", err)
//			return appMetrics, &util.ApiError{InternalMessage: "unable to fetch app level metrics flag"}
//		}
//		appMetrics = appLevelMetrics.AppMetrics
//
//		_, span = otel.Tracer("orchestrator").Start(ctx, "envLevelMetricsRepository.FindByAppIdAndEnvId")
//		envLevelMetrics, err := impl.envLevelMetricsRepository.FindByAppIdAndEnvId(overrideRequest.AppId, envId)
//		span.End()
//		if err != nil && !util.IsErrNoRows(err) {
//			impl.logger.Errorw("err", err)
//			return appMetrics, &util.ApiError{InternalMessage: "unable to fetch env level metrics flag"}
//		}
//		if envLevelMetrics.Id != 0 && envLevelMetrics.AppMetrics != nil {
//			appMetrics = *envLevelMetrics.AppMetrics
//		}
//	}
//	return appMetrics, nil
//}
//
//func (impl *AppManifestServiceImpl) getDeploymentStrategyByTriggerType(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context) (*chartConfig.PipelineStrategy, error) {
//
//	strategy := &chartConfig.PipelineStrategy{}
//	var err error
//	if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
//		_, span := otel.Tracer("orchestrator").Start(ctx, "strategyHistoryRepository.GetHistoryByPipelineIdAndWfrId")
//		strategyHistory, err := impl.strategyHistoryRepository.GetHistoryByPipelineIdAndWfrId(overrideRequest.PipelineId, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
//		span.End()
//		if err != nil {
//			impl.logger.Errorw("error in getting deployed strategy history by pipleinId and wfrId", "err", err, "pipelineId", overrideRequest.PipelineId, "wfrId", overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
//			return nil, err
//		}
//		strategy.Strategy = strategyHistory.Strategy
//		strategy.Config = strategyHistory.Config
//		strategy.PipelineId = overrideRequest.PipelineId
//	} else if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
//		if overrideRequest.ForceTrigger {
//			_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineConfigRepository.GetDefaultStrategyByPipelineId")
//			strategy, err = impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(overrideRequest.PipelineId)
//			span.End()
//		} else {
//			var deploymentTemplate chartRepoRepository.DeploymentStrategy
//			if overrideRequest.DeploymentTemplate == "ROLLING" {
//				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_ROLLING
//			} else if overrideRequest.DeploymentTemplate == "BLUE-GREEN" {
//				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_BLUE_GREEN
//			} else if overrideRequest.DeploymentTemplate == "CANARY" {
//				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_CANARY
//			} else if overrideRequest.DeploymentTemplate == "RECREATE" {
//				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_RECREATE
//			}
//
//			if len(deploymentTemplate) > 0 {
//				_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineConfigRepository.FindByStrategyAndPipelineId")
//				strategy, err = impl.pipelineConfigRepository.FindByStrategyAndPipelineId(deploymentTemplate, overrideRequest.PipelineId)
//				span.End()
//			} else {
//				_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineConfigRepository.GetDefaultStrategyByPipelineId")
//				strategy, err = impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(overrideRequest.PipelineId)
//				span.End()
//			}
//		}
//		if err != nil && errors.IsNotFound(err) == false {
//			impl.logger.Errorf("invalid state", "err", err, "req", strategy)
//			return nil, err
//		}
//	}
//	return strategy, nil
//}
//
//func (impl *AppManifestServiceImpl) getDbMigrationOverride(overrideRequest *bean.ValuesOverrideRequest, artifact *repository3.CiArtifact, isRollback bool) (overrideJson []byte, err error) {
//	if isRollback {
//		return nil, fmt.Errorf("rollback not supported ye")
//	}
//	notConfigured := false
//	config, err := impl.dbMigrationConfigRepository.FindByPipelineId(overrideRequest.PipelineId)
//	if err != nil && !util.IsErrNoRows(err) {
//		impl.logger.Errorw("error in fetching pipeline override config", "req", overrideRequest, "err", err)
//		return nil, err
//	} else if util.IsErrNoRows(err) {
//		notConfigured = true
//	}
//	envVal := &EnvironmentOverride{}
//	if notConfigured {
//		impl.logger.Warnw("no active db migration found", "pipeline", overrideRequest.PipelineId)
//		envVal.Enabled = false
//	} else {
//		materialInfos, err := artifact.ParseMaterialInfo()
//		if err != nil {
//			return nil, err
//		}
//
//		hash, ok := materialInfos[config.GitMaterial.Url]
//		if !ok {
//			impl.logger.Errorf("wrong url map ", "map", materialInfos, "url", config.GitMaterial.Url)
//			return nil, fmt.Errorf("configured url not found in material %s", config.GitMaterial.Url)
//		}
//
//		envVal.Enabled = true
//		if config.GitMaterial.GitProvider.AuthMode != repository3.AUTH_MODE_USERNAME_PASSWORD &&
//			config.GitMaterial.GitProvider.AuthMode != repository3.AUTH_MODE_ACCESS_TOKEN &&
//			config.GitMaterial.GitProvider.AuthMode != repository3.AUTH_MODE_ANONYMOUS {
//			return nil, fmt.Errorf("auth mode %s not supported for migration", config.GitMaterial.GitProvider.AuthMode)
//		}
//		envVal.appendEnvironmentVariable("GIT_REPO_URL", config.GitMaterial.Url)
//		envVal.appendEnvironmentVariable("GIT_USER", config.GitMaterial.GitProvider.UserName)
//		var password string
//		if config.GitMaterial.GitProvider.AuthMode == repository3.AUTH_MODE_USERNAME_PASSWORD {
//			password = config.GitMaterial.GitProvider.Password
//		} else {
//			password = config.GitMaterial.GitProvider.AccessToken
//		}
//		envVal.appendEnvironmentVariable("GIT_AUTH_TOKEN", password)
//		// parse git-tag not required
//		//envVal.appendEnvironmentVariable("GIT_TAG", "")
//		envVal.appendEnvironmentVariable("GIT_HASH", hash)
//		envVal.appendEnvironmentVariable("SCRIPT_LOCATION", config.ScriptSource)
//		envVal.appendEnvironmentVariable("DB_TYPE", string(config.DbConfig.Type))
//		envVal.appendEnvironmentVariable("DB_USER_NAME", config.DbConfig.UserName)
//		envVal.appendEnvironmentVariable("DB_PASSWORD", config.DbConfig.Password)
//		envVal.appendEnvironmentVariable("DB_HOST", config.DbConfig.Host)
//		envVal.appendEnvironmentVariable("DB_PORT", config.DbConfig.Port)
//		envVal.appendEnvironmentVariable("DB_NAME", config.DbConfig.DbName)
//		//Will be used for rollback don't delete it
//		//envVal.appendEnvironmentVariable("MIGRATE_TO_VERSION", strconv.Itoa(overrideRequest.TargetDbVersion))
//	}
//	dbMigrationConfig := map[string]interface{}{"dbMigrationConfig": envVal}
//	confByte, err := json.Marshal(dbMigrationConfig)
//	if err != nil {
//		return nil, err
//	}
//	return confByte, nil
//}
//
//func (impl *AppManifestServiceImpl) mergeOverrideValues(envOverride *chartConfig.EnvConfigOverride,
//	dbMigrationOverride []byte,
//	releaseOverrideJson string,
//	configMapJson []byte,
//	appLabelJsonByte []byte,
//	strategy *chartConfig.PipelineStrategy,
//) (mergedValues []byte, err error) {
//
//	//merge three values on the fly
//	//ordering is important here
//	//global < environment < db< release
//	var merged []byte
//	if !envOverride.IsOverride {
//		merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), []byte(envOverride.Chart.GlobalOverride))
//		if err != nil {
//			return nil, err
//		}
//	} else {
//		merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), []byte(envOverride.EnvOverrideValues))
//		if err != nil {
//			return nil, err
//		}
//	}
//	if strategy != nil && len(strategy.Config) > 0 {
//		merged, err = impl.mergeUtil.JsonPatch(merged, []byte(strategy.Config))
//		if err != nil {
//			return nil, err
//		}
//	}
//	merged, err = impl.mergeUtil.JsonPatch(merged, dbMigrationOverride)
//	if err != nil {
//		return nil, err
//	}
//	merged, err = impl.mergeUtil.JsonPatch(merged, []byte(releaseOverrideJson))
//	if err != nil {
//		return nil, err
//	}
//	if configMapJson != nil {
//		merged, err = impl.mergeUtil.JsonPatch(merged, configMapJson)
//		if err != nil {
//			return nil, err
//		}
//	}
//	if appLabelJsonByte != nil {
//		merged, err = impl.mergeUtil.JsonPatch(merged, appLabelJsonByte)
//		if err != nil {
//			return nil, err
//		}
//	}
//	return merged, nil
//}
//
//func (impl *AppManifestServiceImpl) fetchRequiredReplicaCount(currentReplicaCount float64, reqMaxReplicas float64, reqMinReplicas float64) float64 {
//	var reqReplicaCount float64
//	if currentReplicaCount <= reqMaxReplicas && currentReplicaCount >= reqMinReplicas {
//		reqReplicaCount = currentReplicaCount
//	} else if currentReplicaCount > reqMaxReplicas {
//		reqReplicaCount = reqMaxReplicas
//	} else if currentReplicaCount < reqMinReplicas {
//		reqReplicaCount = reqMinReplicas
//	}
//	return reqReplicaCount
//}
//
//func (impl *AppManifestServiceImpl) setScalingValues(templateMap map[string]interface{}, customScalingKey string, merged []byte, value interface{}) ([]byte, error) {
//	autoscalingJsonPath := templateMap[customScalingKey]
//	autoscalingJsonPathKey := autoscalingJsonPath.(string)
//	mergedRes, err := sjson.Set(string(merged), autoscalingJsonPathKey, value)
//	if err != nil {
//		impl.logger.Errorw("error occurred while setting autoscaling key", "JsonPathKey", autoscalingJsonPathKey, "err", err)
//		return []byte{}, err
//	}
//	return []byte(mergedRes), nil
//}
//
//func (impl *AppManifestServiceImpl) extractParamValue(inputMap map[string]interface{}, key string, merged []byte) (float64, error) {
//	if _, ok := inputMap[key]; !ok {
//		return 0, errors.New("empty-val-err")
//	}
//	floatNumber, err := util2.ParseFloatNumber(gjson.Get(string(merged), inputMap[key].(string)).Value())
//	if err != nil {
//		impl.logger.Errorw("error occurred while parsing float number", "key", key, "err", err)
//	}
//	return floatNumber, err
//}
//
//func (impl *AppManifestServiceImpl) getReplicaCountFromCustomChart(templateMap map[string]interface{}, merged []byte) (float64, error) {
//	autoscalingMinVal, err := impl.extractParamValue(templateMap, bean2.CustomAutoscalingMinPathKey, merged)
//	if err != nil {
//		return 0, err
//	}
//	autoscalingMaxVal, err := impl.extractParamValue(templateMap, bean2.CustomAutoscalingMaxPathKey, merged)
//	if err != nil {
//		return 0, err
//	}
//	autoscalingReplicaCountVal, err := impl.extractParamValue(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged)
//	if err != nil {
//		return 0, err
//	}
//	return impl.fetchRequiredReplicaCount(autoscalingReplicaCountVal, autoscalingMaxVal, autoscalingMinVal), nil
//}
//
//func (impl *AppManifestServiceImpl) autoscalingCheckBeforeTrigger(ctx context.Context, appName string, namespace string, merged []byte, pipeline *pipelineConfig.Pipeline, overrideRequest *bean.ValuesOverrideRequest) []byte {
//	var appId = pipeline.AppId
//	pipelineId := pipeline.Id
//	var appDeploymentType = pipeline.DeploymentAppType
//	var clusterId = pipeline.Environment.ClusterId
//	deploymentType := overrideRequest.DeploymentType
//	templateMap := make(map[string]interface{})
//	err := json.Unmarshal(merged, &templateMap)
//	if err != nil {
//		return merged
//	}
//	if _, ok := templateMap[autoscaling.ServiceName]; ok {
//		as := templateMap[autoscaling.ServiceName]
//		asd := as.(map[string]interface{})
//		isEnable := false
//		if _, ok := asd["enabled"]; ok {
//			isEnable = asd["enabled"].(bool)
//		}
//		if isEnable {
//			reqReplicaCount := templateMap["replicaCount"].(float64)
//			reqMaxReplicas := asd["MaxReplicas"].(float64)
//			reqMinReplicas := asd["MinReplicas"].(float64)
//			version := ""
//			group := autoscaling.ServiceName
//			kind := "HorizontalPodAutoscaler"
//			resourceName := fmt.Sprintf("%s-%s", appName, "hpa")
//			resourceManifest := make(map[string]interface{})
//			if util.IsAcdApp(appDeploymentType) {
//				query := &application2.ApplicationResourceRequest{
//					Name:         &appName,
//					Version:      &version,
//					Group:        &group,
//					Kind:         &kind,
//					ResourceName: &resourceName,
//					Namespace:    &namespace,
//				}
//				recv, err := impl.acdClient.GetResource(ctx, query)
//				impl.logger.Debugw("resource manifest get replica count", "response", recv)
//				if err != nil {
//					impl.logger.Errorw("ACD Get Resource API Failed", "err", err)
//					middleware.AcdGetResourceCounter.WithLabelValues(strconv.Itoa(appId), namespace, appName).Inc()
//					return merged
//				}
//				if recv != nil && len(*recv.Manifest) > 0 {
//					err := json.Unmarshal([]byte(*recv.Manifest), &resourceManifest)
//					if err != nil {
//						impl.logger.Errorw("unmarshal failed for hpa check", "err", err)
//						return merged
//					}
//				}
//			} else {
//				version = "v2beta2"
//				k8sResource, err := impl.k8sApplicationService.GetResource(ctx, &k8s.ResourceRequestBean{ClusterId: clusterId,
//					K8sRequest: &application3.K8sRequestBean{ResourceIdentifier: application3.ResourceIdentifier{Name: resourceName,
//						Namespace: namespace, GroupVersionKind: schema.GroupVersionKind{Group: group, Kind: kind, Version: version}}}})
//				if err != nil {
//					impl.logger.Errorw("error occurred while fetching resource for app", "resourceName", resourceName, "err", err)
//					return merged
//				}
//				resourceManifest = k8sResource.Manifest.Object
//			}
//			if len(resourceManifest) > 0 {
//				statusMap := resourceManifest["status"].(map[string]interface{})
//				currentReplicaVal := statusMap["currentReplicas"]
//				currentReplicaCount, err := util2.ParseFloatNumber(currentReplicaVal)
//				if err != nil {
//					impl.logger.Errorw("error occurred while parsing replica count", "currentReplicas", currentReplicaVal, "err", err)
//					return merged
//				}
//
//				reqReplicaCount = impl.fetchRequiredReplicaCount(currentReplicaCount, reqMaxReplicas, reqMinReplicas)
//				templateMap["replicaCount"] = reqReplicaCount
//				merged, err = json.Marshal(&templateMap)
//				if err != nil {
//					impl.logger.Errorw("marshaling failed for hpa check", "err", err)
//					return merged
//				}
//			}
//		} else {
//			impl.logger.Errorw("autoscaling is not enabled", "pipelineId", pipelineId)
//		}
//	}
//	//check for custom chart support
//	if autoscalingEnabledPath, ok := templateMap[bean2.CustomAutoScalingEnabledPathKey]; ok {
//		if deploymentType == models.DEPLOYMENTTYPE_STOP {
//			merged, err = impl.setScalingValues(templateMap, bean2.CustomAutoScalingEnabledPathKey, merged, false)
//			if err != nil {
//				return merged
//			}
//			merged, err = impl.setScalingValues(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged, 0)
//			if err != nil {
//				return merged
//			}
//		} else {
//			autoscalingEnabled := false
//			autoscalingEnabledValue := gjson.Get(string(merged), autoscalingEnabledPath.(string)).Value()
//			if val, ok := autoscalingEnabledValue.(bool); ok {
//				autoscalingEnabled = val
//			}
//			if autoscalingEnabled {
//				// extract replica count, min, max and check for required value
//				replicaCount, err := impl.getReplicaCountFromCustomChart(templateMap, merged)
//				if err != nil {
//					return merged
//				}
//				merged, err = impl.setScalingValues(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged, replicaCount)
//				if err != nil {
//					return merged
//				}
//			}
//		}
//	}
//
//	return merged
//}
//
//func (impl *AppManifestServiceImpl) GetValuesOverrideForTrigger(overrideRequest *bean.ValuesOverrideRequest, triggeredAt time.Time, ctx context.Context) (ValuesOverrideResponse, error) {
//	if overrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
//		overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
//	}
//	if len(overrideRequest.DeploymentWithConfig) == 0 {
//		overrideRequest.DeploymentWithConfig = bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED
//	}
//	valuesOverrideResponse := ValuesOverrideResponse{}
//	pipeline, err := impl.pipelineRepository.FindById(overrideRequest.PipelineId)
//	if err != nil {
//		impl.logger.Errorw("error in fetching pipeline by pipelineId", "err", err)
//		return valuesOverrideResponse, err
//	}
//	envOverride, err := impl.GetEnvOverrideByTriggerType(overrideRequest, pipeline.EnvironmentId, triggeredAt, ctx)
//	if err != nil {
//		impl.logger.Errorw("error in getting env override by trigger type", "err", err)
//		return valuesOverrideResponse, err
//	}
//	appMetrics, err := impl.GetAppMetricsByTriggerType(overrideRequest, pipeline.EnvironmentId, ctx)
//	if err != nil {
//		impl.logger.Errorw("error in getting app metrics by trigger type", "err", err)
//		return valuesOverrideResponse, err
//	}
//	strategy, err := impl.getDeploymentStrategyByTriggerType(overrideRequest, ctx)
//	if err != nil {
//		impl.logger.Errorw("error in getting strategy by trigger type", "err", err)
//		return valuesOverrideResponse, err
//	}
//	_, span := otel.Tracer("orchestrator").Start(ctx, "ciArtifactRepository.Get")
//	artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
//	span.End()
//	if err != nil {
//		return valuesOverrideResponse, err
//	}
//	_, span = otel.Tracer("orchestrator").Start(ctx, "getDbMigrationOverride")
//	//FIXME: how to determine rollback
//	//we can't depend on ciArtifact ID because CI pipeline can be manually triggered in any order regardless of sourcecode status
//	dbMigrationOverride, err := impl.getDbMigrationOverride(overrideRequest, artifact, false)
//	span.End()
//	if err != nil {
//		impl.logger.Errorw("error in fetching db migration config", "req", overrideRequest, "err", err)
//		return valuesOverrideResponse, err
//	}
//	chartVersion := envOverride.Chart.ChartVersion
//	_, span = otel.Tracer("orchestrator").Start(ctx, "getConfigMapAndSecretJsonV2")
//	configMapJson, err := impl.getConfigMapAndSecretJsonV2(overrideRequest.AppId, envOverride.TargetEnvironment, overrideRequest.PipelineId, chartVersion, overrideRequest.DeploymentWithConfig, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
//	span.End()
//	if err != nil {
//		impl.logger.Errorw("error in fetching config map n secret ", "err", err)
//		configMapJson = nil
//	}
//	_, span = otel.Tracer("orchestrator").Start(ctx, "appCrudOperationService.GetLabelsByAppIdForDeployment")
//	appLabelJsonByte, err := impl.appCrudOperationService.GetLabelsByAppIdForDeployment(overrideRequest.AppId)
//	span.End()
//	if err != nil {
//		impl.logger.Errorw("error in fetching app labels for gitOps commit", "err", err)
//		appLabelJsonByte = nil
//	}
//	_, span = otel.Tracer("orchestrator").Start(ctx, "mergeAndSave")
//	pipelineOverride, err := impl.savePipelineOverride(overrideRequest, envOverride.Id, triggeredAt)
//	if err != nil {
//		return valuesOverrideResponse, err
//	}
//	//TODO: check status and apply lock
//	releaseOverrideJson, err := impl.getReleaseOverride(envOverride, overrideRequest, artifact, pipeline, pipelineOverride, strategy, &appMetrics)
//	if err != nil {
//		return valuesOverrideResponse, err
//	}
//	mergedValues, err := impl.mergeOverrideValues(envOverride, dbMigrationOverride, releaseOverrideJson, configMapJson, appLabelJsonByte, strategy)
//
//	appName := fmt.Sprintf("%s-%s", pipeline.App.AppName, envOverride.Environment.Name)
//	mergedValues = impl.autoscalingCheckBeforeTrigger(ctx, appName, envOverride.Namespace, mergedValues, pipeline, overrideRequest)
//
//	_, span = otel.Tracer("orchestrator").Start(ctx, "dockerRegistryIpsConfigService.HandleImagePullSecretOnApplicationDeployment")
//	// handle image pull secret if access given
//	mergedValues, err = impl.dockerRegistryIpsConfigService.HandleImagePullSecretOnApplicationDeployment(envOverride.Environment, overrideRequest.PipelineId, mergedValues)
//	span.End()
//	if err != nil {
//		return valuesOverrideResponse, err
//	}
//	//valuesOverrideResponse.
//	valuesOverrideResponse.MergedValues = string(mergedValues)
//	valuesOverrideResponse.Pipeline = pipeline
//	valuesOverrideResponse.EnvOverride = envOverride
//	valuesOverrideResponse.PipelineOverride = pipelineOverride
//	valuesOverrideResponse.AppMetrics = appMetrics
//	valuesOverrideResponse.PipelineStrategy = strategy
//	valuesOverrideResponse.ReleaseOverrideJSON = releaseOverrideJson
//	valuesOverrideResponse.Artifact = artifact
//	return valuesOverrideResponse, err
//}
