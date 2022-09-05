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
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	casbin2 "github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

type RoleGroupService interface {
	CreateRoleGroup(request *bean.RoleGroup) (*bean.RoleGroup, error)
	UpdateRoleGroup(request *bean.RoleGroup, token string, managerAuth func(token string, object string) bool) (*bean.RoleGroup, error)
	FetchDetailedRoleGroups() ([]*bean.RoleGroup, error)
	FetchRoleGroupsById(id int32) (*bean.RoleGroup, error)
	FetchRoleGroups() ([]*bean.RoleGroup, error)
	FetchRoleGroupsByName(name string) ([]*bean.RoleGroup, error)
	DeleteRoleGroup(model *bean.RoleGroup) (bool, error)
	FetchRolesForGroups(groupNames []string) ([]*bean.RoleFilter, error)
}

type RoleGroupServiceImpl struct {
	userAuthRepository  repository2.UserAuthRepository
	logger              *zap.SugaredLogger
	userRepository      repository2.UserRepository
	roleGroupRepository repository2.RoleGroupRepository
	userCommonService   UserCommonService
}

func NewRoleGroupServiceImpl(userAuthRepository repository2.UserAuthRepository,
	logger *zap.SugaredLogger, userRepository repository2.UserRepository,
	roleGroupRepository repository2.RoleGroupRepository, userCommonService UserCommonService) *RoleGroupServiceImpl {
	serviceImpl := &RoleGroupServiceImpl{
		userAuthRepository:  userAuthRepository,
		logger:              logger,
		userRepository:      userRepository,
		roleGroupRepository: roleGroupRepository,
		userCommonService:   userCommonService,
	}
	cStore = sessions.NewCookieStore(randKey())
	return serviceImpl
}

func (impl RoleGroupServiceImpl) CreateRoleGroup(request *bean.RoleGroup) (*bean.RoleGroup, error) {
	dbConnection := impl.roleGroupRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	if request.Id > 0 {
		_, err := impl.roleGroupRepository.GetRoleGroupById(request.Id)
		if err != nil {
			impl.logger.Errorw("error while fetching user from db", "error", err)
			return nil, err
		}
	} else {

		//create new user in our db on d basis of info got from google api or hex. assign a basic role
		model := &repository2.RoleGroup{
			Name:        request.Name,
			Description: request.Description,
		}
		rgName := strings.ToLower(request.Name)
		object := "group:" + strings.ReplaceAll(rgName, " ", "_")
		model.CasbinName = object
		model.CreatedBy = request.UserId
		model.UpdatedBy = request.UserId
		model.CreatedOn = time.Now()
		model.UpdatedOn = time.Now()
		model.Active = true
		model, err := impl.roleGroupRepository.CreateRoleGroup(model, tx)
		request.Id = model.Id
		if err != nil {
			impl.logger.Errorw("error in creating new user group", "error", err)
			err = &util.ApiError{
				Code:            constants.UserCreateDBFailed,
				InternalMessage: "failed to create new user in db",
				UserMessage:     fmt.Sprintf("requested by"),
			}
			return request, err
		}
		model.Id = model.Id

		//Starts Role and Mapping
		var policies []casbin2.Policy
		for _, roleFilter := range request.RoleFilters {
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
						impl.logger.Errorw("Error in fetching role by filter", "user", request)
						return nil, err
					}
					if roleModel.Id == 0 {
						impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
						//userInfo.Status = "role not fount for any given filter: " + roleFilter.Team + "," + roleFilter.Environment + "," + roleFilter.Application + "," + roleFilter.Action

						if len(roleFilter.Team) > 0 && len(roleFilter.Environment) > 0 {
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
								impl.logger.Errorw("Error in fetching role by filter", "user", request)
								return nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
								request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
								continue
							}
						} else if len(roleFilter.Entity) > 0 {
							flag, err := impl.userAuthRepository.CreateDefaultPoliciesForGlobalEntity(roleFilter.Entity, entityName, roleFilter.Action, tx)
							if err != nil || flag == false {
								return nil, err
							}
							roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action, roleFilter.AccessType)
							if err != nil {
								impl.logger.Errorw("Error in fetching role by filter", "user", request)
								return nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
								request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
								continue
							}
						} else {
							continue
						}
					}

					if roleModel.Id > 0 {
						roleGroupMappingModel := &repository2.RoleGroupRoleMapping{RoleGroupId: model.Id, RoleId: roleModel.Id}
						roleGroupMappingModel.CreatedBy = request.UserId
						roleGroupMappingModel.UpdatedBy = request.UserId
						roleGroupMappingModel.CreatedOn = time.Now()
						roleGroupMappingModel.UpdatedOn = time.Now()
						roleGroupMappingModel, err = impl.roleGroupRepository.CreateRoleGroupRoleMapping(roleGroupMappingModel, tx)
						if err != nil {
							return nil, err
						}
						policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.CasbinName), Obj: casbin2.Object(roleModel.Role)})
					}
				}
			}
		}

		if len(policies) > 0 {
			pRes := casbin2.AddPolicy(policies)
			println(pRes)
		}
		//Ends
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (impl RoleGroupServiceImpl) UpdateRoleGroup(request *bean.RoleGroup, token string, managerAuth func(token string, object string) bool) (*bean.RoleGroup, error) {
	dbConnection := impl.roleGroupRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	roleGroup, err := impl.roleGroupRepository.GetRoleGroupById(request.Id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	//policyGroup.Name = request.Name
	roleGroup.Description = request.Description
	roleGroup.UpdatedOn = time.Now()
	roleGroup.UpdatedBy = request.UserId
	roleGroup.Active = true
	roleGroup, err = impl.roleGroupRepository.UpdateRoleGroup(roleGroup, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	roleGroupMappingModels, err := impl.roleGroupRepository.GetRoleGroupRoleMappingByRoleGroupId(roleGroup.Id)
	if err != nil {
		return nil, err
	}

	existingRoles := make(map[int]*repository2.RoleGroupRoleMapping)
	eliminatedRoles := make(map[int]*repository2.RoleGroupRoleMapping)
	for _, item := range roleGroupMappingModels {
		existingRoles[item.RoleId] = item
		eliminatedRoles[item.RoleId] = item
	}

	// DELETE PROCESS STARTS
	var eliminatedPolicies []casbin2.Policy
	items, err := impl.userCommonService.RemoveRolesAndReturnEliminatedPoliciesForGroups(request, existingRoles, eliminatedRoles, tx, token, managerAuth)
	if err != nil {
		return nil, err
	}
	eliminatedPolicies = append(eliminatedPolicies, items...)
	if len(eliminatedPolicies) > 0 {
		pRes := casbin2.RemovePolicy(eliminatedPolicies)
		println(pRes)
	}
	// DELETE PROCESS ENDS

	//Adding New Policies
	var policies []casbin2.Policy
	for _, roleFilter := range request.RoleFilters {
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
					impl.logger.Errorw("Error in fetching role by filter", "user", request)
					return nil, err
				}
				if roleModel.Id == 0 {
					impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
					request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action

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
							impl.logger.Errorw("Error in fetching role by filter", "user", request)
							return nil, err
						}
						if roleModel.Id == 0 {
							impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
							request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
							continue
						}
					} else if len(roleFilter.Entity) > 0 {
						flag, err := impl.userAuthRepository.CreateDefaultPoliciesForGlobalEntity(roleFilter.Entity, entityName, roleFilter.Action, tx)
						if err != nil || flag == false {
							return nil, err
						}
						roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action, roleFilter.AccessType)
						if err != nil {
							impl.logger.Errorw("Error in fetching role by filter", "user", request)
							return nil, err
						}
						if roleModel.Id == 0 {
							impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
							request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
							continue
						}
					} else {
						continue
					}
				}

				if _, ok := existingRoles[roleModel.Id]; ok {
					//Adding policies which is removed
					policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(roleGroup.CasbinName), Obj: casbin2.Object(roleModel.Role)})
				} else {
					if roleModel.Id > 0 {
						//new role ids in new array, add it
						roleGroupMappingModel := &repository2.RoleGroupRoleMapping{RoleGroupId: request.Id, RoleId: roleModel.Id}
						roleGroupMappingModel.CreatedBy = request.UserId
						roleGroupMappingModel.UpdatedBy = request.UserId
						roleGroupMappingModel.CreatedOn = time.Now()
						roleGroupMappingModel.UpdatedOn = time.Now()
						roleGroupMappingModel, err = impl.roleGroupRepository.CreateRoleGroupRoleMapping(roleGroupMappingModel, tx)
						if err != nil {
							return nil, err
						}
						policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(roleGroup.CasbinName), Obj: casbin2.Object(roleModel.Role)})
					}
				}
			}
		}
	}

	//updating in casbin
	if len(policies) > 0 {
		casbin2.AddPolicy(policies)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return request, nil
}

const AllEnvironment string = ""

func (impl RoleGroupServiceImpl) FetchRoleGroupsById(id int32) (*bean.RoleGroup, error) {
	roleGroup, err := impl.roleGroupRepository.GetRoleGroupById(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	roleFilters := impl.getRoleGroupMetadata(roleGroup)
	bean := &bean.RoleGroup{
		Id:          roleGroup.Id,
		Name:        roleGroup.Name,
		Description: roleGroup.Description,
		RoleFilters: roleFilters,
	}
	return bean, nil
}

func (impl RoleGroupServiceImpl) getRoleGroupMetadata(roleGroup *repository2.RoleGroup) []bean.RoleFilter {
	roles, err := impl.userAuthRepository.GetRolesByGroupId(roleGroup.Id)
	if err != nil {
		impl.logger.Errorw("No Roles Found for user", "roleGroupId", roleGroup.Id)
	}
	var roleFilters []bean.RoleFilter
	roleFilterMap := make(map[string]*bean.RoleFilter)
	for _, role := range roles {
		key := ""
		if len(role.Team) > 0 {
			key = fmt.Sprintf("%s_%s_%s", role.Team, role.Action, role.AccessType)
		} else if len(role.Entity) > 0 {
			key = fmt.Sprintf("%s_%s", role.Entity, role.Action)
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
	}
	for _, v := range roleFilterMap {
		roleFilters = append(roleFilters, *v)
	}
	if len(roleFilters) == 0 {
		roleFilters = make([]bean.RoleFilter, 0)
	}
	return roleFilters
}

func (impl RoleGroupServiceImpl) FetchDetailedRoleGroups() ([]*bean.RoleGroup, error) {
	roleGroups, err := impl.roleGroupRepository.GetAllRoleGroup()
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	var list []*bean.RoleGroup
	for _, roleGroup := range roleGroups {
		roleFilters := impl.getRoleGroupMetadata(roleGroup)
		roleGrp := &bean.RoleGroup{
			Id:          roleGroup.Id,
			Name:        roleGroup.Name,
			Description: roleGroup.Description,
			RoleFilters: roleFilters,
		}
		list = append(list, roleGrp)
	}

	if len(list) == 0 {
		list = make([]*bean.RoleGroup, 0)
	}
	return list, nil
}

func (impl RoleGroupServiceImpl) FetchRoleGroups() ([]*bean.RoleGroup, error) {
	roleGroup, err := impl.roleGroupRepository.GetAllRoleGroup()
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	var list []*bean.RoleGroup
	for _, item := range roleGroup {
		bean := &bean.RoleGroup{
			Id:          item.Id,
			Name:        item.Name,
			Description: item.Description,
			RoleFilters: make([]bean.RoleFilter, 0),
		}
		list = append(list, bean)
	}

	if len(list) == 0 {
		list = make([]*bean.RoleGroup, 0)
	}
	return list, nil
}

func (impl RoleGroupServiceImpl) FetchRoleGroupsByName(name string) ([]*bean.RoleGroup, error) {
	roleGroup, err := impl.roleGroupRepository.GetRoleGroupListByName(name)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	var list []*bean.RoleGroup
	for _, item := range roleGroup {
		bean := &bean.RoleGroup{
			Id:          item.Id,
			Name:        item.Name,
			Description: item.Description,
			RoleFilters: make([]bean.RoleFilter, 0),
		}
		list = append(list, bean)

	}
	return list, nil
}

func (impl RoleGroupServiceImpl) DeleteRoleGroup(bean *bean.RoleGroup) (bool, error) {

	dbConnection := impl.roleGroupRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return false, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	model, err := impl.roleGroupRepository.GetRoleGroupById(bean.Id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return false, err
	}
	model.Active = false
	model.UpdatedOn = time.Now()
	model.UpdatedBy = bean.UserId
	_, err = impl.roleGroupRepository.UpdateRoleGroup(model, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (impl RoleGroupServiceImpl) FetchRolesForGroups(groupNames []string) ([]*bean.RoleFilter, error) {
	if len(groupNames) == 0 {
		return nil, nil
	}
	roleGroups, err := impl.roleGroupRepository.GetRoleGroupListByNames(groupNames)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		impl.logger.Warnw("no result found for role groups", "groups", groupNames)
		return nil, nil
	}

	var roleGroupIds []int32
	for _, roleGroup := range roleGroups {
		roleGroupIds = append(roleGroupIds, roleGroup.Id)
	}

	roles, err := impl.roleGroupRepository.GetRoleGroupRoleMappingByRoleGroupIds(roleGroupIds)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	list := make([]*bean.RoleFilter, 0)
	if roles == nil {
		return list, nil
	}
	for _, role := range roles {
		bean := &bean.RoleFilter{
			EntityName:  role.EntityName,
			Entity:      role.Entity,
			Action:      role.Action,
			Environment: role.Environment,
			Team:        role.Team,
			AccessType:  role.AccessType,
		}
		list = append(list, bean)
	}
	return list, nil
}
