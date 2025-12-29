package asyncProvider

import (
	"github.com/devtron-labs/common-lib/constants"
	"github.com/devtron-labs/common-lib/workerPool"
	"go.uber.org/zap"
)

func NewBatchWorker[T any](batchSize int, logger *zap.SugaredLogger) *workerPool.WorkerPool[T] {
	return workerPool.NewWorkerPool[T](batchSize, constants.Orchestrator, logger)
}
