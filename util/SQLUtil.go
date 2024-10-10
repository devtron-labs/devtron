package util

import "fmt"

// GetLIKEClauseQueryParam converts string "abc" into "%abc%".
// This is used for SQL queries and we have taken this approach instead of ex- .Where("name = %s%", "abc") because
// it will result into query : where name = %'abc'% since string params are added with quotes.
func GetLIKEClauseQueryParam(s string) string {
	return fmt.Sprintf("%%%s%%", s)
}

func GetCopyByValueObject[T any](input []T) []T {
	res := make([]T, 0, len(input))
	for _, item := range input {
		res = append(res, item)
	}
	return res
}
