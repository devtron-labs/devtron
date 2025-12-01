/*
 * Copyright (c) 2024. Devtron Inc.
 */

package overview

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	deploymentConfigRepo "github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/overview/bean"
	"github.com/devtron-labs/devtron/pkg/overview/constants"
	overviewUtil "github.com/devtron-labs/devtron/pkg/overview/util"
	workflowStageRepository "github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/repository"
	teamRepository "github.com/devtron-labs/devtron/pkg/team/repository"
	"go.uber.org/zap"
)

type AppManagementService interface {
	GetAppsOverview(ctx context.Context) (*bean.AppsOverviewResponse, error)
	GetWorkflowOverview(ctx context.Context) (*bean.WorkflowOverviewResponse, error)
	GetBuildDeploymentActivity(ctx context.Context, request *bean.BuildDeploymentActivityRequest) (*bean.BuildDeploymentActivityResponse, error)
	GetBuildDeploymentActivityDetailed(ctx context.Context, request *bean.BuildDeploymentActivityDetailedRequest) (*bean.BuildDeploymentActivityDetailedResponse, error)
}

type AppManagementServiceImpl struct {
	logger                     *zap.SugaredLogger
	appRepository              app.AppRepository
	pipelineRepository         pipelineConfig.PipelineRepository
	ciPipelineRepository       pipelineConfig.CiPipelineRepository
	ciWorkflowRepository       pipelineConfig.CiWorkflowRepository
	cdWorkflowRepository       pipelineConfig.CdWorkflowRepository
	environmentRepository      repository.EnvironmentRepository
	teamRepository             teamRepository.TeamRepository
	workflowStageRepository    workflowStageRepository.WorkflowStageRepository
	deploymentConfigRepository deploymentConfigRepo.Repository
	trendCalculator            *overviewUtil.TrendCalculator
}

func NewAppManagementServiceImpl(
	logger *zap.SugaredLogger,
	appRepository app.AppRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	environmentRepository repository.EnvironmentRepository,
	teamRepository teamRepository.TeamRepository,
	workflowStageRepository workflowStageRepository.WorkflowStageRepository,
	deploymentConfigRepository deploymentConfigRepo.Repository,
) *AppManagementServiceImpl {
	return &AppManagementServiceImpl{
		logger:                     logger,
		appRepository:              appRepository,
		pipelineRepository:         pipelineRepository,
		ciPipelineRepository:       ciPipelineRepository,
		ciWorkflowRepository:       ciWorkflowRepository,
		cdWorkflowRepository:       cdWorkflowRepository,
		environmentRepository:      environmentRepository,
		teamRepository:             teamRepository,
		workflowStageRepository:    workflowStageRepository,
		deploymentConfigRepository: deploymentConfigRepository,
		trendCalculator:            overviewUtil.NewTrendCalculator(),
	}
}

func (impl *AppManagementServiceImpl) GetAppsOverview(ctx context.Context) (*bean.AppsOverviewResponse, error) {

	allProjects, err := impl.teamRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error getting all projects", "err", err)
		return nil, err
	}

	allDevtronApps, err := impl.appRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error getting all devtron apps", "err", err)
		return nil, err
	}

	allHelmApps, err := impl.appRepository.FindAllChartStoreApps()
	if err != nil {
		impl.logger.Errorw("error getting all helm apps", "err", err)
		return nil, err
	}

	allEnvironments, err := impl.environmentRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error getting all environments", "err", err)
		return nil, err
	}

	response := &bean.AppsOverviewResponse{
		Projects: &bean.AtAGlanceMetric{
			Total: len(allProjects),
		},
		YourApplications: &bean.AtAGlanceMetric{
			Total: len(allDevtronApps),
		},
		HelmApplications: &bean.AtAGlanceMetric{
			Total: len(allHelmApps),
		},
		Environments: &bean.AtAGlanceMetric{
			Total: len(allEnvironments),
		},
	}

	return response, nil
}

func (impl *AppManagementServiceImpl) GetWorkflowOverview(ctx context.Context) (*bean.WorkflowOverviewResponse, error) {
	allTimeMetrics, err := impl.fetchAllWorkflowMetrics(ctx)
	if err != nil {
		return nil, err
	}

	response := impl.buildWorkflowOverviewResponse(allTimeMetrics)
	return response, nil
}

func (impl *AppManagementServiceImpl) GetBuildDeploymentActivity(ctx context.Context, request *bean.BuildDeploymentActivityRequest) (*bean.BuildDeploymentActivityResponse, error) {
	impl.logger.Infow("getting build deployment activity overview", "request", request)

	// Get current period counts - now tracking only CI_BUILD pipeline builds (including failed ones)
	currentTotalBuilds, err := impl.ciWorkflowRepository.GetCiBuildCountInTimeRange(request.From, request.To)
	if err != nil {
		impl.logger.Errorw("error getting current total builds count", "err", err)
		return nil, err
	}

	// Get average build time (only for CI_BUILD pipelines)
	avgBuildTime, err := impl.calculateAverageBuildTime(request.From, request.To)
	if err != nil {
		impl.logger.Errorw("error calculating average build time", "err", err)
		// Don't fail the request, just set to 0
		avgBuildTime = 0
	}

	// Get current total deployments count - now tracking ALL triggered deployments (including failed ones)
	currentTotalDeployments, err := impl.cdWorkflowRepository.GetDeploymentCountInTimeRange(request.From, request.To)
	if err != nil {
		impl.logger.Errorw("error getting current total deployments count", "err", err)
		return nil, err
	}

	response := &bean.BuildDeploymentActivityResponse{
		TotalBuildTriggers:      currentTotalBuilds,
		AverageBuildTime:        avgBuildTime,
		TotalDeploymentTriggers: currentTotalDeployments,
	}

	return response, nil
}

func (impl *AppManagementServiceImpl) GetBuildDeploymentActivityDetailed(ctx context.Context, request *bean.BuildDeploymentActivityDetailedRequest) (*bean.BuildDeploymentActivityDetailedResponse, error) {
	impl.logger.Infow("getting build deployment activity detailed", "request", request)

	response := &bean.BuildDeploymentActivityDetailedResponse{
		ActivityKind:    request.ActivityKind,
		AggregationType: request.AggregationType,
	}

	// Based on activityKind, fetch only the requested data
	switch request.ActivityKind {
	case bean.ActivityKindBuildTrigger:
		buildTriggersTrend, err := impl.getAggregatedBuildStatusTrend(request.From, request.To, request.AggregationType)
		if err != nil {
			impl.logger.Errorw("error getting aggregated build status trend", "err", err)
			return nil, err
		}
		response.BuildTriggersTrend = buildTriggersTrend

	case bean.ActivityKindDeploymentTrigger:
		deploymentTriggersTrend, err := impl.getAggregatedDeploymentStatusTrend(request.From, request.To, request.AggregationType)
		if err != nil {
			impl.logger.Errorw("error getting aggregated deployment status trend", "err", err)
			return nil, err
		}
		response.DeploymentTriggersTrend = deploymentTriggersTrend

	case bean.ActivityKindAvgBuildTime:
		avgBuildTimeTrend, err := impl.getAggregatedBuildTimeTrend(request.From, request.To, request.AggregationType)
		if err != nil {
			impl.logger.Errorw("error getting aggregated build time trend", "err", err)
			return nil, err
		}
		response.AvgBuildTimeTrend = avgBuildTimeTrend

	default:
		return nil, fmt.Errorf("invalid activityKind: %s", request.ActivityKind)
	}

	return response, nil
}

func (impl *AppManagementServiceImpl) getProjectMetrics(ctx context.Context, from, to *time.Time) (*bean.ProjectMetrics, error) {
	teams, err := impl.teamRepository.FindAllActiveInTimeRange(from, to)
	if err != nil {
		impl.logger.Errorw("error in getting projects", "err", err)
		return nil, err
	}

	details := make([]bean.EntityMetadata, 0, len(teams))
	for _, team := range teams {
		details = append(details, bean.EntityMetadata{
			Name:      team.Name,
			CreatedOn: team.CreatedOn,
		})
	}

	return &bean.ProjectMetrics{
		Total:   len(teams),
		Details: details,
	}, nil
}

func (impl *AppManagementServiceImpl) getAppMetrics(ctx context.Context, from, to *time.Time) (*bean.AppMetrics, error) {
	// Get normal apps (CI/CD apps with appType = 0) with details in time range
	devtronApps, err := impl.appRepository.FindAllActiveDevtronAppsInTimeRange(from, to)
	if err != nil {
		impl.logger.Errorw("error in getting all devtron apps", "err", err)
		return nil, err
	}

	normalAppsDetails := make([]bean.EntityMetadata, 0, len(devtronApps))
	for _, app := range devtronApps {
		normalAppsDetails = append(normalAppsDetails, bean.EntityMetadata{
			Name:      app.AppName,
			CreatedOn: app.CreatedOn,
		})
	}

	// Get chart store apps (external apps with appType = 1) with details in time range
	chartStoreApps, err := impl.appRepository.FindAllActiveChartStoreAppsInTimeRange(from, to)
	if err != nil {
		impl.logger.Errorw("error in getting all chart store apps", "err", err)
		return nil, err
	}

	chartStoreAppsDetails := make([]bean.EntityMetadata, 0, len(chartStoreApps))
	for _, app := range chartStoreApps {
		chartStoreAppsDetails = append(chartStoreAppsDetails, bean.EntityMetadata{
			Name:      app.AppName,
			CreatedOn: app.CreatedOn,
		})
	}

	totalApps := len(devtronApps) + len(chartStoreApps)

	return &bean.AppMetrics{
		Total: totalApps,
		YourApps: &bean.AppTypeMetrics{
			Total:   len(devtronApps),
			Details: normalAppsDetails,
		},
		ThirdPartyApps: &bean.AppTypeMetrics{
			Total:   len(chartStoreApps),
			Details: chartStoreAppsDetails,
		},
	}, nil
}

func (impl *AppManagementServiceImpl) getEnvironmentMetrics(ctx context.Context, from, to *time.Time) (*bean.EnvironmentMetrics, error) {
	environments, err := impl.environmentRepository.FindAllActiveInTimeRange(from, to)
	if err != nil {
		impl.logger.Errorw("error in getting environments", "err", err)
		return nil, err
	}

	details := make([]bean.EntityMetadata, 0, len(environments))
	for _, env := range environments {
		details = append(details, bean.EntityMetadata{
			Name:      env.Name,
			CreatedOn: env.CreatedOn,
		})
	}

	return &bean.EnvironmentMetrics{
		Total:   len(environments),
		Details: details,
	}, nil
}

func (impl *AppManagementServiceImpl) getBuildPipelineMetrics(ctx context.Context, from, to *time.Time) (*bean.BuildPipelineMetrics, error) {
	// Get counts directly instead of fetching full structs
	normalCiCount, err := impl.ciPipelineRepository.GetActiveCiPipelineCountInTimeRange(from, to)
	if err != nil {
		impl.logger.Errorw("error getting normal CI pipelines count", "err", err)
		return nil, err
	}

	externalCiCount, err := impl.ciPipelineRepository.GetActiveExternalCiPipelineCountInTimeRange(from, to)
	if err != nil {
		impl.logger.Errorw("error getting external CI pipelines count", "err", err)
		return nil, err
	}

	// For details, we still need to fetch some data, but only if details are actually needed
	// For now, we'll provide empty details arrays since the main use case is just counts
	var normalPipelines []bean.EntityMetadata
	var externalPipelines []bean.EntityMetadata

	total := normalCiCount + externalCiCount

	return &bean.BuildPipelineMetrics{
		Total: total,
		NormalCiPipelines: &bean.CiPipelineTypeMetrics{
			Total:   normalCiCount,
			Details: normalPipelines, // Empty for performance - can be populated if needed
		},
		ExternalCiPipelines: &bean.CiPipelineTypeMetrics{
			Total:   externalCiCount,
			Details: externalPipelines, // Empty for performance - can be populated if needed
		},
	}, nil
}

func (impl *AppManagementServiceImpl) getCdPipelineMetrics(ctx context.Context, from, to *time.Time) (*bean.CdPipelineMetrics, error) {
	// Get counts directly instead of fetching full structs
	prodCount, err := impl.pipelineRepository.GetActivePipelineCountByEnvironmentTypeInTimeRange(true, from, to)
	if err != nil {
		impl.logger.Errorw("error getting production pipelines count", "err", err)
		return nil, err
	}

	nonProdCount, err := impl.pipelineRepository.GetActivePipelineCountByEnvironmentTypeInTimeRange(false, from, to)
	if err != nil {
		impl.logger.Errorw("error getting non-production pipelines count", "err", err)
		return nil, err
	}

	// For details, we provide empty arrays since the main use case is just counts
	var prodDetails []bean.EntityMetadata
	var nonProdDetails []bean.EntityMetadata

	total := prodCount + nonProdCount

	return &bean.CdPipelineMetrics{
		Total: total,
		Production: &bean.PipelineEnvironmentMetrics{
			Total:   prodCount,
			Details: prodDetails, // Empty for performance - can be populated if needed
		},
		NonProduction: &bean.PipelineEnvironmentMetrics{
			Total:   nonProdCount,
			Details: nonProdDetails, // Empty for performance - can be populated if needed
		},
	}, nil
}

func (impl *AppManagementServiceImpl) getEnvironmentTrendMetrics(ctx context.Context, from, to *time.Time, aggregationType constants.AggregationType) (*bean.EnvironmentTrendMetrics, error) {
	// Get aggregated environment trend data
	trendData, err := impl.environmentRepository.GetAggregatedEnvironmentTrendWithParams(from, to, aggregationType)
	if err != nil {
		impl.logger.Errorw("error getting environment trend data", "err", err)
		return nil, err
	}

	// Calculate total
	total := 0
	for _, data := range trendData {
		total += data.Count
	}

	return &bean.EnvironmentTrendMetrics{
		Total: total,
		Trend: trendData,
	}, nil
}

func (impl *AppManagementServiceImpl) getAggregatedBuildStatusTrend(from, to *time.Time, aggregationType constants.AggregationType) ([]bean.BuildStatusDataPoint, error) {
	workflows, err := impl.ciWorkflowRepository.GetCIBuildsForStatusTrend(from, to)
	if err != nil {
		impl.logger.Errorw("error fetching CI builds for status trend", "err", err)
		return nil, err
	}

	statusMap := make(map[string]map[string]int) // timeKey -> status -> count

	targetLocation := from.Location()

	for _, workflow := range workflows {
		// Convert UTC workflow.StartedOn to the target timezone for proper time bucketing
		localStartedOn := workflow.StartedOn.In(targetLocation)

		var timeKey string
		if aggregationType == constants.AggregateByHour {
			timeKey = localStartedOn.Truncate(time.Hour).Format("2006-01-02T15:04:05Z")
		} else if aggregationType == constants.AggregateByMonth {
			timeKey = time.Date(localStartedOn.Year(), localStartedOn.Month(), 1, 0, 0, 0, 0, targetLocation).Format("2006-01-02T15:04:05Z")
		} else {
			timeKey = localStartedOn.Truncate(24 * time.Hour).Format("2006-01-02T15:04:05Z")
		}

		if statusMap[timeKey] == nil {
			statusMap[timeKey] = make(map[string]int)
		}

		// Categorize status
		switch workflow.Status {
		case "Succeeded":
			statusMap[timeKey]["successful"]++
		case "Failed", "Error", "Cancelled", "CANCELLED":
			statusMap[timeKey]["failed"]++
		}
		statusMap[timeKey]["total"]++
	}
	var trendData []bean.BuildStatusDataPoint

	if aggregationType == constants.AggregateByHour {
		current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), 0, 0, 0, from.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Format("2006-01-02T15:04:05Z")
			counts := statusMap[timeKey]

			trendData = append(trendData, bean.BuildStatusDataPoint{
				Timestamp:  current,
				Total:      counts["total"],
				Successful: counts["successful"],
				Failed:     counts["failed"],
			})

			current = current.Add(time.Hour)
		}
	} else if aggregationType == constants.AggregateByMonth {
		// Generate monthly series
		current := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, to.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Format("2006-01-02T15:04:05Z")
			counts := statusMap[timeKey]

			trendData = append(trendData, bean.BuildStatusDataPoint{
				Timestamp:  current,
				Total:      counts["total"],
				Successful: counts["successful"],
				Failed:     counts["failed"],
			})

			current = current.AddDate(0, 1, 0) // Add one month
		}
	} else {
		// Generate daily series
		current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Format("2006-01-02T15:04:05Z")
			counts := statusMap[timeKey]

			trendData = append(trendData, bean.BuildStatusDataPoint{
				Timestamp:  current,
				Total:      counts["total"],
				Successful: counts["successful"],
				Failed:     counts["failed"],
			})

			current = current.AddDate(0, 0, 1) // Add one day
		}
	}

	return trendData, nil
}

// Helper method to get aggregated deployment status trend with success/failed breakdown
func (impl *AppManagementServiceImpl) getAggregatedDeploymentStatusTrend(from, to *time.Time, aggregationType constants.AggregationType) ([]bean.DeploymentStatusDataPoint, error) {
	// Fetch all deployment workflows in the date range from repository
	deployments, err := impl.cdWorkflowRepository.GetDeploymentWorkflowsForStatusTrend(from, to)
	if err != nil {
		impl.logger.Errorw("error fetching deployment workflows for status trend", "err", err)
		return nil, err
	}

	// Group deployments by time period and count statuses
	statusMap := make(map[string]map[string]int) // timeKey -> status -> count

	// Get the timezone from the from/to parameters for proper time bucketing
	targetLocation := from.Location()

	for _, deployment := range deployments {
		// Convert UTC deployment.StartedOn to the target timezone for proper time bucketing
		localStartedOn := deployment.StartedOn.In(targetLocation)

		var timeKey string
		if aggregationType == constants.AggregateByHour {
			timeKey = localStartedOn.Truncate(time.Hour).Format("2006-01-02T15:04:05Z")
		} else if aggregationType == constants.AggregateByMonth {
			timeKey = time.Date(localStartedOn.Year(), localStartedOn.Month(), 1, 0, 0, 0, 0, targetLocation).Format("2006-01-02T15:04:05Z")
		} else {
			timeKey = localStartedOn.Truncate(24 * time.Hour).Format("2006-01-02T15:04:05Z")
		}

		if statusMap[timeKey] == nil {
			statusMap[timeKey] = make(map[string]int)
		}

		// Categorize status
		switch deployment.Status {
		case "Succeeded":
			statusMap[timeKey]["successful"]++
		case "Failed", "Error", "Cancelled", "CANCELLED":
			statusMap[timeKey]["failed"]++
		}
		statusMap[timeKey]["total"]++
	}
	// Generate complete time series and populate with counts
	var trendData []bean.DeploymentStatusDataPoint

	if aggregationType == constants.AggregateByHour {
		current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), 0, 0, 0, from.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Format("2006-01-02T15:04:05Z")
			counts := statusMap[timeKey]

			trendData = append(trendData, bean.DeploymentStatusDataPoint{
				Timestamp:  current,
				Total:      counts["total"],
				Successful: counts["successful"],
				Failed:     counts["failed"],
			})

			current = current.Add(time.Hour)
		}
	} else if aggregationType == constants.AggregateByMonth {
		// Generate monthly series
		current := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, to.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Format("2006-01-02T15:04:05Z")
			counts := statusMap[timeKey]

			trendData = append(trendData, bean.DeploymentStatusDataPoint{
				Timestamp:  current,
				Total:      counts["total"],
				Successful: counts["successful"],
				Failed:     counts["failed"],
			})

			current = current.AddDate(0, 1, 0) // Add one month
		}
	} else {
		// Generate daily series
		current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Format("2006-01-02T15:04:05Z")
			counts := statusMap[timeKey]

			trendData = append(trendData, bean.DeploymentStatusDataPoint{
				Timestamp:  current,
				Total:      counts["total"],
				Successful: counts["successful"],
				Failed:     counts["failed"],
			})

			current = current.AddDate(0, 0, 1) // Add one day
		}
	}

	return trendData, nil
}

func (impl *AppManagementServiceImpl) getAggregatedBuildTimeTrend(from, to *time.Time, aggregationType constants.AggregationType) ([]bean.BuildTimeDataPoint, error) {
	workflows, err := impl.getSuccessfulBuildsFromStages(from, to)
	if err != nil {
		impl.logger.Errorw("error fetching successful workflows from stages", "err", err)
		// Fallback to original method if new method fails
		impl.logger.Infow("falling back to original method for build time trend calculation")
		workflows, err = impl.ciWorkflowRepository.GetSuccessfulCIBuildsForBuildTime(from, to)
		if err != nil {
			impl.logger.Errorw("error fetching successful workflows (fallback)", "err", err)
			return nil, err
		}
	}

	// Calculate build times and group by time period
	buildTimeMap := make(map[string][]float64)

	// Get the timezone from the from/to parameters for proper time bucketing
	targetLocation := from.Location()

	for _, workflow := range workflows {
		// Calculate build time in minutes
		duration := workflow.FinishedOn.Sub(workflow.StartedOn)
		buildTimeMinutes := duration.Minutes()

		// Convert UTC workflow.StartedOn to the target timezone for proper time bucketing
		localStartedOn := workflow.StartedOn.In(targetLocation)

		var timeKey string
		if aggregationType == constants.AggregateByHour {
			timeKey = time.Date(localStartedOn.Year(), localStartedOn.Month(), localStartedOn.Day(), localStartedOn.Hour(), 0, 0, 0, targetLocation).Format("2006-01-02T15:04:05Z")
		} else if aggregationType == constants.AggregateByMonth {
			timeKey = time.Date(localStartedOn.Year(), localStartedOn.Month(), 1, 0, 0, 0, 0, targetLocation).Format("2006-01-02T15:04:05Z")
		} else {
			// For daily aggregation, get midnight of the local date
			timeKey = time.Date(localStartedOn.Year(), localStartedOn.Month(), localStartedOn.Day(), 0, 0, 0, 0, targetLocation).Format("2006-01-02T15:04:05Z")
		}

		buildTimeMap[timeKey] = append(buildTimeMap[timeKey], buildTimeMinutes)
	}
	// Generate complete time series and calculate averages
	var trendData []bean.BuildTimeDataPoint

	if aggregationType == constants.AggregateByHour {
		current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), to.Day(), to.Hour(), 0, 0, 0, from.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Format("2006-01-02T15:04:05Z")
			avgBuildTime := 0.0

			if buildTimes, exists := buildTimeMap[timeKey]; exists && len(buildTimes) > 0 {
				sum := 0.0
				for _, bt := range buildTimes {
					sum += bt
				}
				avgBuildTime = sum / float64(len(buildTimes))
			}

			trendData = append(trendData, bean.BuildTimeDataPoint{
				Timestamp:        current,
				AverageBuildTime: math.Round(avgBuildTime*100) / 100, // Round to 2 decimal places
			})

			current = current.Add(time.Hour)
		}
	} else if aggregationType == constants.AggregateByMonth {
		// Generate monthly series
		current := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, to.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Format("2006-01-02T15:04:05Z")
			avgBuildTime := 0.0

			if buildTimes, exists := buildTimeMap[timeKey]; exists && len(buildTimes) > 0 {
				sum := 0.0
				for _, bt := range buildTimes {
					sum += bt
				}
				avgBuildTime = sum / float64(len(buildTimes))
			}

			trendData = append(trendData, bean.BuildTimeDataPoint{
				Timestamp:        current,
				AverageBuildTime: math.Round(avgBuildTime*100) / 100, // Round to 2 decimal places
			})

			current = current.AddDate(0, 1, 0) // Add one month
		}
	} else {
		// Generate daily series
		current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())

		for current.Before(end) || current.Equal(end) {
			timeKey := current.Format("2006-01-02T15:04:05Z")
			avgBuildTime := 0.0

			if buildTimes, exists := buildTimeMap[timeKey]; exists && len(buildTimes) > 0 {
				sum := 0.0
				for _, bt := range buildTimes {
					sum += bt
				}
				avgBuildTime = sum / float64(len(buildTimes))
			}

			trendData = append(trendData, bean.BuildTimeDataPoint{
				Timestamp:        current,
				AverageBuildTime: math.Round(avgBuildTime*100) / 100, // Round to 2 decimal places
			})

			current = current.AddDate(0, 0, 1) // Add one day
		}
	}

	return trendData, nil
}

// WorkflowMetrics holds aggregated workflow metrics for a time period
type WorkflowMetrics struct {
	BuildPipelinesCount       int // CI pipelines count
	ProductionPipelinesCount  int // Production deployment pipelines count
	NonProdPipelinesCount     int // Non-production deployment pipelines count
	ExternalCICount           int
	ScanningEnabledPercentage float64 // Percentage of build pipelines where scanning is enabled
	GitOpsComplianceCount     int     // Count of GitOps enabled pipelines
	GitOpsCoveragePercentage  float64 // Percentage of GitOps coverage
}

func (impl *AppManagementServiceImpl) fetchAllWorkflowMetrics(ctx context.Context) (*WorkflowMetrics, error) {
	buildPipelinesCount, err := impl.ciPipelineRepository.GetActiveCiPipelineCount()
	if err != nil {
		impl.logger.Errorw("error getting build pipelines count", "err", err)
		return nil, err
	}

	productionPipelinesCount, err := impl.pipelineRepository.GetPipelineCountByEnvironmentType(true)
	if err != nil {
		impl.logger.Errorw("error getting production pipelines count", "err", err)
		return nil, err
	}

	nonProdPipelinesCount, err := impl.pipelineRepository.GetPipelineCountByEnvironmentType(false)
	if err != nil {
		impl.logger.Errorw("error getting non-production pipelines count", "err", err)
		return nil, err
	}

	// Get scanning enabled count directly
	scanningEnabledCount, err := impl.ciPipelineRepository.GetScanEnabledCiPipelineCount()
	if err != nil {
		impl.logger.Errorw("error getting scanning enabled count", "err", err)
		return nil, err
	}

	// Calculate scanning enabled percentage
	var scanningEnabledPercentage float64
	if buildPipelinesCount > 0 {
		scanningEnabledPercentage = (float64(scanningEnabledCount) / float64(buildPipelinesCount)) * 100
	}

	// Get external CI count directly
	externalCICount, err := impl.ciPipelineRepository.GetActiveExternalCiPipelineCount()
	if err != nil {
		impl.logger.Errorw("error getting external CI count", "err", err)
		return nil, err
	}

	// Get GitOps compliance count directly
	gitOpsComplianceCount, err := impl.deploymentConfigRepository.GetGitOpsEnabledPipelineCount()
	if err != nil {
		impl.logger.Errorw("error getting GitOps compliance count", "err", err)
		return nil, err
	}

	// Calculate GitOps coverage percentage
	totalActivePipelines := productionPipelinesCount + nonProdPipelinesCount
	var gitOpsCoveragePercentage float64
	if totalActivePipelines > 0 {
		gitOpsCoveragePercentage = (float64(gitOpsComplianceCount) / float64(totalActivePipelines)) * 100
	}

	metrics := &WorkflowMetrics{
		BuildPipelinesCount:       buildPipelinesCount,
		ProductionPipelinesCount:  productionPipelinesCount,
		NonProdPipelinesCount:     nonProdPipelinesCount,
		ExternalCICount:           externalCICount,
		ScanningEnabledPercentage: scanningEnabledPercentage,
		GitOpsComplianceCount:     gitOpsComplianceCount,
		GitOpsCoveragePercentage:  gitOpsCoveragePercentage,
	}

	return metrics, nil
}

func (impl *AppManagementServiceImpl) buildWorkflowOverviewResponse(allTime *WorkflowMetrics) *bean.WorkflowOverviewResponse {
	allTimeAllDeployments := allTime.ProductionPipelinesCount + allTime.NonProdPipelinesCount

	scanningMetric := &bean.AtAGlanceMetric{
		Percentage: allTime.ScanningEnabledPercentage,
	}

	gitOpsMetric := &bean.AtAGlanceMetric{
		Total:      allTime.GitOpsComplianceCount, // All-time count
		Percentage: allTime.GitOpsCoveragePercentage,
	}

	productionMetric := &bean.AtAGlanceMetric{
		Total: allTime.ProductionPipelinesCount, // All-time count
	}

	return &bean.WorkflowOverviewResponse{
		BuildPipelines: &bean.AtAGlanceMetric{
			Total: allTime.BuildPipelinesCount,
		},
		ExternalImageSource: &bean.AtAGlanceMetric{
			Total: allTime.ExternalCICount,
		},
		AllDeploymentPipelines: &bean.AtAGlanceMetric{
			Total: allTimeAllDeployments,
		},
		ScanningEnabledInWorkflows:    scanningMetric,
		GitOpsComplianceProdPipelines: gitOpsMetric,
		ProductionPipelines:           productionMetric,
	}
}

// calculateAverageBuildTime calculates the average build time for successful CI_BUILD pipelines
// in the given time range using accurate timing data from workflow_execution_stage table.
// This provides more precise build times by using the actual start_time and end_time from
// the Execution stage where workflow_type=CI, stage_name=Execution, status=SUCCEEDED, status_for=workflow.
func (impl *AppManagementServiceImpl) calculateAverageBuildTime(from, to *time.Time) (float64, error) {
	// Fetch successful builds from workflow_execution_stage table for accurate timing
	successfulBuilds, err := impl.getSuccessfulBuildsFromStages(from, to)
	if err != nil {
		impl.logger.Errorw("error getting successful builds for build time calculation from stages", "from", from, "to", to, "err", err)
		// Fallback to original method if new method fails
		impl.logger.Infow("falling back to original method for build time calculation")
		successfulBuilds, err = impl.ciWorkflowRepository.GetSuccessfulCIBuildsForBuildTime(from, to)
		if err != nil {
			impl.logger.Errorw("error getting successful builds for build time calculation (fallback)", "from", from, "to", to, "err", err)
			return 0, err
		}
	}

	// Return 0 if no successful builds found
	if len(successfulBuilds) == 0 {
		impl.logger.Infow("no successful builds found for average build time calculation", "from", from, "to", to)
		return 0, nil
	}

	// Calculate average build time in code for better precision and error handling
	totalBuildTimeMinutes := float64(0)
	validBuilds := 0

	for _, build := range successfulBuilds {
		// Ensure both timestamps are valid
		if !build.StartedOn.IsZero() && !build.FinishedOn.IsZero() && build.FinishedOn.After(build.StartedOn) {
			// Calculate duration in minutes with millisecond precision
			duration := build.FinishedOn.Sub(build.StartedOn)
			buildTimeMinutes := duration.Minutes()

			// Only include positive build times (sanity check)
			if buildTimeMinutes > 0 {
				totalBuildTimeMinutes += buildTimeMinutes
				validBuilds++
			}
		}
	}

	// Calculate average if we have valid builds
	avgBuildTime := float64(0)
	if validBuilds > 0 {
		avgBuildTime = totalBuildTimeMinutes / float64(validBuilds)
	}

	return avgBuildTime, nil
}

// getSuccessfulBuildsFromStages fetches successful CI builds from workflow_execution_stage table
// and converts them to WorkflowBuildTime format for compatibility with existing logic
func (impl *AppManagementServiceImpl) getSuccessfulBuildsFromStages(from, to *time.Time) ([]pipelineConfig.WorkflowBuildTime, error) {
	stages, err := impl.workflowStageRepository.GetSuccessfulCIExecutionStages(from, to)
	if err != nil {
		return nil, err
	}

	var workflows []pipelineConfig.WorkflowBuildTime
	for _, stage := range stages {
		startTime, err := impl.parseTimeString(stage.StartTime)
		if err != nil {
			impl.logger.Warnw("failed to parse start_time, skipping stage", "workflowId", stage.WorkflowId, "startTime", stage.StartTime, "err", err)
			continue
		}

		endTime, err := impl.parseTimeString(stage.EndTime)
		if err != nil {
			impl.logger.Warnw("failed to parse end_time, skipping stage", "workflowId", stage.WorkflowId, "endTime", stage.EndTime, "err", err)
			continue
		}

		if !endTime.After(startTime) {
			impl.logger.Warnw("end_time is not after start_time, skipping stage", "workflowId", stage.WorkflowId, "startTime", startTime, "endTime", endTime)
			continue
		}

		workflows = append(workflows, pipelineConfig.WorkflowBuildTime{
			StartedOn:  startTime,
			FinishedOn: endTime,
		})
	}

	return workflows, nil
}

// parseTimeString parses time string that can be either ISO format or Unix timestamp
func (impl *AppManagementServiceImpl) parseTimeString(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	// Try parsing as ISO format first (e.g., "2024-01-15T10:30:45Z")
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t, nil
	}

	// Try parsing as Unix timestamp in milliseconds
	if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		return time.Unix(timestamp/1000, (timestamp%1000)*1000000), nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time string: %s", timeStr)
}
