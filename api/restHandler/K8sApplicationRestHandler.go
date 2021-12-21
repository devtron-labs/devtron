package restHandler

import (
	"go.uber.org/zap"
	"net/http"
)

type K8sApplicationRestHandler interface {
	GetResource(w http.ResponseWriter, r *http.Request)
	UpdateResource(w http.ResponseWriter, r *http.Request)
	DeleteResource(w http.ResponseWriter, r *http.Request)
}

type K8sApplicationRestHandlerImpl struct {
	logger                 *zap.SugaredLogger
}

func NewK8sApplicationRestHandlerImpl(logger *zap.SugaredLogger,) *K8sApplicationRestHandlerImpl {
	return &K8sApplicationRestHandlerImpl{
		logger:                 logger,
	}
}

func(impl K8sApplicationRestHandlerImpl) GetResource(w http.ResponseWriter, r *http.Request){

}

func(impl K8sApplicationRestHandlerImpl) UpdateResource(w http.ResponseWriter, r *http.Request){

}

func(impl K8sApplicationRestHandlerImpl) DeleteResource(w http.ResponseWriter, r *http.Request){

}