package async

import (
	"fmt"
	"github.com/devtron-labs/common-lib/pubsub-lib/metrics"
	"go.uber.org/zap"
	"runtime/debug"
)

type Async struct {
	logger *zap.SugaredLogger
}

func NewAsync(logger *zap.SugaredLogger) *Async {
	return &Async{
		logger: logger,
	}
}

func (impl *Async) RunAsync(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				metrics.IncPanicRecoveryCount("go-routine", "", "", string(debug.Stack()))
				if impl.logger == nil {
					fmt.Println("panic recovered", "error", r, "stack:", string(debug.Stack()))
				} else {
					impl.logger.Errorw("panic recovered", "error", r)
				}
			}
		}()
		fn()
	}()
}
