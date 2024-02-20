package helper

import (
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	"strconv"
)

func GetCasbinFormattedTimeAndFormat(timeoutWindowConfig *repository.TimeoutWindowConfiguration) (string, string) {
	if timeoutWindowConfig == nil {
		return "", ""
	}
	formattedString := timeoutWindowConfig.TimeoutWindowExpression
	expressionFormatString := strconv.Itoa(int(timeoutWindowConfig.ExpressionFormat))
	return formattedString, expressionFormatString
}
