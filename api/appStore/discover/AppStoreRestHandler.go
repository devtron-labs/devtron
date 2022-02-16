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

package appStoreDiscover

import (
	"github.com/devtron-labs/devtron/api/restHandler/common"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDiscover "github.com/devtron-labs/devtron/pkg/appStore/discover"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type AppStoreRestHandler interface {
	FindAllApps(w http.ResponseWriter, r *http.Request)
	GetChartDetailsForVersion(w http.ResponseWriter, r *http.Request)
	GetChartVersions(w http.ResponseWriter, r *http.Request)
	GetReadme(w http.ResponseWriter, r *http.Request)
	SearchAppStoreChartByName(w http.ResponseWriter, r *http.Request)
}

type AppStoreRestHandlerImpl struct {
	Logger           *zap.SugaredLogger
	appStoreService  appStoreDiscover.AppStoreService
	userAuthService  user.UserService
	enforcer         casbin.Enforcer
}

func NewAppStoreRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService, appStoreService appStoreDiscover.AppStoreService,
	enforcer casbin.Enforcer) *AppStoreRestHandlerImpl {
	return &AppStoreRestHandlerImpl{
		Logger:           Logger,
		appStoreService:  appStoreService,
		userAuthService:  userAuthService,
		enforcer:         enforcer,
	}
}

func (handler *AppStoreRestHandlerImpl) FindAllApps(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	v := r.URL.Query()
	deprecated := false
	deprecatedStr := v.Get("includeDeprecated")
	if len(deprecatedStr) > 0 {
		deprecated, err = strconv.ParseBool(deprecatedStr)
		if err != nil {
			deprecated = false
		}
	}

	var chartRepoIds []int
	chartRepoIdsStr := v.Get("chartRepoId")
	if len(chartRepoIdsStr) > 0 {
		chartRepoIdStrArr := strings.Split(chartRepoIdsStr, ",")
		for _, chartRepoIdStr := range chartRepoIdStrArr {
			chartRepoId, err := strconv.Atoi(chartRepoIdStr)
			if err == nil {
				chartRepoIds = append(chartRepoIds, chartRepoId)
			}
		}
	}
	appStoreName := v.Get("appStoreName")

	offset := 0
	offsetStr := v.Get("offset")
	if len(offsetStr) > 0 {
		offset, _ = strconv.Atoi(offsetStr)
	}
	size := 0
	sizeStr := v.Get("size")
	if len(sizeStr) > 0 {
		size, _ = strconv.Atoi(sizeStr)
	}
	filter := &appStoreBean.AppStoreFilter{IncludeDeprecated: deprecated, ChartRepoId: chartRepoIds, AppStoreName: appStoreName}
	if size > 0 {
		filter.Size = size
		filter.Offset = offset
	}
	handler.Logger.Infow("request payload, FindAllApps, app store", "userId", userId)
	res, err := handler.appStoreService.FindAllApps(filter)
	if err != nil {
		handler.Logger.Errorw("service err, FindAllApps, app store", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) GetChartDetailsForVersion(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.Logger.Errorw("request err, GetChartDetailsForVersion", "err", err, "appStoreApplicationVersionId", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetChartDetailsForVersion, app store", "appStoreApplicationVersionId", id)
	res, err := handler.appStoreService.FindChartDetailsById(id)
	if err != nil {
		handler.Logger.Errorw("service err, GetChartDetailsForVersion, app store", "err", err, "appStoreApplicationVersionId", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) GetChartVersions(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["appStoreId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetChartVersions", "err", err, "appStoreId", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetChartVersions, app store", "appStoreId", id)
	res, err := handler.appStoreService.FindChartVersionsByAppStoreId(id)
	if err != nil {
		handler.Logger.Errorw("service err, GetChartVersions, app store", "err", err, "appStoreId", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) GetReadme(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["appStoreApplicationVersionId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetReadme", "err", err, "appStoreApplicationVersionId", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetReadme, app store", "appStoreApplicationVersionId", id)
	res, err := handler.appStoreService.GetReadMeByAppStoreApplicationVersionId(id)
	if err != nil {
		handler.Logger.Errorw("service err, GetReadme, fetching resource tree", "err", err, "appStoreApplicationVersionId", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) SearchAppStoreChartByName(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	chartName := vars["chartName"]
	handler.Logger.Infow("request payload, SearchAppStoreChartByName, app store", "chartName", chartName)
	res, err := handler.appStoreService.SearchAppStoreChartByName(chartName)
	if err != nil {
		handler.Logger.Errorw("service err, SearchAppStoreChartByName, app store", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}