package k8s

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

	k8sCapacityRouter.Path("/cluster/list").
		HandlerFunc(impl.k8sCapacityRestHandler.GetClusterList).Methods("GET")

	k8sCapacityRouter.Path("/cluster/{clusterId}").
		HandlerFunc(impl.k8sCapacityRestHandler.GetClusterDetail).Methods("GET")

	k8sCapacityRouter.Path("/node/list").
		HandlerFunc(impl.k8sCapacityRestHandler.GetNodeList).Methods("GET")

	k8sCapacityRouter.Path("/node").
		HandlerFunc(impl.k8sCapacityRestHandler.GetNodeDetail).Methods("GET")

	k8sCapacityRouter.Path("/node").
		HandlerFunc(impl.k8sCapacityRestHandler.UpdateNodeManifest).Methods("PUT")
}
