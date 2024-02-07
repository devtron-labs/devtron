package eventProcessor

import (
	"github.com/devtron-labs/devtron/pkg/eventProcessor/in"
	"go.uber.org/zap"
)

type CentralEventProcessor struct {
	logger                 *zap.SugaredLogger
	workflowEventProcessor *in.WorkflowEventProcessorImpl
}

func NewCentralEventProcessor(workflowEventProcessor *in.WorkflowEventProcessorImpl,
	logger *zap.SugaredLogger) (*CentralEventProcessor, error) {
	cep := &CentralEventProcessor{
		workflowEventProcessor: workflowEventProcessor,
		logger:                 logger,
	}
	err := cep.SubscribeAll()
	if err != nil {
		return nil, err
	}
	return cep, nil
}

func (impl *CentralEventProcessor) SubscribeAll() error {
	var err error
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
	return nil
}
