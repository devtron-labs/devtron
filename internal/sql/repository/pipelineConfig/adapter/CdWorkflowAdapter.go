package adapter

import "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"

type UpdateOptions = func(cdWfr *pipelineConfig.CdWorkflowRunner)

func WithMessage(msg string) UpdateOptions {
	return func(cdWfr *pipelineConfig.CdWorkflowRunner) {
		cdWfr.Message = msg
	}
}

func WithStatus(status string) UpdateOptions {
	return func(cdWfr *pipelineConfig.CdWorkflowRunner) {
		cdWfr.Status = status
	}
}
