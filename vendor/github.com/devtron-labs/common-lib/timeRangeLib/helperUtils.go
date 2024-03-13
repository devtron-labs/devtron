package timeRangeLib

import (
	"strconv"
	"time"
)

func (tr TimeRange) isMonthOverlapping() bool {
	dayFrom := tr.DayFrom
	dayTo := tr.DayTo
	if dayFrom > 0 && dayTo > 0 && dayTo < dayFrom {
		return true
	} else if dayFrom < 0 && dayTo > 0 {
		return true
	}
	return false
}

func (tr TimeRange) isToHourMinuteBeforeWindowEnd(targetTime time.Time) bool {

	currentHourMinute, _ := time.Parse(hourMinuteFormat, targetTime.Format(hourMinuteFormat))

	parsedHourTo, _ := time.Parse(hourMinuteFormat, tr.HourMinuteTo)

	return currentHourMinute.Before(parsedHourTo)
}

func (tr TimeRange) getDaysCount(monthEnd int) int {

	windowEndDay := tr.DayTo
	if windowEndDay < 0 {
		windowEndDay = monthEnd + 1 + windowEndDay
	}

	windowStartDay := tr.DayFrom
	if windowStartDay < 0 {
		windowStartDay = monthEnd + 1 + windowStartDay
	}

	totalDays := windowEndDay - windowStartDay
	if tr.isMonthOverlapping() {
		totalDays = totalDays + monthEnd
	}
	return totalDays
}

func getLastDayOfMonth(targetYear int, targetMonth time.Month) int {
	firstDayOfNextMonth := time.Date(targetYear, targetMonth+1, 1, 0, 0, 0, 0, time.UTC)
	lastDayOfMonth := firstDayOfNextMonth.Add(-time.Hour * 24).Day()
	return lastDayOfMonth
}

func constructDateTime(hourMinute string, days int) time.Time {
	now := time.Now()
	dateTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	hour, minute := parseHourMinute(hourMinute)
	fromHour, _ := strconv.Atoi(hour)
	fromMinute, _ := strconv.Atoi(minute)
	dateTime = dateTime.Add(time.Duration(fromHour+24*days)*time.Hour + time.Duration(fromMinute)*time.Minute)
	return dateTime
}

func isToBeforeFrom(from, to string) bool {
	parseHourFrom, _ := time.Parse(hourMinuteFormat, from)
	parsedHourTo, _ := time.Parse(hourMinuteFormat, to)
	return parsedHourTo.Before(parseHourFrom)
}
