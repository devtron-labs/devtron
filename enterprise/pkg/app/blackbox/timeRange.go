package blackbox

import (
	"time"
)

type TimeRange struct {
	TimeFrom       time.Time
	TimeTo         time.Time
	HourMinuteFrom string
	HourMinuteTo   string
	DayFrom        int
	DayTo          int
	WeekdayFrom    time.Weekday
	WeekdayTo      time.Weekday
	Weekdays       []time.Weekday
	Frequency      Frequency
}

func (timeRange TimeRange) GetScheduleSpec(targetTime time.Time) (time.Time, bool) {
	return time.Time{}, true
}

type Frequency string

const (
	FIXED        Frequency = "FIXED"
	DAILY        Frequency = "DAILY"
	WEEKLY       Frequency = "WEEKLY"
	WEEKLY_RANGE Frequency = "WEEKLY_RANGE"
	MONTHLY      Frequency = "MONTHLY"
)
