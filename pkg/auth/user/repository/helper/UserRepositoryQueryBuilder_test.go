/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package helper

import (
	"testing"

	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/stretchr/testify/assert"
)

func TestGetQueryForUserListingWithFilters(t *testing.T) {
	t.Run("QueryWithSearchKeyAndPagination", func(t *testing.T) {
		req := &bean2.ListingRequest{
			SearchKey:  "devtron",
			SortBy:     bean2.Email,
			SortOrder:  bean2.Asc,
			Size:       20,
			Offset:     0,
			CountCheck: false,
		}

		query, params := GetQueryForUserListingWithFilters(req)
		assert.Contains(t, query, "where active = true")
		assert.Contains(t, query, "email_id ilike ?")
		assert.Contains(t, query, "order by  email_id")
		assert.Contains(t, query, "limit ? offset ?")
		assert.Equal(t, 3, len(params))
		assert.Equal(t, "%devtron%", params[0])
		assert.Equal(t, 20, params[1])
		assert.Equal(t, 0, params[2])
	})

	t.Run("CountCheckQuery", func(t *testing.T) {
		req := &bean2.ListingRequest{
			SearchKey:  "test",
			CountCheck: true,
		}

		query, params := GetQueryForUserListingWithFilters(req)
		assert.Contains(t, query, "select count(*)")
		assert.Contains(t, query, "email_id ilike ?")
		assert.NotContains(t, query, "limit ? offset ?")
		assert.Equal(t, 1, len(params))
		assert.Equal(t, "%test%", params[0])
	})
}

func TestGetQueryForGroupListingWithFilters(t *testing.T) {
	t.Run("GroupListingWithSearchKeyAndSorting", func(t *testing.T) {
		req := &bean2.ListingRequest{
			SearchKey:  "admin-group",
			SortBy:     bean2.GroupName,
			SortOrder:  bean2.Desc,
			Size:       10,
			Offset:     5,
			CountCheck: false,
		}

		query, params := GetQueryForGroupListingWithFilters(req)
		assert.Contains(t, query, "where active = ?")
		assert.Contains(t, query, "name ilike ?")
		assert.Contains(t, query, "order by name DESC")
		assert.Contains(t, query, "limit ? offset ?")
		assert.Equal(t, 4, len(params))
		assert.Equal(t, true, params[0])
		assert.Equal(t, "%admin-group%", params[1])
		assert.Equal(t, 10, params[2])
		assert.Equal(t, 5, params[3])
	})
}
