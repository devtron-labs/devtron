/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
