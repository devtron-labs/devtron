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
	"encoding/json"
	"fmt"
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeployment "github.com/devtron-labs/devtron/pkg/appStore/deployment"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"time"
)

type CommonDeploymentRestHandler interface {
	GetDeploymentHistory(w http.ResponseWriter, r *http.Request)
	GetDeploymentHistoryValues(w http.ResponseWriter, r *http.Request)
	RollbackApplication(w http.ResponseWriter, r *http.Request)
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
	appOfferingMode, err := handler.getAppOfferingMode(vars["installedAppId"], vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, "bad request", http.StatusBadRequest)
	}
	if appOfferingMode == util2.SERVER_MODE_HYPERION {
		handler.helmAppRestHandler.GetDeploymentHistory(w, r)
	} else if appOfferingMode == util2.SERVER_MODE_FULL {
		handler.GetDeploymentHistoryFromDb(w, r)
	} else {
		common.WriteJsonResp(w, err, "bad request", http.StatusBadRequest)
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

func (handler *CommonDeploymentRestHandlerImpl) getAppOfferingMode(installedAppId string, appId string) (string, error) {
	var installedApp *appStoreBean.InstallAppVersionDTO
	var appOfferingMode string
	if len(installedAppId) == 0 && len(appId) > 0 {
		appIdentifier, err := handler.helmAppService.DecodeAppId(appId)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "invalid app id"}
			return appOfferingMode, err
		}
		installedApp, err = handler.appStoreDeploymentServiceC.GetInstalledAppByClusterNamespaceAndName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
		if err != nil && err != pg.ErrNoRows {
			err = &util.ApiError{HttpStatusCode: http.StatusInternalServerError, UserMessage: "unable to find app in database"}
			return appOfferingMode, err
		}
	} else if len(installedAppId) > 0 && len(appId) == 0 {
		installedAppId, err := strconv.Atoi(installedAppId)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "invalid installed app id"}
			return appOfferingMode, err
		}
		installedApp, err = handler.appStoreDeploymentService.GetInstalledApp(installedAppId)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: http.StatusInternalServerError, UserMessage: "unable to find app in database"}
			return appOfferingMode, err
		}
	}
	if installedApp.Id > 0 {
		appOfferingMode = installedApp.AppOfferingMode
	}
	return appOfferingMode, nil
}
func (handler *CommonDeploymentRestHandlerImpl) GetDeploymentHistoryValues(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appOfferingMode, err := handler.getAppOfferingMode(vars["installedAppId"], vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, "bad request", http.StatusBadRequest)
	}
	if appOfferingMode == util2.SERVER_MODE_HYPERION {
		handler.helmAppRestHandler.GetDeploymentDetail(w, r)
	} else if appOfferingMode == util2.SERVER_MODE_FULL {
		handler.GetDeploymentHistoryValuesFromDb(w, r)
	} else {
		common.WriteJsonResp(w, err, "bad request", http.StatusBadRequest)
	}
}

func (handler *CommonDeploymentRestHandlerImpl) GetDeploymentHistoryValuesFromDb(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installedAppId"])
	if err != nil {
		handler.Logger.Errorw("request err", "error", err, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	installedAppVersionHistoryId, err := strconv.Atoi(vars["version"])
	if err != nil {
		handler.Logger.Errorw("request err", "error", err, "installedAppVersionHistoryId", installedAppVersionHistoryId)
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
	response, err := handler.appStoreDeploymentService.GetInstalledAppVersionHistoryValues(installedAppVersionHistoryId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, response, http.StatusOK)
}

func (handler *CommonDeploymentRestHandlerImpl) RollbackApplication(w http.ResponseWriter, r *http.Request) {
	request := &openapi2.RollbackReleaseRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// get installed app which can not be null
	installedApp, err := handler.appStoreDeploymentService.GetInstalledApp(int(request.GetInstalledAppId()))
	if err != nil {
		handler.Logger.Errorw("Error in getting installed app", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	installedApp.UserId = userId
	if installedApp == nil {
		handler.Logger.Errorw("Installed App can not be null", "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//rbac block starts from here
	var rbacObject string
	token := r.Header.Get("token")
	if installedApp.AppOfferingMode == util2.SERVER_MODE_HYPERION {
		rbacObject = handler.enforcerUtilHelm.GetHelmObjectByClusterId(installedApp.ClusterId, installedApp.Namespace, installedApp.AppName)
	} else {
		rbacObject = handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(installedApp.AppName, installedApp.EnvironmentId)
	}
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rbac block ends here

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	success, err := handler.appStoreDeploymentService.RollbackApplication(ctx, request, installedApp, userId)
	if err != nil {
		handler.Logger.Errorw("Error in Rollback release", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	res := &openapi2.RollbackReleaseResponse{
		Success: &success,
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
