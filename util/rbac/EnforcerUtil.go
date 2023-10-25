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
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/util"

	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
)

type EnforcerUtil interface {
	GetAppRBACName(appName string) string
	GetRbacObjectsForAllApps() map[int]string
	GetRbacObjectsForAllAppsWithTeamID(teamID int) map[int]string
	GetAppRBACNameByAppId(appId int) string
	GetAppRBACByAppNameAndEnvId(appName string, envId int) string
	GetAppRBACByAppIdAndPipelineId(appId int, pipelineId int) string
	GetTeamEnvRBACNameByAppId(appId int, envId int) string
	GetEnvRBACNameByAppId(appId int, envId int) string
	GetTeamRBACByCiPipelineId(pipelineId int) string
	GetEnvRBACArrayByAppId(appId int) []string
	GetEnvRBACNameByCiPipelineIdAndEnvId(ciPipelineId int, envId int) string
	GetTeamRbacObjectByCiPipelineId(ciPipelineId int) string
	GetTeamAndEnvironmentRbacObjectByCDPipelineId(pipelineId int) (string, string)
	GetRbacObjectsForAllAppsAndEnvironments() (map[int]string, map[string]string)
	GetProjectAdminRBACNameBYAppName(appName string) string
	GetHelmObject(appId int, envId int) (string, string)
	GetHelmObjectByAppNameAndEnvId(appName string, envId int) (string, string)
	GetHelmObjectByProjectIdAndEnvId(teamId int, envId int) (string, string)
	GetEnvRBACNameByCdPipelineIdAndEnvId(cdPipelineId int) string
	GetAppRBACNameByTeamIdAndAppId(teamId int, appId int) string
	GetRBACNameForClusterEntity(clusterName string, resourceIdentifier k8s.ResourceIdentifier) (resourceName, objectName string)
	GetAppObjectByCiPipelineIds(ciPipelineIds []int) map[int]string
	GetAppAndEnvObjectByPipelineIds(cdPipelineIds []int) map[int][]string
	GetRbacObjectsForAllAppsWithMatchingAppName(appNameMatch string) map[int]string
	GetAppAndEnvObjectByPipeline(cdPipelines []*bean.CDPipelineConfigObject) map[int][]string
	GetAppAndEnvObjectByDbPipeline(cdPipelines []*pipelineConfig.Pipeline) map[int][]string
	GetRbacObjectsByAppIds(appIds []int) map[int]string
	GetAllActiveTeamNames() ([]string, error)
	GetRbacObjectsByEnvIdsAndAppId(envIds []int, appId int) (map[int]string, map[string]string)
	GetAppRBACNameByAppAndProjectName(projectName, appName string) string
}

type EnforcerUtilImpl struct {
	logger                *zap.SugaredLogger
	teamRepository        team.TeamRepository
	appRepo               app.AppRepository
	environmentRepository repository.EnvironmentRepository
	pipelineRepository    pipelineConfig.PipelineRepository
	ciPipelineRepository  pipelineConfig.CiPipelineRepository
	clusterRepository     repository.ClusterRepository
	*EnforcerUtilHelmImpl
}

func NewEnforcerUtilImpl(logger *zap.SugaredLogger, teamRepository team.TeamRepository,
	appRepo app.AppRepository, environmentRepository repository.EnvironmentRepository,
	pipelineRepository pipelineConfig.PipelineRepository, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	clusterRepository repository.ClusterRepository) *EnforcerUtilImpl {
	return &EnforcerUtilImpl{
		logger:                logger,
		teamRepository:        teamRepository,
		appRepo:               appRepo,
		environmentRepository: environmentRepository,
		pipelineRepository:    pipelineRepository,
		ciPipelineRepository:  ciPipelineRepository,
		clusterRepository:     clusterRepository,
		EnforcerUtilHelmImpl: &EnforcerUtilHelmImpl{
			logger:            logger,
			clusterRepository: clusterRepository,
		},
	}
}

func (impl EnforcerUtilImpl) GetRbacObjectsByEnvIdsAndAppId(envIds []int, appId int) (map[int]string, map[string]string) {

	objects := make(map[int]string)
	envObjectToName := make(map[string]string)
	application, err := impl.appRepo.FindById(appId)
	if err != nil {
		impl.logger.Errorw("error occurred in fetching appa", "appId", appId)
		return objects, envObjectToName
	}

	var appName = application.AppName
	envs, err := impl.environmentRepository.FindByIds(util.GetReferencedArray(envIds))
	if err != nil {
		impl.logger.Errorw("error occurred in fetching environments", "envIds", envIds)
		return objects, envObjectToName
	}

	for _, env := range envs {
		if _, ok := objects[env.Id]; !ok {
			objects[env.Id] = fmt.Sprintf("%s/%s", strings.ToLower(env.EnvironmentIdentifier), strings.ToLower(appName))
			envObjectToName[objects[env.Id]] = env.Name
		}
	}
	return objects, envObjectToName
}

func (impl EnforcerUtilImpl) GetRbacObjectsByAppIds(appIds []int) map[int]string {
	objects := make(map[int]string)
	result, err := impl.appRepo.FindAppAndProjectByIdsIn(appIds)
	if err != nil {
		impl.logger.Errorw("error occurred in fetching apps", "appIds", appIds)
		return objects
	}
	for _, item := range result {
		if _, ok := objects[item.Id]; !ok {
			objects[item.Id] = fmt.Sprintf("%s/%s", strings.ToLower(item.Team.Name), strings.ToLower(item.AppName))
		}
	}
	return objects
}

func (impl EnforcerUtilImpl) GetAppRBACName(appName string) string {
	application, err := impl.appRepo.FindAppAndProjectByAppName(appName)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(appName))
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(application.Team.Name), strings.ToLower(appName))
}

func (impl EnforcerUtilImpl) GetProjectAdminRBACNameBYAppName(appName string) string {
	application, err := impl.appRepo.FindAppAndProjectByAppName(appName)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(appName))
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(application.Team.Name), "*")
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

func (impl EnforcerUtilImpl) GetRbacObjectsForAllAppsWithTeamID(teamID int) map[int]string {
	objects := make(map[int]string)
	result, err := impl.appRepo.FindAllActiveAppsWithTeamWithTeamId(teamID)
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
	application, err := impl.appRepo.FindAppAndProjectByAppId(appId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(application.Team.Name), strings.ToLower(application.AppName))
}

func (impl EnforcerUtilImpl) GetAppRBACByAppNameAndEnvId(appName string, envId int) string {
	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(appName))
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(env.EnvironmentIdentifier), strings.ToLower(appName))
}

func (impl EnforcerUtilImpl) GetAppRBACByAppIdAndPipelineId(appId int, pipelineId int) string {
	application, err := impl.appRepo.FindById(appId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	pipeline, err := impl.pipelineRepository.FindById(pipelineId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(application.AppName))
	}
	env, err := impl.environmentRepository.FindById(pipeline.EnvironmentId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(application.AppName))
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(env.EnvironmentIdentifier), strings.ToLower(application.AppName))
}

func (impl EnforcerUtilImpl) GetEnvRBACNameByAppId(appId int, envId int) string {
	application, err := impl.appRepo.FindById(appId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	var appName = application.AppName
	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(appName))
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(env.EnvironmentIdentifier), strings.ToLower(appName))
}

func (impl EnforcerUtilImpl) GetTeamEnvRBACNameByAppId(appId int, envId int) string {
	application, err := impl.appRepo.FindAppAndProjectByAppId(appId)
	if err != nil {
		return fmt.Sprintf("%s/%s/%s", "", "", "")
	}
	var appName = application.AppName
	var teamName = application.Team.Name
	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		return fmt.Sprintf("%s/%s/%s", strings.ToLower(teamName), "", strings.ToLower(appName))
	}
	return fmt.Sprintf("%s/%s/%s", strings.ToLower(teamName), strings.ToLower(env.EnvironmentIdentifier), strings.ToLower(appName))
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
	application, err := impl.appRepo.FindById(ciPipeline.AppId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	appName := application.AppName
	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", strings.ToLower(appName))
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(env.EnvironmentIdentifier), strings.ToLower(appName))
}

func (impl EnforcerUtilImpl) GetEnvRBACNameByCdPipelineIdAndEnvId(cdPipelineId int) string {
	pipeline, err := impl.pipelineRepository.FindById(cdPipelineId)
	if err != nil {
		impl.logger.Error(err)
		return fmt.Sprintf("%s/%s", "", "")
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(pipeline.Environment.EnvironmentIdentifier), strings.ToLower(pipeline.App.AppName))
}

func (impl EnforcerUtilImpl) GetTeamRbacObjectByCiPipelineId(ciPipelineId int) string {
	ciPipeline, err := impl.ciPipelineRepository.FindById(ciPipelineId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	application, err := impl.appRepo.FindAppAndProjectByAppId(ciPipeline.AppId)
	if err != nil {
		return fmt.Sprintf("%s/%s", "", "")
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(application.Team.Name), strings.ToLower(ciPipeline.App.AppName))
}

func (impl EnforcerUtilImpl) GetTeamAndEnvironmentRbacObjectByCDPipelineId(pipelineId int) (string, string) {
	pipeline, err := impl.pipelineRepository.FindById(pipelineId)
	if err != nil {
		impl.logger.Error(err)
		return "", ""
	}
	application, err := impl.appRepo.FindAppAndProjectByAppId(pipeline.AppId)
	if err != nil {
		impl.logger.Errorw("error on fetching data for rbac object", "err", err)
		return "", ""
	}
	teamRbac := fmt.Sprintf("%s/%s", strings.ToLower(application.Team.Name), strings.ToLower(pipeline.App.AppName))
	envRbac := fmt.Sprintf("%s/%s", strings.ToLower(pipeline.Environment.EnvironmentIdentifier), strings.ToLower(pipeline.App.AppName))
	return teamRbac, envRbac
}

func (impl EnforcerUtilImpl) GetRbacObjectsForAllAppsAndEnvironments() (map[int]string, map[string]string) {
	appObjects := make(map[int]string)
	envObjects := make(map[string]string)
	apps, err := impl.appRepo.FindAllActiveAppsWithTeam()
	if err != nil {
		impl.logger.Errorw("exception while fetching all active apps for rbac objects", "err", err)
		return appObjects, envObjects
	}
	for _, item := range apps {
		if _, ok := appObjects[item.Id]; !ok {
			appObjects[item.Id] = fmt.Sprintf("%s/%s", strings.ToLower(item.Team.Name), strings.ToLower(item.AppName))
		}
	}

	envs, err := impl.environmentRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("exception while fetching all active env for rbac objects", "err", err)
		return appObjects, envObjects
	}
	for _, env := range envs {
		for _, app := range apps {
			key := fmt.Sprintf("%d-%d", env.Id, app.Id)
			if _, ok := envObjects[key]; !ok {
				envObjects[key] = fmt.Sprintf("%s/%s", strings.ToLower(env.EnvironmentIdentifier), strings.ToLower(app.AppName))
			}
		}
	}
	return appObjects, envObjects
}

func (impl EnforcerUtilImpl) GetHelmObject(appId int, envId int) (string, string) {
	application, err := impl.appRepo.FindAppAndProjectByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error on fetching data for rbac object", "err", err)
		return fmt.Sprintf("%s/%s/%s", "", "", ""), ""
	}
	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.logger.Errorw("error on fetching data for rbac object", "err", err)
		return fmt.Sprintf("%s/%s/%s", "", "", ""), ""
	}
	clusterName := env.Cluster.ClusterName
	namespace := env.Namespace

	environmentIdentifier := env.EnvironmentIdentifier

	// Fix for futuristic permissions, environmentIdentifier2 will return a string object with cluster name if futuristic permissions are given,
	//otherwise it will return an empty string object

	environmentIdentifier2 := ""
	envIdentifierByClusterAndNamespace := clusterName + "__" + namespace
	if !env.IsVirtualEnvironment { //because for virtual_environment environment identifier is equal to environment name (As namespace is optional)
		if environmentIdentifier != envIdentifierByClusterAndNamespace { // for futuristic permission cluster name is not present in environment identifier
			environmentIdentifier2 = envIdentifierByClusterAndNamespace
		}
	}
	//TODO - FIX required for futuristic permission for cluster__* all environment for migrated environment identifier only
	/*//here cluster, env, namespace must not have double underscore in names, as we are using that for separator.
	if !strings.HasPrefix(env.EnvironmentIdentifier, fmt.Sprintf("%s__", env.Cluster.ClusterName)) {
		environmentIdentifier = fmt.Sprintf("%s__%s", env.Cluster.ClusterName, env.EnvironmentIdentifier)
	}*/

	if environmentIdentifier2 == "" {
		return fmt.Sprintf("%s/%s/%s", strings.ToLower(application.Team.Name), environmentIdentifier, strings.ToLower(application.AppName)), ""
	}

	return fmt.Sprintf("%s/%s/%s", strings.ToLower(application.Team.Name), environmentIdentifier, strings.ToLower(application.AppName)),
		fmt.Sprintf("%s/%s/%s", strings.ToLower(application.Team.Name), environmentIdentifier2, strings.ToLower(application.AppName))
}

func (impl EnforcerUtilImpl) GetHelmObjectByAppNameAndEnvId(appName string, envId int) (string, string) {
	application, err := impl.appRepo.FindAppAndProjectByAppName(appName)
	if err != nil {
		impl.logger.Errorw("error on fetching data for rbac object", "err", err)
		return fmt.Sprintf("%s/%s/%s", "", "", ""), ""
	}
	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.logger.Errorw("error on fetching data for rbac object", "err", err)
		return fmt.Sprintf("%s/%s/%s", "", "", ""), ""
	}
	clusterName := env.Cluster.ClusterName
	namespace := env.Namespace
	environmentIdentifier := env.EnvironmentIdentifier

	// Fix for futuristic permissions, environmentIdentifier2 will return a string object with cluster name if futuristic permissions are given,
	//otherwise it will return an empty string object
	environmentIdentifier2 := ""
	if !env.IsVirtualEnvironment {
		if environmentIdentifier != clusterName+"__"+namespace { // for futuristic permission cluster name is not present in environment identifier
			environmentIdentifier2 = clusterName + "__" + namespace
		}
	}
	if environmentIdentifier2 == "" {
		return fmt.Sprintf("%s/%s/%s", strings.ToLower(application.Team.Name), environmentIdentifier, strings.ToLower(application.AppName)), ""
	}

	//TODO - FIX required for futuristic permission for cluster__* all environment for migrated environment identifier only
	/*//here cluster, env, namespace must not have double underscore in names, as we are using that for separator.
	if !strings.HasPrefix(env.EnvironmentIdentifier, fmt.Sprintf("%s__", env.Cluster.ClusterName)) {
		environmentIdentifier = fmt.Sprintf("%s__%s", env.Cluster.ClusterName, env.EnvironmentIdentifier)
	}*/
	return fmt.Sprintf("%s/%s/%s", strings.ToLower(application.Team.Name), environmentIdentifier, strings.ToLower(application.AppName)),
		fmt.Sprintf("%s/%s/%s", strings.ToLower(application.Team.Name), environmentIdentifier2, strings.ToLower(application.AppName))
}

func (impl EnforcerUtilImpl) GetHelmObjectByProjectIdAndEnvId(teamId int, envId int) (string, string) {
	team, err := impl.teamRepository.FindOne(teamId)
	if err != nil {
		impl.logger.Errorw("error on fetching data for rbac object", "err", err)
		return fmt.Sprintf("%s/%s/%s", "", "", ""), fmt.Sprintf("%s/%s/%s", "", "", "")
	}
	env, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.logger.Errorw("error on fetching data for rbac object", "err", err)
		return fmt.Sprintf("%s/%s/%s", "", "", ""), fmt.Sprintf("%s/%s/%s", "", "", "")
	}

	clusterName := env.Cluster.ClusterName
	namespace := env.Namespace

	environmentIdentifier := env.EnvironmentIdentifier

	// Fix for futuristic permissions, environmentIdentifier2 will return a string object with cluster name if futuristic permissions are given,
	//otherwise it will return an empty string object

	environmentIdentifier2 := ""
	if !env.IsVirtualEnvironment {
		if environmentIdentifier != clusterName+"__"+namespace { // for futuristic permission cluster name is not present in environment identifier
			environmentIdentifier2 = clusterName + "__" + namespace
		}
	}

	if environmentIdentifier2 == "" {
		return fmt.Sprintf("%s/%s/%s", strings.ToLower(team.Name), environmentIdentifier, "*"), ""
	}

	//TODO - FIX required for futuristic permission for cluster__* all environment for migrated environment identifier only
	/*//here cluster, env, namespace must not have double underscore in names, as we are using that for separator.
	if !strings.HasPrefix(env.EnvironmentIdentifier, fmt.Sprintf("%s__", env.Cluster.ClusterName)) {
		environmentIdentifier = fmt.Sprintf("%s__%s", env.Cluster.ClusterName, env.EnvironmentIdentifier)
	}*/
	return fmt.Sprintf("%s/%s/%s", strings.ToLower(team.Name), environmentIdentifier, "*"),
		fmt.Sprintf("%s/%s/%s", strings.ToLower(team.Name), environmentIdentifier2, "*")
}

func (impl EnforcerUtilImpl) GetAppRBACNameByTeamIdAndAppId(teamId int, appId int) string {
	application, err := impl.appRepo.FindById(appId)
	if err != nil {
		impl.logger.Errorw("error on fetching data for rbac object", "err", err)
		return fmt.Sprintf("%s/%s", "", "")
	}
	team, err := impl.teamRepository.FindOne(teamId)
	if err != nil {
		impl.logger.Errorw("error on fetching data for rbac object", "err", err)
		return fmt.Sprintf("%s/%s", "", "")
	}
	return fmt.Sprintf("%s/%s", strings.ToLower(team.Name), strings.ToLower(application.AppName))
}

func (impl EnforcerUtilImpl) GetRBACNameForClusterEntity(clusterName string, resourceIdentifier k8s.ResourceIdentifier) (resourceName, objectName string) {
	namespace := resourceIdentifier.Namespace
	objectName = resourceIdentifier.Name
	groupVersionKind := resourceIdentifier.GroupVersionKind
	groupName := groupVersionKind.Group
	kindName := groupVersionKind.Kind
	if groupName == "" {
		groupName = casbin.ClusterEmptyGroupPlaceholder
	}
	if namespace == "" { //empty value means all namespace access would occur for non-namespace resources
		namespace = "*"
	}
	resourceName = fmt.Sprintf(casbin.ClusterResourceRegex, clusterName, namespace)
	objectName = fmt.Sprintf(casbin.ClusterObjectRegex, groupName, kindName, objectName)
	return resourceName, objectName
}

func (impl EnforcerUtilImpl) GetAppObjectByCiPipelineIds(ciPipelineIds []int) map[int]string {
	objects := make(map[int]string)
	models, err := impl.ciPipelineRepository.FindAppAndProjectByCiPipelineIds(ciPipelineIds)
	if err != nil {
		impl.logger.Error(err)
		return objects
	}
	for _, pipeline := range models {
		if _, ok := objects[pipeline.Id]; !ok {
			appObject := fmt.Sprintf("%s/%s", strings.ToLower(pipeline.App.Team.Name), strings.ToLower(pipeline.App.AppName))
			objects[pipeline.Id] = appObject
		}
	}
	return objects
}

func (impl EnforcerUtilImpl) GetAppAndEnvObjectByPipelineIds(cdPipelineIds []int) map[int][]string {
	objects := make(map[int][]string)
	models, err := impl.pipelineRepository.FindAppAndEnvironmentAndProjectByPipelineIds(cdPipelineIds)
	if err != nil {
		impl.logger.Error(err)
		return objects
	}
	for _, pipeline := range models {
		if _, ok := objects[pipeline.Id]; !ok {
			appObject := fmt.Sprintf("%s/%s", strings.ToLower(pipeline.App.Team.Name), strings.ToLower(pipeline.App.AppName))
			envObject := fmt.Sprintf("%s/%s", strings.ToLower(pipeline.Environment.EnvironmentIdentifier), strings.ToLower(pipeline.App.AppName))
			objects[pipeline.Id] = []string{appObject, envObject}
		}
	}
	return objects
}

func (impl EnforcerUtilImpl) GetRbacObjectsForAllAppsWithMatchingAppName(appNameMatch string) map[int]string {
	objects := make(map[int]string)
	result, err := impl.appRepo.FindAllActiveAppsWithTeamByAppNameMatch(appNameMatch)
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
func (impl EnforcerUtilImpl) GetAppAndEnvObjectByPipeline(cdPipelines []*bean.CDPipelineConfigObject) map[int][]string {
	objects := make(map[int][]string)
	var teamIds []*int
	teamMap := make(map[int]string)
	for _, pipeline := range cdPipelines {
		teamIds = append(teamIds, &pipeline.TeamId)
	}
	teams, err := impl.teamRepository.FindByIds(teamIds)
	if err != nil {
		return objects
	}
	for _, team := range teams {
		if _, ok := teamMap[team.Id]; !ok {
			teamMap[team.Id] = team.Name
		}
	}
	for _, pipeline := range cdPipelines {
		if _, ok := objects[pipeline.Id]; !ok {
			appObject := fmt.Sprintf("%s/%s", strings.ToLower(teamMap[pipeline.TeamId]), strings.ToLower(pipeline.AppName))
			envObject := fmt.Sprintf("%s/%s", strings.ToLower(pipeline.EnvironmentIdentifier), strings.ToLower(pipeline.AppName))
			objects[pipeline.Id] = []string{appObject, envObject}
		}
	}
	return objects
}

// GetAppAndEnvObjectByDbPipeline TODO - This function will be merge into GetAppAndEnvObjectByPipeline
func (impl EnforcerUtilImpl) GetAppAndEnvObjectByDbPipeline(cdPipelines []*pipelineConfig.Pipeline) map[int][]string {
	objects := make(map[int][]string)
	var teamIds []*int
	teamMap := make(map[int]string)
	for _, pipeline := range cdPipelines {
		teamIds = append(teamIds, &pipeline.App.TeamId)
	}
	teams, err := impl.teamRepository.FindByIds(teamIds)
	if err != nil {
		return objects
	}
	for _, team := range teams {
		if _, ok := teamMap[team.Id]; !ok {
			teamMap[team.Id] = team.Name
		}
	}
	for _, pipeline := range cdPipelines {
		if _, ok := objects[pipeline.Id]; !ok {
			appObject := fmt.Sprintf("%s/%s", strings.ToLower(teamMap[pipeline.App.TeamId]), strings.ToLower(pipeline.App.AppName))
			envObject := fmt.Sprintf("%s/%s", strings.ToLower(pipeline.Environment.EnvironmentIdentifier), strings.ToLower(pipeline.App.AppName))
			objects[pipeline.Id] = []string{appObject, envObject}
		}
	}
	return objects
}

func (impl EnforcerUtilImpl) GetAllActiveTeamNames() ([]string, error) {
	teamNames, err := impl.teamRepository.FindAllActiveTeamNames()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching active team names", "err", err)
		return nil, err
	}
	for i, teamName := range teamNames {
		teamNames[i] = strings.ToLower(teamName)
	}
	return teamNames, nil
}

func (impl EnforcerUtilImpl) GetAppRBACNameByAppAndProjectName(projectName, appName string) string {
	return fmt.Sprintf("%s/%s", strings.ToLower(projectName), strings.ToLower(appName))
}
