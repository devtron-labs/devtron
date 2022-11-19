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

package globalTag

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/globalTag"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type GlobalTagRestHandler interface {
	GetAllActiveTags(w http.ResponseWriter, r *http.Request)
	CreateTags(w http.ResponseWriter, r *http.Request)
	UpdateTags(w http.ResponseWriter, r *http.Request)
	DeleteTags(w http.ResponseWriter, r *http.Request)
}

type GlobalTagRestHandlerImpl struct {
	logger           *zap.SugaredLogger
	userService      user.UserService
	globalTagService globalTag.GlobalTagService
	enforcer         casbin.Enforcer
	validator        *validator.Validate
}

func NewGlobalTagRestHandlerImpl(logger *zap.SugaredLogger, userService user.UserService, globalTagService globalTag.GlobalTagService,
	enforcer casbin.Enforcer, validator *validator.Validate) *GlobalTagRestHandlerImpl {
	return &GlobalTagRestHandlerImpl{
		logger:           logger,
		userService:      userService,
		globalTagService: globalTagService,
		enforcer:         enforcer,
		validator:        validator,
	}
}

func (impl GlobalTagRestHandlerImpl) GetAllActiveTags(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// service call
	res, err := impl.globalTagService.GetAllActiveTags()
	if err != nil {
		impl.logger.Errorw("service err, GetAllActiveTags", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl GlobalTagRestHandlerImpl) CreateTags(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// decode request
	decoder := json.NewDecoder(r.Body)
	var request *globalTag.CreateGlobalTagsRequest
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("err in decoding request in CreateTags", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// validate request
	err = impl.validator.Struct(request)
	if err != nil {
		impl.logger.Errorw("validation err in CreateTags", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// service call
	res, err := impl.globalTagService.CreateTags(request, userId)
	if err != nil {
		impl.logger.Errorw("service err, CreateTags", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl GlobalTagRestHandlerImpl) UpdateTags(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// decode request
	decoder := json.NewDecoder(r.Body)
	var request *globalTag.UpdateGlobalTagsRequest
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("err in decoding request in UpdateTags", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// validate request
	err = impl.validator.Struct(request)
	if err != nil {
		impl.logger.Errorw("validation err in UpdateTags", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// service call
	res, err := impl.globalTagService.UpdateTags(request, userId)
	if err != nil {
		impl.logger.Errorw("service err, UpdateTags", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl GlobalTagRestHandlerImpl) DeleteTags(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// decode request
	decoder := json.NewDecoder(r.Body)
	var request *globalTag.DeleteGlobalTagsRequest
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("err in decoding request in DeleteTags", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// validate request
	err = impl.validator.Struct(request)
	if err != nil {
		impl.logger.Errorw("validation err in DeleteTags", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// service call
	res, err := impl.globalTagService.DeleteTags(request, userId)
	if err != nil {
		impl.logger.Errorw("service err, DeleteTags", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
