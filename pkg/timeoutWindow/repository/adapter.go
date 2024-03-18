package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"time"
)

func GetTimeoutWindowConfigModel(timeoutExpression string, expressionFormat bean.ExpressionFormat, loggedInUserId int32) *TimeoutWindowConfiguration {
	model := &TimeoutWindowConfiguration{
		TimeoutWindowExpression: timeoutExpression,
		ExpressionFormat:        expressionFormat,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: loggedInUserId,
			UpdatedOn: time.Now(),
			UpdatedBy: loggedInUserId,
		},
	}
	return model
}
