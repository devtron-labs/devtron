package timeRangeLib

import "time"

func (tr TimeRange) getTimeRangeInstant(targetTime time.Time) timeRangeInstant {

	common := tr.buildTimeInstantCommon(targetTime)
	switch tr.Frequency {
	case Daily:
		return &TimeRangeInstantDaily{common}
	case Weekly:
		return &TimeRangeInstantWeekly{common}
	case WeeklyRange:
		return &TimeRangeInstantWeeklyRange{common}
	case Monthly:
		return &TimeRangeInstantMonthly{common, tr.calculateLastDayOfMonth(targetTime)}
	}
	return nil
}

func (tr TimeRange) buildTimeInstantCommon(time time.Time) TimeRangeInstantCommon {
	hour, minute := parseHourMinute(tr.HourMinuteFrom)
	return TimeRangeInstantCommon{
		TimeRange:  tr,
		TargetTime: time,
		TimeRangeInstantCalculated: TimeRangeInstantCalculated{
			Hour:   hour,
			Minute: minute,
		},
	}

}

type timeRangeInstant interface {
	getCron() string
	getDuration() time.Duration
}

type TimeRangeInstantCalculated struct {
	Hour   string
	Minute string
}

type TimeRangeInstantCommon struct {
	TimeRange  TimeRange
	TargetTime time.Time
	TimeRangeInstantCalculated
}

type TimeRangeInstantDaily struct {
	TimeRangeInstantCommon
}

type TimeRangeInstantWeekly struct {
	TimeRangeInstantCommon
}

type TimeRangeInstantWeeklyRange struct {
	TimeRangeInstantCommon
}

type TimeRangeInstantMonthly struct {
	TimeRangeInstantCommon
	lastDayOfMonth int
}

func (td TimeRangeInstantDaily) getCron() string {
	return dailyCron(td.TimeRangeInstantCalculated.Minute, td.TimeRangeInstantCalculated.Hour)
}

func (td TimeRangeInstantDaily) getDuration() time.Duration {
	return td.TimeRange.getDurationForHourMinute()
}

func (tw TimeRangeInstantWeekly) getCron() string {
	return weeklyCron(tw.TimeRangeInstantCalculated.Minute, tw.TimeRangeInstantCalculated.Hour, tw.TimeRange.Weekdays)
}

func (tw TimeRangeInstantWeekly) getDuration() time.Duration {
	return tw.TimeRange.getDurationForHourMinute()
}

func (twr TimeRangeInstantWeeklyRange) getCron() string {
	return weeklyRangeCron(twr.TimeRangeInstantCalculated.Minute, twr.TimeRangeInstantCalculated.Hour, toString(twr.TimeRange.WeekdayFrom))
}

func (twr TimeRangeInstantWeeklyRange) getDuration() time.Duration {
	return twr.TimeRange.getDurationBetweenWeekdays()
}

func (tm TimeRangeInstantMonthly) getCron() string {

	return monthlyCron(tm.TimeRangeInstantCalculated.Minute, tm.TimeRangeInstantCalculated.Hour, tm.TimeRange.DayFrom, tm.lastDayOfMonth)
}

func (tm TimeRangeInstantMonthly) getDuration() time.Duration {
	return tm.TimeRange.getDurationBetweenWeekDates(tm.TargetTime)
}
