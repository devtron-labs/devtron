/*
 * Copyright (c) 2024. Devtron Inc.
 */

package restHandler

import (
	"net/http"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/overview"
	"github.com/devtron-labs/devtron/pkg/overview/bean"
	"github.com/devtron-labs/devtron/pkg/overview/cache"
	"github.com/gorilla/schema"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type InfraOverviewRestHandler interface {
	GetClusterOverview(w http.ResponseWriter, r *http.Request)
	DeleteClusterOverviewCache(w http.ResponseWriter, r *http.Request)
	RefreshClusterOverviewCache(w http.ResponseWriter, r *http.Request)
	GetClusterOverviewDetailedNodeInfo(w http.ResponseWriter, r *http.Request)
}

type InfraOverviewRestHandlerImpl struct {
	logger                 *zap.SugaredLogger
	clusterOverviewService overview.ClusterOverviewService
	clusterCacheService    cache.ClusterCacheService
	userService            user.UserService
	validator              *validator.Validate
	enforcer               casbin.Enforcer
}

func NewInfraOverviewRestHandlerImpl(
	logger *zap.SugaredLogger,
	clusterOverviewService overview.ClusterOverviewService,
	clusterCacheService cache.ClusterCacheService,
	userService user.UserService,
	validator *validator.Validate,
	enforcer casbin.Enforcer,
) *InfraOverviewRestHandlerImpl {
	return &InfraOverviewRestHandlerImpl{
		logger:                 logger,
		clusterOverviewService: clusterOverviewService,
		clusterCacheService:    clusterCacheService,
		userService:            userService,
		validator:              validator,
		enforcer:               enforcer,
	}
}

// GetClusterOverview handles cluster management overview requests
func (handler *InfraOverviewRestHandlerImpl) GetClusterOverview(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	result, err := handler.clusterOverviewService.GetClusterOverview(r.Context())
	if err != nil {
		handler.logger.Errorw("error in getting cluster overview", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

// DeleteClusterOverviewCache handles cluster overview cache deletion requests
func (handler *InfraOverviewRestHandlerImpl) DeleteClusterOverviewCache(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	handler.clusterCacheService.InvalidateClusterOverview()
	handler.logger.Infow("cluster overview cache deleted successfully", "userId", userId)
	common.WriteJsonResp(w, nil, map[string]string{"message": "Cluster overview cache deleted successfully"}, http.StatusOK)
}

// RefreshClusterOverviewCache handles cluster overview cache refresh requests
func (handler *InfraOverviewRestHandlerImpl) RefreshClusterOverviewCache(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	err = handler.clusterOverviewService.RefreshClusterOverviewCache(r.Context())
	if err != nil {
		handler.logger.Errorw("error in refreshing cluster overview cache", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	handler.logger.Infow("cluster overview cache refreshed successfully", "userId", userId)
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

func (handler *InfraOverviewRestHandlerImpl) GetClusterOverviewDetailedNodeInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	var request bean.ClusterOverviewDetailRequest
	decoder := schema.NewDecoder()
	if err := decoder.Decode(&request, r.URL.Query()); err != nil {
		handler.logger.Errorw("error in decoding request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// Validate the request
	if err := handler.validator.Struct(request); err != nil {
		handler.logger.Errorw("validation error", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	result, err := handler.clusterOverviewService.GetClusterOverviewDetailedNodeInfo(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in getting cluster overview detail", "err", err, "groupBy", request.GroupBy)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}
