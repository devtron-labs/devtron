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
	"fmt"
	"github.com/devtron-labs/authenticator/jwt"
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	casbin2 "github.com/devtron-labs/devtron/pkg/user/casbin"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type UserService interface {
	CreateUser(userInfo *bean.UserInfo, token string, managerAuth func(token string, object string) bool) ([]*bean.UserInfo, error)
	SelfRegisterUserIfNotExists(userInfo *bean.UserInfo) ([]*bean.UserInfo, error)
	UpdateUser(userInfo *bean.UserInfo, token string, managerAuth func(token string, object string) bool) (*bean.UserInfo, error)
	GetById(id int32) (*bean.UserInfo, error)
	GetAll() ([]bean.UserInfo, error)
	GetAllDetailedUsers() ([]bean.UserInfo, error)
	GetEmailFromToken(token string) (string, error)
	GetLoggedInUser(r *http.Request) (int32, error)
	GetByIds(ids []int32) ([]bean.UserInfo, error)
	DeleteUser(userInfo *bean.UserInfo) (bool, error)
	CheckUserRoles(id int32) ([]string, error)
	SyncOrchestratorToCasbin() (bool, error)
	GetUserByToken(token string) (int32, string, error)
	IsSuperAdmin(userId int) (bool, error)
	GetByIdIncludeDeleted(id int32) (*bean.UserInfo, error)
	UserExists(emailId string) bool
	UpdateTriggerPolicyForTerminalAccess() (err error)
}

type UserServiceImpl struct {
	userAuthRepository  repository2.UserAuthRepository
	logger              *zap.SugaredLogger
	userRepository      repository2.UserRepository
	roleGroupRepository repository2.RoleGroupRepository
	sessionManager2     *middleware.SessionManager
	userCommonService   UserCommonService
	userAuditService    UserAuditService
}

func NewUserServiceImpl(userAuthRepository repository2.UserAuthRepository,
	logger *zap.SugaredLogger,
	userRepository repository2.UserRepository,
	userGroupRepository repository2.RoleGroupRepository,
	sessionManager2 *middleware.SessionManager, userCommonService UserCommonService, userAuditService UserAuditService) *UserServiceImpl {
	serviceImpl := &UserServiceImpl{
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

func (impl UserServiceImpl) validateUserRequest(userInfo *bean.UserInfo) (bool, error) {
	if len(userInfo.RoleFilters) == 1 &&
		userInfo.RoleFilters[0].Team == "" && userInfo.RoleFilters[0].Environment == "" && userInfo.RoleFilters[0].Action == "" {
		//skip
	} else {
		invalid := false
		for _, roleFilter := range userInfo.RoleFilters {
			if len(roleFilter.Team) > 0 && len(roleFilter.Action) > 0 {
				//
			} else if len(roleFilter.Entity) > 0 {
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

func (impl UserServiceImpl) SelfRegisterUserIfNotExists(userInfo *bean.UserInfo) ([]*bean.UserInfo, error) {
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
			userRoleModel := &repository2.UserRoleModel{UserId: userInfo.Id, RoleId: roleModel.Id}
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
		pRes := casbin2.AddPolicy(policies)
		println(pRes)
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return userResponse, nil
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
	model := &repository2.UserModel{
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

func (impl UserServiceImpl) CreateUser(userInfo *bean.UserInfo, token string, managerAuth func(token string, object string) bool) ([]*bean.UserInfo, error) {
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
			userInfo, err = impl.createUserIfNotExists(userInfo, emailId)
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

func (impl UserServiceImpl) updateUserIfExists(userInfo *bean.UserInfo, dbUser *repository2.UserModel, emailId string,
	token string, managerAuth func(token string, object string) bool) (*bean.UserInfo, error) {
	updateUserInfo, err := impl.GetById(dbUser.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	if dbUser.Active == false {
		updateUserInfo = &bean.UserInfo{Id: dbUser.Id}
		userInfo.Id = dbUser.Id
	}
	updateUserInfo.RoleFilters = impl.mergeRoleFilter(updateUserInfo.RoleFilters, userInfo.RoleFilters)
	updateUserInfo.Groups = impl.mergeGroups(updateUserInfo.Groups, userInfo.Groups)
	updateUserInfo.UserId = userInfo.UserId
	updateUserInfo.EmailId = emailId // override case sensitivity
	updateUserInfo, err = impl.UpdateUser(updateUserInfo, token, managerAuth)
	if err != nil {
		impl.logger.Errorw("error while update user", "error", err)
		return nil, err
	}
	return userInfo, nil
}

func (impl UserServiceImpl) createUserIfNotExists(userInfo *bean.UserInfo, emailId string) (*bean.UserInfo, error) {
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
	model := &repository2.UserModel{
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

	//Starts Role and Mapping
	var policies []casbin2.Policy
	if userInfo.SuperAdmin == false {
		for _, roleFilter := range userInfo.RoleFilters {

			if roleFilter.EntityName == "" {
				roleFilter.EntityName = "NONE"
			}
			if roleFilter.Environment == "" {
				roleFilter.Environment = "NONE"
			}
			entityNames := strings.Split(roleFilter.EntityName, ",")
			environments := strings.Split(roleFilter.Environment, ",")
			for _, environment := range environments {
				for _, entityName := range entityNames {
					if entityName == "NONE" {
						entityName = ""
					}
					if environment == "NONE" {
						environment = ""
					}
					roleModel, err := impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action, roleFilter.AccessType)
					if err != nil {
						impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
						return nil, err
					}
					if roleModel.Id == 0 {
						impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
						//userInfo.Status = "role not fount for any given filter: " + roleFilter.Team + "," + roleFilter.Environment + "," + roleFilter.Application + "," + roleFilter.Action

						if len(roleFilter.Team) > 0 {
							if roleFilter.AccessType == bean.APP_ACCESS_TYPE_HELM {
								flag, err := impl.userAuthRepository.CreateDefaultHelmPolicies(roleFilter.Team, entityName, environment, tx)
								if err != nil || flag == false {
									return nil, err
								}
							} else {
								flag, err := impl.userAuthRepository.CreateDefaultPolicies(roleFilter.Team, entityName, environment, tx)
								if err != nil || flag == false {
									return nil, err
								}
							}
							roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action, roleFilter.AccessType)
							if err != nil {
								impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
								return nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
								userInfo.Status = "role not found for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
								continue
							}
						} else if len(roleFilter.Entity) > 0 && roleFilter.Entity == "chart-group" {
							flag, err := impl.userAuthRepository.CreateDefaultPoliciesForGlobalEntity(roleFilter.Entity, entityName, roleFilter.Action, tx)
							if err != nil || flag == false {
								return nil, err
							}
							roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action, roleFilter.AccessType)
							if err != nil {
								impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
								return nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
								userInfo.Status = "role not found for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
								continue
							}
						} else {
							continue
						}
					}
					//roleModel := roleModels[0]
					if roleModel.Id > 0 {
						userRoleModel := &repository2.UserRoleModel{UserId: model.Id, RoleId: roleModel.Id}
						userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
						if err != nil {
							return nil, err
						}
						policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
					}
				}
			}
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
		flag, err := impl.userAuthRepository.CreateRoleForSuperAdminIfNotExists(tx)
		if err != nil || flag == false {
			return nil, err
		}
		roleModel, err := impl.userAuthRepository.GetRoleByFilter("", "", "", "", "super-admin", "")
		if err != nil {
			impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
			return nil, err
		}
		if roleModel.Id > 0 {
			userRoleModel := &repository2.UserRoleModel{UserId: model.Id, RoleId: roleModel.Id}
			userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
			if err != nil {
				return nil, err
			}
			policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
		}

	}
	if len(policies) > 0 {
		pRes := casbin2.AddPolicy(policies)
		println(pRes)
	}
	//Ends

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return userInfo, nil
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
		})
		key := fmt.Sprintf("%s-%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment, role.EntityName, role.Action, role.AccessType)
		keysMap[key] = true
	}
	for _, role := range newR {
		key := fmt.Sprintf("%s-%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment, role.EntityName, role.Action, role.AccessType)
		if _, ok := keysMap[key]; !ok {
			roleFilters = append(roleFilters, bean.RoleFilter{
				Entity:      role.Entity,
				Team:        role.Team,
				Environment: role.Environment,
				EntityName:  role.EntityName,
				Action:      role.Action,
				AccessType:  role.AccessType,
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

func (impl UserServiceImpl) UpdateUser(userInfo *bean.UserInfo, token string, managerAuth func(token string, object string) bool) (*bean.UserInfo, error) {
	//validating if action user is not admin and trying to update user who has super admin polices, return 403
	isUserSuperAdmin, err := impl.IsSuperAdmin(int(userInfo.Id))
	if err != nil {
		return nil, err
	}
	isActionPerformingUserSuperAdmin, err := impl.IsSuperAdmin(int(userInfo.UserId))
	if err != nil {
		return nil, err
	}
	//if request comes to make user as a super admin or user already a super admin (who'is going to be updated), action performing user should have super admin access
	if userInfo.SuperAdmin || isUserSuperAdmin {
		if !isActionPerformingUserSuperAdmin {
			err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to update super admin type user"}
			return nil, err
		}
	}

	dbConnection := impl.userRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	model, err := impl.userRepository.GetByIdIncludeDeleted(userInfo.Id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	var addedPolicies []casbin2.Policy
	var eliminatedPolicies []casbin2.Policy
	if userInfo.SuperAdmin == false {
		//Starts Role and Mapping
		userRoleModels, err := impl.userAuthRepository.GetUserRoleMappingByUserId(model.Id)
		if err != nil {
			return nil, err
		}
		existingRoleIds := make(map[int]repository2.UserRoleModel)
		eliminatedRoleIds := make(map[int]*repository2.UserRoleModel)
		for i := range userRoleModels {
			existingRoleIds[userRoleModels[i].RoleId] = *userRoleModels[i]
			eliminatedRoleIds[userRoleModels[i].RoleId] = userRoleModels[i]
		}

		//validate role filters
		_, err = impl.validateUserRequest(userInfo)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, please provide role filters"}
			return nil, err
		}

		// DELETE Removed Items
		items, err := impl.userCommonService.RemoveRolesAndReturnEliminatedPolicies(userInfo, existingRoleIds, eliminatedRoleIds, tx, token, managerAuth)
		if err != nil {
			return nil, err
		}
		eliminatedPolicies = append(eliminatedPolicies, items...)
		//Adding New Policies
		for _, roleFilter := range userInfo.RoleFilters {
			if len(roleFilter.Team) > 0 {
				// check auth only for apps permission, skip for chart group
				rbacObject := fmt.Sprintf("%s", strings.ToLower(roleFilter.Team))
				isValidAuth := managerAuth(token, rbacObject)
				if !isValidAuth {
					continue
				}
			}

			if roleFilter.EntityName == "" {
				roleFilter.EntityName = "NONE"
			}
			if roleFilter.Environment == "" {
				roleFilter.Environment = "NONE"
			}
			entityNames := strings.Split(roleFilter.EntityName, ",")
			environments := strings.Split(roleFilter.Environment, ",")
			for _, environment := range environments {
				for _, entityName := range entityNames {
					if entityName == "NONE" {
						entityName = ""
					}
					if environment == "NONE" {
						environment = ""
					}
					roleModel, err := impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action, roleFilter.AccessType)
					if err != nil {
						impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
						return nil, err
					}
					if roleModel.Id == 0 {
						impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
						userInfo.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action

						if len(roleFilter.Team) > 0 {
							if roleFilter.AccessType == bean.APP_ACCESS_TYPE_HELM {
								flag, err := impl.userAuthRepository.CreateDefaultHelmPolicies(roleFilter.Team, entityName, environment, tx)
								if err != nil || flag == false {
									return nil, err
								}
							} else {
								flag, err := impl.userAuthRepository.CreateDefaultPolicies(roleFilter.Team, entityName, environment, tx)
								if err != nil || flag == false {
									return nil, err
								}
							}
							roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action, roleFilter.AccessType)
							if err != nil {
								impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
								return nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
								userInfo.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
								continue
							}
						} else if len(roleFilter.Entity) > 0 {
							flag, err := impl.userAuthRepository.CreateDefaultPoliciesForGlobalEntity(roleFilter.Entity, entityName, roleFilter.Action, tx)
							if err != nil || flag == false {
								return nil, err
							}
							roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action, roleFilter.AccessType)
							if err != nil {
								impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
								return nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
								userInfo.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
								continue
							}
						} else {
							continue
						}
					}
					if _, ok := existingRoleIds[roleModel.Id]; ok {
						//Adding policies which is removed
						addedPolicies = append(addedPolicies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
					} else {
						if roleModel.Id > 0 {
							userRoleModel := &repository2.UserRoleModel{UserId: model.Id, RoleId: roleModel.Id}
							userRoleModel.CreatedBy = userInfo.UserId
							userRoleModel.UpdatedBy = userInfo.UserId
							userRoleModel.CreatedOn = time.Now()
							userRoleModel.UpdatedOn = time.Now()
							userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
							if err != nil {
								return nil, err
							}
							addedPolicies = append(addedPolicies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
						}
					}
				}
			}
		}

		//ROLE GROUP SETUP
		newGroupMap := make(map[string]string)
		oldGroupMap := make(map[string]string)
		userCasbinRoles, err := impl.CheckUserRoles(userInfo.Id)
		if err != nil {
			return nil, err
		}
		for _, oldItem := range userCasbinRoles {
			oldGroupMap[oldItem] = oldItem
		}
		// START GROUP POLICY
		for _, item := range userInfo.Groups {
			userGroup, err := impl.roleGroupRepository.GetRoleGroupByName(item)
			if err != nil {
				return nil, err
			}
			newGroupMap[userGroup.CasbinName] = userGroup.CasbinName
			if _, ok := oldGroupMap[userGroup.CasbinName]; !ok {
				//check permission for new group which is going to add
				hasAccessToGroup := impl.checkGroupAuth(userGroup.CasbinName, token, managerAuth, isActionPerformingUserSuperAdmin)
				if hasAccessToGroup {
					addedPolicies = append(addedPolicies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(userInfo.EmailId), Obj: casbin2.Object(userGroup.CasbinName)})
				}
			}
		}
		for _, item := range userCasbinRoles {
			if _, ok := newGroupMap[item]; !ok {
				if item != bean.SUPERADMIN {
					//check permission for group which is going to eliminate
					hasAccessToGroup := impl.checkGroupAuth(item, token, managerAuth, isActionPerformingUserSuperAdmin)
					if hasAccessToGroup {
						eliminatedPolicies = append(eliminatedPolicies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(userInfo.EmailId), Obj: casbin2.Object(item)})
					}
				}
			}
		}
		// END GROUP POLICY

	} else if userInfo.SuperAdmin == true {
		flag, err := impl.userAuthRepository.CreateRoleForSuperAdminIfNotExists(tx)
		if err != nil || flag == false {
			return nil, err
		}
		roleModel, err := impl.userAuthRepository.GetRoleByFilter("", "", "", "", "super-admin", "")
		if err != nil {
			impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
			return nil, err
		}
		if roleModel.Id > 0 {
			userRoleModel := &repository2.UserRoleModel{UserId: model.Id, RoleId: roleModel.Id}
			userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
			if err != nil {
				return nil, err
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
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

func (impl UserServiceImpl) GetById(id int32) (*bean.UserInfo, error) {
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

func (impl UserServiceImpl) getUserMetadata(model *repository2.UserModel) (bool, []bean.RoleFilter, []string) {
	roles, err := impl.userAuthRepository.GetRolesByUserId(model.Id)
	if err != nil {
		impl.logger.Debugw("No Roles Found for user", "id", model.Id)
	}
	isSuperAdmin := false
	var roleFilters []bean.RoleFilter
	roleFilterMap := make(map[string]*bean.RoleFilter)
	for _, role := range roles {
		key := ""
		if len(role.Team) > 0 {
			key = fmt.Sprintf("%s_%s_%s", role.Team, role.Action, role.AccessType)
		} else if len(role.Entity) > 0 {
			key = fmt.Sprintf("%s_%s_%s", role.Entity, role.Action)
		}
		if _, ok := roleFilterMap[key]; ok {
			envArr := strings.Split(roleFilterMap[key].Environment, ",")
			if containsArr(envArr, AllEnvironment) {
				roleFilterMap[key].Environment = AllEnvironment
			} else if !containsArr(envArr, role.Environment) {
				roleFilterMap[key].Environment = fmt.Sprintf("%s,%s", roleFilterMap[key].Environment, role.Environment)
			}
			entityArr := strings.Split(roleFilterMap[key].EntityName, ",")
			if !containsArr(entityArr, role.EntityName) {
				roleFilterMap[key].EntityName = fmt.Sprintf("%s,%s", roleFilterMap[key].EntityName, role.EntityName)
			}
		} else {
			roleFilterMap[key] = &bean.RoleFilter{
				Entity:      role.Entity,
				Team:        role.Team,
				Environment: role.Environment,
				EntityName:  role.EntityName,
				Action:      role.Action,
				AccessType:  role.AccessType,
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

func (impl UserServiceImpl) GetAllDetailedUsers() ([]bean.UserInfo, error) {
	models, err := impl.userRepository.GetAllExcludingApiTokenUser()
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	var response []bean.UserInfo
	for _, model := range models {
		isSuperAdmin, roleFilters, filterGroups := impl.getUserMetadata(&model)
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
func (impl UserServiceImpl) GetLoggedInUser(r *http.Request) (int32, error) {
	token := r.Header.Get("token")
	userId, userType, err := impl.GetUserByToken(token)
	// if user is of api-token type, then update lastUsedBy and lastUsedAt
	if err == nil && userType == bean.USER_TYPE_API_TOKEN {
		go impl.saveUserAudit(r, userId)
	}
	return userId, err
}

func (impl UserServiceImpl) GetUserByToken(token string) (int32, string, error) {
	email, err := impl.GetEmailFromToken(token)
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

func (impl UserServiceImpl) GetEmailFromToken(token string) (string, error) {
	if token == "" {
		impl.logger.Infow("no token provided", "token", token)
		err := &util.ApiError{
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "no token provided",
		}
		return "", err
	}

	//claims, err := impl.sessionManager.VerifyToken(token)
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

func (impl UserServiceImpl) GetByIds(ids []int32) ([]bean.UserInfo, error) {
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
	for _, item := range groups {
		flag := casbin2.DeleteRoleForUser(model.EmailId, item)
		if flag == false {
			impl.logger.Warnw("unable to delete role:", "user", model.EmailId, "role", item)
		}
	}

	return true, nil
}

func (impl UserServiceImpl) CheckUserRoles(id int32) ([]string, error) {
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
	impl.logger.Infow("total roles processed for sync", "len", processed)
	return true, nil
}

func (impl UserServiceImpl) IsSuperAdmin(userId int) (bool, error) {
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

func (impl UserServiceImpl) UpdateTriggerPolicyForTerminalAccess() (err error) {
	err = impl.userAuthRepository.UpdateTriggerPolicyForTerminalAccess()
	if err != nil {
		impl.logger.Errorw("error in updating policy for terminal access to trigger role", "err", err)
		return err
	}
	return nil
}

func (impl UserServiceImpl) saveUserAudit(r *http.Request, userId int32) {
	clientIp := util2.GetClientIP(r)
	userAudit := &UserAudit{
		UserId:    userId,
		ClientIp:  clientIp,
		CreatedOn: time.Now(),
	}
	impl.userAuditService.Save(userAudit)
}

func (impl UserServiceImpl) checkGroupAuth(groupName string, token string, managerAuth func(token string, object string) bool, isActionUserSuperAdmin bool) bool {
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
			rbacObject := fmt.Sprintf("%s", strings.ToLower(role.Team))
			isValidAuth := managerAuth(token, rbacObject)
			if !isValidAuth {
				hasAccessToGroup = false
				continue
			}
		}
	}
	return hasAccessToGroup
}
