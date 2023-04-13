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
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/appGroup"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type AppGroupRestHandler interface {
	GetActiveAppGroupList(w http.ResponseWriter, r *http.Request)
	GetApplicationsForAppGroup(w http.ResponseWriter, r *http.Request)
	CreateAppGroup(w http.ResponseWriter, r *http.Request)
	UpdateAppGroup(w http.ResponseWriter, r *http.Request)
}

type AppGroupRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	enforcer        casbin.Enforcer
	userService     user.UserService
	appGroupService appGroup.AppGroupService
}

func NewAppGroupRestHandlerImpl(logger *zap.SugaredLogger, enforcer casbin.Enforcer,
	userService user.UserService, appGroupService appGroup.AppGroupService) *AppGroupRestHandlerImpl {
	userAuthHandler := &AppGroupRestHandlerImpl{
		logger:          logger,
		enforcer:        enforcer,
		userService:     userService,
		appGroupService: appGroupService,
	}
	return userAuthHandler
}

func (handler AppGroupRestHandlerImpl) GetActiveAppGroupList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	res, err := handler.appGroupService.GetActiveAppGroupList()
	if err != nil {
		handler.logger.Errorw("service err, GetActiveAppGroupList", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}
func (handler AppGroupRestHandlerImpl) GetApplicationsForAppGroup(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["appGroupId"])
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusBadRequest)
		return
	}
	res, err := handler.appGroupService.GetApplicationsForAppGroup(id)
	if err != nil {
		handler.logger.Errorw("service err, GetApplicationsForAppGroup", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}
func (handler AppGroupRestHandlerImpl) CreateAppGroup(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var dto appGroup.AppGroupDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, CreateAppGroup", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	dto.UserId=userId
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, CreateAppGroup", "payload", dto)
	resp, err := handler.appGroupService.CreateAppGroup(&dto)
	if err != nil {
		handler.logger.Errorw("service err, CreateAppGroup", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}
func (handler AppGroupRestHandlerImpl) UpdateAppGroup(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var dto appGroup.AppGroupDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, UpdateAppGroup", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	dto.UserId=userId
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, UpdateAppGroup", "payload", dto)
	resp, err := handler.appGroupService.UpdateAppGroup(&dto)
	if err != nil {
		handler.logger.Errorw("service err, UpdateAppGroup", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}
