/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type PolicyRouter interface {
	InitPolicyRouter(configRouter *mux.Router)
}
type PolicyRouterImpl struct {
	policyRestHandler restHandler.PolicyRestHandler
}

func NewPolicyRouterImpl(policyRestHandler restHandler.PolicyRestHandler) *PolicyRouterImpl {
	return &PolicyRouterImpl{
		policyRestHandler: policyRestHandler,
	}
}
func (impl PolicyRouterImpl) InitPolicyRouter(configRouter *mux.Router) {
	configRouter.Path("/save").HandlerFunc(impl.policyRestHandler.SavePolicy).Methods("POST")
	configRouter.Path("/update").HandlerFunc(impl.policyRestHandler.UpdatePolicy).Methods("POST")
	configRouter.Path("/list").HandlerFunc(impl.policyRestHandler.GetPolicy).Methods("GET")
	configRouter.Path("/verify/webhook").HandlerFunc(impl.policyRestHandler.VerifyImage).Methods("POST")
}
