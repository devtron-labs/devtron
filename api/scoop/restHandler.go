package scoop

import (
	"context"
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/autoRemediation"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/devtron-labs/devtron/util/response"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type RestHandler interface {
	HandleInterceptedEvent(w http.ResponseWriter, r *http.Request)
	GetWatchersByClusterId(w http.ResponseWriter, r *http.Request)
	HandleNotificationEvent(w http.ResponseWriter, r *http.Request)
}

type RestHandlerImpl struct {
	logger         *zap.SugaredLogger
	watcherService autoRemediation.WatcherService
	service        Service
	enforcerUtil   rbac.EnforcerUtil
	enforcer       casbin.Enforcer
}

func NewRestHandler(service Service, watcherService autoRemediation.WatcherService,
	enforcerUtil rbac.EnforcerUtil,
	enforcer casbin.Enforcer) *RestHandlerImpl {
	return &RestHandlerImpl{
		service:        service,
		watcherService: watcherService,
		enforcerUtil:   enforcerUtil,
		enforcer:       enforcer,
	}
}

func (handler *RestHandlerImpl) HandleInterceptedEvent(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceJobs, casbin.ActionTrigger, "*")
	if !isSuperAdmin {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
	}

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

	token := r.Header.Get("token")
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if !isSuperAdmin {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
	}

	vars := mux.Vars(r)
	clusterId, err := strconv.Atoi(vars["clusterId"])
	if err != nil {
		handler.logger.Errorw("error in getting clusterId from query param", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	watchers, err := handler.watcherService.GetWatchersByClusterId(clusterId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, watchers, http.StatusOK)
}

func (handler *RestHandlerImpl) HandleNotificationEvent(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var notification map[string]interface{}
	err := decoder.Decode(&notification)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	clusterIdStr := r.URL.Query().Get("clusterId")
	clusterId, err := strconv.Atoi(clusterIdStr)
	if err != nil {
		handler.logger.Errorw("error in getting clusterId from query param", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = handler.service.HandleNotificationEvent(context.Background(), clusterId, notification)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
