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

package apiToken

import (
	"encoding/json"
	openapi "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/apiToken"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
)

type ApiTokenRestHandler interface {
	GetAllApiTokens(w http.ResponseWriter, r *http.Request)
	CreateApiToken(w http.ResponseWriter, r *http.Request)
	UpdateApiToken(w http.ResponseWriter, r *http.Request)
	DeleteApiToken(w http.ResponseWriter, r *http.Request)
}

type ApiTokenRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	apiTokenService apiToken.ApiTokenService
	userService     user.UserService
	enforcer        casbin.Enforcer
	validator       *validator.Validate
}

func NewApiTokenRestHandlerImpl(logger *zap.SugaredLogger, apiTokenService apiToken.ApiTokenService, userService user.UserService,
	enforcer casbin.Enforcer, validator *validator.Validate) *ApiTokenRestHandlerImpl {
	return &ApiTokenRestHandlerImpl{
		logger:          logger,
		apiTokenService: apiTokenService,
		userService:     userService,
		enforcer:        enforcer,
		validator:       validator,
	}
}

func (impl ApiTokenRestHandlerImpl) GetAllApiTokens(w http.ResponseWriter, r *http.Request) {
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
	res, err := impl.apiTokenService.GetAllActiveApiTokens()
	if err != nil {
		impl.logger.Errorw("service err, GetAllActiveApiTokens", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl ApiTokenRestHandlerImpl) CreateApiToken(w http.ResponseWriter, r *http.Request) {
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
	var request *openapi.CreateApiTokenRequest
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("err in decoding request in CreateApiToken", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// validate request
	err = impl.validator.Struct(request)
	if err != nil {
		impl.logger.Errorw("validation err in CreateApiToken", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if len(*request.Name) == 0 {
		common.WriteJsonResp(w, errors.New("name cannot be blank in the request"), nil, http.StatusBadRequest)
		return
	}

	// service call
	res, err := impl.apiTokenService.CreateApiToken(request, userId, impl.checkManagerAuth)
	if err != nil {
		impl.logger.Errorw("service err, CreateApiToken", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl ApiTokenRestHandlerImpl) UpdateApiToken(w http.ResponseWriter, r *http.Request) {
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

	// get api-token Id
	vars := mux.Vars(r)
	apiTokenId, err := strconv.Atoi(vars["id"])
	if err != nil {
		impl.logger.Errorw("request err in getting apiTokenId in UpdateApiToken", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// decode request
	decoder := json.NewDecoder(r.Body)
	var request *openapi.UpdateApiTokenRequest
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("err in decoding request, UpdateApiToken", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// validate request
	err = impl.validator.Struct(request)
	if err != nil {
		impl.logger.Errorw("validation err in UpdateApiToken", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	res, err := impl.apiTokenService.UpdateApiToken(apiTokenId, request, userId)
	if err != nil {
		impl.logger.Errorw("service err, UpdateApiToken", "err", err, "apiTokenId", apiTokenId, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl ApiTokenRestHandlerImpl) DeleteApiToken(w http.ResponseWriter, r *http.Request) {
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

	// get api-token Id
	vars := mux.Vars(r)
	apiTokenId, err := strconv.Atoi(vars["id"])
	if err != nil {
		impl.logger.Errorw("request err in getting apiTokenId in DeleteApiToken", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	res, err := impl.apiTokenService.DeleteApiToken(apiTokenId, userId)
	if err != nil {
		impl.logger.Errorw("service err, DeleteApiToken", "err", err, "apiTokenId", apiTokenId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler ApiTokenRestHandlerImpl) checkManagerAuth(token string, object string) bool {
	if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionUpdate, strings.ToLower(object)); !ok {
		return false
	}
	return true
}
