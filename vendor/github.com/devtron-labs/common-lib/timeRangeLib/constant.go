package timeRangeLib

type ErrorMessage string

const (
	InvalidFrequencyType                 ErrorMessage = "invalid Frequency type"
	HourMinuteNotPresent                 ErrorMessage = "HourMinuteFrom and HourMinuteTo must be present for Daily frequency"
	TimeFromOrToNotPresent               ErrorMessage = "TimeFrom and TimeTo must be present for Fixed frequency"
	TimeFromLessThanTimeTo               ErrorMessage = "TimeFrom must be less than TimeTo for Fixed frequency"
	TimeFromEqualToTimeTo                ErrorMessage = "TimeFrom must not be equal to TimeTo for Fixed frequency"
	WeekDayOutsideRange                  ErrorMessage = "one or both of the values are outside the range of 0 to 6"
	WeekDaysNotPresent                   ErrorMessage = "weekdays, must be present for Weekly frequency"
	WeekDayFromOrToNotPresent            ErrorMessage = "WeekdayFrom, must be present for WeeklyRange frequency"
	DayFromOrToNotPresent                ErrorMessage = "DayFrom, DayTo, must be present for Monthly frequency"
	ToBeforeFrom                         ErrorMessage = "Invalid value of hourMinuteFrom or hourMinuteTo  for same day ,hourMinuteFrom >hourMinuteTo"
	BothLessThanZeroAndFromGreaterThanTo ErrorMessage = "invalid value of DayFrom or DayTo,DayFrom and DayTo is less than zero and  DayFrom > DayTo"
	DayFromOrToNotValid                  ErrorMessage = "invalid value of DayFrom or DayTo"
)
