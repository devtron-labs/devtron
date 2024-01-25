package timeoutWindow

import (
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"go.uber.org/zap"
)

type TimeoutWindowService interface {
	GetAndCreateIfNotPresent()
	GetAllWithIds(ids []int) ([]*repository.TimeoutWindowConfiguration, error)
	UpdateTimeoutExpressionForIds(timeoutExpression string, ids []int) error
	CreateWithTimeoutExpression(timeoutExpression string, count int) ([]*repository.TimeoutWindowConfiguration, error)
}

type TimeWindowServiceImpl struct {
	logger               *zap.SugaredLogger
	timeWindowRepository repository.TimeWindowRepository
}

func NewTimeWindowServiceImpl(logger *zap.SugaredLogger,
	timeWindowRepository repository.TimeWindowRepository) *TimeWindowServiceImpl {
	timeoutWindowServiceImpl := &TimeWindowServiceImpl{
		logger:               logger,
		timeWindowRepository: timeWindowRepository,
	}
	return timeoutWindowServiceImpl
}

func (impl TimeWindowServiceImpl) GetAndCreateIfNotPresent() {
	// get with desired

}

func (impl TimeWindowServiceImpl) GetAllWithIds(ids []int) ([]*repository.TimeoutWindowConfiguration, error) {
	timeWindows, err := impl.timeWindowRepository.GetWithIds(ids)
	if err != nil {
		impl.logger.Errorw("error in GetAllWithIds", "err", err, "timeWindowIds", ids)
		return nil, err
	}
	return timeWindows, err
}

func (impl TimeWindowServiceImpl) UpdateTimeoutExpressionForIds(timeoutExpression string, ids []int) error {
	err := impl.timeWindowRepository.UpdateTimeoutExpressionForIds(timeoutExpression, ids)
	if err != nil {
		impl.logger.Errorw("error in UpdateTimeoutExpressionForIds", "err", err, "timeoutExpression", timeoutExpression)
		return err
	}
	return err
}

func (impl TimeWindowServiceImpl) CreateWithTimeoutExpression(timeoutExpression string, count int) ([]*repository.TimeoutWindowConfiguration, error) {
	// Considering timestamp expression format for now, if other formats are added support can be added
	var models []*repository.TimeoutWindowConfiguration
	for i := 0; i < count; i++ {
		model := &repository.TimeoutWindowConfiguration{
			TimeoutWindowExpression: timeoutExpression,
			ExpressionFormat:        bean.TimeStamp,
		}
		models = append(models, model)
	}
	// create in batch
	models, err := impl.timeWindowRepository.CreateInBatch(models)
	if err != nil {
		impl.logger.Errorw("error in CreateWithTimeoutExpression", "err", err, "timeoutExpression", timeoutExpression, "countToBeCreated", count)
		return nil, err
	}
	return models, nil

}
