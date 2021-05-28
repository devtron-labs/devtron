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
	request "github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type EnvironmentRestHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetAllActive(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	FindById(w http.ResponseWriter, r *http.Request)
	GetEnvironmentListForAutocomplete(w http.ResponseWriter, r *http.Request)
}

type EnvironmentRestHandlerImpl struct {
	environmentClusterMappingsService request.EnvironmentService
	logger                            *zap.SugaredLogger
	userService                       user.UserService
	validator                         *validator.Validate
	enforcer                          rbac.Enforcer
	enforcerUtil                      rbac.EnforcerUtil
	userAuthService                   user.UserAuthService
}

func NewEnvironmentRestHandlerImpl(svc request.EnvironmentService, logger *zap.SugaredLogger, userService user.UserService,
	validator *validator.Validate, enforcer rbac.Enforcer, enforcerUtil rbac.EnforcerUtil, userAuthService user.UserAuthService) *EnvironmentRestHandlerImpl {
	return &EnvironmentRestHandlerImpl{
		environmentClusterMappingsService: svc,
		logger:                            logger,
		userService:                       userService,
		validator:                         validator,
		enforcer:                          enforcer,
		enforcerUtil:                      enforcerUtil,
		userAuthService:                   userAuthService,
	}
}

func (impl EnvironmentRestHandlerImpl) validateNamespace(namespace string) bool {
	hostnameRegexString := `^$|^[a-z]+[a-z0-9\-\?]*[a-z0-9]+$`
	hostnameRegexRFC952 := regexp.MustCompile(hostnameRegexString)
	return hostnameRegexRFC952.MatchString(namespace)
}

func (impl EnvironmentRestHandlerImpl) Create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean request.EnvironmentBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, Create", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Errorw("request payload, Create", "payload", bean)

	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, Create", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !impl.validateNamespace(bean.Namespace) {
		impl.logger.Errorw("validation err, Create", "err", err, "namespace", bean.Namespace)
		writeJsonResp(w, errors.New("invalid ns"), nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobalEnvironment, rbac.ActionCreate, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := impl.environmentClusterMappingsService.Create(&bean, userId)
	if err != nil {
		impl.logger.Errorw("service err, Create", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl EnvironmentRestHandlerImpl) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	environment := vars["environment"]

	bean, err := impl.environmentClusterMappingsService.FindOne(environment)
	if err != nil {
		impl.logger.Errorw("service err, Get", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobalEnvironment, rbac.ActionGet, strings.ToLower(bean.Environment)); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	writeJsonResp(w, err, bean, http.StatusOK)
}

func (impl EnvironmentRestHandlerImpl) GetAll(w http.ResponseWriter, r *http.Request) {
	bean, err := impl.environmentClusterMappingsService.GetAll()
	if err != nil {
		impl.logger.Errorw("service err, GetAll", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	var result []request.EnvironmentBean
	token := r.Header.Get("token")
	for _, item := range bean {
		// RBAC enforcer applying
		if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobalEnvironment, rbac.ActionGet, strings.ToLower(item.Environment)); ok {
			result = append(result, item)
		}
		//RBAC enforcer Ends
	}

	writeJsonResp(w, err, result, http.StatusOK)
}

func (impl EnvironmentRestHandlerImpl) GetAllActive(w http.ResponseWriter, r *http.Request) {
	bean, err := impl.environmentClusterMappingsService.GetAllActive()
	if err != nil {
		impl.logger.Errorw("service err, GetAllActive", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	var result []request.EnvironmentBean
	token := r.Header.Get("token")
	for _, item := range bean {
		// RBAC enforcer applying
		if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobalEnvironment, rbac.ActionGet, strings.ToLower(item.Environment)); ok {
			result = append(result, item)
		}
		//RBAC enforcer Ends
	}

	writeJsonResp(w, err, result, http.StatusOK)
}

func (impl EnvironmentRestHandlerImpl) Update(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	var bean request.EnvironmentBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("service err, Update", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, Update", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, Update", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobalEnvironment, rbac.ActionUpdate, strings.ToLower(bean.Environment)); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := impl.environmentClusterMappingsService.Update(&bean, userId)
	if err != nil {
		impl.logger.Errorw("service err, Update", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl EnvironmentRestHandlerImpl) FindById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	envId, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean, err := impl.environmentClusterMappingsService.FindById(envId)
	if err != nil {
		impl.logger.Errorw("service err, FindById", "err", err, "envId", envId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobalEnvironment, rbac.ActionGet, strings.ToLower(bean.Environment)); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	writeJsonResp(w, err, bean, http.StatusOK)
}

func (impl EnvironmentRestHandlerImpl) GetEnvironmentListForAutocomplete(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	environments, err := impl.environmentClusterMappingsService.GetEnvironmentListForAutocomplete()
	if err != nil {
		impl.logger.Errorw("service err, GetEnvironmentListForAutocomplete", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

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
	token := r.Header.Get("token")
	// RBAC enforcer applying
	var grantedEnvironment []request.EnvironmentBean
	for _, item := range environments {
		if authEnabled == true {
			if ok := impl.enforcer.Enforce(token, rbac.ResourceGlobalEnvironment, rbac.ActionGet, strings.ToLower(item.Environment)); ok {
				grantedEnvironment = append(grantedEnvironment, item)
			}
		} else {
			grantedEnvironment = append(grantedEnvironment, item)
		}
	}
	//RBAC enforcer Ends
	if len(grantedEnvironment) == 0 {
		grantedEnvironment = make([]request.EnvironmentBean, 0)
	}
	writeJsonResp(w, err, grantedEnvironment, http.StatusOK)
}
