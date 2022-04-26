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

package server

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/server"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type ServerRestHandler interface {
	GetServerInfo(w http.ResponseWriter, r *http.Request)
	HandleServerAction(w http.ResponseWriter, r *http.Request)
}

type ServerRestHandlerImpl struct {
	logger        *zap.SugaredLogger
	serverService server.ServerService
	userService   user.UserService
	enforcer      casbin.Enforcer
	validator     *validator.Validate
}

func NewServerRestHandlerImpl(logger *zap.SugaredLogger,
	serverService server.ServerService,
	userService user.UserService,
	enforcer casbin.Enforcer,
	validator *validator.Validate,
) *ServerRestHandlerImpl {
	return &ServerRestHandlerImpl{
		logger:        logger,
		serverService: serverService,
		userService:   userService,
		enforcer:      enforcer,
		validator:     validator,
	}
}

func (impl ServerRestHandlerImpl) GetServerInfo(w http.ResponseWriter, r *http.Request) {
	// check if user is logged in or not
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// service call
	res, err := impl.serverService.GetServerInfo()
	if err != nil {
		impl.logger.Errorw("service err, GetServerInfo", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl ServerRestHandlerImpl) HandleServerAction(w http.ResponseWriter, r *http.Request) {
	// check if user is logged in or not
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// decode request
	decoder := json.NewDecoder(r.Body)
	var serverActionRequestDto *serverBean.ServerActionRequestDto
	err = decoder.Decode(&serverActionRequestDto)
	if err != nil {
		impl.logger.Errorw("error in decoding request in HandleServerAction", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = impl.validator.Struct(serverActionRequestDto)
	if err != nil {
		impl.logger.Errorw("error in validating request in HandleServerAction", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// service call
	res, err := impl.serverService.HandleServerAction(userId, serverActionRequestDto)
	if err != nil {
		impl.logger.Errorw("service err, HandleServerAction", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
