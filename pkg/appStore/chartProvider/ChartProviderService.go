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
	"github.com/devtron-labs/devtron/pkg/chartRepo"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type ChartProviderService interface {
	GetChartProviderList() ([]*ChartProviderResponseDto, error)
	ToggleChartProvider(request *ChartProviderRequestDto) error
	SyncChartProvider(request *ChartProviderRequestDto) error
}

type ChartProviderServiceImpl struct {
	logger                      *zap.SugaredLogger
	repoRepository              chartRepoRepository.ChartRepoRepository
	chartRepositoryService      chartRepo.ChartRepositoryService
	registryRepository          dockerRegistryRepository.DockerArtifactStoreRepository
	ociRegistryConfigRepository dockerRegistryRepository.OCIRegistryConfigRepository
}

func NewChartProviderServiceImpl(logger *zap.SugaredLogger, repoRepository chartRepoRepository.ChartRepoRepository, chartRepositoryService chartRepo.ChartRepositoryService) *ChartProviderServiceImpl {
	return &ChartProviderServiceImpl{
		logger:                 logger,
		repoRepository:         repoRepository,
		chartRepositoryService: chartRepositoryService,
	}
}

func UpdateChartRepoList(models []*chartRepoRepository.ChartRepoWithDeploymentCount, chartProviders []*ChartProviderResponseDto) []*ChartProviderResponseDto {
	for _, model := range models {
		chartRepo := &ChartProviderResponseDto{}
		chartRepo.Id = strconv.Itoa(model.Id)
		chartRepo.Name = model.Name
		chartRepo.Active = model.Active
		chartRepo.IsEditable = true
		chartRepo.IsOCIRegistry = false
		if model.ActiveDeploymentCount > 0 {
			chartRepo.IsEditable = false
		}
		chartProviders = append(chartProviders, chartRepo)
	}
	return chartProviders
}

func (impl *ChartProviderServiceImpl) GetChartProviderList() ([]*ChartProviderResponseDto, error) {
	var chartProviders []*ChartProviderResponseDto
	store, err := impl.registryRepository.FindAllChartProviders()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in listing artifact", "err", err)
		return nil, err
	}
	for _, model := range store {
		if !util.IsOCIRegistryChartProvider(model) {
			continue
		}
		chartRepo := &ChartProviderResponseDto{}
		chartRepo.Id = model.Id
		chartRepo.Name = model.Id
		chartRepo.IsEditable = true
		chartRepo.Active = model.OCIRegistryConfig[0].IsChartPullActive
		chartRepo.IsOCIRegistry = true
		chartRepo.RegistryProvider = model.RegistryType
		chartProviders = append(chartProviders, chartRepo)
	}

	models, err := impl.repoRepository.FindAllWithDeploymentCount()
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	return UpdateChartRepoList(models, chartProviders), nil
}

func (impl *ChartProviderServiceImpl) ToggleChartProvider(request *ChartProviderRequestDto) error {
	dbConnection := impl.repoRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	if !request.IsOCIRegistry {
		chartRepo, err := impl.repoRepository.FindById(request.ChartRepoId)
		if err != nil {
			return err
		}
		chartRepo.Active = request.Active
		chartRepo.UpdatedBy = request.UserId
		chartRepo.UpdatedOn = time.Now()
		err = impl.repoRepository.Update(chartRepo, tx)
		if err != nil {
			return err
		}
	} else {
		ociRegistryConfigs, err := impl.ociRegistryConfigRepository.FindByDockerRegistryId(request.Id)
		if err != nil {
			impl.logger.Errorw("find OCI config service err, ToggleChartProvider", "err", err, "DockerArtifactStoreId", request.Id)
			return err
		}
		found := false
		for _, ociRegistryConfig := range ociRegistryConfigs {
			if util.IsOCIConfigChartProvider(ociRegistryConfig) {
				ociRegistryConfig.IsChartPullActive = request.Active
				ociRegistryConfig.UpdatedBy = request.UserId
				ociRegistryConfig.UpdatedOn = time.Now()
				found = true
				err = impl.ociRegistryConfigRepository.Update(ociRegistryConfig, tx)
				if err != nil {
					impl.logger.Errorw("update OCI config service err, ToggleChartProvider", "err", err, "DockerArtifactStoreId", request.Id)
					return err
				}
				break
			}
		}
		if !found {
			impl.logger.Errorw("no OCI config found for update, ToggleChartProvider", "err", pg.ErrNoRows, "DockerArtifactStoreId", request.Id)
			return pg.ErrNoRows
		}
	}

	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in db transaction commit, ToggleChartProvider", "err", err)
		return err
	}
	return nil
}

func (impl *ChartProviderServiceImpl) SyncChartProvider(request *ChartProviderRequestDto) error {
	chartProviderConfig := &chartRepo.ChartProviderConfig{
		ChartProviderId: request.Id,
		IsOCIRegistry:   request.IsOCIRegistry,
	}
	err := impl.chartRepositoryService.TriggerChartSyncManual(chartProviderConfig)
	if err != nil {
		impl.logger.Errorw("error creating chart sync job, SyncChartProvider", "err", err)
		return err
	}
	return nil
}
