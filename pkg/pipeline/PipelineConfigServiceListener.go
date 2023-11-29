package pipeline

import (
	"go.uber.org/zap"
	"reflect"
)

type PipelineConfigListenerService interface {
	RegisterPipelineDeleteListener(listener PipelineConfigListener)
	HandleCdPipelineDelete(pipelineId int, triggeredBy int32)
}

type PipelineConfigListenerServiceImpl struct {
	logger                    *zap.SugaredLogger
	deleteCdPipelineListeners []PipelineConfigListener
}

type PipelineConfigListener interface {
	OnDeleteCdPipelineEvent(pipelineId int, triggeredBy int32)
}

func NewPipelineConfigListenerServiceImpl(logger *zap.SugaredLogger) *PipelineConfigListenerServiceImpl {
	return &PipelineConfigListenerServiceImpl{
		logger: logger,
	}
}

func (impl *PipelineConfigListenerServiceImpl) RegisterPipelineDeleteListener(listener PipelineConfigListener) {
	impl.logger.Infof("registering listener %s, service: PipelineConfigListenerService", reflect.TypeOf(listener))
	impl.deleteCdPipelineListeners = append(impl.deleteCdPipelineListeners, listener)
}

func (impl *PipelineConfigListenerServiceImpl) HandleCdPipelineDelete(pipelineId int, triggeredBy int32) {
	impl.logger.Infow("cd pipeline delete process", "pipelineId", pipelineId, "triggeredBy", triggeredBy)
	for _, deleteCdPipelineListener := range impl.deleteCdPipelineListeners {
		deleteCdPipelineListener.OnDeleteCdPipelineEvent(pipelineId, triggeredBy)
	}
}
