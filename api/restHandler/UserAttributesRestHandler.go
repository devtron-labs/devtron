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
	user2 "github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
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
	userAttributesService user2.UserAttributesService
}

func NewUserAttributesRestHandlerImpl(logger *zap.SugaredLogger, enforcer casbin.Enforcer,
	userService user.UserService, userAttributesService user2.UserAttributesService) *UserAttributesRestHandlerImpl {
	userAuthHandler := &UserAttributesRestHandlerImpl{
		logger:                logger,
		enforcer:              enforcer,
		userService:           userService,
		userAttributesService: userAttributesService,
	}
	return userAuthHandler
}
func (handler UserAttributesRestHandlerImpl) AddUserAttributes(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var dto user2.UserAttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, AddUserAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	dto.UserId = userId
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, AddUserAttributes", "payload", dto)
	resp, err := handler.userAttributesService.AddUserAttributes(&dto)
	if err != nil {
		handler.logger.Errorw("service err, AddUserAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler UserAttributesRestHandlerImpl) UpdateUserAttributes(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var dto user2.UserAttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, UpdateUserAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	dto.UserId = userId
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, UpdateUserAttributes", "payload", dto)
	resp, err := handler.userAttributesService.UpdateUserAttributes(&dto)
	if err != nil {
		handler.logger.Errorw("service err, UpdateUserAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler UserAttributesRestHandlerImpl) GetUserAttribute(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var dto user2.UserAttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, GetUserAttribute", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	res, err := handler.userAttributesService.GetUserAttribute(&dto)
	if err != nil {
		handler.logger.Errorw("service err, GetAttributesById", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}
