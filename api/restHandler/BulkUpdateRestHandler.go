/*
 * Copyright (c) 2024. Devtron Inc.
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

package restHandler

import (
	"encoding/json"
	"fmt"
	bean4 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/repository"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider"
	"github.com/devtron-labs/devtron/pkg/bulkAction/bean"
	"github.com/devtron-labs/devtron/pkg/bulkAction/service"
	"github.com/devtron-labs/devtron/pkg/cluster/environment"
	"github.com/devtron-labs/devtron/util"
	"net/http"
	"strconv"
	"strings"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/appClone"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type BulkUpdateRestHandler interface {
	FindBulkUpdateReadme(w http.ResponseWriter, r *http.Request)
	GetImpactedAppsName(w http.ResponseWriter, r *http.Request)
	BulkUpdate(w http.ResponseWriter, r *http.Request)

	BulkHibernate(w http.ResponseWriter, r *http.Request)
	BulkUnHibernate(w http.ResponseWriter, r *http.Request)
	BulkDeploy(w http.ResponseWriter, r *http.Request)
	BulkBuildTrigger(w http.ResponseWriter, r *http.Request)

	HandleCdPipelineBulkAction(w http.ResponseWriter, r *http.Request)
}
type BulkUpdateRestHandlerImpl struct {
	pipelineBuilder         pipeline.PipelineBuilder
	ciPipelineRepository    pipelineConfig.CiPipelineRepository
	ciHandler               pipeline.CiHandler
	logger                  *zap.SugaredLogger
	bulkUpdateService       service.BulkUpdateService
	chartService            chart.ChartService
	propertiesConfigService pipeline.PropertiesConfigService
	userAuthService         user.UserService
	validator               *validator.Validate
	teamService             team.TeamService
	enforcer                casbin.Enforcer
	gitSensorClient         gitSensor.Client
	pipelineRepository      pipelineConfig.PipelineRepository
	appWorkflowService      appWorkflow.AppWorkflowService
	enforcerUtil            rbac.EnforcerUtil
	envService              environment.EnvironmentService
	gitRegistryConfig       gitProvider.GitRegistryConfig
	dockerRegistryConfig    pipeline.DockerRegistryConfig
	cdHandelr               pipeline.CdHandler
	appCloneService         appClone.AppCloneService
	materialRepository      repository.MaterialRepository
}

func NewBulkUpdateRestHandlerImpl(pipelineBuilder pipeline.PipelineBuilder, logger *zap.SugaredLogger,
	bulkUpdateService service.BulkUpdateService,
	chartService chart.ChartService,
	propertiesConfigService pipeline.PropertiesConfigService,
	userAuthService user.UserService,
	teamService team.TeamService,
	enforcer casbin.Enforcer,
	ciHandler pipeline.CiHandler,
	validator *validator.Validate,
	gitSensorClient gitSensor.Client,
	ciPipelineRepository pipelineConfig.CiPipelineRepository, pipelineRepository pipelineConfig.PipelineRepository,
	enforcerUtil rbac.EnforcerUtil, envService environment.EnvironmentService,
	gitRegistryConfig gitProvider.GitRegistryConfig, dockerRegistryConfig pipeline.DockerRegistryConfig,
	cdHandelr pipeline.CdHandler,
	appCloneService appClone.AppCloneService,
	appWorkflowService appWorkflow.AppWorkflowService,
	materialRepository repository.MaterialRepository,
) *BulkUpdateRestHandlerImpl {
	return &BulkUpdateRestHandlerImpl{
		pipelineBuilder:         pipelineBuilder,
		logger:                  logger,
		bulkUpdateService:       bulkUpdateService,
		chartService:            chartService,
		propertiesConfigService: propertiesConfigService,
		userAuthService:         userAuthService,
		validator:               validator,
		teamService:             teamService,
		enforcer:                enforcer,
		ciHandler:               ciHandler,
		gitSensorClient:         gitSensorClient,
		ciPipelineRepository:    ciPipelineRepository,
		pipelineRepository:      pipelineRepository,
		enforcerUtil:            enforcerUtil,
		envService:              envService,
		gitRegistryConfig:       gitRegistryConfig,
		dockerRegistryConfig:    dockerRegistryConfig,
		cdHandelr:               cdHandelr,
		appCloneService:         appCloneService,
		appWorkflowService:      appWorkflowService,
		materialRepository:      materialRepository,
	}
}

func (handler BulkUpdateRestHandlerImpl) FindBulkUpdateReadme(w http.ResponseWriter, r *http.Request) {
	var operation string
	vars := mux.Vars(r)
	apiVersion := vars["apiVersion"]
	kind := vars["kind"]
	operation = fmt.Sprintf("%s/%s", apiVersion, kind)
	response, err := handler.bulkUpdateService.FindBulkUpdateReadme(operation)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//auth free, only login required
	var responseArr []*bean.BulkUpdateSeeExampleResponse
	responseArr = append(responseArr, response)
	common.WriteJsonResp(w, nil, responseArr, http.StatusOK)
}

func (handler BulkUpdateRestHandlerImpl) CheckAuthForImpactedObjects(AppId int, EnvId int, appResourceObjects map[int]string, envResourceObjects map[string]string, token string) bool {
	resourceName := appResourceObjects[AppId]
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		return ok
	}
	if EnvId > 0 {
		key := fmt.Sprintf("%d-%d", EnvId, AppId)
		envResourceName := envResourceObjects[key]
		if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, envResourceName); !ok {
			return ok
		}
	}
	return true

}
func (handler BulkUpdateRestHandlerImpl) GetImpactedAppsName(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var script bean.BulkUpdateScript
	err := decoder.Decode(&script)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(script)
	if err != nil {
		handler.logger.Errorw("validation err, Script", "err", err, "BulkUpdateScript", script)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	impactedApps, err := handler.bulkUpdateService.GetBulkAppName(script.Spec)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	appResourceObjects, envResourceObjects := handler.enforcerUtil.GetRbacObjectsForAllAppsAndEnvironments()
	for _, deploymentTemplateImpactedApp := range impactedApps.DeploymentTemplate {
		ok := handler.CheckAuthForImpactedObjects(deploymentTemplateImpactedApp.AppId, deploymentTemplateImpactedApp.EnvId, appResourceObjects, envResourceObjects, token)
		if !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	for _, configMapImpactedApp := range impactedApps.ConfigMap {
		ok := handler.CheckAuthForImpactedObjects(configMapImpactedApp.AppId, configMapImpactedApp.EnvId, appResourceObjects, envResourceObjects, token)
		if !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	for _, secretImpactedApp := range impactedApps.Secret {
		ok := handler.CheckAuthForImpactedObjects(secretImpactedApp.AppId, secretImpactedApp.EnvId, appResourceObjects, envResourceObjects, token)
		if !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	common.WriteJsonResp(w, err, impactedApps, http.StatusOK)
}

func (handler BulkUpdateRestHandlerImpl) CheckAuthForBulkUpdate(AppId int, EnvId int, AppName string, rbacObjects map[int]string, token string) bool {
	resourceName := rbacObjects[AppId]
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, resourceName); !ok {
		return ok
	}
	if EnvId > 0 {
		resourceName := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(AppName, EnvId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, resourceName); !ok {
			return ok
		}
	}
	return true

}
func (handler BulkUpdateRestHandlerImpl) BulkUpdate(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var script bean.BulkUpdateScript
	err = decoder.Decode(&script)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(script)
	if err != nil {
		handler.logger.Errorw("validation err, Script", "err", err, "BulkUpdateScript", script)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	impactedApps, err := handler.bulkUpdateService.GetBulkAppName(script.Spec)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	rbacObjects := handler.enforcerUtil.GetRbacObjectsForAllApps(helper.CustomApp)
	for _, deploymentTemplateImpactedApp := range impactedApps.DeploymentTemplate {
		ok := handler.CheckAuthForBulkUpdate(deploymentTemplateImpactedApp.AppId, deploymentTemplateImpactedApp.EnvId, deploymentTemplateImpactedApp.AppName, rbacObjects, token)
		if !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	for _, configMapImpactedApp := range impactedApps.ConfigMap {
		ok := handler.CheckAuthForBulkUpdate(configMapImpactedApp.AppId, configMapImpactedApp.EnvId, configMapImpactedApp.AppName, rbacObjects, token)
		if !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	for _, secretImpactedApp := range impactedApps.Secret {
		ok := handler.CheckAuthForBulkUpdate(secretImpactedApp.AppId, secretImpactedApp.EnvId, secretImpactedApp.AppName, rbacObjects, token)
		if !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	userEmail := util.GetEmailFromContext(r.Context())
	userMetadata := &bean4.UserMetadata{
		UserEmailId:      userEmail,
		IsUserSuperAdmin: isSuperAdmin,
		UserId:           userId,
	}
	response := handler.bulkUpdateService.BulkUpdate(script.Spec, userMetadata)
	common.WriteJsonResp(w, nil, response, http.StatusOK)
}

func (handler BulkUpdateRestHandlerImpl) BulkHibernate(w http.ResponseWriter, r *http.Request) {
	request, err := handler.decodeAndValidateBulkRequest(w, r)
	if err != nil {
		return // response already written by the helper on error.
	}
	token := r.Header.Get("token")
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	userEmail := util.GetEmailFromContext(r.Context())
	userMetadata := &bean4.UserMetadata{
		UserEmailId:      userEmail,
		IsUserSuperAdmin: isSuperAdmin,
		UserId:           request.UserId,
	}

	response, err := handler.bulkUpdateService.BulkHibernate(r.Context(), request, handler.checkAuthForBulkHibernateAndUnhibernate, userMetadata)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, response, http.StatusOK)
}

// decodeAndValidateBulkRequest is a helper to decode and validate the request.
func (handler BulkUpdateRestHandlerImpl) decodeAndValidateBulkRequest(w http.ResponseWriter, r *http.Request) (*bean.BulkApplicationForEnvironmentPayload, error) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return nil, err
	}

	decoder := json.NewDecoder(r.Body)
	var request bean.BulkApplicationForEnvironmentPayload
	if err = decoder.Decode(&request); err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, err
	}
	request.UserId = userId
	if err = handler.validator.Struct(request); err != nil {
		handler.logger.Errorw("validation error", "request", request, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, err
	}
	return &request, nil
}

func (handler BulkUpdateRestHandlerImpl) BulkUnHibernate(w http.ResponseWriter, r *http.Request) {
	request, err := handler.decodeAndValidateBulkRequest(w, r)
	if err != nil {
		return // response already written by the helper on error.
	}
	token := r.Header.Get("token")
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	userEmail := util.GetEmailFromContext(r.Context())
	userMetadata := &bean4.UserMetadata{
		UserEmailId:      userEmail,
		IsUserSuperAdmin: isSuperAdmin,
		UserId:           request.UserId,
	}
	response, err := handler.bulkUpdateService.BulkUnHibernate(r.Context(), request, handler.checkAuthForBulkHibernateAndUnhibernate, userMetadata)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, response, http.StatusOK)
}
func (handler BulkUpdateRestHandlerImpl) BulkDeploy(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request bean.BulkApplicationForEnvironmentPayload
	err = decoder.Decode(&request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*")
	userEmail := util.GetEmailFromContext(r.Context())
	userMetadata := &bean4.UserMetadata{
		UserEmailId:      userEmail,
		IsUserSuperAdmin: isSuperAdmin,
		UserId:           userId,
	}
	response, err := handler.bulkUpdateService.BulkDeploy(&request, token, handler.checkAuthBatch, userMetadata)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, response, http.StatusOK)
}

func (handler BulkUpdateRestHandlerImpl) BulkBuildTrigger(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request bean.BulkApplicationForEnvironmentPayload
	err = decoder.Decode(&request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	response, err := handler.bulkUpdateService.BulkBuildTrigger(&request, r.Context(), w, token, handler.checkAuthForBulkActions)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, response, http.StatusOK)
}

func (handler BulkUpdateRestHandlerImpl) checkAuthForBulkActions(token string, appObject string, envObject string) bool {
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, appObject); !ok {
		return false
	}
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, envObject); !ok {
		return false
	}
	return true
}

func (handler BulkUpdateRestHandlerImpl) checkAuthForBulkHibernateAndUnhibernate(token string, appObject string, envObject string) bool {
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, strings.ToLower(appObject)); !ok {
		return false
	}
	if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, strings.ToLower(envObject)); !ok {
		return false
	}
	return true
}

func (handler BulkUpdateRestHandlerImpl) HandleCdPipelineBulkAction(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var cdPipelineBulkActionReq bean.CdBulkActionRequestDto
	err = decoder.Decode(&cdPipelineBulkActionReq)
	cdPipelineBulkActionReq.UserId = userId
	if err != nil {
		handler.logger.Errorw("request err, HandleCdPipelineBulkAction", "err", err, "payload", cdPipelineBulkActionReq)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	if cdPipelineBulkActionReq.ForceDelete {
		cdPipelineBulkActionReq.NonCascadeDelete = true
	}

	v := r.URL.Query()
	dryRun := false
	dryRunParam := v.Get("dryRun")
	if len(dryRunParam) > 0 {
		dryRun, err = strconv.ParseBool(dryRunParam)
		if err != nil {
			handler.logger.Errorw("request err, HandleCdPipelineBulkAction", "err", err, "payload", cdPipelineBulkActionReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	handler.logger.Infow("request payload, HandleCdPipelineBulkAction", "payload", cdPipelineBulkActionReq)
	impactedPipelines, impactedAppWfIds, impactedCiPipelineIds, err := handler.bulkUpdateService.GetBulkActionImpactedPipelinesAndWfs(&cdPipelineBulkActionReq)
	if err != nil {
		handler.logger.Errorw("service err, GetBulkActionImpactedPipelinesAndWfs", "err", err, "payload", cdPipelineBulkActionReq)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	token := r.Header.Get("token")

	appsHavingRbacChecked := make(map[string]bool)
	for _, impactedPipeline := range impactedPipelines {
		//check to avoid same rbac matching multiple times
		if _, ok := appsHavingRbacChecked[impactedPipeline.App.AppName]; !ok {
			resourceName := handler.enforcerUtil.GetAppRBACName(impactedPipeline.App.AppName)
			if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, resourceName); !ok {
				common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
				return
			} else {
				appsHavingRbacChecked[impactedPipeline.App.AppName] = true
			}
		}
		object := handler.enforcerUtil.GetAppRBACByAppNameAndEnvId(impactedPipeline.App.AppName, impactedPipeline.EnvironmentId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	}

	resp, err := handler.bulkUpdateService.PerformBulkActionOnCdPipelines(&cdPipelineBulkActionReq, impactedPipelines, r.Context(), dryRun, impactedAppWfIds, impactedCiPipelineIds)
	if err != nil {
		handler.logger.Errorw("service err, HandleCdPipelineBulkAction", "err", err, "payload", cdPipelineBulkActionReq)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler BulkUpdateRestHandlerImpl) checkAuthBatch(token string, appObject []string, envObject []string) (map[string]bool, map[string]bool) {
	var appResult map[string]bool
	var envResult map[string]bool
	if len(appObject) > 0 {
		appResult = handler.enforcer.EnforceInBatch(token, casbin.ResourceApplications, casbin.ActionGet, appObject)
	}
	if len(envObject) > 0 {
		envResult = handler.enforcer.EnforceInBatch(token, casbin.ResourceEnvironment, casbin.ActionGet, envObject)
	}
	return appResult, envResult
}
