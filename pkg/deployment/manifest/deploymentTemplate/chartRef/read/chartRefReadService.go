/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package read

import (
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	"go.uber.org/zap"
)

type ChartRefReadService interface {
	FindById(chartRefId int) (*bean.ChartRefDto, error)
	FindByVersionAndName(version, name string) (*bean.ChartRefDto, error)
}

type ChartRefReadServiceImpl struct {
	logger             *zap.SugaredLogger
	chartRefRepository chartRepoRepository.ChartRefRepository
}

func NewChartRefReadServiceImpl(logger *zap.SugaredLogger,
	chartRefRepository chartRepoRepository.ChartRefRepository) *ChartRefReadServiceImpl {
	return &ChartRefReadServiceImpl{
		logger:             logger,
		chartRefRepository: chartRefRepository,
	}
}

func (impl *ChartRefReadServiceImpl) FindById(chartRefId int) (*bean.ChartRefDto, error) {
	chartRef, err := impl.chartRefRepository.FindById(chartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting chartRef by id", "err", err, "chartRefId", chartRefId)
		return nil, err
	}
	return adapter.ConvertChartRefDbObjToBean(chartRef), nil
}

func (impl *ChartRefReadServiceImpl) FindByVersionAndName(version, name string) (*bean.ChartRefDto, error) {
	chartRef, err := impl.chartRefRepository.FindByVersionAndName(name, version)
	if err != nil {
		impl.logger.Errorw("error in getting chartRef by version and name", "err", err, "version", version, "name", name)
		return nil, err
	}
	return adapter.ConvertChartRefDbObjToBean(chartRef), nil
}
