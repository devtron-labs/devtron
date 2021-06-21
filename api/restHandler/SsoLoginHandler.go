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
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/pkg/sso"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type SsoLoginRestHandler interface {
	CreateSSOLoginConfig(w http.ResponseWriter, r *http.Request)
	UpdateSSOLoginConfig(w http.ResponseWriter, r *http.Request)
	GetAllSSOLoginConfig(w http.ResponseWriter, r *http.Request)
	GetSSOLoginConfig(w http.ResponseWriter, r *http.Request)
	GetSSOLoginConfigByName(w http.ResponseWriter, r *http.Request)
}

type SsoLoginRestHandlerImpl struct {
	userAuthService user.UserAuthService
	validator       *validator.Validate
	logger          *zap.SugaredLogger
	enforcer        rbac.Enforcer
	natsClient      *pubsub.PubSubClient
	userService     user.UserService
	ssoLoginService sso.SSOLoginService
}


func NewSsoLoginRestHandlerImpl(userAuthService user.UserAuthService, validator *validator.Validate,
	logger *zap.SugaredLogger, enforcer rbac.Enforcer, natsClient *pubsub.PubSubClient, userService user.UserService,
	ssoLoginService sso.SSOLoginService) *SsoLoginRestHandlerImpl {
	handler := &SsoLoginRestHandlerImpl{userAuthService: userAuthService, validator: validator, logger: logger,
		enforcer: enforcer, natsClient: natsClient, userService: userService, ssoLoginService: ssoLoginService}
	return handler
}

func (handler SsoLoginRestHandlerImpl) CreateSSOLoginConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var dto bean.SSOLoginDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, CreateSSOLoginConfig", "err", err, "payload", dto)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionCreate, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, CreateSSOLoginConfig", "payload", dto)
	resp, err := handler.ssoLoginService.CreateSSOLogin(&dto)
	if err != nil {
		handler.logger.Errorw("service err, CreateSSOLoginConfig", "err", err, "payload", dto)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, resp, http.StatusOK)
}

func (handler SsoLoginRestHandlerImpl) UpdateSSOLoginConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var dto bean.SSOLoginDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, UpdateSSOLoginConfig", "err", err, "payload", dto)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionUpdate, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, UpdateSSOLoginConfig", "payload", dto)
	resp, err := handler.ssoLoginService.UpdateSSOLogin(&dto)
	if err != nil {
		handler.logger.Errorw("service err, UpdateSSOLoginConfig", "err", err, "payload", dto)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, resp, http.StatusOK)
}

func (handler SsoLoginRestHandlerImpl) GetAllSSOLoginConfig(w http.ResponseWriter, r *http.Request) {
	res, err := handler.ssoLoginService.GetAll()
	if err != nil {
		handler.logger.Errorw("service err, GetAllSSOLoginConfig", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}

func (handler SsoLoginRestHandlerImpl) GetSSOLoginConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	/* #nosec */
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, GetSSOLoginConfig", "err", err, "id", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionGet, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	res, err := handler.ssoLoginService.GetById(int32(id))
	if err != nil {
		handler.logger.Errorw("service err, GetSSOLoginConfig", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}

func (handler SsoLoginRestHandlerImpl) GetSSOLoginConfigByName(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionGet, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]
	res, err := handler.ssoLoginService.GetByName(name)
	if err != nil {
		handler.logger.Errorw("service err, GetSSOLoginConfigByName", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}
