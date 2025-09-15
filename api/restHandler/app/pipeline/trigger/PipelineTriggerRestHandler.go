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

package trigger

import (
	"encoding/json"
	"net/http"

	util2 "github.com/devtron-labs/devtron/internal/util"
	bean5 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/deployedApp/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/out"
	bean4 "github.com/devtron-labs/devtron/pkg/eventProcessor/out/bean"

	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util"
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
		return util2.NewApiError(http.StatusForbidden,
			"Access denied for application trigger operation",
			"forbidden").WithCode("11011")
	}
	object = handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(appId, cdPipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, object); !ok {
		return util2.NewApiError(http.StatusForbidden,
			"Access denied for environment trigger operation",
			"forbidden").WithCode("11011")
	}
	// RBAC block ends here
	return nil
}

func (handler PipelineTriggerRestHandlerImpl) OverrideConfig(w http.ResponseWriter, r *http.Request) {
	// 1. Authentication check
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	// 2. Request body parsing
	decoder := json.NewDecoder(r.Body)
	var overrideRequest bean.ValuesOverrideRequest
	err = decoder.Decode(&overrideRequest)
	if err != nil {
		handler.logger.Errorw("request parsing error", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	overrideRequest.UserId = userId

	// 3. Struct validation
	err = handler.validator.Struct(overrideRequest)
	if err != nil {
		handler.logger.Errorw("validation error", "err", err, "payload", overrideRequest)
		// Enhanced validation error handling
		common.HandleValidationErrors(w, r, err)
		return
	}

	// 4. RBAC validation
	token := r.Header.Get("token")
	if rbacErr := handler.validateCdTriggerRBAC(token, overrideRequest.AppId, overrideRequest.PipelineId); rbacErr != nil {
		common.WriteJsonResp(w, rbacErr, nil, rbacErr.(*util2.ApiError).HttpStatusCode)
		return
	}
	// 5. Service call with enhanced error handling
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
		handler.logger.Errorw("service error", "err", err, "payload", overrideRequest)

		// Use enhanced error response builder
		errBuilder := common.NewErrorResponseBuilder(w, r).
			WithOperation("CD pipeline trigger").
			WithResourceFromId("pipeline", overrideRequest.PipelineId)
		errBuilder.HandleError(err)
		return
	}

	// 6. Success response
	res := map[string]interface{}{"releaseId": mergeResp, "helmPackageName": helmPackageName}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) RotatePods(w http.ResponseWriter, r *http.Request) {
	// 1. Authentication check
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	// 2. Request body parsing
	decoder := json.NewDecoder(r.Body)
	var podRotateRequest bean2.PodRotateRequest
	err = decoder.Decode(&podRotateRequest)
	if err != nil {
		handler.logger.Errorw("request parsing error", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	podRotateRequest.UserId = userId

	// 3. Struct validation
	err = handler.validator.Struct(podRotateRequest)
	if err != nil {
		handler.logger.Errorw("validation error", "err", err, "payload", podRotateRequest)
		// Enhanced validation error handling
		common.HandleValidationErrors(w, r, err)
		return
	}
	// 4. RBAC validation
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(podRotateRequest.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
		common.WriteForbiddenError(w, "pod rotation", "application")
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(podRotateRequest.AppId, podRotateRequest.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, object); !ok {
		common.WriteForbiddenError(w, "pod rotation", "environment")
		return
	}
	// 5. Service call with enhanced error handling
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	userEmail := util.GetEmailFromContext(r.Context())
	userMetadata := &bean5.UserMetadata{
		UserEmailId:      userEmail,
		IsUserSuperAdmin: isSuperAdmin,
		UserId:           userId,
	}
	rotatePodResponse, err := handler.deployedAppService.RotatePods(r.Context(), &podRotateRequest, userMetadata)
	if err != nil {
		handler.logger.Errorw("service error", "err", err, "payload", podRotateRequest)

		// Use enhanced error response builder
		errBuilder := common.NewErrorResponseBuilder(w, r).
			WithOperation("pod rotation").
			WithResourceFromId("application", podRotateRequest.AppId)
		errBuilder.HandleError(err)
		return
	}

	// 6. Success response
	common.WriteJsonResp(w, nil, rotatePodResponse, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) StartStopApp(w http.ResponseWriter, r *http.Request) {
	// 1. Authentication check
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	// 2. Request body parsing
	decoder := json.NewDecoder(r.Body)
	var overrideRequest bean2.StopAppRequest
	err = decoder.Decode(&overrideRequest)
	if err != nil {
		handler.logger.Errorw("request parsing error", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	overrideRequest.UserId = userId

	// 3. Struct validation
	err = handler.validator.Struct(overrideRequest)
	if err != nil {
		handler.logger.Errorw("validation error", "err", err, "payload", overrideRequest)
		// Enhanced validation error handling
		common.HandleValidationErrors(w, r, err)
		return
	}
	// 4. RBAC validation
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(overrideRequest.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
		common.WriteForbiddenError(w, "start/stop", "application")
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByAppId(overrideRequest.AppId, overrideRequest.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, object); !ok {
		common.WriteForbiddenError(w, "start/stop", "environment")
		return
	}
	// 5. Service call with enhanced error handling
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
		handler.logger.Errorw("service error", "err", err, "payload", overrideRequest)

		// Use enhanced error response builder
		errBuilder := common.NewErrorResponseBuilder(w, r).
			WithOperation("application start/stop").
			WithResourceFromId("application", overrideRequest.AppId)
		errBuilder.HandleError(err)
		return
	}

	// 6. Success response
	res := map[string]interface{}{"releaseId": mergeResp}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) StartStopDeploymentGroup(w http.ResponseWriter, r *http.Request) {
	// 1. Authentication check
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	// 2. Request body parsing
	decoder := json.NewDecoder(r.Body)
	var stopDeploymentGroupRequest bean4.StopDeploymentGroupRequest
	err = decoder.Decode(&stopDeploymentGroupRequest)
	if err != nil {
		handler.logger.Errorw("request parsing error", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	stopDeploymentGroupRequest.UserId = userId

	// 3. Struct validation
	err = handler.validator.Struct(stopDeploymentGroupRequest)
	if err != nil {
		handler.logger.Errorw("validation error", "err", err, "payload", stopDeploymentGroupRequest)
		// Enhanced validation error handling
		common.HandleValidationErrors(w, r, err)
		return
	}

	// 4. Deployment group retrieval and RBAC validation
	dg, err := handler.deploymentGroupService.GetDeploymentGroupById(stopDeploymentGroupRequest.DeploymentGroupId)
	if err != nil {
		handler.logger.Errorw("deployment group retrieval error", "err", err, "deploymentGroupId", stopDeploymentGroupRequest.DeploymentGroupId)

		// Use enhanced error response with resource context
		common.WriteJsonRespWithResourceContextFromId(w, err, nil, 0, "deployment group", stopDeploymentGroupRequest.DeploymentGroupId)
		return
	}

	token := r.Header.Get("token")
	// RBAC enforcer applying
	object := handler.enforcerUtil.GetTeamRBACByCiPipelineId(dg.CiPipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
		common.WriteForbiddenError(w, "deployment group start/stop", "application")
		return
	}
	object = handler.enforcerUtil.GetEnvRBACNameByCiPipelineIdAndEnvId(dg.CiPipelineId, dg.EnvironmentId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, object); !ok {
		common.WriteForbiddenError(w, "deployment group start/stop", "environment")
		return
	}
	// 5. Service call with enhanced error handling
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	userEmail := util.GetEmailFromContext(r.Context())
	userMetadata := &bean5.UserMetadata{
		UserEmailId:      userEmail,
		IsUserSuperAdmin: isSuperAdmin,
		UserId:           userId,
	}
	res, err := handler.workflowEventPublishService.TriggerBulkHibernateAsync(stopDeploymentGroupRequest, userMetadata)
	if err != nil {
		handler.logger.Errorw("service error", "err", err, "payload", stopDeploymentGroupRequest)

		// Use enhanced error response builder
		errBuilder := common.NewErrorResponseBuilder(w, r).
			WithOperation("deployment group start/stop").
			WithResourceFromId("deployment group", stopDeploymentGroupRequest.DeploymentGroupId)
		errBuilder.HandleError(err)
		return
	}

	// 6. Success response
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) ReleaseStatusUpdate(w http.ResponseWriter, r *http.Request) {
	// 1. Authentication check
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	// 2. Request body parsing
	decoder := json.NewDecoder(r.Body)
	var releaseStatusUpdateRequest bean.ReleaseStatusUpdateRequest
	err = decoder.Decode(&releaseStatusUpdateRequest)
	if err != nil {
		handler.logger.Errorw("request parsing error", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// 3. Struct validation (if validator is available for this struct)
	err = handler.validator.Struct(releaseStatusUpdateRequest)
	if err != nil {
		handler.logger.Errorw("validation error", "err", err, "payload", releaseStatusUpdateRequest)
		// Enhanced validation error handling
		common.HandleValidationErrors(w, r, err)
		return
	}

	// 4. Service call with enhanced error handling
	res, err := handler.appService.UpdateReleaseStatus(&releaseStatusUpdateRequest)
	if err != nil {
		handler.logger.Errorw("service error", "err", err, "payload", releaseStatusUpdateRequest)

		// Use enhanced error response builder
		errBuilder := common.NewErrorResponseBuilder(w, r).
			WithOperation("release status update")
		errBuilder.HandleError(err)
		return
	}

	// 5. Success response
	response := map[string]bool{"status": res}
	common.WriteJsonResp(w, nil, response, http.StatusOK)
}

func (handler PipelineTriggerRestHandlerImpl) GetAllLatestDeploymentConfiguration(w http.ResponseWriter, r *http.Request) {
	// 1. Authentication check
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	// 2. Enhanced path parameter extraction
	appId, err := common.ExtractIntPathParamWithContext(w, r, "appId")
	if err != nil {
		// Error already written by ExtractIntPathParamWithContext
		return
	}

	pipelineId, err := common.ExtractIntPathParamWithContext(w, r, "pipelineId")
	if err != nil {
		// Error already written by ExtractIntPathParamWithContext
		return
	}
	// 3. RBAC validation
	token := r.Header.Get("token")
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteForbiddenError(w, "view", "application deployment configuration")
		return
	}

	// 4. Service call with enhanced error handling
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	ctx := r.Context()
	ctx = util.SetSuperAdminInContext(ctx, isSuperAdmin)
	//checking if user has admin access
	userHasAdminAccess := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, resourceName)
	allDeploymentconfig, err := handler.deploymentConfigService.GetLatestDeploymentConfigurationByPipelineId(ctx, pipelineId, userHasAdminAccess)
	if err != nil {
		handler.logger.Errorw("service error", "err", err, "pipelineId", pipelineId)

		// Use enhanced error response with resource context
		common.WriteJsonRespWithResourceContextFromId(w, err, nil, 0, "pipeline", pipelineId)
		return
	}

	// 5. Success response
	common.WriteJsonResp(w, nil, allDeploymentconfig, http.StatusOK)
}
