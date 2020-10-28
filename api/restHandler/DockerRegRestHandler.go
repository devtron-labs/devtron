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
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strings"
)

type DockerRegRestHandler interface {
	SaveDockerRegistryConfig(w http.ResponseWriter, r *http.Request)
	GetDockerArtifactStore(w http.ResponseWriter, r *http.Request)
	FetchAllDockerAccounts(w http.ResponseWriter, r *http.Request)
	FetchOneDockerAccounts(w http.ResponseWriter, r *http.Request)
	UpdateDockerRegistryConfig(w http.ResponseWriter, r *http.Request)
	FetchAllDockerRegistryForAutocomplete(w http.ResponseWriter, r *http.Request)
}
type DockerRegRestHandlerImpl struct {
	dockerRegistryConfig pipeline.DockerRegistryConfig
	logger               *zap.SugaredLogger
	gitRegistryConfig    pipeline.GitRegistryConfig
	dbConfigService      pipeline.DbConfigService
	userAuthService      user.UserService
	validator            *validator.Validate
	enforcer             rbac.Enforcer
	teamService          team.TeamService
}

func NewDockerRegRestHandlerImpl(dockerRegistryConfig pipeline.DockerRegistryConfig,
	logger *zap.SugaredLogger,
	gitRegistryConfig pipeline.GitRegistryConfig,
	dbConfigService pipeline.DbConfigService, userAuthService user.UserService,
	validator *validator.Validate, enforcer rbac.Enforcer, teamService team.TeamService) *DockerRegRestHandlerImpl {
	return &DockerRegRestHandlerImpl{
		dockerRegistryConfig: dockerRegistryConfig,
		logger:               logger,
		gitRegistryConfig:    gitRegistryConfig,
		dbConfigService:      dbConfigService,
		userAuthService:      userAuthService,
		validator:            validator,
		enforcer:             enforcer,
		teamService:          teamService,
	}
}

func (impl DockerRegRestHandlerImpl) SaveDockerRegistryConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean pipeline.DockerArtifactStoreBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.User = userId
	impl.logger.Infow("request payload, SaveDockerRegistryConfig", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceDocker, rbac.ActionCreate, "*"); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := impl.dockerRegistryConfig.Create(&bean)
	if err != nil {
		impl.logger.Errorw("service err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)

}

func (impl DockerRegRestHandlerImpl) GetDockerArtifactStore(w http.ResponseWriter, r *http.Request) {
	res, err := impl.dockerRegistryConfig.ListAllActive()
	if err != nil {
		impl.logger.Errorw("service err, GetDockerArtifactStore", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	var result []pipeline.DockerArtifactStoreBean
	for _, item := range res {
		if ok := impl.enforcer.Enforce(token, rbac.ResourceDocker, rbac.ActionGet, strings.ToLower(item.Id)); ok {
			result = append(result, item)
		}
	}
	//RBAC enforcer Ends

	writeJsonResp(w, err, result, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) FetchAllDockerAccounts(w http.ResponseWriter, r *http.Request) {
	res, err := impl.dockerRegistryConfig.FetchAllDockerAccounts()
	if err != nil {
		impl.logger.Errorw("service err, FetchAllDockerAccounts", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	var result []pipeline.DockerArtifactStoreBean
	for _, item := range res {
		if ok := impl.enforcer.Enforce(token, rbac.ResourceDocker, rbac.ActionGet, strings.ToLower(item.Id)); ok {
			result = append(result, item)
		}
	}
	//RBAC enforcer Ends

	writeJsonResp(w, err, result, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) FetchOneDockerAccounts(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	res, err := impl.dockerRegistryConfig.FetchOneDockerAccount(id)
	if err != nil {
		impl.logger.Errorw("service err, FetchOneDockerAccounts", "err", err, "id", id)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceDocker, rbac.ActionGet, strings.ToLower(res.Id)); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) UpdateDockerRegistryConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean pipeline.DockerArtifactStoreBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, UpdateDockerRegistryConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.User = userId
	impl.logger.Infow("request payload, UpdateDockerRegistryConfig", "err", err, "payload", bean)

	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, UpdateDockerRegistryConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceDocker, rbac.ActionUpdate, strings.ToLower(bean.Id)); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := impl.dockerRegistryConfig.Update(&bean)
	if err != nil {
		impl.logger.Errorw("service err, UpdateDockerRegistryConfig", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)

}

func (impl DockerRegRestHandlerImpl) FetchAllDockerRegistryForAutocomplete(w http.ResponseWriter, r *http.Request) {
	res, err := impl.dockerRegistryConfig.ListAllActive()
	if err != nil {
		impl.logger.Errorw("service err, FetchAllDockerRegistryForAutocomplete", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	writeJsonResp(w, err, res, http.StatusOK)
}
