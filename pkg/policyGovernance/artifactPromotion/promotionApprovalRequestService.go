package artifactPromotion

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	repository1 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ArtifactPromotionApprovalService interface {
	HandleArtifactPromotionRequest(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) ([]bean.EnvironmentResponse, error)
	GetByPromotionRequestId(artifactPromotionApprovalRequest *repository.ArtifactPromotionApprovalRequest) (*bean.ArtifactPromotionApprovalResponse, error)
	FetchEnvironmentsList(workflowId, artifactId int) ([]bean.EnvironmentResponse, error)
}

type ArtifactPromotionApprovalServiceImpl struct {
	artifactPromotionApprovalRequestRepository repository.ArtifactPromotionApprovalRequestRepository
	logger                                     *zap.SugaredLogger
	ciPipelineRepository                       pipelineConfig.CiPipelineRepository
	pipelineRepository                         pipelineConfig.PipelineRepository
	userService                                user.UserService
	ciArtifactRepository                       repository2.CiArtifactRepository
	appWorkflowRepository                      appWorkflow.AppWorkflowRepository
	cdWorkflowRepository                       pipelineConfig.CdWorkflowRepository
	resourceFilterConditionsEvaluator          resourceFilter.ResourceFilterEvaluator
	imageTaggingService                        pipeline.ImageTaggingService
}

func NewArtifactPromotionApprovalServiceImpl(
	ArtifactPromotionApprovalRequestRepository repository.ArtifactPromotionApprovalRequestRepository,
	logger *zap.SugaredLogger,
	CiPipelineRepository pipelineConfig.CiPipelineRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	userService user.UserService,
	ciArtifactRepository repository2.CiArtifactRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	resourceFilterConditionsEvaluator resourceFilter.ResourceFilterEvaluator,
	imageTaggingService pipeline.ImageTaggingService,
) *ArtifactPromotionApprovalServiceImpl {
	return &ArtifactPromotionApprovalServiceImpl{
		artifactPromotionApprovalRequestRepository: ArtifactPromotionApprovalRequestRepository,
		logger:                            logger,
		ciPipelineRepository:              CiPipelineRepository,
		pipelineRepository:                pipelineRepository,
		userService:                       userService,
		ciArtifactRepository:              ciArtifactRepository,
		appWorkflowRepository:             appWorkflowRepository,
		cdWorkflowRepository:              cdWorkflowRepository,
		resourceFilterConditionsEvaluator: resourceFilterConditionsEvaluator,
		imageTaggingService:               imageTaggingService,
	}
}

func (impl ArtifactPromotionApprovalServiceImpl) FetchEnvironmentsList(workflowId, artifactId int) ([]bean.EnvironmentResponse, error) {
	allAppWorkflowMappings, err := impl.appWorkflowRepository.FindWFAllMappingByWorkflowId(workflowId)
	if err != nil {
		impl.logger.Errorw("error in finding the app workflow mappings using appWorkflowId", "workflowId", workflowId, "err", err)
		return nil, err
	}

	pipelineIds := make([]int, 0, len(allAppWorkflowMappings))
	for _, mapping := range allAppWorkflowMappings {
		if mapping.Type == appWorkflow.CDPIPELINE {
			pipelineIds = append(pipelineIds, mapping.ComponentId)
		}
	}

	pipelines, err := impl.pipelineRepository.FindByIdsIn(pipelineIds)
	if err != nil {
		return nil, err
	}
	envMap := make(map[string]repository1.Environment)
	for _, pipeline := range pipelines {
		envMap[pipeline.Name] = pipeline.Environment
	}

	// todo: fetch the policies by appId and envIds.
	policiesMap := make(map[string]bean.PromotionPolicy)
	if artifactId != 0 {
		ciArtifact, err := impl.ciArtifactRepository.Get(artifactId)
		if err != nil {
			impl.logger.Errorw("error in finding the artifact using id", "artifactId", artifactId, "err", err)
			errorResp := &util.ApiError{
				HttpStatusCode:  http.StatusInternalServerError,
				InternalMessage: fmt.Sprintf("error in finding artifact , err : %s", err.Error()),
				UserMessage:     "error in finding artifact",
			}
			if errors.Is(err, pg.ErrNoRows) {
				errorResp.UserMessage = "artifact not found"
				errorResp.HttpStatusCode = http.StatusConflict
			}

			return nil, errorResp
		}
		return impl.evaluatePoliciesOnArtifact(ciArtifact, envMap, policiesMap)
	}

	responseMap := make(map[string]bean.EnvironmentResponse)
	for envName, _ := range envMap {
		responseMap[envName] = bean.EnvironmentResponse{
			Name:                       envName,
			PromotionPossible:          false,
			PromotionValidationMessage: string(bean.POLICY_NOT_CONFIGURED),
			PromotionValidationState:   bean.POLICY_NOT_CONFIGURED,
			IsVirtualEnvironment:       envMap[envName].IsVirtualEnvironment,
		}
	}
	for envName, policy := range policiesMap {
		responseMap[envName] = bean.EnvironmentResponse{
			Name:                       envName,
			ApprovalCount:              policy.ApprovalMetaData.ApprovalCount,
			IsVirtualEnvironment:       envMap[envName].IsVirtualEnvironment,
			PromotionValidationMessage: "",
			PromotionValidationState:   bean.EMPTY,
		}
	}

	result := make([]bean.EnvironmentResponse, 0, len(responseMap))
	for _, envResponse := range responseMap {
		result = append(result, envResponse)
	}

	return result, nil

}

func (impl ArtifactPromotionApprovalServiceImpl) computeFilterParams(ciArtifact *repository2.CiArtifact) ([]resourceFilter.ExpressionParam, error) {
	var ciMaterials []repository2.CiMaterialInfo
	err := json.Unmarshal([]byte(ciArtifact.MaterialInfo), &ciMaterials)
	if err != nil {
		impl.logger.Errorw("error in parsing ci artifact material info")
		return nil, err
	}

	commitDetailsList := make([]resourceFilter.CommitDetails, 0, len(ciMaterials))
	for _, ciMaterial := range ciMaterials {
		repoUrl := ciMaterial.Material.ScmConfiguration.URL
		commitMessage := ""
		branch := ""
		if ciMaterial.Material.Type == "git" {
			repoUrl = ciMaterial.Material.GitConfiguration.URL
		}
		if ciMaterial.Modifications != nil && len(ciMaterial.Modifications) > 0 {
			modification := ciMaterial.Modifications[0]
			commitMessage = modification.Message
			branch = modification.Branch
		}
		commitDetailsList = append(commitDetailsList, resourceFilter.CommitDetails{
			Repo:          repoUrl,
			CommitMessage: commitMessage,
			Branch:        branch,
		})
	}

	imageTags, err := impl.imageTaggingService.GetTagsByArtifactId(ciArtifact.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching the image tags using artifact id", "artifactId", ciArtifact.Id, "err", err)
		return nil, err
	}

	releaseTags := make([]string, 0, len(imageTags))
	for _, imageTag := range imageTags {
		releaseTags = append(releaseTags, imageTag.TagName)
	}
	params := resourceFilter.GetParamsFromArtifact(ciArtifact.Image, releaseTags, commitDetailsList...)
	return params, nil
}

func (impl ArtifactPromotionApprovalServiceImpl) evaluatePoliciesOnArtifact(ciArtifact *repository2.CiArtifact, envMap map[string]repository1.Environment, policiesMap map[string]bean.PromotionPolicy) ([]bean.EnvironmentResponse, error) {
	params, err := impl.computeFilterParams(ciArtifact)
	if err != nil {
		impl.logger.Errorw("error in finding the required CEL expression parameters for using ciArtifact", "err", err)
		return nil, err
	}
	responseMap := make(map[string]bean.EnvironmentResponse)
	for envName, _ := range envMap {
		responseMap[envName] = bean.EnvironmentResponse{
			Name:                       envName,
			PromotionPossible:          false,
			PromotionValidationMessage: string(bean.POLICY_NOT_CONFIGURED),
			PromotionValidationState:   bean.POLICY_NOT_CONFIGURED,
			IsVirtualEnvironment:       envMap[envName].IsVirtualEnvironment,
		}
	}

	for envName, policy := range policiesMap {
		evaluationResult, err := impl.resourceFilterConditionsEvaluator.EvaluateFilter(policy.Conditions, resourceFilter.ExpressionMetadata{Params: params})
		if err != nil {
			impl.logger.Errorw("evaluation failed with error", "policyConditions", policy.Conditions, "envName", envName, policy.Conditions, "params", params, "err", err)
			responseMap[envName] = bean.EnvironmentResponse{
				ApprovalCount:            policy.ApprovalMetaData.ApprovalCount,
				PromotionPossible:        false,
				PromotionValidationState: bean.POLICY_EVALUATION_ERRORED,
			}
			continue
		}
		envResp := bean.EnvironmentResponse{
			ApprovalCount:     policy.ApprovalMetaData.ApprovalCount,
			PromotionPossible: evaluationResult,
		}
		if !evaluationResult {
			envResp.PromotionValidationMessage = string(bean.BLOCKED_BY_POLICY)
			envResp.PromotionValidationState = bean.BLOCKED_BY_POLICY
		}
		responseMap[envName] = envResp
	}
	result := make([]bean.EnvironmentResponse, 0, len(responseMap))
	for _, envResponse := range responseMap {
		result = append(result, envResponse)
	}
	return result, nil
}

func (impl ArtifactPromotionApprovalServiceImpl) approveArtifactPromotion(request *bean.ArtifactPromotionRequest) ([]bean.EnvironmentResponse, error) {
	// get request and check if it is promoted already.
	// attempt approving this by creating new resource_approval_user_data
	return nil, nil
}

func (impl ArtifactPromotionApprovalServiceImpl) HandleArtifactPromotionRequest(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) ([]bean.EnvironmentResponse, error) {
	switch request.Action {

	case bean.ACTION_PROMOTE:
		return impl.promoteArtifact(request)
	case bean.ACTION_APPROVE:
		return impl.approveArtifactPromotion(request)
	case bean.ACTION_CANCEL:

		_, err := impl.cancelPromotionApprovalRequest(request)
		if err != nil {
			impl.logger.Errorw("error in canceling artifact promotion approval request", "promotionRequestId", request.PromotionRequestId, "err", err)
			return nil, err
		}
		return nil, nil

	}
	return nil, errors.New("unknown action")
}

func (impl ArtifactPromotionApprovalServiceImpl) validateSourceAndFetchAppWorkflow(request *bean.ArtifactPromotionRequest) (*appWorkflow.AppWorkflow, error) {
	// ciArtifact, err := impl.ciArtifactRepository.Get(request.ArtifactId)
	// if err != nil {
	// 	impl.logger.Errorw("error in finding the artifact using id", "artifactId", request.ArtifactId, "err", err)
	// 	errorResp := &util.ApiError{
	// 		HttpStatusCode:  http.StatusInternalServerError,
	// 		InternalMessage: fmt.Sprintf("error in finding artifact , err : %s", err.Error()),
	// 		UserMessage:     "error in finding artifact",
	// 	}
	// 	if errors.Is(err, pg.ErrNoRows) {
	// 		errorResp.UserMessage = "artifact not found"
	// 		errorResp.HttpStatusCode = http.StatusConflict
	// 	}
	//
	// 	return nil, errorResp
	// }
	appWorkflowMapping := &appWorkflow.AppWorkflowMapping{}
	var err error
	if request.SourceType == bean.SOURCE_TYPE_CI || request.SourceType == bean.SOURCE_TYPE_WEBHOOK {
		appWorkflowMapping, err = impl.appWorkflowRepository.FindByWorkflowIdAndCiSource(request.WorkflowId)
		if err != nil {
			// log
			return nil, err
		}
	} else {
		// source type will be cd and source name will be envName.
		// get pipeline using appId and env name and get the workflowMapping
		pipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(request.AppId, request.SourcePipelineId)
		if err != nil {
			// log
			return nil, err
		}
		if len(pipelines) == 0 {
			//  throw error that source is not found
			return nil, err
		}

		pipeline := pipelines[0]
		appWorkflowMapping, err = impl.appWorkflowRepository.FindWFMappingByComponent(appWorkflow.CDPIPELINE, pipeline.Id)
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				// log that could not find pipeline for env in the given app.
			}
			// log
			return nil, err
		}
	}

	if request.WorkflowId != appWorkflowMapping.AppWorkflowId {
		// log evaluation failed
		return nil, errors.New("source is not in the given workflow")
	}
	workflow, err := impl.appWorkflowRepository.FindById(appWorkflowMapping.Id)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			// log workflow not exists error
			return nil, err
		}

		// log error
	}

	return workflow, nil
}
func (impl ArtifactPromotionApprovalServiceImpl) promoteArtifact(request *bean.ArtifactPromotionRequest) ([]bean.EnvironmentResponse, error) {
	// 	step1: validate if artifact is deployed/created at the source pipeline.
	//      step1: if source is cd , check if this artifact is deployed on these environments
	//  step2: check if destination pipeline is topologically downwards from the source pipeline and also source and destination are on the same subtree.
	// 	step3: check if promotion request for this artifact on this destination pipeline has already been raised.
	//  step4: check if this artifact on this destination pipeline has already been promoted
	//  step5: raise request.

	// fetch artifact
	response := make(map[string]bean.EnvironmentResponse)
	for _, env := range request.EnvironmentNames {
		envResponse := bean.EnvironmentResponse{
			Name:                       env,
			PromotionValidationState:   bean.PIPELINE_NOT_FOUND,
			PromotionValidationMessage: string(bean.PIPELINE_NOT_FOUND),
		}
		response[env] = envResponse
	}

	allowedEnvs := make([]int, 0, len(request.EnvNameIdMap))
	for _, envId := range request.EnvNameIdMap {
		allowedEnvs = append(allowedEnvs, envId)
	}

	ciArtifact, err := impl.ciArtifactRepository.Get(request.ArtifactId)
	if err != nil {
		impl.logger.Errorw("error in finding the artifact using id", "artifactId", request.ArtifactId, "err", err)
		errorResp := &util.ApiError{
			HttpStatusCode:  http.StatusInternalServerError,
			InternalMessage: fmt.Sprintf("error in finding artifact , err : %s", err.Error()),
			UserMessage:     "error in finding artifact",
		}
		if errors.Is(err, pg.ErrNoRows) {
			errorResp.UserMessage = "artifact not found"
			errorResp.HttpStatusCode = http.StatusNotFound
		}

		return nil, errorResp
	}
	// componentId := ciArtifact.PipelineId
	// componentType := appWorkflow.CIPIPELINE
	// if ciArtifact.ExternalCiPipelineId != 0 {
	// 	componentType = appWorkflow.WEBHOOK
	// 	componentId = ciArtifact.ExternalCiPipelineId
	// }
	// workflowMapping, err := impl.appWorkflowRepository.FindWFMappingByComponent(componentType, componentId)
	// if err != nil {
	// 	impl.logger.Errorw("error in finding the app workflow mapping using componentId and componentType", "componentType", componentType, "componentId", componentId, "err", err)
	// 	errorResp := &util.ApiError{
	// 		HttpStatusCode:  http.StatusInternalServerError,
	// 		InternalMessage: fmt.Sprintf("error in finding app worlflow mapping , err : %s", err.Error()),
	// 		UserMessage:     "error occurred in promoting artifact, could not resolve workflow",
	// 	}
	// 	if errors.Is(err, pg.ErrNoRows) {
	// 		errorResp.UserMessage = "workflow not found"
	// 		errorResp.HttpStatusCode = http.StatusConflict
	// 	}
	// 	return nil, errorResp
	// }

	workflow, err := impl.validateSourceAndFetchAppWorkflow(request)
	if err != nil {
		return nil, err
	}
	pipelines, err := impl.pipelineRepository.FindByAppIdsAndEnvironmentIds([]int{request.AppId}, allowedEnvs)
	if err != nil {
		impl.logger.Errorw("error in finding the pipelines for the app on given environments", "appId", request.AppId, "envIds", allowedEnvs, "err", err)
		errorResp := &util.ApiError{
			HttpStatusCode:  http.StatusInternalServerError,
			InternalMessage: fmt.Sprintf("error in finding the pipelines for the app on given environments , err : %s", err.Error()),
			UserMessage:     "error occurred in promoting artifact, error in resolving environments on ",
		}
		if errors.Is(err, pg.ErrNoRows) {
			errorResp.UserMessage = "could not find the given environments in this app "
			errorResp.HttpStatusCode = http.StatusNotFound
		}
		return nil, errorResp
	}

	pipelineIdVsEnvNameMap := make(map[int]string)
	pipelineIds := make([]int, 0, len(pipelines))
	for _, pipeline := range pipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
		pipelineIdVsEnvNameMap[pipeline.Id] = request.EnvIdNameMap[pipeline.EnvironmentId]
		EnvResponse := response[pipelineIdVsEnvNameMap[pipeline.Id]]
		EnvResponse.PromotionValidationState = bean.EMPTY
		response[pipelineIdVsEnvNameMap[pipeline.Id]] = EnvResponse
	}

	sourcePipelineId := 0
	allAppWorkflowMappings, err := impl.appWorkflowRepository.FindWFAllMappingByWorkflowId(workflow.Id)
	if err != nil {
		impl.logger.Errorw("error in finding the app workflow mappings", "err", err)
		return nil, err
	}
	// for sourceType CI/Webhook, we don't have to validate as this will be the root node of the DAG.
	if request.SourceType == bean.SOURCE_TYPE_CD {
		tree := make(map[int][]int)
		for _, appWorkflowMapping := range allAppWorkflowMappings {
			if appWorkflowMapping.Type == appWorkflow.CDPIPELINE {
				envName, ok := pipelineIdVsEnvNameMap[appWorkflowMapping.ComponentId]
				if ok && envName == request.SourceName {
					// setting sourcePipelineId here
					sourcePipelineId = appWorkflowMapping.ComponentId
				}
			}

			// create the tree from the DAG excluding the ci source
			if appWorkflowMapping.Type == appWorkflow.CDPIPELINE && appWorkflowMapping.ParentType == appWorkflow.CDPIPELINE {
				tree[appWorkflowMapping.ParentId] = append(tree[appWorkflowMapping.ParentId], appWorkflowMapping.ComponentId)
			}
		}

		// if sourcePipelineId is 0, then the source pipeline given by user is not found in the workflow.
		if sourcePipelineId == 0 {
			return nil, &util.ApiError{
				HttpStatusCode:  http.StatusBadRequest,
				InternalMessage: fmt.Sprintf("no pipeline found against given source environment %s", request.SourceName),
				UserMessage:     fmt.Sprintf("no pipeline found against given source environment %s", request.SourceName),
			}
		}

		deployed, err := impl.checkIfDeployedAtSource(ciArtifact.Id, sourcePipelineId)
		if err != nil {
			// todo: log error
			return nil, err
		}
		if !deployed {
			return nil, &util.ApiError{
				HttpStatusCode:  http.StatusConflict,
				InternalMessage: fmt.Sprintf("artifact is not deployed on the source environment %s", request.SourceName),
				UserMessage:     fmt.Sprintf("artifact is not deployed on the source environment %s", request.SourceName),
			}
		}

		for _, pipelineId := range pipelineIds {
			if !util.IsAncestor(tree, sourcePipelineId, pipelineId) {
				EnvResponse := response[pipelineIdVsEnvNameMap[pipelineId]]
				EnvResponse.PromotionValidationState = bean.SOURCE_AND_DESTINATION_PIPELINE_MISMATCH
			}
		}
	}

	impl.checkPromotionPolicyGovernance(request.AppId, allowedEnvs)
	for _, pipelineId := range pipelineIds {

		EnvResponse := response[pipelineIdVsEnvNameMap[pipelineId]]
		// these
		if EnvResponse.PromotionValidationState == bean.EMPTY {
			// todo send policy of this pipeline
			state, msg, err := impl.raisePromoteRequest(request, pipelineId, nil)
			if err != nil {
				impl.logger.Errorw("error in raising promotion request for the pipeline", "pipelineId", pipelineId, "artifactId", ciArtifact.Id, "err", err)
				EnvResponse.PromotionValidationState = bean.ERRORED
				EnvResponse.PromotionValidationMessage = err.Error()
			}
			EnvResponse.PromotionValidationState = state
			EnvResponse.PromotionValidationMessage = msg
		}

		response[pipelineIdVsEnvNameMap[pipelineId]] = EnvResponse
	}
	envResponses := make([]bean.EnvironmentResponse, len(response))
	for _, resp := range response {
		envResponses = append(envResponses, resp)
	}
	return envResponses, nil
}

func (impl ArtifactPromotionApprovalServiceImpl) checkPromotionPolicyGovernance(appId int, envIds []int) map[int]interface{} {
	// todo: implement me
	return nil
}

func (impl ArtifactPromotionApprovalServiceImpl) raisePromoteRequest(request *bean.ArtifactPromotionRequest, pipelineId int, promotionPolicyMetadata interface{}) (bean.PromotionValidationState, string, error) {
	requests, err := impl.artifactPromotionApprovalRequestRepository.FindAwaitedRequestByPipelineIdAndArtifactId(pipelineId, request.ArtifactId)
	if err != nil {
		impl.logger.Errorw("error in finding the pending promotion request using pipelineId and artifactId", "pipelineId", pipelineId, "artifactId", request.ArtifactId)
		return bean.ERRORED, err.Error(), err
	}

	if len(requests) >= 1 {
		return bean.ALREADY_REQUEST_RAISED, string(bean.ALREADY_REQUEST_RAISED), nil
	}

	promotedRequest, err := impl.artifactPromotionApprovalRequestRepository.FindPromotedRequestByPipelineIdAndArtifactId(pipelineId, request.ArtifactId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in finding the promoted request using pipelineId and artifactId", "pipelineId", pipelineId, "artifactId", request.ArtifactId)
		return bean.ERRORED, err.Error(), err
	}

	if promotedRequest.Id > 0 {
		return bean.ARTIFACT_ALREADY_PROMOTED, string(bean.ARTIFACT_ALREADY_PROMOTED), nil
	}

	promotionRequest := &repository.ArtifactPromotionApprovalRequest{
		SourceType:            bean.CI,
		SourcePipelineId:      request.SourcePipelineId,
		DestinationPipelineId: pipelineId,
		Status:                bean.AWAITING_APPROVAL,
		Active:                true,
		ArtifactId:            request.ArtifactId,
		// todo: update below fields
		PolicyId:                0,
		PolicyEvaluationAuditId: 0,
	}

	status := bean.SENT_FOR_APPROVAL
	// todo: change this to promotionPolicyMetadata.approvalCount
	if promotionPolicyMetadata == 0 {
		promotionRequest.Status = bean.PROMOTED
		status = bean.PROMOTION_SUCCESSFUL
	}
	_, err = impl.artifactPromotionApprovalRequestRepository.Create(promotionRequest)
	if err != nil {
		impl.logger.Errorw("error in finding the pending promotion request using pipelineId and artifactId", "pipelineId", pipelineId, "artifactId", request.ArtifactId)
		return bean.ERRORED, err.Error(), err
	}

	return status, string(status), nil

}

func (impl ArtifactPromotionApprovalServiceImpl) checkIfDeployedAtSource(ciArtifactId, sourcePipelineId int) (bool, error) {
	// todo: implement me

	return true, nil
}

func (impl ArtifactPromotionApprovalServiceImpl) cancelPromotionApprovalRequest(request *bean.ArtifactPromotionRequest) (*bean.ArtifactPromotionRequest, error) {
	artifactPromotionDao, err := impl.artifactPromotionApprovalRequestRepository.FindById(request.PromotionRequestId)
	if err == pg.ErrNoRows {
		impl.logger.Errorw("artifact promotion approval request not found for given id", "promotionRequestId", request.PromotionRequestId, "err", err)
		return nil, &util.ApiError{
			HttpStatusCode:  http.StatusNotFound,
			InternalMessage: bean.ArtifactPromotionRequestNotFoundErr,
			UserMessage:     bean.ArtifactPromotionRequestNotFoundErr,
		}
	}
	if err != nil {
		impl.logger.Errorw("error in fetching artifact promotion request by id", "artifactPromotionRequestId", request.PromotionRequestId, "err", err)
		return nil, err
	}
	artifactPromotionDao.Active = false
	artifactPromotionDao.Status = bean.CANCELED
	artifactPromotionDao.UpdatedOn = time.Now()
	artifactPromotionDao.UpdatedBy = request.UserId
	_, err = impl.artifactPromotionApprovalRequestRepository.Update(artifactPromotionDao)
	if err != nil {
		impl.logger.Errorw("error in updating artifact promotion approval request", "artifactPromotionRequestId", request.PromotionRequestId, "err", err)
		return nil, err
	}
	return nil, err
}

func (impl ArtifactPromotionApprovalServiceImpl) GetByPromotionRequestId(artifactPromotionApprovalRequest *repository.ArtifactPromotionApprovalRequest) (*bean.ArtifactPromotionApprovalResponse, error) {

	sourceType := artifactPromotionApprovalRequest.SourceType.GetSourceType()

	var source string
	if artifactPromotionApprovalRequest.SourceType == bean.CD {
		cdPipeline, err := impl.pipelineRepository.FindById(artifactPromotionApprovalRequest.SourcePipelineId)
		if err != nil {
			impl.logger.Errorw("error in fetching cdPipeline by Id", "cdPipelineId", artifactPromotionApprovalRequest.SourcePipelineId, "err", err)
			return nil, err
		}
		source = cdPipeline.Environment.Name
	}

	destCDPipeline, err := impl.pipelineRepository.FindById(artifactPromotionApprovalRequest.DestinationPipelineId)
	if err != nil {
		impl.logger.Errorw("error in fetching cdPipeline by Id", "cdPipelineId", artifactPromotionApprovalRequest.DestinationPipelineId, "err", err)
		return nil, err
	}

	artifactPromotionRequestUser, err := impl.userService.GetByIdWithoutGroupClaims(artifactPromotionApprovalRequest.CreatedBy)
	if err != nil {
		impl.logger.Errorw("error in fetching user details by id", "userId", artifactPromotionApprovalRequest.CreatedBy, "err", err)
		return nil, err
	}

	artifactPromotionApprovalResponse := &bean.ArtifactPromotionApprovalResponse{
		SourceType:      sourceType,
		Source:          source,
		Destination:     destCDPipeline.Environment.Name,
		RequestedBy:     artifactPromotionRequestUser.EmailId,
		ApprovedUsers:   make([]string, 0), // get by deployment_approval_user_data
		RequestedOn:     artifactPromotionApprovalRequest.CreatedOn,
		PromotedOn:      artifactPromotionApprovalRequest.UpdatedOn,
		PromotionPolicy: "", // todo
	}

	return artifactPromotionApprovalResponse, nil

}
