package artifactPromotion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/cluster"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/constants"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/read"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/workflow/dag"
	"github.com/go-pg/pg"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type ApprovalRequestService interface {
	HandleArtifactPromotionRequest(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) ([]bean.EnvironmentPromotionMetaData, error)
	GetPromotionRequestById(promotionRequestId int) (*bean.ArtifactPromotionApprovalResponse, error)
	FetchWorkflowPromoteNodeList(token string, workflowId int, artifactId int, rbacChecker func(token string, appName string, envNames []string) map[string]bool) (*bean.EnvironmentListingResponse, error)
	FetchApprovalAllowedEnvList(artifactId int, environmentName string, userId int32, token string, promotionApproverAuth func(string, string) bool) ([]bean.EnvironmentApprovalMetadata, error)
}

type ApprovalRequestServiceImpl struct {
	artifactPromotionApprovalRequestRepository repository.RequestRepository
	logger                                     *zap.SugaredLogger
	ciPipelineRepository                       pipelineConfig.CiPipelineRepository
	pipelineRepository                         pipelineConfig.PipelineRepository
	pipelineStageService                       pipeline.PipelineStageService
	environmentService                         cluster.EnvironmentService
	userService                                user.UserService
	ciArtifactRepository                       repository2.CiArtifactRepository
	appWorkflowRepository                      appWorkflow.AppWorkflowRepository
	cdWorkflowRepository                       pipelineConfig.CdWorkflowRepository
	resourceFilterConditionsEvaluator          resourceFilter.ResourceFilterEvaluator
	resourceFilterEvaluationAuditService       resourceFilter.FilterEvaluationAuditService
	imageTaggingService                        pipeline.ImageTaggingService
	promotionPolicyDataReadService             read.ArtifactPromotionDataReadService
	requestApprovalUserdataRepo                pipelineConfig.RequestApprovalUserdataRepository
	workflowDagExecutor                        dag.WorkflowDagExecutor
}

func NewApprovalRequestServiceImpl(
	ArtifactPromotionApprovalRequestRepository repository.RequestRepository,
	logger *zap.SugaredLogger,
	CiPipelineRepository pipelineConfig.CiPipelineRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	userService user.UserService,
	ciArtifactRepository repository2.CiArtifactRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	resourceFilterConditionsEvaluator resourceFilter.ResourceFilterEvaluator,
	imageTaggingService pipeline.ImageTaggingService,
	promotionPolicyService read.ArtifactPromotionDataReadService,
	requestApprovalUserdataRepo pipelineConfig.RequestApprovalUserdataRepository,
	workflowDagExecutor dag.WorkflowDagExecutor,
	promotionPolicyCUDService PolicyCUDService,
	pipelineStageService pipeline.PipelineStageService,
	environmentService cluster.EnvironmentService,
	resourceFilterEvaluationAuditService resourceFilter.FilterEvaluationAuditService,
) *ApprovalRequestServiceImpl {

	artifactApprovalService := &ApprovalRequestServiceImpl{
		artifactPromotionApprovalRequestRepository: ArtifactPromotionApprovalRequestRepository,
		logger:                               logger,
		ciPipelineRepository:                 CiPipelineRepository,
		pipelineRepository:                   pipelineRepository,
		userService:                          userService,
		ciArtifactRepository:                 ciArtifactRepository,
		appWorkflowRepository:                appWorkflowRepository,
		cdWorkflowRepository:                 cdWorkflowRepository,
		resourceFilterConditionsEvaluator:    resourceFilterConditionsEvaluator,
		imageTaggingService:                  imageTaggingService,
		promotionPolicyDataReadService:       promotionPolicyService,
		requestApprovalUserdataRepo:          requestApprovalUserdataRepo,
		workflowDagExecutor:                  workflowDagExecutor,
		pipelineStageService:                 pipelineStageService,
		environmentService:                   environmentService,
		resourceFilterEvaluationAuditService: resourceFilterEvaluationAuditService,
	}

	// register hooks
	promotionPolicyCUDService.AddDeleteHook(artifactApprovalService.onPolicyDelete)
	promotionPolicyCUDService.AddUpdateHook(artifactApprovalService.onPolicyUpdate)
	return artifactApprovalService
}

func (impl *ApprovalRequestServiceImpl) getSourceInfoAndPipelineIds(workflowId int) (*bean.SourceMetaData, []int, error) {
	allAppWorkflowMappings, err := impl.appWorkflowRepository.FindWFAllMappingByWorkflowId(workflowId)
	if err != nil {
		impl.logger.Errorw("error in finding the app workflow mappings using appWorkflowId", "workflowId", workflowId, "err", err)
		return nil, nil, err
	}

	sourcePipelineMapping := &appWorkflow.AppWorkflowMapping{}
	pipelineIds := make([]int, 0, len(allAppWorkflowMappings))
	for _, mapping := range allAppWorkflowMappings {
		if mapping.Type == appWorkflow.CDPIPELINE {
			pipelineIds = append(pipelineIds, mapping.ComponentId)
		}
		if mapping.ParentId == 0 {
			sourcePipelineMapping = mapping
		}
	}

	sourceId := sourcePipelineMapping.ComponentId
	var sourceName string
	var sourceType constants.SourceTypeStr
	if sourcePipelineMapping.Type == appWorkflow.CIPIPELINE {
		ciPipeline, err := impl.ciPipelineRepository.FindById(sourceId)
		if err != nil {
			impl.logger.Errorw("error in fetching ci pipeline by id", "ciPipelineId", sourceId, "err", err)
			return nil, nil, err
		}
		sourceName = ciPipeline.Name
		sourceType = constants.SOURCE_TYPE_CI
	} else if sourcePipelineMapping.Type == appWorkflow.WEBHOOK {
		sourceType = constants.SOURCE_TYPE_WEBHOOK
	}

	// set source metadata
	sourceInfo := &bean.SourceMetaData{}
	sourceInfo = sourceInfo.WithName(sourceName).WithType(sourceType).WithId(sourceId)
	return sourceInfo, pipelineIds, nil
}

// todo: can move to appworkflow mapping service
func (impl *ApprovalRequestServiceImpl) fetchEnvMetaDataListingRequestMetadata(token string, workflowId int, artifactId int, rbacChecker func(token string, appName string, envNames []string) map[string]bool) (*bean.RequestMetaData, error) {

	sourceInfo, pipelineIds, err := impl.getSourceInfoAndPipelineIds(workflowId)
	if err != nil {
		impl.logger.Errorw("error in finding source info and pipelinesIds", "sourceInfo", sourceInfo, "pipelineIds", pipelineIds, "err", err)
		return nil, err
	}
	pipelines, err := impl.pipelineRepository.FindByIdsIn(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error in finding pipelines", "pipelineIds", pipelineIds, "err", err)
		return nil, err
	}
	environmentNames := make([]string, 0, len(pipelines))
	appName := ""
	appId := 0
	environments := make([]*cluster.EnvironmentBean, 0, len(pipelines))
	for _, pipeline := range pipelines {
		environmentNames = append(environmentNames, pipeline.Environment.Name)
		environment := &cluster.EnvironmentBean{}
		environment.AdaptFromEnvironment(&pipeline.Environment)
		environments = append(environments, environment)
		appName = pipeline.App.AppName
		appId = pipeline.AppId
	}
	authorizedEnvironments := rbacChecker(token, appName, environmentNames)
	cdPipelines := make([]*pipelineConfig.Pipeline, 0, len(pipelines))
	for _, pipeline := range pipelines {
		if authorizedEnvironments[pipeline.Environment.Name] {
			cdPipelines = append(cdPipelines, pipeline)
		}
	}

	requestMetaData := &bean.RequestMetaData{}
	requestMetaData = requestMetaData.WithAppId(appId)
	requestMetaData.SetSourceMetaData(sourceInfo)
	requestMetaData.SetActiveEnvironments(environmentNames, authorizedEnvironments, environments)
	requestMetaData.SetDestinationPipelineMetaData(cdPipelines)
	if artifactId > 0 {
		ciArtifact, err := impl.ciArtifactRepository.Get(artifactId)
		if err != nil {
			impl.logger.Errorw("error in finding the artifact using id", "artifactId", artifactId, "err", err)
			errorResp := util.NewApiError().WithHttpStatusCode(http.StatusInternalServerError).WithInternalMessage(fmt.Sprintf("error in finding artifact , err : %s", err.Error())).WithUserMessage("error in finding artifact")
			if errors.Is(err, pg.ErrNoRows) {
				errorResp = errorResp.WithHttpStatusCode(http.StatusConflict).WithUserMessage("artifact not found")
			}
			return nil, errorResp
		}
		requestMetaData = requestMetaData.WithCiArtifact(ciArtifact)
	}

	return requestMetaData, nil
}

func (impl *ApprovalRequestServiceImpl) FetchWorkflowPromoteNodeList(token string, workflowId int, artifactId int, rbacChecker func(token string, appName string, envNames []string) map[string]bool) (*bean.EnvironmentListingResponse, error) {
	metadata, err := impl.fetchEnvMetaDataListingRequestMetadata(token, workflowId, artifactId, rbacChecker)
	if err != nil {
		impl.logger.Errorw("error in fetching envMetaDataListing request metadata", "token", token, "workflowId", workflowId, "artifactId", artifactId, "err", err)
		return nil, err
	}
	envMap := metadata.GetActiveEnvironmentsMap()
	result := &bean.EnvironmentListingResponse{}
	result.CiSource = metadata.GetSourceMetaData().GetCiSourceMeta()
	policiesMap, err := impl.promotionPolicyDataReadService.GetPromotionPolicyByAppAndEnvIds(metadata.GetAppId(), metadata.GetActiveAuthorisedEnvIds())
	if err != nil {
		impl.logger.Errorw("error in getting the policies", "appId", metadata.GetAppId(), "envIds", metadata.GetActiveAuthorisedEnvIds(), "err", err)
		return nil, err
	}
	if artifactId != 0 {
		responses, err := impl.evaluatePoliciesOnArtifact(metadata, policiesMap)
		if err != nil {
			impl.logger.Errorw("error in evaluating policies on an ciArtifact", "ciArtifactId", artifactId, "policiesMap", policiesMap, "authorizedEnvironments", metadata.GetActiveAuthorisedEnvIds(), "err", err)
			return nil, err
		}
		result.Environments = responses
		return result, nil
	}

	responseMap := getDefaultEnvironmentPromotionMetaDataResponseMap(metadata)
	for envName, resp := range responseMap {
		resp.IsVirtualEnvironment = envMap[envName].IsVirtualEnvironment
		responseMap[envName] = resp
	}

	for envName, policy := range policiesMap {
		responseMap[envName] = bean.EnvironmentPromotionMetaData{
			PromotionPossible:          true,
			Name:                       envName,
			ApprovalCount:              policy.ApprovalMetaData.ApprovalCount,
			IsVirtualEnvironment:       envMap[envName].IsVirtualEnvironment,
			PromotionValidationMessage: "",
			PromotionValidationState:   constants.EMPTY,
		}
	}

	responses := make([]bean.EnvironmentPromotionMetaData, 0, len(responseMap))
	for _, envResponse := range responseMap {
		responses = append(responses, envResponse)
	}

	result.Environments = responses
	return result, nil
}

// todo: move this method away from this service to param Extractor service
func (impl *ApprovalRequestServiceImpl) computeFilterParams(ciArtifact *repository2.CiArtifact) ([]resourceFilter.ExpressionParam, error) {
	var ciMaterials []repository2.CiMaterialInfo
	err := json.Unmarshal([]byte(ciArtifact.MaterialInfo), &ciMaterials)
	if err != nil {
		impl.logger.Errorw("error in parsing ci artifact material info")
		return nil, err
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
	params := resourceFilter.GetParamsFromArtifact(ciArtifact.Image, releaseTags, ciMaterials)
	return params, nil
}

func (impl *ApprovalRequestServiceImpl) evaluatePoliciesOnArtifact(metadata *bean.RequestMetaData, policiesMap map[string]*bean.PromotionPolicy) ([]bean.EnvironmentPromotionMetaData, error) {
	params, err := impl.computeFilterParams(metadata.GetCiArtifact())
	if err != nil {
		impl.logger.Errorw("error in finding the required CEL expression parameters for using ciArtifact", "err", err)
		return nil, err
	}
	envMap := metadata.GetActiveEnvironmentsMap()
	responseMap := getDefaultEnvironmentPromotionMetaDataResponseMap(metadata)
	for envName, resp := range responseMap {
		if env, ok := envMap[envName]; ok {
			resp.PromotionValidationState = constants.POLICY_NOT_CONFIGURED
			resp.PromotionValidationMessage = string(constants.POLICY_NOT_CONFIGURED)
			resp.IsVirtualEnvironment = env.IsVirtualEnvironment
			responseMap[envName] = resp
		}
	}

	// can be concurrent,
	// todo: simplify the below loop
	for envName, policy := range policiesMap {
		evaluationResult, err := impl.resourceFilterConditionsEvaluator.EvaluateFilter(policy.Conditions, resourceFilter.ExpressionMetadata{Params: params})
		if err != nil {
			impl.logger.Errorw("evaluation failed with error", "policyConditions", policy.Conditions, "envName", envName, policy.Conditions, "params", params, "err", err)
			responseMap[envName] = bean.EnvironmentPromotionMetaData{
				Name:                     envName,
				ApprovalCount:            policy.ApprovalMetaData.ApprovalCount,
				PromotionPossible:        false,
				PromotionValidationState: constants.POLICY_EVALUATION_ERRORED,
			}
			continue
		}
		envResp := responseMap[envName]
		envResp.ApprovalCount = policy.ApprovalMetaData.ApprovalCount
		envResp.PromotionValidationState = constants.EMPTY
		envResp.PromotionValidationMessage = ""
		envResp.PromotionPossible = evaluationResult
		// checks on metadata not needed as this is just an evaluation flow (kinda validation)
		if !evaluationResult {
			envResp.PromotionValidationMessage = string(constants.BLOCKED_BY_POLICY)
			envResp.PromotionValidationState = constants.BLOCKED_BY_POLICY
		}
		responseMap[envName] = envResp
	}
	result := make([]bean.EnvironmentPromotionMetaData, 0, len(responseMap))
	for _, envResponse := range responseMap {
		result = append(result, envResponse)
	}
	return result, nil
}

func (impl *ApprovalRequestServiceImpl) approveArtifactPromotion(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) ([]bean.EnvironmentPromotionMetaData, error) {
	// get request and check if it is promoted already.
	// attempt approving this by creating new resource_approval_user_data, if unique constraint error ,current user already did something.
	// attempt success , then get the approval count and check no of approvals got
	//  promote if approvalCount > approvals received
	metadata, err := impl.constructPromotionMetaData(request, authorizedEnvironments)
	if err != nil {
		impl.logger.Errorw("error in getting metadata for the request", "request", request, "err", err)
		return nil, err
	}
	responseMap := getDefaultEnvironmentPromotionMetaDataResponseMap(metadata)

	promotionRequests, err := impl.artifactPromotionApprovalRequestRepository.FindByDestinationPipelineIds(metadata.GetActiveAuthorisedPipelineIds())
	if err != nil {
		impl.logger.Errorw("error in getting artifact promotion request object by id", "promotionRequestId", request.PromotionRequestId, "err", err)
		if errors.Is(err, pg.ErrNoRows) {
			return nil, util.NewApiError().WithHttpStatusCode(http.StatusConflict).WithUserMessage(constants.ArtifactPromotionRequestNotFoundErr).WithInternalMessage(constants.ArtifactPromotionRequestNotFoundErr)
		}
		return nil, err
	}

	// policies fetched form above policy ids
	policies, err := impl.promotionPolicyDataReadService.GetPromotionPolicyByAppAndEnvIds(request.AppId, metadata.GetActiveAuthorisedPipelineEnvIds())
	if err != nil {
		impl.logger.Errorw("error in finding the promotionPolicy by appId and envId names", "appid", request.AppId, "envIds", metadata.GetActiveAuthorisedPipelineEnvIds(), "err", err)
		return nil, err
	}

	// map the policies for O(1) access
	policyIdMap := make(map[int]*bean.PromotionPolicy)
	for _, policy := range policies {
		policyIdMap[policy.Id] = policy
	}

	environmentResponses, err := impl.initiateApprovalProcess(request.UserId, metadata, promotionRequests, responseMap, policyIdMap)
	if err != nil {
		impl.logger.Errorw("error in finding approving the artifact promotion requests", "promotionRequests", promotionRequests, "err", err)
		return nil, err
	}
	return environmentResponses, nil
}

func (impl *ApprovalRequestServiceImpl) approveRequests(userId int32, metadata *bean.RequestMetaData, validRequestIds []int, policyIdMap map[int]*bean.PromotionPolicy, promotionRequests []*repository.ArtifactPromotionApprovalRequest, responses map[string]bean.EnvironmentPromotionMetaData) map[string]bean.EnvironmentPromotionMetaData {
	validRequestsMap := make(map[int]bool)
	for _, requestId := range validRequestIds {
		validRequestsMap[requestId] = true
	}

	pipelineIdVsEnvMap := metadata.GetActiveAuthorisedPipelineIdEnvMap()
	for _, promotionRequest := range promotionRequests {
		// skip the invalid requests
		if ok := validRequestsMap[promotionRequest.Id]; !ok {
			continue
		}
		resp := responses[pipelineIdVsEnvMap[promotionRequest.DestinationPipelineId]]

		if !policyIdMap[promotionRequest.PolicyId].CanApprove(promotionRequest.CreatedBy, 1, userId) {
			resp.PromotionValidationState = constants.BLOCKED_BY_POLICY
			resp.PromotionValidationMessage = string(constants.BLOCKED_BY_POLICY)
		}
		promotionRequestApprovedUserData := &pipelineConfig.RequestApprovalUserData{
			ApprovalRequestId: promotionRequest.Id,
			RequestType:       repository2.ARTIFACT_PROMOTION_APPROVAL,
			UserId:            userId,
			UserResponse:      pipelineConfig.APPROVED,
		}
		// have to do this in loop as we have to ensure partial approval even in case of partial failure
		err := impl.requestApprovalUserdataRepo.SaveRequestApprovalUserData(promotionRequestApprovedUserData)
		if err != nil {
			impl.logger.Errorw("error in saving promotion approval user data", "promotionRequestId", promotionRequest.Id, "err", err)
			if strings.Contains(err.Error(), string(pipelineConfig.UNIQUE_USER_REQUEST_ACTION)) {
				resp.PromotionValidationState = constants.ALREADY_APPROVED
				resp.PromotionValidationMessage = string(constants.ALREADY_APPROVED)
			} else {
				resp.PromotionValidationState = constants.ERRORED_APPROVAL
				resp.PromotionValidationMessage = string(constants.ERRORED_APPROVAL)
			}
			continue
		}

		resp.PromotionValidationState = constants.APPROVED
		resp.PromotionValidationMessage = string(constants.APPROVED)
	}
	return responses
}

func (impl *ApprovalRequestServiceImpl) initiateApprovalProcess(userId int32, metadata *bean.RequestMetaData, promotionRequests []*repository.ArtifactPromotionApprovalRequest, responses map[string]bean.EnvironmentPromotionMetaData, policyIdMap map[int]*bean.PromotionPolicy) ([]bean.EnvironmentPromotionMetaData, error) {

	pipelineIdVsEnvMap := metadata.GetActiveAuthorisedPipelineIdEnvMap()
	staleRequestIds, validRequestIds, responses := impl.filterValidAndStaleRequests(promotionRequests, responses, pipelineIdVsEnvMap, policyIdMap)

	tx, err := impl.artifactPromotionApprovalRequestRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting the transaction", "promotionRequests", promotionRequests, "err", err)
		return nil, err
	}
	defer impl.artifactPromotionApprovalRequestRepository.RollbackTx(tx)
	responses = impl.approveRequests(userId, metadata, validRequestIds, policyIdMap, promotionRequests, responses)
	if len(staleRequestIds) > 0 {
		// attempt deleting the stale requests in bulk
		err = impl.artifactPromotionApprovalRequestRepository.MarkStaleByIds(tx, staleRequestIds)
		if err != nil {
			impl.logger.Errorw("error in deleting the request raised using a deleted promotion policy (stale requests)", "staleRequestIds", staleRequestIds, "err", err)
			// not returning by choice, don't interrupt the user flow
		}
	}

	// fetch all the approved users data for the valid requestIds
	approvedUsersData, err := impl.requestApprovalUserdataRepo.FetchApprovalDataForRequests(validRequestIds, repository2.ARTIFACT_PROMOTION_APPROVAL)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in finding the approved users data for a artifact promotion request", "promotionRequestIds", validRequestIds, "err", err)
		return nil, err
	}

	// club the approved users(we just need count for now) per requestId
	promotionRequestIdVsApprovedUserCount := make(map[int]int)
	for _, _approvedUsersData := range approvedUsersData {
		count := promotionRequestIdVsApprovedUserCount[_approvedUsersData.ApprovalRequestId]
		promotionRequestIdVsApprovedUserCount[_approvedUsersData.ApprovalRequestId] = count + 1
	}

	// filter out promotable requests.
	// we will promote if the current number approvals got for any request exceeds the current configured no of approvals in the policy
	promotableRequestIds := make([]int, 0, len(validRequestIds))
	for _, promotionRequest := range promotionRequests {
		approvalCount := promotionRequestIdVsApprovedUserCount[promotionRequest.Id]
		if policyIdMap[promotionRequest.PolicyId].CanBePromoted(approvalCount) {
			promotableRequestIds = append(promotableRequestIds, promotionRequest.Id)
		}
	}

	// promote the promotableRequestIds
	err = impl.artifactPromotionApprovalRequestRepository.MarkPromoted(tx, promotableRequestIds)
	if err != nil {
		impl.logger.Errorw("error in promoting the approval requests", "promotableRequestIds", promotableRequestIds, "err", err)
		return nil, err
	}

	promotionRequestIdToDaoMap := make(map[int]*repository.ArtifactPromotionApprovalRequest)
	for _, promotionRequest := range promotionRequests {
		promotionRequestIdToDaoMap[promotionRequest.Id] = promotionRequest
	}

	if len(promotableRequestIds) > 0 {
		err = impl.handleArtifactPromotionSuccess(promotableRequestIds, promotionRequestIdToDaoMap, metadata.GetActiveAuthorisedPipelineDaoMap())
		if err != nil {
			impl.logger.Errorw("error in handling the successful artifact promotion event for promotedRequests", "promotableRequestIds", promotableRequestIds, "err", err)
			return nil, err
		}
	}

	err = impl.artifactPromotionApprovalRequestRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing the transaction", "promotableRequestIds", promotableRequestIds, "err", err)
		return nil, err
	}

	result := make([]bean.EnvironmentPromotionMetaData, 0, len(responses))
	for _, resp := range responses {
		result = append(result, resp)
	}
	return result, nil
}

func (impl *ApprovalRequestServiceImpl) filterValidAndStaleRequests(promotionRequests []*repository.ArtifactPromotionApprovalRequest, responses map[string]bean.EnvironmentPromotionMetaData, pipelineIdVsEnvMap map[int]string, policyIdMap map[int]*bean.PromotionPolicy) ([]int, []int, map[string]bean.EnvironmentPromotionMetaData) {
	staleRequestIds := make([]int, 0)
	validRequestIds := make([]int, 0)
	for _, promotionRequest := range promotionRequests {
		resp := responses[pipelineIdVsEnvMap[promotionRequest.DestinationPipelineId]]
		_, ok := policyIdMap[promotionRequest.PolicyId]
		if ok {
			validRequestIds = append(validRequestIds, promotionRequest.Id)
		}
		if !ok {
			// policy is not found in the map, and the request is still in awaiting state.
			// although the policy is no longer governing the current pipeline
			// this is a stale case.
			// mark it stale
			staleRequestIds = append(staleRequestIds, promotionRequest.Id)

			// also set the response messages
			resp.PromotionPossible = false
			resp.PromotionValidationMessage = "request is no longer valid as the policy is no longer governing this pipeline, state: stale"
			resp.PromotionValidationState = constants.PromotionValidationState(resp.PromotionValidationMessage)
		} else if promotionRequest.Status != constants.AWAITING_APPROVAL {
			resp.PromotionValidationMessage = fmt.Sprintf("artifact is in %s state", promotionRequest.Status.Status())
			resp.PromotionValidationState = constants.PromotionValidationState(resp.PromotionValidationMessage)
		}
		responses[pipelineIdVsEnvMap[promotionRequest.DestinationPipelineId]] = resp
	}
	return staleRequestIds, validRequestIds, responses
}

func (impl *ApprovalRequestServiceImpl) handleArtifactPromotionSuccess(promotableRequestIds []int, promotionRequestIdToDaoMap map[int]*repository.ArtifactPromotionApprovalRequest, pipelineIdToDaoMap map[int]*pipelineConfig.Pipeline) error {
	promotedCiArtifactIds := make([]int, 0)
	for _, id := range promotableRequestIds {
		promotableRequest := promotionRequestIdToDaoMap[id]
		promotedCiArtifactIds = append(promotedCiArtifactIds, promotableRequest.ArtifactId)
	}

	artifacts, err := impl.ciArtifactRepository.GetByIds(promotedCiArtifactIds)
	if err != nil {
		impl.logger.Errorw("error in fetching the artifacts by ids", "artifactIds", promotedCiArtifactIds, "err", err)
		return err
	}

	artifactsMap := make(map[int]*repository2.CiArtifact)
	for _, artifact := range artifacts {
		artifactsMap[artifact.Id] = artifact
	}
	for _, id := range promotableRequestIds {
		promotableRequest := promotionRequestIdToDaoMap[id]
		pipelineDao := pipelineIdToDaoMap[promotableRequest.DestinationPipelineId]
		triggerRequest := bean2.TriggerRequest{
			CdWf:        nil,
			Pipeline:    pipelineDao,
			Artifact:    artifactsMap[promotableRequest.ArtifactId],
			TriggeredBy: 1,
			TriggerContext: bean2.TriggerContext{
				Context: context.Background(),
			},
		}
		impl.workflowDagExecutor.HandleArtifactPromotionEvent(triggerRequest)
	}
	return nil
}

func (impl *ApprovalRequestServiceImpl) constructPromotionMetaData(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) (*bean.RequestMetaData, error) {
	requestMetaData := &bean.RequestMetaData{}

	// set source metadata
	sourceMeta, err := impl.fetchSourceMeta(request.SourceName, request.SourceType, request.AppId, request.WorkflowId)
	if err != nil {
		impl.logger.Errorw("error in validating the request", "request", request, "err", err)
		return nil, err
	}
	requestMetaData.SetSourceMetaData(sourceMeta)

	// set environment metadata
	environments, err := impl.environmentService.FindByNames(request.EnvironmentNames)
	if err != nil {
		impl.logger.Errorw("error in fetching the environment details", "environmentNames", request.EnvironmentNames, "err", err)
		return nil, err
	}
	requestMetaData.SetActiveEnvironments(request.EnvironmentNames, authorizedEnvironments, environments)

	// set destination pipelines metadata
	cdPipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvNames(request.AppId, requestMetaData.GetActiveAuthorisedEnvNames())
	if err != nil {
		impl.logger.Errorw("error in finding the cd pipelines using appID and environment names", "appId", request.AppId, "envNames", requestMetaData.GetActiveAuthorisedEnvNames(), "err", err)
		return nil, err
	}
	requestMetaData.SetDestinationPipelineMetaData(cdPipelines)
	return requestMetaData, nil
}

func (impl *ApprovalRequestServiceImpl) HandleArtifactPromotionRequest(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) ([]bean.EnvironmentPromotionMetaData, error) {

	switch request.Action {

	case constants.ACTION_PROMOTE:
		return impl.promoteArtifact(request, authorizedEnvironments)
	case constants.ACTION_APPROVE:
		return impl.approveArtifactPromotion(request, authorizedEnvironments)
	case constants.ACTION_CANCEL:

		_, err := impl.cancelPromotionApprovalRequest(request)
		if err != nil {
			impl.logger.Errorw("error in canceling artifact promotion approval request", "promotionRequestId", request.PromotionRequestId, "err", err)
			return nil, err
		}
		return nil, nil

	}
	return nil, errors.New("unknown action")
}

func (impl *ApprovalRequestServiceImpl) fetchSourceMeta(sourceName string, sourceType constants.SourceTypeStr, appId int, workflowId int) (*bean.SourceMetaData, error) {
	sourceInfo := &bean.SourceMetaData{}
	sourceInfo = sourceInfo.WithName(sourceName).WithType(sourceType)
	if sourceType == constants.SOURCE_TYPE_CI || sourceType == constants.SOURCE_TYPE_WEBHOOK {
		appWorkflowMapping, err := impl.appWorkflowRepository.FindByWorkflowIdAndCiSource(workflowId)
		if err != nil {
			impl.logger.Errorw("error in getting the workflow mapping of ci-source/webhook using workflow id", "workflowId", workflowId, "err", err)
			if errors.Is(err, pg.ErrNoRows) {
				return nil, util.NewApiError().WithHttpStatusCode(http.StatusConflict).WithUserMessage("given workflow not found for the provided source").WithInternalMessage("given workflow not found for the provided source")
			}
			return nil, err
		}
		sourceInfo = sourceInfo.WithId(appWorkflowMapping.ComponentId).WithSourceWorkflowId(appWorkflowMapping.AppWorkflowId)
	} else {
		// source type will be cd and source name will be envName.
		// get pipeline using appId and env name and get the workflowMapping
		pipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvNames(appId, []string{sourceName})
		if err != nil {
			impl.logger.Errorw("error in getting the pipelines using appId and source environment name ", "workflowId", workflowId, "appId", appId, "source", sourceName, "err", err)
			return nil, err
		}
		if len(pipelines) == 0 {
			return nil, util.NewApiError().WithHttpStatusCode(http.StatusConflict).WithUserMessage("source pipeline with given environment not found in the workflow").WithInternalMessage("source pipeline with given environment not found in workflow")
		}

		pipeline := pipelines[0]
		appWorkflowMapping, err := impl.appWorkflowRepository.FindWFMappingByComponent(appWorkflow.CDPIPELINE, pipeline.Id)
		if err != nil {
			impl.logger.Errorw("error in getting the app workflow mapping using workflow id and cd component id", "workflowId", workflowId, "appId", appId, "pipelineId", pipeline.Id, "err", err)
			if errors.Is(err, pg.ErrNoRows) {
				return nil, util.NewApiError().WithHttpStatusCode(http.StatusConflict).WithUserMessage("source pipeline not found in the given workflow").WithInternalMessage("source pipeline not found in the given workflow")
			}
			return nil, err
		}
		sourceInfo = sourceInfo.WithId(pipeline.Id).WithSourceWorkflowId(appWorkflowMapping.AppWorkflowId).WithCdPipeline(pipeline)
	}
	return sourceInfo, nil
}

// todo: naming can be better
// todo: gireesh, make this method on RequestMetaData struct
func getDefaultEnvironmentPromotionMetaDataResponseMap(metadata *bean.RequestMetaData) map[string]bean.EnvironmentPromotionMetaData {
	response := make(map[string]bean.EnvironmentPromotionMetaData)
	for _, env := range metadata.GetUserGivenEnvNames() {
		envResponse := bean.EnvironmentPromotionMetaData{
			Name:                       env,
			PromotionValidationState:   constants.PIPELINE_NOT_FOUND,
			PromotionValidationMessage: string(constants.PIPELINE_NOT_FOUND),
		}
		if !metadata.GetAuthorisedEnvMap()[env] {
			envResponse.PromotionValidationState = constants.NO_PERMISSION
			envResponse.PromotionValidationMessage = string(constants.NO_PERMISSION)
		}
		response[env] = envResponse
	}
	for _, pipelineId := range metadata.GetActiveAuthorisedPipelineIds() {
		envName := metadata.GetActiveAuthorisedPipelineIdEnvMap()[pipelineId]
		resp := response[envName]
		resp.PromotionValidationState = constants.EMPTY
		resp.PromotionValidationMessage = ""
		response[envName] = resp
	}
	return response
}

func (impl *ApprovalRequestServiceImpl) validatePromotion(requestedWorkflowId int, ciArtifact *repository2.CiArtifact, metadata *bean.RequestMetaData) (map[string]bean.EnvironmentPromotionMetaData, error) {
	if requestedWorkflowId != metadata.GetWorkflowId() {
		// handle throw api error with conflict status code
		return nil, util.NewApiError().WithHttpStatusCode(http.StatusConflict).WithUserMessage("provided source is not linked to the given workflow").WithInternalMessage("provided source is not linked to the given workflow")
	}

	allAppWorkflowMappings, err := impl.appWorkflowRepository.FindWFAllMappingByWorkflowId(metadata.GetWorkflowId())
	if err != nil {
		impl.logger.Errorw("error in finding the app workflow mappings", "err", err)
		return nil, err
	}
	if metadata.GetSourceType() == constants.SOURCE_TYPE_CD {

		sourcePipeline := metadata.GetSourceCdPipeline()
		deployed, err := impl.checkIfDeployedAtSource(ciArtifact.Id, sourcePipeline)
		if err != nil {
			impl.logger.Errorw("error in checking if artifact is available for promotion at source pipeline", "ciArtifactId", ciArtifact.Id, "sourcePipelineId", sourcePipeline.Id, "err", err)
			return nil, err
		}

		if !deployed {
			errMsg := fmt.Sprintf("artifact is not deployed on the source environment %s", metadata.GetSourceName())
			return nil, util.NewApiError().WithHttpStatusCode(http.StatusConflict).WithUserMessage(errMsg).WithInternalMessage(errMsg)
		}

		tree := make(map[int][]int)
		for _, appWorkflowMapping := range allAppWorkflowMappings {
			// create the tree from the DAG excluding the ci source
			if appWorkflowMapping.Type == appWorkflow.CDPIPELINE && appWorkflowMapping.ParentType == appWorkflow.CDPIPELINE {
				tree[appWorkflowMapping.ParentId] = append(tree[appWorkflowMapping.ParentId], appWorkflowMapping.ComponentId)
			}
		}

		// if sourcePipelineId is 0, then the source pipeline given by user is not found in the workflow.
		if metadata.GetSourcePipelineId() == 0 {
			errMsg := fmt.Sprintf("no pipeline found against given source environment %s", metadata.GetSourceName())
			return nil, util.NewApiError().WithHttpStatusCode(http.StatusBadRequest).WithUserMessage(errMsg).WithInternalMessage(errMsg)
		}

		responseMap := make(map[string]bean.EnvironmentPromotionMetaData)
		for _, pipelineId := range metadata.GetActiveAuthorisedPipelineIds() {
			if !util.IsAncestor(tree, metadata.GetSourcePipelineId(), pipelineId) {
				envName := metadata.GetActiveAuthorisedPipelineIdEnvMap()[pipelineId]
				resp := bean.EnvironmentPromotionMetaData{
					Name:                       envName,
					PromotionValidationState:   constants.SOURCE_AND_DESTINATION_PIPELINE_MISMATCH,
					PromotionValidationMessage: string(constants.SOURCE_AND_DESTINATION_PIPELINE_MISMATCH),
				}
				cdPipeline := metadata.GetActiveAuthorisedPipelineDaoMap()[pipelineId]
				if cdPipeline != nil {
					resp.IsVirtualEnvironment = cdPipeline.Environment.IsVirtualEnvironment
				}
				responseMap[envName] = resp
			}
		}
		return responseMap, nil
	}
	return nil, nil
}

func (impl *ApprovalRequestServiceImpl) promoteArtifact(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) ([]bean.EnvironmentPromotionMetaData, error) {
	// 	step1: validate if artifact is deployed/created at the source pipeline.
	//      step1: if source is cd , check if this artifact is deployed on these environments
	//  step2: check if destination pipeline is topologically downwards from the source pipeline and also source and destination are on the same subtree.
	// 	step3: check if promotion request for this artifact on this destination pipeline has already been raised.
	//  step4: check if this artifact on this destination pipeline has already been promoted
	//  step5: raise request.

	// fetch artifact
	metadata, err := impl.constructPromotionMetaData(request, authorizedEnvironments)
	if err != nil {
		impl.logger.Errorw("error in getting metadata for the request", "request", request, "err", err)
		return nil, err
	}
	responseMap := getDefaultEnvironmentPromotionMetaDataResponseMap(metadata)
	ciArtifact, err := impl.ciArtifactRepository.Get(request.ArtifactId)
	if err != nil {
		impl.logger.Errorw("error in finding the artifact using id", "artifactId", request.ArtifactId, "err", err)
		// todo: create new error type and create a construct on that which will accept user msg and err and overridable status code
		errRes := util.NewApiError().WithHttpStatusCode(http.StatusInternalServerError).WithInternalMessage(fmt.Sprintf("error in finding artifact , err : %s", err.Error())).WithUserMessage("error in finding artifact")
		if errors.Is(err, pg.ErrNoRows) {
			errRes = errRes.WithHttpStatusCode(http.StatusConflict).WithUserMessage("artifact not found")
		}
		return nil, errRes
	}

	validationResponseMap, err := impl.validatePromotion(request.WorkflowId, ciArtifact, metadata)
	if err != nil {
		impl.logger.Errorw("error in validating the workflowPromotion request", "metadata", metadata, "err", err)
		return nil, err
	}

	for envName, resp := range validationResponseMap {
		responseMap[envName] = resp
	}

	policiesMap, err := impl.promotionPolicyDataReadService.GetPromotionPolicyByAppAndEnvIds(request.AppId, metadata.GetActiveAuthorisedEnvIds())
	if err != nil {
		impl.logger.Errorw("error in getting policies for some environments in an app", "appName", request.AppName, "envNames", metadata.GetActiveAuthorisedEnvNames(), "err", err)
		return nil, err
	}

	promotableEnvs := make([]string, 0)
	for _, resp := range responseMap {
		if resp.PromotionValidationState == constants.EMPTY {
			promotableEnvs = append(promotableEnvs, resp.Name)
		}
	}
	promoteResponseMap := impl.raisePromoteRequestHelper(request.UserId, policiesMap, ciArtifact, metadata)
	for envName, resp := range promoteResponseMap {
		responseMap[envName] = resp
	}
	envResponses := make([]bean.EnvironmentPromotionMetaData, 0, len(responseMap))
	for _, resp := range responseMap {
		envResponses = append(envResponses, resp)
	}
	return envResponses, nil
}

func (impl *ApprovalRequestServiceImpl) raisePromoteRequestHelper(userId int32, policiesMap map[string]*bean.PromotionPolicy, ciArtifact *repository2.CiArtifact, metadata *bean.RequestMetaData) map[string]bean.EnvironmentPromotionMetaData {
	responseMap := make(map[string]bean.EnvironmentPromotionMetaData)
	for _, pipelineId := range metadata.GetActiveAuthorisedPipelineIds() {

		pipelineIdVsEnvNameMap := metadata.GetActiveAuthorisedPipelineIdEnvMap()
		pipelineIdToDaoMap := metadata.GetActiveAuthorisedPipelineDaoMap()
		EnvResponse := bean.EnvironmentPromotionMetaData{
			Name:                 pipelineIdVsEnvNameMap[pipelineId],
			IsVirtualEnvironment: metadata.GetActiveAuthorisedPipelineDaoMap()[pipelineId].Environment.IsVirtualEnvironment,
		}

		policy := policiesMap[pipelineIdVsEnvNameMap[pipelineId]]
		if policy == nil {
			EnvResponse.PromotionValidationState = constants.POLICY_NOT_CONFIGURED
			EnvResponse.PromotionValidationMessage = string(constants.POLICY_NOT_CONFIGURED)
		} else if EnvResponse.PromotionValidationState == constants.EMPTY {
			state, msg, err := impl.raisePromoteRequest(userId, policy, ciArtifact, pipelineIdToDaoMap[pipelineId], metadata)
			if err != nil {
				impl.logger.Errorw("error in raising promotion request for the pipeline", "pipelineId", pipelineId, "artifactId", ciArtifact.Id, "err", err)
				EnvResponse.PromotionValidationState = constants.ERRORED
				EnvResponse.PromotionValidationMessage = err.Error()
			}
			EnvResponse.PromotionPossible = true
			EnvResponse.PromotionValidationState = state
			EnvResponse.PromotionValidationMessage = msg
		}
		responseMap[pipelineIdVsEnvNameMap[pipelineId]] = EnvResponse
	}

	return responseMap
}

func (impl *ApprovalRequestServiceImpl) raisePromoteRequest(userId int32, promotionPolicy *bean.PromotionPolicy, ciArtifact *repository2.CiArtifact, cdPipeline *pipelineConfig.Pipeline, metadata *bean.RequestMetaData) (constants.PromotionValidationState, string, error) {
	// todo : handle it in single db call
	requests, err := impl.artifactPromotionApprovalRequestRepository.FindAwaitedRequestByPipelineIdAndArtifactId(cdPipeline.Id, ciArtifact.Id)
	if err != nil {
		impl.logger.Errorw("error in finding the pending promotion request using pipelineId and artifactId", "pipelineId", cdPipeline.Id, "artifactId", ciArtifact.Id)
		return constants.ERRORED, err.Error(), err
	}

	if len(requests) >= 1 {
		return constants.ALREADY_REQUEST_RAISED, string(constants.ALREADY_REQUEST_RAISED), nil
	}

	promotedRequest, err := impl.artifactPromotionApprovalRequestRepository.FindPromotedRequestByPipelineIdAndArtifactId(cdPipeline.Id, ciArtifact.Id)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in finding the promoted request using pipelineId and artifactId", "pipelineId", cdPipeline.Id, "artifactId", ciArtifact.Id)
		return constants.ERRORED, err.Error(), err
	}

	if promotedRequest.Id > 0 {
		return constants.ARTIFACT_ALREADY_PROMOTED, string(constants.ARTIFACT_ALREADY_PROMOTED), nil
	}

	// todo end

	params, err := impl.computeFilterParams(ciArtifact)
	if err != nil {
		impl.logger.Errorw("error in finding the required CEL expression parameters for using ciArtifact", "err", err)
		return constants.POLICY_EVALUATION_ERRORED, string(constants.POLICY_EVALUATION_ERRORED), err
	}

	evaluationResult, err := impl.resourceFilterConditionsEvaluator.EvaluateFilter(promotionPolicy.Conditions, resourceFilter.ExpressionMetadata{Params: params})
	if err != nil {
		impl.logger.Errorw("evaluation failed with error", "policyConditions", promotionPolicy.Conditions, "pipelineId", cdPipeline.Id, promotionPolicy.Conditions, "params", params, "err", err)
		return constants.POLICY_EVALUATION_ERRORED, string(constants.POLICY_EVALUATION_ERRORED), err
	}

	if !evaluationResult {
		return constants.BLOCKED_BY_POLICY, string(constants.BLOCKED_BY_POLICY), nil
	}

	evaluationAuditJsonString, err := evaluationJsonString(evaluationResult, promotionPolicy)
	if err != nil {
		return constants.ERRORED, err.Error(), err
	}

	tx, err := impl.artifactPromotionApprovalRequestRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting the transaction", "evaluationResult", evaluationResult, "promotionPolicy", promotionPolicy, "err", err)
		return constants.ERRORED, err.Error(), err
	}
	defer impl.artifactPromotionApprovalRequestRepository.RollbackTx(tx)

	// save evaluation audit
	evaluationAuditEntry, err := impl.resourceFilterEvaluationAuditService.SaveFilterEvaluationAudit(tx, resourceFilter.Artifact, ciArtifact.Id, cdPipeline.Id, resourceFilter.Pipeline, userId, evaluationAuditJsonString, resourceFilter.ARTIFACT_PROMOTION_POLICY)
	if err != nil {
		impl.logger.Errorw("error in saving policy evaluation audit data", "evaluationAuditEntry", evaluationAuditEntry, "err", err)
		return constants.ERRORED, err.Error(), err
	}
	promotionRequest := &repository.ArtifactPromotionApprovalRequest{
		SourceType:              metadata.GetSourceType().GetSourceType(),
		SourcePipelineId:        metadata.GetSourcePipelineId(),
		DestinationPipelineId:   cdPipeline.Id,
		Status:                  constants.AWAITING_APPROVAL,
		ArtifactId:              ciArtifact.Id,
		PolicyId:                promotionPolicy.Id,
		PolicyEvaluationAuditId: evaluationAuditEntry.Id,
		AuditLog:                sql.NewDefaultAuditLog(userId),
	}

	var status constants.PromotionValidationState
	if promotionPolicy.CanBePromoted(0) {
		promotionRequest.Status = constants.PROMOTED
		status = constants.PROMOTION_SUCCESSFUL
	} else {
		status = constants.SENT_FOR_APPROVAL
	}
	_, err = impl.artifactPromotionApprovalRequestRepository.Create(tx, promotionRequest)
	if err != nil {
		impl.logger.Errorw("error in finding the pending promotion request using pipelineId and artifactId", "pipelineId", cdPipeline.Id, "artifactId", ciArtifact.Id)
		return constants.ERRORED, err.Error(), err
	}

	err = impl.artifactPromotionApprovalRequestRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing the db transaction", "pipelineId", cdPipeline.Id, "artifactId", ciArtifact.Id, "err", err)
		return constants.ERRORED, err.Error(), err
	}
	if promotionRequest.Status == constants.PROMOTED {
		triggerRequest := bean2.TriggerRequest{
			CdWf:        nil,
			Pipeline:    cdPipeline,
			Artifact:    ciArtifact,
			TriggeredBy: 1,
			TriggerContext: bean2.TriggerContext{
				Context: context.Background(),
			},
		}
		// todo: ayush
		impl.workflowDagExecutor.HandleArtifactPromotionEvent(triggerRequest)
	}
	return status, string(status), nil

}

func evaluationJsonString(evaluationResult bool, promotionPolicy *bean.PromotionPolicy) (string, error) {
	evaluationAudit := make(map[string]interface{})
	evaluationAudit["result"] = evaluationResult
	evaluationAudit["policy"] = promotionPolicy
	evaluationAuditJsonBytes, err := json.Marshal(&evaluationAudit)
	if err != nil {
		return "", err
	}
	return string(evaluationAuditJsonBytes), nil
}

func (impl *ApprovalRequestServiceImpl) checkIfDeployedAtSource(ciArtifactId int, pipeline *pipelineConfig.Pipeline) (bool, error) {
	if pipeline == nil {
		return false, errors.New("invalid cd pipeline")
	}
	postStage, err := impl.pipelineStageService.GetCdStageByCdPipelineIdAndStageType(pipeline.Id, repository3.PIPELINE_STAGE_TYPE_POST_CD)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in finding the post-cd existence for the pipeline", "pipelineId", pipeline.Id, "err", err)
		return false, err
	}
	workflowType := bean3.CD_WORKFLOW_TYPE_DEPLOY
	if len(pipeline.PostStageConfig) > 0 || (postStage != nil && postStage.Id > 0) {
		workflowType = bean3.CD_WORKFLOW_TYPE_POST
	}

	deployed, err := impl.cdWorkflowRepository.IsArtifactDeployedOnStage(ciArtifactId, pipeline.Id, workflowType)
	if err != nil {
		impl.logger.Errorw("error in finding if the artifact is successfully deployed on a pipeline", "ciArtifactId", ciArtifactId, "pipelineId", pipeline.Id, "workflowType", workflowType, "err", err)
		return false, err
	}
	return deployed, nil
}

func (impl *ApprovalRequestServiceImpl) cancelPromotionApprovalRequest(request *bean.ArtifactPromotionRequest) (*bean.ArtifactPromotionRequest, error) {
	// todo: accept environment name instead of requestId
	artifactPromotionDao, err := impl.artifactPromotionApprovalRequestRepository.FindById(request.PromotionRequestId)
	if errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("artifact promotion approval request not found for given id", "promotionRequestId", request.PromotionRequestId, "err", err)
		return nil, &util.ApiError{
			HttpStatusCode:  http.StatusNotFound,
			InternalMessage: constants.ArtifactPromotionRequestNotFoundErr,
			UserMessage:     constants.ArtifactPromotionRequestNotFoundErr,
		}
	}
	if err != nil {
		impl.logger.Errorw("error in fetching artifact promotion request by id", "artifactPromotionRequestId", request.PromotionRequestId, "err", err)
		return nil, err
	}

	if artifactPromotionDao.CreatedBy != request.UserId {
		return nil, util.NewApiError().WithHttpStatusCode(http.StatusUnprocessableEntity).WithInternalMessage(constants.UserCannotCancelRequest).WithUserMessage(constants.UserCannotCancelRequest)
	}

	artifactPromotionDao.Status = constants.CANCELED
	artifactPromotionDao.UpdatedOn = time.Now()
	artifactPromotionDao.UpdatedBy = request.UserId
	_, err = impl.artifactPromotionApprovalRequestRepository.Update(artifactPromotionDao)
	if err != nil {
		impl.logger.Errorw("error in updating artifact promotion approval request", "artifactPromotionRequestId", request.PromotionRequestId, "err", err)
		return nil, err
	}
	return nil, err
}

func (impl *ApprovalRequestServiceImpl) GetPromotionRequestById(promotionRequestId int) (*bean.ArtifactPromotionApprovalResponse, error) {
	promotionRequest, err := impl.artifactPromotionApprovalRequestRepository.FindById(promotionRequestId)
	if err != nil {
		impl.logger.Errorw("error in getting promotion request by id", "promotionRequestId", promotionRequestId, "err", err)
		return nil, err
	}
	artifactPromotionResponse := &bean.ArtifactPromotionApprovalResponse{
		Id:                      promotionRequest.Id,
		PolicyId:                promotionRequest.PolicyId,
		PolicyEvaluationAuditId: promotionRequest.PolicyEvaluationAuditId,
		ArtifactId:              promotionRequest.ArtifactId,
		SourceType:              promotionRequest.SourceType,
		SourcePipelineId:        promotionRequest.SourcePipelineId,
		DestinationPipelineId:   promotionRequest.DestinationPipelineId,
		Status:                  promotionRequest.Status,
	}
	return artifactPromotionResponse, nil
}

func (impl *ApprovalRequestServiceImpl) FetchApprovalAllowedEnvList(artifactId int, environmentName string, userId int32, token string, promotionApproverAuth func(string, string) bool) ([]bean.EnvironmentApprovalMetadata, error) {

	artifact, err := impl.ciArtifactRepository.Get(artifactId)
	if err != nil {
		impl.logger.Errorw("artifact not found for given id", "artifactId", artifactId, "err", err)
		return nil, util.NewApiError().WithHttpStatusCode(http.StatusUnprocessableEntity).WithUserMessage("artifact not found for given id").WithInternalMessage("artifact not found for given id")
	}

	promotionRequests, err := impl.artifactPromotionApprovalRequestRepository.FindRequestsByArtifactIdAndEnvName(artifactId, environmentName, constants.AWAITING_APPROVAL)
	if err != nil {
		impl.logger.Errorw("error in finding promotion requests in awaiting state for given artifactId ")
		return nil, err
	}

	environmentApprovalMetadata := make([]bean.EnvironmentApprovalMetadata, 0)

	if len(promotionRequests) == 0 {
		return environmentApprovalMetadata, nil
	}

	// TODO: remove lo
	destinationPipelineIds := lo.Map(promotionRequests, func(item *repository.ArtifactPromotionApprovalRequest, index int) int {
		return item.DestinationPipelineId
	})

	pipelines, err := impl.pipelineRepository.FindAppAndEnvironmentAndProjectByPipelineIds(destinationPipelineIds)
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines by ids", "pipelineIds", destinationPipelineIds, "err", err)
		return nil, err
	}

	pipelineIdMap := make(map[int]*pipelineConfig.Pipeline)
	for _, pipelineDao := range pipelines {
		pipelineIdMap[pipelineDao.Id] = pipelineDao
	}

	//envIds := make([]int, len(pipelines))
	//for i, p := range pipelines {
	//	envIds[i] = p.EnvironmentId
	//}
	//
	//policies, err := impl.promotionPolicyDataReadService.GetPromotionPolicyByAppAndEnvIds(pipelines[0].AppId, envIds)
	//if err != nil {
	//	impl.logger.Errorw("error in fetching policies by appId and envIds", "appId", pipelines[0].AppId, "envIds", envIds, "err", err)
	//	return nil, err
	//}

	for _, request := range promotionRequests {

		pipelineDao := pipelineIdMap[request.DestinationPipelineId]

		environmentMetadata := bean.EnvironmentApprovalMetadata{
			Name:            pipelineDao.Environment.Name,
			ApprovalAllowed: true,
			Reasons:         make([]string, 0),
		}
		// TODO: fetch policies in bulk
		policy, err := impl.promotionPolicyDataReadService.GetPromotionPolicyByAppAndEnvId(pipelineDao.AppId, pipelineDao.EnvironmentId)
		if err != nil {
			impl.logger.Errorw("error in fetching promotion policy for given appId and envId", "appId", pipelineDao.AppId, "envId", pipelineDao.EnvironmentId, "err", err)
			return nil, err
		}
		// TODO abstract logic to policyBean
		if !policy.ApprovalMetaData.AllowImageBuilderFromApprove && request.CreatedBy == artifact.CreatedBy {
			environmentMetadata.ApprovalAllowed = false
			// TODO: reason constant
			environmentMetadata.Reasons = append(environmentMetadata.Reasons, "User who has built the image cannot approve promotion request for this environment")
		}

		if !policy.ApprovalMetaData.AllowRequesterFromApprove && request.CreatedBy == userId {
			environmentMetadata.ApprovalAllowed = false
			environmentMetadata.Reasons = append(environmentMetadata.Reasons, "User who has raised the image cannot approve promotion request for this environment")
		}
		// TODO: rbac batch
		rbacObject := fmt.Sprintf("%s/%s/%s", pipelineDao.App.Team.Name, pipelineDao.Environment.EnvironmentIdentifier, pipelineDao.App.AppName)
		if ok := promotionApproverAuth(token, rbacObject); !ok {
			environmentMetadata.ApprovalAllowed = false
			environmentMetadata.Reasons = append(environmentMetadata.Reasons, "user does not have image promoter access for given app and env")
		}

		environmentApprovalMetadata = append(environmentApprovalMetadata, environmentMetadata)
	}
	return environmentApprovalMetadata, nil
}

func (impl *ApprovalRequestServiceImpl) onPolicyDelete(tx *pg.Tx, policyId int) error {
	err := impl.artifactPromotionApprovalRequestRepository.MarkStaleByPolicyId(tx, policyId)
	if err != nil {
		impl.logger.Errorw("error in marking artifact promotion requests stale", "policyId", policyId, "err", err)
	}
	return err
}

func (impl *ApprovalRequestServiceImpl) onPolicyUpdate(tx *pg.Tx, policy *bean.PromotionPolicy) error {
	// get all the requests whose id is policy.id
	existingRequests, err := impl.artifactPromotionApprovalRequestRepository.FindAwaitedRequestByPolicyId(policy.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching the awaiting artifact promotion requests using policy Id", "policyId", policy.Id, "err", err)
		return err
	}
	artifactIds := make([]int, 0, len(existingRequests))
	for _, request := range existingRequests {
		artifactIds = append(artifactIds, request.ArtifactId)
	}
	if len(artifactIds) == 0 {
		impl.logger.Debugw("no awaiting requests found for the policy", "policyId", policy.Id)
		return nil
	}

	artifacts, err := impl.ciArtifactRepository.GetByIds(artifactIds)
	if err != nil {
		impl.logger.Errorw("error in fetching the artifacts by ids", "artifactIds", artifactIds, "err", err)
		return err
	}

	artifactsMap := make(map[int]*repository2.CiArtifact)
	for _, artifact := range artifacts {
		artifactsMap[artifact.Id] = artifact
	}

	// get all the artifacts using request.artifactId
	// re-evaluate the artifacts using the policy

	requestsToBeUpdatedAsStaled, err := impl.reEvaluatePolicyAndUpdateRequests(tx, policy, artifactsMap, existingRequests)
	if err != nil {
		return err
	}

	err = impl.artifactPromotionApprovalRequestRepository.UpdateInBulk(tx, requestsToBeUpdatedAsStaled)
	if err != nil {
		impl.logger.Errorw("error in marking artifact promotion requests stale", "policyId", policy.Id, "err", err)
	}
	return err
}

func (impl *ApprovalRequestServiceImpl) reEvaluatePolicyAndUpdateRequests(tx *pg.Tx, policy *bean.PromotionPolicy, artifactsMap map[int]*repository2.CiArtifact, existingRequests []*repository.ArtifactPromotionApprovalRequest) ([]*repository.ArtifactPromotionApprovalRequest, error) {
	requestsToBeUpdatedAsStaled := make([]*repository.ArtifactPromotionApprovalRequest, 0, len(existingRequests))
	for _, request := range existingRequests {
		artifact := artifactsMap[request.ArtifactId]
		params, err := impl.computeFilterParams(artifact)
		if err != nil {
			continue
		}

		evaluationResult, err := impl.resourceFilterConditionsEvaluator.EvaluateFilter(policy.Conditions, resourceFilter.ExpressionMetadata{Params: params})
		if err != nil {
			impl.logger.Errorw("evaluation failed with error", "policyConditions", policy.Conditions, "pipelineId", request.DestinationPipelineId, "policyConditions", policy.Conditions, "params", params, "err", err)
			continue
		}

		// policy is blocking the request, so need to update these as staled requests
		if !evaluationResult {
			evaluationAuditJsonString, err := evaluationJsonString(evaluationResult, policy)
			if err != nil {
				continue
			}

			// save evaluation audit
			evaluationAuditEntry, err := impl.resourceFilterEvaluationAuditService.SaveFilterEvaluationAudit(tx, resourceFilter.Artifact, request.ArtifactId, request.DestinationPipelineId, resourceFilter.Pipeline, 1, evaluationAuditJsonString, resourceFilter.ARTIFACT_PROMOTION_POLICY)
			if err != nil {
				impl.logger.Errorw("error in saving policy evaluation audit data", "evaluationAuditEntry", evaluationAuditEntry, "err", err)
				continue
			}
			request.UpdatedOn = time.Now()
			request.Status = constants.STALE
			request.PolicyEvaluationAuditId = evaluationAuditEntry.Id
			requestsToBeUpdatedAsStaled = append(requestsToBeUpdatedAsStaled)
		}

	}

	return requestsToBeUpdatedAsStaled, nil

}
