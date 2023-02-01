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
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/cron"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	service1 "github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/deploymentGroup"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AppListingRestHandler interface {
	FetchAppsByEnvironment(w http.ResponseWriter, r *http.Request)
	FetchAppDetails(w http.ResponseWriter, r *http.Request)
	FetchAllDevtronManagedApps(w http.ResponseWriter, r *http.Request)
	FetchAppTriggerView(w http.ResponseWriter, r *http.Request)
	FetchAppStageStatus(w http.ResponseWriter, r *http.Request)

	FetchOtherEnvironment(w http.ResponseWriter, r *http.Request)
	RedirectToLinkouts(w http.ResponseWriter, r *http.Request)
	GetHostUrlsByBatch(w http.ResponseWriter, r *http.Request)

	ManualSyncAcdPipelineDeploymentStatus(w http.ResponseWriter, r *http.Request)
}

type AppListingRestHandlerImpl struct {
	application                      application.ServiceClient
	appListingService                app.AppListingService
	teamService                      team.TeamService
	enforcer                         casbin.Enforcer
	pipeline                         pipeline.PipelineBuilder
	logger                           *zap.SugaredLogger
	enforcerUtil                     rbac.EnforcerUtil
	deploymentGroupService           deploymentGroup.DeploymentGroupService
	userService                      user.UserService
	helmAppClient                    client.HelmAppClient
	clusterService                   cluster.ClusterService
	helmAppService                   client.HelmAppService
	argoUserService                  argo.ArgoUserService
	k8sApplicationService            k8s.K8sApplicationService
	installedAppService              service1.InstalledAppService
	cdApplicationStatusUpdateHandler cron.CdApplicationStatusUpdateHandler
	pipelineRepository               pipelineConfig.PipelineRepository
	appStatusService                 appStatus.AppStatusService
}

type AppStatus struct {
	name       string
	status     string
	message    string
	err        error
	conditions []v1alpha1.ApplicationCondition
}

func NewAppListingRestHandlerImpl(application application.ServiceClient,
	appListingService app.AppListingService,
	teamService team.TeamService,
	enforcer casbin.Enforcer,
	pipeline pipeline.PipelineBuilder,
	logger *zap.SugaredLogger, enforcerUtil rbac.EnforcerUtil,
	deploymentGroupService deploymentGroup.DeploymentGroupService, userService user.UserService,
	helmAppClient client.HelmAppClient, clusterService cluster.ClusterService, helmAppService client.HelmAppService,
	argoUserService argo.ArgoUserService, k8sApplicationService k8s.K8sApplicationService, installedAppService service1.InstalledAppService,
	cdApplicationStatusUpdateHandler cron.CdApplicationStatusUpdateHandler,
	pipelineRepository pipelineConfig.PipelineRepository,
	appStatusService appStatus.AppStatusService) *AppListingRestHandlerImpl {
	appListingHandler := &AppListingRestHandlerImpl{
		application:                      application,
		appListingService:                appListingService,
		logger:                           logger,
		teamService:                      teamService,
		pipeline:                         pipeline,
		enforcer:                         enforcer,
		enforcerUtil:                     enforcerUtil,
		deploymentGroupService:           deploymentGroupService,
		userService:                      userService,
		helmAppClient:                    helmAppClient,
		clusterService:                   clusterService,
		helmAppService:                   helmAppService,
		argoUserService:                  argoUserService,
		k8sApplicationService:            k8sApplicationService,
		installedAppService:              installedAppService,
		cdApplicationStatusUpdateHandler: cdApplicationStatusUpdateHandler,
		pipelineRepository:               pipelineRepository,
		appStatusService:                 appStatusService,
	}
	return appListingHandler
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	(*w).Header().Set("Content-Type", "text/html; charset=utf-8")
}
func (handler AppListingRestHandlerImpl) FetchAllDevtronManagedApps(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	handler.logger.Infow("got request to fetch all devtron managed apps ", "userId", userId)
	//RBAC starts
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		handler.logger.Infow("user forbidden to fetch all devtron managed apps", "userId", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC ends
	res, err := handler.appListingService.FetchAllDevtronManagedApps()
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
func (handler AppListingRestHandlerImpl) FetchAppsByEnvironment(w http.ResponseWriter, r *http.Request) {
	//Allow CORS here By * or specific origin
	setupResponse(&w, r)
	token := r.Header.Get("token")
	t0 := time.Now()
	t1 := time.Now()
	handler.logger.Infow("api response time testing", "time", time.Now().String(), "stage", "1")
	newCtx, span := otel.Tracer("userService").Start(r.Context(), "GetLoggedInUser")
	userId, err := handler.userService.GetLoggedInUser(r)
	span.End()
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	newCtx, span = otel.Tracer("userService").Start(newCtx, "GetById")
	user, err := handler.userService.GetById(userId)
	span.End()
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	userEmailId := strings.ToLower(user.EmailId)
	var fetchAppListingRequest app.FetchAppListingRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&fetchAppListingRequest)
	if err != nil {
		handler.logger.Errorw("request err, FetchAppsByEnvironment", "err", err, "payload", fetchAppListingRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	newCtx, span = otel.Tracer("fetchAppListingRequest").Start(newCtx, "GetNamespaceClusterMapping")
	_, _, err = fetchAppListingRequest.GetNamespaceClusterMapping()
	span.End()
	if err != nil {
		handler.logger.Errorw("request err, GetNamespaceClusterMapping", "err", err, "payload", fetchAppListingRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	var dg *deploymentGroup.DeploymentGroupDTO
	if fetchAppListingRequest.DeploymentGroupId > 0 {
		newCtx, span = otel.Tracer("deploymentGroupService").Start(newCtx, "FindById")
		dg, err = handler.deploymentGroupService.FindById(fetchAppListingRequest.DeploymentGroupId)
		span.End()
		if err != nil {
			handler.logger.Errorw("service err, FetchAppsByEnvironment", "err", err, "payload", fetchAppListingRequest)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		}
	}

	newCtx, span = otel.Tracer("appListingService").Start(newCtx, "FetchAppsByEnvironment")
	envContainers, err := handler.appListingService.FetchAppsByEnvironment(fetchAppListingRequest, w, r, token)
	span.End()
	if err != nil {
		handler.logger.Errorw("service err, FetchAppsByEnvironment", "err", err, "payload", fetchAppListingRequest)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
	}
	t2 := time.Now()
	handler.logger.Infow("api response time testing", "time", time.Now().String(), "time diff", t2.Unix()-t1.Unix(), "stage", "2")
	t1 = t2

	newCtx, span = otel.Tracer("userService").Start(newCtx, "IsSuperAdmin")
	isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	span.End()
	if err != nil {
		handler.logger.Errorw("request err, FetchAppsByEnvironment", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}
	appEnvContainers := make([]*bean.AppEnvironmentContainer, 0)
	if isActionUserSuperAdmin {
		appEnvContainers = append(appEnvContainers, envContainers...)
	} else {
		uniqueTeams := make(map[int]string)
		authorizedTeams := make(map[int]bool)
		for _, envContainer := range envContainers {
			if _, ok := uniqueTeams[envContainer.TeamId]; !ok {
				uniqueTeams[envContainer.TeamId] = envContainer.TeamName
			}
		}

		objectArray := make([]string, len(uniqueTeams))
		for _, teamName := range uniqueTeams {
			object := strings.ToLower(teamName)
			objectArray = append(objectArray, object)
		}

		newCtx, span = otel.Tracer("enforcer").Start(newCtx, "EnforceByEmailInBatchForTeams")
		resultMap := handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceTeam, casbin.ActionGet, objectArray)
		span.End()
		for teamId, teamName := range uniqueTeams {
			object := strings.ToLower(teamName)
			if ok := resultMap[object]; ok {
				authorizedTeams[teamId] = true
			}
		}

		filteredAppEnvContainers := make([]*bean.AppEnvironmentContainer, 0)
		for _, envContainer := range envContainers {
			if _, ok := authorizedTeams[envContainer.TeamId]; ok {
				filteredAppEnvContainers = append(filteredAppEnvContainers, envContainer)
			}
		}

		objectArray = make([]string, len(filteredAppEnvContainers))
		for _, filteredAppEnvContainer := range filteredAppEnvContainers {
			if fetchAppListingRequest.DeploymentGroupId > 0 {
				if filteredAppEnvContainer.EnvironmentId != 0 && filteredAppEnvContainer.EnvironmentId != dg.EnvironmentId {
					continue
				}
			}
			object := fmt.Sprintf("%s/%s", filteredAppEnvContainer.TeamName, filteredAppEnvContainer.AppName)
			object = strings.ToLower(object)
			objectArray = append(objectArray, object)
		}

		newCtx, span = otel.Tracer("enforcer").Start(newCtx, "EnforceByEmailInBatchForApps")
		resultMap = handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceApplications, casbin.ActionGet, objectArray)
		span.End()
		for _, filteredAppEnvContainer := range filteredAppEnvContainers {
			if fetchAppListingRequest.DeploymentGroupId > 0 {
				if filteredAppEnvContainer.EnvironmentId != 0 && filteredAppEnvContainer.EnvironmentId != dg.EnvironmentId {
					continue
				}
			}
			object := fmt.Sprintf("%s/%s", filteredAppEnvContainer.TeamName, filteredAppEnvContainer.AppName)
			object = strings.ToLower(object)
			if ok := resultMap[object]; ok {
				appEnvContainers = append(appEnvContainers, filteredAppEnvContainer)
			}
		}

	}
	t2 = time.Now()
	handler.logger.Infow("api response time testing", "time", time.Now().String(), "time diff", t2.Unix()-t1.Unix(), "stage", "3")
	t1 = t2
	newCtx, span = otel.Tracer("appListingService").Start(newCtx, "BuildAppListingResponse")
	apps, err := handler.appListingService.BuildAppListingResponse(fetchAppListingRequest, appEnvContainers)
	span.End()
	if err != nil {
		handler.logger.Errorw("service err, FetchAppsByEnvironment", "err", err, "payload", fetchAppListingRequest)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
	}

	// Apply pagination
	appsCount := len(apps)
	offset := fetchAppListingRequest.Offset
	limit := fetchAppListingRequest.Size

	if offset+limit <= len(apps) {
		apps = apps[offset : offset+limit]
	} else {
		apps = apps[offset:]
	}

	appContainerResponse := bean.AppContainerResponse{
		AppContainers: apps,
		AppCount:      appsCount,
	}
	if fetchAppListingRequest.DeploymentGroupId > 0 {
		var ciMaterialDTOs []bean.CiMaterialDTO
		for _, ci := range dg.CiMaterialDTOs {
			ciMaterialDTOs = append(ciMaterialDTOs, bean.CiMaterialDTO{
				Name:        ci.Name,
				SourceValue: ci.SourceValue,
				SourceType:  ci.SourceType,
			})
		}
		appContainerResponse.DeploymentGroupDTO = bean.DeploymentGroupDTO{
			Id:             dg.Id,
			Name:           dg.Name,
			AppCount:       dg.AppCount,
			NoOfApps:       dg.NoOfApps,
			EnvironmentId:  dg.EnvironmentId,
			CiPipelineId:   dg.CiPipelineId,
			CiMaterialDTOs: ciMaterialDTOs,
		}
	}
	t2 = time.Now()
	handler.logger.Infow("api response time testing", "time", time.Now().String(), "time diff", t2.Unix()-t1.Unix(), "stage", "4")
	t1 = t2
	handler.logger.Infow("api response time testing", "total time", time.Now().String(), "total time", t1.Unix()-t0.Unix())
	common.WriteJsonResp(w, err, appContainerResponse, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) FetchAppDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := r.Header.Get("token")
	appId, err := strconv.Atoi(vars["app-id"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	envId, err := strconv.Atoi(vars["env-id"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appDetail, err := handler.appListingService.FetchAppDetails(r.Context(), appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, FetchAppDetails", "err", err, "appId", appId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	appDetail = handler.fetchResourceTree(w, r, appId, envId, appDetail)
	common.WriteJsonResp(w, err, appDetail, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) FetchAppTriggerView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := r.Header.Get("token")
	appId, err := strconv.Atoi(vars["app-id"])
	if err != nil {
		handler.logger.Errorw("request err, FetchAppTriggerView", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchAppTriggerView", "appId", appId)

	triggerView, err := handler.appListingService.FetchAppTriggerView(appId)
	if err != nil {
		handler.logger.Errorw("service err, FetchAppTriggerView", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//TODO: environment based auth, purge data of environment on which user doesnt have access, only show environment name
	// RBAC enforcer applying
	if len(triggerView) > 0 {
		object := handler.enforcerUtil.GetAppRBACName(triggerView[0].AppName)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	//RBAC enforcer Ends

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
		handler.logger.Errorw("error in getting acd token", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	ctx = context.WithValue(ctx, "token", acdToken)
	defer cancel()

	response := make(chan AppStatus)
	qCount := len(triggerView)
	responses := map[string]AppStatus{}

	for i := 0; i < len(triggerView); i++ {
		acdAppName := triggerView[i].AppName + "-" + triggerView[i].EnvironmentName
		go func(pipelineName string) {
			ctxt, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()
			query := application2.ApplicationQuery{Name: &pipelineName}
			app, conn, err := handler.application.Watch(ctxt, &query)
			defer conn.Close()
			if err != nil {
				response <- AppStatus{name: pipelineName, status: "", message: "", err: err, conditions: make([]v1alpha1.ApplicationCondition, 0)}
				return
			}
			if app != nil {
				resp, err := app.Recv()
				if err != nil {
					response <- AppStatus{name: pipelineName, status: "", message: "", err: err, conditions: make([]v1alpha1.ApplicationCondition, 0)}
					return
				}
				if resp != nil {
					healthStatus := resp.Application.Status.Health.Status
					status := AppStatus{
						name:       pipelineName,
						status:     string(healthStatus),
						message:    resp.Application.Status.Health.Message,
						err:        nil,
						conditions: resp.Application.Status.Conditions,
					}
					response <- status
					return
				}
				response <- AppStatus{name: pipelineName, status: "", message: "", err: fmt.Errorf("Missing Application"), conditions: make([]v1alpha1.ApplicationCondition, 0)}
				return
			}
			response <- AppStatus{name: pipelineName, status: "", message: "", err: fmt.Errorf("Connection Closed by Client"), conditions: make([]v1alpha1.ApplicationCondition, 0)}

		}(acdAppName)
	}
	rCount := 0

	for {
		select {
		case msg, ok := <-response:
			if ok {
				if msg.err == nil {
					responses[msg.name] = msg
				}
			}
			rCount++
		}
		if qCount == rCount {
			break
		}
	}

	for i := 0; i < len(triggerView); i++ {
		acdAppName := triggerView[i].AppName + "-" + triggerView[i].EnvironmentName
		if val, ok := responses[acdAppName]; ok {
			status := val.status
			conditions := val.conditions
			for _, condition := range conditions {
				if condition.Type != v1alpha1.ApplicationConditionSharedResourceWarning {
					status = "Degraded"
				}
			}
			triggerView[i].Status = status
			triggerView[i].StatusMessage = val.message
			triggerView[i].Conditions = val.conditions
		}
		if triggerView[i].Status == "" {
			triggerView[i].Status = "Unknown"
		}
		if triggerView[i].Status == string(health.HealthStatusDegraded) {
			triggerView[i].Status = "Not Deployed"
		}
	}
	common.WriteJsonResp(w, err, triggerView, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) FetchAppStageStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["app-id"])
	if err != nil {
		handler.logger.Errorw("request err, FetchAppStageStatus", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, FetchAppStageStatus", "appId", appId)
	token := r.Header.Get("token")
	app, err := handler.pipeline.GetApp(appId)
	if err != nil {
		handler.logger.Errorw("service err, FetchAppStageStatus", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	object := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	triggerView, err := handler.appListingService.FetchAppStageStatus(appId)
	if err != nil {
		handler.logger.Errorw("service err, FetchAppStageStatus", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, triggerView, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) FetchOtherEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["app-id"])
	if err != nil {
		handler.logger.Errorw("request err, FetchOtherEnvironment", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	newCtx, span := otel.Tracer("pipeline").Start(r.Context(), "GetApp")
	app, err := handler.pipeline.GetApp(appId)
	span.End()
	if err != nil {
		handler.logger.Errorw("service err, FetchOtherEnvironment", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	object := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "unauthorized user", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	newCtx, span = otel.Tracer("appListingService").Start(newCtx, "FetchOtherEnvironment")
	otherEnvironment, err := handler.appListingService.FetchOtherEnvironment(newCtx, appId)
	span.End()
	if err != nil {
		handler.logger.Errorw("service err, FetchOtherEnvironment", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//TODO - rbac env level

	common.WriteJsonResp(w, err, otherEnvironment, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) RedirectToLinkouts(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	Id, err := strconv.Atoi(vars["Id"])
	if err != nil {
		handler.logger.Errorw("request err, RedirectToLinkouts", "err", err, "id", Id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, RedirectToLinkouts", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.logger.Errorw("request err, RedirectToLinkouts", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	podName := vars["podName"]
	containerName := vars["containerName"]
	app, err := handler.pipeline.GetApp(appId)
	if err != nil {
		handler.logger.Errorw("bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	object := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "unauthorized user", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	link, err := handler.appListingService.RedirectToLinkouts(Id, appId, envId, podName, containerName)
	if err != nil || len(link) == 0 {
		handler.logger.Errorw("service err, RedirectToLinkouts", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, link, http.StatusOK)
}
func (handler AppListingRestHandlerImpl) fetchResourceTreeFromInstallAppService(w http.ResponseWriter, r *http.Request, appDetail bean.AppDetailContainer) bean.AppDetailContainer {
	rctx := r.Context()
	cn, _ := w.(http.CloseNotifier)
	return handler.installedAppService.FetchResourceTree(rctx, cn, &appDetail)
}
func (handler AppListingRestHandlerImpl) GetHostUrlsByBatch(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	appIdParam := vars.Get("appId")
	installedAppIdParam := vars.Get("installedAppId")
	envIdParam := vars.Get("envId")

	if (appIdParam == "" && installedAppIdParam == "") || (appIdParam != "" && installedAppIdParam != "") {
		handler.logger.Error("error in decoding batch request body", "appId", appIdParam, "installedAppId", installedAppIdParam)
		common.WriteJsonResp(w, fmt.Errorf("only one of the appId or installedAppId should be valid appId: %s installedAppId: %s", appIdParam, installedAppIdParam), nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	if appIdParam != "" {
		appId, err := strconv.Atoi(appIdParam)
		if err != nil {
			handler.logger.Errorw("error in parsing appId from request body", "appId", appIdParam, "err", err)
			common.WriteJsonResp(w, fmt.Errorf("error in parsing appId : %s must be integer", appIdParam), nil, http.StatusBadRequest)
			return
		}
		object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
			return
		}
	}
	var appDetail bean.AppDetailContainer
	var appId, envId int
	envId, err := strconv.Atoi(envIdParam)
	if err != nil {
		handler.logger.Errorw("error in parsing envId from request body", "envId", envIdParam, "err", err)
		common.WriteJsonResp(w, fmt.Errorf("error in parsing envId : %s must be integer", envIdParam), nil, http.StatusBadRequest)
		return
	}
	appDetail, err, appId = handler.getAppDetails(r.Context(), appIdParam, installedAppIdParam, envId)
	if err != nil {
		handler.logger.Errorw("error occurred while getting app details", "appId", appIdParam, "installedAppId", installedAppIdParam, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	isAcdApp := len(appDetail.AppName) > 0 && len(appDetail.EnvironmentName) > 0 && util.IsAcdApp(appDetail.DeploymentAppType)
	isHelmApp := len(appDetail.AppName) > 0 && len(appDetail.EnvironmentName) > 0 && util.IsHelmApp(appDetail.DeploymentAppType)
	if !isAcdApp && !isHelmApp {
		handler.logger.Errorw("Invalid app type", "appId", appIdParam, "envId", envId, "installedAppId", installedAppIdParam)
		common.WriteJsonResp(w, fmt.Errorf("app is neither helm app or devtron app"), nil, http.StatusBadRequest)
		return
	}
	//check user authorization for this app
	if installedAppIdParam != "" {
		object, object2 := handler.enforcerUtil.GetHelmObjectByAppNameAndEnvId(appDetail.AppName, appDetail.EnvironmentId)
		ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object2)
		if !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
			return
		}
	}

	if installedAppIdParam != "" {
		appDetail = handler.fetchResourceTreeFromInstallAppService(w, r, appDetail)
	} else {
		appDetail = handler.fetchResourceTree(w, r, appId, envId, appDetail)
	}

	resourceTree := appDetail.ResourceTree
	_, ok := resourceTree["nodes"]
	if !ok {
		err = fmt.Errorf("no nodes found for this resource tree appName:%s , envName:%s", appDetail.AppName, appDetail.EnvironmentName)
		handler.logger.Errorw("no nodes found for this resource tree", "appName", appDetail.AppName, "envName", appDetail.EnvironmentName)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//valid batch requests, only valid requests will be sent for batch processing
	validRequests := make([]k8s.ResourceRequestBean, 0)
	validRequests = handler.k8sApplicationService.FilterServiceAndIngress(r.Context(), resourceTree, validRequests, appDetail, "")
	if len(validRequests) == 0 {
		handler.logger.Error("neither service nor ingress found for", "appId", appIdParam, "envId", envIdParam, "installedAppId", installedAppIdParam)
		common.WriteJsonResp(w, err, nil, http.StatusNoContent)
		return
	}
	resp, err := handler.k8sApplicationService.GetManifestsByBatch(r.Context(), validRequests)
	if err != nil {
		handler.logger.Errorw("error in getting manifests in batch", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	result := handler.k8sApplicationService.GetUrlsByBatch(r.Context(), resp)
	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) getAppDetails(ctx context.Context, appIdParam, installedAppIdParam string, envId int) (bean.AppDetailContainer, error, int) {
	var appDetail bean.AppDetailContainer
	if appIdParam != "" {
		appId, err := strconv.Atoi(appIdParam)
		if err != nil {
			handler.logger.Errorw("error in parsing appId from request body", "appId", appIdParam, "err", err)
			return appDetail, err, appId
		}
		appDetail, err = handler.appListingService.FetchAppDetails(ctx, appId, envId)
		return appDetail, err, appId
	}

	appId, err := strconv.Atoi(installedAppIdParam)
	if err != nil {
		handler.logger.Errorw("error in parsing installedAppId from request body", "installedAppId", installedAppIdParam, "err", err)
		return appDetail, err, appId
	}
	appDetail, err = handler.installedAppService.FindAppDetailsForAppstoreApplication(appId, envId)
	return appDetail, err, appId
}

// TODO: move this to service
func (handler AppListingRestHandlerImpl) fetchResourceTree(w http.ResponseWriter, r *http.Request, appId int, envId int, appDetail bean.AppDetailContainer) bean.AppDetailContainer {
	if len(appDetail.AppName) > 0 && len(appDetail.EnvironmentName) > 0 && util.IsAcdApp(appDetail.DeploymentAppType) {
		//RBAC enforcer Ends
		cdPipelines, err := handler.pipelineRepository.FindActiveByAppIdAndEnvironmentId(appId, envId)
		if err != nil {
			handler.logger.Errorw("error in getting cdPipeline by appId and envId", "err", err, "appid", appId, "envId", envId)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
			return appDetail
		}
		if len(cdPipelines) != 1 {
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
			return appDetail
		}
		cdPipeline := cdPipelines[0]
		query := &application2.ResourcesQuery{
			ApplicationName: &cdPipeline.DeploymentAppName,
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
		defer cancel()
		acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			handler.logger.Errorw("error in getting acd token", "err", err)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
			return appDetail
		}
		ctx = context.WithValue(ctx, "token", acdToken)
		start := time.Now()
		resp, err := handler.application.ResourceTree(ctx, query)
		elapsed := time.Since(start)
		if err != nil {
			handler.logger.Errorw("service err, FetchAppDetails, resource tree", "err", err, "app", appId, "env", envId)
			err = &util.ApiError{
				Code:            constants.AppDetailResourceTreeNotFound,
				InternalMessage: "app detail fetched, failed to get resource tree from acd",
				UserMessage:     "Error fetching detail, if you have recently created this deployment pipeline please try after sometime.",
			}
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
			return appDetail
		}
		if resp.Status == string(health.HealthStatusHealthy) {
			status, err := handler.appListingService.ISLastReleaseStopType(appId, envId)
			if err != nil {
				handler.logger.Errorw("service err, FetchAppDetails", "err", err, "app", appId, "env", envId)
			} else if status {
				resp.Status = application.HIBERNATING
			}
		}
		handler.logger.Debugw("FetchAppDetails, time elapsed in fetching application for environment ", "elapsed", elapsed, "appId", appId, "envId", envId)

		if resp.Status == string(health.HealthStatusDegraded) {
			count, err := handler.appListingService.GetReleaseCount(appId, envId)
			if err != nil {
				handler.logger.Errorw("service err, FetchAppDetails, release count", "err", err, "app", appId, "env", envId)
			} else if count == 0 {
				resp.Status = app.NotDeployed
			}
		}
		appDetail.ResourceTree = util2.InterfaceToMapAdapter(resp)
		if resp.Status == string(health.HealthStatusHealthy) {
			err = handler.cdApplicationStatusUpdateHandler.SyncPipelineStatusForResourceTreeCall(cdPipeline)
			if err != nil {
				handler.logger.Errorw("error in syncing pipeline status", "err", err)
			}
		}
		//updating app_status table here
		err = handler.appStatusService.UpdateStatusWithAppIdEnvId(appDetail.AppId, appDetail.EnvironmentId, resp.Status)
		if err != nil {
			handler.logger.Warnw("error in updating app status", "err", err, "appId", appDetail.AppId, "envId", appDetail.EnvironmentId)
		}

	} else if len(appDetail.AppName) > 0 && len(appDetail.EnvironmentName) > 0 && util.IsHelmApp(appDetail.DeploymentAppType) {
		config, err := handler.helmAppService.GetClusterConf(appDetail.ClusterId)
		if err != nil {
			handler.logger.Errorw("error in fetching cluster detail", "err", err)
		}
		req := &client.AppDetailRequest{
			ClusterConfig: config,
			Namespace:     appDetail.Namespace,
			ReleaseName:   fmt.Sprintf("%s-%s", appDetail.AppName, appDetail.EnvironmentName),
		}
		detail, err := handler.helmAppClient.GetAppDetail(context.Background(), req)
		if err != nil {
			handler.logger.Errorw("error in fetching app detail", "err", err)
		}
		if detail != nil {
			resourceTree := util2.InterfaceToMapAdapter(detail.ResourceTreeResponse)
			resourceTree["status"] = detail.ApplicationStatus
			appDetail.ResourceTree = resourceTree
			handler.logger.Warnw("appName and envName not found - avoiding resource tree call", "app", appDetail.AppName, "env", appDetail.EnvironmentName)
		} else {
			appDetail.ResourceTree = map[string]interface{}{}
		}
	} else {
		appDetail.ResourceTree = map[string]interface{}{}
		handler.logger.Warnw("appName and envName not found - avoiding resource tree call", "app", appDetail.AppName, "env", appDetail.EnvironmentName)
	}
	return appDetail
}

func (handler AppListingRestHandlerImpl) ManualSyncAcdPipelineDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, ManualSyncAcdPipelineDeploymentStatus", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.logger.Errorw("request err, ManualSyncAcdPipelineDeploymentStatus", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipeline.GetApp(appId)
	if err != nil {
		handler.logger.Errorw("bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	object := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "unauthorized user", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	err = handler.cdApplicationStatusUpdateHandler.ManualSyncPipelineStatus(appId, envId, userId)
	if err != nil {
		handler.logger.Errorw("service err, ManualSyncAcdPipelineDeploymentStatus", "err", err, "appId", appId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "App synced successfully.", http.StatusOK)
}
