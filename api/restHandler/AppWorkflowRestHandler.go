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
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type AppWorkflowRestHandler interface {
	CreateAppWorkflow(w http.ResponseWriter, r *http.Request)
	FindAppWorkflow(w http.ResponseWriter, r *http.Request)
	DeleteAppWorkflow(w http.ResponseWriter, r *http.Request)
}

type AppWorkflowRestHandlerImpl struct {
	Logger             *zap.SugaredLogger
	appWorkflowService appWorkflow.AppWorkflowService
	userAuthService    user.UserService
	teamService        team.TeamService
	enforcer           rbac.Enforcer
	pipelineBuilder    pipeline.PipelineBuilder
	appRepository      pipelineConfig.AppRepository
	enforcerUtil       rbac.EnforcerUtil
}

func NewAppWorkflowRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService, appWorkflowService appWorkflow.AppWorkflowService,
	teamService team.TeamService, enforcer rbac.Enforcer, pipelineBuilder pipeline.PipelineBuilder,
	appRepository pipelineConfig.AppRepository, enforcerUtil rbac.EnforcerUtil) *AppWorkflowRestHandlerImpl {
	return &AppWorkflowRestHandlerImpl{
		Logger:             Logger,
		appWorkflowService: appWorkflowService,
		userAuthService:    userAuthService,
		teamService:        teamService,
		enforcer:           enforcer,
		pipelineBuilder:    pipelineBuilder,
		appRepository:      appRepository,
		enforcerUtil:       enforcerUtil,
	}
}

func (handler AppWorkflowRestHandlerImpl) CreateAppWorkflow(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	handler.Logger.Debugw("request by user", "userId", userId)
	if userId == 0 || err != nil {
		return
	}
	var request appWorkflow.AppWorkflowDto

	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("decode err", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	//rbac block starts from here
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(request.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback block ends here
	request.UserId = userId

	res, err := handler.appWorkflowService.CreateAppWorkflow(request)
	if err != nil {
		handler.Logger.Errorw("error on creating", "err", err)
		common.WriteJsonResp(w, err, []byte("Creation Failed"), http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler AppWorkflowRestHandlerImpl) DeleteAppWorkflow(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	handler.Logger.Debugw("request by user", "userId", userId)
	if userId == 0 || err != nil {
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["app-id"])
	if err != nil {
		handler.Logger.Errorw("bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	appWorkflowId, err := strconv.Atoi(vars["app-wf-id"])
	if err != nil {
		handler.Logger.Errorw("bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	//rbac block starts from here
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionDelete, resourceName); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback block ends here

	err = handler.appWorkflowService.DeleteAppWorkflow(appId, appWorkflowId, userId)
	if err != nil {
		if _, ok := err.(*util.ApiError); ok {
			handler.Logger.Warnw("error on deleting", "err", err)
		} else {
			handler.Logger.Errorw("error on deleting", "err", err)
		}
		common.WriteJsonResp(w, err, []byte("Creation Failed"), http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (impl AppWorkflowRestHandlerImpl) FindAppWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["app-id"])
	if err != nil {
		impl.Logger.Errorw("bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := impl.pipelineBuilder.GetApp(appId)
	if err != nil {
		impl.Logger.Errorw("bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	object := impl.enforcerUtil.GetAppRBACName(app.AppName)
	impl.Logger.Debugw("rbac object for other environment list", "object", object)
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "unauthorized user", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	workflows := make(map[string]interface{})
	workflowsList, err := impl.appWorkflowService.FindAppWorkflows(appId)
	if err != nil {
		impl.Logger.Errorw("error in fetching workflows for app", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	workflows["appId"] = app.Id
	workflows["appName"] = app.AppName
	if len(workflowsList) > 0 {
		workflows["workflows"] = workflowsList
	} else {
		workflows["workflows"] = []appWorkflow.AppWorkflowDto{}
	}
	common.WriteJsonResp(w, err, workflows, http.StatusOK)
}
