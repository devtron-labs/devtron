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

package timeRangeLib

import "time"

func (tr TimeRange) getTimeRangeExpressionEvaluator(targetTime time.Time) timeRangeExpressionEvaluator {

	base := tr.buildBaseTimeRangeExpressionEvaluator(targetTime)
	switch tr.Frequency {
	case Daily:
		return &DailyTimeRangeExpressionEvaluator{base}
	case Weekly:
		return &WeeklyTimeRangeExpressionEvaluator{base}
	case WeeklyRange:
		return &WeeklyRangeTimeRangeExpressionEvaluator{base}
	case Monthly:
		return &MonthlyTimeRangeExpressionEvaluator{base}
	}
	return nil
}

func (tr TimeRange) buildBaseTimeRangeExpressionEvaluator(time time.Time) BaseTimeRangeExpressionEvaluator {
	hour, minute := parseHourMinute(tr.HourMinuteFrom)
	return BaseTimeRangeExpressionEvaluator{
		TimeRange:  tr,
		TargetTime: time,
		parsedValues: parsedValues{
			Hour:   hour,
			Minute: minute,
		},
	}

}

type timeRangeExpressionEvaluator interface {
	getCron() string
	getDuration() time.Duration
	getDurationOfPreviousWindow(time.Duration) time.Duration
}

//change name

type parsedValues struct {
	Hour   string
	Minute string
}

type BaseTimeRangeExpressionEvaluator struct {
	TimeRange  TimeRange
	TargetTime time.Time
	parsedValues
}

type DailyTimeRangeExpressionEvaluator struct {
	BaseTimeRangeExpressionEvaluator
}

type WeeklyTimeRangeExpressionEvaluator struct {
	BaseTimeRangeExpressionEvaluator
}

type WeeklyRangeTimeRangeExpressionEvaluator struct {
	BaseTimeRangeExpressionEvaluator
}

type MonthlyTimeRangeExpressionEvaluator struct {
	BaseTimeRangeExpressionEvaluator
}

func (td DailyTimeRangeExpressionEvaluator) getCron() string {
	return dailyCron(td.parsedValues.Minute, td.parsedValues.Hour)
}

func (td DailyTimeRangeExpressionEvaluator) getDuration() time.Duration {
	return td.getDurationForHourMinute()
}

func (tw WeeklyTimeRangeExpressionEvaluator) getCron() string {
	return weeklyCron(tw.parsedValues.Minute, tw.parsedValues.Hour, tw.TimeRange.Weekdays)
}

func (tw WeeklyTimeRangeExpressionEvaluator) getDuration() time.Duration {
	return tw.getDurationForHourMinute()
}

func (twr WeeklyRangeTimeRangeExpressionEvaluator) getCron() string {
	return weeklyRangeCron(twr.parsedValues.Minute, twr.parsedValues.Hour, toString(twr.TimeRange.WeekdayFrom))
}

func (twr WeeklyRangeTimeRangeExpressionEvaluator) getDuration() time.Duration {
	return twr.getDurationBetweenWeekdays()
}

func (tm MonthlyTimeRangeExpressionEvaluator) getCron() string {

	return monthlyCron(tm.parsedValues.Minute, tm.parsedValues.Hour, tm.TimeRange.DayFrom, tm.BaseTimeRangeExpressionEvaluator.calculateLastDayOfMonth())
}

func (tm MonthlyTimeRangeExpressionEvaluator) getDuration() time.Duration {
	return tm.getDurationBetweenWeekDates()
}
