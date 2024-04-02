package helper

import (
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/util"
	"time"
)

func GetStatusFromTimeoutWindowExpression(expression string, recordedTime time.Time, expressionFormat bean.ExpressionFormat) (bean2.Status, time.Time) {
	parsedTime, err := util.GetParsedTimeFromExpression(expression, expressionFormat)
	if err != nil {
		return bean2.Inactive, parsedTime
	}
	if parsedTime.IsZero() || parsedTime.Before(recordedTime) {
		// sending time zero in case of inactive status,ignoring time expression in db,in automatic expire case
		return bean2.Inactive, time.Time{}
	}
	return bean2.Active, parsedTime
}
