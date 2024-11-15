package devtronApps

import (
	"context"
	"fmt"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean6 "github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"time"
)

func (impl *TriggerServiceImpl) checkFeasibilityForPreStage(pipeline *pipelineConfig.Pipeline, request *bean2.TriggerRequest,
	env *repository.Environment, artifact *repository2.CiArtifact, triggeredBy int32) (*ResourceFilterEvaluationAudit, error) {
	return nil, nil
}

func (impl *TriggerServiceImpl) createAuditDataForDeploymentWindowBypass(request bean2.TriggerRequest, wfrId int) error {
	return nil
}

func (impl *TriggerServiceImpl) getManifestPushTemplateForPreStage(ctx context.Context, envDeploymentConfig *bean3.DeploymentConfig,
	pipeline *pipelineConfig.Pipeline, artifact *repository2.CiArtifact, jobHelmPackagePath string,
	cdWf *pipelineConfig.CdWorkflow, runner *pipelineConfig.CdWorkflowRunner, triggeredBy int32, triggeredAt time.Time,
	request bean2.TriggerRequest) (*bean6.ManifestPushTemplate, error) {
	return nil, nil
}

func (impl *TriggerServiceImpl) setCloningModeInCIProjectDetail(ciProjectDetail *bean.CiProjectDetails, appId int,
	m *pipelineConfig.CiPipelineMaterial) error {
	return nil
}

func (impl *TriggerServiceImpl) getPreStageBuildRegistryConfigIfSourcePipelineNotPresent(appId int) (*types.DockerArtifactStoreBean, error) {
	return nil, fmt.Errorf("soucePipeline is mandatory, corrupt data")
}
