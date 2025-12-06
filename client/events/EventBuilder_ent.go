package client

import "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"

func (impl *EventSimpleFactoryImpl) addExtraCdDataForEnterprise(event Event, wfr *pipelineConfig.CdWorkflowRunner) Event {
	return event
}
