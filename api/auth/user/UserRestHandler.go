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
	"github.com/devtron-labs/devtron/util/commonEnforcementFunctionsUtil"
	"github.com/gorilla/schema"
	"net/http"
	"strconv"
	"strings"

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
	userService         user2.UserService
	validator           *validator.Validate
	logger              *zap.SugaredLogger
	enforcer            casbin.Enforcer
	roleGroupService    user2.RoleGroupService
	userCommonService   user2.UserCommonService
	rbacEnforcementUtil commonEnforcementFunctionsUtil.CommonEnforcementUtil
}

func NewUserRestHandlerImpl(userService user2.UserService, validator *validator.Validate,
	logger *zap.SugaredLogger, enforcer casbin.Enforcer, roleGroupService user2.RoleGroupService,
	userCommonService user2.UserCommonService,
	rbacEnforcementUtil commonEnforcementFunctionsUtil.CommonEnforcementUtil) *UserRestHandlerImpl {
	userAuthHandler := &UserRestHandlerImpl{
		userService:         userService,
		validator:           validator,
		logger:              logger,
		enforcer:            enforcer,
		roleGroupService:    roleGroupService,
		userCommonService:   userCommonService,
		rbacEnforcementUtil: rbacEnforcementUtil,
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
	var userInfo bean2.UserInfo
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
	var userInfo bean2.UserInfo
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
	res, err := handler.userService.GetByIdWithoutGroupClaims(int32(id))
	if err != nil {
		handler.logger.Errorw("service err, GetById", "err", err, "id", id)
		common.WriteJsonResp(w, err, "Failed to get by id", http.StatusInternalServerError)
		return
	}

	token := r.Header.Get("token")
	// NOTE: if no role assigned, user will be visible to all manager.
	// RBAC enforcer applying
	filteredRoleFilter := handler.GetFilteredRoleFiltersAccordingToAccess(token, res.RoleFilters)
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
	isAuthorised, err := handler.rbacEnforcementUtil.CheckRbacForMangerAndAboveAccess(token, userId)
	if err != nil {
		handler.logger.Errorw("err, CheckRbacForMangerAndAboveAccess", "err", err)
		common.WriteJsonResp(w, err, "Failed to Get", http.StatusInternalServerError)
		return
	}
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	req := &bean2.ListingRequest{}
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
	isAuthorised, err := handler.rbacEnforcementUtil.CheckRbacForMangerAndAboveAccess(token, userId)
	if err != nil {
		handler.logger.Errorw("err, CheckRbacForMangerAndAboveAccess", "err", err)
		common.WriteJsonResp(w, err, "Failed to Get", http.StatusInternalServerError)
		return
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
	user, err := handler.userService.GetByIdWithoutGroupClaims(int32(id))
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	isAuthorised := handler.CheckRbacForUserDelete(token, user)
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
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
	var request *bean2.BulkDeleteRequest
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
	filteredRoleFilter := handler.GetFilteredRoleFiltersAccordingToAccess(token, res.RoleFilters)

	res.RoleFilters = filteredRoleFilter
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
	var request bean2.RoleGroup
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
	var request bean2.RoleGroup
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
	isAuthorised, err := handler.rbacEnforcementUtil.CheckRbacForMangerAndAboveAccess(token, userId)
	if err != nil {
		handler.logger.Errorw("err, CheckRbacForMangerAndAboveAccess", "err", err)
		common.WriteJsonResp(w, err, "Failed to Get", http.StatusInternalServerError)
		return
	}
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	req := &bean2.ListingRequest{}
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
	isAuthorised, err := handler.rbacEnforcementUtil.CheckRbacForMangerAndAboveAccess(token, userId)
	if err != nil {
		handler.logger.Errorw("err, CheckRbacForMangerAndAboveAccess", "err", err)
		common.WriteJsonResp(w, err, "Failed to Get", http.StatusInternalServerError)
		return
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
	req := &bean2.ListingRequest{ShowAll: true}
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
	isAuthorised, err := handler.checkRBACForRoleGroupDelete(token, userGroup)
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
	var request *bean2.BulkDeleteRequest
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
			if role == bean2.SUPERADMIN {
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
		if item == bean2.SUPERADMIN {
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
