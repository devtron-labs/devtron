package policyGovernance

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v9"
	"testing"
)

func TestAppEnvPolicyMappingsListFilterStructValidation(t *testing.T) {
	validator := validator.New()
	t.Run("default values", func(tt *testing.T) {
		filter := AppEnvPolicyMappingsListFilter{}
		err := validator.Struct(&filter)
		assert.Nil(tt, err)
	})

	t.Run("invalid sortBy", func(tt *testing.T) {
		filter := AppEnvPolicyMappingsListFilter{
			SortOrder: "desc",
			SortBy:    "pipelineName",
			Offset:    -1,
			Size:      -1,
		}
		err := validator.Struct(&filter)
		assert.NotNil(tt, err)
	})

}

func TestBulkPromotionPolicyApplyRequestStructValidation(t *testing.T) {
	validator := validator.New()
	t.Run("default zero values", func(tt *testing.T) {
		request := BulkPromotionPolicyApplyRequest{}

		err := validator.Struct(&request)
		assert.NotNil(tt, err)
	})

	t.Run("default zero values with invalid ApplyToPolicyName", func(tt *testing.T) {
		request := BulkPromotionPolicyApplyRequest{
			ApplyToPolicyName: "tt",
		}
		err := validator.Struct(&request)
		assert.NotNil(tt, err)
	})

	t.Run("default zero values with valid ApplyToPolicyName", func(tt *testing.T) {
		request := BulkPromotionPolicyApplyRequest{
			ApplyToPolicyName: "test",
		}
		err := validator.Struct(&request)
		assert.Nil(tt, err)
	})

	t.Run("invalid sortBy", func(tt *testing.T) {
		filter := AppEnvPolicyMappingsListFilter{
			SortOrder: "desc",
			SortBy:    "pipelineName",
			Offset:    -1,
			Size:      -1,
		}
		request := BulkPromotionPolicyApplyRequest{
			ApplyToPolicyName:      "test",
			AppEnvPolicyListFilter: filter,
		}
		err := validator.Struct(&request)
		assert.NotNil(tt, err)
	})

}
