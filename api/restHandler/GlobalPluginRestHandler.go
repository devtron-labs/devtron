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

package restHandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/plugin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type GlobalPluginRestHandler interface {
	PatchPlugin(w http.ResponseWriter, r *http.Request)

	GetAllGlobalVariables(w http.ResponseWriter, r *http.Request)
	ListAllPlugins(w http.ResponseWriter, r *http.Request)
	GetPluginDetailById(w http.ResponseWriter, r *http.Request)
	GetDetailedPluginInfoByPluginId(w http.ResponseWriter, r *http.Request)
	GetAllDetailedPluginInfo(w http.ResponseWriter, r *http.Request)

	ListAllPluginsV2(w http.ResponseWriter, r *http.Request)
	GetPluginDetailByIds(w http.ResponseWriter, r *http.Request)
}

func NewGlobalPluginRestHandler(logger *zap.SugaredLogger, globalPluginService plugin.GlobalPluginService,
	enforcerUtil rbac.EnforcerUtil, enforcer casbin.Enforcer, pipelineBuilder pipeline.PipelineBuilder,
	userService user.UserService) *GlobalPluginRestHandlerImpl {
	return &GlobalPluginRestHandlerImpl{
		logger:              logger,
		globalPluginService: globalPluginService,
		enforcerUtil:        enforcerUtil,
		enforcer:            enforcer,
		pipelineBuilder:     pipelineBuilder,
		userService:         userService,
	}
}

type GlobalPluginRestHandlerImpl struct {
	logger              *zap.SugaredLogger
	globalPluginService plugin.GlobalPluginService
	enforcerUtil        rbac.EnforcerUtil
	enforcer            casbin.Enforcer
	pipelineBuilder     pipeline.PipelineBuilder
	userService         user.UserService
}

func (handler *GlobalPluginRestHandlerImpl) PatchPlugin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var pluginDataDto plugin.PluginMetadataDto
	err = decoder.Decode(&pluginDataDto)
	if err != nil {
		handler.logger.Errorw("request err, PatchPlugin", "error", err, "payload", pluginDataDto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload received for patching plugins", pluginDataDto, "userId", userId)
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized user"), nil, http.StatusForbidden)
		return
	}

	//RBAC enforcer Ends
	pluginData, err := handler.globalPluginService.PatchPlugin(&pluginDataDto, userId)
	if err != nil {
		handler.logger.Errorw("error in patching plugin data", "action", pluginDataDto.Action, "pluginMetadataPayloadDto", pluginDataDto, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, pluginData, http.StatusOK)

}
func (handler *GlobalPluginRestHandlerImpl) GetDetailedPluginInfoByPluginId(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pluginId, err := strconv.Atoi(vars["pluginId"])
	if err != nil {
		handler.logger.Errorw("error in converting from string to integer", "pluginId", vars["pluginId"], "userId", userId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	//RBAC enforcer Ends
	pluginMetaData, err := handler.globalPluginService.GetDetailedPluginInfoByPluginId(pluginId)
	if err != nil {
		handler.logger.Errorw("error in getting plugin metadata", "pluginId", pluginId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, pluginMetaData, http.StatusOK)
}
func (handler *GlobalPluginRestHandlerImpl) GetAllDetailedPluginInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	pluginMetaData, err := handler.globalPluginService.GetAllDetailedPluginInfo()
	if err != nil {
		handler.logger.Errorw("error in getting all plugins metadata", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, pluginMetaData, http.StatusOK)
}

func (handler *GlobalPluginRestHandlerImpl) GetAllGlobalVariables(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	appIdQueryParam := r.URL.Query().Get("appId")
	appId, err := strconv.Atoi(appIdQueryParam)
	if appIdQueryParam == "" || err != nil {
		common.WriteJsonResp(w, err, "invalid appId", http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.logger.Infow("service error, GetAllGlobalVariables", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//using appId for rbac in plugin(global resource), because this data must be visible to person having create permission
	//on atleast one app & we can't check this without iterating through every app
	//TODO: update plugin as a resource in casbin and make rbac independent of appId
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, resourceName, casbin.ActionCreate)
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	globalVariables, err := handler.globalPluginService.GetAllGlobalVariables(app.AppType)
	if err != nil {
		handler.logger.Errorw("error in getting global variable list", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, globalVariables, http.StatusOK)
}

func (handler *GlobalPluginRestHandlerImpl) ListAllPlugins(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	appIdQueryParam := r.URL.Query().Get("appId")
	appId, err := strconv.Atoi(appIdQueryParam)
	if appIdQueryParam == "" || err != nil {
		common.WriteJsonResp(w, err, "invalid appId", http.StatusBadRequest)
		return
	}
	stageType := r.URL.Query().Get("stage")
	ok, err := handler.IsUserAuthorised(token, appId)
	if err != nil {
		handler.logger.Infow("service error, ListAllPlugins", "appId", appId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	var plugins []*plugin.PluginListComponentDto
	plugins, err = handler.globalPluginService.ListAllPlugins(stageType)
	if err != nil {
		handler.logger.Errorw("error in getting cd plugin list", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, plugins, http.StatusOK)
}

func (handler *GlobalPluginRestHandlerImpl) GetPluginDetailById(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	appIdQueryParam := r.URL.Query().Get("appId")
	appId, err := strconv.Atoi(appIdQueryParam)
	if appIdQueryParam == "" || err != nil {
		common.WriteJsonResp(w, err, "invalid appId", http.StatusBadRequest)
		return
	}
	ok, err := handler.IsUserAuthorised(token, appId)
	if err != nil {
		handler.logger.Infow("service error, GetPluginDetailById", "appId", appId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	pluginId, err := strconv.Atoi(vars["pluginId"])
	if err != nil {
		handler.logger.Errorw("received invalid pluginId, GetPluginDetailById", "err", err, "pluginId", pluginId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pluginDetail, err := handler.globalPluginService.GetPluginDetailById(pluginId)
	if err != nil {
		handler.logger.Errorw("error in getting plugin detail by id", "err", err, "pluginId", pluginId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, pluginDetail, http.StatusOK)
}

func (handler *GlobalPluginRestHandlerImpl) ListAllPluginsV2(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	v := r.URL.Query()
	appIdQueryParam := v.Get("appId")
	appId, err := strconv.Atoi(appIdQueryParam)
	if appIdQueryParam == "" || err != nil {
		common.WriteJsonResp(w, err, "invalid appId", http.StatusBadRequest)
		return
	}

	listFilter, err := handler.getListFilterFromQueryParam(w, r)
	if err != nil {
		return
	}

	ok, err := handler.IsUserAuthorised(token, appId)
	if err != nil {
		handler.logger.Infow("service error, ListAllPluginsV2", "appId", appId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	plugins, err := handler.globalPluginService.ListAllPluginsV2(listFilter)
	if err != nil {
		handler.logger.Errorw("error in getting cd plugin list", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, plugins, http.StatusOK)
}

func (handler *GlobalPluginRestHandlerImpl) GetPluginDetailByIds(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	appIdQueryParam := r.URL.Query().Get("appId")
	appId, err := strconv.Atoi(appIdQueryParam)
	if appIdQueryParam == "" || err != nil {
		common.WriteJsonResp(w, err, "invalid appId", http.StatusBadRequest)
		return
	}
	pluginIds, parentPluginIds, fetchLatestVersionDetailsOnly, err := handler.extractAllRequiredQueryParamsForPluginDetail(w, r)
	if err != nil {
		return
	}

	ok, err := handler.IsUserAuthorised(token, appId)
	if err != nil {
		handler.logger.Infow("service error, GetPluginDetailById", "appId", appId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	pluginDetail, err := handler.globalPluginService.GetPluginDetailV2(pluginIds, parentPluginIds, fetchLatestVersionDetailsOnly)
	if err != nil {
		handler.logger.Errorw("error in getting plugin detail", "pluginIds", pluginIds, "parentPluginIds", parentPluginIds, "fetchLatestVersionDetails", fetchLatestVersionDetailsOnly, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, pluginDetail, http.StatusOK)

}

func (handler *GlobalPluginRestHandlerImpl) IsUserAuthorised(token string, appId int) (bool, error) {
	var ok bool
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		return ok, err
	}
	//using appId for rbac in plugin(global resource), because this data must be visible to person having create permission
	//on atleast one app & we can't check this without iterating through every app
	//TODO: update plugin as a resource in casbin and make rbac independent of appId
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	ok = handler.enforcerUtil.CheckAppRbacForAppOrJob(token, resourceName, casbin.ActionCreate)
	if !ok {
		return ok, err
	}
	return ok, nil
}

func (handler *GlobalPluginRestHandlerImpl) getListFilterFromQueryParam(w http.ResponseWriter, r *http.Request) (*plugin.PluginsListFilter, error) {
	v := r.URL.Query()
	offsetDefault := 0
	offset, err := common.ExtractIntQueryParam(w, r, "offset", &offsetDefault)
	if err != nil {
		common.WriteJsonResp(w, err, "invalid offset value", http.StatusBadRequest)
		return nil, err
	}

	limitDefault := 20
	limit, err := common.ExtractIntQueryParam(w, r, "size", &limitDefault)
	if err != nil {
		common.WriteJsonResp(w, err, "invalid size value", http.StatusBadRequest)
		return nil, err
	}
	searchQueryParam := v.Get("searchKey")
	tags := v.Get("tagNames")
	fetchLatestVersionDetails := true
	isLatest := v.Get("fetchLatestVersionDetails")
	if len(isLatest) > 0 {
		fetchLatestVersionDetails, err = strconv.ParseBool(isLatest)
		if err != nil {
			common.WriteJsonResp(w, err, "invalid size value", http.StatusBadRequest)
			return nil, err
		}
	}

	var tagArray []string
	if tags != "" {
		tagArray = strings.Split(tags, ",")
	}
	listFilter := plugin.NewPluginsListFilter()
	listFilter.WithOffset(offset).WithLimit(limit).WithTags(tagArray).WithSearchKey(searchQueryParam)
	listFilter.FetchLatestVersionDetails = fetchLatestVersionDetails
	return listFilter, nil
}

func (handler *GlobalPluginRestHandlerImpl) extractAllRequiredQueryParamsForPluginDetail(w http.ResponseWriter, r *http.Request) ([]int, []int, bool, error) {
	fetchLatestVersionDetailsOnly := true
	pluginIds, err := common.ExtractIntArrayQueryParam(w, r, "pluginId")
	if err != nil {
		common.WriteJsonResp(w, err, "invalid pluginId value", http.StatusBadRequest)
		return nil, nil, fetchLatestVersionDetailsOnly, err
	}
	parentPluginIds, err := common.ExtractIntArrayQueryParam(w, r, "parentPluginId")
	if err != nil {
		common.WriteJsonResp(w, err, "invalid parentPluginId value", http.StatusBadRequest)
		return nil, nil, fetchLatestVersionDetailsOnly, err
	}
	fetchLatestVersionDetailsOnly, err = common.ExtractBoolQueryParam(w, r, "fetchLatestVersionDetails")
	if err != nil {
		common.WriteJsonResp(w, err, "invalid isLatest value", http.StatusBadRequest)
		return nil, nil, fetchLatestVersionDetailsOnly, err
	}
	return pluginIds, parentPluginIds, fetchLatestVersionDetailsOnly, nil
}
