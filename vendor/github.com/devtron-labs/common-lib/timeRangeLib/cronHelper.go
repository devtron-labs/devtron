package timeRangeLib

import (
	"fmt"
	"strconv"
	"time"
)

func dailyCron(minute, hour string) string {
	return fmt.Sprintf("%s %s * * *", minute, hour)
}

func weeklyCron(minute, hour string, weekdays []time.Weekday) string {
	days := weekdaysToString(weekdays)
	return fmt.Sprintf("%s %s * * %s", minute, hour, days)
}

func weeklyRangeCron(minute, hour string, weekdayFrom string) string {
	return fmt.Sprintf("%s %s * * %s", minute, hour, weekdayFrom)
}

func monthlyCron(minute, hour string, dayFrom int, lastDayOfMonth int) string {
	if dayFrom < 0 {
		dayFrom = getDayForNegativeValueInMonth(dayFrom, lastDayOfMonth)
	}
	day := strconv.Itoa(dayFrom)

	return fmt.Sprintf("%s %s %s * *", minute, hour, day)
}

func getDayForNegativeValueInMonth(dayFrom int, lastDayOfMonth int) int {
	// example for April last day is 30, so -2(second last day) will be 30+1-2=29
	return lastDayOfMonth + 1 + dayFrom
}
