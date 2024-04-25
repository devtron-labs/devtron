package scoop

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"net/http"
)

type RestHandler interface {
	HandleInterceptedEvent(w http.ResponseWriter, r *http.Request)
}

type RestHandlerImpl struct {
	service Service
}

func NewRestHandler(service Service) *RestHandlerImpl {
	return &RestHandlerImpl{
		service: service,
	}
}

func (handler *RestHandlerImpl) HandleInterceptedEvent(w http.ResponseWriter, r *http.Request) {
	// token := r.Header.Get("token")
	// if ok := handler.handleRbac(r, w, request, token, casbin.ActionUpdate); !ok {
	// 	common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
	// 	return
	// }

	decoder := json.NewDecoder(r.Body)
	var interceptedEvent = &InterceptedEvent{}
	err := decoder.Decode(interceptedEvent)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	err = handler.service.HandleInterceptedEvent(r.Context(), interceptedEvent)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
