package app

// MergeChildMapWithParentMap merges child map of type map[string]string into parent map of type map[string]string
// and returns merged mapping if parentMap is nil then nil is returned.
func MergeChildMapWithParentMap(parentMap map[string]string, toMergeMap map[string]string) map[string]string {
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
