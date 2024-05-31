/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type BatchOperationRouter interface {
	initBatchOperationRouter(router *mux.Router)
}

type BatchOperationRouterImpl struct {
	handler restHandler.BatchOperationRestHandler
	logger  *zap.SugaredLogger
}

func NewBatchOperationRouterImpl(handler restHandler.BatchOperationRestHandler, logger *zap.SugaredLogger) *BatchOperationRouterImpl {
	return &BatchOperationRouterImpl{
		handler: handler,
		logger:  logger,
	}
}

func (r BatchOperationRouterImpl) initBatchOperationRouter(router *mux.Router) {
	router.Path("/operate").
		Methods("POST").
		HandlerFunc(r.handler.Operate)
}
