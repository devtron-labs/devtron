package eventProcessor

import (
	"github.com/devtron-labs/devtron/pkg/eventProcessor/in"
	"go.uber.org/zap"
)

type CentralEventProcessor struct {
	logger                   *zap.SugaredLogger
	workflowEventProcessor   *in.WorkflowEventProcessorImpl
	ciPipelineEventProcessor *in.CIPipelineEventProcessorImpl
}

func NewCentralEventProcessor(logger *zap.SugaredLogger,
	workflowEventProcessor *in.WorkflowEventProcessorImpl,
	ciPipelineEventProcessor *in.CIPipelineEventProcessorImpl) (*CentralEventProcessor, error) {
	cep := &CentralEventProcessor{
		logger:                   logger,
		workflowEventProcessor:   workflowEventProcessor,
		ciPipelineEventProcessor: ciPipelineEventProcessor,
	}
	err := cep.SubscribeAll()
	if err != nil {
		return nil, err
	}
	return cep, nil
}

func (impl *CentralEventProcessor) SubscribeAll() error {
	var err error

	//CI pipeline event starts
	err = impl.ciPipelineEventProcessor.SubscribeNewCIMaterialEvent()
	if err != nil {
		impl.logger.Errorw("error, SubscribeNewCIMaterialEvent", "err", err)
		return err
	}
	//CI pipeline event ends

	//Workflow event starts

	err = impl.workflowEventProcessor.SubscribeCDStageCompleteEvent()
	if err != nil {
		impl.logger.Errorw("error, SubscribeCDStageCompleteEvent", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeTriggerBulkAction()
	if err != nil {
		impl.logger.Errorw("error, SubscribeTriggerBulkAction", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeHibernateBulkAction()
	if err != nil {
		impl.logger.Errorw("error, SubscribeHibernateBulkAction", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeCIWorkflowStatusUpdate()
	if err != nil {
		impl.logger.Errorw("error, SubscribeCIWorkflowStatusUpdate", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeCDWorkflowStatusUpdate()
	if err != nil {
		impl.logger.Errorw("error, SubscribeCDWorkflowStatusUpdate", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeCICompleteEvent()
	if err != nil {
		impl.logger.Errorw("error, SubscribeCICompleteEvent", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeDevtronAsyncHelmInstallRequest()
	if err != nil {
		impl.logger.Errorw("error, SubscribeDevtronAsyncHelmInstallRequest", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeCDPipelineDeleteEvent()
	if err != nil {
		impl.logger.Errorw("error, SubscribeCDPipelineDeleteEvent", "err", err)
		return err
	}

	//Workflow event ends
	return nil
}
