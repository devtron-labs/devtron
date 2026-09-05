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
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateBulkDeleteRequest(t *testing.T) {
	t.Run("NilRequest", func(t *testing.T) {
		err := ValidateBulkDeleteRequest(nil)
		assert.NotNil(t, err)
	})

	t.Run("EmptyRequest", func(t *testing.T) {
		req := &bean.BulkDeleteRequest{
			Ids:            []int32{},
			ListingRequest: nil,
		}
		err := ValidateBulkDeleteRequest(req)
		assert.NotNil(t, err)
	})

	t.Run("ValidIdsRequest", func(t *testing.T) {
		req := &bean.BulkDeleteRequest{
			Ids:            []int32{10, 11, 12},
			ListingRequest: nil,
		}
		err := ValidateBulkDeleteRequest(req)
		assert.Nil(t, err)
	})

	t.Run("ValidListingFilterRequest", func(t *testing.T) {
		req := &bean.BulkDeleteRequest{
			ListingRequest: &bean.ListingRequest{
				SearchKey: "test",
				Size:      20,
			},
		}
		err := ValidateBulkDeleteRequest(req)
		assert.Nil(t, err)
	})
}
