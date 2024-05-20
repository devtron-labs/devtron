package adapter

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"strings"
	"time"
)

func GetLastLoginTime(model repository.UserModel) time.Time {
	lastLoginTime := time.Time{}
	if model.UserAudit != nil {
		lastLoginTime = model.UserAudit.UpdatedOn
	}
	return lastLoginTime
}

func CreateRestrictedGroup(roleGroupName string, hasSuperAdminPermission bool) bean.RestrictedGroup {
	trimmedGroup := strings.TrimPrefix(roleGroupName, "group:")
	return bean.RestrictedGroup{
		Group:                   trimmedGroup,
		HasSuperAdminPermission: hasSuperAdminPermission,
	}
}
