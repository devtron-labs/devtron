package read

import (
	bean3 "github.com/devtron-labs/devtron/api/bean"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ArtifactPromotionDataReadService interface {
	FetchPromotionApprovalDataForArtifacts(artifactIds []int, pipelineId int) (map[int]*bean.PromotionApprovalMetaData, error)
}

type ArtifactPromotionDataReadServiceImpl struct {
	logger                                     *zap.SugaredLogger
	artifactPromotionApprovalRequestRepository repository.ArtifactPromotionApprovalRequestRepository
	requestApprovalUserdataRepo                pipelineConfig.RequestApprovalUserdataRepository
	userService                                user.UserService
	pipelineRepository                         pipelineConfig.PipelineRepository
}

func NewArtifactPromotionDataReadServiceImpl(
	ArtifactPromotionApprovalRequestRepository repository.ArtifactPromotionApprovalRequestRepository,
	logger *zap.SugaredLogger,
	requestApprovalUserdataRepo pipelineConfig.RequestApprovalUserdataRepository,
	userService user.UserService,
	pipelineRepository pipelineConfig.PipelineRepository,
) *ArtifactPromotionDataReadServiceImpl {
	return &ArtifactPromotionDataReadServiceImpl{
		artifactPromotionApprovalRequestRepository: ArtifactPromotionApprovalRequestRepository,
		logger:                      logger,
		requestApprovalUserdataRepo: requestApprovalUserdataRepo,
		userService:                 userService,
		pipelineRepository:          pipelineRepository,
	}
}

func (impl ArtifactPromotionDataReadServiceImpl) FetchPromotionApprovalDataForArtifacts(artifactIds []int, pipelineId int) (map[int]*bean.PromotionApprovalMetaData, error) {

	promotionApprovalMetadata := make(map[int]*bean.PromotionApprovalMetaData)

	promotionApprovalRequest, err := impl.artifactPromotionApprovalRequestRepository.FindByPipelineIdAndArtifactIds(pipelineId, artifactIds)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching promotion request for given pipelineId and artifactId", "pipelineId", pipelineId, "artifactIds", artifactIds, "err", err)
		return promotionApprovalMetadata, nil
	}

	if len(promotionApprovalRequest) > 0 {

		var requestedUserIds []int32
		for _, approvalRequest := range promotionApprovalRequest {
			requestedUserIds = append(requestedUserIds, approvalRequest.CreatedBy)
		}

		userInfos, err := impl.userService.GetByIds(requestedUserIds)
		if err != nil {
			impl.logger.Errorw("error occurred while fetching users", "requestedUserIds", requestedUserIds, "err", err)
			return promotionApprovalMetadata, err
		}
		userInfoMap := make(map[int32]bean3.UserInfo)
		for _, userInfo := range userInfos {
			userId := userInfo.Id
			userInfoMap[userId] = userInfo
		}

		for _, approvalRequest := range promotionApprovalRequest {

			var promotedFrom string

			switch approvalRequest.SourceType {
			case bean.CI:
				promotedFrom = bean.SOURCE_TYPE_CI
			case bean.WEBHOOK:
				promotedFrom = bean.SOURCE_TYPE_CI
			case bean.CD:
				pipeline, err := impl.pipelineRepository.FindById(pipelineId)
				if err != nil {
					impl.logger.Errorw("error in fetching pipeline by id", "pipelineId", pipelineId, "err", err)
					return nil, err
				}
				promotedFrom = pipeline.Environment.Name
			}

			approvalMetadata := &bean.PromotionApprovalMetaData{
				ApprovalRequestId:    approvalRequest.Id,
				ApprovalRuntimeState: approvalRequest.Status.Status(),
				PromotedFrom:         promotedFrom,
				PromotedFromType:     approvalRequest.SourceType.GetSourceType(),
			}

			artifactId := approvalRequest.ArtifactId
			requestedUserId := approvalRequest.CreatedBy
			if userInfo, ok := userInfoMap[requestedUserId]; ok {
				approvalMetadata.RequestedUserData = bean.PromotionApprovalUserData{
					UserId:         userInfo.UserId,
					UserEmail:      userInfo.EmailId,
					UserActionTime: approvalRequest.CreatedOn,
				}
			}

			promotionApprovalUserData, err := impl.requestApprovalUserdataRepo.FetchApprovedDataByApprovalId(approvalRequest.Id, repository2.ARTIFACT_PROMOTION_APPROVAL)
			if err != nil {
				impl.logger.Errorw("error in getting promotionApprovalUserData", "err", err, "promotionApprovalRequestId", approvalRequest.Id)
			}

			for _, approvalUserData := range promotionApprovalUserData {
				approvalMetadata.ApprovalUsersData = append(approvalMetadata.ApprovalUsersData, bean.PromotionApprovalUserData{
					UserId:         approvalUserData.UserId,
					UserEmail:      approvalUserData.User.EmailId,
					UserActionTime: approvalUserData.CreatedOn,
				})
			}
			promotionApprovalMetadata[artifactId] = approvalMetadata
		}
	}
	return promotionApprovalMetadata, nil
}
