package restHandler

import (
	"go.uber.org/zap"
	"net/http"
)

type K8sApplicationRestHandler interface {
	GetResource(w http.ResponseWriter, r *http.Request)
	UpdateResource(w http.ResponseWriter, r *http.Request)
}

type K8sApplicationRestHandlerImpl struct {
	logger                 *zap.SugaredLogger
}

func NewK8sApplicationRestHandlerImpl(logger *zap.SugaredLogger,) *K8sApplicationRestHandlerImpl {
	return &K8sApplicationRestHandlerImpl{
		logger:                 logger,
	}
}


UpdateResource(w http.ResponseWriter, r *http.Request){

}