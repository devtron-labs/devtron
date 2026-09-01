/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package apiToken

import (
	"encoding/json"
	"fmt"
	openapi "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/apiToken"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
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
	GetAllApiTokensForWebhook(w http.ResponseWriter, r *http.Request)
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
		common.HandleUnauthorized(w, r)
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
		common.HandleUnauthorized(w, r)
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
		common.WriteJsonResp(w, errors.New("invalid JSON payload "+err.Error()), nil, http.StatusBadRequest)
		return
	}

	// validate request structure
	err = impl.validator.Struct(request)
	if err != nil {
		impl.logger.Errorw("validation err in CreateApiToken ", "err", err, "request", request)
		common.HandleValidationErrors(w, r, err)
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
		common.HandleUnauthorized(w, r)
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
		common.HandleUnauthorized(w, r)
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

func (handler ApiTokenRestHandlerImpl) checkManagerAuth(resource, token, object string) bool {
	if ok := handler.enforcer.Enforce(token, resource, casbin.ActionUpdate, object); !ok {
		return false
	}
	return true
}

func (impl ApiTokenRestHandlerImpl) GetAllApiTokensForWebhook(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	v := r.URL.Query()
	projectName := v.Get("projectName")
	environmentName := v.Get("environmentName")
	appName := v.Get("appName")

	// handle RBAC - verify that the REQUESTING user (not the stored api-tokens
	// being evaluated below) has trigger permission on the requested
	// project/environment/app before any token data is looked up and returned.
	callerToken := r.Header.Get("token")
	projectObject := fmt.Sprintf("%s/%s", projectName, appName)
	for _, environment := range strings.Split(environmentName, ",") {
		envObject := fmt.Sprintf("%s/%s", environment, appName)
		if !impl.CheckAuthorizationForWebhook(callerToken, projectObject, envObject) {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	}

	// per-token authorization: a stored api-token's own Casbin policy always
	// satisfies the project/env check above if that token is a super-admin
	// ("*" matches any object), regardless of how narrow the caller's own
	// access is. So a stored token is only eligible to be returned if, on top
	// of its own project/env check, it is not more privileged than the caller
	// - i.e. its scope must be no broader than the requesting user's own.
	authForToken := func(storedToken string, projObj string, envObj string) bool {
		if !impl.CheckAuthorizationForWebhook(storedToken, projObj, envObj) {
			return false
		}
		return impl.callerDominatesToken(callerToken, storedToken)
	}

	// service call
	res, err := impl.apiTokenService.GetAllApiTokensForWebhook(projectName, environmentName, appName, authForToken)
	if err != nil {
		impl.logger.Errorw("service err, GetAllApiTokensForWebhook", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler ApiTokenRestHandlerImpl) CheckAuthorizationForWebhook(token string, projectObject string, envObject string) bool {
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, projectObject); !ok {
		return false
	}
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, envObject); !ok {
		return false
	}
	return true
}

// callerDominatesToken returns false when storedToken has strictly broader
// access than callerToken, so that a caller can never see an api-token more
// privileged than themselves through this endpoint. A super-admin token's
// policy is "*", so it always passes the per-object checks in
// CheckAuthorizationForWebhook irrespective of the project/env queried; the
// only way to tell it apart from a token that is genuinely scoped to the
// requested object is to check super-admin status directly, using the same
// check the other handlers in this file use to gate super-admin-only actions.
func (handler ApiTokenRestHandlerImpl) callerDominatesToken(callerToken string, storedToken string) bool {
	if !handler.enforcer.Enforce(storedToken, casbin.ResourceGlobal, casbin.ActionUpdate, "*") {
		// stored token is not a super-admin token, nothing further to check
		return true
	}
	return handler.enforcer.Enforce(callerToken, casbin.ResourceGlobal, casbin.ActionUpdate, "*")
}
