package helper

import (
	"go.uber.org/zap"
	"testing"

	"github.com/stretchr/testify/require"
)

func stringPointer(value string) *string {
	return &value
}

func TestBuildAppListingWhereCondition_WithTagFiltersAnd(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	tagFilters := []TagFilter{
		{Key: "owner", Operator: TagFilterOperatorEquals, Value: stringPointer("James")},
		{Key: "env", Operator: TagFilterOperatorDoesNotContain, Value: stringPointer("pro_d%")},
		{Key: "team", Operator: TagFilterOperatorExists, Value: nil},
		{Key: "zone", Operator: TagFilterOperatorDoesNotExist, Value: nil},
	}
	whereClause, queryParams, err := queryBuilder.buildAppListingWhereCondition(AppListingFilter{
		TagFilters: &tagFilters,
	})
	require.NoError(t, err)

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

func TestBuildTagFilterPredicate_DoesNotEqualRequiresKeyAndDifferentValue(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	value := "mayank"

	predicate, queryParams, err := queryBuilder.buildTagFilterPredicate(TagFilter{
		Key:      "owner",
		Operator: TagFilterOperatorDoesNotEqual,
		Value:    &value,
	})
	require.NoError(t, err)

	require.Equal(t, "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value <> ?)", predicate)
	require.Equal(t, []interface{}{"owner", "mayank"}, queryParams)
}

func TestBuildTagFilterPredicate_DoesNotContainRequiresKeyAndNotLike(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	value := "may"

	predicate, queryParams, err := queryBuilder.buildTagFilterPredicate(TagFilter{
		Key:      "owner",
		Operator: TagFilterOperatorDoesNotContain,
		Value:    &value,
	})
	require.NoError(t, err)

	require.Equal(t, "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value NOT LIKE ? ESCAPE '\\')", predicate)
	require.Equal(t, []interface{}{"owner", "%may%"}, queryParams)
}

func TestBuildTagFilterPredicate_InvalidOperatorReturnsError(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	value := "mayank"

	predicate, queryParams, err := queryBuilder.buildTagFilterPredicate(TagFilter{
		Key:      "owner",
		Operator: TagFilterOperator("INVALID"),
		Value:    &value,
	})

	require.Error(t, err)
	require.Empty(t, predicate)
	require.Nil(t, queryParams)
}

func TestBuildTagFiltersWhereConditionAND_NilFiltersReturnsNoClauseAndNoParams(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())

	whereClause, queryParams, err := queryBuilder.buildTagFiltersWhereConditionAND(nil)

	require.NoError(t, err)
	require.Empty(t, whereClause)
	require.NotNil(t, queryParams)
	require.Len(t, queryParams, 0)
}

func TestBuildAppListingWhereCondition_AppNameAndTagFiltersAreAndCombined(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	tagFilters := []TagFilter{
		{Key: "owner", Operator: TagFilterOperatorEquals, Value: stringPointer("James")},
	}

	whereClause, queryParams, err := queryBuilder.buildAppListingWhereCondition(AppListingFilter{
		AppNameSearch: "demo",
		TagFilters:    &tagFilters,
	})

	require.NoError(t, err)
	require.Contains(t, whereClause, "a.app_name like ?")
	require.Contains(t, whereClause, "and EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value = ?)")
	require.Len(t, queryParams, 5)
	require.Equal(t, true, queryParams[0])
	require.Equal(t, CustomApp, queryParams[1])
	require.Equal(t, "%demo%", queryParams[2])
	require.Equal(t, "owner", queryParams[3])
	require.Equal(t, "James", queryParams[4])
}
