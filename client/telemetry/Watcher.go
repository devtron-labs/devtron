package telemetry

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/nats-io/stan"
	"go.uber.org/zap"
	"gopkg.in/robfig/cron.v3"
)

type TelemetryWatcherImpl struct {
	cron       *cron.Cron
	logger     *zap.SugaredLogger
	nats       stan.Conn
	pollConfig *PollConfig
}

type TelemetryWatcher interface {
}

type PollConfig struct {
	PollDuration int `env:"POLL_DURATION" envDefault:"1"`
	PollWorker   int `env:"POLL_WORKER" envDefault:"5"`
}

func NewTelemetryWatcherImpl(logger *zap.SugaredLogger, nats stan.Conn) (*TelemetryWatcherImpl, error) {
	cfg := &PollConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	cronLogger := &CronLoggerImpl{logger: logger}
	cron := cron.New(
		cron.WithChain(
			cron.SkipIfStillRunning(cronLogger),
			cron.Recover(cronLogger)))
	cron.Start()
	watcher := &TelemetryWatcherImpl{
		cron:       cron,
		logger:     logger,
		nats:       nats,
		pollConfig: cfg,
	}
	logger.Info()
	_, err = cron.AddFunc(fmt.Sprintf("@every %dm", cfg.PollDuration), watcher.Watch)
	if err != nil {
		fmt.Println("error in starting cron")
		return nil, err
	}
	return watcher, err
}

func (impl *TelemetryWatcherImpl) StopCron() {
	impl.cron.Stop()
}
func (impl *TelemetryWatcherImpl) Watch() {
	impl.logger.Infow("starting git watch thread")
	impl.logger.Infow("stop git watch thread")
}

type CronLoggerImpl struct {
	logger *zap.SugaredLogger
}

func (impl *CronLoggerImpl) Info(msg string, keysAndValues ...interface{}) {
	impl.logger.Infow(msg, keysAndValues...)
}

func (impl *CronLoggerImpl) Error(err error, msg string, keysAndValues ...interface{}) {
	keysAndValues = append([]interface{}{"err", err}, keysAndValues...)
	impl.logger.Errorw(msg, keysAndValues...)
}
