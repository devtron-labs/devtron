package async

import (
	"fmt"
	"github.com/devtron-labs/common-lib/constants"
	"github.com/devtron-labs/common-lib/pubsub-lib/metrics"
	"go.uber.org/zap"
	"log"
	"runtime/debug"
)

type Async struct {
	logger           *zap.SugaredLogger
	microServiceName MicroServiceName
}

type MicroServiceName string

func (m MicroServiceName) ToString() string {
	return string(m)
}

type RunAsyncMetaData struct {
	Method string
	Path   string
}

func NewAsync(logger *zap.SugaredLogger, microServiceName MicroServiceName) *Async {
	return &Async{
		logger:           logger,
		microServiceName: microServiceName,
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

func (impl *Async) Run(fn func(), metadataOpts ...NewUpdateMetaData) {
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
				metrics.IncPanicRecoveryCount("go-routine", impl.microServiceName.ToString(), metaData.Method, metaData.Path)
				if impl.logger == nil {
					log.Println(constants.GoRoutinePanicMsgLogPrefix, "go-routine recovered from panic", "err:", r, "stack:", string(debug.Stack()))
				} else {
					impl.logger.Errorw(fmt.Sprintf("%s %s", constants.GoRoutinePanicMsgLogPrefix, "go-routine recovered from panic"), "err", r, "stack", string(debug.Stack()))
				}
			}
		}()
		if fn != nil {
			fn()
		}
	}()
}
