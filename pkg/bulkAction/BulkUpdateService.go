package bulkAction

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/bulkUpdate"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	appWorkflow2 "github.com/devtron-labs/devtron/pkg/appWorkflow"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/chart"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	pipeline1 "github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	repository5 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/rbac"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/go-pg/pg"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
	"net/http"
	"sort"
)

type BulkUpdateService interface {
	FindBulkUpdateReadme(operation string) (response *BulkUpdateSeeExampleResponse, err error)
	GetBulkAppName(bulkUpdateRequest *BulkUpdatePayload) (*ImpactedObjectsResponse, error)
	ApplyJsonPatch(patch jsonpatch.Patch, target string) (string, error)
	BulkUpdateDeploymentTemplate(bulkUpdatePayload *BulkUpdatePayload) *DeploymentTemplateBulkUpdateResponse
	BulkUpdateConfigMap(bulkUpdatePayload *BulkUpdatePayload) *CmAndSecretBulkUpdateResponse
	BulkUpdateSecret(bulkUpdatePayload *BulkUpdatePayload) *CmAndSecretBulkUpdateResponse
	BulkUpdate(bulkUpdateRequest *BulkUpdatePayload) (bulkUpdateResponse *BulkUpdateResponse)

	BulkHibernate(request *BulkApplicationForEnvironmentPayload, ctx context.Context, w http.ResponseWriter, token string, checkAuthForBulkActions func(token string, appObject string, envObject string) bool) (*BulkApplicationForEnvironmentResponse, error)
	BulkUnHibernate(request *BulkApplicationForEnvironmentPayload, ctx context.Context, w http.ResponseWriter, token string, checkAuthForBulkActions func(token string, appObject string, envObject string) bool) (*BulkApplicationForEnvironmentResponse, error)
	BulkDeploy(request *BulkApplicationForEnvironmentPayload, emailId string, checkAuthBatch func(emailId string, appObject []string, envObject []string) (map[string]bool, map[string]bool)) (*BulkApplicationForEnvironmentResponse, error)
	BulkBuildTrigger(request *BulkApplicationForEnvironmentPayload, ctx context.Context, w http.ResponseWriter, token string, checkAuthForBulkActions func(token string, appObject string, envObject string) bool) (*BulkApplicationForEnvironmentResponse, error)

	GetBulkActionImpactedPipelinesAndWfs(dto *CdBulkActionRequestDto) ([]*pipelineConfig.Pipeline, []int, []int, error)
	PerformBulkActionOnCdPipelines(dto *CdBulkActionRequestDto, impactedPipelines []*pipelineConfig.Pipeline, ctx context.Context, dryRun bool, impactedAppWfIds []int, impactedCiPipelineIds []int) (*PipelineAndWfBulkActionResponseDto, error)
}

type BulkUpdateServiceImpl struct {
	bulkUpdateRepository             bulkUpdate.BulkUpdateRepository
	chartRepository                  chartRepoRepository.ChartRepository
	logger                           *zap.SugaredLogger
	repoRepository                   chartRepoRepository.ChartRepoRepository
	chartTemplateService             util.ChartTemplateService
	mergeUtil                        util.MergeUtil
	repositoryService                repository.ServiceClient
	defaultChart                     chart.DefaultChart
	chartRefRepository               chartRepoRepository.ChartRefRepository
	envOverrideRepository            chartConfig.EnvConfigOverrideRepository
	pipelineConfigRepository         chartConfig.PipelineConfigRepository
	configMapRepository              chartConfig.ConfigMapRepository
	environmentRepository            repository2.EnvironmentRepository
	pipelineRepository               pipelineConfig.PipelineRepository
	appLevelMetricsRepository        repository3.AppLevelMetricsRepository
	envLevelAppMetricsRepository     repository3.EnvLevelAppMetricsRepository
	client                           *http.Client
	appRepository                    app.AppRepository
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService
	configMapHistoryService          history.ConfigMapHistoryService
	workflowDagExecutor              pipeline.WorkflowDagExecutor
	cdWorkflowRepository             pipelineConfig.CdWorkflowRepository
	pipelineBuilder                  pipeline.PipelineBuilder
	helmAppService                   client.HelmAppService
	enforcerUtil                     rbac.EnforcerUtil
	enforcerUtilHelm                 rbac.EnforcerUtilHelm
	ciHandler                        pipeline.CiHandler
	ciPipelineRepository             pipelineConfig.CiPipelineRepository
	appWorkflowRepository            appWorkflow.AppWorkflowRepository
	appWorkflowService               appWorkflow2.AppWorkflowService
	pubsubClient                     *pubsub.PubSubClientServiceImpl
	argoUserService                  argo.ArgoUserService
	variableEntityMappingService     variables.VariableEntityMappingService
	variableTemplateParser           parsers.VariableTemplateParser
}

func NewBulkUpdateServiceImpl(bulkUpdateRepository bulkUpdate.BulkUpdateRepository,
	chartRepository chartRepoRepository.ChartRepository,
	logger *zap.SugaredLogger,
	chartTemplateService util.ChartTemplateService,
	repoRepository chartRepoRepository.ChartRepoRepository,
	defaultChart chart.DefaultChart,
	mergeUtil util.MergeUtil,
	repositoryService repository.ServiceClient,
	chartRefRepository chartRepoRepository.ChartRefRepository,
	envOverrideRepository chartConfig.EnvConfigOverrideRepository,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	configMapRepository chartConfig.ConfigMapRepository,
	environmentRepository repository2.EnvironmentRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	appLevelMetricsRepository repository3.AppLevelMetricsRepository,
	envLevelAppMetricsRepository repository3.EnvLevelAppMetricsRepository,
	client *http.Client,
	appRepository app.AppRepository,
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService,
	configMapHistoryService history.ConfigMapHistoryService, workflowDagExecutor pipeline.WorkflowDagExecutor,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository, pipelineBuilder pipeline.PipelineBuilder,
	helmAppService client.HelmAppService, enforcerUtil rbac.EnforcerUtil,
	enforcerUtilHelm rbac.EnforcerUtilHelm, ciHandler pipeline.CiHandler,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	appWorkflowService appWorkflow2.AppWorkflowService,
	pubsubClient *pubsub.PubSubClientServiceImpl,
	argoUserService argo.ArgoUserService,
	variableEntityMappingService variables.VariableEntityMappingService,
	variableTemplateParser parsers.VariableTemplateParser) (*BulkUpdateServiceImpl, error) {
	impl := &BulkUpdateServiceImpl{
		bulkUpdateRepository:             bulkUpdateRepository,
		chartRepository:                  chartRepository,
		logger:                           logger,
		chartTemplateService:             chartTemplateService,
		repoRepository:                   repoRepository,
		mergeUtil:                        mergeUtil,
		defaultChart:                     defaultChart,
		repositoryService:                repositoryService,
		chartRefRepository:               chartRefRepository,
		envOverrideRepository:            envOverrideRepository,
		pipelineConfigRepository:         pipelineConfigRepository,
		configMapRepository:              configMapRepository,
		environmentRepository:            environmentRepository,
		pipelineRepository:               pipelineRepository,
		appLevelMetricsRepository:        appLevelMetricsRepository,
		envLevelAppMetricsRepository:     envLevelAppMetricsRepository,
		client:                           client,
		appRepository:                    appRepository,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
		configMapHistoryService:          configMapHistoryService,
		workflowDagExecutor:              workflowDagExecutor,
		cdWorkflowRepository:             cdWorkflowRepository,
		pipelineBuilder:                  pipelineBuilder,
		helmAppService:                   helmAppService,
		enforcerUtil:                     enforcerUtil,
		enforcerUtilHelm:                 enforcerUtilHelm,
		ciHandler:                        ciHandler,
		ciPipelineRepository:             ciPipelineRepository,
		appWorkflowRepository:            appWorkflowRepository,
		appWorkflowService:               appWorkflowService,
		pubsubClient:                     pubsubClient,
		argoUserService:                  argoUserService,
		variableTemplateParser:           variableTemplateParser,
		variableEntityMappingService:     variableEntityMappingService,
	}

	err := impl.SubscribeToCdBulkTriggerTopic()
	return impl, err
}

func (impl BulkUpdateServiceImpl) FindBulkUpdateReadme(operation string) (*BulkUpdateSeeExampleResponse, error) {
	bulkUpdateReadme, err := impl.bulkUpdateRepository.FindBulkUpdateReadme(operation)
	response := &BulkUpdateSeeExampleResponse{}
	if err != nil {
		impl.logger.Errorw("error in fetching batch operation example", "err", err)
		return response, err
	}
	script := &BulkUpdateScript{}
	err = json.Unmarshal([]byte(bulkUpdateReadme.Script), &script)
	if err != nil {
		impl.logger.Errorw("error in script value(in db) of batch operation example", "err", err)
		return response, err
	}
	response = &BulkUpdateSeeExampleResponse{
		Operation: bulkUpdateReadme.Resource,
		Script:    script,
		ReadMe:    bulkUpdateReadme.Readme,
	}
	return response, nil
}

func (impl BulkUpdateServiceImpl) GetBulkAppName(bulkUpdatePayload *BulkUpdatePayload) (*ImpactedObjectsResponse, error) {
	impactedObjectsResponse := &ImpactedObjectsResponse{}
	deploymentTemplateImpactedObjects := []*DeploymentTemplateImpactedObjectsResponseForOneApp{}
	configMapImpactedObjects := []*CmAndSecretImpactedObjectsResponseForOneApp{}
	secretImpactedObjects := []*CmAndSecretImpactedObjectsResponseForOneApp{}
	var appNameIncludes []string
	var appNameExcludes []string
	if bulkUpdatePayload.Includes == nil || len(bulkUpdatePayload.Includes.Names) == 0 {
		return impactedObjectsResponse, nil
	} else {
		appNameIncludes = bulkUpdatePayload.Includes.Names
	}
	if bulkUpdatePayload.Excludes != nil && len(bulkUpdatePayload.Excludes.Names) > 0 {
		appNameExcludes = bulkUpdatePayload.Excludes.Names
	}
	if bulkUpdatePayload.Global {
		//For Deployment Template
		if bulkUpdatePayload.DeploymentTemplate != nil && bulkUpdatePayload.DeploymentTemplate.Spec != nil {
			appsGlobalDT, err := impl.bulkUpdateRepository.
				FindDeploymentTemplateBulkAppNameForGlobal(appNameIncludes, appNameExcludes)
			if err != nil {
				impl.logger.Errorw("error in fetching bulk app names for global", "err", err)
				return nil, err
			}
			for _, app := range appsGlobalDT {
				deploymentTemplateImpactedObject := &DeploymentTemplateImpactedObjectsResponseForOneApp{
					AppId:   app.Id,
					AppName: app.AppName,
				}
				deploymentTemplateImpactedObjects = append(deploymentTemplateImpactedObjects, deploymentTemplateImpactedObject)
			}
		}

		//For ConfigMap
		if bulkUpdatePayload.ConfigMap != nil && bulkUpdatePayload.ConfigMap.Spec != nil && len(bulkUpdatePayload.ConfigMap.Spec.Names) != 0 {
			configMapAppModels, err := impl.bulkUpdateRepository.FindCMBulkAppModelForGlobal(appNameIncludes, appNameExcludes, bulkUpdatePayload.ConfigMap.Spec.Names)
			if err != nil {
				impl.logger.Errorw("error in fetching bulk app model for global", "err", err)
				return nil, err
			}
			configMapSpecNames := make(map[string]bool)
			if len(configMapAppModels) != 0 {
				for _, name := range bulkUpdatePayload.ConfigMap.Spec.Names {
					configMapSpecNames[name] = true
				}
			}
			for _, configMapAppModel := range configMapAppModels {
				var finalConfigMapNames []string
				configMapNames := gjson.Get(configMapAppModel.ConfigMapData, "maps.#.name")
				for _, configMapName := range configMapNames.Array() {
					_, contains := configMapSpecNames[configMapName.String()]
					if contains == true {
						finalConfigMapNames = append(finalConfigMapNames, configMapName.String())
					}
				}
				if len(finalConfigMapNames) != 0 {
					appDetailsById, _ := impl.appRepository.FindById(configMapAppModel.AppId)
					configMapImpactedObject := &CmAndSecretImpactedObjectsResponseForOneApp{
						AppId:   configMapAppModel.AppId,
						AppName: appDetailsById.AppName,
						Names:   finalConfigMapNames,
					}
					configMapImpactedObjects = append(configMapImpactedObjects, configMapImpactedObject)
				}
			}
		}
		//For Secret
		if bulkUpdatePayload.Secret != nil && bulkUpdatePayload.Secret.Spec != nil && len(bulkUpdatePayload.Secret.Spec.Names) != 0 {
			secretAppModels, err := impl.bulkUpdateRepository.FindSecretBulkAppModelForGlobal(appNameIncludes, appNameExcludes, bulkUpdatePayload.Secret.Spec.Names)
			if err != nil {
				impl.logger.Errorw("error in fetching bulk app model for global", "err", err)
				return nil, err
			}
			secretSpecNames := make(map[string]bool)
			if len(secretAppModels) != 0 {
				for _, name := range bulkUpdatePayload.Secret.Spec.Names {
					secretSpecNames[name] = true
				}
			}
			for _, secretAppModel := range secretAppModels {
				var finalSecretNames []string
				secretNames := gjson.Get(secretAppModel.SecretData, "secrets.#.name")
				for _, secretName := range secretNames.Array() {
					_, contains := secretSpecNames[secretName.String()]
					if contains == true {
						finalSecretNames = append(finalSecretNames, secretName.String())
					}
				}
				if len(finalSecretNames) != 0 {
					appDetailsById, _ := impl.appRepository.FindById(secretAppModel.AppId)
					secretImpactedObject := &CmAndSecretImpactedObjectsResponseForOneApp{
						AppId:   secretAppModel.AppId,
						AppName: appDetailsById.AppName,
						Names:   finalSecretNames,
					}
					secretImpactedObjects = append(secretImpactedObjects, secretImpactedObject)
				}
			}
		}
	}

	for _, envId := range bulkUpdatePayload.EnvIds {
		//For Deployment Template
		if bulkUpdatePayload.DeploymentTemplate != nil && bulkUpdatePayload.DeploymentTemplate.Spec != nil {
			appsNotGlobalDT, err := impl.bulkUpdateRepository.
				FindDeploymentTemplateBulkAppNameForEnv(appNameIncludes, appNameExcludes, envId)
			if err != nil {
				impl.logger.Errorw("error in fetching bulk app names for env", "err", err)
				return nil, err
			}
			for _, app := range appsNotGlobalDT {
				deploymentTemplateImpactedObject := &DeploymentTemplateImpactedObjectsResponseForOneApp{
					AppId:   app.Id,
					AppName: app.AppName,
					EnvId:   envId,
				}
				deploymentTemplateImpactedObjects = append(deploymentTemplateImpactedObjects, deploymentTemplateImpactedObject)
			}
		}
		//For ConfigMap
		if bulkUpdatePayload.ConfigMap != nil && bulkUpdatePayload.ConfigMap.Spec != nil && len(bulkUpdatePayload.ConfigMap.Spec.Names) != 0 {
			configMapEnvModels, err := impl.bulkUpdateRepository.FindCMBulkAppModelForEnv(appNameIncludes, appNameExcludes, envId, bulkUpdatePayload.ConfigMap.Spec.Names)
			if err != nil {
				impl.logger.Errorw("error in fetching bulk app model for global", "err", err)
				return nil, err
			}
			configMapSpecNames := make(map[string]bool)
			if len(configMapEnvModels) != 0 {
				for _, name := range bulkUpdatePayload.ConfigMap.Spec.Names {
					configMapSpecNames[name] = true
				}
			}
			for _, configMapEnvModel := range configMapEnvModels {
				var finalConfigMapNames []string
				configMapNames := gjson.Get(configMapEnvModel.ConfigMapData, "maps.#.name")
				for _, configMapName := range configMapNames.Array() {
					_, contains := configMapSpecNames[configMapName.String()]
					if contains == true {
						finalConfigMapNames = append(finalConfigMapNames, configMapName.String())
					}
				}

				if len(finalConfigMapNames) != 0 {
					appDetailsById, _ := impl.appRepository.FindById(configMapEnvModel.AppId)
					configMapImpactedObject := &CmAndSecretImpactedObjectsResponseForOneApp{
						AppId:   configMapEnvModel.AppId,
						AppName: appDetailsById.AppName,
						EnvId:   envId,
						Names:   finalConfigMapNames,
					}
					configMapImpactedObjects = append(configMapImpactedObjects, configMapImpactedObject)
				}
			}
		}
		//For Secret
		if bulkUpdatePayload.Secret != nil && bulkUpdatePayload.Secret.Spec != nil && len(bulkUpdatePayload.Secret.Spec.Names) != 0 {
			secretEnvModels, err := impl.bulkUpdateRepository.FindSecretBulkAppModelForEnv(appNameIncludes, appNameExcludes, envId, bulkUpdatePayload.Secret.Spec.Names)
			if err != nil {
				impl.logger.Errorw("error in fetching bulk app model for global", "err", err)
				return nil, err
			}
			secretSpecNames := make(map[string]bool)
			if len(secretEnvModels) != 0 {
				for _, name := range bulkUpdatePayload.Secret.Spec.Names {
					secretSpecNames[name] = true
				}
			}
			for _, secretEnvModel := range secretEnvModels {
				var finalSecretNames []string
				secretNames := gjson.Get(secretEnvModel.SecretData, "secrets.#.name")
				for _, secretName := range secretNames.Array() {
					_, contains := secretSpecNames[secretName.String()]
					if contains == true {
						finalSecretNames = append(finalSecretNames, secretName.String())
					}
				}

				if len(finalSecretNames) != 0 {
					appDetailsById, _ := impl.appRepository.FindById(secretEnvModel.AppId)
					secretImpactedObject := &CmAndSecretImpactedObjectsResponseForOneApp{
						AppId:   secretEnvModel.AppId,
						AppName: appDetailsById.AppName,
						EnvId:   envId,
						Names:   finalSecretNames,
					}
					secretImpactedObjects = append(secretImpactedObjects, secretImpactedObject)
				}
			}
		}
	}
	impactedObjectsResponse.DeploymentTemplate = deploymentTemplateImpactedObjects
	impactedObjectsResponse.ConfigMap = configMapImpactedObjects
	impactedObjectsResponse.Secret = secretImpactedObjects
	return impactedObjectsResponse, nil
}
func (impl BulkUpdateServiceImpl) ApplyJsonPatch(patch jsonpatch.Patch, target string) (string, error) {
	modified, err := patch.Apply([]byte(target))
	if err != nil {
		impl.logger.Errorw("error in applying JSON patch", "err", err)
		return "Patch Failed", err
	}
	return string(modified), err
}
func (impl BulkUpdateServiceImpl) BulkUpdateDeploymentTemplate(bulkUpdatePayload *BulkUpdatePayload) *DeploymentTemplateBulkUpdateResponse {
	deploymentTemplateBulkUpdateResponse := &DeploymentTemplateBulkUpdateResponse{}
	var appNameIncludes []string
	var appNameExcludes []string
	if bulkUpdatePayload.Includes == nil || len(bulkUpdatePayload.Includes.Names) == 0 {
		deploymentTemplateBulkUpdateResponse.Message = append(deploymentTemplateBulkUpdateResponse.Message, "Please don't leave includes.names array empty")
		return deploymentTemplateBulkUpdateResponse
	} else {
		appNameIncludes = bulkUpdatePayload.Includes.Names
	}
	if bulkUpdatePayload.Excludes != nil && len(bulkUpdatePayload.Excludes.Names) > 0 {
		appNameExcludes = bulkUpdatePayload.Excludes.Names
	}
	deploymentTemplatePatchJson := []byte(bulkUpdatePayload.DeploymentTemplate.Spec.PatchJson)
	deploymentTemplatePatch, err := jsonpatch.DecodePatch(deploymentTemplatePatchJson)
	if err != nil {
		impl.logger.Errorw("error in decoding JSON patch", "err", err)
		deploymentTemplateBulkUpdateResponse.Message = append(deploymentTemplateBulkUpdateResponse.Message, "The patch string you entered seems wrong, please check and try again")
	}
	var charts []*chartRepoRepository.Chart
	if bulkUpdatePayload.Global {
		charts, err = impl.bulkUpdateRepository.FindBulkChartsByAppNameSubstring(appNameIncludes, appNameExcludes)
		if err != nil {
			impl.logger.Error("error in fetching charts by app name substring")
			deploymentTemplateBulkUpdateResponse.Message = append(deploymentTemplateBulkUpdateResponse.Message, fmt.Sprintf("Unable to bulk update apps globally : %s", err.Error()))
		} else {
			if len(charts) == 0 {
				deploymentTemplateBulkUpdateResponse.Message = append(deploymentTemplateBulkUpdateResponse.Message, "No matching apps to update globally")
			} else {
				for _, chart := range charts {
					appDetailsByChart, _ := impl.bulkUpdateRepository.FindAppByChartId(chart.Id)
					modified, err := impl.ApplyJsonPatch(deploymentTemplatePatch, chart.Values)
					if err != nil {
						impl.logger.Errorw("error in applying JSON patch", "err", err)
						bulkUpdateFailedResponse := &DeploymentTemplateBulkUpdateResponseForOneApp{
							AppId:   appDetailsByChart.Id,
							AppName: appDetailsByChart.AppName,
							Message: fmt.Sprintf("Error in applying JSON patch : %s", err.Error()),
						}
						deploymentTemplateBulkUpdateResponse.Failure = append(deploymentTemplateBulkUpdateResponse.Failure, bulkUpdateFailedResponse)
					} else {
						err = impl.bulkUpdateRepository.BulkUpdateChartsValuesYamlAndGlobalOverrideById(chart.Id, modified)
						if err != nil {
							impl.logger.Errorw("error in bulk updating charts", "err", err)
							bulkUpdateFailedResponse := &DeploymentTemplateBulkUpdateResponseForOneApp{
								AppId:   appDetailsByChart.Id,
								AppName: appDetailsByChart.AppName,
								Message: fmt.Sprintf("Error in updating in db : %s", err.Error()),
							}
							deploymentTemplateBulkUpdateResponse.Failure = append(deploymentTemplateBulkUpdateResponse.Failure, bulkUpdateFailedResponse)
						} else {
							bulkUpdateSuccessResponse := &DeploymentTemplateBulkUpdateResponseForOneApp{
								AppId:   appDetailsByChart.Id,
								AppName: appDetailsByChart.AppName,
								Message: "Updated Successfully",
							}
							deploymentTemplateBulkUpdateResponse.Successful = append(deploymentTemplateBulkUpdateResponse.Successful, bulkUpdateSuccessResponse)

							//creating history entry for deployment template
							appLevelAppMetricsEnabled := false
							appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(chart.AppId)
							if err != nil && err != pg.ErrNoRows {
								impl.logger.Errorw("error in getting app level metrics app level", "error", err)
							} else if err == nil {
								appLevelAppMetricsEnabled = appLevelMetrics.AppMetrics
							}
							chart.GlobalOverride = modified
							chart.Values = modified
							err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromGlobalTemplate(chart, nil, appLevelAppMetricsEnabled)
							if err != nil {
								impl.logger.Errorw("error in creating entry for deployment template history", "err", err, "chart", chart)
							}
							//VARIABLE_MAPPING_UPDATE
							//NOTE: this flow is doesn't have the user info, therefore updated by is being set to the last updated by
							err = impl.extractAndMapVariables(chart.GlobalOverride, chart.Id, repository5.EntityTypeDeploymentTemplateAppLevel, chart.UpdatedBy)
							if err != nil {
								return nil
							}
						}
					}
				}
			}
		}
	}
	var chartsEnv []*chartConfig.EnvConfigOverride
	for _, envId := range bulkUpdatePayload.EnvIds {
		chartsEnv, err = impl.bulkUpdateRepository.FindBulkChartsEnvByAppNameSubstring(appNameIncludes, appNameExcludes, envId)
		if err != nil {
			impl.logger.Errorw("error in fetching charts(for env) by app name substring", "err", err)
			deploymentTemplateBulkUpdateResponse.Message = append(deploymentTemplateBulkUpdateResponse.Message, fmt.Sprintf("Unable to bulk update apps for envId = %d , %s", envId, err.Error()))
		} else {
			if len(chartsEnv) == 0 {
				deploymentTemplateBulkUpdateResponse.Message = append(deploymentTemplateBulkUpdateResponse.Message, fmt.Sprintf("No matching apps to update for envId = %d", envId))
			} else {
				for _, chartEnv := range chartsEnv {
					appDetailsByChart, _ := impl.bulkUpdateRepository.FindAppByChartEnvId(chartEnv.Id)
					modified, err := impl.ApplyJsonPatch(deploymentTemplatePatch, chartEnv.EnvOverrideValues)
					if err != nil {
						impl.logger.Errorw("error in applying JSON patch", "err", err)
						bulkUpdateFailedResponse := &DeploymentTemplateBulkUpdateResponseForOneApp{
							AppId:   appDetailsByChart.Id,
							AppName: appDetailsByChart.AppName,
							EnvId:   envId,
							Message: fmt.Sprintf("Error in applying JSON patch : %s", err.Error()),
						}
						deploymentTemplateBulkUpdateResponse.Failure = append(deploymentTemplateBulkUpdateResponse.Failure, bulkUpdateFailedResponse)
					} else {
						err = impl.bulkUpdateRepository.BulkUpdateChartsEnvYamlOverrideById(chartEnv.Id, modified)
						if err != nil {
							impl.logger.Errorw("error in bulk updating charts", "err", err)
							bulkUpdateFailedResponse := &DeploymentTemplateBulkUpdateResponseForOneApp{
								AppId:   appDetailsByChart.Id,
								AppName: appDetailsByChart.AppName,
								EnvId:   envId,
								Message: fmt.Sprintf("Error in updating in db : %s", err.Error()),
							}
							deploymentTemplateBulkUpdateResponse.Failure = append(deploymentTemplateBulkUpdateResponse.Failure, bulkUpdateFailedResponse)
						} else {
							bulkUpdateSuccessResponse := &DeploymentTemplateBulkUpdateResponseForOneApp{
								AppId:   appDetailsByChart.Id,
								AppName: appDetailsByChart.AppName,
								EnvId:   envId,
								Message: "Updated Successfully",
							}
							deploymentTemplateBulkUpdateResponse.Successful = append(deploymentTemplateBulkUpdateResponse.Successful, bulkUpdateSuccessResponse)

							//creating history entry for deployment template
							envLevelAppMetricsEnabled := false
							envLevelAppMetrics, err := impl.envLevelAppMetricsRepository.FindByAppIdAndEnvId(chartEnv.Chart.AppId, chartEnv.TargetEnvironment)
							if err != nil && err != pg.ErrNoRows {
								impl.logger.Errorw("error in getting env level app metrics", "err", err, "appId", chartEnv.Chart.AppId, "envId", chartEnv.TargetEnvironment)
							} else if err == pg.ErrNoRows {
								appLevelAppMetrics, err := impl.appLevelMetricsRepository.FindByAppId(chartEnv.Chart.AppId)
								if err != nil && err != pg.ErrNoRows {
									impl.logger.Errorw("error in getting app level app metrics", "err", err, "appId", chartEnv.Chart.AppId)
								} else if err == nil {
									envLevelAppMetricsEnabled = appLevelAppMetrics.AppMetrics
								}
							} else {
								envLevelAppMetricsEnabled = *envLevelAppMetrics.AppMetrics
							}
							chartEnv.EnvOverrideValues = modified
							err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(chartEnv, nil, envLevelAppMetricsEnabled, 0)
							if err != nil {
								impl.logger.Errorw("error in creating entry for env deployment template history", "err", err, "envOverride", chartEnv)
							}
							//VARIABLE_MAPPING_UPDATE
							err = impl.extractAndMapVariables(chartEnv.EnvOverrideValues, chartEnv.Id, repository5.EntityTypeDeploymentTemplateEnvLevel, chartEnv.UpdatedBy)
							if err != nil {
								return nil
							}
						}
					}
				}
			}
		}
	}
	if len(deploymentTemplateBulkUpdateResponse.Failure) == 0 && len(deploymentTemplateBulkUpdateResponse.Successful) != 0 {
		deploymentTemplateBulkUpdateResponse.Message = append(deploymentTemplateBulkUpdateResponse.Message, "All matching apps are updated successfully")
	}
	return deploymentTemplateBulkUpdateResponse
}

func (impl BulkUpdateServiceImpl) extractAndMapVariables(template string, entityId int, entityType repository5.EntityType, userId int32) error {
	usedVariables, err := impl.variableTemplateParser.ExtractVariables(template, parsers.JsonVariableTemplate)
	if err != nil {
		return err
	}
	err = impl.variableEntityMappingService.UpdateVariablesForEntity(usedVariables, repository5.Entity{
		EntityType: entityType,
		EntityId:   entityId,
	}, userId, nil)
	if err != nil {
		return err
	}
	return nil
}

func (impl BulkUpdateServiceImpl) BulkUpdateConfigMap(bulkUpdatePayload *BulkUpdatePayload) *CmAndSecretBulkUpdateResponse {
	configMapBulkUpdateResponse := &CmAndSecretBulkUpdateResponse{}
	var appNameIncludes []string
	var appNameExcludes []string
	if bulkUpdatePayload.Includes == nil || len(bulkUpdatePayload.Includes.Names) == 0 {
		configMapBulkUpdateResponse.Message = append(configMapBulkUpdateResponse.Message, "Please don't leave includes.names array empty")
		return configMapBulkUpdateResponse
	} else {
		appNameIncludes = bulkUpdatePayload.Includes.Names
	}
	if bulkUpdatePayload.Excludes != nil && len(bulkUpdatePayload.Excludes.Names) > 0 {
		appNameExcludes = bulkUpdatePayload.Excludes.Names
	}

	if bulkUpdatePayload.Global {
		configMapSpecNames := make(map[string]bool)
		for _, name := range bulkUpdatePayload.ConfigMap.Spec.Names {
			configMapSpecNames[name] = true
		}
		configMapAppModels, err := impl.bulkUpdateRepository.FindCMBulkAppModelForGlobal(appNameIncludes, appNameExcludes, bulkUpdatePayload.ConfigMap.Spec.Names)
		if err != nil {
			impl.logger.Errorw("error in fetching bulk app model for global", "err", err)
			configMapBulkUpdateResponse.Message = append(configMapBulkUpdateResponse.Message, fmt.Sprintf("Unable to bulk update apps globally : %s", err.Error()))
		} else {
			if len(configMapAppModels) == 0 {
				configMapBulkUpdateResponse.Message = append(configMapBulkUpdateResponse.Message, "No matching apps to update globally")
			} else {
				for _, configMapAppModel := range configMapAppModels {
					configMapNames := gjson.Get(configMapAppModel.ConfigMapData, "maps.#.name")
					messageCmNamesMap := make(map[string][]string)
					for i, configMapName := range configMapNames.Array() {
						_, contains := configMapSpecNames[configMapName.String()]
						if contains == true {
							configMapPatchJsonString := bulkUpdatePayload.ConfigMap.Spec.PatchJson
							keyNames := gjson.Get(configMapPatchJsonString, "#.path")
							for j, keyName := range keyNames.Array() {
								configMapPatchJsonString, _ = sjson.Set(configMapPatchJsonString, fmt.Sprintf("%d.path", j), fmt.Sprintf("/maps/%d/data%s", i, keyName.String()))
							}
							configMapPatchJson := []byte(configMapPatchJsonString)
							configMapPatch, err := jsonpatch.DecodePatch(configMapPatchJson)
							if err != nil {
								impl.logger.Errorw("error in decoding JSON patch", "err", err)
								if _, ok := messageCmNamesMap["The patch string you entered seems wrong, please check and try again"]; !ok {
									messageCmNamesMap["The patch string you entered seems wrong, please check and try again"] = []string{configMapName.String()}
								} else {
									messageCmNamesMap["The patch string you entered seems wrong, please check and try again"] = append(messageCmNamesMap["The patch string you entered seems wrong, please check and try again"], configMapName.String())
								}
							} else {
								modified, err := impl.ApplyJsonPatch(configMapPatch, configMapAppModel.ConfigMapData)
								if err != nil {
									impl.logger.Errorw("error in applying JSON patch", "err", err)
									if _, ok := messageCmNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())]; !ok {
										messageCmNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())] = []string{configMapName.String()}
									} else {
										messageCmNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())] = append(messageCmNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())], configMapName.String())
									}
								} else {
									configMapAppModel.ConfigMapData = modified
									if _, ok := messageCmNamesMap["Updated Successfully"]; !ok {
										messageCmNamesMap["Updated Successfully"] = []string{configMapName.String()}
									} else {
										messageCmNamesMap["Updated Successfully"] = append(messageCmNamesMap["Updated Successfully"], configMapName.String())
									}
								}
							}
						}
					}
					if _, ok := messageCmNamesMap["Updated Successfully"]; ok {
						err := impl.bulkUpdateRepository.BulkUpdateConfigMapDataForGlobalById(configMapAppModel.Id, configMapAppModel.ConfigMapData)
						if err != nil {
							impl.logger.Errorw("error in bulk updating charts", "err", err)
							messageCmNamesMap[fmt.Sprintf("Error in updating in db : %s", err.Error())] = messageCmNamesMap["Updated Successfully"]
							delete(messageCmNamesMap, "Updated Successfully")
						}
						//creating history for config map history
						err = impl.configMapHistoryService.CreateHistoryFromAppLevelConfig(configMapAppModel, repository4.CONFIGMAP_TYPE)
						if err != nil {
							impl.logger.Errorw("error in creating entry for configmap history", "err", err)
						}
					}
					if len(messageCmNamesMap) != 0 {
						appDetailsById, _ := impl.appRepository.FindById(configMapAppModel.AppId)
						for key, value := range messageCmNamesMap {
							if key == "Updated Successfully" {
								bulkUpdateSuccessResponse := &CmAndSecretBulkUpdateResponseForOneApp{
									AppId:   appDetailsById.Id,
									AppName: appDetailsById.AppName,
									Names:   value,
									Message: key,
								}
								configMapBulkUpdateResponse.Successful = append(configMapBulkUpdateResponse.Successful, bulkUpdateSuccessResponse)
							} else {
								bulkUpdateFailedResponse := &CmAndSecretBulkUpdateResponseForOneApp{
									AppId:   appDetailsById.Id,
									AppName: appDetailsById.AppName,
									Names:   value,
									Message: key,
								}
								configMapBulkUpdateResponse.Failure = append(configMapBulkUpdateResponse.Failure, bulkUpdateFailedResponse)
							}
						}
					}
				}
			}
		}
	}
	for _, envId := range bulkUpdatePayload.EnvIds {
		configMapSpecNames := make(map[string]bool)
		for _, name := range bulkUpdatePayload.ConfigMap.Spec.Names {
			configMapSpecNames[name] = true
		}
		configMapEnvModels, err := impl.bulkUpdateRepository.FindCMBulkAppModelForEnv(appNameIncludes, appNameExcludes, envId, bulkUpdatePayload.ConfigMap.Spec.Names)
		if err != nil {
			impl.logger.Errorw("error in fetching bulk app model for env", "err", err)
			configMapBulkUpdateResponse.Message = append(configMapBulkUpdateResponse.Message, fmt.Sprintf("Unable to bulk update apps for env: %d , %s", envId, err.Error()))
		} else {
			if len(configMapEnvModels) == 0 {
				configMapBulkUpdateResponse.Message = append(configMapBulkUpdateResponse.Message, fmt.Sprintf("No matching apps to update for envId : %d", envId))
			} else {
				for _, configMapEnvModel := range configMapEnvModels {
					configMapNames := gjson.Get(configMapEnvModel.ConfigMapData, "maps.#.name")
					messageCmNamesMap := make(map[string][]string)
					for i, configMapName := range configMapNames.Array() {
						_, contains := configMapSpecNames[configMapName.String()]
						if contains == true {
							configMapPatchJsonString := bulkUpdatePayload.ConfigMap.Spec.PatchJson
							keyNames := gjson.Get(configMapPatchJsonString, "#.path")
							for j, keyName := range keyNames.Array() {
								configMapPatchJsonString, _ = sjson.Set(configMapPatchJsonString, fmt.Sprintf("%d.path", j), fmt.Sprintf("/maps/%d/data%s", i, keyName.String()))
							}
							configMapPatchJson := []byte(configMapPatchJsonString)
							configMapPatch, err := jsonpatch.DecodePatch(configMapPatchJson)
							if err != nil {
								impl.logger.Errorw("error in decoding JSON patch", "err", err)
								if _, ok := messageCmNamesMap["The patch string you entered seems wrong, please check and try again"]; !ok {
									messageCmNamesMap["The patch string you entered seems wrong, please check and try again"] = []string{configMapName.String()}
								} else {
									messageCmNamesMap["The patch string you entered seems wrong, please check and try again"] = append(messageCmNamesMap["The patch string you entered seems wrong, please check and try again"], configMapName.String())
								}
							} else {
								modified, err := impl.ApplyJsonPatch(configMapPatch, configMapEnvModel.ConfigMapData)
								if err != nil {
									impl.logger.Errorw("error in applying JSON patch", "err", err)
									if _, ok := messageCmNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())]; !ok {
										messageCmNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())] = []string{configMapName.String()}
									} else {
										messageCmNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())] = append(messageCmNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())], configMapName.String())
									}
								} else {
									configMapEnvModel.ConfigMapData = modified
									if _, ok := messageCmNamesMap["Updated Successfully"]; !ok {
										messageCmNamesMap["Updated Successfully"] = []string{configMapName.String()}
									} else {
										messageCmNamesMap["Updated Successfully"] = append(messageCmNamesMap["Updated Successfully"], configMapName.String())
									}
								}
							}
						}
					}
					if _, ok := messageCmNamesMap["Updated Successfully"]; ok {
						err := impl.bulkUpdateRepository.BulkUpdateConfigMapDataForEnvById(configMapEnvModel.Id, configMapEnvModel.ConfigMapData)
						if err != nil {
							impl.logger.Errorw("error in bulk updating charts", "err", err)
							messageCmNamesMap[fmt.Sprintf("Error in updating in db : %s", err.Error())] = messageCmNamesMap["Updated Successfully"]
							delete(messageCmNamesMap, "Updated Successfully")
						}
						//creating history for config map history
						err = impl.configMapHistoryService.CreateHistoryFromEnvLevelConfig(configMapEnvModel, repository4.CONFIGMAP_TYPE)
						if err != nil {
							impl.logger.Errorw("error in creating entry for configmap history", "err", err)
						}
					}
					if len(messageCmNamesMap) != 0 {
						appDetailsById, _ := impl.appRepository.FindById(configMapEnvModel.AppId)
						for key, value := range messageCmNamesMap {
							if key == "Updated Successfully" {
								bulkUpdateSuccessResponse := &CmAndSecretBulkUpdateResponseForOneApp{
									AppId:   appDetailsById.Id,
									AppName: appDetailsById.AppName,
									Names:   value,
									Message: key,
									EnvId:   envId,
								}
								configMapBulkUpdateResponse.Successful = append(configMapBulkUpdateResponse.Successful, bulkUpdateSuccessResponse)
							} else {
								bulkUpdateFailedResponse := &CmAndSecretBulkUpdateResponseForOneApp{
									AppId:   appDetailsById.Id,
									AppName: appDetailsById.AppName,
									Names:   value,
									Message: key,
									EnvId:   envId,
								}
								configMapBulkUpdateResponse.Failure = append(configMapBulkUpdateResponse.Failure, bulkUpdateFailedResponse)
							}
						}
					}
				}
			}
		}
	}
	if len(configMapBulkUpdateResponse.Failure) == 0 && len(configMapBulkUpdateResponse.Successful) != 0 {
		configMapBulkUpdateResponse.Message = append(configMapBulkUpdateResponse.Message, "All matching apps are updated successfully")
	}
	return configMapBulkUpdateResponse
}
func (impl BulkUpdateServiceImpl) BulkUpdateSecret(bulkUpdatePayload *BulkUpdatePayload) *CmAndSecretBulkUpdateResponse {
	secretBulkUpdateResponse := &CmAndSecretBulkUpdateResponse{}
	var appNameIncludes []string
	var appNameExcludes []string
	if bulkUpdatePayload.Includes == nil || len(bulkUpdatePayload.Includes.Names) == 0 {
		secretBulkUpdateResponse.Message = append(secretBulkUpdateResponse.Message, "Please don't leave includes.names array empty")
		return secretBulkUpdateResponse
	} else {
		appNameIncludes = bulkUpdatePayload.Includes.Names
	}
	if bulkUpdatePayload.Excludes != nil && len(bulkUpdatePayload.Excludes.Names) > 0 {
		appNameExcludes = bulkUpdatePayload.Excludes.Names
	}

	if bulkUpdatePayload.Global {
		secretSpecNames := make(map[string]bool)
		for _, name := range bulkUpdatePayload.Secret.Spec.Names {
			secretSpecNames[name] = true
		}
		secretAppModels, err := impl.bulkUpdateRepository.FindSecretBulkAppModelForGlobal(appNameIncludes, appNameExcludes, bulkUpdatePayload.Secret.Spec.Names)
		if err != nil {
			impl.logger.Errorw("error in fetching bulk app model for global", "err", err)
			secretBulkUpdateResponse.Message = append(secretBulkUpdateResponse.Message, fmt.Sprintf("Unable to bulk update apps globally : %s", err.Error()))
		} else {
			if len(secretAppModels) == 0 {
				secretBulkUpdateResponse.Message = append(secretBulkUpdateResponse.Message, "No matching apps to update globally")
			} else {
				for _, secretAppModel := range secretAppModels {
					secretNames := gjson.Get(secretAppModel.SecretData, "secrets.#.name")
					messageSecretNamesMap := make(map[string][]string)
					for i, secretName := range secretNames.Array() {
						_, contains := secretSpecNames[secretName.String()]
						if contains == true {
							secretPatchJsonString := bulkUpdatePayload.Secret.Spec.PatchJson
							keyNames := gjson.Get(secretPatchJsonString, "#.path")
							for j, keyName := range keyNames.Array() {
								secretPatchJsonString, _ = sjson.Set(secretPatchJsonString, fmt.Sprintf("%d.path", j), fmt.Sprintf("/secrets/%d/data%s", i, keyName.String()))
							}
							//updating values to their base64 equivalent, on secret save/update operation this logic is implemented on FE
							values := gjson.Get(secretPatchJsonString, "#.value")
							for j, value := range values.Array() {
								base64EncodedValue := base64.StdEncoding.EncodeToString([]byte(value.String()))
								secretPatchJsonString, _ = sjson.Set(secretPatchJsonString, fmt.Sprintf("%d.value", j), base64EncodedValue)
							}
							secretPatchJson := []byte(secretPatchJsonString)
							secretPatch, err := jsonpatch.DecodePatch(secretPatchJson)
							if err != nil {
								impl.logger.Errorw("error in decoding JSON patch", "err", err)
								if _, ok := messageSecretNamesMap["The patch string you entered seems wrong, please check and try again"]; !ok {
									messageSecretNamesMap["The patch string you entered seems wrong, please check and try again"] = []string{secretName.String()}
								} else {
									messageSecretNamesMap["The patch string you entered seems wrong, please check and try again"] = append(messageSecretNamesMap["The patch string you entered seems wrong, please check and try again"], secretName.String())
								}
							} else {
								modified, err := impl.ApplyJsonPatch(secretPatch, secretAppModel.SecretData)
								if err != nil {
									impl.logger.Errorw("error in applying JSON patch", "err", err)
									if _, ok := messageSecretNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())]; !ok {
										messageSecretNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())] = []string{secretName.String()}
									} else {
										messageSecretNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())] = append(messageSecretNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())], secretName.String())
									}
								} else {
									secretAppModel.SecretData = modified
									if _, ok := messageSecretNamesMap["Updated Successfully"]; !ok {
										messageSecretNamesMap["Updated Successfully"] = []string{secretName.String()}
									} else {
										messageSecretNamesMap["Updated Successfully"] = append(messageSecretNamesMap["Updated Successfully"], secretName.String())
									}
								}
							}
						}
					}
					if _, ok := messageSecretNamesMap["Updated Successfully"]; ok {
						err := impl.bulkUpdateRepository.BulkUpdateSecretDataForGlobalById(secretAppModel.Id, secretAppModel.SecretData)
						if err != nil {
							impl.logger.Errorw("error in bulk updating secrets", "err", err)
							messageSecretNamesMap[fmt.Sprintf("Error in updating in db : %s", err.Error())] = messageSecretNamesMap["Updated Successfully"]
							delete(messageSecretNamesMap, "Updated Successfully")
						}
						//creating history for config map history
						err = impl.configMapHistoryService.CreateHistoryFromAppLevelConfig(secretAppModel, repository4.SECRET_TYPE)
						if err != nil {
							impl.logger.Errorw("error in creating entry for secret history", "err", err)
						}
					}
					if len(messageSecretNamesMap) != 0 {
						appDetailsById, _ := impl.appRepository.FindById(secretAppModel.AppId)
						for key, value := range messageSecretNamesMap {
							if key == "Updated Successfully" {
								bulkUpdateSuccessResponse := &CmAndSecretBulkUpdateResponseForOneApp{
									AppId:   appDetailsById.Id,
									AppName: appDetailsById.AppName,
									Names:   value,
									Message: key,
								}
								secretBulkUpdateResponse.Successful = append(secretBulkUpdateResponse.Successful, bulkUpdateSuccessResponse)
							} else {
								bulkUpdateFailedResponse := &CmAndSecretBulkUpdateResponseForOneApp{
									AppId:   appDetailsById.Id,
									AppName: appDetailsById.AppName,
									Names:   value,
									Message: key,
								}
								secretBulkUpdateResponse.Failure = append(secretBulkUpdateResponse.Failure, bulkUpdateFailedResponse)
							}
						}
					}
				}
			}
		}
	}
	for _, envId := range bulkUpdatePayload.EnvIds {
		secretSpecNames := make(map[string]bool)
		for _, name := range bulkUpdatePayload.Secret.Spec.Names {
			secretSpecNames[name] = true
		}
		secretEnvModels, err := impl.bulkUpdateRepository.FindSecretBulkAppModelForEnv(appNameIncludes, appNameExcludes, envId, bulkUpdatePayload.Secret.Spec.Names)
		if err != nil {
			impl.logger.Errorw("error in fetching bulk app model for env", "err", err)
			secretBulkUpdateResponse.Message = append(secretBulkUpdateResponse.Message, fmt.Sprintf("Unable to bulk update apps for env: %d , %s", envId, err.Error()))
		} else {
			if len(secretEnvModels) == 0 {
				secretBulkUpdateResponse.Message = append(secretBulkUpdateResponse.Message, fmt.Sprintf("No matching apps to update for envId : %d", envId))
			} else {
				for _, secretEnvModel := range secretEnvModels {
					secretNames := gjson.Get(secretEnvModel.SecretData, "secrets.#.name")
					messageSecretNamesMap := make(map[string][]string)
					for i, secretName := range secretNames.Array() {
						_, contains := secretSpecNames[secretName.String()]
						if contains == true {
							secretPatchJsonString := bulkUpdatePayload.Secret.Spec.PatchJson
							keyNames := gjson.Get(secretPatchJsonString, "#.path")
							for j, keyName := range keyNames.Array() {
								secretPatchJsonString, _ = sjson.Set(secretPatchJsonString, fmt.Sprintf("%d.path", j), fmt.Sprintf("/secrets/%d/data%s", i, keyName.String()))
							}
							//updating values to their base64 equivalent, on secret save/update operation this logic is implemented on FE
							values := gjson.Get(secretPatchJsonString, "#.value")
							for j, value := range values.Array() {
								base64EncodedValue := base64.StdEncoding.EncodeToString([]byte(value.String()))
								secretPatchJsonString, _ = sjson.Set(secretPatchJsonString, fmt.Sprintf("%d.value", j), base64EncodedValue)
							}
							secretPatchJson := []byte(secretPatchJsonString)
							secretPatch, err := jsonpatch.DecodePatch(secretPatchJson)
							if err != nil {
								impl.logger.Errorw("error in decoding JSON patch", "err", err)
								if _, ok := messageSecretNamesMap["The patch string you entered seems wrong, please check and try again"]; !ok {
									messageSecretNamesMap["The patch string you entered seems wrong, please check and try again"] = []string{secretName.String()}
								} else {
									messageSecretNamesMap["The patch string you entered seems wrong, please check and try again"] = append(messageSecretNamesMap["The patch string you entered seems wrong, please check and try again"], secretName.String())
								}
							} else {
								modified, err := impl.ApplyJsonPatch(secretPatch, secretEnvModel.SecretData)
								if err != nil {
									impl.logger.Errorw("error in applying JSON patch", "err", err)
									if _, ok := messageSecretNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())]; !ok {
										messageSecretNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())] = []string{secretName.String()}
									} else {
										messageSecretNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())] = append(messageSecretNamesMap[fmt.Sprintf("Error in applying JSON patch : %s", err.Error())], secretName.String())
									}
								} else {
									secretEnvModel.SecretData = modified
									if _, ok := messageSecretNamesMap["Updated Successfully"]; !ok {
										messageSecretNamesMap["Updated Successfully"] = []string{secretName.String()}
									} else {
										messageSecretNamesMap["Updated Successfully"] = append(messageSecretNamesMap["Updated Successfully"], secretName.String())
									}
								}
							}
						}
					}
					if _, ok := messageSecretNamesMap["Updated Successfully"]; ok {
						err := impl.bulkUpdateRepository.BulkUpdateSecretDataForEnvById(secretEnvModel.Id, secretEnvModel.SecretData)
						if err != nil {
							impl.logger.Errorw("error in bulk updating charts", "err", err)
							messageSecretNamesMap[fmt.Sprintf("Error in updating in db : %s", err.Error())] = messageSecretNamesMap["Updated Successfully"]
							delete(messageSecretNamesMap, "Updated Successfully")
						}
						//creating history for config map history
						err = impl.configMapHistoryService.CreateHistoryFromEnvLevelConfig(secretEnvModel, repository4.SECRET_TYPE)
						if err != nil {
							impl.logger.Errorw("error in creating entry for secret history", "err", err)
						}
					}
					if len(messageSecretNamesMap) != 0 {
						appDetailsById, _ := impl.appRepository.FindById(secretEnvModel.AppId)
						for key, value := range messageSecretNamesMap {
							if key == "Updated Successfully" {
								bulkUpdateSuccessResponse := &CmAndSecretBulkUpdateResponseForOneApp{
									AppId:   appDetailsById.Id,
									AppName: appDetailsById.AppName,
									Names:   value,
									Message: key,
									EnvId:   envId,
								}
								secretBulkUpdateResponse.Successful = append(secretBulkUpdateResponse.Successful, bulkUpdateSuccessResponse)
							} else {
								bulkUpdateFailedResponse := &CmAndSecretBulkUpdateResponseForOneApp{
									AppId:   appDetailsById.Id,
									AppName: appDetailsById.AppName,
									Names:   value,
									Message: key,
									EnvId:   envId,
								}
								secretBulkUpdateResponse.Failure = append(secretBulkUpdateResponse.Failure, bulkUpdateFailedResponse)
							}
						}
					}
				}
			}
		}
	}
	if len(secretBulkUpdateResponse.Failure) == 0 && len(secretBulkUpdateResponse.Successful) != 0 {
		secretBulkUpdateResponse.Message = append(secretBulkUpdateResponse.Message, "All matching apps are updated successfully")
	}
	return secretBulkUpdateResponse
}
func (impl BulkUpdateServiceImpl) BulkUpdate(bulkUpdatePayload *BulkUpdatePayload) *BulkUpdateResponse {
	bulkUpdateResponse := &BulkUpdateResponse{}
	var deploymentTemplateBulkUpdateResponse *DeploymentTemplateBulkUpdateResponse
	var configMapBulkUpdateResponse *CmAndSecretBulkUpdateResponse
	var secretBulkUpdateResponse *CmAndSecretBulkUpdateResponse
	if bulkUpdatePayload.DeploymentTemplate != nil && bulkUpdatePayload.DeploymentTemplate.Spec != nil && bulkUpdatePayload.DeploymentTemplate.Spec.PatchJson != "" {
		deploymentTemplateBulkUpdateResponse = impl.BulkUpdateDeploymentTemplate(bulkUpdatePayload)
	}
	if bulkUpdatePayload.ConfigMap != nil && bulkUpdatePayload.ConfigMap.Spec != nil && len(bulkUpdatePayload.ConfigMap.Spec.Names) != 0 && bulkUpdatePayload.ConfigMap.Spec.PatchJson != "" {
		configMapBulkUpdateResponse = impl.BulkUpdateConfigMap(bulkUpdatePayload)
	}
	if bulkUpdatePayload.Secret != nil && bulkUpdatePayload.Secret.Spec != nil && len(bulkUpdatePayload.Secret.Spec.Names) != 0 && bulkUpdatePayload.Secret.Spec.PatchJson != "" {
		secretBulkUpdateResponse = impl.BulkUpdateSecret(bulkUpdatePayload)
	}

	bulkUpdateResponse.DeploymentTemplate = deploymentTemplateBulkUpdateResponse
	bulkUpdateResponse.ConfigMap = configMapBulkUpdateResponse
	bulkUpdateResponse.Secret = secretBulkUpdateResponse
	return bulkUpdateResponse
}

func (impl BulkUpdateServiceImpl) BulkHibernate(request *BulkApplicationForEnvironmentPayload, ctx context.Context, w http.ResponseWriter, token string, checkAuthForBulkActions func(token string, appObject string, envObject string) bool) (*BulkApplicationForEnvironmentResponse, error) {
	var pipelines []*pipelineConfig.Pipeline
	var err error
	if len(request.AppIdIncludes) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.EnvId, request.AppIdIncludes)
	} else if len(request.AppIdExcludes) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByNotFilter(request.EnvId, request.AppIdExcludes)
	} else {
		pipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.EnvId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "envId", request.EnvId, "err", err)
		return nil, err
	}
	response := make(map[string]map[string]bool)
	for _, pipeline := range pipelines {
		appKey := fmt.Sprintf("%d_%s", pipeline.AppId, pipeline.App.AppName)
		pipelineKey := fmt.Sprintf("%d_%s", pipeline.Id, pipeline.Name)
		success := true
		if _, ok := response[appKey]; !ok {
			pResponse := make(map[string]bool)
			pResponse[pipelineKey] = false
			response[appKey] = pResponse
		}
		appObject := impl.enforcerUtil.GetAppRBACNameByAppId(pipeline.AppId)
		envObject := impl.enforcerUtil.GetEnvRBACNameByAppId(pipeline.AppId, pipeline.EnvironmentId)
		isValidAuth := checkAuthForBulkActions(token, appObject, envObject)
		if !isValidAuth {
			//skip hibernate for the app if user does not have access on that
			pipelineResponse := response[appKey]
			pipelineResponse[pipelineKey] = false
			response[appKey] = pipelineResponse
			continue
		}
		var hibernateReqError error
		//if pipeline.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD {
		stopRequest := &pipeline1.StopAppRequest{
			AppId:         pipeline.AppId,
			EnvironmentId: pipeline.EnvironmentId,
			UserId:        request.UserId,
			RequestType:   pipeline1.STOP,
		}
		_, hibernateReqError = impl.workflowDagExecutor.StopStartApp(stopRequest, ctx)
		if hibernateReqError != nil {
			impl.logger.Errorw("error in hibernating application", "err", hibernateReqError, "pipeline", pipeline)
			pipelineResponse := response[appKey]
			pipelineResponse[pipelineKey] = false
			response[appKey] = pipelineResponse
			continue
		}
		pipelineResponse := response[appKey]
		pipelineResponse[pipelineKey] = success
		response[appKey] = pipelineResponse
	}
	bulkOperationResponse := &BulkApplicationForEnvironmentResponse{}
	bulkOperationResponse.BulkApplicationForEnvironmentPayload = *request
	bulkOperationResponse.Response = response
	return bulkOperationResponse, nil
}

func (impl BulkUpdateServiceImpl) buildHibernateUnHibernateRequestForHelmPipelines(pipeline *pipelineConfig.Pipeline) (*client.AppIdentifier, *openapi.HibernateRequest, error) {
	appIdentifier := &client.AppIdentifier{
		ClusterId:   pipeline.Environment.ClusterId,
		Namespace:   pipeline.Environment.Namespace,
		ReleaseName: pipeline.DeploymentAppName,
	}

	hibernateRequest := &openapi.HibernateRequest{}
	chartInfo, err := impl.chartRefRepository.FetchInfoOfChartConfiguredInApp(pipeline.AppId)
	if err != nil {
		impl.logger.Errorw("error in getting chart info for chart configured in app", "err", err, "appId", pipeline.AppId)
		return nil, nil, err
	}
	var group, kind, version, name string
	name = fmt.Sprintf("%s-%s", pipeline.App.AppName, pipeline.Environment.Name)
	if chartInfo.Name == "" && chartInfo.UserUploaded == false {
		// rollout type chart
		group = "argoproj.io"
		kind = "Rollout"
		version = "v1alpha1"
		hibernateRequest = &openapi.HibernateRequest{
			Resources: &[]openapi.HibernateTargetObject{
				{
					Group:     &group,
					Kind:      &kind,
					Version:   &version,
					Namespace: &pipeline.Environment.Namespace,
					Name:      &name,
				},
			},
		}
	} else if chartInfo.Name == "Deployment" {
		//deployment type chart
		group = "apps"
		kind = "Deployment"
		version = "v1"
		hibernateRequest = &openapi.HibernateRequest{
			Resources: &[]openapi.HibernateTargetObject{
				{
					Group:     &group,
					Kind:      &kind,
					Version:   &version,
					Namespace: &pipeline.Environment.Namespace,
					Name:      &name,
				},
			},
		}
	} else {
		//chart not supported for hibernation, skipping
		impl.logger.Warnw("unsupported chart found for hibernate request, skipping", "pipelineId", pipeline.Id, "chartInfo", chartInfo)
		return nil, nil, nil
	}
	return appIdentifier, hibernateRequest, nil
}
func (impl BulkUpdateServiceImpl) BulkUnHibernate(request *BulkApplicationForEnvironmentPayload, ctx context.Context, w http.ResponseWriter, token string, checkAuthForBulkActions func(token string, appObject string, envObject string) bool) (*BulkApplicationForEnvironmentResponse, error) {
	var pipelines []*pipelineConfig.Pipeline
	var err error
	if len(request.AppIdIncludes) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.EnvId, request.AppIdIncludes)
	} else if len(request.AppIdExcludes) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByNotFilter(request.EnvId, request.AppIdExcludes)
	} else {
		pipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.EnvId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "envId", request.EnvId, "err", err)
		return nil, err
	}
	response := make(map[string]map[string]bool)
	for _, pipeline := range pipelines {
		appKey := fmt.Sprintf("%d_%s", pipeline.AppId, pipeline.App.AppName)
		pipelineKey := fmt.Sprintf("%d_%s", pipeline.Id, pipeline.Name)
		success := true
		if _, ok := response[appKey]; !ok {
			pResponse := make(map[string]bool)
			pResponse[pipelineKey] = false
			response[appKey] = pResponse
		}
		appObject := impl.enforcerUtil.GetAppRBACNameByAppId(pipeline.AppId)
		envObject := impl.enforcerUtil.GetEnvRBACNameByAppId(pipeline.AppId, pipeline.EnvironmentId)
		isValidAuth := checkAuthForBulkActions(token, appObject, envObject)
		if !isValidAuth {
			//skip hibernate for the app if user does not have access on that
			pipelineResponse := response[appKey]
			pipelineResponse[pipelineKey] = false
			response[appKey] = pipelineResponse
			continue
		}
		var hibernateReqError error
		//if pipeline.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD {
		stopRequest := &pipeline1.StopAppRequest{
			AppId:         pipeline.AppId,
			EnvironmentId: pipeline.EnvironmentId,
			UserId:        request.UserId,
			RequestType:   pipeline1.START,
		}
		_, hibernateReqError = impl.workflowDagExecutor.StopStartApp(stopRequest, ctx)
		if hibernateReqError != nil {
			impl.logger.Errorw("error in un-hibernating application", "err", hibernateReqError, "pipeline", pipeline)
			pipelineResponse := response[appKey]
			pipelineResponse[pipelineKey] = false
			response[appKey] = pipelineResponse
			//return nil, err
		}
		pipelineResponse := response[appKey]
		pipelineResponse[pipelineKey] = success
		response[appKey] = pipelineResponse
	}
	bulkOperationResponse := &BulkApplicationForEnvironmentResponse{}
	bulkOperationResponse.BulkApplicationForEnvironmentPayload = *request
	bulkOperationResponse.Response = response
	return bulkOperationResponse, nil
}

func (impl BulkUpdateServiceImpl) BulkDeploy(request *BulkApplicationForEnvironmentPayload, emailId string, checkAuthBatch func(emailId string, appObject []string, envObject []string) (map[string]bool, map[string]bool)) (*BulkApplicationForEnvironmentResponse, error) {
	var pipelines []*pipelineConfig.Pipeline
	var err error

	if len(request.AppNamesIncludes) > 0 {
		r, err := impl.appRepository.FindIdsByNames(request.AppNamesIncludes)
		if err != nil {
			impl.logger.Errorw("error in fetching Ids", "err", err)
			return nil, err
		}
		for _, id := range r {
			request.AppIdIncludes = append(request.AppIdIncludes, id)
		}
	}
	if len(request.AppNamesExcludes) > 0 {
		r, err := impl.appRepository.FindIdsByNames(request.AppNamesExcludes)
		if err != nil {
			impl.logger.Errorw("error in fetching Ids", "err", err)
			return nil, err
		}
		for _, id := range r {
			request.AppIdExcludes = append(request.AppIdExcludes, id)
		}
	}
	if len(request.EnvName) > 0 {
		r, err := impl.environmentRepository.FindByName(request.EnvName)
		if err != nil {
			impl.logger.Errorw("error in fetching env details", "err", err)
			return nil, err
		}
		if request.EnvId != 0 && request.EnvId != r.Id {
			return nil, errors.New("environment id and environment name is different select only one environment")
		} else if request.EnvId == 0 {
			request.EnvId = r.Id
		}
	}
	if len(request.EnvName) == 0 && request.EnvId == 0 {
		return nil, errors.New("please mention environment id or environment name")
	}
	if len(request.AppIdIncludes) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.EnvId, request.AppIdIncludes)
	} else if len(request.AppIdExcludes) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByNotFilter(request.EnvId, request.AppIdExcludes)
	} else {
		pipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.EnvId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "envId", request.EnvId, "err", err)
		return nil, err
	}

	pipelineIds := make([]int, 0)
	for _, pipeline := range pipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
	}
	if len(pipelineIds) == 0 {
		return nil, fmt.Errorf("no pipeline found for this environment")
	}
	//authorization block starts here
	var appObjectArr []string
	var envObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByPipelineIds(pipelineIds)
	pipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	appResults, envResults := checkAuthBatch(emailId, appObjectArr, envObjectArr)
	//authorization block ends here

	response := make(map[string]map[string]bool)
	for _, pipeline := range pipelines {
		appKey := fmt.Sprintf("%d_%s", pipeline.AppId, pipeline.App.AppName)
		pipelineKey := fmt.Sprintf("%d_%s", pipeline.Id, pipeline.Name)
		success := true
		if _, ok := response[appKey]; !ok {
			pResponse := make(map[string]bool)
			pResponse[pipelineKey] = false
			response[appKey] = pResponse
		}
		appObject := objects[pipeline.Id][0]
		envObject := objects[pipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			//if user unauthorized, skip items
			pipelineResponse := response[appKey]
			pipelineResponse[pipelineKey] = false
			response[appKey] = pipelineResponse
			continue
		}

		artifactResponse, err := impl.pipelineBuilder.RetrieveArtifactsByCDPipeline(pipeline, bean.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil {
			impl.logger.Errorw("service err, GetArtifactsByCDPipeline", "err", err, "cdPipelineId", pipeline.Id)
			//return nil, err
			pipelineResponse := response[appKey]
			pipelineResponse[appKey] = false
			response[appKey] = pipelineResponse
		}

		artifacts := artifactResponse.CiArtifacts
		if len(artifacts) == 0 {
			//there is no artifacts found for this pipeline, skip cd trigger
			pipelineResponse := response[appKey]
			pipelineResponse[pipelineKey] = false
			response[appKey] = pipelineResponse
			continue
		}
		artifact := artifacts[0]
		overrideRequest := &bean.ValuesOverrideRequest{
			PipelineId:     pipeline.Id,
			AppId:          pipeline.AppId,
			CiArtifactId:   artifact.Id,
			UserId:         request.UserId,
			CdWorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
		}
		event := &bean.BulkCdDeployEvent{
			ValuesOverrideRequest: overrideRequest,
			UserId:                overrideRequest.UserId,
		}

		payload, err := json.Marshal(event)
		if err != nil {
			impl.logger.Errorw("failed to marshal cd bulk deploy event request",
				"request", event,
				"err", err)

			pipelineResponse := response[appKey]
			pipelineResponse[pipelineKey] = false
			response[appKey] = pipelineResponse
			continue
		}

		err = impl.pubsubClient.Publish(pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC, string(payload))
		if err != nil {
			impl.logger.Errorw("failed to publish trigger request event",
				"topic", pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC,
				"request", overrideRequest,
				"err", err)

			pipelineResponse := response[appKey]
			pipelineResponse[pipelineKey] = false
			response[appKey] = pipelineResponse
			continue
		}

		pipelineResponse := response[appKey]
		pipelineResponse[pipelineKey] = success
		response[appKey] = pipelineResponse
	}
	bulkOperationResponse := &BulkApplicationForEnvironmentResponse{}
	bulkOperationResponse.BulkApplicationForEnvironmentPayload = *request
	bulkOperationResponse.Response = response
	return bulkOperationResponse, nil
}

func (impl BulkUpdateServiceImpl) SubscribeToCdBulkTriggerTopic() error {

	callback := func(msg *pubsub.PubSubMsg) {
		impl.logger.Infow("Event received",
			"topic", pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC,
			"msg", msg.Data)

		event := &bean.BulkCdDeployEvent{}
		err := json.Unmarshal([]byte(msg.Data), event)
		if err != nil {
			impl.logger.Errorw("Error unmarshalling received event",
				"topic", pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC,
				"msg", msg.Data,
				"err", err)
			return
		}
		event.ValuesOverrideRequest.UserId = event.UserId

		// trigger
		ctx, err := impl.buildACDContext()
		if err != nil {
			impl.logger.Errorw("error in creating acd context",
				"err", err)
			return
		}

		_, err = impl.workflowDagExecutor.ManualCdTrigger(event.ValuesOverrideRequest, ctx)
		if err != nil {
			impl.logger.Errorw("Error triggering CD",
				"topic", pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC,
				"msg", msg.Data,
				"err", err)
		}
	}
	err := impl.pubsubClient.Subscribe(pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC, callback)
	if err != nil {
		impl.logger.Error("failed to subscribe to NATS topic",
			"topic", pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC,
			"err", err)
		return err
	}
	return nil
}

func (impl *BulkUpdateServiceImpl) buildACDContext() (acdContext context.Context, err error) {
	//this part only accessible for acd apps hibernation, if acd configured it will fetch latest acdToken, else it will return error
	acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		impl.logger.Errorw("error in getting acd token", "err", err)
		return nil, err
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", acdToken)
	return ctx, nil
}

func (impl BulkUpdateServiceImpl) BulkBuildTrigger(request *BulkApplicationForEnvironmentPayload, ctx context.Context, w http.ResponseWriter, token string, checkAuthForBulkActions func(token string, appObject string, envObject string) bool) (*BulkApplicationForEnvironmentResponse, error) {
	var pipelines []*pipelineConfig.Pipeline
	var err error
	if len(request.AppIdIncludes) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.EnvId, request.AppIdIncludes)
	} else if len(request.AppIdExcludes) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByNotFilter(request.EnvId, request.AppIdExcludes)
	} else {
		pipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.EnvId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "envId", request.EnvId, "err", err)
		return nil, err
	}

	latestCommitsMap := map[int]bean2.CiTriggerRequest{}
	ciCompletedStatus := map[int]bool{}
	for _, pipeline := range pipelines {
		if _, ok := latestCommitsMap[pipeline.CiPipelineId]; !ok {
			ciPipelineId := 0
			ciPipeline, err := impl.ciPipelineRepository.FindById(pipeline.CiPipelineId)
			if err != nil {
				impl.logger.Errorw("error in fetching ci pipeline", "CiPipelineId", pipeline.CiPipelineId, "err", err)
				return nil, err
			}
			ciPipelineId = ciPipeline.Id
			if ciPipeline.IsExternal {
				if _, ok := latestCommitsMap[ciPipeline.ParentCiPipeline]; ok {
					//skip linked ci pipeline for fetching materials if its parent already fetched.
					continue
				}
				ciPipelineId = ciPipeline.ParentCiPipeline
			}

			//if include/exclude configured showAll will include excluded materials also in list, if not configured it will ignore this flag
			materialResponse, err := impl.ciHandler.FetchMaterialsByPipelineId(ciPipelineId, false)
			if err != nil {
				impl.logger.Errorw("error in fetching ci pipeline materials", "CiPipelineId", ciPipelineId, "err", err)
				return nil, err
			}
			var materialId int
			var commitHash string
			for _, material := range materialResponse {
				materialId = material.Id
				if len(material.History) > 0 {
					commitHash = material.History[0].Commit
				}
			}
			var ciMaterials []bean2.CiPipelineMaterial
			ciMaterials = append(ciMaterials, bean2.CiPipelineMaterial{
				Id:        materialId,
				GitCommit: bean2.GitCommit{Commit: commitHash},
			})
			ciTriggerRequest := bean2.CiTriggerRequest{
				PipelineId:         ciPipelineId,
				CiPipelineMaterial: ciMaterials,
				TriggeredBy:        request.UserId,
				InvalidateCache:    request.InvalidateCache,
			}
			latestCommitsMap[ciPipelineId] = ciTriggerRequest
			ciCompletedStatus[ciPipelineId] = false
		}
	}

	response := make(map[string]map[string]bool)
	for _, pipeline := range pipelines {
		ciCompleted := ciCompletedStatus[pipeline.CiPipelineId]
		if !ciCompleted {
			appKey := fmt.Sprintf("%d_%s", pipeline.AppId, pipeline.App.AppName)
			pipelineKey := fmt.Sprintf("%d", pipeline.CiPipelineId)
			success := true
			if _, ok := response[appKey]; !ok {
				pResponse := make(map[string]bool)
				pResponse[pipelineKey] = false
				response[appKey] = pResponse
			}
			appObject := impl.enforcerUtil.GetAppRBACNameByAppId(pipeline.AppId)
			envObject := impl.enforcerUtil.GetEnvRBACNameByAppId(pipeline.AppId, pipeline.EnvironmentId)
			isValidAuth := checkAuthForBulkActions(token, appObject, envObject)
			if !isValidAuth {
				//skip hibernate for the app if user does not have access on that
				pipelineResponse := response[appKey]
				pipelineResponse[pipelineKey] = false
				response[appKey] = pipelineResponse
				continue
			}

			ciTriggerRequest := latestCommitsMap[pipeline.CiPipelineId]
			_, err = impl.ciHandler.HandleCIManual(ciTriggerRequest)
			if err != nil {
				impl.logger.Errorw("service err, HandleCIManual", "err", err, "ciTriggerRequest", ciTriggerRequest)
				//return nil, err
				pipelineResponse := response[appKey]
				pipelineResponse[appKey] = false
				response[appKey] = pipelineResponse
			}

			pipelineResponse := response[appKey]
			pipelineResponse[pipelineKey] = success
			response[appKey] = pipelineResponse
			ciCompletedStatus[pipeline.CiPipelineId] = true
		}
	}
	bulkOperationResponse := &BulkApplicationForEnvironmentResponse{}
	bulkOperationResponse.BulkApplicationForEnvironmentPayload = *request
	bulkOperationResponse.Response = response
	return bulkOperationResponse, nil
}

func (impl BulkUpdateServiceImpl) GetBulkActionImpactedPipelinesAndWfs(dto *CdBulkActionRequestDto) ([]*pipelineConfig.Pipeline, []int, []int, error) {
	var err error
	if (len(dto.EnvIds) == 0 && len(dto.EnvNames) == 0) || ((len(dto.AppIds) == 0 && len(dto.AppNames) == 0) && (len(dto.ProjectIds) == 0 && len(dto.ProjectNames) == 0)) {
		//invalid payload, envIds or envNames are must and at least one of appIds, appNames, projectIds, projectNames is must
		return nil, nil, nil, &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "invalid payload, can not get pipelines for this filter"}
	}
	if len(dto.ProjectIds) > 0 || len(dto.ProjectNames) > 0 {
		appIdsInProjects, err := impl.appRepository.FindIdsByTeamIdsAndTeamNames(dto.ProjectIds, dto.ProjectNames)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting appIds by projectIds and projectNames", "err", err, "projectIds", dto.ProjectIds, "projectNames", dto.ProjectNames)
			return nil, nil, nil, err
		}
		dto.AppIds = append(dto.AppIds, appIdsInProjects...)
	}
	var impactedWfIds []int
	var impactedPipelineIds []int
	var impactedCiPipelineIds []int
	if (len(dto.AppIds) > 0 || len(dto.AppNames) > 0) && (len(dto.EnvIds) > 0 || len(dto.EnvNames) > 0) {
		if len(dto.AppNames) > 0 {
			appIdsByNames, err := impl.appRepository.FindIdsByNames(dto.AppNames)
			if err != nil {
				impl.logger.Errorw("error in getting appIds by names", "err", err, "names", dto.AppNames)
				return nil, nil, nil, err
			}
			dto.AppIds = append(dto.AppIds, appIdsByNames...)
		}
		if len(dto.EnvNames) > 0 {
			envIdsByNames, err := impl.environmentRepository.FindIdsByNames(dto.EnvNames)
			if err != nil {
				impl.logger.Errorw("error in getting envIds by names", "err", err, "names", dto.EnvNames)
				return nil, nil, nil, err
			}
			dto.EnvIds = append(dto.EnvIds, envIdsByNames...)
		}
		if !dto.DeleteWfAndCiPipeline {
			//getting pipeline IDs for app level deletion request
			impactedPipelineIds, err = impl.pipelineRepository.FindIdsByAppIdsAndEnvironmentIds(dto.AppIds, dto.EnvIds)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in getting cd pipelines by appIds and envIds", "err", err)
				return nil, nil, nil, err
			}
		} else {
			//getting all workflows in given apps which do not have pipelines of other than given environments
			appWfsHavingSpecificCdPipelines, err := impl.appWorkflowRepository.FindAllWfsHavingCdPipelinesFromSpecificEnvsOnly(dto.EnvIds, dto.AppIds)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in getting wfs having cd pipelines from specific env only", "err", err)
				return nil, nil, nil, err
			}
			impactedWfIdsMap := make(map[int]bool)
			for _, appWf := range appWfsHavingSpecificCdPipelines {
				if appWf.Type == appWorkflow.CDPIPELINE {
					impactedPipelineIds = append(impactedPipelineIds, appWf.ComponentId)
				}
				if _, ok := impactedWfIdsMap[appWf.AppWorkflowId]; !ok {
					impactedWfIds = append(impactedWfIds, appWf.AppWorkflowId)
					impactedWfIdsMap[appWf.AppWorkflowId] = true
				}
			}
			if len(impactedWfIds) > 0 {
				impactedCiPipelineIds, err = impl.appWorkflowRepository.FindCiPipelineIdsFromAppWfIds(impactedWfIds)
				if err != nil {
					impl.logger.Errorw("error in getting ciPipelineIds from appWfIds", "err", err, "wfIds", impactedWfIds)
					return nil, nil, nil, err
				}
			}
		}
	}
	var pipelines []*pipelineConfig.Pipeline
	if len(impactedPipelineIds) > 0 {
		pipelines, err = impl.pipelineRepository.FindByIdsIn(impactedPipelineIds)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipelines by ids", "err", err, "ids", impactedPipelineIds)
			return nil, nil, nil, err
		}
	}
	return pipelines, impactedWfIds, impactedCiPipelineIds, nil
}

func (impl BulkUpdateServiceImpl) PerformBulkActionOnCdPipelines(dto *CdBulkActionRequestDto, impactedPipelines []*pipelineConfig.Pipeline,
	ctx context.Context, dryRun bool, impactedAppWfIds []int, impactedCiPipelineIds []int) (*PipelineAndWfBulkActionResponseDto, error) {
	switch dto.Action {
	case CD_BULK_DELETE:
		deleteAction := bean2.CASCADE_DELETE
		if dto.ForceDelete {
			deleteAction = bean2.FORCE_DELETE
		} else if dto.NonCascadeDelete {
			deleteAction = bean2.NON_CASCADE_DELETE
		}
		bulkDeleteResp, err := impl.PerformBulkDeleteActionOnCdPipelines(impactedPipelines, ctx, dryRun, deleteAction, dto.DeleteWfAndCiPipeline, impactedAppWfIds, impactedCiPipelineIds, dto.UserId)
		if err != nil {
			impl.logger.Errorw("error in cd pipelines bulk deletion")
		}
		return bulkDeleteResp, nil
	default:
		return nil, &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "this action is not supported"}
	}
}

func (impl BulkUpdateServiceImpl) PerformBulkDeleteActionOnCdPipelines(impactedPipelines []*pipelineConfig.Pipeline, ctx context.Context, dryRun bool, deleteAction int, deleteWfAndCiPipeline bool, impactedAppWfIds, impactedCiPipelineIds []int, userId int32) (*PipelineAndWfBulkActionResponseDto, error) {
	var cdPipelineRespDtos []*CdBulkActionResponseDto
	var wfRespDtos []*WfBulkActionResponseDto
	var ciPipelineRespDtos []*CiBulkActionResponseDto
	//sorting pipelines in decreasing order to tackle problem of sequential pipelines
	//here we are assuming that for now sequential pipelines can only be made through UI and pipelines can not be moved
	//(thus ids in decreasing order should not create problems when deleting)
	//also sorting does not guarantee deletion because impacted pipelines can be sequential but not necessarily linked to each other
	//TODO: implement stack type solution to order pipeline by index in appWfs if pipeline moving is introduced
	if impactedPipelines != nil {
		sort.SliceStable(impactedPipelines, func(i, j int) bool {
			return impactedPipelines[i].Id > impactedPipelines[j].Id
		})
	}
	for _, pipeline := range impactedPipelines {
		respDto := &CdBulkActionResponseDto{
			PipelineName:    pipeline.Name,
			AppName:         pipeline.App.AppName,
			EnvironmentName: pipeline.Environment.Name,
		}
		if !dryRun {
			// Delete Cd pipeline
			deleteResponse, err := impl.pipelineBuilder.DeleteCdPipeline(pipeline, ctx, deleteAction, true, userId)
			if err != nil {
				impl.logger.Errorw("error in deleting cd pipeline", "err", err, "pipelineId", pipeline.Id)
				respDto.DeletionResult = fmt.Sprintf("Not able to delete pipeline, %v", err)
			} else if !(deleteResponse.DeleteInitiated || deleteResponse.ClusterReachable) {
				respDto.DeletionResult = fmt.Sprintf("Not able to delete pipeline, %s, piplineId, %v", "cluster connection error", pipeline.Id)
			} else {
				respDto.DeletionResult = "Pipeline deleted successfully."
			}

		}
		cdPipelineRespDtos = append(cdPipelineRespDtos, respDto)
	}
	if deleteWfAndCiPipeline {
		for _, impactedCiPipelineId := range impactedCiPipelineIds {
			ciPipeline, err := impl.pipelineBuilder.GetCiPipelineById(impactedCiPipelineId)
			if err != nil {
				impl.logger.Errorw("error in getting ciPipeline by id", "err", err, "id", impactedCiPipelineId)
				return nil, err
			}
			respDto := &CiBulkActionResponseDto{
				PipelineName: ciPipeline.Name,
			}
			if !dryRun {
				deleteReq := &bean2.CiPatchRequest{
					Action:     2, //delete
					CiPipeline: ciPipeline,
					AppId:      ciPipeline.AppId,
				}
				_, err = impl.pipelineBuilder.DeleteCiPipeline(deleteReq)
				if err != nil {
					impl.logger.Errorw("error in deleting ci pipeline", "err", err, "pipelineId", impactedCiPipelineId)
					respDto.DeletionResult = fmt.Sprintf("Not able to delete pipeline, %v", err)
				} else {
					respDto.DeletionResult = "Pipeline deleted successfully."
				}
			}
			ciPipelineRespDtos = append(ciPipelineRespDtos, respDto)
		}

		for _, impactedAppWfId := range impactedAppWfIds {
			respDto := &WfBulkActionResponseDto{
				WorkflowId: impactedAppWfId,
			}
			if !dryRun {
				err := impl.appWorkflowService.DeleteAppWorkflow(impactedAppWfId, userId)
				if err != nil {
					impl.logger.Errorw("error in deleting appWf", "err", err, "appWfId", impactedAppWfId)
					respDto.DeletionResult = fmt.Sprintf("Not able to delete workflow, %v", err)
				} else {
					respDto.DeletionResult = "Workflow deleted successfully."
				}
			}
			wfRespDtos = append(wfRespDtos, respDto)
		}

	}
	respDto := &PipelineAndWfBulkActionResponseDto{
		CdPipelinesRespDtos: cdPipelineRespDtos,
		CiPipelineRespDtos:  ciPipelineRespDtos,
		AppWfRespDtos:       wfRespDtos,
	}
	return respDto, nil

}
