/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/adapter"
	bean4 "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/read"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/variables"
	repository5 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"time"

	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/sql"

	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.uber.org/zap"
)

type PropertiesConfigService interface {
	CreateEnvironmentProperties(appId int, propertiesRequest *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error)
	UpdateEnvironmentProperties(appId int, propertiesRequest *bean.EnvironmentProperties, userId int32) (*bean.EnvironmentProperties, error)
	//create environment entry for each new environment
	CreateIfRequired(request *bean.EnvironmentOverrideCreateInternalDTO, tx *pg.Tx) (*bean4.EnvConfigOverride, bool, error)
	GetEnvironmentProperties(appId, environmentId int, chartRefId int) (environmentPropertiesResponse *bean.EnvironmentPropertiesResponse, err error)
	GetEnvironmentPropertiesById(environmentId int) ([]bean.EnvironmentProperties, error)

	GetAppIdByChartEnvId(chartEnvId int) (*bean4.EnvConfigOverride, error)
	GetLatestEnvironmentProperties(appId, environmentId int) (*bean.EnvironmentProperties, error)
	ResetEnvironmentProperties(id int) (bool, error)
	CreateEnvironmentPropertiesWithNamespace(appId int, propertiesRequest *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error)

	FetchEnvProperties(appId, envId, chartRefId int) (*bean4.EnvConfigOverride, error)
}
type PropertiesConfigServiceImpl struct {
	logger                           *zap.SugaredLogger
	envConfigRepo                    chartConfig.EnvConfigOverrideRepository
	chartRepo                        chartRepoRepository.ChartRepository
	environmentRepository            repository2.EnvironmentRepository
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService
	scopedVariableManager            variables.ScopedVariableManager
	deployedAppMetricsService        deployedAppMetrics.DeployedAppMetricsService
	envConfigOverrideReadService     read.EnvConfigOverrideService
}

func NewPropertiesConfigServiceImpl(logger *zap.SugaredLogger,
	envConfigRepo chartConfig.EnvConfigOverrideRepository,
	chartRepo chartRepoRepository.ChartRepository,
	environmentRepository repository2.EnvironmentRepository,
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService,
	scopedVariableManager variables.ScopedVariableManager,
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService,
	envConfigOverrideReadService read.EnvConfigOverrideService) *PropertiesConfigServiceImpl {
	return &PropertiesConfigServiceImpl{
		logger:                           logger,
		envConfigRepo:                    envConfigRepo,
		chartRepo:                        chartRepo,
		environmentRepository:            environmentRepository,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
		scopedVariableManager:            scopedVariableManager,
		deployedAppMetricsService:        deployedAppMetricsService,
		envConfigOverrideReadService:     envConfigOverrideReadService,
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
	envOverride, err := impl.envConfigOverrideReadService.ActiveEnvConfigOverride(appId, environmentId)
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
		if chartRefId == 0 && envOverride.Chart != nil {
			environmentProperties.ChartRefId = envOverride.Chart.ChartRefId
		}

		if environmentPropertiesResponse.Namespace == "" {
			environmentPropertiesResponse.Namespace = envOverride.Namespace
		}
	}
	ecOverride, err := impl.envConfigOverrideReadService.FindChartByAppIdAndEnvIdAndChartRefId(appId, environmentId, chartRefId)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	if errors.IsNotFound(err) {
		environmentProperties.Id = 0
		environmentProperties.IsOverride = false
		if chartRefId > 0 {
			environmentProperties.ChartRefId = chartRefId
		}
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
		if chartRefId == 0 && ecOverride.Chart != nil {
			environmentProperties.ChartRefId = ecOverride.Chart.ChartRefId
		}
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
	isAppMetricsEnabled, err := impl.deployedAppMetricsService.GetMetricsFlagForAPipelineByAppIdAndEnvId(appId, environmentId)
	if err != nil {
		impl.logger.Errorw("error, GetMetricsFlagForAPipelineByAppIdAndEnvId", "err", err, "appId", appId, "envId", environmentId)
		return nil, err
	}
	environmentPropertiesResponse.AppMetrics = &isAppMetricsEnabled
	return environmentPropertiesResponse, nil
}

func (impl PropertiesConfigServiceImpl) FetchEnvProperties(appId, envId, chartRefId int) (*bean4.EnvConfigOverride, error) {
	return impl.envConfigOverrideReadService.GetByAppIdEnvIdAndChartRefId(appId, envId, chartRefId)
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
	overrideCreateRequest := &bean.EnvironmentOverrideCreateInternalDTO{
		Chart:               chart,
		EnvironmentId:       environmentProperties.EnvironmentId,
		UserId:              environmentProperties.UserId,
		ManualReviewed:      environmentProperties.ManualReviewed,
		ChartStatus:         models.CHARTSTATUS_SUCCESS,
		IsOverride:          true,
		IsAppMetricsEnabled: appMetrics,
		IsBasicViewLocked:   environmentProperties.IsBasicViewLocked,
		Namespace:           environmentProperties.Namespace,
		CurrentViewEditor:   environmentProperties.CurrentViewEditor,
		MergeStrategy:       environmentProperties.MergeStrategy,
	}
	envOverride, appMetrics, err := impl.CreateIfRequired(overrideCreateRequest, nil)
	if err != nil {
		return nil, err
	}
	environmentProperties.AppMetrics = &appMetrics
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

	return environmentProperties, nil
}

func (impl PropertiesConfigServiceImpl) UpdateEnvironmentProperties(appId int, propertiesRequest *bean.EnvironmentProperties, userId int32) (*bean.EnvironmentProperties, error) {
	//check if exists
	oldEnvOverride, err := impl.envConfigOverrideReadService.GetByIdIncludingInactive(propertiesRequest.Id)
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
		envOverrideExisting, err := impl.envConfigOverrideReadService.FindLatestChartForAppByAppIdAndEnvId(appId, oldEnvOverride.TargetEnvironment)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		if envOverrideExisting != nil {
			envOverrideExisting.Latest = false
			envOverrideExisting.IsOverride = false
			envOverrideExisting.UpdatedOn = time.Now()
			envOverrideExisting.UpdatedBy = userId
			envOverrideExistingDBObj := adapter.EnvOverrideDTOToDB(envOverrideExisting)
			envOverrideExistingDBObj, err = impl.envConfigRepo.Update(envOverrideExistingDBObj)
			if err != nil {
				return nil, err
			}
		}
	}

	overrideDbObj := &chartConfig.EnvConfigOverride{
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

	overrideDbObj.Latest = true
	overrideDbObj.IsOverride = true
	impl.logger.Debugw("updating environment override ", "value", overrideDbObj)
	err = impl.envConfigRepo.UpdateProperties(overrideDbObj)

	if oldEnvOverride.Namespace != overrideDbObj.Namespace {
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
	envLevelMetricsUpdateReq := &bean2.DeployedAppMetricsRequest{
		EnableMetrics: isAppMetricsEnabled,
		AppId:         appId,
		EnvId:         oldEnvOverride.TargetEnvironment,
		ChartRefId:    oldEnvOverride.Chart.ChartRefId,
		UserId:        propertiesRequest.UserId,
	}
	err = impl.deployedAppMetricsService.CreateOrUpdateAppOrEnvLevelMetrics(context.Background(), envLevelMetricsUpdateReq)
	if err != nil {
		impl.logger.Errorw("error, CheckAndUpdateAppOrEnvLevelMetrics", "err", err, "req", envLevelMetricsUpdateReq)
		return nil, err
	}

	overrideOverrideDTO := adapter.EnvOverrideDBToDTO(overrideDbObj)
	//creating history
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(overrideOverrideDTO, nil, isAppMetricsEnabled, 0)
	if err != nil {
		impl.logger.Errorw("error in creating entry for env deployment template history", "err", err, "envOverride", overrideOverrideDTO)
		return nil, err
	}
	//VARIABLE_MAPPING_UPDATE
	err = impl.scopedVariableManager.ExtractAndMapVariables(overrideOverrideDTO.EnvOverrideValues, overrideDbObj.Id, repository5.EntityTypeDeploymentTemplateEnvLevel, overrideOverrideDTO.CreatedBy, nil)
	if err != nil {
		return nil, err
	}

	return propertiesRequest, err
}

func (impl PropertiesConfigServiceImpl) CreateIfRequired(request *bean.EnvironmentOverrideCreateInternalDTO, tx *pg.Tx) (*bean4.EnvConfigOverride, bool, error) {

	chart := request.Chart
	environmentId := request.EnvironmentId
	userId := request.UserId
	manualReviewed := request.ManualReviewed
	chartStatus := request.ChartStatus
	isOverride := request.IsOverride
	isAppMetricsEnabled := request.IsAppMetricsEnabled
	IsBasicViewLocked := request.IsBasicViewLocked
	namespace := request.Namespace
	CurrentViewEditor := request.CurrentViewEditor

	env, err := impl.environmentRepository.FindById(environmentId)
	if err != nil {
		return nil, request.IsAppMetricsEnabled, err
	}

	if env != nil && len(env.Namespace) > 0 {
		namespace = env.Namespace
	}

	if isOverride { //case of override, to do app metrics operation
		envLevelMetricsUpdateReq := &bean2.DeployedAppMetricsRequest{
			EnableMetrics: isAppMetricsEnabled,
			AppId:         chart.AppId,
			EnvId:         environmentId,
			ChartRefId:    chart.ChartRefId,
			UserId:        userId,
		}
		err = impl.deployedAppMetricsService.CreateOrUpdateAppOrEnvLevelMetrics(context.Background(), envLevelMetricsUpdateReq)
		if err != nil {
			impl.logger.Errorw("error, CheckAndUpdateAppOrEnvLevelMetrics", "err", err, "req", envLevelMetricsUpdateReq)
			return nil, isAppMetricsEnabled, err
		}
		//updating metrics flag because it might be possible that the chartRef used was not supported and that could have override the metrics flag got in request
		isAppMetricsEnabled = envLevelMetricsUpdateReq.EnableMetrics
	}

	envOverride, err := impl.envConfigOverrideReadService.GetByChartAndEnvironment(chart.Id, environmentId)
	if err != nil && !errors.IsNotFound(err) {
		return nil, isAppMetricsEnabled, err
	}
	if errors.IsNotFound(err) {
		if isOverride {
			// before creating new entry, remove previous one from latest tag
			envOverrideExisting, err := impl.envConfigOverrideReadService.FindLatestChartForAppByAppIdAndEnvId(chart.AppId, environmentId)
			if err != nil && !errors.IsNotFound(err) {
				return nil, isAppMetricsEnabled, err
			}
			if envOverrideExisting != nil {
				envOverrideExisting.Latest = false
				envOverrideExisting.UpdatedOn = time.Now()
				envOverrideExisting.UpdatedBy = userId
				envOverrideExisting.IsOverride = isOverride
				envOverrideExisting.IsBasicViewLocked = IsBasicViewLocked
				envOverrideExisting.CurrentViewEditor = CurrentViewEditor
				//maintaining backward compatibility for while
				envOverrideDBObj := adapter.EnvOverrideDTOToDB(envOverrideExisting)
				if tx != nil {
					envOverrideDBObj, err = impl.envConfigRepo.UpdateWithTxn(envOverrideDBObj, tx)
				} else {
					envOverrideDBObj, err = impl.envConfigRepo.Update(envOverrideDBObj)
				}
				if err != nil {
					return nil, isAppMetricsEnabled, err
				}
			}
		}

		impl.logger.Debugw("env config not found creating new ", "chart", chart.Id, "env", environmentId)
		//create new
		envOverrideDBObj := &chartConfig.EnvConfigOverride{
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
			envOverrideDBObj.EnvOverrideValues = chart.GlobalOverride
			envOverrideDBObj.Latest = true
		} else {
			envOverrideDBObj.EnvOverrideValues = "{}"
		}
		//maintaining backward compatibility for while
		if tx != nil {
			err = impl.envConfigRepo.SaveWithTxn(envOverrideDBObj, tx)
		} else {
			err = impl.envConfigRepo.Save(envOverrideDBObj)
		}
		if err != nil {
			impl.logger.Errorw("error in creating envconfig", "data", envOverride, "error", err)
			return nil, isAppMetricsEnabled, err
		}
		envOverride = adapter.EnvOverrideDBToDTO(envOverrideDBObj)
		err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(envOverride, tx, isAppMetricsEnabled, 0)
		if err != nil {
			impl.logger.Errorw("error in creating entry for env deployment template history", "err", err, "envOverride", envOverride)
			return nil, isAppMetricsEnabled, err
		}

		//VARIABLE_MAPPING_UPDATE
		if envOverride.EnvOverrideValues != "{}" {
			err = impl.scopedVariableManager.ExtractAndMapVariables(envOverride.EnvOverrideValues, envOverride.Id, repository5.EntityTypeDeploymentTemplateEnvLevel, envOverride.CreatedBy, tx)
			if err != nil {
				return nil, isAppMetricsEnabled, err
			}
		}
	}
	return envOverride, isAppMetricsEnabled, nil
}

func (impl PropertiesConfigServiceImpl) GetEnvironmentPropertiesById(envId int) ([]bean.EnvironmentProperties, error) {

	var envProperties []bean.EnvironmentProperties
	envOverrides, err := impl.envConfigOverrideReadService.GetByEnvironment(envId)
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

func (impl PropertiesConfigServiceImpl) GetAppIdByChartEnvId(chartEnvId int) (*bean4.EnvConfigOverride, error) {
	envOverride, err := impl.envConfigOverrideReadService.GetByIdIncludingInactive(chartEnvId)
	if err != nil {
		impl.logger.Error("error fetching override config", "err", err)
		return nil, err
	}
	return envOverride, nil
}

func (impl PropertiesConfigServiceImpl) GetLatestEnvironmentProperties(appId, environmentId int) (environmentProperties *bean.EnvironmentProperties, err error) {
	env, err := impl.environmentRepository.FindById(environmentId)
	if err != nil {
		return nil, err
	}
	// step 1
	envOverride, err := impl.envConfigOverrideReadService.ActiveEnvConfigOverride(appId, environmentId)
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
	envOverride, err := impl.envConfigOverrideReadService.GetByIdIncludingInactive(id)
	if err != nil {
		return false, err
	}
	envOverride.EnvOverrideValues = "{}"
	envOverride.IsOverride = false
	envOverride.Latest = false
	impl.logger.Infow("reset environment override ", "value", envOverride)

	envOverrideDBObj := adapter.EnvOverrideDTOToDB(envOverride)
	err = impl.envConfigRepo.UpdateProperties(envOverrideDBObj)
	if err != nil {
		impl.logger.Warnw("error in update envOverride", "envOverrideId", id)
	}
	err = impl.deployedAppMetricsService.DeleteEnvLevelMetricsIfPresent(envOverride.Chart.AppId, envOverride.TargetEnvironment)
	if err != nil {
		impl.logger.Errorw("error, DeleteEnvLevelMetricsIfPresent", "err", err, "appId", envOverride.Chart.AppId, "envId", envOverride.TargetEnvironment)
		return false, err
	}
	//VARIABLES
	err = impl.scopedVariableManager.RemoveMappedVariables(envOverride.Id, repository5.EntityTypeDeploymentTemplateEnvLevel, envOverride.UpdatedBy, nil)
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

	var envOverride *bean4.EnvConfigOverride
	if environmentProperties.Id == 0 {
		chart.GlobalOverride = "{}"
		appMetrics := false
		if environmentProperties.AppMetrics != nil {
			appMetrics = *environmentProperties.AppMetrics
		}
		overrideCreateRequest := &bean.EnvironmentOverrideCreateInternalDTO{
			Chart:               chart,
			EnvironmentId:       environmentProperties.EnvironmentId,
			UserId:              environmentProperties.UserId,
			ManualReviewed:      environmentProperties.ManualReviewed,
			ChartStatus:         models.CHARTSTATUS_SUCCESS,
			IsOverride:          false,
			IsAppMetricsEnabled: appMetrics,
			IsBasicViewLocked:   environmentProperties.IsBasicViewLocked,
			Namespace:           environmentProperties.Namespace,
			CurrentViewEditor:   environmentProperties.CurrentViewEditor,
			MergeStrategy:       environmentProperties.MergeStrategy,
		}
		envOverride, appMetrics, err = impl.CreateIfRequired(overrideCreateRequest, nil)
		if err != nil {
			return nil, err
		}
		environmentProperties.AppMetrics = &appMetrics
	} else {
		envOverride, err = impl.envConfigOverrideReadService.GetByIdIncludingInactive(environmentProperties.Id)
		if err != nil {
			impl.logger.Errorw("error in fetching envOverride", "err", err)
		}
		envOverride.Namespace = environmentProperties.Namespace
		envOverride.UpdatedBy = environmentProperties.UserId
		envOverride.IsBasicViewLocked = environmentProperties.IsBasicViewLocked
		envOverride.CurrentViewEditor = environmentProperties.CurrentViewEditor
		envOverride.UpdatedOn = time.Now()
		impl.logger.Debugw("updating environment override ", "value", envOverride)
		err = impl.envConfigRepo.UpdateProperties(adapter.EnvOverrideDTOToDB(envOverride))
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
