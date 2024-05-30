package resourceQualifiers

func QualifierComparator(a, b Qualifier) bool {
	return GetPriority(a) < GetPriority(b)
}
func FindMinWithComparator(variableScope []*QualifierMapping, comparator func(a, b Qualifier) bool) *QualifierMapping {
	if len(variableScope) == 0 {
		return nil
	}
	min := variableScope[0]
	for _, val := range variableScope {
		if comparator(Qualifier(val.QualifierId), Qualifier(min.QualifierId)) {
			min = val
		}
	}
	return min
}

func GetPriority(qualifier Qualifier) int {
	switch qualifier {
	case GLOBAL_QUALIFIER:
		return 5
	default:
		return 0
	}
}
