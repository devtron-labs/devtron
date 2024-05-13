package release

import "github.com/google/wire"

var ReleasePolicyWireSet = wire.NewSet(
	NewPolicyEvaluationServiceImpl,
	wire.Bind(new(PolicyEvaluationService), new(*PolicyEvaluationServiceImpl)),
)
