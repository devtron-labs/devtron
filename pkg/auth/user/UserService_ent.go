package user

import (
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
