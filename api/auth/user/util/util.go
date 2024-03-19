package util

func IsGroupsPresent(groups []string) bool {
	if len(groups) > 0 {
		return true
	}
	return false
}
