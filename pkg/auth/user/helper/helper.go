package helper

import (
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"golang.org/x/exp/slices"
	"time"
)

func IsSystemOrAdminUser(userId int32) bool {
	if userId == bean.SystemUserId || userId == bean.AdminUserId {
		return true
	}
	return false
}

func IsSystemOrAdminUserByEmail(email string) bool {
	if email == bean.AdminUser || email == bean.SystemUser {
		return true
	}
	return false
}

func CheckValidationForAdminAndSystemUserId(userIds []int32) error {
	validated := CheckIfUserDevtronManagedOnly(userIds)
	if !validated {
		err := &util.ApiError{Code: "406", HttpStatusCode: 406, UserMessage: "cannot update status for system or admin user"}
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

func CheckIfUserIdsExists(userIds []int32) error {
	var err error
	if len(userIds) == 0 {
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "no user ids provided"}
		return err
	}
	return nil
}

func GetActualStatusFromExpressionAndStatus(status bean2.Status, timeoutWindowExpression time.Time) bean2.Status {
	if status == bean2.Active && timeoutWindowExpression.IsZero() {
		return bean2.Active
	} else if status == bean2.Inactive {
		return bean2.Inactive
	} else if status == bean2.Active && !timeoutWindowExpression.IsZero() {
		return bean2.TemporaryAccess
	}
	return bean2.Active

}

func HasTimeWindowChanged(status bean2.Status, expression time.Time, timeWindowConfiguration *repository.TimeoutWindowConfiguration) bool {
	var timeZero time.Time
	isTimeWindowConfigurationNil := timeWindowConfiguration == nil
	if isTimeWindowConfigurationNil && !(status == bean2.Active) {
		return true
	} else if isTimeWindowConfigurationNil && status == bean2.Active {
		return false
	} else if status == bean2.Inactive && (timeWindowConfiguration.TimeoutWindowExpression == timeZero.String() && timeWindowConfiguration.ExpressionFormat == bean3.TimeZeroFormat) {
		return false
	} else if status == bean2.TemporaryAccess && (timeWindowConfiguration.TimeoutWindowExpression == expression.String() && timeWindowConfiguration.ExpressionFormat == bean3.TimeStamp) {
		return false
	}
	return true
}
