/*
 * Copyright (c) 2024. Devtron Inc.
 */

package read

import (
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
)

func parsePromotionPolicyFromGlobalPolicy(globalPolicies []*bean2.GlobalPolicyBaseModel) ([]*bean.PromotionPolicy, error) {
	promotionPolicies := make([]*bean.PromotionPolicy, 0)
	for _, rawPolicy := range globalPolicies {
		policy := &bean.PromotionPolicy{}
		err := rawPolicy.ParseJsonInto(policy)
		if err != nil {
			return nil, err
		}
		policy.Id = rawPolicy.Id
		promotionPolicies = append(promotionPolicies, policy)
	}
	return promotionPolicies, nil
}
