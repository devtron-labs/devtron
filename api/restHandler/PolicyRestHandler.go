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
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	security2 "github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/security"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type PolicyRestHandler interface {
	SavePolicy(w http.ResponseWriter, r *http.Request)
	UpdatePolicy(w http.ResponseWriter, r *http.Request)
	GetPolicy(w http.ResponseWriter, r *http.Request)
	VerifyImage(w http.ResponseWriter, r *http.Request)
}
type PolicyRestHandlerImpl struct {
	logger             *zap.SugaredLogger
	policyService      security.PolicyService
	userService        user.UserService
	userAuthService    user.UserAuthService
	enforcer           casbin.Enforcer
	enforcerUtil       rbac.EnforcerUtil
	environmentService cluster.EnvironmentService
}

func NewPolicyRestHandlerImpl(logger *zap.SugaredLogger,
	policyService security.PolicyService,
	userService user.UserService, userAuthService user.UserAuthService,
	enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil, environmentService cluster.EnvironmentService) *PolicyRestHandlerImpl {
	return &PolicyRestHandlerImpl{
		logger:             logger,
		policyService:      policyService,
		userService:        userService,
		userAuthService:    userAuthService,
		enforcer:           enforcer,
		enforcerUtil:       enforcerUtil,
		environmentService: environmentService,
	}
}

func (impl PolicyRestHandlerImpl) SavePolicy(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var req bean.CreateVulnerabilityPolicyRequest
	err = decoder.Decode(&req)
	if err != nil {
		impl.logger.Errorw("request err, SavePolicy", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, SavePolicy", "payload", req)
	token := r.Header.Get("token")
	//AUTH - check from casbin db
	if req.AppId > 0 && req.EnvId > 0 {
		object := impl.enforcerUtil.GetAppRBACNameByAppId(req.AppId)
		if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
		object = impl.enforcerUtil.GetEnvRBACNameByAppId(req.AppId, req.EnvId)
		if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionCreate, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	} else if req.AppId == 0 && req.EnvId > 0 {
		// for env level access check env level access.
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionCreate, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	} else {
		// for global and cluster level check super admin access only
		roles, err := impl.userService.CheckUserRoles(userId)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		superAdmin := false
		for _, item := range roles {
			if item == bean.SUPERADMIN {
				superAdmin = true
			}
		}
		if superAdmin == false {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
			return
		}
	}
	//AUTH

	res, err := impl.policyService.SavePolicy(req, userId)
	if err != nil {
		impl.logger.Errorw("service err, SavePolicy", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl PolicyRestHandlerImpl) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var req bean.UpdatePolicyParams

	err = decoder.Decode(&req)
	if err != nil {
		impl.logger.Errorw("request err, UpdatePolicy", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, UpdatePolicy", "err", err, "payload", req)
	policy, err := impl.policyService.GetCvePolicy(req.Id, userId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	token := r.Header.Get("token")
	//AUTH - check from casbin db
	if policy.AppId > 0 && policy.EnvironmentId > 0 {
		object := impl.enforcerUtil.GetAppRBACNameByAppId(policy.AppId)
		if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
		object = impl.enforcerUtil.GetEnvRBACNameByAppId(policy.AppId, policy.EnvironmentId)
		if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	} else if policy.AppId == 0 && policy.EnvironmentId > 0 {
		// for env level access check env level access.
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionUpdate, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	} else {
		// for global and cluster level check super admin access only
		roles, err := impl.userService.CheckUserRoles(userId)
		if err != nil {
			common.WriteJsonResp(w, err, "Failed to get user by id", http.StatusInternalServerError)
			return
		}
		superAdmin := false
		for _, item := range roles {
			if item == bean.SUPERADMIN {
				superAdmin = true
			}
		}
		if superAdmin == false {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
			return
		}
	}
	//AUTH

	res, err := impl.policyService.UpdatePolicy(req, userId)
	if err != nil {
		impl.logger.Errorw("service err, UpdatePolicy", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl PolicyRestHandlerImpl) GetPolicy(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	req := bean.FetchPolicyParams{}
	v := r.URL.Query()
	level := v.Get("level")
	req.Level = bean.ResourceLevel(level)

	id := v.Get("id")
	if len(id) > 0 {
		ids, err := strconv.Atoi(id)
		if err != nil {
			impl.logger.Errorw("request err, GetPolicy", "err", err, "id", id)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		}
		req.Id = ids
	}
	var clusterId, environmentId, appId int
	var policyLevel security2.PolicyLevel
	if level == security2.Global.String() {
		policyLevel = security2.Global
	} else if level == security2.Cluster.String() {
		clusterId = req.Id
		policyLevel = security2.Cluster
	} else if level == security2.Environment.String() {
		environmentId = req.Id
		policyLevel = security2.Environment
	} else if level == security2.Application.String() {
		appId = req.Id
		policyLevel = security2.Application
	}

	token := r.Header.Get("token")
	var vulnerabilityPolicy []*bean.VulnerabilityPolicy
	res, err := impl.policyService.GetPolicies(policyLevel, clusterId, environmentId, appId)
	if err != nil {
		impl.logger.Errorw("service err, GetPolicy", "err", err, "policyLevel", policyLevel, "clusterId", clusterId, "environmentId", environmentId, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	for _, policy := range res.Policies {
		//AUTH - check from casbin db
		pass := true
		if policy.AppId > 0 && policy.EnvId > 0 {
			passCount := 0
			object := impl.enforcerUtil.GetAppRBACNameByAppId(policy.AppId)
			if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); ok {
				passCount = 1
			}
			object = impl.enforcerUtil.GetEnvRBACNameByAppId(policy.AppId, policy.EnvId)
			if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); ok {
				if passCount == 1 {
					passCount = 2
				}
			}
			if passCount == 2 {
				pass = true
			}
		} else if policy.EnvId > 0 {
			// for env level access check env level access.
			environment, err := impl.environmentService.FindById(policy.EnvId)
			if err != nil {
				common.WriteJsonResp(w, err, "Failed to get environment by id", http.StatusInternalServerError)
				return
			}
			if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, environment.Environment); ok {
				pass = true
			}
		} else if clusterId > 0 {
			// for cluster check any of the env access on this cluster
			environments, err := impl.environmentService.GetByClusterId(clusterId)
			if err != nil {
				impl.logger.Errorw("service err, GetPolicy", "err", err, "clusterId", clusterId)
				common.WriteJsonResp(w, err, "Failed to get cluster by id", http.StatusInternalServerError)
				return
			}
			for _, environment := range environments {
				if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, environment.Environment); ok {
					pass = true
					continue
				}
			}
		} else {
			// for global check only logged in user
		}
		//AUTH
		if pass {
			vulnerabilityPolicy = append(vulnerabilityPolicy, policy)
		}
	}
	res.Policies = vulnerabilityPolicy
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

//TODO - move to image-scanner
func (impl PolicyRestHandlerImpl) VerifyImage(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var req security.VerifyImageRequest

	err := decoder.Decode(&req)
	if err != nil {
		impl.logger.Errorw("request err, VerifyImage", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, VerifyImage", "req", req)
	res, err := impl.policyService.VerifyImage(&req)
	if err != nil {
		impl.logger.Errorw("request err, VerifyImage", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
