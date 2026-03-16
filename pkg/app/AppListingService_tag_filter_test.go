package app

import (
	"testing"

	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/stretchr/testify/require"
)

func strPointer(value string) *string {
	return &value
}

func TestNormalizeAndValidateTagFilters_EqualsRequiresValue(t *testing.T) {
	_, err := NormalizeAndValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperatorEquals, Value: nil},
	})

	require.Error(t, err)
	require.Equal(t, "tagFilters[0].value is required for operator EQUALS", err.Error())
}

func TestNormalizeAndValidateTagFilters_EqualsRejectsEmptyString(t *testing.T) {
	_, err := NormalizeAndValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperatorEquals, Value: strPointer("")},
	})

	require.Error(t, err)
	require.Equal(t, "tagFilters[0].value is required for operator EQUALS", err.Error())
}

func TestNormalizeAndValidateTagFilters_ContainsRequiresValue(t *testing.T) {
	_, err := NormalizeAndValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperatorContains, Value: nil},
	})

	require.Error(t, err)
	require.Equal(t, "tagFilters[0].value is required for operator CONTAINS", err.Error())
}

func TestNormalizeAndValidateTagFilters_EmptyKeyReturnsError(t *testing.T) {
	_, err := NormalizeAndValidateTagFilters([]helper.TagFilter{
		{Key: " ", Operator: helper.TagFilterOperatorEquals, Value: strPointer("James")},
	})

	require.Error(t, err)
	require.Equal(t, "tagFilters[0].key is required", err.Error())
}

func TestNormalizeAndValidateTagFilters_InvalidOperatorReturnsError(t *testing.T) {
	_, err := NormalizeAndValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperator("INVALID"), Value: strPointer("James")},
	})

	require.Error(t, err)
	require.Equal(t, "tagFilters[0].operator is invalid: INVALID", err.Error())
}

func TestNormalizeAndValidateTagFilters_ExistsAllowsNilValueOnly(t *testing.T) {
	normalizedFilters, err := NormalizeAndValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperatorExists, Value: nil},
	})

	require.NoError(t, err)
	require.Len(t, normalizedFilters, 1)
	require.Nil(t, normalizedFilters[0].Value)
}

func TestNormalizeAndValidateTagFilters_ExistsRejectsProvidedValue(t *testing.T) {
	_, err := NormalizeAndValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperatorExists, Value: strPointer("James")},
	})

	require.Error(t, err)
	require.Equal(t, "tagFilters[0].value must be empty for operator EXISTS", err.Error())
}

func TestNormalizeAndValidateTagFilters_DoesNotExistRejectsProvidedValue(t *testing.T) {
	_, err := NormalizeAndValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperatorDoesNotExist, Value: strPointer("")},
	})

	require.Error(t, err)
	require.Equal(t, "tagFilters[0].value must be empty for operator DOES_NOT_EXIST", err.Error())
}
