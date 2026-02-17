/*
 * Copyright (c) 2024. Devtron Inc.
 */

package util

import (
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/overview/bean"
	"github.com/devtron-labs/devtron/pkg/overview/constants"
	teamRepository "github.com/devtron-labs/devtron/pkg/team/repository"
)

// TrendCalculator provides utility functions for calculating trend comparisons
type TrendCalculator struct{}

// NewTrendCalculator creates a new instance of TrendCalculator
func NewTrendCalculator() *TrendCalculator {
	return &TrendCalculator{}
}

// CalculateTrendComparison calculates the trend comparison between current and previous period
func (tc *TrendCalculator) CalculateTrendComparison(currentValue, previousValue int, from, to *time.Time) *bean.TrendComparison {
	timePeriod := GetTimePeriodFromTimeRange(from, to)

	if previousValue == 0 && currentValue == 0 {
		return &bean.TrendComparison{
			Value: 0,
			Label: tc.getTrendLabel(timePeriod),
		}
	}

	difference := currentValue - previousValue
	return &bean.TrendComparison{
		Value: difference,
		Label: tc.getTrendLabel(timePeriod),
	}
}

// CalculatePercentageTrendComparison calculates the trend comparison for percentage values
func (tc *TrendCalculator) CalculatePercentageTrendComparison(currentPercentage, previousPercentage float64, from, to *time.Time) *bean.TrendComparison {
	timePeriod := GetTimePeriodFromTimeRange(from, to)

	difference := int(currentPercentage - previousPercentage)
	return &bean.TrendComparison{
		Value: difference,
		Label: tc.getTrendLabel(timePeriod),
	}
}

// GetPreviousPeriodTimeRange calculates the time range for the previous period based on the current period
// It simply subtracts the duration from the current period to get the previous period
func (tc *TrendCalculator) GetPreviousPeriodTimeRange(from, to *time.Time) (*time.Time, *time.Time) {
	if from == nil || to == nil {
		return nil, nil
	}

	// Calculate the duration of the current period
	duration := to.Sub(*from)

	// Previous period ends where current period starts, and starts duration before that
	prevTo := *from
	prevFrom := from.Add(-duration)

	return &prevFrom, &prevTo
}

// getTrendLabel returns the appropriate label for the trend comparison
func (tc *TrendCalculator) getTrendLabel(timePeriod constants.TimePeriod) string {
	switch timePeriod {
	case constants.Today:
		return "today"
	case constants.ThisWeek:
		return "this week"
	case constants.ThisMonth:
		return "this month"
	case constants.ThisQuarter:
		return "this quarter"
	case constants.LastWeek:
		return "last week"
	case constants.LastMonth:
		return "last month"
	default:
		return "this period"
	}
}

// CalculateTrendForTimeDataPoints calculates trend comparison for time-based data points
func (tc *TrendCalculator) CalculateTrendForTimeDataPoints(currentData, previousData []bean.TimeDataPoint, from, to *time.Time) *bean.TrendComparison {
	currentTotal := 0
	for _, point := range currentData {
		currentTotal += point.Count
	}

	previousTotal := 0
	for _, point := range previousData {
		previousTotal += point.Count
	}

	return tc.CalculateTrendComparison(currentTotal, previousTotal, from, to)
}

// CalculatePercentageFromCounts calculates percentage from counts
func (tc *TrendCalculator) CalculatePercentageFromCounts(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0.0
	}
	return (float64(numerator) / float64(denominator)) * 100.0
}

// CalculateAverageFromCounts calculates average from total and count
func (tc *TrendCalculator) CalculateAverageFromCounts(total, count int) float64 {
	if count == 0 {
		return 0.0
	}
	return float64(total) / float64(count)
}

// FilterTeamsByTimeRange filters teams by time range based on created_on field
func FilterTeamsByTimeRange(teams []teamRepository.Team, from, to *time.Time) []teamRepository.Team {
	if from == nil && to == nil {
		return teams
	}

	filtered := make([]teamRepository.Team, 0)
	for _, team := range teams {
		if IsWithinTimeRange(team.CreatedOn, from, to) {
			filtered = append(filtered, team)
		}
	}
	return filtered
}

// FilterAppsByTimeRange filters apps by time range based on created_on field
func FilterAppsByTimeRange(apps []*app.App, from, to *time.Time) []*app.App {
	if from == nil && to == nil {
		return apps
	}

	filtered := make([]*app.App, 0)
	for _, app := range apps {
		if IsWithinTimeRange(app.CreatedOn, from, to) {
			filtered = append(filtered, app)
		}
	}
	return filtered
}

// FilterEnvironmentsByTimeRange filters environments by time range based on created_on field
func FilterEnvironmentsByTimeRange(environments []*repository.Environment, from, to *time.Time) []*repository.Environment {
	if from == nil && to == nil {
		return environments
	}

	filtered := make([]*repository.Environment, 0)
	for _, env := range environments {
		if IsWithinTimeRange(env.CreatedOn, from, to) {
			filtered = append(filtered, env)
		}
	}
	return filtered
}

// FilterCiPipelinesByTimeRange filters CI pipelines by time range based on created_on field
func FilterCiPipelinesByTimeRange(pipelines []*pipelineConfig.CiPipeline, from, to *time.Time) []*pipelineConfig.CiPipeline {
	if from == nil && to == nil {
		return pipelines
	}

	filtered := make([]*pipelineConfig.CiPipeline, 0)
	for _, pipeline := range pipelines {
		if IsWithinTimeRange(pipeline.CreatedOn, from, to) {
			filtered = append(filtered, pipeline)
		}
	}
	return filtered
}

// FilterCdPipelinesByTimeRange filters CD pipelines by time range based on created_on field
func FilterCdPipelinesByTimeRange(pipelines []*pipelineConfig.Pipeline, from, to *time.Time) []*pipelineConfig.Pipeline {
	if from == nil && to == nil {
		return pipelines
	}

	filtered := make([]*pipelineConfig.Pipeline, 0)
	for _, pipeline := range pipelines {
		if IsWithinTimeRange(pipeline.CreatedOn, from, to) {
			filtered = append(filtered, pipeline)
		}
	}
	return filtered
}

// IsWithinTimeRange checks if a timestamp is within the given time range
func IsWithinTimeRange(timestamp time.Time, from, to *time.Time) bool {
	if from != nil && timestamp.Before(*from) {
		return false
	}
	if to != nil && timestamp.After(*to) {
		return false
	}
	return true
}

// CalculateAppTrendFromPeriodComparison calculates trend by comparing current and previous period app counts
func CalculateAppTrendFromPeriodComparison(currentApps, previousApps []*app.App) int {
	currentCount := 0
	for _, app := range currentApps {
		if app.Active {
			currentCount++
		}
	}

	previousCount := 0
	for _, app := range previousApps {
		if app.Active {
			previousCount++
		}
	}

	return currentCount - previousCount
}

// CalculateTeamTrendFromPeriodComparison calculates trend by comparing current and previous period team counts
func CalculateTeamTrendFromPeriodComparison(currentTeams, previousTeams []teamRepository.Team) int {
	currentCount := 0
	for _, team := range currentTeams {
		if team.Active {
			currentCount++
		}
	}

	previousCount := 0
	for _, team := range previousTeams {
		if team.Active {
			previousCount++
		}
	}

	return currentCount - previousCount
}

// CalculateEnvironmentTrendFromPeriodComparison calculates trend by comparing current and previous period environment counts
func CalculateEnvironmentTrendFromPeriodComparison(currentEnvs, previousEnvs []*repository.Environment) int {
	currentCount := 0
	for _, env := range currentEnvs {
		if env.Active {
			currentCount++
		}
	}

	previousCount := 0
	for _, env := range previousEnvs {
		if env.Active {
			previousCount++
		}
	}

	return currentCount - previousCount
}

// CalculateCiPipelineTrendFromPeriodComparison calculates trend by comparing current and previous period CI pipeline counts
func CalculateCiPipelineTrendFromPeriodComparison(currentPipelines, previousPipelines []*pipelineConfig.CiPipeline) int {
	currentCount := 0
	for _, pipeline := range currentPipelines {
		if !pipeline.Deleted {
			currentCount++
		}
	}

	previousCount := 0
	for _, pipeline := range previousPipelines {
		if !pipeline.Deleted {
			previousCount++
		}
	}

	return currentCount - previousCount
}

// CalculateCdPipelineTrendFromPeriodComparison calculates trend by comparing current and previous period CD pipeline counts
func CalculateCdPipelineTrendFromPeriodComparison(currentPipelines, previousPipelines []*pipelineConfig.Pipeline) int {
	currentCount := 0
	for _, pipeline := range currentPipelines {
		if !pipeline.Deleted {
			currentCount++
		}
	}

	previousCount := 0
	for _, pipeline := range previousPipelines {
		if !pipeline.Deleted {
			previousCount++
		}
	}

	return currentCount - previousCount
}
