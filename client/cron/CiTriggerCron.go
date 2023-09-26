package cron

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type CiTriggerCron interface {
	TriggerCiCron()
}

type CiTriggerCronImpl struct {
	logger                  *zap.SugaredLogger
	cron                    *cron.Cron
	cfg                     *CiTriggerCronConfig
	pipelineStageRepository repository.PipelineStageRepository
	ciHandler               pipeline.CiHandler
}

func NewCiTriggerCronImpl(logger *zap.SugaredLogger, cfg *CiTriggerCronConfig, pipelineStageRepository repository.PipelineStageRepository,
	ciHandler pipeline.CiHandler) *CiTriggerCronImpl {
	cron := cron.New(
		cron.WithChain())
	cron.Start()
	impl := &CiTriggerCronImpl{
		logger:                  logger,
		cron:                    cron,
		pipelineStageRepository: pipelineStageRepository,
		ciHandler:               ciHandler,
		cfg:                     cfg,
	}

	_, err := cron.AddFunc(fmt.Sprintf("@every %dm", cfg.SourceControllerCronTime), impl.TriggerCiCron)
	if err != nil {
		logger.Errorw("error while configure cron job for ci workflow status update", "err", err)
		return impl
	}
	return impl
}

type CiTriggerCronConfig struct {
	SourceControllerCronTime int `env:"CI_WORKFLOW_STATUS_UPDATE_CRON" envDefault:"2"`
	PluginIds                int `env:"PLUGIN_IDS"  envDefault:"2"`
}

func GetCiTriggerCronConfig() (*CiTriggerCronConfig, error) {
	cfg := &CiTriggerCronConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse ci trigger cron config: " + err.Error())
		return nil, err
	}
	return cfg, nil
}

// UpdateCiWorkflowStatusFailedCron this function will execute periodically
func (impl *CiTriggerCronImpl) TriggerCiCron() {
	ciPipelineIds, err := impl.pipelineStageRepository.GetAllCiPipelineIdsByPluginIdAndStageType(impl.cfg.PluginIds, "POST_CI")
	if err != nil {
		return
	}
	for _, ciPipelineId := range ciPipelineIds {
		material, err := impl.ciHandler.FetchMaterialsByPipelineId(ciPipelineId, false)
		if err != nil {
			return
		}
		ciTriggerRequest := bean.CiTriggerRequest{
			PipelineId: ciPipelineId,
			CiPipelineMaterial: []bean.CiPipelineMaterial{
				{
					Id:            material[0].Id,
					GitMaterialId: material[0].GitMaterialId,
					GitCommit: bean.GitCommit{
						Commit: material[0].History[0].Commit,
					},
				},
			},
			TriggeredBy:     1,
			InvalidateCache: false,
		}
		_, err = impl.ciHandler.HandleCIManual(ciTriggerRequest)
		if err != nil {
			return
		}
	}
	return
}
