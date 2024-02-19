package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	"strconv"
	"time"
)

func GetCasbinFormattedTimeAndFormat(timeoutWindowConfig *repository.TimeoutWindowConfiguration) (string, string, error) {
	if timeoutWindowConfig == nil {
		return "", "", nil
	}
	parsedTime, err := time.Parse(TimeFormatForParsing, timeoutWindowConfig.TimeoutWindowExpression)
	if err != nil {
		fmt.Errorf("error in GetCasbinFormattedTime", "TimeoutWindowExpression", timeoutWindowConfig.TimeoutWindowExpression, "err", err)
		return "", "", err
	}

	// Formatting the parsed time into the DateTime format
	formattedString := parsedTime.Format(time.DateTime)
	expressionFormatString := strconv.Itoa(int(timeoutWindowConfig.ExpressionFormat))
	return formattedString, expressionFormatString, nil
}
