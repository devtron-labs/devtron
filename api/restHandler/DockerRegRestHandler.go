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
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"net/http"
	"strings"

	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const REG_DELETE_SUCCESS_RESP = "Container Registry deleted successfully."

type DockerRegRestHandler interface {
	SaveDockerRegistryConfig(w http.ResponseWriter, r *http.Request)
	GetDockerArtifactStore(w http.ResponseWriter, r *http.Request)
	FetchAllDockerAccounts(w http.ResponseWriter, r *http.Request)
	FetchOneDockerAccounts(w http.ResponseWriter, r *http.Request)
	UpdateDockerRegistryConfig(w http.ResponseWriter, r *http.Request)
	FetchAllDockerRegistryForAutocomplete(w http.ResponseWriter, r *http.Request)
	IsDockerRegConfigured(w http.ResponseWriter, r *http.Request)
	DeleteDockerRegistryConfig(w http.ResponseWriter, r *http.Request)
}
type DockerRegRestHandlerImpl struct {
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

const secureWithCert = "secure-with-cert"

func NewDockerRegRestHandlerImpl(dockerRegistryConfig pipeline.DockerRegistryConfig,
	logger *zap.SugaredLogger,
	gitRegistryConfig pipeline.GitRegistryConfig,
	dbConfigService pipeline.DbConfigService, userAuthService user.UserService,
	validator *validator.Validate, enforcer casbin.Enforcer, teamService team.TeamService,
	deleteServiceFullMode delete2.DeleteServiceFullMode) *DockerRegRestHandlerImpl {
	return &DockerRegRestHandlerImpl{
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

func (impl DockerRegRestHandlerImpl) SaveDockerRegistryConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean pipeline.DockerArtifactStoreBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.User = userId
	if (bean.Connection == secureWithCert && bean.Cert == "") || (bean.Connection != secureWithCert && bean.Cert != "") {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	} else {
		impl.logger.Infow("request payload, SaveDockerRegistryConfig", "payload", bean)
		err = impl.validator.Struct(bean)
		if err != nil {
			impl.logger.Errorw("validation err, SaveDockerRegistryConfig", "err", err, "payload", bean)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		// RBAC enforcer applying
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionCreate, "*"); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends

		res, err := impl.dockerRegistryConfig.Create(&bean)
		if err != nil {
			impl.logger.Errorw("service err, SaveDockerRegistryConfig", "err", err, "payload", bean)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}

		common.WriteJsonResp(w, err, res, http.StatusOK)
	}

}

func (impl DockerRegRestHandlerImpl) GetDockerArtifactStore(w http.ResponseWriter, r *http.Request) {
	res, err := impl.dockerRegistryConfig.ListAllActive()
	if err != nil {
		impl.logger.Errorw("service err, GetDockerArtifactStore", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	var result []pipeline.DockerArtifactStoreBean
	for _, item := range res {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionGet, strings.ToLower(item.Id)); ok {
			result = append(result, item)
		}
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) FetchAllDockerAccounts(w http.ResponseWriter, r *http.Request) {
	res, err := impl.dockerRegistryConfig.FetchAllDockerAccounts()
	if err != nil {
		impl.logger.Errorw("service err, FetchAllDockerAccounts", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	var result []pipeline.DockerArtifactStoreBean
	for _, item := range res {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionGet, strings.ToLower(item.Id)); ok {
			result = append(result, item)
		}
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) FetchOneDockerAccounts(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	res, err := impl.dockerRegistryConfig.FetchOneDockerAccount(id)
	if err != nil {
		impl.logger.Errorw("service err, FetchOneDockerAccounts", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionGet, strings.ToLower(res.Id)); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) UpdateDockerRegistryConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean pipeline.DockerArtifactStoreBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, UpdateDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.User = userId
	if (bean.Connection == secureWithCert && bean.Cert == "") || (bean.Connection != secureWithCert && bean.Cert != "") {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	} else {
		impl.logger.Infow("request payload, UpdateDockerRegistryConfig", "err", err, "payload", bean)

		err = impl.validator.Struct(bean)
		if err != nil {
			impl.logger.Errorw("validation err, UpdateDockerRegistryConfig", "err", err, "payload", bean)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		// RBAC enforcer applying
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionUpdate, strings.ToLower(bean.Id)); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends

		res, err := impl.dockerRegistryConfig.Update(&bean)
		if err != nil {
			impl.logger.Errorw("service err, UpdateDockerRegistryConfig", "err", err, "payload", bean)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}

		common.WriteJsonResp(w, err, res, http.StatusOK)
	}

}

func (impl DockerRegRestHandlerImpl) FetchAllDockerRegistryForAutocomplete(w http.ResponseWriter, r *http.Request) {
	res, err := impl.dockerRegistryConfig.ListAllActive()
	if err != nil {
		impl.logger.Errorw("service err, FetchAllDockerRegistryForAutocomplete", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) IsDockerRegConfigured(w http.ResponseWriter, r *http.Request) {
	isConfigured := false
	res, err := impl.dockerRegistryConfig.ListAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("service err, IsDockerRegConfigured", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if len(res) > 0 {
		isConfigured = true
	}

	common.WriteJsonResp(w, err, isConfigured, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) DeleteDockerRegistryConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean pipeline.DockerArtifactStoreBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, DeleteDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.User = userId
	impl.logger.Infow("request payload, DeleteDockerRegistryConfig", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, DeleteDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionCreate, strings.ToLower(bean.Id)); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	err = impl.deleteServiceFullMode.DeleteDockerRegistryConfig(&bean)
	if err != nil {
		impl.logger.Errorw("service err, DeleteDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, REG_DELETE_SUCCESS_RESP, http.StatusOK)
}
