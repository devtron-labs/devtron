package resourceFilter

import "go.uber.org/zap"

type FilterEvaluationService interface {
}
type FilterEvaluationServiceImpl struct {
	logger *zap.SugaredLogger
}

func NewFilterServiceImpl(logger *zap.SugaredLogger) *FilterEvaluationServiceImpl {
	return &FilterEvaluationServiceImpl{
		logger: logger,
	}
}
