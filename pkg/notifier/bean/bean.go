/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

type ConditionType int

const (
	FAIL ConditionType = 0
	PASS ConditionType = 1
)

func (condition ConditionType) IsConditionSatisfied(res bool) bool {
	if condition == FAIL {
		return !res
	} else if condition == PASS {
		return res
	}
	return false
}
