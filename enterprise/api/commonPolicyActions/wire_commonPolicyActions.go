package commonPolicyActions

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance"
	"github.com/google/wire"
)

var CommonPolicyActionWireSet = wire.NewSet(
	policyGovernance.NewCommonPolicyActionsService,
	wire.Bind(new(policyGovernance.CommonPolicyActionsService), new(*policyGovernance.CommonPolicyActionsServiceImpl)),
	NewCommonPolicyRestHandlerImpl,
	wire.Bind(new(CommonPolicyRestHandler), new(*CommonPolicyRestHandlerImpl)),
	NewCommonPolicyRouterImpl,
	wire.Bind(new(CommonPolicyRouter), new(*CommonPolicyRouterImpl)),
)
