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
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
)

type GitHostRestHandler interface {
	GetGitHosts(w http.ResponseWriter, r *http.Request)
	GetGitHostById(w http.ResponseWriter, r *http.Request)
	CreateGitHost(w http.ResponseWriter, r *http.Request)
}

type GitHostRestHandlerImpl struct {
	logger               *zap.SugaredLogger
	gitHostConfig    	 pipeline.GitHostConfig
	userAuthService      user.UserService
	validator            *validator.Validate
	enforcer             rbac.Enforcer
}

func NewGitHostRestHandlerImpl(logger *zap.SugaredLogger,
	gitHostConfig pipeline.GitHostConfig, userAuthService user.UserService,
	validator *validator.Validate, enforcer rbac.Enforcer) *GitHostRestHandlerImpl {
	return &GitHostRestHandlerImpl{
		logger:               logger,
		gitHostConfig:    gitHostConfig,
		userAuthService: userAuthService,
		validator: validator,
		enforcer: enforcer,
	}
}

func (impl GitHostRestHandlerImpl) GetGitHosts(w http.ResponseWriter, r *http.Request) {

	// check if user is logged in or not
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	res, err := impl.gitHostConfig.GetAll()
	if err != nil {
		impl.logger.Errorw("service err, GetGitHosts", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	var result []pipeline.GitHostRequest
	for _, item := range res {
		if ok := impl.enforcer.Enforce(token, rbac.ResourceGit, rbac.ActionGet, strings.ToLower(item.Name)); ok {
			result = append(result, item)
		}
	}
	//RBAC enforcer Ends

	writeJsonResp(w, err, res, http.StatusOK)
}



// Need to make this call RBAC free as this API is called from the create app flow (configuring ci)
func (impl GitHostRestHandlerImpl) GetGitHostById(w http.ResponseWriter, r *http.Request) {

	// check if user is logged in or not
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	id, err :=  strconv.Atoi(params["id"])

	if err != nil {
		impl.logger.Errorw("service err in parsing Id , GetGitHostById", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	res, err := impl.gitHostConfig.GetById(id)

	if err != nil {
		impl.logger.Errorw("service err, GetGitHostById", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl GitHostRestHandlerImpl) CreateGitHost(w http.ResponseWriter, r *http.Request) {

	// check if user is logged in or not
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)

	var bean pipeline.GitHostRequest
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, CreateGitHost", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	bean.UserId = userId
	impl.logger.Infow("request payload, CreateGitHost", "err", err, "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, CreateGitHost", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceGit, rbac.ActionCreate, strings.ToLower(bean.Name)); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := impl.gitHostConfig.Create(&bean)
	if err != nil {
		impl.logger.Errorw("service err, CreateGitHost", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}
