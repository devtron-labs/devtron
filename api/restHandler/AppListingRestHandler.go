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
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	bean2 "github.com/devtron-labs/devtron/api/restHandler/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/cron"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	service1 "github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/deploymentGroup"
	"github.com/devtron-labs/devtron/pkg/genericNotes"
	"github.com/devtron-labs/devtron/pkg/k8s"
	application3 "github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
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
	FetchJobs(w http.ResponseWriter, r *http.Request)
	FetchJobOverviewCiPipelines(w http.ResponseWriter, r *http.Request)
	FetchAppDetailsV2(w http.ResponseWriter, r *http.Request)
	FetchResourceTree(w http.ResponseWriter, r *http.Request)
	FetchAllDevtronManagedApps(w http.ResponseWriter, r *http.Request)
	FetchAppTriggerView(w http.ResponseWriter, r *http.Request)
	FetchAppStageStatus(w http.ResponseWriter, r *http.Request)

	FetchOtherEnvironment(w http.ResponseWriter, r *http.Request)
	FetchMinDetailOtherEnvironment(w http.ResponseWriter, r *http.Request)
	RedirectToLinkouts(w http.ResponseWriter, r *http.Request)
	GetHostUrlsByBatch(w http.ResponseWriter, r *http.Request)

	ManualSyncAcdPipelineDeploymentStatus(w http.ResponseWriter, r *http.Request)
	GetClusterTeamAndEnvListForAutocomplete(w http.ResponseWriter, r *http.Request)
	FetchAppsByEnvironmentV2(w http.ResponseWriter, r *http.Request)
	FetchAppsByEnvironmentV1(w http.ResponseWriter, r *http.Request)
	FetchAppsByEnvironmentVersioned(w http.ResponseWriter, r *http.Request)
	FetchOverviewAppsByEnvironment(w http.ResponseWriter, r *http.Request)
}

type AppListingRestHandlerImpl struct {
	application                       application.ServiceClient
	appListingService                 app.AppListingService
	teamService                       team.TeamService
	enforcer                          casbin.Enforcer
	pipeline                          pipeline.PipelineBuilder
	logger                            *zap.SugaredLogger
	enforcerUtil                      rbac.EnforcerUtil
	deploymentGroupService            deploymentGroup.DeploymentGroupService
	userService                       user.UserService
	helmAppClient                     client.HelmAppClient
	clusterService                    cluster.ClusterService
	helmAppService                    client.HelmAppService
	argoUserService                   argo.ArgoUserService
	k8sCommonService                  k8s.K8sCommonService
	installedAppService               service1.InstalledAppService
	cdApplicationStatusUpdateHandler  cron.CdApplicationStatusUpdateHandler
	pipelineRepository                pipelineConfig.PipelineRepository
	appStatusService                  appStatus.AppStatusService
	installedAppRepository            repository.InstalledAppRepository
	environmentClusterMappingsService cluster.EnvironmentService
	genericNoteService                genericNotes.GenericNoteService
	cfg                               *bean.Config
	k8sApplicationService             application3.K8sApplicationService
}

type AppStatus struct {
	name       string
	status     string
	message    string
	err        error
	conditions []v1alpha1.ApplicationCondition
}

type AppAutocomplete struct {
	Teams        []team.TeamRequest
	Environments []cluster.EnvironmentBean
	Clusters     []cluster.ClusterBean
}

func NewAppListingRestHandlerImpl(application application.ServiceClient,
	appListingService app.AppListingService,
	teamService team.TeamService,
	enforcer casbin.Enforcer,
	pipeline pipeline.PipelineBuilder,
	logger *zap.SugaredLogger, enforcerUtil rbac.EnforcerUtil,
	deploymentGroupService deploymentGroup.DeploymentGroupService, userService user.UserService,
	helmAppClient client.HelmAppClient, clusterService cluster.ClusterService, helmAppService client.HelmAppService,
	argoUserService argo.ArgoUserService, k8sCommonService k8s.K8sCommonService, installedAppService service1.InstalledAppService,
	cdApplicationStatusUpdateHandler cron.CdApplicationStatusUpdateHandler,
	pipelineRepository pipelineConfig.PipelineRepository,
	appStatusService appStatus.AppStatusService, installedAppRepository repository.InstalledAppRepository,
	environmentClusterMappingsService cluster.EnvironmentService,
	genericNoteService genericNotes.GenericNoteService,
	k8sApplicationService application3.K8sApplicationService,
) *AppListingRestHandlerImpl {
	cfg := &bean.Config{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Errorw("error occurred while parsing config ", "err", err)
		cfg.IgnoreAuthCheck = false
	}
	logger.Infow("app listing rest handler initialized", "ignoreAuthCheckValue", cfg.IgnoreAuthCheck)
	appListingHandler := &AppListingRestHandlerImpl{
		application:                       application,
		appListingService:                 appListingService,
		logger:                            logger,
		teamService:                       teamService,
		pipeline:                          pipeline,
		enforcer:                          enforcer,
		enforcerUtil:                      enforcerUtil,
		deploymentGroupService:            deploymentGroupService,
		userService:                       userService,
		helmAppClient:                     helmAppClient,
		clusterService:                    clusterService,
		helmAppService:                    helmAppService,
		argoUserService:                   argoUserService,
		k8sCommonService:                  k8sCommonService,
		installedAppService:               installedAppService,
		cdApplicationStatusUpdateHandler:  cdApplicationStatusUpdateHandler,
		pipelineRepository:                pipelineRepository,
		appStatusService:                  appStatusService,
		installedAppRepository:            installedAppRepository,
		environmentClusterMappingsService: environmentClusterMappingsService,
		genericNoteService:                genericNoteService,
		cfg:                               cfg,
		k8sApplicationService:             k8sApplicationService,
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
func (handler AppListingRestHandlerImpl) FetchJobs(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		handler.logger.Errorw("request err, userId", "err", err, "payload", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if !isSuperAdmin || err != nil {
		if err != nil {
			handler.logger.Errorw("request err, CheckSuperAdmin", "err", isSuperAdmin, "isSuperAdmin", isSuperAdmin)
		}
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	var fetchJobListingRequest app.FetchAppListingRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&fetchJobListingRequest)
	if err != nil {
		handler.logger.Errorw("request err, FetchAppsByEnvironment", "err", err, "payload", fetchJobListingRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	jobs, err := handler.appListingService.FetchJobs(fetchJobListingRequest)
	if err != nil {
		handler.logger.Errorw("service err, FetchJobs", "err", err, "payload", fetchJobListingRequest)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}
	jobsCount := len(jobs)
	offset := fetchJobListingRequest.Offset
	limit := fetchJobListingRequest.Size

	if limit > 0 {
		if offset+limit <= len(jobs) {
			jobs = jobs[offset : offset+limit]
		} else {
			jobs = jobs[offset:]
		}
	}
	jobContainerResponse := bean.JobContainerResponse{
		JobContainers: jobs,
		JobCount:      jobsCount,
	}

	common.WriteJsonResp(w, err, jobContainerResponse, http.StatusOK)
}
func (handler AppListingRestHandlerImpl) FetchJobOverviewCiPipelines(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		handler.logger.Errorw("request err, userId", "err", err, "payload", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if !isSuperAdmin || err != nil {
		if err != nil {
			handler.logger.Errorw("request err, CheckSuperAdmin", "err", isSuperAdmin, "isSuperAdmin", isSuperAdmin)
		}
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	jobId, err := strconv.Atoi(vars["jobId"])
	if err != nil {
		handler.logger.Errorw("request err, GetAppMetaInfo", "err", err, "jobId", jobId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	job, err := handler.pipeline.GetApp(jobId)
	if err != nil || job == nil || job.AppType != helper.Job {
		handler.logger.Errorw("Job with the given Id does not exist", "err", err, "jobId", jobId)
		common.WriteJsonResp(w, err, "Job with the given Id does not exist", http.StatusBadRequest)
		return
	}

	jobCi, err := handler.appListingService.FetchOverviewCiPipelines(jobId)
	if err != nil {
		handler.logger.Errorw("request err, GetJobCi", "err", jobCi, "jobCi", jobCi)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	common.WriteJsonResp(w, err, jobCi, http.StatusOK)
}
func (handler AppListingRestHandlerImpl) FetchAppsByEnvironmentVersioned(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version := vars["version"]
	if version == app.APIVersionV1 {
		handler.FetchAppsByEnvironmentV1(w, r)
		return
	}
	if version == app.APIVersionV2 {
		handler.FetchAppsByEnvironmentV2(w, r)
		return
	}
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
	start := time.Now()
	envContainers, err := handler.appListingService.FetchAppsByEnvironment(fetchAppListingRequest, w, r, token, "")
	middleware.AppListingDuration.WithLabelValues("fetchAppsByEnvironment", "devtron").Observe(time.Since(start).Seconds())
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
		start = time.Now()
		resultMap := handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceTeam, casbin.ActionGet, objectArray)
		middleware.AppListingDuration.WithLabelValues("enforceByEmailInBatchResourceTeam", "devtron").Observe(time.Since(start).Seconds())
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
		start = time.Now()
		resultMap = handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceApplications, casbin.ActionGet, objectArray)
		middleware.AppListingDuration.WithLabelValues("enforceByEmailInBatchResourceApplication", "devtron").Observe(time.Since(start).Seconds())
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

	if limit > 0 {
		if offset+limit <= len(apps) {
			apps = apps[offset : offset+limit]
		} else {
			apps = apps[offset:]
		}
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
func (handler AppListingRestHandlerImpl) FetchAppsByEnvironmentV1(w http.ResponseWriter, r *http.Request) {
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
	start := time.Now()
	envContainers, err := handler.appListingService.FetchAppsByEnvironment(fetchAppListingRequest, w, r, token, app.APIVersionV1)
	middleware.AppListingDuration.WithLabelValues("fetchAppsByEnvironment", "devtron").Observe(time.Since(start).Seconds())
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
		start = time.Now()
		resultMap := handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceTeam, casbin.ActionGet, objectArray)
		middleware.AppListingDuration.WithLabelValues("enforceByEmailInBatchResourceTeam", "devtron").Observe(time.Since(start).Seconds())
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
		start = time.Now()
		resultMap = handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceApplications, casbin.ActionGet, objectArray)
		middleware.AppListingDuration.WithLabelValues("enforceByEmailInBatchResourceApplication", "devtron").Observe(time.Since(start).Seconds())
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

	if limit > 0 {
		if offset+limit <= len(apps) {
			apps = apps[offset : offset+limit]
		} else {
			apps = apps[offset:]
		}
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
func (handler AppListingRestHandlerImpl) FetchAppsByEnvironmentV2(w http.ResponseWriter, r *http.Request) {
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
	newCtx, span = otel.Tracer("userService").Start(newCtx, "IsSuperAdmin")
	isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	span.End()
	if err != nil {
		handler.logger.Errorw("request err, FetchAppsByEnvironment", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}

	validAppIds := make([]int, 0)
	//for non super admin users
	if !isActionUserSuperAdmin {
		userEmailId := strings.ToLower(user.EmailId)
		rbacObjectsForAllAppsMap := handler.enforcerUtil.GetRbacObjectsForAllApps()
		rbacObjectToAppIdMap := make(map[string]int)
		rbacObjects := make([]string, len(rbacObjectsForAllAppsMap))
		itr := 0
		for appId, object := range rbacObjectsForAllAppsMap {
			rbacObjects[itr] = object
			rbacObjectToAppIdMap[object] = appId
			itr++
		}

		result := handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceApplications, casbin.ActionGet, rbacObjects)
		//O(n) loop, n = len(rbacObjectsForAllAppsMap)
		for object, ok := range result {
			if ok {
				validAppIds = append(validAppIds, rbacObjectToAppIdMap[object])
			}
		}

		if len(validAppIds) == 0 {
			handler.logger.Infow("user doesn't have access to any app", "userId", userId)
			common.WriteJsonResp(w, err, bean.AppContainerResponse{}, http.StatusOK)
			return
		}
	}

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
	start := time.Now()
	fetchAppListingRequest.AppIds = validAppIds
	envContainers, appsCount, err := handler.appListingService.FetchAppsByEnvironmentV2(fetchAppListingRequest, w, r, token)
	middleware.AppListingDuration.WithLabelValues("fetchAppsByEnvironment", "devtron").Observe(time.Since(start).Seconds())
	span.End()
	if err != nil {
		handler.logger.Errorw("service err, FetchAppsByEnvironment", "err", err, "payload", fetchAppListingRequest)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
	}

	t2 := time.Now()
	handler.logger.Infow("api response time testing", "time", time.Now().String(), "time diff", t2.Unix()-t1.Unix(), "stage", "3")
	t1 = t2
	newCtx, span = otel.Tracer("appListingService").Start(newCtx, "BuildAppListingResponse")
	apps, err := handler.appListingService.BuildAppListingResponseV2(fetchAppListingRequest, envContainers)
	span.End()
	if err != nil {
		handler.logger.Errorw("service err, FetchAppsByEnvironment", "err", err, "payload", fetchAppListingRequest)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
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

func (handler AppListingRestHandlerImpl) FetchOverviewAppsByEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	user, err := handler.userService.GetById(userId)
	if user == nil || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	envId, err := strconv.Atoi(vars["env-id"])
	if err != nil || envId == 0 {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	limit, err := strconv.Atoi(vars["size"])
	if _, ok := vars["size"]; ok && err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	offset, err := strconv.Atoi(vars["offset"])
	if _, ok := vars["offset"]; ok && err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resp, err := handler.appListingService.FetchOverviewAppsByEnvironment(envId, limit, offset)
	if err != nil {
		handler.logger.Errorw("error in getting apps for app-group overview", "envid", envId, "limit", limit, "offset", offset)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resp.AppCount = len(resp.Apps)
	isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if err != nil {
		handler.logger.Errorw("request err, FetchOverviewAppsByEnvironment", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}

	//return if user is super admin
	if isActionUserSuperAdmin {
		common.WriteJsonResp(w, err, resp, http.StatusOK)
		return
	}

	//apply rbac
	userEmailId := strings.ToLower(user.EmailId)
	//get all the appIds
	appIds := make([]int, 0)
	appContainers := resp.Apps
	for _, appBean := range resp.Apps {
		appIds = append(appIds, appBean.AppId)
	}

	//get rbac objects for the appids
	rbacObjectsWithAppId := handler.enforcerUtil.GetRbacObjectsByAppIds(appIds)
	rbacObjects := make([]string, len(rbacObjectsWithAppId))
	itr := 0
	for _, object := range rbacObjectsWithAppId {
		rbacObjects[itr] = object
		itr++
	}
	//enforce rbac in batch
	rbacResult := handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceApplications, casbin.ActionGet, rbacObjects)
	//filter out rbac passed apps
	resp.Apps = make([]*bean.AppEnvironmentContainer, 0)
	for _, appBean := range appContainers {
		rbacObject := rbacObjectsWithAppId[appBean.AppId]
		if rbacResult[rbacObject] {
			resp.Apps = append(resp.Apps, appBean)
		}
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)

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
	pipelines, err := handler.pipelineRepository.FindActiveByAppIdAndEnvironmentId(appId, envId)
	if err == pg.ErrNoRows {
		common.WriteJsonResp(w, err, "pipeline Not found in database", http.StatusNotFound)
		return
	}
	if err != nil {
		handler.logger.Errorw("error in fetching pipelines from db", "appId", appId, "envId", envId)
		common.WriteJsonResp(w, err, "error in fetching pipeline from database", http.StatusInternalServerError)
		return
	}
	if len(pipelines) == 0 {
		common.WriteJsonResp(w, fmt.Errorf("app deleted"), nil, http.StatusNotFound)
		return
	}
	if len(pipelines) != 1 {
		common.WriteJsonResp(w, err, "multiple pipelines found for an envId", http.StatusBadRequest)
		return
	}
	cdPipeline := pipelines[0]
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
	acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
	//acdToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhcmdvY2QiLCJzdWIiOiJkZXZ0cm9uOmFwaUtleSIsIm5iZiI6MTY5MDE4Mjg0MSwiaWF0IjoxNjkwMTgyODQxLCJqdGkiOiJhMDA4YWIzZC05ODdiLTQ4ZWEtYjVhNS1kNTQwOWM3YjQzZmUifQ.ChpTDgoNERC7BBeB2E1HTI1UnYpyWBrn1pjQmWYro-8"
	if err != nil {
		common.WriteJsonResp(w, fmt.Errorf("error in getting acd token"), nil, http.StatusInternalServerError)
		return
	}
	resourceTree, err := handler.fetchResourceTree(w, r, appId, envId, acdToken, cdPipeline)
	if appDetail.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD {
		apiError, ok := err.(*util.ApiError)
		if ok && apiError != nil {
			if apiError.Code == constants.AppDetailResourceTreeNotFound && appDetail.DeploymentAppDeleteRequest == true {
				acdAppFound, _ := handler.pipeline.MarkGitOpsDevtronAppsDeletedWhereArgoAppIsDeleted(appId, envId, acdToken, cdPipeline)
				if acdAppFound {
					common.WriteJsonResp(w, fmt.Errorf("unable to fetch resource tree"), nil, http.StatusInternalServerError)
					return
				} else {
					common.WriteJsonResp(w, fmt.Errorf("app deleted"), nil, http.StatusNotFound)
					return
				}
			}
		}
	}
	if err != nil {
		common.WriteJsonResp(w, fmt.Errorf("unable to fetch resource tree"), nil, http.StatusInternalServerError)
		return
	}
	appDetail.ResourceTree = resourceTree
	common.WriteJsonResp(w, err, appDetail, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) FetchAppDetailsV2(w http.ResponseWriter, r *http.Request) {
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
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	appDetail, err := handler.appListingService.FetchAppDetails(r.Context(), appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, FetchAppDetailsV2", "err", err, "appId", appId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, appDetail, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) FetchResourceTree(w http.ResponseWriter, r *http.Request) {
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
	pipelines, err := handler.pipelineRepository.FindActiveByAppIdAndEnvironmentId(appId, envId)
	if err == pg.ErrNoRows {
		common.WriteJsonResp(w, err, "pipeline Not found in database", http.StatusNotFound)
		return
	}
	if err != nil {
		handler.logger.Errorw("error in fetching pipelines from db", "appId", appId, "envId", envId)
		common.WriteJsonResp(w, err, "error in fetching pipeline from database", http.StatusInternalServerError)
		return
	}
	if len(pipelines) == 0 {
		common.WriteJsonResp(w, fmt.Errorf("app deleted"), nil, http.StatusNotFound)
		return
	}
	if len(pipelines) != 1 {
		common.WriteJsonResp(w, err, "multiple pipelines found for an envId", http.StatusBadRequest)
		return
	}
	cdPipeline := pipelines[0]
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		common.WriteJsonResp(w, fmt.Errorf("error in getting acd token"), nil, http.StatusInternalServerError)
		return
	}
	resourceTree, err := handler.fetchResourceTree(w, r, appId, envId, acdToken, cdPipeline)
	if cdPipeline.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD {
		apiError, ok := err.(*util.ApiError)
		if ok && apiError != nil {
			if apiError.Code == constants.AppDetailResourceTreeNotFound && cdPipeline.DeploymentAppDeleteRequest == true && cdPipeline.DeploymentAppCreated == true {
				acdAppFound, _ := handler.pipeline.MarkGitOpsDevtronAppsDeletedWhereArgoAppIsDeleted(appId, envId, acdToken, cdPipeline)
				if acdAppFound {
					common.WriteJsonResp(w, fmt.Errorf("unable to fetch resource tree"), nil, http.StatusInternalServerError)
					return
				} else {
					common.WriteJsonResp(w, fmt.Errorf("app deleted"), nil, http.StatusNotFound)
					return
				}
			}
		}
	}
	common.WriteJsonResp(w, err, resourceTree, http.StatusOK)
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

	triggerView, err := handler.appListingService.FetchAppStageStatus(appId, int(app.AppType))
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

func (handler AppListingRestHandlerImpl) FetchMinDetailOtherEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["app-id"])
	if err != nil {
		handler.logger.Errorw("request err, FetchMinDetailOtherEnvironment", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipeline.GetApp(appId)
	if err != nil {
		handler.logger.Errorw("service err, FetchMinDetailOtherEnvironment", "err", err, "appId", appId)
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

	otherEnvironment, err := handler.appListingService.FetchMinDetailOtherEnvironment(appId)
	if err != nil {
		handler.logger.Errorw("service err, FetchMinDetailOtherEnvironment", "err", err, "appId", appId)
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
func (handler AppListingRestHandlerImpl) fetchResourceTreeFromInstallAppService(w http.ResponseWriter, r *http.Request, resourceTreeAndNotesContainer bean.ResourceTreeAndNotesContainer, installedApps repository.InstalledApps) (bean.ResourceTreeAndNotesContainer, error) {
	rctx := r.Context()
	cn, _ := w.(http.CloseNotifier)
	err := handler.installedAppService.FetchResourceTree(rctx, cn, &resourceTreeAndNotesContainer, installedApps)
	return resourceTreeAndNotesContainer, err
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
	var resourceTree map[string]interface{}
	if installedAppIdParam != "" {
		installedAppId, err := strconv.Atoi(installedAppIdParam)
		if err != nil {
			handler.logger.Errorw("request err, FetchAppDetailsForInstalledAppV2", "err", err, "installedAppId", installedAppId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		installedApp, err := handler.installedAppService.CheckAppExistsByInstalledAppId(installedAppId)
		if err == pg.ErrNoRows {
			common.WriteJsonResp(w, err, "App not found in database", http.StatusBadRequest)
			return
		}
		resourceTreeAndNotesContainer := bean.ResourceTreeAndNotesContainer{}
		resourceTreeAndNotesContainer, err = handler.fetchResourceTreeFromInstallAppService(w, r, resourceTreeAndNotesContainer, *installedApp)
		if err != nil {
			common.WriteJsonResp(w, fmt.Errorf("error in fetching resource tree"), nil, http.StatusInternalServerError)
			return
		}
		resourceTree = resourceTreeAndNotesContainer.ResourceTree

	} else {
		acdToken, err := handler.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			common.WriteJsonResp(w, fmt.Errorf("error in getting acd token"), nil, http.StatusInternalServerError)
			return
		}
		pipelines, err := handler.pipelineRepository.FindActiveByAppIdAndEnvironmentId(appId, envId)
		if err != nil && err != pg.ErrNoRows {
			handler.logger.Errorw("error in fetching pipelines from db", "appId", appId, "envId", envId)
			common.WriteJsonResp(w, err, "error in fetching pipelines from db", http.StatusInternalServerError)
			return
		}
		if len(pipelines) == 0 {
			common.WriteJsonResp(w, err, "deployment not found, unable to fetch resource tree", http.StatusNotFound)
			return
		}
		if len(pipelines) > 1 {
			common.WriteJsonResp(w, err, "multiple pipelines found for an envId", http.StatusBadRequest)
			return
		}

		cdPipeline := pipelines[0]
		resourceTree, err = handler.fetchResourceTree(w, r, appId, envId, acdToken, cdPipeline)
	}
	_, ok := resourceTree["nodes"]
	if !ok {
		err = fmt.Errorf("no nodes found for this resource tree appName:%s , envName:%s", appDetail.AppName, appDetail.EnvironmentName)
		handler.logger.Errorw("no nodes found for this resource tree", "appName", appDetail.AppName, "envName", appDetail.EnvironmentName)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//valid batch requests, only valid requests will be sent for batch processing
	validRequests := make([]k8s.ResourceRequestBean, 0)
	validRequests = handler.k8sCommonService.FilterK8sResources(r.Context(), resourceTree, validRequests, appDetail, "", []string{k8s.ServiceKind, k8s.IngressKind})
	if len(validRequests) == 0 {
		handler.logger.Error("neither service nor ingress found for", "appId", appIdParam, "envId", envIdParam, "installedAppId", installedAppIdParam)
		common.WriteJsonResp(w, err, nil, http.StatusNoContent)
		return
	}
	resp, err := handler.k8sCommonService.GetManifestsByBatch(r.Context(), validRequests)
	if err != nil {
		handler.logger.Errorw("error in getting manifests in batch", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	result := handler.k8sApplicationService.GetUrlsByBatchForIngress(r.Context(), resp)
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
func (handler AppListingRestHandlerImpl) fetchResourceTree(w http.ResponseWriter, r *http.Request, appId int, envId int, acdToken string, cdPipeline *pipelineConfig.Pipeline) (map[string]interface{}, error) {
	var resourceTree map[string]interface{}
	if len(cdPipeline.DeploymentAppName) > 0 && cdPipeline.EnvironmentId > 0 && util.IsAcdApp(cdPipeline.DeploymentAppType) {
		//RBAC enforcer Ends
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
		ctx = context.WithValue(ctx, "token", acdToken)
		start := time.Now()
		resp, err := handler.application.ResourceTree(ctx, query)
		elapsed := time.Since(start)
		handler.logger.Debugw("FetchAppDetailsV2, time elapsed in fetching application for environment ", "elapsed", elapsed, "appId", appId, "envId", envId)
		if err != nil {
			handler.logger.Errorw("service err, FetchAppDetailsV2, resource tree", "err", err, "app", appId, "env", envId)
			err = &util.ApiError{
				Code:            constants.AppDetailResourceTreeNotFound,
				InternalMessage: "app detail fetched, failed to get resource tree from acd",
				UserMessage:     "Error fetching detail, if you have recently created this deployment pipeline please try after sometime.",
			}
			return resourceTree, err
		}

		//we currently add appId and envId as labels for devtron apps deployed via acd
		label := fmt.Sprintf("appId=%v,envId=%v", cdPipeline.AppId, cdPipeline.EnvironmentId)
		pods, err := handler.k8sApplicationService.GetPodListByLabel(cdPipeline.Environment.ClusterId, cdPipeline.Environment.Namespace, label)
		if err != nil {
			handler.logger.Errorw("error in getting pods by label", "err", err, "clusterId", cdPipeline.Environment.ClusterId, "namespace", cdPipeline.Environment.Namespace, "label", label)
			return resourceTree, err
		}
		ephemeralContainersMap := bean2.ExtractEphemeralContainers(pods)
		for _, metaData := range resp.PodMetadata {
			metaData.EphemeralContainers = ephemeralContainersMap[metaData.Name]
		}

		if resp.Status == string(health.HealthStatusHealthy) {
			status, err := handler.appListingService.ISLastReleaseStopType(appId, envId)
			if err != nil {
				handler.logger.Errorw("service err, FetchAppDetailsV2", "err", err, "app", appId, "env", envId)
			} else if status {
				resp.Status = application.HIBERNATING
			}
		}
		if resp.Status == string(health.HealthStatusDegraded) {
			count, err := handler.appListingService.GetReleaseCount(appId, envId)
			if err != nil {
				handler.logger.Errorw("service err, FetchAppDetailsV2, release count", "err", err, "app", appId, "env", envId)
			} else if count == 0 {
				resp.Status = app.NotDeployed
			}
		}
		resourceTree = util2.InterfaceToMapAdapter(resp)
		go func() {
			if resp.Status == string(health.HealthStatusHealthy) {
				err = handler.cdApplicationStatusUpdateHandler.SyncPipelineStatusForResourceTreeCall(cdPipeline)
				if err != nil {
					handler.logger.Errorw("error in syncing pipeline status", "err", err)
				}
			}
			//updating app_status table here
			err = handler.appStatusService.UpdateStatusWithAppIdEnvId(appId, envId, resp.Status)
			if err != nil {
				handler.logger.Warnw("error in updating app status", "err", err, "appId", cdPipeline.AppId, "envId", cdPipeline.EnvironmentId)
			}
		}()

	} else if len(cdPipeline.DeploymentAppName) > 0 && cdPipeline.EnvironmentId > 0 && util.IsHelmApp(cdPipeline.DeploymentAppType) {
		config, err := handler.helmAppService.GetClusterConf(cdPipeline.Environment.ClusterId)
		if err != nil {
			handler.logger.Errorw("error in fetching cluster detail", "err", err)
		}
		req := &client.AppDetailRequest{
			ClusterConfig: config,
			Namespace:     cdPipeline.Environment.Namespace,
			ReleaseName:   cdPipeline.DeploymentAppName,
		}
		detail, err := handler.helmAppClient.GetAppDetail(context.Background(), req)
		if err != nil {
			handler.logger.Errorw("error in fetching app detail", "err", err)
		}
		if detail != nil {
			resourceTree = util2.InterfaceToMapAdapter(detail.ResourceTreeResponse)
			applicationStatus := detail.ApplicationStatus
			resourceTree["status"] = applicationStatus
			if applicationStatus == application.Healthy {
				status, err := handler.appListingService.ISLastReleaseStopType(appId, envId)
				if err != nil {
					handler.logger.Errorw("service err, FetchAppDetailsV2", "err", err, "app", appId, "env", envId)
				} else if status {
					resourceTree["status"] = application.HIBERNATING
				}
			}
			handler.logger.Warnw("appName and envName not found - avoiding resource tree call", "app", cdPipeline.DeploymentAppName, "env", cdPipeline.Environment.Name)
		}
	} else {
		handler.logger.Warnw("appName and envName not found - avoiding resource tree call", "app", cdPipeline.DeploymentAppName, "env", cdPipeline.Environment.Name)
	}
	if resourceTree != nil {
		version, err := handler.k8sCommonService.GetK8sServerVersion(cdPipeline.Environment.ClusterId)
		if err != nil {
			handler.logger.Errorw("error in fetching k8s version in resource tree call fetching", "clusterId", cdPipeline.Environment.ClusterId, "err", err)
		} else {
			resourceTree["serverVersion"] = version.String()
		}
	}
	validRequests := make([]k8s.ResourceRequestBean, 0)
	k8sAppDetail := bean.AppDetailContainer{
		DeploymentDetailContainer: bean.DeploymentDetailContainer{
			ClusterId: cdPipeline.Environment.ClusterId,
			Namespace: cdPipeline.Environment.Namespace,
		},
	}
	clusterIdString := strconv.Itoa(cdPipeline.Environment.ClusterId)
	validRequest := handler.k8sCommonService.FilterK8sResources(r.Context(), resourceTree, validRequests, k8sAppDetail, clusterIdString, []string{k8s.ServiceKind, k8s.EndpointsKind, k8s.IngressKind})
	resp, err := handler.k8sCommonService.GetManifestsByBatch(r.Context(), validRequest)
	ports := make([]int64, 0)
	for _, portHolder := range resp {
		if portHolder.ManifestResponse.Manifest.Object["kind"] == "Service" {
			spec := portHolder.ManifestResponse.Manifest.Object["spec"].(map[string]interface{})
			if spec != nil {
				portList := spec["ports"].([]interface{})
				for _, portItem := range portList {
					if portItem.(map[string]interface{}) != nil {
						_portNumber := portItem.(map[string]interface{})["port"]
						portNumber := _portNumber.(int64)
						if portNumber != 0 {
							ports = append(ports, portNumber)
						}
					}
				}
			} else {
				handler.logger.Errorw("spec doest not contain data", "err", spec)
			}
		}
		if portHolder.ManifestResponse.Manifest.Object["kind"] == "Endpoints" {
			if portHolder.ManifestResponse.Manifest.Object["subsets"] != nil {
				subsets := portHolder.ManifestResponse.Manifest.Object["subsets"].([]interface{})
				for _, subset := range subsets {
					subsetObj := subset.(map[string]interface{})
					if subsetObj != nil {
						portsIfs := subsetObj["ports"].([]interface{})
						for _, portsIf := range portsIfs {
							portsIfObj := portsIf.(map[string]interface{})
							if portsIfObj != nil {
								port := portsIfObj["port"].(int64)
								ports = append(ports, port)
							}
						}
					}
				}
			}
		}
		if portHolder.ManifestResponse.Manifest.Object["kind"] == "EndpointSlice" {
			if portHolder.ManifestResponse.Manifest.Object["ports"] != nil {
				endPointsSlicePorts := portHolder.ManifestResponse.Manifest.Object["ports"].([]interface{})
				for _, val := range endPointsSlicePorts {
					_portNumber := val.(map[string]interface{})["port"]
					portNumber := _portNumber.(int64)
					if portNumber != 0 {
						ports = append(ports, portNumber)
					}
				}
			}
		}
	}
	if err != nil {
		handler.logger.Errorw("error in fetching manifest", "err", err)
	}
	if val, ok := resourceTree["nodes"]; ok {
		resourceTreeVal := val.([]interface{})
		for _, val := range resourceTreeVal {
			_value := val.(map[string]interface{})
			for key, _type := range _value {
				if key == "kind" && _type == "Endpoints" {
					_value["port"] = ports
				}
				if key == "kind" && _type == "Service" {
					_value["port"] = ports
				}
			}
		}
	}
	return resourceTree, nil
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
	if app.AppType == helper.ChartStoreApp {
		err = handler.cdApplicationStatusUpdateHandler.ManualSyncPipelineStatus(appId, 0, userId)
	} else {
		err = handler.cdApplicationStatusUpdateHandler.ManualSyncPipelineStatus(appId, envId, userId)
	}

	if err != nil {
		handler.logger.Errorw("service err, ManualSyncAcdPipelineDeploymentStatus", "err", err, "appId", appId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "App synced successfully.", http.StatusOK)
}

func (handler AppListingRestHandlerImpl) GetClusterTeamAndEnvListForAutocomplete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	clusterMapping := make(map[string]cluster.ClusterBean)
	start := time.Now()
	clusterList, err := handler.clusterService.FindAllForAutoComplete()
	dbOperationTime := time.Since(start)
	if err != nil {
		handler.logger.Errorw("service err, FindAllForAutoComplete in clusterService layer", "error", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	var granterClusters []cluster.ClusterBean
	v := r.URL.Query()
	authEnabled := true
	auth := v.Get("auth")
	if len(auth) > 0 {
		authEnabled, err = strconv.ParseBool(auth)
		if err != nil {
			authEnabled = true
			err = nil
			//ignore error, apply rbac by default
		}
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	start = time.Now()
	for _, item := range clusterList {
		clusterMapping[strings.ToLower(item.ClusterName)] = item
		if authEnabled == true {
			if ok := handler.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionGet, item.ClusterName); ok {
				granterClusters = append(granterClusters, item)
			}
		} else {
			granterClusters = append(granterClusters, item)
		}

	}
	handler.logger.Infow("Cluster elapsed Time for enforcer", "dbElapsedTime", dbOperationTime, "enforcerTime", time.Since(start), "envSize", len(granterClusters))
	//RBAC enforcer Ends

	if len(granterClusters) == 0 {
		granterClusters = make([]cluster.ClusterBean, 0)
	}

	//getting environment for autocomplete
	start = time.Now()
	environments, err := handler.environmentClusterMappingsService.GetEnvironmentOnlyListForAutocomplete()
	if err != nil {
		handler.logger.Errorw("service err, GetEnvironmentListForAutocomplete at environmentClusterMappingsService layer", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	dbElapsedTime := time.Since(start)
	var grantedEnvironment = environments
	start = time.Now()
	println(dbElapsedTime, grantedEnvironment)
	if !handler.cfg.IgnoreAuthCheck {
		grantedEnvironment = make([]cluster.EnvironmentBean, 0)
		emailId, _ := handler.userService.GetEmailFromToken(token)
		// RBAC enforcer applying
		var envIdentifierList []string
		for index, item := range environments {
			clusterName := strings.ToLower(strings.Split(item.EnvironmentIdentifier, "__")[0])
			if clusterMapping[clusterName].Id != 0 {
				environments[index].CdArgoSetup = clusterMapping[clusterName].IsCdArgoSetup
				environments[index].ClusterName = clusterMapping[clusterName].ClusterName
			}
			envIdentifierList = append(envIdentifierList, strings.ToLower(item.EnvironmentIdentifier))
		}

		result := handler.enforcer.EnforceByEmailInBatch(emailId, casbin.ResourceGlobalEnvironment, casbin.ActionGet, envIdentifierList)
		for _, item := range environments {

			var hasAccess bool
			EnvironmentIdentifier := item.ClusterName + "__" + item.Namespace
			if item.EnvironmentIdentifier != EnvironmentIdentifier {
				// fix for futuristic case
				hasAccess = result[strings.ToLower(EnvironmentIdentifier)] || result[strings.ToLower(item.EnvironmentIdentifier)]
			} else {
				hasAccess = result[strings.ToLower(item.EnvironmentIdentifier)]
			}
			if hasAccess {
				grantedEnvironment = append(grantedEnvironment, item)
			}
		}
		//RBAC enforcer Ends
	}
	elapsedTime := time.Since(start)
	handler.logger.Infow("Env elapsed Time for enforcer", "dbElapsedTime", dbElapsedTime, "elapsedTime",
		elapsedTime, "envSize", len(grantedEnvironment))

	//getting teams for autocomplete
	start = time.Now()
	teams, err := handler.teamService.FetchForAutocomplete()
	if err != nil {
		handler.logger.Errorw("service err, FetchForAutocomplete at teamService layer", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	dbElapsedTime = time.Since(start)
	var grantedTeams = teams
	start = time.Now()
	if !handler.cfg.IgnoreAuthCheck {
		grantedTeams = make([]team.TeamRequest, 0)
		emailId, _ := handler.userService.GetEmailFromToken(token)
		// RBAC enforcer applying
		var teamNameList []string
		for _, item := range teams {
			teamNameList = append(teamNameList, strings.ToLower(item.Name))
		}

		result := handler.enforcer.EnforceByEmailInBatch(emailId, casbin.ResourceTeam, casbin.ActionGet, teamNameList)

		for _, item := range teams {
			if hasAccess := result[strings.ToLower(item.Name)]; hasAccess {
				grantedTeams = append(grantedTeams, item)
			}
		}
	}
	handler.logger.Infow("Team elapsed Time for enforcer", "dbElapsedTime", dbElapsedTime, "elapsedTime", time.Since(start),
		"envSize", len(grantedTeams))

	//RBAC enforcer Ends
	resp := &AppAutocomplete{
		Teams:        grantedTeams,
		Environments: grantedEnvironment,
		Clusters:     granterClusters,
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)

}
