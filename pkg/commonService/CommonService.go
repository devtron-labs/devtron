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

package commonService

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	helper2 "github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/attributes/bean"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/read"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	read3 "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/read"
	read2 "github.com/devtron-labs/devtron/pkg/team/read"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CommonService interface {
	FetchLatestChartVersion(appId int, envId int) (string, error)
	GlobalChecklist() (*GlobalChecklist, error)
	EnvironmentVariableList() (*EnvironmentVariableList, error)
}

type CommonServiceImpl struct {
	logger                       *zap.SugaredLogger
	chartRepository              chartRepoRepository.ChartRepository
	environmentConfigRepository  chartConfig.EnvConfigOverrideRepository
	dockerReg                    dockerRegistryRepository.DockerArtifactStoreRepository
	attributeRepo                repository.AttributesRepository
	gitProviderReadService       read.GitProviderReadService
	environmentRepository        repository3.EnvironmentRepository
	appRepository                app.AppRepository
	gitOpsConfigReadService      config.GitOpsConfigReadService
	commonBaseServiceImpl        *CommonBaseServiceImpl
	envConfigOverrideReadService read3.EnvConfigOverrideService
	teamReadService              read2.TeamReadService
}

func NewCommonServiceImpl(logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository,
	dockerReg dockerRegistryRepository.DockerArtifactStoreRepository,
	attributeRepo repository.AttributesRepository,
	environmentRepository repository3.EnvironmentRepository,
	appRepository app.AppRepository,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	gitProviderReadService read.GitProviderReadService,
	envConfigOverrideReadService read3.EnvConfigOverrideService,
	commonBaseServiceImpl *CommonBaseServiceImpl,
	teamReadService read2.TeamReadService) *CommonServiceImpl {
	serviceImpl := &CommonServiceImpl{
		logger:                       logger,
		chartRepository:              chartRepository,
		environmentConfigRepository:  environmentConfigRepository,
		dockerReg:                    dockerReg,
		attributeRepo:                attributeRepo,
		environmentRepository:        environmentRepository,
		appRepository:                appRepository,
		gitOpsConfigReadService:      gitOpsConfigReadService,
		gitProviderReadService:       gitProviderReadService,
		commonBaseServiceImpl:        commonBaseServiceImpl,
		envConfigOverrideReadService: envConfigOverrideReadService,
		teamReadService:              teamReadService,
	}
	return serviceImpl
}

func (impl *CommonServiceImpl) FetchLatestChartVersion(appId int, envId int) (string, error) {
	var chart *chartRepoRepository.Chart
	if appId > 0 && envId > 0 {
		envOverride, err := impl.envConfigOverrideReadService.ActiveEnvConfigOverride(appId, envId)
		if err != nil {
			return "", err
		}
		//if chart is overrides in env, and not mark as overrides in db, it means it was not completed and refer to latest to the app.
		if (envOverride.Id == 0) || (envOverride.Id > 0 && !envOverride.IsOverride) {
			chart, err = impl.chartRepository.FindLatestChartForAppByAppId(appId)
			if err != nil {
				return "", err
			}
		} else {
			//if chart is overrides in env, it means it may have different version than app level.
			chart = envOverride.Chart
		}
	} else if appId > 0 {
		chartG, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
		if err != nil {
			return "", err
		}
		chart = chartG

		//TODO - note if secret create/update from global with property (new style).
		// there may be older chart version in env overrides (and in that case it will be ignore, property and isBinary)
	}
	if chart == nil {
		return "", nil
	}
	return chart.ChartVersion, nil
}

func (impl *CommonServiceImpl) GlobalChecklist() (*GlobalChecklist, error) {

	dockerReg, err := impl.dockerReg.FindActiveDefaultStore()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	attribute, err := impl.attributeRepo.FindByKey(bean.HostUrlKey)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	env, err := impl.environmentRepository.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	git, err := impl.gitProviderReadService.GetAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}

	project, err := impl.teamReadService.FindAllActive()
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

	apps, err := impl.appRepository.FindAllActiveAppsWithTeam(helper2.CustomApp)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GlobalChecklist, error while getting error", "err", err)
		return nil, err
	}
	if len(apps) > 0 {
		config.IsAppCreated = true
	}
	return config, err
}

func (impl *CommonServiceImpl) EnvironmentVariableList() (*EnvironmentVariableList, error) {
	return impl.commonBaseServiceImpl.EnvironmentVariableList()
}
