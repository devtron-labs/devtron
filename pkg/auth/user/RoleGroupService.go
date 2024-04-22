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
	"github.com/devtron-labs/devtron/pkg/auth/user/repository/helper"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	casbin2 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	util2 "github.com/devtron-labs/devtron/pkg/auth/user/util"
	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

type RoleGroupService interface {
	CreateRoleGroup(request *bean.RoleGroup) (*bean.RoleGroup, error)
	UpdateRoleGroup(request *bean.RoleGroup, token string, managerAuth func(resource, token string, object string) bool) (*bean.RoleGroup, error)
	FetchDetailedRoleGroups(req *bean.ListingRequest) ([]*bean.RoleGroup, error)
	FetchRoleGroupsById(id int32) (*bean.RoleGroup, error)
	FetchRoleGroups() ([]*bean.RoleGroup, error)
	FetchRoleGroupsV2(req *bean.ListingRequest) (*bean.RoleGroupListingResponse, error)
	FetchRoleGroupsWithFilters(request *bean.ListingRequest) (*bean.RoleGroupListingResponse, error)
	FetchRoleGroupsByName(name string) ([]*bean.RoleGroup, error)
	DeleteRoleGroup(model *bean.RoleGroup) (bool, error)
	BulkDeleteRoleGroups(request *bean.BulkDeleteRequest) (bool, error)
	FetchRolesForUserRoleGroups(userRoleGroups []bean.UserRoleGroup) ([]*bean.RoleFilter, error)
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
			roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes("", "", "", "", bean2.SUPER_ADMIN, "", "", "", "", "", "", "", false, "")
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
			pRes := casbin2.AddPolicy(policies)
			println(pRes)
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
					roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, "", "", "", "", accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, false, "")
					if err != nil {
						impl.logger.Errorw("error in getting new role model by filter")
						return policiesToBeAdded, err
					}
					if roleModel.Id == 0 {
						flag, err, policiesAdded := impl.userCommonService.CreateDefaultPoliciesForAllTypes("", "", "", entity, roleFilter.Cluster, namespace, group, kind, resource, actionType, accessType, "", userId)
						if err != nil || flag == false {
							return policiesToBeAdded, err
						}
						policiesToBeAdded = append(policiesToBeAdded, policiesAdded...)
						roleModel, err = impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, "", "", "", "", accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, false, "")
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
	actionType := roleFilter.Action
	accessType := roleFilter.AccessType
	entity := roleFilter.Entity
	entityNames := strings.Split(roleFilter.EntityName, ",")
	environments := strings.Split(roleFilter.Environment, ",")
	var policiesToBeAdded = make([]casbin2.Policy, 0, capacity)
	for _, environment := range environments {
		for _, entityName := range entityNames {
			roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, accessType, "", "", "", "", "", "", false, "")
			if err != nil {
				impl.logger.Errorw("error in getting new role model")
				return nil, err
			}
			if roleModel.Id == 0 {
				request.Status = fmt.Sprintf("%s+%s,%s,%s,%s", bean2.RoleNotFoundStatusPrefix, roleFilter.Team, environment, entityName, actionType)
				if roleFilter.Entity == bean2.ENTITY_APPS || roleFilter.Entity == bean.CHART_GROUP_ENTITY {
					flag, err, policiesAdded := impl.userCommonService.CreateDefaultPoliciesForAllTypes(roleFilter.Team, entityName, environment, entity, "", "", "", "", "", actionType, accessType, "", request.UserId)
					if err != nil || flag == false {
						return nil, err
					}
					policiesToBeAdded = append(policiesToBeAdded, policiesAdded...)
					roleModel, err = impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, accessType, "", "", "", "", "", "", false, "")
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
				//Adding policies which are removed
				policiesToBeAdded = append(policiesToBeAdded, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.CasbinName), Obj: casbin2.Object(roleModel.Role)})
			} else {
				if roleModel.Id > 0 {
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
				roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, accessType, "", "", "", "", "", "", false, workflow)
				if err != nil {
					impl.logger.Errorw("error in getting new role model")
					return nil, err
				}
				if roleModel.Id == 0 {
					flag, err, policiesAdded := impl.userCommonService.CreateDefaultPoliciesForAllTypes(roleFilter.Team, entityName, environment, entity, "", "", "", "", "", actionType, accessType, workflow, userId)
					if err != nil || flag == false {
						return nil, err
					}
					policiesToBeAdded = append(policiesToBeAdded, policiesAdded...)
					roleModel, err = impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, accessType, "", "", "", "", "", "", false, workflow)
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
	existingRoles := make(map[int]*repository.RoleGroupRoleMapping)
	eliminatedRoles := make(map[int]*repository.RoleGroupRoleMapping)
	var eliminatedPolicies []casbin2.Policy
	capacity, mapping := impl.userCommonService.GetCapacityForRoleFilter(request.RoleFilters)
	var policies = make([]casbin2.Policy, 0, capacity)
	if request.SuperAdmin == false {
		roleGroupMappingModels, err := impl.roleGroupRepository.GetRoleGroupRoleMappingByRoleGroupId(roleGroup.Id)
		if err != nil {
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
		roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes("", "", "", "", bean2.SUPER_ADMIN, "", "", "", "", "", "", "", false, "")
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
		pRes := casbin2.AddPolicy(policies)
		impl.logger.Debugw("pres failed policies on add policy", "pres", &pRes)
		println(pRes)
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

	roleFilters, superAdmin := impl.getRoleGroupMetadata(roleGroup)
	bean := &bean.RoleGroup{
		Id:          roleGroup.Id,
		Name:        roleGroup.Name,
		Description: roleGroup.Description,
		RoleFilters: roleFilters,
		SuperAdmin:  superAdmin,
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
				Workflow:    role.Workflow,
			}
		}
		if role.Role == bean.SUPERADMIN {
			isSuperAdmin = true
		}
	}
	for _, v := range roleFilterMap {
		if v.Action == bean2.SUPER_ADMIN {
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
	if len(roleFilters) == 0 {
		roleFilters = make([]bean.RoleFilter, 0)
	}
	return roleFilters, isSuperAdmin
}

func (impl RoleGroupServiceImpl) FetchDetailedRoleGroups(req *bean.ListingRequest) ([]*bean.RoleGroup, error) {
	query := helper.GetQueryForGroupListingWithFilters(req)
	roleGroups, err := impl.roleGroupRepository.GetAllExecutingQuery(query)
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

func (impl RoleGroupServiceImpl) FetchRoleGroupsV2(req *bean.ListingRequest) (*bean.RoleGroupListingResponse, error) {
	list, err := impl.FetchDetailedRoleGroups(req)
	if err != nil {
		impl.logger.Errorw("error in FetchDetailedRoleGroups", "err", err)
		return nil, err
	}
	response := &bean.RoleGroupListingResponse{
		RoleGroups: list,
		TotalCount: len(list),
	}
	return response, nil
}

// FetchRoleGroupsWithFilters takes listing request as input and outputs RoleGroupListingResponse based on the request filters.
func (impl RoleGroupServiceImpl) FetchRoleGroupsWithFilters(request *bean.ListingRequest) (*bean.RoleGroupListingResponse, error) {
	// default values will be used if not provided
	impl.userCommonService.SetDefaultValuesIfNotPresent(request, true)
	if request.ShowAll {
		return impl.FetchRoleGroupsV2(request)
	}

	// setting count check to true for getting only count
	request.CountCheck = true
	query := helper.GetQueryForGroupListingWithFilters(request)
	totalCount, err := impl.userRepository.GetCountExecutingQuery(query)
	if err != nil {
		impl.logger.Errorw("error in FetchRoleGroupsWithFilters", "err", err, "query", query)
		return nil, err
	}
	// setting count check to false for getting data
	request.CountCheck = false

	query = helper.GetQueryForGroupListingWithFilters(request)
	roleGroup, err := impl.roleGroupRepository.GetAllExecutingQuery(query)
	if err != nil {
		impl.logger.Errorw("error while FetchRoleGroupsWithFilters", "error", err, "query", query)
		return nil, err
	}

	response := impl.fetchRoleGroupResponseFromModel(roleGroup, totalCount)
	return response, nil
}

func (impl RoleGroupServiceImpl) fetchRoleGroupResponseFromModel(roleGroup []*repository.RoleGroup, totalCount int) *bean.RoleGroupListingResponse {
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

	response := &bean.RoleGroupListingResponse{
		RoleGroups: list,
		TotalCount: totalCount,
	}
	return response
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
	roleGroupRoleMappingIds, err := impl.roleGroupRepository.GetRoleGroupRoleMappingIdsByRoleGroupId(model.Id)
	if err != nil {
		impl.logger.Errorw("error in getting all role group role mappings or not found", "err", err)
	}
	allRolesForGroup, err := casbin2.GetRolesForUser(model.CasbinName)
	if err != nil {
		impl.logger.Errorw("error in getting all roles for groups", "err", err)
	}
	if len(roleGroupRoleMappingIds) > 0 {
		err = impl.roleGroupRepository.DeleteRoleGroupRoleMappingByIds(roleGroupRoleMappingIds, tx)
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

// BulkDeleteRoleGroups takes in bulk delete request and return error
func (impl RoleGroupServiceImpl) BulkDeleteRoleGroups(request *bean.BulkDeleteRequest) (bool, error) {
	// it handles ListingRequest if filters are applied will delete those users or will consider the given user ids.
	if request.ListingRequest != nil {
		filteredGroupIds, err := impl.getGroupIdsHonoringFilters(request.ListingRequest)
		if err != nil {
			impl.logger.Errorw("error in BulkDeleteRoleGroups", "request", request, "err", err)
			return false, err
		}
		// setting the filtered user ids here for further processing
		request.Ids = filteredGroupIds
	}

	err := impl.deleteRoleGroupsByIds(request)
	if err != nil {
		impl.logger.Errorw("error in BulkDeleteRoleGroups", "request", request, "error", err)
		return false, err
	}
	return true, nil
}

// getGroupIdsHonoringFilters get the filtered group ids according to the request filters and returns groupIds and error(not nil) if any exception is caught.
func (impl *RoleGroupServiceImpl) getGroupIdsHonoringFilters(request *bean.ListingRequest) ([]int32, error) {
	//query to get particular models respecting filters
	query := helper.GetQueryForGroupListingWithFilters(request)
	models, err := impl.roleGroupRepository.GetAllExecutingQuery(query)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db in getGroupIdsHonoringFilters", "error", err)
		return nil, err
	}
	// collecting the required group ids from filtered models
	filteredGroupIds := make([]int32, len(models))
	for i, model := range models {
		filteredGroupIds[i] = model.Id
	}
	return filteredGroupIds, nil
}

// deleteRoleGroupsByIds delete role groups by ids takes in bulk delete request and return error
func (impl RoleGroupServiceImpl) deleteRoleGroupsByIds(request *bean.BulkDeleteRequest) error {
	tx, err := impl.roleGroupRepository.StartATransaction()
	if err != nil {
		impl.logger.Errorw("error in starting a transaction", "err", err)
		return &util.ApiError{Code: "500", HttpStatusCode: 500, UserMessage: "error starting a transaction in db", InternalMessage: "error starting a transaction in db"}
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// get casbin names
	groupCasbinNames, err := impl.roleGroupRepository.GetCasbinNamesById(request.Ids)
	if err != nil {
		impl.logger.Errorw("error in deleteRoleGroupsByIds", "request", request, "err", err)
		return err
	}
	// delete mappings from orchestrator
	err = impl.deleteMappingsFromOrchestrator(request.Ids, tx)
	if err != nil {
		impl.logger.Errorw("error in deleteRoleGroupsByIds", "request", request, "err", err)
		return err
	}
	// update models to inactive with audit
	err = impl.roleGroupRepository.UpdateToInactiveByIds(request.Ids, tx, request.LoggedInUserId)
	if err != nil {
		impl.logger.Errorw("error in deleteMappingsFromOrchestrator", "err", err)
		return err
	}

	// delete from casbin
	err = impl.deleteMappingsFromCasbin(groupCasbinNames, len(request.Ids))
	if err != nil {
		impl.logger.Errorw("error in deleteRoleGroupsByIds", "request", request, "err", err)
		return err
	}
	// commit transaction
	err = impl.roleGroupRepository.CommitATransaction(tx)
	if err != nil {
		impl.logger.Errorw("error in committing a transaction in deleteRoleGroupsByIds", "err", err)
		return &util.ApiError{Code: "500", HttpStatusCode: 500, UserMessage: "error committing a transaction in db", InternalMessage: "error committing a transaction in db"}
	}
	return nil

}

// deleteMappingsFromOrchestrator deletes role group role mapping from orchestrator only, takes in ids and returns error
func (impl RoleGroupServiceImpl) deleteMappingsFromOrchestrator(roleGroupIds []int32, tx *pg.Tx) error {
	mappingIds, err := impl.roleGroupRepository.GetRoleGroupRoleMappingIdsByGroupIds(roleGroupIds)
	if err != nil {
		impl.logger.Errorw("error in deleteMappingsFromOrchestrator", "err", err)
		return err
	}

	if len(mappingIds) > 0 {
		err = impl.roleGroupRepository.DeleteRoleGroupRoleMappingByIds(mappingIds, tx)
		if err != nil {
			impl.logger.Errorw("error in deleteMappingsFromOrchestrator", "err", err)
			return err
		}
	}
	return nil
}

// deleteMappingsFromCasbin delete GROUP-POLICY mappings and USER-GROUP mappings from casbin
func (impl RoleGroupServiceImpl) deleteMappingsFromCasbin(groupCasbinNames []string, totalCount int) error {
	groupNameVsCasbinRolesMap := make(map[string][]string, totalCount)
	groupVsUsersMap := make(map[string][]string, totalCount)
	for _, casbinName := range groupCasbinNames {
		casbinRoles, err := casbin2.GetRolesForUser(casbinName)
		if err != nil {
			impl.logger.Warnw("No Roles Found for user", "casbinName", casbinName, "err", err)
			return err
		}
		allUsersMappedToGroup, err := casbin2.GetUserByRole(casbinName)
		if err != nil {
			impl.logger.Errorw("error while fetching all users mapped to given group", "err", err)
			return err
		}
		groupNameVsCasbinRolesMap[casbinName] = casbinRoles
		groupVsUsersMap[casbinName] = allUsersMappedToGroup

	}
	// GROUP-POLICY mapping deletion from casbin
	success := impl.userCommonService.DeleteRoleForUserFromCasbin(groupNameVsCasbinRolesMap)
	if !success {
		impl.logger.Errorw("error in deleteMappingsFromCasbin, not all mappings removed ", "groupCasbinNames", groupCasbinNames)
		return &util.ApiError{Code: "500", HttpStatusCode: 500, InternalMessage: "Not able to delete mappings from casbin", UserMessage: "Not able to delete mappings from casbin"}
	}

	// USER-GROUP mapping deletion from casbin
	success = impl.userCommonService.DeleteUserForRoleFromCasbin(groupVsUsersMap)
	if !success {
		impl.logger.Errorw("error in deleteMappingsFromCasbin, not all mappings removed ", "groupCasbinNames", groupCasbinNames)
		return &util.ApiError{Code: "500", HttpStatusCode: 500, InternalMessage: "Not able to delete mappings from casbin", UserMessage: "Not able to delete mappings from casbin"}
	}

	return nil
}

func (impl RoleGroupServiceImpl) FetchRolesForUserRoleGroups(userRoleGroups []bean.UserRoleGroup) ([]*bean.RoleFilter, error) {
	groupNames := make([]string, 0)
	for _, userRoleGroup := range userRoleGroups {
		groupNames = append(groupNames, userRoleGroup.RoleGroup.Name)
	}
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

	roles, err := impl.roleGroupRepository.GetRolesByRoleGroupIds(roleGroupIds)
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
			Cluster:     role.Cluster,
			Namespace:   role.Namespace,
			Group:       role.Group,
			Kind:        role.Kind,
			Resource:    role.Resource,
			Workflow:    role.Workflow,
		}
		list = append(list, bean)
	}
	return list, nil
}
