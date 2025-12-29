/*
 * Copyright (c) 2024. Devtron Inc.
 */

package restHandler

import (
	"fmt"
	"net/http"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/overview"
	"github.com/devtron-labs/devtron/pkg/overview/bean"
	"github.com/devtron-labs/devtron/pkg/overview/constants"
	"github.com/devtron-labs/devtron/pkg/overview/util"
	"github.com/gorilla/schema"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type OverviewRestHandler interface {
	GetAppsOverview(w http.ResponseWriter, r *http.Request)
	GetWorkflowOverview(w http.ResponseWriter, r *http.Request)
	GetBuildDeploymentActivity(w http.ResponseWriter, r *http.Request)
	GetBuildDeploymentActivityDetailed(w http.ResponseWriter, r *http.Request)
	GetDoraMetrics(w http.ResponseWriter, r *http.Request)
	GetInsights(w http.ResponseWriter, r *http.Request)

	// Cluster Management Overview
	GetClusterOverview(w http.ResponseWriter, r *http.Request)
	DeleteClusterOverviewCache(w http.ResponseWriter, r *http.Request)
	RefreshClusterOverviewCache(w http.ResponseWriter, r *http.Request)

	// Cluster Overview Detailed Drill-down API (unified endpoint)
	GetClusterOverviewDetailedNodeInfo(w http.ResponseWriter, r *http.Request)

	// Security Overview APIs
	GetSecurityOverview(w http.ResponseWriter, r *http.Request)
	GetSeverityInsights(w http.ResponseWriter, r *http.Request)
	GetDeploymentSecurityStatus(w http.ResponseWriter, r *http.Request)
	GetVulnerabilityTrend(w http.ResponseWriter, r *http.Request)
	GetBlockedDeploymentsTrend(w http.ResponseWriter, r *http.Request)
}

type OverviewRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	overviewService overview.OverviewService
	userService     user.UserService
	validator       *validator.Validate
	enforcer        casbin.Enforcer
}

func NewOverviewRestHandlerImpl(
	logger *zap.SugaredLogger,
	overviewService overview.OverviewService,
	userService user.UserService,
	validator *validator.Validate,
	enforcer casbin.Enforcer,
) *OverviewRestHandlerImpl {
	return &OverviewRestHandlerImpl{
		logger:          logger,
		overviewService: overviewService,
		userService:     userService,
		validator:       validator,
		enforcer:        enforcer,
	}
}

// validateTimeParameters validates that either timeWindow is provided or both from and to are provided
// Returns error if validation fails
func validateTimeParameters(timeWindow, from, to string) error {
	hasTimeWindow := len(timeWindow) > 0
	hasFromTo := len(from) > 0 && len(to) > 0

	if !hasTimeWindow && !hasFromTo {
		return fmt.Errorf("either timeWindow or both from/to parameters must be provided")
	}

	return nil
}

func (handler *OverviewRestHandlerImpl) GetAppsOverview(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	result, err := handler.overviewService.GetAppsOverview(r.Context())
	if err != nil {
		handler.logger.Errorw("error in getting apps overview", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

func (handler *OverviewRestHandlerImpl) GetWorkflowOverview(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	result, err := handler.overviewService.GetWorkflowOverview(r.Context())
	if err != nil {
		handler.logger.Errorw("error in getting workflow overview", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

func (handler *OverviewRestHandlerImpl) GetBuildDeploymentActivity(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// Extract query parameters
	timeWindow := r.URL.Query().Get("timeWindow")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	// Validate time parameters
	if err := validateTimeParameters(timeWindow, from, to); err != nil {
		handler.logger.Errorw("validation error for time parameters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// Parse from and to parameters
	request, err := util.GetCurrentTimePeriodBasedOnTimeWindow(timeWindow, from, to)
	if err != nil {
		handler.logger.Errorw("error in parsing request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	buildDeploymentRequest := &bean.BuildDeploymentActivityRequest{
		From: request.From,
		To:   request.To,
	}

	if err := handler.validator.Struct(buildDeploymentRequest); err != nil {
		handler.logger.Errorw("validation error", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	result, err := handler.overviewService.GetBuildDeploymentActivity(r.Context(), buildDeploymentRequest)
	if err != nil {
		handler.logger.Errorw("error in getting build deployment activity", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

func (handler *OverviewRestHandlerImpl) GetBuildDeploymentActivityDetailed(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// Extract query parameters
	timeWindow := r.URL.Query().Get("timeWindow")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	// Validate time parameters
	if err := validateTimeParameters(timeWindow, from, to); err != nil {
		handler.logger.Errorw("validation error for time parameters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	activityKind := r.URL.Query().Get("activityKind")
	if activityKind == "" {
		handler.logger.Errorw("activityKind query parameter is required")
		common.WriteJsonResp(w, fmt.Errorf("activityKind query parameter is required"), nil, http.StatusBadRequest)
		return
	}

	request, err := util.GetCurrentTimePeriodBasedOnTimeWindow(timeWindow, from, to)
	if err != nil {
		handler.logger.Errorw("error in parsing request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	aggregationType := constants.GetAggregationType(constants.TimePeriod(timeWindow))

	buildDeploymentDetailedRequest := &bean.BuildDeploymentActivityDetailedRequest{
		ActivityKind:    bean.ActivityKind(activityKind),
		AggregationType: aggregationType,
		From:            request.From,
		To:              request.To,
	}

	if err := handler.validator.Struct(buildDeploymentDetailedRequest); err != nil {
		handler.logger.Errorw("validation error", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	result, err := handler.overviewService.GetBuildDeploymentActivityDetailed(r.Context(), buildDeploymentDetailedRequest)
	if err != nil {
		handler.logger.Errorw("error in getting build deployment activity detailed", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

func (handler *OverviewRestHandlerImpl) GetDoraMetrics(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// Extract query parameters
	timeWindow := r.URL.Query().Get("timeWindow")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	// Validate time parameters
	if err := validateTimeParameters(timeWindow, from, to); err != nil {
		handler.logger.Errorw("validation error for time parameters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// Get both current and previous time ranges
	currentTimeWindow, prevTimeWindow, err := util.GetCurrentAndPreviousTimeRangeBasedOnTimeWindow(timeWindow, from, to)
	if err != nil {
		handler.logger.Errorw("error in parsing time periods", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	doraRequest := &bean.DoraMetricsRequest{
		TimeRangeRequest: currentTimeWindow,
		PrevFrom:         prevTimeWindow.From,
		PrevTo:           prevTimeWindow.To,
	}

	if err := handler.validator.Struct(doraRequest); err != nil {
		handler.logger.Errorw("validation error", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	result, err := handler.overviewService.GetDoraMetrics(r.Context(), doraRequest)
	if err != nil {
		handler.logger.Errorw("error in getting DORA metrics", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

func (handler *OverviewRestHandlerImpl) GetInsights(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// Extract query parameters
	timeWindow := r.URL.Query().Get("timeWindow")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	// Validate time parameters
	if err := validateTimeParameters(timeWindow, from, to); err != nil {
		handler.logger.Errorw("validation error for time parameters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	request, err := util.GetCurrentTimePeriodBasedOnTimeWindow(timeWindow, from, to)
	if err != nil {
		handler.logger.Errorw("error in parsing request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// Extract new query parameters
	pipelineTypeStr := r.URL.Query().Get("pipelineType")
	if pipelineTypeStr == "" {
		handler.logger.Errorw("pipelineType parameter is required")
		common.WriteJsonResp(w, fmt.Errorf("pipelineType parameter is required"), nil, http.StatusBadRequest)
		return
	}

	// Validate pipelineType
	var pipelineType bean.PipelineType
	switch pipelineTypeStr {
	case string(bean.BuildPipelines):
		pipelineType = bean.BuildPipelines
	case string(bean.DeploymentPipelines):
		pipelineType = bean.DeploymentPipelines
	default:
		handler.logger.Errorw("invalid pipelineType parameter", "pipelineType", pipelineTypeStr)
		common.WriteJsonResp(w, fmt.Errorf("invalid pipelineType parameter. Must be 'buildPipelines' or 'deploymentPipelines'"), nil, http.StatusBadRequest)
		return
	}

	sortOrderStr := r.URL.Query().Get("sortOrder")
	if sortOrderStr == "" {
		sortOrderStr = string(bean.DESC) // Default to DESC
	}

	// Validate sortOrder
	var sortOrder bean.SortOrder
	switch sortOrderStr {
	case string(bean.ASC):
		sortOrder = bean.ASC
	case string(bean.DESC):
		sortOrder = bean.DESC
	default:
		handler.logger.Errorw("invalid sortOrder parameter", "sortOrder", sortOrderStr)
		common.WriteJsonResp(w, fmt.Errorf("invalid sortOrder parameter. Must be 'ASC' or 'DESC'"), nil, http.StatusBadRequest)
		return
	}

	limit, err := common.ExtractIntQueryParam(w, r, "limit", 10)
	if err != nil {
		handler.logger.Errorw("error in parsing limit parameter", "err", err)
		return
	}

	offset, err := common.ExtractIntQueryParam(w, r, "offset", 0)
	if err != nil {
		handler.logger.Errorw("error in parsing offset parameter", "err", err)
		return
	}

	insightsRequest := &bean.InsightsRequest{
		TimeRangeRequest: request,
		PipelineType:     pipelineType,
		SortOrder:        sortOrder,
		Limit:            limit,
		Offset:           offset,
	}

	if err := handler.validator.Struct(insightsRequest); err != nil {
		handler.logger.Errorw("validation error", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	result, err := handler.overviewService.GetInsights(r.Context(), insightsRequest)
	if err != nil {
		handler.logger.Errorw("error in getting insights", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

// GetClusterOverview handles cluster management overview requests
func (handler *OverviewRestHandlerImpl) GetClusterOverview(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	result, err := handler.overviewService.GetClusterOverview(r.Context())
	if err != nil {
		handler.logger.Errorw("error in getting cluster overview", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

// DeleteClusterOverviewCache handles cluster overview cache deletion requests
func (handler *OverviewRestHandlerImpl) DeleteClusterOverviewCache(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	err = handler.overviewService.DeleteClusterOverviewCache(r.Context())
	if err != nil {
		handler.logger.Errorw("error in deleting cluster overview cache", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	handler.logger.Infow("cluster overview cache deleted successfully", "userId", userId)
	common.WriteJsonResp(w, nil, map[string]string{"message": "Cluster overview cache deleted successfully"}, http.StatusOK)
}

// RefreshClusterOverviewCache handles cluster overview cache refresh requests
func (handler *OverviewRestHandlerImpl) RefreshClusterOverviewCache(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	err = handler.overviewService.RefreshClusterOverviewCache(r.Context())
	if err != nil {
		handler.logger.Errorw("error in refreshing cluster overview cache", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	handler.logger.Infow("cluster overview cache refreshed successfully", "userId", userId)
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

func (handler *OverviewRestHandlerImpl) GetClusterOverviewDetailedNodeInfo(w http.ResponseWriter, r *http.Request) {
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

	result, err := handler.overviewService.GetClusterOverviewDetailedNodeInfo(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in getting cluster overview detail", "err", err, "groupBy", request.GroupBy)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

// ============================================================================
// Security Overview APIs
// ============================================================================

func (handler *OverviewRestHandlerImpl) GetSecurityOverview(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	var request bean.SecurityOverviewRequest
	if err := decoder.Decode(&request, r.URL.Query()); err != nil {
		handler.logger.Errorw("error in decoding request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	result, err := handler.overviewService.GetSecurityOverview(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in getting security overview", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

func (handler *OverviewRestHandlerImpl) GetSeverityInsights(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	var request bean.SeverityInsightsRequest
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

	result, err := handler.overviewService.GetSeverityInsights(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in getting severity insights", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

func (handler *OverviewRestHandlerImpl) GetDeploymentSecurityStatus(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	var request bean.DeploymentSecurityStatusRequest
	if err := decoder.Decode(&request, r.URL.Query()); err != nil {
		handler.logger.Errorw("error in decoding request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	result, err := handler.overviewService.GetDeploymentSecurityStatus(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in getting deployment security status", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

func (handler *OverviewRestHandlerImpl) GetVulnerabilityTrend(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	timeWindow := r.URL.Query().Get("timeWindow")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	envType := r.URL.Query().Get("envType")

	// Validate time parameters
	if err := validateTimeParameters(timeWindow, from, to); err != nil {
		handler.logger.Errorw("validation error for time parameters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// Validate envType
	if envType != string(bean.EnvTypeProd) && envType != string(bean.EnvTypeNonProd) && envType != string(bean.EnvTypeAll) {
		handler.logger.Errorw("invalid envType", "envType", envType)
		common.WriteJsonResp(w, fmt.Errorf("envType must be 'prod', 'non-prod' or 'all'"), nil, http.StatusBadRequest)
		return
	}

	// Get both current and previous time ranges
	currentTimeWindow, _, err := util.GetCurrentAndPreviousTimeRangeBasedOnTimeWindow(timeWindow, from, to)
	if err != nil {
		handler.logger.Errorw("error in parsing time periods", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// Determine aggregation type based on time range
	timePeriod := util.GetTimePeriodFromTimeRange(currentTimeWindow.From, currentTimeWindow.To)
	aggregationType := constants.GetAggregationType(timePeriod)

	result, err := handler.overviewService.GetVulnerabilityTrend(r.Context(), currentTimeWindow, bean.EnvType(envType), aggregationType)
	if err != nil {
		handler.logger.Errorw("error in getting vulnerability trend", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

func (handler *OverviewRestHandlerImpl) GetBlockedDeploymentsTrend(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	timeWindow := r.URL.Query().Get("timeWindow")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	// Validate time parameters
	if err := validateTimeParameters(timeWindow, from, to); err != nil {
		handler.logger.Errorw("validation error for time parameters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// Get current time range
	currentTimeWindow, err := util.GetCurrentTimePeriodBasedOnTimeWindow(timeWindow, from, to)
	if err != nil {
		handler.logger.Errorw("error in parsing time period", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// Determine aggregation type based on time range
	timePeriod := util.GetTimePeriodFromTimeRange(currentTimeWindow.From, currentTimeWindow.To)
	aggregationType := constants.GetAggregationType(timePeriod)

	result, err := handler.overviewService.GetBlockedDeploymentsTrend(r.Context(), currentTimeWindow, aggregationType)
	if err != nil {
		handler.logger.Errorw("error in getting blocked deployments trend", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, result, http.StatusOK)
}
