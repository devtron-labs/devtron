package artifactPromotion

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/enterprise/pkg/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ArtifactPromotionApprovalService interface {
	HandleArtifactPromotionRequest(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) (*bean.ArtifactPromotionRequest, error)
	GetByPromotionRequestId(artifactPromotionApprovalRequest *ArtifactPromotionApprovalRequest) (*bean.ArtifactPromotionApprovalResponse, error)
}

type ArtifactPromotionApprovalServiceImpl struct {
	artifactPromotionApprovalRequestRepository ArtifactPromotionApprovalRequestRepository
	logger                                     *zap.SugaredLogger
	ciPipelineRepository                       pipelineConfig.CiPipelineRepository
	pipelineRepository                         pipelineConfig.PipelineRepository
	userService                                user.UserService
	ciArtifactRepository                       repository.CiArtifactRepository
	appWorkflowRepository                      appWorkflow.AppWorkflowRepository
	cdWorkflowRepository                       pipelineConfig.CdWorkflowRepository
}

func NewArtifactPromotionApprovalServiceImpl(
	ArtifactPromotionApprovalRequestRepository ArtifactPromotionApprovalRequestRepository,
	logger *zap.SugaredLogger,
	CiPipelineRepository pipelineConfig.CiPipelineRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	userService user.UserService,
	ciArtifactRepository repository.CiArtifactRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
) *ArtifactPromotionApprovalServiceImpl {
	return &ArtifactPromotionApprovalServiceImpl{
		artifactPromotionApprovalRequestRepository: ArtifactPromotionApprovalRequestRepository,
		logger:                logger,
		ciPipelineRepository:  CiPipelineRepository,
		pipelineRepository:    pipelineRepository,
		userService:           userService,
		ciArtifactRepository:  ciArtifactRepository,
		appWorkflowRepository: appWorkflowRepository,
		cdWorkflowRepository:  cdWorkflowRepository,
	}
}

func (impl ArtifactPromotionApprovalServiceImpl) HandleArtifactPromotionRequest(request *bean.ArtifactPromotionRequest, authorizedEnvironments map[string]bool) (*bean.ArtifactPromotionRequest, error) {
	switch request.Action {

	case bean.ACTION_PROMOTE:

	case bean.ACTION_APPROVE:

	case bean.ACTION_CANCEL:

		artifactPromotionRequest, err := impl.cancelPromotionApprovalRequest(request)
		if err != nil {
			impl.logger.Errorw("error in canceling artifact promotion approval request", "promotionRequestId", request.PromotionRequestId, "err", err)
			return nil, err
		}
		return artifactPromotionRequest, nil

	}
	return nil, nil
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
func (impl ArtifactPromotionApprovalServiceImpl) promoteArtifact(request *bean.ArtifactPromotionRequest) (*bean.ArtifactPromotionRequest, error) {
	// 	step1: validate if artifact is deployed/created at the source pipeline.
	//      step1: if source is cd , check if this artifact is deployed on these environments
	//  step2: validate if destination pipeline is in the same workflow of the artifact source.
	//  step2.1: check if destination pipeline is topologically downwards from the source pipeline and also source and destination are on the same subtree.
	// 	step3: check if promotion request for this artifact on this destination pipeline has already been raised.
	//  step4: check if this artifact on this destination pipeline has already been promoted
	//  step5: raise request.

	// fetch artifact
	response := make(map[string]bean.EnvironmentResponse)
	for _, env := range request.EnvironmentNames {
		envResponse := bean.EnvironmentResponse{
			Name:                       env,
			PromotionEvaluationState:   bean.PIPELINE_NOT_FOUND,
			PromotionEvaluationMessage: string(bean.PIPELINE_NOT_FOUND),
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
		EnvResponse.PromotionEvaluationState = bean.EMPTY
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
				EnvResponse.PromotionEvaluationState = bean.SOURCE_AND_DESTINATION_PIPELINE_MISMATCH
			}
		}
	}

	impl.checkPromotionPolicyGovernance(request.AppId, allowedEnvs)
	for _, pipelineId := range pipelineIds {

		EnvResponse := response[pipelineIdVsEnvNameMap[pipelineId]]
		// these
		if EnvResponse.PromotionEvaluationState == bean.EMPTY {
			// todo send policy of this pipeline
			state, msg, err := impl.raisePromoteRequest(request, pipelineId, nil)
			if err != nil {
				impl.logger.Errorw("error in raising promotion request for the pipeline", "pipelineId", pipelineId, "artifactId", ciArtifact.Id, "err", err)
				EnvResponse.PromotionEvaluationState = bean.ERRORED
				EnvResponse.PromotionEvaluationMessage = err.Error()
			}
			EnvResponse.PromotionEvaluationState = state
			EnvResponse.PromotionEvaluationMessage = msg
		}

		response[pipelineIdVsEnvNameMap[pipelineId]] = EnvResponse
	}
	return nil, nil
}

func (impl ArtifactPromotionApprovalServiceImpl) checkPromotionPolicyGovernance(appId int, envIds []int) map[int]interface{} {
	// todo: implement me
	return nil
}

func (impl ArtifactPromotionApprovalServiceImpl) raisePromoteRequest(request *bean.ArtifactPromotionRequest, pipelineId int, promotionPolicyMetadata interface{}) (bean.PromotionEvaluationState, string, error) {
	requests, err := impl.artifactPromotionApprovalRequestRepository.FindAwaitedRequestByPipelineIdAndArtifactId(pipelineId, request.ArtifactId)
	if err != nil {
		impl.logger.Errorw("error in finding the pending promotion request using pipelineId and artifactId", "pipelineId", pipelineId, "artifactId", request.ArtifactId)
		return bean.ERRORED, err.Error(), err
	}

	if len(requests) >= 1 {
		return bean.ALREADY_REQUEST_RAISED, bean.ALREADY_REQUEST_RAISED, nil
	}

	promotedRequest, err := impl.artifactPromotionApprovalRequestRepository.FindPromotedRequestByPipelineIdAndArtifactId(pipelineId, request.ArtifactId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in finding the promoted request using pipelineId and artifactId", "pipelineId", pipelineId, "artifactId", request.ArtifactId)
		return bean.ERRORED, err.Error(), err
	}

	if promotedRequest.Id > 0 {
		return bean.ARTIFACT_ALREADY_PROMOTED, bean.ARTIFACT_ALREADY_PROMOTED, nil
	}

	promotionRequest := &ArtifactPromotionApprovalRequest{
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

func (impl ArtifactPromotionApprovalServiceImpl) GetByPromotionRequestId(artifactPromotionApprovalRequest *ArtifactPromotionApprovalRequest) (*bean.ArtifactPromotionApprovalResponse, error) {

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
