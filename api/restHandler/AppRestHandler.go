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
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type AppRestHandlerHandler interface {
	GetAllLabels(w http.ResponseWriter, r *http.Request)
	GetAppMetaInfo(w http.ResponseWriter, r *http.Request)
	UpdateApp(w http.ResponseWriter, r *http.Request)
	UpdateProjectForApps(w http.ResponseWriter, r *http.Request)
}

type AppRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	appService      app.AppCrudOperationService
	userAuthService user.UserService
	validator       *validator.Validate
	enforcerUtil    rbac.EnforcerUtil
	enforcer        casbin.Enforcer
}

func NewAppRestHandlerImpl(logger *zap.SugaredLogger, appService app.AppCrudOperationService,
	userAuthService user.UserService, validator *validator.Validate, enforcerUtil rbac.EnforcerUtil,
	enforcer casbin.Enforcer) *AppRestHandlerImpl {
	handler := &AppRestHandlerImpl{
		logger:          logger,
		appService:      appService,
		userAuthService: userAuthService,
		validator:       validator,
		enforcerUtil:    enforcerUtil,
		enforcer:        enforcer,
	}
	return handler
}

func (handler AppRestHandlerImpl) GetAllLabels(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	results := make([]*bean.AppLabelDto, 0)
	labels, err := handler.appService.FindAll()
	if err != nil {
		handler.logger.Errorw("service err, GetAllLabels", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	objects := handler.enforcerUtil.GetRbacObjectsForAllApps()
	for _, label := range labels {
		object := objects[label.AppId]
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); ok {
			results = append(results, label)
		}
	}
	common.WriteJsonResp(w, nil, results, http.StatusOK)
}

func (handler AppRestHandlerImpl) GetAppMetaInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, GetAppMetaInfo", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//rback implementation starts here
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback implementation ends here

	res, err := handler.appService.GetAppMetaInfo(appId)
	if err != nil {
		handler.logger.Errorw("service err, GetAppMetaInfo", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler AppRestHandlerImpl) UpdateApp(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request bean.CreateAppDTO
	err = decoder.Decode(&request)
	request.UserId = userId
	if err != nil {
		handler.logger.Errorw("request err, UpdateApp", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, UpdateApp", "request", request)

	//rbac implementation starts here
	token := r.Header.Get("token")

	// check for existing project/app permission
	object := handler.enforcerUtil.GetAppRBACNameByAppId(request.Id)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}

	// check for request project/app permission
	object = handler.enforcerUtil.GetAppRBACNameByTeamIdAndAppId(request.TeamId, request.Id)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}

	objects := handler.enforcerUtil.GetEnvRBACArrayByAppId(request.Id)
	for _, object := range objects {
		if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, object); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	//rbac implementation ends here

	res, err := handler.appService.UpdateApp(&request)
	if err != nil {
		handler.logger.Errorw("service err, UpdateApp", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler AppRestHandlerImpl) UpdateProjectForApps(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request bean.UpdateProjectBulkAppsRequest
	err = decoder.Decode(&request)
	request.UserId = userId
	if err != nil {
		handler.logger.Errorw("request err, ProjectChange", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, ProjectChange", "request", request)

	//rbac implementation ends here
	token := r.Header.Get("token")
	for _, appId := range request.AppIds {
		object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, object); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
		objects := handler.enforcerUtil.GetEnvRBACArrayByAppId(appId)
		for _, object := range objects {
			if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, object); !ok {
				common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
				return
			}
		}
	}
	//rbac implementation ends here

	res, err := handler.appService.UpdateProjectForApps(&request)
	if err != nil {
		handler.logger.Errorw("service err, ProjectChange", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}
