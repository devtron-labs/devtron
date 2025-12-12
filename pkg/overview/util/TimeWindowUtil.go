package util

import (
	"fmt"
	"time"

	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/devtron/pkg/overview/constants"
)

// TimeRange represents a time range with from and to timestamps
type TimeRange struct {
	From time.Time
	To   time.Time
}

// ParseTimeString helper function to parse time string in ISO 8601 format
func ParseTimeString(timeStr string) (time.Time, error) {
	// Try parsing with different time formats
	formats := []string{
		time.RFC3339,     // "2006-01-02T15:04:05Z07:00"
		time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.000-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
}

// GetCurrentTimePeriodBasedOnTimeWindow parses the time-based filter request with timeWindow support using individual parameters
func GetCurrentTimePeriodBasedOnTimeWindow(timeWindow, from, to string) (*utils.TimeRangeRequest, error) {
	timeRangeReq := &utils.TimeRangeRequest{}
	if len(timeWindow) > 0 {
		timeWindowType := utils.TimeWindows(timeWindow)
		timeRangeReq = utils.NewTimeWindowRequest(timeWindowType)
	} else {
		if len(from) == 0 || len(to) == 0 {
			return nil, fmt.Errorf("either timeWindow or both from/to parameters must be provided")
		}
		var fromTime, toTime *time.Time
		if parsedTime, err := ParseTimeString(from); err == nil {
			fromTime = &parsedTime
		} else {
			return nil, fmt.Errorf("invalid 'from' time format: %s", from)
		}
		if parsedTime, err := ParseTimeString(to); err == nil {
			toTime = &parsedTime
		} else {
			return nil, fmt.Errorf("invalid 'from' time format: %s", to)
		}
		timeRangeReq = utils.NewTimeRangeRequest(fromTime, toTime)
	}
	timeRange, err := timeRangeReq.ParseAndValidateTimeRange()
	if err != nil {
		return nil, err
	}

	return timeRange, nil
}

// calculatePreviousTimeRangeFromDuration calculates the previous time range based on current time range duration
// currentFrom becomes the previous To, and prevFrom is calculated by subtracting the duration from currentFrom
func calculatePreviousTimeRangeFromDuration(currentFrom, currentTo *time.Time) (*utils.TimeRangeRequest, error) {
	if currentFrom == nil || currentTo == nil {
		return nil, fmt.Errorf("currentFrom and currentTo cannot be nil")
	}

	// Calculate the duration between current from and to
	duration := currentTo.Sub(*currentFrom)

	// Previous To becomes current From
	prevTo := *currentFrom

	// Previous From is calculated by subtracting the duration from current From
	prevFrom := currentFrom.Add(-duration)

	// Create time range request for the calculated previous period
	timeRangeReq := utils.NewTimeRangeRequest(&prevFrom, &prevTo)

	// Parse and validate the time range
	timeRange, err := timeRangeReq.ParseAndValidateTimeRange()
	if err != nil {
		return nil, fmt.Errorf("failed to parse calculated previous time period: %w", err)
	}

	return timeRange, nil
}

// GetPreviousTimePeriodBasedOnTimeWindow calculates the previous from and to using the timeWindow key
// It maps current time windows to their previous equivalents and calls ParseAndValidateTimeRange
// For unknown timeWindows, it falls back to duration-based calculation using currentFrom and currentTo
func GetPreviousTimePeriodBasedOnTimeWindow(timeWindow string, currentFrom, currentTo *time.Time) (*utils.TimeRangeRequest, error) {
	var previousTimeWindow utils.TimeWindows

	switch utils.TimeWindows(timeWindow) {
	case utils.Today:
		// If user provided today, use yesterday
		previousTimeWindow = utils.Yesterday
	case utils.Week:
		// If user provided week, use lastWeek
		previousTimeWindow = utils.LastWeek
	case utils.Month:
		// If user provided month, use lastMonth
		previousTimeWindow = utils.LastMonth
	case utils.Quarter:
		// If user provided quarter, use lastQuarter
		previousTimeWindow = utils.LastQuarter
	default:
		// Fallback to duration-based calculation for unknown timeWindows
		return calculatePreviousTimeRangeFromDuration(currentFrom, currentTo)
	}

	// Create time window request for the previous period
	timeRangeReq := utils.NewTimeWindowRequest(previousTimeWindow)

	// Parse and validate the time range
	timeRange, err := timeRangeReq.ParseAndValidateTimeRange()
	if err != nil {
		return nil, fmt.Errorf("failed to parse previous time period for timeWindow %s: %w", timeWindow, err)
	}

	return timeRange, nil
}

// GetCurrentAndPreviousTimeRangeBasedOnTimeWindow calculates and returns the current and previous time ranges based on a time window.
// It supports parsing and validating a time-based filter with optional "from" and "to" parameters and time window input.
// Returns two time range requests for the current and previous periods, or an error if parsing or validation fails.
func GetCurrentAndPreviousTimeRangeBasedOnTimeWindow(timeWindow, from, to string) (*utils.TimeRangeRequest, *utils.TimeRangeRequest, error) {
	// Get current time range
	currentFromTo, err := GetCurrentTimePeriodBasedOnTimeWindow(timeWindow, from, to)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse current time period: %w", err)
	}

	// Get previous time range using current time range for fallback calculation
	prevFromTo, err := GetPreviousTimePeriodBasedOnTimeWindow(timeWindow, currentFromTo.From, currentFromTo.To)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse previous time period: %w", err)
	}

	return currentFromTo, prevFromTo, nil
}

// GetTimePeriodFromTimeRange determines the time period based on the duration between from and to
func GetTimePeriodFromTimeRange(from, to *time.Time) constants.TimePeriod {
	if from == nil || to == nil {
		return constants.ThisWeek // default
	}

	duration := to.Sub(*from)

	// If the duration is approximately 1 day (within 2 hours tolerance)
	if duration <= 26*time.Hour && duration >= 22*time.Hour {
		return constants.Today
	}

	// If the duration is approximately 1 week (within 1 day tolerance)
	if duration <= 8*24*time.Hour && duration >= 6*24*time.Hour {
		return constants.ThisWeek
	}

	// If the duration is approximately 1 month (within 3 days tolerance)
	if duration <= 33*24*time.Hour && duration >= 28*24*time.Hour {
		return constants.ThisMonth
	}

	// If the duration is approximately 3 months (within 1 week tolerance)
	if duration <= 97*24*time.Hour && duration >= 83*24*time.Hour {
		return constants.ThisQuarter
	}

	// Default to week for other durations
	return constants.ThisWeek
}
