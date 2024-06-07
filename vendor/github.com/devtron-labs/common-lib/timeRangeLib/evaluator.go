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

import (
	"time"
)

func (evaluator BaseTimeRangeExpressionEvaluator) getDurationForHourMinute() time.Duration {

	parsedHourFrom, _ := time.Parse(hourMinuteFormat, evaluator.TimeRange.HourMinuteFrom)
	parsedHourTo, _ := time.Parse(hourMinuteFormat, evaluator.TimeRange.HourMinuteTo)
	if parsedHourTo.Before(parsedHourFrom) || parsedHourTo.Equal(parsedHourFrom) {
		parsedHourTo = parsedHourTo.AddDate(0, 0, 1)
	}
	return parsedHourTo.Sub(parsedHourFrom)
}

func (evaluator BaseTimeRangeExpressionEvaluator) getDurationBetweenWeekdays() time.Duration {
	days := calculateDaysBetweenWeekdays(int(evaluator.TimeRange.WeekdayFrom), int(evaluator.TimeRange.WeekdayTo))
	fromDateTime := constructDateTime(evaluator.TimeRange.HourMinuteFrom, 0)
	toDateTime := constructDateTime(evaluator.TimeRange.HourMinuteTo, days)
	return toDateTime.Sub(fromDateTime)
}

func (evaluator BaseTimeRangeExpressionEvaluator) getDurationBetweenWeekDates() time.Duration {
	lastDayOfMonth := evaluator.calculateLastDayOfMonth()
	days := evaluator.getDaysCount(lastDayOfMonth)
	fromDateTime := constructDateTime(evaluator.TimeRange.HourMinuteFrom, 0)
	toDateTime := constructDateTime(evaluator.TimeRange.HourMinuteTo, days)
	return toDateTime.Sub(fromDateTime)
}

func (evaluator BaseTimeRangeExpressionEvaluator) getDaysCount(monthEnd int) int {

	windowEndDay := evaluator.TimeRange.DayTo
	if windowEndDay < 0 {
		windowEndDay = monthEnd + 1 + windowEndDay
	}

	windowStartDay := evaluator.TimeRange.DayFrom
	if windowStartDay < 0 {
		windowStartDay = monthEnd + 1 + windowStartDay
	}

	totalDays := windowEndDay - windowStartDay
	if evaluator.isMonthOverlapping() {
		totalDays = totalDays + monthEnd
	}
	return totalDays
}

func (evaluator BaseTimeRangeExpressionEvaluator) calculateLastDayOfMonth() int {
	month, year := evaluator.getMonthAndYearForPreviousWindow()
	return getLastDayOfMonth(year, month)
}

// this will determine if the relevant year and month for the last window happens
// in the same month or previous month
func (evaluator BaseTimeRangeExpressionEvaluator) getMonthAndYearForPreviousWindow() (time.Month, int) {
	month := evaluator.TargetTime.Month()
	year := evaluator.TargetTime.Year()

	if evaluator.isMonthOverlapping() && evaluator.isInsideOverLap() {
		month, year = getPreviousMonthAndYear(month, year)
	}
	return month, year
}

func (evaluator BaseTimeRangeExpressionEvaluator) getDurationOfPreviousWindow(duration time.Duration) time.Duration {
	prevDuration := duration
	if evaluator.isMonthOverlapping() && !evaluator.isInsideOverLap() {
		prevDuration = evaluator.getAdjustedDuration(duration, prevDuration)
	}
	return prevDuration
}

func (evaluator BaseTimeRangeExpressionEvaluator) getAdjustedDuration(duration time.Duration, prevDuration time.Duration) time.Duration {
	//adjusting duration when duration for consecutive windows is different
	currentMonth := evaluator.TargetTime.Month()
	currentYear := evaluator.TargetTime.Year()
	previousMonth, previousYear := getPreviousMonthAndYear(currentMonth, currentYear)
	diff := getLastDayOfMonth(currentYear, currentMonth) - getLastDayOfMonth(previousYear, previousMonth)
	prevDuration = duration - time.Duration(diff)*time.Hour*24
	return prevDuration
}

func (evaluator BaseTimeRangeExpressionEvaluator) isInsideOverLap() bool {
	// for an overlapping window if the current time is on the latter part of the overlap then
	// we use the last month for calculation.
	tr := evaluator.TimeRange
	day := evaluator.TargetTime.Day()
	if day < 1 {
		return false
	}
	return day < tr.DayTo || (day == tr.DayTo && isToHourMinuteBeforeWindowEnd(tr.HourMinuteTo, evaluator.TargetTime))
}

func (evaluator BaseTimeRangeExpressionEvaluator) isMonthOverlapping() bool {
	dayFrom := evaluator.TimeRange.DayFrom
	dayTo := evaluator.TimeRange.DayTo
	if dayFrom > 0 && dayTo > 0 && dayTo < dayFrom {
		return true
		// from is -ve and to is +ve
	} else if dayFrom < 0 && dayTo > 0 {
		return true
	}
	return false
}
