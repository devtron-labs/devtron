/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package user

import (
	"encoding/json"
	"errors"
	util2 "github.com/devtron-labs/devtron/api/auth/user/util"
	"github.com/devtron-labs/devtron/pkg/auth/user/helper"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/go-pg/pg"
	"github.com/gorilla/schema"
	"net/http"
	"strconv"
	"strings"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	user2 "github.com/devtron-labs/devtron/pkg/auth/user"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/util/response"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type UserRestHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	GetById(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetAllV2(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	GetAllDetailedUsers(w http.ResponseWriter, r *http.Request)
	BulkDeleteUsers(w http.ResponseWriter, r *http.Request)
	FetchRoleGroupById(w http.ResponseWriter, r *http.Request)
	CreateRoleGroup(w http.ResponseWriter, r *http.Request)
	UpdateRoleGroup(w http.ResponseWriter, r *http.Request)
	FetchRoleGroups(w http.ResponseWriter, r *http.Request)
	FetchRoleGroupsV2(w http.ResponseWriter, r *http.Request)
	FetchDetailedRoleGroups(w http.ResponseWriter, r *http.Request)
	FetchRoleGroupsByName(w http.ResponseWriter, r *http.Request)
	DeleteRoleGroup(w http.ResponseWriter, r *http.Request)
	BulkDeleteRoleGroups(w http.ResponseWriter, r *http.Request)
	CheckUserRoles(w http.ResponseWriter, r *http.Request)
	SyncOrchestratorToCasbin(w http.ResponseWriter, r *http.Request)
	UpdateTriggerPolicyForTerminalAccess(w http.ResponseWriter, r *http.Request)
	GetRoleCacheDump(w http.ResponseWriter, r *http.Request)
	InvalidateRoleCache(w http.ResponseWriter, r *http.Request)
}

type userNamePassword struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type UserRestHandlerImpl struct {
	userService       user2.UserService
	validator         *validator.Validate
	logger            *zap.SugaredLogger
	enforcer          casbin.Enforcer
	roleGroupService  user2.RoleGroupService
	userCommonService user2.UserCommonService
}

func NewUserRestHandlerImpl(userService user2.UserService, validator *validator.Validate,
	logger *zap.SugaredLogger, enforcer casbin.Enforcer, roleGroupService user2.RoleGroupService,
	userCommonService user2.UserCommonService) *UserRestHandlerImpl {
	userAuthHandler := &UserRestHandlerImpl{
		userService:       userService,
		validator:         validator,
		logger:            logger,
		enforcer:          enforcer,
		roleGroupService:  roleGroupService,
		userCommonService: userCommonService,
	}
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

	// struct Validations
	handler.logger.Infow("request payload, CreateUser ", "payload", userInfo)
	err = handler.validator.Struct(userInfo)
	if err != nil {
		handler.logger.Errorw("validation err, CreateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// Doing this as api is not compatible with previous release of dashboard, groups has been migrated to userRoleGroups
	isGroupsPresent := util2.IsGroupsPresent(userInfo.Groups)
	if isGroupsPresent {
		handler.logger.Errorw("validation error , createUser ", "err", err, "payload", userInfo)
		err := &util.ApiError{Code: "406", HttpStatusCode: 406, UserMessage: "Not compatible with request", InternalMessage: "Not compatible with the request payload, as groups has been migrated to userRoleGroups"}
		common.WriteJsonResp(w, err, nil, http.StatusNotAcceptable)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	isAuthorised, err := handler.checkRBACForUserCreate(token, userInfo.SuperAdmin, userInfo.RoleFilters, userInfo.UserRoleGroup)
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}
	if !isAuthorised {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
	}

	//RBAC enforcer Ends
	//In create req, we also check if any email exists already. If yes, then in that case we go on and merge existing roles and groups with the ones in request
	//but rbac is only checked on create request roles and groups as existing roles and groups are assumed to be checked when created/updated before
	res, err := handler.userService.CreateUser(&userInfo, token, handler.CheckManagerAuth)
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
	return
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

	err = handler.validator.Struct(userInfo)
	if err != nil {
		handler.logger.Errorw("validation err, UpdateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// Doing this as api is not compatible with previous release of dashboard,groups has been migrated to userRoleGroups
	isGroupsPresent := util2.IsGroupsPresent(userInfo.Groups)
	if isGroupsPresent {
		handler.logger.Errorw("validation error , createUser ", "err", err, "payload", userInfo)
		err := &util.ApiError{Code: "406", HttpStatusCode: 406, UserMessage: "Not compatible with request, please update to latest version", InternalMessage: "Not compatible with the request payload, as groups has been migrated to userRoleGroups"}
		common.WriteJsonResp(w, err, nil, http.StatusNotAcceptable)
		return
	}

	res, err := handler.userService.UpdateUser(&userInfo, token, handler.checkRBACForUserUpdate, handler.CheckManagerAuth)
	if err != nil {
		handler.logger.Errorw("service err, UpdateUser", "err", err, "payload", userInfo)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
	//if len(restrictedGroups) == 0 {
	//	common.WriteJsonResp(w, err, res, http.StatusOK)
	//} else {
	//	errorMessageForGroupsWithoutSuperAdmin, errorMessageForGroupsWithSuperAdmin := helper.CreateErrorMessageForUserRoleGroups(restrictedGroups)
	//
	//	if rolesChanged || groupsModified {
	//		// warning
	//		message := fmt.Errorf("User permissions updated partially. %s%s", errorMessageForGroupsWithoutSuperAdmin, errorMessageForGroupsWithSuperAdmin)
	//		common.WriteJsonResp(w, message, nil, http.StatusExpectationFailed)
	//
	//	} else {
	//		//error
	//		message := fmt.Errorf("Permission could not be added/removed. %s%s", errorMessageForGroupsWithoutSuperAdmin, errorMessageForGroupsWithSuperAdmin)
	//		common.WriteJsonResp(w, message, nil, http.StatusBadRequest)
	//	}
	//}
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

	token := r.Header.Get("token")
	// NOTE: if no role assigned, user will be visible to all manager.
	// RBAC enforcer applying
	filteredRoleFilter := make([]bean.RoleFilter, 0)
	if res.RoleFilters != nil && len(res.RoleFilters) > 0 {
		isUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
		for _, filter := range res.RoleFilters {
			authPass := handler.checkRbacForFilter(token, filter, isUserSuperAdmin)
			if authPass {
				filteredRoleFilter = append(filteredRoleFilter, filter)
			}
		}
	}
	for index, roleFilter := range filteredRoleFilter {
		if roleFilter.Entity == "" {
			filteredRoleFilter[index].Entity = bean2.ENTITY_APPS
			if roleFilter.AccessType == "" {
				filteredRoleFilter[index].AccessType = bean2.DEVTRON_APP
			}
		}
	}
	res.RoleFilters = filteredRoleFilter
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) GetAllV2(w http.ResponseWriter, r *http.Request) {
	var decoder = schema.NewDecoder()
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	//checking superAdmin access
	isAuthorised := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if !isAuthorised {
		user, err := handler.userService.GetById(userId)
		if err != nil {
			handler.logger.Errorw("error in getting user by id", "err", err)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
			return
		}
		var roleFilters []bean.RoleFilter
		if len(user.UserRoleGroup) > 0 {
			groupRoleFilters, err := handler.userService.GetRoleFiltersByUserRoleGroups(user.UserRoleGroup)
			if err != nil {
				handler.logger.Errorw("Error in getting role filters by group names", "err", err, "UserRoleGroup", user.UserRoleGroup)
				common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
				return
			}
			if len(groupRoleFilters) > 0 {
				roleFilters = append(roleFilters, groupRoleFilters...)
			}
		}
		if user.RoleFilters != nil && len(user.RoleFilters) > 0 {
			roleFilters = append(roleFilters, user.RoleFilters...)
		}
		if len(roleFilters) > 0 {
			for _, filter := range roleFilters {
				if len(filter.Team) > 0 {
					if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, filter.Team); ok {
						isAuthorised = true
						break
					}
				}
				if filter.Entity == bean2.CLUSTER_ENTITIY {
					if ok := handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth); ok {
						isAuthorised = true
						break
					}
				}
			}
		}
	}
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	req := &bean.ListingRequest{}
	err = decoder.Decode(req, r.URL.Query())
	if err != nil {
		handler.logger.Errorw("request err, GetAll", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := handler.userService.GetAllWithFilters(req)
	if err != nil {
		handler.logger.Errorw("service err, GetAll", "err", err)
		common.WriteJsonResp(w, err, "Failed to Get", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) GetAll(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	//checking superAdmin access
	isAuthorised := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if !isAuthorised {
		user, err := handler.userService.GetById(userId)
		if err != nil {
			handler.logger.Errorw("error in getting user by id", "err", err)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
			return
		}
		var roleFilters []bean.RoleFilter
		if len(user.UserRoleGroup) > 0 {
			groupRoleFilters, err := handler.userService.GetRoleFiltersByUserRoleGroups(user.UserRoleGroup)
			if err != nil {
				handler.logger.Errorw("Error in getting role filters by group names", "err", err, "UserRoleGroup", user.UserRoleGroup)
				common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
				return
			}
			if len(groupRoleFilters) > 0 {
				roleFilters = append(roleFilters, groupRoleFilters...)
			}
		}
		if user.RoleFilters != nil && len(user.RoleFilters) > 0 {
			roleFilters = append(roleFilters, user.RoleFilters...)
		}
		if len(roleFilters) > 0 {
			for _, filter := range roleFilters {
				if len(filter.Team) > 0 {
					if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, filter.Team); ok {
						isAuthorised = true
						break
					}
				}
				if filter.Entity == bean2.CLUSTER_ENTITIY {
					if ok := handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth); ok {
						isAuthorised = true
						break
					}
				}
			}
		}
	}
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	res, err := handler.userService.GetAll()
	if err != nil {
		handler.logger.Errorw("service err, GetAll", "err", err)
		common.WriteJsonResp(w, err, "Failed to Get", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) GetAllDetailedUsers(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	isActionUserSuperAdmin := false
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	if !isActionUserSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	res, err := handler.userService.GetAllDetailedUsers()
	if err != nil {
		handler.logger.Errorw("service err, GetAllDetailedUsers", "err", err)
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
			if filter.AccessType == bean2.APP_ACCESS_TYPE_HELM && !isActionUserSuperAdmin {
				common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
				return
			}
			if len(filter.Team) > 0 {
				if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionDelete, filter.Team); !ok {
					common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
					return
				}
			}
			if filter.Entity == bean2.CLUSTER_ENTITIY {
				if ok := handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth); !ok {
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
	//validation
	validated := helper.IsSystemOrAdminUser(int32(id))
	if validated {
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "cannot delete system or admin user"}
		handler.logger.Errorw("request err, DeleteUser, validation failed", "id", id, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//service call
	res, err := handler.userService.DeleteUser(user)
	if err != nil {
		handler.logger.Errorw("service err, DeleteUser", "err", err, "id", id)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) BulkDeleteUsers(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	// request decoding
	var request *bean.BulkDeleteRequest
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, BulkDeleteUsers", "payload", request, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, BulkDeleteUsers", "payload", request)
	// setting logged in user Id for audit logs
	request.LoggedInUserId = userId

	// validations for system and admin user
	err = helper.CheckValidationForAdminAndSystemUserId(request.Ids)
	if err != nil {
		handler.logger.Errorw("request err, BulkDeleteUsers, validation failed", "payload", request, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	// struct validation
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err, BulkDeleteUsers", "payload", request, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// service call
	res, err := handler.userService.BulkDeleteUsers(request)
	if err != nil {
		handler.logger.Errorw("service err, BulkDeleteUsers", "payload", request, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
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
	filteredRoleFilter := make([]bean.RoleFilter, 0)
	if res.RoleFilters != nil && len(res.RoleFilters) > 0 {
		isUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
		for _, filter := range res.RoleFilters {
			authPass := handler.checkRbacForFilter(token, filter, isUserSuperAdmin)
			if authPass {
				filteredRoleFilter = append(filteredRoleFilter, filter)
			}
		}
	}
	for index, roleFilter := range filteredRoleFilter {
		if roleFilter.Entity == "" {
			filteredRoleFilter[index].Entity = bean2.ENTITY_APPS
		}
		if roleFilter.Entity == bean2.ENTITY_APPS && roleFilter.AccessType == "" {
			filteredRoleFilter[index].AccessType = bean2.DEVTRON_APP
		}
	}

	res.RoleFilters = filteredRoleFilter
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) checkRbacForFilter(token string, filter bean.RoleFilter, isUserSuperAdmin bool) bool {
	isAuthorised := true
	switch {
	case isUserSuperAdmin:
		isAuthorised = true
	case filter.AccessType == bean2.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			isAuthorised = false
		}

	case len(filter.Team) > 0:
		// this is case of devtron app
		if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, filter.Team); !ok {
			isAuthorised = false
		}

	case filter.Entity == bean.CLUSTER_ENTITIY:
		isValidAuth := handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
		if !isValidAuth {
			isAuthorised = false
		}
	case filter.Entity == bean.CHART_GROUP_ENTITY:
		isAuthorised = true
	default:
		isAuthorised = false
	}
	return isAuthorised
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
	isAuthorised, err := handler.checkRBACForUserCreate(token, request.SuperAdmin, request.RoleFilters, nil)
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}
	if !isAuthorised {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
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
		} else if err.Error() == bean2.VALIDATION_FAILED_ERROR_MSG {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
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
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err, UpdateRoleGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	isActionUserSuperAdmin := false
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	//Check if user is not superAdmin and updating to userAdmin
	if request.SuperAdmin && !isActionUserSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	res, err := handler.roleGroupService.UpdateRoleGroup(&request, token, handler.checkRBACForRoleGroupUpdate, handler.CheckManagerAuth)
	if err != nil {
		handler.logger.Errorw("service err, UpdateRoleGroup", "err", err, "payload", request)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler UserRestHandlerImpl) FetchRoleGroupsV2(w http.ResponseWriter, r *http.Request) {
	var decoder = schema.NewDecoder()
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	//checking superAdmin access
	isAuthorised := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if !isAuthorised {
		user, err := handler.userService.GetById(userId)
		if err != nil {
			handler.logger.Errorw("error in getting user by id", "err", err)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
			return
		}
		var roleFilters []bean.RoleFilter
		if len(user.UserRoleGroup) > 0 {
			groupRoleFilters, err := handler.userService.GetRoleFiltersByUserRoleGroups(user.UserRoleGroup)
			if err != nil {
				handler.logger.Errorw("Error in getting role filters by group names", "err", err, "UserRoleGroup", user.UserRoleGroup)
				common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
				return
			}
			if len(groupRoleFilters) > 0 {
				roleFilters = append(roleFilters, groupRoleFilters...)
			}
		}
		if user.RoleFilters != nil && len(user.RoleFilters) > 0 {
			roleFilters = append(roleFilters, user.RoleFilters...)
		}
		if len(roleFilters) > 0 {
			for _, filter := range roleFilters {
				if len(filter.Team) > 0 {
					if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, filter.Team); ok {
						isAuthorised = true
						break
					}
				}
				if filter.Entity == bean2.CLUSTER_ENTITIY {
					if isValidAuth := handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth); isValidAuth {
						isAuthorised = true
						break
					}
				}

			}
		}
	}
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	req := &bean.ListingRequest{}
	err = decoder.Decode(req, r.URL.Query())
	if err != nil {
		handler.logger.Errorw("request err, FetchRoleGroups", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := handler.roleGroupService.FetchRoleGroupsWithFilters(req)
	if err != nil {
		handler.logger.Errorw("service err, FetchRoleGroups", "err", err)
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
	// RBAC enforcer applying
	token := r.Header.Get("token")
	//checking superAdmin access
	isAuthorised := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if !isAuthorised {
		user, err := handler.userService.GetById(userId)
		if err != nil {
			handler.logger.Errorw("error in getting user by id", "err", err)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
			return
		}
		var roleFilters []bean.RoleFilter
		if len(user.UserRoleGroup) > 0 {
			groupRoleFilters, err := handler.userService.GetRoleFiltersByUserRoleGroups(user.UserRoleGroup)
			if err != nil {
				handler.logger.Errorw("Error in getting role filters by group names", "err", err, "UserRoleGroup", user.UserRoleGroup)
				common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
				return
			}
			if len(groupRoleFilters) > 0 {
				roleFilters = append(roleFilters, groupRoleFilters...)
			}
		}
		if user.RoleFilters != nil && len(user.RoleFilters) > 0 {
			roleFilters = append(roleFilters, user.RoleFilters...)
		}
		if len(roleFilters) > 0 {
			for _, filter := range roleFilters {
				if len(filter.Team) > 0 {
					if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, filter.Team); ok {
						isAuthorised = true
						break
					}
				}
				if filter.Entity == bean2.CLUSTER_ENTITIY {
					if isValidAuth := handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth); isValidAuth {
						isAuthorised = true
						break
					}
				}

			}
		}
	}
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
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

func (handler UserRestHandlerImpl) FetchDetailedRoleGroups(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	isActionUserSuperAdmin := false
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	if !isActionUserSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	req := &bean.ListingRequest{ShowAll: true}
	res, err := handler.roleGroupService.FetchDetailedRoleGroups(req)
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
	isAuthorised, err := handler.checkRBACForRoleGroupDelete(token, userGroup.RoleFilters)
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}
	if !isAuthorised {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
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

func (handler UserRestHandlerImpl) BulkDeleteRoleGroups(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	// request decoding
	var request *bean.BulkDeleteRequest
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, BulkDeleteRoleGroups", "payload", request, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, BulkDeleteRoleGroups", "payload", request)
	// setting logged in user Id for audit logs
	request.LoggedInUserId = userId

	// struct validation
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err, BulkDeleteRoleGroups", "payload", request, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// service call
	res, err := handler.roleGroupService.BulkDeleteRoleGroups(request)
	if err != nil {
		handler.logger.Errorw("service err, BulkDeleteRoleGroups", "payload", request, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
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

	v := r.URL.Query()

	if v.Has("appName") {
		appName := v.Get("appName")
		result := make(map[string]interface{})
		var isSuperAdmin, isAdmin, isManager, isTrigger bool
		for _, role := range roles {
			if role == bean.SUPERADMIN {
				isSuperAdmin = true
				break
			}
			frags := strings.Split(role, "_")
			n := len(frags)
			if n >= 2 && (frags[n-1] == appName || frags[n-1] == "") {
				isManager = strings.Contains(frags[0], "manager")
				isAdmin = strings.Contains(frags[0], "admin")
				isTrigger = strings.Contains(frags[0], "trigger")
			}
		}
		if isSuperAdmin {
			result["role"] = "SuperAdmin"
		} else if isManager {
			result["role"] = "Manager"
		} else if isAdmin {
			result["role"] = "Admin"
		} else if isTrigger {
			result["role"] = "Trigger"
		} else {
			result["role"] = "View"
		}
		result["roles"] = roles

		common.WriteJsonResp(w, err, result, http.StatusOK)
		return
	}
	result := make(map[string]interface{})
	result["roles"] = roles
	result["superAdmin"] = false
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
	userEmailId, err := handler.userService.GetActiveEmailById(userId)
	if err != nil {
		handler.logger.Errorw("service err, SyncOrchestratorToCasbin", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if userEmailId != "admin" {
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

func (handler UserRestHandlerImpl) GetRoleCacheDump(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	if isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	} else {
		cacheDump := handler.enforcer.GetCacheDump()
		common.WriteJsonResp(w, nil, cacheDump, http.StatusOK)
	}
}

func (handler UserRestHandlerImpl) InvalidateRoleCache(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	if isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	} else {
		handler.enforcer.InvalidateCompleteCache()
		common.WriteJsonResp(w, nil, "Cache Cleaned Successfully", http.StatusOK)
	}
}

func (handler UserRestHandlerImpl) CheckManagerAuth(resource, token string, object string) bool {
	if ok := handler.enforcer.Enforce(token, resource, casbin.ActionUpdate, object); !ok {
		return false
	}
	return true

}

func (handler UserRestHandlerImpl) checkRBACForUserCreate(token string, requestSuperAdmin bool, roleFilters []bean.RoleFilter,
	roleGroups []bean.UserRoleGroup) (isAuthorised bool, err error) {
	isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if requestSuperAdmin && !isActionUserSuperAdmin {
		return false, nil
	}
	isAuthorised = isActionUserSuperAdmin
	if !isAuthorised {
		if roleFilters != nil && len(roleFilters) > 0 { //auth check inside roleFilters
			for _, filter := range roleFilters {
				switch {
				case filter.AccessType == bean.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
					isAuthorised = isActionUserSuperAdmin
				case len(filter.Team) > 0:
					isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, filter.Team)
				case filter.Entity == bean.CLUSTER_ENTITIY:
					isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
				case filter.Entity == bean.CHART_GROUP_ENTITY && len(roleFilters) == 1: //if only chartGroup entity is present in request then access will be judged through super-admin access
					isAuthorised = isActionUserSuperAdmin
				case filter.Entity == bean.CHART_GROUP_ENTITY && len(roleFilters) > 1: //if entities apart from chartGroup entity are present, not checking chartGroup access
					isAuthorised = true
				default:
					isAuthorised = false
				}
				if !isAuthorised {
					return false, nil
				}
			}
		}
		if len(roleGroups) > 0 { // auth check inside groups
			groupRoles, err := handler.roleGroupService.FetchRolesForUserRoleGroups(roleGroups)
			if err != nil && err != pg.ErrNoRows {
				handler.logger.Errorw("service err, UpdateUser", "err", err, "payload", roleGroups)
				return false, err
			}
			if len(groupRoles) > 0 {
				for _, groupRole := range groupRoles {
					switch {
					case groupRole.Action == bean.ACTION_SUPERADMIN:
						isAuthorised = isActionUserSuperAdmin
					case groupRole.AccessType == bean.APP_ACCESS_TYPE_HELM || groupRole.Entity == bean2.EntityJobs:
						isAuthorised = isActionUserSuperAdmin
					case len(groupRole.Team) > 0:
						isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, groupRole.Team)
					case groupRole.Entity == bean.CLUSTER_ENTITIY:
						isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(groupRole.Cluster, groupRole.Namespace, groupRole.Group, groupRole.Kind, groupRole.Resource, token, handler.CheckManagerAuth)
					case groupRole.Entity == bean.CHART_GROUP_ENTITY && len(groupRoles) == 1: //if only chartGroup entity is present in request then access will be judged through super-admin access
						isAuthorised = isActionUserSuperAdmin
					case groupRole.Entity == bean.CHART_GROUP_ENTITY && len(groupRoles) > 1: //if entities apart from chartGroup entity are present, not checking chartGroup access
						isAuthorised = true
					default:
						isAuthorised = false
					}
					if !isAuthorised {
						return false, nil
					}
				}
			} else {
				isAuthorised = false
			}
		}
	}
	return isAuthorised, nil
}

func (handler UserRestHandlerImpl) checkRBACForUserUpdate(token string, userInfo *bean.UserInfo, isUserAlreadySuperAdmin bool, eliminatedRoleFilters,
	eliminatedGroupRoles []*repository.RoleModel) (isAuthorised bool, err error) {
	isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	requestSuperAdmin := userInfo.SuperAdmin
	if (requestSuperAdmin || isUserAlreadySuperAdmin) && !isActionUserSuperAdmin {
		//if user is going to be provided with super-admin access or already a super-admin then the action user should be a super-admin
		return false, nil
	}
	roleFilters := userInfo.RoleFilters
	roleGroups := userInfo.UserRoleGroup
	isAuthorised = isActionUserSuperAdmin
	eliminatedRolesToBeChecked := append(eliminatedRoleFilters, eliminatedGroupRoles...)
	if !isAuthorised {
		if roleFilters != nil && len(roleFilters) > 0 { //auth check inside roleFilters
			for _, filter := range roleFilters {
				switch {
				case filter.AccessType == bean.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
					isAuthorised = isActionUserSuperAdmin
				case len(filter.Team) > 0:
					isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, filter.Team)
				case filter.Entity == bean.CLUSTER_ENTITIY:
					isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
				case filter.Entity == bean.CHART_GROUP_ENTITY:
					isAuthorised = true
				default:
					isAuthorised = false
				}
				if !isAuthorised {
					return false, nil
				}
			}
		}
		if eliminatedRolesToBeChecked != nil && len(eliminatedRolesToBeChecked) > 0 {
			for _, filter := range eliminatedRolesToBeChecked {
				switch {
				case filter.AccessType == bean.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
					isAuthorised = isActionUserSuperAdmin
				case len(filter.Team) > 0:
					isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, filter.Team)
				case filter.Entity == bean.CLUSTER_ENTITIY:
					isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
				case filter.Entity == bean.CHART_GROUP_ENTITY:
					isAuthorised = true
				default:
					isAuthorised = false
				}
				if !isAuthorised {
					return false, nil
				}
			}
		}
		if len(roleGroups) > 0 { // auth check inside groups
			groupRoles, err := handler.roleGroupService.FetchRolesForUserRoleGroups(roleGroups)
			if err != nil && err != pg.ErrNoRows {
				handler.logger.Errorw("service err, UpdateUser", "err", err, "payload", roleGroups)
				return false, err
			}
			if len(groupRoles) > 0 {
				for _, groupRole := range groupRoles {
					switch {
					case groupRole.Action == bean.ACTION_SUPERADMIN:
						isAuthorised = isActionUserSuperAdmin
					case groupRole.AccessType == bean.APP_ACCESS_TYPE_HELM || groupRole.Entity == bean2.EntityJobs:
						isAuthorised = isActionUserSuperAdmin
					case len(groupRole.Team) > 0:
						isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, groupRole.Team)
					case groupRole.Entity == bean.CLUSTER_ENTITIY:
						isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(groupRole.Cluster, groupRole.Namespace, groupRole.Group, groupRole.Kind, groupRole.Resource, token, handler.CheckManagerAuth)
					case groupRole.Entity == bean.CHART_GROUP_ENTITY:
						isAuthorised = true
					default:
						isAuthorised = false
					}
					if !isAuthorised {
						return false, nil
					}
				}
			} else {
				isAuthorised = false
			}
		}
	}
	return isAuthorised, nil
}

func (handler UserRestHandlerImpl) checkRBACForRoleGroupUpdate(token string, groupInfo *bean.RoleGroup,
	eliminatedRoleFilters []*repository.RoleModel) (isAuthorised bool, err error) {
	isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	requestSuperAdmin := groupInfo.SuperAdmin
	if requestSuperAdmin && !isActionUserSuperAdmin {
		//if user is going to be provided with super-admin access or already a super-admin then the action user should be a super-admin
		return false, nil
	}
	isAuthorised = isActionUserSuperAdmin
	if !isAuthorised {
		if groupInfo.RoleFilters != nil && len(groupInfo.RoleFilters) > 0 { //auth check inside roleFilters
			for _, filter := range groupInfo.RoleFilters {
				switch {
				case filter.Action == bean.ACTION_SUPERADMIN:
					isAuthorised = isActionUserSuperAdmin
				case filter.AccessType == bean.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
					isAuthorised = isActionUserSuperAdmin
				case len(filter.Team) > 0:
					isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, filter.Team)
				case filter.Entity == bean.CLUSTER_ENTITIY:
					isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
				case filter.Entity == bean.CHART_GROUP_ENTITY:
					isAuthorised = true
				default:
					isAuthorised = false
				}
				if !isAuthorised {
					return false, nil
				}
			}
		}
		if len(eliminatedRoleFilters) > 0 {
			for _, filter := range eliminatedRoleFilters {
				switch {
				case filter.AccessType == bean.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
					isAuthorised = isActionUserSuperAdmin
				case len(filter.Team) > 0:
					isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, filter.Team)
				case filter.Entity == bean.CLUSTER_ENTITIY:
					isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
				case filter.Entity == bean.CHART_GROUP_ENTITY:
					isAuthorised = true
				default:
					isAuthorised = false
				}
				if !isAuthorised {
					return false, nil
				}
			}
		}
	}
	return isAuthorised, nil
}

func (handler UserRestHandlerImpl) checkRBACForRoleGroupDelete(token string, groupRoles []bean.RoleFilter) (isAuthorised bool, err error) {
	isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	isAuthorised = isActionUserSuperAdmin
	if !isAuthorised {
		if groupRoles != nil && len(groupRoles) > 0 { //auth check inside roleFilters
			for _, filter := range groupRoles {
				switch {
				case filter.Action == bean.ACTION_SUPERADMIN:
					isAuthorised = isActionUserSuperAdmin
				case filter.AccessType == bean.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
					isAuthorised = isActionUserSuperAdmin
				case len(filter.Team) > 0:
					isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, filter.Team)
				case filter.Entity == bean.CLUSTER_ENTITIY:
					isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
				case filter.Entity == bean.CHART_GROUP_ENTITY:
					isAuthorised = true
				default:
					isAuthorised = false
				}
				if !isAuthorised {
					return false, nil
				}
			}
		}
	}
	return isAuthorised, nil
}
