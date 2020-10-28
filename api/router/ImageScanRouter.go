/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
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
