package helper

import (
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
)

func QualifierComparator(a, b resourceQualifiers.Qualifier) bool {
	return GetPriority(a) < GetPriority(b)
}
func FindMinWithComparator(variableScope []*resourceQualifiers.QualifierMapping, comparator func(a, b resourceQualifiers.Qualifier) bool) *resourceQualifiers.QualifierMapping {
	if len(variableScope) == 0 {
		return nil
	}
	min := variableScope[0]
	for _, val := range variableScope {
		if comparator(resourceQualifiers.Qualifier(val.QualifierId), resourceQualifiers.Qualifier(min.QualifierId)) {
			min = val
		}
	}
	return min
}

func GetPriority(qualifier resourceQualifiers.Qualifier) int {
	switch qualifier {
	case resourceQualifiers.GLOBAL_QUALIFIER:
		return 5
	default:
		return 0
	}
}
