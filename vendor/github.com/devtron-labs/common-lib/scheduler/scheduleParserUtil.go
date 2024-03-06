package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func getHour(hourMin string) string {
	return strings.Split(hourMin, ":")[0]
}

func getMinute(hourMin string) string {
	return strings.Split(hourMin, ":")[1]
}

func toString(weekday time.Weekday) string {
	return strconv.Itoa(int(weekday))
}

func intToString(value int) string {
	return strconv.Itoa(value)
}
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

func monthlyCron(minute, hour string, dayFrom int) string {
	day := strconv.Itoa(dayFrom)
	if dayFrom == -1 {
		day = "L"
	} else if dayFrom <= -2 && dayFrom >= -31 {
		day = fmt.Sprintf("L-%s", strconv.Itoa(-dayFrom-1))
	} else {
		day = strconv.Itoa(dayFrom)
	}
	return fmt.Sprintf("%s %s %s * *", minute, hour, day)
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
