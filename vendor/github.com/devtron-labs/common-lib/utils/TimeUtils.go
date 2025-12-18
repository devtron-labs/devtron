/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"fmt"
	"strings"
	"time"
)

type TimeRangeRequest struct {
	From       *time.Time   `json:"from" schema:"from"`
	To         *time.Time   `json:"to" schema:"to"`
	TimeWindow *TimeWindows `json:"timeWindow" schema:"timeWindow" validate:"omitempty,oneof=today yesterday week month quarter lastWeek lastMonth lastQuarter last24Hours last7Days last30Days last90Days"`
}

func NewTimeRangeRequest(from *time.Time, to *time.Time) *TimeRangeRequest {
	return &TimeRangeRequest{
		From: from,
		To:   to,
	}
}

func NewTimeWindowRequest(timeWindow TimeWindows) *TimeRangeRequest {
	return &TimeRangeRequest{
		TimeWindow: &timeWindow,
	}
}

// TimeWindows is a string type that represents different time windows
type TimeWindows string

func (timeRange TimeWindows) String() string {
	return string(timeRange)
}

// Define constants for different time windows
const (
	Today       TimeWindows = "today"
	Yesterday   TimeWindows = "yesterday"
	Week        TimeWindows = "week"
	Month       TimeWindows = "month"
	Quarter     TimeWindows = "quarter"
	LastWeek    TimeWindows = "lastWeek"
	LastMonth   TimeWindows = "lastMonth"
	Year        TimeWindows = "year"
	LastQuarter TimeWindows = "lastQuarter"
	Last24Hours TimeWindows = "last24Hours"
	Last7Days   TimeWindows = "last7Days"
	Last30Days  TimeWindows = "last30Days"
	Last90Days  TimeWindows = "last90Days"
)

func (timeRange *TimeRangeRequest) ParseAndValidateTimeRange() (*TimeRangeRequest, error) {
	if timeRange == nil {
		return NewTimeRangeRequest(&time.Time{}, &time.Time{}), fmt.Errorf("invalid time range request. either from/to or timeWindow must be provided")
	}
	now := time.Now()
	// If timeWindow is provided, it takes preference over from/to
	if timeRange.TimeWindow != nil {
		switch *timeRange.TimeWindow {
		case Today:
			start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			return NewTimeRangeRequest(&start, &now), nil
		case Yesterday:
			start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(-24 * time.Hour)
			end := start.Add(24 * time.Hour)
			return NewTimeRangeRequest(&start, &end), nil
		case Week:
			// Current week (Monday to Sunday)
			weekday := int(now.Weekday())
			if weekday == 0 { // Sunday
				weekday = 7
			}
			start := now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
			return NewTimeRangeRequest(&start, &now), nil
		case Month:
			start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			return NewTimeRangeRequest(&start, &now), nil
		case Quarter:
			quarter := ((int(now.Month()) - 1) / 3) + 1
			quarterStart := time.Month((quarter-1)*3 + 1)
			start := time.Date(now.Year(), quarterStart, 1, 0, 0, 0, 0, now.Location())
			return NewTimeRangeRequest(&start, &now), nil
		case LastWeek:
			weekday := int(now.Weekday())
			if weekday == 0 { // Sunday
				weekday = 7
			}
			thisWeekStart := now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
			lastWeekStart := thisWeekStart.AddDate(0, 0, -7)
			lastWeekEnd := thisWeekStart.Add(-time.Second)
			return NewTimeRangeRequest(&lastWeekStart, &lastWeekEnd), nil
		case LastMonth:
			thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
			lastMonthEnd := thisMonthStart.Add(-time.Second)
			return NewTimeRangeRequest(&lastMonthStart, &lastMonthEnd), nil
		case LastQuarter:
			// Calculate current quarter
			currentQuarter := ((int(now.Month()) - 1) / 3) + 1

			// Calculate previous quarter
			var prevQuarter int
			var prevYear int
			if currentQuarter == 1 {
				// If current quarter is Q1, previous quarter is Q4 of previous year
				prevQuarter = 4
				prevYear = now.Year() - 1
			} else {
				// Otherwise, previous quarter is in the same year
				prevQuarter = currentQuarter - 1
				prevYear = now.Year()
			}

			// Calculate start and end of previous quarter
			prevQuarterStartMonth := time.Month((prevQuarter-1)*3 + 1)
			prevQuarterStart := time.Date(prevYear, prevQuarterStartMonth, 1, 0, 0, 0, 0, now.Location())

			// End of previous quarter is the start of current quarter minus 1 second
			currentQuarterStartMonth := time.Month((currentQuarter-1)*3 + 1)
			currentQuarterStart := time.Date(now.Year(), currentQuarterStartMonth, 1, 0, 0, 0, 0, now.Location())
			if currentQuarter == 1 {
				// If current quarter is Q1, we need to calculate Q4 end of previous year
				currentQuarterStart = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
			}
			prevQuarterEnd := currentQuarterStart.Add(-time.Second)

			return NewTimeRangeRequest(&prevQuarterStart, &prevQuarterEnd), nil
		case Year:
			start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
			return NewTimeRangeRequest(&start, &now), nil
		case Last24Hours:
			start := now.Add(-24 * time.Hour)
			return NewTimeRangeRequest(&start, &now), nil
		case Last7Days:
			start := now.AddDate(0, 0, -7)
			return NewTimeRangeRequest(&start, &now), nil
		case Last30Days:
			start := now.AddDate(0, 0, -30)
			return NewTimeRangeRequest(&start, &now), nil
		case Last90Days:
			start := now.AddDate(0, 0, -90)
			return NewTimeRangeRequest(&start, &now), nil
		default:
			return NewTimeRangeRequest(&time.Time{}, &time.Time{}), fmt.Errorf("unsupported time window: %q", *timeRange.TimeWindow)
		}
	}

	// Use from/to dates if provided
	if timeRange.From != nil && timeRange.To != nil {
		if timeRange.From.After(*timeRange.To) {
			return NewTimeRangeRequest(&time.Time{}, &time.Time{}), fmt.Errorf("from date cannot be after to date")
		}
		return NewTimeRangeRequest(timeRange.From, timeRange.To), nil
	} else {
		return NewTimeRangeRequest(&time.Time{}, &time.Time{}), fmt.Errorf("from and to dates are required if time window is not provided")
	}
}

// TimeBoundariesRequest represents the request for time boundary frames
type TimeBoundariesRequest struct {
	TimeWindowBoundaries []string     `json:"timeWindowBoundaries" schema:"timeWindowBoundaries" validate:"omitempty,min=1"`
	TimeWindow           *TimeWindows `json:"timeWindow" schema:"timeWindow" validate:"omitempty,oneof=week month quarter year"` // week, month, quarter, year
	Iterations           int          `json:"iterations" schema:"iterations" validate:"omitempty,min=1"`
}

// TimeWindowBoundaries represents the start and end times for a time window
type TimeWindowBoundaries struct {
	StartTime time.Time
	EndTime   time.Time
}

func (timeBoundaries *TimeBoundariesRequest) ParseAndValidateTimeBoundaries() ([]TimeWindowBoundaries, error) {
	if timeBoundaries == nil {
		return []TimeWindowBoundaries{}, fmt.Errorf("invalid time boundaries request")
	}
	// If timeWindow is provided, it takes preference over timeWindowBoundaries
	if timeBoundaries.TimeWindow != nil {
		switch *timeBoundaries.TimeWindow {
		case Week:
			return GetWeeklyTimeBoundaries(timeBoundaries.Iterations), nil
		case Month:
			return GetMonthlyTimeBoundaries(timeBoundaries.Iterations), nil
		case Quarter:
			return GetQuarterlyTimeBoundaries(timeBoundaries.Iterations), nil
		case Year:
			return GetYearlyTimeBoundaries(timeBoundaries.Iterations), nil
		default:
			return []TimeWindowBoundaries{}, fmt.Errorf("unsupported time window: %q", *timeBoundaries.TimeWindow)
		}
	} else if len(timeBoundaries.TimeWindowBoundaries) != 0 {
		// Validate time window
		return DecodeAndValidateTimeWindowBoundaries(timeBoundaries.TimeWindowBoundaries)
	} else {
		return []TimeWindowBoundaries{}, fmt.Errorf("time window boundaries are required if time window is not provided")
	}
}

func GetWeeklyTimeBoundaries(iterations int) []TimeWindowBoundaries {
	if iterations <= 0 {
		return []TimeWindowBoundaries{}
	}
	boundaries := make([]TimeWindowBoundaries, iterations)
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	// Get start of this week (Monday)
	weekStart := now.AddDate(0, 0, -(weekday - 1))
	// Set time to midnight
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

	for i := 0; i < iterations; i++ {
		start := weekStart.AddDate(0, 0, -7*i)
		end := start.AddDate(0, 0, 7)
		// For the current week, if now < end, set end = now
		if i == 0 && now.Before(end) {
			end = now
		}
		boundaries[i] = TimeWindowBoundaries{
			StartTime: start,
			EndTime:   end,
		}
	}
	return boundaries
}

func GetMonthlyTimeBoundaries(iterations int) []TimeWindowBoundaries {
	if iterations <= 0 {
		return []TimeWindowBoundaries{}
	}
	boundaries := make([]TimeWindowBoundaries, iterations)
	now := time.Now()
	// Get start of this month (1st)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	for i := 0; i < iterations; i++ {
		start := monthStart.AddDate(0, -i, 0)
		end := start.AddDate(0, 1, 0)
		// For the current month, if now < end, set end = now
		if i == 0 && now.Before(end) {
			end = now
		}
		boundaries[i] = TimeWindowBoundaries{
			StartTime: start,
			EndTime:   end,
		}
	}
	return boundaries
}

func GetQuarterlyTimeBoundaries(iterations int) []TimeWindowBoundaries {
	if iterations <= 0 {
		return []TimeWindowBoundaries{}
	}
	boundaries := make([]TimeWindowBoundaries, iterations)
	now := time.Now()
	quarter := ((int(now.Month()) - 1) / 3) + 1
	quarterMonth := time.Month((quarter-1)*3 + 1)
	// Get start of this quarter (1st of the month)
	quarterStart := time.Date(now.Year(), quarterMonth, 1, 0, 0, 0, 0, now.Location())
	for i := 0; i < iterations; i++ {
		start := quarterStart.AddDate(0, -3*i, 0)
		end := start.AddDate(0, 3, 0)
		// For the current quarter, if now < end, set end = now
		if i == 0 && now.Before(end) {
			end = now
		}
		boundaries[i] = TimeWindowBoundaries{
			StartTime: start,
			EndTime:   end,
		}
	}
	return boundaries
}

func GetYearlyTimeBoundaries(iterations int) []TimeWindowBoundaries {
	if iterations <= 0 {
		return []TimeWindowBoundaries{}
	}
	boundaries := make([]TimeWindowBoundaries, iterations)
	now := time.Now()
	// Get start of this year (1st of January)
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	for i := 0; i < iterations; i++ {
		start := yearStart.AddDate(-i, 0, 0)
		end := start.AddDate(1, 0, 0)
		// For the current year, if now < end, set end = now
		if i == 0 && now.Before(end) {
			end = now
		}
		boundaries[i] = TimeWindowBoundaries{
			StartTime: start,
			EndTime:   end,
		}
	}
	return boundaries
}

func DecodeAndValidateTimeWindowBoundaries(timeWindowBoundaries []string) ([]TimeWindowBoundaries, error) {
	boundaries := make([]TimeWindowBoundaries, 0, len(timeWindowBoundaries))
	for _, boundary := range timeWindowBoundaries {
		parts := strings.Split(boundary, "|")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid time window boundary format: %q", boundary)
		}
		startTime, err := time.Parse(time.RFC3339, parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid start time format: %q. expected format: %q", parts[0], time.RFC3339)
		}
		endTime, err := time.Parse(time.RFC3339, parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid end time format: %q. expected format: %q", parts[1], time.RFC3339)
		}
		if startTime.After(endTime) {
			return nil, fmt.Errorf("start time cannot be after end time: %q", boundary)
		}
		boundaries = append(boundaries, TimeWindowBoundaries{
			StartTime: startTime,
			EndTime:   endTime,
		})
	}
	return boundaries, nil
}
