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
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/gitops"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type GitOpsConfigRestHandler interface {
	CreateGitOpsConfig(w http.ResponseWriter, r *http.Request)
	GetAllGitOpsConfig(w http.ResponseWriter, r *http.Request)
	GetGitOpsConfigById(w http.ResponseWriter, r *http.Request)
	UpdateGitOpsConfig(w http.ResponseWriter, r *http.Request)
	GetGitOpsConfigByProvider(w http.ResponseWriter, r *http.Request)
	GitOpsConfigured(w http.ResponseWriter, r *http.Request)
	GitOpsValidator(w http.ResponseWriter, r *http.Request)
}

type GitOpsConfigRestHandlerImpl struct {
	logger              *zap.SugaredLogger
	gitOpsConfigService gitops.GitOpsConfigService
	userAuthService     user.UserService
	validator           *validator.Validate
	enforcer            rbac.Enforcer
	teamService         team.TeamService
	gitOpsRepository    repository.GitOpsConfigRepository
}

func NewGitOpsConfigRestHandlerImpl(
	logger *zap.SugaredLogger,
	gitOpsConfigService gitops.GitOpsConfigService, userAuthService user.UserService,
	validator *validator.Validate, enforcer rbac.Enforcer, teamService team.TeamService, gitOpsRepository repository.GitOpsConfigRepository) *GitOpsConfigRestHandlerImpl {
	return &GitOpsConfigRestHandlerImpl{
		logger:              logger,
		gitOpsConfigService: gitOpsConfigService,
		userAuthService:     userAuthService,
		validator:           validator,
		enforcer:            enforcer,
		teamService:         teamService,
		gitOpsRepository:    gitOpsRepository,
	}
}

func (impl GitOpsConfigRestHandlerImpl) CreateGitOpsConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionCreate, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	var bean bean2.GitOpsConfigDto
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
	detailedErrorGitOpsConfigResponse, err := impl.gitOpsConfigService.ValidateAndCreateGitOpsConfig(&bean)
	if err != nil {
		impl.logger.Errorw("service err, SaveGitRepoConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	writeJsonResp(w, nil, detailedErrorGitOpsConfigResponse, http.StatusOK)

}

func (impl GitOpsConfigRestHandlerImpl) UpdateGitOpsConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionUpdate, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	var bean bean2.GitOpsConfigDto
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
	detailedErrorGitOpsConfigResponse, err := impl.gitOpsConfigService.ValidateAndUpdateGitOpsConfig(&bean)
	if err != nil {
		impl.logger.Errorw("service err, UpdateGitOpsConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	writeJsonResp(w, nil, detailedErrorGitOpsConfigResponse, http.StatusOK)

}

func (impl GitOpsConfigRestHandlerImpl) GetGitOpsConfigById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		impl.logger.Errorw("request err, GetGitOpsConfigById", "err", err, "chart repo id", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := impl.gitOpsConfigService.GetGitOpsConfigById(id)
	if err != nil {
		impl.logger.Errorw("service err, GetGitOpsConfigById", "err", err, "id", id)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionGet, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	// RBAC enforcer Ends

	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl GitOpsConfigRestHandlerImpl) GitOpsConfigured(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	result, err := impl.gitOpsConfigService.GetAllGitOpsConfig()
	if err != nil {
		impl.logger.Errorw("service err, GetAllGitOpsConfig", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	gitopsConfigured := false
	if len(result) > 0 {
		for _, gitopsConf := range result {
			if gitopsConf.Active {
				gitopsConfigured = true
			}
		}
	}
	res := make(map[string]bool)
	res["exists"] = gitopsConfigured
	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl GitOpsConfigRestHandlerImpl) GetAllGitOpsConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	result, err := impl.gitOpsConfigService.GetAllGitOpsConfig()
	if err != nil {
		impl.logger.Errorw("service err, GetAllGitOpsConfig", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionGet, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	writeJsonResp(w, err, result, http.StatusOK)
}

func (impl GitOpsConfigRestHandlerImpl) GetGitOpsConfigByProvider(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	provider := vars["provider"]
	res, err := impl.gitOpsConfigService.GetGitOpsConfigByProvider(provider)
	if err != nil {
		impl.logger.Errorw("service err, GetGitOpsConfigByProvider", "err", err, "provider", provider)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying

	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionGet, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	// RBAC enforcer Ends

	writeJsonResp(w, err, res, http.StatusOK)
}
func (impl GitOpsConfigRestHandlerImpl) GitOpsValidator(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionCreate, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	var bean bean2.GitOpsConfigDto
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, ValidateGitOpsConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	impl.logger.Infow("request payload, ValidateGitOpsConfig", "err", err, "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, ValidateGitOpsConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	detailedErrorGitOpsConfigResponse := impl.gitOpsConfigService.GitOpsValidateDryRun(&bean)
	writeJsonResp(w, nil, detailedErrorGitOpsConfigResponse, http.StatusOK)
}
