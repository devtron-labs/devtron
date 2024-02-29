package read

import (
	"errors"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
)

type ArtifactPromotionDataReadService interface {
	FetchPromotionApprovalDataForArtifacts(artifactIds []int, pipelineId int) (map[int]*bean.PromotionApprovalMetaData, error)
	GetPromotionPolicyByAppAndEnvId(appId, envId int) (*bean.PromotionPolicy, error)
	GetPromotionPolicyByAppAndEnvIds(appId int, envIds []int) (map[string]*bean.PromotionPolicy, error)
	GetPromotionPolicyById(id int) (*bean.PromotionPolicy, error)
	GetPromotionPolicyByIds(ids []int) ([]*bean.PromotionPolicy, error)
	GetPromotionPolicyByName(name string) (*bean.PromotionPolicy, error)
	GetPoliciesMetadata(policyMetadataRequest bean.PromotionPolicyMetaRequest) (*bean.PromotionPolicyExtraResponse, error)
}

type ArtifactPromotionDataReadServiceImpl struct {
	logger                                     *zap.SugaredLogger
	artifactPromotionApprovalRequestRepository repository.ArtifactPromotionApprovalRequestRepository
	requestApprovalUserdataRepo                pipelineConfig.RequestApprovalUserdataRepository
	userService                                user.UserService
	pipelineRepository                         pipelineConfig.PipelineRepository
	resourceQualifierMappingService            resourceQualifiers.QualifierMappingService
	globalPolicyDataManager                    globalPolicy.GlobalPolicyDataManager
}

func NewArtifactPromotionDataReadServiceImpl(
	ArtifactPromotionApprovalRequestRepository repository.ArtifactPromotionApprovalRequestRepository,
	logger *zap.SugaredLogger,
	requestApprovalUserdataRepo pipelineConfig.RequestApprovalUserdataRepository,
	userService user.UserService,
	pipelineRepository pipelineConfig.PipelineRepository,
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService,
	globalPolicyDataManager globalPolicy.GlobalPolicyDataManager,
) *ArtifactPromotionDataReadServiceImpl {
	return &ArtifactPromotionDataReadServiceImpl{
		artifactPromotionApprovalRequestRepository: ArtifactPromotionApprovalRequestRepository,
		logger:                          logger,
		requestApprovalUserdataRepo:     requestApprovalUserdataRepo,
		userService:                     userService,
		pipelineRepository:              pipelineRepository,
		resourceQualifierMappingService: resourceQualifierMappingService,
		globalPolicyDataManager:         globalPolicyDataManager,
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
				promotedFrom = string(bean.SOURCE_TYPE_CI)
			case bean.WEBHOOK:
				promotedFrom = string(bean.SOURCE_TYPE_CI)
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
				PromotedFromType:     string(approvalRequest.SourceType.GetSourceTypeStr()),
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

func (impl ArtifactPromotionDataReadServiceImpl) GetPromotionPolicyByAppAndEnvId(appId, envId int) (*bean.PromotionPolicy, error) {

	scope := &resourceQualifiers.Scope{AppId: appId, EnvId: envId}
	//
	qualifierMapping, err := impl.resourceQualifierMappingService.GetResourceMappingsForScopes(
		resourceQualifiers.ImagePromotionPolicy,
		resourceQualifiers.ApplicationEnvironmentSelector,
		[]*resourceQualifiers.Scope{scope},
	)
	if err != nil {
		impl.logger.Errorw("error in fetching resource qualifier mapping by scope", "resource", resourceQualifiers.ImagePromotionPolicy, "scope", scope, "err", err)
		return nil, err
	}

	if len(qualifierMapping) == 0 {
		impl.logger.Infow("no artifact promotion policy found for given app and env", "appId", appId, "envId", envId, "err", err)
		return nil, nil
	}

	policyId := qualifierMapping[0].ResourceId
	rawPolicy, err := impl.globalPolicyDataManager.GetPolicyById(policyId)
	if err != nil {
		impl.logger.Errorw("error in finding policies by id", "policyId", policyId, "err", err)
		return nil, err
	}
	policy := &bean.PromotionPolicy{}
	err = policy.UpdateWithGlobalPolicy(rawPolicy)
	if err != nil {
		impl.logger.Errorw("error in extracting policy from globalPolicy json", "policyId", rawPolicy.Id, "err", err)
		return nil, err
	}
	return policy, nil
}

func (impl ArtifactPromotionDataReadServiceImpl) GetPromotionPolicyByAppAndEnvIds(appId int, envIds []int) (map[string]*bean.PromotionPolicy, error) {
	scopes := make([]*resourceQualifiers.Scope, 0, len(envIds))
	for _, envId := range envIds {
		scopes = append(scopes, &resourceQualifiers.Scope{
			AppId: appId,
			EnvId: envId,
		})
	}

	resourceQualifierMappings, err := impl.resourceQualifierMappingService.GetResourceMappingsForScopes(resourceQualifiers.ImagePromotionPolicy, resourceQualifiers.ApplicationEnvironmentSelector, scopes)
	if err != nil {
		impl.logger.Errorw("error in finding resource qualifier mappings from scope", "scopes", scopes, "err", err)
		return nil, err
	}
	resourceIdVsMappings := make(map[int]resourceQualifiers.ResourceQualifierMappings)
	policyIdVsEnvIdMap := make(map[int]int)
	policyIds := make([]int, 0, len(resourceQualifierMappings))
	for _, mapping := range resourceQualifierMappings {
		resourceIdVsMappings[mapping.Scope.EnvId] = mapping
		policyIdVsEnvIdMap[mapping.ResourceId] = mapping.Scope.EnvId
		policyIds = append(policyIds, mapping.ResourceId)
	}
	policiesMap := make(map[string]*bean.PromotionPolicy)
	rawPolicies, err := impl.globalPolicyDataManager.GetPolicyByIds(policyIds)
	if err != nil {
		impl.logger.Errorw("error in finding policies by ids", "ids", policyIds, "err", err)
		return nil, err
	}

	for _, rawPolicy := range rawPolicies {
		policy := &bean.PromotionPolicy{}
		err = policy.UpdateWithGlobalPolicy(rawPolicy)
		if err != nil {
			impl.logger.Errorw("error in extracting policy from globalPolicy json", "policyId", rawPolicy.Id, "err", err)
			return nil, err
		}
		envId := policyIdVsEnvIdMap[policy.Id]
		resourceQualifierMapping := resourceIdVsMappings[envId]
		policiesMap[resourceQualifierMapping.Scope.SystemMetadata.EnvironmentName] = policy
	}

	return policiesMap, err
}

func (impl ArtifactPromotionDataReadServiceImpl) GetPromotionPolicyByIds(ids []int) ([]*bean.PromotionPolicy, error) {
	globalPolicies, err := impl.globalPolicyDataManager.GetPolicyByIds(ids)
	if err != nil {
		impl.logger.Errorw("error in fetching global policies by ids", "policyids", ids, "err", err)
		return nil, err
	}

	promotionPolicies := make([]*bean.PromotionPolicy, 0, len(globalPolicies))
	for _, globalPolicy := range globalPolicies {
		policy := &bean.PromotionPolicy{}
		err = policy.UpdateWithGlobalPolicy(globalPolicy)
		if err != nil {
			impl.logger.Errorw("error in extracting policy from globalPolicy json", "policyId", globalPolicy.Id, "err", err)
			return nil, err
		}
		promotionPolicies = append(promotionPolicies, policy)
	}
	return promotionPolicies, nil
}

func (impl ArtifactPromotionDataReadServiceImpl) GetPromotionPolicyById(id int) (*bean.PromotionPolicy, error) {
	globalPolicy, err := impl.globalPolicyDataManager.GetPolicyById(id)
	if err != nil {
		impl.logger.Errorw("error in fetching global policy by id", "policyid", id, "err", err)
		return nil, err
	}
	policy := &bean.PromotionPolicy{}
	err = policy.UpdateWithGlobalPolicy(globalPolicy)
	if err != nil {
		impl.logger.Errorw("error in extracting policy from globalPolicy json", "policyId", globalPolicy.Id, "err", err)
		return nil, err
	}
	return policy, nil
}

func (impl ArtifactPromotionDataReadServiceImpl) GetPromotionPolicyByName(name string) (*bean.PromotionPolicy, error) {
	globalPolicy, err := impl.globalPolicyDataManager.GetPolicyByName(name)
	if err != nil {
		impl.logger.Errorw("error in fetching global policy by name", "name", name, "err", err)
		if errors.Is(err, pg.ErrNoRows) {
			return nil, &util.ApiError{
				HttpStatusCode:  http.StatusFound,
				InternalMessage: "policy not found",
				UserMessage:     "policy not found",
			}
		}
		return nil, err
	}

	promotionPolicy := &bean.PromotionPolicy{}
	err = promotionPolicy.UpdateWithGlobalPolicy(globalPolicy)
	if err != nil {
		impl.logger.Errorw("error in extracting policy from globalPolicy json", "policyName", globalPolicy.Name, "err", err)
		return nil, err
	}

	return promotionPolicy, nil
}

func (impl ArtifactPromotionDataReadServiceImpl) GetPoliciesMetadata(policyMetadataRequest bean.PromotionPolicyMetaRequest) (*bean.PromotionPolicyExtraResponse, error) {

	promotionPolicies := make([]*bean.PromotionPolicy, 0)

	sortRequest := impl.parseSortByRequest(policyMetadataRequest)
	globalPolicies, err := impl.globalPolicyDataManager.GetAndSort(policyMetadataRequest.Search, sortRequest)
	if err != nil {
		impl.logger.Errorw("error in fetching global policies by name search string", "policyMetadataRequest", policyMetadataRequest, "err", err)
		return nil, err
	}

	promotionPolicies, err = impl.parsePromotionPolicyFromGlobalPolicy(globalPolicies)
	if err != nil {
		impl.logger.Errorw("error in parsing global policy from promotion policy", "globalPolicy", globalPolicies, "err", err)
		return nil, err
	}

	promotionPolicyExtraResponse := &bean.PromotionPolicyExtraResponse{
		IdentifierCount: len(globalPolicies),
		PromotionPolicy: promotionPolicies,
	}

	return promotionPolicyExtraResponse, nil
}

func (impl ArtifactPromotionDataReadServiceImpl) parseSortByRequest(policyMetadataRequest bean.PromotionPolicyMetaRequest) *bean2.SortByRequest {

	sortRequest := &bean2.SortByRequest{
		SortOrderDesc: policyMetadataRequest.SortOrder == bean.DESC,
	}
	switch policyMetadataRequest.SortBy {
	case bean.POLICY_NAME_SORT_KEY:
		sortRequest.SortByType = bean2.GlobalPolicyColumnField
	case bean.APPROVER_COUNT_SORT_KEY:
		sortRequest.SortByType = bean2.GlobalPolicySearchableField
		sortRequest.SearchableField = util2.SearchableField{
			FieldName: bean.PROMOTION_APPROVAL_PENDING_NODE,
			FieldType: util2.NumericType,
		}
	}
	return sortRequest
}

func (impl ArtifactPromotionDataReadServiceImpl) parsePromotionPolicyFromGlobalPolicy(globalPolicies []*bean2.GlobalPolicyBaseModel) ([]*bean.PromotionPolicy, error) {
	promotionPolicies := make([]*bean.PromotionPolicy, 0)
	for _, rawPolicy := range globalPolicies {
		policy := &bean.PromotionPolicy{}
		err := policy.UpdateWithGlobalPolicy(rawPolicy)
		if err != nil {
			impl.logger.Errorw("error in extracting policy from globalPolicy json", "policyId", rawPolicy.Id, "err", err)
			return nil, err
		}
		promotionPolicies = append(promotionPolicies, policy)
	}
	return promotionPolicies, nil
}
