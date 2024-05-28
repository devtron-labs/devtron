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

package appStore

import (
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type AppStoreStatusTimelineRestHandler interface {
	FetchTimelinesForAppStore(w http.ResponseWriter, r *http.Request)
}

type AppStoreStatusTimelineRestHandlerImpl struct {
	logger                        *zap.SugaredLogger
	pipelineStatusTimelineService status.PipelineStatusTimelineService
	enforcerUtil                  rbac.EnforcerUtil
	enforcer                      casbin.Enforcer
}

func NewAppStoreStatusTimelineRestHandlerImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	enforcerUtil rbac.EnforcerUtil,
	enforcer casbin.Enforcer) *AppStoreStatusTimelineRestHandlerImpl {
	return &AppStoreStatusTimelineRestHandlerImpl{
		logger:                        logger,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
		enforcerUtil:                  enforcerUtil,
		enforcer:                      enforcer,
	}
}

func (handler AppStoreStatusTimelineRestHandlerImpl) FetchTimelinesForAppStore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installedAppId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	installedAppVersionHistoryId := 0
	showTimeline := false
	showTimelineParam := r.URL.Query().Get("showTimeline")
	if len(showTimelineParam) > 0 {
		showTimeline, err = strconv.ParseBool(showTimelineParam)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	installedAppVersionHistoryIdParam := r.URL.Query().Get("installedAppVersionHistoryId")
	if len(installedAppVersionHistoryIdParam) != 0 {
		installedAppVersionHistoryId, err = strconv.Atoi(installedAppVersionHistoryIdParam)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	//rbac will already be handled at app level
	timelines, err := handler.pipelineStatusTimelineService.FetchTimelinesForAppStore(installedAppId, envId, installedAppVersionHistoryId, showTimeline)
	if err != nil {
		handler.logger.Errorw("error in getting pipeline status timelines by installedAppVersionHistoryId", "err", err, "installedAppVersionHistoryId", installedAppVersionHistoryId, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, timelines, http.StatusOK)
	return
}
