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
	bean2 "github.com/devtron-labs/devtron/pkg/user/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	util2 "github.com/devtron-labs/devtron/pkg/user/util"
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
	UpdateRoleGroup(request *bean.RoleGroup, token string, managerAuth func(resource, token string, object string) bool) (*bean.RoleGroup, error)
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
		model := &repository2.RoleGroup{
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
		for index, roleFilter := range request.RoleFilters {
			roleFilter = impl.userCommonService.ReplacePlaceHolderForEmptyEntriesInRoleFilter(roleFilter)
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

func (impl RoleGroupServiceImpl) CreateOrUpdateRoleGroupForClusterEntity(roleFilter bean.RoleFilter, userId int32, model *repository2.RoleGroup, existingRoles map[int]*repository2.RoleGroupRoleMapping, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, capacity int) ([]casbin2.Policy, error) {
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
					namespace = impl.userCommonService.RemovePlaceHolderInRoleFilterField(namespace)
					group = impl.userCommonService.RemovePlaceHolderInRoleFilterField(group)
					kind = impl.userCommonService.RemovePlaceHolderInRoleFilterField(kind)
					resource = impl.userCommonService.RemovePlaceHolderInRoleFilterField(resource)
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
							roleGroupMappingModel := &repository2.RoleGroupRoleMapping{RoleGroupId: model.Id, RoleId: roleModel.Id}
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

func (impl RoleGroupServiceImpl) CreateOrUpdateRoleGroupForOtherEntity(roleFilter bean.RoleFilter, request *bean.RoleGroup, model *repository2.RoleGroup, existingRoles map[int]*repository2.RoleGroupRoleMapping, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, capacity int) ([]casbin2.Policy, error) {
	var policiesToBeAdded = make([]casbin2.Policy, 0, capacity)
	accessType := roleFilter.AccessType
	entityNames := strings.Split(roleFilter.EntityName, ",")
	environments := strings.Split(roleFilter.Environment, ",")
	entity := roleFilter.Entity
	actions := strings.Split(roleFilter.Action, ",")
	for _, environment := range environments {
		for _, entityName := range entityNames {
			for _, actionType := range actions {
				entityName = impl.userCommonService.RemovePlaceHolderInRoleFilterField(entityName)
				environment = impl.userCommonService.RemovePlaceHolderInRoleFilterField(environment)
				roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, roleFilter.Team, entityName, environment, actionType, roleFilter.Approver, accessType, "", "", "", "", "", "", false, "")
				if err != nil {
					return nil, err
				}
				if roleModel.Id == 0 {
					request.Status = bean2.RoleNotFoundStatusPrefix + roleFilter.Team + "," + environment + "," + entityName + "," + actionType
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
							request.Status = bean2.RoleNotFoundStatusPrefix + roleFilter.Team + "," + environment + "," + entityName + "," + actionType
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
						roleGroupMappingModel := &repository2.RoleGroupRoleMapping{RoleGroupId: model.Id, RoleId: roleModel.Id}
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

func (impl RoleGroupServiceImpl) CreateOrUpdateRoleGroupForJobsEntity(roleFilter bean.RoleFilter, userId int32, model *repository2.RoleGroup, existingRoles map[int]*repository2.RoleGroupRoleMapping, token string, managerAuth func(resource string, token string, object string) bool, tx *pg.Tx, capacity int) ([]casbin2.Policy, error) {
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
				entityName = impl.userCommonService.RemovePlaceHolderInRoleFilterField(entityName)
				environment = impl.userCommonService.RemovePlaceHolderInRoleFilterField(environment)
				workflow = impl.userCommonService.RemovePlaceHolderInRoleFilterField(workflow)
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
						roleGroupMappingModel := &repository2.RoleGroupRoleMapping{RoleGroupId: model.Id, RoleId: roleModel.Id}
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

	//loading policy for safety
	casbin2.LoadPolicy()

	// DELETE PROCESS STARTS
	var eliminatedPolicies []casbin2.Policy
	items, err := impl.userCommonService.RemoveRolesAndReturnEliminatedPoliciesForGroups(request, existingRoles, eliminatedRoles, tx, token, managerAuth)
	if err != nil {
		return nil, err
	}
	eliminatedPolicies = append(eliminatedPolicies, items...)
	impl.logger.Debugw("eliminated policies", "eliminatedPolicies", eliminatedPolicies)
	if len(eliminatedPolicies) > 0 {
		pRes := casbin2.RemovePolicy(eliminatedPolicies)
		impl.logger.Debugw("pRes : failed policies 1", "pRes", &pRes)
		println(pRes)
	}
	// DELETE PROCESS ENDS

	//Adding New Policies
	capacity, mapping := impl.userCommonService.GetCapacityForRoleFilter(request.RoleFilters)
	var policies = make([]casbin2.Policy, 0, capacity)
	for index, roleFilter := range request.RoleFilters {
		roleFilter = impl.userCommonService.ReplacePlaceHolderForEmptyEntriesInRoleFilter(roleFilter)
		if roleFilter.Entity == bean.CLUSTER_ENTITIY {
			policiesToBeAdded, err := impl.CreateOrUpdateRoleGroupForClusterEntity(roleFilter, request.UserId, roleGroup, existingRoles, token, managerAuth, tx, mapping[index])
			policies = append(policies, policiesToBeAdded...)
			if err != nil {
				impl.logger.Errorw("error in creating updating role group for cluster entity", "err", err, "roleFilter", roleFilter)
			}
		} else {
			if len(roleFilter.Team) > 0 {
				// check auth only for apps permission, skip for chart group
				rbacObject := fmt.Sprintf("%s", strings.ToLower(roleFilter.Team))
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
			key = fmt.Sprintf("%s_%s_%s_%t", role.Team, role.Action, role.AccessType, role.Approver)
		} else if role.Entity == bean2.EntityJobs {
			key = fmt.Sprintf("%s_%s_%s_%s", role.Team, role.Action, role.AccessType, role.Entity)
		} else if len(role.Entity) > 0 {
			if role.Entity == bean.CLUSTER_ENTITIY {
				key = fmt.Sprintf("%s_%s_%s_%s_%s_%s", role.Entity, role.Action, role.Cluster,
					role.Namespace, role.Group, role.Kind)
			} else {
				key = fmt.Sprintf("%s_%s", role.Entity, role.Action)
			}
		}
		if _, ok := roleFilterMap[key]; ok {
			if role.Entity == bean.CLUSTER_ENTITIY {
				namespaceArr := strings.Split(roleFilterMap[key].Namespace, ",")
				if containsArr(namespaceArr, AllNamespace) {
					roleFilterMap[key].Namespace = AllNamespace
				} else if !containsArr(namespaceArr, role.Namespace) {
					roleFilterMap[key].Namespace = fmt.Sprintf("%s,%s", roleFilterMap[key].Namespace, role.Namespace)
				}
				groupArr := strings.Split(roleFilterMap[key].Group, ",")
				if containsArr(groupArr, AllGroup) {
					roleFilterMap[key].Group = AllGroup
				} else if !containsArr(groupArr, role.Group) {
					roleFilterMap[key].Group = fmt.Sprintf("%s,%s", roleFilterMap[key].Group, role.Group)
				}
				kindArr := strings.Split(roleFilterMap[key].Kind, ",")
				if containsArr(kindArr, AllKind) {
					roleFilterMap[key].Kind = AllKind
				} else if !containsArr(kindArr, role.Kind) {
					roleFilterMap[key].Kind = fmt.Sprintf("%s,%s", roleFilterMap[key].Kind, role.Kind)
				}
				resourceArr := strings.Split(roleFilterMap[key].Resource, ",")
				if containsArr(resourceArr, AllResource) {
					roleFilterMap[key].Resource = AllResource
				} else if !containsArr(resourceArr, role.Resource) {
					roleFilterMap[key].Resource = fmt.Sprintf("%s,%s", roleFilterMap[key].Resource, role.Resource)
				}
			} else if role.Entity == bean2.EntityJobs {
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
				workflowArr := strings.Split(roleFilterMap[key].Workflow, ",")
				if containsArr(workflowArr, AllWorkflow) {
					roleFilterMap[key].Workflow = AllWorkflow
				} else if !containsArr(workflowArr, role.Workflow) {
					roleFilterMap[key].Workflow = fmt.Sprintf("%s,%s", roleFilterMap[key].Workflow, role.Workflow)
				}
			} else {
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
			}
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
	}
	for _, v := range roleFilterMap {
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
