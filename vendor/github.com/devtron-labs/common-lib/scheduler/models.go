package scheduler

import (
	"github.com/robfig/cron/v3"
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

const parseFormat = "15:04"

type Frequency string

const (
	FIXED        Frequency = "FIXED"
	DAILY        Frequency = "DAILY"
	WEEKLY       Frequency = "WEEKLY"
	WEEKLY_RANGE Frequency = "WEEKLY_RANGE"
	MONTHLY      Frequency = "MONTHLY"
)
const CRON = cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow

//type CronIndex int
//
//const (
//	Minute CronIndex = 0
//	Hour   CronIndex = 1
//	DOM    CronIndex = 2
//	MONTH  CronIndex = 3
//	DOW    CronIndex = 4
//)

/*
cases on frequency
case 1 : FIXED : if all fields are empty other than TimeFrom and TimeTo
then here I have add validation that TimeFrom must be less than TimeTo and TimeFrom TimeTo both should be present.
case 2 : DAILY : HourMinuteFrom and HourMinuteTo must pe present
case 3 : WEEKLY :  Weekdays must be present and HourMinuteFrom and HourMinuteTo must be present
case 4 : WEEKLY_RANGE : WeekdayFrom must be present , HourMinuteFrom and HourMinuteTo must be present
case 5 : MONTHLY :  DayFrom and  DayTo must me present , HourMinuteFrom and HourMinuteTo must be present
*/
/*
cases on field validation
DayFrom and DayTo must be from 1 to 31 as i month max have 31 days
HourMinuteFrom and HourMinuteTo must HH:MM here HH must be from 0 to 23 and MM form 0 to 59
*/
