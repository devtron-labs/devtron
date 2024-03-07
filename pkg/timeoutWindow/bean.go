package timeoutWindow

import (
	"encoding/json"
	"github.com/devtron-labs/common-lib/scheduler"
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
		return scheduler.FIXED
	case Daily:
		return scheduler.DAILY
	case Weekly:
		return scheduler.WEEKLY
	case WeeklyRange:
		return scheduler.WEEKLY_RANGE
	case Monthly:
		return scheduler.MONTHLY
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
		return time.Weekday(0)
	case Tuesday:
		return time.Weekday(0)
	case Wednesday:
		return time.Weekday(0)
	case Thursday:
		return time.Weekday(0)
	case Friday:
		return time.Weekday(0)
	case Saturday:
		return time.Weekday(0)
	}
	return time.Weekday(-1)
}
