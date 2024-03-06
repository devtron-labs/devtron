package scheduler

import (
	"github.com/robfig/cron/v3"
	"strings"
	"time"
)

func (tr TimeRange) GetScheduleSpec(targetTime time.Time) (nextWindowEdge time.Time, isTimeBetween bool, err error) {
	var windowEnd time.Time
	err = tr.ValidateTimeRange()
	if err != nil {
		return nextWindowEdge, false, err
	}
	if tr.Frequency == FIXED {
		nextWindowEdge, isTimeBetween = getScheduleForFixedTime(targetTime, tr)
		return nextWindowEdge, isTimeBetween, err
	}
	month := targetTime.Month()
	year := targetTime.Year()
	day := targetTime.Day()
	cronExp := tr.getCron()
	if day >= 1 && day < tr.DayTo && tr.isCyclic() {
		if month == 1 {
			month = 12
			year = year - 1
		} else {
			month = month - 1
		}
	}
	lastDayOfMonth := getLastDayOfMonth(year, month, time.Now().Local())
	if strings.Contains(cronExp, "L-2") {
		lastDayOfMonth = lastDayOfMonth - 2
		cronExp = strings.Replace(cronExp, "L-2", intToString(lastDayOfMonth), -1)
	} else if strings.Contains(cronExp, "L-1") {
		lastDayOfMonth = lastDayOfMonth - 1
		cronExp = strings.Replace(cronExp, "L-1", intToString(lastDayOfMonth), -1)
	} else {
		cronExp = strings.Replace(cronExp, "L", intToString(lastDayOfMonth), -1)
	}
	parser := cron.NewParser(CRON)
	schedule, err := parser.Parse(cronExp)
	if err != nil {
		return nextWindowEdge, false, err
	}

	duration, err := tr.getDuration(month, year)
	if err != nil {
		return nextWindowEdge, false, err
	}
	timeMinusDuration := targetTime.Add(-1 * duration)
	windowStart := schedule.Next(timeMinusDuration)
	windowEnd = windowStart.Add(duration)
	if !tr.TimeFrom.IsZero() && windowStart.Before(tr.TimeFrom) {
		windowStart = tr.TimeFrom
	}
	if !tr.TimeTo.IsZero() && windowEnd.After(tr.TimeTo) {
		windowEnd = tr.TimeTo
	}
	if isTimeInBetween(targetTime, windowStart, windowEnd) {
		return windowEnd, true, err
	}
	return windowStart, false, err
}

func isTimeInBetween(timeCurrent, periodStart, periodEnd time.Time) bool {
	return (timeCurrent.After(periodStart) && timeCurrent.Before(periodEnd)) || timeCurrent.Equal(periodStart)
}
