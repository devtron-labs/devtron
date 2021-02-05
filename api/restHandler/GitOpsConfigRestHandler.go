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
	"github.com/devtron-labs/devtron/pkg/gitops"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type GitOpsConfigRestHandler interface {
	CreateGitOpsConfig(w http.ResponseWriter, r *http.Request)
	GetAllGitOpsConfig(w http.ResponseWriter, r *http.Request)
	GetGitOpsConfigById(w http.ResponseWriter, r *http.Request)
	UpdateGitOpsConfig(w http.ResponseWriter, r *http.Request)
}

type GitOpsConfigRestHandlerImpl struct {
	logger              *zap.SugaredLogger
	gitOpsConfigService gitops.GitOpsConfigService
	userAuthService     user.UserService
	validator           *validator.Validate
	enforcer            rbac.Enforcer
	teamService         team.TeamService
}

func NewGitOpsConfigRestHandlerImpl(
	logger *zap.SugaredLogger,
	gitOpsConfigService gitops.GitOpsConfigService, userAuthService user.UserService,
	validator *validator.Validate, enforcer rbac.Enforcer, teamService team.TeamService) *GitOpsConfigRestHandlerImpl {
	return &GitOpsConfigRestHandlerImpl{
		logger:              logger,
		gitOpsConfigService: gitOpsConfigService,
		userAuthService:     userAuthService,
		validator:           validator,
		enforcer:            enforcer,
		teamService:         teamService,
	}
}

func (impl GitOpsConfigRestHandlerImpl) CreateGitOpsConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean gitops.GitOpsConfigDto
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, CreateGitOpsConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	impl.logger.Infow("request payload, CreateGitOpsConfig", "err", err, "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, CreateGitOpsConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying

	//RBAC enforcer Ends

	res, err := impl.gitOpsConfigService.CreateGitOpsConfig(&bean)
	if err != nil {
		impl.logger.Errorw("service err, SaveGitRepoConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl GitOpsConfigRestHandlerImpl) UpdateGitOpsConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean gitops.GitOpsConfigDto
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, UpdateGitOpsConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	impl.logger.Infow("request payload, UpdateGitOpsConfig", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, UpdateGitOpsConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying

	//RBAC enforcer Ends

	err = impl.gitOpsConfigService.UpdateGitOpsConfig(&bean)
	if err != nil {
		impl.logger.Errorw("service err, UpdateGitOpsConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, bean, http.StatusOK)
}

func (impl GitOpsConfigRestHandlerImpl) GetGitOpsConfigById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	res, err := impl.gitOpsConfigService.GetGitOpsConfigById(1)
	if err != nil {
		impl.logger.Errorw("service err, GetGitOpsConfigById", "err", err, "id", id)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	//RBAC enforcer Ends

	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl GitOpsConfigRestHandlerImpl) GetAllGitOpsConfig(w http.ResponseWriter, r *http.Request) {
	result, err := impl.gitOpsConfigService.GetAllGitOpsConfig()
	if err != nil {
		impl.logger.Errorw("service err, GetAllGitOpsConfig", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying

	//RBAC enforcer Ends

	writeJsonResp(w, err, result, http.StatusOK)
}
