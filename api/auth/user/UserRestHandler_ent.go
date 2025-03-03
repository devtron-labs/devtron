package user

import (
	util2 "github.com/devtron-labs/devtron/api/auth/user/util"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/go-pg/pg"
)

func (handler UserRestHandlerImpl) checkRBACForUserCreate(token string, requestSuperAdmin bool, roleFilters []bean2.RoleFilter,
	roleGroups []bean2.UserRoleGroup) (isAuthorised bool, err error) {
	isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if requestSuperAdmin && !isActionUserSuperAdmin {
		return false, nil
	}
	isAuthorised = isActionUserSuperAdmin
	if !isAuthorised {
		if roleFilters != nil && len(roleFilters) > 0 { //auth check inside roleFilters
			for _, filter := range roleFilters {
				switch {
				case filter.AccessType == bean2.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
					isAuthorised = isActionUserSuperAdmin
				case len(filter.Team) > 0:
					isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, filter.Team)
				case filter.Entity == bean2.CLUSTER_ENTITIY:
					isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
				case filter.Entity == bean2.CHART_GROUP_ENTITY && len(roleFilters) == 1: //if only chartGroup entity is present in request then access will be judged through super-admin access
					isAuthorised = isActionUserSuperAdmin
				case filter.Entity == bean2.CHART_GROUP_ENTITY && len(roleFilters) > 1: //if entities apart from chartGroup entity are present, not checking chartGroup access
					isAuthorised = true
				default:
					isAuthorised = false
				}
				if !isAuthorised {
					return false, nil
				}
			}
		}
		if len(roleGroups) > 0 { // auth check inside groups
			groupRoles, err := handler.roleGroupService.FetchRolesForUserRoleGroups(roleGroups)
			if err != nil && err != pg.ErrNoRows {
				handler.logger.Errorw("service err, UpdateUser", "err", err, "payload", roleGroups)
				return false, err
			}
			if len(groupRoles) > 0 {
				for _, groupRole := range groupRoles {
					switch {
					case groupRole.Action == bean2.ACTION_SUPERADMIN:
						isAuthorised = isActionUserSuperAdmin
					case groupRole.AccessType == bean2.APP_ACCESS_TYPE_HELM || groupRole.Entity == bean2.EntityJobs:
						isAuthorised = isActionUserSuperAdmin
					case len(groupRole.Team) > 0:
						isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, groupRole.Team)
					case groupRole.Entity == bean2.CLUSTER_ENTITIY:
						isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(groupRole.Cluster, groupRole.Namespace, groupRole.Group, groupRole.Kind, groupRole.Resource, token, handler.CheckManagerAuth)
					case groupRole.Entity == bean2.CHART_GROUP_ENTITY && len(groupRoles) == 1: //if only chartGroup entity is present in request then access will be judged through super-admin access
						isAuthorised = isActionUserSuperAdmin
					case groupRole.Entity == bean2.CHART_GROUP_ENTITY && len(groupRoles) > 1: //if entities apart from chartGroup entity are present, not checking chartGroup access
						isAuthorised = true
					default:
						isAuthorised = false
					}
					if !isAuthorised {
						return false, nil
					}
				}
			} else {
				isAuthorised = false
			}
		}
	}
	return isAuthorised, nil
}

func (handler UserRestHandlerImpl) checkRBACForUserUpdate(token string, userInfo *bean2.UserInfo, isUserAlreadySuperAdmin bool, eliminatedRoleFilters,
	eliminatedGroupRoles []*repository.RoleModel, mapOfExistingUserRoleGroup map[string]bool) (isAuthorised bool, err error) {
	isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	requestSuperAdmin := userInfo.SuperAdmin
	if (requestSuperAdmin || isUserAlreadySuperAdmin) && !isActionUserSuperAdmin {
		//if user is going to be provided with super-admin access or already a super-admin then the action user should be a super-admin
		return false, nil
	}
	roleFilters := userInfo.RoleFilters
	roleGroups := userInfo.UserRoleGroup
	isAuthorised = isActionUserSuperAdmin
	eliminatedRolesToBeChecked := append(eliminatedRoleFilters, eliminatedGroupRoles...)
	if !isAuthorised {
		if roleFilters != nil && len(roleFilters) > 0 { //auth check inside roleFilters
			for _, filter := range roleFilters {
				switch {
				case filter.AccessType == bean2.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
					isAuthorised = isActionUserSuperAdmin
				case len(filter.Team) > 0:
					isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, filter.Team)
				case filter.Entity == bean2.CLUSTER_ENTITIY:
					isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
				case filter.Entity == bean2.CHART_GROUP_ENTITY:
					isAuthorised = true
				default:
					isAuthorised = false
				}
				if !isAuthorised {
					return false, nil
				}
			}
		}
		if eliminatedRolesToBeChecked != nil && len(eliminatedRolesToBeChecked) > 0 {
			for _, filter := range eliminatedRolesToBeChecked {
				switch {
				case filter.AccessType == bean2.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
					isAuthorised = isActionUserSuperAdmin
				case len(filter.Team) > 0:
					isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, filter.Team)
				case filter.Entity == bean2.CLUSTER_ENTITIY:
					isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
				case filter.Entity == bean2.CHART_GROUP_ENTITY:
					isAuthorised = true
				default:
					isAuthorised = false
				}
				if !isAuthorised {
					return false, nil
				}
			}
		}
		if len(roleGroups) > 0 { // auth check inside groups
			//filter out roleGroups (existing has to be ignore while checking rbac)
			filteredRoleGroups := util2.FilterRoleGroupIfAlreadyPresent(roleGroups, mapOfExistingUserRoleGroup)
			if len(filteredRoleGroups) > 0 {
				groupRoles, err := handler.roleGroupService.FetchRolesForUserRoleGroups(roleGroups)
				if err != nil && err != pg.ErrNoRows {
					handler.logger.Errorw("service err, UpdateUser", "err", err, "filteredRoleGroups", filteredRoleGroups)
					return false, err
				}
				if len(groupRoles) > 0 {
					for _, groupRole := range groupRoles {
						switch {
						case groupRole.Action == bean2.ACTION_SUPERADMIN:
							isAuthorised = isActionUserSuperAdmin
						case groupRole.AccessType == bean2.APP_ACCESS_TYPE_HELM || groupRole.Entity == bean2.EntityJobs:
							isAuthorised = isActionUserSuperAdmin
						case len(groupRole.Team) > 0:
							isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, groupRole.Team)
						case groupRole.Entity == bean2.CLUSTER_ENTITIY:
							isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(groupRole.Cluster, groupRole.Namespace, groupRole.Group, groupRole.Kind, groupRole.Resource, token, handler.CheckManagerAuth)
						case groupRole.Entity == bean2.CHART_GROUP_ENTITY:
							isAuthorised = true
						default:
							isAuthorised = false
						}
						if !isAuthorised {
							return false, nil
						}
					}
				} else {
					isAuthorised = false
				}
			}
		}
	}
	return isAuthorised, nil
}

func (handler UserRestHandlerImpl) CheckRbacForUserDelete(token string, user *bean2.UserInfo) (isAuthorised bool) {
	isActionUserSuperAdmin := false
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	isAuthorised = isActionUserSuperAdmin
	if !isAuthorised {
		if user.RoleFilters != nil && len(user.RoleFilters) > 0 {
			for _, filter := range user.RoleFilters {
				if filter.AccessType == bean2.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs {
					// helm apps, jobs are managed by super admin
					return false
				}
				if len(filter.Team) > 0 {
					if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionDelete, filter.Team); !ok {
						return false
					}
				}
				if filter.Entity == bean2.CLUSTER_ENTITIY {
					if ok := handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth); !ok {
						return false
					}
				}
			}
		} else {
			if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionDelete, ""); !ok {
				return false
			}
		}
		isAuthorised = true
	}
	return isAuthorised
}

func (handler UserRestHandlerImpl) checkRBACForRoleGroupUpdate(token string, groupInfo *bean2.RoleGroup, eliminatedRoleFilters []*repository.RoleModel, isRoleGroupAlreadySuperAdmin bool) (isAuthorised bool, err error) {
	isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	requestSuperAdmin := groupInfo.SuperAdmin
	if (requestSuperAdmin || isRoleGroupAlreadySuperAdmin) && !isActionUserSuperAdmin {
		//if user is going to be provided with super-admin access or already a super-admin then the action user should be a super-admin
		return false, nil
	}
	isAuthorised = isActionUserSuperAdmin
	if !isAuthorised {
		if groupInfo.RoleFilters != nil && len(groupInfo.RoleFilters) > 0 { //auth check inside roleFilters
			for _, filter := range groupInfo.RoleFilters {
				switch {
				case filter.Action == bean2.ACTION_SUPERADMIN:
					isAuthorised = isActionUserSuperAdmin
				case filter.AccessType == bean2.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
					isAuthorised = isActionUserSuperAdmin
				case len(filter.Team) > 0:
					isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, filter.Team)
				case filter.Entity == bean2.CLUSTER_ENTITIY:
					isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
				case filter.Entity == bean2.CHART_GROUP_ENTITY:
					isAuthorised = true
				default:
					isAuthorised = false
				}
				if !isAuthorised {
					return false, nil
				}
			}
		}
		if len(eliminatedRoleFilters) > 0 {
			for _, filter := range eliminatedRoleFilters {
				switch {
				case filter.AccessType == bean2.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
					isAuthorised = isActionUserSuperAdmin
				case len(filter.Team) > 0:
					isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, filter.Team)
				case filter.Entity == bean2.CLUSTER_ENTITIY:
					isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
				case filter.Entity == bean2.CHART_GROUP_ENTITY:
					isAuthorised = true
				default:
					isAuthorised = false
				}
				if !isAuthorised {
					return false, nil
				}
			}
		}
	}
	return isAuthorised, nil
}

func (handler UserRestHandlerImpl) checkRBACForRoleGroupDelete(token string, userGroup *bean2.RoleGroup) (isAuthorised bool, err error) {
	isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if userGroup.SuperAdmin && !isActionUserSuperAdmin {
		return false, nil
	}
	isAuthorised = isActionUserSuperAdmin
	if !isAuthorised {
		if userGroup.RoleFilters != nil && len(userGroup.RoleFilters) > 0 { //auth check inside roleFilters
			for _, filter := range userGroup.RoleFilters {
				switch {
				case filter.Action == bean2.ACTION_SUPERADMIN:
					isAuthorised = isActionUserSuperAdmin
				case filter.AccessType == bean2.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
					isAuthorised = isActionUserSuperAdmin
				case len(filter.Team) > 0:
					isAuthorised = handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, filter.Team)
				case filter.Entity == bean2.CLUSTER_ENTITIY:
					isAuthorised = handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
				case filter.Entity == bean2.CHART_GROUP_ENTITY:
					isAuthorised = true
				default:
					isAuthorised = false
				}
				if !isAuthorised {
					return false, nil
				}
			}
		}
	}
	return isAuthorised, nil
}

func (handler UserRestHandlerImpl) GetFilteredRoleFiltersAccordingToAccess(token string, roleFilters []bean2.RoleFilter) []bean2.RoleFilter {
	filteredRoleFilter := make([]bean2.RoleFilter, 0)
	if len(roleFilters) > 0 {
		isUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
		for _, filter := range roleFilters {
			authPass := handler.checkRbacForFilter(token, filter, isUserSuperAdmin)
			if authPass {
				filteredRoleFilter = append(filteredRoleFilter, filter)
			}
		}
	}
	for index, roleFilter := range filteredRoleFilter {
		if roleFilter.Entity == "" {
			filteredRoleFilter[index].Entity = bean2.ENTITY_APPS
			if roleFilter.AccessType == "" {
				filteredRoleFilter[index].AccessType = bean2.DEVTRON_APP
			}
		}
	}
	return filteredRoleFilter
}

func (handler UserRestHandlerImpl) checkRbacForFilter(token string, filter bean2.RoleFilter, isUserSuperAdmin bool) bool {
	isAuthorised := true
	switch {
	case isUserSuperAdmin:
		isAuthorised = true
	case filter.AccessType == bean2.APP_ACCESS_TYPE_HELM || filter.Entity == bean2.EntityJobs:
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			isAuthorised = false
		}

	case len(filter.Team) > 0:
		// this is case of devtron app
		if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, filter.Team); !ok {
			isAuthorised = false
		}

	case filter.Entity == bean2.CLUSTER_ENTITIY:
		isValidAuth := handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.CheckManagerAuth)
		if !isValidAuth {
			isAuthorised = false
		}
	case filter.Entity == bean2.CHART_GROUP_ENTITY:
		isAuthorised = true
	default:
		isAuthorised = false
	}
	return isAuthorised
}
