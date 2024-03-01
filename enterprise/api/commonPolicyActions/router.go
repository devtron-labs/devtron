package commonPolicyActions

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/policyGovernance"
	"github.com/gorilla/mux"
)

type CommonPolicyRouter interface {
	InitCommonPolicyRouter(router *mux.Router)
}

type CommonPolicyRouterImpl struct {
	restHandler CommonPolicyRestHandler
}

func NewCommonPolicyRouterImpl(restHandler CommonPolicyRestHandler) *CommonPolicyRouterImpl {
	return &CommonPolicyRouterImpl{
		restHandler: restHandler,
	}
}

func (r *CommonPolicyRouterImpl) InitCommonPolicyRouter(router *mux.Router) {
	router.Path(fmt.Sprintf("/policy/{%s}/app-env/list", policyGovernance.PathVariablePolicyTypeVariable)).HandlerFunc(r.restHandler.ListAppEnvPolicies).
		Methods("POST")
	router.Path(fmt.Sprintf("/policy/{%s}/bulk/apply", policyGovernance.PathVariablePolicyTypeVariable)).HandlerFunc(r.restHandler.ApplyPolicyToIdentifiers).
		Methods("POST")
}
