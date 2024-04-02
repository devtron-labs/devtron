package adapter

import (
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"time"
)

func GetLastLoginTime(model repository.UserModel) time.Time {
	lastLoginTime := time.Time{}
	if model.UserAudit != nil {
		lastLoginTime = model.UserAudit.UpdatedOn
	}
	return lastLoginTime
}
