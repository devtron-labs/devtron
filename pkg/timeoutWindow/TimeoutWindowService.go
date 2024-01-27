package timeoutWindow

import (
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type TimeoutWindowService interface {
	GetAndCreateIfNotPresent()
	GetAllWithIds(ids []int) ([]*repository.TimeoutWindowConfiguration, error)
	UpdateTimeoutExpressionAndFormatForIds(tx *pg.Tx, timeoutExpression string, ids []int, expressionFormat bean.ExpressionFormat) error
	CreateWithTimeoutExpressionAndFormat(tx *pg.Tx, timeoutExpression string, count int, expressionFormat bean.ExpressionFormat) ([]*repository.TimeoutWindowConfiguration, error)
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

func (impl TimeWindowServiceImpl) UpdateTimeoutExpressionAndFormatForIds(tx *pg.Tx, timeoutExpression string, ids []int, expressionFormat bean.ExpressionFormat) error {
	err := impl.timeWindowRepository.UpdateTimeoutExpressionAndFormatForIds(tx, timeoutExpression, ids, expressionFormat)
	if err != nil {
		impl.logger.Errorw("error in UpdateTimeoutExpressionForIds", "err", err, "timeoutExpression", timeoutExpression)
		return err
	}
	return err
}

func (impl TimeWindowServiceImpl) CreateWithTimeoutExpressionAndFormat(tx *pg.Tx, timeoutExpression string, count int, expressionFormat bean.ExpressionFormat) ([]*repository.TimeoutWindowConfiguration, error) {
	var models []*repository.TimeoutWindowConfiguration
	for i := 0; i < count; i++ {
		model := &repository.TimeoutWindowConfiguration{
			TimeoutWindowExpression: timeoutExpression,
			ExpressionFormat:        expressionFormat,
		}
		models = append(models, model)
	}
	// create in batch
	models, err := impl.timeWindowRepository.CreateInBatch(tx, models)
	if err != nil {
		impl.logger.Errorw("error in CreateWithTimeoutExpression", "err", err, "timeoutExpression", timeoutExpression, "countToBeCreated", count)
		return nil, err
	}
	return models, nil

}
