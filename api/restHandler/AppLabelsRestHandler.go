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
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type AppLabelsRestHandler interface {
	EditAppLabels(w http.ResponseWriter, r *http.Request)
	GetAllActiveLabels(w http.ResponseWriter, r *http.Request)
	GetAppMetaInfo(w http.ResponseWriter, r *http.Request)
}

type AppLabelsRestHandlerImpl struct {
	logger           *zap.SugaredLogger
	appLabelsService app.AppLabelsService
	userAuthService  user.UserService
	validator        *validator.Validate
}

func NewAppTagRestHandlerImpl(logger *zap.SugaredLogger, appLabelsService app.AppLabelsService,
	userAuthService user.UserService, validator *validator.Validate) *AppLabelsRestHandlerImpl {
	handler := &AppLabelsRestHandlerImpl{
		logger:           logger,
		appLabelsService: appLabelsService,
		userAuthService:  userAuthService,
		validator:        validator,
	}
	return handler
}

func (handler AppLabelsRestHandlerImpl) GetAllActiveLabels(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	res, err := handler.appLabelsService.FindAllActive()
	if err != nil {
		handler.logger.Errorw("service err, GetAllActiveLabels", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}

func (handler AppLabelsRestHandlerImpl) GetAppMetaInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, GetAppMetaInfo", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := handler.appLabelsService.GetAppMetaInfo(appId)
	if err != nil {
		handler.logger.Errorw("service err, GetAppMetaInfo", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}

func (handler AppLabelsRestHandlerImpl) EditAppLabels(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var createRequest app.AppLabelsCreateRequest
	err = decoder.Decode(&createRequest)
	createRequest.UserId = userId
	if err != nil {
		handler.logger.Errorw("request err, EditAppLabels", "err", err, "create request", createRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, EditAppLabels", "create request", createRequest)
	err = handler.validator.Struct(createRequest)
	if err != nil {
		handler.logger.Errorw("validation err, EditAppLabels", "err", err, "create request", createRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := handler.appLabelsService.EditAppLabels(&createRequest)
	if err != nil {
		handler.logger.Errorw("service err, EditAppLabels", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}
