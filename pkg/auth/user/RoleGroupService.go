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
	"errors"
	"fmt"
	"strings"
	"time"

	casbin2 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	util2 "github.com/devtron-labs/devtron/pkg/auth/user/util"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

type RoleGroupService interface {
	CreateRoleGroup(request *bean.RoleGroup) (*bean.RoleGroup, error)
	UpdateRoleGroup(request *bean.RoleGroup, token string, managerAuth func(resource, token string, object string) bool) (*bean.RoleGroup, error)
	FetchDetailedRoleGroups() ([]*bean.RoleGroup, error)
	FetchRoleGroupsById(id int32) (*bean.RoleGroup, error)
	FetchRoleGroups() (*bean.RoleGroupListingResponse, error)
	FetchRoleGroupsWithFilters(req *bean.FetchListingRequest) (*bean.RoleGroupListingResponse, error)
	FetchRoleGroupsByName(name string) ([]*bean.RoleGroup, error)
	DeleteRoleGroup(model *bean.RoleGroup) (bool, error)
	FetchRoleGroupsWithRolesByGroupNames(groupNames []string) ([]*bean.RoleFilter, []bean.RoleGroup, error)
	FetchRoleGroupsWithRolesByGroupCasbinNames(groupCasbinNames []string) ([]*bean.RoleFilter, []bean.RoleGroup, error)
}

type RoleGroupServiceImpl struct {
	userAuthRepository  repository.UserAuthRepository
	logger              *zap.SugaredLogger
	userRepository      repository.UserRepository
	roleGroupRepository repository.RoleGroupRepository
	userCommonService   UserCommonService
}

func NewRoleGroupServiceImpl(userAuthRepository repository.UserAuthRepository,
	logger *zap.SugaredLogger, userRepository repository.UserRepository,
	roleGroupRepository repository.RoleGroupRepository, userCommonService UserCommonService) *RoleGroupServiceImpl {
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
	validationPassed := util2.CheckValidationForRoleGroupCreation(request.Name)
	if !validationPassed {
		return nil, errors.New(bean2.VALIDATION_FAILED_ERROR_MSG)
	}

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
		//loading policy for safety
		casbin2.LoadPolicy()
		//create new user in our db on d basis of info got from google api or hex. assign a basic role
		model := &repository.RoleGroup{
			Name:        request.Name,
			Description: request.Description,
		}
		rgName := strings.ToLower(request.Name)
		object := "group:" + strings.ReplaceAll(rgName, " ", "_")

		exists, err := impl.roleGroupRepository.CheckRoleGroupExistByCasbinName(object)
		if err != nil {
			impl.logger.Errorw("error in getting role group by casbin name", "err", err, "casbinName", object)
			return nil, err
		} else if exists {
			impl.logger.Errorw("role group already present", "err", err, "roleGroup", request.Name)
			return nil, errors.New("role group already exist")
		}
		model.CasbinName = object
		model.CreatedBy = request.UserId
		model.UpdatedBy = request.UserId
		model.CreatedOn = time.Now()
		model.UpdatedOn = time.Now()
		model.Active = true
		model, err = impl.roleGroupRepository.CreateRoleGroup(model, tx)
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

		capacity, mapping := impl.userCommonService.GetCapacityForRoleFilter(request.RoleFilters)
		var policies = make([]casbin2.Policy, 0, capacity)
		if request.SuperAdmin == false {
			for index, roleFilter := range request.RoleFilters {
				entity := roleFilter.Entity
				if entity == bean.CLUSTER_ENTITIY {
					policiesToBeAdded, err := impl.CreateOrUpdateRoleGroupForClusterEntity(roleFilter, request.UserId, model, nil, "", nil, tx, mapping[index])
					policies = append(policies, policiesToBeAdded...)
					if err != nil {
						// making it non-blocking as it is being done for multiple Role filters and does not want this to be blocking.
						impl.logger.Errorw("error in creating updating role group for cluster entity", "err", err, "roleFilter", roleFilter)
					}
				} else if entity == bean2.EntityJobs {
					policiesToBeAdded, err := impl.CreateOrUpdateRoleGroupForJobsEntity(roleFilter, request.UserId, model, nil, "", nil, tx, mapping[index])
					policies = append(policies, policiesToBeAdded...)
					if err != nil {
						// making it non-blocking as it is being done for multiple Role filters and does not want this to be blocking.
						impl.logger.Errorw("error in creating updating role group for jobs entity", "err", err, "roleFilter", roleFilter)
					}
				} else {
					policiesToBeAdded, err := impl.CreateOrUpdateRoleGroupForOtherEntity(roleFilter, request, model, nil, "", nil, tx, mapping[index])
					policies = append(policies, policiesToBeAdded...)
					if err != nil {
						// making it non-blocking as it is being done for multiple Role filters and does not want this to be blocking.
						impl.logger.Errorw("error in creating updating role group for apps entity", "err", err, "roleFilter", roleFilter)
					}
				}
			}
		} else if request.SuperAdmin == true {
			flag, err := impl.userAuthRepository.CreateRoleForSuperAdminIfNotExists(tx, request.UserId)
			if err != nil || flag == false {
				impl.logger.Errorw("error in CreateRoleForSuperAdminIfNotExists ", "err", err, "roleGroupName", request.Name)
				return nil, err
			}
			roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes("", "", "", "", bean2.SUPER_ADMIN, false, "", "", "", "", "", "", "", false, "")
			if err != nil {
				impl.logger.Errorw("error in getting role by filter for all Types for superAdmin", "err", err)
				return nil, err
			}
			if roleModel.Id > 0 {
				roleGroupMappingModel := &repository.RoleGroupRoleMapping{RoleGroupId: model.Id, RoleId: roleModel.Id}
				roleGroupMappingModel.CreatedBy = request.UserId
				roleGroupMappingModel.UpdatedBy = request.UserId
				roleGroupMappingModel.CreatedOn = time.Now()
				roleGroupMappingModel.UpdatedOn = time.Now()
				roleGroupMappingModel, err = impl.roleGroupRepository.CreateRoleGroupRoleMapping(roleGroupMappingModel, tx)
				if err != nil {
					impl.logger.Errorw("error in creating role group role mapping", "err", err, "RoleGroupId", model.Id)
					return nil, err
				}
				policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.CasbinName), Obj: casbin2.Object(roleModel.Role)})
			}
		}

		if len(policies) > 0 {
			err = casbin2.AddPolicy(policies)
			if err != nil {
				impl.logger.Errorw("casbin policy addition failed", "err", err)
				return nil, err
			}
			//loading policy for syncing orchestrator to casbin with newly added policies
			casbin2.LoadPolicy()
		}
		//Ends
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (impl RoleGroupServiceImpl) CreateOrUpdateRoleGroupForClusterEntity(roleFilter bean.RoleFilter, userId int32, model *repository.RoleGroup, existingRoles map[int]*repository.RoleGroupRoleMapping, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, capacity int) ([]casbin2.Policy, error) {
	//var policiesToBeAdded []casbin2.Policy
	namespaces := strings.Split(roleFilter.Namespace, ",")
	groups := strings.Split(roleFilter.Group, ",")
	kinds := strings.Split(roleFilter.Kind, ",")
	resources := strings.Split(roleFilter.Resource, ",")
	actionType := roleFilter.Action
	accessType := roleFilter.AccessType
	entity := roleFilter.Entity
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
					roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, "", "", "", "", false, accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, false, "")
					if err != nil {
						impl.logger.Errorw("error in getting new role model by filter")
						return policiesToBeAdded, err
					}
					if roleModel.Id == 0 {
						flag, err, policiesAdded := impl.userCommonService.CreateDefaultPoliciesForAllTypes("", "", "", entity, roleFilter.Cluster, namespace, group, kind, resource, actionType, accessType, roleFilter.Approver, "", userId)
						if err != nil || flag == false {
							return policiesToBeAdded, err
						}
						policiesToBeAdded = append(policiesToBeAdded, policiesAdded...)
						roleModel, err = impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, "", "", "", "", false, accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, false, "")
						if err != nil {
							return policiesToBeAdded, err
						}
						if roleModel.Id == 0 {
							continue
						}
					}
					if _, ok := existingRoles[roleModel.Id]; ok {
						//Adding policies which are removed
						policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.CasbinName), Obj: casbin2.Object(roleModel.Role)})
					} else {
						if roleModel.Id > 0 {
							//new role ids in new array, add it
							roleGroupMappingModel := &repository.RoleGroupRoleMapping{RoleGroupId: model.Id, RoleId: roleModel.Id}
							roleGroupMappingModel.CreatedBy = userId
							roleGroupMappingModel.UpdatedBy = userId
							roleGroupMappingModel.CreatedOn = time.Now()
							roleGroupMappingModel.UpdatedOn = time.Now()
							roleGroupMappingModel, err = impl.roleGroupRepository.CreateRoleGroupRoleMapping(roleGroupMappingModel, tx)
							if err != nil {
								return nil, err
							}
							policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.CasbinName), Obj: casbin2.Object(roleModel.Role)})
						}
					}
				}
			}
		}
	}
	return policiesToBeAdded, nil
}

func (impl RoleGroupServiceImpl) CreateOrUpdateRoleGroupForOtherEntity(roleFilter bean.RoleFilter, request *bean.RoleGroup, model *repository.RoleGroup, existingRoles map[int]*repository.RoleGroupRoleMapping, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, capacity int) ([]casbin2.Policy, error) {
	var policiesToBeAdded = make([]casbin2.Policy, 0, capacity)
	accessType := roleFilter.AccessType
	entityNames := strings.Split(roleFilter.EntityName, ",")
	environments := strings.Split(roleFilter.Environment, ",")
	entity := roleFilter.Entity
	actions := strings.Split(roleFilter.Action, ",")
	for _, environment := range environments {
		for _, entityName := range entityNames {
			for _, actionType := range actions {
				roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, roleFilter.Approver, accessType, "", "", "", "", "", "", false, "")
				if err != nil {
					return nil, err
				}
				if roleModel.Id == 0 {
					request.Status = fmt.Sprintf("%s+%s,%s,%s,%s", bean2.RoleNotFoundStatusPrefix, roleFilter.Team, environment, entityName, actionType)
					if roleFilter.Entity == bean2.ENTITY_APPS || roleFilter.Entity == bean.CHART_GROUP_ENTITY {
						flag, err, policiesAdded := impl.userCommonService.CreateDefaultPoliciesForAllTypes(roleFilter.Team, entityName, environment, entity, "", "", "", "", "", actionType, accessType, roleFilter.Approver, "", request.UserId)
						policiesToBeAdded = append(policiesToBeAdded, policiesAdded...)
						if err != nil || flag == false {
							return nil, err
						}
						roleModel, err = impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, roleFilter.Approver, accessType, "", "", "", "", "", "", false, "")
						if err != nil {
							return nil, err
						}
						if roleModel.Id == 0 {
							request.Status = fmt.Sprintf("%s+%s,%s,%s,%s", bean2.RoleNotFoundStatusPrefix, roleFilter.Team, environment, entityName, actionType)
							continue
						}
					} else {
						continue
					}
				}
				if _, ok := existingRoles[roleModel.Id]; ok {
					//Adding policies which is removed
					policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.CasbinName), Obj: casbin2.Object(roleModel.Role)})
				} else {
					if roleModel.Id > 0 {
						//new role ids in new array, add it
						roleGroupMappingModel := &repository.RoleGroupRoleMapping{RoleGroupId: model.Id, RoleId: roleModel.Id}
						roleGroupMappingModel.CreatedBy = request.UserId
						roleGroupMappingModel.UpdatedBy = request.UserId
						roleGroupMappingModel.CreatedOn = time.Now()
						roleGroupMappingModel.UpdatedOn = time.Now()
						roleGroupMappingModel, err = impl.roleGroupRepository.CreateRoleGroupRoleMapping(roleGroupMappingModel, tx)
						if err != nil {
							return nil, err
						}
						policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.CasbinName), Obj: casbin2.Object(roleModel.Role)})
					}
				}
			}
		}
	}
	return policiesToBeAdded, nil
}

func (impl RoleGroupServiceImpl) CreateOrUpdateRoleGroupForJobsEntity(roleFilter bean.RoleFilter, userId int32, model *repository.RoleGroup, existingRoles map[int]*repository.RoleGroupRoleMapping, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, capacity int) ([]casbin2.Policy, error) {
	actionType := roleFilter.Action
	accessType := roleFilter.AccessType
	entity := roleFilter.Entity
	entityNames := strings.Split(roleFilter.EntityName, ",")
	environments := strings.Split(roleFilter.Environment, ",")
	workflows := strings.Split(roleFilter.Workflow, ",")
	var policiesToBeAdded = make([]casbin2.Policy, 0, capacity)
	for _, environment := range environments {
		for _, entityName := range entityNames {
			for _, workflow := range workflows {
				roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, false, accessType, "", "", "", "", "", "", false, workflow)
				if err != nil {
					impl.logger.Errorw("error in getting new role model")
					return nil, err
				}
				if roleModel.Id == 0 {
					flag, err, policiesAdded := impl.userCommonService.CreateDefaultPoliciesForAllTypes(roleFilter.Team, entityName, environment, entity, "", "", "", "", "", actionType, accessType, false, workflow, userId)
					if err != nil || flag == false {
						return nil, err
					}
					policiesToBeAdded = append(policiesToBeAdded, policiesAdded...)
					roleModel, err = impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, false, accessType, "", "", "", "", "", "", false, workflow)
					if err != nil {
						return nil, err
					}
					if roleModel.Id == 0 {
						continue
					}
				}
				if _, ok := existingRoles[roleModel.Id]; ok {
					//Adding policies which are removed
					policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.CasbinName), Obj: casbin2.Object(roleModel.Role)})
				} else {
					if roleModel.Id > 0 {
						roleGroupMappingModel := &repository.RoleGroupRoleMapping{RoleGroupId: model.Id, RoleId: roleModel.Id}
						roleGroupMappingModel.CreatedBy = userId
						roleGroupMappingModel.UpdatedBy = userId
						roleGroupMappingModel.CreatedOn = time.Now()
						roleGroupMappingModel.UpdatedOn = time.Now()
						roleGroupMappingModel, err = impl.roleGroupRepository.CreateRoleGroupRoleMapping(roleGroupMappingModel, tx)
						if err != nil {
							return nil, err
						}
						policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.CasbinName), Obj: casbin2.Object(roleModel.Role)})
					}
				}
			}
		}
	}
	return policiesToBeAdded, nil
}

func (impl RoleGroupServiceImpl) UpdateRoleGroup(request *bean.RoleGroup, token string, managerAuth func(resource, token string, object string) bool) (*bean.RoleGroup, error) {
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
	//loading policy for safety
	casbin2.LoadPolicy()

	var eliminatedPolicies []casbin2.Policy
	existingRoles := make(map[int]*repository.RoleGroupRoleMapping)
	eliminatedRoles := make(map[int]*repository.RoleGroupRoleMapping)
	capacity, mapping := impl.userCommonService.GetCapacityForRoleFilter(request.RoleFilters)
	var policies = make([]casbin2.Policy, 0, capacity)
	if request.SuperAdmin == false {
		roleGroupMappingModels, err := impl.roleGroupRepository.GetRoleGroupRoleMappingByRoleGroupId(roleGroup.Id)
		if err != nil {
			impl.logger.Errorw("error in getting roleGroupMappingModels by roleGroupId", "err", err, "roleGroupId", roleGroup.Id)
			return nil, err
		}

		for _, item := range roleGroupMappingModels {
			existingRoles[item.RoleId] = item
			eliminatedRoles[item.RoleId] = item
		}

		// DELETE PROCESS STARTS
		items, err := impl.userCommonService.RemoveRolesAndReturnEliminatedPoliciesForGroups(request, existingRoles, eliminatedRoles, tx, token, managerAuth)
		if err != nil {
			return nil, err
		}
		eliminatedPolicies = append(eliminatedPolicies, items...)
		// DELETE PROCESS ENDS

		//Adding New Policies
		for index, roleFilter := range request.RoleFilters {
			if roleFilter.Entity == bean.CLUSTER_ENTITIY {
				policiesToBeAdded, err := impl.CreateOrUpdateRoleGroupForClusterEntity(roleFilter, request.UserId, roleGroup, existingRoles, token, managerAuth, tx, mapping[index])
				policies = append(policies, policiesToBeAdded...)
				if err != nil {
					impl.logger.Errorw("error in creating updating role group for cluster entity", "err", err, "roleFilter", roleFilter)
				}
			} else {
				if len(roleFilter.Team) > 0 {
					// check auth only for apps permission, skip for chart group
					rbacObject := fmt.Sprintf("%s", roleFilter.Team)
					isValidAuth := managerAuth(casbin2.ResourceUser, token, rbacObject)
					if !isValidAuth {
						continue
					}
				}
				switch roleFilter.Entity {
				case bean2.EntityJobs:
					{
						policiesToBeAdded, err := impl.CreateOrUpdateRoleGroupForJobsEntity(roleFilter, request.UserId, roleGroup, existingRoles, token, managerAuth, tx, mapping[index])
						policies = append(policies, policiesToBeAdded...)
						if err != nil {
							impl.logger.Errorw("error in creating updating role group for jobs entity", "err", err, "roleFilter", roleFilter)
						}
					}
				default:
					{
						policiesToBeAdded, err := impl.CreateOrUpdateRoleGroupForOtherEntity(roleFilter, request, roleGroup, existingRoles, token, managerAuth, tx, mapping[index])
						policies = append(policies, policiesToBeAdded...)
						if err != nil {
							impl.logger.Errorw("error in creating updating role group for other entity", "err", err, "roleFilter", roleFilter)
						}
					}
				}
			}
		}
	} else if request.SuperAdmin == true {
		flag, err := impl.userAuthRepository.CreateRoleForSuperAdminIfNotExists(tx, request.UserId)
		if err != nil || flag == false {
			impl.logger.Errorw("error in CreateRoleForSuperAdminIfNotExists ", "err", err, "roleGroupName", request.Name)
			return nil, err
		}
		roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes("", "", "", "", bean2.SUPER_ADMIN, false, "", "", "", "", "", "", "", false, "")
		if err != nil {
			impl.logger.Errorw("error in getting role by filter for all Types for superAdmin", "err", err)
			return nil, err
		}
		if roleModel.Id > 0 {
			roleGroupMappingModel := &repository.RoleGroupRoleMapping{RoleGroupId: roleGroup.Id, RoleId: roleModel.Id}
			roleGroupMappingModel.CreatedBy = request.UserId
			roleGroupMappingModel.UpdatedBy = request.UserId
			roleGroupMappingModel.CreatedOn = time.Now()
			roleGroupMappingModel.UpdatedOn = time.Now()
			roleGroupMappingModel, err = impl.roleGroupRepository.CreateRoleGroupRoleMapping(roleGroupMappingModel, tx)
			if err != nil {
				impl.logger.Errorw("error in creating role group role mapping", "err", err, "RoleGroupId", roleGroup.Id)
				return nil, err
			}
			policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(roleGroup.CasbinName), Obj: casbin2.Object(roleModel.Role)})
		}
	}
	//deleting policies from casbin
	impl.logger.Debugw("eliminated policies", "eliminatedPolicies", eliminatedPolicies)
	if len(eliminatedPolicies) > 0 {
		pRes := casbin2.RemovePolicy(eliminatedPolicies)
		impl.logger.Debugw("pRes : failed policies 1", "pRes", &pRes)
		println(pRes)
	}
	//updating in casbin
	if len(policies) > 0 {
		err = casbin2.AddPolicy(policies)
		if err != nil {
			impl.logger.Errorw("casbin policy addition failed", "err", err)
			return nil, err
		}
	}
	//loading policy for syncing orchestrator to casbin with newly added policies
	//(not calling this method in above if condition because we are also removing policies in this update service)
	casbin2.LoadPolicy()
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return request, nil
}

const (
	AllEnvironment string = ""
	AllNamespace   string = ""
	AllGroup       string = ""
	AllKind        string = ""
	AllResource    string = ""
	AllWorkflow    string = ""
)

func (impl RoleGroupServiceImpl) FetchRoleGroupsById(id int32) (*bean.RoleGroup, error) {
	roleGroup, err := impl.roleGroupRepository.GetRoleGroupById(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	roleFilters, isSuperAdmin := impl.getRoleGroupMetadata(roleGroup)
	bean := &bean.RoleGroup{
		Id:          roleGroup.Id,
		Name:        roleGroup.Name,
		Description: roleGroup.Description,
		RoleFilters: roleFilters,
		SuperAdmin:  isSuperAdmin,
	}
	return bean, nil
}

func (impl RoleGroupServiceImpl) getRoleGroupMetadata(roleGroup *repository.RoleGroup) ([]bean.RoleFilter, bool) {
	roles, err := impl.userAuthRepository.GetRolesByGroupId(roleGroup.Id)
	if err != nil {
		impl.logger.Errorw("No Roles Found for user", "roleGroupId", roleGroup.Id)
	}
	var roleFilters []bean.RoleFilter
	isSuperAdmin := false
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
	for index, roleFilter := range roleFilters {
		if roleFilter.Entity == "" {
			roleFilters[index].Entity = bean2.ENTITY_APPS
		}
		if roleFilter.Entity == bean2.ENTITY_APPS && roleFilter.AccessType == "" {
			roleFilters[index].AccessType = bean2.DEVTRON_APP
		}
	}
	if len(roleFilters) == 0 {
		roleFilters = make([]bean.RoleFilter, 0)
	}
	return roleFilters, isSuperAdmin
}

func (impl RoleGroupServiceImpl) FetchDetailedRoleGroups() ([]*bean.RoleGroup, error) {
	roleGroups, err := impl.roleGroupRepository.GetAllRoleGroup()
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	var list []*bean.RoleGroup
	for _, roleGroup := range roleGroups {
		roleFilters, isSuperAdmin := impl.getRoleGroupMetadata(roleGroup)
		for index, roleFilter := range roleFilters {
			if roleFilter.Entity == "" {
				roleFilters[index].Entity = bean2.ENTITY_APPS
			}
			if roleFilter.Entity == bean2.ENTITY_APPS && roleFilter.AccessType == "" {
				roleFilters[index].AccessType = bean2.DEVTRON_APP
			}
		}
		roleGrp := &bean.RoleGroup{
			Id:          roleGroup.Id,
			Name:        roleGroup.Name,
			Description: roleGroup.Description,
			RoleFilters: roleFilters,
			SuperAdmin:  isSuperAdmin,
		}
		list = append(list, roleGrp)
	}

	if len(list) == 0 {
		list = make([]*bean.RoleGroup, 0)
	}
	return list, nil
}

func (impl RoleGroupServiceImpl) FetchRoleGroups() (*bean.RoleGroupListingResponse, error) {
	roleGroup, err := impl.roleGroupRepository.GetAllRoleGroup()
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	list := impl.fetchRoleGroupResponseFromModel(roleGroup)
	response := &bean.RoleGroupListingResponse{
		RoleGroups: list,
		TotalCount: len(list),
	}
	return response, nil
}

// FetchRoleGroupsWithFilters takes FetchListingRequest as input and outputs RoleGroupListingResponse based on the request filters.
func (impl RoleGroupServiceImpl) FetchRoleGroupsWithFilters(request *bean.FetchListingRequest) (*bean.RoleGroupListingResponse, error) {
	if request.ShowAll {
		return impl.FetchRoleGroups()
	}
	// Setting size as zero to calculate the total number of results based on request
	size := request.Size
	request.Size = 0
	roleGroup, err := impl.roleGroupRepository.GetAllWithFilters(request)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	request.Size = size
	totalCount := len(roleGroup)

	// if total count is more than diff , then need to query with offset and limit(optimisation)
	if totalCount > (request.Size - request.Offset) {
		roleGroup, err = impl.roleGroupRepository.GetAllWithFilters(request)
		if err != nil {
			impl.logger.Errorw("error while fetching user from db", "error", err)
			return nil, err
		}
	}

	list := impl.fetchRoleGroupResponseFromModel(roleGroup)
	response := &bean.RoleGroupListingResponse{
		RoleGroups: list,
		TotalCount: totalCount,
	}
	return response, nil
}

func (impl RoleGroupServiceImpl) fetchRoleGroupResponseFromModel(roleGroup []*repository.RoleGroup) []*bean.RoleGroup {
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
	return list
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
	allRoleGroupRoleMappings, err := impl.roleGroupRepository.GetRoleGroupRoleMappingByRoleGroupId(model.Id)
	if err != nil {
		impl.logger.Errorw("error in getting all role group role mappings or not found", "err", err)
	}
	allRolesForGroup, err := casbin2.GetRolesForUser(model.CasbinName)
	if err != nil {
		impl.logger.Errorw("error in getting all roles for groups", "err", err)
	}
	for _, roleGroupRoleMapping := range allRoleGroupRoleMappings {
		err = impl.roleGroupRepository.DeleteRoleGroupRoleMappingByRoleId(roleGroupRoleMapping.RoleId, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting role group role mapping by role id", "err", err)
			return false, err
		}
	}
	model.Active = false
	model.UpdatedOn = time.Now()
	model.UpdatedBy = bean.UserId
	_, err = impl.roleGroupRepository.UpdateRoleGroup(model, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return false, err
	}

	allUsersMappedToGroup, err := casbin2.GetUserByRole(model.CasbinName)
	if err != nil {
		impl.logger.Errorw("error while fetching all users mapped to given group", "err", err)

	}
	for _, userMappedToGroup := range allUsersMappedToGroup {
		flag := casbin2.DeleteRoleForUser(userMappedToGroup, model.CasbinName)
		if flag == false {
			impl.logger.Warnw("unable to delete mapping of group and user in casbin", "user", model.CasbinName, "role", userMappedToGroup)
			return false, err
		}
	}

	for _, role := range allRolesForGroup {
		flag := casbin2.DeleteRoleForUser(model.CasbinName, role)
		if flag == false {
			impl.logger.Warnw("unable to delete mapping of group and user in casbin", "user", model.CasbinName, "role", role)
			return false, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (impl RoleGroupServiceImpl) FetchRoleGroupsWithRolesByGroupNames(groupNames []string) ([]*bean.RoleFilter, []bean.RoleGroup, error) {
	if len(groupNames) == 0 {
		return nil, nil, nil
	}
	roleGroups, err := impl.roleGroupRepository.GetRoleGroupListByNames(groupNames)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, nil, err
	}
	if err == pg.ErrNoRows || len(roleGroups) == 0 {
		impl.logger.Warnw("no result found for role groups", "groups", groupNames)
		return nil, nil, nil
	}

	return impl.fetchRolesFromRoleGroups(roleGroups)
}

func (impl RoleGroupServiceImpl) fetchRolesFromRoleGroups(roleGroups []*repository.RoleGroup) ([]*bean.RoleFilter, []bean.RoleGroup, error) {
	var roleGroupIds []int32
	var roleGroupResponse []bean.RoleGroup
	for _, roleGroup := range roleGroups {
		roleGroupIds = append(roleGroupIds, roleGroup.Id)
		roleGroupBean := bean.RoleGroup{
			Id:          roleGroup.Id,
			Name:        roleGroup.Name,
			Description: roleGroup.Description,
		}
		roleGroupResponse = append(roleGroupResponse, roleGroupBean)
	}

	roles, err := impl.roleGroupRepository.GetRoleGroupRoleMappingByRoleGroupIds(roleGroupIds)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, nil, err
	}
	list := make([]*bean.RoleFilter, 0)
	if roles == nil {
		return list, nil, nil
	}
	for _, role := range roles {
		bean := &bean.RoleFilter{
			EntityName:  role.EntityName,
			Entity:      role.Entity,
			Action:      role.Action,
			Environment: role.Environment,
			Team:        role.Team,
			AccessType:  role.AccessType,
			Cluster:     role.Cluster,
			Namespace:   role.Namespace,
			Group:       role.Group,
			Kind:        role.Kind,
			Resource:    role.Resource,
			Workflow:    role.Workflow,
		}
		list = append(list, bean)
	}
	return list, roleGroupResponse, nil
}

func (impl RoleGroupServiceImpl) FetchRoleGroupsWithRolesByGroupCasbinNames(groupCasbinNames []string) ([]*bean.RoleFilter, []bean.RoleGroup, error) {
	if len(groupCasbinNames) == 0 {
		return nil, nil, nil
	}
	roleGroups, err := impl.roleGroupRepository.GetRoleGroupListByCasbinNames(groupCasbinNames)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, nil, err
	}
	if err == pg.ErrNoRows || len(roleGroups) == 0 {
		impl.logger.Warnw("no result found for role groups", "groups", groupCasbinNames)
		return nil, nil, nil
	}

	return impl.fetchRolesFromRoleGroups(roleGroups)
}
