package app

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStatusTimelineService interface {
	SaveTimeline(timeline *pipelineConfig.PipelineStatusTimeline, tx *pg.Tx) error
	FetchTimelines(appId, envId, wfrId int) (*PipelineTimelineDetailDto, error)
}

type PipelineStatusTimelineServiceImpl struct {
	logger                                 *zap.SugaredLogger
	pipelineStatusTimelineRepository       pipelineConfig.PipelineStatusTimelineRepository
	cdWorkflowRepository                   pipelineConfig.CdWorkflowRepository
	userService                            user.UserService
	pipelineStatusTimelineResourcesService PipelineStatusTimelineResourcesService
	pipelineStatusSyncDetailService        PipelineStatusSyncDetailService
}

func NewPipelineStatusTimelineServiceImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	userService user.UserService,
	pipelineStatusTimelineResourcesService PipelineStatusTimelineResourcesService,
	pipelineStatusSyncDetailService PipelineStatusSyncDetailService,
) *PipelineStatusTimelineServiceImpl {
	return &PipelineStatusTimelineServiceImpl{
		logger:                                 logger,
		pipelineStatusTimelineRepository:       pipelineStatusTimelineRepository,
		cdWorkflowRepository:                   cdWorkflowRepository,
		userService:                            userService,
		pipelineStatusTimelineResourcesService: pipelineStatusTimelineResourcesService,
		pipelineStatusSyncDetailService:        pipelineStatusSyncDetailService,
	}
}

type PipelineTimelineDetailDto struct {
	DeploymentStartedOn  time.Time                    `json:"deploymentStartedOn"`
	DeploymentFinishedOn time.Time                    `json:"deploymentFinishedOn"`
	TriggeredBy          string                       `json:"triggeredBy"`
	Timelines            []*PipelineStatusTimelineDto `json:"timelines"`
	StatusLastFetchedAt  time.Time                    `json:"statusLastFetchedAt"`
	StatusFetchCount     int                          `json:"statusFetchCount"`
	WfrStatus            string                       `json:"wfrStatus"`
}

type PipelineStatusTimelineDto struct {
	Id                           int                           `json:"id"`
	InstalledAppVersionHistoryId int                           `json:"InstalledAppVersionHistoryId,omitempty"`
	CdWorkflowRunnerId           int                           `json:"cdWorkflowRunnerId"`
	Status                       pipelineConfig.TimelineStatus `json:"status"`
	StatusDetail                 string                        `json:"statusDetail"`
	StatusTime                   time.Time                     `json:"statusTime"`
	ResourceDetails              []*SyncStageResourceDetailDto `json:"resourceDetails,omitempty"`
}

func (impl *PipelineStatusTimelineServiceImpl) SaveTimeline(timeline *pipelineConfig.PipelineStatusTimeline, tx *pg.Tx) error {
	//get unableToFetch or timedOut timeline
	redundantTimelines, err := impl.pipelineStatusTimelineRepository.FetchTimelineByWfrIdAndStatuses(timeline.CdWorkflowRunnerId, []pipelineConfig.TimelineStatus{pipelineConfig.TIMELINE_STATUS_UNABLE_TO_FETCH_STATUS, pipelineConfig.TIMELINE_STATUS_FETCH_TIMED_OUT})
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting unableToFetch/timedOut timelines", "err", err, "cdWfrId", timeline.CdWorkflowRunnerId)
		return err
	}
	if len(redundantTimelines) > 1 {
		return fmt.Errorf("multiple unableToFetch/timedOut timelines found")
	} else if len(redundantTimelines) == 1 && redundantTimelines[0].Id > 0 {
		timeline.Id = redundantTimelines[0].Id
	} else {
		// do nothing
	}

	//saving/updating timeline
	err = impl.saveOrUpdateTimeline(timeline, tx)
	if err != nil {
		impl.logger.Errorw("error in saving/updating timeline", "err", err, "timeline", timeline)
		return err
	}
	return nil
}

func (impl *PipelineStatusTimelineServiceImpl) saveOrUpdateTimeline(timeline *pipelineConfig.PipelineStatusTimeline, tx *pg.Tx) error {
	if tx == nil {
		if timeline.Id > 0 {
			err := impl.pipelineStatusTimelineRepository.UpdateTimelines([]*pipelineConfig.PipelineStatusTimeline{timeline})
			if err != nil {
				impl.logger.Errorw("error in updating timeline", "err", err, "timeline", timeline)
				return err
			}
		} else {
			err := impl.pipelineStatusTimelineRepository.SaveTimelines([]*pipelineConfig.PipelineStatusTimeline{timeline})
			if err != nil {
				impl.logger.Errorw("error in saving timeline", "err", err, "timeline", timeline)
				return err
			}
		}
	} else {
		if timeline.Id > 0 {
			err := impl.pipelineStatusTimelineRepository.UpdateTimelinesWithTxn([]*pipelineConfig.PipelineStatusTimeline{timeline}, tx)
			if err != nil {
				impl.logger.Errorw("error in updating timeline", "err", err, "timeline", timeline)
				return err
			}
		} else {
			err := impl.pipelineStatusTimelineRepository.SaveTimelinesWithTxn([]*pipelineConfig.PipelineStatusTimeline{timeline}, tx)
			if err != nil {
				impl.logger.Errorw("error in saving timeline", "err", err, "timeline", timeline)
				return err
			}
		}
	}
	return nil
}

func (impl *PipelineStatusTimelineServiceImpl) FetchTimelines(appId, envId, wfrId int) (*PipelineTimelineDetailDto, error) {
	var triggeredBy int32
	var deploymentStartedOn time.Time
	var deploymentFinishedOn time.Time
	var wfrStatus string
	var deploymentAppType string
	var err error
	wfr := &pipelineConfig.CdWorkflowRunner{}
	if wfrId == 0 {
		//fetch latest wfr by app and env
		wfr, err = impl.cdWorkflowRepository.FindLatestWfrByAppIdAndEnvironmentId(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in getting wfr by appId and envId", "err", err, "appId", appId, "envId", envId)
			return nil, err
		}
		wfrId = wfr.Id
	} else {
		//fetch latest wfr by id
		wfr, err = impl.cdWorkflowRepository.FindWorkflowRunnerById(wfrId)
		if err != nil {
			impl.logger.Errorw("error in getting wfr by appId and envId", "err", err, "appId", appId, "envId", envId)
			return nil, err
		}
	}
	deploymentStartedOn = wfr.StartedOn
	deploymentFinishedOn = wfr.FinishedOn
	triggeredBy = wfr.TriggeredBy
	wfrStatus = wfr.Status
	deploymentAppType = wfr.CdWorkflow.Pipeline.DeploymentAppType
	triggeredByUser, err := impl.userService.GetById(triggeredBy)
	if err != nil {
		impl.logger.Errorw("error in getting user detail by id", "err", err, "userId", triggeredBy)
		return nil, err
	}
	var timelineDtos []*PipelineStatusTimelineDto
	var statusLastFetchedAt time.Time
	var statusFetchCount int
	if util.IsAcdApp(deploymentAppType) {
		timelines, err := impl.pipelineStatusTimelineRepository.FetchTimelinesByWfrId(wfrId)
		if err != nil {
			impl.logger.Errorw("error in getting timelines by wfrId", "err", err, "wfrId", wfrId)
			return nil, err
		}
		for _, timeline := range timelines {
			var timelineResourceDetails []*SyncStageResourceDetailDto
			if timeline.Status == pipelineConfig.TIMELINE_STATUS_KUBECTL_APPLY_STARTED {
				timelineResourceDetails, err = impl.pipelineStatusTimelineResourcesService.GetTimelineResourcesForATimeline(timeline.CdWorkflowRunnerId)
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in getting timeline resources details", "err", err, "cdWfrId", timeline.CdWorkflowRunnerId)
					return nil, err
				}
			}
			timelineDto := &PipelineStatusTimelineDto{
				Id:                 timeline.Id,
				CdWorkflowRunnerId: timeline.CdWorkflowRunnerId,
				Status:             timeline.Status,
				StatusTime:         timeline.StatusTime,
				StatusDetail:       timeline.StatusDetail,
				ResourceDetails:    timelineResourceDetails,
			}
			timelineDtos = append(timelineDtos, timelineDto)
		}
		statusLastFetchedAt, statusFetchCount, err = impl.pipelineStatusSyncDetailService.GetSyncTimeAndCountByCdWfrId(wfrId)
		if err != nil {
			impl.logger.Errorw("error in getting pipeline status fetchTime and fetchCount by cdWfrId", "err", err, "cdWfrId", wfrId)
		}
	}
	timelineDetail := &PipelineTimelineDetailDto{
		TriggeredBy:          triggeredByUser.EmailId,
		DeploymentStartedOn:  deploymentStartedOn,
		DeploymentFinishedOn: deploymentFinishedOn,
		Timelines:            timelineDtos,
		StatusLastFetchedAt:  statusLastFetchedAt,
		StatusFetchCount:     statusFetchCount,
		WfrStatus:            wfrStatus,
	}
	return timelineDetail, nil
}
