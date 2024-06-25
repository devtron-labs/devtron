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
