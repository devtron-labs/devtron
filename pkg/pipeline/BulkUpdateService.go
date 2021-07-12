package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/bulkUpdate"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	jsonpatch "github.com/evanphx/json-patch"
	"go.uber.org/zap"
	"net/http"
)

type NameIncludesExcludes struct {
	Names []string `json:"names"`
}
type Spec struct {
	PatchJson string `json:"patchJson"`
}
type Tasks struct {
	Spec Spec `json:"spec"`
}
type BulkUpdatePayload struct {
	Includes           NameIncludesExcludes `json:"includes"`
	Excludes           NameIncludesExcludes `json:"excludes"`
	EnvIds             []int                `json:"envIds"`
	Global             bool                 `json:"global"`
	DeploymentTemplate Tasks                `json:"deploymentTemplate"`
}
type BulkUpdateScript struct {
	ApiVersion string            `json:"apiVersion" validate:"required"`
	Kind       string            `json:"kind" validate:"required"`
	Spec       BulkUpdatePayload `json:"spec" validate:"required"`
}
type BulkUpdateSeeExampleResponse struct {
	Operation string           `json:"operation"`
	Script    BulkUpdateScript `json:"script" validate:"required"`
	ReadMe    string           `json:"readme"`
}
type ImpactedObjectsResponse struct {
	AppId   int    `json:"appId"`
	AppName string `json:"appName"`
	EnvId   int    `json:"envId"`
}
type BulkUpdateResponseStatusForOneApp struct {
	AppId   int    `json:"appId"`
	AppName string `json:"appName"`
	EnvId   int    `json:"envId"`
	Message string `json:"message"`
}
type BulkUpdateResponse struct {
	Message    []string                             `json:"message"`
	Failure    []*BulkUpdateResponseStatusForOneApp `json:"failure"`
	Successful []*BulkUpdateResponseStatusForOneApp `json:"successful"`
}
type BulkUpdateService interface {
	FindBulkUpdateReadme(operation string) (response BulkUpdateSeeExampleResponse, err error)
	GetBulkAppName(bulkUpdateRequest BulkUpdatePayload) ([]*ImpactedObjectsResponse, error)
	ApplyJsonPatch(patch jsonpatch.Patch, target string) (string, error)
	BulkUpdateDeploymentTemplate(bulkUpdateRequest BulkUpdatePayload) (bulkUpdateResponse BulkUpdateResponse)
}

type BulkUpdateServiceImpl struct {
	bulkUpdateRepository      bulkUpdate.BulkUpdateRepository
	chartRepository           chartConfig.ChartRepository
	logger                    *zap.SugaredLogger
	repoRepository            chartConfig.ChartRepoRepository
	chartTemplateService      util.ChartTemplateService
	pipelineGroupRepository   pipelineConfig.AppRepository
	mergeUtil                 util.MergeUtil
	repositoryService         repository.ServiceClient
	refChartDir               RefChartDir
	defaultChart              DefaultChart
	chartRefRepository        chartConfig.ChartRefRepository
	envOverrideRepository     chartConfig.EnvConfigOverrideRepository
	pipelineConfigRepository  chartConfig.PipelineConfigRepository
	configMapRepository       chartConfig.ConfigMapRepository
	environmentRepository     cluster.EnvironmentRepository
	pipelineRepository        pipelineConfig.PipelineRepository
	appLevelMetricsRepository repository3.AppLevelMetricsRepository
	client                    *http.Client
}

func NewBulkUpdateServiceImpl(bulkUpdateRepository bulkUpdate.BulkUpdateRepository,
	chartRepository chartConfig.ChartRepository,
	logger *zap.SugaredLogger,
	chartTemplateService util.ChartTemplateService,
	repoRepository chartConfig.ChartRepoRepository,
	pipelineGroupRepository pipelineConfig.AppRepository,
	refChartDir RefChartDir,
	defaultChart DefaultChart,
	mergeUtil util.MergeUtil,
	repositoryService repository.ServiceClient,
	chartRefRepository chartConfig.ChartRefRepository,
	envOverrideRepository chartConfig.EnvConfigOverrideRepository,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	configMapRepository chartConfig.ConfigMapRepository,
	environmentRepository cluster.EnvironmentRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	appLevelMetricsRepository repository3.AppLevelMetricsRepository,
	client *http.Client,
) *BulkUpdateServiceImpl {
	return &BulkUpdateServiceImpl{
		bulkUpdateRepository:      bulkUpdateRepository,
		chartRepository:           chartRepository,
		logger:                    logger,
		chartTemplateService:      chartTemplateService,
		repoRepository:            repoRepository,
		pipelineGroupRepository:   pipelineGroupRepository,
		mergeUtil:                 mergeUtil,
		refChartDir:               refChartDir,
		defaultChart:              defaultChart,
		repositoryService:         repositoryService,
		chartRefRepository:        chartRefRepository,
		envOverrideRepository:     envOverrideRepository,
		pipelineConfigRepository:  pipelineConfigRepository,
		configMapRepository:       configMapRepository,
		environmentRepository:     environmentRepository,
		pipelineRepository:        pipelineRepository,
		appLevelMetricsRepository: appLevelMetricsRepository,
		client:                    client,
	}
}
func (impl BulkUpdateServiceImpl) FindBulkUpdateReadme(operation string) (BulkUpdateSeeExampleResponse, error) {
	bulkUpdateReadme, err := impl.bulkUpdateRepository.FindBulkUpdateReadme(operation)
	response := BulkUpdateSeeExampleResponse{}
	if err != nil {
		impl.logger.Errorw("error in fetching batch operation example", "err", err)
		return response, err
	}
	script := BulkUpdateScript{}
	err = json.Unmarshal([]byte(bulkUpdateReadme.Script), &script)
	if err != nil {
		impl.logger.Errorw("error in script value(in db) of batch operation example", "err", err)
		return response, err
	}
	response = BulkUpdateSeeExampleResponse{
		Operation: bulkUpdateReadme.Resource,
		Script:    script,
		ReadMe:    bulkUpdateReadme.Readme,
	}
	return response, nil
}
func (impl BulkUpdateServiceImpl) GetBulkAppName(bulkUpdatePayload BulkUpdatePayload) ([]*ImpactedObjectsResponse, error) {
	impactedObjectsResponse := []*ImpactedObjectsResponse{}
	if len(bulkUpdatePayload.Includes.Names) == 0 {
		return impactedObjectsResponse, nil
	}
	if bulkUpdatePayload.Global {
		appsGlobal, err := impl.bulkUpdateRepository.
			FindBulkAppNameForGlobal(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names)
		if err != nil {
			impl.logger.Errorw("error in fetching bulk app names for global", "err", err)
			return nil, err
		}
		for _, app := range appsGlobal {
			impactedObject := &ImpactedObjectsResponse{
				AppId:   app.Id,
				AppName: app.AppName,
			}
			impactedObjectsResponse = append(impactedObjectsResponse, impactedObject)
		}
	}
	for _, envId := range bulkUpdatePayload.EnvIds {
		appsNotGlobal, err := impl.bulkUpdateRepository.
			FindBulkAppNameForEnv(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names, envId)
		if err != nil {
			impl.logger.Errorw("error in fetching bulk app names for env", "err", err)
			return nil, err
		}
		for _, app := range appsNotGlobal {
			impactedObject := &ImpactedObjectsResponse{
				AppId:   app.Id,
				AppName: app.AppName,
				EnvId:   envId,
			}
			impactedObjectsResponse = append(impactedObjectsResponse, impactedObject)
		}
	}
	return impactedObjectsResponse, nil
}
func (impl BulkUpdateServiceImpl) ApplyJsonPatch(patch jsonpatch.Patch, target string) (string, error) {
	modified, err := patch.Apply([]byte(target))
	if err != nil {
		impl.logger.Errorw("error in applying JSON patch","err",err)
		return "Patch Failed", err
	}
	return string(modified), err
}
func (impl BulkUpdateServiceImpl) BulkUpdateDeploymentTemplate(bulkUpdatePayload BulkUpdatePayload) BulkUpdateResponse {
	var bulkUpdateResponse BulkUpdateResponse
	if len(bulkUpdatePayload.Includes.Names) == 0 {
		bulkUpdateResponse.Message = append(bulkUpdateResponse.Message, "Please don't leave includes.names array empty")
		return bulkUpdateResponse
	}
	patchJson := []byte(bulkUpdatePayload.DeploymentTemplate.Spec.PatchJson)
	patch, err := jsonpatch.DecodePatch(patchJson)
	if err != nil {
		impl.logger.Errorw("error in decoding JSON patch", "err", err)
		bulkUpdateResponse.Message = append(bulkUpdateResponse.Message, "The patch string you entered seems wrong, please check and try again")
		return bulkUpdateResponse
	}
	var charts []*chartConfig.Chart
	if bulkUpdatePayload.Global {
		charts, err = impl.bulkUpdateRepository.FindBulkChartsByAppNameSubstring(bulkUpdatePayload.Includes.Names, bulkUpdatePayload.Excludes.Names)
		if err != nil {
			impl.logger.Error("error in fetching charts by app name substring")
			bulkUpdateResponse.Message = append(bulkUpdateResponse.Message, fmt.Sprintf("Unable to bulk update apps globally : %s", err.Error()))
		} else {
			if len(charts) == 0 {
				bulkUpdateResponse.Message = append(bulkUpdateResponse.Message, "No matching apps to update globally")
			} else {
				for _, chart := range charts {
					appDetailsByChart, _ := impl.bulkUpdateRepository.FindAppByChartId(chart.Id)
					modified, err := impl.ApplyJsonPatch(patch, chart.Values)
					if err != nil {
						impl.logger.Errorw("error in applying JSON patch","err",err)
						bulkUpdateFailedResponse := &BulkUpdateResponseStatusForOneApp{
							AppId:   appDetailsByChart.Id,
							AppName: appDetailsByChart.AppName,
							Message: fmt.Sprintf("Error in applying JSON patch : %s", err.Error()),
						}
						bulkUpdateResponse.Failure = append(bulkUpdateResponse.Failure, bulkUpdateFailedResponse)
					} else {
						err = impl.bulkUpdateRepository.BulkUpdateChartsValuesYamlAndGlobalOverrideById(chart.Id, modified)
						if err != nil {
							impl.logger.Errorw("error in bulk updating charts","err",err)
							bulkUpdateFailedResponse := &BulkUpdateResponseStatusForOneApp{
								AppId:   appDetailsByChart.Id,
								AppName: appDetailsByChart.AppName,
								Message: fmt.Sprintf("Error in updating in db : %s", err.Error()),
							}
							bulkUpdateResponse.Failure = append(bulkUpdateResponse.Failure, bulkUpdateFailedResponse)
						} else {
							bulkUpdateSuccessResponse := &BulkUpdateResponseStatusForOneApp{
								AppId:   appDetailsByChart.Id,
								AppName: appDetailsByChart.AppName,
								Message: "Updated Successfully",
							}
							bulkUpdateResponse.Successful = append(bulkUpdateResponse.Successful, bulkUpdateSuccessResponse)
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
			bulkUpdateResponse.Message = append(bulkUpdateResponse.Message, fmt.Sprintf("Unable to bulk update apps for envId = %d , %s", envId, err.Error()))
		} else {
			if len(chartsEnv) == 0 {
				bulkUpdateResponse.Message = append(bulkUpdateResponse.Message, fmt.Sprintf("No matching apps to update for envId = %d", envId))
			} else {
				for _, chartEnv := range chartsEnv {
					appDetailsByChart, _ := impl.bulkUpdateRepository.FindAppByChartEnvId(chartEnv.Id)
					modified, err := impl.ApplyJsonPatch(patch, chartEnv.EnvOverrideValues)
					if err != nil {
						impl.logger.Errorw("error in applying JSON patch", "err", err)
						bulkUpdateFailedResponse := &BulkUpdateResponseStatusForOneApp{
							AppId:   appDetailsByChart.Id,
							AppName: appDetailsByChart.AppName,
							EnvId:   envId,
							Message: fmt.Sprintf("Error in applying JSON patch : %s", err.Error()),
						}
						bulkUpdateResponse.Failure = append(bulkUpdateResponse.Failure, bulkUpdateFailedResponse)
					} else {
						err = impl.bulkUpdateRepository.BulkUpdateChartsEnvYamlOverrideById(chartEnv.Id, modified)
						if err != nil {
							impl.logger.Errorw("error in bulk updating charts","err",err)
							bulkUpdateFailedResponse := &BulkUpdateResponseStatusForOneApp{
								AppId:   appDetailsByChart.Id,
								AppName: appDetailsByChart.AppName,
								EnvId:   envId,
								Message: fmt.Sprintf("Error in updating in db : %s", err.Error()),
							}
							bulkUpdateResponse.Failure = append(bulkUpdateResponse.Failure, bulkUpdateFailedResponse)
						} else {
							bulkUpdateSuccessResponse := &BulkUpdateResponseStatusForOneApp{
								AppId:   appDetailsByChart.Id,
								AppName: appDetailsByChart.AppName,
								EnvId:   envId,
								Message: "Updated Successfully",
							}
							bulkUpdateResponse.Successful = append(bulkUpdateResponse.Successful, bulkUpdateSuccessResponse)
						}
					}
				}
			}
		}
	}
	if len(bulkUpdateResponse.Failure) == 0 {
		bulkUpdateResponse.Message = append(bulkUpdateResponse.Message, "All matching apps are updated successfully")
	}
	return bulkUpdateResponse
}