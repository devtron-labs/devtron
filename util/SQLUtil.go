package util

import "fmt"

func GetLIKEClauseQueryParam(s string) string {
	return fmt.Sprintf("%%%s%%", s)
}
