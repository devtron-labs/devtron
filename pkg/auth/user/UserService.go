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
	"fmt"
	"github.com/devtron-labs/devtron/pkg/auth/user/adapter"
	helper2 "github.com/devtron-labs/devtron/pkg/auth/user/helper"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository/helper"
	"net/http"
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
	SelfRegisterUserIfNotExists(userInfo *bean.UserInfo) ([]*bean.UserInfo, error)
	UpdateUser(userInfo *bean.UserInfo, token string, managerAuth func(resource, token string, object string) bool) (*bean.UserInfo, bool, bool, []string, error)
	GetById(id int32) (*bean.UserInfo, error)
	GetAll() ([]bean.UserInfo, error)
	GetAllWithFilters(request *bean.ListingRequest) (*bean.UserListingResponse, error)
	GetEmailFromToken(token string) (string, error)
	GetEmailById(userId int32) (string, error)
	GetLoggedInUser(r *http.Request) (int32, error)
	GetByIds(ids []int32) ([]bean.UserInfo, error)
	DeleteUser(userInfo *bean.UserInfo) (bool, error)
	BulkDeleteUsers(request *bean.BulkDeleteRequest) (bool, error)
	CheckUserRoles(id int32) ([]string, error)
	SyncOrchestratorToCasbin() (bool, error)
	GetUserByToken(context context.Context, token string) (int32, string, error)
	//IsSuperAdmin(userId int) (bool, error)
	GetByIdIncludeDeleted(id int32) (*bean.UserInfo, error)
	UserExists(emailId string) bool
	UpdateTriggerPolicyForTerminalAccess() (err error)
	GetRoleFiltersByGroupNames(groupNames []string) ([]bean.RoleFilter, error)
	SaveLoginAudit(emailId, clientIp string, id int32)
}

type UserServiceImpl struct {
	userReqLock sync.RWMutex
	//map of userId and current lock-state of their serving ability;
	//if TRUE then it means that some request is ongoing & unable to serve and FALSE then it is open to serve
	userReqState        map[int32]bool
	userAuthRepository  repository.UserAuthRepository
	logger              *zap.SugaredLogger
	userRepository      repository.UserRepository
	roleGroupRepository repository.RoleGroupRepository
	sessionManager2     *middleware.SessionManager
	userCommonService   UserCommonService
	userAuditService    UserAuditService
}

func NewUserServiceImpl(userAuthRepository repository.UserAuthRepository,
	logger *zap.SugaredLogger,
	userRepository repository.UserRepository,
	userGroupRepository repository.RoleGroupRepository,
	sessionManager2 *middleware.SessionManager, userCommonService UserCommonService, userAuditService UserAuditService) *UserServiceImpl {
	serviceImpl := &UserServiceImpl{
		userReqState:        make(map[int32]bool),
		userAuthRepository:  userAuthRepository,
		logger:              logger,
		userRepository:      userRepository,
		roleGroupRepository: userGroupRepository,
		sessionManager2:     sessionManager2,
		userCommonService:   userCommonService,
		userAuditService:    userAuditService,
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

func (impl *UserServiceImpl) validateUserRequest(userInfo *bean.UserInfo) (bool, error) {
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

func (impl *UserServiceImpl) SelfRegisterUserIfNotExists(userInfo *bean.UserInfo) ([]*bean.UserInfo, error) {
	var pass []string
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

		pass = append(pass, emailId)
		userInfo.EmailId = emailId
		userInfo.Exist = dbUser.Active
		userResponse = append(userResponse, &bean.UserInfo{Id: userInfo.Id, EmailId: emailId, Groups: userInfo.Groups, RoleFilters: userInfo.RoleFilters, SuperAdmin: userInfo.SuperAdmin})
	}

	if len(policies) > 0 {
		//loading policy for safety
		casbin2.LoadPolicy()
		pRes := casbin2.AddPolicy(policies)
		println(pRes)
		//loading policy for syncing orchestrator to casbin with newly added policies
		casbin2.LoadPolicy()
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return userResponse, nil
}

func (impl *UserServiceImpl) saveUser(userInfo *bean.UserInfo, emailId string) (*bean.UserInfo, error) {
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

func (impl *UserServiceImpl) CreateUser(userInfo *bean.UserInfo, token string, managerAuth func(resource, token string, object string) bool) ([]*bean.UserInfo, error) {

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

func (impl *UserServiceImpl) updateUserIfExists(userInfo *bean.UserInfo, dbUser *repository.UserModel, emailId string,
	token string, managerAuth func(resource, token, object string) bool) (*bean.UserInfo, error) {
	updateUserInfo, err := impl.GetById(dbUser.Id)
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
	updateUserInfo, _, _, _, err = impl.UpdateUser(updateUserInfo, token, managerAuth)
	if err != nil {
		impl.logger.Errorw("error while update user", "error", err)
		return nil, err
	}
	return userInfo, nil
}

func (impl *UserServiceImpl) createUserIfNotExists(userInfo *bean.UserInfo, emailId string, token string, managerAuth func(resource string, token string, object string) bool) (*bean.UserInfo, error) {
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

		isSuperAdmin, err := impl.IsSuperAdmin(int(userInfo.UserId))
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
		roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes("", "", "", "", bean2.SUPER_ADMIN, "", "", "", "", "", "", "", false, "")
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
		pRes := casbin2.AddPolicy(policies)
		println(pRes)
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

func (impl *UserServiceImpl) CreateOrUpdateUserRolesForAllTypes(roleFilter bean.RoleFilter, userId int32, model *repository.UserModel, existingRoles map[int]repository.UserRoleModel, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, entity string, capacity int) ([]casbin2.Policy, bool, error) {
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

func (impl *UserServiceImpl) createOrUpdateUserRolesForClusterEntity(roleFilter bean.RoleFilter, userId int32, model *repository.UserModel, existingRoles map[int]repository.UserRoleModel, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, entity string, capacity int) ([]casbin2.Policy, bool, error) {

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
					roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, "", "", "", "", accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, false, "")
					if err != nil {
						return policiesToBeAdded, rolesChanged, err
					}
					if roleModel.Id == 0 {
						impl.logger.Infow("Creating Polices for cluster", resource, kind, namespace, group)
						flag, err, policiesAdded := impl.userCommonService.CreateDefaultPoliciesForAllTypes("", "", "", entity, roleFilter.Cluster, namespace, group, kind, resource, actionType, accessType, "", userId)
						if err != nil || flag == false {
							return policiesToBeAdded, rolesChanged, err
						}
						policiesToBeAdded = append(policiesToBeAdded, policiesAdded...)
						impl.logger.Infow("getting role again for cluster")
						roleModel, err = impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, "", "", "", "", accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, false, "")
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

func (impl *UserServiceImpl) mergeRoleFilter(oldR []bean.RoleFilter, newR []bean.RoleFilter) []bean.RoleFilter {
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
			Workflow:    role.Workflow,
		})
		key := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%s-%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment,
			role.EntityName, role.Action, role.AccessType, role.Cluster, role.Namespace, role.Group, role.Kind, role.Resource, role.Workflow)
		keysMap[key] = true
	}
	for _, role := range newR {
		key := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%s-%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment,
			role.EntityName, role.Action, role.AccessType, role.Cluster, role.Namespace, role.Group, role.Kind, role.Resource, role.Workflow)
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
				Workflow:    role.Workflow,
			})
		}
	}
	return roleFilters
}

func (impl *UserServiceImpl) mergeGroups(oldGroups []string, newGroups []string) []string {
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

func (impl *UserServiceImpl) UpdateUser(userInfo *bean.UserInfo, token string, managerAuth func(resource, token string, object string) bool) (*bean.UserInfo, bool, bool, []string, error) {
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
	//validating if action user is not admin and trying to update user who has super admin polices, return 403
	isUserSuperAdmin, err := impl.IsSuperAdmin(int(userInfo.Id))
	if err != nil {
		return nil, false, false, nil, err
	}
	isActionPerformingUserSuperAdmin, err := impl.IsSuperAdmin(int(userInfo.UserId))
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

	dbConnection := impl.userRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, false, false, nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	model, err := impl.userRepository.GetByIdIncludeDeleted(userInfo.Id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
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
		userCasbinRoles, err := impl.CheckUserRoles(userInfo.Id)

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
		roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes("", "", "", "", bean2.SUPER_ADMIN, "", "", "", "", "", "", "", false, "")
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
		pRes := casbin2.RemovePolicy(eliminatedPolicies)
		println(pRes)
	}
	if len(addedPolicies) > 0 {
		pRes := casbin2.AddPolicy(addedPolicies)
		println(pRes)
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

func (impl *UserServiceImpl) GetById(id int32) (*bean.UserInfo, error) {
	model, err := impl.userRepository.GetById(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	isSuperAdmin, roleFilters, filterGroups := impl.getUserMetadata(model)
	for index, roleFilter := range roleFilters {
		if roleFilter.Entity == "" {
			roleFilters[index].Entity = bean2.ENTITY_APPS
			if roleFilter.AccessType == "" {
				roleFilters[index].AccessType = bean2.DEVTRON_APP
			}
		}
	}
	response := &bean.UserInfo{
		Id:          model.Id,
		EmailId:     model.EmailId,
		RoleFilters: roleFilters,
		Groups:      filterGroups,
		SuperAdmin:  isSuperAdmin,
	}

	return response, nil
}

func (impl *UserServiceImpl) getUserMetadata(model *repository.UserModel) (bool, []bean.RoleFilter, []string) {
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
	return isSuperAdmin, roleFilters, filterGroups
}

// GetAll excluding API token user
func (impl *UserServiceImpl) GetAll() ([]bean.UserInfo, error) {
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
	if request.ShowAll {
		response, err := impl.getAllDetailedUsers(request)
		if err != nil {
			impl.logger.Errorw("error in GetAllWithFilters", "err", err)
			return nil, err
		}
		return impl.getAllDetailedUsersAdapter(response), nil
	}
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

	listingResponse, err := impl.getUserResponse(models, totalCount)
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

func (impl UserServiceImpl) getUserResponse(model []repository.UserModel, totalCount int) (*bean.UserListingResponse, error) {
	var response []bean.UserInfo
	for _, m := range model {
		lastLoginTime := adapter.GetLastLoginTime(m)
		response = append(response, bean.UserInfo{
			Id:            m.Id,
			EmailId:       m.EmailId,
			RoleFilters:   make([]bean.RoleFilter, 0),
			Groups:        make([]string, 0),
			LastLoginTime: lastLoginTime,
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

func (impl *UserServiceImpl) getAllDetailedUsers(req *bean.ListingRequest) ([]bean.UserInfo, error) {
	query := helper.GetQueryForUserListingWithFilters(req)
	models, err := impl.userRepository.GetAllExecutingQuery(query)
	if err != nil {
		impl.logger.Errorw("error in GetAllDetailedUsers", "err", err)
		return nil, err
	}
	var response []bean.UserInfo
	for _, model := range models {
		isSuperAdmin, roleFilters, filterGroups := impl.getUserMetadata(&model)
		lastLoginTime := adapter.GetLastLoginTime(model)
		for index, roleFilter := range roleFilters {
			if roleFilter.Entity == "" {
				roleFilters[index].Entity = bean2.ENTITY_APPS
			}
			if roleFilter.Entity == bean2.ENTITY_APPS && roleFilter.AccessType == "" {
				roleFilters[index].AccessType = bean2.DEVTRON_APP
			}
		}
		response = append(response, bean.UserInfo{
			Id:            model.Id,
			EmailId:       model.EmailId,
			RoleFilters:   roleFilters,
			Groups:        filterGroups,
			SuperAdmin:    isSuperAdmin,
			LastLoginTime: lastLoginTime,
		})
	}
	if len(response) == 0 {
		response = make([]bean.UserInfo, 0)
	}
	return response, nil
}

func (impl *UserServiceImpl) UserExists(emailId string) bool {
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
func (impl *UserServiceImpl) SaveLoginAudit(emailId, clientIp string, id int32) {

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

func (impl *UserServiceImpl) GetUserByEmail(emailId string) (*bean.UserInfo, error) {
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

func (impl *UserServiceImpl) GetLoggedInUser(r *http.Request) (int32, error) {
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

func (impl *UserServiceImpl) GetUserByToken(context context.Context, token string) (int32, string, error) {
	_, span := otel.Tracer("userService").Start(context, "GetUserByToken")
	email, err := impl.GetEmailFromToken(token)
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

func (impl *UserServiceImpl) GetEmailFromToken(token string) (string, error) {
	if token == "" {
		impl.logger.Infow("no token provided")
		err := &util.ApiError{
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "no token provided",
		}
		return "", err
	}

	claims, err := impl.sessionManager2.VerifyToken(token)

	if err != nil {
		impl.logger.Errorw("failed to verify token", "error", err)
		err := &util.ApiError{
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "failed to verify token",
			UserMessage:     "token verification failed while getting logged in user",
		}
		return "", err
	}

	mapClaims, err := jwt.MapClaims(claims)
	if err != nil {
		impl.logger.Errorw("failed to MapClaims", "error", err)
		err := &util.ApiError{
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "token invalid",
			UserMessage:     "token verification failed while parsing token",
		}
		return "", err
	}

	email := jwt.GetField(mapClaims, "email")
	sub := jwt.GetField(mapClaims, "sub")

	if email == "" && (sub == "admin" || sub == "admin:login") {
		email = "admin"
	}

	return email, nil
}

func (impl *UserServiceImpl) GetByIds(ids []int32) ([]bean.UserInfo, error) {
	var beans []bean.UserInfo
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

func (impl *UserServiceImpl) DeleteUser(bean *bean.UserInfo) (bool, error) {

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
	for _, item := range groups {
		flag := casbin2.DeleteRoleForUser(model.EmailId, item)
		if flag == false {
			impl.logger.Warnw("unable to delete role:", "user", model.EmailId, "role", item)
		}
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
	emailIdVsCasbinRolesMap := make(map[string][]string, totalCount)
	for _, email := range emailIds {
		casbinRoles, err := casbin2.GetRolesForUser(email)
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

func (impl *UserServiceImpl) CheckUserRoles(id int32) ([]string, error) {
	model, err := impl.userRepository.GetByIdIncludeDeleted(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	groups, err := casbin2.GetRolesForUser(model.EmailId)
	if err != nil {
		impl.logger.Errorw("No Roles Found for user", "id", model.Id)
		return nil, err
	}
	if len(groups) > 0 {
		// getting unique, handling for duplicate roles
		grps, err := impl.getUniquesRolesByGroupCasbinNames(groups)
		if err != nil {
			impl.logger.Errorw("error in getUniquesRolesByGroupCasbinNames", "err", err)
			return nil, err
		}
		groups = append(groups, grps...)
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

func (impl *UserServiceImpl) SyncOrchestratorToCasbin() (bool, error) {
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

func (impl *UserServiceImpl) IsSuperAdmin(userId int) (bool, error) {
	//validating if action user is not admin and trying to update user who has super admin polices, return 403
	isSuperAdmin := false
	userCasbinRoles, err := impl.CheckUserRoles(int32(userId))
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

func (impl *UserServiceImpl) GetByIdIncludeDeleted(id int32) (*bean.UserInfo, error) {
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

func (impl *UserServiceImpl) UpdateTriggerPolicyForTerminalAccess() (err error) {
	err = impl.userAuthRepository.UpdateTriggerPolicyForTerminalAccess()
	if err != nil {
		impl.logger.Errorw("error in updating policy for terminal access to trigger role", "err", err)
		return err
	}
	return nil
}

func (impl *UserServiceImpl) saveUserAudit(r *http.Request, userId int32) {
	clientIp := util2.GetClientIP(r)
	userAudit := &UserAudit{
		UserId:    userId,
		ClientIp:  clientIp,
		CreatedOn: time.Now(),
		UpdatedOn: time.Now(),
	}
	impl.userAuditService.Save(userAudit)
}

func (impl *UserServiceImpl) checkGroupAuth(groupName string, token string, managerAuth func(resource, token string, object string) bool, isActionUserSuperAdmin bool) bool {
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

func (impl *UserServiceImpl) GetRoleFiltersByGroupNames(groupNames []string) ([]bean.RoleFilter, error) {
	roles, err := impl.roleGroupRepository.GetRolesByGroupNames(groupNames)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting roles by group names", "err", err)
		return nil, err
	}
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
	actionType := roleFilter.Action
	accessType := roleFilter.AccessType
	entityNames := strings.Split(roleFilter.EntityName, ",")
	environments := strings.Split(roleFilter.Environment, ",")
	for _, environment := range environments {
		for _, entityName := range entityNames {
			if managerAuth != nil {
				// check auth only for apps permission, skip for chart group
				rbacObject := fmt.Sprintf("%s", roleFilter.Team)
				isValidAuth := managerAuth(casbin2.ResourceUser, token, rbacObject)
				if !isValidAuth {
					continue
				}
			}
			roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, accessType, "", "", "", "", "", actionType, false, "")
			if err != nil {
				impl.logger.Errorw("error in getting role by all type", "err", err, "roleFilter", roleFilter)
				return policiesToBeAdded, rolesChanged, err
			}
			if roleModel.Id == 0 {
				impl.logger.Debugw("no role found for given filter", "filter", "roleFilter", roleFilter)
				flag, err, policiesAdded := impl.userCommonService.CreateDefaultPoliciesForAllTypes(roleFilter.Team, entityName, environment, entity, "", "", "", "", "", actionType, accessType, "", userId)
				if err != nil || flag == false {
					return policiesToBeAdded, rolesChanged, err
				}
				policiesToBeAdded = append(policiesToBeAdded, policiesAdded...)
				roleModel, err = impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, accessType, "", "", "", "", "", actionType, false, "")
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
				roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, accessType, "", "", "", "", "", actionType, false, workflow)
				if err != nil {
					impl.logger.Errorw("error in getting role by all type", "err", err, "roleFilter", roleFilter)
					return policiesToBeAdded, rolesChanged, err
				}
				if roleModel.Id == 0 {
					impl.logger.Debugw("no role found for given filter", "filter", "roleFilter", roleFilter)
					flag, err, policiesAdded := impl.userCommonService.CreateDefaultPoliciesForAllTypes(roleFilter.Team, entityName, environment, entity, "", "", "", "", "", actionType, accessType, workflow, userId)
					if err != nil || flag == false {
						return policiesToBeAdded, rolesChanged, err
					}
					policiesToBeAdded = append(policiesToBeAdded, policiesAdded...)
					roleModel, err = impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, accessType, "", "", "", "", "", actionType, false, workflow)
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
