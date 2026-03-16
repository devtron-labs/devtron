package helper

import (
	"fmt"
	"go.uber.org/zap"
	"testing"

	"github.com/stretchr/testify/require"
)

func stringPointer(value string) *string {
	return &value
}

func TestBuildAppListingWhereCondition_WithTagFiltersAnd(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	whereClause, queryParams := queryBuilder.buildAppListingWhereCondition(AppListingFilter{
		TagFilters: []TagFilter{
			{Key: "owner", Operator: TagFilterOperatorEquals, Value: stringPointer("James")},
			{Key: "env", Operator: TagFilterOperatorDoesNotContain, Value: stringPointer("pro_d%")},
			{Key: "team", Operator: TagFilterOperatorExists, Value: nil},
			{Key: "zone", Operator: TagFilterOperatorDoesNotExist, Value: nil},
		},
	})

	require.Contains(t, whereClause, "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value = ?)")
	require.Contains(t, whereClause, "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value NOT LIKE ? ESCAPE '\\')")
	require.Contains(t, whereClause, "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ?)")
	require.Contains(t, whereClause, "NOT EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ?)")
	require.Len(t, queryParams, 8)
	require.Equal(t, true, queryParams[0])
	require.Equal(t, CustomApp, queryParams[1])
	require.Equal(t, "owner", queryParams[2])
	require.Equal(t, "James", queryParams[3])
	require.Equal(t, "env", queryParams[4])
	require.Equal(t, "%pro\\_d\\%%", queryParams[5])
	require.Equal(t, "team", queryParams[6])
	require.Equal(t, "zone", queryParams[7])
}

func TestBuildTagFiltersWhereConditionOR(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	whereClause, queryParams := queryBuilder.buildTagFiltersWhereConditionOR([]TagFilter{
		{Key: "owner", Operator: TagFilterOperatorEquals, Value: stringPointer("James")},
		{Key: "cost-center", Operator: TagFilterOperatorContains, Value: stringPointer("ENG")},
	})

	require.Equal(t, " and (EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value = ?) OR EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value LIKE ? ESCAPE '\\')) ", whereClause)
	require.Equal(t, []interface{}{"owner", "James", "cost-center", "%ENG%"}, queryParams)
}

func TestBuildTagFilterPredicate_DoesNotEqualRequiresKeyAndDifferentValue(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	value := "mayank"

	predicate, queryParams := queryBuilder.buildTagFilterPredicate(TagFilter{
		Key:      "owner",
		Operator: TagFilterOperatorDoesNotEqual,
		Value:    &value,
	})

	require.Equal(t, "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value <> ?)", predicate)
	require.Equal(t, []interface{}{"owner", "mayank"}, queryParams)
}

func TestBuildTagFilterPredicate_DoesNotContainRequiresKeyAndNotLike(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	value := "may"

	predicate, queryParams := queryBuilder.buildTagFilterPredicate(TagFilter{
		Key:      "owner",
		Operator: TagFilterOperatorDoesNotContain,
		Value:    &value,
	})

	require.Equal(t, "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value NOT LIKE ? ESCAPE '\\')", predicate)
	require.Equal(t, []interface{}{"owner", "%may%"}, queryParams)
}

func BenchmarkBuildAppListingQueryWithTagFilters(b *testing.B) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	tagFilters := make([]TagFilter, 0, 10)
	for i := 0; i < 10; i++ {
		value := fmt.Sprintf("value-%d", i)
		tagFilters = append(tagFilters, TagFilter{
			Key:      fmt.Sprintf("key-%d", i),
			Operator: TagFilterOperatorContains,
			Value:    &value,
		})
	}
	appListingFilter := AppListingFilter{
		TagFilters: tagFilters,
		SortBy:     AppNameSortBy,
		SortOrder:  Asc,
		Offset:     0,
		Size:       20,
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = queryBuilder.GetAppIdsQueryWithPaginationForAppNameSearch(appListingFilter)
	}
}
