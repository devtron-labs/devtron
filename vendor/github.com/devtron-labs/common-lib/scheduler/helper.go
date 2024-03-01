package scheduler

import "time"

func (tr TimeRange) getCron() string {
	minute := getMinute(tr.HourMinuteFrom)
	hour := getHour(tr.HourMinuteFrom)

	switch tr.Frequency {
	case DAILY:
		return dailyCron(minute, hour)
	case WEEKLY:
		return weeklyCron(minute, hour, tr.Weekdays)
	case WEEKLY_RANGE:
		return weeklyRangeCron(minute, hour, tr.WeekdayFrom)
	case MONTHLY:
		return monthlyCron(minute, hour, tr.DayFrom)
	}
	return ""
}

func (tr TimeRange) getDuration(month time.Month, year int) (time.Duration, error) {
	switch tr.Frequency {
	case DAILY:
		return getDurationForHourMinute(tr)
	case WEEKLY:
		return getDurationForHourMinute(tr)
	case WEEKLY_RANGE:
		return getDurationBetweenWeekdays(tr)
	case MONTHLY:
		return getDurationBetweenWeekDates(tr, month, year)
	}
	return 0, nil
}
