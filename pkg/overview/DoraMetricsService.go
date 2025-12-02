/*
 * Copyright (c) 2024. Devtron Inc.
 */

package overview

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/devtron-labs/devtron/client/lens"
	//"github.com/devtron-labs/devtron/client/lens"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/overview/bean"
	"github.com/devtron-labs/devtron/pkg/overview/util"
	"go.uber.org/zap"
)

type DoraMetricsService interface {
	GetDoraMetrics(ctx context.Context, request *bean.DoraMetricsRequest) (*bean.DoraMetricsResponse, error)
}
type DoraMetricsServiceImpl struct {
	logger                *zap.SugaredLogger
	lensClient            lens.LensClient
	appRepository         app.AppRepository
	pipelineRepository    pipelineConfig.PipelineRepository
	environmentRepository repository.EnvironmentRepository
	cdWorkflowRepository  pipelineConfig.CdWorkflowRepository
}

func NewDoraMetricsServiceImpl(
	logger *zap.SugaredLogger,
	lensClient lens.LensClient,
	appRepository app.AppRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	environmentRepository repository.EnvironmentRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
) *DoraMetricsServiceImpl {
	return &DoraMetricsServiceImpl{
		logger:                logger,
		lensClient:            lensClient,
		appRepository:         appRepository,
		pipelineRepository:    pipelineRepository,
		environmentRepository: environmentRepository,
		cdWorkflowRepository:  cdWorkflowRepository,
	}
}

func (impl *DoraMetricsServiceImpl) GetDoraMetrics(ctx context.Context, request *bean.DoraMetricsRequest) (*bean.DoraMetricsResponse, error) {
	impl.logger.Infow("getting DORA metrics", "request", request)

	// Get all apps with production pipelines using optimized query for current period
	appEnvPairs, err := impl.getAppEnvironmentPairsOptimized(ctx, request.TimeRangeRequest.From, request.TimeRangeRequest.To)
	if err != nil {
		impl.logger.Errorw("error getting app-environment pairs", "err", err)
		return nil, err
	}

	if len(appEnvPairs) == 0 {
		impl.logger.Warnw("no production pipelines found with deployment history")
		return bean.NewDoraMetricsResponse(), nil
	}

	// Calculate all DORA metrics using single Lens API call per app-env pair
	allMetrics, err := impl.calculateAllDoraMetricsFromLens(ctx, request, appEnvPairs)
	if err != nil {
		impl.logger.Errorw("error calculating DORA metrics from lens", "err", err)
		return nil, err
	}

	response := &bean.DoraMetricsResponse{
		ProdDeploymentPipelineCount: len(appEnvPairs),
		DeploymentFrequency:         allMetrics.DeploymentFrequency,
		MeanLeadTime:                allMetrics.MeanLeadTime,
		ChangeFailureRate:           allMetrics.ChangeFailureRate,
		MeanTimeToRecovery:          allMetrics.MeanTimeToRecovery,
	}

	return response, nil
}

// getAppEnvironmentPairsOptimized is an optimized version that uses a single query
// to fetch production pipelines with deployment history within the specified time range
// This method only fetches the minimal data needed (AppId and EnvId) for better performance
func (impl *DoraMetricsServiceImpl) getAppEnvironmentPairsOptimized(ctx context.Context, from, to *time.Time) ([]lens.AppEnvPair, error) {
	prodPipelines, err := impl.pipelineRepository.FindProdPipelinesWithAppDataAndDeploymentHistoryInTimeRange(from, to)
	if err != nil {
		impl.logger.Errorw("error getting production pipelines with deployment history in time range", "from", from, "to", to, "err", err)
		return nil, err
	}

	if len(prodPipelines) == 0 {
		impl.logger.Warnw("no production pipelines found with deployment history in time range", "from", from, "to", to)
		return []lens.AppEnvPair{}, nil
	}

	// Convert to our simplified structure, only keeping the IDs
	var appEnvPairs []lens.AppEnvPair
	for _, pipeline := range prodPipelines {
		appEnvPairs = append(appEnvPairs, lens.AppEnvPair{
			AppId: pipeline.AppId,
			EnvId: pipeline.EnvironmentId,
		})
	}

	return appEnvPairs, nil
}

// calculateAllDoraMetricsFromLens calculates all DORA metrics using single Lens API call per app-env pair
func (impl *DoraMetricsServiceImpl) calculateAllDoraMetricsFromLens(ctx context.Context, request *bean.DoraMetricsRequest, appEnvPairs []lens.AppEnvPair) (*bean.AllDoraMetrics, error) {
	currentMetricsData, err := impl.fetchAllMetricsFromLens(ctx, appEnvPairs, request.TimeRangeRequest.From, request.TimeRangeRequest.To)
	if err != nil {
		impl.logger.Errorw("error fetching current period metrics from lens", "err", err)
		return nil, err
	}

	// Get app-env pairs for previous period
	previousAppEnvPairs, err := impl.getAppEnvironmentPairsOptimized(ctx, request.PrevFrom, request.PrevTo)
	if err != nil {
		impl.logger.Errorw("error getting app-environment pairs for previous period", "err", err)
		// Continue without comparison if we can't get previous period data
		return impl.createAllDoraMetricsWithoutComparison(currentMetricsData), nil
	}

	var previousMetricsData map[string]*bean.LensMetrics
	if len(previousAppEnvPairs) > 0 {
		previousMetricsData, err = impl.fetchAllMetricsFromLens(ctx, previousAppEnvPairs, request.PrevFrom, request.PrevTo)
		if err != nil {
			impl.logger.Errorw("error fetching previous period metrics from lens", "err", err)
			// Continue without comparison if we can't get previous period data
			return impl.createAllDoraMetricsWithoutComparison(currentMetricsData), nil
		}
	}

	// Calculate all metrics with comparison
	return impl.createAllDoraMetricsWithComparison(currentMetricsData, previousMetricsData), nil
}

// fetchAllMetricsFromLens fetches all DORA metrics from Lens using single bulk API call
func (impl *DoraMetricsServiceImpl) fetchAllMetricsFromLens(ctx context.Context, bulkAppEnvPairs []lens.AppEnvPair, from, to *time.Time) (map[string]*bean.LensMetrics, error) {
	metricsData := make(map[string]*bean.LensMetrics)

	if len(bulkAppEnvPairs) == 0 {
		return metricsData, nil
	}

	bulkRequest := &lens.BulkMetricRequest{
		AppEnvPairs: bulkAppEnvPairs,
		From:        from,
		To:          to,
	}

	// Make single bulk call to get all metrics for all app-env pairs
	lensResp, resCode, err := impl.lensClient.GetBulkAppMetrics(bulkRequest)
	if err != nil {
		impl.logger.Errorw("error calling lens bulk API for all metrics", "err", err)
		return nil, err
	}

	if !resCode.IsSuccess() {
		impl.logger.Errorw("lens bulk API returned error", "statusCode", *resCode)
		return nil, fmt.Errorf("lens bulk API returned error with status code: %d", *resCode)
	}

	// Parse the new bulk response - now it's directly an array of DoraMetrics
	var doraMetricsArray []*lens.DoraMetrics
	if err := json.Unmarshal(lensResp.Result, &doraMetricsArray); err != nil {
		impl.logger.Errorw("error unmarshaling lens bulk response", "err", err)
		return nil, err
	}

	// Process results and map them to app-env keys
	for _, doraMetric := range doraMetricsArray {
		if doraMetric == nil {
			impl.logger.Warnw("nil dora metric in response")
			continue
		}

		// Convert lens.DoraMetrics to LensMetrics (our internal struct)
		// Map the new field names to the old structure for backward compatibility
		lensMetrics := &bean.LensMetrics{
			AverageCycleTime:    doraMetric.DeploymentFrequency,    // DeploymentFrequency maps to AverageCycleTime
			AverageLeadTime:     doraMetric.MeanLeadTimeForChanges, // MeanLeadTimeForChanges maps to AverageLeadTime
			ChangeFailureRate:   doraMetric.ChangeFailureRate,      // ChangeFailureRate maps directly
			AverageRecoveryTime: doraMetric.MeanTimeToRecovery,     // MeanTimeToRecovery maps to AverageRecoveryTime
		}

		// Store metrics with app-env ID key
		key := fmt.Sprintf("%d-%d", doraMetric.AppId, doraMetric.EnvId)
		metricsData[key] = lensMetrics
	}

	return metricsData, nil
}

// createAllDoraMetricsWithoutComparison creates all DORA metrics without comparison data
func (impl *DoraMetricsServiceImpl) createAllDoraMetricsWithoutComparison(currentMetricsData map[string]*bean.LensMetrics) *bean.AllDoraMetrics {
	// Extract all metric values from current data
	var deploymentFreqValues, leadTimeValues, changeFailureValues, recoveryTimeValues []float64

	for _, metrics := range currentMetricsData {
		deploymentFreqValues = append(deploymentFreqValues, metrics.AverageCycleTime)
		leadTimeValues = append(leadTimeValues, metrics.AverageLeadTime)
		changeFailureValues = append(changeFailureValues, metrics.ChangeFailureRate)
		recoveryTimeValues = append(recoveryTimeValues, metrics.AverageRecoveryTime)
	}

	// Calculate averages
	deploymentFreqAvg := util.CalculateAverageFromValues(deploymentFreqValues)
	leadTimeAvg := util.CalculateAverageFromValues(leadTimeValues)
	changeFailureAvg := util.CalculateAverageFromValues(changeFailureValues)
	recoveryTimeAvg := util.CalculateAverageFromValues(recoveryTimeValues)

	// Calculate performance levels for each metric separately
	deploymentFreqPerformanceLevels := util.CalculatePerformanceLevelsForMetric(currentMetricsData, bean.MetricCategoryDeploymentFrequency)
	leadTimePerformanceLevels := util.CalculatePerformanceLevelsForMetric(currentMetricsData, bean.MetricCategoryMeanLeadTime)
	changeFailurePerformanceLevels := util.CalculatePerformanceLevelsForMetric(currentMetricsData, bean.MetricCategoryChangeFailureRate)
	recoveryTimePerformanceLevels := util.CalculatePerformanceLevelsForMetric(currentMetricsData, bean.MetricCategoryMeanTimeToRecovery)

	deploymentFrequency := util.CreateDoraMetricObject(deploymentFreqAvg, bean.MetricValueUnitNumber, 0, bean.ComparisonUnitPercentage, deploymentFreqPerformanceLevels)
	meanLeadTime := util.CreateDoraMetricObject(leadTimeAvg, bean.MetricValueUnitMinutes, 0, bean.ComparisonUnitMinutes, leadTimePerformanceLevels)
	changeFailureRate := util.CreateDoraMetricObject(changeFailureAvg, bean.MetricValueUnitPercentage, 0, bean.ComparisonUnitPercentage, changeFailurePerformanceLevels)
	meanTimeToRecovery := util.CreateDoraMetricObject(recoveryTimeAvg, bean.MetricValueUnitMinutes, 0, bean.ComparisonUnitMinutes, recoveryTimePerformanceLevels)

	allDoraMetrics := bean.NewAllDoraMetrics().
		WithDeploymentFrequency(deploymentFrequency).
		WithMeanLeadTime(meanLeadTime).
		WithChangeFailureRate(changeFailureRate).
		WithMeanTimeToRecovery(meanTimeToRecovery)

	return allDoraMetrics

}

// createAllDoraMetricsWithComparison creates all DORA metrics with comparison data
func (impl *DoraMetricsServiceImpl) createAllDoraMetricsWithComparison(currentMetricsData, previousMetricsData map[string]*bean.LensMetrics) *bean.AllDoraMetrics {
	// Extract current period values
	var currentDeploymentFreq, currentLeadTime, currentChangeFailure, currentRecoveryTime []float64
	for _, metrics := range currentMetricsData {
		currentDeploymentFreq = append(currentDeploymentFreq, metrics.AverageCycleTime)
		currentLeadTime = append(currentLeadTime, metrics.AverageLeadTime)
		currentChangeFailure = append(currentChangeFailure, metrics.ChangeFailureRate)
		currentRecoveryTime = append(currentRecoveryTime, metrics.AverageRecoveryTime)
	}

	// Extract previous period values
	var previousDeploymentFreq, previousLeadTime, previousChangeFailure, previousRecoveryTime []float64
	for _, metrics := range previousMetricsData {
		previousDeploymentFreq = append(previousDeploymentFreq, metrics.AverageCycleTime)
		previousLeadTime = append(previousLeadTime, metrics.AverageLeadTime)
		previousChangeFailure = append(previousChangeFailure, metrics.ChangeFailureRate)
		previousRecoveryTime = append(previousRecoveryTime, metrics.AverageRecoveryTime)
	}

	// Calculate averages
	currentDeploymentFreqAvg := util.CalculateAverageFromValues(currentDeploymentFreq)
	currentLeadTimeAvg := util.CalculateAverageFromValues(currentLeadTime)
	currentChangeFailureAvg := util.CalculateAverageFromValues(currentChangeFailure)
	currentRecoveryTimeAvg := util.CalculateAverageFromValues(currentRecoveryTime)

	previousDeploymentFreqAvg := util.CalculateAverageFromValues(previousDeploymentFreq)
	previousLeadTimeAvg := util.CalculateAverageFromValues(previousLeadTime)
	previousChangeFailureAvg := util.CalculateAverageFromValues(previousChangeFailure)
	previousRecoveryTimeAvg := util.CalculateAverageFromValues(previousRecoveryTime)

	// Calculate comparisons
	deploymentFreqCompValue := util.CalculateComparison(currentDeploymentFreqAvg, previousDeploymentFreqAvg, bean.MetricCategoryDeploymentFrequency)
	leadTimeCompValue := util.CalculateComparison(currentLeadTimeAvg, previousLeadTimeAvg, bean.MetricCategoryMeanLeadTime)
	changeFailureCompValue := util.CalculateComparison(currentChangeFailureAvg, previousChangeFailureAvg, bean.MetricCategoryChangeFailureRate)
	recoveryTimeCompValue := util.CalculateComparison(currentRecoveryTimeAvg, previousRecoveryTimeAvg, bean.MetricCategoryMeanTimeToRecovery)

	// Calculate performance levels for each metric separately using current period data
	deploymentFreqPerformanceLevels := util.CalculatePerformanceLevelsForMetric(currentMetricsData, bean.MetricCategoryDeploymentFrequency)
	leadTimePerformanceLevels := util.CalculatePerformanceLevelsForMetric(currentMetricsData, bean.MetricCategoryMeanLeadTime)
	changeFailurePerformanceLevels := util.CalculatePerformanceLevelsForMetric(currentMetricsData, bean.MetricCategoryChangeFailureRate)
	recoveryTimePerformanceLevels := util.CalculatePerformanceLevelsForMetric(currentMetricsData, bean.MetricCategoryMeanTimeToRecovery)

	deploymentFrequency := util.CreateDoraMetricObject(currentDeploymentFreqAvg, bean.MetricValueUnitNumber, deploymentFreqCompValue, bean.ComparisonUnitPercentage, deploymentFreqPerformanceLevels)
	meanLeadTime := util.CreateDoraMetricObject(currentLeadTimeAvg, bean.MetricValueUnitMinutes, leadTimeCompValue, bean.ComparisonUnitMinutes, leadTimePerformanceLevels)
	changeFailureRate := util.CreateDoraMetricObject(currentChangeFailureAvg, bean.MetricValueUnitPercentage, changeFailureCompValue, bean.ComparisonUnitPercentage, changeFailurePerformanceLevels)
	meanTimeToRecovery := util.CreateDoraMetricObject(currentRecoveryTimeAvg, bean.MetricValueUnitMinutes, recoveryTimeCompValue, bean.ComparisonUnitMinutes, recoveryTimePerformanceLevels)

	allDoraMetrics := bean.NewAllDoraMetrics().
		WithDeploymentFrequency(deploymentFrequency).
		WithMeanLeadTime(meanLeadTime).
		WithChangeFailureRate(changeFailureRate).
		WithMeanTimeToRecovery(meanTimeToRecovery)

	return allDoraMetrics
}
