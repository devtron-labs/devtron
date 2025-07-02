package trigger

import (
	"encoding/json"
	"github.com/devtron-labs/common-lib/imageScan/bean"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/bean/common"
	bean2 "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	pipelineConfigBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"net/http"
)

func (impl *HandlerServiceImpl) updateRuntimeParamsForAutoCI(ciPipelineId int, runtimeParameters *common.RuntimeParameters) (*common.RuntimeParameters, error) {
	return runtimeParameters, nil
}

func (impl *HandlerServiceImpl) getRuntimeParamsForBuildingManualTriggerHashes(ciTriggerRequest bean3.CiTriggerRequest) *common.RuntimeParameters {
	return common.NewRuntimeParameters()
}

func (impl *HandlerServiceImpl) fetchImageScanExecutionMedium() (*repository.ScanToolMetadata, bean.ScanExecutionMedium, error) {
	return &repository.ScanToolMetadata{}, "", nil
}

func (impl *HandlerServiceImpl) fetchImageScanExecutionStepsForWfRequest(scanToolMetadata *repository.ScanToolMetadata) ([]*types.ImageScanningSteps, []*pipelineConfigBean.RefPluginObject, error) {
	return nil, nil, nil
}

func (impl *HandlerServiceImpl) checkIfCITriggerIsBlocked(pipeline *pipelineConfig.CiPipeline,
	ciMaterials []*pipelineConfig.CiPipelineMaterial, isJob bool) (bool, error) {
	return false, nil
}

func (impl *HandlerServiceImpl) handleWFIfCITriggerIsBlocked(ciWorkflow *pipelineConfig.CiWorkflow) (*pipelineConfig.CiWorkflow, error) {
	impl.Logger.Errorw("cannot trigger pipeline, blocked by mandatory plugin policy", "ciPipelineId", ciWorkflow.CiPipelineId)
	return &pipelineConfig.CiWorkflow{}, util.GetApiErrorAdapter(http.StatusInternalServerError, "500", "Invalid flow access, corrupt data possibility", "Invalid flow access, corrupt data possibility")
}

func (impl *HandlerServiceImpl) checkArgoSetupRequirement(envModal *repository2.Environment) error {
	return nil
}

func (impl *HandlerServiceImpl) updateWorkflowRequestForDigestPull(pipelineId int, workflowRequest *types.WorkflowRequest) (*types.WorkflowRequest, error) {
	return workflowRequest, nil
}

func (impl *HandlerServiceImpl) updateCIProjectDetailWithCloningMode(appId int, ciMaterial *pipelineConfig.CiPipelineMaterial,
	ciProjectDetail pipelineConfigBean.CiProjectDetails) (pipelineConfigBean.CiProjectDetails, error) {
	return ciProjectDetail, nil
}

func (impl *HandlerServiceImpl) updateWorkflowRequestWithRemoteConnConf(dockerRegistry *repository3.DockerArtifactStore,
	workflowRequest *types.WorkflowRequest) (*types.WorkflowRequest, error) {
	return workflowRequest, nil
}

func (impl *HandlerServiceImpl) updateWorkflowRequestWithEntSupportData(workflowRequest *types.WorkflowRequest) *types.WorkflowRequest {
	return workflowRequest
}

func (impl *HandlerServiceImpl) updateWorkflowRequestWithBuildxFlags(workflowRequest *types.WorkflowRequest,
	scope resourceQualifiers.Scope) (*types.WorkflowRequest, error) {
	workflowRequest.BuildxCacheModeMin = impl.buildxGlobalFlags.BuildxCacheModeMin
	workflowRequest.AsyncBuildxCacheExport = impl.buildxGlobalFlags.AsyncBuildxCacheExport
	workflowRequest.BuildxInterruptionMaxRetry = impl.buildxGlobalFlags.BuildxInterruptionMaxRetry
	return workflowRequest, nil
}

func (impl *HandlerServiceImpl) canSetK8sDriverData(workflowRequest *types.WorkflowRequest) bool {
	return impl.config != nil && impl.config.BuildxK8sDriverOptions != "" && workflowRequest.CiBuildConfig != nil &&
		workflowRequest.CiBuildConfig.DockerBuildConfig != nil
}

func (impl *HandlerServiceImpl) getK8sDriverOptions(workflowRequest *types.WorkflowRequest, targetPlatforms string) ([]map[string]string, error) {
	buildxK8sDriverOptions := make([]map[string]string, 0)
	err := json.Unmarshal([]byte(impl.config.BuildxK8sDriverOptions), &buildxK8sDriverOptions)
	if err != nil {
		return nil, err
	}
	return buildxK8sDriverOptions, nil
}

func (impl *HandlerServiceImpl) updateCIBuildConfig(ciBuildConfigBean *bean2.CiBuildConfigBean) *bean2.CiBuildConfigBean {
	defaultTargetPlatform := impl.config.DefaultTargetPlatform
	useBuildx := impl.config.UseBuildx
	if ciBuildConfigBean.DockerBuildConfig != nil {
		if ciBuildConfigBean.DockerBuildConfig.TargetPlatform == "" && useBuildx {
			ciBuildConfigBean.DockerBuildConfig.TargetPlatform = defaultTargetPlatform
			ciBuildConfigBean.DockerBuildConfig.UseBuildx = useBuildx
		}
		ciBuildConfigBean.DockerBuildConfig.BuildxProvenanceMode = impl.config.BuildxProvenanceMode
	}
	return ciBuildConfigBean
}

func updateBuildPrePostStepDataReq(req *pipelineConfigBean.BuildPrePostStepDataRequest, trigger *types.CiTriggerRequest) *pipelineConfigBean.BuildPrePostStepDataRequest {
	return req
}
