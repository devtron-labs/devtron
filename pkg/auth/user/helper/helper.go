package helper

import (
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
)

func CheckIfUserDevtronManaged(userId int32) bool {
	if userId == bean.SystemUserId || userId == bean.AdminUserId {
		return false
	}
	return true
}
