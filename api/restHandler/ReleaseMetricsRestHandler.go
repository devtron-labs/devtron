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
	"github.com/devtron-labs/devtron/client/lens"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"go.uber.org/zap"
	"net/http"
)

type ReleaseMetricsRestHandler interface {
	ResetDataForAppEnvironment(w http.ResponseWriter, r *http.Request)
	ResetDataForAllAppEnvironment(w http.ResponseWriter, r *http.Request)
	GetDeploymentMetrics(w http.ResponseWriter, r *http.Request)
}

type ReleaseMetricsRestHandlerImpl struct {
	logger             *zap.SugaredLogger
	enforcer           rbac.Enforcer
	ReleaseDataService app.ReleaseDataService
	userAuthService    user.UserService
	teamService        team.TeamService
	pipelineRepository pipelineConfig.PipelineRepository
	enforcerUtil       rbac.EnforcerUtil
}

func NewReleaseMetricsRestHandlerImpl(
	logger *zap.SugaredLogger,
	enforcer rbac.Enforcer,
	ReleaseDataService app.ReleaseDataService,
	userAuthService user.UserService,
	teamService team.TeamService,
	pipelineRepository pipelineConfig.PipelineRepository, enforcerUtil rbac.EnforcerUtil) *ReleaseMetricsRestHandlerImpl {
	return &ReleaseMetricsRestHandlerImpl{
		logger:             logger,
		enforcer:           enforcer,
		ReleaseDataService: ReleaseDataService,
		userAuthService:    userAuthService,
		teamService:        teamService,
		pipelineRepository: pipelineRepository,
		enforcerUtil:       enforcerUtil,
	}
}

type MetricsRequest struct {
	AppId         int `json:"appId"`
	EnvironmentId int `json:"environmentId"`
}

func (impl *ReleaseMetricsRestHandlerImpl) ResetDataForAppEnvironment(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var req MetricsRequest
	err = decoder.Decode(&req)
	if err != nil {
		impl.logger.Errorw("request err, ResetDataForAppEnvironment", "err", err, "payload", req)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, ResetDataForAppEnvironment", "err", err, "payload", req)
	//RBAC
	token := r.Header.Get("token")
	appRbacObject := impl.enforcerUtil.GetAppRBACNameByAppId(req.AppId)
	if appRbacObject == "" {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	envRbacObject := impl.enforcerUtil.GetEnvRBACNameByAppId(req.AppId, req.EnvironmentId)
	if envRbacObject == "" {
		writeJsonResp(w, fmt.Errorf("envId is incorrect"), nil, http.StatusBadRequest)
		return
	}
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, appRbacObject); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if ok := impl.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionCreate, envRbacObject); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC end

	err = impl.ReleaseDataService.TriggerEventForAllRelease(req.AppId, req.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("service err, ResetDataForAppEnvironment", "err", err, "payload", req)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, true, http.StatusOK)
}

func (impl *ReleaseMetricsRestHandlerImpl) ResetDataForAllAppEnvironment(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	pipelines, err := impl.pipelineRepository.UniqueAppEnvironmentPipelines()
	if err != nil {
		impl.logger.Errorw("service err, ResetDataForAllAppEnvironment", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	for _, pipeline := range pipelines {
		appRbacObject := impl.enforcerUtil.GetAppRBACNameByAppId(pipeline.AppId)
		if appRbacObject == "" {
			continue
		}
		envRbacObject := impl.enforcerUtil.GetEnvRBACNameByAppId(pipeline.AppId, pipeline.EnvironmentId)
		if envRbacObject == "" {
			continue
		}
		if !impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionCreate, appRbacObject) {
			continue
		}
		if !impl.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionCreate, envRbacObject) {
			continue
		}
		impl.logger.Infow("trigger event, ResetDataForAllAppEnvironment", "app", pipeline.AppId, "env", pipeline.EnvironmentId)
		err = impl.ReleaseDataService.TriggerEventForAllRelease(pipeline.AppId, pipeline.EnvironmentId)
		if err != nil {
			impl.logger.Errorw("service err, ResetDataForAllAppEnvironment, trigger event", "err", err, "app", pipeline.AppId, "env", pipeline.EnvironmentId)
		}
	}
}

//err := decoder.Decode(&employeeStruct, r.URL.Query())
var decoder = schema.NewDecoder()

func (impl *ReleaseMetricsRestHandlerImpl) GetDeploymentMetrics(w http.ResponseWriter, r *http.Request) {
	//decoder := json.NewDecoder(r.Body)
	metricRequest := &lens.MetricRequest{}
	err := decoder.Decode(metricRequest, r.URL.Query())
	if err != nil {
		impl.logger.Errorw("request err, GetDeploymentMetrics", "err", err, "payload", metricRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	appRbacObject := impl.enforcerUtil.GetAppRBACNameByAppId(metricRequest.AppId)
	if appRbacObject == "" {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, appRbacObject); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	resBody, resCode, err := impl.ReleaseDataService.GetDeploymentMetrics(metricRequest)
	if err != nil {
		impl.logger.Errorw("service err, GetDeploymentMetrics", "err", err, "payload", metricRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(*resCode))
	_, err = w.Write(resBody)
	if err != nil {
		impl.logger.Errorw("service err, GetDeploymentMetrics", "err", err, "resCode", resCode)
	}
}
