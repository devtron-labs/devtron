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
	logger *zap.SugaredLogger
}

type RunAsyncMetaData struct {
	MicroServiceName string
	InterfaceName    string
	MethodName       string
}

func NewAsync(logger *zap.SugaredLogger) *Async {
	return &Async{
		logger: logger,
	}
}

func NewRunAsyncMetaData() *RunAsyncMetaData {
	return &RunAsyncMetaData{}
}

func RunningMicroService(microServiceName string) NewUpdateMetaData {
	return func(m *RunAsyncMetaData) {
		m.MicroServiceName = microServiceName
	}
}

func RunningInterface(interfaceName string) NewUpdateMetaData {
	return func(m *RunAsyncMetaData) {
		m.InterfaceName = interfaceName
	}
}

func RunningMethod(methodName string) NewUpdateMetaData {
	return func(m *RunAsyncMetaData) {
		m.MethodName = methodName
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
				metrics.IncPanicRecoveryCount("go-routine", metaData.MicroServiceName, metaData.InterfaceName, metaData.MethodName)
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
