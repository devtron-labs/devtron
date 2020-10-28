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

package commonService

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"go.uber.org/zap"
)

type CommonService interface {
	FetchLatestChart(appId int, envId int) (*chartConfig.Chart, error)
}

type CommonServiceImpl struct {
	logger                      *zap.SugaredLogger
	chartRepository             chartConfig.ChartRepository
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository
}

func NewCommonServiceImpl(logger *zap.SugaredLogger,
	chartRepository chartConfig.ChartRepository,
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository) *CommonServiceImpl {
	serviceImpl := &CommonServiceImpl{
		logger:                      logger,
		chartRepository:             chartRepository,
		environmentConfigRepository: environmentConfigRepository,
	}
	return serviceImpl
}

func (impl *CommonServiceImpl) FetchLatestChart(appId int, envId int) (*chartConfig.Chart, error) {
	var chart *chartConfig.Chart
	if appId > 0 && envId > 0 {
		envOverride, err := impl.environmentConfigRepository.ActiveEnvConfigOverride(appId, envId)
		if err != nil {
			return nil, err
		}
		//if chart is overrides in env, and not mark as overrides in db, it means it was not completed and refer to latest to the app.
		if (envOverride.Id == 0) || (envOverride.Id > 0 && !envOverride.IsOverride) {
			chart, err = impl.chartRepository.FindLatestChartForAppByAppId(appId)
			if err != nil {
				return nil, err
			}
		} else {
			//if chart is overrides in env, it means it may have different version than app level.
			chart = envOverride.Chart
		}
	} else if appId > 0 {
		chartG, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
		if err != nil {
			return nil, err
		}
		chart = chartG

		//TODO - note if secret create/update from global with property (new style).
		// there may be older chart version in env overrides (and in that case it will be ignore, property and isBinary)
	}
	return chart, nil
}
