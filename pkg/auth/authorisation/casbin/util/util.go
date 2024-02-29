package util

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	bean4 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin/bean"
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

func GetUsersForActivePolicy(groupPolicies []bean4.GroupPolicy) ([]string, error) {
	users := make([]string, 0)
	recordedTime := time.Now()
	for _, policy := range groupPolicies {
		isActive, err := IsGroupPolicyActive(policy.TimeoutWindowExpression, policy.ExpressionFormat, recordedTime)
		if err != nil {
			fmt.Println("error in GetUsersForActivePolicy", "err", err, "policy", policy)
			return nil, err
		}
		if isActive {
			users = append(users, policy.User)
		}
	}
	return users, nil
}

func GetStringSliceWithUserAndRole(user, role string) []string {
	return []string{user, role}
}

func GetStringSliceWithUserRoleExpressionAndFormat(user, role, expression, format string) []string {
	return []string{user, role, expression, format}
}
