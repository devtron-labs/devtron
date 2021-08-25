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
	jwt2 "github.com/argoproj/argo-cd/util/jwt"
	"github.com/argoproj/argo-cd/util/session"
	"github.com/devtron-labs/devtron/api/bean"
	session2 "github.com/devtron-labs/devtron/client/argocdServer/session"
	casbin2 "github.com/devtron-labs/devtron/internal/casbin"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type UserService interface {
	CreateUser(userInfo *bean.UserInfo) ([]*bean.UserInfo, error)
	UpdateUser(userInfo *bean.UserInfo) (*bean.UserInfo, error)
	GetById(id int32) (*bean.UserInfo, error)
	GetAll() ([]bean.UserInfo, error)

	GetLoggedInUser(r *http.Request) (int32, error)
	GetByIds(ids []int32) ([]bean.UserInfo, error)
	DeleteUser(userInfo *bean.UserInfo) (bool, error)
	CheckUserRoles(id int32) ([]string, error)
	SyncOrchestratorToCasbin() (bool, error)
	GetUserByToken(token string) (int32, error)
	IsSuperAdmin(userId int) (bool, error)
}

type UserServiceImpl struct {
	sessionManager      *session.SessionManager
	userAuthRepository  repository.UserAuthRepository
	sessionClient       session2.ServiceClient
	logger              *zap.SugaredLogger
	userRepository      repository.UserRepository
	roleGroupRepository repository.RoleGroupRepository
}

func NewUserServiceImpl(userAuthRepository repository.UserAuthRepository, sessionManager *session.SessionManager,
	client session2.ServiceClient, logger *zap.SugaredLogger, userRepository repository.UserRepository,
	userGroupRepository repository.RoleGroupRepository) *UserServiceImpl {
	serviceImpl := &UserServiceImpl{
		userAuthRepository:  userAuthRepository,
		sessionManager:      sessionManager,
		sessionClient:       client,
		logger:              logger,
		userRepository:      userRepository,
		roleGroupRepository: userGroupRepository,
	}
	cStore = sessions.NewCookieStore(randKey())
	return serviceImpl
}

func (impl UserServiceImpl) validateUserRequest(userInfo *bean.UserInfo) (bool, error) {
	if len(userInfo.RoleFilters) == 1 &&
		len(userInfo.RoleFilters[0].Team) == 0 && len(userInfo.RoleFilters[0].Environment) == 0 && len(userInfo.RoleFilters[0].Action) == 0 {
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

func (impl UserServiceImpl) CreateUser(userInfo *bean.UserInfo) ([]*bean.UserInfo, error) {
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
			userInfo, err = impl.updateUserIfExists(userInfo, dbUser, emailId)
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

func (impl UserServiceImpl) updateUserIfExists(userInfo *bean.UserInfo, dbUser *repository.UserModel, emailId string) (*bean.UserInfo, error) {
	updateUserInfo, err := impl.GetById(dbUser.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	if dbUser.Active == false {
		updateUserInfo = &bean.UserInfo{Id: dbUser.Id}
	}
	updateUserInfo.RoleFilters = impl.mergeRoleFilter(updateUserInfo.RoleFilters, userInfo.RoleFilters)
	updateUserInfo.Groups = impl.mergeGroups(updateUserInfo.Groups, userInfo.Groups)
	updateUserInfo.UserId = userInfo.UserId
	updateUserInfo.EmailId = emailId // override case sensitivity
	updateUserInfo, err = impl.UpdateUser(updateUserInfo)
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

			if len(roleFilter.EntityName) == 0 {
				roleFilter.EntityName = "NONE"
			}
			if len(roleFilter.Environment) == 0 {
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
					roleModel, err := impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
					if err != nil {
						impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
						return nil, err
					}
					if roleModel.Id == 0 {
						impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
						//userInfo.Status = "role not fount for any given filter: " + roleFilter.Team + "," + roleFilter.Environment + "," + roleFilter.Application + "," + roleFilter.Action

						//TODO - create roles from here
						if len(roleFilter.Team) > 0 {
							flag, err := impl.userAuthRepository.CreateDefaultPolicies(roleFilter.Team, entityName, environment, tx)
							if err != nil || flag == false {
								return nil, err
							}
							roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
							if err != nil {
								impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
								return nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
								userInfo.Status = "role not found for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
								continue
							}
						} else if len(roleFilter.Entity) > 0 {
							flag, err := impl.userAuthRepository.CreateDefaultPoliciesForGlobalEntity(roleFilter.Entity, entityName, roleFilter.Action, tx)
							if err != nil || flag == false {
								return nil, err
							}
							roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
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
						userRoleModel := &repository.UserRoleModel{UserId: model.Id, RoleId: roleModel.Id}
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

		isSuperAdmin := false
		roles, err := impl.CheckUserRoles(userInfo.UserId)
		if err != nil {
			return nil, err
		}
		for _, item := range roles {
			if item == bean.SUPERADMIN {
				isSuperAdmin = true
			}
		}
		if isSuperAdmin == false {
			err = &util.ApiError{HttpStatusCode: http.StatusForbidden, UserMessage: "Invalid request, not allow to update super admin type user"}
			return nil, err
		}

		flag, err := impl.userAuthRepository.CreateUpdateDefaultPoliciesForSuperAdmin(tx)
		if err != nil || flag == false {
			return nil, err
		}
		roleModel, err := impl.userAuthRepository.GetRoleByFilter("", "", "", "", "super-admin")
		if err != nil {
			impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
			return nil, err
		}
		if roleModel.Id > 0 {
			userRoleModel := &repository.UserRoleModel{UserId: model.Id, RoleId: roleModel.Id}
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
		})
		key := fmt.Sprintf("%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment, role.EntityName, role.Action)
		keysMap[key] = true
	}
	for _, role := range newR {
		key := fmt.Sprintf("%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment, role.EntityName, role.Action)
		if _, ok := keysMap[key]; !ok {
			roleFilters = append(roleFilters, bean.RoleFilter{
				Entity:      role.Entity,
				Team:        role.Team,
				Environment: role.Environment,
				EntityName:  role.EntityName,
				Action:      role.Action,
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

func (impl UserServiceImpl) UpdateUser(userInfo *bean.UserInfo) (*bean.UserInfo, error) {
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

	var policies []casbin2.Policy
	var policiesRemove []casbin2.Policy
	if userInfo.SuperAdmin == false {

		//Starts Role and Mapping
		userRoleModels, err := impl.userAuthRepository.GetUserRoleMappingByUserId(model.Id)
		if err != nil {
			return nil, err
		}
		roleIds := make(map[int]repository.UserRoleModel)
		roleIdsRemaining := make(map[int]*repository.UserRoleModel)
		oldRolesItems := make(map[int]repository.UserRoleModel)

		for i := range userRoleModels {
			roleIds[userRoleModels[i].RoleId] = *userRoleModels[i]
			roleIdsRemaining[userRoleModels[i].RoleId] = userRoleModels[i]
			oldRolesItems[userRoleModels[i].RoleId] = *userRoleModels[i]
		}

		//validate role filters
		_, err = impl.validateUserRequest(userInfo)
		if err != nil {
			err = &util.ApiError{HttpStatusCode: http.StatusBadRequest, UserMessage: "Invalid request, please provide role filters"}
			return nil, err
		}

		// DELETE Removed Items
		for _, roleFilter := range userInfo.RoleFilters {
			if len(roleFilter.EntityName) == 0 {
				roleFilter.EntityName = "NONE"
			}
			if len(roleFilter.Environment) == 0 {
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
					roleModel, err := impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
					if err != nil {
						impl.logger.Errorw("Error in fetching roles by filter", "user", userInfo)
						return nil, err
					}
					if roleModel.Id == 0 {
						impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
						userInfo.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
						continue
					}
					if _, ok := roleIds[roleModel.Id]; ok {
						delete(roleIdsRemaining, roleModel.Id)
					}
				}
			}
		}

		//delete remaining Ids from casbin role mapping table in orchestrator and casbin policy db
		// which are existing but not provided in this request

		for _, userRoleModel := range roleIdsRemaining {
			_, err := impl.userAuthRepository.DeleteUserRoleMapping(userRoleModel, tx)
			if err != nil {
				impl.logger.Errorw("Error in delete user role mapping", "user", userInfo)
				return nil, err
			}
			role, err := impl.userAuthRepository.GetRoleById(userRoleModel.RoleId)
			if err != nil {
				return nil, err
			}
			policiesRemove = append(policiesRemove, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(role.Role)})
		}
		// DELETE ENDS

		//Adding New Policies
		for _, roleFilter := range userInfo.RoleFilters {
			if len(roleFilter.EntityName) == 0 {
				roleFilter.EntityName = "NONE"
			}
			if len(roleFilter.Environment) == 0 {
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
					roleModel, err := impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
					if err != nil {
						impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
						return nil, err
					}
					if roleModel.Id == 0 {
						impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
						userInfo.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action

						//TODO - create roles from here
						if len(roleFilter.Team) > 0 {
							flag, err := impl.userAuthRepository.CreateDefaultPolicies(roleFilter.Team, entityName, environment, tx)
							if err != nil || flag == false {
								return nil, err
							}
							roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
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
							roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
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

					if _, ok := roleIds[roleModel.Id]; ok {
						//Adding policies which is removed
						policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
					} else {
						//new role ids in new array, add it
						if roleModel.Id > 0 {
							userRoleModel := &repository.UserRoleModel{UserId: model.Id, RoleId: roleModel.Id}
							userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
							if err != nil {
								return nil, err
							}
							policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
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
				policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(userInfo.EmailId), Obj: casbin2.Object(userGroup.CasbinName)})
			}
		}
		for _, item := range userCasbinRoles {
			if _, ok := newGroupMap[item]; !ok {
				if item != bean.SUPERADMIN {
					policiesRemove = append(policiesRemove, casbin2.Policy{Type: "g", Sub: casbin2.Subject(userInfo.EmailId), Obj: casbin2.Object(item)})
				}
			}
		}
		// END GROUP POLICY

	} else if userInfo.SuperAdmin == true {

		flag, err := impl.userAuthRepository.CreateUpdateDefaultPoliciesForSuperAdmin(tx)
		if err != nil || flag == false {
			return nil, err
		}
		roleModel, err := impl.userAuthRepository.GetRoleByFilter("", "", "", "", "super-admin")
		if err != nil {
			impl.logger.Errorw("Error in fetching role by filter", "user", userInfo)
			return nil, err
		}
		if roleModel.Id > 0 {
			userRoleModel := &repository.UserRoleModel{UserId: model.Id, RoleId: roleModel.Id}
			userRoleModel, err = impl.userAuthRepository.CreateUserRoleMapping(userRoleModel, tx)
			if err != nil {
				return nil, err
			}
			policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.EmailId), Obj: casbin2.Object(roleModel.Role)})
		}
	}

	//updating in casbin
	if len(policiesRemove) > 0 {
		pRes := casbin2.RemovePolicy(policiesRemove)
		println(pRes)
	}
	if len(policies) > 0 {
		pRes := casbin2.AddPolicy(policies)
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
			key = fmt.Sprintf("%s_%s", role.Team, role.Action)
		} else if len(role.Entity) > 0 {
			key = fmt.Sprintf("%s_%s", role.Entity, role.Action)
		}
		if _, ok := roleFilterMap[key]; ok {
			envArr := strings.Split(roleFilterMap[key].Environment, ",")
			if !containsArr(envArr, role.Environment) {
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

	filterGroupsModels, err := impl.roleGroupRepository.GetRoleGroupListByCasbinNames(filterGroups)
	if err != nil {
		impl.logger.Warnw("No Roles Found for user", "id", model.Id)
	}
	filterGroups = nil
	for _, item := range filterGroupsModels {
		filterGroups = append(filterGroups, item.Name)
	}
	if filterGroups == nil || len(filterGroups) == 0 {
		filterGroups = make([]string, 0)
	}
	if roleFilters == nil || len(roleFilters) == 0 {
		roleFilters = make([]bean.RoleFilter, 0)
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

func containsArr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (impl UserServiceImpl) GetAll() ([]bean.UserInfo, error) {
	model, err := impl.userRepository.GetAll()
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
	if response == nil || len(response) == 0 {
		response = make([]bean.UserInfo, 0)
	}
	return response, nil
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
		AccessToken: model.AccessToken,
		RoleFilters: roleFilters,
	}

	return response, nil
}
func (impl UserServiceImpl) GetLoggedInUser(r *http.Request) (int32, error) {
	token := r.Header.Get("token")
	return impl.GetUserByToken(token)
}

func (impl UserServiceImpl) GetUserByToken(token string) (int32, error) {
	if len(token) == 0 {
		impl.logger.Infow("no token provided", "token", token)
		err := &util.ApiError{
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "no token provided",
		}
		return http.StatusUnauthorized, err
	}

	claims, err := impl.sessionManager.VerifyToken(token)
	if err != nil {
		impl.logger.Errorw("failed to verify token", "error", err)
		err := &util.ApiError{
			Code:            constants.UserNoTokenProvided,
			InternalMessage: "failed to verify token",
			UserMessage:     fmt.Sprintf("token verification failed while getting logged in user: %s", token),
		}
		return http.StatusUnauthorized, err
	}
	mapClaims, err := jwt2.MapClaims(claims)
	if err != nil {
		impl.logger.Errorw("failed to MapClaims", "error", err)
		return http.StatusUnauthorized, err
	}

	email := jwt2.GetField(mapClaims, "email")
	sub := jwt2.GetField(mapClaims, "sub")

	if len(email) == 0 && sub == "admin" {
		email = sub
	}

	userInfo, err := impl.GetUserByEmail(email)
	if err != nil {
		impl.logger.Errorw("unable to fetch user from db", "error", err)
		err := &util.ApiError{
			Code:            constants.UserNotFoundForToken,
			InternalMessage: "user not found for token",
			UserMessage:     fmt.Sprintf("no user found against provided token: %s", token),
		}
		return http.StatusUnauthorized, err
	}
	return userInfo.Id, nil
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
