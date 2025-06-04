package user

import (
	"fmt"
	casbin2 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	bean4 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/adapter"
	userBean "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	userrepo "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"strings"
)

func (impl *UserServiceImpl) UpdateDataForGroupClaims(dto *userBean.SelfRegisterDto) error {
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

	groups, err := casbin2.GetRolesForUser(model.EmailId)
	if err != nil {
		impl.logger.Errorw("No Roles Found for user", "id", model.Id)
		return nil, err
	}
	if len(groups) > 0 {
		// getting unique, handling for duplicate roles
		roleFromGroups, err := impl.getUniquesRolesByGroupCasbinNames(groups)
		if err != nil {
			impl.logger.Errorw("error in getUniquesRolesByGroupCasbinNames", "err", err)
			return nil, err
		}
		groups = append(groups, roleFromGroups...)
	}

	return groups, nil
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
