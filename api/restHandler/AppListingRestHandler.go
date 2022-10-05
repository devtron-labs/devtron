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
	"github.com/devtron-labs/devtron/api/appStore"
	"github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	application1 "github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
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
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AppListingRestHandler interface {
	FetchAppsByEnvironment(w http.ResponseWriter, r *http.Request)
	FetchAppDetails(w http.ResponseWriter, r *http.Request)

	FetchAppTriggerView(w http.ResponseWriter, r *http.Request)
	FetchAppStageStatus(w http.ResponseWriter, r *http.Request)

	FetchOtherEnvironment(w http.ResponseWriter, r *http.Request)
	RedirectToLinkouts(w http.ResponseWriter, r *http.Request)
	GetManifestsByBatch(w http.ResponseWriter, r *http.Request)
}

type AppListingRestHandlerImpl struct {
	application             application.ServiceClient
	appListingService       app.AppListingService
	teamService             team.TeamService
	enforcer                casbin.Enforcer
	pipeline                pipeline.PipelineBuilder
	logger                  *zap.SugaredLogger
	enforcerUtil            rbac.EnforcerUtil
	deploymentGroupService  deploymentGroup.DeploymentGroupService
	userService             user.UserService
	helmAppClient           client.HelmAppClient
	clusterService          cluster.ClusterService
	helmAppService          client.HelmAppService
	argoUserService         argo.ArgoUserService
	k8sApplicationService   k8s.K8sApplicationService
	installedAppService     service1.InstalledAppService
	installedAppRestHandler appStore.InstalledAppRestHandler
}

type AppStatus struct {
	name       string
	status     string
	message    string
	err        error
	conditions []v1alpha1.ApplicationCondition
}

type BatchRequest struct {
	AppId          string `json:"appId,omitempty"`
	EnvId          string `json:"envId,omitempty"`
	InstalledAppId string `json:"installedAppId"`
	//ResourceRequests []k8s.ResourceRequestBean `json:"resourceRequests"`
}

func NewAppListingRestHandlerImpl(application application.ServiceClient,
	appListingService app.AppListingService,
	teamService team.TeamService,
	enforcer casbin.Enforcer,
	pipeline pipeline.PipelineBuilder,
	logger *zap.SugaredLogger, enforcerUtil rbac.EnforcerUtil,
	deploymentGroupService deploymentGroup.DeploymentGroupService, userService user.UserService,
	helmAppClient client.HelmAppClient, clusterService cluster.ClusterService, helmAppService client.HelmAppService,
	argoUserService argo.ArgoUserService, k8sApplicationService k8s.K8sApplicationService, installedAppService service1.InstalledAppService, installedAppRestHandler appStore.InstalledAppRestHandler) *AppListingRestHandlerImpl {
	appListingHandler := &AppListingRestHandlerImpl{
		application:             application,
		appListingService:       appListingService,
		logger:                  logger,
		teamService:             teamService,
		pipeline:                pipeline,
		enforcer:                enforcer,
		enforcerUtil:            enforcerUtil,
		deploymentGroupService:  deploymentGroupService,
		userService:             userService,
		helmAppClient:           helmAppClient,
		clusterService:          clusterService,
		helmAppService:          helmAppService,
		argoUserService:         argoUserService,
		k8sApplicationService:   k8sApplicationService,
		installedAppService:     installedAppService,
		installedAppRestHandler: installedAppRestHandler,
	}
	return appListingHandler
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	(*w).Header().Set("Content-Type", "text/html; charset=utf-8")
}

func (handler AppListingRestHandlerImpl) FetchAppsByEnvironment(w http.ResponseWriter, r *http.Request) {
	//Allow CORS here By * or specific origin
	setupResponse(&w, r)
	token := r.Header.Get("token")
	t0 := time.Now()
	t1 := time.Now()
	handler.logger.Infow("api response time testing", "time", time.Now().String(), "stage", "1")
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	user, err := handler.userService.GetById(userId)
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
	_, _, err = fetchAppListingRequest.GetNamespaceClusterMapping()
	if err != nil {
		handler.logger.Errorw("request err, GetNamespaceClusterMapping", "err", err, "payload", fetchAppListingRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	var dg *deploymentGroup.DeploymentGroupDTO
	if fetchAppListingRequest.DeploymentGroupId > 0 {
		dg, err = handler.deploymentGroupService.FindById(fetchAppListingRequest.DeploymentGroupId)
		if err != nil {
			handler.logger.Errorw("service err, FetchAppsByEnvironment", "err", err, "payload", fetchAppListingRequest)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		}
	}

	envContainers, err := handler.appListingService.FetchAppsByEnvironment(fetchAppListingRequest, w, r, token)
	if err != nil {
		handler.logger.Errorw("service err, FetchAppsByEnvironment", "err", err, "payload", fetchAppListingRequest)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
	}
	t2 := time.Now()
	handler.logger.Infow("api response time testing", "time", time.Now().String(), "time diff", t2.Unix()-t1.Unix(), "stage", "2")
	t1 = t2

	isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
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

		resultMap := handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceTeam, casbin.ActionGet, objectArray)
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

		resultMap = handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceApplications, casbin.ActionGet, objectArray)
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
	apps, err := handler.appListingService.BuildAppListingResponse(fetchAppListingRequest, appEnvContainers)
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
	appDetail, err := handler.appListingService.FetchAppDetails(appId, envId)
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
	appDetail = handler.fetchResourceTree(w, r, token, appId, envId, appDetail)
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
	app, err := handler.pipeline.GetApp(appId)
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

	otherEnvironment, err := handler.appListingService.FetchOtherEnvironment(appId)
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

func (handler AppListingRestHandlerImpl) GetManifestsByBatch(w http.ResponseWriter, r *http.Request) {
	var batchRequest BatchRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&batchRequest)
	if err != nil {
		handler.logger.Errorw("error in decoding batch request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if (batchRequest.AppId == "" && batchRequest.InstalledAppId == "") || (batchRequest.AppId != "" && batchRequest.InstalledAppId != "") {
		handler.logger.Error("error in decoding batch request body")
		common.WriteJsonResp(w, fmt.Errorf("only one of the appId or envId should be valid"), nil, http.StatusBadRequest)
		return
	}
	var appDetail bean.AppDetailContainer
	var appId, envId int
	envId, err = strconv.Atoi(batchRequest.EnvId)
	if err != nil {
		handler.logger.Errorw("error in env-id from request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	if batchRequest.AppId != "" {
		appId, err = strconv.Atoi(batchRequest.AppId)
		if err != nil {
			handler.logger.Errorw("error in app-id from request body", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		appDetail, err = handler.appListingService.FetchAppDetails(appId, envId)
	}

	if batchRequest.InstalledAppId != "" {
		appId, err = strconv.Atoi(batchRequest.InstalledAppId)
		if err != nil {
			handler.logger.Errorw("error in app-id from request body", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		appDetail, err = handler.installedAppService.FindAppDetailsForAppstoreApplication(appId, envId)
	}

	if err != nil {
		handler.logger.Errorw("error occurred while getting app details", "appId", batchRequest.AppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	isAcdApp := len(appDetail.AppName) > 0 && len(appDetail.EnvironmentName) > 0 && util.IsAcdApp(appDetail.DeploymentAppType)
	isHelmApp := len(appDetail.AppName) > 0 && len(appDetail.EnvironmentName) > 0 && util.IsHelmApp(appDetail.DeploymentAppType)
	if !isAcdApp && !isHelmApp {
		handler.logger.Errorw("Invalid app type", "appId", batchRequest.AppId)
		common.WriteJsonResp(w, fmt.Errorf("app is neither helm app or devtron app"), nil, http.StatusBadRequest)
		return
	}
	if isHelmApp {
		handler.installedAppRestHandler.FetchResourceTreeHelper(w, r, "", &appDetail)
	} else {
		appDetail = handler.fetchResourceTree(w, r, "", appId, envId, appDetail)
	}

	resourceTree := appDetail.ResourceTree
	_, ok := resourceTree["nodes"]
	if !ok {
		handler.logger.Errorw("no nodes found for this resource tree", "appName", appDetail.AppName, "envName", appDetail.EnvironmentName)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//valid batch requests, only valid requests will be sent for batch processing
	validRequests := make([]k8s.ResourceRequestAndGroupVersionKind, 0)
	version := ""
	if len(appDetail.K8sVersion) != 0 {
		version = strings.Split(appDetail.K8sVersion, ".")[0]
	}

	noOfNodes := len(resourceTree["nodes"].([]interface{}))
	for i := 0; i < noOfNodes; i++ {
		resourceI := resourceTree["nodes"].([]interface{})[i].(map[string]interface{})
		kind, name, namespace := resourceI["kind"].(string), resourceI["name"].(string), resourceI["namespace"].(string)
		if strings.Compare(kind, "Service") == 0 || strings.Compare(kind, "Ingress") == 0 {
			req := k8s.ResourceRequestAndGroupVersionKind{
				ResourceRequestBean: k8s.ResourceRequestBean{
					AppId: strconv.Itoa(appDetail.ClusterId) + "|" + namespace + "|" + (appDetail.AppName + "-" + appDetail.EnvironmentName),
					AppIdentifier: &client.AppIdentifier{
						ClusterId: appDetail.ClusterId,
					},
					K8sRequest: &application1.K8sRequestBean{
						ResourceIdentifier: application1.ResourceIdentifier{
							Name:      name,
							Namespace: namespace,
						},
					},
				},
				Version: version,
				Kind:    kind,
			}
			validRequests = append(validRequests, req)
		}
	}

	if len(validRequests) == 0 {
		handler.logger.Error("Invalid requests in whole batch")
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resp := handler.k8sApplicationService.GetManifestsInBatch(validRequests, 5)
	result := make([]interface{}, 0)
	for _, res := range resp {
		err = res.Err
		if err != nil {
			continue
		}
		urlRes := handler.getUrls(res.ManifestResponse)
		result = append(result, urlRes)
	}
	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

type Response struct {
	Kind     string   `json:"kind"`
	Name     string   `json:"name"`
	PointsTo string   `json:"pointsTo"`
	Urls     []string `json:"urls"`
}

func (handler AppListingRestHandlerImpl) getUrls(manifest *application1.ManifestResponse) Response {
	var res Response
	kind := manifest.Manifest.Object["kind"]
	if _, ok := manifest.Manifest.Object["metadata"]; ok {
		metadata := manifest.Manifest.Object["metadata"].(map[string]interface{})
		if metadata != nil {
			name := metadata["name"]
			if name != nil {
				res.Name = name.(string)
			}
		}
	}

	if kind != nil {
		res.Kind = kind.(string)
	}
	res.PointsTo = ""
	urls := make([]string, 0)
	if res.Kind == "Ingress" {
		if manifest.Manifest.Object["spec"] != nil {
			spec := manifest.Manifest.Object["spec"].(map[string]interface{})
			if spec["rules"] != nil {
				rules := spec["rules"].([]interface{})
				for _, rule := range rules {
					ruleMap := rule.(map[string]interface{})
					url := ""
					if ruleMap["host"] != nil {
						url = ruleMap["host"].(string)
					}
					var httpPaths []interface{}
					if ruleMap["http"] != nil && ruleMap["http"].(map[string]interface{})["paths"] != nil {
						httpPaths = ruleMap["http"].(map[string]interface{})["paths"].([]interface{})
					} else {
						continue
					}
					for _, httpPath := range httpPaths {
						path := httpPath.(map[string]interface{})["path"]
						if path != nil {
							url = url + path.(string)
						}
						urls = append(urls, url)
					}
				}
			}
		}
	}

	if manifest.Manifest.Object["status"] != nil {
		status := manifest.Manifest.Object["status"].(map[string]interface{})
		if status["loadBalancer"] != nil {
			loadBalancer := status["loadBalancer"].(map[string]interface{})
			if loadBalancer["ingress"] != nil {
				ingressArray := loadBalancer["ingress"].([]interface{})
				if len(ingressArray) > 0 {
					if hostname, ok := ingressArray[0].(map[string]interface{})["hostname"]; ok {
						res.PointsTo = hostname.(string)
					} else if ip, ok := ingressArray[0].(map[string]interface{})["ip"]; ok {
						res.PointsTo = ip.(string)
					}
				}
			}
		}
	}
	res.Urls = urls
	return res
}

func (handler AppListingRestHandlerImpl) fetchResourceTree(w http.ResponseWriter, r *http.Request, token string, appId int, envId int, appDetail bean.AppDetailContainer) bean.AppDetailContainer {
	if len(appDetail.AppName) > 0 && len(appDetail.EnvironmentName) > 0 && util.IsAcdApp(appDetail.DeploymentAppType) {
		//RBAC enforcer Ends
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
		handler.logger.Debugw("application environment status", "appId", appId, "envId", envId, "resp", resp)
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
			resourceTree["status"] = detail.ReleaseStatus.Status
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
