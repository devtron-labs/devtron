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

package module

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/module"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type ModuleRestHandler interface {
	GetModuleInfo(w http.ResponseWriter, r *http.Request)
	GetModuleConfig(w http.ResponseWriter, r *http.Request)
	HandleModuleAction(w http.ResponseWriter, r *http.Request)
}

type ModuleRestHandlerImpl struct {
	logger        *zap.SugaredLogger
	moduleService module.ModuleService
	userService   user.UserService
	enforcer      casbin.Enforcer
	validator     *validator.Validate
}

func NewModuleRestHandlerImpl(logger *zap.SugaredLogger,
	moduleService module.ModuleService,
	userService user.UserService,
	enforcer casbin.Enforcer,
	validator *validator.Validate,
) *ModuleRestHandlerImpl {
	return &ModuleRestHandlerImpl{
		logger:        logger,
		moduleService: moduleService,
		userService:   userService,
		enforcer:      enforcer,
		validator:     validator,
	}
}

func (impl ModuleRestHandlerImpl) GetModuleConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// check query param
	params := mux.Vars(r)
	moduleName := params["name"]
	if len(moduleName) == 0 {
		impl.logger.Error("module name is not supplied")
		common.WriteJsonResp(w, errors.New("module name is not supplied"), nil, http.StatusBadRequest)
		return
	}

	config, err := impl.moduleService.GetModuleConfig(moduleName)
	if err != nil {
		impl.logger.Errorw("service err, GetModuleConfig", "name", moduleName, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, config, http.StatusOK)
}

func (impl ModuleRestHandlerImpl) GetModuleInfo(w http.ResponseWriter, r *http.Request) {
	// check if user is logged in or not
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// check query param
	params := mux.Vars(r)
	moduleName := params["name"]
	if len(moduleName) == 0 {
		impl.logger.Error("module name is not supplied")
		common.WriteJsonResp(w, errors.New("module name is not supplied"), nil, http.StatusBadRequest)
		return
	}

	// service call
	res, err := impl.moduleService.GetModuleInfo(moduleName)
	if err != nil {
		impl.logger.Errorw("service err, GetModuleInfo", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl ModuleRestHandlerImpl) HandleModuleAction(w http.ResponseWriter, r *http.Request) {
	// check if user is logged in or not
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// check query param
	params := mux.Vars(r)
	moduleName := params["name"]
	if len(moduleName) == 0 {
		impl.logger.Error("module name is not supplied")
		common.WriteJsonResp(w, errors.New("module name is not supplied"), nil, http.StatusBadRequest)
		return
	}

	// decode request
	decoder := json.NewDecoder(r.Body)
	var moduleActionRequestDto *module.ModuleActionRequestDto
	err = decoder.Decode(&moduleActionRequestDto)
	if err != nil {
		impl.logger.Errorw("error in decoding request in HandleModuleAction", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = impl.validator.Struct(moduleActionRequestDto)
	if err != nil {
		impl.logger.Errorw("error in validating request in HandleModuleAction", "err", err)
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
	res, err := impl.moduleService.HandleModuleAction(userId, moduleName, moduleActionRequestDto)
	if err != nil {
		impl.logger.Errorw("service err, HandleModuleAction", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
