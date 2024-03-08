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
	"fmt"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	bean4 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin/bean"
	util4 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin/util"
	auth "github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig"
	bean5 "github.com/devtron-labs/devtron/pkg/auth/common/bean"
	helper3 "github.com/devtron-labs/devtron/pkg/auth/common/helper"
	"github.com/devtron-labs/devtron/pkg/auth/user/adapter"
	helper2 "github.com/devtron-labs/devtron/pkg/auth/user/helper"
	adapter2 "github.com/devtron-labs/devtron/pkg/auth/user/repository/adapter"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository/helper"
	util3 "github.com/devtron-labs/devtron/pkg/auth/user/util"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow"
	bean6 "github.com/devtron-labs/devtron/pkg/timeoutWindow/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	jwt2 "github.com/golang-jwt/jwt/v4"
	"golang.org/x/exp/slices"
	"net/http"
	"strconv"
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
	GetAll() ([]bean.UserInfo, error) //this is only being used for summary event for now , in use is GetAllWithFilters
	GetAllDetailedUsers() ([]bean.UserInfo, error)
	GetAllWithFilters(request *bean.ListingRequest) (*bean.UserListingResponse, error)
	GetEmailById(userId int32) (string, error)
	GetEmailAndGroupClaimsFromToken(token string) (string, []string, error)
	GetLoggedInUser(r *http.Request) (int32, error)
	GetByIds(ids []int32) ([]bean.UserInfo, error)
	DeleteUser(userInfo *bean.UserInfo) (bool, error)
	CheckUserRoles(id int32, token string) ([]string, error)
	BulkDeleteUsers(request *bean.BulkDeleteRequest) (bool, error)
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
	IsUserAdminOrManagerForAnyApp(userId int32, token string) (bool, error)
	GetFieldValuesFromToken(token string) ([]byte, error)
	BulkUpdateStatus(request *bean.BulkStatusUpdateRequest) (*bean.ActionResponse, error)
	CheckUserStatusAndUpdateLoginAudit(token string) (bool, int32, error)
	GetUserBasicDataByEmailId(emailId string) (*bean.UserInfo, error)
	GetActiveRolesAttachedToUser(emailId string, recordedTime time.Time) ([]string, error)
	GetUserRoleGroupsForEmail(emailId string, recordedTime time.Time) ([]bean.UserRoleGroup, []string, error)
	GetActiveUserRolesByEntityAndUserId(entity string, userId int32) ([]*repository.RoleModel, error)
}

type UserServiceImpl struct {
	userReqLock sync.RWMutex
	//map of userId and current lock-state of their serving ability;
	//if TRUE then it means that some request is ongoing & unable to serve and FALSE then it is open to serve
	userReqState                     map[int32]bool
	userAuthRepository               repository.UserAuthRepository
	logger                           *zap.SugaredLogger
	userRepository                   repository.UserRepository
	roleGroupRepository              repository.RoleGroupRepository
	sessionManager2                  *middleware.SessionManager
	userCommonService                UserCommonService
	userAuditService                 UserAuditService
	globalAuthorisationConfigService auth.GlobalAuthorisationConfigService
	roleGroupService                 RoleGroupService
	userGroupMapRepository           repository.UserGroupMapRepository
	enforcer                         casbin.Enforcer
	timeoutWindowService             timeoutWindow.TimeoutWindowService
}

func NewUserServiceImpl(userAuthRepository repository.UserAuthRepository,
	logger *zap.SugaredLogger,
	userRepository repository.UserRepository,
	userGroupRepository repository.RoleGroupRepository,
	sessionManager2 *middleware.SessionManager, userCommonService UserCommonService, userAuditService UserAuditService,
	globalAuthorisationConfigService auth.GlobalAuthorisationConfigService,
	roleGroupService RoleGroupService, userGroupMapRepository repository.UserGroupMapRepository,
	enforcer casbin.Enforcer,
	timeoutWindowService timeoutWindow.TimeoutWindowService,
) *UserServiceImpl {
	serviceImpl := &UserServiceImpl{
		userReqState:                     make(map[int32]bool),
		userAuthRepository:               userAuthRepository,
		logger:                           logger,
		userRepository:                   userRepository,
		roleGroupRepository:              userGroupRepository,
		sessionManager2:                  sessionManager2,
		userCommonService:                userCommonService,
		userAuditService:                 userAuditService,
		globalAuthorisationConfigService: globalAuthorisationConfigService,
		roleGroupService:                 roleGroupService,
		userGroupMapRepository:           userGroupMapRepository,
		enforcer:                         enforcer,
		timeoutWindowService:             timeoutWindowService,
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

func (impl UserServiceImpl) validateUserRequest(userInfo *bean.UserInfo) error {
	if len(userInfo.RoleFilters) == 1 &&
		userInfo.RoleFilters[0].Team == "" && userInfo.RoleFilters[0].Environment == "" && userInfo.RoleFilters[0].Action == "" {
		//skip
	} else {
		invalid := false
		conflictingRolefilters := false
		roleFilterKeyMap := make(map[string]bean.RoleFilter)
		for _, roleFilter := range userInfo.RoleFilters {
			key := util3.GetUniqueKeyForRoleFilter(roleFilter)
			if filter, ok := roleFilterKeyMap[key]; ok {
				hasTimeoutConfigChanged := !(roleFilter.TimeoutWindowExpression == filter.TimeoutWindowExpression && roleFilter.Status == filter.Status)
				if hasTimeoutConfigChanged {
					conflictingRolefilters = true
					break
				}
			}
			if len(roleFilter.Team) > 0 && len(roleFilter.Action) > 0 {
				//
			} else if len(roleFilter.Entity) > 0 { //this will pass roleFilter for clusterEntity as well as chart-group
				//
			} else {
				invalid = true
			}
			roleFilterKeyMap[key] = roleFilter
		}
		if invalid {
			err := &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, please provide role filters, "}
			return err
		}
		if conflictingRolefilters {
			err := &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, please provide non-conflicting role filters "}
			return err
		}
	}
	// validation for checking conflicting user RoleGroups
	err := impl.validateUserRoleGroupRequest(userInfo.UserRoleGroup)
	if err != nil {
		impl.logger.Errorw("error in validateUserRequest", "err", err)
		return err
	}
	return nil
}

// validateUserRoleGroupRequest :validates conflicting userRoleGroups( Conflicting: same combination with different timeout window configuration), return error if validation fails
func (impl UserServiceImpl) validateUserRoleGroupRequest(userRoleGroups []bean.UserRoleGroup) error {
	mapKey := make(map[string]bean.UserRoleGroup)
	var err error
	for _, userRoleGroup := range userRoleGroups {
		if val, ok := mapKey[userRoleGroup.RoleGroup.Name]; ok {
			hasTimeoutConfigChanged := !(userRoleGroup.TimeoutWindowExpression == val.TimeoutWindowExpression && userRoleGroup.Status == val.Status)
			if hasTimeoutConfigChanged {
				err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, please correct groups with different status"}
				break
			}
		}
		mapKey[userRoleGroup.RoleGroup.Name] = userRoleGroup
	}
	return err
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

	var policies []bean4.Policy
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
				policies = append(policies, bean4.Policy{Type: "g", Sub: bean4.Subject(userInfo.EmailId), Obj: bean4.Object(roleModel.Role)})
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

	err = impl.validateUserRequest(userInfo)
	if err != nil {
		impl.logger.Errorw("error in saveUser", "request", userInfo, "err", err)
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
		userResponse = append(userResponse, &bean.UserInfo{Id: userInfo.Id, EmailId: emailId, Groups: userInfo.Groups, RoleFilters: userInfo.RoleFilters, SuperAdmin: userInfo.SuperAdmin, UserRoleGroup: userInfo.UserRoleGroup, UserStatus: userInfo.UserStatus, TimeoutWindowExpression: userInfo.TimeoutWindowExpression})
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
	updateUserInfo.UserRoleGroup = impl.mergeUserRoleGroup(updateUserInfo.UserRoleGroup, userInfo.UserRoleGroup)
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

	err = impl.validateUserRequest(userInfo)
	if err != nil {
		impl.logger.Errorw("error in createUserIfNotExists", "request", userInfo, "err", err)
		return nil, err
	}

	//create new user in our db on d basis of info got from google api or hex. assign a basic role
	model := adapter2.GetUserModelBasicAdapter(emailId, userInfo.AccessToken, userInfo.UserType)
	model.Active = true
	model.CreatedBy = userInfo.UserId
	model.UpdatedBy = userInfo.UserId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()

	timeoutWindowConfig, err := impl.getOrCreateTimeoutWindowConfiguration(userInfo.UserStatus, userInfo.TimeoutWindowExpression, tx, userInfo.UserId)
	if err != nil {
		impl.logger.Errorw("error encountered in createUserIfNotExists", "userStatus", userInfo.UserStatus, "TimeoutWindowExpression", userInfo.TimeoutWindowExpression, "err", err)
		return nil, err
	}

	var timeoutWindowConfigId int
	if timeoutWindowConfig != nil {
		timeoutWindowConfigId = timeoutWindowConfig.Id
	}
	model.TimeoutWindowConfigurationId = timeoutWindowConfigId
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
	// check for global authorisationConfig and perform operations.
	operationDone, err := impl.checkAndPerformOperationsForGroupClaims(tx, userInfo)
	if err != nil {
		impl.logger.Errorw("error encountered in createUserIfNotExists", "err", err)
		return nil, err
	}
	// this is in case of active directory ,we are allowed to create user but not no direct permissions will be assigned
	if operationDone {
		return userInfo, nil
	}
	//loading policy for safety
	casbin2.LoadPolicy()

	//Starts Role and Mapping
	policies := make([]bean4.Policy, 0)
	if userInfo.SuperAdmin == false {
		// case: non-super admin policies addition
		policiesToBeAdded, err := impl.CreateAndAddPoliciesForNonSuperAdmin(tx, userInfo.RoleFilters, userInfo.UserRoleGroup, emailId, userInfo.UserId, token, model, managerAuth)
		if err != nil {
			impl.logger.Errorw("error encountered in createUserIfNotExists", "err", err)
			return nil, err
		}
		policies = append(policies, policiesToBeAdded...)
	} else if userInfo.SuperAdmin == true {
		// case: super admin policies addition
		err = impl.validateSuperAdminUser(userInfo.UserId, token)
		if err != nil {
			impl.logger.Errorw("error in createUserIfNotExists", "userId", userInfo.UserId, "err", err)
			return nil, err
		}
		policesToBeAdded, err := impl.CreateAndAddPoliciesForSuperAdmin(tx, userInfo.UserId, model.EmailId, model.Id)
		if err != nil {
			impl.logger.Errorw("error in createUserIfNotExists", "userId", userInfo.UserId, "err", err)
			return nil, err
		}
		policies = append(policies, policesToBeAdded...)

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

// checkAndPerformOperationsForGroupClaims : checks which globalAuthorisationConfig is active and performs operations accordingly
func (impl UserServiceImpl) checkAndPerformOperationsForGroupClaims(tx *pg.Tx, userInfo *bean.UserInfo) (bool, error) {
	isGroupClaimsActive := impl.globalAuthorisationConfigService.IsGroupClaimsConfigActive()
	isSystemManagedActive := impl.globalAuthorisationConfigService.IsDevtronSystemManagedConfigActive()
	operationDone := false
	// case: when system managed is not active and group claims is active , so only create user, not permissions.
	if !isSystemManagedActive && isGroupClaimsActive {
		userInfo.RoleFilters = []bean.RoleFilter{}
		userInfo.Groups = []string{}
		userInfo.UserRoleGroup = []bean.UserRoleGroup{}
		userInfo.SuperAdmin = false
		err := tx.Commit()
		if err != nil {
			impl.logger.Errorw("error encountered in checkAndPerformOperationsForGroupClaims", "err", err)
			return operationDone, err
		}
		operationDone = true
		return operationDone, nil
	}
	return operationDone, nil
}

// CreateAndAddPoliciesForNonSuperAdmin : iterates over every roleFilter and adds corresponding mappings in orchestrator and return polcies to be added in casbin.
func (impl UserServiceImpl) CreateAndAddPoliciesForNonSuperAdmin(tx *pg.Tx, roleFilters []bean.RoleFilter, userRoleGroup []bean.UserRoleGroup, emailId string, userLoggedInId int32, token string, model *repository.UserModel, managerAuth func(resource string, token string, object string) bool) ([]bean4.Policy, error) {
	capacity, mapping := impl.userCommonService.GetCapacityForRoleFilter(roleFilters)
	var policies = make([]bean4.Policy, 0, capacity)
	for index, roleFilter := range roleFilters {
		impl.logger.Infow("Creating Or updating User Roles for RoleFilter ")
		entity := roleFilter.Entity
		policiesToBeAdded, _, err := impl.CreateOrUpdateUserRolesForAllTypes(roleFilter, userLoggedInId, model, nil, token, managerAuth, tx, entity, mapping[index])
		if err != nil {
			impl.logger.Errorw("error in CreateAndAddPoliciesForNonSuperAdmin", "err", err, "rolefilter", roleFilters)
			return nil, err
		}
		policies = append(policies, policiesToBeAdded...)

	}

	// UserRoleGroup Addition flow starts
	policiesToBeAdded, err := impl.AddUserGroupPoliciesForCasbin(userRoleGroup, emailId, userLoggedInId, tx)
	if err != nil {
		impl.logger.Errorw("error encountered in CreateAndAddPoliciesForNonSuperAdmin", "err", err, "emailId", emailId)
		return nil, err
	}
	policies = append(policies, policiesToBeAdded...)
	// UserRoleGroup Addition flow starts ends
	return policies, nil
}

// UpdateAndAddPoliciesForNonSuperAdmin : creates corresponding mappings in orchestrator and return policies to be added, removed from casbin with flags indicating roles changes, groups changed, groups which were restricted with error if any
func (impl UserServiceImpl) UpdateAndAddPoliciesForNonSuperAdmin(tx *pg.Tx, model *repository.UserModel, token string, managerAuth func(resource string, token string, object string) bool, userInfo *bean.UserInfo, isActionPerformingUserSuperAdmin bool) ([]bean4.Policy, []bean4.Policy, []string, bool, bool, error) {
	capacity, mapping := impl.userCommonService.GetCapacityForRoleFilter(userInfo.RoleFilters)
	addedPolicies := make([]bean4.Policy, 0, capacity)
	eliminatedPolicies := make([]bean4.Policy, 0, capacity)
	restrictedGroups := []string{}
	rolesChanged := false

	//Starts Role and Mapping
	userRoleModels, err := impl.userAuthRepository.GetUserRoleMappingByUserId(model.Id)
	if err != nil {
		impl.logger.Errorw("error in UpdateAndAddPoliciesForNonSuperAdmin", "request", userInfo, "err", err)
		return nil, nil, nil, false, false, err
	}
	existingRoleIds := make(map[int]repository.UserRoleModel)
	eliminatedRoleIds := make(map[int]*repository.UserRoleModel)
	for i := range userRoleModels {
		existingRoleIds[userRoleModels[i].RoleId] = *userRoleModels[i]
		eliminatedRoleIds[userRoleModels[i].RoleId] = userRoleModels[i]
	}

	//validate role filters and user role Group for conflicts and bad payload
	err = impl.validateUserRequest(userInfo)
	if err != nil {
		impl.logger.Errorw("error in UpdateAndAddPoliciesForNonSuperAdmin", "request", userInfo, "err", err)
		return nil, nil, nil, false, false, err
	}

	// DELETE Removed Items
	items, err := impl.userCommonService.RemoveRolesAndReturnEliminatedPolicies(userInfo, existingRoleIds, eliminatedRoleIds, tx, token, managerAuth)
	if err != nil {
		impl.logger.Errorw("error in UpdateAndAddPoliciesForNonSuperAdmin", "request", userInfo, "err", err)
		return nil, nil, nil, false, false, err
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
			impl.logger.Errorw("error in UpdateAndAddPoliciesForNonSuperAdmin", "request", userInfo, "err", err)
			return nil, nil, nil, false, false, err
		}
		addedPolicies = append(addedPolicies, policiesToBeAdded...)
		rolesChanged = rolesChanged || rolesChangedFromRoleUpdate

	}

	//ROLE GROUP SETUP

	policiesToBeAdded, policiesToBeEliminated, groups, groupsModified, err := impl.createOrUpdateUserRoleGroupsPolices(userInfo.UserRoleGroup, userInfo.EmailId, token, managerAuth, isActionPerformingUserSuperAdmin, tx, userInfo.UserId)
	if err != nil {
		impl.logger.Errorw("error in UpdateAndAddPoliciesForNonSuperAdmin", "request", userInfo, "err", err)
		return nil, nil, nil, false, false, err
	}
	addedPolicies = append(addedPolicies, policiesToBeAdded...)
	restrictedGroups = append(restrictedGroups, groups...)
	eliminatedPolicies = append(eliminatedPolicies, policiesToBeEliminated...)
	// END GROUP POLICY

	return addedPolicies, eliminatedPolicies, restrictedGroups, rolesChanged, groupsModified, nil
}

// validateSuperAdminUser: validates if current user is superAdmin, if not returns a custom error.
func (impl UserServiceImpl) validateSuperAdminUser(userLoggedInId int32, token string) error {
	isSuperAdmin, err := impl.IsSuperAdmin(int(userLoggedInId), token)
	if err != nil {
		impl.logger.Errorw("error in validateSuperAdminUser ", "err", err)
		return err
	}
	if isSuperAdmin == false {
		err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to update super admin type user"}
		impl.logger.Debugw("validateSuperAdminUser failed, Not a super Admin user", "err", err, "userLoggeedId", userLoggedInId)
		return err
	}
	return nil
}

// CreateAndAddPoliciesForSuperAdmin : checks if super Admin roles else creates and creates mapping in orchestrator , returns casbin polices
func (impl UserServiceImpl) CreateAndAddPoliciesForSuperAdmin(tx *pg.Tx, userLoggedInId int32, emailId string, userModelId int32) ([]bean4.Policy, error) {
	policies := make([]bean4.Policy, 0)
	flag, err := impl.userAuthRepository.CreateRoleForSuperAdminIfNotExists(tx, userLoggedInId)
	if err != nil || flag == false {
		return nil, err
	}
	roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes("", "", "", "", bean2.SUPER_ADMIN, false, "", "", "", "", "", "", "", false, "")
	if err != nil {
		return nil, err
	}
	if roleModel.Id > 0 {
		userRoleModel := &repository.UserRoleModel{UserId: userModelId, RoleId: roleModel.Id, AuditLog: sql.AuditLog{
			CreatedBy: userLoggedInId,
			CreatedOn: time.Now(),
			UpdatedBy: userLoggedInId,
			UpdatedOn: time.Now(),
		}}
		userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
		if err != nil {
			return nil, err
		}
		policies = append(policies, bean4.Policy{Type: "g", Sub: bean4.Subject(emailId), Obj: bean4.Object(roleModel.Role)})
	}
	return policies, nil
}

// AddUserGroupPoliciesForCasbin : returns user and group mapping for casbin with timeout window config.
func (impl UserServiceImpl) AddUserGroupPoliciesForCasbin(userRoleGroup []bean.UserRoleGroup, emailId string, userLoggedInId int32, tx *pg.Tx) ([]bean4.Policy, error) {
	var policies = make([]bean4.Policy, 0)
	groupIdRoleGroupMap, err := impl.getGroupIdRoleGroupMap(userRoleGroup)
	if err != nil {
		impl.logger.Errorw("error in AddUserGroupPoliciesForCasbin", "userGroups", userRoleGroup, "err", err)
		return nil, err
	}
	// TODO : optimise this legacy flow
	for _, item := range userRoleGroup {
		userGroup := groupIdRoleGroupMap[item.RoleGroup.Id]
		twConfig, err := impl.getOrCreateTimeoutWindowConfiguration(item.Status, item.TimeoutWindowExpression, tx, userLoggedInId)
		if err != nil {
			impl.logger.Errorw("error in AddUserGroupPoliciesForCasbin", "item", item, "err", err)
			return nil, err
		}
		timeExpression, expressionFormat := helper.GetCasbinFormattedTimeAndFormat(twConfig)
		casbinPolicy := adapter.GetCasbinGroupPolicy(emailId, userGroup.CasbinName, timeExpression, expressionFormat)
		policies = append(policies, casbinPolicy)
	}
	return policies, nil
}

func (impl UserServiceImpl) getGroupIdRoleGroupMap(userRoleGroups []bean.UserRoleGroup) (map[int32]*repository.RoleGroup, error) {
	groupIdRoleGroupMap := make(map[int32]*repository.RoleGroup)
	if len(userRoleGroups) > 0 {
		ids := make([]int32, 0, len(userRoleGroups))
		for _, userGroup := range userRoleGroups {
			ids = append(ids, userGroup.RoleGroup.Id)
		}
		var err error
		groupIdRoleGroupMap, err = impl.roleGroupService.GetGroupIdVsRoleGroupMapForIds(ids)
		if err != nil {
			impl.logger.Errorw("error in getGroupIdRoleGroupMap", "userRoleGroups", userRoleGroups, "err", err)
			return nil, err
		}
	}
	return groupIdRoleGroupMap, nil
}

func (impl UserServiceImpl) CreateOrUpdateUserRolesForAllTypes(roleFilter bean.RoleFilter, userId int32, model *repository.UserModel, existingRoles map[int]repository.UserRoleModel, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, entity string, capacity int) ([]bean4.Policy, bool, error) {
	//var policiesToBeAdded []casbin2.Policy
	var policiesToBeAdded = make([]bean4.Policy, 0, capacity)
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

func (impl UserServiceImpl) createOrUpdateUserRolesForClusterEntity(roleFilter bean.RoleFilter, userId int32, model *repository.UserModel, existingRoles map[int]repository.UserRoleModel, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, entity string, capacity int) ([]bean4.Policy, bool, error) {

	//var policiesToBeAdded []casbin2.Policy
	rolesChanged := false
	namespaces := strings.Split(roleFilter.Namespace, ",")
	groups := strings.Split(roleFilter.Group, ",")
	kinds := strings.Split(roleFilter.Kind, ",")
	resources := strings.Split(roleFilter.Resource, ",")

	//capacity := len(namespaces) * len(groups) * len(kinds) * len(resources) * 2
	actionType := roleFilter.Action
	accessType := roleFilter.AccessType
	var policiesToBeAdded = make([]bean4.Policy, 0, capacity)
	timeoutWindowConfig, err := impl.getOrCreateTimeoutWindowConfiguration(roleFilter.Status, roleFilter.TimeoutWindowExpression, tx, userId)
	if err != nil {
		impl.logger.Errorw("error encountered in createOrUpdateUserRolesForClusterEntity", "roleFilter", roleFilter, "err", err)
		return policiesToBeAdded, rolesChanged, err
	}
	timeExpression, expressionFormat := helper.GetCasbinFormattedTimeAndFormat(timeoutWindowConfig)
	var timeoutWindowConfigId int
	if timeoutWindowConfig != nil {
		timeoutWindowConfigId = timeoutWindowConfig.Id
	}

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
						casbinPolicy := adapter.GetCasbinGroupPolicy(model.EmailId, roleModel.Role, timeExpression, expressionFormat)
						policiesToBeAdded = append(policiesToBeAdded, casbinPolicy)

					} else {
						if roleModel.Id > 0 {
							rolesChanged = true
							userRoleModel := &repository.UserRoleModel{
								UserId:                       model.Id,
								RoleId:                       roleModel.Id,
								TimeoutWindowConfigurationId: timeoutWindowConfigId,
								AuditLog: sql.AuditLog{
									CreatedBy: userId,
									CreatedOn: time.Now(),
									UpdatedBy: userId,
									UpdatedOn: time.Now(),
								}}
							userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
							if err != nil {
								impl.logger.Errorw("error in createOrUpdateUserRolesForClusterEntity", "userId", model.Id, "roleModelId", roleModel.Id, "err", err)
								return nil, rolesChanged, err
							}

							casbinPolicy := adapter.GetCasbinGroupPolicy(model.EmailId, roleModel.Role, timeExpression, expressionFormat)
							policiesToBeAdded = append(policiesToBeAdded, casbinPolicy)
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
			Entity:                  role.Entity,
			Team:                    role.Team,
			Environment:             role.Environment,
			EntityName:              role.EntityName,
			Action:                  role.Action,
			AccessType:              role.AccessType,
			Cluster:                 role.Cluster,
			Namespace:               role.Namespace,
			Group:                   role.Group,
			Kind:                    role.Kind,
			Resource:                role.Resource,
			Approver:                role.Approver,
			Workflow:                role.Workflow,
			Status:                  role.Status,
			TimeoutWindowExpression: role.TimeoutWindowExpression,
		})
		key := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%t-%s-%s-%s-%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment,
			role.EntityName, role.Action, role.AccessType, role.Approver, role.Cluster, role.Namespace, role.Group, role.Kind, role.Resource, role.Workflow, role.Status, role.TimeoutWindowExpression)
		keysMap[key] = true
	}
	for _, role := range newR {
		key := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%t-%s-%s-%s-%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment,
			role.EntityName, role.Action, role.AccessType, role.Approver, role.Cluster, role.Namespace, role.Group, role.Kind, role.Resource, role.Workflow, role.Status, role.TimeoutWindowExpression)
		if _, ok := keysMap[key]; !ok {
			roleFilters = append(roleFilters, bean.RoleFilter{
				Entity:                  role.Entity,
				Team:                    role.Team,
				Environment:             role.Environment,
				EntityName:              role.EntityName,
				Action:                  role.Action,
				AccessType:              role.AccessType,
				Cluster:                 role.Cluster,
				Namespace:               role.Namespace,
				Group:                   role.Group,
				Kind:                    role.Kind,
				Resource:                role.Resource,
				Approver:                role.Approver,
				Workflow:                role.Workflow,
				Status:                  role.Status,
				TimeoutWindowExpression: role.TimeoutWindowExpression,
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

// mergeUserRoleGroup : patches the existing userRoleGroups and new userRoleGroups with unique key name-status-expression, conflicting roleGroup are handled in validations.
func (impl UserServiceImpl) mergeUserRoleGroup(oldUserRoleGroups []bean.UserRoleGroup, newUserRoleGroups []bean.UserRoleGroup) []bean.UserRoleGroup {
	finalUserRoleGroups := make([]bean.UserRoleGroup, 0)
	keyMap := make(map[string]bool)
	for _, userRoleGroup := range oldUserRoleGroups {
		key := fmt.Sprintf("%s-%s-%s", userRoleGroup.RoleGroup.Name, userRoleGroup.Status, userRoleGroup.TimeoutWindowExpression)
		finalUserRoleGroups = append(finalUserRoleGroups, userRoleGroup)
		keyMap[key] = true
	}
	for _, userRoleGroup := range newUserRoleGroups {
		key := fmt.Sprintf("%s-%s-%s", userRoleGroup.RoleGroup.Name, userRoleGroup.Status, userRoleGroup.TimeoutWindowExpression)
		if _, ok := keyMap[key]; !ok {
			finalUserRoleGroups = append(finalUserRoleGroups, userRoleGroup)
		}
	}
	return finalUserRoleGroups
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

	timeoutWindowConfigId, err := impl.getTimeoutWindowID(userInfo.UserStatus, userInfo.TimeoutWindowExpression, tx, userInfo.UserId)
	if err != nil {
		impl.logger.Errorw("error in UpdateUser ", "status", userInfo.UserStatus, "timeoutExpression", userInfo.TimeoutWindowExpression, "err", err)
		return nil, false, false, nil, err
	}

	isSystemManagedActive := impl.globalAuthorisationConfigService.IsDevtronSystemManagedConfigActive()
	isUserActive := model.Active
	isApiToken := util3.CheckIfApiToken(userInfo.EmailId)
	userLevelTimeoutChanged := model.TimeoutWindowConfigurationId != timeoutWindowConfigId
	// case: system managed permissions is not active and user is inActive , mark user as active and update user timeWindowConfig id
	if !isSystemManagedActive && (!isUserActive || userLevelTimeoutChanged) && !isApiToken {
		err = impl.updateUserAndCommitTransaction(model, tx, userInfo, timeoutWindowConfigId)
		if err != nil {
			impl.logger.Errorw("error in updating user to active", "err", err, "EmailId", userInfo.EmailId)
			return userInfo, false, false, nil, err
		}
		return userInfo, false, false, nil, nil

	} else if !isSystemManagedActive && isUserActive && !isApiToken {
		// case: system managed permissions is not active and user is active , update is not allowed
		err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, user permissions are managed by your SSO Provider groups"}
		impl.logger.Errorw("Invalid request,User permissions are managed by SSO Provider", "error", err)
		return nil, false, false, nil, err
	}

	isActionPerformingUserSuperAdmin, err := impl.IsSuperAdmin(int(userInfo.UserId), token)
	if err != nil {
		return nil, false, false, nil, err
	}

	isUserSuperAdmin, err := impl.IsSuperAdmin(int(userInfo.Id), token)
	if err != nil {
		return nil, false, false, nil, err
	}

	//validating if action user is not admin and trying to update user who has super admin polices, return 403
	err = impl.validationIfAllowedToUpdateSuperAdmin(isActionPerformingUserSuperAdmin, isUserSuperAdmin, userInfo.SuperAdmin)
	if err != nil {
		impl.logger.Errorw("error in UpdateUser", "err", err, "isActionPerformingUserSuperAdmin", isActionPerformingUserSuperAdmin, "isUserSuperAdmin", isUserSuperAdmin, "payload", userInfo)
		return nil, false, false, nil, err
	}
	// case if user is superadmin already, check is user status changed,update user if yes or return custom error
	isUserStatusChanged := model.TimeoutWindowConfigurationId != timeoutWindowConfigId
	err = impl.validateIfUserAlreadyASuperAdmin(userInfo.SuperAdmin, isUserSuperAdmin)
	if err != nil && !isUserStatusChanged {
		impl.logger.Errorw("error in UpdateUser", "err", err)
		return nil, false, false, nil, err
	} else if err != nil && isUserStatusChanged {
		err2 := impl.updateUserAndCommitTransaction(model, tx, userInfo, timeoutWindowConfigId)
		if err2 != nil {
			impl.logger.Errorw("error in UpdateUser", "err", err)
			return nil, false, false, nil, err2
		}
		return userInfo, false, false, nil, nil
	}

	var eliminatedPolicies = make([]bean4.Policy, 0)
	var addedPolicies = make([]bean4.Policy, 0)
	restrictedGroups := []string{}
	rolesChanged := false
	groupsModified := false
	//loading policy for safety
	casbin2.LoadPolicy()
	if userInfo.SuperAdmin == false {
		//Starts Role and Mapping
		policiesToBeAdded, policiesToBeEliminated, groups, isRolesModified, isGroupModified, err := impl.UpdateAndAddPoliciesForNonSuperAdmin(tx, model, token, managerAuth, userInfo, isActionPerformingUserSuperAdmin)
		if err != nil {
			impl.logger.Errorw("error encountered in UpdateUser", "request", userInfo, "err", err)
			return nil, false, false, nil, err
		}
		addedPolicies = append(addedPolicies, policiesToBeAdded...)
		eliminatedPolicies = append(eliminatedPolicies, policiesToBeEliminated...)
		restrictedGroups = append(restrictedGroups, groups...)
		rolesChanged = rolesChanged || isRolesModified
		groupsModified = groupsModified || isGroupModified

	} else if userInfo.SuperAdmin == true {
		policiesToBeAdded, err := impl.CreateAndAddPoliciesForSuperAdmin(tx, userInfo.UserId, model.EmailId, model.Id)
		if err != nil {
			impl.logger.Errorw("error in UpdateUser", "userId", userInfo.UserId, "err", err)
			return nil, false, false, nil, err
		}
		addedPolicies = append(addedPolicies, policiesToBeAdded...)
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
	model.TimeoutWindowConfigurationId = timeoutWindowConfigId
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
func (impl UserServiceImpl) updateUserAndCommitTransaction(model *repository.UserModel, tx *pg.Tx, userInfo *bean.UserInfo, timeoutWindowConfigId int) error {
	err := impl.UpdateUserToActive(model, tx, userInfo.EmailId, userInfo.UserId, timeoutWindowConfigId)
	if err != nil {
		impl.logger.Errorw("error in UpdateUserAndCommitTransaction", "err", err, "EmailId", userInfo.EmailId)
		return err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in UpdateUserAndCommitTransaction", "err", err)
		return err
	}
	return nil
}

// createOrUpdateUserRoleGroupsPolices : gives policies which are to be added and which are to be eliminated from casbin, with support of timewindow Config changed fromm existing
func (impl UserServiceImpl) createOrUpdateUserRoleGroupsPolices(requestUserRoleGroups []bean.UserRoleGroup, emailId string, token string, managerAuth func(resource string, token string, object string) bool, isActionPerformingUserSuperAdmin bool, tx *pg.Tx, loggedInUser int32) ([]bean4.Policy, []bean4.Policy, []string, bool, error) {
	// getting existing userRoleGroups mapped to user by email with status according to current time
	userRoleGroups, _, err := impl.GetUserRoleGroupsForEmail(emailId, time.Now())
	if err != nil {
		impl.logger.Errorw("error encountered in createOrUpdateUserRoleGroupsPolices", "userRoleGroups", userRoleGroups, "emailId", emailId, "err", err)
		return nil, nil, nil, false, err
	}
	// initialisation
	newUserRoleGroupMap := make(map[int32]bean.UserRoleGroup, len(requestUserRoleGroups))
	oldUserRoleGroupMap := make(map[int32]bean.UserRoleGroup, len(userRoleGroups))
	restrictedGroups := make([]string, 0)
	groupsModified := false
	addedPolicies := make([]bean4.Policy, 0)
	eliminatedPolicies := make([]bean4.Policy, 0)

	// map for existing user role group mapping
	for _, userRoleGroup := range userRoleGroups {
		oldUserRoleGroupMap[userRoleGroup.RoleGroup.Id] = userRoleGroup
	}
	// bulk fetch roleGroups(which came in request)
	groupIdRoleGroupMap, err := impl.getGroupIdRoleGroupMap(requestUserRoleGroups)
	if err != nil {
		impl.logger.Errorw("error in createOrUpdateUserRoleGroupsPolices", "requestUserRoleGroups", requestUserRoleGroups, "err", err)
		return nil, nil, nil, false, err
	}
	// adding polices which are not existing in the system
	for _, userRoleGroup := range requestUserRoleGroups {
		roleGroup, ok := groupIdRoleGroupMap[userRoleGroup.RoleGroup.Id]
		if !ok {
			// doing this as to handle the deleted group in request
			continue
		}
		newUserRoleGroupMap[userRoleGroup.RoleGroup.Id] = userRoleGroup
		if val, ok := oldUserRoleGroupMap[userRoleGroup.RoleGroup.Id]; !ok {
			// case: when newly polices has to be created previously dont exist
			policiesToBeAdded, restrictedGroupsToBeAdded, isGroupModified, err := impl.CheckAccessAndReturnAdditionPolices(roleGroup.CasbinName, token, isActionPerformingUserSuperAdmin, tx, loggedInUser, emailId, userRoleGroup, managerAuth)
			if err != nil {
				impl.logger.Errorw("error in createOrUpdateUserRoleGroupsPolices", "err", err, "emailId", emailId, "userRoleGroup", userRoleGroup)
				return nil, nil, nil, false, err
			}
			addedPolicies = append(addedPolicies, policiesToBeAdded...)
			restrictedGroups = append(restrictedGroups, restrictedGroupsToBeAdded...)
			groupsModified = groupsModified || isGroupModified
		} else {
			//  case: when user group exist but just time config is changed we add polices for new configuration
			hasTimeoutChanged := helper2.HasTimeWindowChangedForUserRoleGroup(userRoleGroup, val)
			if hasTimeoutChanged {
				policiesToBeAdded, restrictedGroupsToBeAdded, isGroupModified, err := impl.CheckAccessAndReturnAdditionPolices(roleGroup.CasbinName, token, isActionPerformingUserSuperAdmin, tx, loggedInUser, emailId, userRoleGroup, managerAuth)
				if err != nil {
					impl.logger.Errorw("error in createOrUpdateUserRoleGroupsPolices", "err", err, "emailId", emailId, "userRoleGroup", userRoleGroup)
					return nil, nil, nil, false, err
				}
				addedPolicies = append(addedPolicies, policiesToBeAdded...)
				restrictedGroups = append(restrictedGroups, restrictedGroupsToBeAdded...)
				groupsModified = groupsModified || isGroupModified
			}
		}

	}
	// removing the existing policies which are not present in current request
	for _, item := range userRoleGroups {
		if val, ok := newUserRoleGroupMap[item.RoleGroup.Id]; !ok {
			// case: when existing policies has to be removed as request dont have this user roleGroup now
			policiesToBeEliminated, restrictedGroupsToBeAdded, isGroupModified := impl.CheckAccessAndReturnEliminatedPolices(token, isActionPerformingUserSuperAdmin, emailId, item, managerAuth)
			eliminatedPolicies = append(eliminatedPolicies, policiesToBeEliminated...)
			restrictedGroups = append(restrictedGroups, restrictedGroupsToBeAdded...)
			groupsModified = groupsModified || isGroupModified
		} else {
			// case: when existing policies has been given but with differnt timeoutWindow Configration , existing policies has to be removed.
			hasTimeoutChanged := helper2.HasTimeWindowChangedForUserRoleGroup(item, val)
			if hasTimeoutChanged {
				policiesToBeEliminated, restrictedGroupsToBeAdded, isGroupModified := impl.CheckAccessAndReturnEliminatedPolices(token, isActionPerformingUserSuperAdmin, emailId, item, managerAuth)
				eliminatedPolicies = append(eliminatedPolicies, policiesToBeEliminated...)
				restrictedGroups = append(restrictedGroups, restrictedGroupsToBeAdded...)
				groupsModified = groupsModified || isGroupModified
			}
		}
	}

	return addedPolicies, eliminatedPolicies, restrictedGroups, groupsModified, nil
}

// validationForSuperAdminForUpdate returns custom error when logged in user is not allowed to update user to superadmin
func (impl UserServiceImpl) validationIfAllowedToUpdateSuperAdmin(isActionPerformingUserSuperAdmin bool, isUserSuperAdmin bool, superAdminFromRequest bool) error {
	if superAdminFromRequest || isUserSuperAdmin {
		if !isActionPerformingUserSuperAdmin {
			err := &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to update super admin type user"}
			impl.logger.Errorw("Invalid request, not allow to update super admin type user", "error", err)
			return err
		}
	}
	return nil

}

// validateIfUserAlreadyASuperAdmin returns custom error when user already a super admin and request to make the user super-admin
func (impl UserServiceImpl) validateIfUserAlreadyASuperAdmin(superAdminFromRequest bool, isUserSuperAdmin bool) error {
	//if request comes to make user as a super admin or user already a super admin (who'is going to be updated), action performing user should have super admin access
	if superAdminFromRequest && isUserSuperAdmin {
		err := &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "User Already A Super Admin"}
		impl.logger.Errorw("user already a superAdmin", "error", err)
		return err
	}
	return nil
}

// CheckAccessAndReturnAdditionPolices : checks group access and return policies which are to be added in casbin
func (impl UserServiceImpl) CheckAccessAndReturnAdditionPolices(casbinName, token string, isActionPerformingUserSuperAdmin bool, tx *pg.Tx, loggedInUser int32, emailId string, userRoleGroup bean.UserRoleGroup, managerAuth func(resource string, token string, object string) bool) ([]bean4.Policy, []string, bool, error) {
	addedPolicies := make([]bean4.Policy, 0)
	restrictedGroups := make([]string, 0)
	groupsModified := false
	hasAccessToGroup := impl.checkGroupAuth(casbinName, token, managerAuth, isActionPerformingUserSuperAdmin)
	if hasAccessToGroup {
		groupsModified = true
		timeoutWindowConfig, err := impl.getOrCreateTimeoutWindowConfiguration(userRoleGroup.Status, userRoleGroup.TimeoutWindowExpression, tx, loggedInUser)
		if err != nil {
			impl.logger.Errorw("error in CheckAccessAndReturnAdditionPolices", "userRoleGroup", userRoleGroup, "err", err)
			return nil, nil, false, err
		}
		timeExpression, expressionFormat := helper.GetCasbinFormattedTimeAndFormat(timeoutWindowConfig)
		casbinPolicy := adapter.GetCasbinGroupPolicy(emailId, casbinName, timeExpression, expressionFormat)
		addedPolicies = append(addedPolicies, casbinPolicy)
	} else {
		trimmedGroup := strings.TrimPrefix(casbinName, bean5.GroupPrefix)
		restrictedGroups = append(restrictedGroups, trimmedGroup)
	}
	return addedPolicies, restrictedGroups, groupsModified, nil
}

// CheckAccessAndReturnEliminatedPolices checks access for the group and returns policies which are to be eliminateeed from casbin
func (impl UserServiceImpl) CheckAccessAndReturnEliminatedPolices(token string, isActionPerformingUserSuperAdmin bool, emailId string, userRoleGroup bean.UserRoleGroup, managerAuth func(resource string, token string, object string) bool) ([]bean4.Policy, []string, bool) {
	eliminatedPolicies := make([]bean4.Policy, 0)
	restrictedGroups := make([]string, 0)
	groupsModified := false
	//check permission for group which is going to eliminate
	hasAccessToGroup := impl.checkGroupAuth(userRoleGroup.RoleGroup.CasbinName, token, managerAuth, isActionPerformingUserSuperAdmin)
	if hasAccessToGroup {
		groupsModified = true
		// getting casbin policy only for email vs role mapping as to handle ttl cases will result in status inactive, but in casbin we have ttl expression and format
		casbinPolicy := adapter.GetCasbinGroupPolicyForEmailAndRoleOnly(emailId, userRoleGroup.RoleGroup.CasbinName)
		eliminatedPolicies = append(eliminatedPolicies, casbinPolicy)
	} else {
		trimmedGroup := strings.TrimPrefix(userRoleGroup.RoleGroup.CasbinName, bean5.GroupPrefix)
		restrictedGroups = append(restrictedGroups, trimmedGroup)
	}
	return eliminatedPolicies, restrictedGroups, groupsModified
}

func (impl UserServiceImpl) getTimeoutWindowID(status bean.Status, timeoutWindowExpression time.Time, tx *pg.Tx, loggedInUserId int32) (int, error) {
	var timeoutWindowConfigId int
	timeoutWindowConfig, err := impl.getOrCreateTimeoutWindowConfiguration(status, timeoutWindowExpression, tx, loggedInUserId)
	if err != nil {
		impl.logger.Errorw("error encountered in createUserIfNotExists", "Status", status, "TimeoutWindowExpression", timeoutWindowExpression, "err", err)
		return timeoutWindowConfigId, err
	}

	if timeoutWindowConfig != nil {
		timeoutWindowConfigId = timeoutWindowConfig.Id
	}
	return timeoutWindowConfigId, nil

}

func (impl UserServiceImpl) UpdateUserToActive(model *repository.UserModel, tx *pg.Tx, emailId string, userId int32, twcId int) error {

	model.EmailId = emailId // override case sensitivity
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userId
	model.Active = true
	model.TimeoutWindowConfigurationId = twcId
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

	isSuperAdmin, roleFilters, filterGroups, userRoleGroups := impl.getUserMetadata(model, time.Now())
	response := &bean.UserInfo{
		Id:            model.Id,
		EmailId:       model.EmailId,
		RoleFilters:   roleFilters,
		Groups:        filterGroups,
		SuperAdmin:    isSuperAdmin,
		UserRoleGroup: userRoleGroups,
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
	model, err := impl.userRepository.GetByIdWithTimeoutWindowConfig(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	var roleFilters []bean.RoleFilter
	var filterGroups []string
	var isSuperAdmin bool
	var userRoleGroups []bean.UserRoleGroup
	recordedTime := time.Now()
	isGroupClaimsActive := impl.globalAuthorisationConfigService.IsGroupClaimsConfigActive()
	isApiToken := util3.CheckIfApiToken(model.EmailId)
	if isGroupClaimsActive && !isApiToken {
		userRoleGroups, err = impl.getRoleGroupsForGroupClaims(id)
		if err != nil {
			impl.logger.Errorw("error in getRoleGroupsForGroupClaims ", "err", err, "id", id)
			return nil, err
		}
	} else {
		// Intentionally considering ad or devtron managed here to avoid conflicts
		isSuperAdmin, roleFilters, filterGroups, userRoleGroups = impl.getUserMetadata(model, recordedTime)
	}
	status, timeoutWindowExpression := getStatusAndTTL(model.TimeoutWindowConfiguration, recordedTime)
	response := &bean.UserInfo{
		Id:                      model.Id,
		EmailId:                 model.EmailId,
		RoleFilters:             roleFilters,
		Groups:                  filterGroups,
		SuperAdmin:              isSuperAdmin,
		UserRoleGroup:           userRoleGroups,
		UserStatus:              status,
		TimeoutWindowExpression: timeoutWindowExpression,
	}

	return response, nil
}

func (impl UserServiceImpl) fetchUserRoleGroupsByGroupClaims(groupClaims []string) ([]bean.UserRoleGroup, error) {
	_, roleGroups, err := impl.roleGroupService.FetchRoleGroupsWithRolesByGroupCasbinNames(groupClaims)
	if err != nil {
		impl.logger.Errorw("error in fetchRoleGroupsByGroupClaims", "err", err, "groupClaims", groupClaims)
		return nil, err
	}
	var userRoleGroups []bean.UserRoleGroup
	for _, roleGroup := range roleGroups {
		// Sending hardcoded status active and time zero here in case of Active directory
		userRoleGroups = append(userRoleGroups, adapter.GetUserRoleGroupAdapter(roleGroup, bean.Active, time.Time{}))
	}
	return userRoleGroups, err
}

func (impl UserServiceImpl) getUserMetadata(model *repository.UserModel, recordedTime time.Time) (bool, []bean.RoleFilter, []string, []bean.UserRoleGroup) {
	userRoles, err := impl.userRepository.GetRolesWithTimeoutWindowConfigurationByUserId(model.Id)
	if err != nil {
		impl.logger.Debugw("No Roles Found for user", "id", model.Id)
	}
	isSuperAdmin := false
	var roleFilters []bean.RoleFilter
	roleFilterMap := make(map[string]*bean.RoleFilter)
	for _, userRole := range userRoles {
		status, timeoutExpression := getStatusAndTTL(userRole.TimeoutWindowConfiguration, recordedTime)
		key := GetUniqueKeyForAllEntityWithTimeAndStatus(userRole.Role, status, timeoutExpression)
		if _, ok := roleFilterMap[key]; ok {
			impl.userCommonService.BuildRoleFilterForAllTypes(roleFilterMap, userRole.Role, key)
		} else {
			roleFilterMap[key] = &bean.RoleFilter{
				Entity:                  userRole.Role.Entity,
				Team:                    userRole.Role.Team,
				Environment:             userRole.Role.Environment,
				EntityName:              userRole.Role.EntityName,
				Action:                  userRole.Role.Action,
				AccessType:              userRole.Role.AccessType,
				Cluster:                 userRole.Role.Cluster,
				Namespace:               userRole.Role.Namespace,
				Group:                   userRole.Role.Group,
				Kind:                    userRole.Role.Kind,
				Resource:                userRole.Role.Resource,
				Approver:                userRole.Role.Approver,
				Workflow:                userRole.Role.Workflow,
				Status:                  status,
				TimeoutWindowExpression: timeoutExpression,
			}

		}
		if userRole.Role.Role == bean.SUPERADMIN {
			isSuperAdmin = true
		}
	}
	for _, v := range roleFilterMap {
		if v.Action == bean2.SUPER_ADMIN {
			continue
		}
		roleFilters = append(roleFilters, *v)
	}
	roleFilters = impl.userCommonService.MergeCustomRoleFilters(roleFilters)
	userRoleGroups, filterGroups, err := impl.GetUserRoleGroupsForEmail(model.EmailId, recordedTime)
	if err != nil {
		impl.logger.Errorw("error in getUserMetadata", "err", err)
	}

	if len(filterGroups) == 0 {
		filterGroups = make([]string, 0)
	}
	if len(roleFilters) == 0 {
		roleFilters = make([]bean.RoleFilter, 0)
	}
	if len(userRoleGroups) == 0 {
		userRoleGroups = make([]bean.UserRoleGroup, 0)
	}
	for index, roleFilter := range roleFilters {
		if roleFilter.Entity == "" {
			roleFilters[index].Entity = bean2.ENTITY_APPS
			if roleFilter.AccessType == "" {
				roleFilters[index].AccessType = bean2.DEVTRON_APP
			}
		}
	}
	return isSuperAdmin, roleFilters, filterGroups, userRoleGroups
}

// GetUserRoleGroupsForEmail : returns existing userRoleGroup with status and timeoutExpression with respect to recordedTime
func (impl UserServiceImpl) GetUserRoleGroupsForEmail(emailId string, recordedTime time.Time) ([]bean.UserRoleGroup, []string, error) {
	groupPolicy, err := casbin2.GetGroupsAttachedToUser(emailId)
	if err != nil {
		impl.logger.Warnw("error in getUserRoleGroupsForEmail", "emailId", emailId)
		return nil, nil, err
	}
	var filterGroups []string
	var userRoleGroups []bean.UserRoleGroup
	if len(groupPolicy) > 0 {
		userRoleGroups, filterGroups, err = impl.getUserRoleGroupAndActiveGroups(groupPolicy, recordedTime)
		if err != nil {
			impl.logger.Errorw("error in getUserRoleGroupsForEmail", "emailId", emailId)
			return nil, nil, err
		}
	}
	return userRoleGroups, filterGroups, nil

}

func (impl UserServiceImpl) getUserRoleGroupAndActiveGroups(groupPolicy []bean4.GroupPolicy, recordedTime time.Time) ([]bean.UserRoleGroup, []string, error) {
	var userRoleGroup []bean.UserRoleGroup
	var activeGroup []string
	var casbinGroupName []string
	groupCasbinNameVsStatusMap := make(map[string]bean4.GroupPolicy)
	for _, policy := range groupPolicy {
		groupCasbinNameVsStatusMap[policy.Role] = policy
		casbinGroupName = append(casbinGroupName, policy.Role)
	}
	filterGroupsModels, err := impl.roleGroupRepository.GetRoleGroupListByCasbinNames(casbinGroupName)
	if err != nil {
		impl.logger.Errorw("error in getUserRoleGroupAndActiveGroups", "casbinGroupName", casbinGroupName, "err", err)
		return userRoleGroup, activeGroup, err
	}

	for _, roleGroup := range filterGroupsModels {
		group := adapter.GetBasicRoleGroupDetailsAdapter(roleGroup.Name, roleGroup.Description, roleGroup.Id, roleGroup.CasbinName)
		status, timeWindowExpression, err := getStatusAndTimeoutExpressionFromCasbinValues(groupCasbinNameVsStatusMap[roleGroup.CasbinName].TimeoutWindowExpression, groupCasbinNameVsStatusMap[roleGroup.CasbinName].ExpressionFormat, recordedTime)
		if err != nil {
			impl.logger.Errorw("error in getUserRoleGroup", "groupPolicy", groupPolicy, "err", err)
			return userRoleGroup, activeGroup, err
		}
		userRoleGroup = append(userRoleGroup, bean.UserRoleGroup{RoleGroup: group, Status: status, TimeoutWindowExpression: timeWindowExpression})
		if status != bean.Inactive {
			activeGroup = append(activeGroup, roleGroup.Name)
		}
	}

	return userRoleGroup, activeGroup, nil
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

// GetAllWithFilters takes filter request  gives UserListingResponse as output with some operations like filter, sorting, searching,pagination support inbuilt
func (impl UserServiceImpl) GetAllWithFilters(request *bean.ListingRequest) (*bean.UserListingResponse, error) {
	//  default values will be used if not provided
	impl.userCommonService.SetDefaultValuesIfNotPresent(request, false)
	// setting filter status type
	impl.setStatusFilterType(request)
	if request.ShowAll {
		response, err := impl.getAllDetailedUsers(request)
		if err != nil {
			impl.logger.Errorw("error in GetAllWithFilters", "err", err)
			return nil, err
		}
		return impl.getAllDetailedUsersAdapter(response), nil
	}

	// Recording time here for overall consistency
	request.CurrentTime = time.Now()

	// setting count check to true for only count
	request.CountCheck = true
	// Build query from query builder
	query := helper.GetQueryForUserListingWithFilters(request)
	totalCount, err := impl.userRepository.GetCountExecutingQuery(query)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db in GetAllWithFilters", "error", err)
		return nil, err
	}

	// setting count check to false for getting data
	request.CountCheck = false

	query = helper.GetQueryForUserListingWithFilters(request)
	models, err := impl.userRepository.GetAllExecutingQuery(query)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db in GetAllWithFilters", "error", err)
		return nil, err
	}

	listingResponse, err := impl.getUserResponse(models, request.CurrentTime, totalCount)
	if err != nil {
		impl.logger.Errorw("error in GetAllWithFilters", "err", err)
		return nil, err
	}

	return listingResponse, nil

}

func (impl UserServiceImpl) getAllDetailedUsersAdapter(detailedUsers []bean.UserInfo) *bean.UserListingResponse {
	listingResponse := &bean.UserListingResponse{
		Users:      detailedUsers,
		TotalCount: len(detailedUsers),
	}
	return listingResponse
}

func (impl UserServiceImpl) setStatusFilterType(request *bean.ListingRequest) {
	if len(request.Status) == 0 {
		return
	}
	statues := request.Status
	containsActive := slices.Contains(statues, bean.Active)
	containsInactive := slices.Contains(statues, bean.Inactive)
	containsTemporaryAccess := slices.Contains(statues, bean.TemporaryAccess)
	// setting status type for all cases
	if containsActive && containsInactive && containsTemporaryAccess {
		// case when all three filters are selected
		request.StatusType = bean2.Active_Inactive_TemporaryAccess
	} else if containsActive && containsInactive {
		// case when all two filters are selected, active and inactive
		request.StatusType = bean2.Active_InActive
	} else if containsActive && containsTemporaryAccess {
		// case when all two filters are selected, active and temporaryAccess
		request.StatusType = bean2.Active_TemporaryAccess
	} else if containsInactive && containsTemporaryAccess {
		// case when all two filters are selected, inactive and temporaryAccess
		request.StatusType = bean2.Inactive_TemporaryAccess
	} else if containsActive {
		//case when active filter is selected
		request.StatusType = bean2.Active
	} else if containsInactive {
		//case when inactive filter is selected
		request.StatusType = bean2.Inactive
	} else if containsTemporaryAccess {
		//case when temporaryAccess filter is selected
		request.StatusType = bean2.TemporaryAccess
	}
}

func (impl UserServiceImpl) getUserResponse(model []repository.UserModel, recordedTime time.Time, totalCount int) (*bean.UserListingResponse, error) {
	var response []bean.UserInfo
	for _, m := range model {
		lastLoginTime := adapter.GetLastLoginTime(m)
		userStatus, ttlTime := getStatusAndTTL(m.TimeoutWindowConfiguration, recordedTime)
		response = append(response, bean.UserInfo{
			Id:                      m.Id,
			EmailId:                 m.EmailId,
			RoleFilters:             make([]bean.RoleFilter, 0),
			Groups:                  make([]string, 0),
			UserRoleGroup:           make([]bean.UserRoleGroup, 0),
			LastLoginTime:           lastLoginTime,
			UserStatus:              userStatus,
			TimeoutWindowExpression: ttlTime,
		})
	}
	if len(response) == 0 {
		response = make([]bean.UserInfo, 0)
	}

	listingResponse := &bean.UserListingResponse{
		Users:      response,
		TotalCount: totalCount,
	}
	return listingResponse, nil
}

func (impl UserServiceImpl) getAllDetailedUsers(req *bean.ListingRequest) ([]bean.UserInfo, error) {
	query := helper.GetQueryForUserListingWithFilters(req)
	models, err := impl.userRepository.GetAllExecutingQuery(query)
	if err != nil {
		impl.logger.Errorw("error in GetAllDetailedUsers", "err", err)
		return nil, err
	}

	var response []bean.UserInfo
	// recording time here for overall status consistency
	recordedTime := time.Now()

	for _, model := range models {
		isSuperAdmin, roleFilters, filterGroups, userRoleGroups := impl.getUserMetadata(&model, recordedTime)
		userStatus, ttlTime := getStatusAndTTL(model.TimeoutWindowConfiguration, recordedTime)
		lastLoginTime := adapter.GetLastLoginTime(model)
		userResp := getUserInfoAdapter(model.Id, model.EmailId, roleFilters, filterGroups, isSuperAdmin, lastLoginTime, ttlTime, userStatus, userRoleGroups)
		response = append(response, userResp)
	}
	if len(response) == 0 {
		response = make([]bean.UserInfo, 0)
	}
	return response, nil
}

func getUserInfoAdapter(id int32, emailId string, roleFilters []bean.RoleFilter, filterGroups []string, isSuperAdmin bool, lastLoginTime, ttlTime time.Time, userStatus bean.Status, userRoleGroups []bean.UserRoleGroup) bean.UserInfo {
	user := bean.UserInfo{
		Id:                      id,
		EmailId:                 emailId,
		RoleFilters:             roleFilters,
		Groups:                  filterGroups,
		SuperAdmin:              isSuperAdmin,
		LastLoginTime:           lastLoginTime,
		UserStatus:              userStatus,
		TimeoutWindowExpression: ttlTime,
		UserRoleGroup:           userRoleGroups,
	}
	return user
}

func (impl UserServiceImpl) getLastLoginTime(model repository.UserModel) time.Time {
	lastLoginTime := time.Time{}
	if model.UserAudit != nil {
		lastLoginTime = model.UserAudit.UpdatedOn
	}
	return lastLoginTime
}

func (impl UserServiceImpl) GetAllDetailedUsers() ([]bean.UserInfo, error) {
	models, err := impl.userRepository.GetAllExcludingApiTokenUser()
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	var response []bean.UserInfo
	for _, model := range models {
		isSuperAdmin, roleFilters, filterGroups, _ := impl.getUserMetadata(&model, time.Now())
		response = append(response, bean.UserInfo{
			Id:          model.Id,
			EmailId:     model.EmailId,
			RoleFilters: roleFilters,
			Groups:      filterGroups,
			SuperAdmin:  isSuperAdmin,
		})
	}
	if len(response) == 0 {
		response = make([]bean.UserInfo, 0)
	}
	return response, nil
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
		polices, err := casbin2.GetUserAttachedToRoleWithTimeoutExpressionAndFormat(groupName)
		if err != nil {
			impl.logger.Errorw("error in extractEmailIds", "err", err, "groupName", groupName)
			return emailIds, err
		}
		userEmails, err := util4.GetUsersForActivePolicy(polices)
		if err != nil {
			impl.logger.Errorw("error in extractEmailIds", "err", err, "polices", polices)
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
		user, err := impl.GetUserBasicDataByEmailId(emailId)
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

func (impl UserServiceImpl) getUserWithTimeoutWindowConfiguration(emailId string) (int32, bool, error) {
	isInactive := true
	user, err := impl.userRepository.GetUserWithTimeoutWindowConfiguration(emailId)
	if err != nil {
		err = &util.ApiError{HttpStatusCode: 401, UserMessage: "Invalid User", InternalMessage: "failed to fetch user by email id"}
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return 0, isInactive, err
	}

	if user.TimeoutWindowConfigurationId == 0 {
		//no timeout window configuration available refer infinite active
		isInactive = false
		return user.Id, isInactive, nil
	} else {
		expiryDate, err := time.Parse(bean6.TimeFormatForParsing, user.TimeoutWindowConfiguration.TimeoutWindowExpression)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: 401, UserMessage: "Invalid User", InternalMessage: "failed to parse TimeoutWindowExpression"}
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
	userInfo, err := impl.GetUserBasicDataByEmailId(email)
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
	userRolesMappingIds, err := impl.userAuthRepository.GetUserRoleMappingIdsByUserId(bean.Id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return false, err
	}
	if len(userRolesMappingIds) > 0 {
		err = impl.userAuthRepository.DeleteUserRoleMappingByIds(userRolesMappingIds, tx)
		if err != nil {
			impl.logger.Errorw("error in DeleteUser", "userRolesMappingIds", userRolesMappingIds, "err", err)
			return false, err
		}
	}
	model.Active = false
	model.UpdatedBy = bean.UserId
	model.UpdatedOn = time.Now()
	model.TimeoutWindowConfigurationId = 0
	model, err = impl.userRepository.UpdateUser(model, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}

	groupsPolicies, err := casbin2.GetRolesAndGroupsAttachedToUserWithTimeoutExpressionAndFormat(model.EmailId)
	if err != nil {
		impl.logger.Warnw("No Roles Found for user", "id", model.Id)
	}
	var eliminatedPolicies []bean4.Policy
	for _, policy := range groupsPolicies {
		flag := casbin2.DeleteRoleForUser(model.EmailId, policy.Role, policy.TimeoutWindowExpression, policy.ExpressionFormat)
		if flag == false {
			impl.logger.Warnw("unable to delete role:", "user", model.EmailId, "role", policy.Role)
		}
		eliminatedPolicies = append(eliminatedPolicies, bean4.Policy{Type: "g", Sub: bean4.Subject(model.EmailId), Obj: bean4.Object(policy.Role)})
	}
	// updating in casbin
	if len(eliminatedPolicies) > 0 {
		pRes := casbin2.RemovePolicy(eliminatedPolicies)
		impl.logger.Infow("Failed to remove policies", "policies", pRes)
	}

	return true, nil
}

// BulkDeleteUsers takes in BulkDeleteRequest and return success and error
func (impl *UserServiceImpl) BulkDeleteUsers(request *bean.BulkDeleteRequest) (bool, error) {
	// it handles ListingRequest if filters are applied will delete those users or will consider the given user ids.
	if request.ListingRequest != nil {
		filteredUserIds, err := impl.getUserIdsHonoringFilters(request.ListingRequest)
		if err != nil {
			impl.logger.Errorw("error in BulkDeleteUsers", "request", request, "err", err)
			return false, err
		}
		// setting the filtered user ids here for further processing
		request.Ids = filteredUserIds
	}
	err := impl.deleteUsersByIds(request)
	if err != nil {
		impl.logger.Errorw("error in BulkDeleteUsers", "err", err)
		return false, err
	}
	return true, nil
}

// getUserIdsHonoringFilters get the filtered user ids according to the request filters and returns userIds and error(not nil) if any exception is caught.
func (impl *UserServiceImpl) getUserIdsHonoringFilters(request *bean.ListingRequest) ([]int32, error) {
	//query to get particular models respecting filters
	query := helper.GetQueryForUserListingWithFilters(request)
	models, err := impl.userRepository.GetAllExecutingQuery(query)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db in GetAllWithFilters", "error", err)
		return nil, err
	}
	// collecting the required user ids from filtered models
	filteredUserIds := make([]int32, 0, len(models))
	for _, model := range models {
		if !helper2.IsSystemOrAdminUserByEmail(model.EmailId) {
			filteredUserIds = append(filteredUserIds, model.Id)
		}
	}
	return filteredUserIds, nil
}

// deleteUsersByIds bulk delete all the users with their user role mappings in orchestrator and user-role and user-group mappings from casbin, takes in BulkDeleteRequest request and return success and error in return
func (impl *UserServiceImpl) deleteUsersByIds(request *bean.BulkDeleteRequest) error {
	tx, err := impl.roleGroupRepository.StartATransaction()
	if err != nil {
		impl.logger.Errorw("error in starting a transaction", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	emailIds, err := impl.userRepository.GetEmailByIds(request.Ids)
	if err != nil {
		impl.logger.Errorw("error in DeleteUsersForIds", "userIds", request.Ids, "err", err)
		return err
	}

	// operations in orchestrator and getting emails ids for corresponding user ids
	err = impl.deleteMappingsFromOrchestrator(request.Ids, tx)
	if err != nil {
		impl.logger.Errorw("error encountered in deleteUsersByIds", "request", request, "err", err)
		return err
	}
	// updating models to inactive
	err = impl.userRepository.UpdateToInactiveByIds(request.Ids, tx, request.LoggedInUserId)
	if err != nil {
		impl.logger.Errorw("error encountered in DeleteUsersForIds", "err", err)
		return err
	}
	// deleting from the group mappings from casbin
	err = impl.deleteMappingsFromCasbin(emailIds, len(request.Ids))
	if err != nil {
		impl.logger.Errorw("error encountered in deleteUsersByIds", "request", request, "err", err)
		return err
	}

	err = impl.roleGroupRepository.CommitATransaction(tx)
	if err != nil {
		impl.logger.Errorw("error in committing a transaction", "err", err)
		return err
	}

	return nil
}

// deleteMappingsFromCasbin gets all mappings for all email ids and delete that mapping one by one as no bulk support from casbin library.
func (impl *UserServiceImpl) deleteMappingsFromCasbin(emailIds []string, totalCount int) error {
	emailIdVsCasbinRolesMap := make(map[string][]bean4.GroupPolicy, totalCount)
	for _, email := range emailIds {
		casbinRoles, err := casbin2.GetRolesAndGroupsAttachedToUserWithTimeoutExpressionAndFormat(email)
		if err != nil {
			impl.logger.Warnw("No Roles Found for user", "email", email, "err", err)
			return err
		}
		emailIdVsCasbinRolesMap[email] = casbinRoles
	}

	success := impl.userCommonService.DeleteRoleForUserFromCasbin(emailIdVsCasbinRolesMap)
	if !success {
		impl.logger.Errorw("error in deleting from casbin in deleteMappingsFromCasbin ", "emailIds", emailIds)
		return &util.ApiError{Code: "500", HttpStatusCode: 500, InternalMessage: "Not able to delete mappings from casbin", UserMessage: "Not able to delete mappings from casbin"}
	}
	return nil
}

// deleteMappingsFromOrchestrator takes in userIds to be deleted and transaction returns error in case of any issue else nil
func (impl *UserServiceImpl) deleteMappingsFromOrchestrator(userIds []int32, tx *pg.Tx) error {
	urmIds, err := impl.userAuthRepository.GetUserRoleMappingIdsByUserIds(userIds)
	if err != nil {
		impl.logger.Errorw("error in DeleteUsersForIds", "err", err)
		return err
	}

	if len(urmIds) > 0 {
		err = impl.userAuthRepository.DeleteUserRoleMappingByIds(urmIds, tx)
		if err != nil {
			impl.logger.Errorw("error encountered in DeleteUsersForIds", "urmIds", urmIds, "err", err)
			return err
		}
	}
	return nil
}
func (impl UserServiceImpl) CheckUserRoles(id int32, token string) ([]string, error) {
	model, err := impl.userRepository.GetByIdIncludeDeleted(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	isGroupClaimsActive := impl.globalAuthorisationConfigService.IsGroupClaimsConfigActive()
	isDevtronSystemActive := impl.globalAuthorisationConfigService.IsDevtronSystemManagedConfigActive()
	recordedTime := time.Now()
	var groups []string
	if isDevtronSystemActive || util3.CheckIfAdminOrApiToken(model.EmailId) {
		activeRoles, err := impl.GetActiveRolesAttachedToUser(model.EmailId, recordedTime)
		if err != nil {
			impl.logger.Errorw("error encountered in getActiveRolesAttachedToUser", "id", model.Id, "err", err)
			return nil, err
		}
		groups = append(groups, activeRoles...)
		if len(groups) > 0 {
			grps, err := impl.getUniquesRolesByGroupCasbinNames(groups)
			if err != nil {
				impl.logger.Errorw("error in getUniquesRolesByGroupCasbinNames", "err", err, "groups", groups)
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

// GetActiveRolesAttachedToUser returns only active roles attached to user in casbin
func (impl UserServiceImpl) GetActiveRolesAttachedToUser(emailId string, recordedTime time.Time) ([]string, error) {
	groupPolicies, err := casbin2.GetRolesAndGroupsAttachedToUserWithTimeoutExpressionAndFormat(emailId)
	if err != nil {
		impl.logger.Warnw("error in getUserRoleGroupsForEmail", "emailId", emailId)
		return nil, err
	}
	var filterRoles []string
	for _, groupPolicy := range groupPolicies {
		status, _, err := getStatusAndTimeoutExpressionFromCasbinValues(groupPolicy.TimeoutWindowExpression, groupPolicy.ExpressionFormat, recordedTime)
		if err != nil {
			impl.logger.Errorw("error in getActiveRolesAttachedToUser", "groupPolicy", groupPolicy, "err", err)
			return nil, err
		}
		if status != bean.Inactive {
			filterRoles = append(filterRoles, groupPolicy.Role)
		}
	}
	return filterRoles, nil
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
	isSuperAdmin, err := impl.IsSuperAdmin(userId, "")
	if err != nil {
		return isSuperAdmin, err
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
		key := GetUniqueKeyForAllEntity(*role)
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

func (impl *UserServiceImpl) createOrUpdateUserRolesForOtherEntity(roleFilter bean.RoleFilter, userId int32, model *repository.UserModel, existingRoles map[int]repository.UserRoleModel, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, entity string, capacity int) ([]bean4.Policy, bool, error) {
	rolesChanged := false
	var policiesToBeAdded = make([]bean4.Policy, 0, capacity)
	accessType := roleFilter.AccessType
	entityNames := strings.Split(roleFilter.EntityName, ",")
	environments := strings.Split(roleFilter.Environment, ",")
	actions := strings.Split(roleFilter.Action, ",")
	timeoutWindowConfig, err := impl.getOrCreateTimeoutWindowConfiguration(roleFilter.Status, roleFilter.TimeoutWindowExpression, tx, userId)
	if err != nil {
		impl.logger.Errorw("error encountered in createOrUpdateUserRolesForOtherEntity", "roleFilter", roleFilter, "err", err)
		return policiesToBeAdded, rolesChanged, err
	}
	timeExpression, expressionFormat := helper.GetCasbinFormattedTimeAndFormat(timeoutWindowConfig)
	var timeoutWindowConfigId int
	if timeoutWindowConfig != nil {
		timeoutWindowConfigId = timeoutWindowConfig.Id
	}
	for _, environment := range environments {
		for _, entityName := range entityNames {
			for _, actionType := range actions {
				if managerAuth != nil && entity != bean.CHART_GROUP_ENTITY {
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
					casbinPolicy := adapter.GetCasbinGroupPolicy(model.EmailId, roleModel.Role, timeExpression, expressionFormat)
					policiesToBeAdded = append(policiesToBeAdded, casbinPolicy)
				} else if roleModel.Id > 0 {
					rolesChanged = true
					userRoleModel := &repository.UserRoleModel{
						UserId:                       model.Id,
						RoleId:                       roleModel.Id,
						TimeoutWindowConfigurationId: timeoutWindowConfigId,
						AuditLog: sql.AuditLog{
							CreatedBy: userId,
							CreatedOn: time.Now(),
							UpdatedBy: userId,
							UpdatedOn: time.Now(),
						}}
					userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
					if err != nil {
						impl.logger.Errorw("error in createOrUpdateUserRolesForOtherEntity", "userId", model.Id, "roleModelId", roleModel.Id, "err", err)
						return nil, rolesChanged, err
					}
					casbinPolicy := adapter.GetCasbinGroupPolicy(model.EmailId, roleModel.Role, timeExpression, expressionFormat)
					policiesToBeAdded = append(policiesToBeAdded, casbinPolicy)
				}
			}
		}
	}
	return policiesToBeAdded, rolesChanged, nil
}

func (impl *UserServiceImpl) createOrUpdateUserRolesForJobsEntity(roleFilter bean.RoleFilter, userId int32, model *repository.UserModel, existingRoles map[int]repository.UserRoleModel, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, entity string, capacity int) ([]bean4.Policy, bool, error) {

	rolesChanged := false
	actionType := roleFilter.Action
	accessType := roleFilter.AccessType
	var policiesToBeAdded = make([]bean4.Policy, 0, capacity)
	entityNames := strings.Split(roleFilter.EntityName, ",")
	environments := strings.Split(roleFilter.Environment, ",")
	workflows := strings.Split(roleFilter.Workflow, ",")
	timeoutWindowConfig, err := impl.getOrCreateTimeoutWindowConfiguration(roleFilter.Status, roleFilter.TimeoutWindowExpression, tx, userId)
	if err != nil {
		impl.logger.Errorw("error encountered in createOrUpdateUserRolesForJobsEntity", "roleFilter", roleFilter, "err", err)
		return policiesToBeAdded, rolesChanged, err
	}
	timeExpression, expressionFormat := helper.GetCasbinFormattedTimeAndFormat(timeoutWindowConfig)
	var timeoutWindowConfigId int
	if timeoutWindowConfig != nil {
		timeoutWindowConfigId = timeoutWindowConfig.Id
	}
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
					casbinPolicy := adapter.GetCasbinGroupPolicy(model.EmailId, roleModel.Role, timeExpression, expressionFormat)
					policiesToBeAdded = append(policiesToBeAdded, casbinPolicy)
				} else if roleModel.Id > 0 {
					rolesChanged = true
					userRoleModel := &repository.UserRoleModel{
						UserId:                       model.Id,
						RoleId:                       roleModel.Id,
						TimeoutWindowConfigurationId: timeoutWindowConfigId,
						AuditLog: sql.AuditLog{
							CreatedBy: userId,
							CreatedOn: time.Now(),
							UpdatedBy: userId,
							UpdatedOn: time.Now(),
						}}
					userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
					if err != nil {
						impl.logger.Errorw("error in createOrUpdateUserRolesForJobsEntity ", "userId", model.Id, "roleModelId", roleModel.Id, "err", err)
						return nil, rolesChanged, err
					}
					casbinPolicy := adapter.GetCasbinGroupPolicy(model.EmailId, roleModel.Role, timeExpression, expressionFormat)
					policiesToBeAdded = append(policiesToBeAdded, casbinPolicy)
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

func (impl UserServiceImpl) getRoleGroupsForGroupClaims(id int32) ([]bean.UserRoleGroup, error) {
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
	var userRoleGroups []bean.UserRoleGroup
	if len(groupClaims) > 0 {
		groupCasbinNames := util3.GetGroupCasbinName(groupClaims)
		userRoleGroups, err = impl.fetchUserRoleGroupsByGroupClaims(groupCasbinNames)
		if err != nil {
			impl.logger.Errorw("error in fetchRoleGroupsByGroupClaims ", "err", err, "groupClaims", groupClaims)
			return nil, err
		}
	}
	return userRoleGroups, nil
}

func (impl UserServiceImpl) getRolefiltersForDevtronManaged(model *repository.UserModel) ([]bean.RoleFilter, error) {
	_, roleFilters, filterGroups, _ := impl.getUserMetadata(model, time.Now())
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

func (impl UserServiceImpl) IsUserAdminOrManagerForAnyApp(userId int32, token string) (bool, error) {

	isAuthorised := false

	//checking superAdmin access
	isAuthorised, err := impl.IsSuperAdmin(int(userId), token)
	if err != nil {
		impl.logger.Errorw("error in checking superAdmin access of user", "err", err, "userId", userId)
		return false, err
	}
	if !isAuthorised {
		user, err := impl.GetRoleFiltersForAUserById(userId)
		if err != nil {
			impl.logger.Errorw("error in getting user by id", "err", err, "userId", userId)
			return false, err
		}
		// ApplicationResource pe Create
		var roleFilters []bean.RoleFilter
		if user.RoleFilters != nil && len(user.RoleFilters) > 0 {
			roleFilters = append(roleFilters, user.RoleFilters...)
		}
		if len(roleFilters) > 0 {
			resourceObjects := impl.getProjectsOrAppAdminRBACNamesByAppNamesAndTeamNames(roleFilters)
			resourceObjectsMap := impl.enforcer.EnforceInBatch(token, casbin.ResourceApplications, casbin.ActionCreate, resourceObjects)
			for _, value := range resourceObjectsMap {
				if value {
					isAuthorised = true
					break
				}
			}
		}
	}
	return isAuthorised, nil
}

func (impl UserServiceImpl) getProjectOrAppAdminRBACNameByAppNameAndTeamName(appName, teamName string) string {
	if appName == "" {
		return fmt.Sprintf("%s/%s", teamName, "*")
	}
	return fmt.Sprintf("%s/%s", teamName, appName)
}

func (impl UserServiceImpl) getProjectsOrAppAdminRBACNamesByAppNamesAndTeamNames(roleFilters []bean.RoleFilter) []string {
	var resourceObjects []string
	for _, filter := range roleFilters {
		if len(filter.Team) > 0 {
			entityNames := strings.Split(filter.EntityName, ",")
			if len(entityNames) > 0 {
				for _, val := range entityNames {
					resourceName := impl.getProjectOrAppAdminRBACNameByAppNameAndTeamName(val, filter.Team)
					resourceObjects = append(resourceObjects, resourceName)
				}
			}
		}
	}
	return resourceObjects
}

// BulkUpdateStatus updates the status for the users or filters given in bulk , return ActionResponse and error in response
func (impl UserServiceImpl) BulkUpdateStatus(request *bean.BulkStatusUpdateRequest) (*bean.ActionResponse, error) {
	// it handles ListingRequest if filters are applied will delete those users or will consider the given user ids.
	if request.ListingRequest != nil {
		filteredUserIds, err := impl.getUserIdsHonoringFilters(request.ListingRequest)
		if err != nil {
			impl.logger.Errorw("error in BulkDeleteUsers", "request", request, "err", err)
			return nil, err
		}
		// setting the filtered user ids here for further processing
		request.UserIds = filteredUserIds
	}
	err := impl.bulkUpdateTimeoutWindowConfigForIds(request)
	if err != nil {
		impl.logger.Errorw("error in BulkUpdateStatus", "request", request, "err", err)
		return nil, err
	}
	resp := &bean.ActionResponse{
		Suceess: true,
	}
	return resp, nil

}
func (impl UserServiceImpl) getStatusFromExpression(status bean.Status, timeoutWindowExpression time.Time) (activeStatus, inactiveStatus, timeExpressionStatus bool) {
	activeStatus = (status == bean.Active && timeoutWindowExpression.IsZero()) || len(status) == 0
	inactiveStatus = status == bean.Inactive
	timeExpressionStatus = status == bean.Active && !timeoutWindowExpression.IsZero()
	return activeStatus, inactiveStatus, timeExpressionStatus
}

func (impl UserServiceImpl) bulkUpdateStatusForIds(request *bean.BulkStatusUpdateRequest) error {
	activeStatus, inactiveStatus, timeExpressionStatus := impl.getStatusFromExpression(request.Status, request.TimeoutWindowExpression)
	if activeStatus {
		// active case
		// set foreign key to null for every user
		err := impl.userRepository.UpdateWindowIdToNull(request.UserIds, request.LoggedInUserId, nil)
		if err != nil {
			impl.logger.Errorw("error in BulkUpdateStatusForUsers", "err", err, "status", request.Status)
			return err
		}
	} else if timeExpressionStatus || inactiveStatus {
		// case: time out expression or inactive

		// getting expression from request configuration
		timeOutExpression, expressionFormat := getTimeoutExpressionAndFormatforReq(timeExpressionStatus, inactiveStatus, request.TimeoutWindowExpression)
		err := impl.createAndUpdateWindowID(request.UserIds, timeOutExpression, expressionFormat, request.LoggedInUserId)
		if err != nil {
			impl.logger.Errorw("error in BulkUpdateStatusForUsers", "err", err, "status", request.Status)
			return err
		}
	} else {
		return &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "status not supported"}
	}

	return nil
}

func (impl UserServiceImpl) createAndUpdateWindowID(userIds []int32, timeoutExpression string, expressionFormat bean3.ExpressionFormat, loggedInUserId int32) error {
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
		err = impl.timeoutWindowService.UpdateTimeoutExpressionAndFormatForIds(tx, timeoutExpression, windowIds, expressionFormat, loggedInUserId)
		if err != nil {
			impl.logger.Errorw("error in updateOrCreateAndUpdateWindowID", "err", err, "userIds", userIds)
			return err
		}

	}
	countWithoutWindowId := len(idsWithoutWindowId)
	// case when no fk exist , will create it and update the fk constraint for user
	if countWithoutWindowId > 0 {
		err = impl.createAndMapTimeoutWindow(tx, timeoutExpression, countWithoutWindowId, idsWithoutWindowId, expressionFormat, loggedInUserId)
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

// bulkUpdateTimeoutWindowConfigForIds getOrCreates timeout Config for request and bulk update timeoutConfigId for given userIds
func (impl UserServiceImpl) bulkUpdateTimeoutWindowConfigForIds(request *bean.BulkStatusUpdateRequest) error {
	tx, err := impl.userRepository.StartATransaction()
	if err != nil {
		impl.logger.Errorw("error in starting a transaction", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	twcModel, err := impl.getOrCreateTimeoutWindowConfiguration(request.Status, request.TimeoutWindowExpression, tx, request.LoggedInUserId)
	if err != nil {
		impl.logger.Errorw("error in updateTimeoutWindowConfigIdForUserIds", "userIds", request.UserIds, "status", request.Status, "timeoutExpression", request.TimeoutWindowExpression, "err", err)
		return err
	}
	// case: for active twcModel will always be null, mapping to 0(nil) foreign key to null
	if twcModel == nil {
		err = impl.userRepository.UpdateWindowIdToNull(request.UserIds, request.LoggedInUserId, tx)
		if err != nil {
			impl.logger.Errorw("error in updateTimeoutWindowConfigIdForUserIds", "userIds", request.UserIds, "status", request.Status, "err", err)
			return err
		}
	} else {
		// case for inactive and temporary access will be updating window id in bulk.
		err = impl.userRepository.UpdateWindowIdForIds(request.UserIds, request.LoggedInUserId, twcModel.Id, tx)
		if err != nil {
			impl.logger.Errorw("error in updateTimeoutWindowConfigIdForUserIds", "userIds", request.UserIds, "status", request.Status, "timeoutConfigurationId", twcModel.Id, "err", err)
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

func (impl UserServiceImpl) createAndMapTimeoutWindow(tx *pg.Tx, timeoutExpression string, countWithoutWindowId int, idsWithoutWindowId []int32, expressionFormat bean3.ExpressionFormat, loggedInUserId int32) error {
	models, err := impl.timeoutWindowService.BulkCreateWithTimeoutExpressionAndFormat(tx, timeoutExpression, countWithoutWindowId, expressionFormat, loggedInUserId)
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
	err = impl.userRepository.UpdateTimeWindowIdInBatch(tx, idsWithoutWindowId, windowMapping, loggedInUserId)
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
		return nil, &util.ApiError{Code: "500", HttpStatusCode: 500, UserMessage: "something went wrong, length does not match for userIds given and db users"}
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
	idsWithWindowId, idsWithoutWindowId, windowIds := getUserIdsAndWindowIds(users)
	return idsWithWindowId, idsWithoutWindowId, windowIds, nil

}

func getUserIdsAndWindowIds(users []repository.UserModel) ([]int32, []int32, []int) {
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

func getTimeoutExpressionAndFormatforReq(timeExpressionStatus, inactiveStatus bool, requestTime time.Time) (string, bean3.ExpressionFormat) {
	if timeExpressionStatus {
		return requestTime.String(), bean3.TimeStamp
	} else if inactiveStatus {
		return time.Time{}.String(), bean3.TimeZeroFormat
	}
	return "", bean3.TimeStamp
}

func getStatusAndTTL(twcModel *repository2.TimeoutWindowConfiguration, recordedTime time.Time) (bean.Status, time.Time) {
	status := bean.Active
	var ttlTime time.Time
	if twcModel != nil && len(twcModel.TimeoutWindowExpression) > 0 {
		status, ttlTime = helper3.GetStatusFromTimeoutWindowExpression(twcModel.TimeoutWindowExpression, recordedTime, twcModel.ExpressionFormat)
	}
	return status, ttlTime
}

func getStatusAndTimeoutExpressionFromCasbinValues(expression, format string, recordedTime time.Time) (bean.Status, time.Time, error) {
	status := bean.Active
	var timeoutExpression time.Time
	if len(expression) > 0 && len(format) > 0 {
		expressionFormat, err := strconv.Atoi(format)
		if err != nil {
			fmt.Println("error in parsing casbin expression format", "err", err, "format", format)
			return status, timeoutExpression, err
		}
		status, timeoutExpression = helper3.GetStatusFromTimeoutWindowExpression(expression, recordedTime, bean3.ExpressionFormat(expressionFormat))
	}
	return status, timeoutExpression, nil

}

func (impl UserServiceImpl) GetUserBasicDataByEmailId(emailId string) (*bean.UserInfo, error) {
	model, err := impl.userRepository.FetchActiveUserByEmail(emailId)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	response := &bean.UserInfo{
		Id:       model.Id,
		EmailId:  model.EmailId,
		UserType: model.UserType,
	}
	return response, nil
}

func (impl UserServiceImpl) CheckUserStatusAndUpdateLoginAudit(token string) (bool, int32, error) {
	emailId, _, err := impl.GetEmailAndGroupClaimsFromToken(token)
	if err != nil {
		impl.logger.Error("unable to fetch user by token")
		err = &util.ApiError{HttpStatusCode: 401, UserMessage: "Invalid User", InternalMessage: "unable to fetch user by token"}
		return false, 0, err
	}
	userId, isInactive, err := impl.getUserWithTimeoutWindowConfiguration(emailId)
	if err != nil {
		impl.logger.Errorw("unable to fetch user by email, %s", token)
		return isInactive, userId, err
	}

	//if user is inactive, no need to store audit log
	if !isInactive {
		impl.SaveLoginAudit(emailId, "localhost", userId)
	}

	return isInactive, userId, nil
}

func (impl UserServiceImpl) getOrCreateTimeoutWindowConfiguration(status bean.Status, timeoutWindowExpression time.Time, tx *pg.Tx, loggedInUserId int32) (*repository2.TimeoutWindowConfiguration, error) {
	active, inactive, timeExpressionStatus := impl.getStatusFromExpression(status, timeoutWindowExpression)
	if active {
		return nil, nil
	} else if inactive || timeExpressionStatus {
		// getting expression from request configuration
		timeOutExpressionString, expressionFormat := getTimeoutExpressionAndFormatforReq(timeExpressionStatus, inactive, timeoutWindowExpression)
		model, err := impl.timeoutWindowService.GetOrCreateWithExpressionAndFormat(tx, timeOutExpressionString, expressionFormat, loggedInUserId)
		if err != nil {
			impl.logger.Errorw("error in createTimeoutWindowConfiguration", "status", status, "timeoutExpression", timeoutWindowExpression, "err", err)
			return nil, err
		}
		return model, nil
	}
	return nil, &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "not able to identify status", InternalMessage: "status not supported"}

}

func (impl UserServiceImpl) GetActiveUserRolesByEntityAndUserId(entity string, userId int32) ([]*repository.RoleModel, error) {
	userRoles, err := impl.userRepository.GetRolesWithTimeoutWindowConfigurationByUserIdAndEntityType(userId, entity)
	if err != nil {
		impl.logger.Errorw("error in GetActiveUserRolesByEntityAndUserId", "entity", entity, "userId", userId, "err", err)
		return nil, err
	}
	activeRolesModels := make([]*repository.RoleModel, 0, len(userRoles))
	recordedTime := time.Now()
	for _, userRole := range userRoles {
		status, _ := getStatusAndTTL(userRole.TimeoutWindowConfiguration, recordedTime)
		if status != bean.Inactive {
			activeRolesModels = append(activeRolesModels, &userRole.Role)
		}
	}
	return activeRolesModels, nil
}
