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
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/infraConfig/adapter"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v0"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	errors2 "github.com/devtron-labs/devtron/pkg/infraConfig/errors"
	"github.com/devtron-labs/devtron/pkg/infraConfig/service"
	"github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
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

	InfraConfigRestHandlerEnt
}

type InfraConfigRestHandlerImpl struct {
	logger              *zap.SugaredLogger
	infraProfileService service.InfraConfigService
	userService         user.UserService
	enforcer            casbin.Enforcer
	validator           *validator.Validate
}

func NewInfraConfigRestHandlerImpl(logger *zap.SugaredLogger, infraProfileService service.InfraConfigService, userService user.UserService, enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, validator *validator.Validate) *InfraConfigRestHandlerImpl {
	return &InfraConfigRestHandlerImpl{
		logger:              logger,
		infraProfileService: infraProfileService,
		userService:         userService,
		enforcer:            enforcer,
		validator:           validator,
	}
}

func (handler *InfraConfigRestHandlerImpl) GetProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	identifier := r.URL.Query().Get("name")
	if len(identifier) == 0 {
		common.WriteJsonResp(w, errors.New(errors2.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	profileName := strings.ToLower(identifier)
	var profile *v1.ProfileBeanDto
	if profileName != v1.GLOBAL_PROFILE_NAME {
		profile, err = handler.infraProfileService.GetProfileByName(profileName)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if errors.Is(err, pg.ErrNoRows) {
				err = errors.New(fmt.Sprintf("profile %s not found", profileName))
				statusCode = http.StatusNotFound
			}
			common.WriteJsonResp(w, err, nil, statusCode)
			return
		}
	}

	defaultProfile, err := handler.infraProfileService.GetProfileByName(v1.GLOBAL_PROFILE_NAME)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, pg.ErrNoRows) {
			err = errors.New(fmt.Sprintf("profile %s not found", v1.GLOBAL_PROFILE_NAME))
			statusCode = http.StatusNotFound
		}
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	if profileName == v1.GLOBAL_PROFILE_NAME {
		profile = defaultProfile
	}
	resp := v1.ProfileResponse{
		Profile: *profile,
	}
	resp.ConfigurationUnits, err = handler.infraProfileService.GetConfigurationUnits()
	if err != nil {
		handler.logger.Errorw("error in getting configuration units", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resp.DefaultConfigurations = defaultProfile.GetConfigurations()
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *InfraConfigRestHandlerImpl) UpdateInfraProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
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
		common.WriteJsonResp(w, errors.New("name is required"), nil, http.StatusBadRequest)
		return
	}
	profileName := strings.ToLower(val)
	if profileName == "" {
		common.WriteJsonResp(w, errors.New(errors2.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	payload := &v1.ProfileBeanDto{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(payload)
	if err != nil {
		handler.logger.Errorw("error in decoding the request payload", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(payload)
	if err != nil {
		err = errors.Wrap(err, errors2.PayloadValidationError)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
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

// Deprecated: GetProfileV0 is deprecated in favour of GetProfile
func (handler *InfraConfigRestHandlerImpl) GetProfileV0(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
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
		common.WriteJsonResp(w, errors.New(errors2.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}

	var profileV0 *v0.ProfileBeanV0
	if profileName != v1.DEFAULT_PROFILE_NAME {
		profileV1, err := handler.infraProfileService.GetProfileByName(profileName)
		profileV0 = adapter.GetV0ProfileBean(profileV1)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if errors.Is(err, pg.ErrNoRows) {
				err = errors.New(fmt.Sprintf("profile %s not found", profileName))
				statusCode = http.StatusNotFound
			}
			common.WriteJsonResp(w, err, nil, statusCode)
			return
		}
	}

	defaultProfileV1, err := handler.infraProfileService.GetProfileByName(v1.GLOBAL_PROFILE_NAME)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, pg.ErrNoRows) {
			err = errors.New(fmt.Sprintf("profile %s not found", v1.GLOBAL_PROFILE_NAME))
			statusCode = http.StatusNotFound
		}
		common.WriteJsonResp(w, err, nil, statusCode)
		return
	}
	defaultProfileV0 := adapter.GetV0ProfileBean(defaultProfileV1)
	if profileName == v1.DEFAULT_PROFILE_NAME {
		profileV0 = defaultProfileV0
	}
	resp := v0.ProfileResponseV0{
		Profile: *profileV0,
	}
	resp.ConfigurationUnits, err = handler.infraProfileService.GetConfigurationUnits()
	if err != nil {
		handler.logger.Errorw("error in getting configuration units", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//returning the default configuration for UI inheriting purpose
	resp.DefaultConfigurations = defaultProfileV0.Configurations
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *InfraConfigRestHandlerImpl) UpdateInfraProfileV0(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
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
		common.WriteJsonResp(w, errors.New(errors2.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	payload := &v0.ProfileBeanV0{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(payload)
	if err != nil {
		handler.logger.Errorw("error in decoding the request payload", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(payload)
	if err != nil {
		err = errors.Wrap(err, errors2.PayloadValidationError)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !util.IsValidProfileNameRequestedV0(profileName, payload.GetName()) {
		common.WriteJsonResp(w, errors.New(errors2.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	if profileName == v1.DEFAULT_PROFILE_NAME && payload.GetName() != v1.DEFAULT_PROFILE_NAME {
		common.WriteJsonResp(w, errors.New(errors2.InvalidProfileName), nil, http.StatusBadRequest)
		return
	}
	payloadV1 := adapter.ConvertToV1ProfileBean(payload)
	err = handler.infraProfileService.UpdateProfileV0(userId, profileName, payloadV1)
	if err != nil {
		handler.logger.Errorw("error in updating profile and configurations", "profileName", profileName, "payLoad", payload, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
