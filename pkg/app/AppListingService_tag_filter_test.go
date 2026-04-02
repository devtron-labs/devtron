package app

import (
	"testing"

	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/stretchr/testify/assert"
)

func strPointer(value string) *string {
	return &value
}

func TestValidateTagFilters_EqualsRequiresValue(t *testing.T) {
	err := ValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperatorEquals, Value: nil},
	})

	assert.Error(t, err)
	assert.Equal(t, "tagFilters[0].value is required for operator EQUALS", err.Error())
}

func TestValidateTagFilters_EqualsRejectsEmptyString(t *testing.T) {
	err := ValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperatorEquals, Value: strPointer("")},
	})

	assert.Error(t, err)
	assert.Equal(t, "tagFilters[0].value is required for operator EQUALS", err.Error())
}

func TestValidateTagFilters_ContainsRequiresValue(t *testing.T) {
	err := ValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperatorContains, Value: nil},
	})

	assert.Error(t, err)
	assert.Equal(t, "tagFilters[0].value is required for operator CONTAINS", err.Error())
}

func TestValidateTagFilters_EmptyKeyReturnsError(t *testing.T) {
	err := ValidateTagFilters([]helper.TagFilter{
		{Key: " ", Operator: helper.TagFilterOperatorEquals, Value: strPointer("James")},
	})

	assert.Error(t, err)
	assert.Equal(t, "tagFilters[0].key is required", err.Error())
}

func TestValidateTagFilters_InvalidOperatorReturnsError(t *testing.T) {
	err := ValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperator("INVALID"), Value: strPointer("James")},
	})

	assert.Error(t, err)
	assert.Equal(t, "tagFilters[0].operator is invalid: INVALID", err.Error())
}

func TestValidateTagFilters_ExistsAllowsNilValueOnly(t *testing.T) {
	err := ValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperatorExists, Value: nil},
	})

	assert.NoError(t, err)
}

func TestValidateTagFilters_ExistsRejectsProvidedValue(t *testing.T) {
	err := ValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperatorExists, Value: strPointer("James")},
	})

	assert.Error(t, err)
	assert.Equal(t, "tagFilters[0].value must be empty for operator EXISTS", err.Error())
}

func TestValidateTagFilters_DoesNotExistRejectsProvidedValue(t *testing.T) {
	err := ValidateTagFilters([]helper.TagFilter{
		{Key: "owner", Operator: helper.TagFilterOperatorDoesNotExist, Value: strPointer("")},
	})

	assert.Error(t, err)
	assert.Equal(t, "tagFilters[0].value must be empty for operator DOES_NOT_EXIST", err.Error())
}

func TestNormalizeTagFilters_TrimsKey(t *testing.T) {
	filters := []helper.TagFilter{
		{Key: " owner ", Operator: helper.TagFilterOperatorEquals, Value: strPointer("James")},
	}

	normalizedFilters := NormalizeTagFilters(filters)

	assert.Len(t, normalizedFilters, 1)
	assert.Equal(t, "owner", normalizedFilters[0].Key)
	// Ensure input is not modified by normalization.
	assert.Equal(t, " owner ", filters[0].Key)
}
