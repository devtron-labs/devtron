package devtronApps

import (
	"fmt"
	"github.com/devtron-labs/devtron/enterprise/pkg/deploymentWindow"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"time"
)

func (impl *TriggerServiceImpl) CheckFeasibility(triggerRequirementRequest *bean.TriggerRequirementRequestDto) (*bean.TriggerFeasibilityResponse, error) {
	var approvalRequestId int
	var err error
	artifactId := triggerRequirementRequest.TriggerRequest.Artifact.Id
	cdPipeline := triggerRequirementRequest.TriggerRequest.Pipeline
	triggeredBy := triggerRequirementRequest.TriggerRequest.TriggeredBy
	IsDeployStage := triggerRequirementRequest.Stage == resourceFilter.Deploy

	if IsDeployStage {
		// checking approval node only for deployment
		approvalRequestId, err = impl.checkApprovalNodeForDeployment(triggeredBy, cdPipeline, artifactId)
		if err != nil {
			impl.logger.Errorw("error encountered in CheckFeasibility", "artifactId", artifactId, "err", err)
			return nil, err
		}
	}
	filters, err := impl.resourceFilterService.GetFiltersByScope(triggerRequirementRequest.Scope)
	if err != nil {
		impl.logger.Errorw("error in getting resource filters for the pipeline", "scope", triggerRequirementRequest.Scope, "err", err)
		return nil, err
	}

	// get releaseTags from imageTaggingService
	imageTagNames, err := impl.imageTaggingService.GetTagNamesByArtifactId(artifactId)
	if err != nil {
		impl.logger.Errorw("error in getting image tags for the given artifact id", "artifactId", artifactId, "err", err)
		return nil, err
	}

	filterState, filterIdVsState, err := impl.resourceFilterService.CheckForResource(filters, triggerRequirementRequest.TriggerRequest.Artifact.Image, imageTagNames)
	if err != nil {
		impl.logger.Errorw("error encountered in CheckFeasibility", "imageTagNames", imageTagNames, "filters", filters, "err", err)
		return nil, err
	}

	// allow or block w.r.t filterState
	if filterState != resourceFilter.ALLOW {
		return nil, &util.ApiError{Code: constants.FilteringConditionFail, InternalMessage: "the artifact does not pass filtering condition", UserMessage: "the artifact does not pass filtering condition"}
	}

	triggerRequest, err := impl.checkForDeploymentWindow(triggerRequirementRequest.TriggerRequest, triggerRequirementRequest.Stage)
	if err != nil {
		impl.logger.Errorw("error encountered in CheckFeasibility", "triggerRequest", triggerRequirementRequest.TriggerRequest)
		return nil, err
	}

	return adapter.GetTriggerFeasibilityResponse(approvalRequestId, triggerRequest, filterIdVsState, filters), nil
}

func (impl *TriggerServiceImpl) checkForDeploymentWindow(triggerRequest bean.TriggerRequest, stage resourceFilter.ReferenceType) (bean.TriggerRequest, error) {
	triggerTime := time.Now()
	actionState, envState, err := impl.deploymentWindowService.GetStateForAppEnv(triggerTime, triggerRequest.Pipeline.AppId, triggerRequest.Pipeline.EnvironmentId, triggerRequest.TriggeredBy)
	if err != nil {
		return triggerRequest, fmt.Errorf("failed to fetch deployment window state %s %d %d %d %v", triggerTime, triggerRequest.Pipeline.AppId, triggerRequest.Pipeline.EnvironmentId, triggerRequest.TriggeredBy, err)
	}
	triggerRequest.TriggerMessage = actionState.GetBypassActionMessageForProfileAndState(envState)
	triggerRequest.DeploymentWindowState = envState

	if !isDeploymentAllowed(triggerRequest, actionState) {
		err = impl.handleBlockedTrigger(triggerRequest, stage)
		if err != nil {
			return triggerRequest, err
		}
		return triggerRequest, deploymentWindow.GetActionBlockedError(actionState.GetErrorMessageForProfileAndState(envState), constants.DeploymentWindowFail)
	}
	return triggerRequest, nil
}
