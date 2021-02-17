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
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type AttributesRestHandler interface {
	AddAttributes(w http.ResponseWriter, r *http.Request)
	UpdateAttributes(w http.ResponseWriter, r *http.Request)
	GetAttributesById(w http.ResponseWriter, r *http.Request)
	GetAttributesActiveList(w http.ResponseWriter, r *http.Request)
	GetAttributesByKey(w http.ResponseWriter, r *http.Request)
}

type AttributesRestHandlerImpl struct {
	logger            *zap.SugaredLogger
	enforcer          rbac.Enforcer
	userService       user.UserService
	attributesService attributes.AttributesService
}

func NewAttributesRestHandlerImpl(logger *zap.SugaredLogger, enforcer rbac.Enforcer,
	userService user.UserService, attributesService attributes.AttributesService) *AttributesRestHandlerImpl {
	userAuthHandler := &AttributesRestHandlerImpl{
		logger:            logger,
		enforcer:          enforcer,
		userService:       userService,
		attributesService: attributesService,
	}
	return userAuthHandler
}
func (handler AttributesRestHandlerImpl) AddAttributes(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var dto attributes.AttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, AddAttributes", "err", err, "payload", dto)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	/*isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if err != nil {
		handler.logger.Errorw("request err, AddAttributes", "err", err, "userId", userId)
		writeJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}

	if !isActionUserSuperAdmin {
		err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to perform operation"}
		writeJsonResp(w, err, "", http.StatusForbidden)
		return
	}*/
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionCreate, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, AddAttributes", "payload", dto)
	resp, err := handler.attributesService.AddAttributes(&dto)
	if err != nil {
		handler.logger.Errorw("service err, AddAttributes", "err", err, "payload", dto)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, resp, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) UpdateAttributes(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var dto attributes.AttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, UpdateAttributes", "err", err, "payload", dto)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	/*isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if err != nil {
		handler.logger.Errorw("request err, UpdateAttributes", "err", err, "userId", userId)
		writeJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}

	if !isActionUserSuperAdmin {
		err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to perform operation"}
		writeJsonResp(w, err, "", http.StatusForbidden)
		return
	}*/
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionUpdate, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, UpdateAttributes", "payload", dto)
	resp, err := handler.attributesService.UpdateAttributes(&dto)
	if err != nil {
		handler.logger.Errorw("service err, UpdateAttributes", "err", err, "payload", dto)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, resp, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) GetAttributesById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	/*isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if err != nil {
		handler.logger.Errorw("request err, GetAttributesById", "err", err, "userId", userId)
		writeJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}

	if !isActionUserSuperAdmin {
		err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to perform operation"}
		writeJsonResp(w, err, "", http.StatusForbidden)
		return
	}*/

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionGet, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeJsonResp(w, err, "", http.StatusBadRequest)
		return
	}
	res, err := handler.attributesService.GetById(id)
	if err != nil {
		handler.logger.Errorw("service err, GetAttributesById", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) GetAttributesActiveList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	/*isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if err != nil {
		handler.logger.Errorw("request err, GetHostUrlActive", "err", err, "userId", userId)
		writeJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}

	if !isActionUserSuperAdmin {
		err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to perform operation"}
		writeJsonResp(w, err, "", http.StatusForbidden)
		return
	}*/

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionGet, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	res, err := handler.attributesService.GetActiveList()
	if err != nil {
		handler.logger.Errorw("service err, GetHostUrlActive", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) GetAttributesByKey(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	/*isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if err != nil {
		handler.logger.Errorw("request err, GetAttributesById", "err", err, "userId", userId)
		writeJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}

	if !isActionUserSuperAdmin {
		err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to perform operation"}
		writeJsonResp(w, err, "", http.StatusForbidden)
		return
	}*/

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionGet, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	key := vars["key"]
	res, err := handler.attributesService.GetByKey(key)
	if err != nil {
		handler.logger.Errorw("service err, GetAttributesById", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}
