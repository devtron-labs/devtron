package config

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiPipelineConfigReadService interface {
	FindLinkedCiCount(ciPipelineId int) (int, error)
	FindNumberOfAppsWithCiPipeline(appIds []int) (count int, err error)
	FindAllPipelineCreatedCountInLast24Hour() (pipelineCount int, err error)
	FindAllDeletedPipelineCountInLast24Hour() (pipelineCount int, err error)
	GetChildrenCiCount(parentCiPipelineId int) (int, error)
}

type CiPipelineConfigReadServiceImpl struct {
	logger               *zap.SugaredLogger
	ciPipelineRepository pipelineConfig.CiPipelineRepository
}

func NewCiPipelineConfigReadServiceImpl(
	logger *zap.SugaredLogger,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
) *CiPipelineConfigReadServiceImpl {
	return &CiPipelineConfigReadServiceImpl{
		logger:               logger,
		ciPipelineRepository: ciPipelineRepository,
	}
}

func (impl *CiPipelineConfigReadServiceImpl) FindLinkedCiCount(ciPipelineId int) (int, error) {
	return impl.ciPipelineRepository.FindLinkedCiCount(ciPipelineId)
}

func (impl *CiPipelineConfigReadServiceImpl) FindNumberOfAppsWithCiPipeline(appIds []int) (count int, err error) {
	return impl.ciPipelineRepository.FindNumberOfAppsWithCiPipeline(appIds)
}

func (impl *CiPipelineConfigReadServiceImpl) FindAllPipelineCreatedCountInLast24Hour() (pipelineCount int, err error) {
	return impl.ciPipelineRepository.FindAllPipelineCreatedCountInLast24Hour()
}

func (impl *CiPipelineConfigReadServiceImpl) FindAllDeletedPipelineCountInLast24Hour() (pipelineCount int, err error) {
	return impl.ciPipelineRepository.FindAllDeletedPipelineCountInLast24Hour()
}

func (impl *CiPipelineConfigReadServiceImpl) GetChildrenCiCount(parentCiPipelineId int) (int, error) {
	count, err := impl.ciPipelineRepository.GetChildrenCiCount(parentCiPipelineId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("failed to get children ci count", "parentCiPipelineId", parentCiPipelineId, "error", err)
		return 0, err
	} else if errors.Is(err, pg.ErrNoRows) {
		impl.logger.Debugw("no children ci found", "parentCiPipelineId", parentCiPipelineId)
		return 0, nil
	}
	return count, nil
}
