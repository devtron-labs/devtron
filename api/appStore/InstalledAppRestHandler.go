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
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/devtron-labs/devtron/util/response"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type InstalledAppRestHandler interface {
	GetAllInstalledApp(w http.ResponseWriter, r *http.Request)
	DeployBulk(w http.ResponseWriter, r *http.Request)
	CheckAppExists(w http.ResponseWriter, r *http.Request)
	DefaultComponentInstallation(w http.ResponseWriter, r *http.Request)
	FetchAppDetailsForInstalledApp(w http.ResponseWriter, r *http.Request)
}

type InstalledAppRestHandlerImpl struct {
	Logger                    *zap.SugaredLogger
	userAuthService           user.UserService
	enforcer                  casbin.Enforcer
	enforcerUtil              rbac.EnforcerUtil
	installedAppService       service.InstalledAppService
	validator                 *validator.Validate
	clusterService            cluster.ClusterService
	acdServiceClient          application.ServiceClient
	appStoreDeploymentService service.AppStoreDeploymentService
	helmAppClient             client.HelmAppClient
	helmAppService            client.HelmAppService
	argoUserService           argo.ArgoUserService
}

func NewInstalledAppRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService,
	enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, installedAppService service.InstalledAppService,
	validator *validator.Validate, clusterService cluster.ClusterService, acdServiceClient application.ServiceClient,
	appStoreDeploymentService service.AppStoreDeploymentService, helmAppClient client.HelmAppClient, helmAppService client.HelmAppService,
	argoUserService argo.ArgoUserService,
) *InstalledAppRestHandlerImpl {
	return &InstalledAppRestHandlerImpl{
		Logger:                    Logger,
		userAuthService:           userAuthService,
		enforcer:                  enforcer,
		enforcerUtil:              enforcerUtil,
		installedAppService:       installedAppService,
		validator:                 validator,
		clusterService:            clusterService,
		acdServiceClient:          acdServiceClient,
		appStoreDeploymentService: appStoreDeploymentService,
		helmAppService:            helmAppService,
		helmAppClient:             helmAppClient,
		argoUserService:           argoUserService,
	}
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
	filter := &appStoreBean.AppStoreFilter{OnlyDeprecated: onlyDeprecated, ChartRepoId: chartRepoIds, AppStoreName: appStoreName, EnvIds: envIds, AppName: appName, ClusterIds: clusterIds}
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

	authorizedApp := make([]openapi.HelmApp, 0)
	for _, app := range *res.HelmApps {
		appName := *app.AppName
		envId := (*app.EnvironmentDetail).EnvironmentId
		//rbac block starts from here
		object := handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(appName, int(*envId))
		if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object); !ok {
			continue
		}
		authorizedApp = append(authorizedApp, app)
		//rback block ends here
	}
	res.HelmApps = &authorizedApp

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

	for _, item := range request.ChartGroupInstallChartRequest {
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
	if ok := impl.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionUpdate, strings.ToLower(cluster.ClusterName)); !ok {
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

	appDetail, err := handler.installedAppService.FindAppDetailsForAppstoreApplication(installedAppId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchAppDetailsForInstalledApp, app store", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//rbac block starts from here
	object := handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(appDetail.AppName, appDetail.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rback block ends here

	if len(appDetail.AppName) > 0 && len(appDetail.EnvironmentName) > 0 {
		handler.fetchResourceTree(w, r, token, &appDetail)
	} else {
		appDetail.ResourceTree = map[string]interface{}{}
		handler.Logger.Warnw("appName and envName not found - avoiding resource tree call", "app", appDetail.AppName, "env", appDetail.EnvironmentName)
	}
	common.WriteJsonResp(w, err, appDetail, http.StatusOK)
}

func (handler *InstalledAppRestHandlerImpl) fetchResourceTree(w http.ResponseWriter, r *http.Request, token string, appDetail *bean2.AppDetailContainer) {
	if appDetail.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD {
		acdAppName := appDetail.AppName + "-" + appDetail.EnvironmentName
		query := &application2.ResourcesQuery{
			ApplicationName: &acdAppName,
		}
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
		acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			handler.Logger.Errorw("error in getting acd token", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		ctx = context.WithValue(ctx, "token", acdToken)
		defer cancel()
		start := time.Now()
		resp, err := handler.acdServiceClient.ResourceTree(ctx, query)
		elapsed := time.Since(start)
		handler.Logger.Debugf("Time elapsed %s in fetching app-store installed application %s for environment %s", elapsed, appDetail.InstalledAppId, appDetail.EnvironmentId)
		if err != nil {
			handler.Logger.Errorw("service err, FetchAppDetailsForInstalledApp, fetching resource tree", "err", err, "installedAppId", appDetail.InstalledAppId, "envId", appDetail.EnvironmentId)
			err = &util.ApiError{
				Code:            constants.AppDetailResourceTreeNotFound,
				InternalMessage: "app detail fetched, failed to get resource tree from acd",
				UserMessage:     "app detail fetched, failed to get resource tree from acd",
			}
			appDetail.ResourceTree = map[string]interface{}{}
			return
		}
		appDetail.ResourceTree = util2.InterfaceToMapAdapter(resp)
		handler.Logger.Debugf("application %s in environment %s had status %+v\n", appDetail.InstalledAppId, appDetail.EnvironmentId, resp)
	} else if appDetail.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_HELM {
		config, err := handler.helmAppService.GetClusterConf(appDetail.ClusterId)
		if err != nil {
			handler.Logger.Errorw("error in fetching cluster detail", "err", err)
		}
		req := &client.AppDetailRequest{
			ClusterConfig: config,
			Namespace:     appDetail.Namespace,
			ReleaseName:   fmt.Sprintf("%s", appDetail.AppName),
		}
		detail, err := handler.helmAppClient.GetAppDetail(context.Background(), req)
		if err != nil {
			handler.Logger.Errorw("error in fetching app detail", "err", err)
		}
		if detail != nil {
			resourceTree := util2.InterfaceToMapAdapter(detail.ResourceTreeResponse)
			resourceTree["status"] = detail.ReleaseStatus.Status
			appDetail.ResourceTree = resourceTree
			handler.Logger.Warnw("appName and envName not found - avoiding resource tree call", "app", appDetail.AppName, "env", appDetail.EnvironmentName)
		} else {
			appDetail.ResourceTree = map[string]interface{}{}
		}
	}
	return
}
