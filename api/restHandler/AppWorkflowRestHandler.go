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
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	appWorkflow2 "github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/util"
	bean3 "github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	resourceGroup2 "github.com/devtron-labs/devtron/pkg/resourceGroup"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type AppWorkflowRestHandler interface {
	CreateAppWorkflow(w http.ResponseWriter, r *http.Request)
	FindAppWorkflow(w http.ResponseWriter, r *http.Request)
	DeleteAppWorkflow(w http.ResponseWriter, r *http.Request)
	FindAllWorkflows(w http.ResponseWriter, r *http.Request)
	FindAppWorkflowByEnvironment(w http.ResponseWriter, r *http.Request)
	GetWorkflowsViewData(w http.ResponseWriter, r *http.Request)
	FindAllWorkflowsForApps(w http.ResponseWriter, r *http.Request)
}

type AppWorkflowRestHandlerImpl struct {
	Logger             *zap.SugaredLogger
	appWorkflowService appWorkflow.AppWorkflowService
	userAuthService    user.UserService
	teamService        team.TeamService
	enforcer           casbin.Enforcer
	pipelineBuilder    pipeline.PipelineBuilder
	appRepository      app.AppRepository
	enforcerUtil       rbac.EnforcerUtil
}

func NewAppWorkflowRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService, appWorkflowService appWorkflow.AppWorkflowService,
	teamService team.TeamService, enforcer casbin.Enforcer, pipelineBuilder pipeline.PipelineBuilder,
	appRepository app.AppRepository, enforcerUtil rbac.EnforcerUtil) *AppWorkflowRestHandlerImpl {
	return &AppWorkflowRestHandlerImpl{
		Logger:             Logger,
		appWorkflowService: appWorkflowService,
		userAuthService:    userAuthService,
		teamService:        teamService,
		enforcer:           enforcer,
		pipelineBuilder:    pipelineBuilder,
		appRepository:      appRepository,
		enforcerUtil:       enforcerUtil,
	}
}

func (handler AppWorkflowRestHandlerImpl) CreateAppWorkflow(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	handler.Logger.Debugw("request by user", "userId", userId)
	if userId == 0 || err != nil {
		return
	}
	var request appWorkflow.AppWorkflowDto

	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("decode err", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	//rbac block starts from here
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(request.AppId)
	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, resourceName, casbin.ActionCreate)
	if !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback block ends here
	request.UserId = userId

	res, err := handler.appWorkflowService.CreateAppWorkflow(request)
	if err != nil {
		if err.Error() == bean3.WORKFLOW_EXIST_ERROR {
			common.WriteJsonResp(w, err, bean3.WORKFLOW_EXIST_ERROR, http.StatusBadRequest)
			return
		}
		handler.Logger.Errorw("error on creating", "err", err)
		common.WriteJsonResp(w, err, []byte("Creation Failed"), http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler AppWorkflowRestHandlerImpl) DeleteAppWorkflow(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	handler.Logger.Debugw("request by user", "userId", userId)
	if userId == 0 || err != nil {
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["app-id"])
	if err != nil {
		handler.Logger.Errorw("bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	appWorkflowId, err := strconv.Atoi(vars["app-wf-id"])
	if err != nil {
		handler.Logger.Errorw("bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appWorkflow, err := handler.appWorkflowService.FindAppWorkflowById(appWorkflowId, appId)
	if err != nil {
		handler.Logger.Errorw("error in finding appWorkflow by appWorkflowId and appId", "err", err, "appWorkflowId", appWorkflowId, "appid", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	token := r.Header.Get("token")
	//rbac block starts from here
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	workflowResourceName := handler.enforcerUtil.GetRbacObjectNameByAppIdAndWorkflow(appId, appWorkflow.Name)
	ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionDelete, resourceName)
	if !ok {
		ok = handler.enforcer.Enforce(token, casbin.ResourceJobs, casbin.ActionDelete, resourceName) && handler.enforcer.Enforce(token, casbin.ResourceWorkflow, casbin.ActionDelete, workflowResourceName)
	}
	if !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//rback block ends here

	err = handler.appWorkflowService.DeleteAppWorkflow(appWorkflowId, userId)
	if err != nil {
		if _, ok := err.(*util.ApiError); ok {
			handler.Logger.Warnw("error on deleting", "err", err)
			common.WriteJsonResp(w, err, []byte("Deletion Failed"), http.StatusOK)
			return
		} else {
			handler.Logger.Errorw("error on deleting", "err", err)
		}
		common.WriteJsonResp(w, err, []byte("Creation Failed"), http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (impl AppWorkflowRestHandlerImpl) FindAppWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["app-id"])
	if err != nil {
		impl.Logger.Errorw("bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := impl.pipelineBuilder.GetApp(appId)
	if err != nil {
		impl.Logger.Errorw("bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	v := r.URL.Query()
	envIdsString := v.Get("envIds")
	envIds := make([]int, 0)
	if len(envIdsString) > 0 {
		envIds, err = util2.SplitCommaSeparatedIntValues(envIdsString)
		if err != nil {
			common.WriteJsonResp(w, err, "please provide valid envIds", http.StatusBadRequest)
			return
		}
	}

	// RBAC enforcer applying
	object := impl.enforcerUtil.GetAppRBACName(app.AppName)
	impl.Logger.Debugw("rbac object for other environment list", "object", object)
	ok := impl.enforcerUtil.CheckAppRbacForAppOrJob(token, object, casbin.ActionGet)
	if !ok {
		common.WriteJsonResp(w, err, "unauthorized user", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	workflows := make(map[string]interface{})
	workflowsList, err := impl.appWorkflowService.FindAppWorkflows(appId)
	if err != nil {
		impl.Logger.Errorw("error in fetching workflows for app", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	if len(envIds) > 0 {
		cdPipelineWfData, err := impl.appWorkflowService.FindCdPipelinesByAppId(appId)
		if err != nil {
			impl.Logger.Errorw("error in fetching cdPipelines for app", "appId", appId, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		triggerViewPayload := &appWorkflow.TriggerViewWorkflowConfig{
			Workflows:   workflowsList,
			CdPipelines: cdPipelineWfData,
		}
		//filter based on envIds
		response, err := impl.appWorkflowService.FilterWorkflows(triggerViewPayload, envIds)
		if err != nil {
			impl.Logger.Errorw("service err, FindAppWorkflow", "appId", appId, "envIds", envIds, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		workflowsList = response.Workflows
	}

	workflows["appId"] = app.Id
	workflows["appName"] = app.AppName
	if len(workflowsList) > 0 && app.AppType == helper.Job {
		// RBAC
		userEmailId, err := impl.userAuthService.GetEmailFromToken(token)
		if err != nil {
			impl.Logger.Errorw("error in getting user emailId from token", "err", err)
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
			return
		}

		var workflowNames []string
		var workflowIds []int
		var updatedWorkflowList []appWorkflow.AppWorkflowDto
		var rbacObjects []string
		workNameObjectMap := make(map[string]appWorkflow.AppWorkflowDto)

		for _, workflow := range workflowsList {
			workflowNames = append(workflowNames, workflow.Name)
			workflowIds = append(workflowIds, workflow.Id)
		}
		workflowIdToObjectMap := impl.enforcerUtil.GetAllWorkflowRBACObjectsByAppId(appId, workflowNames, workflowIds)
		itr := 0
		for _, val := range workflowIdToObjectMap {
			rbacObjects = append(rbacObjects, val)
			workNameObjectMap[val] = workflowsList[itr]
			itr++
		}

		enforcedMap := impl.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceWorkflow, casbin.ActionGet, rbacObjects)
		for obj, passed := range enforcedMap {
			if passed {
				updatedWorkflowList = append(updatedWorkflowList, workNameObjectMap[obj])
			}
		}
		if len(updatedWorkflowList) == 0 {
			updatedWorkflowList = []appWorkflow.AppWorkflowDto{}
		}
		workflows[bean3.Workflows] = updatedWorkflowList
	} else if len(workflowsList) > 0 {
		workflows[bean3.Workflows] = workflowsList
	} else {
		workflows[bean3.Workflows] = []appWorkflow.AppWorkflowDto{}
	}
	common.WriteJsonResp(w, err, workflows, http.StatusOK)
}

func (impl AppWorkflowRestHandlerImpl) FindAllWorkflows(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		impl.Logger.Errorw("Bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := impl.pipelineBuilder.GetApp(appId)
	if err != nil {
		impl.Logger.Errorw("Bad request, invalid appId", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	object := impl.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized user", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	resp, err := impl.appWorkflowService.FindAllWorkflowsComponentDetails(appId)
	if err != nil {
		impl.Logger.Errorw("error in getting all wf component details by appId", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (impl AppWorkflowRestHandlerImpl) FindAllWorkflowsForApps(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, err, "Unauthorized user", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	var request appWorkflow.WorkflowNamesRequest
	err = decoder.Decode(&request)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	resp, err := impl.appWorkflowService.FindAllWorkflowsForApps(request)
	if err != nil {
		impl.Logger.Errorw("error in getting all wf component details by appId", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (impl AppWorkflowRestHandlerImpl) FindAppWorkflowByEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	user, err := impl.userAuthService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	userEmailId := strings.ToLower(user.EmailId)
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		impl.Logger.Errorw("bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	v := r.URL.Query()
	appIdsString := v.Get("appIds")
	var appIds []int
	if len(appIdsString) > 0 {
		appIdsSlices := strings.Split(appIdsString, ",")
		for _, appId := range appIdsSlices {
			id, err := strconv.Atoi(appId)
			if err != nil {
				common.WriteJsonResp(w, err, "please provide valid appIds", http.StatusBadRequest)
				return
			}
			appIds = append(appIds, id)
		}
	}
	var appGroupId int
	appGroupIdStr := v.Get("appGroupId")
	if len(appGroupIdStr) > 0 {
		appGroupId, err = strconv.Atoi(appGroupIdStr)
		if err != nil {
			common.WriteJsonResp(w, err, "please provide valid appGroupId", http.StatusBadRequest)
			return
		}
	}

	request := resourceGroup2.ResourceGroupingRequest{
		ParentResourceId:  envId,
		ResourceGroupId:   appGroupId,
		ResourceGroupType: resourceGroup2.APP_GROUP,
		ResourceIds:       appIds,
		EmailId:           userEmailId,
		CheckAuthBatch:    impl.checkAuthBatch,
		UserId:            userId,
		Ctx:               r.Context(),
	}
	workflows := make(map[string]interface{})
	_, span := otel.Tracer("orchestrator").Start(r.Context(), "ciHandler.FetchAppWorkflowsInAppGrouping")
	workflowsList, err := impl.appWorkflowService.FindAppWorkflowsByEnvironmentId(request)
	span.End()
	if err != nil {
		impl.Logger.Errorw("error in fetching workflows for app", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	workflows["envId"] = envId
	if len(workflowsList) > 0 {
		workflows["workflows"] = workflowsList
	} else {
		workflows["workflows"] = []appWorkflow.AppWorkflowDto{}
	}
	common.WriteJsonResp(w, err, workflows, http.StatusOK)
}

func (handler *AppWorkflowRestHandlerImpl) GetWorkflowsViewData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["app-id"])
	if err != nil {
		handler.Logger.Errorw("error in parsing app-id", "appId", appId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("error in getting app details", "appId", appId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query()
	envIdsString := v.Get("envIds")
	envIds := make([]int, 0)
	if len(envIdsString) > 0 {
		envIds, err = util2.SplitCommaSeparatedIntValues(envIdsString)
		if err != nil {
			common.WriteJsonResp(w, err, "please provide valid envIds", http.StatusBadRequest)
			return
		}
	}

	// RBAC enforcer applying
	object := handler.enforcerUtil.GetAppRBACName(app.AppName)
	handler.Logger.Debugw("rbac object for workflows view data", "object", object)
	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, object, casbin.ActionGet)
	if !ok {
		common.WriteJsonResp(w, err, "unauthorized user", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	appWorkflows, err := handler.appWorkflowService.FindAppWorkflows(appId)
	if err != nil {
		handler.Logger.Errorw("error in fetching workflows for app", "appId", appId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	ciPipelineViewData, err := handler.pipelineBuilder.GetTriggerViewCiPipeline(appId)
	if err != nil {
		if _, ok := err.(*util.ApiError); !ok {
			handler.Logger.Errorw("error in fetching trigger view ci pipeline data for app", "appId", appId, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		} else {
			err = nil
		}
	}

	cdPipelinesForApp, err := handler.pipelineBuilder.GetTriggerViewCdPipelinesForApp(appId)
	if err != nil {
		if _, ok := err.(*util.ApiError); !ok {
			handler.Logger.Errorw("error in fetching trigger view cd pipeline data for app", "appId", appId, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		} else {
			err = nil
		}
	}

	containsExternalCi := handler.containsExternalCi(appWorkflows)
	var externalCiData []*bean.ExternalCiConfig
	if containsExternalCi {
		externalCiData, err = handler.pipelineBuilder.GetExternalCi(appId)
		if err != nil {
			handler.Logger.Errorw("service err, GetExternalCi", "appId", appId, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}

	response := &appWorkflow.TriggerViewWorkflowConfig{
		Workflows:        appWorkflows,
		CiConfig:         ciPipelineViewData,
		CdPipelines:      cdPipelinesForApp,
		ExternalCiConfig: externalCiData,
	}

	if len(envIds) > 0 {
		//filter based on envIds
		response, err = handler.appWorkflowService.FilterWorkflows(response, envIds)
		if err != nil {
			handler.Logger.Errorw("service err, FilterResponseOnEnvIds", "appId", appId, "envIds", envIds, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}

	common.WriteJsonResp(w, err, response, http.StatusOK)
}

func (handler *AppWorkflowRestHandlerImpl) checkAuthBatch(emailId string, appObject []string, envObject []string) (map[string]bool, map[string]bool) {
	var appResult map[string]bool
	var envResult map[string]bool
	if len(appObject) > 0 {
		appResult = handler.enforcer.EnforceByEmailInBatch(emailId, casbin.ResourceApplications, casbin.ActionGet, appObject)
	}
	if len(envObject) > 0 {
		envResult = handler.enforcer.EnforceByEmailInBatch(emailId, casbin.ResourceEnvironment, casbin.ActionGet, envObject)
	}
	return appResult, envResult
}

func (handler AppWorkflowRestHandlerImpl) containsExternalCi(appWorkflows []appWorkflow.AppWorkflowDto) bool {
	for _, appWorkflowDto := range appWorkflows {
		for _, workflowMappingDto := range appWorkflowDto.AppWorkflowMappingDto {
			if workflowMappingDto.Type == appWorkflow2.WEBHOOK {
				return true
			}
		}
	}
	return false
}
