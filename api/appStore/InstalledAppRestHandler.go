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

package appStore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	bean2 "github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/cron"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/middleware"
	util2 "github.com/devtron-labs/devtron/internal/util"
	app2 "github.com/devtron-labs/devtron/pkg/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/cluster"
	application2 "github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/devtron-labs/devtron/util/response"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type InstalledAppRestHandler interface {
	FetchAppOverview(w http.ResponseWriter, r *http.Request)
	GetAllInstalledApp(w http.ResponseWriter, r *http.Request)
	DeployBulk(w http.ResponseWriter, r *http.Request)
	CheckAppExists(w http.ResponseWriter, r *http.Request)
	DefaultComponentInstallation(w http.ResponseWriter, r *http.Request)
	FetchAppDetailsForInstalledApp(w http.ResponseWriter, r *http.Request)
	DeleteArgoInstalledAppWithNonCascade(w http.ResponseWriter, r *http.Request)
	FetchAppDetailsForInstalledAppV2(w http.ResponseWriter, r *http.Request)
	FetchResourceTree(w http.ResponseWriter, r *http.Request)
	FetchResourceTreeForACDApp(w http.ResponseWriter, r *http.Request)
	FetchNotesForArgoInstalledApp(w http.ResponseWriter, r *http.Request)
}

type InstalledAppRestHandlerImpl struct {
	Logger                           *zap.SugaredLogger
	userAuthService                  user.UserService
	enforcer                         casbin.Enforcer
	enforcerUtil                     rbac.EnforcerUtil
	enforcerUtilHelm                 rbac.EnforcerUtilHelm
	installedAppService              service.InstalledAppService
	validator                        *validator.Validate
	clusterService                   cluster.ClusterService
	acdServiceClient                 application.ServiceClient
	appStoreDeploymentService        service.AppStoreDeploymentService
	helmAppClient                    client.HelmAppClient
	argoUserService                  argo.ArgoUserService
	cdApplicationStatusUpdateHandler cron.CdApplicationStatusUpdateHandler
	installedAppRepository           repository.InstalledAppRepository
	K8sApplicationService            application2.K8sApplicationService
	appCrudOperationService          app2.AppCrudOperationService
}

func NewInstalledAppRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService,
	enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, enforcerUtilHelm rbac.EnforcerUtilHelm, installedAppService service.InstalledAppService,
	validator *validator.Validate, clusterService cluster.ClusterService, acdServiceClient application.ServiceClient,
	appStoreDeploymentService service.AppStoreDeploymentService, helmAppClient client.HelmAppClient,
	argoUserService argo.ArgoUserService,
	cdApplicationStatusUpdateHandler cron.CdApplicationStatusUpdateHandler,
	installedAppRepository repository.InstalledAppRepository,
	appCrudOperationService app2.AppCrudOperationService) *InstalledAppRestHandlerImpl {
	return &InstalledAppRestHandlerImpl{
		Logger:                           Logger,
		userAuthService:                  userAuthService,
		enforcer:                         enforcer,
		enforcerUtil:                     enforcerUtil,
		enforcerUtilHelm:                 enforcerUtilHelm,
		installedAppService:              installedAppService,
		validator:                        validator,
		clusterService:                   clusterService,
		acdServiceClient:                 acdServiceClient,
		appStoreDeploymentService:        appStoreDeploymentService,
		helmAppClient:                    helmAppClient,
		argoUserService:                  argoUserService,
		cdApplicationStatusUpdateHandler: cdApplicationStatusUpdateHandler,
		installedAppRepository:           installedAppRepository,
		appCrudOperationService:          appCrudOperationService,
	}
}
func (handler *InstalledAppRestHandlerImpl) FetchAppOverview(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installedAppId"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchAppOverview", "err", err, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	handler.Logger.Infow("request payload, FindAppOverview", "installedAppId", installedAppId)
	installedApp, err := handler.installedAppService.CheckAppExistsByInstalledAppId(installedAppId)
	appOverview, err := handler.appCrudOperationService.GetAppMetaInfo(installedApp.AppId, installedAppId, installedApp.EnvironmentId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchAppOverview", "err", err, "appId", installedApp.AppId, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//rbac block starts from here
	object, object2 := handler.enforcerUtil.GetHelmObject(appOverview.AppId, installedApp.EnvironmentId)
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
	common.WriteJsonResp(w, nil, appOverview, http.StatusOK)
}

func (handler InstalledAppRestHandlerImpl) GetAllInstalledApp(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	v := r.URL.Query()
	token := r.Header.Get("token")
	var envIds []int
	envsQueryParam := v.Get("envIds")
	if envsQueryParam != "" {
		envsStr := strings.Split(envsQueryParam, ",")
		for _, t := range envsStr {
			envId, err := strconv.Atoi(t)
			if err != nil {
				handler.Logger.Errorw("request err, GetAllInstalledApp", "err", err, "envsQueryParam", envsQueryParam)
				response.WriteResponse(http.StatusBadRequest, "please send valid envs", w, errors.New("env id invalid"))
				return
			}
			envIds = append(envIds, envId)
		}
	}
	clusterIdString := v.Get("clusterIds")
	var clusterIds []int
	if clusterIdString != "" {
		clusterIdSlices := strings.Split(clusterIdString, ",")
		for _, clusterId := range clusterIdSlices {
			id, err := strconv.Atoi(clusterId)
			if err != nil {
				handler.Logger.Errorw("request err, GetAllInstalledApp", "err", err, "clusterIdString", clusterIdString)
				common.WriteJsonResp(w, err, "please send valid cluster Ids", http.StatusBadRequest)
				return
			}
			clusterIds = append(clusterIds, id)
		}
	}
	onlyDeprecated := false
	deprecatedStr := v.Get("onlyDeprecated")
	if len(deprecatedStr) > 0 {
		onlyDeprecated, err = strconv.ParseBool(deprecatedStr)
		if err != nil {
			onlyDeprecated = false
		}
	}

	var chartRepoIds []int
	chartRepoIdsStr := v.Get("chartRepoIds")
	if len(chartRepoIdsStr) > 0 {
		chartRepoIdStrArr := strings.Split(chartRepoIdsStr, ",")
		for _, chartRepoIdStr := range chartRepoIdStrArr {
			chartRepoId, err := strconv.Atoi(chartRepoIdStr)
			if err == nil {
				chartRepoIds = append(chartRepoIds, chartRepoId)
			}
		}
	}

	var appStatuses []string
	appStatusesStr := v.Get("appStatuses")
	if len(appStatusesStr) > 0 {
		appStatuses = strings.Split(appStatusesStr, ",")
	}
	appStoreName := v.Get("appStoreName")
	appName := v.Get("appName")
	offset := 0
	offsetStr := v.Get("offset")
	if len(offsetStr) > 0 {
		offset, _ = strconv.Atoi(offsetStr)
	}
	size := 0
	sizeStr := v.Get("size")
	if len(sizeStr) > 0 {
		size, _ = strconv.Atoi(sizeStr)
	}
	filter := &appStoreBean.AppStoreFilter{
		OnlyDeprecated: onlyDeprecated,
		ChartRepoId:    chartRepoIds,
		AppStoreName:   appStoreName,
		EnvIds:         envIds,
		AppName:        appName,
		ClusterIds:     clusterIds,
		AppStatuses:    appStatuses,
	}
	if size > 0 {
		filter.Size = size
		filter.Offset = offset
	}
	handler.Logger.Infow("request payload, GetAllInstalledApp", "envsQueryParam", envsQueryParam)
	res, err := handler.installedAppService.GetAll(filter)
	if err != nil {
		handler.Logger.Errorw("service err, GetAllInstalledApp", "err", err, "envsQueryParam", envsQueryParam)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, "*")
	if isActionUserSuperAdmin {
		common.WriteJsonResp(w, err, res, http.StatusOK)
		return
	}

	appIdToAppMap := make(map[string]openapi.HelmApp)

	//the value of this map is array of strings because the GetHelmObjectByAppNameAndEnvId method may return "//" for error cases
	//so different apps may contain same object, to handle that we are using (map[string] []string)
	rbacObjectToAppIdMap1 := make(map[string][]string)
	rbacObjectToAppIdMap2 := make(map[string][]string)

	objectArray1 := make([]string, 0)
	objectArray2 := make([]string, 0)

	for _, app := range *res.HelmApps {

		appIdToAppMap[*app.AppId] = app
		appName := *app.AppName
		envId := (*app.EnvironmentDetail).EnvironmentId
		object1, object2 := handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(appName, int(*envId))
		objectArray1 = append(objectArray1, object1)
		_, ok := rbacObjectToAppIdMap1[object1]
		if !ok {
			rbacObjectToAppIdMap1[object1] = make([]string, 0)
		}
		rbacObjectToAppIdMap1[object1] = append(rbacObjectToAppIdMap1[object1], *app.AppId)
		if object2 != "" {
			_, ok := rbacObjectToAppIdMap2[object2]
			if !ok {
				rbacObjectToAppIdMap2[object2] = make([]string, 0)
			}
			rbacObjectToAppIdMap2[object2] = append(rbacObjectToAppIdMap2[object2], *app.AppId)
			objectArray2 = append(objectArray2, object2)
		}

	}
	start := time.Now()
	resultObjectMap1 := handler.enforcer.EnforceInBatch(token, casbin.ResourceHelmApp, casbin.ActionGet, objectArray1)
	resultObjectMap2 := handler.enforcer.EnforceInBatch(token, casbin.ResourceHelmApp, casbin.ActionGet, objectArray2)
	middleware.AppListingDuration.WithLabelValues("enforceByEmailInBatch", "helm").Observe(time.Since(start).Seconds())
	authorizedAppIdSet := make(map[string]bool)
	//O(n) time loop , at max we will only iterate through all the apps
	for obj, ok := range resultObjectMap1 {
		if ok {
			appIds := rbacObjectToAppIdMap1[obj]
			for _, appId := range appIds {
				authorizedAppIdSet[appId] = true
			}

		}
	}
	for obj, ok := range resultObjectMap2 {
		if ok {
			appIds := rbacObjectToAppIdMap2[obj]
			for _, appId := range appIds {
				authorizedAppIdSet[appId] = true
			}
		}
	}

	authorizedApps := make([]openapi.HelmApp, 0)
	for appId, _ := range authorizedAppIdSet {
		authorizedApp := appIdToAppMap[appId]
		authorizedApps = append(authorizedApps, authorizedApp)
	}

	res.HelmApps = &authorizedApps
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *InstalledAppRestHandlerImpl) DeployBulk(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request appStoreBean.ChartGroupInstallRequest
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, DeployBulk", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, DeployBulk", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	handler.Logger.Infow("request payload, DeployBulk", "payload", request)
	//RBAC block starts from here
	token := r.Header.Get("token")
	rbacObject := ""
	if ok := handler.enforcer.Enforce(token, casbin.ResourceChartGroup, casbin.ActionUpdate, rbacObject); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC block ends here

	visited := make(map[string]bool)

	for _, item := range request.ChartGroupInstallChartRequest {
		if visited[item.AppName] {
			handler.Logger.Errorw("service err, CreateInstalledApp", "err", err, "payload", request)
			common.WriteJsonResp(w, errors.New("duplicate appName found"), nil, http.StatusBadRequest)
			return
		} else {
			visited[item.AppName] = true
		}
		isChartRepoActive, err := handler.appStoreDeploymentService.IsChartRepoActive(item.AppStoreVersion)
		if err != nil {
			handler.Logger.Errorw("service err, CreateInstalledApp", "err", err, "payload", request)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		if !isChartRepoActive {
			common.WriteJsonResp(w, fmt.Errorf("chart repo is disabled"), nil, http.StatusNotAcceptable)
			return
		}
	}
	res, err := handler.installedAppService.DeployBulk(&request)
	if err != nil {
		handler.Logger.Errorw("service err, DeployBulk", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *InstalledAppRestHandlerImpl) CheckAppExists(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request []*appStoreBean.AppNames
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, CheckAppExists", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, CheckAppExists", "payload", request)
	res, err := handler.installedAppService.CheckAppExists(request)
	if err != nil {
		handler.Logger.Errorw("service err, CheckAppExists", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl *InstalledAppRestHandlerImpl) DefaultComponentInstallation(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		impl.Logger.Errorw("service err, DefaultComponentInstallation", "error", err, "userId", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	clusterId, err := strconv.Atoi(vars["clusterId"])
	if err != nil {
		impl.Logger.Errorw("request err, DefaultComponentInstallation", "error", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.Logger.Errorw("request payload, DefaultComponentInstallation", "clusterId", clusterId)
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.Logger.Errorw("service err, DefaultComponentInstallation", "error", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	if ok := impl.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionUpdate, cluster.ClusterName); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	// RBAC enforcer ends
	isTriggered, err := impl.installedAppService.DeployDefaultChartOnCluster(cluster, userId)
	if err != nil {
		impl.Logger.Errorw("service err, DefaultComponentInstallation", "error", err, "cluster", cluster)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, isTriggered, http.StatusOK)
}
func (handler *InstalledAppRestHandlerImpl) FetchNotesForArgoInstalledApp(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installed-app-id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchNotesForArgoInstalledApp", "err", err, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	envId, err := strconv.Atoi(vars["env-id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchNotesForArgoInstalledApp", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchNotesForArgoInstalledApp, app store", "installedAppId", installedAppId, "envId", envId)
	notes, err := handler.installedAppService.FetchChartNotes(installedAppId, envId, token, handler.checkNotesAuth)
	if err != nil {
		handler.Logger.Errorw("service err, FetchNotesFromdb, app store", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, &bean2.Notes{Notes: notes}, http.StatusOK)

}

func (handler *InstalledAppRestHandlerImpl) DeleteArgoInstalledAppWithNonCascade(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installedAppId"])
	if err != nil {
		handler.Logger.Errorw("request err, DeleteArgoInstalledAppWithNonCascade", "err", err, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, delete app", "appId", installedAppId)
	v := r.URL.Query()
	forceDelete := false
	force := v.Get("force")
	if len(force) > 0 {
		forceDelete, err = strconv.ParseBool(force)
		if err != nil {
			handler.Logger.Errorw("request err, NonCascadeDeleteCdPipeline", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	installedApp, err := handler.appStoreDeploymentService.GetInstalledApp(installedAppId)
	if err != nil {
		handler.Logger.Error("request err, NonCascadeDeleteCdPipeline", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if util.IsBaseStack() || util.IsHelmApp(installedApp.AppOfferingMode) || util2.IsHelmApp(installedApp.DeploymentAppType) {
		handler.Logger.Errorw("request err, NonCascadeDeleteCdPipeline", "err", fmt.Errorf("nocascade delete is not supported for %s", installedApp.AppName))
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//rbac block starts from here
	rbacObject, rbacObject2 := handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(installedApp.AppName, installedApp.EnvironmentId)
	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionDelete, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionDelete, rbacObject2)
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rback block ends here
	acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		handler.Logger.Errorw("error in getting acd token", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	ctx := context.WithValue(r.Context(), "token", acdToken)
	request := &appStoreBean.InstallAppVersionDTO{}
	request.InstalledAppId = installedAppId
	request.AppName = installedApp.AppName
	request.AppId = installedApp.AppId
	request.EnvironmentId = installedApp.EnvironmentId
	request.UserId = userId
	request.ForceDelete = forceDelete
	request.NonCascadeDelete = true
	request.AppOfferingMode = installedApp.AppOfferingMode
	request.ClusterId = installedApp.ClusterId
	request.Namespace = installedApp.Namespace
	request.AcdPartialDelete = true

	request, err = handler.appStoreDeploymentService.DeleteInstalledApp(ctx, request)
	if err != nil {
		handler.Logger.Errorw("service err, DeleteInstalledApp", "err", err, "installAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)

}

func (handler *InstalledAppRestHandlerImpl) checkNotesAuth(token string, appName string, envId int) bool {

	object, object2 := handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(appName, envId)
	var ok bool
	if object2 == "" {
		ok = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object)
	} else {
		ok = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object2)
	}
	return ok
}

func (handler *InstalledAppRestHandlerImpl) FetchAppDetailsForInstalledApp(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installed-app-id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchAppDetailsForInstalledApp", "err", err, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	envId, err := strconv.Atoi(vars["env-id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchAppDetailsForInstalledApp", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchAppDetailsForInstalledApp, app store", "installedAppId", installedAppId, "envId", envId)

	installedApp, err := handler.installedAppService.CheckAppExistsByInstalledAppId(installedAppId)
	if err == pg.ErrNoRows {
		common.WriteJsonResp(w, err, "App not found in database", http.StatusBadRequest)
		return
	}

	appDetail, err := handler.installedAppService.FindAppDetailsForAppstoreApplication(installedAppId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchAppDetailsForInstalledApp, app store", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//rbac block starts from here
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
	//rback block ends here
	resourceTreeAndNotesContainer := bean2.AppDetailsContainer{}
	resourceTreeAndNotesContainer.ResourceTree = map[string]interface{}{}

	if len(installedApp.App.AppName) > 0 && len(installedApp.Environment.Name) > 0 {
		err = handler.fetchResourceTree(w, r, &resourceTreeAndNotesContainer, *installedApp, "", "")
		if installedApp.DeploymentAppType == util2.PIPELINE_DEPLOYMENT_TYPE_ACD {
			apiError, ok := err.(*util2.ApiError)
			if ok && apiError != nil {
				if apiError.Code == constants.AppDetailResourceTreeNotFound && installedApp.DeploymentAppDeleteRequest == true {
					err = handler.installedAppService.MarkGitOpsInstalledAppsDeletedIfArgoAppIsDeleted(installedAppId, envId)
					appDeleteErr, appDeleteErrOk := err.(*util2.ApiError)
					if appDeleteErrOk && appDeleteErr != nil {
						handler.Logger.Errorw(appDeleteErr.InternalMessage)
						return
					}
				}
			}
		} else if err != nil {
			common.WriteJsonResp(w, fmt.Errorf("error in fetching resource tree"), nil, http.StatusInternalServerError)
			return
		}
	}
	appDetail.ResourceTree = resourceTreeAndNotesContainer.ResourceTree
	appDetail.Notes = resourceTreeAndNotesContainer.Notes
	common.WriteJsonResp(w, nil, appDetail, http.StatusOK)
}

func (handler *InstalledAppRestHandlerImpl) FetchAppDetailsForInstalledAppV2(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installed-app-id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchAppDetailsForInstalledAppV2", "err", err, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	envId, err := strconv.Atoi(vars["env-id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchAppDetailsForInstalledAppV2", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchAppDetailsForInstalledAppV2, app store", "installedAppId", installedAppId, "envId", envId)
	appDetail, err := handler.installedAppService.FindAppDetailsForAppstoreApplication(installedAppId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchAppDetailsForInstalledAppV2, app store", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//rbac block starts from here
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
	common.WriteJsonResp(w, nil, appDetail, http.StatusOK)
}

func (handler *InstalledAppRestHandlerImpl) FetchResourceTree(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installed-app-id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchAppDetailsForInstalledAppV2", "err", err, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["env-id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchAppDetailsForInstalledAppV2", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchAppDetailsForInstalledAppV2, app store", "installedAppId", installedAppId, "envId", envId)
	installedApp, err := handler.installedAppService.CheckAppExistsByInstalledAppId(installedAppId)
	if err == pg.ErrNoRows {
		common.WriteJsonResp(w, err, "App not found in database", http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	object, object2 := handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(installedApp.App.AppName, installedApp.EnvironmentId)
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
	appDetail, err := handler.installedAppService.FindAppDetailsForAppstoreApplication(installedAppId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchAppDetailsForInstalledApp, app store", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	resourceTreeAndNotesContainer := bean2.AppDetailsContainer{}
	resourceTreeAndNotesContainer.ResourceTree = map[string]interface{}{}

	if len(installedApp.App.AppName) > 0 && len(installedApp.Environment.Name) > 0 {
		err = handler.fetchResourceTree(w, r, &resourceTreeAndNotesContainer, *installedApp, appDetail.HelmReleaseInstallStatus, appDetail.Status)
		if installedApp.DeploymentAppType == util2.PIPELINE_DEPLOYMENT_TYPE_ACD {
			//resource tree has been fetched now prepare to sync application deployment status with this resource tree call
			handler.syncDeploymentStatusWithResourceTreeCall(appDetail)
			apiError, ok := err.(*util2.ApiError)
			if ok && apiError != nil {
				if apiError.Code == constants.AppDetailResourceTreeNotFound && installedApp.DeploymentAppDeleteRequest == true {
					err = handler.installedAppService.MarkGitOpsInstalledAppsDeletedIfArgoAppIsDeleted(installedAppId, envId)
					appDeleteErr, appDeleteErrOk := err.(*util2.ApiError)
					if appDeleteErrOk && appDeleteErr != nil {
						handler.Logger.Errorw(appDeleteErr.InternalMessage)
						return
					}
				}
			}
		} else if err != nil {
			common.WriteJsonResp(w, fmt.Errorf("error in fetching resource tree"), nil, http.StatusInternalServerError)
			return
		}
	}
	common.WriteJsonResp(w, nil, resourceTreeAndNotesContainer, http.StatusOK)
}

func (handler *InstalledAppRestHandlerImpl) syncDeploymentStatusWithResourceTreeCall(appDetail bean2.AppDetailContainer) {
	go func() {
		installedAppVersion, err := handler.installedAppRepository.GetInstalledAppVersion(appDetail.AppStoreInstalledAppVersionId)
		if err != nil {
			handler.Logger.Errorw("error in getting installed_app_version in FetchAppDetailsForInstalledApp", "err", err)
		}
		err = handler.cdApplicationStatusUpdateHandler.SyncPipelineStatusForAppStoreForResourceTreeCall(installedAppVersion)
		if err != nil {
			handler.Logger.Errorw("error in syncing deployment status for installed_app ", "err", err, "installedAppVersion", installedAppVersion)
		}
	}()
}

func (handler *InstalledAppRestHandlerImpl) FetchResourceTreeForACDApp(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installed-app-id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchAppDetailsForInstalledAppV2", "err", err, "installedAppId", installedAppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	envId, err := strconv.Atoi(vars["env-id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchAppDetailsForInstalledAppV2", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchAppDetailsForInstalledAppV2, app store", "installedAppId", installedAppId, "envId", envId)

	appDetail, err := handler.installedAppService.FindAppDetailsForAppstoreApplication(installedAppId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchAppDetailsForInstalledAppV2, app store", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//rbac block starts from here
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
	//rback block ends here
	if len(appDetail.AppName) > 0 && len(appDetail.EnvironmentName) > 0 {
		handler.fetchResourceTreeWithHibernateForACD(w, r, &appDetail)
	} else {
		appDetail.ResourceTree = map[string]interface{}{}
		handler.Logger.Warnw("appName and envName not found - avoiding resource tree call", "app", appDetail.AppName, "env", appDetail.EnvironmentName)
	}
	common.WriteJsonResp(w, err, appDetail, http.StatusOK)
}

func (handler *InstalledAppRestHandlerImpl) fetchResourceTree(w http.ResponseWriter, r *http.Request, resourceTreeAndNotesContainer *bean2.AppDetailsContainer, installedApp repository.InstalledApps, helmReleaseInstallStatus string, status string) error {
	ctx := r.Context()
	cn, _ := w.(http.CloseNotifier)
	err := handler.installedAppService.FetchResourceTree(ctx, cn, resourceTreeAndNotesContainer, installedApp, helmReleaseInstallStatus, status)
	return err
}

func (handler *InstalledAppRestHandlerImpl) fetchResourceTreeWithHibernateForACD(w http.ResponseWriter, r *http.Request, appDetail *bean2.AppDetailContainer) {
	ctx := r.Context()
	cn, _ := w.(http.CloseNotifier)
	handler.installedAppService.FetchResourceTreeWithHibernateForACD(ctx, cn, appDetail)
}
