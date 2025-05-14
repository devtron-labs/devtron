/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package commonEnforcementFunctionsUtil

import (
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
)

func (impl *CommonEnforcementUtilImpl) CheckRbacForMangerAndAboveAccess(token string, userId int32) (bool, error) {
	isAuthorised := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if !isAuthorised {
		user, err := impl.userService.GetByIdWithoutGroupClaims(userId)
		if err != nil {
			impl.logger.Errorw("error in getting user by id", "err", err)
			return false, err
		}
		var roleFilters []bean2.RoleFilter
		if len(user.UserRoleGroup) > 0 {
			groupRoleFilters, err := impl.userService.GetRoleFiltersByUserRoleGroups(user.UserRoleGroup)
			if err != nil {
				impl.logger.Errorw("Error in getting role filters by group names", "err", err, "UserRoleGroup", user.UserRoleGroup)
				return false, err
			}
			if len(groupRoleFilters) > 0 {
				roleFilters = append(roleFilters, groupRoleFilters...)
			}
		}
		if user.RoleFilters != nil && len(user.RoleFilters) > 0 {
			roleFilters = append(roleFilters, user.RoleFilters...)
		}
		if len(roleFilters) > 0 {
			for _, filter := range roleFilters {
				if len(filter.Team) > 0 {
					if ok := impl.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, filter.Team); ok {
						isAuthorised = true
						break
					}
				}
				if filter.Entity == bean2.CLUSTER_ENTITIY {
					if ok := impl.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, impl.checkManagerAuth); ok {
						isAuthorised = true
						break
					}
				}
			}
		}
	}
	return isAuthorised, nil
}

func (impl *CommonEnforcementUtilImpl) checkManagerAuth(resource, token string, object string) bool {
	if ok := impl.enforcer.Enforce(token, resource, casbin.ActionUpdate, object); !ok {
		return false
	}
	return true

}
