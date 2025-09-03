/*
 * Copyright (c) 2024. Devtron Inc.
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
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/read"
	"github.com/devtron-labs/devtron/pkg/cluster/environment"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	repository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"k8s.io/utils/strings/slices"
)

type DevtronAppAutoCompleteRestHandler interface {
	GetAppListForAutocomplete(w http.ResponseWriter, r *http.Request)
	EnvironmentListAutocomplete(w http.ResponseWriter, r *http.Request)
	GitListAutocomplete(w http.ResponseWriter, r *http.Request)
	RegistriesListAutocomplete(w http.ResponseWriter, r *http.Request)
	TeamListAutocomplete(w http.ResponseWriter, r *http.Request)
}

type DevtronAppAutoCompleteRestHandlerImpl struct {
	Logger                  *zap.SugaredLogger
	userAuthService         user.UserService
	teamService             team.TeamService
	enforcer                casbin.Enforcer
	enforcerUtil            rbac.EnforcerUtil
	devtronAppConfigService pipeline.DevtronAppConfigService
	envService              environment.EnvironmentService
	gitProviderReadService  read.GitProviderReadService
	dockerRegistryConfig    pipeline.DockerRegistryConfig
}

func NewDevtronAppAutoCompleteRestHandlerImpl(
	Logger *zap.SugaredLogger,
	userAuthService user.UserService,
	teamService team.TeamService,
	enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil,
	devtronAppConfigService pipeline.DevtronAppConfigService,
	envService environment.EnvironmentService,
	dockerRegistryConfig pipeline.DockerRegistryConfig,
	gitProviderReadService read.GitProviderReadService) *DevtronAppAutoCompleteRestHandlerImpl {
	return &DevtronAppAutoCompleteRestHandlerImpl{
		Logger:                  Logger,
		userAuthService:         userAuthService,
		teamService:             teamService,
		enforcer:                enforcer,
		enforcerUtil:            enforcerUtil,
		devtronAppConfigService: devtronAppConfigService,
		envService:              envService,
		dockerRegistryConfig:    dockerRegistryConfig,
		gitProviderReadService:  gitProviderReadService,
	}
}

func (handler DevtronAppAutoCompleteRestHandlerImpl) GetAppListForAutocomplete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
	token := r.Header.Get("token")
	isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")

	v := r.URL.Query()
	teamId := v.Get("teamId")
	appName := v.Get("appName")
	offset := 0
	size := 0 // default value is 0, it means if not provided in query param it will fetch all
	sizeStr := v.Get("size")
	if sizeStr != "" {
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 0 {
			common.WriteJsonResp(w, errors.New("invalid size"), nil, http.StatusBadRequest)
			return
		}
	}
	offsetStr := v.Get("offset")
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			common.WriteJsonResp(w, errors.New("invalid offset"), nil, http.StatusBadRequest)
			return
		}
	}
	appTypeParam := v.Get("appType")
	var appType int
	if appTypeParam != "" {
		appType, err = strconv.Atoi(appTypeParam)
		if err != nil {
			handler.Logger.Errorw("service err, GetAppListForAutocomplete", "err", err, "teamId", teamId, "appTypeParam", appTypeParam)
			common.WriteJsonResp(w, err, "Failed to parse appType param", http.StatusInternalServerError)
			return
		}
	} else {
		// if appType not provided we are considering it as customApp for now, doing this because to get all apps by team id rbac objects
		appType = int(helper.CustomApp)
	}
	var teamIdInt int
	handler.Logger.Infow("request payload, GetAppListForAutocomplete", "teamId", teamId)
	var apps []*pipeline.AppBean
	if len(teamId) == 0 {
		apps, err = handler.devtronAppConfigService.FindAllMatchesByAppName(appName, helper.AppType(appType))
		if err != nil {
			handler.Logger.Errorw("service err, GetAppListForAutocomplete", "err", err, "teamId", teamId)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	} else {
		teamIdInt, err = common.ExtractIntPathParamWithContext(w, r, "teamId", teamId+" team")
		if err != nil {
			// Error already written by ExtractIntPathParamWithContext
			return
		}
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		} else {
			apps, err = handler.devtronAppConfigService.FindAppsByTeamId(teamIdInt)
			if err != nil {
				handler.Logger.Errorw("service err, GetAppListForAutocomplete", "err", err, "teamId", teamId)
				common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}
		}
	}
	if isActionUserSuperAdmin {
		apps = handler.getPaginatedResultsForApps(offset, size, apps)
		common.WriteJsonResp(w, err, apps, http.StatusOK)
		return
	}

	// RBAC
	_, span := otel.Tracer("autoCompleteAppAPI").Start(context.Background(), "RBACForAutoCompleteAppAPI")
	accessedApps := make([]*pipeline.AppBean, 0)
	rbacObjects := make([]string, 0)

	var appIdToObjectMap map[int]string
	if len(teamId) == 0 {
		appIdToObjectMap = handler.enforcerUtil.GetRbacObjectsForAllAppsWithMatchingAppName(appName, helper.AppType(appType))
	} else {
		appIdToObjectMap = handler.enforcerUtil.GetRbacObjectsForAllAppsWithTeamID(teamIdInt, helper.AppType(appType))
	}

	for _, app := range apps {
		object := appIdToObjectMap[app.Id]
		rbacObjects = append(rbacObjects, object)
	}

	enforcedMap := handler.enforcerUtil.CheckAppRbacForAppOrJobInBulk(token, casbin.ActionGet, rbacObjects, helper.AppType(appType))
	for _, app := range apps {
		object := appIdToObjectMap[app.Id]
		if enforcedMap[object] {
			accessedApps = append(accessedApps, app)
		}
	}
	span.End()
	// RBAC
	accessedApps = handler.getPaginatedResultsForApps(offset, size, accessedApps)
	common.WriteJsonResp(w, err, accessedApps, http.StatusOK)
}

func (handler DevtronAppAutoCompleteRestHandlerImpl) EnvironmentListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, EnvironmentListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	showDeploymentOptionsParam := false
	param := r.URL.Query().Get("showDeploymentOptions")
	if param != "" {
		showDeploymentOptionsParam, _ = strconv.ParseBool(param)
	}
	result, err := handler.envService.GetEnvironmentListForAutocomplete(showDeploymentOptionsParam)
	if err != nil {
		handler.Logger.Errorw("service err, EnvironmentListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (handler DevtronAppAutoCompleteRestHandlerImpl) GitListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GitListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, object, casbin.ActionGet)
	if !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	res, err := handler.gitProviderReadService.GetAll()
	if err != nil {
		handler.Logger.Errorw("service err, GitListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler DevtronAppAutoCompleteRestHandlerImpl) RegistriesListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	// Use enhanced parameter parsing with context
	appId, err := common.ExtractIntPathParamWithContext(w, r, "appId", "application")
	if err != nil {
		// Error already written by ExtractIntPathParamWithContext
		return
	}
	v := r.URL.Query()
	storageType := v.Get("storageType")
	if storageType == "" {
		storageType = repository.OCI_REGISRTY_REPO_TYPE_CONTAINER
	}
	if !slices.Contains(repository.OCI_REGISRTY_REPO_TYPE_LIST, storageType) {
		common.WriteJsonResp(w, fmt.Errorf("invalid query parameters"), nil, http.StatusBadRequest)
		return
	}
	storageAction := v.Get("storageAction")
	if storageAction == "" {
		storageAction = repository.STORAGE_ACTION_TYPE_PUSH
	}
	if !(storageAction == repository.STORAGE_ACTION_TYPE_PULL || storageAction == repository.STORAGE_ACTION_TYPE_PUSH) {
		common.WriteJsonResp(w, fmt.Errorf("invalid query parameters"), nil, http.StatusBadRequest)
		return
	}

	handler.Logger.Infow("request payload, DockerListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	registryConfigs, err := handler.dockerRegistryConfig.ListAllActive()
	if err != nil {
		handler.Logger.Errorw("service err, DockerListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	res := handler.dockerRegistryConfig.FilterRegistryBeanListBasedOnStorageTypeAndAction(registryConfigs, storageType, storageAction, repository.STORAGE_ACTION_TYPE_PULL_AND_PUSH)
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler DevtronAppAutoCompleteRestHandlerImpl) TeamListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, TeamListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	result, err := handler.teamService.FetchForAutocomplete()
	if err != nil {
		handler.Logger.Errorw("service err, TeamListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (handler DevtronAppAutoCompleteRestHandlerImpl) getPaginatedResultsForApps(offset int, size int, apps []*pipeline.AppBean) []*pipeline.AppBean {
	if size > 0 {
		if offset+size <= len(apps) {
			apps = apps[offset : offset+size]
		} else {
			apps = apps[offset:]
		}
	}
	return apps
}
