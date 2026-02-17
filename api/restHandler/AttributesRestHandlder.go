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

package restHandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/attributes/bean"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type AttributesRestHandler interface {
	AddAttributes(w http.ResponseWriter, r *http.Request)
	UpdateAttributes(w http.ResponseWriter, r *http.Request)
	GetAttributesById(w http.ResponseWriter, r *http.Request)
	GetAttributesActiveList(w http.ResponseWriter, r *http.Request)
	GetAttributesByKey(w http.ResponseWriter, r *http.Request)
	AddDeploymentEnforcementConfig(w http.ResponseWriter, r *http.Request)
}

type AttributesRestHandlerImpl struct {
	logger            *zap.SugaredLogger
	enforcer          casbin.Enforcer
	userService       user.UserService
	attributesService attributes.AttributesService
}

func NewAttributesRestHandlerImpl(logger *zap.SugaredLogger, enforcer casbin.Enforcer,
	userService user.UserService, attributesService attributes.AttributesService) *AttributesRestHandlerImpl {
	userAuthHandler := &AttributesRestHandlerImpl{
		logger:            logger,
		enforcer:          enforcer,
		userService:       userService,
		attributesService: attributesService,
	}
	return userAuthHandler
}

// isInternalOnlyKey checks if the given key is internal-only and should not be exposed
func (handler AttributesRestHandlerImpl) isInternalOnlyKey(key string) bool {
	return bean.InternalOnlyKeys[key]
}

// filterInternalAttributes removes internal-only attributes from the list
func (handler AttributesRestHandlerImpl) filterInternalAttributes(attributes []*bean.AttributesDto) []*bean.AttributesDto {
	filtered := make([]*bean.AttributesDto, 0, len(attributes))
	return sliceUtil.Filter(filtered, attributes, func(attr *bean.AttributesDto) bool {
		return !handler.isInternalOnlyKey(attr.Key)
	})
}

func (handler AttributesRestHandlerImpl) AddAttributes(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var dto bean.AttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, AddAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// Check if the key is internal-only (not allowed to be created via API)
	if handler.isInternalOnlyKey(dto.Key) {
		handler.logger.Warnw("attempt to create internal-only attribute", "key", dto.Key, "userId", userId)
		common.WriteJsonResp(w, fmt.Errorf("forbidden: cannot create attribute with key: %q", dto.Key), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, AddAttributes", "payload", dto)
	resp, err := handler.attributesService.AddAttributes(&dto)
	if err != nil {
		handler.logger.Errorw("service err, AddAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) UpdateAttributes(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var dto bean.AttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, UpdateAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")

	// Check if the key is internal-only (not allowed to be created via API)
	if handler.isInternalOnlyKey(dto.Key) {
		handler.logger.Warnw("attempt to create internal-only attribute", "key", dto.Key, "userId", userId)
		common.WriteJsonResp(w, fmt.Errorf("forbidden: cannot edit attribute with key: %q", dto.Key), nil, http.StatusForbidden)
		return
	}

	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, UpdateAttributes", "payload", dto)
	resp, err := handler.attributesService.UpdateAttributes(&dto)
	if err != nil {
		handler.logger.Errorw("service err, UpdateAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) GetAttributesById(w http.ResponseWriter, r *http.Request) {
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
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusBadRequest)
		return
	}
	res, err := handler.attributesService.GetById(id)
	if err != nil {
		handler.logger.Errorw("service err, GetAttributesById", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// Filter out internal-only attributes
	if res != nil && handler.isInternalOnlyKey(res.Key) {
		handler.logger.Warnw("attempt to read internal-only attribute", "key", res.Key, "userId", userId)
		common.WriteJsonResp(w, fmt.Errorf("forbidden: cannot read attribute with key: %q", res.Key), nil, http.StatusForbidden)
		return
	}

	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) GetAttributesActiveList(w http.ResponseWriter, r *http.Request) {
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

	res, err := handler.attributesService.GetActiveList()
	if err != nil {
		handler.logger.Errorw("service err, GetHostUrlActive", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// Filter out internal-only attributes from the list
	filteredRes := handler.filterInternalAttributes(res)
	common.WriteJsonResp(w, nil, filteredRes, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) GetAttributesByKey(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	/*token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionGet, "*"); !ok {
		WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}*/

	vars := mux.Vars(r)
	key := vars["key"]

	// Check if the key is internal-only (not allowed to be read via API)
	if handler.isInternalOnlyKey(key) {
		handler.logger.Warnw("attempt to read internal-only attribute by key", "key", key, "userId", userId)
		common.WriteJsonResp(w, fmt.Errorf("forbidden: cannot read attribute with key: %q", key), nil, http.StatusForbidden)
		return
	}

	res, err := handler.attributesService.GetByKey(key)
	if err != nil {
		handler.logger.Errorw("service err, GetAttributesByKey", "key", key, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) AddDeploymentEnforcementConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var dto bean.AttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, AddDeploymentEnforcementConfig", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// Check if the key is enforce deployment type config
	if dto.Key != bean.ENFORCE_DEPLOYMENT_TYPE_CONFIG {
		common.WriteJsonResp(w, fmt.Errorf("invalid key: %q", dto.Key), nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, AddDeploymentEnforcementConfig", "payload", dto)
	resp, err := handler.attributesService.AddDeploymentEnforcementConfig(&dto)
	if err != nil {
		handler.logger.Errorw("service err, AddDeploymentEnforcementConfig", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}
