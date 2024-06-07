/*
 * Copyright (c) 2024. Devtron Inc.
 */

package helper

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"strconv"
	"strings"
	"time"
)

func GetCasbinFormattedTimeAndFormat(timeoutWindowConfig *repository.TimeoutWindowConfiguration) (string, string) {
	if timeoutWindowConfig == nil {
		return "", ""
	}
	formattedString := timeoutWindowConfig.TimeoutWindowExpression
	expressionFormatString := strconv.Itoa(int(timeoutWindowConfig.ExpressionFormat))
	return formattedString, expressionFormatString
}

func GetCasbinFormattedTimeAndFormatFromStatusAndExpression(status bean.Status, timeoutWindowExpression time.Time) (string, string) {
	if status == bean.Active && timeoutWindowExpression.IsZero() {
		return "", ""
	} else if status == bean.Active && !timeoutWindowExpression.IsZero() {
		return strings.ToLower(timeoutWindowExpression.String()), strconv.Itoa(int(bean2.TimeStamp))
	} else if status == bean.Inactive {
		return strings.ToLower(timeoutWindowExpression.String()), strconv.Itoa(int(bean2.TimeZeroFormat))
	}
	return "", ""
}
