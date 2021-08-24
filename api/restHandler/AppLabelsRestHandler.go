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
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type AppLabelRestHandler interface {
	UpdateLabelsInApp(w http.ResponseWriter, r *http.Request)
	GetAllLabels(w http.ResponseWriter, r *http.Request)
	GetAppMetaInfo(w http.ResponseWriter, r *http.Request)
}

type AppLabelRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	appLabelService app.AppLabelService
	userAuthService user.UserService
	validator       *validator.Validate
	enforcerUtil    rbac.EnforcerUtil
	enforcer        rbac.Enforcer
}

func NewAppLabelRestHandlerImpl(logger *zap.SugaredLogger, appLabelService app.AppLabelService,
	userAuthService user.UserService, validator *validator.Validate, enforcerUtil rbac.EnforcerUtil,
	enforcer rbac.Enforcer) *AppLabelRestHandlerImpl {
	handler := &AppLabelRestHandlerImpl{
		logger:          logger,
		appLabelService: appLabelService,
		userAuthService: userAuthService,
		validator:       validator,
		enforcerUtil:    enforcerUtil,
		enforcer:        enforcer,
	}
	return handler
}

func (handler AppLabelRestHandlerImpl) GetAllLabels(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	results := make([]*bean.AppLabelDto, 0)
	labels, err := handler.appLabelService.FindAll()
	if err != nil {
		handler.logger.Errorw("service err, GetAllLabels", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	objects := handler.enforcerUtil.GetRbacObjectsForAllApps()
	for _, label := range labels {
		object := objects[label.AppId]
		if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); ok {
			results = append(results, label)
		}
	}
	writeJsonResp(w, nil, results, http.StatusOK)
}

func (handler AppLabelRestHandlerImpl) GetAppMetaInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, GetAppMetaInfo", "err", err, "appId", appId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//rback implementation starts here
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback implementation ends here

	res, err := handler.appLabelService.GetAppMetaInfo(appId)
	if err != nil {
		handler.logger.Errorw("service err, GetAppMetaInfo", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}

func (handler AppLabelRestHandlerImpl) UpdateLabelsInApp(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request bean.AppLabelsDto
	err = decoder.Decode(&request)
	request.UserId = userId
	if err != nil {
		handler.logger.Errorw("request err, UpdateLabelsInApp", "err", err, "request", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, UpdateLabelsInApp", "request", request)
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err, UpdateLabelsInApp", "err", err, "request", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//rback implementation starts here
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(request.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionUpdate, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback implementation ends here
	res, err := handler.appLabelService.UpdateLabelsInApp(&request)
	if err != nil {
		handler.logger.Errorw("service err, UpdateLabelsInApp", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}
