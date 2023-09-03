/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	repository5 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"time"

	chartService "github.com/devtron-labs/devtron/pkg/chart"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/sql"

	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.uber.org/zap"
)

type PropertiesConfigService interface {
	CreateEnvironmentProperties(appId int, propertiesRequest *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error)
	UpdateEnvironmentProperties(appId int, propertiesRequest *bean.EnvironmentProperties, userId int32) (*bean.EnvironmentProperties, error)
	//create environment entry for each new environment
	CreateIfRequired(chart *chartRepoRepository.Chart, environmentId int, userId int32, manualReviewed bool, chartStatus models.ChartStatus, isOverride, isAppMetricsEnabled bool, namespace string, IsBasicViewLocked bool, CurrentViewEditor models.ChartsViewEditorType, tx *pg.Tx) (*chartConfig.EnvConfigOverride, error)
	GetEnvironmentProperties(appId, environmentId int, chartRefId int) (environmentPropertiesResponse *bean.EnvironmentPropertiesResponse, err error)
	GetEnvironmentPropertiesById(environmentId int) ([]bean.EnvironmentProperties, error)

	GetAppIdByChartEnvId(chartEnvId int) (*chartConfig.EnvConfigOverride, error)
	GetLatestEnvironmentProperties(appId, environmentId int) (*bean.EnvironmentProperties, error)
	FindEnvLevelAppMetricsByAppIdAndEnvId(appId int, envId int) (*repository.EnvLevelAppMetrics, error)
	ResetEnvironmentProperties(id int) (bool, error)
	CreateEnvironmentPropertiesWithNamespace(appId int, propertiesRequest *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error)

	EnvMetricsEnableDisable(appMetricRequest *chartService.AppMetricEnableDisableRequest) (*chartService.AppMetricEnableDisableRequest, error)
	FetchEnvProperties(appId, envId, chartRefId int) (*chartConfig.EnvConfigOverride, error)
}
type PropertiesConfigServiceImpl struct {
	logger                           *zap.SugaredLogger
	envConfigRepo                    chartConfig.EnvConfigOverrideRepository
	chartRepo                        chartRepoRepository.ChartRepository
	chartRefRepository               chartRepoRepository.ChartRefRepository
	mergeUtil                        util.MergeUtil
	environmentRepository            repository2.EnvironmentRepository
	ciCdPipelineOrchestrator         CiCdPipelineOrchestrator
	application                      application.ServiceClient
	envLevelAppMetricsRepository     repository.EnvLevelAppMetricsRepository
	appLevelMetricsRepository        repository.AppLevelMetricsRepository
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService
	variableEntityMappingService     variables.VariableEntityMappingService
	variableTemplateParser           parsers.VariableTemplateParser
}

func NewPropertiesConfigServiceImpl(logger *zap.SugaredLogger,
	envConfigRepo chartConfig.EnvConfigOverrideRepository,
	chartRepo chartRepoRepository.ChartRepository,
	chartRefRepository chartRepoRepository.ChartRefRepository,
	mergeUtil util.MergeUtil,
	environmentRepository repository2.EnvironmentRepository,
	ciCdPipelineOrchestrator CiCdPipelineOrchestrator,
	application application.ServiceClient,
	envLevelAppMetricsRepository repository.EnvLevelAppMetricsRepository,
	appLevelMetricsRepository repository.AppLevelMetricsRepository,
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService,
	variableEntityMappingService variables.VariableEntityMappingService,
	variableTemplateParser parsers.VariableTemplateParser) *PropertiesConfigServiceImpl {
	return &PropertiesConfigServiceImpl{
		logger:                           logger,
		envConfigRepo:                    envConfigRepo,
		chartRepo:                        chartRepo,
		chartRefRepository:               chartRefRepository,
		mergeUtil:                        mergeUtil,
		environmentRepository:            environmentRepository,
		ciCdPipelineOrchestrator:         ciCdPipelineOrchestrator,
		application:                      application,
		envLevelAppMetricsRepository:     envLevelAppMetricsRepository,
		appLevelMetricsRepository:        appLevelMetricsRepository,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
		variableEntityMappingService:     variableEntityMappingService,
		variableTemplateParser:           variableTemplateParser,
	}

}

func (impl PropertiesConfigServiceImpl) GetEnvironmentProperties(appId, environmentId int, chartRefId int) (environmentPropertiesResponse *bean.EnvironmentPropertiesResponse, err error) {
	environmentPropertiesResponse = &bean.EnvironmentPropertiesResponse{}
	env, err := impl.environmentRepository.FindById(environmentId)
	if err != nil {
		return nil, err
	}
	if len(env.Namespace) > 0 {
		environmentPropertiesResponse.Namespace = env.Namespace
	}

	// step 1
	envOverride, err := impl.envConfigRepo.ActiveEnvConfigOverride(appId, environmentId)
	if err != nil {
		return nil, err
	}
	environmentProperties := &bean.EnvironmentProperties{}
	if envOverride.Id > 0 {
		r := json.RawMessage{}
		if envOverride.IsOverride {
			err = r.UnmarshalJSON([]byte(envOverride.EnvOverrideValues))
			if err != nil {
				return nil, err
			}
			environmentPropertiesResponse.IsOverride = true
		} else {
			err = r.UnmarshalJSON([]byte(envOverride.EnvOverrideValues))
			if err != nil {
				return nil, err
			}
		}
		environmentProperties = &bean.EnvironmentProperties{
			//Id:                envOverride.Id,
			Status:            envOverride.Status,
			EnvOverrideValues: r,
			ManualReviewed:    envOverride.ManualReviewed,
			Active:            envOverride.Active,
			Namespace:         env.Namespace,
			Description:       env.Description,
			EnvironmentId:     environmentId,
			EnvironmentName:   env.Name,
			Latest:            envOverride.Latest,
			//ChartRefId:        chartRefId,
			IsOverride:        envOverride.IsOverride,
			IsBasicViewLocked: envOverride.IsBasicViewLocked,
			CurrentViewEditor: envOverride.CurrentViewEditor,
		}

		if environmentPropertiesResponse.Namespace == "" {
			environmentPropertiesResponse.Namespace = envOverride.Namespace
		}
	}
	ecOverride, err := impl.envConfigRepo.FindChartByAppIdAndEnvIdAndChartRefId(appId, environmentId, chartRefId)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	if errors.IsNotFound(err) {
		environmentProperties.Id = 0
		environmentProperties.ChartRefId = chartRefId
		environmentProperties.IsOverride = false
	} else {
		environmentProperties.Id = ecOverride.Id
		environmentProperties.Latest = ecOverride.Latest
		environmentProperties.IsOverride = ecOverride.IsOverride
		environmentProperties.ChartRefId = chartRefId
		environmentProperties.ManualReviewed = ecOverride.ManualReviewed
		environmentProperties.Status = ecOverride.Status
		environmentProperties.Namespace = ecOverride.Namespace
		environmentProperties.Active = ecOverride.Active
		environmentProperties.IsBasicViewLocked = ecOverride.IsBasicViewLocked
		environmentProperties.CurrentViewEditor = ecOverride.CurrentViewEditor
	}
	environmentPropertiesResponse.ChartRefId = chartRefId
	environmentPropertiesResponse.EnvironmentConfig = *environmentProperties

	//setting global config
	chart, err := impl.chartRepo.FindLatestChartForAppByAppId(appId)
	if err != nil {
		return nil, err
	}
	if chart != nil && chart.Id > 0 {
		globalOverride := []byte(chart.GlobalOverride)
		environmentPropertiesResponse.GlobalConfig = globalOverride
		environmentPropertiesResponse.GlobalChartRefId = chart.ChartRefId
		if !environmentPropertiesResponse.IsOverride {
			environmentPropertiesResponse.EnvironmentConfig.IsBasicViewLocked = chart.IsBasicViewLocked
			environmentPropertiesResponse.EnvironmentConfig.CurrentViewEditor = chart.CurrentViewEditor
		}
	}

	envLevelMetrics, err := impl.envLevelAppMetricsRepository.FindByAppIdAndEnvId(appId, environmentId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return nil, err
	}
	if util.IsErrNoRows(err) {
		appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(appId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in app metrics app level flag", "error", err)
			return nil, err
		}
		if util.IsErrNoRows(err) {
			flag := false
			environmentPropertiesResponse.AppMetrics = &flag
		} else {
			environmentPropertiesResponse.AppMetrics = &appLevelMetrics.AppMetrics
		}
	} else {
		environmentPropertiesResponse.AppMetrics = envLevelMetrics.AppMetrics
	}
	return environmentPropertiesResponse, nil
}

func (impl PropertiesConfigServiceImpl) FetchEnvProperties(appId, envId, chartRefId int) (*chartConfig.EnvConfigOverride, error) {
	return impl.envConfigRepo.GetByAppIdEnvIdAndChartRefId(appId, envId, chartRefId)
}

func (impl PropertiesConfigServiceImpl) CreateEnvironmentProperties(appId int, environmentProperties *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error) {
	chart, err := impl.chartRepo.FindChartByAppIdAndRefId(appId, environmentProperties.ChartRefId)
	if err != nil && pg.ErrNoRows != err {
		return nil, err
	}
	if pg.ErrNoRows == err {
		impl.logger.Errorw("create new chart set latest=false", "a", "b")
		return nil, fmt.Errorf("NOCHARTEXIST")
	}
	chart.GlobalOverride = string(environmentProperties.EnvOverrideValues)
	appMetrics := false
	if environmentProperties.AppMetrics != nil {
		appMetrics = *environmentProperties.AppMetrics
	}
	envOverride, err := impl.CreateIfRequired(chart, environmentProperties.EnvironmentId, environmentProperties.UserId, environmentProperties.ManualReviewed, models.CHARTSTATUS_SUCCESS, true, appMetrics, environmentProperties.Namespace, environmentProperties.IsBasicViewLocked, environmentProperties.CurrentViewEditor, nil)
	if err != nil {
		return nil, err
	}

	r := json.RawMessage{}
	err = r.UnmarshalJSON([]byte(envOverride.EnvOverrideValues))
	if err != nil {
		return nil, err
	}
	env, err := impl.environmentRepository.FindById(environmentProperties.EnvironmentId)
	if err != nil {
		return nil, err
	}
	environmentProperties = &bean.EnvironmentProperties{
		Id:                envOverride.Id,
		Status:            envOverride.Status,
		EnvOverrideValues: r,
		ManualReviewed:    envOverride.ManualReviewed,
		Active:            envOverride.Active,
		Namespace:         env.Namespace,
		EnvironmentId:     environmentProperties.EnvironmentId,
		EnvironmentName:   env.Name,
		Latest:            envOverride.Latest,
		ChartRefId:        environmentProperties.ChartRefId,
		IsOverride:        envOverride.IsOverride,
	}

	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err, "chartVersion", chart.ChartVersion)
		return nil, err
	}
	chartRefValue, err := impl.chartRefRepository.FindById(environmentProperties.ChartRefId)
	if err != nil {
		impl.logger.Errorw("error in finding ref chart by id", "err", err)
		return nil, err
	}
	if !(chartRefValue.IsAppMetricsSupported) {
		appMetricsRequest := chartService.AppMetricEnableDisableRequest{UserId: environmentProperties.UserId, AppId: appId, EnvironmentId: environmentProperties.EnvironmentId, IsAppMetricsEnabled: false}
		_, err = impl.EnvMetricsEnableDisable(&appMetricsRequest)
		if err != nil {
			impl.logger.Errorw("err while disable app metrics", "err", err, "appId", appId, "chartVersion", chart.ChartVersion)
			return nil, err
		}
	} else {
		appMetricsRequest := chartService.AppMetricEnableDisableRequest{UserId: environmentProperties.UserId, AppId: appId, EnvironmentId: environmentProperties.EnvironmentId, IsAppMetricsEnabled: appMetrics}
		_, err = impl.EnvMetricsEnableDisable(&appMetricsRequest)
		if err != nil {
			impl.logger.Errorw("err while updating app metrics", "err", err)
			return nil, err
		}
	}

	return environmentProperties, nil
}

func (impl PropertiesConfigServiceImpl) UpdateEnvironmentProperties(appId int, propertiesRequest *bean.EnvironmentProperties, userId int32) (*bean.EnvironmentProperties, error) {
	//check if exists
	oldEnvOverride, err := impl.envConfigRepo.Get(propertiesRequest.Id)
	if err != nil {
		return nil, err
	}
	overrideByte, err := propertiesRequest.EnvOverrideValues.MarshalJSON()
	if err != nil {
		return nil, err
	}
	env, err := impl.environmentRepository.FindById(oldEnvOverride.TargetEnvironment)
	if err != nil {
		return nil, err
	}
	//FIXME add check for restricted NS also like (kube-system, devtron, monitoring, etc)
	if env.Namespace != "" && env.Namespace != propertiesRequest.Namespace {
		return nil, fmt.Errorf("enviremnt is restricted to namespace: %s only, cant deploy to: %s", env.Namespace, propertiesRequest.Namespace)
	}

	if !oldEnvOverride.Latest {
		envOverrideExisting, err := impl.envConfigRepo.FindLatestChartForAppByAppIdAndEnvId(appId, oldEnvOverride.TargetEnvironment)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		if envOverrideExisting != nil {
			envOverrideExisting.Latest = false
			envOverrideExisting.IsOverride = false
			envOverrideExisting.UpdatedOn = time.Now()
			envOverrideExisting.UpdatedBy = userId
			envOverrideExisting, err = impl.envConfigRepo.Update(envOverrideExisting)
			if err != nil {
				return nil, err
			}
		}
	}

	override := &chartConfig.EnvConfigOverride{
		Active:            propertiesRequest.Active,
		Id:                propertiesRequest.Id,
		ChartId:           oldEnvOverride.ChartId,
		EnvOverrideValues: string(overrideByte),
		Status:            propertiesRequest.Status,
		ManualReviewed:    propertiesRequest.ManualReviewed,
		Namespace:         propertiesRequest.Namespace,
		TargetEnvironment: propertiesRequest.EnvironmentId,
		IsBasicViewLocked: propertiesRequest.IsBasicViewLocked,
		CurrentViewEditor: propertiesRequest.CurrentViewEditor,
		AuditLog:          sql.AuditLog{UpdatedBy: propertiesRequest.UserId, UpdatedOn: time.Now()},
	}

	override.Latest = true
	override.IsOverride = true
	impl.logger.Debugw("updating environment override ", "value", override)
	err = impl.envConfigRepo.UpdateProperties(override)

	if oldEnvOverride.Namespace != override.Namespace {
		return nil, fmt.Errorf("namespace name update not supported")
	}

	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return nil, err
	}

	isAppMetricsEnabled := false
	if propertiesRequest.AppMetrics != nil {
		isAppMetricsEnabled = *propertiesRequest.AppMetrics
	}
	chartRefValue, err := impl.chartRefRepository.FindById(oldEnvOverride.Chart.ChartRefId)
	if err != nil {
		impl.logger.Errorw("error in finding ref chart by id", "err", err)
		return nil, err
	}
	if !(chartRefValue.IsAppMetricsSupported) {
		appMetricsRequest := chartService.AppMetricEnableDisableRequest{UserId: propertiesRequest.UserId, AppId: appId, EnvironmentId: oldEnvOverride.TargetEnvironment, IsAppMetricsEnabled: false}
		_, err = impl.EnvMetricsEnableDisable(&appMetricsRequest)
		if err != nil {
			impl.logger.Errorw("err while disable app metrics for lower versions", err)
			return nil, err
		}
	} else {
		appMetricsRequest := chartService.AppMetricEnableDisableRequest{UserId: propertiesRequest.UserId, AppId: appId, EnvironmentId: oldEnvOverride.TargetEnvironment, IsAppMetricsEnabled: isAppMetricsEnabled}
		_, err = impl.EnvMetricsEnableDisable(&appMetricsRequest)
		if err != nil {
			impl.logger.Errorw("err while updating app metrics", "err", err)
			return nil, err
		}
	}

	//creating history
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(override, nil, isAppMetricsEnabled, 0)
	if err != nil {
		impl.logger.Errorw("error in creating entry for env deployment template history", "err", err, "envOverride", override)
		return nil, err
	}
	//VARIABLE_MAPPING_UPDATE
	err = impl.extractAndMapVariables(override.EnvOverrideValues, override.Id, repository5.EntityTypeDeploymentTemplateEnvLevel, override.CreatedBy)
	if err != nil {
		return nil, err
	}

	return propertiesRequest, err
}

func (impl PropertiesConfigServiceImpl) buildAppMetricsJson() ([]byte, error) {
	appMetricsEnabled := chartService.AppMetricsEnabled{
		AppMetrics: true,
	}
	appMetricsJson, err := json.Marshal(appMetricsEnabled)
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	return appMetricsJson, nil
}

func (impl PropertiesConfigServiceImpl) CreateIfRequired(chart *chartRepoRepository.Chart, environmentId int, userId int32, manualReviewed bool, chartStatus models.ChartStatus, isOverride, isAppMetricsEnabled bool, namespace string, IsBasicViewLocked bool, CurrentViewEditor models.ChartsViewEditorType, tx *pg.Tx) (*chartConfig.EnvConfigOverride, error) {
	env, err := impl.environmentRepository.FindById(environmentId)
	if err != nil {
		return nil, err
	}

	if env != nil && len(env.Namespace) > 0 {
		namespace = env.Namespace
	}

	envOverride, err := impl.envConfigRepo.GetByChartAndEnvironment(chart.Id, environmentId)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	if errors.IsNotFound(err) {
		if isOverride {
			// before creating new entry, remove previous one from latest tag
			envOverrideExisting, err := impl.envConfigRepo.FindLatestChartForAppByAppIdAndEnvId(chart.AppId, environmentId)
			if err != nil && !errors.IsNotFound(err) {
				return nil, err
			}
			if envOverrideExisting != nil {
				envOverrideExisting.Latest = false
				envOverrideExisting.UpdatedOn = time.Now()
				envOverrideExisting.UpdatedBy = userId
				envOverrideExisting.IsOverride = isOverride
				envOverrideExisting.IsBasicViewLocked = IsBasicViewLocked
				envOverrideExisting.CurrentViewEditor = CurrentViewEditor
				//maintaining backward compatibility for while
				if tx != nil {
					envOverrideExisting, err = impl.envConfigRepo.UpdateWithTxn(envOverrideExisting, tx)
				} else {
					envOverrideExisting, err = impl.envConfigRepo.Update(envOverrideExisting)
				}
				if err != nil {
					return nil, err
				}
			}
		}

		impl.logger.Debugw("env config not found creating new ", "chart", chart.Id, "env", environmentId)
		//create new
		envOverride = &chartConfig.EnvConfigOverride{
			Active:            true,
			ManualReviewed:    manualReviewed,
			Status:            chartStatus,
			TargetEnvironment: environmentId,
			ChartId:           chart.Id,
			AuditLog:          sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now(), CreatedBy: userId},
			Namespace:         namespace,
			IsOverride:        isOverride,
			IsBasicViewLocked: IsBasicViewLocked,
			CurrentViewEditor: CurrentViewEditor,
		}
		if isOverride {
			envOverride.EnvOverrideValues = chart.GlobalOverride
			envOverride.Latest = true
		} else {
			envOverride.EnvOverrideValues = "{}"
		}
		//maintaining backward compatibility for while
		if tx != nil {
			err = impl.envConfigRepo.SaveWithTxn(envOverride, tx)
		} else {
			err = impl.envConfigRepo.Save(envOverride)
		}
		if err != nil {
			impl.logger.Errorw("error in creating envconfig", "data", envOverride, "error", err)
			return nil, err
		}
		err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(envOverride, tx, isAppMetricsEnabled, 0)
		if err != nil {
			impl.logger.Errorw("error in creating entry for env deployment template history", "err", err, "envOverride", envOverride)
			return nil, err
		}

		//VARIABLE_MAPPING_UPDATE
		if envOverride.EnvOverrideValues != "{}" {
			err = impl.extractAndMapVariables(envOverride.EnvOverrideValues, envOverride.Id, repository5.EntityTypeDeploymentTemplateEnvLevel, envOverride.CreatedBy)
			if err != nil {
				return nil, err
			}
		}
	}
	return envOverride, nil
}

func (impl PropertiesConfigServiceImpl) GetEnvironmentPropertiesById(envId int) ([]bean.EnvironmentProperties, error) {

	var envProperties []bean.EnvironmentProperties
	envOverrides, err := impl.envConfigRepo.GetByEnvironment(envId)
	if err != nil {
		impl.logger.Error("error fetching override config", "err", err)
		return nil, err
	}

	for _, envOverride := range envOverrides {
		envProperties = append(envProperties, bean.EnvironmentProperties{
			Id:             envOverride.Id,
			Status:         envOverride.Status,
			ManualReviewed: envOverride.ManualReviewed,
			Active:         envOverride.Active,
			Namespace:      envOverride.Namespace,
		})
	}

	return envProperties, nil
}

func (impl PropertiesConfigServiceImpl) GetAppIdByChartEnvId(chartEnvId int) (*chartConfig.EnvConfigOverride, error) {

	envOverride, err := impl.envConfigRepo.Get(chartEnvId)
	if err != nil {
		impl.logger.Error("error fetching override config", "err", err)
		return nil, err
	}

	return envOverride, nil
}

func (impl PropertiesConfigServiceImpl) FindEnvLevelAppMetricsByAppIdAndEnvId(appId int, envId int) (*repository.EnvLevelAppMetrics, error) {
	return impl.envLevelAppMetricsRepository.FindByAppIdAndEnvId(appId, envId)
}

func (impl PropertiesConfigServiceImpl) GetLatestEnvironmentProperties(appId, environmentId int) (environmentProperties *bean.EnvironmentProperties, err error) {
	env, err := impl.environmentRepository.FindById(environmentId)
	if err != nil {
		return nil, err
	}
	// step 1
	envOverride, err := impl.envConfigRepo.ActiveEnvConfigOverride(appId, environmentId)
	if err != nil {
		return nil, err
	}
	if envOverride.Id == 0 {
		//return nil, errors.New("No env config exists with tag latest for given appId and envId")
		impl.logger.Warnw("No env config exists with tag latest for given appId and envId", "envId", environmentId)
	} else {
		r := json.RawMessage{}
		err = r.UnmarshalJSON([]byte(envOverride.EnvOverrideValues))
		if err != nil {
			return nil, err
		}

		environmentProperties = &bean.EnvironmentProperties{
			Id:                envOverride.Id,
			EnvOverrideValues: r,
			Status:            envOverride.Status,
			ManualReviewed:    envOverride.ManualReviewed,
			Active:            envOverride.Active,
			Namespace:         env.Namespace,
			EnvironmentId:     environmentId,
			EnvironmentName:   env.Name,
			Latest:            envOverride.Latest,
			IsOverride:        envOverride.IsOverride,
			IsBasicViewLocked: envOverride.IsBasicViewLocked,
			CurrentViewEditor: envOverride.CurrentViewEditor,
			ChartRefId:        envOverride.Chart.ChartRefId,
			ClusterId:         env.ClusterId,
		}
	}

	return environmentProperties, nil
}

func (impl PropertiesConfigServiceImpl) ResetEnvironmentProperties(id int) (bool, error) {
	envOverride, err := impl.envConfigRepo.Get(id)
	if err != nil {
		return false, err
	}
	envOverride.EnvOverrideValues = "{}"
	envOverride.IsOverride = false
	envOverride.Latest = false
	impl.logger.Infow("reset environment override ", "value", envOverride)
	err = impl.envConfigRepo.UpdateProperties(envOverride)
	if err != nil {
		impl.logger.Warnw("error in update envOverride", "envOverrideId", id)
	}
	envLevelAppMetrics, err := impl.envLevelAppMetricsRepository.FindByAppIdAndEnvId(envOverride.Chart.AppId, envOverride.TargetEnvironment)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetching env level app metric", "err", err)
		return false, err
	}
	if envLevelAppMetrics.Id > 0 {
		err = impl.envLevelAppMetricsRepository.Delete(envLevelAppMetrics)
		if err != nil {
			impl.logger.Errorw("error while deletion of app metric at env level", "err", err)
			return false, err
		}
	}
	//VARIABLES
	err = impl.removeMappedVariables(envOverride.Id, repository5.EntityTypeDeploymentTemplateEnvLevel, envOverride.UpdatedBy)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (impl PropertiesConfigServiceImpl) CreateEnvironmentPropertiesWithNamespace(appId int, environmentProperties *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error) {
	chart, err := impl.chartRepo.FindChartByAppIdAndRefId(appId, environmentProperties.ChartRefId)
	if err != nil && pg.ErrNoRows != err {
		return nil, err
	}
	if pg.ErrNoRows == err {
		impl.logger.Warnw("no chart found this ref id", "refId", environmentProperties.ChartRefId)
		chart, err = impl.chartRepo.FindLatestChartForAppByAppId(appId)
		if err != nil && pg.ErrNoRows != err {
			return nil, err
		}
	}

	var envOverride *chartConfig.EnvConfigOverride
	if environmentProperties.Id == 0 {
		chart.GlobalOverride = "{}"
		appMetrics := false
		if environmentProperties.AppMetrics != nil {
			appMetrics = *environmentProperties.AppMetrics
		}
		envOverride, err = impl.CreateIfRequired(chart, environmentProperties.EnvironmentId, environmentProperties.UserId, environmentProperties.ManualReviewed, models.CHARTSTATUS_SUCCESS, false, appMetrics, environmentProperties.Namespace, environmentProperties.IsBasicViewLocked, environmentProperties.CurrentViewEditor, nil)
		if err != nil {
			return nil, err
		}
	} else {
		envOverride, err = impl.envConfigRepo.Get(environmentProperties.Id)
		if err != nil {
			impl.logger.Errorw("error in fetching envOverride", "err", err)
		}
		envOverride.Namespace = environmentProperties.Namespace
		envOverride.UpdatedBy = environmentProperties.UserId
		envOverride.IsBasicViewLocked = environmentProperties.IsBasicViewLocked
		envOverride.CurrentViewEditor = environmentProperties.CurrentViewEditor
		envOverride.UpdatedOn = time.Now()
		impl.logger.Debugw("updating environment override ", "value", envOverride)
		err = impl.envConfigRepo.UpdateProperties(envOverride)
	}

	r := json.RawMessage{}
	err = r.UnmarshalJSON([]byte(envOverride.EnvOverrideValues))
	if err != nil {
		return nil, err
	}
	env, err := impl.environmentRepository.FindById(environmentProperties.EnvironmentId)
	if err != nil {
		return nil, err
	}
	environmentProperties = &bean.EnvironmentProperties{
		Id:                envOverride.Id,
		Status:            envOverride.Status,
		EnvOverrideValues: r,
		ManualReviewed:    envOverride.ManualReviewed,
		Active:            envOverride.Active,
		Namespace:         env.Namespace,
		EnvironmentId:     environmentProperties.EnvironmentId,
		EnvironmentName:   env.Name,
		Latest:            envOverride.Latest,
		ChartRefId:        environmentProperties.ChartRefId,
		IsOverride:        envOverride.IsOverride,
		ClusterId:         env.ClusterId,
	}
	return environmentProperties, nil
}

//below method is deprecated

func (impl PropertiesConfigServiceImpl) EnvMetricsEnableDisable(appMetricRequest *chartService.AppMetricEnableDisableRequest) (*chartService.AppMetricEnableDisableRequest, error) {
	// validate app metrics compatibility
	var currentChart *chartConfig.EnvConfigOverride
	var err error
	currentChart, err = impl.envConfigRepo.FindLatestChartForAppByAppIdAndEnvId(appMetricRequest.AppId, appMetricRequest.EnvironmentId)
	if err != nil && !errors.IsNotFound(err) {
		impl.logger.Error(err)
		return nil, err
	}
	if errors.IsNotFound(err) {
		impl.logger.Errorw("no chart configured for this app", "appId", appMetricRequest.AppId)
		err = &util.ApiError{
			InternalMessage: "no chart configured for this app",
			UserMessage:     "no chart configured for this app",
		}
		return nil, err
	}
	if appMetricRequest.IsAppMetricsEnabled == true {
		chartRefValue, err := impl.chartRefRepository.FindById(currentChart.Chart.ChartRefId)
		if err != nil {
			impl.logger.Errorw("error in finding ref chart by id", "err", err)
			return nil, err
		}
		if !(chartRefValue.IsAppMetricsSupported) {
			err = &util.ApiError{
				InternalMessage: "chart version in not compatible for app metrics",
				UserMessage:     "chart version in not compatible for app metrics",
			}
			return nil, err
		}
	}
	// update and create env level app metrics
	envLevelAppMetrics, err := impl.envLevelAppMetricsRepository.FindByAppIdAndEnvId(appMetricRequest.AppId, appMetricRequest.EnvironmentId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error("err", err)
		return nil, err
	}
	if envLevelAppMetrics == nil || envLevelAppMetrics.Id == 0 {
		infraMetrics := true
		envLevelAppMetrics = &repository.EnvLevelAppMetrics{
			AppId:        appMetricRequest.AppId,
			EnvId:        appMetricRequest.EnvironmentId,
			AppMetrics:   &appMetricRequest.IsAppMetricsEnabled,
			InfraMetrics: &infraMetrics,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				CreatedBy: appMetricRequest.UserId,
				UpdatedBy: appMetricRequest.UserId,
			},
		}
		err = impl.envLevelAppMetricsRepository.Save(envLevelAppMetrics)
		if err != nil {
			impl.logger.Error("err", err)
			return nil, err
		}
	} else {
		envLevelAppMetrics.AppMetrics = &appMetricRequest.IsAppMetricsEnabled
		envLevelAppMetrics.UpdatedOn = time.Now()
		envLevelAppMetrics.UpdatedBy = appMetricRequest.UserId
		err = impl.envLevelAppMetricsRepository.Update(envLevelAppMetrics)
		if err != nil {
			impl.logger.Error("err", err)
			return nil, err
		}
	}
	//updating audit log details of chart as history service uses it
	currentChart.UpdatedOn = time.Now()
	currentChart.UpdatedBy = appMetricRequest.UserId
	//creating history entry
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(currentChart, nil, appMetricRequest.IsAppMetricsEnabled, 0)
	if err != nil {
		impl.logger.Errorw("error in creating entry for env deployment template history", "err", err, "envOverride", currentChart)
		return nil, err
	}
	//VARIABLE_MAPPING_UPDATE - not needed?
	//err = impl.extractAndMapVariables(override.EnvOverrideValues, override.Id, repository5.EntityTypeDeploymentTemplateEnvLevel, override.CreatedBy)
	//if err != nil {
	//	return nil, err
	//}

	return appMetricRequest, err
}

func (impl PropertiesConfigServiceImpl) extractAndMapVariables(template string, entityId int, entityType repository5.EntityType, userId int32) error {
	usedVariables, err := impl.variableTemplateParser.ExtractVariables(template)
	if err != nil {
		return err
	}
	err = impl.variableEntityMappingService.UpdateVariablesForEntity(usedVariables, repository5.Entity{
		EntityType: entityType,
		EntityId:   entityId,
	}, userId)
	if err != nil {
		return err
	}
	return nil
}

func (impl PropertiesConfigServiceImpl) removeMappedVariables(entityId int, entityType repository5.EntityType, userId int32) error {

	err := impl.variableEntityMappingService.DeleteMappingsForEntities([]repository5.Entity{{
		EntityType: entityType,
		EntityId:   entityId,
	}}, userId)
	if err != nil {
		return err
	}
	return nil
}
