package executors

import (
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"go.uber.org/zap"
)

type WorkflowExecutor interface {
	ExecuteWorkflow(workflowTemplate bean.WorkflowTemplate) error
}

type ArgoWorkflowExecutor interface {
	WorkflowExecutor
}

type ArgoWorkflowExecutorImpl struct {
	logger *zap.SugaredLogger
}

func NewArgoWorkflowExecutorImpl(logger *zap.SugaredLogger) *ArgoWorkflowExecutorImpl {
	return &ArgoWorkflowExecutorImpl{logger: logger}
}

func (impl *ArgoWorkflowExecutorImpl) ExecuteWorkflow(workflowTemplate bean.WorkflowTemplate) error {

	// get cm and cs argo step templates

}
