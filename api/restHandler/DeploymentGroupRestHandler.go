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
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/deploymentGroup"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
)

type DeploymentGroupRestHandler interface {
	CreateDeploymentGroup(w http.ResponseWriter, r *http.Request)
	FetchParentCiForDG(w http.ResponseWriter, r *http.Request)
	FetchEnvApplicationsForDG(w http.ResponseWriter, r *http.Request)
	FetchAllDeploymentGroups(w http.ResponseWriter, r *http.Request)
	DeleteDeploymentGroup(w http.ResponseWriter, r *http.Request)
	TriggerReleaseForDeploymentGroup(w http.ResponseWriter, r *http.Request)
	UpdateDeploymentGroup(w http.ResponseWriter, r *http.Request)
	GetArtifactsByCiPipeline(w http.ResponseWriter, r *http.Request)
	GetDeploymentGroupById(w http.ResponseWriter, r *http.Request)
}

type DeploymentGroupRestHandlerImpl struct {
	deploymentGroupService deploymentGroup.DeploymentGroupService
	logger                 *zap.SugaredLogger
	validator              *validator.Validate
	enforcer               casbin.Enforcer
	teamService            team.TeamService
	userAuthService        user.UserService
	enforcerUtil           rbac.EnforcerUtil
}

func NewDeploymentGroupRestHandlerImpl(deploymentGroupService deploymentGroup.DeploymentGroupService, logger *zap.SugaredLogger,
	validator *validator.Validate, enforcer casbin.Enforcer, teamService team.TeamService, userAuthService user.UserService, enforcerUtil rbac.EnforcerUtil) *DeploymentGroupRestHandlerImpl {
	return &DeploymentGroupRestHandlerImpl{deploymentGroupService: deploymentGroupService, logger: logger, validator: validator,
		enforcer: enforcer, teamService: teamService, userAuthService: userAuthService, enforcerUtil: enforcerUtil}
}

func (handler *DeploymentGroupRestHandlerImpl) CreateDeploymentGroup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean deploymentGroup.DeploymentGroupRequest

	err = decoder.Decode(&bean)
	if err != nil {
		handler.logger.Errorw("request err, CreateDeploymentGroup", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	handler.logger.Errorw("request payload, CreateDeploymentGroup", "payload", bean)

	err = handler.validator.Struct(bean)
	if err != nil {
		handler.logger.Errorw("validation err, CreateDeploymentGroup", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	for _, item := range bean.AppIds {
		resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(item)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	// RBAC enforcer Ends

	res, err := handler.deploymentGroupService.CreateDeploymentGroup(&bean)
	if err != nil {
		handler.logger.Errorw("service err, CreateDeploymentGroup", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *DeploymentGroupRestHandlerImpl) FetchParentCiForDG(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	vars := mux.Vars(r)
	deploymentGroupId, err := strconv.Atoi(vars["deploymentGroupId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchParentCiForDG", "err", err, "deploymentGroupId", deploymentGroupId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	resp, err := handler.deploymentGroupService.FetchParentCiForDG(deploymentGroupId)
	if err != nil {
		handler.logger.Errorw("service err, FetchParentCiForDG", "err", err, "deploymentGroupId", deploymentGroupId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC filter CI List
	finalResp := make([]*deploymentGroup.CiPipelineResponseForDG, 0)
	for _, item := range resp {
		resourceName := handler.enforcerUtil.GetTeamRBACByCiPipelineId(item.CiPipelineId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); ok {
			finalResp = append(finalResp, item)
		}
	}
	// RBAC filter CI List Ends

	common.WriteJsonResp(w, err, finalResp, http.StatusOK)
}

func (handler *DeploymentGroupRestHandlerImpl) FetchEnvApplicationsForDG(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	ciPipelineId, err := strconv.Atoi(vars["ciPipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchEnvApplicationsForDG", "err", err, "ciPipelineId", ciPipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	resourceName := handler.enforcerUtil.GetTeamRBACByCiPipelineId(ciPipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	result, err := handler.deploymentGroupService.FetchEnvApplicationsForDG(ciPipelineId)
	if err != nil {
		handler.logger.Errorw("service err, FetchEnvApplicationsForDG", "err", err, "ciPipelineId", ciPipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	finalResp := make([]*deploymentGroup.EnvironmentAppListForDG, 0)
	for _, item := range result {
		// RBAC enforcer applying
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, strings.ToLower(item.EnvironmentIdentifier)); ok {
			passCount := 0
			for _, app := range item.Apps {
				resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(app.Id)
				if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); ok {
					passCount = passCount + 1
				}
			}
			if len(item.Apps) == passCount {
				finalResp = append(finalResp, item)
			}
		}
		//RBAC enforcer Ends
	}

	common.WriteJsonResp(w, err, finalResp, http.StatusOK)

}

func (handler *DeploymentGroupRestHandlerImpl) FetchAllDeploymentGroups(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	resp, err := handler.deploymentGroupService.FetchAllDeploymentGroups()
	if err != nil {
		handler.logger.Errorw("request err, FetchAllDeploymentGroups", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC filter CI List
	finalResp := make([]deploymentGroup.DeploymentGroupDTO, 0)
	for _, item := range resp {
		pass := 0
		resourceName := handler.enforcerUtil.GetTeamRBACByCiPipelineId(item.CiPipelineId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); ok {
			pass = 1
		}
		resourceName = handler.enforcerUtil.GetEnvRBACNameByCiPipelineIdAndEnvId(item.CiPipelineId, item.EnvironmentId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, resourceName); ok {
			pass = 2
		}
		if pass == 2 {
			finalResp = append(finalResp, item)
		}
	}
	// RBAC filter CI List Ends

	common.WriteJsonResp(w, err, finalResp, http.StatusOK)
}

func (handler *DeploymentGroupRestHandlerImpl) DeleteDeploymentGroup(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	token := r.Header.Get("token")
	deploymentGroupId, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("service err, DeleteDeploymentGroup", "err", err, "deploymentGroupId", deploymentGroupId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, DeleteDeploymentGroup", "deploymentGroupId", deploymentGroupId)

	dg, err := handler.deploymentGroupService.FindById(deploymentGroupId)
	if err != nil {
		handler.logger.Errorw("service err, DeleteDeploymentGroup", "err", err, "deploymentGroupId", deploymentGroupId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	resourceName := handler.enforcerUtil.GetTeamRBACByCiPipelineId(dg.CiPipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionDelete, resourceName); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	// RBAC enforcer Ends

	res, err := handler.deploymentGroupService.DeleteDeploymentGroup(deploymentGroupId)
	if err != nil {
		handler.logger.Errorw("service err, DeleteDeploymentGroup", "err", err, "deploymentGroupId", deploymentGroupId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *DeploymentGroupRestHandlerImpl) TriggerReleaseForDeploymentGroup(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean *deploymentGroup.DeploymentGroupTriggerRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&bean)
	if err != nil {
		handler.logger.Errorw("request err, TriggerReleaseForDeploymentGroup", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	handler.logger.Errorw("request payload, TriggerReleaseForDeploymentGroup", "payload", bean)

	dg, err := handler.deploymentGroupService.GetDeploymentGroupById(bean.DeploymentGroupId)
	if err != nil {
		handler.logger.Errorw("service err, TriggerReleaseForDeploymentGroup", "err", err, "payload", bean)
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
	// RBAC enforcer Ends

	res, err := handler.deploymentGroupService.TriggerReleaseForDeploymentGroup(bean)
	if err != nil {
		handler.logger.Errorw("service err, TriggerReleaseForDeploymentGroup", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *DeploymentGroupRestHandlerImpl) UpdateDeploymentGroup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean deploymentGroup.DeploymentGroupRequest
	err = decoder.Decode(&bean)
	if err != nil {
		handler.logger.Errorw("request err, UpdateDeploymentGroup", "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	handler.logger.Infow("request payload, UpdateDeploymentGroup", "payload", bean)
	err = handler.validator.Struct(bean)
	if err != nil {
		handler.logger.Errorw("validation err, UpdateDeploymentGroup", "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	resourceName := handler.enforcerUtil.GetTeamRBACByCiPipelineId(bean.CiPipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, resourceName); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	// RBAC enforcer Ends

	res, err := handler.deploymentGroupService.UpdateDeploymentGroup(&bean)
	if err != nil {
		handler.logger.Errorw("service err, UpdateDeploymentGroup", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *DeploymentGroupRestHandlerImpl) GetArtifactsByCiPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deploymentGroupId, err := strconv.Atoi(vars["deploymentGroupId"])
	if err != nil {
		handler.logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	dg, err := handler.deploymentGroupService.FindById(deploymentGroupId)
	if err != nil {
		handler.logger.Errorw("request err, GetArtifactsByCiPipeline", "deploymentGroupId", deploymentGroupId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	resourceName := handler.enforcerUtil.GetTeamRBACByCiPipelineId(dg.CiPipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	ciArtifactResponse, err := handler.deploymentGroupService.GetArtifactsByCiPipeline(dg.CiPipelineId)
	if err != nil {
		handler.logger.Errorw("service err, GetArtifactsByCiPipeline", "err", err, "deploymentGroupId", deploymentGroupId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, ciArtifactResponse, http.StatusOK)
}

func (handler *DeploymentGroupRestHandlerImpl) GetDeploymentGroupById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deploymentGroupId, err := strconv.Atoi(vars["deploymentGroupId"])
	if err != nil {
		handler.logger.Errorw("request err, GetDeploymentGroupById", "deploymentGroupId", deploymentGroupId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	dg, err := handler.deploymentGroupService.FindById(deploymentGroupId)
	if err != nil {
		handler.logger.Errorw("service err, GetDeploymentGroupById", "err", err, "deploymentGroupId", deploymentGroupId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	resourceName := handler.enforcerUtil.GetTeamRBACByCiPipelineId(dg.CiPipelineId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	deploymentGroup, err := handler.deploymentGroupService.GetDeploymentGroupById(deploymentGroupId)
	if err != nil {
		handler.logger.Errorw("service err, GetDeploymentGroupById", "err", err, "deploymentGroupId", deploymentGroupId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, deploymentGroup, http.StatusOK)
}
