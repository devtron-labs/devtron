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

package infraConfig

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/infraConfig/adapter"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/service"
	util2 "github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strings"
)

type InfraConfigRestHandler interface {
	GetProfile(w http.ResponseWriter, r *http.Request)
	UpdateInfraProfile(w http.ResponseWriter, r *http.Request)

	// Deprecated: GetProfileV0 is deprecated in favour of GetProfile
	GetProfileV0(w http.ResponseWriter, r *http.Request)
	// Deprecated: UpdateInfraProfileV0 is deprecated in favour of UpdateInfraProfile
	UpdateInfraProfileV0(w http.ResponseWriter, r *http.Request)
}
type InfraConfigRestHandlerImpl struct {
	logger              *zap.SugaredLogger
	infraProfileService service.InfraConfigService
	userService         user.UserService
	enforcer            casbin.Enforcer
	enforcerUtil        rbac.EnforcerUtil
	validator           *validator.Validate
}

func NewInfraConfigRestHandlerImpl(logger *zap.SugaredLogger, infraProfileService service.InfraConfigService, userService user.UserService, enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, validator *validator.Validate) *InfraConfigRestHandlerImpl {
	return &InfraConfigRestHandlerImpl{
		logger:              logger,
		infraProfileService: infraProfileService,
		userService:         userService,
		enforcer:            enforcer,
		enforcerUtil:        enforcerUtil,
		validator:           validator,
	}
}

func (handler *InfraConfigRestHandlerImpl) GetProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	identifier := r.URL.Query().Get("name")

	if len(identifier) == 0 {
		common.WriteJsonResp(w, errors.New(bean.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	profileName := strings.ToLower(identifier)
	var profile *bean.ProfileBeanDto
	if profileName != bean.GLOBAL_PROFILE_NAME {
		common.WriteJsonResp(w, errors.New(bean.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	defaultProfile, err := handler.infraProfileService.GetProfileByName(profileName)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	profile = defaultProfile

	resp := bean.ProfileResponse{
		Profile: *profile,
	}
	resp.ConfigurationUnits = handler.infraProfileService.GetConfigurationUnits()
	//TODO: why below line ??
	resp.DefaultConfigurations = defaultProfile.Configurations
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *InfraConfigRestHandlerImpl) UpdateInfraProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	//vars := mux.Vars(r)
	val := r.URL.Query().Get("name")
	if len(val) == 0 {
		common.WriteJsonResp(w, errors.New(bean.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	profileName := strings.ToLower(val)

	payload := &bean.ProfileBeanDto{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(payload)
	if err != nil {
		handler.logger.Errorw("error in decoding the request payload", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//handler.validator.RegisterValidation("validateValue", ValidateValue)
	payload.Name = strings.ToLower(payload.Name)
	err = handler.validator.Struct(payload)
	if err != nil {
		err = errors.Wrap(err, bean.PayloadValidationError)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	}
	if !util2.IsValidProfileNameRequested(profileName, payload.Name) {
		common.WriteJsonResp(w, errors.New(bean.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	err = handler.infraProfileService.UpdateProfile(userId, profileName, payload)
	if err != nil {
		handler.logger.Errorw("error in updating profile and configurations", "profileName", profileName, "payLoad", payload, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

// Deprecated
func (handler *InfraConfigRestHandlerImpl) GetProfileV0(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	profileName := strings.ToLower(vars["name"])
	if profileName == "" {
		common.WriteJsonResp(w, errors.New(bean.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}

	if profileName != bean.DEFAULT_PROFILE_NAME {
		common.WriteJsonResp(w, errors.New(bean.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}

	profileName = bean.GLOBAL_PROFILE_NAME

	var profile *bean.ProfileBeanV0
	defaultProfileV1, err := handler.infraProfileService.GetProfileByName(profileName)
	defaultProfileV0 := adapter.GetV0ProfileBean(defaultProfileV1)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	profile = defaultProfileV0
	resp := bean.ProfileResponseV0{
		Profile: *profile,
	}
	resp.ConfigurationUnits = handler.infraProfileService.GetConfigurationUnits()
	resp.DefaultConfigurations = defaultProfileV0.Configurations
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

// Deprecated
func (handler *InfraConfigRestHandlerImpl) UpdateInfraProfileV0(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	profileName := strings.ToLower(vars["name"])
	if profileName == "" {
		common.WriteJsonResp(w, errors.New(bean.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	payload := &bean.ProfileBeanV0{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(payload)
	if err != nil {
		handler.logger.Errorw("error in decoding the request payload", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	payload.Name = strings.ToLower(payload.Name)
	err = handler.validator.Struct(payload)
	if err != nil {
		err = errors.Wrap(err, bean.PayloadValidationError)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	}
	if !util2.IsValidProfileNameRequestedV0(profileName, payload.Name) {
		common.WriteJsonResp(w, errors.New(bean.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	if payload.Name != bean.DEFAULT_PROFILE_NAME {
		common.WriteJsonResp(w, errors.New(bean.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}

	profileName = bean.GLOBAL_PROFILE_NAME
	payloadV1 := adapter.GetV1ProfileBean(payload)
	err = handler.infraProfileService.UpdateProfile(userId, profileName, payloadV1)
	if err != nil {
		handler.logger.Errorw("error in updating profile and configurations", "profileName", profileName, "payLoad", payload, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
