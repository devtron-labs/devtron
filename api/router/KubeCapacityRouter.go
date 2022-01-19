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

type KubeCapacityRouter interface {
	InitKubeCapacityRouter(configRouter *mux.Router)
}
type KubeCapacityRouterImpl struct {
	kubeCapacityRestHandler restHandler.KubeCapacityRestHandler
}

func NewKubeCapacityRouterImpl(kubeCapacityRestHandler restHandler.KubeCapacityRestHandler) *KubeCapacityRouterImpl {
	return &KubeCapacityRouterImpl{
		kubeCapacityRestHandler: kubeCapacityRestHandler,
	}
}
func (impl KubeCapacityRouterImpl) InitKubeCapacityRouter(configRouter *mux.Router) {
	configRouter.Path("/default").HandlerFunc(impl.kubeCapacityRestHandler.KubeCapacityDefault).Methods("GET")
	configRouter.Path("/pods").HandlerFunc(impl.kubeCapacityRestHandler.KubeCapacityPods).Methods("GET")
	configRouter.Path("/util").HandlerFunc(impl.kubeCapacityRestHandler.KubeCapacityUtilization).Methods("GET")
	configRouter.Path("/available").HandlerFunc(impl.kubeCapacityRestHandler.AvailableResources).Methods("GET")
	configRouter.Path("/pods-util").HandlerFunc(impl.kubeCapacityRestHandler.PodsAndUtil).Methods("GET")
	configRouter.Path("/get-nodes").HandlerFunc(impl.kubeCapacityRestHandler.GetNodes).Methods("GET")
	configRouter.Path("/get-pods").HandlerFunc(impl.kubeCapacityRestHandler.GetPodsOfNodes).Methods("GET")
}
