package app

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStatusSyncDetailService interface {
	SaveOrUpdateSyncDetail(cdWfrId int, userId int32) error
	GetSyncTimeAndCountByCdWfrId(cdWfrId int) (time.Time, int, error)
}

type PipelineStatusSyncDetailServiceImpl struct {
	logger                             *zap.SugaredLogger
	pipelineStatusSyncDetailRepository pipelineConfig.PipelineStatusSyncDetailRepository
}

func NewPipelineStatusSyncDetailServiceImpl(logger *zap.SugaredLogger,
	pipelineStatusSyncDetailRepository pipelineConfig.PipelineStatusSyncDetailRepository,
) *PipelineStatusSyncDetailServiceImpl {
	return &PipelineStatusSyncDetailServiceImpl{
		logger:                             logger,
		pipelineStatusSyncDetailRepository: pipelineStatusSyncDetailRepository,
	}
}

func (impl *PipelineStatusSyncDetailServiceImpl) SaveOrUpdateSyncDetail(cdWfrId int, userId int32) error {
	syncDetailModel, err := impl.pipelineStatusSyncDetailRepository.GetByCdWfrId(cdWfrId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting pipeline status sync detail", "err", err, "cdWfrId", cdWfrId)
		return err
	}
	if syncDetailModel != nil {
		syncDetailModel.LastSyncedAt = time.Now()
		syncDetailModel.SyncCount += 1
		syncDetailModel.UpdatedBy = userId
		syncDetailModel.UpdatedOn = time.Now()
		err = impl.pipelineStatusSyncDetailRepository.Update(syncDetailModel)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline status sync detail", "err", err, "model", syncDetailModel)
			return err
		}
	} else {
		syncDetailModelNew := &pipelineConfig.PipelineStatusSyncDetail{
			CdWorkflowRunnerId: cdWfrId,
			LastSyncedAt:       time.Now(),
			SyncCount:          1,
			AuditLog: sql.AuditLog{
				CreatedBy: userId,
				CreatedOn: time.Now(),
				UpdatedBy: userId,
				UpdatedOn: time.Now(),
			},
		}
		err = impl.pipelineStatusSyncDetailRepository.Save(syncDetailModelNew)
		if err != nil {
			impl.logger.Errorw("error in saving pipeline status sync detail", "err", err, "model", syncDetailModelNew)
			return err
		}
	}
	return nil
}
func (impl *PipelineStatusSyncDetailServiceImpl) GetSyncTimeAndCountByCdWfrId(cdWfrId int) (time.Time, int, error) {
	syncTime := time.Time{}
	syncCount := 0
	syncDetailModel, err := impl.pipelineStatusSyncDetailRepository.GetByCdWfrId(cdWfrId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting pipeline status sync detail", "err", err, "cdWfrId", cdWfrId)
		return syncTime, syncCount, err
	}
	if syncDetailModel != nil {
		syncTime = syncDetailModel.LastSyncedAt
		syncCount = syncDetailModel.SyncCount
	}
	return syncTime, syncCount, nil
}
