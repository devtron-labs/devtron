/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type ReleaseMetricsRouter interface {
	initReleaseMetricsRouter(router *mux.Router)
}

type ReleaseMetricsRouterImpl struct {
	logger                    *zap.SugaredLogger
	releaseMetricsRestHandler restHandler.ReleaseMetricsRestHandler
}

func NewReleaseMetricsRouterImpl(logger *zap.SugaredLogger,
	releaseMetricsRestHandler restHandler.ReleaseMetricsRestHandler) *ReleaseMetricsRouterImpl {
	return &ReleaseMetricsRouterImpl{
		logger:                    logger,
		releaseMetricsRestHandler: releaseMetricsRestHandler,
	}
}

func (impl ReleaseMetricsRouterImpl) initReleaseMetricsRouter(router *mux.Router) {
	router.Path("/reset-app-environment").
		HandlerFunc(impl.releaseMetricsRestHandler.ResetDataForAppEnvironment).
		Methods("POST")
	router.Path("/reset-all-app-environment").
		HandlerFunc(impl.releaseMetricsRestHandler.ResetDataForAllAppEnvironment).
		Methods("POST")
	router.Path("/").
		HandlerFunc(impl.releaseMetricsRestHandler.GetDeploymentMetrics).
		Methods("GET")
}
