package asyncWrapper

import (
	"fmt"
	"github.com/devtron-labs/common-lib/async"
	"github.com/devtron-labs/common-lib/constants"
	"github.com/devtron-labs/devtron/util/runTimeUtil"
	"go.uber.org/zap"
)

type AsyncGoFuncService interface {
	Execute(callback func())
}

type AsyncGoFuncServiceImpl struct {
	async *async.Async
}

func NewAsyncGoFuncServiceImpl(logger *zap.SugaredLogger) *AsyncGoFuncServiceImpl {
	newAsync := async.NewAsync(logger, constants.OrchestratorMicroService)
	return &AsyncGoFuncServiceImpl{
		async: newAsync,
	}
}

func (impl *AsyncGoFuncServiceImpl) Execute(callback func()) {
	impl.async.Run(callback,
		async.CallerMethod(runTimeUtil.GetCallerFunctionName()),
		async.CallerPath(fmt.Sprintf("%s:%d", runTimeUtil.GetCallerFileName(), runTimeUtil.GetCallerLineNumber())),
	)
}
