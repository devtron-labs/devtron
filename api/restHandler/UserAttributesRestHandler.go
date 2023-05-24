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
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type UserAttributesRestHandler interface {
	AddUserAttributes(w http.ResponseWriter, r *http.Request)
	UpdateUserAttributes(w http.ResponseWriter, r *http.Request)
	GetUserAttribute(w http.ResponseWriter, r *http.Request)
}

type UserAttributesRestHandlerImpl struct {
	logger                *zap.SugaredLogger
	enforcer              casbin.Enforcer
	userService           user.UserService
	userAttributesService attributes.UserAttributesService
}

func NewUserAttributesRestHandlerImpl(logger *zap.SugaredLogger, enforcer casbin.Enforcer,
	userService user.UserService, userAttributesService attributes.UserAttributesService) *UserAttributesRestHandlerImpl {
	userAuthHandler := &UserAttributesRestHandlerImpl{
		logger:                logger,
		enforcer:              enforcer,
		userService:           userService,
		userAttributesService: userAttributesService,
	}
	return userAuthHandler
}

func (handler *UserAttributesRestHandlerImpl) AddUserAttributes(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var dto attributes.UserAttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, AddUserAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	dto.UserId = userId
	token := r.Header.Get("token")
	//if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
	//	common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
	//	return
	//}
	emailId, err := handler.userService.GetEmailFromToken(token)
	if err != nil {
		handler.logger.Errorw("request err, UpdateUserAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	dto.EmailId = emailId

	handler.logger.Infow("request payload, AddUserAttributes", "payload", dto)
	resp, err := handler.userAttributesService.AddUserAttributes(&dto)
	if err != nil {
		handler.logger.Errorw("service err, AddUserAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

// @Summary update user attributes
// @version 1.0
// @produce application/json
// @Param payload body attributes.UserAttributesDto true "Input key"
// @Success 200 {object} attributes.UserAttributesDto
// @Router /orchestrator/attributes/user/update [POST]
func (handler *UserAttributesRestHandlerImpl) UpdateUserAttributes(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var dto attributes.UserAttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, UpdateUserAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	dto.UserId = userId
	token := r.Header.Get("token")
	//if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
	//	common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
	//	return
	//}

	emailId, err := handler.userService.GetEmailFromToken(token)
	if err != nil {
		handler.logger.Errorw("request err, UpdateUserAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	dto.EmailId = emailId

	handler.logger.Infow("request payload, UpdateUserAttributes", "payload", dto)
	resp, err := handler.userAttributesService.UpdateUserAttributes(&dto)
	if err != nil {
		handler.logger.Errorw("service err, UpdateUserAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

// @Summary get user attributes
// @version 1.0
// @produce application/json
// @Param name query string true "Input key"
// @Success 200 {object} attributes.UserAttributesDto
// @Router /orchestrator/attributes/user/get [GET]
func (handler *UserAttributesRestHandlerImpl) GetUserAttribute(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	key := vars["key"]
	if key == "" {
		handler.logger.Errorw("request err, GetUserAttribute", "err", err, "key", key)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	//if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
	//	common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
	//	return
	//}

	dto := attributes.UserAttributesDto{}

	emailId, err := handler.userService.GetEmailFromToken(token)
	if err != nil {
		handler.logger.Errorw("request err, UpdateUserAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	dto.EmailId = emailId
	dto.Key = key

	res, err := handler.userAttributesService.GetUserAttribute(&dto)
	if err != nil {
		handler.logger.Errorw("service err, GetAttributesById", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}
