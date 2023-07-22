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

package chartProvider

import (
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/util"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"go.uber.org/zap"
	"strconv"
)

type ChartProviderService interface {
	GetChartProviderList() ([]*ChartProviderResponseDto, error)
	ToggleChartProvider(request *ChartProviderToggleRequestDto) (*ChartProviderResponseDto, error)
	SyncChartProvider(request *ChartProviderSyncRequestDto) error
}

type ChartProviderServiceImpl struct {
	logger             *zap.SugaredLogger
	repoRepository     chartRepoRepository.ChartRepoRepository
	registryRepository dockerRegistryRepository.DockerArtifactStoreRepository
}

func NewChartProviderServiceImpl(logger *zap.SugaredLogger, repoRepository chartRepoRepository.ChartRepoRepository, registryRepository dockerRegistryRepository.DockerArtifactStoreRepository) *ChartProviderServiceImpl {
	return &ChartProviderServiceImpl{
		logger:             logger,
		repoRepository:     repoRepository,
		registryRepository: registryRepository,
	}
}

func (impl *ChartProviderServiceImpl) GetChartProviderList() ([]*ChartProviderResponseDto, error) {
	var chartProviders []*ChartProviderResponseDto
	store, err := impl.registryRepository.FindAllChartProviders()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in listing artifact", "err", err)
		return nil, err
	}
	for _, model := range store {
		chartRepo := &ChartProviderResponseDto{}
		chartRepo.Id = model.Id
		chartRepo.Name = model.Id
		chartRepo.Active = model.Active
		chartRepo.IsEditable = true
		chartProviders = append(chartProviders, chartRepo)
	}

	models, err := impl.repoRepository.FindAllWithDeploymentCount()
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	for _, model := range models {
		chartRepo := &ChartProviderResponseDto{}
		chartRepo.Id = strconv.Itoa(model.Id)
		chartRepo.Name = model.Name
		chartRepo.Active = model.Active
		chartRepo.IsEditable = true
		if model.ActiveDeploymentCount > 0 {
			chartRepo.IsEditable = false
		}
		chartProviders = append(chartProviders, chartRepo)
	}
	return chartProviders, nil
}

func (impl *ChartProviderServiceImpl) ToggleChartProvider(request *ChartProviderToggleRequestDto) (*ChartProviderResponseDto, error) {
	return &ChartProviderResponseDto{}, nil
}

func (impl *ChartProviderServiceImpl) SyncChartProvider(request *ChartProviderSyncRequestDto) error {
	return nil
}
