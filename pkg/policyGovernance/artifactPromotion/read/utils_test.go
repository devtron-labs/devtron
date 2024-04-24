package read

import (
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParsePromotionPolicyFromGlobalPolicy(t *testing.T) {
	t.Run("Pass case: valid jsons", func(tt *testing.T) {
		globalPolicies := []*bean.GlobalPolicyBaseModel{
			{
				JsonData: `{"id":0,"name":"test1","description":"imager builder cannot approve this artifact","conditions":[{"conditionType":1,"expression":"true"}],"approvalMetadata":{"approverCount":1,"allowImageBuilderFromApprove":false,"allowRequesterFromApprove":true,"allowApproverFromDeploy":true}}`,
			},

			{
				JsonData: `{"id":0,"name":"test2","description":"imager builder cannot approve this artifact","conditions":[{"conditionType":1,"expression":"true"}],"approvalMetadata":{"approverCount":1,"allowImageBuilderFromApprove":false,"allowRequesterFromApprove":true,"allowApproverFromDeploy":true}}`,
			},
		}

		policies, err := parsePromotionPolicyFromGlobalPolicy(globalPolicies)
		assert.Nil(t, err)
		assert.Equal(t, len(globalPolicies), len(policies))
	})

	t.Run("Fail case: inValid jsons", func(tt *testing.T) {
		globalPolicies := []*bean.GlobalPolicyBaseModel{
			{
				JsonData: `{"id":0,"name":"test1","description":"imager builder cannot approve this artifact","conditions":[{"conditionType":1,"expression":"true"}],"approvalMetadata":{"approverCount":"1","allowImageBuilderFromApprove":false,"allowRequesterFromApprove":true,"allowApproverFromDeploy":true}}`,
			},

			{
				JsonData: `{"id":0,"name":"test2","description":"imager builder cannot approve this artifact","conditions":[{"conditionType":1,"expression":"true"}],"approvalMetadata":{"approverCount":1,"allowImageBuilderFromApprove":"false","allowRequesterFromApprove":true,"allowApproverFromDeploy":true}}`,
			},
		}

		policies, err := parsePromotionPolicyFromGlobalPolicy(globalPolicies)
		assert.NotNil(t, err)
		assert.Nil(t, policies)
		assert.Equal(t, 0, len(policies))
	})
}
