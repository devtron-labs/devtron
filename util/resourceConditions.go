/*
 * Copyright (c) 2024. Devtron Inc.
 */

package util

type ResourceConditionType int

const (
	FAIL ResourceConditionType = iota
	PASS
)

type ResourceCondition struct {
	ConditionType ResourceConditionType `json:"conditionType" validate:"min=0,max=1"`
	Expression    string                `json:"expression" validate:"required,min=1"`
	ErrorMsg      string                `json:"errorMsg,omitempty"`
}

func (condition ResourceCondition) IsFailCondition() bool {
	return condition.ConditionType == FAIL
}
