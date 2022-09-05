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
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/k8sObjectsUtil"
	"github.com/devtron-labs/devtron/util/rbac"
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
	appStoreDeploymentService  service.AppStoreDeploymentService
	appStoreDeploymentServiceC appStoreDeploymentCommon.AppStoreDeploymentCommonService
	validator                  *validator.Validate
	helmAppService             client.HelmAppService
	helmAppRestHandler         client.HelmAppRestHandler
}

func NewCommonDeploymentRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService,
	enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, enforcerUtilHelm rbac.EnforcerUtilHelm, appStoreDeploymentService service.AppStoreDeploymentService,
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
func (handler *CommonDeploymentRestHandlerImpl) getAppOfferingMode(installedAppId string, appId string) (string, *appStoreBean.InstallAppVersionDTO, error) {
	installedAppDto := &appStoreBean.InstallAppVersionDTO{}
	var appOfferingMode string
	if len(appId) > 0 {
		appIdentifier, err := handler.helmAppService.DecodeAppId(appId)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "invalid app id"}
			return appOfferingMode, installedAppDto, err
		}
		installedAppDto, err = handler.appStoreDeploymentServiceC.GetInstalledAppByClusterNamespaceAndName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: http.StatusInternalServerError, UserMessage: "unable to find app in database"}
			return appOfferingMode, installedAppDto, err
		}
		// this is the case when hyperion apps does not linked yet
		if installedAppDto == nil {
			installedAppDto = &appStoreBean.InstallAppVersionDTO{}
			appOfferingMode = util2.SERVER_MODE_HYPERION
			installedAppDto.InstalledAppId = 0
			installedAppDto.AppOfferingMode = appOfferingMode
			appIdentifier, err := handler.helmAppService.DecodeAppId(appId)
			if err != nil {
				err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "invalid app id"}
				return appOfferingMode, installedAppDto, err
			}
			installedAppDto.ClusterId = appIdentifier.ClusterId
			installedAppDto.Namespace = appIdentifier.Namespace
			installedAppDto.AppName = appIdentifier.ReleaseName
		}
	} else if len(installedAppId) > 0 {
		installedAppId, err := strconv.Atoi(installedAppId)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "invalid installed app id"}
			return appOfferingMode, installedAppDto, err
		}
		installedAppDto, err = handler.appStoreDeploymentServiceC.GetInstalledAppByInstalledAppId(installedAppId)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: http.StatusInternalServerError, UserMessage: "unable to find app in database"}
			return appOfferingMode, installedAppDto, err
		}
	} else {
		err := &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "app id missing in request"}
		return appOfferingMode, installedAppDto, err
	}
	if installedAppDto != nil && installedAppDto.InstalledAppId > 0 {
		appOfferingMode = installedAppDto.AppOfferingMode
	}
	return appOfferingMode, installedAppDto, nil
}

func (handler *CommonDeploymentRestHandlerImpl) GetDeploymentHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	v := r.URL.Query()
	installedAppId := v.Get("installedAppId")
	appId := v.Get("appId")
	appOfferingMode, installedAppDto, err := handler.getAppOfferingMode(installedAppId, appId)
	if err != nil {
		common.WriteJsonResp(w, err, "bad request", http.StatusBadRequest)
		return
	}
	installedAppDto.UserId = userId
	//rbac block starts from here
	var rbacObject string
	token := r.Header.Get("token")
	if util2.IsHelmApp(appOfferingMode) {
		rbacObject = handler.enforcerUtilHelm.GetHelmObjectByClusterId(installedAppDto.ClusterId, installedAppDto.Namespace, installedAppDto.AppName)
	} else {
		rbacObject = handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(installedAppDto.AppName, installedAppDto.EnvironmentId)
	}
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rbac block ends here

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := handler.appStoreDeploymentService.GetDeploymentHistory(ctx, installedAppDto)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *CommonDeploymentRestHandlerImpl) GetDeploymentHistoryValues(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	v := r.URL.Query()
	installedAppId := v.Get("installedAppId")
	appId := v.Get("appId")
	appOfferingMode, installedAppDto, err := handler.getAppOfferingMode(installedAppId, appId)
	if err != nil {
		common.WriteJsonResp(w, err, "bad request", http.StatusBadRequest)
		return
	}
	installedAppDto.UserId = userId
	installedAppVersionHistoryId, err := strconv.Atoi(vars["version"])
	if err != nil {
		handler.Logger.Errorw("request err", "error", err, "installedAppVersionHistoryId", installedAppVersionHistoryId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//rbac block starts from here
	var rbacObject string
	token := r.Header.Get("token")
	if util2.IsHelmApp(appOfferingMode) {
		rbacObject = handler.enforcerUtilHelm.GetHelmObjectByClusterId(installedAppDto.ClusterId, installedAppDto.Namespace, installedAppDto.AppName)
	} else {
		rbacObject = handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(installedAppDto.AppName, installedAppDto.EnvironmentId)
	}
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rbac block ends here

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := handler.appStoreDeploymentService.GetDeploymentHistoryInfo(ctx, installedAppDto, installedAppVersionHistoryId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if util2.IsHelmApp(appOfferingMode) {
		canUpdate := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject)
		if !canUpdate && res != nil && res.Manifest != nil {
			modifiedManifest, err := k8sObjectsUtil.HideValuesIfSecretForWholeYamlInput(*res.Manifest)
			if err != nil {
				handler.Logger.Errorw("error in hiding secret values", "err", err)
				common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}
			res.Manifest = &modifiedManifest
		}
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *CommonDeploymentRestHandlerImpl) RollbackApplication(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	request := &openapi2.RollbackReleaseRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	installedAppId := ""
	if request.GetInstalledAppId() > 0 {
		installedAppId = fmt.Sprint(request.GetInstalledAppId())
	}
	appOfferingMode, installedAppDto, err := handler.getAppOfferingMode(installedAppId, *request.HAppId)
	if err != nil {
		common.WriteJsonResp(w, err, "bad request", http.StatusBadRequest)
	}
	installedAppDto.UserId = userId
	//rbac block starts from here
	var rbacObject string
	token := r.Header.Get("token")
	if util2.IsHelmApp(appOfferingMode) {
		rbacObject = handler.enforcerUtilHelm.GetHelmObjectByClusterId(installedAppDto.ClusterId, installedAppDto.Namespace, installedAppDto.AppName)
	} else {
		rbacObject = handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(installedAppDto.AppName, installedAppDto.EnvironmentId)
	}
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rbac block ends here

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	success, err := handler.appStoreDeploymentService.RollbackApplication(ctx, request, installedAppDto, userId)
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
