package timeRangeLib

import (
	"strconv"
	"strings"
	"time"
)

func parseHourMinute(hourMin string) (string, string) {
	return strings.Split(hourMin, ":")[0], strings.Split(hourMin, ":")[1]
}

func toString(weekday time.Weekday) string {
	return strconv.Itoa(int(weekday))
}

func weekdaysToString(weekdays []time.Weekday) string {
	days := ""
	for _, day := range weekdays {
		days += toString(day) + ","
	}
	return days[:len(days)-1]
}

func calculateDaysBetweenWeekdays(from, to int) int {
	days := to - from
	if days < 0 {
		days += 7
	}
	return days
}

func getPreviousMonthAndYear(currentMonth time.Month, currentYear int) (time.Month, int) {
	previousMonth := currentMonth - 1
	previousYear := currentYear
	if previousMonth == 0 {
		previousMonth = 12
		previousYear = currentYear - 1
	}
	return previousMonth, previousYear
}
