package app

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStatusFetchDetailService interface {
	SaveOrUpdateFetchDetail(cdWfrId int, userId int32) error
	GetFetchTimeAndCountByCdWfrId(cdWfrId int) (time.Time, int, error)
}

type PipelineStatusFetchDetailServiceImpl struct {
	logger                              *zap.SugaredLogger
	pipelineStatusFetchDetailRepository pipelineConfig.PipelineStatusFetchDetailRepository
}

func NewPipelineStatusFetchDetailServiceImpl(logger *zap.SugaredLogger,
	pipelineStatusFetchDetailRepository pipelineConfig.PipelineStatusFetchDetailRepository,
) *PipelineStatusFetchDetailServiceImpl {
	return &PipelineStatusFetchDetailServiceImpl{
		logger:                              logger,
		pipelineStatusFetchDetailRepository: pipelineStatusFetchDetailRepository,
	}
}

func (impl *PipelineStatusFetchDetailServiceImpl) SaveOrUpdateFetchDetail(cdWfrId int, userId int32) error {
	fetchDetailModel, err := impl.pipelineStatusFetchDetailRepository.GetByCdWfrId(cdWfrId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting pipeline status fetch detail", "err", err, "cdWfrId", cdWfrId)
		return err
	}
	if fetchDetailModel != nil {
		fetchDetailModel.LastFetchedAt = time.Now()
		fetchDetailModel.FetchCount += 1
		fetchDetailModel.UpdatedBy = userId
		fetchDetailModel.UpdatedOn = time.Now()
		err = impl.pipelineStatusFetchDetailRepository.Update(fetchDetailModel)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline status fetch detail", "err", err, "model", fetchDetailModel)
			return err
		}
	} else {
		fetchDetailModelNew := &pipelineConfig.PipelineStatusFetchDetail{
			CdWorkflowRunnerId: cdWfrId,
			LastFetchedAt:      time.Now(),
			FetchCount:         1,
			AuditLog: sql.AuditLog{
				CreatedBy: userId,
				CreatedOn: time.Now(),
				UpdatedBy: userId,
				UpdatedOn: time.Now(),
			},
		}
		if err != nil {
			impl.logger.Errorw("error in saving pipeline status fetch detail", "err", err, "model", fetchDetailModelNew)
			return err
		}
	}
	return nil
}
func (impl *PipelineStatusFetchDetailServiceImpl) GetFetchTimeAndCountByCdWfrId(cdWfrId int) (time.Time, int, error) {
	fetchTime := time.Time{}
	fetchCount := 0
	fetchDetailModel, err := impl.pipelineStatusFetchDetailRepository.GetByCdWfrId(cdWfrId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting pipeline status fetch detail", "err", err, "cdWfrId", cdWfrId)
		return fetchTime, fetchCount, err
	}
	if fetchDetailModel != nil {
		fetchTime = fetchDetailModel.LastFetchedAt
		fetchCount = fetchDetailModel.FetchCount
	}
	return fetchTime, fetchCount, nil
}
