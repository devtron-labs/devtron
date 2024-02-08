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

package appList

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	client "github.com/devtron-labs/devtron/api/helm-app/service"
	util3 "github.com/devtron-labs/devtron/api/util"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode/resource"
	"net/http"
	"strconv"
	"time"

	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	k8sObjectUtils "github.com/devtron-labs/common-lib/utils/k8sObjectsUtil"
	"github.com/devtron-labs/devtron/api/bean"
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
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/deploymentGroup"
	"github.com/devtron-labs/devtron/pkg/generateManifest"
	"github.com/devtron-labs/devtron/pkg/genericNotes"
	"github.com/devtron-labs/devtron/pkg/k8s"
	application3 "github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type AppListingRestHandler interface {
	FetchAppDetails(w http.ResponseWriter, r *http.Request)
	FetchJobs(w http.ResponseWriter, r *http.Request)
	FetchJobOverviewCiPipelines(w http.ResponseWriter, r *http.Request)
	FetchAppDetailsV2(w http.ResponseWriter, r *http.Request)
	FetchResourceTree(w http.ResponseWriter, r *http.Request)
	FetchAllDevtronManagedApps(w http.ResponseWriter, r *http.Request)
	FetchAppStageStatus(w http.ResponseWriter, r *http.Request)

	FetchOtherEnvironment(w http.ResponseWriter, r *http.Request)
	FetchMinDetailOtherEnvironment(w http.ResponseWriter, r *http.Request)
	RedirectToLinkouts(w http.ResponseWriter, r *http.Request)
	GetHostUrlsByBatch(w http.ResponseWriter, r *http.Request)

	FetchAppsByEnvironmentV2(w http.ResponseWriter, r *http.Request)
	FetchOverviewAppsByEnvironment(w http.ResponseWriter, r *http.Request)
}

type AppListingRestHandlerImpl struct {
	application            application.ServiceClient
	appListingService      app.AppListingService
	enforcer               casbin.Enforcer
	pipeline               pipeline.PipelineBuilder
	logger                 *zap.SugaredLogger
	enforcerUtil           rbac.EnforcerUtil
	deploymentGroupService deploymentGroup.DeploymentGroupService
	userService            user.UserService
	// TODO fix me next
	helmAppClient                    gRPC.HelmAppClient // TODO refactoring: use HelmAppService
	helmAppService                   client.HelmAppService
	argoUserService                  argo.ArgoUserService
	k8sCommonService                 k8s.K8sCommonService
	installedAppService              FullMode.InstalledAppDBExtendedService
	installedAppResourceService      resource.InstalledAppResourceService
	cdApplicationStatusUpdateHandler cron.CdApplicationStatusUpdateHandler
	pipelineRepository               pipelineConfig.PipelineRepository
	appStatusService                 appStatus.AppStatusService
	installedAppRepository           repository.InstalledAppRepository
	genericNoteService               genericNotes.GenericNoteService
	k8sApplicationService            application3.K8sApplicationService
	deploymentTemplateService        generateManifest.DeploymentTemplateService
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
	enforcer casbin.Enforcer,
	pipeline pipeline.PipelineBuilder,
	logger *zap.SugaredLogger, enforcerUtil rbac.EnforcerUtil,
	deploymentGroupService deploymentGroup.DeploymentGroupService, userService user.UserService,
	helmAppClient gRPC.HelmAppClient, helmAppService client.HelmAppService,
	argoUserService argo.ArgoUserService, k8sCommonService k8s.K8sCommonService,
	installedAppService FullMode.InstalledAppDBExtendedService,
	installedAppResourceService resource.InstalledAppResourceService,
	cdApplicationStatusUpdateHandler cron.CdApplicationStatusUpdateHandler,
	pipelineRepository pipelineConfig.PipelineRepository,
	appStatusService appStatus.AppStatusService, installedAppRepository repository.InstalledAppRepository,
	genericNoteService genericNotes.GenericNoteService,
	k8sApplicationService application3.K8sApplicationService, deploymentTemplateService generateManifest.DeploymentTemplateService,
) *AppListingRestHandlerImpl {
	appListingHandler := &AppListingRestHandlerImpl{
		application:                      application,
		appListingService:                appListingService,
		logger:                           logger,
		pipeline:                         pipeline,
		enforcer:                         enforcer,
		enforcerUtil:                     enforcerUtil,
		deploymentGroupService:           deploymentGroupService,
		userService:                      userService,
		helmAppClient:                    helmAppClient,
		helmAppService:                   helmAppService,
		argoUserService:                  argoUserService,
		k8sCommonService:                 k8sCommonService,
		installedAppService:              installedAppService,
		installedAppResourceService:      installedAppResourceService,
		cdApplicationStatusUpdateHandler: cdApplicationStatusUpdateHandler,
		pipelineRepository:               pipelineRepository,
		appStatusService:                 appStatusService,
		installedAppRepository:           installedAppRepository,
		genericNoteService:               genericNoteService,
		k8sApplicationService:            k8sApplicationService,
		deploymentTemplateService:        deploymentTemplateService,
	}
	return appListingHandler
}

func (handler AppListingRestHandlerImpl) FetchAllDevtronManagedApps(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	handler.logger.Infow("got request to fetch all devtron managed apps ", "userId", userId)
	// RBAC starts
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		handler.logger.Infow("user forbidden to fetch all devtron managed apps", "userId", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	// RBAC ends
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
	token := r.Header.Get("token")
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	var validAppIds []int
	// for non super admin users
	if !isSuperAdmin {
		rbacObjectsForAllAppsMap := handler.enforcerUtil.GetRbacObjectsForAllApps(helper.Job)
		rbacObjectToAppIdMap := make(map[string]int)
		rbacObjects := make([]string, len(rbacObjectsForAllAppsMap))
		itr := 0
		for appId, object := range rbacObjectsForAllAppsMap {
			rbacObjects[itr] = object
			rbacObjectToAppIdMap[object] = appId
			itr++
		}

		result := handler.enforcer.EnforceInBatch(token, casbin.ResourceJobs, casbin.ActionGet, rbacObjects)
		// O(n) loop, n = len(rbacObjectsForAllAppsMap)
		for object, ok := range result {
			if ok {
				validAppIds = append(validAppIds, rbacObjectToAppIdMap[object])
			}
		}

		if len(validAppIds) == 0 {
			handler.logger.Infow("user doesn't have access to any app", "userId", userId)
			common.WriteJsonResp(w, err, bean.JobContainerResponse{}, http.StatusOK)
			return
		}
	}
	var fetchJobListingRequest app.FetchAppListingRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&fetchJobListingRequest)
	if err != nil {
		handler.logger.Errorw("request err, FetchAppsByEnvironment", "err", err, "payload", fetchJobListingRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// fetching only those jobs whose access user has by setting valid app Ids.
	fetchJobListingRequest.AppIds = validAppIds

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
	vars := mux.Vars(r)
	jobId, err := strconv.Atoi(vars["jobId"])
	if err != nil {
		handler.logger.Errorw("request err, GetAppMetaInfo", "err", err, "jobId", jobId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(jobId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceJobs, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	// RBAC ENDS
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

func (handler AppListingRestHandlerImpl) FetchAppsByEnvironmentV2(w http.ResponseWriter, r *http.Request) {
	//Allow CORS here By * or specific origin
	util3.SetupCorsOriginHeader(&w)
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
	span.End()
	newCtx, span = otel.Tracer("userService").Start(newCtx, "IsSuperAdmin")
	isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	span.End()
	validAppIds := make([]int, 0)
	// for non super admin users
	if !isActionUserSuperAdmin {
		rbacObjectsForAllAppsMap := handler.enforcerUtil.GetRbacObjectsForAllApps(helper.CustomApp)
		rbacObjectToAppIdMap := make(map[string]int)
		rbacObjects := make([]string, len(rbacObjectsForAllAppsMap))
		itr := 0
		for appId, object := range rbacObjectsForAllAppsMap {
			rbacObjects[itr] = object
			rbacObjectToAppIdMap[object] = appId
			itr++
		}

		result := handler.enforcer.EnforceInBatch(token, casbin.ResourceApplications, casbin.ActionGet, rbacObjects)
		// O(n) loop, n = len(rbacObjectsForAllAppsMap)
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
			Id:                   dg.Id,
			Name:                 dg.Name,
			AppCount:             dg.AppCount,
			NoOfApps:             dg.NoOfApps,
			EnvironmentId:        dg.EnvironmentId,
			CiPipelineId:         dg.CiPipelineId,
			CiMaterialDTOs:       ciMaterialDTOs,
			IsVirtualEnvironment: dg.IsVirtualEnvironment,
		}
	}
	t2 = time.Now()
	handler.logger.Infow("api response time testing", "time", time.Now().String(), "time diff", t2.Unix()-t1.Unix(), "stage", "4")
	t1 = t2
	handler.logger.Infow("api response time testing", "total time", time.Now().String(), "total time", t1.Unix()-t0.Unix())
	common.WriteJsonResp(w, err, appContainerResponse, http.StatusOK)
}

// TODO refactoring: use schema.NewDecoder().Decode(&queryStruct, r.URL.Query())
func (handler AppListingRestHandlerImpl) FetchOverviewAppsByEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
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
	// return all if user is super admin
	if isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); isActionUserSuperAdmin {
		common.WriteJsonResp(w, err, resp, http.StatusOK)
		return
	}

	// get all the appIds
	appIds := make([]int, 0)
	appContainers := resp.Apps
	for _, appBean := range resp.Apps {
		appIds = append(appIds, appBean.AppId)
	}

	// get rbac objects for the appids
	rbacObjectsWithAppId := handler.enforcerUtil.GetRbacObjectsByAppIds(appIds)
	rbacObjects := make([]string, len(rbacObjectsWithAppId))
	itr := 0
	for _, object := range rbacObjectsWithAppId {
		rbacObjects[itr] = object
		itr++
	}
	// enforce rbac in batch
	rbacResult := handler.enforcer.EnforceInBatch(token, casbin.ResourceApplications, casbin.ActionGet, rbacObjects)
	// filter out rbac passed apps
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
	if cdPipeline.Environment.IsVirtualEnvironment {
		common.WriteJsonResp(w, nil, nil, http.StatusOK)
		return
	}
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
	if err != nil {
		handler.logger.Errorw("error in fetching resource tree", "err", err, "appId", appId, "envId", envId)
		handler.handleResourceTreeErrAndDeletePipelineIfNeeded(w, err, appId, envId, acdToken, cdPipeline)
		return
	}
	common.WriteJsonResp(w, err, resourceTree, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) handleResourceTreeErrAndDeletePipelineIfNeeded(w http.ResponseWriter, err error,
	appId int, envId int, acdToken string, cdPipeline *pipelineConfig.Pipeline) {
	if cdPipeline.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD {
		apiError, ok := err.(*util.ApiError)
		if ok && apiError != nil {
			if apiError.Code == constants.AppDetailResourceTreeNotFound && cdPipeline.DeploymentAppDeleteRequest == true && cdPipeline.DeploymentAppCreated == true {
				acdAppFound, appDeleteErr := handler.pipeline.MarkGitOpsDevtronAppsDeletedWhereArgoAppIsDeleted(appId, envId, acdToken, cdPipeline)
				if appDeleteErr != nil {
					common.WriteJsonResp(w, fmt.Errorf("error in deleting devtron pipeline for deleted argocd app"), nil, http.StatusInternalServerError)
					return
				} else if appDeleteErr == nil && !acdAppFound {
					common.WriteJsonResp(w, fmt.Errorf("argocd app deleted"), nil, http.StatusNotFound)
					return
				}
			}
		}
	}
	// not returned yet therefore no specific error to be handled, send error in internal message
	common.WriteJsonResp(w, &util.ApiError{
		InternalMessage: err.Error(),
		UserMessage:     "unable to fetch resource tree",
	}, nil, http.StatusInternalServerError)
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
	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, object, casbin.ActionGet)
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	// RBAC enforcer Ends

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
	// RBAC enforcer Ends

	newCtx, span = otel.Tracer("appListingService").Start(newCtx, "FetchOtherEnvironment")
	otherEnvironment, err := handler.appListingService.FetchOtherEnvironment(newCtx, appId)
	span.End()
	if err != nil {
		handler.logger.Errorw("service err, FetchOtherEnvironment", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// TODO - rbac env level

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
	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, object, casbin.ActionGet)
	if !ok {
		common.WriteJsonResp(w, err, "unauthorized user", http.StatusForbidden)
		return
	}
	// RBAC enforcer Ends

	otherEnvironment, err := handler.appListingService.FetchMinDetailOtherEnvironment(appId)
	if err != nil {
		handler.logger.Errorw("service err, FetchMinDetailOtherEnvironment", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// TODO - rbac env level

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
	// RBAC enforcer Ends

	link, err := handler.appListingService.RedirectToLinkouts(Id, appId, envId, podName, containerName)
	if err != nil || len(link) == 0 {
		handler.logger.Errorw("service err, RedirectToLinkouts", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, link, http.StatusOK)
}
func (handler AppListingRestHandlerImpl) fetchResourceTreeFromInstallAppService(w http.ResponseWriter, r *http.Request, resourceTreeAndNotesContainer bean.AppDetailsContainer, installedApps repository.InstalledApps) (bean.AppDetailsContainer, error) {
	rctx := r.Context()
	cn, _ := w.(http.CloseNotifier)
	err := handler.installedAppResourceService.FetchResourceTree(rctx, cn, &resourceTreeAndNotesContainer, installedApps, "", "")
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
	// check user authorization for this app
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
		resourceTreeAndNotesContainer := bean.AppDetailsContainer{}
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
	// valid batch requests, only valid requests will be sent for batch processing
	validRequests := handler.k8sCommonService.FilterK8sResources(r.Context(), resourceTree, appDetail, "", []string{k8sCommonBean.ServiceKind, k8sCommonBean.IngressKind})
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
		// RBAC enforcer Ends
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

		// we currently add appId and envId as labels for devtron apps deployed via acd
		label := fmt.Sprintf("appId=%v,envId=%v", cdPipeline.AppId, cdPipeline.EnvironmentId)
		pods, err := handler.k8sApplicationService.GetPodListByLabel(cdPipeline.Environment.ClusterId, cdPipeline.Environment.Namespace, label)
		if err != nil {
			handler.logger.Errorw("error in getting pods by label", "err", err, "clusterId", cdPipeline.Environment.ClusterId, "namespace", cdPipeline.Environment.Namespace, "label", label)
			return resourceTree, err
		}
		ephemeralContainersMap := k8sObjectUtils.ExtractEphemeralContainers(pods)
		for _, metaData := range resp.PodMetadata {
			metaData.EphemeralContainers = ephemeralContainersMap[metaData.Name]
		}

		if resp.Status == string(health.HealthStatusHealthy) {
			status, err := handler.appListingService.ISLastReleaseStopType(appId, envId)
			if err != nil {
				handler.logger.Errorw("service err, FetchAppDetailsV2", "err", err, "app", appId, "env", envId)
			} else if status {
				resp.Status = argoApplication.HIBERNATING
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
			// updating app_status table here
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
		req := &gRPC.AppDetailRequest{
			ClusterConfig: config,
			Namespace:     cdPipeline.Environment.Namespace,
			ReleaseName:   cdPipeline.DeploymentAppName,
		}
		detail, err := handler.helmAppClient.GetAppDetail(context.Background(), req)
		if err != nil {
			handler.logger.Errorw("error in fetching app detail", "err", err)
		}
		if detail != nil && detail.ReleaseExist {
			resourceTree = util2.InterfaceToMapAdapter(detail.ResourceTreeResponse)
			releaseStatus := util2.InterfaceToMapAdapter(detail.ReleaseStatus)
			applicationStatus := detail.ApplicationStatus
			resourceTree["releaseStatus"] = releaseStatus
			resourceTree["status"] = applicationStatus
			if applicationStatus == argoApplication.Healthy {
				status, err := handler.appListingService.ISLastReleaseStopType(appId, envId)
				if err != nil {
					handler.logger.Errorw("service err, FetchAppDetailsV2", "err", err, "app", appId, "env", envId)
				} else if status {
					resourceTree["status"] = argoApplication.HIBERNATING
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
	k8sAppDetail := bean.AppDetailContainer{
		DeploymentDetailContainer: bean.DeploymentDetailContainer{
			ClusterId: cdPipeline.Environment.ClusterId,
			Namespace: cdPipeline.Environment.Namespace,
		},
	}
	clusterIdString := strconv.Itoa(cdPipeline.Environment.ClusterId)
	validRequest := handler.k8sCommonService.FilterK8sResources(r.Context(), resourceTree, k8sAppDetail, clusterIdString, []string{k8sCommonBean.ServiceKind, k8sCommonBean.EndpointsKind, k8sCommonBean.IngressKind})
	resp, err := handler.k8sCommonService.GetManifestsByBatch(r.Context(), validRequest)
	if err != nil {
		handler.logger.Errorw("error in getting manifest by batch", "err", err, "clusterId", clusterIdString)
		return nil, err
	}
	newResourceTree := handler.k8sCommonService.PortNumberExtraction(resp, resourceTree)
	return newResourceTree, nil
}
