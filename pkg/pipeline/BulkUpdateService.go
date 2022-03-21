package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/go-pg/pg"
	"net/http"

	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/bulkUpdate"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

type NameIncludesExcludes struct {
	Names []string `json:"names"`
}
type DeploymentTemplateSpec struct {
	PatchJson string `json:"patchJson"`
}
type DeploymentTemplateTask struct {
	Spec *DeploymentTemplateSpec `json:"spec"`
}
type CmAndSecretSpec struct {
	Names     []string `json:"names"`
	PatchJson string   `json:"patchJson"`
}
type CmAndSecretTask struct {
	Spec *CmAndSecretSpec `json:"spec"`
}
type BulkUpdatePayload struct {
	Includes           *NameIncludesExcludes   `json:"includes"`
	Excludes           *NameIncludesExcludes   `json:"excludes"`
	EnvIds             []int                   `json:"envIds"`
	Global             bool                    `json:"global"`
	DeploymentTemplate *DeploymentTemplateTask `json:"deploymentTemplate"`
	ConfigMap          *CmAndSecretTask        `json:"configMap"`
	Secret             *CmAndSecretTask        `json:"secret"`
}
type BulkUpdateScript struct {
	ApiVersion string             `json:"apiVersion" validate:"required"`
	Kind       string             `json:"kind" validate:"required"`
	Spec       *BulkUpdatePayload `json:"spec" validate:"required"`
}
type BulkUpdateSeeExampleResponse struct {
	Operation string            `json:"operation"`
	Script    *BulkUpdateScript `json:"script" validate:"required"`
	ReadMe    string            `json:"readme"`
}
type ImpactedObjectsResponse struct {
	DeploymentTemplate []*DeploymentTemplateImpactedObjectsResponseForOneApp `json:"deploymentTemplate"`
	ConfigMap          []*CmAndSecretImpactedObjectsResponseForOneApp        `json:"configMap"`
	Secret             []*CmAndSecretImpactedObjectsResponseForOneApp        `json:"secret"`
}
type DeploymentTemplateImpactedObjectsResponseForOneApp struct {
	AppId   int    `json:"appId"`
	AppName string `json:"appName"`
	EnvId   int    `json:"envId"`
}
type CmAndSecretImpactedObjectsResponseForOneApp struct {
	AppId   int      `json:"appId"`
	AppName string   `json:"appName"`
	EnvId   int      `json:"envId"`
	Names   []string `json:"names"`
}
type DeploymentTemplateBulkUpdateResponseForOneApp struct {
	AppId   int    `json:"appId"`
	AppName string `json:"appName"`
	EnvId   int    `json:"envId"`
	Message string `json:"message"`
}
type CmAndSecretBulkUpdateResponseForOneApp struct {
	AppId   int      `json:"appId"`
	AppName string   `json:"appName"`
	EnvId   int      `json:"envId"`
	Names   []string `json:"names"`
	Message string   `json:"message"`
}
type BulkUpdateResponse struct {
	DeploymentTemplate *DeploymentTemplateBulkUpdateResponse `json:"deploymentTemplate"`
	ConfigMap          *CmAndSecretBulkUpdateResponse        `json:"configMap"`
	Secret             *CmAndSecretBulkUpdateResponse        `json:"secret"`
}
type DeploymentTemplateBulkUpdateResponse struct {
	Message    []string                                         `json:"message"`
	Failure    []*DeploymentTemplateBulkUpdateResponseForOneApp `json:"failure"`
	Successful []*DeploymentTemplateBulkUpdateResponseForOneApp `json:"successful"`
}
type CmAndSecretBulkUpdateResponse struct {
	Message    []string                                  `json:"message"`
	Failure    []*CmAndSecretBulkUpdateResponseForOneApp `json:"failure"`
	Successful []*CmAndSecretBulkUpdateResponseForOneApp `json:"successful"`
}
type BulkUpdateService interface {
	FindBulkUpdateReadme(operation string) (response *BulkUpdateSeeExampleResponse, err error)
	GetBulkAppName(bulkUpdateRequest *BulkUpdatePayload) (*ImpactedObjectsResponse, error)
	ApplyJsonPatch(patch jsonpatch.Patch, target string) (string, error)
	BulkUpdateDeploymentTemplate(bulkUpdatePayload *BulkUpdatePayload) *DeploymentTemplateBulkUpdateResponse
	BulkUpdateConfigMap(bulkUpdatePayload *BulkUpdatePayload) *CmAndSecretBulkUpdateResponse
	BulkUpdateSecret(bulkUpdatePayload *BulkUpdatePayload) *CmAndSecretBulkUpdateResponse
	BulkUpdate(bulkUpdateRequest *BulkUpdatePayload) (bulkUpdateResponse *BulkUpdateResponse)
}

type BulkUpdateServiceImpl struct {
	bulkUpdateRepository             bulkUpdate.BulkUpdateRepository
	chartRepository                  chartRepoRepository.ChartRepository
	logger                           *zap.SugaredLogger
	repoRepository                   chartRepoRepository.ChartRepoRepository
	chartTemplateService             util.ChartTemplateService
	mergeUtil                        util.MergeUtil
	repositoryService                repository.ServiceClient
	refChartDir                      RefChartDir
	defaultChart                     DefaultChart
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
}

func NewBulkUpdateServiceImpl(bulkUpdateRepository bulkUpdate.BulkUpdateRepository,
	chartRepository chartRepoRepository.ChartRepository,
	logger *zap.SugaredLogger,
	chartTemplateService util.ChartTemplateService,
	repoRepository chartRepoRepository.ChartRepoRepository,
	refChartDir RefChartDir,
	defaultChart DefaultChart,
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
	configMapHistoryService history.ConfigMapHistoryService) *BulkUpdateServiceImpl {
	return &BulkUpdateServiceImpl{
		bulkUpdateRepository:             bulkUpdateRepository,
		chartRepository:                  chartRepository,
		logger:                           logger,
		chartTemplateService:             chartTemplateService,
		repoRepository:                   repoRepository,
		mergeUtil:                        mergeUtil,
		refChartDir:                      refChartDir,
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
	}
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

	if len(bulkUpdatePayload.Includes.Names) == 0 {
		return impactedObjectsResponse, nil
	}
	if bulkUpdatePayload.Global {
		//For Deployment Template
		if bulkUpdatePayload.DeploymentTemplate != nil && bulkUpdatePayload.DeploymentTemplate.Spec != nil {
			appsGlobalDT, err := impl.bulkUpdateRepository.
				FindDeploymentTemplateBulkAppNameForGlobal(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names)
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
			configMapAppModels, err := impl.bulkUpdateRepository.FindCMBulkAppModelForGlobal(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names, bulkUpdatePayload.ConfigMap.Spec.Names)
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
			secretAppModels, err := impl.bulkUpdateRepository.FindSecretBulkAppModelForGlobal(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names, bulkUpdatePayload.Secret.Spec.Names)
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
				FindDeploymentTemplateBulkAppNameForEnv(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names, envId)
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
			configMapEnvModels, err := impl.bulkUpdateRepository.FindCMBulkAppModelForEnv(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names, envId, bulkUpdatePayload.ConfigMap.Spec.Names)
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
			secretEnvModels, err := impl.bulkUpdateRepository.FindSecretBulkAppModelForEnv(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names, envId, bulkUpdatePayload.Secret.Spec.Names)
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
	if len(bulkUpdatePayload.Includes.Names) == 0 {
		deploymentTemplateBulkUpdateResponse.Message = append(deploymentTemplateBulkUpdateResponse.Message, "Please don't leave includes.names array empty")
		return deploymentTemplateBulkUpdateResponse
	}
	deploymentTemplatePatchJson := []byte(bulkUpdatePayload.DeploymentTemplate.Spec.PatchJson)
	deploymentTemplatePatch, err := jsonpatch.DecodePatch(deploymentTemplatePatchJson)
	if err != nil {
		impl.logger.Errorw("error in decoding JSON patch", "err", err)
		deploymentTemplateBulkUpdateResponse.Message = append(deploymentTemplateBulkUpdateResponse.Message, "The patch string you entered seems wrong, please check and try again")
	}
	var charts []*chartRepoRepository.Chart
	if bulkUpdatePayload.Global {
		charts, err = impl.bulkUpdateRepository.FindBulkChartsByAppNameSubstring(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names)
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
						}
					}
				}
			}
		}
	}
	var chartsEnv []*chartConfig.EnvConfigOverride
	for _, envId := range bulkUpdatePayload.EnvIds {
		chartsEnv, err = impl.bulkUpdateRepository.FindBulkChartsEnvByAppNameSubstring(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names, envId)
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

func (impl BulkUpdateServiceImpl) BulkUpdateConfigMap(bulkUpdatePayload *BulkUpdatePayload) *CmAndSecretBulkUpdateResponse {
	configMapBulkUpdateResponse := &CmAndSecretBulkUpdateResponse{}
	if len(bulkUpdatePayload.Includes.Names) == 0 {
		configMapBulkUpdateResponse.Message = append(configMapBulkUpdateResponse.Message, "Please don't leave includes.names array empty")
		return configMapBulkUpdateResponse
	}
	if bulkUpdatePayload.Global {
		configMapSpecNames := make(map[string]bool)
		for _, name := range bulkUpdatePayload.ConfigMap.Spec.Names {
			configMapSpecNames[name] = true
		}
		configMapAppModels, err := impl.bulkUpdateRepository.FindCMBulkAppModelForGlobal(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names, bulkUpdatePayload.ConfigMap.Spec.Names)
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
		configMapEnvModels, err := impl.bulkUpdateRepository.FindCMBulkAppModelForEnv(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names, envId, bulkUpdatePayload.ConfigMap.Spec.Names)
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
	if len(bulkUpdatePayload.Includes.Names) == 0 {
		secretBulkUpdateResponse.Message = append(secretBulkUpdateResponse.Message, "Please don't leave includes.names array empty")
		return secretBulkUpdateResponse
	}
	if bulkUpdatePayload.Global {
		secretSpecNames := make(map[string]bool)
		for _, name := range bulkUpdatePayload.Secret.Spec.Names {
			secretSpecNames[name] = true
		}
		secretAppModels, err := impl.bulkUpdateRepository.FindSecretBulkAppModelForGlobal(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names, bulkUpdatePayload.Secret.Spec.Names)
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
							impl.logger.Errorw("error in bulk updating charts", "err", err)
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
		secretEnvModels, err := impl.bulkUpdateRepository.FindSecretBulkAppModelForEnv(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names, envId, bulkUpdatePayload.Secret.Spec.Names)
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
