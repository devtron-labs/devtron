package util

import (
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/timeoutWindow/bean"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"time"
)

func GetParsedTimeFromExpression(expression string, format bean.ExpressionFormat) (time.Time, error) {
	// considering default to timestamp , will add support for other formats here in future
	switch format {
	case bean.TimeStamp:
		return parseExpressionToTime(expression)
	case bean.TimeZeroFormat:
		// Considering format timeZeroFormat for extremities, kept it in other format but represents UTC time
		return parseExpressionToTime(expression)
	default:
		return parseExpressionToTime(expression)
	}

	return time.Time{}, &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "expression format not supported"}

}

func parseExpressionToTime(expression string) (time.Time, error) {
	parsedTime, err := time.Parse(bean2.TimeFormatForParsing, expression)
	if err != nil {
		return parsedTime, err
	}
	return parsedTime, err
}
