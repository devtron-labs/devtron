/*
 * Copyright (c) 2024. Devtron Inc.
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

package globalConfig

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	auth "github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util/commonEnforcementFunctionsUtil"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type AuthorisationConfigRestHandler interface {
	CreateOrUpdateAuthorisationConfig(w http.ResponseWriter, r *http.Request)
	GetAllActiveAuthorisationConfig(w http.ResponseWriter, r *http.Request)
}

type AuthorisationConfigRestHandlerImpl struct {
	validator                        *validator.Validate
	logger                           *zap.SugaredLogger
	enforcer                         casbin.Enforcer
	userService                      user.UserService
	userCommonService                user.UserCommonService
	globalAuthorisationConfigService auth.GlobalAuthorisationConfigService
	rbacEnforcementUtil              commonEnforcementFunctionsUtil.CommonEnforcementUtil
}

func NewGlobalAuthorisationConfigRestHandlerImpl(validator *validator.Validate,
	logger *zap.SugaredLogger, enforcer casbin.Enforcer,
	userService user.UserService,
	globalAuthorisationConfigService auth.GlobalAuthorisationConfigService,
	userCommonService user.UserCommonService,
	rbacEnforcementUtil commonEnforcementFunctionsUtil.CommonEnforcementUtil,
) *AuthorisationConfigRestHandlerImpl {
	return &AuthorisationConfigRestHandlerImpl{
		validator:                        validator,
		logger:                           logger,
		enforcer:                         enforcer,
		userService:                      userService,
		globalAuthorisationConfigService: globalAuthorisationConfigService,
		userCommonService:                userCommonService,
		rbacEnforcementUtil:              rbacEnforcementUtil,
	}
}

func (handler *AuthorisationConfigRestHandlerImpl) CreateOrUpdateAuthorisationConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var globalConfigPayload bean.GlobalAuthorisationConfig
	err = decoder.Decode(&globalConfigPayload)
	if err != nil {
		handler.logger.Errorw("request err, CreateOrUpdateAuthorisationConfig", "err", err, "payload", globalConfigPayload)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	isValidationError, err := handler.validateGlobalAuthorisationConfigPayload(globalConfigPayload)
	if err != nil {
		handler.logger.Errorw("error, validateGlobalAuthorisationConfigPayload", "payload", globalConfigPayload, "err", err)
		if isValidationError {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	globalConfigPayload.UserId = userId
	resp, err := handler.globalAuthorisationConfigService.CreateOrUpdateGlobalAuthConfig(globalConfigPayload, nil)
	if err != nil {
		handler.logger.Errorw("service error, CreateOrUpdateAuthorisationConfig", "err", err, "payload", globalConfigPayload)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *AuthorisationConfigRestHandlerImpl) GetAllActiveAuthorisationConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
	token := r.Header.Get("token")
	isAuthorised, err := handler.rbacEnforcementUtil.CheckRbacForMangerAndAboveAccess(token, userId)
	if err != nil {
		handler.logger.Errorw("error, GetAllActiveAuthorisationConfig", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	resp, err := handler.globalAuthorisationConfigService.GetAllActiveAuthorisationConfig()
	if err != nil {
		handler.logger.Errorw("service error, GetAllActiveAuthorisationConfig", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *AuthorisationConfigRestHandlerImpl) validateGlobalAuthorisationConfigPayload(globalConfigPayload bean.GlobalAuthorisationConfig) (bool, error) {
	err := handler.validator.Struct(globalConfigPayload)
	if err != nil {
		handler.logger.Errorw("err, validateGlobalAuthorisationConfigPayload", "payload", globalConfigPayload, "err", err)
		return true, err
	}
	if len(globalConfigPayload.ConfigTypes) == 0 {
		handler.logger.Errorw("err, validation failed on validateGlobalAuthorisationConfigPayload due to no configType provided", "payload", globalConfigPayload)
		return true, errors.New("no configTypes provided in request")
	}
	return false, nil
}
