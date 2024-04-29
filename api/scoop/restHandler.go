package scoop

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/autoRemediation"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type RestHandler interface {
	HandleInterceptedEvent(w http.ResponseWriter, r *http.Request)
}

type RestHandlerImpl struct {
	logger         *zap.SugaredLogger
	watcherService autoRemediation.WatcherService
	service        Service
}

func NewRestHandler(service Service, watcherService autoRemediation.WatcherService) *RestHandlerImpl {
	return &RestHandlerImpl{
		service:        service,
		watcherService: watcherService,
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

func (handler *RestHandlerImpl) GetWatchersByClusterId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterId, err := strconv.Atoi(vars["clusterId"])
	if err != nil {
		handler.logger.Errorw("error in getting clusterId from query param", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// this is an safe api, currently there is no RBAC applied here
	handler.watcherService.GetWatchersByClusterId(clusterId)
}
