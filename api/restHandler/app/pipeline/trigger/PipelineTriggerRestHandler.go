/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package trigger

import (
	"encoding/json"
	"fmt"
	util2 "github.com/devtron-labs/devtron/internal/util"
	bean5 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/deployedApp/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/out"
	bean4 "github.com/devtron-labs/devtron/pkg/eventProcessor/out/bean"
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/deploymentGroup"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type PipelineTriggerRestHandler interface {
	OverrideConfig(w http.ResponseWriter, r *http.Request)
	ReleaseStatusUpdate(w http.ResponseWriter, r *http.Request)
	StartStopApp(w http.ResponseWriter, r *http.Request)
	StartStopDeploymentGroup(w http.ResponseWriter, r *http.Request)
	GetAllLatestDeploymentConfiguration(w http.ResponseWriter, r *http.Request)
	RotatePods(w http.ResponseWriter, r *http.Request)
}

type PipelineTriggerRestHandlerImpl struct {
	appService                  app.AppService
	userAuthService             user.UserService
	validator                   *validator.Validate
	enforcer                    casbin.Enforcer
	teamService                 team.TeamService
	logger                      *zap.SugaredLogger
	enforcerUtil                rbac.EnforcerUtil
	deploymentGroupService      deploymentGroup.DeploymentGroupService
	deploymentConfigService     pipeline.PipelineDeploymentConfigService
	deployedAppService          deployedApp.DeployedAppService
	cdHandlerService            devtronApps.HandlerService
	workflowEventPublishService out.WorkflowEventPublishService
}

func NewPipelineRestHandler(appService app.AppService, userAuthService user.UserService, validator *validator.Validate,
	enforcer casbin.Enforcer, teamService team.TeamService, logger *zap.SugaredLogger, enforcerUtil rbac.EnforcerUtil,
	deploymentGroupService deploymentGroup.DeploymentGroupService,
	deploymentConfigService pipeline.PipelineDeploymentConfigService,
	deployedAppService deployedApp.DeployedAppService,
	cdHandlerService devtronApps.HandlerService,
	workflowEventPublishService out.WorkflowEventPublishService) *PipelineTriggerRestHandlerImpl {
	pipelineHandler := &PipelineTriggerRestHandlerImpl{
		appService:                  appService,
		userAuthService:             userAuthService,
		validator:                   validator,
		enforcer:                    enforcer,
		teamService:                 teamService,
		logger:                      logger,
		enforcerUtil:                enforcerUtil,
		deploymentGroupService:      deploymentGroupService,
		deploymentConfigService:     deploymentConfigService,
		deployedAppService:          deployedAppService,
		cdHandlerService:            cdHandlerService,
		workflowEventPublishService: workflowEventPublishService,
	}
	return pipelineHandler
}

func (handler PipelineTriggerRestHandlerImpl) validateCdTriggerRBAC(token string, appId, cdPipelineId int) error {
	// RBAC block starts from here
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
		return util2.NewApiError(http.StatusForbidden, common.UnAuthorisedUser, common.UnAuthorisedUser)
	}
	object = handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(appId, cdPipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, object); !ok {
		return util2.NewApiError(http.StatusForbidden, common.UnAuthorisedUser, common.UnAuthorisedUser)
	}
	// RBAC block ends here
	return nil
}

func (handler PipelineTriggerRestHandlerImpl) OverrideConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var overrideRequest bean.ValuesOverrideRequest
	err = decoder.Decode(&overrideRequest)
	if err != nil {
		handler.logger.Errorw("request err, OverrideConfig", "err", err, "payload", overrideRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	overrideRequest.UserId = userId
	handler.logger.Infow("request for OverrideConfig", "payload", overrideRequest)
	err = handler.validator.Struct(overrideRequest)
	if err != nil {
		handler.logger.Errorw("request err, OverrideConfig", "err", err, "payload", overrideRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	if rbacErr := handler.validateCdTriggerRBAC(token, overrideRequest.AppId, overrideRequest.PipelineId); rbacErr != nil {
		common.WriteJsonResp(w, rbacErr, nil, http.StatusForbidden)
		return
	}
	ctx := r.Context()
	_, span := otel.Tracer("orchestrator").Start(ctx, "workflowDagExecutor.ManualCdTrigger")
	triggerContext := bean3.TriggerContext{
		Context: ctx,
	}
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	userEmail := util.GetEmailFromContext(ctx)
	userMetadata := &bean5.UserMetadata{
		UserEmailId:      userEmail,
		IsUserSuperAdmin: isSuperAdmin,
		UserId:           userId,
	}
	mergeResp, helmPackageName, _, err := handler.cdHandlerService.ManualCdTrigger(triggerContext, &overrideRequest, userMetadata)
	span.End()
	if err != nil {
		handler.logger.Errorw("request err, OverrideConfig", "err", err, "payload", overrideRequest)
		common.WriteJsonResp(w, err, err.Error(), http.StatusInternalServerError)
		return
	}
	res := map[string]interface{}{"releaseId": mergeResp, "helmPackageName": helmPackageName}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) RotatePods(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var podRotateRequest bean2.PodRotateRequest
	err = decoder.Decode(&podRotateRequest)
	if err != nil {
		handler.logger.Errorw("request err, RotatePods", "err", err, "payload", podRotateRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	podRotateRequest.UserId = userId
	handler.logger.Infow("request payload, RotatePods", "err", err, "payload", podRotateRequest)
	err = handler.validator.Struct(podRotateRequest)
	if err != nil {
		handler.logger.Errorw("validation err, RotatePods", "err", err, "payload", podRotateRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(podRotateRequest.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(podRotateRequest.AppId, podRotateRequest.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	userEmail := util.GetEmailFromContext(r.Context())
	userMetadata := &bean5.UserMetadata{
		UserEmailId:      userEmail,
		IsUserSuperAdmin: isSuperAdmin,
		UserId:           userId,
	}
	rotatePodResponse, err := handler.deployedAppService.RotatePods(r.Context(), &podRotateRequest, userMetadata)
	if err != nil {
		handler.logger.Errorw("service err, RotatePods", "err", err, "payload", podRotateRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, rotatePodResponse, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) StartStopApp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var overrideRequest bean2.StopAppRequest
	err = decoder.Decode(&overrideRequest)
	if err != nil {
		handler.logger.Errorw("request err, StartStopApp", "err", err, "payload", overrideRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	overrideRequest.UserId = userId
	handler.logger.Infow("request payload, StartStopApp", "err", err, "payload", overrideRequest)
	err = handler.validator.Struct(overrideRequest)
	if err != nil {
		handler.logger.Errorw("validation err, StartStopApp", "err", err, "payload", overrideRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	//rbac block starts from here
	object := handler.enforcerUtil.GetAppRBACNameByAppId(overrideRequest.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(overrideRequest.AppId, overrideRequest.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback block ends here
	ctx := r.Context()
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	userEmail := util.GetEmailFromContext(ctx)
	userMetadata := &bean5.UserMetadata{
		UserEmailId:      userEmail,
		IsUserSuperAdmin: isSuperAdmin,
		UserId:           userId,
	}
	mergeResp, err := handler.deployedAppService.StopStartApp(ctx, &overrideRequest, userMetadata)
	if err != nil {
		handler.logger.Errorw("service err, StartStopApp", "err", err, "payload", overrideRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	res := map[string]interface{}{"releaseId": mergeResp}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) StartStopDeploymentGroup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var stopDeploymentGroupRequest bean4.StopDeploymentGroupRequest
	err = decoder.Decode(&stopDeploymentGroupRequest)
	if err != nil {
		handler.logger.Errorw("request err, StartStopDeploymentGroup", "err", err, "payload", stopDeploymentGroupRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	stopDeploymentGroupRequest.UserId = userId
	err = handler.validator.Struct(stopDeploymentGroupRequest)
	if err != nil {
		handler.logger.Errorw("validation err, StartStopDeploymentGroup", "err", err, "payload", stopDeploymentGroupRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, StartStopDeploymentGroup", "err", err, "payload", stopDeploymentGroupRequest)

	//rbac block starts from here
	dg, err := handler.deploymentGroupService.GetDeploymentGroupById(stopDeploymentGroupRequest.DeploymentGroupId)
	if err != nil {
		handler.logger.Errorw("request err, StartStopDeploymentGroup", "err", err, "payload", stopDeploymentGroupRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	token := r.Header.Get("token")
	// RBAC enforcer applying
	object := handler.enforcerUtil.GetTeamRBACByCiPipelineId(dg.CiPipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByCiPipelineIdAndEnvId(dg.CiPipelineId, dg.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback block ends here
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	userEmail := util.GetEmailFromContext(r.Context())
	userMetadata := &bean5.UserMetadata{
		UserEmailId:      userEmail,
		IsUserSuperAdmin: isSuperAdmin,
		UserId:           userId,
	}
	res, err := handler.workflowEventPublishService.TriggerBulkHibernateAsync(stopDeploymentGroupRequest, userMetadata)
	if err != nil {
		handler.logger.Errorw("service err, StartStopDeploymentGroup", "err", err, "payload", stopDeploymentGroupRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) ReleaseStatusUpdate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var releaseStatusUpdateRequest bean.ReleaseStatusUpdateRequest
	err = decoder.Decode(&releaseStatusUpdateRequest)
	if err != nil {
		handler.logger.Errorw("request err, ReleaseStatusUpdate", "err", err, "payload", releaseStatusUpdateRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, ReleaseStatusUpdate, override request ----", "err", err, "payload", releaseStatusUpdateRequest)
	res, err := handler.appService.UpdateReleaseStatus(&releaseStatusUpdateRequest)
	if err != nil {
		handler.logger.Errorw("service err, ReleaseStatusUpdate", "err", err, "payload", releaseStatusUpdateRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	m := map[string]bool{
		"status": res}
	resJson, err := json.Marshal(m)
	if err != nil {
		handler.logger.Errorw("marshal err, ReleaseStatusUpdate", "err", err, "payload", m)
	}
	common.WriteJsonResp(w, err, resJson, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) GetAllLatestDeploymentConfiguration(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	handler.logger.Infow("request payload, GetAllLatestDeploymentConfiguration")

	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, GetAllDeployedConfigurationHistoryForSpecificWfrIdForPipeline", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, GetAllDeployedConfigurationHistoryForSpecificWfrIdForPipeline", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//RBAC START
	token := r.Header.Get("token")
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	ctx := r.Context()
	ctx = util.SetSuperAdminInContext(ctx, isSuperAdmin)
	//checking if user has admin access
	userHasAdminAccess := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, resourceName)
	allDeploymentconfig, err := handler.deploymentConfigService.GetLatestDeploymentConfigurationByPipelineId(ctx, pipelineId, userHasAdminAccess)
	if err != nil {
		handler.logger.Errorw("error in getting latest deployment config, GetAllDeployedConfigurationHistoryForSpecificWfrIdForPipeline", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, allDeploymentconfig, http.StatusOK)
}
