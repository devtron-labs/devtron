package timeRangeLib

import (
	"time"
)

func (tr TimeRange) getDurationForHourMinute() time.Duration {

	parsedHourFrom, _ := time.Parse(hourMinuteFormat, tr.HourMinuteFrom)
	parsedHourTo, _ := time.Parse(hourMinuteFormat, tr.HourMinuteTo)
	if parsedHourTo.Before(parsedHourFrom) || parsedHourTo.Equal(parsedHourFrom) {
		parsedHourTo = parsedHourTo.AddDate(0, 0, 1)
	}
	return parsedHourTo.Sub(parsedHourFrom)
}

func (tr TimeRange) getDurationBetweenWeekdays() time.Duration {
	days := calculateDaysBetweenWeekdays(int(tr.WeekdayFrom), int(tr.WeekdayTo))
	fromDateTime := constructDateTime(tr.HourMinuteFrom, 0)
	toDateTime := constructDateTime(tr.HourMinuteTo, days)
	return toDateTime.Sub(fromDateTime)
}

func (tr TimeRange) getDurationBetweenWeekDates(targetTime time.Time) time.Duration {
	lastDayOfMonth := tr.calculateLastDayOfMonth(targetTime)
	days := tr.getDaysCount(lastDayOfMonth)
	fromDateTime := constructDateTime(tr.HourMinuteFrom, 0)
	toDateTime := constructDateTime(tr.HourMinuteTo, days)
	return toDateTime.Sub(fromDateTime)
}
