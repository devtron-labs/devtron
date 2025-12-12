package workerPool

import (
	"fmt"
	"github.com/devtron-labs/common-lib/constants"
	"github.com/devtron-labs/common-lib/pubsub-lib/metrics"
	"github.com/devtron-labs/common-lib/utils/reflectUtils"
	"github.com/devtron-labs/common-lib/utils/runTime"
	"github.com/gammazero/workerpool"
	"go.uber.org/zap"
	"reflect"
	"runtime/debug"
	"sync"
)

type WorkerPool[T any] struct {
	logger           *zap.SugaredLogger
	service          constants.ServiceName
	wp               *workerpool.WorkerPool
	mu               *sync.Mutex
	err              chan error
	response         []T
	includeZeroValue bool
}

func NewWorkerPool[T any](maxWorkers int, serviceName constants.ServiceName, logger *zap.SugaredLogger) *WorkerPool[T] {
	wp := &WorkerPool[T]{
		logger:  logger,
		service: serviceName,
		wp:      workerpool.New(maxWorkers),
		mu:      &sync.Mutex{},
		err:     make(chan error, 1),
	}
	return wp
}

func (wp *WorkerPool[T]) InitializeResponse() *WorkerPool[T] {
	wp.response = []T{}
	return wp
}

func (wp *WorkerPool[T]) IncludeZeroValue() *WorkerPool[T] {
	wp.includeZeroValue = true
	return wp
}

func (wp *WorkerPool[T]) Submit(task func() (T, error)) {
	if task == nil {
		return
	}
	wp.wp.Submit(func() {
		defer func() {
			if r := recover(); r != nil {
				metrics.IncPanicRecoveryCount("go-routine", wp.service.ToString(), runTime.GetCallerFunctionName(), fmt.Sprintf("%s:%d", runTime.GetCallerFileName(), runTime.GetCallerLineNumber()))
				wp.logger.Errorw(fmt.Sprintf("%s %s", constants.GoRoutinePanicMsgLogPrefix, "go-routine recovered from panic"), "err", r, "stack", string(debug.Stack()))
			}
		}()
		if wp.Error() != nil {
			return
		}
		res, err := task()
		if err != nil {
			wp.logger.Errorw("error in worker pool task", "err", err)
			wp.setError(err)
			return
		}
		wp.updateResponse(res)
	})
}

func (wp *WorkerPool[T]) updateResponse(res T) {
	wp.lock()
	defer wp.unlock()
	val := reflect.ValueOf(res)
	if reflectUtils.IsNullableValue(val) && val.IsNil() {
		return
	} else if !wp.includeZeroValue && val.IsZero() {
		return
	} else {
		wp.response = append(wp.response, res)
		return
	}
}

func (wp *WorkerPool[_]) StopWait() error {
	wp.wp.StopWait()
	// return error from workerPool error channel
	return wp.Error()
}

func (wp *WorkerPool[_]) lock() {
	wp.mu.Lock()
}

func (wp *WorkerPool[_]) unlock() {
	wp.mu.Unlock()
}

func (wp *WorkerPool[_]) Error() error {
	select {
	case err := <-wp.err:
		return err
	default:
		return nil
	}
}

func (wp *WorkerPool[_]) setError(err error) {
	if err != nil && wp.Error() == nil {
		wp.err <- err
	}
}

func (wp *WorkerPool[T]) GetResponse() []T {
	return wp.response
}
