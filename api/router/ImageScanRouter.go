/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type ImageScanRouter interface {
	InitImageScanRouter(gocdRouter *mux.Router)
}
type ImageScanRouterImpl struct {
	imageScanRestHandler restHandler.ImageScanRestHandler
}

func NewImageScanRouterImpl(imageScanRestHandler restHandler.ImageScanRestHandler) *ImageScanRouterImpl {
	return &ImageScanRouterImpl{imageScanRestHandler: imageScanRestHandler}
}

func (impl ImageScanRouterImpl) InitImageScanRouter(configRouter *mux.Router) {
	configRouter.Path("/list").HandlerFunc(impl.imageScanRestHandler.ScanExecutionList).Methods("POST")

	//image=image:abc&envId=3&appId=100&artifactId=100&executionId=100
	configRouter.Path("/executionDetail").HandlerFunc(impl.imageScanRestHandler.FetchExecutionDetail).Methods("GET")
	configRouter.Path("/executionDetail/min").HandlerFunc(impl.imageScanRestHandler.FetchMinScanResultByAppIdAndEnvId).Methods("GET")

	configRouter.Path("/cve/exposure").HandlerFunc(impl.imageScanRestHandler.VulnerabilityExposure).Methods("POST")

}
