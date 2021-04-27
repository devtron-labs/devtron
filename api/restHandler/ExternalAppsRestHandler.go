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
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/deploymentGroup"
	"github.com/devtron-labs/devtron/pkg/external"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type ExternalAppsRestHandler interface {
	FindById(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)

	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
}

type ExternalAppsRestHandlerImpl struct {
	application            application.ServiceClient
	appListingService      app.AppListingService
	teamService            team.TeamService
	enforcer               rbac.Enforcer
	pipeline               pipeline.PipelineBuilder
	logger                 *zap.SugaredLogger
	enforcerUtil           rbac.EnforcerUtil
	deploymentGroupService deploymentGroup.DeploymentGroupService
	userAuthService        user.UserService
	externalAppsService    external.ExternalAppsService
}

func NewExternalAppsRestHandlerImpl(application application.ServiceClient,
	appListingService app.AppListingService,
	teamService team.TeamService,
	enforcer rbac.Enforcer,
	pipeline pipeline.PipelineBuilder,
	logger *zap.SugaredLogger, enforcerUtil rbac.EnforcerUtil,
	deploymentGroupService deploymentGroup.DeploymentGroupService, userAuthService user.UserService,
	externalAppsService external.ExternalAppsService) *ExternalAppsRestHandlerImpl {
	appListingHandler := &ExternalAppsRestHandlerImpl{
		application:            application,
		appListingService:      appListingService,
		logger:                 logger,
		teamService:            teamService,
		pipeline:               pipeline,
		enforcer:               enforcer,
		enforcerUtil:           enforcerUtil,
		deploymentGroupService: deploymentGroupService,
		userAuthService:        userAuthService,
		externalAppsService:    externalAppsService,
	}
	return appListingHandler
}

func (handler *ExternalAppsRestHandlerImpl) FindAll(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	res, err := handler.externalAppsService.FindAll()
	if err != nil {
		handler.logger.Errorw("service err, FindAllApps, app store", "err", err, "userId", userId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler *ExternalAppsRestHandlerImpl) FindById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, FindById", "err", err, " id", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, FindById", " id", id)
	res, err := handler.externalAppsService.FindById(id)
	if err != nil {
		handler.logger.Errorw("service err, FindById", "err", err, "userId", userId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler *ExternalAppsRestHandlerImpl) Create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request *external.ExternalAppsDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, Create", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionCreate, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//rback block ends here
	request.UserId = userId
	res, err := handler.externalAppsService.Create(request)
	if err != nil {
		handler.logger.Errorw("service err, Create", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler *ExternalAppsRestHandlerImpl) Update(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request *external.ExternalAppsDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, Update", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionUpdate, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//rback block ends here
	request.UserId = userId
	res, err := handler.externalAppsService.Update(request)
	if err != nil {
		handler.logger.Errorw("service err, Update", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}
