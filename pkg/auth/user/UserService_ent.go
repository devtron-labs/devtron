package user

import (
	userBean "github.com/devtron-labs/devtron/pkg/auth/user/bean"
)

func (impl *UserServiceImpl) UpdateDataForGroupClaims(dto *userBean.SelfRegisterDto) error {
	return nil
}

func (impl *UserServiceImpl) mergeAccessRoleFiltersAndUserGroups(requestUserInfo, currentUserInfo *userBean.UserInfo) {
	return
}
