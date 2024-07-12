package asyncProvider

import (
	"github.com/devtron-labs/common-lib/async"
	"github.com/devtron-labs/common-lib/constants"
	"go.uber.org/zap"
)

func NewAsyncRunnable(logger *zap.SugaredLogger) *async.Runnable {
	return async.NewAsyncRunnable(logger, constants.Orchestrator)
}
