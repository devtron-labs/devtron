package util

import "fmt"

func ProcessAppStatuses(appStatuses []string) string {
	query := ""
	n := len(appStatuses)
	for i, status := range appStatuses {
		query += fmt.Sprintf("'%s'", status)
		if i < n-1 {
			query += ","
		}
	}

	return query
}
