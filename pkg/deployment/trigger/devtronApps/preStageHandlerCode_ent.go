package devtronApps

import (
	"context"
	"fmt"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean6 "github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"time"
)

func (impl *HandlerServiceImpl) checkFeasibilityForPreStage(pipeline *pipelineConfig.Pipeline, request *bean2.CdTriggerRequest,
	env *repository.Environment, artifact *repository2.CiArtifact, triggeredBy int32) (interface{}, error) {
	//here return type is interface as ResourceFilterEvaluationAudit is not present in this version
	return nil, nil
}

func (impl *HandlerServiceImpl) createAuditDataForDeploymentWindowBypass(request bean2.CdTriggerRequest, wfrId int) error {
	return nil
}

func (impl *HandlerServiceImpl) getManifestPushTemplateForPreStage(ctx context.Context, envDeploymentConfig *bean3.DeploymentConfig,
	pipeline *pipelineConfig.Pipeline, artifact *repository2.CiArtifact, jobHelmPackagePath string,
	cdWf *pipelineConfig.CdWorkflow, runner *pipelineConfig.CdWorkflowRunner, triggeredBy int32, triggeredAt time.Time,
	request bean2.CdTriggerRequest) (*bean6.ManifestPushTemplate, error) {
	return nil, nil
}

func (impl *HandlerServiceImpl) setCloningModeInCIProjectDetail(ciProjectDetail *bean.CiProjectDetails, appId int,
	m *pipelineConfig.CiPipelineMaterial) error {
	return nil
}

func (impl *HandlerServiceImpl) getPreStageBuildRegistryConfigIfSourcePipelineNotPresent(appId int) (*types.DockerArtifactStoreBean, error) {
	return nil, fmt.Errorf("soucePipeline is mandatory, corrupt data")
}

func (impl *HandlerServiceImpl) handlerFilterEvaluationAudit(filterEvaluationAudit interface{},
	runner *pipelineConfig.CdWorkflowRunner) error {
	//here ip type of filterEvaluationAudit is interface as ResourceFilterEvaluationAudit is not present in this version
	return nil
}
