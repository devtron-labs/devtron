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
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
)

type AppGroupRestHandler interface {
	GetActiveAppGroupList(w http.ResponseWriter, r *http.Request)
	GetApplicationsForAppGroup(w http.ResponseWriter, r *http.Request)
	CreateAppGroup(w http.ResponseWriter, r *http.Request)
	UpdateAppGroup(w http.ResponseWriter, r *http.Request)
	DeleteAppGroup(w http.ResponseWriter, r *http.Request)
	CheckAppGroupPermissions(w http.ResponseWriter, r *http.Request)
}

type AppGroupRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	enforcer        casbin.Enforcer
	userService     user.UserService
	appGroupService appGroup.AppGroupService
	validator       *validator.Validate
}

func NewAppGroupRestHandlerImpl(logger *zap.SugaredLogger, enforcer casbin.Enforcer,
	userService user.UserService, appGroupService appGroup.AppGroupService,
	validator *validator.Validate) *AppGroupRestHandlerImpl {
	userAuthHandler := &AppGroupRestHandlerImpl{
		logger:          logger,
		enforcer:        enforcer,
		userService:     userService,
		appGroupService: appGroupService,
		validator:       validator,
	}
	return userAuthHandler
}

func (handler AppGroupRestHandlerImpl) GetActiveAppGroupList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	user, err := handler.userService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	emailId := strings.ToLower(user.EmailId)
	vars := mux.Vars(r)
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := handler.appGroupService.GetActiveAppGroupList(emailId, handler.checkAuthBatch, envId)
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
	var request appGroup.AppGroupDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, CreateAppGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	user, err := handler.userService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	emailId := strings.ToLower(user.EmailId)
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation error", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.EnvironmentId = envId
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation error", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, CreateAppGroup", "payload", request)
	request.CheckAuthBatch = handler.checkAuthBatch
	request.EmailId = emailId
	resp, err := handler.appGroupService.CreateAppGroup(&request)
	if err != nil {
		handler.logger.Errorw("service err, CreateAppGroup", "err", err, "payload", request)
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
	var request appGroup.AppGroupDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, UpdateAppGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	user, err := handler.userService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	emailId := strings.ToLower(user.EmailId)
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation error", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, UpdateAppGroup", "payload", request)
	request.CheckAuthBatch = handler.checkAuthBatch
	request.EmailId = emailId
	resp, err := handler.appGroupService.UpdateAppGroup(&request)
	if err != nil {
		handler.logger.Errorw("service err, UpdateAppGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}
func (handler AppGroupRestHandlerImpl) DeleteAppGroup(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appGroupId, err := strconv.Atoi(vars["appGroupId"])
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusBadRequest)
		return
	}
	user, err := handler.userService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	emailId := strings.ToLower(user.EmailId)
	handler.logger.Infow("request payload, DeleteAppGroup", "appGroupId", appGroupId)
	resp, err := handler.appGroupService.DeleteAppGroup(appGroupId, emailId, handler.checkAuthBatch)
	if err != nil {
		handler.logger.Errorw("service err, DeleteAppGroup", "err", err, "appGroupId", appGroupId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}
func (handler AppGroupRestHandlerImpl) CheckAppGroupPermissions(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request appGroup.AppGroupDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, CreateAppGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	user, err := handler.userService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	emailId := strings.ToLower(user.EmailId)
	vars := mux.Vars(r)
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.EnvironmentId = envId
	handler.logger.Infow("request payload, CreateAppGroup", "payload", request)
	request.CheckAuthBatch = handler.checkAuthBatch
	request.EmailId = emailId
	resp, err := handler.appGroupService.CheckAppGroupPermissions(&request)
	if err != nil {
		handler.logger.Errorw("service err", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler AppGroupRestHandlerImpl) checkAuthBatch(emailId string, appObject []string, action string) map[string]bool {
	var appResult map[string]bool
	if len(appObject) > 0 {
		appResult = handler.enforcer.EnforceByEmailInBatch(emailId, casbin.ResourceApplications, action, appObject)
	}
	return appResult
}
