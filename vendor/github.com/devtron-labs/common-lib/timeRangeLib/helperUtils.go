package timeRangeLib

import (
	"strconv"
	"time"
)

func isToHourMinuteBeforeWindowEnd(hourMinute string, targetTime time.Time) bool {

	currentHourMinute, _ := time.Parse(hourMinuteFormat, targetTime.Format(hourMinuteFormat))
	parsedHourTo, _ := time.Parse(hourMinuteFormat, hourMinute)

	return currentHourMinute.Before(parsedHourTo)
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

func isTimeInBetween(timeCurrent, periodStart, periodEnd time.Time) bool {
	return (timeCurrent.After(periodStart) && timeCurrent.Before(periodEnd)) || timeCurrent.Equal(periodStart)
}
