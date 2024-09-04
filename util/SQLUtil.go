package util

import "fmt"

// GetLIKEClauseQueryParam converts string "abc" into "%abc%".
//This is used for SQL queries and we have taken this approach instead of ex- .Where("name = %s%", "abc") because
//it will result into query : where name = %'abc'% since string params are added with quotes.
func GetLIKEClauseQueryParam(s string) string {
	return fmt.Sprintf("%%%s%%", s)
}
