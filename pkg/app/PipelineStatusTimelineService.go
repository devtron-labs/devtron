package app

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
	"time"
)

type PipelineStatusTimelineService interface {
	FetchTimelines(appId, envId, wfrId int) (*PipelineTimelineDetailDto, error)
}

type PipelineStatusTimelineServiceImpl struct {
	logger                           *zap.SugaredLogger
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository
	cdWorkflowRepository             pipelineConfig.CdWorkflowRepository
	userService                      user.UserService
}

func NewPipelineStatusTimelineServiceImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	userService user.UserService) *PipelineStatusTimelineServiceImpl {
	return &PipelineStatusTimelineServiceImpl{
		logger:                           logger,
		pipelineStatusTimelineRepository: pipelineStatusTimelineRepository,
		cdWorkflowRepository:             cdWorkflowRepository,
		userService:                      userService,
	}
}

type PipelineTimelineDetailDto struct {
	DeploymentStartedOn  time.Time                    `json:"deploymentStartedOn"`
	DeploymentFinishedOn time.Time                    `json:"deploymentFinishedOn"`
	TriggeredBy          string                       `json:"triggeredBy"`
	Timelines            []*PipelineStatusTimelineDto `json:"timelines"`
}

type PipelineStatusTimelineDto struct {
	Id                           int                           `json:"id"`
	InstalledAppVersionHistoryId int                           `json:"InstalledAppVersionHistoryId,omitempty"`
	CdWorkflowRunnerId           int                           `json:"cdWorkflowRunnerId"`
	Status                       pipelineConfig.TimelineStatus `json:"status"`
	StatusDetail                 string                        `json:"statusDetail"`
	StatusTime                   time.Time                     `json:"statusTime"`
}

func (impl *PipelineStatusTimelineServiceImpl) FetchTimelines(appId, envId, wfrId int) (*PipelineTimelineDetailDto, error) {
	var triggeredBy int32
	var deploymentStartedOn time.Time
	var deploymentFinishedOn time.Time
	if wfrId == 0 {
		//fetch latest wfr by app and env
		wfr, err := impl.cdWorkflowRepository.FindLatestWfrByAppIdAndEnvironmentId(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in getting wfr by appId and envId", "err", err, "appId", appId, "envId", envId)
			return nil, err
		}
		wfrId = wfr.Id
		deploymentStartedOn = wfr.StartedOn
		deploymentFinishedOn = wfr.FinishedOn
		triggeredBy = wfr.TriggeredBy
	} else {
		//fetch latest wfr by id
		wfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(wfrId)
		if err != nil {
			impl.logger.Errorw("error in getting wfr by appId and envId", "err", err, "appId", appId, "envId", envId)
			return nil, err
		}
		deploymentStartedOn = wfr.StartedOn
		deploymentFinishedOn = wfr.FinishedOn
		triggeredBy = wfr.TriggeredBy
	}

	triggeredByUser, err := impl.userService.GetById(triggeredBy)
	if err != nil {
		impl.logger.Errorw("error in getting user detail by id", "err", err, "userId", triggeredBy)
		return nil, err
	}
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
	timelineDetail := &PipelineTimelineDetailDto{
		TriggeredBy:          triggeredByUser.EmailId,
		DeploymentStartedOn:  deploymentStartedOn,
		DeploymentFinishedOn: deploymentFinishedOn,
		Timelines:            timelineDtos,
	}
	return timelineDetail, nil
}
