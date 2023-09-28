package cron

import (
	"fmt"
	"github.com/caarlos0/env"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"time"
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
	ciArtifactRepository    repository2.CiArtifactRepository
}

func NewCiTriggerCronImpl(logger *zap.SugaredLogger, cfg *CiTriggerCronConfig, pipelineStageRepository repository.PipelineStageRepository,
	ciHandler pipeline.CiHandler, ciArtifactRepository repository2.CiArtifactRepository) *CiTriggerCronImpl {
	cron := cron.New(
		cron.WithChain())
	cron.Start()
	impl := &CiTriggerCronImpl{
		logger:                  logger,
		cron:                    cron,
		pipelineStageRepository: pipelineStageRepository,
		ciHandler:               ciHandler,
		cfg:                     cfg,
		ciArtifactRepository:    ciArtifactRepository,
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
	ciPipelineIds, err := impl.pipelineStageRepository.GetAllCiPipelineIdsByPluginIdAndStageType(impl.cfg.PluginIds, string(repository.PIPELINE_STAGE_TYPE_POST_CI))
	if err != nil {
		return
	}
	artifacts, err := impl.ciArtifactRepository.GetLatestArtifactTimeByCiPipelineIds(ciPipelineIds)
	mp := make(map[int]time.Time)
	for _, artifact := range artifacts {
		mp[artifact.PipelineId] = artifact.CreatedOn
	}
	for _, ciPipelineId := range ciPipelineIds {
		//_, err := impl.ciHandler.FetchMaterialsByPipelineId(ciPipelineId, false)
		//if err != nil {
		//	return
		//}
		var ciPipelineMaterials []bean.CiPipelineMaterial

		//for _, material := range materials {
		//	if len(material.History) == 0 {
		//		return
		//	}
		//	ciPipelineMaterial := bean.CiPipelineMaterial{
		//		Id:            material.Id,
		//		GitMaterialId: material.GitMaterialId,
		//		GitCommit: bean.GitCommit{
		//			Commit: material.History[0].Commit,
		//		},
		//	}
		//	ciPipelineMaterials = append(ciPipelineMaterials, ciPipelineMaterial)
		//}
		ciTriggerRequest := bean.CiTriggerRequest{
			PipelineId:          ciPipelineId,
			CiPipelineMaterial:  ciPipelineMaterials,
			TriggeredBy:         1,
			InvalidateCache:     false,
			CiArtifactLastFetch: mp[ciPipelineId],
			PipelineType:        bean2.JOB_CI,
		}
		_, err = impl.ciHandler.HandleCIManual(ciTriggerRequest)
		if err != nil {
			return
		}
	}
	return
}
