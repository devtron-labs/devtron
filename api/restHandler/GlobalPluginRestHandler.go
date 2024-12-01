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
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/plugin"
	"github.com/devtron-labs/devtron/pkg/plugin/bean"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type GlobalPluginRestHandler interface {
	CreatePlugin(w http.ResponseWriter, r *http.Request)

	GetAllGlobalVariables(w http.ResponseWriter, r *http.Request)
	ListAllPlugins(w http.ResponseWriter, r *http.Request)
	GetPluginDetailById(w http.ResponseWriter, r *http.Request)
	GetDetailedPluginInfoByPluginId(w http.ResponseWriter, r *http.Request)
	GetAllDetailedPluginInfo(w http.ResponseWriter, r *http.Request)

	ListAllPluginsV2(w http.ResponseWriter, r *http.Request)
	GetPluginDetailByIds(w http.ResponseWriter, r *http.Request)
	GetAllUniqueTags(w http.ResponseWriter, r *http.Request)
	MigratePluginData(w http.ResponseWriter, r *http.Request)
	GetAllPluginMinData(w http.ResponseWriter, r *http.Request)
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

// Deprecated: method patchPlugin
// The below API was initially designed to handle the older design of global plugins.
// The API is not yet used in UI.
// The CODE is not yet tested for all the cases.
// TODO: remove this dead code and all the related handling.
func (handler *GlobalPluginRestHandlerImpl) patchPlugin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var pluginDataDto bean.PluginMetadataDto
	err = decoder.Decode(&pluginDataDto)
	if err != nil {
		handler.logger.Errorw("request err, patchPlugin", "error", err, "payload", pluginDataDto)
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
	ok, err := handler.IsUserAuthorisedForThisApp(token, appId)
	if err != nil {
		handler.logger.Infow("service error, ListAllPlugins", "appId", appId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	var plugins []*bean.PluginListComponentDto
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
	ok, err := handler.IsUserAuthorisedForThisApp(token, appId)
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
	appId, err := common.ExtractIntQueryParam(w, r, "appId", 0)
	if err != nil {
		return
	}
	ok, err := handler.IsUserAuthorized(token, appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	listFilter, err := handler.getListFilterFromQueryParam(w, r)
	if err != nil {
		common.WriteJsonResp(w, err, "invalid filter value in query param", http.StatusBadRequest)
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

func (handler *GlobalPluginRestHandlerImpl) GetAllUniqueTags(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	appId, err := common.ExtractIntQueryParam(w, r, "appId", 0)
	if err != nil {
		return
	}
	ok, err := handler.IsUserAuthorized(token, appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	pluginDetail, err := handler.globalPluginService.GetAllUniqueTags()
	if err != nil {
		handler.logger.Errorw("error in getting all unique tags", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, pluginDetail, http.StatusOK)
}

func (handler *GlobalPluginRestHandlerImpl) GetPluginDetailByIds(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	request, err := handler.getPluginDetailsRequestDto(r)
	if err != nil {
		common.WriteJsonResp(w, err, err.Error(), http.StatusBadRequest)
		return
	}

	ok, err := handler.IsUserAuthorized(token, request.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	pluginDetail, err := handler.globalPluginService.GetPluginDetailV2(request)
	if err != nil {
		handler.logger.Errorw("error in getting plugin detail", "request", request, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, pluginDetail, http.StatusOK)

}

func (handler *GlobalPluginRestHandlerImpl) IsUserAuthorisedForThisApp(token string, appId int) (bool, error) {
	var ok bool
	//using appId for rbac in plugin(global resource), because this data must be visible to person having create permission
	//on atleast one app & we can't check this without iterating through every app
	//TODO: update plugin as a resource in casbin and make rbac independent of appId
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	ok = handler.enforcerUtil.CheckAppRbacForAppOrJob(token, resourceName, casbin.ActionCreate)
	return ok, nil
}

func (handler *GlobalPluginRestHandlerImpl) getListFilterFromQueryParam(w http.ResponseWriter, r *http.Request) (*bean.PluginsListFilter, error) {
	v := r.URL.Query()
	offset, err := common.ExtractIntQueryParam(w, r, "offset", 0)
	if err != nil {
		return nil, err
	}

	limit, err := common.ExtractIntQueryParam(w, r, "size", 20)
	if err != nil {
		return nil, err
	}
	searchQueryParam := v.Get("searchKey")
	tagArray := v["tag"]

	fetchAllVersionDetails, err := common.ExtractBoolQueryParam(r, "fetchAllVersionDetails")
	if err != nil {
		return nil, err
	}

	listFilter := bean.NewPluginsListFilter()
	listFilter.WithOffset(offset).WithLimit(limit).WithTags(tagArray).WithSearchKey(searchQueryParam)
	listFilter.FetchAllVersionDetails = fetchAllVersionDetails
	return listFilter, nil
}

func (handler *GlobalPluginRestHandlerImpl) getPluginDetailsRequestDto(r *http.Request) (bean.GlobalPluginDetailsRequest, error) {
	request := bean.GlobalPluginDetailsRequest{}
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(&request)
	if err != nil && err != io.EOF {
		handler.logger.Errorw("request err, CreateOrUpdateGlobalPolicy", "err", err, "payload", request)
		return request, err
	} else if err == io.EOF {
		var schemaDecoder = schema.NewDecoder()
		schemaDecoder.IgnoreUnknownKeys(true)
		err = schemaDecoder.Decode(&request, r.URL.Query())
		if err != nil {
			handler.logger.Errorw("error in parsing query param", "err", err)
			return request, err
		}
		parentPluginIdentifiers := strings.Split(request.ParentPluginIdentifier, ",")
		for _, identifiers := range parentPluginIdentifiers {
			request.ParentPluginIdentifiers = append(request.ParentPluginIdentifiers, strings.TrimSpace(identifiers))
		}
	}
	return request, nil
}

func (handler *GlobalPluginRestHandlerImpl) IsUserAuthorized(token string, appId int) (bool, error) {
	var isAuthorised bool
	var err error
	if appId > 0 {
		isAuthorised, err = handler.IsUserAuthorisedForThisApp(token, appId)
		if err != nil {
			return isAuthorised, err
		}
	} else { //check for super-admin, to be used in global policy
		isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	}
	return isAuthorised, nil
}

func (handler *GlobalPluginRestHandlerImpl) MigratePluginData(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	err := handler.globalPluginService.MigratePluginData()
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

func (handler *GlobalPluginRestHandlerImpl) CreatePlugin(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	appId, err := common.ExtractIntQueryParam(w, r, "appId", 0)
	if err != nil {
		return
	}
	ok, err := handler.IsUserAuthorized(token, appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var pluginDataDto bean.PluginParentMetadataDto
	err = decoder.Decode(&pluginDataDto)
	if err != nil {
		handler.logger.Errorw("request err, CreatePlugin", "error", err, "payload", pluginDataDto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload received for creating plugins", pluginDataDto, "userId", userId)

	pluginVersionId, err := handler.globalPluginService.CreatePluginOrVersions(&pluginDataDto, userId)
	if err != nil {
		handler.logger.Errorw("service error, error in creating plugin", "pluginCreateRequestDto", pluginDataDto, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, bean.NewPluginMinDto().WithPluginVersionId(pluginVersionId), http.StatusOK)
}

func (handler *GlobalPluginRestHandlerImpl) GetAllPluginMinData(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	v := r.URL.Query()
	var schemaDecoder = schema.NewDecoder()
	schemaDecoder.IgnoreUnknownKeys(true)
	queryParams := bean.PluginDetailsMinQuery{}
	err := schemaDecoder.Decode(&queryParams, v)
	if err != nil {
		handler.logger.Errorw("error in parsing query param", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	ok, err := handler.IsUserAuthorized(token, queryParams.AppId)
	if err != nil {
		handler.logger.Errorw("error in verifying rbac", "appId", queryParams.AppId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if !queryParams.IsValidPluginType() {
		common.WriteJsonResp(w, fmt.Errorf("invalid query param 'type'"), "invalid query param 'type'", http.StatusBadRequest)
		return
	}
	pluginDetail, err := handler.globalPluginService.GetAllPluginMinData(queryParams.GetPluginType())
	if err != nil {
		handler.logger.Errorw("error in getting all unique tags", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, pluginDetail, http.StatusOK)
}
