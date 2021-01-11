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

package rbac

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/team"
	"go.uber.org/zap"
	"strings"
)

type EnforcerUtil interface {
	GetAppRBACName(appName string) string
	GetRbacObjectsForAllApps() map[int]string
	GetAppRBACNameByAppId(appId int) string
	GetAppRBACByAppNameAndEnvId(appName string, envId int) string
	GetAppRBACByAppIdAndPipelineId(appId int, pipelineId int) string
	GetEnvRBACNameByAppId(appId int, envId int) string
	GetTeamRBACByCiPipelineId(pipelineId int) string
	GetEnvRBACArrayByAppId(appId int) []string
	GetEnvRBACNameByCiPipelineIdAndEnvId(ciPipelineId int, envId int) string
	GetTeamRbacObjectByCiPipelineId(ciPipelineId int) string
	GetTeamAndEnvironmentRbacObjectByCDPipelineId(pipelineId int) (string, string)
}
type EnforcerUtilImpl struct {
	logger                *zap.SugaredLogger
	teamRepository        team.TeamRepository
	appRepo               pipelineConfig.AppRepository
	environmentRepository cluster.EnvironmentRepository
	pipelineRepository    pipelineConfig.PipelineRepository
	ciPipelineRepository  pipelineConfig.CiPipelineRepository
}

func NewEnforcerUtilImpl(logger *zap.SugaredLogger, teamRepository team.TeamRepository,
	appRepo pipelineConfig.AppRepository, environmentRepository cluster.EnvironmentRepository,
	pipelineRepository pipelineConfig.PipelineRepository, ciPipelineRepository pipelineConfig.CiPipelineRepository) *EnforcerUtilImpl {
	return &EnforcerUtilImpl{
		logger:                logger,
		teamRepository:        teamRepository,
		appRepo:               appRepo,
		environmentRepository: environmentRepository,
		pipelineRepository:    pipelineRepository,
		ciPipelineRepository:  ciPipelineRepository,
	}
}

func (impl EnforcerUtilImpl) GetAppRBACName(appName string) string {
	team, err := impl.teamRepository.FindTeamByAppName(appName)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(appName))
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(team.Name), strings.ToLower(appName))
}

func (impl EnforcerUtilImpl) GetRbacObjectsForAllApps() map[int]string {
	objects := make(map[int]string)
	result, err := impl.appRepo.FindAllActiveAppsWithTeam()
	if err != nil {
		return objects
	}
	for _, item := range result {
		if _, ok := objects[item.Id]; !ok {
			objects[item.Id] = fmt.Sprintf("%s/%s", strings.ToLower(item.Team.Name), strings.ToLower(item.AppName))
		}
	}
	return objects
}

func (impl EnforcerUtilImpl) GetAppRBACNameByAppId(appId int) string {
	team, err := impl.teamRepository.FindTeamByAppId(appId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	apps, err := impl.appRepo.FindAppsByTeamId(team.Id)
	if err != nil {
		return fmt.Sprintf("%s/%s", strings.ToLower(team.Name), "")
	}
	var appName = ""
	for _, item := range apps {
		if item.Id == appId {
			appName = item.AppName
		}
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(team.Name), strings.ToLower(appName))
}

func (impl EnforcerUtilImpl) GetAppRBACByAppNameAndEnvId(appName string, envId int) string {
	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(appName))
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(env.Name), strings.ToLower(appName))
}

func (impl EnforcerUtilImpl) GetAppRBACByAppIdAndPipelineId(appId int, pipelineId int) string {
	team, err := impl.teamRepository.FindTeamByAppId(appId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	apps, err := impl.appRepo.FindAppsByTeamId(team.Id)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	var appName = ""
	for _, item := range apps {
		if item.Id == appId {
			appName = item.AppName
		}
	}
	pipeline, err := impl.pipelineRepository.FindById(pipelineId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(appName))
	}
	env, err := impl.environmentRepository.FindById(pipeline.EnvironmentId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(appName))
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(env.Name), strings.ToLower(appName))
}

func (impl EnforcerUtilImpl) GetEnvRBACNameByAppId(appId int, envId int) string {
	team, err := impl.teamRepository.FindTeamByAppId(appId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	apps, err := impl.appRepo.FindAppsByTeamId(team.Id)
	if err != nil {
		return fmt.Sprintf("%s/%s", strings.ToLower(team.Name), "")
	}
	var appName = ""
	for _, item := range apps {
		if item.Id == appId {
			appName = item.AppName
		}
	}
	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(appName))
	}

	return fmt.Sprintf("%s/%s", strings.ToLower(env.Name), strings.ToLower(appName))
}

func (impl EnforcerUtilImpl) GetTeamRBACByCiPipelineId(pipelineId int) string {
	ciPipeline, err := impl.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		impl.logger.Error(err)
		return ""
	}
	return impl.GetAppRBACNameByAppId(ciPipeline.AppId)
}

func (impl EnforcerUtilImpl) GetEnvRBACArrayByAppId(appId int) []string {
	var rbacObjects []string

	pipelines, err := impl.pipelineRepository.FindActiveByAppId(appId)
	if err != nil {
		impl.logger.Error(err)
		return rbacObjects
	}
	for _, item := range pipelines {
		rbacObjects = append(rbacObjects, impl.GetAppRBACByAppIdAndPipelineId(appId, item.Id))
	}

	return rbacObjects
}

func (impl EnforcerUtilImpl) GetEnvRBACNameByCiPipelineIdAndEnvId(ciPipelineId int, envId int) string {
	ciPipeline, err := impl.ciPipelineRepository.FindById(ciPipelineId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	team, err := impl.teamRepository.FindTeamByAppId(ciPipeline.AppId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	apps, err := impl.appRepo.FindAppsByTeamId(team.Id)
	if err != nil {
		return fmt.Sprintf("%s/%s", strings.ToLower(team.Name), "")
	}
	var appName = ""
	for _, item := range apps {
		if item.Id == ciPipeline.AppId {
			appName = item.AppName
		}
	}
	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(appName))
	}

	return fmt.Sprintf("%s/%s", strings.ToLower(env.Name), strings.ToLower(appName))
}

func (impl EnforcerUtilImpl) GetTeamRbacObjectByCiPipelineId(ciPipelineId int) string {
	ciPipeline, err := impl.ciPipelineRepository.FindById(ciPipelineId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	team, err := impl.teamRepository.FindTeamByAppId(ciPipeline.AppId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(team.Name), strings.ToLower(ciPipeline.App.AppName))
}

func (impl EnforcerUtilImpl) GetTeamAndEnvironmentRbacObjectByCDPipelineId(pipelineId int) (string, string) {
	pipeline, err := impl.pipelineRepository.FindById(pipelineId)
	if err != nil {
		impl.logger.Error(err)
		return "", ""
	}
	team, err := impl.teamRepository.FindTeamByAppId(pipeline.AppId)
	if err != nil {
		return "", ""
	}
	teamRbac := fmt.Sprintf("%s/%s", strings.ToLower(team.Name), strings.ToLower(pipeline.App.AppName))
	envRbac := fmt.Sprintf("%s/%s", strings.ToLower(pipeline.Environment.Name), strings.ToLower(pipeline.App.AppName))
	return teamRbac, envRbac
}
