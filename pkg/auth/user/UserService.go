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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	auth "github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository/helper"
	util3 "github.com/devtron-labs/devtron/pkg/auth/user/util"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow"
	repository2 "github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	jwt2 "github.com/golang-jwt/jwt/v4"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/devtron-labs/authenticator/jwt"
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"

	casbin2 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

const (
	ConcurrentRequestLockError   = "there is an ongoing request for this user, please try after some time"
	ConcurrentRequestUnlockError = "cannot block request that is not in process"
)

type UserService interface {
	CreateUser(userInfo *bean.UserInfo, token string, managerAuth func(resource, token string, object string) bool) ([]*bean.UserInfo, error)
	SelfRegisterUserIfNotExists(userInfo *bean.UserInfo, groupsFromClaims []string, groupClaimsConfigActive bool) ([]*bean.UserInfo, error)
	UpdateUserGroupMappingIfActiveUser(emailId string, groups []string) error
	UpdateUser(userInfo *bean.UserInfo, token string, managerAuth func(resource, token string, object string) bool) (*bean.UserInfo, bool, bool, []string, error)
	GetByIdWithoutGroupClaims(id int32) (*bean.UserInfo, error)
	GetRoleFiltersForAUserById(id int32) (*bean.UserInfo, error)
	GetByIdForGroupClaims(id int32) (*bean.UserInfo, error)
	GetAll() ([]bean.UserInfo, error)
	GetAllWithFilters(status string, sortOrder string, sortBy string, offset int, totalSize int, showAll bool, searchKey string) (*bean.UserListingResponse, error)
	GetAllDetailedUsers() ([]bean.UserInfo, error)
	GetEmailById(userId int32) (string, error)
	GetEmailAndGroupClaimsFromToken(token string) (string, []string, error)
	GetLoggedInUser(r *http.Request) (int32, error)
	GetByIds(ids []int32) ([]bean.UserInfo, error)
	DeleteUser(userInfo *bean.UserInfo) (bool, error)
	CheckUserRoles(id int32, token string) ([]string, error)
	SyncOrchestratorToCasbin() (bool, error)
	GetUserByToken(context context.Context, token string) (int32, string, error)
	IsSuperAdminForDevtronManaged(userId int) (bool, error)
	GetByIdIncludeDeleted(id int32) (*bean.UserInfo, error)
	UserExists(emailId string) bool
	GetRoleFiltersByGroupNames(groupNames []string) ([]bean.RoleFilter, error)
	GetRoleFiltersByGroupCasbinNames(groupNames []string) ([]bean.RoleFilter, error)
	SaveLoginAudit(emailId, clientIp string, id int32)
	GetApprovalUsersByEnv(appName, envName string) ([]string, error)
	CheckForApproverAccess(appName, envName string, userId int32) bool
	GetConfigApprovalUsersByEnv(appName, envName, team string) ([]string, error)
	GetFieldValuesFromToken(token string) ([]byte, error)
	BulkUpdateStatusForUsers(request *bean.BulkStatusUpdateRequest) (*bean.ActionResponse, error)
	GetUserWithTimeoutWindowConfiguration(emailId string) (int32, bool, error)
}

type UserServiceImpl struct {
	userReqLock sync.RWMutex
	//map of userId and current lock-state of their serving ability;
	//if TRUE then it means that some request is ongoing & unable to serve and FALSE then it is open to serve
	userReqState                      map[int32]bool
	userAuthRepository                repository.UserAuthRepository
	logger                            *zap.SugaredLogger
	userRepository                    repository.UserRepository
	roleGroupRepository               repository.RoleGroupRepository
	sessionManager2                   *middleware.SessionManager
	userCommonService                 UserCommonService
	userAuditService                  UserAuditService
	globalAuthorisationConfigService  auth.GlobalAuthorisationConfigService
	roleGroupService                  RoleGroupService
	userGroupMapRepository            repository.UserGroupMapRepository
	userListingRepositoryQueryBuilder helper.UserRepositoryQueryBuilder
	timeoutWindowService              timeoutWindow.TimeoutWindowService
}

func NewUserServiceImpl(userAuthRepository repository.UserAuthRepository,
	logger *zap.SugaredLogger,
	userRepository repository.UserRepository,
	userGroupRepository repository.RoleGroupRepository,
	sessionManager2 *middleware.SessionManager, userCommonService UserCommonService, userAuditService UserAuditService,
	globalAuthorisationConfigService auth.GlobalAuthorisationConfigService,
	roleGroupService RoleGroupService, userGroupMapRepository repository.UserGroupMapRepository,
	userListingRepositoryQueryBuilder helper.UserRepositoryQueryBuilder,
	timeoutWindowService timeoutWindow.TimeoutWindowService,
) *UserServiceImpl {
	serviceImpl := &UserServiceImpl{
		userReqState:                      make(map[int32]bool),
		userAuthRepository:                userAuthRepository,
		logger:                            logger,
		userRepository:                    userRepository,
		roleGroupRepository:               userGroupRepository,
		sessionManager2:                   sessionManager2,
		userCommonService:                 userCommonService,
		userAuditService:                  userAuditService,
		globalAuthorisationConfigService:  globalAuthorisationConfigService,
		roleGroupService:                  roleGroupService,
		userGroupMapRepository:            userGroupMapRepository,
		userListingRepositoryQueryBuilder: userListingRepositoryQueryBuilder,
		timeoutWindowService:              timeoutWindowService,
	}
	cStore = sessions.NewCookieStore(randKey())
	return serviceImpl
}

func (impl *UserServiceImpl) getUserReqLockStateById(userId int32) bool {
	defer impl.userReqLock.RUnlock()
	impl.userReqLock.RLock()
	return impl.userReqState[userId]
}

// FreeUnfreeUserReqState - free sets the userId free for serving, meaning removing the lock(removing entry). Unfree locks the user for other requests
func (impl *UserServiceImpl) lockUnlockUserReqState(userId int32, lock bool) error {
	var err error
	defer impl.userReqLock.Unlock()
	impl.userReqLock.Lock()
	if lock {
		//checking again if someone changed or not
		if !impl.userReqState[userId] {
			//available to serve, locking
			impl.userReqState[userId] = true
		} else {
			err = &util.ApiError{Code: "409", HttpStatusCode: http.StatusConflict, UserMessage: ConcurrentRequestLockError}
		}
	} else {
		if impl.userReqState[userId] {
			//in serving state, unlocking
			delete(impl.userReqState, userId)
		} else {
			err = &util.ApiError{Code: "409", HttpStatusCode: http.StatusConflict, UserMessage: ConcurrentRequestUnlockError}
		}
	}
	return err
}

func (impl UserServiceImpl) validateUserRequest(userInfo *bean.UserInfo) (bool, error) {
	if len(userInfo.RoleFilters) == 1 &&
		userInfo.RoleFilters[0].Team == "" && userInfo.RoleFilters[0].Environment == "" && userInfo.RoleFilters[0].Action == "" {
		//skip
	} else {
		invalid := false
		for _, roleFilter := range userInfo.RoleFilters {
			if len(roleFilter.Team) > 0 && len(roleFilter.Action) > 0 {
				//
			} else if len(roleFilter.Entity) > 0 { //this will pass roleFilter for clusterEntity as well as chart-group
				//
			} else {
				invalid = true
			}
		}
		if invalid {
			err := &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, please provide role filters"}
			return false, err
		}
	}
	return true, nil
}

func (impl UserServiceImpl) SelfRegisterUserIfNotExists(userInfo *bean.UserInfo, groupsFromClaims []string, groupClaimsConfigActive bool) ([]*bean.UserInfo, error) {
	var userResponse []*bean.UserInfo
	emailIds := strings.Split(userInfo.EmailId, ",")
	dbConnection := impl.userRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	var policies []casbin2.Policy
	for _, emailId := range emailIds {
		dbUser, err := impl.userRepository.FetchActiveOrDeletedUserByEmail(emailId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error while fetching user from db", "error", err)
			return nil, err
		}

		//if found, update it with new roles
		if dbUser != nil && dbUser.Id > 0 {
			return nil, fmt.Errorf("existing user, cant self register")
		}

		// if not found, create new user
		userInfo, err = impl.saveUser(userInfo, emailId)
		if err != nil {
			err = &util.ApiError{
				Code:            constants.UserCreateDBFailed,
				InternalMessage: "failed to create new user in db",
				UserMessage:     fmt.Sprintf("requested by %d", userInfo.UserId),
			}
			return nil, err
		}
		if groupClaimsConfigActive {
			err = impl.updateDataForUserGroupClaimsMap(userInfo.Id, groupsFromClaims)
			if err != nil {
				impl.logger.Errorw("error in updating data for user group claims map", "err", err, "userId", userInfo.Id)
				return nil, err
			}
		}

		if len(userInfo.Roles) > 0 { //checking this because in self registration service we have set roles as per requirement, like in group claims type auth no need of roles
			roles, err := impl.userAuthRepository.GetRoleByRoles(userInfo.Roles)
			if err != nil {
				err = &util.ApiError{
					Code:            constants.UserCreateDBFailed,
					InternalMessage: "configured roles for selfregister are wrong",
					UserMessage:     fmt.Sprintf("requested by %d", userInfo.UserId),
				}
				return nil, err
			}
			for _, roleModel := range roles {
				userRoleModel := &repository.UserRoleModel{UserId: userInfo.Id, RoleId: roleModel.Id}
				userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
				if err != nil {
					return nil, err
				}
				policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(userInfo.EmailId), Obj: casbin2.Object(roleModel.Role)})
			}
		}
		userInfo.EmailId = emailId
		userInfo.Exist = dbUser.Active
		userResponse = append(userResponse, &bean.UserInfo{Id: userInfo.Id, EmailId: emailId, Groups: userInfo.Groups, RoleFilters: userInfo.RoleFilters, SuperAdmin: userInfo.SuperAdmin})
	}

	if len(policies) > 0 {
		//loading policy for safety
		casbin2.LoadPolicy()
		err = casbin2.AddPolicy(policies)
		if err != nil {
			impl.logger.Errorw("casbin policy addition failed", "err", err)
			return nil, err
		}
		//loading policy for syncing orchestrator to casbin with newly added policies
		casbin2.LoadPolicy()
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return userResponse, nil
}

func (impl UserServiceImpl) UpdateUserGroupMappingIfActiveUser(emailId string, groups []string) error {
	user, err := impl.userRepository.FetchActiveUserByEmail(emailId)
	if err != nil {
		impl.logger.Errorw("error in getting active user by email", "err", err, "emailId", emailId)
		return err
	}
	err = impl.updateDataForUserGroupClaimsMap(user.Id, groups)
	if err != nil {
		impl.logger.Errorw("error in updating data for user group claims map", "err", err, "userId", user.Id)
		return err
	}
	return nil
}

func (impl UserServiceImpl) updateDataForUserGroupClaimsMap(userId int32, groups []string) error {
	//updating groups received in claims
	mapOfGroups := make(map[string]bool, len(groups))
	for _, group := range groups {
		mapOfGroups[group] = true
	}
	groupMappings, err := impl.userGroupMapRepository.GetByUserId(userId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting user group mapping by userId", "err", err, "userId", userId)
		return err
	}
	dbConnection := impl.userGroupMapRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating transaction", "err", err)
		return err
	}
	defer tx.Rollback()
	modelsToBeSaved := make([]*repository.UserGroup, 0)
	modelsToBeUpdated := make([]*repository.UserGroup, 0)
	timeNow := time.Now()
	for i := range groupMappings {
		groupMapping := groupMappings[i]
		//checking if mapping present in groups from claims
		if _, ok := mapOfGroups[groupMapping.GroupName]; ok {
			//present so marking active flag true
			groupMapping.Active = true
			//deleting entry from map now
			delete(mapOfGroups, groupMapping.GroupName)
		} else {
			//not present so marking active flag false
			groupMapping.Active = false
		}
		groupMapping.UpdatedOn = timeNow
		groupMapping.UpdatedBy = 1 //system user

		//adding this group mapping to updated models irrespective of active
		modelsToBeUpdated = append(modelsToBeUpdated, groupMapping)
	}

	//iterating through remaining groups from the map, they are not found in current entries so need to be saved
	for group := range mapOfGroups {
		modelsToBeSaved = append(modelsToBeSaved, &repository.UserGroup{
			UserId:            userId,
			GroupName:         group,
			IsGroupClaimsData: true,
			Active:            true,
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: timeNow,
				UpdatedBy: 1,
				UpdatedOn: timeNow,
			},
		})
	}
	if len(modelsToBeUpdated) > 0 {
		err = impl.userGroupMapRepository.Update(modelsToBeUpdated, tx)
		if err != nil {
			impl.logger.Errorw("error in updating user group mapping", "err", err)
			return err
		}
	}
	if len(modelsToBeSaved) > 0 {
		err = impl.userGroupMapRepository.Save(modelsToBeSaved, tx)
		if err != nil {
			impl.logger.Errorw("error in saving user group mapping", "err", err)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return err
	}
	return nil
}

func (impl UserServiceImpl) saveUser(userInfo *bean.UserInfo, emailId string) (*bean.UserInfo, error) {
	dbConnection := impl.userRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	_, err = impl.validateUserRequest(userInfo)
	if err != nil {
		err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, please provide role filters"}
		return nil, err
	}

	//create new user in our db on d basis of info got from google api or hex. assign a basic role
	model := &repository.UserModel{
		EmailId:     emailId,
		AccessToken: userInfo.AccessToken,
	}
	model.Active = true
	model.CreatedBy = userInfo.UserId
	model.UpdatedBy = userInfo.UserId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	model, err = impl.userRepository.CreateUser(model, tx)
	if err != nil {
		impl.logger.Errorw("error in creating new user", "error", err)
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	userInfo.Id = model.Id
	return userInfo, nil
}

func (impl UserServiceImpl) CreateUser(userInfo *bean.UserInfo, token string, managerAuth func(resource, token string, object string) bool) ([]*bean.UserInfo, error) {

	var pass []string
	var userResponse []*bean.UserInfo
	emailIds := strings.Split(userInfo.EmailId, ",")
	for _, emailId := range emailIds {
		dbUser, err := impl.userRepository.FetchActiveOrDeletedUserByEmail(emailId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error while fetching user from db", "error", err)
			return nil, err
		}

		//if found, update it with new roles
		if dbUser != nil && dbUser.Id > 0 {
			userInfo, err = impl.updateUserIfExists(userInfo, dbUser, emailId, token, managerAuth)
			if err != nil {
				impl.logger.Errorw("error while create user if exists in db", "error", err)
				return nil, err
			}
		}

		// if not found, create new user
		if err == pg.ErrNoRows {
			userInfo, err = impl.createUserIfNotExists(userInfo, emailId, token, managerAuth)
			if err != nil {
				impl.logger.Errorw("error while create user if not exists in db", "error", err)
				return nil, err
			}
		}

		pass = append(pass, emailId)
		userInfo.EmailId = emailId
		userInfo.Exist = dbUser.Active
		userResponse = append(userResponse, &bean.UserInfo{Id: userInfo.Id, EmailId: emailId, Groups: userInfo.Groups, RoleFilters: userInfo.RoleFilters, SuperAdmin: userInfo.SuperAdmin})
	}

	return userResponse, nil
}

func (impl UserServiceImpl) updateUserIfExists(userInfo *bean.UserInfo, dbUser *repository.UserModel, emailId string,
	token string, managerAuth func(resource, token, object string) bool) (*bean.UserInfo, error) {
	updateUserInfo, err := impl.GetByIdWithoutGroupClaims(dbUser.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	if dbUser.Active == false {
		updateUserInfo = &bean.UserInfo{Id: dbUser.Id}
		userInfo.Id = dbUser.Id
		updateUserInfo.SuperAdmin = userInfo.SuperAdmin
	}
	updateUserInfo.RoleFilters = impl.mergeRoleFilter(updateUserInfo.RoleFilters, userInfo.RoleFilters)
	updateUserInfo.Groups = impl.mergeGroups(updateUserInfo.Groups, userInfo.Groups)
	updateUserInfo.UserId = userInfo.UserId
	updateUserInfo.EmailId = emailId // override case sensitivity
	impl.logger.Debugw("update user called through create user flow", "user", updateUserInfo)
	updateUserInfo, _, _, _, err = impl.UpdateUser(updateUserInfo, token, managerAuth)
	if err != nil {
		impl.logger.Errorw("error while update user", "error", err)
		return nil, err
	}
	return userInfo, nil
}

func (impl UserServiceImpl) createUserIfNotExists(userInfo *bean.UserInfo, emailId string, token string, managerAuth func(resource string, token string, object string) bool) (*bean.UserInfo, error) {
	// if not found, create new user
	dbConnection := impl.userRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	_, err = impl.validateUserRequest(userInfo)
	if err != nil {
		err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, please provide role filters"}
		return nil, err
	}

	//create new user in our db on d basis of info got from google api or hex. assign a basic role
	model := &repository.UserModel{
		EmailId:     emailId,
		AccessToken: userInfo.AccessToken,
		UserType:    userInfo.UserType,
	}
	model.Active = true
	model.CreatedBy = userInfo.UserId
	model.UpdatedBy = userInfo.UserId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	model, err = impl.userRepository.CreateUser(model, tx)
	if err != nil {
		impl.logger.Errorw("error in creating new user", "error", err)
		err = &util.ApiError{
			Code:            constants.UserCreateDBFailed,
			InternalMessage: "failed to create new user in db",
			UserMessage:     fmt.Sprintf("requested by %d", userInfo.UserId),
		}
		return nil, err
	}
	userInfo.Id = model.Id
	isGroupClaimsActive := impl.globalAuthorisationConfigService.IsGroupClaimsConfigActive()
	isSystemManagedActive := impl.globalAuthorisationConfigService.IsDevtronSystemManagedConfigActive()
	// case: when system managed is not active and group claims is active , so only create user, not permissions.
	if !isSystemManagedActive && isGroupClaimsActive {
		userInfo.RoleFilters = []bean.RoleFilter{}
		userInfo.Groups = []string{}
		userInfo.SuperAdmin = false
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
		return userInfo, nil
	}
	//loading policy for safety
	casbin2.LoadPolicy()

	//Starts Role and Mapping
	capacity, mapping := impl.userCommonService.GetCapacityForRoleFilter(userInfo.RoleFilters)
	//var policies []casbin2.Policy
	var policies = make([]casbin2.Policy, 0, capacity)
	if userInfo.SuperAdmin == false {
		for index, roleFilter := range userInfo.RoleFilters {
			impl.logger.Infow("Creating Or updating User Roles for RoleFilter ")
			entity := roleFilter.Entity
			policiesToBeAdded, _, err := impl.CreateOrUpdateUserRolesForAllTypes(roleFilter, userInfo.UserId, model, nil, token, managerAuth, tx, entity, mapping[index])
			if err != nil {
				impl.logger.Errorw("error in creating user roles for Alltypes", "err", err)
				return nil, err
			}
			policies = append(policies, policiesToBeAdded...)

		}

		// START GROUP POLICY
		for _, item := range userInfo.Groups {
			userGroup, err := impl.roleGroupRepository.GetRoleGroupByName(item)
			if err != nil {
				return nil, err
			}
			//object := "group:" + strings.ReplaceAll(item, " ", "_")
			policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(emailId), Obj: casbin2.Object(userGroup.CasbinName)})
		}
		// END GROUP POLICY
	} else if userInfo.SuperAdmin == true {

		isSuperAdmin, err := impl.IsSuperAdmin(int(userInfo.UserId), token)
		if err != nil {
			return nil, err
		}
		if isSuperAdmin == false {
			err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to update super admin type user"}
			return nil, err
		}
		flag, err := impl.userAuthRepository.CreateRoleForSuperAdminIfNotExists(tx, userInfo.UserId)
		if err != nil || flag == false {
			return nil, err
		}
		roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes("", "", "", "", bean2.SUPER_ADMIN, false, "", "", "", "", "", "", "", false, "")
		if err != nil {
			return nil, err
		}
		if roleModel.Id > 0 {
			userRoleModel := &repository.UserRoleModel{UserId: model.Id, RoleId: roleModel.Id, AuditLog: sql.AuditLog{
				CreatedBy: userInfo.UserId,
				CreatedOn: time.Now(),
				UpdatedBy: userInfo.UserId,
				UpdatedOn: time.Now(),
			}}
			userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
			if err != nil {
				return nil, err
			}
			policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
		}

	}
	impl.logger.Infow("Checking the length of policies to be added and Adding in casbin ")
	if len(policies) > 0 {
		impl.logger.Infow("Adding policies in casbin")
		err = casbin2.AddPolicy(policies)
		if err != nil {
			impl.logger.Errorw("casbin policy addition failed", "err", err)
			return nil, err
		}
	}
	//Ends
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	//loading policy for syncing orchestrator to casbin with newly added policies
	casbin2.LoadPolicy()
	return userInfo, nil
}

func (impl UserServiceImpl) CreateOrUpdateUserRolesForAllTypes(roleFilter bean.RoleFilter, userId int32, model *repository.UserModel, existingRoles map[int]repository.UserRoleModel, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, entity string, capacity int) ([]casbin2.Policy, bool, error) {
	//var policiesToBeAdded []casbin2.Policy
	var policiesToBeAdded = make([]casbin2.Policy, 0, capacity)
	var err error
	rolesChanged := false
	if entity == bean2.CLUSTER {
		policiesToBeAdded, rolesChanged, err = impl.createOrUpdateUserRolesForClusterEntity(roleFilter, userId, model, existingRoles, token, managerAuth, tx, entity, capacity)
		if err != nil {
			return nil, false, err
		}
	} else if entity == bean2.EntityJobs {
		policiesToBeAdded, rolesChanged, err = impl.createOrUpdateUserRolesForJobsEntity(roleFilter, userId, model, existingRoles, token, managerAuth, tx, entity, capacity)
		if err != nil {
			return nil, false, err
		}
	} else {
		policiesToBeAdded, rolesChanged, err = impl.createOrUpdateUserRolesForOtherEntity(roleFilter, userId, model, existingRoles, token, managerAuth, tx, entity, capacity)
		if err != nil {
			return nil, false, err
		}
	}
	return policiesToBeAdded, rolesChanged, nil
}

func (impl UserServiceImpl) createOrUpdateUserRolesForClusterEntity(roleFilter bean.RoleFilter, userId int32, model *repository.UserModel, existingRoles map[int]repository.UserRoleModel, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, entity string, capacity int) ([]casbin2.Policy, bool, error) {

	//var policiesToBeAdded []casbin2.Policy
	rolesChanged := false
	namespaces := strings.Split(roleFilter.Namespace, ",")
	groups := strings.Split(roleFilter.Group, ",")
	kinds := strings.Split(roleFilter.Kind, ",")
	resources := strings.Split(roleFilter.Resource, ",")

	//capacity := len(namespaces) * len(groups) * len(kinds) * len(resources) * 2
	actionType := roleFilter.Action
	accessType := roleFilter.AccessType
	var policiesToBeAdded = make([]casbin2.Policy, 0, capacity)
	for _, namespace := range namespaces {
		for _, group := range groups {
			for _, kind := range kinds {
				for _, resource := range resources {
					if managerAuth != nil {
						isValidAuth := impl.userCommonService.CheckRbacForClusterEntity(roleFilter.Cluster, namespace, group, kind, resource, token, managerAuth)
						if !isValidAuth {
							continue
						}
					}
					impl.logger.Infow("Getting Role by filter for cluster")
					roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, "", "", "", "", false, accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, false, "")
					if err != nil {
						return policiesToBeAdded, rolesChanged, err
					}
					if roleModel.Id == 0 {
						impl.logger.Infow("Creating Polices for cluster", resource, kind, namespace, group)
						flag, err, policiesAdded := impl.userCommonService.CreateDefaultPoliciesForAllTypes("", "", "", entity, roleFilter.Cluster, namespace, group, kind, resource, actionType, accessType, false, "", userId)
						if err != nil || flag == false {
							return policiesToBeAdded, rolesChanged, err
						}
						policiesToBeAdded = append(policiesToBeAdded, policiesAdded...)
						impl.logger.Infow("getting role again for cluster")
						roleModel, err = impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, "", "", "", "", false, accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, false, "")
						if err != nil {
							return policiesToBeAdded, rolesChanged, err
						}
						if roleModel.Id == 0 {
							continue
						}
					}
					if _, ok := existingRoles[roleModel.Id]; ok {
						//Adding policies which are removed
						policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
					} else {
						if roleModel.Id > 0 {
							rolesChanged = true
							userRoleModel := &repository.UserRoleModel{
								UserId: model.Id,
								RoleId: roleModel.Id,
								AuditLog: sql.AuditLog{
									CreatedBy: userId,
									CreatedOn: time.Now(),
									UpdatedBy: userId,
									UpdatedOn: time.Now(),
								}}
							userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
							if err != nil {
								return nil, rolesChanged, err
							}
							policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
						}
					}
				}
			}
		}
	}
	return policiesToBeAdded, rolesChanged, nil
}

func (impl UserServiceImpl) mergeRoleFilter(oldR []bean.RoleFilter, newR []bean.RoleFilter) []bean.RoleFilter {
	var roleFilters []bean.RoleFilter
	keysMap := make(map[string]bool)
	for _, role := range oldR {
		roleFilters = append(roleFilters, bean.RoleFilter{
			Entity:      role.Entity,
			Team:        role.Team,
			Environment: role.Environment,
			EntityName:  role.EntityName,
			Action:      role.Action,
			AccessType:  role.AccessType,
			Cluster:     role.Cluster,
			Namespace:   role.Namespace,
			Group:       role.Group,
			Kind:        role.Kind,
			Resource:    role.Resource,
			Approver:    role.Approver,
			Workflow:    role.Workflow,
		})
		key := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%t-%s-%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment,
			role.EntityName, role.Action, role.AccessType, role.Approver, role.Cluster, role.Namespace, role.Group, role.Kind, role.Resource, role.Workflow)
		keysMap[key] = true
	}
	for _, role := range newR {
		key := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%t-%s-%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment,
			role.EntityName, role.Action, role.AccessType, role.Approver, role.Cluster, role.Namespace, role.Group, role.Kind, role.Resource, role.Workflow)
		if _, ok := keysMap[key]; !ok {
			roleFilters = append(roleFilters, bean.RoleFilter{
				Entity:      role.Entity,
				Team:        role.Team,
				Environment: role.Environment,
				EntityName:  role.EntityName,
				Action:      role.Action,
				AccessType:  role.AccessType,
				Cluster:     role.Cluster,
				Namespace:   role.Namespace,
				Group:       role.Group,
				Kind:        role.Kind,
				Resource:    role.Resource,
				Approver:    role.Approver,
				Workflow:    role.Workflow,
			})
		}
	}
	return roleFilters
}

func (impl UserServiceImpl) mergeGroups(oldGroups []string, newGroups []string) []string {
	var groups []string
	keysMap := make(map[string]bool)
	for _, group := range oldGroups {
		groups = append(groups, group)
		key := fmt.Sprintf(group)
		keysMap[key] = true
	}
	for _, group := range newGroups {
		key := fmt.Sprintf(group)
		if _, ok := keysMap[key]; !ok {
			groups = append(groups, group)
		}
	}
	return groups
}

func (impl UserServiceImpl) UpdateUser(userInfo *bean.UserInfo, token string, managerAuth func(resource, token string, object string) bool) (*bean.UserInfo, bool, bool, []string, error) {
	//checking if request for same user is being processed
	isLocked := impl.getUserReqLockStateById(userInfo.Id)
	if isLocked {
		impl.logger.Errorw("received concurrent request for user update, UpdateUser", "userId", userInfo.Id)
		return nil, false, false, nil, &util.ApiError{
			Code:           "409",
			HttpStatusCode: http.StatusConflict,
			UserMessage:    ConcurrentRequestLockError,
		}
	} else {
		//locking state for this user since it's ready to serve
		err := impl.lockUnlockUserReqState(userInfo.Id, true)
		if err != nil {
			impl.logger.Errorw("error in locking, lockUnlockUserReqState", "userId", userInfo.Id)
			return nil, false, false, nil, err
		}
		defer func() {
			err = impl.lockUnlockUserReqState(userInfo.Id, false)
			if err != nil {
				impl.logger.Errorw("error in unlocking, lockUnlockUserReqState", "userId", userInfo.Id)
			}
		}()
	}
	dbConnection := impl.userRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, false, false, nil, err
	}
	model, err := impl.userRepository.GetByIdIncludeDeleted(userInfo.Id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, false, false, nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	isSystemManagedActive := impl.globalAuthorisationConfigService.IsDevtronSystemManagedConfigActive()
	isUserActive := model.Active
	isApiToken := util3.CheckIfApiToken(userInfo.EmailId)
	// case: system managed permissions is not active and user is inActive , mark user as active
	if !isSystemManagedActive && !isUserActive && !isApiToken {
		err = impl.UpdateUserToActive(model, tx, userInfo.EmailId, userInfo.UserId)
		if err != nil {
			impl.logger.Errorw("error in updating user to active", "err", err, "EmailId", userInfo.EmailId)
			return userInfo, false, false, nil, err
		}
		err = tx.Commit()
		if err != nil {
			return nil, false, false, nil, err
		}
		return userInfo, false, false, nil, nil

	} else if !isSystemManagedActive && isUserActive && !isApiToken {
		// case: system managed permissions is not active and user is active , update is not allowed
		err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, user permissions are managed by your SSO Provider groups"}
		impl.logger.Errorw("Invalid request,User permissions are managed by SSO Provider", "error", err)
		return nil, false, false, nil, err
	}
	//validating if action user is not admin and trying to update user who has super admin polices, return 403
	isUserSuperAdmin, err := impl.IsSuperAdmin(int(userInfo.Id), token)
	if err != nil {
		return nil, false, false, nil, err
	}
	isActionPerformingUserSuperAdmin, err := impl.IsSuperAdmin(int(userInfo.UserId), token)
	if err != nil {
		return nil, false, false, nil, err
	}
	//if request comes to make user as a super admin or user already a super admin (who'is going to be updated), action performing user should have super admin access
	if userInfo.SuperAdmin || isUserSuperAdmin {
		if !isActionPerformingUserSuperAdmin {
			err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to update super admin type user"}
			impl.logger.Errorw("Invalid request, not allow to update super admin type user", "error", err)
			return nil, false, false, nil, err
		}
	}
	if userInfo.SuperAdmin && isUserSuperAdmin {
		err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "User Already A Super Admin"}
		impl.logger.Errorw("user already a superAdmin", "error", err)
		return nil, false, false, nil, err
	}

	var eliminatedPolicies []casbin2.Policy
	capacity, mapping := impl.userCommonService.GetCapacityForRoleFilter(userInfo.RoleFilters)
	var addedPolicies = make([]casbin2.Policy, 0, capacity)
	restrictedGroups := []string{}
	rolesChanged := false
	groupsModified := false
	//loading policy for safety
	casbin2.LoadPolicy()
	if userInfo.SuperAdmin == false {
		//Starts Role and Mapping
		userRoleModels, err := impl.userAuthRepository.GetUserRoleMappingByUserId(model.Id)
		if err != nil {
			return nil, false, false, nil, err
		}
		existingRoleIds := make(map[int]repository.UserRoleModel)
		eliminatedRoleIds := make(map[int]*repository.UserRoleModel)
		for i := range userRoleModels {
			existingRoleIds[userRoleModels[i].RoleId] = *userRoleModels[i]
			eliminatedRoleIds[userRoleModels[i].RoleId] = userRoleModels[i]
		}

		//validate role filters
		_, err = impl.validateUserRequest(userInfo)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, please provide role filters"}
			return nil, false, false, nil, err
		}

		// DELETE Removed Items
		items, err := impl.userCommonService.RemoveRolesAndReturnEliminatedPolicies(userInfo, existingRoleIds, eliminatedRoleIds, tx, token, managerAuth)
		if err != nil {
			return nil, false, false, nil, err
		}
		eliminatedPolicies = append(eliminatedPolicies, items...)
		if len(eliminatedPolicies) > 0 {
			impl.logger.Debugw("casbin policies to remove for the request", "policies: ", eliminatedPolicies, "userInfo", userInfo)
			rolesChanged = true
		}

		//Adding New Policies
		for index, roleFilter := range userInfo.RoleFilters {
			entity := roleFilter.Entity

			policiesToBeAdded, rolesChangedFromRoleUpdate, err := impl.CreateOrUpdateUserRolesForAllTypes(roleFilter, userInfo.UserId, model, existingRoleIds, token, managerAuth, tx, entity, mapping[index])
			if err != nil {
				impl.logger.Errorw("error in creating user roles for All Types", "err", err)
				return nil, false, false, nil, err
			}
			addedPolicies = append(addedPolicies, policiesToBeAdded...)
			rolesChanged = rolesChangedFromRoleUpdate

		}

		//ROLE GROUP SETUP
		newGroupMap := make(map[string]string)
		oldGroupMap := make(map[string]string)
		userCasbinRoles, err := impl.CheckUserRoles(userInfo.Id, token)

		if err != nil {
			return nil, false, false, nil, err
		}
		for _, oldItem := range userCasbinRoles {
			oldGroupMap[oldItem] = oldItem
		}
		// START GROUP POLICY
		for _, item := range userInfo.Groups {
			userGroup, err := impl.roleGroupRepository.GetRoleGroupByName(item)
			if err != nil {
				return nil, false, false, nil, err
			}
			newGroupMap[userGroup.CasbinName] = userGroup.CasbinName
			if _, ok := oldGroupMap[userGroup.CasbinName]; !ok {
				//check permission for new group which is going to add
				hasAccessToGroup := impl.checkGroupAuth(userGroup.CasbinName, token, managerAuth, isActionPerformingUserSuperAdmin)
				if hasAccessToGroup {
					groupsModified = true
					addedPolicies = append(addedPolicies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(userInfo.EmailId), Obj: casbin2.Object(userGroup.CasbinName)})
				} else {
					trimmedGroup := strings.TrimPrefix(item, "group:")
					restrictedGroups = append(restrictedGroups, trimmedGroup)
				}
			}
		}

		for _, item := range userCasbinRoles {
			if _, ok := newGroupMap[item]; !ok {
				if item != bean.SUPERADMIN {
					//check permission for group which is going to eliminate
					if strings.HasPrefix(item, "group:") {
						hasAccessToGroup := impl.checkGroupAuth(item, token, managerAuth, isActionPerformingUserSuperAdmin)
						if hasAccessToGroup {
							if strings.HasPrefix(item, "group:") {
								groupsModified = true
							}
							eliminatedPolicies = append(eliminatedPolicies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(userInfo.EmailId), Obj: casbin2.Object(item)})
						} else {
							trimmedGroup := strings.TrimPrefix(item, "group:")
							restrictedGroups = append(restrictedGroups, trimmedGroup)
						}
					}
				}
			}
		}
		// END GROUP POLICY

	} else if userInfo.SuperAdmin == true {
		flag, err := impl.userAuthRepository.CreateRoleForSuperAdminIfNotExists(tx, userInfo.UserId)
		if err != nil || flag == false {
			return nil, false, false, nil, err
		}
		roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes("", "", "", "", bean2.SUPER_ADMIN, false, "", "", "", "", "", "", "", false, "")
		if err != nil {
			return nil, false, false, nil, err
		}
		if roleModel.Id > 0 {
			userRoleModel := &repository.UserRoleModel{UserId: model.Id, RoleId: roleModel.Id}
			userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
			if err != nil {
				return nil, false, false, nil, err
			}
			addedPolicies = append(addedPolicies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
		}
	}

	//updating in casbin
	if len(eliminatedPolicies) > 0 {
		impl.logger.Debugw("casbin policies being eliminated", "policies: ", eliminatedPolicies, "userInfo", userInfo)
		pRes := casbin2.RemovePolicy(eliminatedPolicies)
		println(pRes)
	}
	if len(addedPolicies) > 0 {
		impl.logger.Debugw("casbin policies being added", "policies: ", addedPolicies)
		err = casbin2.AddPolicy(addedPolicies)
		if err != nil {
			impl.logger.Errorw("casbin policy addition failed", "err", err)
			return nil, false, false, nil, err
		}
	}
	//Ends

	model.EmailId = userInfo.EmailId // override case sensitivity
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userInfo.UserId
	model.Active = true
	model, err = impl.userRepository.UpdateUser(model, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, false, false, nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, false, false, nil, err
	}
	//loading policy for syncing orchestrator to casbin with newly added policies
	casbin2.LoadPolicy()
	return userInfo, rolesChanged, groupsModified, restrictedGroups, nil
}
func (impl UserServiceImpl) UpdateUserToActive(model *repository.UserModel, tx *pg.Tx, emailId string, userId int32) error {

	model.EmailId = emailId // override case sensitivity
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userId
	model.Active = true
	model, err := impl.userRepository.UpdateUser(model, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return err
	}
	return nil

}

func (impl UserServiceImpl) GetByIdWithoutGroupClaims(id int32) (*bean.UserInfo, error) {
	model, err := impl.userRepository.GetById(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	isSuperAdmin, roleFilters, filterGroups := impl.getUserMetadata(model)
	response := &bean.UserInfo{
		Id:          model.Id,
		EmailId:     model.EmailId,
		RoleFilters: roleFilters,
		Groups:      filterGroups,
		SuperAdmin:  isSuperAdmin,
	}

	return response, nil
}

func (impl UserServiceImpl) GetRoleFiltersForAUserById(id int32) (*bean.UserInfo, error) {
	model, err := impl.userRepository.GetById(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	var roleFilters []bean.RoleFilter
	isGroupClaimsActive := impl.globalAuthorisationConfigService.IsGroupClaimsConfigActive()
	if isGroupClaimsActive {
		roleFiltersFromGroupClaims, err := impl.getRoleFiltersForGroupClaims(id)
		if err != nil {
			impl.logger.Errorw("error in getRoleFiltersForGroupClaims", "err", err, "userId", id)
			return nil, err
		}
		if len(roleFiltersFromGroupClaims) > 0 {
			roleFilters = append(roleFilters, roleFiltersFromGroupClaims...)
		}

	} else {
		roleFiltersFromDevtronManaged, err := impl.getRolefiltersForDevtronManaged(model)
		if err != nil {
			impl.logger.Errorw("error while getRolefiltersForDevtronManaged", "error", err, "id", model.Id)
			return nil, err
		}
		if len(roleFiltersFromDevtronManaged) > 0 {
			roleFilters = append(roleFilters, roleFiltersFromDevtronManaged...)
		}
	}

	response := &bean.UserInfo{
		Id:          model.Id,
		EmailId:     model.EmailId,
		RoleFilters: roleFilters,
	}

	return response, nil
}

func (impl UserServiceImpl) GetByIdForGroupClaims(id int32) (*bean.UserInfo, error) {
	model, err := impl.userRepository.GetById(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	var roleFilters []bean.RoleFilter
	var filterGroups []string
	var isSuperAdmin bool
	var roleGroups []bean.RoleGroup
	isGroupClaimsActive := impl.globalAuthorisationConfigService.IsGroupClaimsConfigActive()
	if isGroupClaimsActive {
		roleGroups, err = impl.getRoleGroupsForGroupClaims(id)
		if err != nil {
			impl.logger.Errorw("error in getRoleGroupsForGroupClaims ", "err", err, "id", id)
			return nil, err
		}
	} else {
		// Intentionally considering ad or devtron managed here to avoid conflicts
		isSuperAdmin, roleFilters, filterGroups = impl.getUserMetadata(model)
	}
	response := &bean.UserInfo{
		Id:          model.Id,
		EmailId:     model.EmailId,
		RoleFilters: roleFilters,
		Groups:      filterGroups,
		SuperAdmin:  isSuperAdmin,
		RoleGroups:  roleGroups,
	}

	return response, nil
}

func (impl UserServiceImpl) fetchRoleGroupsByGroupClaims(groupClaims []string) ([]bean.RoleGroup, error) {
	_, roleGroups, err := impl.roleGroupService.FetchRoleGroupsWithRolesByGroupCasbinNames(groupClaims)
	if err != nil {
		impl.logger.Errorw("error in fetchRoleGroupsByGroupClaims", "err", err, "groupClaims", groupClaims)
		return nil, err
	}
	return roleGroups, err
}

func (impl UserServiceImpl) getUserMetadata(model *repository.UserModel) (bool, []bean.RoleFilter, []string) {
	roles, err := impl.userAuthRepository.GetRolesByUserId(model.Id)
	if err != nil {
		impl.logger.Debugw("No Roles Found for user", "id", model.Id)
	}

	isSuperAdmin := false
	var roleFilters []bean.RoleFilter
	roleFilterMap := make(map[string]*bean.RoleFilter)
	for _, role := range roles {
		key := impl.userCommonService.GetUniqueKeyForAllEntity(role)
		if _, ok := roleFilterMap[key]; ok {
			impl.userCommonService.BuildRoleFilterForAllTypes(roleFilterMap, role, key)
		} else {
			roleFilterMap[key] = &bean.RoleFilter{
				Entity:      role.Entity,
				Team:        role.Team,
				Environment: role.Environment,
				EntityName:  role.EntityName,
				Action:      role.Action,
				AccessType:  role.AccessType,
				Cluster:     role.Cluster,
				Namespace:   role.Namespace,
				Group:       role.Group,
				Kind:        role.Kind,
				Resource:    role.Resource,
				Approver:    role.Approver,
				Workflow:    role.Workflow,
			}

		}
		if role.Role == bean.SUPERADMIN {
			isSuperAdmin = true
		}
	}
	for _, v := range roleFilterMap {
		if v.Action == "super-admin" {
			continue
		}
		roleFilters = append(roleFilters, *v)
	}
	roleFilters = impl.userCommonService.MergeCustomRoleFilters(roleFilters)
	groups, err := casbin2.GetRolesForUser(model.EmailId)
	if err != nil {
		impl.logger.Warnw("No Roles Found for user", "id", model.Id)
	}

	var filterGroups []string
	for _, item := range groups {
		if strings.Contains(item, "group:") {
			filterGroups = append(filterGroups, item)
		}
	}

	if len(filterGroups) > 0 {
		filterGroupsModels, err := impl.roleGroupRepository.GetRoleGroupListByCasbinNames(filterGroups)
		if err != nil {
			impl.logger.Warnw("No Roles Found for user", "id", model.Id)
		}
		filterGroups = nil
		for _, item := range filterGroupsModels {
			filterGroups = append(filterGroups, item.Name)
		}
	} else {
		impl.logger.Warnw("no roles found for user", "email", model.EmailId)
	}

	if len(filterGroups) == 0 {
		filterGroups = make([]string, 0)
	}
	if len(roleFilters) == 0 {
		roleFilters = make([]bean.RoleFilter, 0)
	}
	for index, roleFilter := range roleFilters {
		if roleFilter.Entity == "" {
			roleFilters[index].Entity = bean2.ENTITY_APPS
			if roleFilter.AccessType == "" {
				roleFilters[index].AccessType = bean2.DEVTRON_APP
			}
		}
	}
	return isSuperAdmin, roleFilters, filterGroups
}

// GetAll excluding API token user
func (impl UserServiceImpl) GetAll() ([]bean.UserInfo, error) {
	model, err := impl.userRepository.GetAllExcludingApiTokenUser()
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	var response []bean.UserInfo
	for _, m := range model {
		response = append(response, bean.UserInfo{
			Id:          m.Id,
			EmailId:     m.EmailId,
			RoleFilters: make([]bean.RoleFilter, 0),
			Groups:      make([]string, 0),
		})
	}
	if len(response) == 0 {
		response = make([]bean.UserInfo, 0)
	}
	return response, nil
}

// GetAllWithFilters takes filter arguments gives UserListingResponse as output with some operations like filter, sorting, searching,pagination support inbuilt
func (impl UserServiceImpl) GetAllWithFilters(status string, sortOrder string, sortBy string, offset int, totalSize int, showAll bool, searchKey string) (*bean.UserListingResponse, error) {
	// get req from arguments
	request := impl.getRequestWithFiltersArgsOrDefault(status, sortOrder, sortBy, offset, totalSize, showAll, searchKey)
	if request.ShowAll {
		return impl.getAllDetailedUsers()
	}
	// Setting size as zero to calculate the total number of results based on request
	size := request.Size
	request.Size = 0

	// Recording time here for overall consistency
	request.CurrentTime = time.Now()

	// Build query from query builder
	query := impl.userListingRepositoryQueryBuilder.GetQueryForUserListingWithFilters(request)
	totalCount, err := impl.userRepository.GetCountExecutingQuery(query)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	request.Size = size

	query = impl.userListingRepositoryQueryBuilder.GetQueryForUserListingWithFilters(request)
	models, err := impl.userRepository.GetAllExecutingQuery(query)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	response, err := impl.getUserResponse(models, request.CurrentTime)
	if err != nil {
		impl.logger.Errorw("error in getUserResponseWithLoginAudit", "err", err)
		return nil, err
	}

	listingResponse := &bean.UserListingResponse{
		Users:      response,
		TotalCount: totalCount,
	}
	return listingResponse, nil

}
func (impl UserServiceImpl) getDesiredResponseWithOffSetAndSize(resp []bean.UserInfo, offset, size int) []bean.UserInfo {
	sizeOfResponse := len(resp)
	if sizeOfResponse == 0 {
		return resp
	}
	start, end := util3.FindStartAndEndForOffsetAndSize(sizeOfResponse, offset, size)
	if start >= end {
		return []bean.UserInfo{}
	}
	return resp[start:end]

}
func (impl UserServiceImpl) getRequestWithFiltersArgsOrDefault(status string, sortOrder string, sortBy string, offset int, totalSize int, showAll bool, searchKey string) *bean.FetchListingRequest {
	sortByRes, size := impl.userCommonService.GetDefaultValuesIfNotPresent(sortBy, totalSize, false)
	request := &bean.FetchListingRequest{
		Status:    bean.Status(status),
		SortOrder: bean2.SortOrder(sortOrder),
		SortBy:    sortByRes,
		Offset:    offset,
		Size:      size,
		ShowAll:   showAll,
		SearchKey: searchKey,
	}
	return request
}

func (impl UserServiceImpl) getAllDetailedUsers() (*bean.UserListingResponse, error) {
	response, err := impl.GetAllDetailedUsers()
	if err != nil {
		impl.logger.Errorw("error in getAllDetailedUsers", "err", err)
		return nil, err
	}

	listingResponse := &bean.UserListingResponse{
		Users:      response,
		TotalCount: len(response),
	}
	return listingResponse, err
}

func (impl UserServiceImpl) getUserResponse(model []repository.UserModel, recordedTime time.Time) ([]bean.UserInfo, error) {
	var response []bean.UserInfo
	for _, m := range model {
		userStatus, ttlTime := impl.getStatusAndTTL(m, recordedTime)
		lastLoginTime := impl.getLastLoginTime(m)
		response = append(response, bean.UserInfo{
			Id:            m.Id,
			EmailId:       m.EmailId,
			RoleFilters:   make([]bean.RoleFilter, 0),
			Groups:        make([]string, 0),
			LastLoginTime: lastLoginTime,
			UserStatus:    userStatus,
			TimeToLive:    ttlTime,
		})
	}
	if len(response) == 0 {
		response = make([]bean.UserInfo, 0)
	}
	return response, nil
}

func (impl UserServiceImpl) sortByLoginTime(users []bean.UserInfo, sortOrder bean2.SortOrder) []bean.UserInfo {
	if sortOrder == bean2.Asc {
		sort.Slice(users, func(i, j int) bool {
			return users[i].LastLoginTime.Before(users[j].LastLoginTime)
		})
	} else if sortOrder == bean2.Desc {
		sort.Slice(users, func(i, j int) bool {
			return users[i].LastLoginTime.After(users[j].LastLoginTime)
		})
	}
	return users

}

func (impl UserServiceImpl) GetAllDetailedUsers() ([]bean.UserInfo, error) {
	query := impl.userListingRepositoryQueryBuilder.GetQueryForAllUserWithAudit()
	models, err := impl.userRepository.GetAllExecutingQuery(query)
	if err != nil {
		impl.logger.Errorw("error in GetAllDetailedUsers", "err", err)
		return nil, err
	}

	var response []bean.UserInfo
	// recording time here for overall status consistency
	recordedTime := time.Now()

	for _, model := range models {
		isSuperAdmin, roleFilters, filterGroups := impl.getUserMetadata(&model)
		userStatus, ttlTime := impl.getStatusAndTTL(model, recordedTime)
		lastLoginTime := impl.getLastLoginTime(model)
		response = append(response, bean.UserInfo{
			Id:            model.Id,
			EmailId:       model.EmailId,
			RoleFilters:   roleFilters,
			Groups:        filterGroups,
			SuperAdmin:    isSuperAdmin,
			LastLoginTime: lastLoginTime,
			UserStatus:    userStatus,
			TimeToLive:    ttlTime,
		})
	}
	if len(response) == 0 {
		response = make([]bean.UserInfo, 0)
	}
	return response, nil
}
func (impl UserServiceImpl) getLastLoginTime(model repository.UserModel) time.Time {
	lastLoginTime := time.Time{}
	if model.UserAudit != nil {
		lastLoginTime = model.UserAudit.UpdatedOn
	}
	return lastLoginTime
}

func (impl UserServiceImpl) UserExists(emailId string) bool {
	model, err := impl.userRepository.FetchActiveUserByEmail(emailId)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return false
	}
	if model.Id == 0 {
		impl.logger.Errorw("no user found ", "email", emailId)
		return false
	} else {
		return true
	}
}

func (impl UserServiceImpl) CheckForApproverAccess(appName, envName string, userId int32) bool {
	allowedUsers, err := impl.GetApprovalUsersByEnv(appName, envName)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching approval users", "appName", appName, "envName", envName, "err", err)
		return false
	}
	email, err := impl.GetEmailById(userId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching user details", "userId", userId, "err", err)
		return false
	}
	allowed := userId == 2 // admin user
	if !allowed {
		for _, allowedUser := range allowedUsers {
			if email == allowedUser {
				allowed = true
				break
			}
		}
	}
	return allowed
}

func (impl UserServiceImpl) GetConfigApprovalUsersByEnv(appName, envName, team string) ([]string, error) {
	emailIds, permissionGroupNames, err := impl.userAuthRepository.GetConfigApprovalUsersByEnv(appName, envName, team)
	if err != nil {
		return emailIds, err
	}
	finalEmails, err := impl.extractEmailIds(permissionGroupNames, emailIds)
	if err != nil {
		return emailIds, err
	}
	return finalEmails, nil
}

func (impl UserServiceImpl) GetApprovalUsersByEnv(appName, envName string) ([]string, error) {
	emailIds, permissionGroupNames, err := impl.userAuthRepository.GetApprovalUsersByEnv(appName, envName)
	if err != nil {
		return emailIds, err
	}
	finalEmails, err := impl.extractEmailIds(permissionGroupNames, emailIds)
	if err != nil {
		return emailIds, err
	}
	return finalEmails, nil
}

func (impl UserServiceImpl) extractEmailIds(permissionGroupNames []string, emailIds []string) ([]string, error) {
	for _, groupName := range permissionGroupNames {
		userEmails, err := casbin2.GetUserByRole(groupName)
		if err != nil {
			return emailIds, err
		}
		emailIds = append(emailIds, userEmails...)
	}
	uniqueEmails := make(map[string]bool)
	for _, emailId := range emailIds {
		_, ok := uniqueEmails[emailId]
		if !ok {
			uniqueEmails[emailId] = true
		}
	}
	var finalEmails []string
	for emailId, _ := range uniqueEmails {
		finalEmails = append(finalEmails, emailId)
	}
	return finalEmails, nil
}
func (impl UserServiceImpl) SaveLoginAudit(emailId, clientIp string, id int32) {

	if emailId != "" && id <= 0 {
		user, err := impl.GetUserByEmail(emailId)
		if err != nil {
			impl.logger.Errorw("error in getting userInfo by emailId", "err", err, "emailId", emailId)
			return
		}
		id = user.Id
	}
	if id <= 0 {
		impl.logger.Errorw("Invalid id to save login audit of sso user", "Id", id)
		return
	}
	model := UserAudit{
		UserId:   id,
		ClientIp: clientIp,
	}
	err := impl.userAuditService.Update(&model)
	if err != nil {
		impl.logger.Errorw("error occurred while saving user audit", "err", err)
	}
}

func (impl UserServiceImpl) GetUserWithTimeoutWindowConfiguration(emailId string) (int32, bool, error) {
	isInactive := true
	user, err := impl.userRepository.GetUserWithTimeoutWindowConfiguration(emailId)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return user.Id, isInactive, err
	}

	if user.TimeoutWindowConfigurationId == 0 {
		isInactive = false
		return user.Id, isInactive, nil
	} else {
		expiryDate, err := time.Parse(helper.TimeFormatForParsing, user.TimeoutWindowConfiguration.TimeoutWindowExpression)
		if err != nil {
			impl.logger.Errorw("error while parsing date time", "error", err)
			return user.Id, isInactive, err
		}
		if expiryDate.After(time.Now()) {
			isInactive = false
		}
	}
	return user.Id, isInactive, nil
}

func (impl UserServiceImpl) GetUserByEmail(emailId string) (*bean.UserInfo, error) {
	model, err := impl.userRepository.FetchActiveUserByEmail(emailId)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	roles, err := impl.userAuthRepository.GetRolesByUserId(model.Id)
	if err != nil {
		impl.logger.Warnw("No Roles Found for user", "id", model.Id)
	}
	var roleFilters []bean.RoleFilter
	for _, role := range roles {
		roleFilters = append(roleFilters, bean.RoleFilter{
			Entity:      role.Entity,
			Team:        role.Team,
			Environment: role.Environment,
			EntityName:  role.EntityName,
			Action:      role.Action,
			Cluster:     role.Cluster,
			Namespace:   role.Namespace,
			Group:       role.Group,
			Kind:        role.Kind,
			Resource:    role.Resource,
			Workflow:    role.Workflow,
		})
	}

	response := &bean.UserInfo{
		Id:          model.Id,
		EmailId:     model.EmailId,
		UserType:    model.UserType,
		AccessToken: model.AccessToken,
		RoleFilters: roleFilters,
	}

	return response, nil
}

func (impl *UserServiceImpl) GetEmailById(userId int32) (string, error) {
	var emailId string
	model, err := impl.userRepository.GetById(userId)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return emailId, err
	}
	if model != nil {
		emailId = model.EmailId
	}
	return emailId, nil
}

func (impl UserServiceImpl) GetLoggedInUser(r *http.Request) (int32, error) {
	_, span := otel.Tracer("userService").Start(r.Context(), "GetLoggedInUser")
	defer span.End()
	token := ""
	if strings.Contains(r.URL.Path, "/orchestrator/webhook/ext-ci/") {
		token = r.Header.Get("api-token")
	} else {
		token = r.Header.Get("token")
	}
	userId, userType, err := impl.GetUserByToken(r.Context(), token)
	// if user is of api-token type, then update lastUsedBy and lastUsedAt
	if err == nil && userType == bean.USER_TYPE_API_TOKEN {
		go impl.saveUserAudit(r, userId)
	}
	return userId, err
}

func (impl UserServiceImpl) GetUserByToken(context context.Context, token string) (int32, string, error) {
	_, span := otel.Tracer("userService").Start(context, "GetUserByToken")
	email, _, err := impl.GetEmailAndGroupClaimsFromToken(token)
	span.End()
	if err != nil {
		return http.StatusUnauthorized, "", err
	}
	userInfo, err := impl.GetUserByEmail(email)
	if err != nil {
		impl.logger.Errorw("unable to fetch user from db", "error", err)
		err := &util.ApiError{
			Code:            constants.UserNotFoundForToken,
			InternalMessage: "user not found for token",
			UserMessage:     fmt.Sprintf("no user found against provided token: %s", token),
		}
		return http.StatusUnauthorized, "", err
	}
	return userInfo.Id, userInfo.UserType, nil
}
func (impl UserServiceImpl) GetFieldValuesFromToken(token string) ([]byte, error) {
	var claimBytes []byte
	mapClaims, err := impl.getMapClaims(token)
	if err != nil {
		return claimBytes, err
	}
	impl.logger.Infow("got map claims", "mapClaims", mapClaims)
	claimBytes, err = json.Marshal(mapClaims)
	if err != nil {
		return nil, err
	}
	return claimBytes, nil
}
func (impl UserServiceImpl) getMapClaims(token string) (jwt2.MapClaims, error) {
	if token == "" {
		impl.logger.Infow("no token provided")
		err := &util.ApiError{
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "no token provided",
		}
		return nil, err
	}

	claims, err := impl.sessionManager2.VerifyToken(token)

	if err != nil {
		impl.logger.Errorw("failed to verify token", "error", err)
		err := &util.ApiError{
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "failed to verify token",
			UserMessage:     "token verification failed while getting logged in user",
		}
		return nil, err
	}
	mapClaims, err := jwt.MapClaims(claims)
	if err != nil {
		impl.logger.Errorw("failed to MapClaims", "error", err)
		err := &util.ApiError{
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "token invalid",
			UserMessage:     "token verification failed while parsing token",
		}
		return nil, err
	}
	return mapClaims, nil
}

func (impl UserServiceImpl) GetEmailAndGroupClaimsFromToken(token string) (string, []string, error) {
	mapClaims, err := impl.getMapClaims(token)
	if err != nil {
		impl.logger.Errorw("error in fetching map claims", "err", err)
		return "", nil, err
	}
	groupsClaims := make([]string, 0)
	email, groups := impl.globalAuthorisationConfigService.GetEmailAndGroupsFromClaims(mapClaims)
	if impl.globalAuthorisationConfigService.IsGroupClaimsConfigActive() {
		groupsClaims = groups
	}
	return email, groupsClaims, nil
}

func (impl UserServiceImpl) GetByIds(ids []int32) ([]bean.UserInfo, error) {
	var beans []bean.UserInfo
	if len(ids) == 0 {
		return beans, nil
	}
	models, err := impl.userRepository.GetByIds(ids)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	if len(models) > 0 {
		for _, item := range models {
			beans = append(beans, bean.UserInfo{Id: item.Id, EmailId: item.EmailId})
		}
	}
	return beans, nil
}

func (impl UserServiceImpl) DeleteUser(bean *bean.UserInfo) (bool, error) {

	dbConnection := impl.roleGroupRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return false, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	model, err := impl.userRepository.GetById(bean.Id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return false, err
	}
	urm, err := impl.userAuthRepository.GetUserRoleMappingByUserId(bean.Id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return false, err
	}
	for _, item := range urm {
		_, err = impl.userAuthRepository.DeleteUserRoleMapping(item, tx)
		if err != nil {
			impl.logger.Errorw("error while fetching user from db", "error", err)
			return false, err
		}
	}
	model.Active = false
	model.UpdatedBy = bean.UserId
	model.UpdatedOn = time.Now()
	model, err = impl.userRepository.UpdateUser(model, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}

	groups, err := casbin2.GetRolesForUser(model.EmailId)
	if err != nil {
		impl.logger.Warnw("No Roles Found for user", "id", model.Id)
	}
	var eliminatedPolicies []casbin2.Policy
	for _, item := range groups {
		flag := casbin2.DeleteRoleForUser(model.EmailId, item)
		if flag == false {
			impl.logger.Warnw("unable to delete role:", "user", model.EmailId, "role", item)
		}
		eliminatedPolicies = append(eliminatedPolicies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(item)})
	}
	// updating in casbin
	if len(eliminatedPolicies) > 0 {
		pRes := casbin2.RemovePolicy(eliminatedPolicies)
		impl.logger.Infow("Failed to remove policies", "policies", pRes)
	}

	return true, nil
}

func (impl UserServiceImpl) CheckUserRoles(id int32, token string) ([]string, error) {
	model, err := impl.userRepository.GetByIdIncludeDeleted(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	isGroupClaimsActive := impl.globalAuthorisationConfigService.IsGroupClaimsConfigActive()
	isDevtronSystemActive := impl.globalAuthorisationConfigService.IsDevtronSystemManagedConfigActive()
	var groups []string
	if isDevtronSystemActive || util3.CheckIfAdminOrApiToken(model.EmailId) {
		groupsCasbin, err := casbin2.GetRolesForUser(model.EmailId)
		if err != nil {
			impl.logger.Errorw("No Roles Found for user", "id", model.Id)
			return nil, err
		}
		if len(groupsCasbin) > 0 {
			groups = append(groups, groupsCasbin...)
			grps, err := impl.getUniquesRolesByGroupCasbinNames(groups)
			if err != nil {
				impl.logger.Errorw("error in getUniquesRolesByGroupCasbinNames", "err", err)
				return nil, err
			}
			groups = append(groups, grps...)
		}
	}

	if isGroupClaimsActive {
		_, groupClaims, err := impl.GetEmailAndGroupClaimsFromToken(token)
		if err != nil {
			impl.logger.Errorw("error in GetEmailAndGroupClaimsFromToken", "err", err)
			return nil, err
		}
		if len(groupClaims) > 0 {
			groupsCasbinNames := util3.GetGroupCasbinName(groupClaims)
			grps, err := impl.getUniquesRolesByGroupCasbinNames(groupsCasbinNames)
			if err != nil {
				impl.logger.Errorw("error in getUniquesRolesByGroupCasbinNames", "err", err)
				return nil, err
			}
			groups = append(groups, grps...)
		}
	}

	return groups, nil
}

func (impl UserServiceImpl) getUniquesRolesByGroupCasbinNames(groupCasbinNames []string) ([]string, error) {
	var groups []string
	rolesModels, err := impl.roleGroupRepository.GetRolesByGroupCasbinNames(groupCasbinNames)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting roles by group names", "err", err)
		return nil, err
	}
	uniqueRolesFromGroupMap := make(map[string]bool)
	rolesFromGroup := make([]string, 0, len(rolesModels))
	for _, roleModel := range rolesModels {
		uniqueRolesFromGroupMap[roleModel.Role] = true
	}
	for role, _ := range uniqueRolesFromGroupMap {
		rolesFromGroup = append(rolesFromGroup, role)
	}
	if len(rolesFromGroup) > 0 {
		groups = append(groups, rolesFromGroup...)
	}
	return groups, nil
}

func (impl UserServiceImpl) SyncOrchestratorToCasbin() (bool, error) {
	roles, err := impl.userAuthRepository.GetAllRole()
	if err != nil {
		impl.logger.Errorw("error while fetching roles from db", "error", err)
		return false, err
	}
	total := len(roles)
	processed := 0
	impl.logger.Infow("total roles found for sync", "len", total)
	//loading policy for safety
	casbin2.LoadPolicy()
	for _, role := range roles {
		if len(role.Team) > 0 {
			flag, err := impl.userAuthRepository.SyncOrchestratorToCasbin(role.Team, role.EntityName, role.Environment, nil)
			if err != nil {
				impl.logger.Errorw("error sync orchestrator to casbin", "error", err)
				return false, err
			}
			if !flag {
				impl.logger.Infow("sync failed orchestrator to db", "roleId", role.Id)
			}
		}
		processed = processed + 1
	}
	//loading policy for syncing orchestrator to casbin with updated policies(if any)
	casbin2.LoadPolicy()
	impl.logger.Infow("total roles processed for sync", "len", processed)
	return true, nil
}

// TODO Kripansh remove this
func (impl UserServiceImpl) IsSuperAdminForDevtronManaged(userId int) (bool, error) {
	//validating if action user is not admin and trying to update user who has super admin polices, return 403
	isSuperAdmin := false
	// TODO Kripansh: passing empty token in not allowed for Active directory, fix this
	userCasbinRoles, err := impl.CheckUserRoles(int32(userId), "")
	if err != nil {
		return isSuperAdmin, err
	}
	//if user which going to updated is super admin, action performing user also be super admin
	for _, item := range userCasbinRoles {
		if item == bean.SUPERADMIN {
			isSuperAdmin = true
			break
		}
	}
	return isSuperAdmin, nil
}

func (impl UserServiceImpl) IsSuperAdmin(userId int, token string) (bool, error) {
	//validating if action user is not admin and trying to update user who has super admin polices, return 403
	isSuperAdmin := false
	userCasbinRoles, err := impl.CheckUserRoles(int32(userId), token)
	if err != nil {
		return isSuperAdmin, err
	}
	//if user which going to updated is super admin, action performing user also be super admin
	for _, item := range userCasbinRoles {
		if item == bean.SUPERADMIN {
			isSuperAdmin = true
			break
		}
	}
	return isSuperAdmin, nil
}

func (impl UserServiceImpl) GetByIdIncludeDeleted(id int32) (*bean.UserInfo, error) {
	model, err := impl.userRepository.GetByIdIncludeDeleted(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	response := &bean.UserInfo{
		Id:      model.Id,
		EmailId: model.EmailId,
	}
	return response, nil
}

func (impl UserServiceImpl) saveUserAudit(r *http.Request, userId int32) {
	clientIp := util2.GetClientIP(r)
	userAudit := &UserAudit{
		UserId:    userId,
		ClientIp:  clientIp,
		CreatedOn: time.Now(),
		UpdatedOn: time.Now(),
	}
	impl.userAuditService.Save(userAudit)
}

func (impl UserServiceImpl) checkGroupAuth(groupName string, token string, managerAuth func(resource, token string, object string) bool, isActionUserSuperAdmin bool) bool {
	//check permission for group which is going to add/eliminate
	roles, err := impl.roleGroupRepository.GetRolesByGroupCasbinName(groupName)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return false
	}
	hasAccessToGroup := true
	for _, role := range roles {
		if role.AccessType == bean.APP_ACCESS_TYPE_HELM && !isActionUserSuperAdmin {
			hasAccessToGroup = false
		}
		if len(role.Team) > 0 {
			rbacObject := fmt.Sprintf("%s", role.Team)
			isValidAuth := managerAuth(casbin2.ResourceUser, token, rbacObject)
			if !isValidAuth {
				hasAccessToGroup = false
			}
		}
		if role.Entity == bean.CLUSTER_ENTITIY && !isActionUserSuperAdmin {
			isValidAuth := impl.userCommonService.CheckRbacForClusterEntity(role.Cluster, role.Namespace, role.Group, role.Kind, role.Resource, token, managerAuth)
			if !isValidAuth {
				hasAccessToGroup = false
			}
		}

	}
	return hasAccessToGroup
}

func (impl UserServiceImpl) GetRoleFiltersByGroupNames(groupNames []string) ([]bean.RoleFilter, error) {
	roles, err := impl.roleGroupRepository.GetRolesByGroupNames(groupNames)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting roles by group names", "err", err)
		return nil, err
	}

	return impl.getRoleFiltersFromRoles(roles)
}

func (impl UserServiceImpl) GetRoleFiltersByGroupCasbinNames(groupNames []string) ([]bean.RoleFilter, error) {
	roles, err := impl.roleGroupRepository.GetRolesByGroupCasbinNames(groupNames)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting roles by group names", "err", err)
		return nil, err
	}
	return impl.getRoleFiltersFromRoles(roles)
}

func (impl *UserServiceImpl) getRoleFiltersFromRoles(roles []*repository.RoleModel) ([]bean.RoleFilter, error) {
	var roleFilters []bean.RoleFilter
	roleFilterMap := make(map[string]*bean.RoleFilter)
	for _, role := range roles {
		key := impl.userCommonService.GetUniqueKeyForAllEntity(*role)
		if _, ok := roleFilterMap[key]; ok {
			impl.userCommonService.BuildRoleFilterForAllTypes(roleFilterMap, *role, key)
		} else {
			roleFilterMap[key] = &bean.RoleFilter{
				Entity:      role.Entity,
				Team:        role.Team,
				Environment: role.Environment,
				EntityName:  role.EntityName,
				Action:      role.Action,
				AccessType:  role.AccessType,
				Cluster:     role.Cluster,
				Namespace:   role.Namespace,
				Group:       role.Group,
				Kind:        role.Kind,
				Resource:    role.Resource,
				Workflow:    role.Workflow,
			}

		}
	}
	for _, v := range roleFilterMap {
		if v.Action == "super-admin" {
			continue
		}
		roleFilters = append(roleFilters, *v)
	}
	roleFilters = impl.userCommonService.MergeCustomRoleFilters(roleFilters)
	for index, roleFilter := range roleFilters {
		if roleFilter.Entity == "" {
			roleFilters[index].Entity = bean2.ENTITY_APPS
		}
		if roleFilter.Entity == bean2.ENTITY_APPS && roleFilter.AccessType == "" {
			roleFilters[index].AccessType = bean2.DEVTRON_APP
		}
	}
	return roleFilters, nil
}

func (impl *UserServiceImpl) createOrUpdateUserRolesForOtherEntity(roleFilter bean.RoleFilter, userId int32, model *repository.UserModel, existingRoles map[int]repository.UserRoleModel, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, entity string, capacity int) ([]casbin2.Policy, bool, error) {
	rolesChanged := false
	var policiesToBeAdded = make([]casbin2.Policy, 0, capacity)
	accessType := roleFilter.AccessType
	entityNames := strings.Split(roleFilter.EntityName, ",")
	environments := strings.Split(roleFilter.Environment, ",")
	actions := strings.Split(roleFilter.Action, ",")
	for _, environment := range environments {
		for _, entityName := range entityNames {
			for _, actionType := range actions {
				if managerAuth != nil {
					// check auth only for apps permission, skip for chart group
					rbacObject := fmt.Sprintf("%s", strings.ToLower(roleFilter.Team))
					isValidAuth := managerAuth(casbin2.ResourceUser, token, rbacObject)
					if !isValidAuth {
						continue
					}
				}
				roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, roleFilter.Approver, accessType, "", "", "", "", "", actionType, false, "")
				if err != nil {
					impl.logger.Errorw("error in getting role by all type", "err", err, "roleFilter", roleFilter)
					return policiesToBeAdded, rolesChanged, err
				}
				if roleModel.Id == 0 {
					impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
					flag, err, policiesAdded := impl.userCommonService.CreateDefaultPoliciesForAllTypes(roleFilter.Team, entityName, environment, entity, "", "", "", "", "", actionType, accessType, roleFilter.Approver, "", userId)
					if err != nil || flag == false {
						return policiesToBeAdded, rolesChanged, err
					}
					policiesToBeAdded = append(policiesToBeAdded, policiesAdded...)
					roleModel, err = impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, roleFilter.Approver, accessType, "", "", "", "", "", actionType, false, "")
					if err != nil {
						return policiesToBeAdded, rolesChanged, err
					}
					if roleModel.Id == 0 {
						continue
					}
				}
				if _, ok := existingRoles[roleModel.Id]; ok {
					//Adding policies which is removed
					policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
				} else if roleModel.Id > 0 {
					rolesChanged = true
					userRoleModel := &repository.UserRoleModel{
						UserId: model.Id,
						RoleId: roleModel.Id,
						AuditLog: sql.AuditLog{
							CreatedBy: userId,
							CreatedOn: time.Now(),
							UpdatedBy: userId,
							UpdatedOn: time.Now(),
						}}
					userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
					if err != nil {
						return nil, rolesChanged, err
					}
					policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
				}
			}
		}
	}
	return policiesToBeAdded, rolesChanged, nil
}

func (impl *UserServiceImpl) createOrUpdateUserRolesForJobsEntity(roleFilter bean.RoleFilter, userId int32, model *repository.UserModel, existingRoles map[int]repository.UserRoleModel, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, entity string, capacity int) ([]casbin2.Policy, bool, error) {

	rolesChanged := false
	actionType := roleFilter.Action
	accessType := roleFilter.AccessType
	var policiesToBeAdded = make([]casbin2.Policy, 0, capacity)
	entityNames := strings.Split(roleFilter.EntityName, ",")
	environments := strings.Split(roleFilter.Environment, ",")
	workflows := strings.Split(roleFilter.Workflow, ",")
	for _, environment := range environments {
		for _, entityName := range entityNames {
			for _, workflow := range workflows {
				if managerAuth != nil {
					// check auth only for apps permission, skip for chart group
					rbacObject := fmt.Sprintf("%s", roleFilter.Team)
					isValidAuth := managerAuth(casbin2.ResourceUser, token, rbacObject)
					if !isValidAuth {
						continue
					}
				}
				roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, false, accessType, "", "", "", "", "", actionType, false, workflow)
				if err != nil {
					impl.logger.Errorw("error in getting role by all type", "err", err, "roleFilter", roleFilter)
					return policiesToBeAdded, rolesChanged, err
				}
				if roleModel.Id == 0 {
					impl.logger.Debugw("no role found for given filter", "filter", "roleFilter", roleFilter)
					flag, err, policiesAdded := impl.userCommonService.CreateDefaultPoliciesForAllTypes(roleFilter.Team, entityName, environment, entity, "", "", "", "", "", actionType, accessType, false, workflow, userId)
					if err != nil || flag == false {
						return policiesToBeAdded, rolesChanged, err
					}
					policiesToBeAdded = append(policiesToBeAdded, policiesAdded...)
					roleModel, err = impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, false, accessType, "", "", "", "", "", actionType, false, workflow)
					if err != nil {
						return policiesToBeAdded, rolesChanged, err
					}
					if roleModel.Id == 0 {
						continue
					}
				}
				if _, ok := existingRoles[roleModel.Id]; ok {
					//Adding policies which is removed
					policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
				} else if roleModel.Id > 0 {
					rolesChanged = true
					userRoleModel := &repository.UserRoleModel{
						UserId: model.Id,
						RoleId: roleModel.Id,
						AuditLog: sql.AuditLog{
							CreatedBy: userId,
							CreatedOn: time.Now(),
							UpdatedBy: userId,
							UpdatedOn: time.Now(),
						}}
					userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
					if err != nil {
						return nil, rolesChanged, err
					}
					policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
				}
			}
		}
	}
	return policiesToBeAdded, rolesChanged, nil
}

func (impl UserServiceImpl) getRoleFiltersForGroupClaims(id int32) ([]bean.RoleFilter, error) {
	var roleFilters []bean.RoleFilter
	userGroups, err := impl.userGroupMapRepository.GetActiveByUserId(id)
	if err != nil {
		impl.logger.Errorw("error in GetActiveByUserId", "err", err, "userId", id)
		return nil, err
	}
	groupClaims := make([]string, 0, len(userGroups))
	for _, userGroup := range userGroups {
		groupClaims = append(groupClaims, userGroup.GroupName)
	}
	// checking by group casbin name (considering case insensitivity here)
	if len(groupClaims) > 0 {
		groupCasbinNames := util3.GetGroupCasbinName(groupClaims)
		groupFilters, err := impl.GetRoleFiltersByGroupCasbinNames(groupCasbinNames)
		if err != nil {
			impl.logger.Errorw("error while GetRoleFiltersByGroupNames", "error", err, "groupCasbinNames", groupCasbinNames)
			return nil, err
		}
		if len(groupFilters) > 0 {
			roleFilters = append(roleFilters, groupFilters...)
		}
	}
	return roleFilters, nil
}

func (impl UserServiceImpl) getRoleGroupsForGroupClaims(id int32) ([]bean.RoleGroup, error) {
	userGroups, err := impl.userGroupMapRepository.GetActiveByUserId(id)
	if err != nil {
		impl.logger.Errorw("error in GetActiveByUserId", "err", err, "userId", id)
		return nil, err
	}
	groupClaims := make([]string, 0, len(userGroups))
	for _, userGroup := range userGroups {
		groupClaims = append(groupClaims, userGroup.GroupName)
	}
	// checking by group casbin name (considering case insensitivity here)
	var roleGroups []bean.RoleGroup
	if len(groupClaims) > 0 {
		groupCasbinNames := util3.GetGroupCasbinName(groupClaims)
		roleGroups, err = impl.fetchRoleGroupsByGroupClaims(groupCasbinNames)
		if err != nil {
			impl.logger.Errorw("error in fetchRoleGroupsByGroupClaims ", "err", err, "groupClaims", groupClaims)
			return nil, err
		}
	}
	return roleGroups, nil
}

func (impl UserServiceImpl) getRolefiltersForDevtronManaged(model *repository.UserModel) ([]bean.RoleFilter, error) {
	_, roleFilters, filterGroups := impl.getUserMetadata(model)
	if len(filterGroups) > 0 {
		groupRoleFilters, err := impl.GetRoleFiltersByGroupNames(filterGroups)
		if err != nil {
			impl.logger.Errorw("error while GetRoleFiltersByGroupNames", "error", err, "filterGroups", filterGroups)
			return nil, err
		}
		if len(groupRoleFilters) > 0 {
			roleFilters = append(roleFilters, groupRoleFilters...)
		}
	}
	return roleFilters, nil
}

func (impl UserServiceImpl) BulkUpdateStatusForUsers(request *bean.BulkStatusUpdateRequest) (*bean.ActionResponse, error) {
	if len(request.UserIds) == 0 {
		return nil, errors.New("bad request ,no user Ids provided")
	}

	var err error
	activeStatus := request.Status == bean.Active && request.TimeToLive.IsZero()
	inactiveStatus := request.Status == bean.Inactive
	timeExpressionStatus := request.Status == bean.Active && !request.TimeToLive.IsZero()
	if activeStatus {
		// active case
		// set foreign key to null for every user
		err = impl.statusUpdateToActive(request.UserIds)
		if err != nil {
			impl.logger.Errorw("error in BulkUpdateStatusForUsers", "err", err, "status", request.Status)
			return nil, err
		}
	} else if timeExpressionStatus || inactiveStatus {
		// case: time out expression or inactive

		// getting expression from request configuration
		timeOutExpression, expressionFormat := impl.getTimeoutExpressionAndFormatforReq(timeExpressionStatus, inactiveStatus, request.TimeToLive)
		err = impl.updateOrCreateAndUpdateWindowID(request.UserIds, timeOutExpression, expressionFormat)
		if err != nil {
			impl.logger.Errorw("error in BulkUpdateStatusForUsers", "err", err, "status", request.Status)
			return nil, err
		}
	} else {
		return nil, errors.New("bad request ,status not supported")
	}

	resp := &bean.ActionResponse{
		Suceess: true,
	}
	return resp, nil
}

func (impl UserServiceImpl) statusUpdateToActive(userIds []int32) error {
	err := impl.userRepository.UpdateWindowIdToNull(userIds)
	if err != nil {
		impl.logger.Errorw("error in statusUpdateToActive", "err", err)
		return err
	}
	return nil
}

func (impl UserServiceImpl) updateOrCreateAndUpdateWindowID(userIds []int32, timeoutExpression string, expressionFormat bean3.ExpressionFormat) error {
	idsWithWindowId, idsWithoutWindowId, windowIds, err := impl.getIdsWithAndWithoutWindowId(userIds)
	if err != nil {
		impl.logger.Errorw("error in updateOrCreateAndUpdateWindowID", "err", err, "userIds", userIds)
		return err
	}
	tx, err := impl.userRepository.StartATransaction()
	if err != nil {
		impl.logger.Errorw("error in starting a transaction", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	// case when fk exist , just update the configuration in the timeout window for fks
	if len(idsWithWindowId) > 0 && len(windowIds) > 0 {
		err = impl.timeoutWindowService.UpdateTimeoutExpressionAndFormatForIds(tx, timeoutExpression, windowIds, expressionFormat)
		if err != nil {
			impl.logger.Errorw("error in updateOrCreateAndUpdateWindowID", "err", err, "userIds", userIds)
			return err
		}

	}
	countWithoutWindowId := len(idsWithoutWindowId)
	// case when no fk exist , will create it and update the fk constraint for user
	if countWithoutWindowId > 0 {
		err = impl.createAndMapTimeoutWindow(tx, timeoutExpression, countWithoutWindowId, idsWithoutWindowId, expressionFormat)
		if err != nil {
			impl.logger.Errorw("error in updateOrCreateAndUpdateWindowID", "err", err, "userIds", userIds, "timeoutExpression", timeoutExpression)
			return err
		}
	}
	err = impl.userRepository.CommitATransaction(tx)
	if err != nil {
		impl.logger.Errorw("error in committing a transaction", "err", err)
		return err
	}
	return nil

}

func (impl UserServiceImpl) createAndMapTimeoutWindow(tx *pg.Tx, timeoutExpression string, countWithoutWindowId int, idsWithoutWindowId []int32, expressionFormat bean3.ExpressionFormat) error {
	models, err := impl.timeoutWindowService.CreateWithTimeoutExpressionAndFormat(tx, timeoutExpression, countWithoutWindowId, expressionFormat)
	if err != nil {
		impl.logger.Errorw("error in updateOrCreateAndUpdateWindowID", "err", err)
		return err
	}
	// user id vs windowId map
	windowMapping, err := impl.getUserIdVsWindowIdMapping(idsWithoutWindowId, models)
	if err != nil {
		impl.logger.Errorw("error in updateOrCreateAndUpdateWindowID", "err", err)
		return err
	}
	err = impl.updateWindowIdForId(tx, windowMapping, idsWithoutWindowId)
	if err != nil {
		impl.logger.Errorw("error in updateOrCreateAndUpdateWindowID", "err", err)
		return err
	}
	return nil
}

func (impl UserServiceImpl) getUserIdVsWindowIdMapping(userIds []int32, models []*repository2.TimeoutWindowConfiguration) (map[int32]int, error) {
	length := len(userIds)
	if length != len(models) {
		impl.logger.Errorw("created time models differ with what was required to be created")
		return nil, errors.New("something went wrong")
	}
	mapping := make(map[int32]int, len(userIds))
	for i := 0; i < length; i++ {
		mapping[userIds[i]] = models[i].Id
	}
	return mapping, nil

}

func (impl UserServiceImpl) getIdsWithAndWithoutWindowId(userIds []int32) ([]int32, []int32, []int, error) {
	// Get all users with users ids to check for which users fk exist
	users, err := impl.userRepository.GetByIds(userIds)
	if err != nil {
		impl.logger.Errorw("error in statusUpdateToActiveWithTimeExpression ", "err", err, "userIds", userIds)
		return []int32{}, []int32{}, []int{}, err
	}
	idsWithWindowId, idsWithoutWindowId, windowIds := impl.getUserIdsAndWindowIds(users)
	return idsWithWindowId, idsWithoutWindowId, windowIds, nil

}

func (impl UserServiceImpl) getUserIdsAndWindowIds(users []repository.UserModel) ([]int32, []int32, []int) {
	totalLen := len(users)
	idsWithWindowId := make([]int32, 0, totalLen)
	idsWithoutWindowId := make([]int32, 0, totalLen)
	windowIds := make([]int, 0, totalLen)
	for _, user := range users {
		if user.TimeoutWindowConfigurationId != 0 {
			idsWithWindowId = append(idsWithWindowId, user.Id)
			windowIds = append(windowIds, user.TimeoutWindowConfigurationId)
		} else {
			idsWithoutWindowId = append(idsWithoutWindowId, user.Id)
		}
	}
	return idsWithWindowId, idsWithoutWindowId, windowIds
}

func (impl UserServiceImpl) getTimeoutExpressionAndFormatforReq(timeExpressionStatus, inactiveStatus bool, requestTime time.Time) (string, bean3.ExpressionFormat) {
	if timeExpressionStatus {
		return requestTime.String(), bean3.TimeStamp
	} else if inactiveStatus {
		return time.Time{}.String(), bean3.TimeZeroFormat
	}
	return "", bean3.TimeStamp
}

func (impl UserServiceImpl) updateWindowIdForId(tx *pg.Tx, mapping map[int32]int, idsWithoutWindowId []int32) error {
	err := impl.userRepository.UpdateTimeWindowIdInBatch(tx, idsWithoutWindowId, mapping)
	if err != nil {
		impl.logger.Errorw("updateWindowIdForId failed", "err", err, "idsWithoutWindowId", idsWithoutWindowId)
		return err
	}
	return nil
}

func (impl UserServiceImpl) getStatusAndTTL(model repository.UserModel, recordedTime time.Time) (bean.Status, time.Time) {
	status := bean.Active
	var ttlTime time.Time
	if model.TimeoutWindowConfiguration != nil && len(model.TimeoutWindowConfiguration.TimeoutWindowExpression) > 0 {
		status, ttlTime = impl.getUserStatusFromTimeoutWindowExpression(model.TimeoutWindowConfiguration.TimeoutWindowExpression, recordedTime, model.TimeoutWindowConfiguration.ExpressionFormat)
	}
	return status, ttlTime
}

func (impl UserServiceImpl) getUserStatusFromTimeoutWindowExpression(expression string, recordedTime time.Time, expressionFormat bean3.ExpressionFormat) (bean.Status, time.Time) {
	parsedTime, err := impl.getParsedTimeFromExpression(expression, expressionFormat)
	if err != nil {
		impl.logger.Errorw("error in getUserStatusFromTimeoutWindowExpression", "err", err, "expression", expression, "expressionFormat", expressionFormat)
		return bean.Inactive, parsedTime
	}
	if parsedTime.IsZero() || parsedTime.Before(recordedTime) {
		return bean.Inactive, parsedTime
	}
	return bean.Active, parsedTime
}

func (impl UserServiceImpl) getParsedTimeFromExpression(expression string, format bean3.ExpressionFormat) (time.Time, error) {
	// considering default to timestamp , will add support for other formats here in future
	switch format {
	case bean3.TimeStamp:
		return impl.parseExpressionToTime(expression)
	case bean3.TimeZeroFormat:
		// Considering format timeZeroFormat for extremities, kept it in other format but represents UTC time
		return impl.parseExpressionToTime(expression)
	default:
		return impl.parseExpressionToTime(expression)
	}

	return time.Time{}, errors.New("expression format not supported")
}

func (impl UserServiceImpl) parseExpressionToTime(expression string) (time.Time, error) {
	parsedTime, err := time.Parse(helper.TimeFormatForParsing, expression)
	if err != nil {
		impl.logger.Errorw("error in parsing time from expression :getParsedTimeFromExpression", "err", err, "expression", expression)
		return parsedTime, err
	}
	return parsedTime, err
}
