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

package appStoreDeployment

import (
	"context"
	"errors"
	"fmt"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	appStoreDeployment "github.com/devtron-labs/devtron/pkg/appStore/deployment"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type CommonDeploymentRestHandler interface {
	GetDeploymentHistory(w http.ResponseWriter, r *http.Request)
}

type CommonDeploymentRestHandlerImpl struct {
	Logger                     *zap.SugaredLogger
	userAuthService            user.UserService
	enforcer                   casbin.Enforcer
	enforcerUtil               rbac.EnforcerUtil
	enforcerUtilHelm           rbac.EnforcerUtilHelm
	appStoreDeploymentService  appStoreDeployment.AppStoreDeploymentService
	appStoreDeploymentServiceC appStoreDeploymentCommon.AppStoreDeploymentCommonService
	validator                  *validator.Validate
	helmAppService             client.HelmAppService
	helmAppRestHandler         client.HelmAppRestHandler
}

func NewCommonDeploymentRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService,
	enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, enforcerUtilHelm rbac.EnforcerUtilHelm, appStoreDeploymentService appStoreDeployment.AppStoreDeploymentService,
	validator *validator.Validate, helmAppService client.HelmAppService, appStoreDeploymentServiceC appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	helmAppRestHandler client.HelmAppRestHandler) *CommonDeploymentRestHandlerImpl {
	return &CommonDeploymentRestHandlerImpl{
		Logger:                     Logger,
		userAuthService:            userAuthService,
		enforcer:                   enforcer,
		enforcerUtil:               enforcerUtil,
		enforcerUtilHelm:           enforcerUtilHelm,
		appStoreDeploymentService:  appStoreDeploymentService,
		validator:                  validator,
		helmAppService:             helmAppService,
		appStoreDeploymentServiceC: appStoreDeploymentServiceC,
		helmAppRestHandler:         helmAppRestHandler,
	}
}

func (handler *CommonDeploymentRestHandlerImpl) GetDeploymentHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mode := vars["mode"]
	if mode == util2.SERVER_MODE_HYPERION {
		//handler.GetDeploymentHistoryFromHelm(w, r)
		handler.helmAppRestHandler.GetDeploymentHistory(w, r)
	} else if mode == util2.SERVER_MODE_FULL {
		handler.GetDeploymentHistoryFromDb(w, r)
	}
}

func (handler *CommonDeploymentRestHandlerImpl) GetDeploymentHistoryFromDb(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installedAppId"])
	if err != nil {
		handler.Logger.Errorw("request err", "error", err, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	installedApp, err := handler.appStoreDeploymentService.GetInstalledApp(installedAppId)
	if err != nil {
		handler.Logger.Errorw("service err, GetDeploymentHistoryValues", "err", err, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//rbac block starts from here
	object := handler.enforcerUtil.GetHelmObject(installedApp.AppId, installedApp.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rback block ends here
	response, err := handler.appStoreDeploymentService.GetInstalledAppVersionHistory(installedAppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, response, http.StatusOK)
}
func (handler *CommonDeploymentRestHandlerImpl) GetDeploymentHistoryFromHelm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId := vars["appId"]
	appIdentifier, err := handler.helmAppService.DecodeAppId(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject := handler.enforcerUtilHelm.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	deploymentHistory, err := handler.helmAppService.GetDeploymentHistory(context.Background(), appIdentifier)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	installedApp, err := handler.appStoreDeploymentServiceC.GetInstalledAppByClusterNamespaceAndName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	res := &client.DeploymentHistoryAndInstalledAppInfo{
		DeploymentHistory: deploymentHistory.GetDeploymentHistory(),
		InstalledAppInfo:  convertToInstalledAppInfo(installedApp),
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
