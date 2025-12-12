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

package constants

import (
	"time"
)

// TimePeriod represents the predefined time periods
type TimePeriod string

// TimeWindow represents the predefined time windows (same as TimePeriod but for API consistency)
type TimeWindow string

const (
	Today       TimePeriod = "today"
	ThisWeek    TimePeriod = "week"
	ThisMonth   TimePeriod = "month"
	ThisQuarter TimePeriod = "quarter"
	LastWeek    TimePeriod = "lastWeek"
	LastMonth   TimePeriod = "lastMonth"
)

// AggregationType represents how data should be aggregated
type AggregationType string

const (
	AggregateByHour  AggregationType = "HOUR"
	AggregateByDay   AggregationType = "DAY"
	AggregateByMonth AggregationType = "MONTH"
)

// TimeRange represents a time range with from and to timestamps
type TimeRange struct {
	From            time.Time
	To              time.Time
	AggregationType AggregationType
}

// IsValidTimePeriod checks if the given string is a valid time period
func IsValidTimePeriod(period string) bool {
	switch TimePeriod(period) {
	case Today, ThisWeek, ThisMonth, ThisQuarter, LastWeek, LastMonth:
		return true
	default:
		return false
	}
}

// IsValidTimeWindow checks if the given string is a valid time window
func IsValidTimeWindow(window string) bool {
	switch window {
	case "today", "week", "month", "quarter", "lastWeek", "lastMonth":
		return true
	default:
		return false
	}
}

// GetAggregationType returns the aggregation type for a given time period
// This is used to determine whether to aggregate data by hour, day, or month
func GetAggregationType(period TimePeriod) AggregationType {
	switch period {
	case Today:
		return AggregateByHour
	case ThisWeek, ThisMonth, LastWeek, LastMonth:
		return AggregateByDay
	case ThisQuarter:
		return AggregateByMonth
	default:
		return AggregateByDay
	}
}
