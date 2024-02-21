package artifactPromotionApprovalRequest

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
)

type ArtifactPromotionApprovalService interface {
	HandleArtifactPromotionRequest(request *ArtifactPromotionRequest, envAuthorized map[string]bool) (*ArtifactPromotionRequest, error)
	GetByPromotionRequestId(artifactPromotionApprovalRequest *ArtifactPromotionApprovalRequest) (*ArtifactPromotionApprovalResponse, error)
}

type ArtifactPromotionApprovalServiceImpl struct {
	artifactPromotionApprovalRequestRepository ArtifactPromotionApprovalRequestRepository
	logger                                     *zap.SugaredLogger
	CiPipelineRepository                       pipelineConfig.CiPipelineRepository
	pipelineRepository                         pipelineConfig.PipelineRepository
	userService                                user.UserService
}

func NewArtifactPromotionApprovalServiceImpl(
	ArtifactPromotionApprovalRequestRepository ArtifactPromotionApprovalRequestRepository,
	logger *zap.SugaredLogger,
	CiPipelineRepository pipelineConfig.CiPipelineRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	userService user.UserService,
) *ArtifactPromotionApprovalServiceImpl {
	return &ArtifactPromotionApprovalServiceImpl{
		artifactPromotionApprovalRequestRepository: ArtifactPromotionApprovalRequestRepository,
		logger:               logger,
		CiPipelineRepository: CiPipelineRepository,
		pipelineRepository:   pipelineRepository,
		userService:          userService,
	}
}

func (impl ArtifactPromotionApprovalServiceImpl) HandleArtifactPromotionRequest(request *ArtifactPromotionRequest, envAuthorized map[string]bool) (*ArtifactPromotionRequest, error) {
	switch request.Action {

	case ACTION_PROMOTE:

	case ACTION_APPROVE:

	case ACTION_CANCEL:

		artifactPromotionRequest, err := impl.cancelPromotionApprovalRequest(request)
		if err != nil {
			impl.logger.Errorw("error in canceling artifact promotion approval request", "promotionRequestId", request.PromotionRequestId, "err", err)
			return nil, err
		}
		return artifactPromotionRequest, nil

	}
	return nil, nil
}

func (impl ArtifactPromotionApprovalRequest) promoteArtifact(request *ArtifactPromotionRequest) (*ArtifactPromotionRequest, error) {
	// TODO: add validations on artifactId, sourceId and destinationId
	return nil, nil
}

func (impl ArtifactPromotionApprovalServiceImpl) cancelPromotionApprovalRequest(request *ArtifactPromotionRequest) (*ArtifactPromotionRequest, error) {
	artifactPromotionDao, err := impl.artifactPromotionApprovalRequestRepository.FindById(request.PromotionRequestId)
	if err == pg.ErrNoRows {
		impl.logger.Errorw("artifact promotion approval request not found for given id", "promotionRequestId", request.PromotionRequestId, "err", err)
		return nil, &util.ApiError{
			HttpStatusCode:  http.StatusNotFound,
			InternalMessage: ArtifactPromotionRequestNotFoundErr,
			UserMessage:     ArtifactPromotionRequestNotFoundErr,
		}
	}
	if err != nil {
		impl.logger.Errorw("error in fetching artifact promotion request by id", "artifactPromotionRequestId", request.PromotionRequestId, "err", err)
		return nil, err
	}
	artifactPromotionDao.Active = false
	artifactPromotionDao.Status = CANCELED
	_, err = impl.artifactPromotionApprovalRequestRepository.Update(artifactPromotionDao)
	if err != nil {
		impl.logger.Errorw("error in updating artifact promotion approval request", "artifactPromotionRequestId", request.PromotionRequestId, "err", err)
		return nil, err
	}
	return nil, err
}

func (impl ArtifactPromotionApprovalServiceImpl) GetByPromotionRequestId(artifactPromotionApprovalRequest *ArtifactPromotionApprovalRequest) (*ArtifactPromotionApprovalResponse, error) {

	sourceType := getSourceType(artifactPromotionApprovalRequest.SourceType)

	var source string
	if artifactPromotionApprovalRequest.SourceType == CD {
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

	artifactPromotionApprovalResponse := &ArtifactPromotionApprovalResponse{
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
