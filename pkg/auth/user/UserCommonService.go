/*
 * Copyright (c) 2024. Devtron Inc.
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
	"fmt"
	bean3 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/adapter"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository/bean"
	"golang.org/x/exp/maps"
	"math"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

type UserCommonService interface {
	CreateDefaultPoliciesForAllTypes(roleFieldsDto *bean.RoleModelFieldsDto, userId int32) (bool, error, []bean3.Policy)
	RemoveRolesAndReturnEliminatedPolicies(userInfo *bean2.UserInfo, existingRoleIds map[int]repository.UserRoleModel, eliminatedRoleIds map[int]*repository.UserRoleModel, tx *pg.Tx, token string, managerAuth func(resource string, token string, object string) bool) ([]bean3.Policy, []*repository.RoleModel, error)
	RemoveRolesAndReturnEliminatedPoliciesForGroups(request *bean2.RoleGroup, existingRoles map[int]*repository.RoleGroupRoleMapping, eliminatedRoles map[int]*repository.RoleGroupRoleMapping, tx *pg.Tx, token string, managerAuth func(resource string, token string, object string) bool) ([]bean3.Policy, []*repository.RoleModel, error)
	CheckRbacForClusterEntity(cluster, namespace, group, kind, resource, token string, managerAuth func(resource, token, object string) bool) bool
	GetCapacityForRoleFilter(roleFilters []bean2.RoleFilter) (int, map[int]int)
	BuildRoleFilterForAllTypes(roleFilterMap map[string]*bean2.RoleFilter, entityProcessor EntityKeyProcessor, key string, basetoConsider bean2.MergingBaseKey)
	GetUniqueKeyForAllEntity(entityProcessor EntityKeyProcessor, baseToConsider bean2.MergingBaseKey) string
	SetDefaultValuesIfNotPresent(request *bean2.ListingRequest, isRoleGroup bool)
	DeleteRoleForUserFromCasbin(mappings map[string][]string) bool
	DeleteUserForRoleFromCasbin(mappings map[string][]string) bool
	BuildRoleFiltersAfterMerging(entityProcessors []EntityKeyProcessor, baseKey bean2.MergingBaseKey) []bean2.RoleFilter
}

// Defining a common interface for role and rolefilter
type EntityKeyProcessor interface {
	GetTeam() string
	GetEntity() string
	GetAction() string
	GetAccessType() string
	GetEnvironment() string
	GetCluster() string
	GetGroup() string
	GetKind() string
	GetEntityName() string
	GetResource() string
	GetWorkflow() string
	GetNamespace() string
}

type UserCommonServiceImpl struct {
	userAuthRepository          repository.UserAuthRepository
	logger                      *zap.SugaredLogger
	userRepository              repository.UserRepository
	roleGroupRepository         repository.RoleGroupRepository
	sessionManager2             *middleware.SessionManager
	defaultRbacDataCacheFactory repository.RbacDataCacheFactory
	userRbacConfig              *UserRbacConfig
}

func NewUserCommonServiceImpl(userAuthRepository repository.UserAuthRepository,
	logger *zap.SugaredLogger,
	userRepository repository.UserRepository,
	userGroupRepository repository.RoleGroupRepository,
	sessionManager2 *middleware.SessionManager,
	defaultRbacDataCacheFactory repository.RbacDataCacheFactory) (*UserCommonServiceImpl, error) {
	userConfig := &UserRbacConfig{}
	err := env.Parse(userConfig)
	if err != nil {
		logger.Errorw("error occurred while parsing user config", err)
		return nil, err
	}
	serviceImpl := &UserCommonServiceImpl{
		userAuthRepository:          userAuthRepository,
		logger:                      logger,
		userRepository:              userRepository,
		roleGroupRepository:         userGroupRepository,
		sessionManager2:             sessionManager2,
		defaultRbacDataCacheFactory: defaultRbacDataCacheFactory,
		userRbacConfig:              userConfig,
	}
	cStore = sessions.NewCookieStore(randKey())
	defaultRbacDataCacheFactory.SyncPolicyCache()
	defaultRbacDataCacheFactory.SyncRoleDataCache()
	return serviceImpl, nil
}

type UserRbacConfig struct {
	UseRbacCreationV2 bool `env:"USE_RBAC_CREATION_V2" envDefault:"true" description:"To use the V2 for RBAC creation"`
}

func (impl UserCommonServiceImpl) CreateDefaultPoliciesForAllTypes(roleFieldsDto *bean.RoleModelFieldsDto, userId int32) (bool, error, []bean3.Policy) {
	team, entityName, env, entity, cluster, namespace, group, kind,
		resource, actionType, accessType, workflow := roleFieldsDto.Team, roleFieldsDto.App, roleFieldsDto.Env,
		roleFieldsDto.Entity, roleFieldsDto.Cluster, roleFieldsDto.Namespace, roleFieldsDto.Group,
		roleFieldsDto.Kind, roleFieldsDto.Resource, roleFieldsDto.Action, roleFieldsDto.AccessType, roleFieldsDto.Workflow
	if impl.userRbacConfig.UseRbacCreationV2 {
		impl.logger.Debugw("using rbac creation v2 for creating default policies")
		return impl.CreateDefaultPoliciesForAllTypesV2(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType, workflow)
	} else {
		return impl.userAuthRepository.CreateDefaultPoliciesForAllTypes(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType, userId)
	}
}

func (impl UserCommonServiceImpl) CreateDefaultPoliciesForAllTypesV2(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType, workflow string) (bool, error, []bean3.Policy) {
	//TODO: below txn is making this process slow, need to do bulk operation for role creation.
	//For detail - https://github.com/devtron-labs/devtron/blob/main/pkg/user/benchmarking-results

	renderedRole, renderedPolicyDetails, err := impl.getRenderedRoleAndPolicy(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType, workflow)
	if err != nil {
		return false, err, nil
	}
	_, err = impl.userAuthRepository.CreateRole(renderedRole)
	if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
		return false, err, nil
	}
	return true, nil, renderedPolicyDetails
}

func (impl UserCommonServiceImpl) getRenderedRoleAndPolicy(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType, workflow string) (*repository.RoleModel, []bean3.Policy, error) {
	//getting map of values to be used for rendering
	pValUpdateMap := getPValUpdateMap(team, entityName, env, entity, cluster, namespace, group, kind, resource, workflow)

	//getting default role data and policy
	defaultRoleData, defaultPolicy, err := impl.getDefaultRbacRoleAndPolicyByRoleFilter(entity, accessType, actionType)
	if err != nil {
		return nil, nil, err
	}
	//getting rendered role and policy data
	renderedRoleData := getRenderedRoleData(defaultRoleData, pValUpdateMap)
	renderedPolicy := getRenderedPolicy(defaultPolicy, pValUpdateMap)

	return renderedRoleData, renderedPolicy, nil
}

func (impl UserCommonServiceImpl) getDefaultRbacRoleAndPolicyByRoleFilter(entity, accessType, action string) (repository.RoleCacheDetailObj, repository.PolicyCacheDetailObj, error) {
	//getting default role and policy data from cache
	return impl.defaultRbacDataCacheFactory.
		GetDefaultRoleDataAndPolicyByEntityAccessTypeAndRoleType(entity, accessType, action)
}

func getRenderedRoleData(defaultRoleData repository.RoleCacheDetailObj, pValUpdateMap map[repository.PValUpdateKey]string) *repository.RoleModel {
	renderedRoleData := &repository.RoleModel{
		Role:        getResolvedValueFromPValDetailObject(defaultRoleData.Role, pValUpdateMap),
		Entity:      getResolvedValueFromPValDetailObject(defaultRoleData.Entity, pValUpdateMap),
		EntityName:  getResolvedValueFromPValDetailObject(defaultRoleData.EntityName, pValUpdateMap),
		Team:        getResolvedValueFromPValDetailObject(defaultRoleData.Team, pValUpdateMap),
		Environment: getResolvedValueFromPValDetailObject(defaultRoleData.Environment, pValUpdateMap),
		AccessType:  getResolvedValueFromPValDetailObject(defaultRoleData.AccessType, pValUpdateMap),
		Action:      getResolvedValueFromPValDetailObject(defaultRoleData.Action, pValUpdateMap),
		Cluster:     getResolvedValueFromPValDetailObject(defaultRoleData.Cluster, pValUpdateMap),
		Namespace:   getResolvedValueFromPValDetailObject(defaultRoleData.Namespace, pValUpdateMap),
		Group:       getResolvedValueFromPValDetailObject(defaultRoleData.Group, pValUpdateMap),
		Kind:        getResolvedValueFromPValDetailObject(defaultRoleData.Kind, pValUpdateMap),
		Resource:    getResolvedValueFromPValDetailObject(defaultRoleData.Resource, pValUpdateMap),
		Workflow:    getResolvedValueFromPValDetailObject(defaultRoleData.Workflow, pValUpdateMap),
		AuditLog: sql.AuditLog{ //not storing user information because this role can be mapped to other users in future and hence can lead to confusion
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
		},
	}
	return renderedRoleData
}

func getRenderedPolicy(defaultPolicy repository.PolicyCacheDetailObj, pValUpdateMap map[repository.PValUpdateKey]string) []bean3.Policy {
	renderedPolicies := make([]bean3.Policy, 0, len(defaultPolicy.ResActObjSet))
	policyType := getResolvedValueFromPValDetailObject(defaultPolicy.Type, pValUpdateMap)
	policySub := getResolvedValueFromPValDetailObject(defaultPolicy.Sub, pValUpdateMap)
	for _, v := range defaultPolicy.ResActObjSet {
		policyRes := getResolvedValueFromPValDetailObject(v.Res, pValUpdateMap)
		policyAct := getResolvedValueFromPValDetailObject(v.Act, pValUpdateMap)
		policyObj := getResolvedValueFromPValDetailObject(v.Obj, pValUpdateMap)
		renderedPolicy := bean3.Policy{
			Type: bean3.PolicyType(policyType),
			Sub:  bean3.Subject(policySub),
			Res:  bean3.Resource(policyRes),
			Act:  bean3.Action(policyAct),
			Obj:  bean3.Object(policyObj),
		}
		renderedPolicies = append(renderedPolicies, renderedPolicy)
	}
	return renderedPolicies
}

func getResolvedValueFromPValDetailObject(pValDetailObj repository.PValDetailObj, pValUpdateMap map[repository.PValUpdateKey]string) string {
	if len(pValDetailObj.IndexKeyMap) == 0 {
		return pValDetailObj.Value
	}
	pValBytes := []byte(pValDetailObj.Value)
	var resolvedValueInBytes []byte
	for i, pValByte := range pValBytes {
		if pValByte == '%' {
			valUpdateKey := pValDetailObj.IndexKeyMap[i]
			val := pValUpdateMap[valUpdateKey]
			resolvedValueInBytes = append(resolvedValueInBytes, []byte(val)...)
		} else {
			resolvedValueInBytes = append(resolvedValueInBytes, pValByte)
		}
	}
	return string(resolvedValueInBytes)
}

func getPValUpdateMap(team, entityName, env, entity, cluster, namespace, group, kind, resource, workflow string) map[repository.PValUpdateKey]string {
	pValUpdateMap := make(map[repository.PValUpdateKey]string)
	pValUpdateMap[repository.EntityPValUpdateKey] = entity
	if entity == bean2.CLUSTER_ENTITIY {
		pValUpdateMap[repository.ClusterPValUpdateKey] = cluster
		pValUpdateMap[repository.NamespacePValUpdateKey] = namespace
		pValUpdateMap[repository.GroupPValUpdateKey] = group
		pValUpdateMap[repository.KindPValUpdateKey] = kind
		pValUpdateMap[repository.ResourcePValUpdateKey] = resource
		pValUpdateMap[repository.ClusterObjPValUpdateKey] = getResolvedPValMapValue(cluster)
		pValUpdateMap[repository.NamespaceObjPValUpdateKey] = getResolvedPValMapValue(namespace)
		pValUpdateMap[repository.GroupObjPValUpdateKey] = getResolvedPValMapValue(group)
		pValUpdateMap[repository.KindObjPValUpdateKey] = getResolvedPValMapValue(kind)
		pValUpdateMap[repository.ResourceObjPValUpdateKey] = getResolvedPValMapValue(resource)
	} else {
		pValUpdateMap[repository.EntityNamePValUpdateKey] = entityName
		pValUpdateMap[repository.TeamPValUpdateKey] = team
		pValUpdateMap[repository.AppPValUpdateKey] = entityName
		pValUpdateMap[repository.EnvPValUpdateKey] = env
		pValUpdateMap[repository.TeamObjPValUpdateKey] = getResolvedPValMapValue(team)
		pValUpdateMap[repository.AppObjPValUpdateKey] = getResolvedPValMapValue(entityName)
		pValUpdateMap[repository.EnvObjPValUpdateKey] = getResolvedPValMapValue(env)
		if entity == bean2.EntityJobs {
			pValUpdateMap[repository.WorkflowPValUpdateKey] = workflow
			pValUpdateMap[repository.WorkflowObjPValUpdateKey] = getResolvedPValMapValue(workflow)
		}
	}
	return pValUpdateMap
}

func getResolvedPValMapValue(rawValue string) string {
	resolvedVal := rawValue
	if rawValue == "" {
		resolvedVal = "*"
	}
	return resolvedVal
}

func (impl UserCommonServiceImpl) RemoveRolesAndReturnEliminatedPolicies(userInfo *bean2.UserInfo,
	existingRoleIds map[int]repository.UserRoleModel, eliminatedRoleIds map[int]*repository.UserRoleModel,
	tx *pg.Tx, token string, managerAuth func(resource string, token string, object string) bool) ([]bean3.Policy, []*repository.RoleModel, error) {
	var eliminatedPolicies []bean3.Policy
	// DELETE Removed Items
	for _, roleFilter := range userInfo.RoleFilters {
		if roleFilter.Entity == bean2.CLUSTER_ENTITIY {
			namespaces := strings.Split(roleFilter.Namespace, ",")
			groups := strings.Split(roleFilter.Group, ",")
			kinds := strings.Split(roleFilter.Kind, ",")
			resources := strings.Split(roleFilter.Resource, ",")
			accessType := roleFilter.AccessType
			actionType := roleFilter.Action
			subAction := getSubactionFromRoleFilter(roleFilter)
			subActions := strings.Split(subAction, ",")
			for _, namespace := range namespaces {
				for _, group := range groups {
					for _, kind := range kinds {
						for _, resource := range resources {
							for _, subaction := range subActions {
								roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(adapter.BuildClusterRoleFieldsDto(roleFilter.Entity, accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, subaction))
								if err != nil {
									impl.logger.Errorw("Error in fetching roles by filter", "roleFilter", roleFilter)
									return nil, nil, err
								}
								if roleModel.Id == 0 {
									impl.logger.Warnw("no role found for given filter", "filter", roleFilter)
									continue
								}
								if _, ok := existingRoleIds[roleModel.Id]; ok {
									delete(eliminatedRoleIds, roleModel.Id)
								}
							}
						}
					}
				}
			}
		} else if roleFilter.Entity == bean2.EntityJobs {
			entityNames := strings.Split(roleFilter.EntityName, ",")
			environments := strings.Split(roleFilter.Environment, ",")
			workflows := strings.Split(roleFilter.Workflow, ",")
			actionType := roleFilter.Action
			accessType := roleFilter.AccessType
			subAction := getSubactionFromRoleFilter(roleFilter)
			subActions := strings.Split(subAction, ",")
			for _, environment := range environments {
				for _, entityName := range entityNames {
					for _, workflow := range workflows {
						for _, subaction := range subActions {
							roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(adapter.BuildJobsRoleFieldsDto(roleFilter.Entity, roleFilter.Team, entityName, environment, actionType, accessType, workflow, subaction))
							if err != nil {
								impl.logger.Errorw("Error in fetching roles by filter", "user", userInfo)
								return nil, nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
								continue
							}
							if _, ok := existingRoleIds[roleModel.Id]; ok {
								delete(eliminatedRoleIds, roleModel.Id)
							}
						}
					}
				}
			}
		} else {
			entityNames := strings.Split(roleFilter.EntityName, ",")
			environments := strings.Split(roleFilter.Environment, ",")
			actionType := roleFilter.Action
			accessType := roleFilter.AccessType
			subAction := getSubactionFromRoleFilter(roleFilter)
			subActions := strings.Split(subAction, ",")
			for _, environment := range environments {
				for _, entityName := range entityNames {
					for _, subaction := range subActions {
						roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(adapter.BuildOtherRoleFieldsDto(roleFilter.Entity, roleFilter.Team, entityName, environment, actionType, accessType, false, subaction, false))
						if err != nil {
							impl.logger.Errorw("Error in fetching roles by filter", "user", userInfo)
							return nil, nil, err
						}
						oldRoleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(adapter.BuildOtherRoleFieldsDto(roleFilter.Entity, roleFilter.Team, entityName, environment, actionType, accessType, true, subaction, false))
						if err != nil {
							return nil, nil, err
						}
						if roleModel.Id == 0 {
							impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
							continue
						}
						if _, ok := existingRoleIds[roleModel.Id]; ok {
							delete(eliminatedRoleIds, roleModel.Id)
						}
						isChartGroupEntity := roleFilter.Entity == bean2.CHART_GROUP_ENTITY
						if _, ok := existingRoleIds[oldRoleModel.Id]; ok && !isChartGroupEntity {
							//delete old role mapping from existing but not from eliminated roles (so that it gets deleted)
							delete(existingRoleIds, oldRoleModel.Id)
						}
					}
				}
			}
		}
	}

	// delete remaining Ids from casbin role mapping table in orchestrator and casbin policy db
	// which are existing but not provided in this request
	eliminatedRoles := make([]*repository.RoleModel, 0, len(eliminatedRoleIds))
	roleIdVsRoleMap, err := impl.getMapOfRoleIdVsRoleForRoleIds(maps.Keys(eliminatedRoleIds))
	if err != nil {
		impl.logger.Errorw("error encountered in RemoveRolesAndReturnEliminatedPolicies", "eliminatedRoleIds", eliminatedRoleIds, "err", err)
		return nil, nil, err
	}
	toBeDeletedUserRolesIds := make([]int, 0, len(eliminatedRoleIds))
	for _, userRoleModel := range eliminatedRoleIds {
		if role, ok := roleIdVsRoleMap[userRoleModel.RoleId]; ok {
			isValidAuth := impl.checkRbacForARole(role, token, managerAuth)
			if !isValidAuth {
				continue
			}
			toBeDeletedUserRolesIds = append(toBeDeletedUserRolesIds, userRoleModel.Id)
			eliminatedRoles = append(eliminatedRoles, role)
			eliminatedPolicies = append(eliminatedPolicies, bean3.Policy{Type: "g", Sub: bean3.Subject(userInfo.EmailId), Obj: bean3.Object(role.Role)})
		}
	}

	if len(toBeDeletedUserRolesIds) > 0 {
		err = impl.userAuthRepository.DeleteUserRoleMappingByIds(toBeDeletedUserRolesIds, tx)
		if err != nil {
			impl.logger.Errorw("error encountered in RemoveRolesAndReturnEliminatedPolicies", "toBeDeletedUserRolesIds", toBeDeletedUserRolesIds, "err", err)
			return nil, nil, err
		}
	}
	// DELETE ENDS
	return eliminatedPolicies, eliminatedRoles, nil
}

func (impl UserCommonServiceImpl) getMapOfRoleIdVsRoleForRoleIds(roleIds []int) (map[int]*repository.RoleModel, error) {
	roleIdVsRoleMap := make(map[int]*repository.RoleModel, len(roleIds))
	if len(roleIds) > 0 {
		roles, err := impl.userAuthRepository.GetRolesByIds(roleIds)
		if err != nil {
			impl.logger.Errorw("Error in fetching roles by ids", "roleIds", roleIds)
			return nil, err
		}
		for _, role := range roles {
			roleIdVsRoleMap[role.Id] = role
		}
	}
	return roleIdVsRoleMap, nil
}

func (impl UserCommonServiceImpl) RemoveRolesAndReturnEliminatedPoliciesForGroups(request *bean2.RoleGroup, existingRoles map[int]*repository.RoleGroupRoleMapping, eliminatedRoles map[int]*repository.RoleGroupRoleMapping, tx *pg.Tx, token string, managerAuth func(resource string, token string, object string) bool) ([]bean3.Policy, []*repository.RoleModel, error) {
	// Filter out removed items in current request
	//var policies []casbin.Policy
	for _, roleFilter := range request.RoleFilters {
		entity := roleFilter.Entity
		if entity == bean2.CLUSTER_ENTITIY {
			namespaces := strings.Split(roleFilter.Namespace, ",")
			groups := strings.Split(roleFilter.Group, ",")
			kinds := strings.Split(roleFilter.Kind, ",")
			resources := strings.Split(roleFilter.Resource, ",")
			actionType := roleFilter.Action
			accessType := roleFilter.AccessType
			subAction := getSubactionFromRoleFilter(roleFilter)
			subActions := strings.Split(subAction, ",")
			for _, namespace := range namespaces {
				for _, group := range groups {
					for _, kind := range kinds {
						for _, resource := range resources {
							for _, subaction := range subActions {
								roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(adapter.BuildClusterRoleFieldsDto(entity, accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, subaction))
								if err != nil {
									impl.logger.Errorw("Error in fetching roles by filter", "user", request)
									return nil, nil, err
								}
								if roleModel.Id == 0 {
									impl.logger.Warnw("no role found for given filter", "filter", roleFilter)
									continue
								}
								if _, ok := existingRoles[roleModel.Id]; ok {
									delete(eliminatedRoles, roleModel.Id)
								}
							}
						}
					}
				}
			}
		} else if entity == bean2.EntityJobs {
			entityNames := strings.Split(roleFilter.EntityName, ",")
			environments := strings.Split(roleFilter.Environment, ",")
			workflows := strings.Split(roleFilter.Workflow, ",")
			accessType := roleFilter.AccessType
			actionType := roleFilter.Action
			subAction := getSubactionFromRoleFilter(roleFilter)
			subActions := strings.Split(subAction, ",")
			for _, environment := range environments {
				for _, entityName := range entityNames {
					for _, workflow := range workflows {
						for _, subaction := range subActions {
							roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(adapter.BuildJobsRoleFieldsDto(entity, roleFilter.Team, entityName, environment, actionType, accessType, workflow, subaction))
							if err != nil {
								impl.logger.Errorw("Error in fetching roles by filter", "user", request)
								return nil, nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Warnw("no role found for given filter", "filter", roleFilter)
								continue
							}
							if _, ok := existingRoles[roleModel.Id]; ok {
								delete(eliminatedRoles, roleModel.Id)
							}
						}
					}
				}
			}
		} else {
			entityNames := strings.Split(roleFilter.EntityName, ",")
			environments := strings.Split(roleFilter.Environment, ",")
			accessType := roleFilter.AccessType
			actionType := roleFilter.Action
			subAction := getSubactionFromRoleFilter(roleFilter)
			subActions := strings.Split(subAction, ",")
			for _, environment := range environments {
				for _, entityName := range entityNames {
					for _, subaction := range subActions {
						roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(adapter.BuildOtherRoleFieldsDto(roleFilter.Entity, roleFilter.Team, entityName, environment, actionType, accessType, false, subaction, false))
						if err != nil {
							impl.logger.Errorw("Error in fetching roles by filter", "user", request)
							return nil, nil, err
						}
						oldRoleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(adapter.BuildOtherRoleFieldsDto(roleFilter.Entity, roleFilter.Team, entityName, environment, actionType, accessType, true, subaction, false))
						if err != nil {
							impl.logger.Errorw("Error in fetching roles by filter by old values", "user", request)
							return nil, nil, err
						}
						if roleModel.Id == 0 && oldRoleModel.Id == 0 {
							impl.logger.Warnw("no role found for given filter", "filter", roleFilter)
							continue
						}
						if _, ok := existingRoles[roleModel.Id]; ok {
							delete(eliminatedRoles, roleModel.Id)
						}
						isChartGroupEntity := roleFilter.Entity == bean2.CHART_GROUP_ENTITY
						if _, ok := existingRoles[oldRoleModel.Id]; ok && !isChartGroupEntity {
							//delete old role mapping from existing but not from eliminated roles (so that it gets deleted)
							delete(existingRoles, oldRoleModel.Id)
						}
					}
				}
			}
		}
	}

	//delete remaining Ids from casbin role mapping table in orchestrator and casbin policy db
	// which are existing but not provided in this request
	var eliminatedPolicies []bean3.Policy
	eliminatedRoleModels := make([]*repository.RoleModel, 0, len(eliminatedRoles))
	toBeDeletedRoleGroupRoleMappingsIds := make([]int, 0, len(eliminatedRoles))
	roleIdVsRoleMap, err := impl.getMapOfRoleIdVsRoleForRoleIds(maps.Keys(eliminatedRoles))
	if err != nil {
		impl.logger.Errorw("error encountered in RemoveRolesAndReturnEliminatedPoliciesForGroups", "eliminatedRoles", eliminatedRoles, "err", err)
		return nil, nil, err
	}
	policyGroup, err := impl.roleGroupRepository.GetRoleGroupById(request.Id)
	if err != nil {
		impl.logger.Errorw("error encountered in RemoveRolesAndReturnEliminatedPoliciesForGroups", "roleGroupId", request.Id, "err", err)
		return nil, nil, err
	}
	for _, model := range eliminatedRoles {
		if role, ok := roleIdVsRoleMap[model.RoleId]; ok {
			isValidAuth := impl.checkRbacForARole(role, token, managerAuth)
			if !isValidAuth {
				continue
			}
			toBeDeletedRoleGroupRoleMappingsIds = append(toBeDeletedRoleGroupRoleMappingsIds, model.Id)
			eliminatedRoleModels = append(eliminatedRoleModels, role)
			eliminatedPolicies = append(eliminatedPolicies, bean3.Policy{Type: "g", Sub: bean3.Subject(policyGroup.CasbinName), Obj: bean3.Object(role.Role)})
		}
	}
	if len(toBeDeletedRoleGroupRoleMappingsIds) > 0 {
		err = impl.roleGroupRepository.DeleteRoleGroupRoleMappingsByIds(tx, toBeDeletedRoleGroupRoleMappingsIds)
		if err != nil {
			impl.logger.Errorw("error encountered in RemoveRolesAndReturnEliminatedPoliciesForGroups", "toBeDeletedRoleGroupRoleMappingsIds", toBeDeletedRoleGroupRoleMappingsIds, "err", err)
			return nil, nil, err
		}
	}
	return eliminatedPolicies, eliminatedRoleModels, nil
}

func (impl UserCommonServiceImpl) checkRbacForARole(role *repository.RoleModel, token string, managerAuth func(resource string, token string, object string) bool) bool {
	isAuthorised := true
	switch {
	case role.Action == bean2.SUPER_ADMIN || role.AccessType == bean2.APP_ACCESS_TYPE_HELM || role.Entity == bean2.EntityJobs:
		isValidAuth := managerAuth(casbin.ResourceGlobal, token, "*")
		if !isValidAuth {
			isAuthorised = false
		}

	case len(role.Team) > 0:
		// this is case of devtron app
		rbacObject := fmt.Sprintf("%s", role.Team)
		isValidAuth := managerAuth(casbin.ResourceUser, token, rbacObject)
		if !isValidAuth {
			isAuthorised = false
		}

	case role.Entity == bean2.CLUSTER_ENTITIY:
		isValidAuth := impl.CheckRbacForClusterEntity(role.Cluster, role.Namespace, role.Group, role.Kind, role.Resource, token, managerAuth)
		if !isValidAuth {
			isAuthorised = false
		}
	case role.Entity == bean2.CHART_GROUP_ENTITY:
		isAuthorised = true
	default:
		isAuthorised = false
	}
	return isAuthorised
}

func containsArr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (impl UserCommonServiceImpl) CheckRbacForClusterEntity(cluster, namespace, group, kind, resource, token string, managerAuth func(resource, token, object string) bool) bool {
	if namespace == "NONE" {
		namespace = ""
	}
	if group == "NONE" {
		group = ""
	}
	if kind == "NONE" {
		kind = ""
	}
	if resource == "NONE" {
		resource = ""
	}
	namespaceObj := namespace
	groupObj := group
	kindObj := kind
	resourceObj := resource
	if namespace == "" {
		namespaceObj = "*"
	}
	if group == "" {
		groupObj = "*"
	}
	if kind == "" {
		kindObj = "*"
	}
	if resource == "" {
		resourceObj = "*"
	}

	resourcesArray := strings.Split(resourceObj, ",")
	namespacesArray := strings.Split(namespaceObj, ",")
	for _, namespaceInArray := range namespacesArray {
		rbacResource := fmt.Sprintf("%s/%s/%s", strings.ToLower(cluster), strings.ToLower(namespaceInArray), casbin.ResourceUser)
		for _, resourceVal := range resourcesArray {
			rbacObject := fmt.Sprintf("%s/%s/%s", groupObj, kindObj, resourceVal)
			allowed := managerAuth(rbacResource, token, rbacObject)
			if !allowed {
				return false
			}
		}
	}
	return true
}

func (impl UserCommonServiceImpl) GetCapacityForRoleFilter(roleFilters []bean2.RoleFilter) (int, map[int]int) {
	capacity := 0

	m := make(map[int]int)
	for index, roleFilter := range roleFilters {
		namespaces := strings.Split(roleFilter.Namespace, ",")
		groups := strings.Split(roleFilter.Group, ",")
		kinds := strings.Split(roleFilter.Kind, ",")
		resources := strings.Split(roleFilter.Resource, ",")
		entityNames := strings.Split(roleFilter.EntityName, ",")
		environments := strings.Split(roleFilter.Environment, ",")
		workflows := strings.Split(roleFilter.Workflow, ",")
		value := math.Max(float64(len(namespaces)*len(groups)*len(kinds)*len(resources)*2), math.Max(float64(len(entityNames)*len(environments)*6), float64(len(entityNames)*len(environments)*len(workflows)*8)))
		m[index] = int(value)
		capacity += int(value)
	}
	return capacity, m
}

func (impl UserCommonServiceImpl) BuildRoleFilterForAllTypes(roleFilterMap map[string]*bean2.RoleFilter, entityProcessor EntityKeyProcessor, key string, basetoConsider bean2.MergingBaseKey) {
	switch entityProcessor.GetEntity() {
	case bean2.CLUSTER_ENTITIY:
		{
			BuildRoleFilterKeyForCluster(roleFilterMap, entityProcessor, key)
		}
	case bean2.EntityJobs:
		{
			BuildRoleFilterKeyForJobs(roleFilterMap, entityProcessor, key, basetoConsider)
		}
	default:
		{
			BuildRoleFilterKeyForOtherEntity(roleFilterMap, entityProcessor, key, basetoConsider)
		}
	}
}

func BuildRoleFilterKeyForCluster(roleFilterMap map[string]*bean2.RoleFilter, entityProcessor EntityKeyProcessor, key string) {
	namespaceArr := strings.Split(roleFilterMap[key].Namespace, ",")
	if containsArr(namespaceArr, AllNamespace) {
		roleFilterMap[key].Namespace = AllNamespace
	} else if !containsArr(namespaceArr, entityProcessor.GetNamespace()) {
		roleFilterMap[key].Namespace = fmt.Sprintf("%s,%s", roleFilterMap[key].Namespace, entityProcessor.GetNamespace())
	}
	groupArr := strings.Split(roleFilterMap[key].Group, ",")
	if containsArr(groupArr, AllGroup) {
		roleFilterMap[key].Group = AllGroup
	} else if !containsArr(groupArr, entityProcessor.GetGroup()) {
		roleFilterMap[key].Group = fmt.Sprintf("%s,%s", roleFilterMap[key].Group, entityProcessor.GetGroup())
	}
	kindArr := strings.Split(roleFilterMap[key].Kind, ",")
	if containsArr(kindArr, AllKind) {
		roleFilterMap[key].Kind = AllKind
	} else if !containsArr(kindArr, entityProcessor.GetKind()) {
		roleFilterMap[key].Kind = fmt.Sprintf("%s,%s", roleFilterMap[key].Kind, entityProcessor.GetKind())
	}
	resourceArr := strings.Split(roleFilterMap[key].Resource, ",")
	if containsArr(resourceArr, AllResource) {
		roleFilterMap[key].Resource = AllResource
	} else if !containsArr(resourceArr, entityProcessor.GetResource()) {
		roleFilterMap[key].Resource = fmt.Sprintf("%s,%s", roleFilterMap[key].Resource, entityProcessor.GetResource())
	}
}

func BuildRoleFilterKeyForJobs(roleFilterMap map[string]*bean2.RoleFilter, entityProcessor EntityKeyProcessor, key string, baseToConsider bean2.MergingBaseKey) {
	switch baseToConsider {
	case bean2.ApplicationBasedKey:
		envArr := strings.Split(roleFilterMap[key].Environment, ",")
		if containsArr(envArr, AllEnvironment) {
			roleFilterMap[key].Environment = AllEnvironment
		} else if !containsArr(envArr, entityProcessor.GetEnvironment()) {
			roleFilterMap[key].Environment = fmt.Sprintf("%s,%s", roleFilterMap[key].Environment, entityProcessor.GetEnvironment())
		}
	case bean2.EnvironmentBasedKey:
		entityArr := strings.Split(roleFilterMap[key].EntityName, ",")
		if containsArr(entityArr, bean2.EmptyStringIndicatingAll) {
			roleFilterMap[key].EntityName = bean2.EmptyStringIndicatingAll
		} else if !containsArr(entityArr, entityProcessor.GetEntityName()) {
			roleFilterMap[key].EntityName = fmt.Sprintf("%s,%s", roleFilterMap[key].EntityName, entityProcessor.GetEntityName())
		}
	default:
		envArr := strings.Split(roleFilterMap[key].Environment, ",")
		if containsArr(envArr, AllEnvironment) {
			roleFilterMap[key].Environment = AllEnvironment
		} else if !containsArr(envArr, entityProcessor.GetEnvironment()) {
			roleFilterMap[key].Environment = fmt.Sprintf("%s,%s", roleFilterMap[key].Environment, entityProcessor.GetEnvironment())
		}
		entityArr := strings.Split(roleFilterMap[key].EntityName, ",")
		if containsArr(entityArr, bean2.EmptyStringIndicatingAll) {
			roleFilterMap[key].EntityName = bean2.EmptyStringIndicatingAll
		} else if !containsArr(entityArr, entityProcessor.GetEntityName()) {
			roleFilterMap[key].EntityName = fmt.Sprintf("%s,%s", roleFilterMap[key].EntityName, entityProcessor.GetEntityName())
		}
	}
	workflowArr := strings.Split(roleFilterMap[key].Workflow, ",")
	if containsArr(workflowArr, AllWorkflow) {
		roleFilterMap[key].Workflow = AllWorkflow
	} else if !containsArr(workflowArr, entityProcessor.GetWorkflow()) {
		roleFilterMap[key].Workflow = fmt.Sprintf("%s,%s", roleFilterMap[key].Workflow, entityProcessor.GetWorkflow())
	}
}

func BuildRoleFilterKeyForOtherEntity(roleFilterMap map[string]*bean2.RoleFilter, entityProcessor EntityKeyProcessor, key string, baseToConsider bean2.MergingBaseKey) {
	switch baseToConsider {
	case bean2.ApplicationBasedKey:
		envArr := strings.Split(roleFilterMap[key].Environment, ",")
		if containsArr(envArr, AllEnvironment) {
			roleFilterMap[key].Environment = AllEnvironment
		} else if !containsArr(envArr, entityProcessor.GetEnvironment()) {
			roleFilterMap[key].Environment = fmt.Sprintf("%s,%s", roleFilterMap[key].Environment, entityProcessor.GetEnvironment())
		}
	case bean2.EnvironmentBasedKey:
		entityArr := strings.Split(roleFilterMap[key].EntityName, ",")
		if containsArr(entityArr, bean2.EmptyStringIndicatingAll) {
			roleFilterMap[key].EntityName = bean2.EmptyStringIndicatingAll
		} else if !containsArr(entityArr, entityProcessor.GetEntityName()) {
			roleFilterMap[key].EntityName = fmt.Sprintf("%s,%s", roleFilterMap[key].EntityName, entityProcessor.GetEntityName())
		}
	default:
		envArr := strings.Split(roleFilterMap[key].Environment, ",")
		if containsArr(envArr, AllEnvironment) {
			roleFilterMap[key].Environment = AllEnvironment
		} else if !containsArr(envArr, entityProcessor.GetEnvironment()) {
			roleFilterMap[key].Environment = fmt.Sprintf("%s,%s", roleFilterMap[key].Environment, entityProcessor.GetEnvironment())
		}
		entityArr := strings.Split(roleFilterMap[key].EntityName, ",")
		if containsArr(entityArr, bean2.EmptyStringIndicatingAll) {
			roleFilterMap[key].EntityName = bean2.EmptyStringIndicatingAll
		} else if !containsArr(entityArr, entityProcessor.GetEntityName()) {
			roleFilterMap[key].EntityName = fmt.Sprintf("%s,%s", roleFilterMap[key].EntityName, entityProcessor.GetEntityName())
		}
	}
}
func (impl UserCommonServiceImpl) GetUniqueKeyForAllEntity(entityProcessor EntityKeyProcessor, baseToConsider bean2.MergingBaseKey) string {
	key := ""
	if len(entityProcessor.GetTeam()) > 0 && entityProcessor.GetEntity() != bean2.EntityJobs {
		switch baseToConsider {
		case bean2.EnvironmentBasedKey:
			key = fmt.Sprintf("%s_%s_%s_%s", entityProcessor.GetTeam(), entityProcessor.GetEnvironment(), entityProcessor.GetAction(), entityProcessor.GetAccessType())
		case bean2.ApplicationBasedKey:
			key = fmt.Sprintf("%s_%s_%s_%s", entityProcessor.GetTeam(), entityProcessor.GetEntityName(), entityProcessor.GetAction(), entityProcessor.GetAccessType())
		default:
			key = fmt.Sprintf("%s_%s_%s", entityProcessor.GetTeam(), entityProcessor.GetAction(), entityProcessor.GetAccessType())
		}
	} else if entityProcessor.GetEntity() == bean2.EntityJobs {
		switch baseToConsider {
		case bean2.EnvironmentBasedKey:
			key = fmt.Sprintf("%s_%s_%s_%s_%s", entityProcessor.GetTeam(), entityProcessor.GetEnvironment(), entityProcessor.GetAction(), entityProcessor.GetAccessType(), entityProcessor.GetEntity())
		case bean2.ApplicationBasedKey:
			key = fmt.Sprintf("%s_%s_%s_%s_%s", entityProcessor.GetTeam(), entityProcessor.GetEntityName(), entityProcessor.GetAction(), entityProcessor.GetAccessType(), entityProcessor.GetEntity())
		default:
			key = fmt.Sprintf("%s_%s_%s_%s", entityProcessor.GetTeam(), entityProcessor.GetAction(), entityProcessor.GetAccessType(), entityProcessor.GetEntity())
		}
	} else if len(entityProcessor.GetEntity()) > 0 {
		if entityProcessor.GetEntity() == bean2.CLUSTER_ENTITIY {
			key = fmt.Sprintf("%s_%s_%s_%s_%s", entityProcessor.GetEntity(), entityProcessor.GetAction(), entityProcessor.GetCluster(),
				entityProcessor.GetGroup(), entityProcessor.GetKind())
		} else {
			key = fmt.Sprintf("%s_%s", entityProcessor.GetEntity(), entityProcessor.GetAction())
		}
	}
	return key
}

func (impl UserCommonServiceImpl) SetDefaultValuesIfNotPresent(request *bean2.ListingRequest, isRoleGroup bool) {
	if len(request.SortBy) == 0 {
		if isRoleGroup {
			request.SortBy = bean2.GroupName
		} else {
			request.SortBy = bean2.Email
		}
	}
	if request.Size == 0 {
		request.Size = bean2.DefaultSize
	}
}

func (impl UserCommonServiceImpl) DeleteRoleForUserFromCasbin(mappings map[string][]string) bool {
	successful := true
	for v0, v1s := range mappings {
		for _, v1 := range v1s {
			flag := casbin.DeleteRoleForUser(v0, v1)
			if flag == false {
				impl.logger.Warnw("unable to delete role:", "v0", v0, "v1", v1)
				successful = false
				return successful
			}
		}
	}
	return successful
}

func (impl UserCommonServiceImpl) DeleteUserForRoleFromCasbin(mappings map[string][]string) bool {
	successful := true
	for v1, v0s := range mappings {
		for _, v0 := range v0s {
			flag := casbin.DeleteRoleForUser(v0, v1)
			if flag == false {
				impl.logger.Warnw("unable to delete role:", "v0", v0, "v1", v1)
				successful = false
				return successful
			}
		}
	}
	return successful
}

func (impl UserCommonServiceImpl) BuildRoleFiltersAfterMerging(entityProcessors []EntityKeyProcessor, baseKey bean2.MergingBaseKey) []bean2.RoleFilter {
	roleFilterMap := make(map[string]*bean2.RoleFilter, len(entityProcessors))
	roleFilters := make([]bean2.RoleFilter, 0, len(entityProcessors))
	for _, entityProcessor := range entityProcessors {
		key := impl.GetUniqueKeyForAllEntity(entityProcessor, baseKey)
		if _, ok := roleFilterMap[key]; ok {
			impl.BuildRoleFilterForAllTypes(roleFilterMap, entityProcessor, key, baseKey)
		} else {
			roleFilterMap[key] = &bean2.RoleFilter{
				Entity:      entityProcessor.GetEntity(),
				Team:        entityProcessor.GetTeam(),
				Environment: entityProcessor.GetEnvironment(),
				EntityName:  entityProcessor.GetEntityName(),
				Action:      entityProcessor.GetAction(),
				AccessType:  entityProcessor.GetAccessType(),
				Cluster:     entityProcessor.GetCluster(),
				Namespace:   entityProcessor.GetNamespace(),
				Group:       entityProcessor.GetGroup(),
				Kind:        entityProcessor.GetKind(),
				Resource:    entityProcessor.GetResource(),
				Workflow:    entityProcessor.GetWorkflow(),
			}
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
	return roleFilters
}

func ConvertRoleFiltersToEntityProcessors(filters []bean2.RoleFilter) []EntityKeyProcessor {
	processors := make([]EntityKeyProcessor, len(filters))
	for i, filter := range filters {
		processors[i] = filter // Each RoleFilter is an EntityKeyProcessor
	}
	return processors
}

func ConvertRolesToEntityProcessors(roles []*repository.RoleModel) []EntityKeyProcessor {
	processors := make([]EntityKeyProcessor, len(roles))
	for i, filter := range roles {
		processors[i] = filter // Each role is an EntityKeyProcessor
	}
	return processors
}
