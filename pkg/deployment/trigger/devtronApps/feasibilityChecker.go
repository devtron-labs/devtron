/*
 * Copyright (c) 2024. Devtron Inc.
 */

package devtronApps

import (
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/deploymentWindow"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	constants2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/constants"
	"github.com/go-pg/pg"
	"net/http"
	"time"
)

func (impl *TriggerServiceImpl) CheckFeasibility(triggerRequirementRequest *bean.TriggerRequirementRequestDto) (*bean.TriggerFeasibilityResponse, bool, bool, error) {
	var approvalRequestId int
	var err error
	artifactId := triggerRequirementRequest.TriggerRequest.Artifact.Id
	cdPipeline := triggerRequirementRequest.TriggerRequest.Pipeline
	triggeredBy := triggerRequirementRequest.TriggerRequest.TriggeredBy
	IsDeployStage := triggerRequirementRequest.Stage == resourceFilter.Deploy
	artifact := triggerRequirementRequest.TriggerRequest.Artifact

	if triggerRequirementRequest.DeploymentType == models.DEPLOYMENTTYPE_PRE || triggerRequirementRequest.DeploymentType == models.DEPLOYMENTTYPE_POST {
		err = impl.CheckIfPreOrPostExists(cdPipeline, triggerRequirementRequest.TriggerRequest.WorkflowType, triggerRequirementRequest.DeploymentType)
		if err != nil {
			impl.logger.Errorw("error encountered in CheckFeasibility", "err", err, "cdPipelineId", cdPipeline.Id)
			return nil, false, false, err
		}
	}

	if bean2.CheckIfDeploymentTypePrePostOrDeployOrUnknown(triggerRequirementRequest.DeploymentType) {
		isArtifactAvailable, err := impl.isArtifactDeploymentAllowed(cdPipeline, artifact, triggerRequirementRequest.TriggerRequest.WorkflowType)
		if err != nil {
			impl.logger.Errorw("error in checking artifact availability on cdPipeline", "artifactId", artifactId, "cdPipelineId", cdPipeline.Id, "err", err)
			return nil, false, false, err
		}
		if !isArtifactAvailable {
			return nil, false, false, util.NewApiError().WithHttpStatusCode(http.StatusConflict).WithUserMessage(constants2.ARTIFACT_UNAVAILABLE_MESSAGE).WithCode(constants.ArtifactNotAvailable)
		}

		_, err = impl.isImagePromotionPolicyViolated(cdPipeline, artifact.Id, triggeredBy)
		if err != nil {
			impl.logger.Errorw("error in checking if image promotion policy violated", "artifactId", artifactId, "cdPipelineId", cdPipeline.Id, "err", err)
			return nil, false, false, err
		}
	}

	if IsDeployStage {
		// checking approval node only for deployment
		approvalRequestId, err = impl.checkApprovalNodeForDeployment(triggeredBy, cdPipeline, artifactId)
		if err != nil {
			impl.logger.Errorw("error encountered in CheckFeasibility", "artifactId", artifactId, "err", err)
			return nil, false, false, err
		}
	}
	filters, err := impl.resourceFilterService.GetFiltersByScope(triggerRequirementRequest.Scope)
	if err != nil {
		impl.logger.Errorw("error in getting resource filters for the pipeline", "scope", triggerRequirementRequest.Scope, "err", err)
		return nil, false, false, err
	}

	// get releaseTags from imageTaggingService
	imageTagNames, err := impl.imageTaggingService.GetTagNamesByArtifactId(artifactId)
	if err != nil {
		impl.logger.Errorw("error in getting image tags for the given artifact id", "artifactId", artifactId, "err", err)
		return nil, false, false, err
	}

	materialInfos, err := artifact.GetMaterialInfo()
	if err != nil {
		impl.logger.Errorw("error in getting material info for the given artifact", "artifactId", artifactId, "materialInfo", artifact.MaterialInfo, "err", err)
		return nil, false, false, err
	}

	filterState, filterIdVsState, err := impl.resourceFilterService.CheckForResource(filters, triggerRequirementRequest.TriggerRequest.Artifact.Image, imageTagNames, materialInfos)
	if err != nil {
		impl.logger.Errorw("error encountered in CheckFeasibility", "imageTagNames", imageTagNames, "filters", filters, "err", err)
		return nil, false, false, err
	}

	// allow or block w.r.t filterState
	if filterState != resourceFilter.ALLOW {
		return adapter.GetTriggerFeasibilityResponse(approvalRequestId, triggerRequirementRequest.TriggerRequest, filterIdVsState, filters), true, false, &util.ApiError{Code: constants.FilteringConditionFail, InternalMessage: "the artifact does not pass filtering condition", UserMessage: "the artifact does not pass filtering condition"}
	}

	triggerRequest, isDeploymentWindowByPassed, err := impl.checkForDeploymentWindow(triggerRequirementRequest.TriggerRequest, triggerRequirementRequest.Stage)
	if err != nil {
		impl.logger.Errorw("error encountered in CheckFeasibility", "triggerRequest", triggerRequirementRequest.TriggerRequest)
		return adapter.GetTriggerFeasibilityResponse(approvalRequestId, triggerRequirementRequest.TriggerRequest, filterIdVsState, filters), true, false, err
	}

	return adapter.GetTriggerFeasibilityResponse(approvalRequestId, triggerRequest, filterIdVsState, filters), true, isDeploymentWindowByPassed, nil
}

func (impl *TriggerServiceImpl) checkForDeploymentWindow(triggerRequest bean.TriggerRequest, stage resourceFilter.ReferenceType) (bean.TriggerRequest, bool, error) {
	triggerTime := time.Now()
	actionState, envState, err := impl.deploymentWindowService.GetStateForAppEnv(triggerTime, triggerRequest.Pipeline.AppId, triggerRequest.Pipeline.EnvironmentId, triggerRequest.TriggeredBy)
	if err != nil {
		return triggerRequest, false, fmt.Errorf("failed to fetch deployment window state %s %d %d %d %v", triggerTime, triggerRequest.Pipeline.AppId, triggerRequest.Pipeline.EnvironmentId, triggerRequest.TriggeredBy, err)
	}
	triggerRequest.TriggerMessage = actionState.GetBypassActionMessageForProfileAndState(envState)
	triggerRequest.DeploymentWindowState = envState

	if !isDeploymentAllowed(triggerRequest, actionState) {
		err = impl.handleBlockedTrigger(triggerRequest, stage)
		if err != nil {
			return triggerRequest, false, err
		}
		return triggerRequest, false, deploymentWindow.GetActionBlockedError(actionState.GetErrorMessageForProfileAndState(envState), constants.DeploymentWindowFail)
	}
	return triggerRequest, actionState.IsActionBypass(), nil
}

func (impl *TriggerServiceImpl) isImagePromotionPolicyViolated(cdPipeline *pipelineConfig.Pipeline, artifactId int, userId int32) (bool, error) {
	promotionPolicy, err := impl.artifactPromotionDataReadService.GetPromotionPolicyByAppAndEnvId(cdPipeline.AppId, cdPipeline.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in fetching image promotion policy for checking trigger access", "cdPipelineId", cdPipeline.Id, "err", err)
		return false, err
	}
	if promotionPolicy != nil && promotionPolicy.Id > 0 {
		if promotionPolicy.BlockApproverFromDeploy() {
			isUserApprover, err := impl.artifactPromotionDataReadService.IsUserApprover(artifactId, cdPipeline.Id, userId)
			if err != nil {
				impl.logger.Errorw("error in checking if user is approver or not", "artifactId", artifactId, "cdPipelineId", cdPipeline.Id, "err", err)
				return false, err
			}
			if isUserApprover {
				impl.logger.Errorw("error in cd trigger, user who has approved the image for promotion cannot deploy")
				return true, util.NewApiError().WithHttpStatusCode(http.StatusForbidden).WithUserMessage(bean.ImagePromotionPolicyValidationErr).WithInternalMessage(bean.ImagePromotionPolicyValidationErr)
			}
		}
	}
	return false, nil
}

func (impl *TriggerServiceImpl) CheckIfPreOrPostExists(cdPipeline *pipelineConfig.Pipeline, workflowType bean2.WorkflowType, deploymentType models.DeploymentType) error {

	switch deploymentType {
	case models.DEPLOYMENTTYPE_PRE:
		if len(cdPipeline.PreStageConfig) != 0 {
			return nil
		}

	case models.DEPLOYMENTTYPE_POST:
		if len(cdPipeline.PostStageConfig) != 0 {
			return nil
		}
	default:
		return util.NewApiError().WithHttpStatusCode(http.StatusBadRequest).WithUserMessage(constants2.DEPLOYMENT_TYPE_NOT_SUPPORTED_MESSAGE)

	}
	pipelineStage, err := impl.pipelineStageService.GetCdStageByCdPipelineIdAndStageType(cdPipeline.Id, workflowType.WorkflowTypeToStageType())
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error encountered in CheckFeasibility", "err", err)
		return err
	}
	exists := pipelineStage != nil && pipelineStage.Id != 0
	if !exists && deploymentType == models.DEPLOYMENTTYPE_PRE {
		return util.NewApiError().WithHttpStatusCode(http.StatusNotFound).WithUserMessage(constants2.PRE_DEPLOYMENT_DOES_NOT_EXIST_MESSAGE).WithCode(constants.PreCDDoesNotExists)
	}
	if !exists && deploymentType == models.DEPLOYMENTTYPE_POST {
		return util.NewApiError().WithHttpStatusCode(http.StatusNotFound).WithUserMessage(constants2.POST_DEPLOYMENT_DOES_NOT_EXIST_MESSAGE).WithCode(constants.PostCDDoesNotExists)
	}
	return nil
}
