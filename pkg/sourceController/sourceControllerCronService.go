package sourceController

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type SourceControllerCronService interface {
}
type SourceControllerCronImpl struct {
	logger                  *zap.SugaredLogger
	sourceControllerService SourceControllerService
}

type SourceControllerCronConfig struct {
	SourceControllerCronTime int `env:"FETCH_LATEST_TAGS_CRON_TIME" envDefault:"5"`
}

func NewSourceControllerCronServiceImpl(logger *zap.SugaredLogger,
	sourceControllerService SourceControllerService) (*SourceControllerCronImpl, error) {
	sourceControllerServiceImpl := &SourceControllerCronImpl{
		logger:                  logger,
		sourceControllerService: sourceControllerService,
	}
	// initialise cron
	newCron := cron.New(cron.WithChain())
	newCron.Start()
	cfg := &SourceControllerCronConfig{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Errorw("failed to parse server cluster status config: ", "err", err.Error())
	}
	// add function into cron
	_, err = newCron.AddFunc(fmt.Sprintf("@every %dm", cfg.SourceControllerCronTime), sourceControllerService.ReconcileSourceWrapper)
	if err != nil {
		logger.Errorw("error in adding cron function into SourceControllerCronService", "err", err)
		return sourceControllerServiceImpl, err
	}
	return sourceControllerServiceImpl, nil

}
