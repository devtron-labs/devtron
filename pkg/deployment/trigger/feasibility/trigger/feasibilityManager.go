package trigger

import "go.uber.org/zap"

type FeasibilityManagerImpl struct {
	logger *zap.SugaredLogger
}

func NewFeasibilityManagerImpl(logger *zap.SugaredLogger) *FeasibilityManagerImpl {
	return &FeasibilityManagerImpl{
		logger: logger,
	}
}

type FeasibilityManager interface {
	checkFeasibility() error
}

func (impl *FeasibilityManagerImpl) checkFeasibility() error {
	return nil
}
