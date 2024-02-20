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
	HandleArtifactPromotionRequest(request *ArtifactPromotionRequest) (*ArtifactPromotionRequest, error)
	GetByPromotionRequestId(artifactPromotionApprovalRequestId int) (*ArtifactPromotionApprovalResponse, error)
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

func (impl ArtifactPromotionApprovalServiceImpl) HandleArtifactPromotionRequest(request *ArtifactPromotionRequest) (*ArtifactPromotionRequest, error) {
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

func (impl ArtifactPromotionApprovalServiceImpl) GetByPromotionRequestId(artifactPromotionApprovalRequestId int) (*ArtifactPromotionApprovalResponse, error) {

	artifactPromotionDao, err := impl.artifactPromotionApprovalRequestRepository.FindById(artifactPromotionApprovalRequestId)
	if err != nil {
		impl.logger.Errorw("error in fetching artifact promotion request by id", "artifactPromotionRequestId", artifactPromotionApprovalRequestId, "err", err)
		return nil, err
	}

	sourceType := getSourceType(artifactPromotionDao.SourceType)

	var source string
	if artifactPromotionDao.SourceType == CD {
		cdPipeline, err := impl.pipelineRepository.FindById(artifactPromotionDao.SourcePipelineId)
		if err != nil {
			impl.logger.Errorw("error in fetching cdPipeline by Id", "cdPipelineId", artifactPromotionDao.SourcePipelineId, "err", err)
			return nil, err
		}
		source = cdPipeline.Environment.Name
	}

	destCDPipeline, err := impl.pipelineRepository.FindById(artifactPromotionDao.DestinationPipelineId)
	if err != nil {
		impl.logger.Errorw("error in fetching cdPipeline by Id", "cdPipelineId", artifactPromotionDao.DestinationPipelineId, "err", err)
		return nil, err
	}

	artifactPromotionRequestUser, err := impl.userService.GetByIdWithoutGroupClaims(artifactPromotionDao.CreatedBy)
	if err != nil {
		impl.logger.Errorw("error in fetching user details by id", "userId", artifactPromotionDao.CreatedBy, "err", err)
		return nil, err
	}

	artifactPromotionApprovalResponse := &ArtifactPromotionApprovalResponse{
		SourceType:      sourceType,
		Source:          source,
		Destination:     destCDPipeline.Environment.Name,
		RequestedBy:     artifactPromotionRequestUser.EmailId,
		ApprovedUsers:   make([]string, 0), // get by deployment_approval_user_data
		RequestedOn:     artifactPromotionDao.CreatedOn,
		PromotedOn:      artifactPromotionDao.UpdatedOn,
		PromotionPolicy: "", // todo
	}

	return artifactPromotionApprovalResponse, nil

}
