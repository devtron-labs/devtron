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
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/internal/casbin"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sso"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/devtron-labs/devtron/util/response"
	"github.com/gorilla/mux"
	"github.com/nats-io/stan"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
)

type UserAuthHandler interface {
	LoginHandler(w http.ResponseWriter, r *http.Request)
	CallbackHandler(w http.ResponseWriter, r *http.Request)
	AddPolicy(w http.ResponseWriter, r *http.Request)
	RemovePolicy(w http.ResponseWriter, r *http.Request)
	RefreshTokenHandler(w http.ResponseWriter, r *http.Request)
	//SessionOptions(w http.ResponseWriter, r *http.Request)

	GetAllSubjectsFromCasbin(w http.ResponseWriter, r *http.Request)
	GetRolesForUserFromCasbin(w http.ResponseWriter, r *http.Request)
	GetUserByRoleFromCasbin(w http.ResponseWriter, r *http.Request)
	DeleteRoleForUserFromCasbin(w http.ResponseWriter, r *http.Request)

	CreateRoleFromOrchestrator(w http.ResponseWriter, r *http.Request)
	UpdateRoleFromOrchestrator(w http.ResponseWriter, r *http.Request)

	GetRolesByUserIdFromOrchestrator(w http.ResponseWriter, r *http.Request)
	GetAllRoleFromOrchestrator(w http.ResponseWriter, r *http.Request)
	GetRoleByFilterFromOrchestrator(w http.ResponseWriter, r *http.Request)
	DeleteRoleFromOrchestrator(w http.ResponseWriter, r *http.Request)

	AddDefaultPolicyAndRoles(w http.ResponseWriter, r *http.Request)
	Subscribe() error
	AuthVerification(w http.ResponseWriter, r *http.Request)
	CreateSSOLoginConfig(w http.ResponseWriter, r *http.Request)
	UpdateSSOLoginConfig(w http.ResponseWriter, r *http.Request)
	GetAllSSOLoginConfig(w http.ResponseWriter, r *http.Request)
	GetSSOLoginConfig(w http.ResponseWriter, r *http.Request)
	GetSSOLoginConfigByName(w http.ResponseWriter, r *http.Request)
}

type UserAuthHandlerImpl struct {
	userAuthService user.UserAuthService
	validator       *validator.Validate
	logger          *zap.SugaredLogger
	enforcer        rbac.Enforcer
	natsClient      *pubsub.PubSubClient
	userService     user.UserService
	ssoLoginService sso.SSOLoginService
}

const POLICY_UPDATE_TOPIC = "Policy.Update"

func NewUserAuthHandlerImpl(userAuthService user.UserAuthService, validator *validator.Validate,
	logger *zap.SugaredLogger, enforcer rbac.Enforcer, natsClient *pubsub.PubSubClient, userService user.UserService,
	ssoLoginService sso.SSOLoginService) *UserAuthHandlerImpl {
	userAuthHandler := &UserAuthHandlerImpl{userAuthService: userAuthService, validator: validator, logger: logger,
		enforcer: enforcer, natsClient: natsClient, userService: userService, ssoLoginService: ssoLoginService}

	err := userAuthHandler.Subscribe()
	if err != nil {
		logger.Errorw("subscribe err, POLICY_UPDATE_TOPIC", "err", err)
		return nil
	}
	return userAuthHandler
}

func (handler UserAuthHandlerImpl) LoginHandler(w http.ResponseWriter, r *http.Request) {
	up := &userNamePassword{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(up)
	if err != nil {
		handler.logger.Errorw("request err, LoginHandler", "err", err, "payload", up)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
	}

	err = handler.validator.Struct(up)
	if err != nil {
		handler.logger.Errorw("validation err, LoginHandler", "err", err, "payload", up)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token, err := handler.userAuthService.HandleLogin(up.Username, up.Password)
	if err != nil {
		writeJsonResp(w, fmt.Errorf("invalid username or password"), nil, http.StatusForbidden)
		return
	}
	response := make(map[string]interface{})
	response["token"] = token
	http.SetCookie(w, &http.Cookie{Name: "argocd.token", Value: token, Path: "/"})
	writeJsonResp(w, nil, response, http.StatusOK)
}

func (handler UserAuthHandlerImpl) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	handler.userAuthService.HandleDexCallback(w, r)
}

func (handler UserAuthHandlerImpl) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	handler.userAuthService.HandleRefresh(w, r)
}

func (handler UserAuthHandlerImpl) AddPolicy(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var policies bean.PolicyRequest
	err = decoder.Decode(&policies)
	if err != nil {
		handler.logger.Errorw("request err, AddPolicy", "err", err, "payload", policies)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, AddPolicy", "payload", policies)

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceAdmin, rbac.ActionCreate, string(userId)); !ok {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
	}
	//RBAC enforcer Ends

	failedPolicies := casbin.AddPolicy(policies.Data)
	//UPDATE CASBIN SUBJECT TABLE IN ORCHESTRATOR
	if len(policies.Data) != len(failedPolicies) {
		//e.LoadPolicy()
		err := handler.natsClient.Conn.Publish(POLICY_UPDATE_TOPIC, []byte("{}"))
		if err != nil {
			handler.logger.Error("err", err)
			return
		}
	}
	if len(failedPolicies) > 0 {
		writeJsonResp(w, err, "Few Polices Failed To Add", http.StatusPartialContent)
	} else {
		writeJsonResp(w, err, "Policies Successfully Added", http.StatusOK)
	}
	return
}
func (handler UserAuthHandlerImpl) RemovePolicy(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var policies bean.PolicyRequest
	err = decoder.Decode(&policies)
	if err != nil {
		handler.logger.Errorw("request err, RemovePolicy", "err", err, "payload", policies)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, RemovePolicy", "payload", policies)
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceAdmin, rbac.ActionDelete, userId); !ok {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
	}
	//RBAC enforcer Ends
	mergeResp := casbin.RemovePolicy(policies.Data)
	if len(mergeResp) == 0 {
		writeJsonResp(w, err, []byte{}, http.StatusOK)
	}
	resJson, err := json.Marshal(mergeResp)
	if err != nil {
		handler.logger.Errorw("marshal err, RemovePolicy", "err", err, "payload", mergeResp)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, resJson, http.StatusOK)
}

func (handler UserAuthHandlerImpl) Subscribe() error {
	_, err := handler.natsClient.Conn.Subscribe(POLICY_UPDATE_TOPIC, func(msg *stan.Msg) {
		handler.logger.Debugw("msg received by subscriber for - Policy Load", "msg", msg)
		casbin.LoadPolicy()
	})
	if err != nil {
		handler.logger.Errorw("subscribe err, POLICY_UPDATE_TOPIC", "err", err)
		return err
	}
	return nil
}

func (handler UserAuthHandlerImpl) GetAllSubjectsFromCasbin(w http.ResponseWriter, r *http.Request) {
	pRes := casbin.GetAllSubjects()
	writeJsonResp(w, nil, pRes, http.StatusOK)
}
func (handler UserAuthHandlerImpl) DeleteRoleForUserFromCasbin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	role := vars["role"]
	pRes := casbin.DeleteRoleForUser(user, role)
	writeJsonResp(w, nil, pRes, http.StatusOK)
}
func (handler UserAuthHandlerImpl) GetRolesForUserFromCasbin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	pRes, err := casbin.GetRolesForUser(user)
	if err != nil {
		handler.logger.Errorw("service err, GetRolesForUserFromCasbin", "err", err, "user", user)
		writeJsonResp(w, err, pRes, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, pRes, http.StatusOK)
}
func (handler UserAuthHandlerImpl) GetUserByRoleFromCasbin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	role := vars["role"]
	pRes, err := casbin.GetUserByRole(role)
	if err != nil {
		handler.logger.Errorw("service err, GetUserByRoleFromCasbin", "err", err, "role", role)
		writeJsonResp(w, err, pRes, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, pRes, http.StatusOK)
}

func (handler UserAuthHandlerImpl) CreateRoleFromOrchestrator(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var roleData bean.RoleData
	err := decoder.Decode(&roleData)
	if err != nil {
		handler.logger.Errorw("request err, CreateRoleFromOrchestrator", "err", err, "payload", roleData)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, CreateRoleFromOrchestrator", "payload", roleData)
	_, err = handler.userAuthService.CreateRole(&roleData)
	if err != nil {
		handler.logger.Errorw("service err, CreateRoleFromOrchestrator", "err", err, "payload", roleData)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, nil, http.StatusOK)
}
func (handler UserAuthHandlerImpl) UpdateRoleFromOrchestrator(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var roleData bean.RoleData
	err := decoder.Decode(&roleData)
	if err != nil {
		handler.logger.Errorw("request err, UpdateRoleFromOrchestrator", "err", err, "payload", roleData)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, UpdateRoleFromOrchestrator", "payload", roleData)
	_, err = handler.userAuthService.UpdateRole(&roleData)
	if err != nil {
		handler.logger.Errorw("service err, UpdateRoleFromOrchestrator", "err", err, "payload", roleData)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, nil, http.StatusOK)
}
func (handler UserAuthHandlerImpl) GetRolesByUserIdFromOrchestrator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	/* #nosec */
	userId, err := strconv.Atoi(vars["userId"])
	if err != nil {
		handler.logger.Errorw("request err, GetRolesByUserIdFromOrchestrator", "err", err, "userId", userId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := handler.userAuthService.GetRolesByUserId(int32(userId))
	if err != nil {
		handler.logger.Errorw("service err, GetRolesByUserIdFromOrchestrator", "err", err, "userId", userId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}
func (handler UserAuthHandlerImpl) GetAllRoleFromOrchestrator(w http.ResponseWriter, r *http.Request) {
	res, err := handler.userAuthService.GetAllRole()
	if err != nil {
		handler.logger.Errorw("service err, GetAllRoleFromOrchestrator", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}
func (handler UserAuthHandlerImpl) GetRoleByFilterFromOrchestrator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entity := vars["entity"]
	team := vars["team"]
	app := vars["app"]
	env := vars["env"]
	act := vars["act"]
	res, err := handler.userAuthService.GetRoleByFilter(entity, team, app, env, act)
	if err != nil {
		handler.logger.Errorw("service err, GetRoleByFilterFromOrchestrator", "err", err, "team", team, "app", app, "env", env, "act", act)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}
func (handler UserAuthHandlerImpl) DeleteRoleFromOrchestrator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	role := vars["role"]
	flag, err := handler.userAuthService.DeleteRole(role)
	if err != nil {
		handler.logger.Errorw("service err, DeleteRoleFromOrchestrator", "err", err, "role", role)
		writeJsonResp(w, err, false, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, flag, http.StatusOK)
}

func (handler UserAuthHandlerImpl) AddDefaultPolicyAndRoles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	team := vars["team"]
	app := vars["app"]
	env := vars["env"]
	handler.logger.Infow("request payload, AddDefaultPolicyAndRoles", "team", team, "app", app, "env", env)
	adminPolicies := "{\n    \"data\": [\n        {\n            \"type\": \"p\",\n            \"sub\": \"role:admin_<TEAM>_<ENV>_<APP>\",\n            \"res\": \"applications\",\n            \"act\": \"*\",\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\n        },\n        {\n            \"type\": \"p\",\n            \"sub\": \"role:admin_<TEAM>_<ENV>_<APP>\",\n            \"res\": \"environment\",\n            \"act\": \"trigger\",\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\n        },\n        {\n            \"type\": \"p\",\n            \"sub\": \"role:admin_<TEAM>_<ENV>_<APP>\",\n            \"res\": \"team\",\n            \"act\": \"get\",\n            \"obj\": \"<TEAM_OBJ>\"\n        }\n    ]\n}"
	triggerPolicies := "{\n    \"data\": [\n        {\n            \"type\": \"p\",\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\n            \"res\": \"applications\",\n            \"act\": \"get\",\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\n        },\n        {\n            \"type\": \"p\",\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\n            \"res\": \"applications\",\n            \"act\": \"trigger\",\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\n        },\n        {\n            \"type\": \"p\",\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\n            \"res\": \"environment\",\n            \"act\": \"trigger\",\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\n        }\n    ]\n}"
	viewPolicies := "{\n    \"data\": [\n        {\n            \"type\": \"p\",\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\n            \"res\": \"applications\",\n            \"act\": \"get\",\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\n        }\n    ]\n}"

	adminPolicies = strings.ReplaceAll(adminPolicies, "<TEAM>", team)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<ENV>", env)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<APP>", app)

	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<TEAM>", team)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<ENV>", env)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<APP>", app)

	viewPolicies = strings.ReplaceAll(viewPolicies, "<TEAM>", team)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<ENV>", env)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<APP>", app)

	//for START in Casbin Object
	teamObj := team
	envObj := env
	appObj := app
	if len(teamObj) == 0 {
		teamObj = "*"
	}
	if len(envObj) == 0 {
		envObj = "*"
	}
	if len(appObj) == 0 {
		appObj = "*"
	}
	adminPolicies = strings.ReplaceAll(adminPolicies, "<TEAM_OBJ>", teamObj)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<ENV_OBJ>", envObj)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<APP_OBJ>", appObj)

	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<TEAM_OBJ>", teamObj)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<ENV_OBJ>", envObj)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<APP_OBJ>", appObj)

	viewPolicies = strings.ReplaceAll(viewPolicies, "<TEAM_OBJ>", teamObj)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<ENV_OBJ>", envObj)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<APP_OBJ>", appObj)
	//for START in Casbin Object Ends Here

	var policiesAdmin bean.PolicyRequest
	err := json.Unmarshal([]byte(adminPolicies), &policiesAdmin)
	if err != nil {
		handler.logger.Errorw("request err, AddDefaultPolicyAndRoles", "err", err, "payload", policiesAdmin)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, AddDefaultPolicyAndRoles", "policiesAdmin", policiesAdmin)
	casbin.AddPolicy(policiesAdmin.Data)

	var policiesTrigger bean.PolicyRequest
	err = json.Unmarshal([]byte(triggerPolicies), &policiesTrigger)
	if err != nil {
		handler.logger.Errorw("request err, AddDefaultPolicyAndRoles", "err", err, "payload", policiesTrigger)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, AddDefaultPolicyAndRoles", "policiesTrigger", policiesTrigger)
	casbin.AddPolicy(policiesTrigger.Data)

	var policiesView bean.PolicyRequest
	err = json.Unmarshal([]byte(viewPolicies), &policiesView)
	if err != nil {
		handler.logger.Errorw("request err, AddDefaultPolicyAndRoles", "err", err, "payload", policiesView)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, AddDefaultPolicyAndRoles", "policiesView", policiesView)
	casbin.AddPolicy(policiesView.Data)

	//Creating ROLES
	roleAdmin := "{\n    \"role\": \"role:admin_<TEAM>_<ENV>_<APP>\",\n    \"casbinSubjects\": [\n        \"role:admin_<TEAM>_<ENV>_<APP>\"\n    ],\n    \"team\": \"<TEAM>\",\n    \"application\": \"<APP>\",\n    \"environment\": \"<ENV>\",\n    \"action\": \"*\"\n}"
	roleTrigger := "{\n    \"role\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\n    \"casbinSubjects\": [\n        \"role:trigger_<TEAM>_<ENV>_<APP>\"\n    ],\n    \"team\": \"<TEAM>\",\n    \"application\": \"<APP>\",\n    \"environment\": \"<ENV>\",\n    \"action\": \"trigger\"\n}"
	roleView := "{\n    \"role\": \"role:view_<TEAM>_<ENV>_<APP>\",\n    \"casbinSubjects\": [\n        \"role:view_<TEAM>_<ENV>_<APP>\"\n    ],\n    \"team\": \"<TEAM>\",\n    \"application\": \"<APP>\",\n    \"environment\": \"<ENV>\",\n    \"action\": \"view\"\n}"
	roleAdmin = strings.ReplaceAll(roleAdmin, "<TEAM>", team)
	roleAdmin = strings.ReplaceAll(roleAdmin, "<ENV>", env)
	roleAdmin = strings.ReplaceAll(roleAdmin, "<APP>", app)

	roleTrigger = strings.ReplaceAll(roleTrigger, "<TEAM>", team)
	roleTrigger = strings.ReplaceAll(roleTrigger, "<ENV>", env)
	roleTrigger = strings.ReplaceAll(roleTrigger, "<APP>", app)

	roleView = strings.ReplaceAll(roleView, "<TEAM>", team)
	roleView = strings.ReplaceAll(roleView, "<ENV>", env)
	roleView = strings.ReplaceAll(roleView, "<APP>", app)

	var roleAdminData bean.RoleData
	err = json.Unmarshal([]byte(roleAdmin), &roleAdminData)
	if err != nil {
		handler.logger.Errorw("request err, AddDefaultPolicyAndRoles", "err", err, "payload", roleAdminData)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	_, err = handler.userAuthService.CreateRole(&roleAdminData)
	if err != nil {
		handler.logger.Errorw("service err, AddDefaultPolicyAndRoles", "err", err, "payload", roleAdminData)
		writeJsonResp(w, err, "Role Creation Failed", http.StatusInternalServerError)
		return
	}

	var roleTriggerData bean.RoleData
	err = json.Unmarshal([]byte(roleTrigger), &roleTriggerData)
	if err != nil {
		handler.logger.Errorw("request err, AddDefaultPolicyAndRoles", "err", err, "payload", roleTriggerData)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	_, err = handler.userAuthService.CreateRole(&roleTriggerData)
	if err != nil {
		handler.logger.Errorw("service err, AddDefaultPolicyAndRoles", "err", err, "payload", roleTriggerData)
		writeJsonResp(w, err, "Role Creation Failed", http.StatusInternalServerError)
		return
	}

	var roleViewData bean.RoleData
	err = json.Unmarshal([]byte(roleView), &roleViewData)
	if err != nil {
		handler.logger.Errorw("request err, AddDefaultPolicyAndRoles", "err", err, "payload", roleViewData)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	_, err = handler.userAuthService.CreateRole(&roleViewData)
	if err != nil {
		handler.logger.Errorw("service err, AddDefaultPolicyAndRoles", "err", err, "payload", roleTriggerData)
		writeJsonResp(w, err, "Role Creation Failed", http.StatusInternalServerError)
		return
	}

	return
}

func (handler UserAuthHandlerImpl) AuthVerification(w http.ResponseWriter, r *http.Request) {
	res, err := handler.userAuthService.AuthVerification(r)
	if err != nil {
		handler.logger.Errorw("service err, AuthVerification", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}

func (handler UserAuthHandlerImpl) CreateSSOLoginConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var dto bean.SSOLoginDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, CreateSSOLoginConfig", "err", err, "payload", dto)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if err != nil {
		handler.logger.Errorw("request err, CreateSSOLoginConfig", "err", err, "userId", userId)
		writeJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}

	if !isActionUserSuperAdmin {
		err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to perform operation"}
		writeJsonResp(w, err, "", http.StatusForbidden)
		return
	}

	dto.Url = fmt.Sprintf("%s://%s", r.URL.Scheme, r.Host)
	handler.logger.Infow("request payload, CreateSSOLoginConfig", "payload", dto)
	resp, err := handler.ssoLoginService.CreateSSOLogin(&dto)
	if err != nil {
		handler.logger.Errorw("service err, CreateSSOLoginConfig", "err", err, "payload", dto)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, resp, http.StatusOK)
}
func (handler UserAuthHandlerImpl) UpdateSSOLoginConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var dto bean.SSOLoginDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, UpdateSSOLoginConfig", "err", err, "payload", dto)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if err != nil {
		handler.logger.Errorw("request err, UpdateSSOLoginConfig", "err", err, "userId", userId)
		writeJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}

	if !isActionUserSuperAdmin {
		err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to perform operation"}
		writeJsonResp(w, err, "", http.StatusForbidden)
		return
	}

	dto.Url = fmt.Sprintf("%s://%s", r.URL.Scheme, r.Host)
	handler.logger.Infow("request payload, UpdateSSOLoginConfig", "payload", dto)
	resp, err := handler.ssoLoginService.UpdateSSOLogin(&dto)
	if err != nil {
		handler.logger.Errorw("service err, UpdateSSOLoginConfig", "err", err, "payload", dto)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, resp, http.StatusOK)
}

func (handler UserAuthHandlerImpl) GetAllSSOLoginConfig(w http.ResponseWriter, r *http.Request) {
	res, err := handler.ssoLoginService.GetAll()
	if err != nil {
		handler.logger.Errorw("service err, GetAllSSOLoginConfig", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}

func (handler UserAuthHandlerImpl) GetSSOLoginConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	/* #nosec */
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, GetSSOLoginConfig", "err", err, "id", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if err != nil {
		handler.logger.Errorw("request err, GetSSOLoginConfig", "err", err, "userId", userId)
		writeJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}

	if !isActionUserSuperAdmin {
		err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to perform operation"}
		writeJsonResp(w, err, "", http.StatusForbidden)
		return
	}

	res, err := handler.ssoLoginService.GetById(int32(id))
	if err != nil {
		handler.logger.Errorw("service err, GetSSOLoginConfig", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}

func (handler UserAuthHandlerImpl) GetSSOLoginConfigByName(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isActionUserSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if err != nil {
		handler.logger.Errorw("request err, GetSSOLoginConfigByName", "err", err, "userId", userId)
		writeJsonResp(w, err, "Failed to check is super admin", http.StatusInternalServerError)
		return
	}

	if !isActionUserSuperAdmin {
		err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to perform operation"}
		writeJsonResp(w, err, "", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]
	res, err := handler.ssoLoginService.GetByName(name)
	if err != nil {
		handler.logger.Errorw("service err, GetSSOLoginConfigByName", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}
