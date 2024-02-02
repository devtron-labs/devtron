package helper

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"golang.org/x/exp/slices"
)

func CheckIfUserDevtronManaged(userId int32) bool {
	if userId == bean.SystemUserId || userId == bean.AdminUserId {
		return false
	}
	return true
}

func CheckValidationForAdminAndSystemUserId(userIds []int32) error {
	if len(userIds) == 0 {
		err := &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "no user ids provided"}
		return err
	}
	validated := CheckIfUserDevtronManagedOnly(userIds)
	if !validated {
		err := &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "cannot update status for system or admin user"}
		return err
	}
	return nil
}

func CheckIfUserDevtronManagedOnly(userIds []int32) bool {
	if slices.Contains(userIds, bean.AdminUserId) || slices.Contains(userIds, bean.SystemUserId) {
		return false
	}
	return true
}
