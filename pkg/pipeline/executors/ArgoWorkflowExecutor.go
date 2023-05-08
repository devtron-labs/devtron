package executors

import (
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"go.uber.org/zap"
	"strconv"
)

const (
	CD_WORKFLOW_NAME        = "cd"
	CD_WORKFLOW_WITH_STAGES = "cd-stages-with-env"
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

	entryPoint := CD_WORKFLOW_NAME

	// get cm and cs argo step templates
	templates, steps, err := impl.getArgoTemplatesAndSteps(workflowTemplate.ConfigMaps, workflowTemplate.Secrets)

}

func (impl *ArgoWorkflowExecutorImpl) getArgoTemplatesAndSteps(configMaps []bean2.ConfigSecretMap, secrets []bean2.ConfigSecretMap) ([]v1alpha1.Template, []v1alpha1.ParallelSteps, error) {
	var templates []v1alpha1.Template
	var steps []v1alpha1.ParallelSteps
	cmIndex := 0
	csIndex := 0
	for _, configMap := range configMaps {
		i, parallelSteps, err2 := impl.funcName(configMap, templates, steps, cmIndex)
		if err2 != nil {
			return i, parallelSteps, err2
		}
		cmIndex++
	}
}

func (impl *ArgoWorkflowExecutorImpl) funcName(configMap bean2.ConfigSecretMap, templates []v1alpha1.Template, steps []v1alpha1.ParallelSteps, cmIndex int) ([]v1alpha1.Template, []v1alpha1.ParallelSteps, error) {
	configDataMap, err := configMap.GetDataMap()
	if err != nil {
		impl.logger.Errorw("error occurred while extracting data map", "Data", configMap.Data, "err", err)
		return templates, steps, err
	}
	cmJson, err := pipeline.GetConfigMapJson(pipeline.ConfigMapSecretDto{Name: configMap.Name, Data: configDataMap, OwnerRef: pipeline.ArgoWorkflowOwnerRef})
	if err != nil {
		impl.logger.Errorw("error occurred while extracting cm json", "configName", configMap.Name, "err", err)
		return templates, steps, err
	}
	steps = append(steps, v1alpha1.ParallelSteps{
		Steps: []v1alpha1.WorkflowStep{
			{
				Name:     "create-env-cm-gb-" + strconv.Itoa(cmIndex),
				Template: "cm-gb-" + strconv.Itoa(cmIndex),
			},
		},
	})
	templates = append(templates, v1alpha1.Template{
		Name: "cm-gb-" + strconv.Itoa(cmIndex),
		Resource: &v1alpha1.ResourceTemplate{
			Action:            "create",
			SetOwnerReference: true,
			Manifest:          string(cmJson),
		},
	})
	return nil, nil, nil
}
