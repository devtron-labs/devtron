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

package user

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"net/http"
	"strconv"
	"strings"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/response"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type UserRestHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	GetById(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)

	FetchRoleGroupById(w http.ResponseWriter, r *http.Request)
	CreateRoleGroup(w http.ResponseWriter, r *http.Request)
	UpdateRoleGroup(w http.ResponseWriter, r *http.Request)
	FetchRoleGroups(w http.ResponseWriter, r *http.Request)
	FetchRoleGroupsByName(w http.ResponseWriter, r *http.Request)
	DeleteRoleGroup(w http.ResponseWriter, r *http.Request)
	CheckUserRoles(w http.ResponseWriter, r *http.Request)
	SyncOrchestratorToCasbin(w http.ResponseWriter, r *http.Request)
	UpdateTriggerPolicyForTerminalAccess(w http.ResponseWriter, r *http.Request)
}

type userNamePassword struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type UserRestHandlerImpl struct {
	userService      user.UserService
	validator        *validator.Validate
	logger           *zap.SugaredLogger
	enforcer         casbin.Enforcer
	roleGroupService user.RoleGroupService
}

func NewUserRestHandlerImpl(userService user.UserService, validator *validator.Validate,
	logger *zap.SugaredLogger, enforcer casbin.Enforcer, roleGroupService user.RoleGroupService) *UserRestHandlerImpl {
	userAuthHandler := &UserRestHandlerImpl{userService: userService, validator: validator, logger: logger,
		enforcer: enforcer, roleGroupService: roleGroupService}
	return userAuthHandler
}

func (handler UserRestHandlerImpl) CreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var userInfo bean.UserInfo
	err = decoder.Decode(&userInfo)
	if err != nil {
		handler.logger.Errorw("request err, CreateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	userInfo.UserId = userId
	handler.logger.Infow("request payload, CreateUser", "payload", userInfo)

	// RBAC enforcer applying
	token := r.Header.Get("token")
	isActionUserSuperAdmin := false
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	if userInfo.RoleFilters != nil && len(userInfo.RoleFilters) > 0 {
		for _, filter := range userInfo.RoleFilters {
			if filter.AccessType == bean.APP_ACCESS_TYPE_HELM && !isActionUserSuperAdmin {
				response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
				return
			}
			if len(filter.Team) > 0 {
				if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, strings.ToLower(filter.Team)); !ok {
					response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
					return
				}
			}
		}
	} else {
		if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, "*"); !ok {
			response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
			return
		}
	}

	// auth check inside groups
	if len(userInfo.Groups) > 0 {
		groupRoles, err := handler.roleGroupService.FetchRolesForGroups(userInfo.Groups)
		if err != nil && err != pg.ErrNoRows {
			handler.logger.Errorw("service err, UpdateUser", "err", err, "payload", userInfo)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
			return
		}

		if len(groupRoles) > 0 {
			for _, groupRole := range groupRoles {
				if groupRole.AccessType == bean.APP_ACCESS_TYPE_HELM && !isActionUserSuperAdmin {
					response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
					return
				}
				if len(groupRole.Team) > 0 {
					if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, strings.ToLower(groupRole.Team)); !ok {
						response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
						return
					}
				}
			}
		} else {
			if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, "*"); !ok {
				response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
				return
			}
		}
	}
	//RBAC enforcer Ends

	handler.logger.Infow("request payload, CreateUser ", "payload", userInfo)
	err = handler.validator.Struct(userInfo)
	if err != nil {
		handler.logger.Errorw("validation err, CreateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	res, err := handler.userService.CreateUser(&userInfo)
	if err != nil {
		handler.logger.Errorw("service err, CreateUser", "err", err, "payload", userInfo)
		if _, ok := err.(*util.ApiError); ok {
			common.WriteJsonResp(w, err, "User Creation Failed", http.StatusOK)
		} else {
			handler.logger.Errorw("error on creating new user", "err", err)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		}
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) UpdateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var userInfo bean.UserInfo
	err = decoder.Decode(&userInfo)
	if err != nil {
		handler.logger.Errorw("request err, UpdateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	userInfo.UserId = userId
	handler.logger.Infow("request payload, UpdateUser", "payload", userInfo)

	// RBAC enforcer applying
	token := r.Header.Get("token")
	isActionUserSuperAdmin := false
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	if userInfo.RoleFilters != nil && len(userInfo.RoleFilters) > 0 {
		for _, filter := range userInfo.RoleFilters {
			if filter.AccessType == bean.APP_ACCESS_TYPE_HELM && !isActionUserSuperAdmin {
				response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
				return
			}
			if len(filter.Team) > 0 {
				if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionUpdate, strings.ToLower(filter.Team)); !ok {
					common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
					return
				}
			}
		}
	} else {
		if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionUpdate, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	}

	// auth check inside groups
	if len(userInfo.Groups) > 0 {
		groupRoles, err := handler.roleGroupService.FetchRolesForGroups(userInfo.Groups)
		if err != nil && err != pg.ErrNoRows {
			handler.logger.Errorw("service err, UpdateUser", "err", err, "payload", userInfo)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
			return
		}

		if len(groupRoles) > 0 {
			for _, groupRole := range groupRoles {
				if groupRole.AccessType == bean.APP_ACCESS_TYPE_HELM && !isActionUserSuperAdmin {
					response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
					return
				}
				if len(groupRole.Team) > 0 {
					if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionUpdate, strings.ToLower(groupRole.Team)); !ok {
						response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
						return
					}
				}
			}
		} else {
			if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionUpdate, "*"); !ok {
				response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
				return
			}
		}
	}
	//RBAC enforcer Ends

	if userInfo.EmailId == "admin" {
		userInfo.EmailId = "admin@github.com/devtron-labs"
	}
	err = handler.validator.Struct(userInfo)
	if err != nil {
		handler.logger.Errorw("validation err, UpdateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	if userInfo.EmailId == "admin@github.com/devtron-labs" {
		userInfo.EmailId = "admin"
	}
	res, err := handler.userService.UpdateUser(&userInfo)
	if err != nil {
		handler.logger.Errorw("service err, UpdateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) GetById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	/* #nosec */
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, GetById", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := handler.userService.GetById(int32(id))
	if err != nil {
		handler.logger.Errorw("service err, GetById", "err", err, "id", id)
		common.WriteJsonResp(w, err, "Failed to get by id", http.StatusInternalServerError)
		return
	}

	isActionUserSuperAdmin := false
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}

	// NOTE: if no role assigned, user will be visible to all manager.
	// RBAC enforcer applying
	if res.RoleFilters != nil && len(res.RoleFilters) > 0 {
		authPass := false
		for _, filter := range res.RoleFilters {
			if len(filter.Team) > 0 {
				if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, strings.ToLower(filter.Team)); ok {
					authPass = true
				}
			}
		}
		if len(res.RoleFilters) == 1 && res.RoleFilters[0].Entity == casbin.ResourceChartGroup {
			authPass = true
		}
		if isActionUserSuperAdmin {
			authPass = true
		}
		if authPass == false {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) GetAll(w http.ResponseWriter, r *http.Request) {
	res, err := handler.userService.GetAll()
	if err != nil {
		handler.logger.Errorw("service err, GetAll", "err", err)
		common.WriteJsonResp(w, err, "Failed to Get", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	/* #nosec */
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, DeleteUser", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, DeleteUser", "err", err, "id", id)
	user, err := handler.userService.GetById(int32(id))
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	isActionUserSuperAdmin := false
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	if user.RoleFilters != nil && len(user.RoleFilters) > 0 {
		for _, filter := range user.RoleFilters {
			if filter.AccessType == bean.APP_ACCESS_TYPE_HELM && !isActionUserSuperAdmin {
				common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
				return
			}
			if len(filter.Team) > 0 {
				if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionDelete, strings.ToLower(filter.Team)); !ok {
					common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
					return
				}
			}
		}
	} else {
		if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionDelete, ""); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	}
	//RBAC enforcer Ends

	res, err := handler.userService.DeleteUser(user)
	if err != nil {
		handler.logger.Errorw("service err, DeleteUser", "err", err, "id", id)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) FetchRoleGroupById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	/* #nosec */
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, FetchRoleGroupById", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := handler.roleGroupService.FetchRoleGroupsById(int32(id))
	if err != nil {
		handler.logger.Errorw("service err, FetchRoleGroupById", "err", err, "id", id)
		common.WriteJsonResp(w, err, "Failed to get by id", http.StatusInternalServerError)
		return
	}

	// NOTE: if no role assigned, user will be visible to all manager.
	// RBAC enforcer applying
	token := r.Header.Get("token")
	isActionUserSuperAdmin := false
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	if res.RoleFilters != nil && len(res.RoleFilters) > 0 {
		for _, filter := range res.RoleFilters {
			if filter.AccessType == bean.APP_ACCESS_TYPE_HELM && !isActionUserSuperAdmin {
				common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
				return
			}
			if len(filter.Team) > 0 {
				if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, strings.ToLower(filter.Team)); !ok {
					common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
					return
				}
			}
		}
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) CreateRoleGroup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request bean.RoleGroup
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, CreateRoleGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	handler.logger.Infow("request payload, CreateRoleGroup", "err", err, "payload", request)

	// RBAC enforcer applying
	token := r.Header.Get("token")
	isActionUserSuperAdmin := false
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	if request.RoleFilters != nil && len(request.RoleFilters) > 0 {
		for _, filter := range request.RoleFilters {
			if filter.AccessType == bean.APP_ACCESS_TYPE_HELM && !isActionUserSuperAdmin {
				common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
				return
			}
			if len(filter.Team) > 0 {
				if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, strings.ToLower(filter.Team)); !ok {
					common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
					return
				}
			}
		}
	} else {
		if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	}
	//RBAC enforcer Ends
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err, CreateRoleGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	res, err := handler.roleGroupService.CreateRoleGroup(&request)
	if err != nil {
		handler.logger.Errorw("service err, CreateRoleGroup", "err", err, "payload", request)
		if _, ok := err.(*util.ApiError); ok {
			common.WriteJsonResp(w, err, nil, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) UpdateRoleGroup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request bean.RoleGroup
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, UpdateRoleGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	handler.logger.Infow("request payload, UpdateRoleGroup", "err", err, "payload", request)
	// RBAC enforcer applying
	token := r.Header.Get("token")
	isActionUserSuperAdmin := false
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	if request.RoleFilters != nil && len(request.RoleFilters) > 0 {
		for _, filter := range request.RoleFilters {
			if filter.AccessType == bean.APP_ACCESS_TYPE_HELM && !isActionUserSuperAdmin {
				common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
				return
			}
			if len(filter.Team) > 0 {
				if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionUpdate, strings.ToLower(filter.Team)); !ok {
					common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
					return
				}
			}
		}
	} else {
		if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionUpdate, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	}
	//RBAC enforcer Ends

	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err, UpdateRoleGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	res, err := handler.roleGroupService.UpdateRoleGroup(&request)
	if err != nil {
		handler.logger.Errorw("service err, UpdateRoleGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) FetchRoleGroups(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	res, err := handler.roleGroupService.FetchRoleGroups()
	if err != nil {
		handler.logger.Errorw("service err, FetchRoleGroups", "err", err)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) FetchRoleGroupsByName(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	userGroupName := vars["name"]
	res, err := handler.roleGroupService.FetchRoleGroupsByName(userGroupName)
	if err != nil {
		handler.logger.Errorw("service err, FetchRoleGroupsByName", "err", err, "userGroupName", userGroupName)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) DeleteRoleGroup(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	/* #nosec */
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, DeleteRoleGroup", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, DeleteRoleGroup", "id", id)
	userGroup, err := handler.roleGroupService.FetchRoleGroupsById(int32(id))
	if err != nil {
		handler.logger.Errorw("service err, DeleteRoleGroup", "err", err, "id", id)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}

	token := r.Header.Get("token")
	isActionUserSuperAdmin := false
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	if userGroup.RoleFilters != nil && len(userGroup.RoleFilters) > 0 {
		for _, filter := range userGroup.RoleFilters {
			if filter.AccessType == bean.APP_ACCESS_TYPE_HELM && !isActionUserSuperAdmin {
				common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
				return
			}
			if len(filter.Team) > 0 {
				if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionDelete, strings.ToLower(filter.Team)); !ok {
					common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
					return
				}
			}
		}
	}
	//RBAC enforcer Ends

	res, err := handler.roleGroupService.DeleteRoleGroup(userGroup)
	if err != nil {
		handler.logger.Errorw("service err, DeleteRoleGroup", "err", err, "id", id)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) CheckUserRoles(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	roles, err := handler.userService.CheckUserRoles(userId)
	if err != nil {
		handler.logger.Errorw("service err, CheckUserRoles", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	result := make(map[string]interface{})
	result["roles"] = roles
	result["superAdmin"] = true
	for _, item := range roles {
		if item == bean.SUPERADMIN {
			result["superAdmin"] = true
		}
	}
	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (handler UserRestHandlerImpl) SyncOrchestratorToCasbin(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	user, err := handler.userService.GetById(userId)
	if err != nil {
		handler.logger.Errorw("service err, SyncOrchestratorToCasbin", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if user.EmailId != "admin" {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	flag, err := handler.userService.SyncOrchestratorToCasbin()
	if err != nil {
		handler.logger.Errorw("service err, SyncOrchestratorToCasbin", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, flag, http.StatusOK)
}

func (handler UserRestHandlerImpl) UpdateTriggerPolicyForTerminalAccess(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		handler.logger.Errorw("unauthorized user, UpdateTriggerPolicyForTerminalAccess", "userId", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		handler.logger.Errorw("unauthorized user, UpdateTriggerPolicyForTerminalAccess", "userId", userId)
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	err = handler.userService.UpdateTriggerPolicyForTerminalAccess()
	if err != nil {
		handler.logger.Errorw("error in updating trigger policy for terminal access", "err", err)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "Trigger policy updated successfully.", http.StatusOK)
}
