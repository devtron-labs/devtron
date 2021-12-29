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
	application2 "github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/constants"
	appstore2 "github.com/devtron-labs/devtron/internal/sql/repository/appstore"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appstore"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AppStoreRestHandler interface {
	FindAllApps(w http.ResponseWriter, r *http.Request)
	GetChartDetailsForVersion(w http.ResponseWriter, r *http.Request)
	GetChartVersions(w http.ResponseWriter, r *http.Request)
	FetchAppDetailsForInstalledApp(w http.ResponseWriter, r *http.Request)
	GetReadme(w http.ResponseWriter, r *http.Request)
	SearchAppStoreChartByName(w http.ResponseWriter, r *http.Request)
	GetChartRepoById(w http.ResponseWriter, r *http.Request)
	GetChartRepoList(w http.ResponseWriter, r *http.Request)
	CreateChartRepo(w http.ResponseWriter, r *http.Request)
	UpdateChartRepo(w http.ResponseWriter, r *http.Request)
	ValidateChartRepo(w http.ResponseWriter, r *http.Request)
	TriggerChartSyncManual(w http.ResponseWriter, r *http.Request)
}

type AppStoreRestHandlerImpl struct {
	Logger           *zap.SugaredLogger
	appStoreService  appstore.AppStoreService
	userAuthService  user.UserService
	teamService      team.TeamService
	enforcer         casbin.Enforcer
	acdServiceClient application.ServiceClient
	enforcerUtil     rbac.EnforcerUtil
	validator        *validator.Validate
	client           *http.Client
}

func NewAppStoreRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService, appStoreService appstore.AppStoreService,
	acdServiceClient application.ServiceClient, teamService team.TeamService,
	enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil,
	validator *validator.Validate, client *http.Client) *AppStoreRestHandlerImpl {
	return &AppStoreRestHandlerImpl{
		Logger:           Logger,
		appStoreService:  appStoreService,
		userAuthService:  userAuthService,
		teamService:      teamService,
		acdServiceClient: acdServiceClient,
		enforcer:         enforcer,
		enforcerUtil:     enforcerUtil,
		validator:        validator,
		client:           client,
	}
}

func (handler *AppStoreRestHandlerImpl) FindAllApps(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	v := r.URL.Query()
	deprecated := false
	deprecatedStr := v.Get("includeDeprecated")
	if len(deprecatedStr) > 0 {
		deprecated, err = strconv.ParseBool(deprecatedStr)
		if err != nil {
			deprecated = false
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
	filter := &appstore2.AppStoreFilter{IncludeDeprecated: deprecated, ChartRepoId: chartRepoIds, AppStoreName: appStoreName}
	if size > 0 {
		filter.Size = size
		filter.Offset = offset
	}
	handler.Logger.Infow("request payload, FindAllApps, app store", "userId", userId)
	res, err := handler.appStoreService.FindAllApps(filter)
	if err != nil {
		handler.Logger.Errorw("service err, FindAllApps, app store", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) GetChartDetailsForVersion(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.Logger.Errorw("request err, GetChartDetailsForVersion", "err", err, "appStoreApplicationVersionId", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetChartDetailsForVersion, app store", "appStoreApplicationVersionId", id)
	res, err := handler.appStoreService.FindChartDetailsById(id)
	if err != nil {
		handler.Logger.Errorw("service err, GetChartDetailsForVersion, app store", "err", err, "appStoreApplicationVersionId", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) GetChartVersions(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["appStoreId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetChartVersions", "err", err, "appStoreId", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetChartVersions, app store", "appStoreId", id)
	res, err := handler.appStoreService.FindChartVersionsByAppStoreId(id)
	if err != nil {
		handler.Logger.Errorw("service err, GetChartVersions, app store", "err", err, "appStoreId", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) FetchAppDetailsForInstalledApp(w http.ResponseWriter, r *http.Request) {
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

	appDetail, err := handler.appStoreService.FindAppDetailsForAppstoreApplication(installedAppId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchAppDetailsForInstalledApp, app store", "err", err, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//rbac block starts from here
	object := handler.enforcerUtil.GetAppRBACName(appDetail.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(appDetail.AppName, appDetail.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//rback block ends here

	if len(appDetail.AppName) > 0 && len(appDetail.EnvironmentName) > 0 {
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
		ctx = context.WithValue(ctx, "token", token)
		defer cancel()
		start := time.Now()
		resp, err := handler.acdServiceClient.ResourceTree(ctx, query)
		elapsed := time.Since(start)
		handler.Logger.Debugf("Time elapsed %s in fetching app-store installed application %s for environment %s", elapsed, installedAppId, envId)
		if err != nil {
			handler.Logger.Errorw("service err, FetchAppDetailsForInstalledApp, fetching resource tree", "err", err, "installedAppId", installedAppId, "envId", envId)
			err = &util.ApiError{
				Code:            constants.AppDetailResourceTreeNotFound,
				InternalMessage: "app detail fetched, failed to get resource tree from acd",
				UserMessage:     "app detail fetched, failed to get resource tree from acd",
			}
			appDetail.ResourceTree = &application.ResourceTreeResponse{}
			common.WriteJsonResp(w, nil, appDetail, http.StatusOK)
			return
		}
		appDetail.ResourceTree = resp
		handler.Logger.Debugf("application %s in environment %s had status %+v\n", installedAppId, envId, resp)
	} else {
		handler.Logger.Infow("appName and envName not found - avoiding resource tree call", "app", appDetail.AppName, "env", appDetail.EnvironmentName)
	}
	common.WriteJsonResp(w, err, appDetail, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) GetReadme(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["appStoreApplicationVersionId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetReadme", "err", err, "appStoreApplicationVersionId", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetReadme, app store", "appStoreApplicationVersionId", id)
	res, err := handler.appStoreService.GetReadMeByAppStoreApplicationVersionId(id)
	if err != nil {
		handler.Logger.Errorw("service err, GetReadme, fetching resource tree", "err", err, "appStoreApplicationVersionId", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) SearchAppStoreChartByName(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	chartName := vars["chartName"]
	handler.Logger.Infow("request payload, SearchAppStoreChartByName, app store", "chartName", chartName)
	res, err := handler.appStoreService.SearchAppStoreChartByName(chartName)
	if err != nil {
		handler.Logger.Errorw("service err, SearchAppStoreChartByName, app store", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) GetChartRepoById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.Logger.Errorw("request err, GetChartRepoById", "err", err, "chart repo id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetChartRepoById, app store", "chart repo id", id)
	res, err := handler.appStoreService.GetChartRepoById(id)
	if err != nil {
		handler.Logger.Errorw("service err, GetChartRepoById, app store", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) GetChartRepoList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	handler.Logger.Infow("request payload, GetChartRepoList, app store")
	res, err := handler.appStoreService.GetChartRepoList()
	if err != nil {
		handler.Logger.Errorw("service err, GetChartRepoList, app store", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) CreateChartRepo(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request *appstore.ChartRepoDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, CreateChartRepo", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, CreateChartRepo", "err", err, "payload", request)
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "data validation error", InternalMessage: err.Error()}
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	//rback block ends here
	request.UserId = userId
	handler.Logger.Infow("request payload, CreateChartRepo", "payload", request)
	res, err, validationResult := handler.appStoreService.ValidateAndCreateChartRepo(request)
	if validationResult.CustomErrMsg != appstore.ValidationSuccessMsg {
		common.WriteJsonResp(w, nil, validationResult, http.StatusOK)
		return
	}
	if err != nil {
		handler.Logger.Errorw("service err, CreateChartRepo", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) UpdateChartRepo(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request *appstore.ChartRepoDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, UpdateChartRepo", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateChartRepo", "err", err, "payload", request)
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "data validation error", InternalMessage: err.Error()}
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	//rback block ends here
	request.UserId = userId
	handler.Logger.Infow("request payload, UpdateChartRepo", "payload", request)
	res, err, validationResult := handler.appStoreService.ValidateAndUpdateChartRepo(request)
	if validationResult.CustomErrMsg != appstore.ValidationSuccessMsg {
		common.WriteJsonResp(w, nil, validationResult, http.StatusOK)
		return
	}
	if err != nil {
		handler.Logger.Errorw("service err, UpdateChartRepo", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) ValidateChartRepo(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request *appstore.ChartRepoDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, ValidateChartRepo", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, ValidateChartRepo", "err", err, "payload", request)
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "data validation error", InternalMessage: err.Error()}
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	request.UserId = userId
	handler.Logger.Infow("request payload, ValidateChartRepo", "payload", request)
	validationResult := handler.appStoreService.ValidateChartRepo(request)
	common.WriteJsonResp(w, nil, validationResult, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) TriggerChartSyncManual(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	err2 := handler.appStoreService.TriggerChartSyncManual()
	if err2 != nil {
		common.WriteJsonResp(w, err2, nil, http.StatusInternalServerError)
	} else {
		common.WriteJsonResp(w, nil, map[string]string{"status": "ok"}, http.StatusOK)
	}
}
