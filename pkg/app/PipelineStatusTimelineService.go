package app

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"go.uber.org/zap"
	"time"
)

type PipelineStatusTimelineService interface {
	FetchTimelinesForLatestTriggerByAppIdAndEnvId(appId, envId int) ([]*PipelineStatusTimelineDto, error)
	FetchTimelinesByWfrId(wfrId int) ([]*PipelineStatusTimelineDto, error)
}

type PipelineStatusTimelineServiceImpl struct {
	logger                           *zap.SugaredLogger
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository
	cdWorkflowRepository             pipelineConfig.CdWorkflowRepository
}

func NewPipelineStatusTimelineServiceImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository) *PipelineStatusTimelineServiceImpl {
	return &PipelineStatusTimelineServiceImpl{
		logger:                           logger,
		pipelineStatusTimelineRepository: pipelineStatusTimelineRepository,
		cdWorkflowRepository:             cdWorkflowRepository,
	}
}

type PipelineStatusTimelineDto struct {
	Id                           int                           `json:"id"`
	InstalledAppVersionHistoryId int                           `json:"InstalledAppVersionHistoryId,omitempty"`
	CdWorkflowRunnerId           int                           `json:"cdWorkflowRunnerId"`
	Status                       pipelineConfig.TimelineStatus `json:"status"`
	StatusDetail                 string                        `json:"statusDetail"`
	StatusTime                   time.Time                     `json:"statusTime"`
}

func (impl *PipelineStatusTimelineServiceImpl) FetchTimelinesByWfrId(wfrId int) ([]*PipelineStatusTimelineDto, error) {
	timelines, err := impl.pipelineStatusTimelineRepository.FetchTimelinesByWfrId(wfrId)
	if err != nil {
		impl.logger.Errorw("error in getting timelines by wfrId", "err", err, "wfrId", wfrId)
		return nil, err
	}
	var timelineDtos []*PipelineStatusTimelineDto
	for _, timeline := range timelines {
		timelineDto := &PipelineStatusTimelineDto{
			Id:                 timeline.Id,
			CdWorkflowRunnerId: timeline.CdWorkflowRunnerId,
			Status:             timeline.Status,
			StatusTime:         timeline.StatusTime,
			StatusDetail:       timeline.StatusDetail,
		}
		timelineDtos = append(timelineDtos, timelineDto)
	}
	return timelineDtos, nil
}

func (impl *PipelineStatusTimelineServiceImpl) FetchTimelinesForLatestTriggerByAppIdAndEnvId(appId, envId int) ([]*PipelineStatusTimelineDto, error) {
	//fetch latest wfr id by app and env
	wfr, err := impl.cdWorkflowRepository.FindLatestWfrByAppIdAndEnvironmentId(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting wfr by appId and envId", "err", err, "appId", appId, "envId", envId)
		return nil, err
	}
	timelines, err := impl.pipelineStatusTimelineRepository.FetchTimelinesByWfrId(wfr.Id)
	if err != nil {
		impl.logger.Errorw("error in getting timelines by wfrId", "err", err, "wfrId", wfr.Id)
		return nil, err
	}

	var timelineDtos []*PipelineStatusTimelineDto
	for _, timeline := range timelines {
		timelineDto := &PipelineStatusTimelineDto{
			Id:                 timeline.Id,
			CdWorkflowRunnerId: timeline.CdWorkflowRunnerId,
			Status:             timeline.Status,
			StatusTime:         timeline.StatusTime,
			StatusDetail:       timeline.StatusDetail,
		}
		timelineDtos = append(timelineDtos, timelineDto)
	}
	return timelineDtos, nil
}
