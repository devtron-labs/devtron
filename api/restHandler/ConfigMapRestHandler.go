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

package restHandler

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type ConfigMapRestHandler interface {
	CMGlobalAddUpdate(w http.ResponseWriter, r *http.Request)
	CMEnvironmentAddUpdate(w http.ResponseWriter, r *http.Request)
	CMGlobalFetch(w http.ResponseWriter, r *http.Request)
	CMEnvironmentFetch(w http.ResponseWriter, r *http.Request)

	CSGlobalAddUpdate(w http.ResponseWriter, r *http.Request)
	CSEnvironmentAddUpdate(w http.ResponseWriter, r *http.Request)
	CSGlobalFetch(w http.ResponseWriter, r *http.Request)
	CSEnvironmentFetch(w http.ResponseWriter, r *http.Request)

	CMGlobalDelete(w http.ResponseWriter, r *http.Request)
	CMEnvironmentDelete(w http.ResponseWriter, r *http.Request)
	CSGlobalDelete(w http.ResponseWriter, r *http.Request)
	CSEnvironmentDelete(w http.ResponseWriter, r *http.Request)

	CSGlobalFetchForEdit(w http.ResponseWriter, r *http.Request)
	CSEnvironmentFetchForEdit(w http.ResponseWriter, r *http.Request)
	ConfigSecretBulkPatch(w http.ResponseWriter, r *http.Request)
	ConfigSecretBulkPatchV2(w http.ResponseWriter, r *http.Request)
}

type ConfigMapRestHandlerImpl struct {
	pipelineBuilder    pipeline.PipelineBuilder
	Logger             *zap.SugaredLogger
	chartService       pipeline.ChartService
	userAuthService    user.UserService
	teamService        team.TeamService
	enforcer           rbac.Enforcer
	pipelineRepository pipelineConfig.PipelineRepository
	enforcerUtil       rbac.EnforcerUtil
	configMapService   pipeline.ConfigMapService
}

func NewConfigMapRestHandlerImpl(pipelineBuilder pipeline.PipelineBuilder, Logger *zap.SugaredLogger,
	chartService pipeline.ChartService, userAuthService user.UserService, teamService team.TeamService,
	enforcer rbac.Enforcer, pipelineRepository pipelineConfig.PipelineRepository,
	enforcerUtil rbac.EnforcerUtil, configMapService pipeline.ConfigMapService) *ConfigMapRestHandlerImpl {
	return &ConfigMapRestHandlerImpl{
		pipelineBuilder:    pipelineBuilder,
		Logger:             Logger,
		chartService:       chartService,
		userAuthService:    userAuthService,
		teamService:        teamService,
		enforcer:           enforcer,
		pipelineRepository: pipelineRepository,
		enforcerUtil:       enforcerUtil,
		configMapService:   configMapService,
	}
}

func (handler ConfigMapRestHandlerImpl) CMGlobalAddUpdate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var configMapRequest pipeline.ConfigDataRequest

	err = decoder.Decode(&configMapRequest)
	if err != nil {
		handler.Logger.Errorw("request err, CMGlobalAddUpdate", "err", err, "payload", configMapRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	configMapRequest.UserId = userId
	handler.Logger.Errorw("request payload, CMGlobalAddUpdate", "payload", configMapRequest)

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(configMapRequest.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CMGlobalAddUpdate(&configMapRequest)
	if err != nil {
		handler.Logger.Errorw("service err, CMGlobalAddUpdate", "err", err, "payload", configMapRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CMEnvironmentAddUpdate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var configMapRequest pipeline.ConfigDataRequest
	err = decoder.Decode(&configMapRequest)
	if err != nil {
		handler.Logger.Errorw("request err, CMEnvironmentAddUpdate", "err", err, "payload", configMapRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	configMapRequest.UserId = userId
	handler.Logger.Errorw("request payload, CMEnvironmentAddUpdate", "payload", configMapRequest)

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(configMapRequest.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(configMapRequest.AppId, configMapRequest.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionCreate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CMEnvironmentAddUpdate(&configMapRequest)
	if err != nil {
		handler.Logger.Errorw("service err, CMEnvironmentAddUpdate", "err", err, "payload", configMapRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CMGlobalFetch(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, CMGlobalFetch", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CMGlobalFetch(appId)
	if err != nil {
		handler.Logger.Errorw("service err, CMGlobalFetch", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CMEnvironmentFetch(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, CMEnvironmentFetch", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.Logger.Errorw("request err, CMEnvironmentFetch", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CMEnvironmentFetch(appId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, CMEnvironmentFetch", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CSGlobalAddUpdate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var configMapRequest pipeline.ConfigDataRequest

	err = decoder.Decode(&configMapRequest)
	if err != nil {
		handler.Logger.Errorw("request err, CSGlobalAddUpdate", "err", err, "payload", configMapRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	configMapRequest.UserId = userId
	handler.Logger.Errorw("request payload, CSGlobalAddUpdate", "payload", configMapRequest)

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(configMapRequest.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CSGlobalAddUpdate(&configMapRequest)
	if err != nil {
		handler.Logger.Errorw("service err, CSGlobalAddUpdate", "err", err, "payload", configMapRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CSEnvironmentAddUpdate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var configMapRequest pipeline.ConfigDataRequest

	err = decoder.Decode(&configMapRequest)
	if err != nil {
		handler.Logger.Errorw("request err, CSEnvironmentAddUpdate", "err", err, "payload", configMapRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	configMapRequest.UserId = userId
	handler.Logger.Errorw("request payload, CSEnvironmentAddUpdate", "payload", configMapRequest)

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(configMapRequest.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(configMapRequest.AppId, configMapRequest.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionCreate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CSEnvironmentAddUpdate(&configMapRequest)
	if err != nil {
		handler.Logger.Errorw("service err, CSEnvironmentAddUpdate", "err", err, "payload", configMapRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CSGlobalFetch(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, CSGlobalFetch", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CSGlobalFetch(appId)
	if err != nil {
		handler.Logger.Errorw("service err, CSGlobalFetch", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CSEnvironmentFetch(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, CSEnvironmentFetch", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.Logger.Errorw("bad request", "err", err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CSEnvironmentFetch(appId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, CSEnvironmentFetch", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CMGlobalDelete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, CMGlobalDelete", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.Logger.Errorw("request err, CMGlobalDelete", "err", err, "id", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	name := vars["name"]
	handler.Logger.Errorw("request payload, CMGlobalDelete", "appId", appId, "id", id, "name", name)

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionDelete, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CMGlobalDelete(name, id, userId)
	if err != nil {
		handler.Logger.Errorw("service err, CMGlobalDelete", "err", err, "appId", appId, "id", id, "name", name)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CMEnvironmentDelete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, CMEnvironmentDelete", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.Logger.Errorw("request err, CMEnvironmentDelete", "err", err, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.Logger.Errorw("request err, CMEnvironmentDelete", "err", err, "id", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	name := vars["name"]
	handler.Logger.Errorw("request payload, CMEnvironmentDelete", "appId", appId, "envId", envId, "id", id, "name", name)

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionDelete, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionDelete, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CMEnvironmentDelete(name, id, userId)
	if err != nil {
		handler.Logger.Errorw("service err, CMEnvironmentDelete", "err", err, "appId", appId, "envId", envId, "id", id)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CSGlobalDelete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, CSGlobalDelete", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.Logger.Errorw("request err, CSGlobalDelete", "err", err, "id", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	name := vars["name"]
	handler.Logger.Errorw("request payload, CSGlobalDelete", "appId", appId, "id", id, "name", name)

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionDelete, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CSGlobalDelete(name, id, userId)
	if err != nil {
		handler.Logger.Errorw("service err, CSGlobalDelete", "err", err, "appId", appId, "id", id, "name", name)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CSEnvironmentDelete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, CSEnvironmentDelete", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.Logger.Errorw("request err, CSEnvironmentDelete", "err", err, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.Logger.Errorw("request err, CSEnvironmentDelete", "err", err, "id", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	name := vars["name"]
	handler.Logger.Errorw("request payload, CSEnvironmentDelete", "appId", appId, "envId", envId, "id", id, "name", name)

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionDelete, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionDelete, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CSEnvironmentDelete(name, id, userId)
	if err != nil {
		handler.Logger.Errorw("service err, CSEnvironmentDelete", "err", err, "appId", appId, "envId", envId, "id", id)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CSGlobalFetchForEdit(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, CSGlobalFetchForEdit", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.Logger.Errorw("request err, CSGlobalFetchForEdit", "err", err, "id", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	name := vars["name"]
	handler.Logger.Errorw("request payload, CSGlobalFetchForEdit", "appId", appId, "id", id, "name", name)

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionUpdate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CSGlobalFetchForEdit(name, id, userId)
	if err != nil {
		handler.Logger.Errorw("service err, CSGlobalFetchForEdit", "err", err, "appId", appId, "id", id, "name", name)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) CSEnvironmentFetchForEdit(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.Logger.Errorw("request err, CSEnvironmentFetchForEdit", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.Logger.Errorw("request err, CSEnvironmentFetchForEdit", "err", err, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.Logger.Errorw("request err, CSEnvironmentFetchForEdit", "err", err, "id", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	name := vars["name"]
	handler.Logger.Errorw("request payload, CSEnvironmentFetchForEdit", "appId", appId, "envId", envId, "id", id, "name", name)

	//RBAC START
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionUpdate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionUpdate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapService.CSEnvironmentFetchForEdit(name, id, appId, envId, userId)
	if err != nil {
		handler.Logger.Errorw("service err, CSEnvironmentFetchForEdit", "err", err, "appId", appId, "envId", envId, "id", id)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) ConfigSecretBulkPatch(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	//AUTH - check from casbin db
	roles, err := handler.userAuthService.CheckUserRoles(userId)
	if err != nil {
		writeJsonResp(w, err, []byte("Failed to get user by id"), http.StatusInternalServerError)
		return
	}
	superAdmin := false
	for _, item := range roles {
		if item == bean.SUPERADMIN {
			superAdmin = true
		}
	}
	if superAdmin == false {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//AUTH

	var bulkPatchRequest pipeline.BulkPatchRequest
	err = decoder.Decode(&bulkPatchRequest)
	if err != nil {
		handler.Logger.Errorw("request err, ConfigSecretBulkPatch", "err", err, "payload", bulkPatchRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Errorw("request payload, ConfigSecretBulkPatch", "payload", bulkPatchRequest)

	if bulkPatchRequest.ProjectId > 0 && bulkPatchRequest.Global {
		apps, err := handler.pipelineBuilder.FindAppsByTeamId(bulkPatchRequest.ProjectId)
		if err != nil {
			handler.Logger.Errorw("service err, ConfigSecretBulkPatch", "err", err, "payload", bulkPatchRequest)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		var payload []*pipeline.BulkPatchPayload
		for _, app := range apps {
			payload = append(payload, &pipeline.BulkPatchPayload{AppId: app.Id})
		}
		bulkPatchRequest.Payload = payload
	}

	bulkPatchRequest.UserId = userId
	if bulkPatchRequest.Global {
		_, err := handler.configMapService.ConfigSecretGlobalBulkPatch(&bulkPatchRequest)
		if err != nil {
			handler.Logger.Errorw("service err, ConfigSecretBulkPatch", "err", err, "payload", bulkPatchRequest)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	} else {
		_, err := handler.configMapService.ConfigSecretEnvironmentBulkPatch(&bulkPatchRequest)
		if err != nil {
			handler.Logger.Errorw("service err, ConfigSecretBulkPatch", "err", err, "payload", bulkPatchRequest)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	writeJsonResp(w, err, true, http.StatusOK)
}

func (handler ConfigMapRestHandlerImpl) ConfigSecretBulkPatchV2(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	//AUTH - check from casbin db
	roles, err := handler.userAuthService.CheckUserRoles(userId)
	if err != nil {
		writeJsonResp(w, err, []byte("Failed to get user by id"), http.StatusInternalServerError)
		return
	}
	superAdmin := false
	for _, item := range roles {
		if item == bean.SUPERADMIN {
			superAdmin = true
		}
	}
	if superAdmin == false {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//AUTH

	var bulkPatchRequest pipeline.BulkPatchRequest
	err = decoder.Decode(&bulkPatchRequest)
	if err != nil {
		handler.Logger.Errorw("request err, ConfigSecretBulkPatch", "err", err, "payload", bulkPatchRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Errorw("request payload, ConfigSecretBulkPatch", "payload", bulkPatchRequest)

	if bulkPatchRequest.ProjectId > 0 && bulkPatchRequest.Global {
		apps, err := handler.pipelineBuilder.FindAppsByTeamId(bulkPatchRequest.ProjectId)
		if err != nil {
			handler.Logger.Errorw("service err, ConfigSecretBulkPatch", "err", err, "payload", bulkPatchRequest)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		var payload []*pipeline.BulkPatchPayload
		for _, app := range apps {
			payload = append(payload, &pipeline.BulkPatchPayload{AppId: app.Id})
		}
		bulkPatchRequest.Payload = payload
	}

	bulkPatchRequest.UserId = userId
	if bulkPatchRequest.Global {
		_, err := handler.configMapService.ConfigSecretGlobalBulkPatch(&bulkPatchRequest)
		if err != nil {
			handler.Logger.Errorw("service err, ConfigSecretBulkPatch", "err", err, "payload", bulkPatchRequest)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	} else {
		_, err := handler.configMapService.ConfigSecretEnvironmentBulkPatch(&bulkPatchRequest)
		if err != nil {
			handler.Logger.Errorw("service err, ConfigSecretBulkPatch", "err", err, "payload", bulkPatchRequest)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	writeJsonResp(w, err, true, http.StatusOK)
}
