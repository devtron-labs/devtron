package user

import (
	"fmt"
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
	casbinPolicy := adapter.GetCasbinGroupPolicy(emailId, userGroupCasbinName)
	return casbinPolicy, nil
}

func getUniqueKeyForRoleFilter(role userBean.RoleFilter) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s-%s-%s-%s-%s-%s-%s-%s", role.Entity, role.Team, role.Environment,
		role.EntityName, role.Action, role.AccessType, role.Cluster, role.Namespace, role.Group, role.Kind, role.Resource, role.Workflow)
}

func getUniqueKeyForUserRoleGroup(userRoleGroup userBean.UserRoleGroup) string {
	return fmt.Sprintf("%s", userRoleGroup.RoleGroup.Name)
}
