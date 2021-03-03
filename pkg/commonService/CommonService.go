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
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/team"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CommonService interface {
	FetchLatestChart(appId int, envId int) (*chartConfig.Chart, error)
	GlobalChecklist() (*GlobalChecklist, error)
}

type CommonServiceImpl struct {
	logger                      *zap.SugaredLogger
	chartRepository             chartConfig.ChartRepository
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository
	gitOpsRepository            repository.GitOpsConfigRepository
	dockerReg                   repository.DockerArtifactStoreRepository
	attributeRepo               repository.AttributesRepository
	gitProviderRepository       repository.GitProviderRepository
	environmentRepository       cluster.EnvironmentRepository
	teamRepository              team.TeamRepository
	appRepository               pipelineConfig.AppRepository
}

func NewCommonServiceImpl(logger *zap.SugaredLogger,
	chartRepository chartConfig.ChartRepository,
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository,
	gitOpsRepository repository.GitOpsConfigRepository,
	dockerReg repository.DockerArtifactStoreRepository,
	attributeRepo repository.AttributesRepository,
	gitProviderRepository repository.GitProviderRepository,
	environmentRepository cluster.EnvironmentRepository, teamRepository team.TeamRepository,
	appRepository pipelineConfig.AppRepository) *CommonServiceImpl {
	serviceImpl := &CommonServiceImpl{
		logger:                      logger,
		chartRepository:             chartRepository,
		environmentConfigRepository: environmentConfigRepository,
		gitOpsRepository:            gitOpsRepository,
		dockerReg:                   dockerReg,
		attributeRepo:               attributeRepo,
		gitProviderRepository:       gitProviderRepository,
		environmentRepository:       environmentRepository,
		teamRepository:              teamRepository,
		appRepository:               appRepository,
	}
	return serviceImpl
}

type GlobalChecklist struct {
	AppChecklist   *AppChecklist   `json:"appChecklist"`
	ChartChecklist *ChartChecklist `json:"chartChecklist"`
	IsAppCreated   bool            `json:"isAppCreated"`
	UserId         int32           `json:"-"`
}

type ChartChecklist struct {
	GitOps      int `json:"gitOps"`
	Project     int `json:"project"`
	Environment int `json:"environment"`
}

type AppChecklist struct {
	GitOps      int `json:"gitOps"`
	Project     int `json:"project"`
	Git         int `json:"git"`
	Environment int `json:"environment"`
	Docker      int `json:"docker"`
	HostUrl     int `json:"hostUrl"`
	//ChartChecklist *ChartChecklist `json:",inline"`
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

func (impl *CommonServiceImpl) GlobalChecklist() (*GlobalChecklist, error) {
	gitOps, err := impl.gitOpsRepository.GetGitOpsConfigActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	dockerReg, err := impl.dockerReg.FindActiveDefaultStore()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	attribute, err := impl.attributeRepo.FindByKey(attributes.HostUrlKey)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	env, err := impl.environmentRepository.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	git, err := impl.gitProviderRepository.FindAllActiveForAutocomplete()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	project, err := impl.teamRepository.FindAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	chartChecklist := &ChartChecklist{
		Project:     1,
		Environment: 1,
	}
	appChecklist := &AppChecklist{
		Project:     1,
		Git:         1,
		Environment: 1,
	}
	if len(env) > 0 {
		chartChecklist.Environment = 1
		appChecklist.Environment = 1
	}

	if len(git) > 0 {
		appChecklist.Git = 1
	}

	if len(project) > 0 {
		chartChecklist.Project = 1
		appChecklist.Project = 1
	}

	if gitOps.Id > 0 {
		chartChecklist.GitOps = 1
		appChecklist.GitOps = 1
	}
	if len(dockerReg.Id) > 0 {
		appChecklist.Docker = 1
	}
	if attribute.Id > 0 {
		appChecklist.HostUrl = 1
	}
	config := &GlobalChecklist{
		AppChecklist:   appChecklist,
		ChartChecklist: chartChecklist,
	}

	apps, err := impl.appRepository.FindAllActiveAppsWithTeam()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}
	if len(apps) > 0 {
		config.IsAppCreated = true
	}
	return config, err
}
