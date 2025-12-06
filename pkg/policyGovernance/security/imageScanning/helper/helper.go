package helper

import "strings"

func GetSeverityToSkipMap(severityList []string) map[string]bool {
	severityToSkipMap := make(map[string]bool, len(severityList))

	for _, severity := range severityList {
		if _, ok := severityToSkipMap[strings.ToLower(severity)]; !ok {
			severityToSkipMap[severity] = true
		}
	}
	return severityToSkipMap
}
