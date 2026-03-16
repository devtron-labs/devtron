/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package appList

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/devtron-labs/devtron/api/bean/AppView"
	util3 "github.com/devtron-labs/devtron/api/util"
	"github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode/resource"
	appStoreUtil "github.com/devtron-labs/devtron/pkg/appStore/util"
	util2 "github.com/devtron-labs/devtron/pkg/auth/user/util"
	clusterBean "github.com/devtron-labs/devtron/pkg/cluster/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/environment/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp/status/resourceTree"
	"github.com/devtron-labs/devtron/pkg/deployment/resourceTree/devtronApp"
	"github.com/devtron-labs/devtron/pkg/featureFlag"
	"github.com/devtron-labs/devtron/pkg/featureFlag/model"
	ffService "github.com/devtron-labs/devtron/pkg/featureFlag/service"
	"github.com/devtron-labs/devtron/pkg/globalFlag"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	read2 "github.com/devtron-labs/devtron/pkg/policyGovernance/approvalConfig/read"
	devtronUtil "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac/filter"
	"github.com/gorilla/schema"
	"golang.org/x/exp/maps"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	_ "github.com/devtron-labs/devtron/pkg/cluster"
	common2 "github.com/devtron-labs/devtron/pkg/deployment/common"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/pkg/deploymentGroup"
	"github.com/devtron-labs/devtron/pkg/k8s"
	k8sApplication "github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean5 "github.com/devtron-labs/devtron/pkg/team/bean"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type AppListingRestHandler interface {
	FetchJobs(w http.ResponseWriter, r *http.Request)
	FetchJobOverviewCiPipelines(w http.ResponseWriter, r *http.Request)
	FetchAppDetailsV2(w http.ResponseWriter, r *http.Request)
	FetchResourceTree(w http.ResponseWriter, r *http.Request)
	FetchAllDevtronManagedApps(w http.ResponseWriter, r *http.Request)
	FetchAppStageStatus(w http.ResponseWriter, r *http.Request)

	FetchEnvMinData(w http.ResponseWriter, r *http.Request)
	FetchOtherEnvironment(w http.ResponseWriter, r *http.Request)
	FetchMinDetailOtherEnvironment(w http.ResponseWriter, r *http.Request)
	RedirectToLinkouts(w http.ResponseWriter, r *http.Request)
	GetHostUrlsByBatch(w http.ResponseWriter, r *http.Request)

	FetchAppsByEnvironmentV2(w http.ResponseWriter, r *http.Request)
	FetchOverviewAppsByEnvironment(w http.ResponseWriter, r *http.Request)

	//ent
	FetchAppPolicyConsequences(w http.ResponseWriter, r *http.Request)
	FetchAutocompleteJobCiPipelines(w http.ResponseWriter, r *http.Request)
	GetAllAppEnvsFromResourceNames(w http.ResponseWriter, r *http.Request)
}

type AppListingRestHandlerImpl struct {
	appListingService           app.AppListingService
	enforcer                    casbin.Enforcer
	pipeline                    pipeline.PipelineBuilder
	logger                      *zap.SugaredLogger
	enforcerUtil                rbac.EnforcerUtil
	deploymentGroupService      deploymentGroup.DeploymentGroupService
	userService                 user.UserService
	k8sCommonService            k8s.K8sCommonService
	installedAppService         FullMode.InstalledAppDBExtendedService
	installedAppResourceService resource.InstalledAppResourceService
	pipelineRepository          pipelineConfig.PipelineRepository
	k8sApplicationService       k8sApplication.K8sApplicationService
	deploymentConfigService     common2.DeploymentConfigService
	resourceTreeService         resourceTree.Service

	//ent
	virtualEnvResourceTreeService           devtronApp.VirtualEnvResourceTreeService
	featureFlagService                      ffService.FeatureFlagService
	approvalConfigurationEnforcementService read2.ApprovalPolicyReadService
	rbacFilterUtil                          filter.RbacFilterUtil
	globalFlagService                       globalFlag.GlobalFlagService
	globalPolicyDataManager                 globalPolicy.GlobalPolicyDataManager
}

type AppStatus struct {
	name       string
	status     string
	message    string
	err        error
	conditions []v1alpha1.ApplicationCondition
}

type AppAutocomplete struct {
	Teams        []bean5.TeamRequest
	Environments []bean2.EnvironmentBean
	Clusters     []clusterBean.ClusterBean
}

func NewAppListingRestHandlerImpl(
	appListingService app.AppListingService,
	enforcer casbin.Enforcer,
	pipeline pipeline.PipelineBuilder,
	logger *zap.SugaredLogger, enforcerUtil rbac.EnforcerUtil,
	deploymentGroupService deploymentGroup.DeploymentGroupService, userService user.UserService,
	k8sCommonService k8s.K8sCommonService,
	installedAppService FullMode.InstalledAppDBExtendedService,
	installedAppResourceService resource.InstalledAppResourceService,
	pipelineRepository pipelineConfig.PipelineRepository,
	k8sApplicationService k8sApplication.K8sApplicationService,
	deploymentConfigService common2.DeploymentConfigService,
	virtualEnvResourceTreeService devtronApp.VirtualEnvResourceTreeService,
	featureFlagService ffService.FeatureFlagService,
	approvalConfigurationEnforcementService read2.ApprovalPolicyReadService,
	resourceTreeService resourceTree.Service,
	rbacFilterUtil filter.RbacFilterUtil,
	globalFlagService globalFlag.GlobalFlagService,
	globalPolicyDataManager globalPolicy.GlobalPolicyDataManager,
) *AppListingRestHandlerImpl {
	appListingHandler := &AppListingRestHandlerImpl{
		appListingService:           appListingService,
		logger:                      logger,
		pipeline:                    pipeline,
		enforcer:                    enforcer,
		enforcerUtil:                enforcerUtil,
		deploymentGroupService:      deploymentGroupService,
		userService:                 userService,
		k8sCommonService:            k8sCommonService,
		installedAppService:         installedAppService,
		installedAppResourceService: installedAppResourceService,
		pipelineRepository:          pipelineRepository,
		k8sApplicationService:       k8sApplicationService,
		deploymentConfigService:     deploymentConfigService,
		resourceTreeService:         resourceTreeService,

		virtualEnvResourceTreeService:           virtualEnvResourceTreeService,
		featureFlagService:                      featureFlagService,
		approvalConfigurationEnforcementService: approvalConfigurationEnforcementService,
		rbacFilterUtil:                          rbacFilterUtil,
		globalFlagService:                       globalFlagService,
		globalPolicyDataManager:                 globalPolicyDataManager,
	}
	return appListingHandler
}

func (handler AppListingRestHandlerImpl) FetchAllDevtronManagedApps(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
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
		common.HandleUnauthorized(w, r)
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
			common.WriteJsonResp(w, err, AppView.JobContainerResponse{}, http.StatusOK)
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
	jobContainerResponse := AppView.JobContainerResponse{
		JobContainers: jobs,
		JobCount:      jobsCount,
	}

	common.WriteJsonResp(w, err, jobContainerResponse, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) FetchJobOverviewCiPipelines(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		handler.logger.Errorw("request err, userId", "err", err, "payload", userId)
		common.HandleUnauthorized(w, r)
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
		common.HandleUnauthorized(w, r)
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
			common.WriteJsonResp(w, err, AppView.AppContainerResponse{}, http.StatusOK)
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
	normalizedTagFilters, err := app.NormalizeAndValidateTagFilters(fetchAppListingRequest.TagFilters)
	if err != nil {
		handler.logger.Errorw("request err, ValidateTagFilters", "err", err, "payload", fetchAppListingRequest.TagFilters)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	fetchAppListingRequest.TagFilters = normalizedTagFilters
	newCtx, span = otel.Tracer("fetchAppListingRequest").Start(newCtx, "GetNamespaceClusterMapping")
	_, _, err = app.GetNamespaceClusterMapping(fetchAppListingRequest.Namespaces)
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
			return
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
		return
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
		return
	}

	// RBAC start
	//canOnlyViewPermittedData, err := handler.globalFlagService.CanOnlyViewPermittedData(userId)
	//if err != nil {
	//	handler.logger.Errorw("service err, FetchAppsByEnvironment", "err", err, "payload", fetchAppListingRequest)
	//	common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
	//	return
	//}
	//
	//if canOnlyViewPermittedData {
	//	filteredEnvContainers := handler.filterAppEnvironmentContainers(envContainers, token)
	//}

	// TODO logic to update the response with filtered envContainers
	//RBAC ends

	appContainerResponse := AppView.AppContainerResponse{
		AppContainers: apps,
		AppCount:      appsCount,
	}
	if fetchAppListingRequest.DeploymentGroupId > 0 {
		var ciMaterialDTOs []AppView.CiMaterialDTO
		for _, ci := range dg.CiMaterialDTOs {
			ciMaterialDTOs = append(ciMaterialDTOs, AppView.CiMaterialDTO{
				Name:        ci.Name,
				SourceValue: ci.SourceValue,
				SourceType:  ci.SourceType,
			})
		}
		appContainerResponse.DeploymentGroupDTO = AppView.DeploymentGroupDTO{
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

// this function gives only those envContainers which are permitted to the user, this can be used in future in app listing to
// filter out the envContainers which are not permitted to the user, after building the response, iterate over apps -> it's envContainers
// and filter out the envContainers which are not permitted to the user or put a particular message instead of envContainers
func (impl AppListingRestHandlerImpl) filterAppEnvironmentContainers(envContainers []*AppView.AppEnvironmentContainer, token string) []*AppView.AppEnvironmentContainer {
	// make envRBACObjects
	envRBACObjects := impl.enforcerUtil.GetEnvRbacObjectsByEnvContainers(envContainers)

	envResults := impl.enforcer.EnforceInBatch(token, casbin.ResourceEnvironment, casbin.ActionGet, envRBACObjects)
	filteredEnvContainers := make([]*AppView.AppEnvironmentContainer, 0)
	for _, envContainer := range envContainers {
		rbacObject := impl.enforcerUtil.GetEnvRbacObjectsByEnvContainer(envContainer)
		// if rbac object exist into envResults then add to filteredEnvContainers
		if envResults[rbacObject] {
			filteredEnvContainers = append(filteredEnvContainers, envContainer)
		}

	}
	return filteredEnvContainers
}

// TODO refactoring: use schema.NewDecoder().Decode(&queryStruct, r.URL.Query())
func (handler AppListingRestHandlerImpl) FetchOverviewAppsByEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
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
	resp.Apps = make([]*AppView.AppEnvironmentContainer, 0)
	for _, appBean := range appContainers {
		rbacObject := rbacObjectsWithAppId[appBean.AppId]
		if rbacResult[rbacObject] {
			resp.Apps = append(resp.Apps, appBean)
		}
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)

}

func (handler AppListingRestHandlerImpl) FetchAppDetailsV2(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
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
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	userMetadata := util2.GetUserMetadata(r.Context(), userId, isSuperAdmin)
	appDetail, err := handler.appListingService.FetchAppDetails(r.Context(), appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, FetchAppDetailsV2", "err", err, "appId", appId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	appDetail, err = handler.updateApprovalConfigDataInAppDetailResp(r.Context(), appDetail, appId, envId, userMetadata)
	if err != nil {
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
	ctx = context.WithValue(ctx, "token", token)
	ctx = devtronUtil.SetSuperAdminInContext(ctx, handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"))

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
	if cdPipeline.Environment.IsVirtualEnvironment {
		resourceTreeResp, err := handler.virtualEnvResourceTreeService.GetResourceTreeForPipeline(cdPipeline.Id)
		if err != nil {
			handler.logger.Errorw("error in fetching resource tree", "err", err, "appId", appId, "envId", envId)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		common.WriteJsonResp(w, nil, resourceTreeResp, http.StatusOK)
		return
	}

	envDeploymentConfig, err := handler.deploymentConfigService.GetConfigForDevtronApps(nil, appId, envId)
	if err != nil {
		handler.logger.Errorw("error in fetching deployment config", "appId", appId, "envId", envId, "err", err)
		common.WriteJsonResp(w, fmt.Errorf("error in getting deployment config for env"), nil, http.StatusInternalServerError)
		return
	}
	// flag based drift compute in resource tree
	enableConfigDrift := false
	resourceTree := map[string]any{}
	enableConfigDrift, err = handler.featureFlagService.GetFeatureFlagBoolValueFor(
		featureFlag.NewFeatureFlagHandlerRequest(model.EnableConfigDrift).
			WithAppId(appId).
			WithEnvId(envId))

	if enableConfigDrift {
		resourceTree, err = handler.resourceTreeService.FetchResourceTreeWithDrift(ctx, appId, envId, cdPipeline, envDeploymentConfig)
	} else {
		resourceTree, err = handler.resourceTreeService.FetchResourceTree(ctx, appId, envId, cdPipeline, envDeploymentConfig)
	}

	if err != nil {
		handler.logger.Errorw("error in fetching resource tree", "err", err, "appId", appId, "envId", envId)
		handler.handleResourceTreeErrAndDeletePipelineIfNeeded(w, err, cdPipeline, envDeploymentConfig)
		return
	}

	common.WriteJsonResp(w, err, resourceTree, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) handleResourceTreeErrAndDeletePipelineIfNeeded(w http.ResponseWriter, err error,
	cdPipeline *pipelineConfig.Pipeline, deploymentConfig *commonBean.DeploymentConfig) {
	var apiError *util.ApiError
	ok := errors.As(err, &apiError)
	if deploymentConfig.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD {
		if ok && apiError != nil {
			if apiError.Code == constants.AppDetailResourceTreeNotFound && cdPipeline.DeploymentAppDeleteRequest == true && cdPipeline.DeploymentAppCreated == true {
				acdAppFound, appDeleteErr := handler.pipeline.MarkGitOpsDevtronAppsDeletedWhereArgoAppIsDeleted(cdPipeline)
				if appDeleteErr != nil {
					apiError.UserMessage = constants.ErrorDeletingPipelineForDeletedArgoAppMsg
					common.WriteJsonResp(w, apiError, nil, http.StatusInternalServerError)
					return
				} else if appDeleteErr == nil && !acdAppFound {
					apiError.UserMessage = constants.ArgoAppDeletedErrMsg
					common.WriteJsonResp(w, apiError, nil, http.StatusNotFound)
					return
				}
			}
		}
	}
	// not returned yet therefore no specific error to be handled, send error in internal message
	if ok && apiError != nil {
		apiError.UserMessage = constants.UnableToFetchResourceTreeErrMsg
	} else {
		apiError = &util.ApiError{
			InternalMessage: err.Error(),
			UserMessage:     constants.UnableToFetchResourceTreeErrMsg,
		}
	}
	common.WriteJsonResp(w, apiError, nil, http.StatusInternalServerError)
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
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
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

	// RBAC enforcer applying
	otherEnvironment, err = handler.applyEnvRBACEnforcer(userId, appId, otherEnvironment, token)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC enforcer Ends

	common.WriteJsonResp(w, err, otherEnvironment, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) FetchMinDetailOtherEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
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

	// RBAC enforcer applying
	otherEnvironment, err = handler.applyEnvRBACEnforcer(userId, appId, otherEnvironment, token)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC enforcer Ends
	common.WriteJsonResp(w, err, otherEnvironment, http.StatusOK)
}

func (handler AppListingRestHandlerImpl) applyEnvRBACEnforcer(userId int32, appId int, otherEnvironment []*AppView.Environment, token string) ([]*AppView.Environment, error) {
	canViewPermittedEnv, err := handler.globalFlagService.CanOnlyViewPermittedData(userId)
	if err != nil {
		handler.logger.Errorw("error in fetching global flag", "err", err)
		return nil, err
	}
	if canViewPermittedEnv {
		envIds := make([]int, len(otherEnvironment))
		for _, env := range otherEnvironment {
			envIds = append(envIds, env.EnvironmentId)
		}
		authorizedResourceIds := handler.rbacFilterUtil.FilterAuthorizedResources(envIds, appId, token)
		otherEnvironment = handler.filterEnvironments(otherEnvironment, authorizedResourceIds)
	}
	return otherEnvironment, nil
}

func (handler AppListingRestHandlerImpl) filterEnvironments(envs []*AppView.Environment, envIds []int) []*AppView.Environment {
	envIdSet := make(map[int]struct{}, len(envIds))
	for _, envId := range envIds {
		envIdSet[envId] = struct{}{}
	}

	authorizedEnvs := make([]*AppView.Environment, 0)
	for _, env := range envs {
		if _, exists := envIdSet[env.EnvironmentId]; exists {
			authorizedEnvs = append(authorizedEnvs, env)
		}
	}
	return authorizedEnvs
}

func (handler AppListingRestHandlerImpl) FetchEnvMinData(w http.ResponseWriter, r *http.Request) {
	type AppIdsParam struct {
		AppIds []int `schema:"appId"`
	}

	var schemaDecoder = schema.NewDecoder()
	schemaDecoder.IgnoreUnknownKeys(true)
	queryParams := AppIdsParam{}
	v := r.URL.Query()
	err := schemaDecoder.Decode(&queryParams, v)
	if err != nil {
		handler.logger.Errorw("error in parsing query param", "err", err)
		return
	}

	token := r.Header.Get("token")
	appEnvIds, err := handler.appListingService.FetchEnvMinData(queryParams.AppIds)
	if err != nil {
		handler.logger.Errorw("service err, FetchEnvMinData", "appIds", queryParams.AppIds, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	authorisedAppEnvs := make([]*bean.AppEnvMin, 0)
	appRbacObjs := handler.enforcerUtil.GetRbacObjectsByAppIds(queryParams.AppIds)
	allowedApps := handler.enforcer.EnforceInBatch(token, casbin.ResourceApplications, casbin.ActionGet, maps.Values(appRbacObjs))
	for _, appEnvObj := range appEnvIds {
		if !appEnvObj.IsPipelineDeleted && allowedApps[appRbacObjs[appEnvObj.AppId]] {
			authorisedAppEnvs = append(authorisedAppEnvs, bean.GetAppEnvMin(appEnvObj))
		}
	}
	common.WriteJsonResp(w, nil, authorisedAppEnvs, http.StatusOK)
	return
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
func (handler AppListingRestHandlerImpl) fetchResourceTreeFromInstallAppService(w http.ResponseWriter, r *http.Request, resourceTreeAndNotesContainer AppView.AppDetailsContainer, installedApps repository.InstalledApps, deploymentConfig *commonBean.DeploymentConfig) (AppView.AppDetailsContainer, error) {
	rctx := r.Context()
	cn, _ := w.(http.CloseNotifier)
	err := handler.installedAppResourceService.FetchResourceTree(rctx, cn, &resourceTreeAndNotesContainer, installedApps, deploymentConfig, "", "")
	return resourceTreeAndNotesContainer, err
}
func (handler AppListingRestHandlerImpl) GetHostUrlsByBatch(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	appIdParam := vars.Get("appId")
	installedAppIdParam := vars.Get("installedAppId")
	envIdParam := vars.Get("envId")

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
	var appDetail AppView.AppDetailContainer
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
		installedApp, err := handler.installedAppService.GetInstalledAppById(installedAppId)
		if err == pg.ErrNoRows {
			common.WriteJsonResp(w, err, "App not found in database", http.StatusBadRequest)
			return
		}
		if appStoreUtil.IsExternalChartStoreApp(installedApp.App.DisplayName) {
			//this is external app case where app_name is a unique identifier, and we want to fetch resource based on display_name
			handler.installedAppService.ChangeAppNameToDisplayNameForInstalledApp(installedApp)
		}
		resourceTreeAndNotesContainer := AppView.AppDetailsContainer{}
		resourceTreeAndNotesContainer, err = handler.fetchResourceTreeFromInstallAppService(w, r, resourceTreeAndNotesContainer, *installedApp, appDetail.DeploymentConfig)
		if err != nil {
			common.WriteJsonResp(w, fmt.Errorf("error in fetching resource tree"), nil, http.StatusInternalServerError)
			return
		}
		resourceTree = resourceTreeAndNotesContainer.ResourceTree

	} else {
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
		envDeploymentConfig, err := handler.deploymentConfigService.GetConfigForDevtronApps(nil, appId, envId)
		if err != nil {
			handler.logger.Errorw("error in fetching deployment config", "appId", appId, "envId", envId, "err", err)
			common.WriteJsonResp(w, fmt.Errorf("error in getting deployment config for env"), nil, http.StatusInternalServerError)
			return
		}
		resourceTree, err = handler.resourceTreeService.FetchResourceTree(ctx, appId, envId, cdPipeline, envDeploymentConfig)
	}
	_, ok := resourceTree["nodes"]
	if !ok {
		err = fmt.Errorf("no nodes found for this resource tree appName:%s , envName:%s", appDetail.AppName, appDetail.EnvironmentName)
		handler.logger.Errorw("no nodes found for this resource tree", "appName", appDetail.AppName, "envName", appDetail.EnvironmentName)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// valid batch requests, only valid requests will be sent for batch processing
	validRequests := handler.k8sCommonService.FilterK8sResources(r.Context(), resourceTree, appDetail, "", []string{k8sCommonBean.ServiceKind, k8sCommonBean.IngressKind}, "")
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

func (handler AppListingRestHandlerImpl) getAppDetails(ctx context.Context, appIdParam, installedAppIdParam string, envId int) (AppView.AppDetailContainer, error, int) {
	var appDetail AppView.AppDetailContainer
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
