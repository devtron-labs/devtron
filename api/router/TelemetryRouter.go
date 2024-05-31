/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type TelemetryRouter interface {
	InitTelemetryRouter(router *mux.Router)
}

type TelemetryRouterImpl struct {
	logger  *zap.SugaredLogger
	handler restHandler.TelemetryRestHandler
}

func NewTelemetryRouterImpl(logger *zap.SugaredLogger, handler restHandler.TelemetryRestHandler) *TelemetryRouterImpl {
	router := &TelemetryRouterImpl{
		logger:  logger,
		handler: handler,
	}
	return router
}

func (router TelemetryRouterImpl) InitTelemetryRouter(telemetryRouter *mux.Router) {
	telemetryRouter.Path("/meta").
		HandlerFunc(router.handler.GetTelemetryMetaInfo).Methods("GET")
	telemetryRouter.Path("/event").
		HandlerFunc(router.handler.SendTelemetryData).Methods("POST")
	telemetryRouter.Path("/summary").
		HandlerFunc(router.handler.SendSummaryEvent).Methods("POST")

}
