package artifactPromotion

import (
	"context"
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
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/read"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/repository"
	"github.com/devtron-labs/devtron/pkg/workflow/dag"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
	"net/http"
	"strings"
	"time"
)

type ArtifactPromotionApprovalService interface {
	HandleArtifactPromotionRequest(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) ([]bean.EnvironmentResponse, error)
	GetByPromotionRequestId(artifactPromotionApprovalRequest *repository.ArtifactPromotionApprovalRequest) (*bean.ArtifactPromotionApprovalResponse, error)
	FetchEnvironmentsList(envMap map[string]repository1.Environment, appId int, appName string, authorizedEnvironments map[string]bool, artifactId int) (*bean.EnvironmentListingResponse, error)
	FetchApprovalAllowedEnvList(artifactId int, userId int32, token string, promotionApproverAuth func(string, string) bool) ([]bean.EnvironmentApprovalMetadata, error)
	GetAppAndEnvsMapByWorkflowId(workflowId int) (*bean.WorkflowMetaData, error)
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
	promotionPolicyDataReadService             read.ArtifactPromotionDataReadService
	requestApprovalUserdataRepo                pipelineConfig.RequestApprovalUserdataRepository
	workflowDagExecutor                        dag.WorkflowDagExecutor
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
	promotionPolicyService read.ArtifactPromotionDataReadService,
	requestApprovalUserdataRepo pipelineConfig.RequestApprovalUserdataRepository,
	workflowDagExecutor dag.WorkflowDagExecutor,
	promotionPolicyCUDService PromotionPolicyCUDService,
) *ArtifactPromotionApprovalServiceImpl {

	artifactApprovalService := &ArtifactPromotionApprovalServiceImpl{
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
		promotionPolicyDataReadService:    promotionPolicyService,
		requestApprovalUserdataRepo:       requestApprovalUserdataRepo,
		workflowDagExecutor:               workflowDagExecutor,
	}

	// register hooks
	promotionPolicyCUDService.AddPreDeleteHook(artifactApprovalService.onPolicyDelete)
	promotionPolicyCUDService.AddPreUpdateHook(artifactApprovalService.onPolicyUpdate)
	return artifactApprovalService
}

func (impl ArtifactPromotionApprovalServiceImpl) GetAppAndEnvsMapByWorkflowId(workflowId int) (*bean.WorkflowMetaData, error) {
	allAppWorkflowMappings, err := impl.appWorkflowRepository.FindWFAllMappingByWorkflowId(workflowId)
	if err != nil {
		impl.logger.Errorw("error in finding the app workflow mappings using appWorkflowId", "workflowId", workflowId, "err", err)
		return nil, err
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

	sourceType := sourcePipelineMapping.Type
	sourceId := sourcePipelineMapping.Id
	pipelines, err := impl.pipelineRepository.FindByIdsIn(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error in finding pipelines", "pipelineIds", pipelineIds, "err", err)
		return nil, err
	}
	envMap := make(map[string]repository1.Environment)
	appName := ""
	appId := 0
	for _, pipeline := range pipelines {
		envMap[pipeline.Environment.Name] = pipeline.Environment
		appName = pipeline.App.AppName
		appId = pipeline.AppId
	}
	wfMeta := &bean.WorkflowMetaData{
		WorkflowId: workflowId,
		AppId:      appId,
		AppName:    appName,
		EnvMap:     envMap,
		CiSourceData: bean.CiSourceMetaData{
			Id:   sourceId,
			Type: bean.SourceTypeStr(sourceType),
		},
	}
	return wfMeta, nil
}

func (impl ArtifactPromotionApprovalServiceImpl) FetchEnvironmentsList(envMap map[string]repository1.Environment, appId int, appName string, authorizedEnvironments map[string]bool, artifactId int) (*bean.EnvironmentListingResponse, error) {
	envNames := make([]string, 0, len(envMap))
	envIds := make([]int, 0, len(envMap))
	for envName, env := range envMap {
		envIds = append(envIds, env.Id)
		envNames = append(envNames, envName)
	}
	result := &bean.EnvironmentListingResponse{}
	policiesMap, err := impl.promotionPolicyDataReadService.GetPromotionPolicyByAppAndEnvIds(appId, envIds)
	if err != nil {
		impl.logger.Errorw("error in getting the policies", "appId", appName, "envIds", envNames, "err", err)
		return nil, err
	}
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
		responses, err := impl.evaluatePoliciesOnArtifact(ciArtifact, envMap, authorizedEnvironments, policiesMap)
		if err != nil {
			return nil, err
		}
		result.Environments = responses
		return result, nil
	}

	responseMap := make(map[string]bean.EnvironmentResponse)
	for envName, _ := range envMap {
		resp := bean.EnvironmentResponse{
			Name:                       envName,
			PromotionPossible:          pointer.Bool(false),
			PromotionValidationMessage: string(bean.POLICY_NOT_CONFIGURED),
			PromotionValidationState:   bean.POLICY_NOT_CONFIGURED,
			IsVirtualEnvironment:       pointer.Bool(envMap[envName].IsVirtualEnvironment),
		}
		if !authorizedEnvironments[envName] {
			resp.PromotionValidationState = bean.NO_PERMISSION
			resp.PromotionValidationMessage = string(bean.NO_PERMISSION)
		}
		responseMap[envName] = resp
	}

	for envName, policy := range policiesMap {
		responseMap[envName] = bean.EnvironmentResponse{
			Name:                       envName,
			ApprovalCount:              policy.ApprovalMetaData.ApprovalCount,
			IsVirtualEnvironment:       pointer.Bool(envMap[envName].IsVirtualEnvironment),
			PromotionValidationMessage: "",
			PromotionValidationState:   bean.EMPTY,
		}
	}

	responses := make([]bean.EnvironmentResponse, 0, len(responseMap))
	for _, envResponse := range responseMap {
		responses = append(responses, envResponse)
	}

	result.Environments = responses
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

func (impl ArtifactPromotionApprovalServiceImpl) evaluatePoliciesOnArtifact(ciArtifact *repository2.CiArtifact, envMap map[string]repository1.Environment, authorizedEnvironments map[string]bool, policiesMap map[string]*bean.PromotionPolicy) ([]bean.EnvironmentResponse, error) {
	params, err := impl.computeFilterParams(ciArtifact)
	if err != nil {
		impl.logger.Errorw("error in finding the required CEL expression parameters for using ciArtifact", "err", err)
		return nil, err
	}
	responseMap := make(map[string]bean.EnvironmentResponse)
	for envName, _ := range envMap {
		resp := bean.EnvironmentResponse{
			Name:                       envName,
			PromotionPossible:          pointer.Bool(false),
			PromotionValidationMessage: string(bean.POLICY_NOT_CONFIGURED),
			PromotionValidationState:   bean.POLICY_NOT_CONFIGURED,
			IsVirtualEnvironment:       pointer.Bool(envMap[envName].IsVirtualEnvironment),
		}
		if !authorizedEnvironments[envName] {
			resp.PromotionValidationState = bean.NO_PERMISSION
			resp.PromotionValidationMessage = string(bean.NO_PERMISSION)
		}
		responseMap[envName] = resp
	}

	for envName, policy := range policiesMap {
		evaluationResult, err := impl.resourceFilterConditionsEvaluator.EvaluateFilter(policy.Conditions, resourceFilter.ExpressionMetadata{Params: params})
		if err != nil {
			impl.logger.Errorw("evaluation failed with error", "policyConditions", policy.Conditions, "envName", envName, policy.Conditions, "params", params, "err", err)
			responseMap[envName] = bean.EnvironmentResponse{
				ApprovalCount:            policy.ApprovalMetaData.ApprovalCount,
				PromotionPossible:        pointer.Bool(false),
				PromotionValidationState: bean.POLICY_EVALUATION_ERRORED,
			}
			continue
		}
		envResp := bean.EnvironmentResponse{
			ApprovalCount:     policy.ApprovalMetaData.ApprovalCount,
			PromotionPossible: pointer.Bool(evaluationResult),
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

func (impl ArtifactPromotionApprovalServiceImpl) approveArtifactPromotion(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) ([]bean.EnvironmentResponse, error) {
	// get request and check if it is promoted already.
	// attempt approving this by creating new resource_approval_user_data, if unique constraint error ,current user already did something.
	// attempt success , then get the approval count and check no of approvals got
	//  promote if approvalCount > approvals received
	responses := make(map[string]*bean.EnvironmentResponse)
	environmentNames := make([]string, 0)
	for envName, authorised := range authorizedEnvironments {
		envResponse := &bean.EnvironmentResponse{
			Name: envName,
		}
		if !authorised {
			envResponse.PromotionValidationState = bean.NO_PERMISSION
			envResponse.PromotionValidationMessage = string(bean.NO_PERMISSION)
		} else {
			environmentNames = append(environmentNames, envName)
		}
		responses[envName] = envResponse
	}

	pipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvNames(request.AppId, environmentNames)
	if err != nil {
		impl.logger.Errorw("error in finding the cd pipelines using appID and environment names", "appId", request.AppId, "envNames", environmentNames, "err", err)
		return nil, err
	}

	pipelineIdVsEnvMap := make(map[int]string)
	pipelineIdToDaoMap := make(map[int]*pipelineConfig.Pipeline)
	pipelineIds := make([]int, 0, len(pipelines))
	for _, pipeline := range pipelines {
		pipelineIdVsEnvMap[pipeline.Id] = pipeline.Environment.Name
		pipelineIds = append(pipelineIds, pipeline.Id)
		pipelineIdToDaoMap[pipeline.Id] = pipeline
	}

	promotionRequests, err := impl.artifactPromotionApprovalRequestRepository.FindByDestinationPipelineIds(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error in getting artifact promotion request object by id", "promotionRequestId", request.PromotionRequestId, "err", err)
		if errors.Is(err, pg.ErrNoRows) {
			return nil, errors.New("promotion request not found")
		}
		return nil, err
	}

	// policy Ids extracted from the requests
	policyIds := make([]int, 0, len(promotionRequests))
	for _, promotionRequest := range promotionRequests {
		policyIds = append(policyIds, promotionRequest.PolicyId)
	}

	// policies fetched form above policy ids
	policies, err := impl.promotionPolicyDataReadService.GetPromotionPolicyByIds(policyIds)
	if err != nil {
		impl.logger.Errorw("error in finding the promotionPolicy by ids", "policyIds", policyIds, "err", err)
		return nil, err
	}

	// map the policies for O(1) access
	policyIdMap := make(map[int]*bean.PromotionPolicy)
	for _, policy := range policies {
		policyIdMap[policy.Id] = policy
	}

	staleRequestIds := make([]int, 0)
	validRequestIds := make([]int, 0)
	validRequestPolicyMap := make(map[int]int)
	for _, promotionRequest := range promotionRequests {
		resp := responses[pipelineIdVsEnvMap[promotionRequest.DestinationPipelineId]]
		_, ok := policyIdMap[promotionRequest.PolicyId]
		if ok {
			validRequestIds = append(validRequestIds, promotionRequest.Id)
			validRequestPolicyMap[promotionRequest.Id] = promotionRequest.PolicyId
		}
		if !ok {
			// policy is not found in the map, and the request is still in awaiting state.
			// this is a stale case.
			// mark it stale
			staleRequestIds = append(staleRequestIds, promotionRequest.Id)

			// also set the response messages
			resp.PromotionPossible = pointer.Bool(false)
			resp.PromotionValidationMessage = "request is no longer valid, state: stale"
			resp.PromotionValidationState = bean.PromotionValidationState(resp.PromotionValidationMessage)
		} else if promotionRequest.Status != bean.AWAITING_APPROVAL {
			resp.PromotionValidationMessage = fmt.Sprintf("artifact is in %s state", promotionRequest.Status.Status())
			resp.PromotionValidationState = bean.PromotionValidationState(resp.PromotionValidationMessage)
		}
	}

	if len(staleRequestIds) > 0 {
		// attempt deleting the stale requests in bulk
		err = impl.artifactPromotionApprovalRequestRepository.MarkStaleByIds(staleRequestIds)
		if err != nil {
			impl.logger.Errorw("error in deleting the request raised using a deleted promotion policy (stale requests)", "staleRequestIds", staleRequestIds, "err", err)
			// not returning by choice, don't interrupt the user flow
		}
	}

	tx, err := impl.artifactPromotionApprovalRequestRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting the transaction", "promotionRequests", promotionRequests, "err", err)
		return nil, err
	}
	defer impl.artifactPromotionApprovalRequestRepository.RollbackTx(tx)

	for _, promotionRequest := range promotionRequests {
		promotionRequestApprovedUserData := &pipelineConfig.RequestApprovalUserData{
			ApprovalRequestId: promotionRequest.Id,
			RequestType:       repository2.ARTIFACT_PROMOTION_APPROVAL,
			UserId:            request.UserId,
			UserResponse:      pipelineConfig.APPROVED,
		}
		resp := responses[pipelineIdVsEnvMap[promotionRequest.DestinationPipelineId]]
		// have to do this in loop as we have to ensure partial approval even in case of partial failure
		err = impl.requestApprovalUserdataRepo.SaveDeploymentUserData(promotionRequestApprovedUserData)
		if err != nil {
			impl.logger.Errorw("error in saving promotion approval user data", "promotionRequestId", request.PromotionRequestId, "err", err)
			if strings.Contains(err.Error(), "unique_user_request_action") {
				err = errors.New("you have already approved this")
				resp.PromotionValidationState = bean.APPROVED
				resp.PromotionValidationMessage = err.Error()
			} else {
				err = errors.New(fmt.Sprintf("server error: %s", err.Error()))
				resp.PromotionValidationState = bean.ERRORED
				resp.PromotionValidationMessage = err.Error()
			}
			resp.PromotionValidationState = bean.APPROVED
			resp.PromotionValidationMessage = string(bean.APPROVED)
			continue
		}

		resp.PromotionValidationState = bean.APPROVED
		resp.PromotionValidationMessage = string(bean.APPROVED)
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
		if approvalCount >= policyIdMap[promotionRequest.PolicyId].ApprovalMetaData.ApprovalCount {
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

		promotedCiArtifactIds := make([]int, 0)
		for _, id := range promotableRequestIds {
			promotableRequest := promotionRequestIdToDaoMap[id]
			promotedCiArtifactIds = append(promotedCiArtifactIds, promotableRequest.ArtifactId)
		}

		ciArtifact, err := impl.ciArtifactRepository.Get(request.ArtifactId)
		if err != nil {
			impl.logger.Errorw("error in fetching promoted ci artifact by artifact id", "artifactId", request.ArtifactId, "err", err)
			return nil, err
		}

		for _, id := range promotableRequestIds {
			promotableRequest := promotionRequestIdToDaoMap[id]
			pipelineDao := pipelineIdToDaoMap[promotableRequest.DestinationPipelineId]
			triggerRequest := bean2.TriggerRequest{
				CdWf:        nil,
				Pipeline:    pipelineDao,
				Artifact:    ciArtifact,
				TriggeredBy: 1,
				TriggerContext: bean2.TriggerContext{
					Context: context.Background(),
				},
			}
			impl.workflowDagExecutor.HandleArtifactPromotionEvent(triggerRequest)
		}
	}

	err = impl.artifactPromotionApprovalRequestRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing the transaction", "promotableRequestIds", promotableRequestIds, "err", err)
		return nil, err
	}

	if len(promotableRequestIds) > 0 {
		// todo: trigger release
	}
	return nil, nil
}

func (impl ArtifactPromotionApprovalServiceImpl) HandleArtifactPromotionRequest(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) ([]bean.EnvironmentResponse, error) {

	switch request.Action {

	case bean.ACTION_PROMOTE:
		return impl.promoteArtifact(request, authorizedEnvironments)
	case bean.ACTION_APPROVE:
		return impl.approveArtifactPromotion(request, authorizedEnvironments)
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

func (impl ArtifactPromotionApprovalServiceImpl) promoteArtifact(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) ([]bean.EnvironmentResponse, error) {
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
		if !authorizedEnvironments[env] {
			envResponse.PromotionValidationState = bean.NO_PERMISSION
			envResponse.PromotionValidationMessage = string(bean.NO_PERMISSION)
		}
		response[env] = envResponse
	}

	allowedEnvs := make([]int, 0, len(request.EnvNameIdMap))
	allowedEnvNames := make([]string, 0, len(request.EnvNameIdMap))
	for envName, envId := range request.EnvNameIdMap {
		if authorizedEnvironments[envName] {
			allowedEnvs = append(allowedEnvs, envId)
			allowedEnvNames = append(allowedEnvNames, envName)
		}
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
	pipelineIdToDaoMap := make(map[int]*pipelineConfig.Pipeline)
	for _, pipeline := range pipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
		pipelineIdVsEnvNameMap[pipeline.Id] = request.EnvIdNameMap[pipeline.EnvironmentId]
		EnvResponse := response[pipelineIdVsEnvNameMap[pipeline.Id]]
		EnvResponse.PromotionValidationState = bean.EMPTY
		response[pipelineIdVsEnvNameMap[pipeline.Id]] = EnvResponse
		pipelineIdToDaoMap[pipeline.Id] = pipeline
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
			impl.logger.Errorw("error in checking if artifact is available for promotion at source pipeline", "ciArtifactId", ciArtifact.Id, "sourcePipelineId", sourcePipelineId, "err", err)
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

	policiesMap, err := impl.promotionPolicyDataReadService.GetPromotionPolicyByAppAndEnvIds(request.AppId, allowedEnvs)
	if err != nil {
		impl.logger.Errorw("error in getting policies for some environments in an app", "appName", request.AppName, "envNames", allowedEnvNames, "err", err)
		return nil, err
	}
	for _, pipelineId := range pipelineIds {

		EnvResponse := response[pipelineIdVsEnvNameMap[pipelineId]]
		// these
		if EnvResponse.PromotionValidationState == bean.EMPTY {
			policy := policiesMap[pipelineIdVsEnvNameMap[pipelineId]]
			state, msg, err := impl.raisePromoteRequest(request, pipelineId, policy, ciArtifact, pipelineIdToDaoMap[pipelineId])
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

func (impl ArtifactPromotionApprovalServiceImpl) raisePromoteRequest(request *bean.ArtifactPromotionRequest, pipelineId int, promotionPolicy *bean.PromotionPolicy, ciArtifact *repository2.CiArtifact, cdPipeline *pipelineConfig.Pipeline) (bean.PromotionValidationState, string, error) {
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
		SourceType:              request.SourceType.GetSourceType(),
		SourcePipelineId:        request.SourcePipelineId,
		DestinationPipelineId:   pipelineId,
		Status:                  bean.AWAITING_APPROVAL,
		Active:                  true,
		ArtifactId:              request.ArtifactId,
		PolicyId:                promotionPolicy.Id,
		PolicyEvaluationAuditId: promotionPolicy.PolicyEvaluationId,
	}

	var status bean.PromotionValidationState
	if promotionPolicy.ApprovalMetaData.ApprovalCount == 0 {
		promotionRequest.Status = bean.PROMOTED
		status = bean.PROMOTION_SUCCESSFUL
	} else {
		status = bean.SENT_FOR_APPROVAL
	}
	_, err = impl.artifactPromotionApprovalRequestRepository.Create(promotionRequest)
	if err != nil {
		impl.logger.Errorw("error in finding the pending promotion request using pipelineId and artifactId", "pipelineId", pipelineId, "artifactId", request.ArtifactId)
		return bean.ERRORED, err.Error(), err
	}

	if promotionRequest.Status == bean.PROMOTED {
		triggerRequest := bean2.TriggerRequest{
			CdWf:        nil,
			Pipeline:    cdPipeline,
			Artifact:    ciArtifact,
			TriggeredBy: 1,
			TriggerContext: bean2.TriggerContext{
				Context: context.Background(),
			},
		}
		impl.workflowDagExecutor.HandleArtifactPromotionEvent(triggerRequest)
	}
	return status, string(status), nil

}

func (impl ArtifactPromotionApprovalServiceImpl) checkIfDeployedAtSource(ciArtifactId, sourcePipelineId int) (bool, error) {
	// todo: implement me
	return true, nil
}

func (impl ArtifactPromotionApprovalServiceImpl) cancelPromotionApprovalRequest(request *bean.ArtifactPromotionRequest) (*bean.ArtifactPromotionRequest, error) {
	artifactPromotionDao, err := impl.artifactPromotionApprovalRequestRepository.FindById(request.PromotionRequestId)
	if errors.Is(err, pg.ErrNoRows) {
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

	if artifactPromotionDao.CreatedBy != request.UserId {
		return nil, &util.ApiError{
			HttpStatusCode:  http.StatusUnprocessableEntity,
			InternalMessage: "only user who has raised the promotion request can cancel it",
			UserMessage:     "only user who has raised the promotion request can cancel it",
		}
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

	sourceType := artifactPromotionApprovalRequest.SourceType.GetSourceTypeStr()

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

	policyMap, err := impl.promotionPolicyDataReadService.GetPromotionPolicyByAppAndEnvIds(destCDPipeline.AppId, []int{destCDPipeline.EnvironmentId})
	if err != nil {
		impl.logger.Errorw("error in fetching policies", "appName", destCDPipeline.App.AppName, "envName", destCDPipeline.Environment.Name, "err", err)
		return nil, err
	}
	policy := policyMap[destCDPipeline.Environment.Name]
	artifactPromotionApprovalResponse := &bean.ArtifactPromotionApprovalResponse{
		SourceType:      sourceType,
		Source:          source,
		Destination:     destCDPipeline.Environment.Name,
		RequestedBy:     artifactPromotionRequestUser.EmailId,
		ApprovedUsers:   make([]string, 0), // get by deployment_approval_user_data
		RequestedOn:     artifactPromotionApprovalRequest.CreatedOn,
		PromotedOn:      artifactPromotionApprovalRequest.UpdatedOn,
		PromotionPolicy: policy.Name,
	}

	return artifactPromotionApprovalResponse, nil

}

func (impl ArtifactPromotionApprovalServiceImpl) FetchApprovalAllowedEnvList(artifactId int, userId int32, token string, promotionApproverAuth func(string, string) bool) ([]bean.EnvironmentApprovalMetadata, error) {

	artifact, err := impl.ciArtifactRepository.Get(artifactId)
	if err != nil {
		impl.logger.Errorw("artifact not found for given id", "artifactId", artifactId, "err", err)
		return nil, &util.ApiError{
			HttpStatusCode:  http.StatusUnprocessableEntity,
			InternalMessage: "artifact not found for given id",
			UserMessage:     "artifact not found for given id",
		}
	}

	promotionRequests, err := impl.artifactPromotionApprovalRequestRepository.FindAwaitedRequestsByArtifactId(artifactId)
	if err != nil {
		impl.logger.Errorw("error in finding promotion requests in awaiting state for given artifactId ")
	}

	environmentApprovalMetadata := make([]bean.EnvironmentApprovalMetadata, 0)

	if len(promotionRequests) == 0 {
		return environmentApprovalMetadata, nil
	}

	destinationPipelineIds := make([]int, 0)
	for _, request := range promotionRequests {
		destinationPipelineIds = append(destinationPipelineIds, request.Id)
	}

	pipelines, err := impl.pipelineRepository.FindAppAndEnvironmentAndProjectByPipelineIds(destinationPipelineIds)
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines by ids", "pipelineIds", destinationPipelineIds, "err", err)
		return nil, err
	}

	pipelineIdMap := make(map[int]*pipelineConfig.Pipeline)
	for _, pipelineDao := range pipelines {
		pipelineIdMap[pipelineDao.Id] = pipelineDao
	}

	for _, request := range promotionRequests {

		pipelineDao := pipelineIdMap[request.DestinationPipelineId]

		environmentMetadata := bean.EnvironmentApprovalMetadata{
			Name:            pipelineDao.Environment.Name,
			ApprovalAllowed: true,
			Reasons:         make([]string, 0),
		}

		policy, err := impl.promotionPolicyDataReadService.GetPromotionPolicyByAppAndEnvId(pipelineDao.AppId, pipelineDao.EnvironmentId)
		if err != nil {
			impl.logger.Errorw("error in fetching promotion policy for given appId and envId", "appId", pipelineDao.AppId, "envId", pipelineDao.EnvironmentId, "err", err)
			return nil, err
		}

		if !policy.ApprovalMetaData.AllowImageBuilderFromApprove && request.CreatedBy == artifact.CreatedBy {
			environmentMetadata.ApprovalAllowed = false
			environmentMetadata.Reasons = append(environmentMetadata.Reasons, "User who has built the image cannot approve promotion request for this environment")
		}

		if !policy.ApprovalMetaData.AllowRequesterFromApprove && request.CreatedBy == userId {
			environmentMetadata.ApprovalAllowed = false
			environmentMetadata.Reasons = append(environmentMetadata.Reasons, "User who has raised the image cannot approve promotion request for this environment")
		}

		rbacObject := fmt.Sprintf("%s/%s/%s", pipelineDao.App.Team.Name, pipelineDao.Environment.EnvironmentIdentifier, pipelineDao.App.AppName)
		if ok := promotionApproverAuth(token, rbacObject); !ok {
			environmentMetadata.ApprovalAllowed = false
			environmentMetadata.Reasons = append(environmentMetadata.Reasons, "user does not have image promoter access for given app and env")
		}
	}
	return environmentApprovalMetadata, nil
}

func (impl ArtifactPromotionApprovalServiceImpl) onPolicyDelete(tx *pg.Tx, policyId int) error {
	err := impl.artifactPromotionApprovalRequestRepository.MarkStaleByPolicyId(tx, policyId)
	if err != nil {
		impl.logger.Errorw("error in marking artifact promotion requests stale", "policyId", policyId, "err", err)
	}
	return err
}

func (impl ArtifactPromotionApprovalServiceImpl) onPolicyUpdate(tx *pg.Tx, policyId int) error {
	err := impl.artifactPromotionApprovalRequestRepository.MarkStaleByPolicyId(tx, policyId)
	if err != nil {
		impl.logger.Errorw("error in marking artifact promotion requests stale", "policyId", policyId, "err", err)
	}
	return err
}
