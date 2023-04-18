package status

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/sync/common"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStatusTimelineResourcesService interface {
	SaveOrUpdatePipelineTimelineResources(pipelineId int, application *v1alpha1.Application, tx *pg.Tx, userId int32, isAppStore bool) error
	GetTimelineResourcesForATimeline(cdWfrIds []int) (map[int][]*SyncStageResourceDetailDto, error)
	GetTimelineResourcesForATimelineForAppStore(installedAppVersionHistoryId int) ([]*SyncStageResourceDetailDto, error)
}

type PipelineStatusTimelineResourcesServiceImpl struct {
	dbConnection                              *pg.DB
	logger                                    *zap.SugaredLogger
	pipelineStatusTimelineResourcesRepository pipelineConfig.PipelineStatusTimelineResourcesRepository
}

func NewPipelineStatusTimelineResourcesServiceImpl(dbConnection *pg.DB, logger *zap.SugaredLogger,
	pipelineStatusTimelineResourcesRepository pipelineConfig.PipelineStatusTimelineResourcesRepository) *PipelineStatusTimelineResourcesServiceImpl {
	return &PipelineStatusTimelineResourcesServiceImpl{
		dbConnection: dbConnection,
		logger:       logger,
		pipelineStatusTimelineResourcesRepository: pipelineStatusTimelineResourcesRepository,
	}
}

type SyncStageResourceDetailDto struct {
	Id                           int                                  `json:"id"`
	InstalledAppVersionHistoryId int                                  `json:"installedAppVersionHistoryId,omitempty"`
	CdWorkflowRunnerId           int                                  `json:"cdWorkflowRunnerId,omitempty"`
	ResourceName                 string                               `json:"resourceName"`
	ResourceKind                 string                               `json:"resourceKind"`
	ResourceGroup                string                               `json:"resourceGroup"`
	ResourceStatus               string                               `json:"resourceStatus"`
	ResourcePhase                string                               `json:"resourcePhase"`
	StatusMessage                string                               `json:"statusMessage"`
	TimelineStage                pipelineConfig.ResourceTimelineStage `json:"timelineStage,omitempty"`
}

func (impl *PipelineStatusTimelineResourcesServiceImpl) SaveOrUpdatePipelineTimelineResources(pipelineId int, application *v1alpha1.Application, tx *pg.Tx, userId int32, isAppStore bool) error {
	var err error
	var timelineResources []*pipelineConfig.PipelineStatusTimelineResources
	if isAppStore {
		//getting all timeline resources by installedAppVersionHistoryId
		timelineResources, err = impl.pipelineStatusTimelineResourcesRepository.GetByInstalledAppVersionHistoryIdAndTimelineStage(pipelineId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting timelineResources for installedAppVersionHistoryId", "err", err, "installedAppVersionHistoryId", pipelineId)
			return err
		}
	} else {
		//getting all timeline resources by cdWfrId
		timelineResources, err = impl.pipelineStatusTimelineResourcesRepository.GetByCdWfrIdAndTimelineStage(pipelineId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting timelineResources for wfrId", "err", err, "wfrId", pipelineId)
			return err
		}
	}

	//map of resourceName and its index
	oldTimelineResourceMap := make(map[string]int)

	for i, timelineResource := range timelineResources {
		oldTimelineResourceMap[timelineResource.ResourceName] = i
	}

	var timelineResourcesToBeSaved []*pipelineConfig.PipelineStatusTimelineResources
	var timelineResourcesToBeUpdated []*pipelineConfig.PipelineStatusTimelineResources

	if application != nil && application.Status.OperationState != nil && application.Status.OperationState.SyncResult != nil {
		for _, resource := range application.Status.OperationState.SyncResult.Resources {
			if resource != nil {
				if index, ok := oldTimelineResourceMap[resource.Name]; ok {
					timelineResources[index].ResourceStatus = string(resource.HookPhase)
					timelineResources[index].StatusMessage = resource.Message
					timelineResources[index].UpdatedBy = userId
					timelineResources[index].UpdatedOn = time.Now()
					timelineResourcesToBeUpdated = append(timelineResourcesToBeUpdated, timelineResources[index])
				} else {
					newTimelineResource := &pipelineConfig.PipelineStatusTimelineResources{
						ResourceName:   resource.Name,
						ResourceKind:   resource.Kind,
						ResourceGroup:  resource.Group,
						ResourceStatus: string(resource.HookPhase),
						StatusMessage:  resource.Message,
						AuditLog: sql.AuditLog{
							CreatedBy: userId,
							CreatedOn: time.Now(),
							UpdatedBy: userId,
							UpdatedOn: time.Now(),
						},
					}
					if isAppStore {
						newTimelineResource.InstalledAppVersionHistoryId = pipelineId
					} else {
						newTimelineResource.CdWorkflowRunnerId = pipelineId
					}
					if resource.HookType != "" {
						newTimelineResource.ResourcePhase = string(resource.HookType)
					} else {
						//since hookType for non-hook resources is empty and always come under sync phase, hard-coding it
						newTimelineResource.ResourcePhase = string(common.HookTypeSync)
					}
					timelineResourcesToBeSaved = append(timelineResourcesToBeSaved, newTimelineResource)
				}
			}
		}
	}
	if len(timelineResourcesToBeSaved) > 0 {
		if tx != nil {
			err = impl.pipelineStatusTimelineResourcesRepository.SaveTimelineResourcesWithTxn(timelineResourcesToBeSaved, tx)
			if err != nil {
				impl.logger.Errorw("error in saving timelineResources", "err", err, "timelineResources", timelineResourcesToBeSaved)
				return err
			}
		} else {
			err = impl.pipelineStatusTimelineResourcesRepository.SaveTimelineResources(timelineResourcesToBeSaved)
			if err != nil {
				impl.logger.Errorw("error in saving timelineResources", "err", err, "timelineResources", timelineResourcesToBeSaved)
				return err
			}
		}
	}
	if len(timelineResourcesToBeUpdated) > 0 {
		if tx != nil {
			err = impl.pipelineStatusTimelineResourcesRepository.UpdateTimelineResourcesWithTxn(timelineResourcesToBeUpdated, tx)
			if err != nil {
				impl.logger.Errorw("error in updating timelineResources", "err", err, "timelineResources", timelineResourcesToBeUpdated)
				return err
			}
		} else {
			err = impl.pipelineStatusTimelineResourcesRepository.UpdateTimelineResources(timelineResourcesToBeUpdated)
			if err != nil {
				impl.logger.Errorw("error in updating timelineResources", "err", err, "timelineResources", timelineResourcesToBeUpdated)
				return err
			}
		}
	}
	return nil
}

func (impl *PipelineStatusTimelineResourcesServiceImpl) GetTimelineResourcesForATimeline(cdWfrIds []int) (map[int][]*SyncStageResourceDetailDto, error) {
	timelineResources, err := impl.pipelineStatusTimelineResourcesRepository.GetByCdWfrIds(cdWfrIds)
	if err != nil {
		impl.logger.Errorw("error in getting timeline resources", "err", err, "cdWfrIds", cdWfrIds)
		return nil, err
	}
	timelineResourcesMap := make(map[int][]*SyncStageResourceDetailDto)
	var timelineResourcesDtos []*SyncStageResourceDetailDto
	for _, timelineResource := range timelineResources {
		dto := &SyncStageResourceDetailDto{
			Id:                 timelineResource.Id,
			CdWorkflowRunnerId: timelineResource.CdWorkflowRunnerId,
			ResourceKind:       timelineResource.ResourceKind,
			ResourceName:       timelineResource.ResourceName,
			ResourceGroup:      timelineResource.ResourceGroup,
			ResourceStatus:     timelineResource.ResourceStatus,
			ResourcePhase:      timelineResource.ResourcePhase,
			StatusMessage:      timelineResource.StatusMessage,
		}
		timelineResourcesDtos = append(timelineResourcesDtos, dto)
		timelineResourcesMap[timelineResource.CdWorkflowRunnerId] = append(timelineResourcesMap[timelineResource.CdWorkflowRunnerId], dto)
	}
	return timelineResourcesMap, nil
}

func (impl *PipelineStatusTimelineResourcesServiceImpl) GetTimelineResourcesForATimelineForAppStore(installedAppVersionHistoryId int) ([]*SyncStageResourceDetailDto, error) {
	timelineResources, err := impl.pipelineStatusTimelineResourcesRepository.GetByInstalledAppVersionHistoryIdAndTimelineStage(installedAppVersionHistoryId)
	if err != nil {
		impl.logger.Errorw("error in getting timeline resources", "err", err, "installedAppVersionHistoryId", installedAppVersionHistoryId)
		return nil, err
	}
	var timelineResourcesDtos []*SyncStageResourceDetailDto
	for _, timelineResource := range timelineResources {
		dto := &SyncStageResourceDetailDto{
			Id:                           timelineResource.Id,
			InstalledAppVersionHistoryId: timelineResource.InstalledAppVersionHistoryId,
			ResourceKind:                 timelineResource.ResourceKind,
			ResourceName:                 timelineResource.ResourceName,
			ResourceGroup:                timelineResource.ResourceGroup,
			ResourceStatus:               timelineResource.ResourceStatus,
			ResourcePhase:                timelineResource.ResourcePhase,
			StatusMessage:                timelineResource.StatusMessage,
		}
		timelineResourcesDtos = append(timelineResourcesDtos, dto)
	}
	return timelineResourcesDtos, nil
}
