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
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/deploymentGroup"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type PipelineTriggerRestHandler interface {
	OverrideConfig(w http.ResponseWriter, r *http.Request)
	ReleaseStatusUpdate(w http.ResponseWriter, r *http.Request)
	StartStopApp(w http.ResponseWriter, r *http.Request)
	StartStopDeploymentGroup(w http.ResponseWriter, r *http.Request)
}

type PipelineTriggerRestHandlerImpl struct {
	appService             app.AppService
	userAuthService        user.UserService
	validator              *validator.Validate
	enforcer               rbac.Enforcer
	teamService            team.TeamService
	logger                 *zap.SugaredLogger
	workflowDagExecutor    pipeline.WorkflowDagExecutor
	enforcerUtil           rbac.EnforcerUtil
	deploymentGroupService deploymentGroup.DeploymentGroupService
}

func NewPipelineRestHandler(appService app.AppService, userAuthService user.UserService, validator *validator.Validate,
	enforcer rbac.Enforcer, teamService team.TeamService, logger *zap.SugaredLogger, enforcerUtil rbac.EnforcerUtil,
	workflowDagExecutor pipeline.WorkflowDagExecutor, deploymentGroupService deploymentGroup.DeploymentGroupService) *PipelineTriggerRestHandlerImpl {
	pipelineHandler := &PipelineTriggerRestHandlerImpl{
		appService:             appService,
		userAuthService:        userAuthService,
		validator:              validator,
		enforcer:               enforcer,
		teamService:            teamService,
		logger:                 logger,
		workflowDagExecutor:    workflowDagExecutor,
		enforcerUtil:           enforcerUtil,
		deploymentGroupService: deploymentGroupService,
	}
	return pipelineHandler
}

func (handler PipelineTriggerRestHandlerImpl) OverrideConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var overrideRequest bean.ValuesOverrideRequest
	err = decoder.Decode(&overrideRequest)
	if err != nil {
		handler.logger.Errorw("request err, OverrideConfig", "err", err, "payload", overrideRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	overrideRequest.UserId = userId
	handler.logger.Infow("request err, OverrideConfig", "err", err, "payload", overrideRequest)
	err = handler.validator.Struct(overrideRequest)
	if err != nil {
		handler.logger.Errorw("request err, OverrideConfig", "err", err, "payload", overrideRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")

	//rbac block starts from here
	object := handler.enforcerUtil.GetAppRBACNameByAppId(overrideRequest.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionTrigger, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(overrideRequest.AppId, overrideRequest.PipelineId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionTrigger, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback block ends here

	ctx := context.WithValue(r.Context(), "token", token)
	mergeResp, err := handler.workflowDagExecutor.ManualCdTrigger(&overrideRequest, ctx)
	if err != nil {
		handler.logger.Errorw("request err, OverrideConfig", "err", err, "payload", overrideRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	res := map[string]interface{}{"releaseId": mergeResp}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) StartStopApp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var overrideRequest pipeline.StopAppRequest
	err = decoder.Decode(&overrideRequest)
	if err != nil {
		handler.logger.Errorw("request err, StartStopApp", "err", err, "payload", overrideRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	overrideRequest.UserId = userId
	handler.logger.Infow("request payload, StartStopApp", "err", err, "payload", overrideRequest)
	err = handler.validator.Struct(overrideRequest)
	if err != nil {
		handler.logger.Errorw("validation err, StartStopApp", "err", err, "payload", overrideRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	//rbac block starts from here
	object := handler.enforcerUtil.GetAppRBACNameByAppId(overrideRequest.AppId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionTrigger, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(overrideRequest.AppId, overrideRequest.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionTrigger, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback block ends here

	ctx := context.WithValue(r.Context(), "token", token)
	mergeResp, err := handler.workflowDagExecutor.StopStartApp(&overrideRequest, ctx)
	if err != nil {
		handler.logger.Errorw("service err, StartStopApp", "err", err, "payload", overrideRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	res := map[string]interface{}{"releaseId": mergeResp}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) StartStopDeploymentGroup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var stopDeploymentGroupRequest pipeline.StopDeploymentGroupRequest
	err = decoder.Decode(&stopDeploymentGroupRequest)
	if err != nil {
		handler.logger.Errorw("request err, StartStopDeploymentGroup", "err", err, "payload", stopDeploymentGroupRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	stopDeploymentGroupRequest.UserId = userId
	err = handler.validator.Struct(stopDeploymentGroupRequest)
	if err != nil {
		handler.logger.Errorw("validation err, StartStopDeploymentGroup", "err", err, "payload", stopDeploymentGroupRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, StartStopDeploymentGroup", "err", err, "payload", stopDeploymentGroupRequest)

	//rbac block starts from here
	dg, err := handler.deploymentGroupService.GetDeploymentGroupById(stopDeploymentGroupRequest.DeploymentGroupId)
	if err != nil {
		handler.logger.Errorw("request err, StartStopDeploymentGroup", "err", err, "payload", stopDeploymentGroupRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	token := r.Header.Get("token")
	// RBAC enforcer applying
	object := handler.enforcerUtil.GetTeamRBACByCiPipelineId(dg.CiPipelineId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionTrigger, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByCiPipelineIdAndEnvId(dg.CiPipelineId, dg.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, rbac.ResourceEnvironment, rbac.ActionTrigger, object); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback block ends here

	ctx := context.WithValue(r.Context(), "token", token)
	res, err := handler.workflowDagExecutor.TriggerBulkHibernateAsync(stopDeploymentGroupRequest, ctx)
	if err != nil {
		handler.logger.Errorw("service err, StartStopDeploymentGroup", "err", err, "payload", stopDeploymentGroupRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) ReleaseStatusUpdate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var releaseStatusUpdateRequest bean.ReleaseStatusUpdateRequest
	err = decoder.Decode(&releaseStatusUpdateRequest)
	if err != nil {
		handler.logger.Errorw("request err, ReleaseStatusUpdate", "err", err, "payload", releaseStatusUpdateRequest)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, ReleaseStatusUpdate, override request ----", "err", err, "payload", releaseStatusUpdateRequest)
	res, err := handler.appService.UpdateReleaseStatus(&releaseStatusUpdateRequest)
	if err != nil {
		handler.logger.Errorw("service err, ReleaseStatusUpdate", "err", err, "payload", releaseStatusUpdateRequest)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	m := map[string]bool{
		"status": res}
	resJson, err := json.Marshal(m)
	if err != nil {
		handler.logger.Errorw("marshal err, ReleaseStatusUpdate", "err", err, "payload", m)
	}
	writeJsonResp(w, err, resJson, http.StatusOK)
}
