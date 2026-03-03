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
	"github.com/devtron-labs/authenticator/jwt"
	casbin2 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	bean4 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/adapter"
	userBean "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	userrepo "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	util3 "github.com/devtron-labs/devtron/pkg/auth/user/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/juju/errors"
	"strings"
	"time"
)

func (impl *UserServiceImpl) UpdateDataForGroupClaims(dto *userBean.SelfRegisterDto) error {
	userInfo := dto.UserInfo
	if dto.GroupClaimsConfigActive {
		err := impl.updateDataForUserGroupClaimsMap(userInfo.Id, dto.GroupsFromClaims)
		if err != nil {
			impl.logger.Errorw("error in updating data for user group claims map", "err", err, "userId", userInfo.Id)
			return err
		}
	}
	return nil
}

func (impl *UserServiceImpl) mergeAccessRoleFiltersAndUserGroups(currentUserInfo, requestUserInfo *userBean.UserInfo) {
	return
}

func (impl *UserServiceImpl) setTimeoutWindowConfigIdInUserModel(tx *pg.Tx, userInfo *userBean.UserInfo, model *userrepo.UserModel) error {
	return nil
}

func (impl *UserServiceImpl) assignUserGroups(tx *pg.Tx, userInfo *userBean.UserInfo, model *userrepo.UserModel) error {
	return nil
}

func (impl *UserServiceImpl) checkAndPerformOperationsForGroupClaims(tx *pg.Tx, userInfo *userBean.UserInfo) (bool, error) {
	return false, nil
}

func getFinalRoleFiltersToBeConsidered(userInfo *userBean.UserInfo) []userBean.RoleFilter {
	return userInfo.RoleFilters
}

func validateAccessRoleFilters(info *userBean.UserInfo) error {
	return nil
}

func (impl *UserServiceImpl) createAuditForSelfRegisterOperation(tx *pg.Tx, userResponseInfo *userBean.UserInfo) error {
	return nil
}

func (impl *UserServiceImpl) createAuditForCreateOperation(tx *pg.Tx, userResponseInfo *userBean.UserInfo, model *userrepo.UserModel) error {
	return nil
}

func (impl *UserServiceImpl) getCasbinPolicyForGroup(tx *pg.Tx, emailId, userGroupCasbinName string, userRoleGroup userBean.UserRoleGroup, userLoggedInId int32) (bean4.Policy, error) {
	casbinPolicy := adapter.GetCasbinGroupPolicy(emailId, userGroupCasbinName, nil)
	return casbinPolicy, nil
}

func getUniqueKeyForRoleFilter(role userBean.RoleFilter) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s-%s-%s-%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment,
		role.EntityName, role.Action, role.AccessType, role.Cluster, role.Namespace, role.Group, role.Kind, role.Resource, role.Workflow)
}

func getUniqueKeyForUserRoleGroup(userRoleGroup userBean.UserRoleGroup) string {
	return fmt.Sprintf("%s", userRoleGroup.RoleGroup.Name)
}

func (impl *UserServiceImpl) updateUserGroupForUser(tx *pg.Tx, userInfo *userBean.UserInfo, model *userrepo.UserModel) (bool, error) {
	return false, nil
}

func (impl *UserServiceImpl) saveAuditBasedOnActiveOrInactiveUser(tx *pg.Tx, isUserActive bool, model *userrepo.UserModel, userInfo *userBean.UserInfo) error {
	return nil
}

func setStatusFilterType(request *userBean.ListingRequest) {
	return
}

func setCurrentTimeInUserInfo(request *userBean.ListingRequest) {
	return
}

func (impl *UserServiceImpl) getTimeoutWindowConfig(tx *pg.Tx, roleFilter userBean.RoleFilter, userLoggedInId int32) (*userBean.TimeoutWindowConfigDto, error) {
	return nil, nil
}

func getSubactionFromRoleFilter(roleFilter userBean.RoleFilter) string {
	return ""
}

func (impl *UserServiceImpl) CheckUserRoles(id int32, token string) ([]string, error) {
	model, err := impl.userRepository.GetByIdIncludeDeleted(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	var groups []string
	// devtron-system-managed path: get roles from casbin directly
	activeRoles, err := casbin2.GetRolesForUser(model.EmailId)
	if err != nil {
		impl.logger.Errorw("No Roles Found for user", "id", model.Id)
		return nil, err
	}
	groups = append(groups, activeRoles...)
	if len(groups) > 0 {
		// getting unique, handling for duplicate roles
		roleFromGroups, err := impl.getUniquesRolesByGroupCasbinNames(groups)
		if err != nil {
			impl.logger.Errorw("error in getUniquesRolesByGroupCasbinNames", "err", err)
			return nil, err
		}
		groups = append(groups, roleFromGroups...)
	}

	// group claims path: check group claims active and add role groups from JWT claims
	isGroupClaimsActive := impl.globalAuthorisationConfigService.IsGroupClaimsConfigActive()
	if isGroupClaimsActive && !strings.HasPrefix(model.EmailId, userBean.API_TOKEN_USER_EMAIL_PREFIX) {
		_, groupClaims, err := impl.GetEmailAndGroupClaimsFromToken(token)
		if err != nil {
			impl.logger.Errorw("error in GetEmailAndGroupClaimsFromToken", "err", err)
			return nil, err
		}
		if len(groupClaims) > 0 {
			groupsCasbinNames := util3.GetGroupCasbinName(groupClaims)
			grps, err := impl.getUniquesRolesByGroupCasbinNames(groupsCasbinNames)
			if err != nil {
				impl.logger.Errorw("error in getUniquesRolesByGroupCasbinNames", "err", err)
				return nil, err
			}
			groups = append(groups, grps...)
		}
	}

	return groups, nil
}

func (impl *UserServiceImpl) UpdateUserGroupMappingIfActiveUser(emailId string, groups []string) error {
	user, err := impl.userRepository.FetchActiveUserByEmail(emailId)
	if err != nil {
		impl.logger.Errorw("error in getting active user by email", "err", err, "emailId", emailId)
		return err
	}
	err = impl.updateDataForUserGroupClaimsMap(user.Id, groups)
	if err != nil {
		impl.logger.Errorw("error in updating data for user group claims map", "err", err, "userId", user.Id)
		return err
	}
	return nil
}

func (impl *UserServiceImpl) updateDataForUserGroupClaimsMap(userId int32, groups []string) error {
	//updating groups received in claims
	mapOfGroups := make(map[string]bool, len(groups))
	for _, group := range groups {
		mapOfGroups[group] = true
	}
	groupMappings, err := impl.userGroupMapRepository.GetByUserId(userId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting user group mapping by userId", "err", err, "userId", userId)
		return err
	}
	dbConnection := impl.userGroupMapRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating transaction", "err", err)
		return err
	}
	defer tx.Rollback()
	modelsToBeSaved := make([]*userrepo.UserAutoAssignedGroup, 0)
	modelsToBeUpdated := make([]*userrepo.UserAutoAssignedGroup, 0)
	timeNow := time.Now()
	for i := range groupMappings {
		groupMapping := groupMappings[i]
		//checking if mapping present in groups from claims
		if _, ok := mapOfGroups[groupMapping.GroupName]; ok {
			//present so marking active flag true
			groupMapping.Active = true
			//deleting entry from map now
			delete(mapOfGroups, groupMapping.GroupName)
		} else {
			//not present so marking active flag false
			groupMapping.Active = false
		}
		groupMapping.UpdatedOn = timeNow
		groupMapping.UpdatedBy = userBean.SystemUserId //system user

		//adding this group mapping to updated models irrespective of active
		modelsToBeUpdated = append(modelsToBeUpdated, groupMapping)
	}

	//iterating through remaining groups from the map, they are not found in current entries so need to be saved
	for group := range mapOfGroups {
		modelsToBeSaved = append(modelsToBeSaved, &userrepo.UserAutoAssignedGroup{
			UserId:            userId,
			GroupName:         group,
			IsGroupClaimsData: true,
			Active:            true,
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: timeNow,
				UpdatedBy: 1,
				UpdatedOn: timeNow,
			},
		})
	}
	if len(modelsToBeUpdated) > 0 {
		err = impl.userGroupMapRepository.Update(modelsToBeUpdated, tx)
		if err != nil {
			impl.logger.Errorw("error in updating user group mapping", "err", err)
			return err
		}
	}
	if len(modelsToBeSaved) > 0 {
		err = impl.userGroupMapRepository.Save(modelsToBeSaved, tx)
		if err != nil {
			impl.logger.Errorw("error in saving user group mapping", "err", err)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return err
	}
	return nil
}

func (impl *UserServiceImpl) getRoleFiltersForGroupClaims(id int32) ([]userBean.RoleFilter, error) {
	var roleFilters []userBean.RoleFilter
	userGroups, err := impl.userGroupMapRepository.GetActiveByUserId(id)
	if err != nil {
		impl.logger.Errorw("error in GetActiveByUserId", "err", err, "userId", id)
		return nil, err
	}
	groupClaims := make([]string, 0, len(userGroups))
	for _, userGroup := range userGroups {
		groupClaims = append(groupClaims, userGroup.GroupName)
	}
	// checking by group casbin name (considering case insensitivity here)
	if len(groupClaims) > 0 {
		groupCasbinNames := util3.GetGroupCasbinName(groupClaims)
		groupFilters, err := impl.GetRoleFiltersByGroupCasbinNames(groupCasbinNames)
		if err != nil {
			impl.logger.Errorw("error while GetRoleFiltersByGroupCasbinNames", "error", err, "groupCasbinNames", groupCasbinNames)
			return nil, err
		}
		if len(groupFilters) > 0 {
			roleFilters = append(roleFilters, groupFilters...)
		}
	}
	return roleFilters, nil
}

func (impl *UserServiceImpl) getRoleGroupsForGroupClaims(id int32) ([]userBean.UserRoleGroup, error) {
	userGroups, err := impl.userGroupMapRepository.GetActiveByUserId(id)
	if err != nil {
		impl.logger.Errorw("error in GetActiveByUserId", "err", err, "userId", id)
		return nil, err
	}
	groupClaims := make([]string, 0, len(userGroups))
	for _, userGroup := range userGroups {
		groupClaims = append(groupClaims, userGroup.GroupName)
	}
	// checking by group casbin name (considering case insensitivity here)
	var userRoleGroups []userBean.UserRoleGroup
	if len(groupClaims) > 0 {
		groupCasbinNames := util3.GetGroupCasbinName(groupClaims)
		userRoleGroups, err = impl.fetchUserRoleGroupsByGroupClaims(groupCasbinNames)
		if err != nil {
			impl.logger.Errorw("error in fetchUserRoleGroupsByGroupClaims ", "err", err, "groupClaims", groupClaims)
			return nil, err
		}
	}
	return userRoleGroups, nil
}

func (impl *UserServiceImpl) fetchUserRoleGroupsByGroupClaims(groupCasbinNames []string) ([]userBean.UserRoleGroup, error) {
	roleGroups, err := impl.roleGroupRepository.GetRoleGroupListByCasbinNames(groupCasbinNames)
	if err != nil {
		impl.logger.Errorw("error in fetchUserRoleGroupsByGroupClaims", "err", err, "groupCasbinNames", groupCasbinNames)
		return nil, err
	}
	userRoleGroups := make([]userBean.UserRoleGroup, 0, len(roleGroups))
	for _, roleGroup := range roleGroups {
		userRoleGroups = append(userRoleGroups, userBean.UserRoleGroup{
			RoleGroup: &userBean.RoleGroup{
				Id:          roleGroup.Id,
				Name:        roleGroup.Name,
				Description: roleGroup.Description,
			},
		})
	}
	return userRoleGroups, nil
}

// GetRoleFiltersByGroupCasbinNames returns role filters for the given group casbin names
func (impl *UserServiceImpl) GetRoleFiltersByGroupCasbinNames(groupCasbinNames []string) ([]userBean.RoleFilter, error) {
	roles, err := impl.roleGroupRepository.GetRolesByGroupCasbinNames(groupCasbinNames)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting roles by group casbin names", "err", err)
		return nil, err
	}
	var roleFilters []userBean.RoleFilter
	// merging considering env as base first
	roleFilters = impl.userCommonService.BuildRoleFiltersAfterMerging(ConvertRolesToEntityProcessors(roles), userBean.EnvironmentBasedKey)
	// merging role filters based on application
	roleFilters = impl.userCommonService.BuildRoleFiltersAfterMerging(ConvertRoleFiltersToEntityProcessors(roleFilters), userBean.ApplicationBasedKey)
	return roleFilters, nil
}

// GetEmailAndGroupClaimsFromToken extracts email and group claims from the JWT token
func (impl *UserServiceImpl) GetEmailAndGroupClaimsFromToken(token string) (string, []string, error) {
	if token == "" {
		return "", nil, nil
	}
	mapClaims, err := impl.getMapClaims(token)
	if err != nil {
		return "", nil, err
	}
	email, groups := impl.globalAuthorisationConfigService.GetEmailAndGroupsFromClaims(mapClaims)
	return email, groups, nil
}

func (impl *UserServiceImpl) getMapClaims(token string) (jwtv4.MapClaims, error) {
	claims, err := impl.sessionManager2.VerifyToken(token)
	if err != nil {
		impl.logger.Errorw("failed to verify token", "error", err)
		return nil, err
	}
	mapClaims, err := jwt.MapClaims(claims)
	if err != nil {
		impl.logger.Errorw("failed to MapClaims", "error", err)
		return nil, err
	}
	return mapClaims, nil
}

func (impl *UserServiceImpl) getUserGroupMapFromModels(model []userrepo.UserModel) (*userBean.UserGroupMapDto, error) {
	return nil, nil
}

func setTwcId(model *userrepo.UserModel, twcId int) {
	return
}

func (impl *UserServiceImpl) getTimeoutWindowID(tx *pg.Tx, userInfo *userBean.UserInfo) (int, error) {
	return 0, nil
}

// createOrUpdateUserRoleGroupsPolices : gives policies which are to be added and which are to be eliminated from casbin, with support of timewindow Config changed fromm existing
func (impl *UserServiceImpl) createOrUpdateUserRoleGroupsPolices(requestUserRoleGroups []userBean.UserRoleGroup, emailId string, tx *pg.Tx, loggedInUser int32, userInfoId int32) ([]bean4.Policy, []bean4.Policy, []*userrepo.RoleModel, map[string]bool, error) {
	userCasbinRoles, err := impl.CheckUserRoles(userInfoId, "")
	if err != nil {
		impl.logger.Errorw("error encountered in createOrUpdateUserRoleGroupsPolices", "userRoleGroups", requestUserRoleGroups, "emailId", emailId, "err", err)
		return nil, nil, nil, nil, err
	}
	// initialisation

	newGroupMap := make(map[string]string)
	oldGroupMap := make(map[string]string)
	mapOfExistingUserRoleGroup := make(map[string]bool, len(userCasbinRoles))
	addedPolicies := make([]bean4.Policy, 0)
	eliminatedPolicies := make([]bean4.Policy, 0)
	eliminatedGroupCasbinNames := make([]string, 0, len(newGroupMap))
	var eliminatedGroupRoles []*userrepo.RoleModel
	for _, oldItem := range userCasbinRoles {
		oldGroupMap[oldItem] = oldItem
		mapOfExistingUserRoleGroup[oldItem] = true
	}
	// START GROUP POLICY
	for _, item := range requestUserRoleGroups {
		userGroup, err := impl.roleGroupRepository.GetRoleGroupByName(item.RoleGroup.Name)
		if err != nil {
			impl.logger.Errorw("error encountered in createOrUpdateUserRoleGroupsPolices", "userRoleGroups", requestUserRoleGroups, "emailId", emailId, "err", err)
			return nil, nil, nil, nil, err
		}
		newGroupMap[userGroup.CasbinName] = userGroup.CasbinName
		if _, ok := oldGroupMap[userGroup.CasbinName]; !ok {
			addedPolicies = append(addedPolicies, bean4.Policy{Type: "g", Sub: bean4.Subject(emailId), Obj: bean4.Object(userGroup.CasbinName)})
		}
	}
	for _, item := range userCasbinRoles {
		if _, ok := newGroupMap[item]; !ok {
			if item != userBean.SUPERADMIN {
				//check permission for group which is going to eliminate
				if strings.HasPrefix(item, "group:") {
					eliminatedPolicies = append(eliminatedPolicies, bean4.Policy{Type: "g", Sub: bean4.Subject(emailId), Obj: bean4.Object(item)})
					eliminatedGroupCasbinNames = append(eliminatedGroupCasbinNames, item)
				}
			}
		}
	} // END GROUP POLICY
	if len(eliminatedGroupCasbinNames) > 0 {
		eliminatedGroupRoles, err = impl.roleGroupRepository.GetRolesByGroupCasbinNames(eliminatedGroupCasbinNames)
		if err != nil {
			impl.logger.Errorw("error encountered in createOrUpdateUserRoleGroupsPolices", "userRoleGroups", requestUserRoleGroups, "emailId", emailId, "err", err)
			return nil, nil, nil, nil, err
		}
	}
	return addedPolicies, eliminatedPolicies, eliminatedGroupRoles, mapOfExistingUserRoleGroup, nil
}

func (impl *UserServiceImpl) deleteUserCasbinPolices(model *userrepo.UserModel) error {
	groups, err := casbin2.GetRolesForUser(model.EmailId)
	if err != nil {
		impl.logger.Warnw("No Roles Found for user", "id", model.Id)
		return err
	}
	for _, item := range groups {
		flag := casbin2.DeleteRoleForUser(model.EmailId, item)
		if flag == false {
			impl.logger.Warnw("unable to delete role:", "user", model.EmailId, "role", item)
		}
	}
	return nil
}

func getApproverFromRoleFilter(roleFilter userBean.RoleFilter) bool {
	return false
}

func (impl *UserServiceImpl) checkValidationAndPerformOperationsForUpdate(token string, tx *pg.Tx, model *userrepo.UserModel, userInfo *userBean.UserInfo, userGroupsUpdated bool, timeoutWindowConfigId int) (operationCompleted bool, isUserSuperAdmin bool, err error) {
	//validating if action user is not admin and trying to update user who has super admin polices, return 403
	// isUserSuperAdminOrManageAllAccess only super-admin is checked as manage all access is not applicable for user
	isUserSuperAdmin, err = impl.IsSuperAdmin(int(userInfo.Id))
	if err != nil {
		return false, isUserSuperAdmin, err
	}
	return false, isUserSuperAdmin, nil
}

func (impl *UserServiceImpl) getUserWithTimeoutWindowConfiguration(emailId string) (int32, bool, error) {
	user, err := impl.userRepository.FetchActiveUserByEmail(emailId)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		errMsg := fmt.Sprintf("failed to fetch user by email id, err: %s", err.Error())
		return 0, false, errors.New(errMsg)
	}
	// here false is always returned to match signature of authoriser function.
	return user.Id, false, nil
}
