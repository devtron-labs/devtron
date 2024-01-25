package cron

import (
	"github.com/devtron-labs/common-lib/constants"
	"go.uber.org/zap"
)

type CronLoggerImpl struct {
	logger *zap.SugaredLogger
}

func (impl *CronLoggerImpl) Info(msg string, keysAndValues ...interface{}) {
	impl.logger.Infow(msg, keysAndValues...)
}

func (impl *CronLoggerImpl) Error(err error, msg string, keysAndValues ...interface{}) {
	keysAndValues = append([]interface{}{"err", err}, keysAndValues...)
	impl.logger.Errorw(constants.PanicLogIdentifier+": "+msg, keysAndValues...)
}

func NewCronLoggerImpl(logger *zap.SugaredLogger) *CronLoggerImpl {
	return &CronLoggerImpl{logger: logger}
}
