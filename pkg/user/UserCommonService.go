package user

import (
	"fmt"
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/user/bean"
	casbin2 "github.com/devtron-labs/devtron/pkg/user/casbin"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"math"
	"strings"
)

type UserCommonService interface {
	RemoveRolesAndReturnEliminatedPolicies(userInfo *bean.UserInfo, existingRoleIds map[int]repository2.UserRoleModel, eliminatedRoleIds map[int]*repository2.UserRoleModel, tx *pg.Tx, token string, managerAuth func(resource, token, object string) bool) ([]casbin2.Policy, error)
	RemoveRolesAndReturnEliminatedPoliciesForGroups(request *bean.RoleGroup, existingRoles map[int]*repository2.RoleGroupRoleMapping, eliminatedRoles map[int]*repository2.RoleGroupRoleMapping, tx *pg.Tx, token string, managerAuth func(resource string, token string, object string) bool) ([]casbin2.Policy, error)
	CheckRbacForClusterEntity(cluster, namespace, group, kind, resource, token string, managerAuth func(resource, token, object string) bool) bool
	ReplacePlaceHolderForEmptyEntriesInRoleFilter(roleFilter bean.RoleFilter) bean.RoleFilter
	RemovePlaceHolderInRoleFilterField(roleFilterField string) string
	GetCapacityForRoleFilter(roleFilters []bean.RoleFilter) (int, map[int]int)
}

type UserCommonServiceImpl struct {
	userAuthRepository  repository2.UserAuthRepository
	logger              *zap.SugaredLogger
	userRepository      repository2.UserRepository
	roleGroupRepository repository2.RoleGroupRepository
	sessionManager2     *middleware.SessionManager
}

func NewUserCommonServiceImpl(userAuthRepository repository2.UserAuthRepository,
	logger *zap.SugaredLogger,
	userRepository repository2.UserRepository,
	userGroupRepository repository2.RoleGroupRepository,
	sessionManager2 *middleware.SessionManager) *UserCommonServiceImpl {
	serviceImpl := &UserCommonServiceImpl{
		userAuthRepository:  userAuthRepository,
		logger:              logger,
		userRepository:      userRepository,
		roleGroupRepository: userGroupRepository,
		sessionManager2:     sessionManager2,
	}
	cStore = sessions.NewCookieStore(randKey())
	return serviceImpl
}

func (impl UserCommonServiceImpl) RemoveRolesAndReturnEliminatedPolicies(userInfo *bean.UserInfo,
	existingRoleIds map[int]repository2.UserRoleModel, eliminatedRoleIds map[int]*repository2.UserRoleModel,
	tx *pg.Tx, token string, managerAuth func(resource, token, object string) bool) ([]casbin2.Policy, error) {
	var eliminatedPolicies []casbin2.Policy
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
							oldRoleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(roleFilter.Entity, "", "", "", "", accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, true)
							if err != nil {
								return nil, err
							}
							if roleModel.Id == 0 && oldRoleModel.Id == 0 {
								impl.logger.Warnw("no role found for given filter", "filter", roleFilter)
								continue
							}
							if _, ok := existingRoleIds[roleModel.Id]; ok {
								delete(eliminatedRoleIds, roleModel.Id)
							}
							if _, ok := existingRoleIds[oldRoleModel.Id]; ok {
								//delete old role mapping from existing but not from eliminated roles (so that it gets deleted)
								delete(existingRoleIds, oldRoleModel.Id)
							}
						}
					}
				}
			}
		} else {
			if len(roleFilter.Team) > 0 { // check auth only for apps permission, skip for chart group
				rbacObject := fmt.Sprintf("%s", strings.ToLower(roleFilter.Team))
				isValidAuth := managerAuth(casbin2.ResourceUser, token, rbacObject)
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
					if _, ok := existingRoleIds[oldRoleModel.Id]; ok {
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
			isValidAuth := managerAuth(casbin2.ResourceUser, token, rbacObject)
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
		eliminatedPolicies = append(eliminatedPolicies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(userInfo.EmailId), Obj: casbin2.Object(role.Role)})
	}
	// DELETE ENDS
	return eliminatedPolicies, nil
}

func (impl UserCommonServiceImpl) RemoveRolesAndReturnEliminatedPoliciesForGroups(request *bean.RoleGroup, existingRoles map[int]*repository2.RoleGroupRoleMapping, eliminatedRoles map[int]*repository2.RoleGroupRoleMapping, tx *pg.Tx, token string, managerAuth func(resource string, token string, object string) bool) ([]casbin2.Policy, error) {
	// Filter out removed items in current request
	//var policies []casbin2.Policy
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
							oldRoleModel, err := impl.userAuthRepository.GetRoleByFilterForAllTypes(entity, "", "", "", "", accessType, roleFilter.Cluster, namespace, group, kind, resource, actionType, true)
							if err != nil {
								impl.logger.Errorw("Error in fetching roles by filter", "user", request)
								return nil, err
							}
							if roleModel.Id == 0 && oldRoleModel.Id == 0 {
								impl.logger.Warnw("no role found for given filter", "filter", roleFilter)
								continue
							}
							if _, ok := existingRoles[roleModel.Id]; ok {
								delete(eliminatedRoles, roleModel.Id)
							}
							if _, ok := existingRoles[oldRoleModel.Id]; ok {
								//delete old role mapping from existing but not from eliminated roles (so that it gets deleted)
								delete(existingRoles, oldRoleModel.Id)
							}
						}
					}
				}
			}
		} else {
			if len(roleFilter.Team) > 0 { // check auth only for apps permission, skip for chart group
				rbacObject := fmt.Sprintf("%s", strings.ToLower(roleFilter.Team))
				isValidAuth := managerAuth(casbin2.ResourceUser, token, rbacObject)
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
					if _, ok := existingRoles[oldRoleModel.Id]; ok {
						//delete old role mapping from existing but not from eliminated roles (so that it gets deleted)
						delete(existingRoles, oldRoleModel.Id)
					}

				}
			}
		}
	}

	//delete remaining Ids from casbin role mapping table in orchestrator and casbin policy db
	// which are existing but not provided in this request
	var eliminatedPolicies []casbin2.Policy
	for _, model := range eliminatedRoles {
		role, err := impl.userAuthRepository.GetRoleById(model.RoleId)
		if err != nil {
			return nil, err
		}
		if len(role.Team) > 0 {
			rbacObject := fmt.Sprintf("%s", strings.ToLower(role.Team))
			isValidAuth := managerAuth(casbin2.ResourceUser, token, rbacObject)
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
		eliminatedPolicies = append(eliminatedPolicies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(policyGroup.CasbinName), Obj: casbin2.Object(role.Role)})
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

	rbacResource := fmt.Sprintf("%s/%s/%s", strings.ToLower(cluster), strings.ToLower(namespaceObj), casbin2.ResourceUser)
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
