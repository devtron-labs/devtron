package timeoutWindow

import (
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository"
	"github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type TimeoutWindowService interface {
	GetAllWithIds(ids []int) ([]*repository.TimeoutWindowConfiguration, error)
	UpdateTimeoutExpressionAndFormatForIds(tx *pg.Tx, timeoutExpression string, ids []int, expressionFormat bean.ExpressionFormat, loggedInUserId int32) error
	BulkCreateWithTimeoutExpressionAndFormat(tx *pg.Tx, timeoutExpression string, count int, expressionFormat bean.ExpressionFormat, loggedInUserId int32) ([]*repository.TimeoutWindowConfiguration, error)
	CreateWithTimeoutExpressionAndFormat(tx *pg.Tx, timeoutExpression string, expressionFormat bean.ExpressionFormat, loggedInUserId int32) (*repository.TimeoutWindowConfiguration, error)
	GetOrCreateWithExpressionAndFormat(tx *pg.Tx, timeoutExpression string, expressionFormat bean.ExpressionFormat, loggedInUserId int32) (*repository.TimeoutWindowConfiguration, error)
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

func (impl TimeWindowServiceImpl) GetAllWithIds(ids []int) ([]*repository.TimeoutWindowConfiguration, error) {
	timeWindows, err := impl.timeWindowRepository.GetWithIds(ids)
	if err != nil {
		impl.logger.Errorw("error in GetAllWithIds", "err", err, "timeWindowIds", ids)
		return nil, err
	}
	return timeWindows, err
}

func (impl TimeWindowServiceImpl) UpdateTimeoutExpressionAndFormatForIds(tx *pg.Tx, timeoutExpression string, ids []int, expressionFormat bean.ExpressionFormat, loggedInUserId int32) error {
	err := impl.timeWindowRepository.UpdateTimeoutExpressionAndFormatForIds(tx, timeoutExpression, ids, expressionFormat, loggedInUserId)
	if err != nil {
		impl.logger.Errorw("error in UpdateTimeoutExpressionForIds", "err", err, "timeoutExpression", timeoutExpression)
		return err
	}
	return err
}

func (impl TimeWindowServiceImpl) BulkCreateWithTimeoutExpressionAndFormat(tx *pg.Tx, timeoutExpression string, count int, expressionFormat bean.ExpressionFormat, loggedInUserId int32) ([]*repository.TimeoutWindowConfiguration, error) {
	var models []*repository.TimeoutWindowConfiguration
	for i := 0; i < count; i++ {
		model := repository.GetTimeoutWindowConfigModel(timeoutExpression, expressionFormat, loggedInUserId)
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

func (impl TimeWindowServiceImpl) CreateWithTimeoutExpressionAndFormat(tx *pg.Tx, timeoutExpression string, expressionFormat bean.ExpressionFormat, loggedInUserId int32) (*repository.TimeoutWindowConfiguration, error) {
	model := repository.GetTimeoutWindowConfigModel(timeoutExpression, expressionFormat, loggedInUserId)
	model, err := impl.timeWindowRepository.Create(tx, model)
	if err != nil {
		impl.logger.Errorw("error in CreateWithTimeoutExpression", "err", err, "timeoutExpression", timeoutExpression)
		return nil, err
	}
	return model, nil

}

func (impl TimeWindowServiceImpl) GetOrCreateWithExpressionAndFormat(tx *pg.Tx, timeoutExpression string, expressionFormat bean.ExpressionFormat, loggedInUserId int32) (*repository.TimeoutWindowConfiguration, error) {
	timeoutWindow, err := impl.timeWindowRepository.GetWithExpressionAndFormat(tx, timeoutExpression, expressionFormat)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error encountered in GetOrCreateWithExpressionAndFormat", "timeoutExpression", timeoutExpression, "expressionFormat", expressionFormat, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		timeoutWindow, err = impl.CreateWithTimeoutExpressionAndFormat(tx, timeoutExpression, expressionFormat, loggedInUserId)
		if err != nil {
			impl.logger.Errorw("error encountered in GetOrCreateWithExpressionAndFormat", "timeoutExpression", timeoutExpression, "expressionFormat", expressionFormat, "err", err)
			return nil, err
		}
	}
	return timeoutWindow, nil

}
