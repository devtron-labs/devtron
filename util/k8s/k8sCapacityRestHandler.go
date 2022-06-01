package k8s

import (
	"go.uber.org/zap"
	"net/http"
)

type K8sCapacityRestHandler interface {
	GetClusterList(w http.ResponseWriter, r *http.Request)
	GetClusterDetail(w http.ResponseWriter, r *http.Request)
	GetNodeList(w http.ResponseWriter, r *http.Request)
	GetNodeDetail(w http.ResponseWriter, r *http.Request)
	GetNodeManifest(w http.ResponseWriter, r *http.Request)
	UpdateNodeManifest(w http.ResponseWriter, r *http.Request)
}
type K8sCapacityRestHandlerImpl struct {
	logger             *zap.SugaredLogger
	k8sCapacityService K8sCapacityService
}

func NewK8sCapacityRestHandlerImpl(logger *zap.SugaredLogger,
	k8sCapacityService K8sCapacityService) *K8sCapacityRestHandlerImpl {
	return &K8sCapacityRestHandlerImpl{
		logger:             logger,
		k8sCapacityService: k8sCapacityService,
	}
}

func (handler *K8sApplicationRestHandlerImpl) GetClusterList(w http.ResponseWriter, r *http.Request) {

}

func (handler *K8sApplicationRestHandlerImpl) GetClusterDetail(w http.ResponseWriter, r *http.Request) {

}

func (handler *K8sApplicationRestHandlerImpl) GetNodeList(w http.ResponseWriter, r *http.Request) {

}

func (handler *K8sApplicationRestHandlerImpl) GetNodeDetail(w http.ResponseWriter, r *http.Request) {

}

func (handler *K8sApplicationRestHandlerImpl) GetNodeManifest(w http.ResponseWriter, r *http.Request) {

}

func (handler *K8sApplicationRestHandlerImpl) UpdateNodeManifest(w http.ResponseWriter, r *http.Request) {

}
