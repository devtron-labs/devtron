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
	"github.com/devtron-labs/devtron/api/restHandler/common"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strings"
)

const GIT_ACCOUNT_DELETE_SUCCESS_RESP = "Git account deleted successfully."

type GitProviderRestHandler interface {
	SaveGitRepoConfig(w http.ResponseWriter, r *http.Request)
	GetGitProviders(w http.ResponseWriter, r *http.Request)
	FetchAllGitProviders(w http.ResponseWriter, r *http.Request)
	FetchOneGitProviders(w http.ResponseWriter, r *http.Request)
	UpdateGitRepoConfig(w http.ResponseWriter, r *http.Request)
	DeleteGitRepoConfig(w http.ResponseWriter, r *http.Request)
}

type GitProviderRestHandlerImpl struct {
	dockerRegistryConfig  pipeline.DockerRegistryConfig
	logger                *zap.SugaredLogger
	gitRegistryConfig     pipeline.GitRegistryConfig
	dbConfigService       pipeline.DbConfigService
	userAuthService       user.UserService
	validator             *validator.Validate
	enforcer              casbin.Enforcer
	teamService           team.TeamService
	deleteServiceFullMode delete2.DeleteServiceFullMode
}

func NewGitProviderRestHandlerImpl(dockerRegistryConfig pipeline.DockerRegistryConfig,
	logger *zap.SugaredLogger,
	gitRegistryConfig pipeline.GitRegistryConfig,
	dbConfigService pipeline.DbConfigService, userAuthService user.UserService,
	validator *validator.Validate, enforcer casbin.Enforcer, teamService team.TeamService,
	deleteServiceFullMode delete2.DeleteServiceFullMode) *GitProviderRestHandlerImpl {
	return &GitProviderRestHandlerImpl{
		dockerRegistryConfig:  dockerRegistryConfig,
		logger:                logger,
		gitRegistryConfig:     gitRegistryConfig,
		dbConfigService:       dbConfigService,
		userAuthService:       userAuthService,
		validator:             validator,
		enforcer:              enforcer,
		teamService:           teamService,
		deleteServiceFullMode: deleteServiceFullMode,
	}
}

func (handler *GitProviderRestHandlerImpl) SaveGitRepoConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean pipeline.GitRegistry
	err = decoder.Decode(&bean)
	if err != nil {
		handler.logger.Errorw("request err, SaveGitRepoConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	handler.logger.Infow("request payload, SaveGitRepoConfig", "err", err, "payload", bean)
	err = handler.validator.Struct(bean)
	if err != nil {
		handler.logger.Errorw("validation err, SaveGitRepoConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGit, casbin.ActionCreate, strings.ToLower(bean.Name)); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := handler.gitRegistryConfig.Create(&bean)
	if err != nil {
		handler.logger.Errorw("service err, SaveGitRepoConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *GitProviderRestHandlerImpl) GetGitProviders(w http.ResponseWriter, r *http.Request) {
	res, err := handler.gitRegistryConfig.GetAll()
	if err != nil {
		handler.logger.Errorw("service err, GetGitProviders", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *GitProviderRestHandlerImpl) FetchAllGitProviders(w http.ResponseWriter, r *http.Request) {
	res, err := handler.gitRegistryConfig.FetchAllGitProviders()
	if err != nil {
		handler.logger.Errorw("service err, FetchAllGitProviders", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	result := make([]pipeline.GitRegistry, 0)
	for _, item := range res {
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGit, casbin.ActionGet, strings.ToLower(item.Name)); ok {
			result = append(result, item)
		}
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (handler *GitProviderRestHandlerImpl) FetchOneGitProviders(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	res, err := handler.gitRegistryConfig.FetchOneGitProvider(id)
	if err != nil {
		handler.logger.Errorw("service err, FetchOneGitProviders", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGit, casbin.ActionGet, strings.ToLower(res.Name)); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *GitProviderRestHandlerImpl) UpdateGitRepoConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean pipeline.GitRegistry
	err = decoder.Decode(&bean)
	if err != nil {
		handler.logger.Errorw("request err, UpdateGitRepoConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	handler.logger.Infow("request payload, UpdateGitRepoConfig", "payload", bean)
	err = handler.validator.Struct(bean)
	if err != nil {
		handler.logger.Errorw("validation err, UpdateGitRepoConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGit, casbin.ActionUpdate, strings.ToLower(bean.Name)); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := handler.gitRegistryConfig.Update(&bean)
	if err != nil {
		handler.logger.Errorw("service err, UpdateGitRepoConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *GitProviderRestHandlerImpl) DeleteGitRepoConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean pipeline.GitRegistry
	err = decoder.Decode(&bean)
	if err != nil {
		handler.logger.Errorw("request err, DeleteGitRepoConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	handler.logger.Infow("request payload, DeleteGitRepoConfig", "payload", bean)
	err = handler.validator.Struct(bean)
	if err != nil {
		handler.logger.Errorw("validation err, DeleteGitRepoConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGit, casbin.ActionCreate, strings.ToLower(bean.Name)); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	err = handler.deleteServiceFullMode.DeleteGitProvider(&bean)
	if err != nil {
		handler.logger.Errorw("error in deleting git account", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, GIT_ACCOUNT_DELETE_SUCCESS_RESP, http.StatusOK)
}
