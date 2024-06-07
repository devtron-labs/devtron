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

import (
	"github.com/robfig/cron/v3"
	"time"
)

// case 1: Fixed frequency:
// Only TimeFrom and TimeTo are allowed.

// case 2: Daily frequency:
// HourMinuteFrom and HourMinuteTo, TimeFrom, and TimeTo are allowed.

// case 3: Weekly frequency:
// Weekdays must be present along with HourMinuteFrom and HourMinuteTo.
// TimeFrom and TimeTo are also allowed.

// case 4: WeeklyRange frequency:
// WeekdayFrom must be present along with HourMinuteFrom and HourMinuteTo.
// TimeFrom and TimeTo are also allowed.

// case 5: Monthly frequency:
// DayFrom and DayTo must be present along with HourMinuteFrom and HourMinuteTo.
// TimeFrom and TimeTo are also allowed.

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

// random values for  for understanding HH:MM format
const hourMinuteFormat = "15:04"

type Frequency string

const (
	Fixed       Frequency = "Fixed"
	Daily       Frequency = "Daily"
	Weekly      Frequency = "Weekly"
	WeeklyRange Frequency = "WeeklyRange"
	Monthly     Frequency = "Monthly"
)

var AllowedFrequencies = []Frequency{Fixed, Daily, Weekly, WeeklyRange, Monthly}

const CRON = cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow
