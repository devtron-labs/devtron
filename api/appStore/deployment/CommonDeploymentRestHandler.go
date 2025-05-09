/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package appStoreDeployment

import (
	"context"
	"encoding/json"
	"fmt"
	service2 "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/bean"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"net/http"
	"strconv"
	"time"

	"github.com/devtron-labs/common-lib/utils/k8sObjectsUtil"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type CommonDeploymentRestHandler interface {
	GetDeploymentHistory(w http.ResponseWriter, r *http.Request)
	GetDeploymentHistoryValues(w http.ResponseWriter, r *http.Request)
	RollbackApplication(w http.ResponseWriter, r *http.Request)
}

type CommonDeploymentRestHandlerImpl struct {
	Logger                    *zap.SugaredLogger
	userAuthService           user.UserService
	enforcer                  casbin.Enforcer
	enforcerUtil              rbac.EnforcerUtil
	enforcerUtilHelm          rbac.EnforcerUtilHelm
	appStoreDeploymentService service.AppStoreDeploymentService
	installedAppService       EAMode.InstalledAppDBService
	validator                 *validator.Validate
	helmAppService            service2.HelmAppService
	attributesService         attributes.AttributesService
}

func NewCommonDeploymentRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService,
	enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, enforcerUtilHelm rbac.EnforcerUtilHelm,
	appStoreDeploymentService service.AppStoreDeploymentService, installedAppService EAMode.InstalledAppDBService,
	validator *validator.Validate, helmAppService service2.HelmAppService,
	attributesService attributes.AttributesService) *CommonDeploymentRestHandlerImpl {
	return &CommonDeploymentRestHandlerImpl{
		Logger:                    Logger,
		userAuthService:           userAuthService,
		enforcer:                  enforcer,
		enforcerUtil:              enforcerUtil,
		enforcerUtilHelm:          enforcerUtilHelm,
		appStoreDeploymentService: appStoreDeploymentService,
		installedAppService:       installedAppService,
		validator:                 validator,
		helmAppService:            helmAppService,
		attributesService:         attributesService,
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
		installedAppDto, err = handler.installedAppService.GetInstalledAppByClusterNamespaceAndName(appIdentifier)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "unable to find app in database"}
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
		installedAppDto, err = handler.installedAppService.GetInstalledAppByInstalledAppId(installedAppId)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "unable to find app in database"}
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
	var rbacObject2 string
	token := r.Header.Get("token")
	if util2.IsHelmApp(appOfferingMode) {
		rbacObject, rbacObject2 = handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(installedAppDto.ClusterId, installedAppDto.Namespace, installedAppDto.AppName)
	} else {
		rbacObject, rbacObject2 = handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(installedAppDto.AppName, installedAppDto.EnvironmentId)
	}

	var ok bool
	if rbacObject2 == "" {
		ok = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject)
	} else {
		ok = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rbac block ends here

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
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
	var rbacObject2 string
	token := r.Header.Get("token")
	if util2.IsHelmApp(appOfferingMode) {
		rbacObject, rbacObject2 = handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(installedAppDto.ClusterId, installedAppDto.Namespace, installedAppDto.AppName)
	} else {
		rbacObject, rbacObject2 = handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(installedAppDto.AppName, installedAppDto.EnvironmentId)
	}

	var ok bool
	if rbacObject2 == "" {
		ok = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject)
	} else {
		ok = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rbac block ends here

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, span := otel.Tracer("orchestrator").Start(ctx, "appStoreDeploymentService.GetDeploymentHistoryInfo")
	res, err := handler.appStoreDeploymentService.GetDeploymentHistoryInfo(ctx, installedAppDto, installedAppVersionHistoryId)
	span.End()
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if util2.IsHelmApp(appOfferingMode) {
		canUpdate := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject)
		if !canUpdate && res != nil && res.Manifest != nil {
			_, span = otel.Tracer("orchestrator").Start(ctx, "k8sObjectsUtil.HideValuesIfSecretForWholeYamlInput")
			modifiedManifest, err := k8sObjectsUtil.HideValuesIfSecretForWholeYamlInput(*res.Manifest)
			span.End()
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
		return
	}
	installedAppDto.UserId = userId
	//rbac block starts from here
	var rbacObject string
	var rbacObject2 string
	token := r.Header.Get("token")
	if util2.IsHelmApp(appOfferingMode) {
		rbacObject, rbacObject2 = handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(installedAppDto.ClusterId, installedAppDto.Namespace, installedAppDto.AppName)
	} else {
		rbacObject, rbacObject2 = handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(installedAppDto.AppName, installedAppDto.EnvironmentId)
	}
	var ok bool
	if rbacObject2 == "" {
		ok = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject)
	} else {
		ok = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject2)
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rbac block ends here
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	if util2.IsBaseStack() || util2.IsHelmApp(appOfferingMode) {
		ctx = context.WithValue(r.Context(), "token", token)
	}

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
	handler.attributesService.UpdateKeyValueByOne(bean.HELM_APP_UPDATE_COUNTER)
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
