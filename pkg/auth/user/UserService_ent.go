package user

import (
	"fmt"
	casbin2 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	bean4 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/adapter"
	userBean "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	userrepo "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/go-pg/pg"
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

func (impl *UserServiceImpl) CheckUserRoles(id int32) ([]string, error) {
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
