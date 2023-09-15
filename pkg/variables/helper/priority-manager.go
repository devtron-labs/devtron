package helper

import (
	repository2 "github.com/devtron-labs/devtron/pkg/variables/repository"
)

func QualifierComparator(a, b repository2.Qualifier) bool {
	return GetPriority(a) < GetPriority(b)
}
func FindMinWithComparator(variableScope []*repository2.VariableScope, comparator func(a, b repository2.Qualifier) bool) *repository2.VariableScope {
	if len(variableScope) == 0 {
		return nil
	}
	min := variableScope[0]
	for _, val := range variableScope {
		if comparator(repository2.Qualifier(val.QualifierId), repository2.Qualifier(min.QualifierId)) {
			min = val
		}
	}
	return min
}

func GetPriority(qualifier repository2.Qualifier) int {
	switch qualifier {
	case repository2.GLOBAL_QUALIFIER:
		return 5
	default:
		return 0
	}
}
