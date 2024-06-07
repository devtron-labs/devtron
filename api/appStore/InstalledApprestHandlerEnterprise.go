/*
 * Copyright (c) 2024. Devtron Inc.
 */

package appStore

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode/deployment"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

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

type InstalledAppRestHandlerEnterprise interface {
	InstalledAppRestHandler
	GetChartForLatestDeployment(w http.ResponseWriter, r *http.Request)
	GetChartForParticularTrigger(w http.ResponseWriter, r *http.Request)
}

type InstalledAppRestHandlerEnterpriseImpl struct {
	InstalledAppRestHandler
	Logger                       *zap.SugaredLogger
	userAuthService              user.UserService
	enforcer                     casbin.Enforcer
	enforcerUtil                 rbac.EnforcerUtil
	installedAppService          FullMode.InstalledAppDBExtendedService
	fullModeDeploymentServiceEnt deployment.FullModeDeploymentServiceEnterprise
}

func NewInstalledAppRestHandlerEnterpriseImpl(Logger *zap.SugaredLogger, userAuthService user.UserService,
	enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, installedAppService FullMode.InstalledAppDBExtendedService,
	fullModeDeploymentServiceEnt deployment.FullModeDeploymentServiceEnterprise, InstalledAppRestHandler InstalledAppRestHandler) *InstalledAppRestHandlerEnterpriseImpl {
	return &InstalledAppRestHandlerEnterpriseImpl{
		Logger:                       Logger,
		userAuthService:              userAuthService,
		enforcer:                     enforcer,
		enforcerUtil:                 enforcerUtil,
		installedAppService:          installedAppService,
		fullModeDeploymentServiceEnt: fullModeDeploymentServiceEnt,
		InstalledAppRestHandler:      InstalledAppRestHandler,
	}
}

func (handler *InstalledAppRestHandlerEnterpriseImpl) GetChartForLatestDeployment(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installedAppId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetChartForLatestDeployment", "err", err, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])

	appDetail, err := handler.installedAppService.FindAppDetailsForAppstoreApplication(installedAppId, envId)
	if err != nil {
		handler.Logger.Errorw("request err, GetChartForLatestDeployment", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	object, object2 := handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(appDetail.AppName, appDetail.EnvironmentId)
	var ok bool
	if object2 == "" {
		ok = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object)
	} else {
		ok = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object2)
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	manifestByteArr, err := handler.fullModeDeploymentServiceEnt.GetChartBytesForLatestDeployment(installedAppId, appDetail.AppStoreInstalledAppVersionId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//TODO: move below to custom function
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(manifestByteArr)
	return
}

func (handler *InstalledAppRestHandlerEnterpriseImpl) GetChartForParticularTrigger(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installedAppId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetChartForLatestDeployment", "err", err, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetChartForLatestDeployment", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	installedAppVersionHistoryId, err := strconv.Atoi(vars["installedAppVersionHistoryId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetChartForLatestDeployment", "err", err, "installedAppVersionHistoryId", installedAppVersionHistoryId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	appDetail, err := handler.installedAppService.FindAppDetailsForAppstoreApplication(installedAppId, envId)
	if err != nil {
		handler.Logger.Errorw("request err, GetChartForLatestDeployment", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	object, object2 := handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(appDetail.AppName, appDetail.EnvironmentId)
	var ok bool
	if object2 == "" {
		ok = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object)
	} else {
		ok = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object2)
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	manifestByteArr, err := handler.fullModeDeploymentServiceEnt.GetChartBytesForParticularDeployment(installedAppId, appDetail.AppStoreInstalledAppVersionId, installedAppVersionHistoryId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//TODO: move below to custom function
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(manifestByteArr)
	return
}
