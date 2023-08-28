package user

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	bean2 "github.com/devtron-labs/devtron/pkg/user/bean"
	casbin "github.com/devtron-labs/devtron/pkg/user/casbin"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"math"
	"strings"
	"time"
)

type UserCommonService interface {
	CreateDefaultPoliciesForAllTypes(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType string, userId int32) (bool, error, []casbin.Policy)
	RemoveRolesAndReturnEliminatedPolicies(userInfo *bean.UserInfo, existingRoleIds map[int]repository2.UserRoleModel, eliminatedRoleIds map[int]*repository2.UserRoleModel, tx *pg.Tx, token string, managerAuth func(resource, token, object string) bool) ([]casbin.Policy, error)
	RemoveRolesAndReturnEliminatedPoliciesForGroups(request *bean.RoleGroup, existingRoles map[int]*repository2.RoleGroupRoleMapping, eliminatedRoles map[int]*repository2.RoleGroupRoleMapping, tx *pg.Tx, token string, managerAuth func(resource string, token string, object string) bool) ([]casbin.Policy, error)
	CheckRbacForClusterEntity(cluster, namespace, group, kind, resource, token string, managerAuth func(resource, token, object string) bool) bool
	ReplacePlaceHolderForEmptyEntriesInRoleFilter(roleFilter bean.RoleFilter) bean.RoleFilter
	RemovePlaceHolderInRoleFilterField(roleFilterField string) string
	GetCapacityForRoleFilter(roleFilters []bean.RoleFilter) (int, map[int]int)
}

type UserCommonServiceImpl struct {
	userAuthRepository          repository2.UserAuthRepository
	logger                      *zap.SugaredLogger
	userRepository              repository2.UserRepository
	roleGroupRepository         repository2.RoleGroupRepository
	sessionManager2             *middleware.SessionManager
	defaultRbacDataCacheFactory repository2.RbacDataCacheFactory
	userRbacConfig              *UserRbacConfig
}

func NewUserCommonServiceImpl(userAuthRepository repository2.UserAuthRepository,
	logger *zap.SugaredLogger,
	userRepository repository2.UserRepository,
	userGroupRepository repository2.RoleGroupRepository,
	sessionManager2 *middleware.SessionManager,
	defaultRbacDataCacheFactory repository2.RbacDataCacheFactory) *UserCommonServiceImpl {
	userConfig := &UserRbacConfig{}
	err := env.Parse(userConfig)
	if err != nil {
		logger.Fatal("error occurred while parsing user config", err)
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
	return serviceImpl
}

type UserRbacConfig struct {
	UseRbacCreationV2 bool `env:"USE_RBAC_CREATION_V2" envDefault:"false"`
}

func (impl UserCommonServiceImpl) CreateDefaultPoliciesForAllTypes(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType string, userId int32) (bool, error, []casbin.Policy) {
	if impl.userRbacConfig.UseRbacCreationV2 {
		impl.logger.Debugw("using rbac creation v2 for creating default policies")
		return impl.CreateDefaultPoliciesForAllTypesV2(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType)
	} else {
		return impl.userAuthRepository.CreateDefaultPoliciesForAllTypes(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType, userId)
	}
}

func (impl UserCommonServiceImpl) CreateDefaultPoliciesForAllTypesV2(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType string) (bool, error, []casbin.Policy) {
	//TODO: below txn is making this process slow, need to do bulk operation for role creation.
	//For detail - https://github.com/devtron-labs/devtron/blob/main/pkg/user/benchmarking-results

	renderedRole, renderedPolicyDetails, err := impl.getRenderedRoleAndPolicy(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType)
	if err != nil {
		return false, err, nil
	}
	_, err = impl.userAuthRepository.CreateRole(renderedRole)
	if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
		return false, err, nil
	}
	return true, nil, renderedPolicyDetails
}

func (impl UserCommonServiceImpl) getRenderedRoleAndPolicy(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType string) (*repository2.RoleModel, []casbin.Policy, error) {
	//getting map of values to be used for rendering
	pValUpdateMap := getPValUpdateMap(team, entityName, env, entity, cluster, namespace, group, kind, resource)

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

func (impl UserCommonServiceImpl) getDefaultRbacRoleAndPolicyByRoleFilter(entity, accessType, action string) (repository2.RoleCacheDetailObj, repository2.PolicyCacheDetailObj, error) {
	//getting default role and policy data from cache
	return impl.defaultRbacDataCacheFactory.
		GetDefaultRoleDataAndPolicyByEntityAccessTypeAndRoleType(entity, accessType, action)
}

func getRenderedRoleData(defaultRoleData repository2.RoleCacheDetailObj, pValUpdateMap map[repository2.PValUpdateKey]string) *repository2.RoleModel {
	renderedRoleData := &repository2.RoleModel{
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
		AuditLog: sql.AuditLog{ //not storing user information because this role can be mapped to other users in future and hence can lead to confusion
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
		},
	}
	return renderedRoleData
}

func getRenderedPolicy(defaultPolicy repository2.PolicyCacheDetailObj, pValUpdateMap map[repository2.PValUpdateKey]string) []casbin.Policy {
	renderedPolicies := make([]casbin.Policy, 0, len(defaultPolicy.ResActObjSet))
	policyType := getResolvedValueFromPValDetailObject(defaultPolicy.Type, pValUpdateMap)
	policySub := getResolvedValueFromPValDetailObject(defaultPolicy.Sub, pValUpdateMap)
	for _, v := range defaultPolicy.ResActObjSet {
		policyRes := getResolvedValueFromPValDetailObject(v.Res, pValUpdateMap)
		policyAct := getResolvedValueFromPValDetailObject(v.Act, pValUpdateMap)
		policyObj := getResolvedValueFromPValDetailObject(v.Obj, pValUpdateMap)
		renderedPolicy := casbin.Policy{
			Type: casbin.PolicyType(policyType),
			Sub:  casbin.Subject(policySub),
			Res:  casbin.Resource(policyRes),
			Act:  casbin.Action(policyAct),
			Obj:  casbin.Object(policyObj),
		}
		renderedPolicies = append(renderedPolicies, renderedPolicy)
	}
	return renderedPolicies
}

func getResolvedValueFromPValDetailObject(pValDetailObj repository2.PValDetailObj, pValUpdateMap map[repository2.PValUpdateKey]string) string {
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

func getPValUpdateMap(team, entityName, env, entity, cluster,
	namespace, group, kind, resource string) map[repository2.PValUpdateKey]string {
	pValUpdateMap := make(map[repository2.PValUpdateKey]string)
	pValUpdateMap[repository2.EntityPValUpdateKey] = entity
	if entity == bean.CLUSTER_ENTITIY {
		pValUpdateMap[repository2.ClusterPValUpdateKey] = cluster
		pValUpdateMap[repository2.NamespacePValUpdateKey] = namespace
		pValUpdateMap[repository2.GroupPValUpdateKey] = group
		pValUpdateMap[repository2.KindPValUpdateKey] = kind
		pValUpdateMap[repository2.ResourcePValUpdateKey] = resource
		pValUpdateMap[repository2.ClusterObjPValUpdateKey] = getResolvedPValMapValue(cluster)
		pValUpdateMap[repository2.NamespaceObjPValUpdateKey] = getResolvedPValMapValue(namespace)
		pValUpdateMap[repository2.GroupObjPValUpdateKey] = getResolvedPValMapValue(group)
		pValUpdateMap[repository2.KindObjPValUpdateKey] = getResolvedPValMapValue(kind)
		pValUpdateMap[repository2.ResourceObjPValUpdateKey] = getResolvedPValMapValue(resource)
	} else {
		pValUpdateMap[repository2.EntityNamePValUpdateKey] = entityName
		pValUpdateMap[repository2.TeamPValUpdateKey] = team
		pValUpdateMap[repository2.AppPValUpdateKey] = entityName
		pValUpdateMap[repository2.EnvPValUpdateKey] = env
		pValUpdateMap[repository2.TeamObjPValUpdateKey] = getResolvedPValMapValue(team)
		pValUpdateMap[repository2.AppObjPValUpdateKey] = getResolvedPValMapValue(entityName)
		pValUpdateMap[repository2.EnvObjPValUpdateKey] = getResolvedPValMapValue(env)
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

func (impl UserCommonServiceImpl) RemoveRolesAndReturnEliminatedPolicies(userInfo *bean.UserInfo,
	existingRoleIds map[int]repository2.UserRoleModel, eliminatedRoleIds map[int]*repository2.UserRoleModel,
	tx *pg.Tx, token string, managerAuth func(resource, token, object string) bool) ([]casbin.Policy, error) {
	var eliminatedPolicies []casbin.Policy
	// DELETE Removed Items
	for _, roleFilter := range userInfo.RoleFilters {
		if roleFilter.Entity == bean.CLUSTER_ENTITIY {
			if roleFilter.Namespace == "" {
				roleFilter.Namespace = "NONE"
			}
			if roleFilter.Group == "" {
				roleFilter.Group = "NONE"
			}
			if roleFilter.Kind == "" {
				roleFilter.Kind = "NONE"
			}
			if roleFilter.Resource == "" {
				roleFilter.Resource = "NONE"
			}
			namespaces := strings.Split(roleFilter.Namespace, ",")
			groups := strings.Split(roleFilter.Group, ",")
			kinds := strings.Split(roleFilter.Kind, ",")
			resources := strings.Split(roleFilter.Resource, ",")
			accessType := roleFilter.AccessType
			actionType := roleFilter.Action
			for _, namespace := range namespaces {
				for _, group := range groups {
					for _, kind := range kinds {
						for _, resource := range resources {
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
							isValidAuth := impl.CheckRbacForClusterEntity(roleFilter.Cluster, namespace, group, kind, resource, token, managerAuth)
							if !isValidAuth {
								continue
							}
							roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(roleFilter.Entity, "", "", "", "", accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, false)
							if err != nil {
								impl.logger.Errorw("Error in fetching roles by filter", "roleFilter", roleFilter)
								return nil, err
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
		} else {
			if len(roleFilter.Team) > 0 { // check auth only for apps permission, skip for chart group
				rbacObject := fmt.Sprintf("%s", strings.ToLower(roleFilter.Team))
				isValidAuth := managerAuth(casbin.ResourceUser, token, rbacObject)
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
			actionType := roleFilter.Action
			accessType := roleFilter.AccessType
			for _, environment := range environments {
				for _, entityName := range entityNames {
					if entityName == "NONE" {
						entityName = ""
					}
					if environment == "NONE" {
						environment = ""
					}
					roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(roleFilter.Entity, roleFilter.Team, entityName, environment, actionType, accessType, "", "", "", "", "", actionType, false)
					if err != nil {
						impl.logger.Errorw("Error in fetching roles by filter", "user", userInfo)
						return nil, err
					}
					oldRoleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(roleFilter.Entity, roleFilter.Team, entityName, environment, actionType, accessType, "", "", "", "", "", actionType, true)
					if err != nil {
						return nil, err
					}
					if roleModel.Id == 0 {
						impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
						userInfo.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
						continue
					}
					if _, ok := existingRoleIds[roleModel.Id]; ok {
						delete(eliminatedRoleIds, roleModel.Id)
					}
					isChartGroupEntity := roleFilter.Entity == bean.CHART_GROUP_ENTITY
					if _, ok := existingRoleIds[oldRoleModel.Id]; ok && !isChartGroupEntity {
						//delete old role mapping from existing but not from eliminated roles (so that it gets deleted)
						delete(existingRoleIds, oldRoleModel.Id)
					}

				}
			}
		}
	}

	// delete remaining Ids from casbin role mapping table in orchestrator and casbin policy db
	// which are existing but not provided in this request

	for _, userRoleModel := range eliminatedRoleIds {
		role, err := impl.userAuthRepository.GetRoleById(userRoleModel.RoleId)
		if err != nil {
			return nil, err
		}
		if len(role.Team) > 0 {
			rbacObject := fmt.Sprintf("%s", strings.ToLower(role.Team))
			isValidAuth := managerAuth(casbin.ResourceUser, token, rbacObject)
			if !isValidAuth {
				continue
			}
		}
		if role.Entity == bean.CLUSTER_ENTITIY {
			isValidAuth := impl.CheckRbacForClusterEntity(role.Cluster, role.Namespace, role.Group, role.Kind, role.Resource, token, managerAuth)
			if !isValidAuth {
				continue
			}
		}
		_, err = impl.userAuthRepository.DeleteUserRoleMapping(userRoleModel, tx)
		if err != nil {
			impl.logger.Errorw("Error in delete user role mapping", "user", userInfo)
			return nil, err
		}
		eliminatedPolicies = append(eliminatedPolicies, casbin.Policy{Type: "g", Sub: casbin.Subject(userInfo.EmailId), Obj: casbin.Object(role.Role)})
	}
	// DELETE ENDS
	return eliminatedPolicies, nil
}

func (impl UserCommonServiceImpl) RemoveRolesAndReturnEliminatedPoliciesForGroups(request *bean.RoleGroup, existingRoles map[int]*repository2.RoleGroupRoleMapping, eliminatedRoles map[int]*repository2.RoleGroupRoleMapping, tx *pg.Tx, token string, managerAuth func(resource string, token string, object string) bool) ([]casbin.Policy, error) {
	// Filter out removed items in current request
	//var policies []casbin.Policy
	for _, roleFilter := range request.RoleFilters {
		if roleFilter.Entity == bean.CLUSTER_ENTITIY {
			if roleFilter.Namespace == "" {
				roleFilter.Namespace = "NONE"
			}
			if roleFilter.Group == "" {
				roleFilter.Group = "NONE"
			}
			if roleFilter.Kind == "" {
				roleFilter.Kind = "NONE"
			}
			if roleFilter.Resource == "" {
				roleFilter.Resource = "NONE"
			}
			namespaces := strings.Split(roleFilter.Namespace, ",")
			groups := strings.Split(roleFilter.Group, ",")
			kinds := strings.Split(roleFilter.Kind, ",")
			resources := strings.Split(roleFilter.Resource, ",")
			entity := roleFilter.Entity
			actionType := roleFilter.Action
			accessType := roleFilter.AccessType
			for _, namespace := range namespaces {
				for _, group := range groups {
					for _, kind := range kinds {
						for _, resource := range resources {
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
							isValidAuth := impl.CheckRbacForClusterEntity(roleFilter.Cluster, namespace, group, kind, resource, token, managerAuth)
							if !isValidAuth {
								continue
							}
							roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, "", "", "", "", accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, false)
							if err != nil {
								impl.logger.Errorw("Error in fetching roles by filter", "user", request)
								return nil, err
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
			if len(roleFilter.Team) > 0 { // check auth only for apps permission, skip for chart group
				rbacObject := fmt.Sprintf("%s", strings.ToLower(roleFilter.Team))
				isValidAuth := managerAuth(casbin.ResourceUser, token, rbacObject)
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
			accessType := roleFilter.AccessType
			actionType := roleFilter.Action
			for _, environment := range environments {
				for _, entityName := range entityNames {
					if entityName == "NONE" {
						entityName = ""
					}
					if environment == "NONE" {
						environment = ""
					}
					roleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(roleFilter.Entity, roleFilter.Team, entityName, environment, actionType, accessType, "", "", "", "", "", "", false)
					if err != nil {
						impl.logger.Errorw("Error in fetching roles by filter", "user", request)
						return nil, err
					}
					oldRoleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(roleFilter.Entity, roleFilter.Team, entityName, environment, actionType, accessType, "", "", "", "", "", "", true)
					if err != nil {
						impl.logger.Errorw("Error in fetching roles by filter by old values", "user", request)
						return nil, err
					}
					if roleModel.Id == 0 && oldRoleModel.Id == 0 {
						impl.logger.Warnw("no role found for given filter", "filter", roleFilter)
						request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + actionType
						continue
					}
					if _, ok := existingRoles[roleModel.Id]; ok {
						delete(eliminatedRoles, roleModel.Id)
					}
					isChartGroupEntity := roleFilter.Entity == bean.CHART_GROUP_ENTITY
					if _, ok := existingRoles[oldRoleModel.Id]; ok && !isChartGroupEntity {
						//delete old role mapping from existing but not from eliminated roles (so that it gets deleted)
						delete(existingRoles, oldRoleModel.Id)
					}

				}
			}
		}
	}

	//delete remaining Ids from casbin role mapping table in orchestrator and casbin policy db
	// which are existing but not provided in this request
	var eliminatedPolicies []casbin.Policy
	for _, model := range eliminatedRoles {
		role, err := impl.userAuthRepository.GetRoleById(model.RoleId)
		if err != nil {
			return nil, err
		}
		if len(role.Team) > 0 {
			rbacObject := fmt.Sprintf("%s", strings.ToLower(role.Team))
			isValidAuth := managerAuth(casbin.ResourceUser, token, rbacObject)
			if !isValidAuth {
				continue
			}
		}
		if role.Entity == bean.CLUSTER_ENTITIY {
			isValidAuth := impl.CheckRbacForClusterEntity(role.Cluster, role.Namespace, role.Group, role.Kind, role.Resource, token, managerAuth)
			if !isValidAuth {
				continue
			}
		}
		_, err = impl.roleGroupRepository.DeleteRoleGroupRoleMapping(model, tx)
		if err != nil {
			return nil, err
		}
		policyGroup, err := impl.roleGroupRepository.GetRoleGroupById(model.RoleGroupId)
		if err != nil {
			return nil, err
		}
		eliminatedPolicies = append(eliminatedPolicies, casbin.Policy{Type: "g", Sub: casbin.Subject(policyGroup.CasbinName), Obj: casbin.Object(role.Role)})
	}
	return eliminatedPolicies, nil
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

	rbacResource := fmt.Sprintf("%s/%s/%s", strings.ToLower(cluster), strings.ToLower(namespaceObj), casbin.ResourceUser)
	resourcesArray := strings.Split(resourceObj, ",")
	for _, resourceVal := range resourcesArray {
		rbacObject := fmt.Sprintf("%s/%s/%s", strings.ToLower(groupObj), strings.ToLower(kindObj), strings.ToLower(resourceVal))
		allowed := managerAuth(rbacResource, token, rbacObject)
		if !allowed {
			return false
		}
	}
	return true
}

func (impl UserCommonServiceImpl) ReplacePlaceHolderForEmptyEntriesInRoleFilter(roleFilter bean.RoleFilter) bean.RoleFilter {
	if roleFilter.EntityName == "" {
		roleFilter.EntityName = bean2.EMPTY_ROLEFILTER_ENTRY_PLACEHOLDER
	}
	if roleFilter.Environment == "" {
		roleFilter.Environment = bean2.EMPTY_ROLEFILTER_ENTRY_PLACEHOLDER
	}
	if roleFilter.Namespace == "" {
		roleFilter.Namespace = bean2.EMPTY_ROLEFILTER_ENTRY_PLACEHOLDER
	}
	if roleFilter.Group == "" {
		roleFilter.Group = bean2.EMPTY_ROLEFILTER_ENTRY_PLACEHOLDER
	}
	if roleFilter.Kind == "" {
		roleFilter.Kind = bean2.EMPTY_ROLEFILTER_ENTRY_PLACEHOLDER
	}
	if roleFilter.Resource == "" {
		roleFilter.Resource = bean2.EMPTY_ROLEFILTER_ENTRY_PLACEHOLDER
	}
	return roleFilter
}

func (impl UserCommonServiceImpl) RemovePlaceHolderInRoleFilterField(roleFilterField string) string {
	if roleFilterField == bean2.EMPTY_ROLEFILTER_ENTRY_PLACEHOLDER {
		return ""
	}
	return roleFilterField
}
func (impl UserCommonServiceImpl) GetCapacityForRoleFilter(roleFilters []bean.RoleFilter) (int, map[int]int) {
	capacity := 0

	m := make(map[int]int)
	for index, roleFilter := range roleFilters {
		roleFilter = impl.ReplacePlaceHolderForEmptyEntriesInRoleFilter(roleFilter)
		namespaces := strings.Split(roleFilter.Namespace, ",")
		groups := strings.Split(roleFilter.Group, ",")
		kinds := strings.Split(roleFilter.Kind, ",")
		resources := strings.Split(roleFilter.Resource, ",")
		entityNames := strings.Split(roleFilter.EntityName, ",")
		environments := strings.Split(roleFilter.Environment, ",")
		value := math.Max(float64(len(namespaces)*len(groups)*len(kinds)*len(resources)*2), float64(len(entityNames)*len(environments)*6))
		m[index] = int(value)
		capacity += int(value)
	}
	return capacity, m
}
