/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type CommonRouter interface {
	InitCommonRouter(router *mux.Router)
}
type CommonRouterImpl struct {
	commonRestHandler restHandler.CommonRestHandler
}

func NewCommonRouterImpl(commonRestHandler restHandler.CommonRestHandler) *CommonRouterImpl {
	return &CommonRouterImpl{commonRestHandler: commonRestHandler}
}
func (impl CommonRouterImpl) InitCommonRouter(router *mux.Router) {
	router.Path("/checklist").
		HandlerFunc(impl.commonRestHandler.GlobalChecklist).
		Methods("GET")
	router.Path("/environment-variables").
		HandlerFunc(impl.commonRestHandler.EnvironmentVariableList).
		Methods("GET")
}
