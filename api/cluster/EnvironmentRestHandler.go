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

package cluster

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/api/bean"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	request "github.com/devtron-labs/devtron/pkg/cluster"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const ENV_DELETE_SUCCESS_RESP = "Environment deleted successfully."

type EnvironmentRestHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetAllActive(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	FindById(w http.ResponseWriter, r *http.Request)
	GetEnvironmentListForAutocomplete(w http.ResponseWriter, r *http.Request)
	GetCombinedEnvironmentListForDropDown(w http.ResponseWriter, r *http.Request)
	DeleteEnvironment(w http.ResponseWriter, r *http.Request)
	GetCombinedEnvironmentListForDropDownByClusterIds(w http.ResponseWriter, r *http.Request)
}

type EnvironmentRestHandlerImpl struct {
	environmentClusterMappingsService request.EnvironmentService
	logger                            *zap.SugaredLogger
	userService                       user.UserService
	validator                         *validator.Validate
	enforcer                          casbin.Enforcer
	deleteService                     delete2.DeleteService
	cfg                               *bean.Config
}

func NewEnvironmentRestHandlerImpl(svc request.EnvironmentService, logger *zap.SugaredLogger, userService user.UserService,
	validator *validator.Validate, enforcer casbin.Enforcer,
	deleteService delete2.DeleteService,
) *EnvironmentRestHandlerImpl {
	cfg := &bean.Config{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Errorw("error occurred while parsing config ", "err", err)
		cfg.IgnoreAuthCheck = false
	}
	logger.Infow("evironment rest handler initialized", "ignoreAuthCheckValue", cfg.IgnoreAuthCheck)
	return &EnvironmentRestHandlerImpl{
		environmentClusterMappingsService: svc,
		logger:                            logger,
		userService:                       userService,
		validator:                         validator,
		enforcer:                          enforcer,
		deleteService:                     deleteService,
		cfg:                               cfg,
	}
}

func (handler *EnvironmentRestHandlerImpl) Create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean request.EnvironmentBean
	err = decoder.Decode(&bean)
	if err != nil {
		handler.logger.Errorw("request err, Create", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Errorw("request payload, Create", "payload", bean)

	err = handler.validator.Struct(bean)
	if err != nil {
		handler.logger.Errorw("validation err, Create", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := handler.environmentClusterMappingsService.Create(&bean, userId)
	if err != nil {
		handler.logger.Errorw("service err, Create", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *EnvironmentRestHandlerImpl) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	environment := vars["environment"]

	bean, err := handler.environmentClusterMappingsService.FindOne(environment)
	if err != nil {
		handler.logger.Errorw("service err, Get", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, strings.ToLower(bean.EnvironmentIdentifier)); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, bean, http.StatusOK)
}

func (handler *EnvironmentRestHandlerImpl) GetAll(w http.ResponseWriter, r *http.Request) {
	bean, err := handler.environmentClusterMappingsService.GetAll()
	if err != nil {
		handler.logger.Errorw("service err, GetAll", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	var result []request.EnvironmentBean
	token := r.Header.Get("token")
	for _, item := range bean {
		// RBAC enforcer applying
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, strings.ToLower(item.EnvironmentIdentifier)); ok {
			result = append(result, item)
		}
		//RBAC enforcer Ends
	}

	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (handler *EnvironmentRestHandlerImpl) GetAllActive(w http.ResponseWriter, r *http.Request) {
	bean, err := handler.environmentClusterMappingsService.GetAllActive()
	if err != nil {
		handler.logger.Errorw("service err, GetAllActive", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	var result []request.EnvironmentBean
	token := r.Header.Get("token")
	for _, item := range bean {
		// RBAC enforcer applying
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, strings.ToLower(item.EnvironmentIdentifier)); ok {
			result = append(result, item)
		}
		//RBAC enforcer Ends
	}

	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (handler *EnvironmentRestHandlerImpl) Update(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	var bean request.EnvironmentBean
	err = decoder.Decode(&bean)
	if err != nil {
		handler.logger.Errorw("service err, Update", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, Update", "payload", bean)
	err = handler.validator.Struct(bean)
	if err != nil {
		handler.logger.Errorw("validation err, Update", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	modifiedEnvironment, err := handler.environmentClusterMappingsService.FindById(bean.Id)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionUpdate, strings.ToLower(modifiedEnvironment.EnvironmentIdentifier)); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := handler.environmentClusterMappingsService.Update(&bean, userId)
	if err != nil {
		handler.logger.Errorw("service err, Update", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *EnvironmentRestHandlerImpl) FindById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	envId, err := strconv.Atoi(vars["id"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean, err := handler.environmentClusterMappingsService.FindById(envId)
	if err != nil {
		handler.logger.Errorw("service err, FindById", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, strings.ToLower(bean.EnvironmentIdentifier)); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, bean, http.StatusOK)
}

func (handler *EnvironmentRestHandlerImpl) GetEnvironmentListForAutocomplete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	start := time.Now()
	environments, err := handler.environmentClusterMappingsService.GetEnvironmentListForAutocomplete()
	if err != nil {
		handler.logger.Errorw("service err, GetEnvironmentListForAutocomplete", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	dbElapsedTime := time.Since(start)

	token := r.Header.Get("token")
	var grantedEnvironment = environments
	start = time.Now()
	if !handler.cfg.IgnoreAuthCheck {
		grantedEnvironment = make([]request.EnvironmentBean, 0)
		emailId, _ := handler.userService.GetEmailFromToken(token)
		// RBAC enforcer applying
		var envIdentifierList []string
		for _, item := range environments {
			envIdentifierList = append(envIdentifierList, strings.ToLower(item.EnvironmentIdentifier))
		}

		result := handler.enforcer.EnforceByEmailInBatch(emailId, casbin.ResourceGlobalEnvironment, casbin.ActionGet, envIdentifierList)
		for _, item := range environments {
			if hasAccess := result[strings.ToLower(item.EnvironmentIdentifier)]; hasAccess {
				grantedEnvironment = append(grantedEnvironment, item)
			}
		}
		//RBAC enforcer Ends
	}
	elapsedTime := time.Since(start)
	handler.logger.Infow("Env elapsed Time for enforcer", "dbElapsedTime", dbElapsedTime, "elapsedTime",
		elapsedTime, "token", token, "envSize", len(grantedEnvironment))

	common.WriteJsonResp(w, err, grantedEnvironment, http.StatusOK)
}

func (handler *EnvironmentRestHandlerImpl) GetCombinedEnvironmentListForDropDown(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check admin check", http.StatusInternalServerError)
		return
	}
	token := r.Header.Get("token")
	clusters, err := handler.environmentClusterMappingsService.GetCombinedEnvironmentListForDropDown(token, isActionUserSuperAdmin, handler.CheckAuthorizationForGlobalEnvironment)
	if err != nil {
		handler.logger.Errorw("service err, GetCombinedEnvironmentListForDropDown", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if len(clusters) == 0 {
		clusters = make([]*request.ClusterEnvDto, 0)
	}
	common.WriteJsonResp(w, err, clusters, http.StatusOK)
}

func (handler *EnvironmentRestHandlerImpl) CheckAuthorizationForGlobalEnvironment(token string, object string) bool {
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, strings.ToLower(object)); !ok {
		return false
	}
	return true
}

func (handler *EnvironmentRestHandlerImpl) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean request.EnvironmentBean
	err = decoder.Decode(&bean)
	if err != nil {
		handler.logger.Errorw("request err, Delete", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, Delete", "payload", bean)

	err = handler.validator.Struct(bean)
	if err != nil {
		handler.logger.Errorw("validation err, Delete", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobalEnvironment, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	err = handler.deleteService.DeleteEnvironment(&bean, userId)
	if err != nil {
		handler.logger.Errorw("service err, Delete", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, ENV_DELETE_SUCCESS_RESP, http.StatusOK)
}

func (handler *EnvironmentRestHandlerImpl) GetCombinedEnvironmentListForDropDownByClusterIds(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	v := r.URL.Query()
	clusterIdString := v.Get("ids")
	var clusterIds []int
	if clusterIdString != "" {
		clusterIdSlices := strings.Split(clusterIdString, ",")
		for _, clusterId := range clusterIdSlices {
			id, err := strconv.Atoi(clusterId)
			if err != nil {
				handler.logger.Errorw("request err, GetCombinedEnvironmentListForDropDownByClusterIds", "err", err, "clusterIdString", clusterIdString)
				common.WriteJsonResp(w, err, "please send valid cluster Ids", http.StatusBadRequest)
				return
			}
			clusterIds = append(clusterIds, id)
		}
	}
	token := r.Header.Get("token")
	clusters, err := handler.environmentClusterMappingsService.GetCombinedEnvironmentListForDropDownByClusterIds(token, clusterIds, handler.CheckAuthorizationForGlobalEnvironment)
	if err != nil {
		handler.logger.Errorw("service err, GetCombinedEnvironmentListForDropDown", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	if len(clusters) == 0 {
		clusters = make([]*request.ClusterEnvDto, 0)
	}
	common.WriteJsonResp(w, err, clusters, http.StatusOK)
}
