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
	"github.com/devtron-labs/devtron/pkg/externalLink"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"net/http"
	"strconv"
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
}

func NewApiTokenRestHandlerImpl(logger *zap.SugaredLogger, apiTokenService apiToken.ApiTokenService, userService user.UserService, enforcer casbin.Enforcer) *ApiTokenRestHandlerImpl {
	return &ApiTokenRestHandlerImpl{
		logger:          logger,
		apiTokenService: apiTokenService,
		userService:     userService,
		enforcer:        enforcer,
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

	decoder := json.NewDecoder(r.Body)
	var request *openapi.CreateApiTokenRequest
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("request err, SaveLink", "err", err, "payload", beans)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	res, err := impl.externalLinkService.Create(beans, userId)
	if err != nil {
		impl.logger.Errorw("service err, SaveLink", "err", err, "payload", beans)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl ExternalLinkRestHandlerImpl) UpdateExternalLink(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var bean externalLink.ExternalLinkDto
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, Update Link", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId

	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	impl.logger.Infow("request payload, UpdateLink", "err", err, "bean", bean)
	res, err := impl.externalLinkService.Update(&bean)
	if err != nil {
		impl.logger.Errorw("service err, Update Links", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
func (impl ExternalLinkRestHandlerImpl) DeleteExternalLink(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(r)
	id := params["id"]
	idi, err := strconv.Atoi(id)
	if err != nil {
		impl.logger.Errorw("request err, DeleteExternalLink", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	res, err := impl.externalLinkService.DeleteLink(idi, userId)
	if err != nil {
		impl.logger.Errorw("service err, delete Links", "err", err, "id", idi)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
