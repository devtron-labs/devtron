/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package timeRangeLib

type ErrorMessage string

const (
	InvalidFrequencyType                 ErrorMessage = "invalid Frequency type"
	HourMinuteNotPresent                 ErrorMessage = "hourMinuteFrom and hourMinuteTo must be present for Daily frequency"
	TimeFromOrToNotPresent               ErrorMessage = "timeFrom and timeTo must be present for Fixed frequency"
	TimeFromLessThanTimeTo               ErrorMessage = "timeFrom must be less than timeTo for Fixed frequency"
	TimeFromEqualToTimeTo                ErrorMessage = "timeFrom must not be equal to timeTo for Fixed frequency"
	WeekDayOutsideRange                  ErrorMessage = "one or both of the values are outside the range of 0 to 6"
	WeekDaysNotPresent                   ErrorMessage = "weekdays, must be present for Weekly frequency"
	WeekDayFromOrToNotPresent            ErrorMessage = "weekdayFrom, must be present for WeeklyRange frequency"
	DayFromOrToNotPresent                ErrorMessage = "dayFrom, dayTo, must be present for Monthly frequency"
	ToBeforeFrom                         ErrorMessage = "Invalid value of hourMinuteFrom or hourMinuteTo  for same day ,hourMinuteFrom >hourMinuteTo"
	BothLessThanZeroAndFromGreaterThanTo ErrorMessage = "invalid value of DayFrom or DayTo,DayFrom and DayTo is less than zero and  dayFrom > dayTo"
	DayFromOrToNotValid                  ErrorMessage = "invalid value of dayFrom or dayTo"
)
