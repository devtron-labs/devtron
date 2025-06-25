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
	errors2 "errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	bean5 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	bean3 "github.com/devtron-labs/devtron/pkg/chart/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/adapter"
	bean4 "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/read"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/variables"
	repository5 "github.com/devtron-labs/devtron/pkg/variables/repository"
	globalUtil "github.com/devtron-labs/devtron/util"
	"go.opentelemetry.io/otel"
	"net/http"
	"time"

	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/sql"

	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.uber.org/zap"
)

type PropertiesConfigService interface {
	CreateEnvironmentPropertiesAndBaseIfNeeded(ctx context.Context, appId int, environmentProperties *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error)
	CreateEnvironmentProperties(appId int, propertiesRequest *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error)
	UpdateEnvironmentProperties(appId int, propertiesRequest *bean.EnvironmentProperties, userId int32) (*bean.EnvironmentProperties, error)
	//create environment entry for each new environment
	CreateIfRequired(request *bean.EnvironmentOverrideCreateInternalDTO, tx *pg.Tx) (*bean4.EnvConfigOverride, bool, error)
	GetEnvironmentProperties(appId, environmentId int, chartRefId int) (environmentPropertiesResponse *bean.EnvironmentPropertiesResponse, err error)
	GetEnvironmentPropertiesById(environmentId int) ([]bean.EnvironmentProperties, error)

	GetAppIdByChartEnvId(chartEnvId int) (*bean4.EnvConfigOverride, error)
	GetLatestEnvironmentProperties(appId, environmentId int) (*bean.EnvironmentProperties, error)
	ResetEnvironmentProperties(id int, userId int32) (bool, error)
	CreateEnvironmentPropertiesWithNamespace(appId int, propertiesRequest *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error)

	FetchEnvProperties(appId, envId, chartRefId int) (*bean4.EnvConfigOverride, error)
	ChangeChartRefForEnvConfigOverride(ctx context.Context, request *bean3.ChartRefChangeRequest, userId int32) (*bean.EnvironmentProperties, error)

	PropertiesConfigServiceEnt
}
type PropertiesConfigServiceImpl struct {
	logger                           *zap.SugaredLogger
	envConfigRepo                    chartConfig.EnvConfigOverrideRepository
	chartRepo                        chartRepoRepository.ChartRepository
	environmentRepository            repository.EnvironmentRepository
	deploymentTemplateHistoryService deploymentTemplate.DeploymentTemplateHistoryService
	scopedVariableManager            variables.ScopedVariableManager
	deployedAppMetricsService        deployedAppMetrics.DeployedAppMetricsService
	envConfigOverrideReadService     read.EnvConfigOverrideService
	deploymentConfigService          common.DeploymentConfigService
	chartService                     chartService.ChartService
}

func NewPropertiesConfigServiceImpl(logger *zap.SugaredLogger,
	envConfigRepo chartConfig.EnvConfigOverrideRepository,
	chartRepo chartRepoRepository.ChartRepository,
	environmentRepository repository.EnvironmentRepository,
	deploymentTemplateHistoryService deploymentTemplate.DeploymentTemplateHistoryService,
	scopedVariableManager variables.ScopedVariableManager,
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService,
	envConfigOverrideReadService read.EnvConfigOverrideService,
	deploymentConfigService common.DeploymentConfigService,
	chartService chartService.ChartService) *PropertiesConfigServiceImpl {
	return &PropertiesConfigServiceImpl{
		logger:                           logger,
		envConfigRepo:                    envConfigRepo,
		chartRepo:                        chartRepo,
		environmentRepository:            environmentRepository,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
		scopedVariableManager:            scopedVariableManager,
		deployedAppMetricsService:        deployedAppMetricsService,
		envConfigOverrideReadService:     envConfigOverrideReadService,
		deploymentConfigService:          deploymentConfigService,
		chartService:                     chartService,
	}

}

func (impl *PropertiesConfigServiceImpl) GetEnvironmentProperties(appId, environmentId int, chartRefId int) (environmentPropertiesResponse *bean.EnvironmentPropertiesResponse, err error) {
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
	chart, err := impl.chartRepo.FindLatestChartForAppByAppId(nil, appId)
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

	externalReleaseType, err := impl.deploymentConfigService.GetExternalReleaseType(appId, environmentId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config by appId and envId", "appId", appId, "envId", environmentId, "err", err)
		return nil, err
	}
	if len(externalReleaseType) != 0 {
		environmentPropertiesResponse.EnvironmentConfig.MigratedFrom = &externalReleaseType
	}
	return environmentPropertiesResponse, nil
}

func (impl *PropertiesConfigServiceImpl) FetchEnvProperties(appId, envId, chartRefId int) (*bean4.EnvConfigOverride, error) {
	return impl.envConfigOverrideReadService.GetByAppIdEnvIdAndChartRefId(appId, envId, chartRefId)
}

func (impl *PropertiesConfigServiceImpl) CreateEnvironmentPropertiesAndBaseIfNeeded(ctx context.Context, appId int, environmentProperties *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error) {
	createResp, err := impl.CreateEnvironmentProperties(appId, environmentProperties)
	if err != nil {
		if err.Error() == bean5.NOCHARTEXIST {
			appMetrics := false
			if environmentProperties.AppMetrics != nil {
				appMetrics = *environmentProperties.AppMetrics
			}
			templateRequest := bean3.TemplateRequest{
				AppId:               appId,
				ChartRefId:          environmentProperties.ChartRefId,
				ValuesOverride:      globalUtil.GetEmptyJSON(),
				UserId:              environmentProperties.UserId,
				IsAppMetricsEnabled: appMetrics,
			}
			_, err = impl.chartService.CreateChartFromEnvOverride(ctx, templateRequest)
			if err != nil {
				impl.logger.Errorw("service err, EnvConfigOverrideCreate", "err", err, "payload", environmentProperties)
				return nil, err
			}
			createResp, err = impl.CreateEnvironmentProperties(appId, environmentProperties)
			if err != nil {
				impl.logger.Errorw("service err, EnvConfigOverrideCreate", "err", err, "payload", environmentProperties)
				return nil, err
			}
		} else {
			impl.logger.Errorw("service err, EnvConfigOverrideCreate", "err", err, "payload", environmentProperties)
			return nil, err
		}
	}
	return createResp, nil
}

func (impl *PropertiesConfigServiceImpl) CreateEnvironmentProperties(appId int, environmentProperties *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error) {
	chart, err := impl.chartRepo.FindChartByAppIdAndRefId(appId, environmentProperties.ChartRefId)
	if err != nil && !errors2.Is(err, pg.ErrNoRows) {
		return nil, err
	} else if errors2.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("create new chart set latest=false", "a", "b")
		return nil, fmt.Errorf("NOCHARTEXIST")
	}

	externalReleaseType, err := impl.deploymentConfigService.GetExternalReleaseType(chart.AppId, environmentProperties.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config by appId and envId", "appId", chart.AppId, "envId", environmentProperties.EnvironmentId, "err", err)
		return nil, err
	}
	if externalReleaseType.IsArgoApplication() {
		return nil, util.NewApiError(http.StatusConflict,
			"chart version change is not allowed for external argo application",
			"chart version change is not allowed for external argo application")
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
	dbConnection := impl.envConfigRepo.GetDbConnection()

	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in beginning db tx", "err", err)
		return nil, err
	}

	envOverride, appMetrics, err := impl.CreateIfRequired(overrideCreateRequest, tx)
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
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in commiting tx", "err", err)
		return nil, err
	}

	return environmentProperties, nil
}

func (impl *PropertiesConfigServiceImpl) UpdateEnvironmentProperties(appId int, propertiesRequest *bean.EnvironmentProperties, userId int32) (*bean.EnvironmentProperties, error) {
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

	tx, err := impl.envConfigRepo.GetDbConnection().Begin()
	if err != nil {
		impl.logger.Errorw("error in beginning db tx", "err", err)
		return nil, err
	}

	if !oldEnvOverride.Latest {
		envOverrideExisting, err := impl.envConfigOverrideReadService.FindLatestChartForAppByAppIdAndEnvId(nil, appId, oldEnvOverride.TargetEnvironment)
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		if envOverrideExisting != nil {
			envOverrideExisting.Latest = false
			envOverrideExisting.IsOverride = false
			envOverrideExisting.UpdatedOn = time.Now()
			envOverrideExisting.UpdatedBy = userId
			envOverrideExistingDBObj := adapter.EnvOverrideDTOToDB(envOverrideExisting)
			envOverrideExistingDBObj, err = impl.envConfigRepo.Update(tx, envOverrideExistingDBObj)
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
		MergeStrategy:     propertiesRequest.MergeStrategy,
		AuditLog:          sql.AuditLog{UpdatedBy: propertiesRequest.UserId, UpdatedOn: time.Now()},
	}

	overrideDbObj.Latest = true
	overrideDbObj.IsOverride = true
	impl.logger.Debugw("updating environment override ", "value", overrideDbObj)
	err = impl.envConfigRepo.UpdateProperties(tx, overrideDbObj)

	if oldEnvOverride.Namespace != overrideDbObj.Namespace {
		return nil, fmt.Errorf("namespace name update not supported")
	}

	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return nil, err
	}

	chart, err := impl.chartRepo.FindById(overrideDbObj.ChartId)
	if err != nil {
		impl.logger.Errorw("error in chartRefRepository.FindById", "chartRefId", chart.ChartRefId, "err", err)
		return nil, err
	}
	err = impl.deploymentConfigService.UpdateChartLocationInDeploymentConfig(tx, appId, overrideDbObj.TargetEnvironment, chart.ChartRefId, userId, chart.ChartVersion)
	if err != nil {
		impl.logger.Errorw("error in UpdateChartLocationInDeploymentConfig", "appId", appId, "envId", overrideDbObj.TargetEnvironment, "err", err)
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
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(overrideOverrideDTO, tx, isAppMetricsEnabled, 0)
	if err != nil {
		impl.logger.Errorw("error in creating entry for env deployment template history", "err", err, "envOverride", overrideOverrideDTO)
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	//VARIABLE_MAPPING_UPDATE
	err = impl.scopedVariableManager.ExtractAndMapVariables(overrideOverrideDTO.EnvOverrideValues, overrideDbObj.Id, repository5.EntityTypeDeploymentTemplateEnvLevel, overrideOverrideDTO.CreatedBy, nil)
	if err != nil {
		return nil, err
	}

	return propertiesRequest, err
}

func (impl *PropertiesConfigServiceImpl) CreateIfRequired(request *bean.EnvironmentOverrideCreateInternalDTO, tx *pg.Tx) (*bean4.EnvConfigOverride, bool, error) {

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
			envOverrideExisting, err := impl.envConfigOverrideReadService.FindLatestChartForAppByAppIdAndEnvId(nil, chart.AppId, environmentId)
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
					envOverrideDBObj, err = impl.envConfigRepo.Update(nil, envOverrideDBObj)
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
			MergeStrategy:     request.MergeStrategy,
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
		envOverrideDBObj.Chart = chart
		envOverride = adapter.EnvOverrideDBToDTO(envOverrideDBObj)

		err = impl.deploymentConfigService.UpdateChartLocationInDeploymentConfig(tx, chart.AppId, envOverride.TargetEnvironment, chart.ChartRefId, userId, envOverride.Chart.ChartVersion)
		if err != nil {
			impl.logger.Errorw("error in UpdateChartLocationInDeploymentConfig", "appId", chart.AppId, "envId", envOverride.TargetEnvironment, "err", err)
			return nil, isAppMetricsEnabled, err
		}

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

func (impl *PropertiesConfigServiceImpl) GetEnvironmentPropertiesById(envId int) ([]bean.EnvironmentProperties, error) {

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

func (impl *PropertiesConfigServiceImpl) GetAppIdByChartEnvId(chartEnvId int) (*bean4.EnvConfigOverride, error) {
	envOverride, err := impl.envConfigOverrideReadService.GetByIdIncludingInactive(chartEnvId)
	if err != nil {
		impl.logger.Error("error fetching override config", "err", err)
		return nil, err
	}
	return envOverride, nil
}

func (impl *PropertiesConfigServiceImpl) GetLatestEnvironmentProperties(appId, environmentId int) (environmentProperties *bean.EnvironmentProperties, err error) {
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
		r := json.RawMessage("{}")
		err = r.UnmarshalJSON([]byte(envOverride.EnvOverrideValues))
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
			Description:       env.Description,
			EnvironmentId:     environmentId,
			EnvironmentName:   env.Name,
			Latest:            envOverride.Latest,
			ChartRefId:        envOverride.Chart.ChartRefId,
			IsOverride:        envOverride.IsOverride,
			IsBasicViewLocked: envOverride.IsBasicViewLocked,
			CurrentViewEditor: envOverride.CurrentViewEditor,
			MergeStrategy:     envOverride.MergeStrategy,
			ClusterId:         env.ClusterId,
		}
	}
	return environmentProperties, nil
}

func (impl *PropertiesConfigServiceImpl) ResetEnvironmentProperties(id int, userId int32) (bool, error) {
	envOverride, err := impl.envConfigOverrideReadService.GetByIdIncludingInactive(id)
	if err != nil {
		return false, err
	}
	envOverride.EnvOverrideValues = "{}"
	envOverride.IsOverride = false
	envOverride.Latest = false
	impl.logger.Infow("reset environment override ", "value", envOverride)

	tx, err := impl.environmentRepository.GetConnection().Begin()
	if err != nil {
		impl.logger.Errorw("error in beginning db tx", "err", err)
		return false, err
	}

	envOverrideDBObj := adapter.EnvOverrideDTOToDB(envOverride)
	err = impl.envConfigRepo.UpdateProperties(tx, envOverrideDBObj)
	if err != nil {
		impl.logger.Warnw("error in update envOverride", "envOverrideId", id)
	}
	err = impl.deployedAppMetricsService.DeleteEnvLevelMetricsIfPresent(envOverride.Chart.AppId, envOverride.TargetEnvironment)
	if err != nil {
		impl.logger.Errorw("error, DeleteEnvLevelMetricsIfPresent", "err", err, "appId", envOverride.Chart.AppId, "envId", envOverride.TargetEnvironment)
		return false, err
	}

	chart, err := impl.chartRepo.FindLatestChartForAppByAppId(nil, envOverride.Chart.AppId)
	if err != nil {
		impl.logger.Errorw("error in chartRefRepository.FindById", "chartRefId", envOverride.Chart.ChartRefId, "err", err)
		return false, err
	}
	err = impl.deploymentConfigService.UpdateChartLocationInDeploymentConfig(tx, envOverride.Chart.AppId, envOverride.TargetEnvironment, chart.ChartRefId, userId, chart.ChartVersion)
	if err != nil {
		impl.logger.Errorw("error in UpdateChartLocationInDeploymentConfig", "appId", envOverride.Chart.AppId, "envId", envOverride.TargetEnvironment, "err", err)
		return false, err
	}

	//VARIABLES
	err = impl.scopedVariableManager.RemoveMappedVariables(envOverride.Id, repository5.EntityTypeDeploymentTemplateEnvLevel, envOverride.UpdatedBy, tx)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (impl *PropertiesConfigServiceImpl) CreateEnvironmentPropertiesWithNamespace(appId int, environmentProperties *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error) {
	chart, err := impl.chartRepo.FindChartByAppIdAndRefId(appId, environmentProperties.ChartRefId)
	if err != nil && !errors2.Is(err, pg.ErrNoRows) {
		return nil, err
	}
	if errors2.Is(err, pg.ErrNoRows) {
		impl.logger.Warnw("no chart found this ref id", "refId", environmentProperties.ChartRefId)
		chart, err = impl.chartRepo.FindLatestChartForAppByAppId(nil, appId)
		if err != nil && !errors2.Is(err, pg.ErrNoRows) {
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
		err = impl.envConfigRepo.UpdateProperties(nil, adapter.EnvOverrideDTOToDB(envOverride))
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

func (impl *PropertiesConfigServiceImpl) ChangeChartRefForEnvConfigOverride(ctx context.Context, request *bean3.ChartRefChangeRequest, userId int32) (*bean.EnvironmentProperties, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "PropertiesConfigServiceImpl.ChangeChartRefForEnvConfigOverride")
	defer span.End()
	envConfigPropertiesOld, err := impl.FetchEnvProperties(request.AppId, request.EnvId, request.TargetChartRefId)
	if err != nil && !errors2.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("service err, ChangeChartRef", "err", err, "payload", request)
		return nil, fmt.Errorf("could not fetch env properties. error: %v", err)
	} else if errors2.Is(err, pg.ErrNoRows) {
		createResp, err := impl.createEnvConfigOverrideWithChart(newCtx, request, userId)
		if err != nil {
			impl.logger.Errorw("service err, ChangeChartRef", "err", err, "payload", request)
			return nil, err
		}
		return createResp, nil
	}
	envConfigProperties := request.EnvConfigProperties
	envConfigProperties.Id = envConfigPropertiesOld.Id
	createResp, err := impl.UpdateEnvironmentProperties(request.AppId, envConfigProperties, userId)
	if err != nil {
		impl.logger.Errorw("service err, EnvConfigOverrideUpdate", "err", err, "payload", envConfigProperties)
		return nil, fmt.Errorf("could not update env override, error: %v", err)
	}
	return createResp, nil
}

func (impl *PropertiesConfigServiceImpl) createEnvConfigOverrideWithChart(ctx context.Context, request *bean3.ChartRefChangeRequest, userId int32) (*bean.EnvironmentProperties, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "PropertiesConfigServiceImpl.createEnvConfigOverrideWithChart")
	defer span.End()
	createResp, err := impl.CreateEnvironmentProperties(request.AppId, request.EnvConfigProperties)
	if err != nil && err.Error() != bean5.NOCHARTEXIST {
		impl.logger.Errorw("service err, EnvConfigOverrideCreate", "err", err, "payload", request)
		return nil, fmt.Errorf("could not create env override, error: %v", err)
	} else if err != nil && err.Error() == bean5.NOCHARTEXIST {
		appMetrics := false
		if request.EnvConfigProperties.AppMetrics != nil {
			appMetrics = request.EnvMetrics
		}
		templateRequest := bean3.TemplateRequest{
			AppId:               request.AppId,
			ChartRefId:          request.TargetChartRefId,
			ValuesOverride:      globalUtil.GetEmptyJSON(),
			UserId:              userId,
			IsAppMetricsEnabled: appMetrics,
		}
		_, err := impl.chartService.CreateChartFromEnvOverride(newCtx, templateRequest)
		if err != nil {
			impl.logger.Errorw("service err, CreateChartFromEnvOverride", "err", err, "payload", request)
			return nil, fmt.Errorf("could not create chart from env override, error: %v", err)
		}
		return impl.CreateEnvironmentProperties(request.AppId, request.EnvConfigProperties)
	}
	return createResp, nil
}
