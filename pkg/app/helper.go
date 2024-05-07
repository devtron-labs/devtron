package app

// MergeChildMapToParentMap merges child map of generic type map into parent map of generic type
// and returns merged mapping, if parentMap is nil then nil is returned.
func MergeChildMapToParentMap[T comparable](parentMap map[T]T, toMergeMap map[T]T) map[T]T {
	if parentMap == nil {
		return nil
	}
	for key, value := range toMergeMap {
		if _, ok := parentMap[key]; !ok {
			parentMap[key] = value
		}
	}
	return parentMap
}
