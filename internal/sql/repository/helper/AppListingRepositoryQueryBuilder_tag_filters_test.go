package helper

import (
	"go.uber.org/zap"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, err)

	assert.Contains(t, whereClause, "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value = ?)")
	assert.Contains(t, whereClause, "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value NOT LIKE ? ESCAPE '\\')")
	assert.Contains(t, whereClause, "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ?)")
	assert.Contains(t, whereClause, "NOT EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ?)")
	assert.Len(t, queryParams, 8)
	assert.Equal(t, true, queryParams[0])
	assert.Equal(t, CustomApp, queryParams[1])
	assert.Equal(t, "owner", queryParams[2])
	assert.Equal(t, "James", queryParams[3])
	assert.Equal(t, "env", queryParams[4])
	assert.Equal(t, "%pro\\_d\\%%", queryParams[5])
	assert.Equal(t, "team", queryParams[6])
	assert.Equal(t, "zone", queryParams[7])
}

func TestBuildTagFilterPredicate_DoesNotEqualRequiresKeyAndDifferentValue(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	value := "mayank"

	predicate, queryParams, err := queryBuilder.buildTagFilterPredicate(TagFilter{
		Key:      "owner",
		Operator: TagFilterOperatorDoesNotEqual,
		Value:    &value,
	})
	assert.NoError(t, err)

	assert.Equal(t, "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value <> ?)", predicate)
	assert.Equal(t, []interface{}{"owner", "mayank"}, queryParams)
}

func TestBuildTagFilterPredicate_DoesNotContainRequiresKeyAndNotLike(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	value := "may"

	predicate, queryParams, err := queryBuilder.buildTagFilterPredicate(TagFilter{
		Key:      "owner",
		Operator: TagFilterOperatorDoesNotContain,
		Value:    &value,
	})
	assert.NoError(t, err)

	assert.Equal(t, "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value NOT LIKE ? ESCAPE '\\')", predicate)
	assert.Equal(t, []interface{}{"owner", "%may%"}, queryParams)
}

func TestBuildTagFilterPredicate_InvalidOperatorReturnsError(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())
	value := "mayank"

	predicate, queryParams, err := queryBuilder.buildTagFilterPredicate(TagFilter{
		Key:      "owner",
		Operator: TagFilterOperator("INVALID"),
		Value:    &value,
	})

	assert.Error(t, err)
	assert.Empty(t, predicate)
	assert.Nil(t, queryParams)
}

func TestBuildTagFiltersWhereConditionAND_NilFiltersReturnsNoClauseAndNoParams(t *testing.T) {
	queryBuilder := NewAppListingRepositoryQueryBuilder(zap.NewNop().Sugar())

	whereClause, queryParams, err := queryBuilder.buildTagFiltersWhereConditionAND(nil)

	assert.NoError(t, err)
	assert.Empty(t, whereClause)
	assert.NotNil(t, queryParams)
	assert.Len(t, queryParams, 0)
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

	assert.NoError(t, err)
	assert.Contains(t, whereClause, "a.app_name like ?")
	assert.Contains(t, whereClause, "and EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value = ?)")
	assert.Len(t, queryParams, 5)
	assert.Equal(t, true, queryParams[0])
	assert.Equal(t, CustomApp, queryParams[1])
	assert.Equal(t, "%demo%", queryParams[2])
	assert.Equal(t, "owner", queryParams[3])
	assert.Equal(t, "James", queryParams[4])
}
