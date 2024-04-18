package bean

import (
	"github.com/devtron-labs/devtron/internal/util"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolicyBean(t *testing.T) {
	policyBean1 := PromotionPolicy{
		Name:       "hello",
		Conditions: []util2.ResourceCondition{},
	}

	policyBean2 := PromotionPolicy{
		Name:       "   ",
		Conditions: nil,
	}
	val, _ := util.IntValidator()
	err := val.Struct(&policyBean1)
	assert.NotNil(t, err)

	err = val.Struct(&policyBean2)
	assert.NotNil(t, err)
}
