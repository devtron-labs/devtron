/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bean

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/timelineStatus"
)

type TimelineGetRequest struct {
	cdWfrId           int
	excludingStatuses []timelineStatus.TimelineStatus
}

func NewTimelineGetRequest() *TimelineGetRequest {
	return &TimelineGetRequest{}
}

func (req *TimelineGetRequest) GetCdWfrId() int {
	return req.cdWfrId
}

func (req *TimelineGetRequest) GetExcludingStatuses() []timelineStatus.TimelineStatus {
	return req.excludingStatuses
}

func (req *TimelineGetRequest) WithCdWfrId(id int) *TimelineGetRequest {
	req.cdWfrId = id
	return req
}

func (req *TimelineGetRequest) ExcludingStatuses(statuses ...timelineStatus.TimelineStatus) *TimelineGetRequest {
	req.excludingStatuses = append(req.excludingStatuses, statuses...)
	return req
}
