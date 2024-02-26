package util

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/auth/common/helper"
	bean3 "github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"strconv"
	"strings"
	"time"
)

func GetStatusFromCasbinExpressionAndFormat(expression string, format string, recordedTime time.Time) (bean.Status, error) {
	status := bean.Active
	if len(expression) > 0 && len(format) > 0 {
		expressionFormat, err := strconv.Atoi(format)
		if err != nil {
			fmt.Println("error in converting expression format from casbin to Expression Format", "err", err)
			return bean.Inactive, err
		}
		status, _ = helper.GetStatusFromTimeoutWindowExpression(strings.ToUpper(expression), recordedTime, bean3.ExpressionFormat(expressionFormat))
	}
	return status, nil
}

func IsGroupPolicyActive(expression string, format string, recordedTime time.Time) (bool, error) {
	status, err := GetStatusFromCasbinExpressionAndFormat(expression, format, recordedTime)
	if err != nil {
		fmt.Println("error in converting expression format from casbin to Expression Format in IsGroupPolicyActive", "err", err)
		return false, err
	}
	if status == bean.Active {
		return true, nil
	}
	return false, nil
}
