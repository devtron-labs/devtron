/*
 * Copyright (c) 2024. Devtron Inc.
 */

package scoop

import (
	"context"
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/apiToken"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	user2 "github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/autoRemediation"
	"github.com/devtron-labs/devtron/pkg/cluster"
	types2 "github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/devtron-labs/scoop/types"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type RestHandler interface {
	HandleInterceptedEvent(w http.ResponseWriter, r *http.Request)
	HandleNotificationEvent(w http.ResponseWriter, r *http.Request)
	GetWatchersByClusterId(w http.ResponseWriter, r *http.Request)
	GetNamespacesByClusterId(w http.ResponseWriter, r *http.Request)
}

type RestHandlerImpl struct {
	logger             *zap.SugaredLogger
	watcherService     autoRemediation.WatcherService
	environmentService cluster.EnvironmentService
	service            Service
	enforcerUtil       rbac.EnforcerUtil
	enforcer           casbin.Enforcer
	userService        user2.UserService
	ciCdConfig         *types2.CiCdConfig
}

func NewRestHandler(service Service, watcherService autoRemediation.WatcherService,
	enforcerUtil rbac.EnforcerUtil,
	userService user2.UserService,
	logger *zap.SugaredLogger,
	enforcer casbin.Enforcer,
	ciCdConfig *types2.CiCdConfig,
	environmentService cluster.EnvironmentService,
) *RestHandlerImpl {
	return &RestHandlerImpl{
		service:            service,
		watcherService:     watcherService,
		enforcerUtil:       enforcerUtil,
		enforcer:           enforcer,
		userService:        userService,
		logger:             logger,
		ciCdConfig:         ciCdConfig,
		environmentService: environmentService,
	}
}

// HandleInterceptedEvent
// we are maintaining the same handler.ciCdConfig.OrchestratorToken at scoop.
func (handler *RestHandlerImpl) HandleInterceptedEvent(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	if token != handler.ciCdConfig.OrchestratorToken {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var interceptedEvent = &types.InterceptedEvent{}
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

// HandleNotificationEvent
// 1) scoop -> HandleInterceptedEvent
// 2) HandleInterceptedEvent -> creates a token with expiry and trigger a job with notification plugin configured
// 3) plugin -> HandleNotificationEvent. here we will verify the token
func (handler *RestHandlerImpl) HandleNotificationEvent(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	claimBytes, err := handler.userService.GetFieldValuesFromToken(token)
	if err != nil {
		handler.logger.Errorw("error in getting field values from token", "err", err)
		common.WriteJsonResp(w, errors.New("invalid token"), nil, http.StatusUnauthorized)
		return
	}

	tokenClaims := apiToken.ApiTokenCustomClaims{}
	err = json.Unmarshal(claimBytes, &tokenClaims)
	if err != nil {
		handler.logger.Errorw("error in un marshalling token claims", "claimBytes", claimBytes, "err", err)
		common.WriteJsonResp(w, errors.New("invalid token"), nil, http.StatusBadRequest)
		return
	}

	if expired := tokenClaims.ExpiresAt.Before(time.Now()); expired {
		common.WriteJsonResp(w, errors.New("token expired"), nil, http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var notification map[string]interface{}
	err = decoder.Decode(&notification)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = handler.service.HandleNotificationEvent(context.Background(), notification)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

// GetWatchersByClusterId
// we are maintaining the same handler.ciCdConfig.OrchestratorToken at scoop.
func (handler *RestHandlerImpl) GetWatchersByClusterId(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("token")
	if token != handler.ciCdConfig.OrchestratorToken {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
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

func (handler *RestHandlerImpl) GetNamespacesByClusterId(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	if token != handler.ciCdConfig.OrchestratorToken {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	clusterId, err := strconv.Atoi(vars["clusterId"])
	if err != nil {
		handler.logger.Errorw("error in getting clusterId from query param", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	results, err := handler.environmentService.GetCombinedEnvironmentListForDropDownByClusterIds(token, []int{clusterId}, func(token string, object string) bool {
		return true
	})

	if err != nil {
		common.WriteJsonResp(w, err, results, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, results, http.StatusOK)

}
