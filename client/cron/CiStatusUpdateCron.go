package cron

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type CiStatusUpdateCron interface {
	UpdateStatusForStuckCi()
}

type CiStatusUpdateCronImpl struct {
	logger                   *zap.SugaredLogger
	cron                     *cron.Cron
	appService               app.AppService
	ciStatusUpdateCronConfig *CiStatusUpdateCronConfig
	ciPipelineRepository     pipelineConfig.CiPipelineRepository
}

func NewCiStatusUpdateCronImpl(logger *zap.SugaredLogger, appService app.AppService,
	ciStatusUpdateCronConfig *CiStatusUpdateCronConfig, ciPipelineRepository pipelineConfig.CiPipelineRepository) *CiStatusUpdateCronImpl {
	cron := cron.New(
		cron.WithChain())
	cron.Start()
	impl := &CiStatusUpdateCronImpl{
		logger:                   logger,
		cron:                     cron,
		appService:               appService,
		ciStatusUpdateCronConfig: ciStatusUpdateCronConfig,
		ciPipelineRepository:     ciPipelineRepository,
	}
	_, err := cron.AddFunc(ciStatusUpdateCronConfig.CiPipelineStatusCronTime, impl.UpdateStatusForStuckCi)
	if err != nil {
		logger.Errorw("error in starting ci application status update cron job", "err", err)
		return nil
	}
	return impl
}

type CiStatusUpdateCronConfig struct {
	CiPipelineStatusCronTime string `env:"CI_PIPELINE_STATUS_CRON_TIME" envDefault:"*/2 * * * *"`
	PipelineFailedTime       string `env:"PIPELINE_FAILED_TIME" envDefault:"10"` //in minutes
}

func GetCiStatusUpdateCronConfig() (*CiStatusUpdateCronConfig, error) {
	cfg := &CiStatusUpdateCronConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse server app status config: " + err.Error())
		return nil, err
	}
	return cfg, nil
}

func (impl *CiStatusUpdateCronImpl) UpdateStatusForStuckCi() {

	return
}
