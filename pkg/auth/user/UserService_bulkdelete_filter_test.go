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

package user

import (
	"fmt"
	"strings"
	"testing"

	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository/helper"
	"github.com/stretchr/testify/assert"
)

// fakeUserRepoForFilterTest is a minimal repository.UserRepository fake that only answers
// GetAllExecutingQuery - the single repository call getUserIdsHonoringFilters makes. Every other
// method is left to panic on the embedded nil interface, since these tests never expect them to run.
type fakeUserRepoForFilterTest struct {
	repository.UserRepository
	called      bool
	gotQuery    string
	gotParams   []interface{}
	modelsToRet []repository.UserModel
}

func (f *fakeUserRepoForFilterTest) GetAllExecutingQuery(query string, queryParams []interface{}) ([]repository.UserModel, error) {
	f.called = true
	f.gotQuery = query
	f.gotParams = queryParams
	return f.modelsToRet, nil
}

func newTestUserServiceForFilterTest(repo *fakeUserRepoForFilterTest) *UserServiceImpl {
	return &UserServiceImpl{
		userRepository:    repo,
		userCommonService: &UserCommonServiceImpl{},
	}
}

// TestGetUserIdsHonoringFilters_RejectsUnscopedFilter is a regression test for the production incident
// where `DELETE /orchestrator/user/bulk` with body {"listingRequest":{"searchKey":"","status":["inactive"]}}
// (an empty search key, intended to be narrowed down by status) deleted every active user on the platform
// instead of only the intended inactive ones. The delete path resolved the filter directly through the
// query builder without the normalization/bound the listing endpoint applies, so an effectively empty
// filter silently widened into "every user". With the fix, a listingRequest carrying no concrete filter
// (status isn't a filterable field in this OSS build) is rejected outright instead of being resolved at all.
func TestGetUserIdsHonoringFilters_RejectsUnscopedFilter(t *testing.T) {
	repo := &fakeUserRepoForFilterTest{}
	impl := newTestUserServiceForFilterTest(repo)

	// mirrors the exact incident payload once decoded: SearchKey is empty and there is no other
	// concrete filter field available on this build's ListingRequest.
	request := &bean.ListingRequest{SearchKey: ""}

	ids, err := impl.getUserIdsHonoringFilters(request)

	assert.Error(t, err)
	assert.Nil(t, ids)
	assert.False(t, repo.called, "no query should ever be executed for an unscoped filter based bulk delete")
}

// TestGetUserIdsHonoringFilters_MatchesListingQuery asserts that once a concrete filter is supplied, the
// query getUserIdsHonoringFilters resolves is identical - query text and params - to the query the listing
// endpoint (GetAllWithFilters) would build for the exact same input, other than pagination (see
// TestGetUserIdsHonoringFilters_DoesNotPaginate). This is the core of the fix: delete-side
// filter resolution can no longer diverge from listing-side filter resolution (e.g. via a missing default
// page size), so the delete path's resolved id set only ever contains records actually matching the filter.
func TestGetUserIdsHonoringFilters_MatchesListingQuery(t *testing.T) {
	matchingUser := repository.UserModel{Id: 42, EmailId: "inactive-user@example.com"}
	repo := &fakeUserRepoForFilterTest{modelsToRet: []repository.UserModel{matchingUser}}
	impl := newTestUserServiceForFilterTest(repo)

	deleteRequest := &bean.ListingRequest{SearchKey: "inactive-user"}
	ids, err := impl.getUserIdsHonoringFilters(deleteRequest)
	assert.NoError(t, err)
	assert.Equal(t, []int32{42}, ids)
	assert.True(t, repo.called)

	// Build the query the listing endpoint would run for the exact same input, applying the exact
	// same normalization GetAllWithFilters applies before calling the shared query builder.
	// The listing endpoint paginates by default, whereas the delete path must intentionally NOT paginate
	// (a filter based bulk delete applies to every matching user, not one page), so the delete path's
	// request is the listing request with ShowAll forced on - nothing else may differ.
	listingRequest := &bean.ListingRequest{SearchKey: "inactive-user"}
	(&UserCommonServiceImpl{}).SetDefaultValuesIfNotPresent(listingRequest, false)
	setStatusFilterType(listingRequest)
	listingRequest.ShowAll = true
	listingQuery, listingParams := helper.GetQueryForUserListingWithFilters(listingRequest)

	assert.Equal(t, listingQuery, repo.gotQuery, "delete path query must match listing path query (minus pagination) for the same filter")
	assert.Equal(t, listingParams, repo.gotParams, "delete path query params must match listing path query params (minus pagination) for the same filter")
}

// TestGetUserIdsHonoringFilters_DoesNotPaginate is a regression test for a filter based bulk delete being
// silently capped to the first listing page. getUserIdsHonoringFilters applies the listing normalization,
// which defaults Size to bean.DefaultSize (20) when unset, and the query builder then appends
// "limit ? offset ?" unless ShowAll is set. A bulk delete such as {"listingRequest":{"searchKey":"abc"}}
// therefore resolved (and deleted) only the first 20 matching users and left the rest behind.
func TestGetUserIdsHonoringFilters_DoesNotPaginate(t *testing.T) {
	const matchingUsers = 25 // deliberately more than the default page size
	assert.Greater(t, matchingUsers, bean.DefaultSize, "test must use more matching users than the default page size")

	models := make([]repository.UserModel, 0, matchingUsers)
	wantIds := make([]int32, 0, matchingUsers)
	for i := 1; i <= matchingUsers; i++ {
		models = append(models, repository.UserModel{Id: int32(i), EmailId: fmt.Sprintf("abc-user-%d@example.com", i)})
		wantIds = append(wantIds, int32(i))
	}
	repo := &fakeUserRepoForFilterTest{modelsToRet: models}
	impl := newTestUserServiceForFilterTest(repo)

	request := &bean.ListingRequest{SearchKey: "abc"}
	ids, err := impl.getUserIdsHonoringFilters(request)

	assert.NoError(t, err)
	assert.True(t, repo.called)
	assert.Equal(t, wantIds, ids, "every matching user must be resolved for deletion, not just the first page")
	assert.True(t, request.ShowAll, "filter based bulk delete must force ShowAll so the query is not paginated")
	assert.NotContains(t, repo.gotQuery, "limit ? offset ?", "delete path query must not be paginated")
	// the only param is the search key pattern(s); size/offset must never be bound
	assert.NotContains(t, repo.gotParams, bean.DefaultSize, "page size must not be bound as a query param")

	// Proof the test would fail if the fix were reverted: the exact same input run through the same
	// normalization WITHOUT ShowAll (the pre-fix behaviour) produces a paginated query bound to the
	// default page size, which is what capped the delete to 20 users.
	preFix := &bean.ListingRequest{SearchKey: "abc"}
	(&UserCommonServiceImpl{}).SetDefaultValuesIfNotPresent(preFix, false)
	setStatusFilterType(preFix)
	assert.False(t, preFix.ShowAll)
	preFixQuery, preFixParams := helper.GetQueryForUserListingWithFilters(preFix)
	assert.Equal(t, bean.DefaultSize, preFix.Size)
	assert.True(t, strings.Contains(preFixQuery, "limit ? offset ?"), "pre-fix query paginates by default")
	assert.Contains(t, preFixParams, bean.DefaultSize, "pre-fix query binds the default page size")
	assert.NotEqual(t, preFixQuery, repo.gotQuery, "fixed query must differ from the paginated pre-fix query")
}
