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
	"fmt"
	application2 "github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appstore"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type AppStoreRestHandler interface {
	FindAllApps(w http.ResponseWriter, r *http.Request)
	GetChartDetailsForVersion(w http.ResponseWriter, r *http.Request)
	GetChartVersions(w http.ResponseWriter, r *http.Request)
	FetchAppDetailsForInstalledApp(w http.ResponseWriter, r *http.Request)
	GetReadme(w http.ResponseWriter, r *http.Request)
}

type AppStoreRestHandlerImpl struct {
	Logger           *zap.SugaredLogger
	appStoreService  appstore.AppStoreService
	userAuthService  user.UserService
	teamService      team.TeamService
	enforcer         rbac.Enforcer
	acdServiceClient application.ServiceClient
	enforcerUtil     rbac.EnforcerUtil
}

func NewAppStoreRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService, appStoreService appstore.AppStoreService,
	acdServiceClient application.ServiceClient, teamService team.TeamService,
	enforcer rbac.Enforcer, enforcerUtil rbac.EnforcerUtil) *AppStoreRestHandlerImpl {
	return &AppStoreRestHandlerImpl{
		Logger:           Logger,
		appStoreService:  appStoreService,
		userAuthService:  userAuthService,
		teamService:      teamService,
		acdServiceClient: acdServiceClient,
		enforcer:         enforcer,
		enforcerUtil:     enforcerUtil,
	}
}

func (handler *AppStoreRestHandlerImpl) FindAllApps(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	handler.Logger.Infow("request payload, FindAllApps, app store", "userId", userId)
	res, err := handler.appStoreService.FindAllApps()
	if err != nil {
		handler.Logger.Errorw("service err, FindAllApps, app store", "err", err, "userId", userId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) GetChartDetailsForVersion(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.Logger.Errorw("request err, GetChartDetailsForVersion", "err", err, "appStoreApplicationVersionId", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetChartDetailsForVersion, app store", "appStoreApplicationVersionId", id)
	res, err := handler.appStoreService.FindChartDetailsById(id)
	if err != nil {
		handler.Logger.Errorw("service err, GetChartDetailsForVersion, app store", "err", err, "appStoreApplicationVersionId", id)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) GetChartVersions(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["appStoreId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetChartVersions", "err", err, "appStoreId", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetChartVersions, app store", "appStoreId", id)
	res, err := handler.appStoreService.FindChartVersionsByAppStoreId(id)
	if err != nil {
		handler.Logger.Errorw("service err, GetChartVersions, app store", "err", err, "appStoreId", id)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) FetchAppDetailsForInstalledApp(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installed-app-id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchAppDetailsForInstalledApp", "err", err, "installedAppId", installedAppId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	envId, err := strconv.Atoi(vars["env-id"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchAppDetailsForInstalledApp", "err", err, "installedAppId", installedAppId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchAppDetailsForInstalledApp, app store", "installedAppId", installedAppId, "envId", envId)

	appDetail, err := handler.appStoreService.FindAppDetailsForAppstoreApplication(installedAppId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchAppDetailsForInstalledApp, app store", "err", err, "installedAppId", installedAppId, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//rbac block starts from here
	object := handler.enforcerUtil.GetAppRBACName(appDetail.AppName)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(appDetail.AppName, appDetail.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionGet, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
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
			writeJsonResp(w, nil, appDetail, http.StatusOK)
			return
		}
		appDetail.ResourceTree = resp
		handler.Logger.Debugf("application %s in environment %s had status %+v\n", installedAppId, envId, resp)
	} else {
		handler.Logger.Infow("appName and envName not found - avoiding resource tree call", "app", appDetail.AppName, "env", appDetail.EnvironmentName)
	}
	writeJsonResp(w, err, appDetail, http.StatusOK)
}

func (handler *AppStoreRestHandlerImpl) GetReadme(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["appStoreApplicationVersionId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetReadme", "err", err, "appStoreApplicationVersionId", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetReadme, app store", "appStoreApplicationVersionId", id)
	res, err := handler.appStoreService.GetReadMeByAppStoreApplicationVersionId(id)
	if err != nil {
		handler.Logger.Errorw("service err, GetReadme, fetching resource tree", "err", err, "appStoreApplicationVersionId", id)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}
