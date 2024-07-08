package async

import (
	"fmt"
	"github.com/devtron-labs/common-lib/constants"
	"github.com/devtron-labs/common-lib/pubsub-lib/metrics"
	"github.com/devtron-labs/common-lib/utils/runTime"
	"go.uber.org/zap"
	"runtime/debug"
)

type Runnable struct {
	logger      *zap.SugaredLogger
	serviceName constants.ServiceName
}

type RunAsyncMetaData struct {
	Method string
	Path   string
}

func NewAsyncRunnable(logger *zap.SugaredLogger, serviceName constants.ServiceName) *Runnable {
	return &Runnable{
		logger:      logger,
		serviceName: serviceName,
	}
}

func NewRunAsyncMetaData() *RunAsyncMetaData {
	return &RunAsyncMetaData{}
}

func CallerMethod(methodName string) NewUpdateMetaData {
	return func(m *RunAsyncMetaData) {
		m.Method = methodName
	}
}

func CallerPath(pathName string) NewUpdateMetaData {
	return func(m *RunAsyncMetaData) {
		m.Path = pathName
	}
}

type NewUpdateMetaData func(*RunAsyncMetaData)

func (impl *Runnable) Execute(runnableFunc func()) {
	impl.run(runnableFunc,
		CallerMethod(runTime.GetCallerFunctionName()),
		CallerPath(fmt.Sprintf("%s:%d", runTime.GetCallerFileName(), runTime.GetCallerLineNumber())),
	)
}

func (impl *Runnable) run(fn func(), metadataOpts ...NewUpdateMetaData) {
	metaData := NewRunAsyncMetaData()
	for _, metadataOpt := range metadataOpts {
		// updating meta data
		if metadataOpt != nil {
			metadataOpt(metaData)
		}
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				metrics.IncPanicRecoveryCount("go-routine", impl.serviceName.ToString(), metaData.Method, metaData.Path)
				impl.logger.Errorw(fmt.Sprintf("%s %s", constants.GoRoutinePanicMsgLogPrefix, "go-routine recovered from panic"), "err", r, "stack", string(debug.Stack()))
			}
		}()
		if fn != nil {
			fn()
		}
	}()
}
