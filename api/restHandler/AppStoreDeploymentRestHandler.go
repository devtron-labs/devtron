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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	appstore2 "github.com/devtron-labs/devtron/internal/sql/repository/appstore"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appstore"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/devtron-labs/devtron/util/response"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
)

type InstalledAppRestHandler interface {
	CreateInstalledApp(w http.ResponseWriter, r *http.Request)
	UpdateInstalledApp(w http.ResponseWriter, r *http.Request)
	GetAllInstalledApp(w http.ResponseWriter, r *http.Request)
	GetInstalledAppsByAppStoreId(w http.ResponseWriter, r *http.Request)
	GetInstalledAppVersion(w http.ResponseWriter, r *http.Request)
	DeleteInstalledApp(w http.ResponseWriter, r *http.Request)
	DeployBulk(w http.ResponseWriter, r *http.Request)
	CheckAppExists(w http.ResponseWriter, r *http.Request)
}

type InstalledAppRestHandlerImpl struct {
	pipelineBuilder     pipeline.PipelineBuilder
	Logger              *zap.SugaredLogger
	chartService        pipeline.ChartService
	userAuthService     user.UserService
	teamService         team.TeamService
	enforcer            rbac.Enforcer
	pipelineRepository  pipelineConfig.PipelineRepository
	enforcerUtil        rbac.EnforcerUtil
	configMapService    pipeline.ConfigMapService
	installedAppService appstore.InstalledAppService
	validator           *validator.Validate
}

func NewInstalledAppRestHandlerImpl(pipelineBuilder pipeline.PipelineBuilder, Logger *zap.SugaredLogger,
	chartService pipeline.ChartService, userAuthService user.UserService, teamService team.TeamService,
	enforcer rbac.Enforcer, pipelineRepository pipelineConfig.PipelineRepository,
	enforcerUtil rbac.EnforcerUtil, configMapService pipeline.ConfigMapService,
	installedAppService appstore.InstalledAppService,
	validator *validator.Validate) *InstalledAppRestHandlerImpl {
	return &InstalledAppRestHandlerImpl{
		pipelineBuilder:     pipelineBuilder,
		Logger:              Logger,
		chartService:        chartService,
		userAuthService:     userAuthService,
		teamService:         teamService,
		enforcer:            enforcer,
		pipelineRepository:  pipelineRepository,
		enforcerUtil:        enforcerUtil,
		configMapService:    configMapService,
		installedAppService: installedAppService,
		validator:           validator,
	}
}

func (handler InstalledAppRestHandlerImpl) CreateInstalledApp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request appstore.InstallAppVersionDTO

	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, CreateInstalledApp", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, CreateInstalledApp", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	//rbac block starts from here
	team, err := handler.teamService.FetchOne(request.TeamId)
	if err != nil {
		handler.Logger.Errorw("service err, CreateInstalledApp", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	teamRbac := team.Name + "/" + request.AppName
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, teamRbac); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(request.AppName, request.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionCreate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rback block ends here

	isChartRepoActive, err := handler.installedAppService.IsChartRepoActive(request.AppStoreVersion)
	if err != nil {
		handler.Logger.Errorw("service err, CreateInstalledApp", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !isChartRepoActive {
		writeJsonResp(w, fmt.Errorf("chart repo is disabled"), nil, http.StatusNotAcceptable)
		return
	}

	request.UserId = userId
	handler.Logger.Infow("request payload, CreateInstalledApp", "payload", request)
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
	ctx = context.WithValue(r.Context(), "token", token)
	defer cancel()
	res, err := handler.installedAppService.CreateInstalledAppV2(&request, ctx)
	if err != nil {
		if strings.Contains(err.Error(), "application spec is invalid") {
			err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "application spec is invalid, please check provided chart values"}
		}
		handler.Logger.Errorw("service err, CreateInstalledApp", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler InstalledAppRestHandlerImpl) UpdateInstalledApp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request appstore.InstallAppVersionDTO
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, UpdateInstalledApp", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateInstalledApp", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	handler.Logger.Infow("request payload, UpdateInstalledApp", "payload", request)
	installedApp, err := handler.installedAppService.GetInstalledApp(request.InstalledAppId)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateInstalledApp", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//rbac block starts from here
	object := handler.enforcerUtil.GetAppRBACNameByAppId(installedApp.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionUpdate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(installedApp.AppId, installedApp.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionUpdate, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rback block ends here

	request.UserId = userId
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
	ctx = context.WithValue(r.Context(), "token", token)
	res, err := handler.installedAppService.UpdateInstalledApp(ctx, &request)
	if err != nil {
		if strings.Contains(err.Error(), "application spec is invalid") {
			err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "application spec is invalid, please check provided chart values"}
		}
		handler.Logger.Errorw("service err, UpdateInstalledApp", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler InstalledAppRestHandlerImpl) GetAllInstalledApp(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	v := r.URL.Query()
	token := r.Header.Get("token")
	var envs []int
	envsQueryParam := v.Get("envs")
	if envsQueryParam != "" {
		envsStr := strings.Split(envsQueryParam, ",")
		for _, t := range envsStr {
			env, err := strconv.Atoi(t)
			if err != nil {
				handler.Logger.Errorw("request err, GetAllInstalledApp", "err", err, "envsQueryParam", envsQueryParam)
				response.WriteResponse(http.StatusBadRequest, "please send valid envs", w, errors.New("env id invalid"))
				return
			}
			envs = append(envs, env)
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
	chartRepoIdsStr := v.Get("chartRepoId")
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
	filter := &appstore2.AppStoreFilter{OnlyDeprecated: onlyDeprecated, ChartRepoId: chartRepoIds, AppStoreName: appStoreName, EnvIds: envs, AppName: appName}
	if size > 0 {
		filter.Size = size
		filter.Offset = offset
	}
	handler.Logger.Infow("request payload, GetAllInstalledApp", "envsQueryParam", envsQueryParam)
	res, err := handler.installedAppService.GetAll(filter)
	if err != nil {
		handler.Logger.Errorw("service err, GetAllInstalledApp", "err", err, "envsQueryParam", envsQueryParam)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	var installedAppsResponse []appstore.InstalledAppsResponse
	for _, app := range res {
		//rbac block starts from here
		object := handler.enforcerUtil.GetAppRBACName(app.AppName)
		if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
			continue
		}
		object = handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, app.EnvironmentId)
		if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionGet, object); !ok {
			continue
		}
		//rback block ends here
		installedAppsResponse = append(installedAppsResponse, app)
	}

	writeJsonResp(w, err, installedAppsResponse, http.StatusOK)
}

func (handler InstalledAppRestHandlerImpl) GetInstalledAppsByAppStoreId(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	appStoreId, err := strconv.Atoi(vars["appStoreId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetInstalledAppsByAppStoreId", "err", err, "appStoreId", appStoreId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	handler.Logger.Infow("request payload, GetInstalledAppsByAppStoreId", "appStoreId", appStoreId)
	res, err := handler.installedAppService.GetAllInstalledAppsByAppStoreId(w, r, token, appStoreId)
	if err != nil {
		handler.Logger.Errorw("service err, GetInstalledAppsByAppStoreId", "err", err, "appStoreId", appStoreId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	var installedAppsResponse []appstore.InstalledAppsResponse
	for _, app := range res {
		//rbac block starts from here
		object := handler.enforcerUtil.GetAppRBACName(app.AppName)
		if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
			continue
		}
		object = handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(app.AppName, app.EnvironmentId)
		if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionGet, object); !ok {
			continue
		}
		//rback block ends here
		installedAppsResponse = append(installedAppsResponse, app)
	}

	writeJsonResp(w, err, installedAppsResponse, http.StatusOK)
}

func (handler InstalledAppRestHandlerImpl) GetInstalledAppVersion(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installedAppVersionId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetInstalledAppVersion", "err", err, "installedAppVersionId", installedAppId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	handler.Logger.Infow("request payload, GetInstalledAppVersion", "installedAppVersionId", installedAppId)
	dto, err := handler.installedAppService.GetInstalledAppVersion(installedAppId)
	if err != nil {
		handler.Logger.Errorw("service err, GetInstalledAppVersion", "err", err, "installedAppVersionId", installedAppId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//rbac block starts from here
	object := handler.enforcerUtil.GetAppRBACName(dto.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(dto.AppName, dto.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rback block ends here

	writeJsonResp(w, err, dto, http.StatusOK)
}

func (handler InstalledAppRestHandlerImpl) DeleteInstalledApp(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	installAppId, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.Logger.Errorw("request err, DeleteInstalledApp", "err", err, "installAppId", installAppId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	v := r.URL.Query()
	forceDelete := false
	force := v.Get("force")
	if len(force) > 0 {
		forceDelete, err = strconv.ParseBool(force)
		if err != nil {
			handler.Logger.Errorw("request err, DeleteInstalledApp", "err", err, "installAppId", installAppId)
			writeJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	handler.Logger.Infow("request payload, DeleteInstalledApp", "installAppId", installAppId)
	token := r.Header.Get("token")
	//rbac block starts from here
	installedApp, err := handler.installedAppService.GetInstalledApp(installAppId)
	if err != nil {
		handler.Logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	object := handler.enforcerUtil.GetAppRBACNameByAppId(installedApp.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionDelete, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(installedApp.AppName, installedApp.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionDelete, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rback block ends here
	request := appstore.InstallAppVersionDTO{}
	request.InstalledAppId = installAppId
	request.AppId = installedApp.AppId
	request.EnvironmentId = installedApp.EnvironmentId
	request.UserId = userId
	request.ForceDelete = forceDelete
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
	ctx = context.WithValue(r.Context(), "token", token)
	res, err := handler.installedAppService.DeleteInstalledApp(ctx, &request)
	if err != nil {
		handler.Logger.Errorw("service err, DeleteInstalledApp", "err", err, "installAppId", installAppId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler *InstalledAppRestHandlerImpl) DeployBulk(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request appstore.ChartGroupInstallRequest
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, DeployBulk", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, DeployBulk", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	handler.Logger.Infow("request payload, DeployBulk", "payload", request)
	//RBAC block starts from here
	token := r.Header.Get("token")
	rbacObject := ""
	if ok := handler.enforcer.Enforce(token, rbac.ResourceChartGroup, rbac.ActionGet, rbacObject); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC block ends here

	for _, item := range request.ChartGroupInstallChartRequest {
		isChartRepoActive, err := handler.installedAppService.IsChartRepoActive(item.AppStoreVersion)
		if err != nil {
			handler.Logger.Errorw("service err, CreateInstalledApp", "err", err, "payload", request)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		if !isChartRepoActive {
			writeJsonResp(w, fmt.Errorf("chart repo is disabled"), nil, http.StatusNotAcceptable)
			return
		}
	}
	res, err := handler.installedAppService.DeployBulk(&request)
	if err != nil {
		handler.Logger.Errorw("service err, DeployBulk", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler *InstalledAppRestHandlerImpl) CheckAppExists(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request []*appstore.AppNames
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, CheckAppExists", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, CheckAppExists", "payload", request)
	res, err := handler.installedAppService.CheckAppExists(request)
	if err != nil {
		handler.Logger.Errorw("service err, CheckAppExists", "err", err, "payload", request)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}
