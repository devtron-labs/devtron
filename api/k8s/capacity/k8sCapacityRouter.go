/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package capacity

import (
	"github.com/gorilla/mux"
)

type K8sCapacityRouter interface {
	InitK8sCapacityRouter(helmRouter *mux.Router)
}
type K8sCapacityRouterImpl struct {
	k8sCapacityRestHandler K8sCapacityRestHandler
}

func NewK8sCapacityRouterImpl(k8sCapacityRestHandler K8sCapacityRestHandler) *K8sCapacityRouterImpl {
	return &K8sCapacityRouterImpl{
		k8sCapacityRestHandler: k8sCapacityRestHandler,
	}
}

func (impl *K8sCapacityRouterImpl) InitK8sCapacityRouter(k8sCapacityRouter *mux.Router) {
	k8sCapacityRouter.Path("/cluster/list/raw").
		HandlerFunc(impl.k8sCapacityRestHandler.GetClusterListRaw).Methods("GET")

	k8sCapacityRouter.Path("/cluster/list").
		HandlerFunc(impl.k8sCapacityRestHandler.GetClusterListWithDetail).Methods("GET")

	k8sCapacityRouter.Path("/cluster/{clusterId}").
		HandlerFunc(impl.k8sCapacityRestHandler.GetClusterDetail).Methods("GET")

	k8sCapacityRouter.Path("/node/list").
		HandlerFunc(impl.k8sCapacityRestHandler.GetNodeList).Methods("GET")

	k8sCapacityRouter.Path("/node").
		HandlerFunc(impl.k8sCapacityRestHandler.GetNodeDetail).Methods("GET")

	k8sCapacityRouter.Path("/node").
		HandlerFunc(impl.k8sCapacityRestHandler.UpdateNodeManifest).Methods("PUT")

	k8sCapacityRouter.Path("/node").
		HandlerFunc(impl.k8sCapacityRestHandler.DeleteNode).Methods("DELETE")

	k8sCapacityRouter.Path("/node/cordon").
		HandlerFunc(impl.k8sCapacityRestHandler.CordonOrUnCordonNode).Methods("PUT")

	k8sCapacityRouter.Path("/node/drain").
		HandlerFunc(impl.k8sCapacityRestHandler.DrainNode).Methods("PUT")

	k8sCapacityRouter.Path("/node/taints/edit").
		HandlerFunc(impl.k8sCapacityRestHandler.EditNodeTaints).Methods("PUT")
}
