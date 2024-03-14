package timeoutWindow

import (
	"encoding/json"
	scheduler "github.com/devtron-labs/common-lib/timeRangeLib"
	"github.com/samber/lo"
	"time"
)

type Frequency string

const (
	Fixed       Frequency = "FIXED"
	Daily       Frequency = "DAILY"
	Monthly     Frequency = "MONTHLY"
	Weekly      Frequency = "WEEKLY"
	WeeklyRange Frequency = "WEEKLY_RANGE"
)

// TimeWindow defines model for TimeWindow.
type TimeWindow struct {
	//Id        int       `json:"id"`
	Frequency Frequency `json:"frequency"`

	// relevant for daily and monthly
	DayFrom int `json:"dayFrom"`
	DayTo   int `json:"dayTo"`

	// relevant for
	HourMinuteFrom string `json:"hourMinuteFrom"`
	HourMinuteTo   string `json:"hourMinuteTo"`

	// optional for frequencies other than FIXED, otherwise required
	TimeFrom time.Time `json:"timeFrom"`
	TimeTo   time.Time `json:"timeTo"`

	// relevant for weekly range
	WeekdayFrom DayOfWeek `json:"weekdayFrom"`
	WeekdayTo   DayOfWeek `json:"weekdayTo"`

	// relevant for weekly
	Weekdays []DayOfWeek `json:"weekdays"`
}

func (window *TimeWindow) toJsonString() string {
	marshal, err := json.Marshal(window)
	if err != nil {
		return ""
	}
	return string(marshal)
}
func (window *TimeWindow) setFromJsonString(jsonString string) {
	json.Unmarshal([]byte(jsonString), window)
}

func (timeWindow *TimeWindow) toTimeRange() scheduler.TimeRange {
	return scheduler.TimeRange{
		TimeFrom:       timeWindow.TimeFrom,
		TimeTo:         timeWindow.TimeTo,
		HourMinuteFrom: timeWindow.HourMinuteFrom,
		HourMinuteTo:   timeWindow.HourMinuteTo,
		DayFrom:        timeWindow.DayFrom,
		DayTo:          timeWindow.DayTo,
		WeekdayFrom:    timeWindow.WeekdayFrom.toWeekday(),
		WeekdayTo:      timeWindow.WeekdayTo.toWeekday(),
		Weekdays:       lo.Map(timeWindow.Weekdays, func(item DayOfWeek, index int) time.Weekday { return item.toWeekday() }),
		Frequency:      timeWindow.Frequency.toTimeRangeFrequency(),
	}
}

func (f Frequency) toTimeRangeFrequency() scheduler.Frequency {
	switch f {
	case Fixed:
		return scheduler.Fixed
	case Daily:
		return scheduler.Daily
	case Weekly:
		return scheduler.Weekly
	case WeeklyRange:
		return scheduler.WeeklyRange
	case Monthly:
		return scheduler.Monthly
	}
	return ""
}

type DayOfWeek string

const (
	Sunday    DayOfWeek = "Sunday"
	Monday    DayOfWeek = "Monday"
	Tuesday   DayOfWeek = "Tuesday"
	Wednesday DayOfWeek = "Wednesday"
	Thursday  DayOfWeek = "Thursday"
	Friday    DayOfWeek = "Friday"
	Saturday  DayOfWeek = "Saturday"
)

func (day DayOfWeek) toWeekday() time.Weekday {
	switch day {
	case Sunday:
		return time.Weekday(0)
	case Monday:
		return time.Weekday(1)
	case Tuesday:
		return time.Weekday(2)
	case Wednesday:
		return time.Weekday(3)
	case Thursday:
		return time.Weekday(4)
	case Friday:
		return time.Weekday(5)
	case Saturday:
		return time.Weekday(6)
	}
	return time.Weekday(-1)
}
