package read

import (
	"encoding/json"
	"errors"
	"fmt"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appWorkflow/read"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	repository3 "github.com/devtron-labs/devtron/pkg/globalPolicy/repository"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/constants"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/team"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
)

type ArtifactPromotionDataReadService interface {
	FetchPromotionApprovalDataForArtifacts(artifactIds []int, pipelineId int, status constants.ArtifactPromotionRequestStatus) (map[int]*bean.PromotionApprovalMetaData, error)
	GetPromotionPolicyByAppAndEnvId(appId, envId int) (*bean.PromotionPolicy, error)
	GetPromotionPolicyByAppAndEnvIds(ctx *util2.RequestCtx, appId int, envIds []int) (map[string]*bean.PromotionPolicy, error)
	GetPromotionPolicyById(ctx *util2.RequestCtx, id int) (*bean.PromotionPolicy, error)
	GetPromotionPolicyByIds(ctx *util2.RequestCtx, ids []int) ([]*bean.PromotionPolicy, error)
	GetPromotionPolicyByName(ctx *util2.RequestCtx, name string) (*bean.PromotionPolicy, error)
	GetPoliciesMetadata(ctx *util2.RequestCtx, policyMetadataRequest bean.PromotionPolicyMetaRequest) ([]*bean.PromotionPolicy, error)
	GetAllPoliciesNameForAutocomplete(ctx *util2.RequestCtx) ([]string, error)
	GetPromotionPendingRequestMapping(ctx *util2.RequestCtx, pipelineIds []int) (map[int]int, error)
	GetPolicyHistoryIdsByPolicyIds(policyIds []int) ([]int, error)
}

type ArtifactPromotionDataReadServiceImpl struct {
	logger                                     *zap.SugaredLogger
	artifactPromotionApprovalRequestRepository repository.RequestRepository
	requestApprovalUserdataRepo                pipelineConfig.RequestApprovalUserdataRepository
	userService                                user.UserService
	pipelineRepository                         pipelineConfig.PipelineRepository
	resourceQualifierMappingService            resourceQualifiers.QualifierMappingService
	resourceFilterEvaluationAuditService       resourceFilter.FilterEvaluationAuditService
	globalPolicyDataManager                    globalPolicy.GlobalPolicyDataManager
	appWorkflowDataReadService                 read.AppWorkflowDataReadService
	teamService                                team.TeamService
}

func NewArtifactPromotionDataReadServiceImpl(
	ArtifactPromotionApprovalRequestRepository repository.RequestRepository,
	logger *zap.SugaredLogger,
	requestApprovalUserdataRepo pipelineConfig.RequestApprovalUserdataRepository,
	userService user.UserService,
	pipelineRepository pipelineConfig.PipelineRepository,
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService,
	globalPolicyDataManager globalPolicy.GlobalPolicyDataManager,
	appWorkflowDataReadService read.AppWorkflowDataReadService,
	teamService team.TeamService,
	resourceFilterEvaluationAuditService resourceFilter.FilterEvaluationAuditService,
) *ArtifactPromotionDataReadServiceImpl {
	return &ArtifactPromotionDataReadServiceImpl{
		artifactPromotionApprovalRequestRepository: ArtifactPromotionApprovalRequestRepository,
		logger:                               logger,
		requestApprovalUserdataRepo:          requestApprovalUserdataRepo,
		userService:                          userService,
		pipelineRepository:                   pipelineRepository,
		resourceQualifierMappingService:      resourceQualifierMappingService,
		resourceFilterEvaluationAuditService: resourceFilterEvaluationAuditService,
		globalPolicyDataManager:              globalPolicyDataManager,
		appWorkflowDataReadService:           appWorkflowDataReadService,
		teamService:                          teamService,
	}
}

func (impl *ArtifactPromotionDataReadServiceImpl) FetchPromotionApprovalDataForArtifacts(artifactIds []int, pipelineId int, status constants.ArtifactPromotionRequestStatus) (map[int]*bean.PromotionApprovalMetaData, error) {

	promotionApprovalMetadata := make(map[int]*bean.PromotionApprovalMetaData)

	promotionApprovalRequest, err := impl.artifactPromotionApprovalRequestRepository.FindByPipelineIdAndArtifactIds(pipelineId, artifactIds, status)
	if err != nil {
		impl.logger.Errorw("error in fetching promotion request for given pipelineId and artifactId", "pipelineId", pipelineId, "artifactIds", artifactIds, "err", err)
		return promotionApprovalMetadata, err
	}

	if len(promotionApprovalRequest) > 0 {

		var requestedUserIds []int32
		var approvalRequestIds []int

		for _, approvalRequest := range promotionApprovalRequest {
			requestedUserIds = append(requestedUserIds, approvalRequest.CreatedBy)
			approvalRequestIds = append(approvalRequestIds, approvalRequest.Id)
		}

		requestIdToApprovalUserDataMapping, err := impl.getPromotionApprovalUserMetadata(approvalRequestIds)
		if err != nil {
			impl.logger.Errorw("error in fetching approval user data mapping for approval request ids", "approvalRequestIds", approvalRequestIds, "err", err)
			return promotionApprovalMetadata, err
		}

		requestedUserInfoMap, err := impl.getUserInfoMap(requestedUserIds)
		if err != nil {
			impl.logger.Errorw("error in getting user info map by user ids", "requestedUserIds", requestedUserIds, "err", err)
			return promotionApprovalMetadata, err
		}

		pipeline, err := impl.pipelineRepository.FindById(pipelineId)
		if err != nil {
			impl.logger.Errorw("error in fetching pipeline by id", "pipelineId", pipelineId, "err", err)
			return promotionApprovalMetadata, err
		}

		requestIdToPolicyMapping, err := impl.getRequestIdToPolicyMapping(promotionApprovalRequest)
		if err != nil {
			return promotionApprovalMetadata, err
		}

		for _, approvalRequest := range promotionApprovalRequest {
			approvalMetadata := impl.getPromotionApprovalMetadata(approvalRequest, pipeline.Environment.Name, requestedUserInfoMap, requestIdToApprovalUserDataMapping, requestIdToPolicyMapping)
			promotionApprovalMetadata[approvalRequest.ArtifactId] = approvalMetadata
		}
	}
	return promotionApprovalMetadata, nil
}

func (impl *ArtifactPromotionDataReadServiceImpl) getRequestIdToPolicyMapping(promotionApprovalRequest []*repository.ArtifactPromotionApprovalRequest) (map[int]*bean.PromotionPolicy, error) {
	policyEvaluationIds := util2.GetArrayObject(promotionApprovalRequest, func(service *repository.ArtifactPromotionApprovalRequest) int {
		return service.PolicyEvaluationAuditId
	})

	policyAuditHistoryIds := make([]int, 0, len(policyEvaluationIds))

	filterEvaluationAudits, err := impl.resourceFilterEvaluationAuditService.GetByIds(policyEvaluationIds)
	if err != nil {
		impl.logger.Errorw("error in fetching policy filter evaluation audits using policyEvaluationIds", "policyEvaluationIds", policyEvaluationIds, "err", err)
		return nil, err
	}
	for _, filterEvaluationAudit := range filterEvaluationAudits {
		historyObj := resourceFilter.FilterHistoryObject{}
		err = json.Unmarshal([]byte(filterEvaluationAudit.FilterHistoryObjects), &historyObj)
		if err != nil {
			impl.logger.Errorw("error in un-marshaling policy evaluation audits filterHistory object", "filterHistoryObjects", filterEvaluationAudit.FilterHistoryObjects, "err", err)
			return nil, err
		}
		policyAuditHistoryIds = append(policyAuditHistoryIds, historyObj.FilterHistoryId)
	}

	rawPolicies, err := impl.globalPolicyDataManager.GetPoliciesByHistoryIds(policyAuditHistoryIds)
	if err != nil {
		impl.logger.Errorw("error in fetching global policies by policyEvaluationIds", "policyEvaluationIds", policyEvaluationIds, "err", err)
		return nil, err
	}

	globalPolicies, err := impl.parsePromotionPolicyFromGlobalPolicy(rawPolicies)
	if err != nil {
		impl.logger.Errorw("error in parsing promotion policy from global policy", "err", err)
		return nil, err
	}

	policyIdToPolicyMapping := util2.GetIdToObjectMap(globalPolicies, func(policy *bean.PromotionPolicy) int {
		return policy.Id
	})

	requestIdToPolicyMapping := make(map[int]*bean.PromotionPolicy)
	for _, approvalRequest := range promotionApprovalRequest {
		requestIdToPolicyMapping[approvalRequest.Id] = policyIdToPolicyMapping[approvalRequest.PolicyId]
	}
	return requestIdToPolicyMapping, nil
}

func (impl *ArtifactPromotionDataReadServiceImpl) getPromotionApprovalMetadata(approvalRequest *repository.ArtifactPromotionApprovalRequest, envName string, userInfoMap map[int]bean3.UserInfo, requestIdToApprovalUserDataMapping map[int][]*pipelineConfig.RequestApprovalUserData, requestIdToPolicyMapping map[int]*bean.PromotionPolicy) *bean.PromotionApprovalMetaData {
	promotedFrom := impl.getSource(approvalRequest, envName)

	policy := requestIdToPolicyMapping[approvalRequest.Id]

	approvalMetadata := &bean.PromotionApprovalMetaData{
		ApprovalRequestId:    approvalRequest.Id,
		ApprovalRuntimeState: approvalRequest.Status,
		PromotedFrom:         promotedFrom,
		PromotedFromType:     string(approvalRequest.SourceType.GetSourceTypeStr()),
	}

	// don't want to block if policy not found.
	if policy != nil {
		approvalMetadata.Policy = *policy
	}

	if userInfo, ok := userInfoMap[int(approvalRequest.CreatedBy)]; ok {
		approvalMetadata.RequestedUserData = bean.PromotionApprovalUserData{
			UserId:         userInfo.Id,
			UserEmail:      userInfo.EmailId,
			UserActionTime: approvalRequest.CreatedOn,
		}
	}

	promotionApprovalUserData := requestIdToApprovalUserDataMapping[approvalRequest.Id]
	for _, approvalUserData := range promotionApprovalUserData {
		approvalMetadata.ApprovalUsersData = append(approvalMetadata.ApprovalUsersData, bean.PromotionApprovalUserData{
			UserId:         approvalUserData.UserId,
			UserEmail:      approvalUserData.User.EmailId,
			UserActionTime: approvalUserData.CreatedOn,
		})
	}

	return approvalMetadata
}

func (impl *ArtifactPromotionDataReadServiceImpl) getPromotionPolicy(policyId int) (bean.PromotionPolicy, error) {
	globalPolicyData, err := impl.globalPolicyDataManager.GetPolicyById(policyId)
	if err != nil {
		impl.logger.Errorw("error in fetching globalPolicy by id", "globalPolicyId", policyId, "err", err)
		return bean.PromotionPolicy{}, err
	}

	policy := bean.PromotionPolicy{}
	err = policy.UpdateWithGlobalPolicy(globalPolicyData)
	if err != nil {
		impl.logger.Errorw("error in parsing promotion policy from globalPolicy", "err", err)
		return bean.PromotionPolicy{}, err
	}
	return policy, nil
}

func (impl *ArtifactPromotionDataReadServiceImpl) getPromotionApprovalUserMetadata(approvalRequestIds []int) (map[int][]*pipelineConfig.RequestApprovalUserData, error) {
	requestIdToApprovalUserDataMapping := make(map[int][]*pipelineConfig.RequestApprovalUserData)
	promotionApprovalUserDataArray, err := impl.requestApprovalUserdataRepo.FetchApprovalDataForRequests(approvalRequestIds, repository2.ARTIFACT_PROMOTION_APPROVAL)
	if err != nil {
		impl.logger.Errorw("error in getting promotionApprovalUserData", "err", err, "promotionApprovalRequestIds", approvalRequestIds)
		return requestIdToApprovalUserDataMapping, err
	}
	for _, promotionApprovalUserData := range promotionApprovalUserDataArray {
		requestIdToApprovalUserDataMapping[promotionApprovalUserData.ApprovalRequestId] = append(
			requestIdToApprovalUserDataMapping[promotionApprovalUserData.ApprovalRequestId],
			promotionApprovalUserData)
	}
	return requestIdToApprovalUserDataMapping, nil
}

func (impl *ArtifactPromotionDataReadServiceImpl) getUserInfoMap(requestedUserIds []int32) (map[int]bean3.UserInfo, error) {
	userInfos, err := impl.userService.GetByIds(requestedUserIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching users", "requestedUserIds", requestedUserIds, "err", err)
		return nil, err
	}
	userInfoMap := util2.GetIdToObjectMap(userInfos, func(info bean3.UserInfo) int { return int(info.Id) })

	return userInfoMap, nil
}

func (impl *ArtifactPromotionDataReadServiceImpl) getSource(approvalRequest *repository.ArtifactPromotionApprovalRequest, envName string) string {

	if approvalRequest.SourceType == constants.CD {
		return envName
	}
	return string(approvalRequest.SourceType.GetSourceTypeStr())
}

func (impl *ArtifactPromotionDataReadServiceImpl) GetPromotionPolicyByAppAndEnvId(appId, envId int) (*bean.PromotionPolicy, error) {

	scope := &resourceQualifiers.SelectionIdentifier{AppId: appId, EnvId: envId}
	//
	qualifierMapping, err := impl.resourceQualifierMappingService.GetResourceMappingsForSelections(
		resourceQualifiers.ImagePromotionPolicy,
		resourceQualifiers.ApplicationEnvironmentSelector,
		[]*resourceQualifiers.SelectionIdentifier{scope},
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

func (impl *ArtifactPromotionDataReadServiceImpl) GetPromotionPolicyByAppAndEnvIds(ctx *util2.RequestCtx, appId int, envIds []int) (map[string]*bean.PromotionPolicy, error) {
	scopes := make([]*resourceQualifiers.SelectionIdentifier, 0, len(envIds))
	for _, envId := range envIds {
		scopes = append(scopes, &resourceQualifiers.SelectionIdentifier{
			AppId: appId,
			EnvId: envId,
		})
	}

	if len(scopes) == 0 {
		return nil, nil
	}

	resourceQualifierMappings, err := impl.resourceQualifierMappingService.GetResourceMappingsForSelections(resourceQualifiers.ImagePromotionPolicy, resourceQualifiers.ApplicationEnvironmentSelector, scopes)
	if err != nil {
		impl.logger.Errorw("error in finding resource qualifier mappings from scope", "scopes", scopes, "err", err)
		return nil, err
	}
	policyIds := make([]int, 0, len(resourceQualifierMappings))
	for _, mapping := range resourceQualifierMappings {
		policyIds = append(policyIds, mapping.ResourceId)
	}
	rawPolicies, err := impl.globalPolicyDataManager.GetPolicyByIds(policyIds)
	if err != nil {
		impl.logger.Errorw("error in finding policies by ids", "ids", policyIds, "err", err)
		return nil, err
	}

	policyIdVsPolicyMap := make(map[int]*bean.PromotionPolicy)
	envVsPolicyMap := make(map[string]*bean.PromotionPolicy)
	for _, rawPolicy := range rawPolicies {
		policy := &bean.PromotionPolicy{}
		err = policy.UpdateWithGlobalPolicy(rawPolicy)
		if err != nil {
			impl.logger.Errorw("error in extracting policy from globalPolicy json", "policyId", rawPolicy.Id, "err", err)
			return nil, err
		}
		policyIdVsPolicyMap[policy.Id] = policy
	}

	for _, mapping := range resourceQualifierMappings {
		policy := policyIdVsPolicyMap[mapping.ResourceId]
		if policy == nil {
			continue
		}
		envVsPolicyMap[mapping.SelectionIdentifier.SelectionIdentifierName.EnvironmentName] = policy
		policyIds = append(policyIds, mapping.ResourceId)
	}
	return envVsPolicyMap, err
}

func (impl *ArtifactPromotionDataReadServiceImpl) GetPromotionPolicyByIds(ctx *util2.RequestCtx, ids []int) ([]*bean.PromotionPolicy, error) {
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

func (impl *ArtifactPromotionDataReadServiceImpl) GetPromotionPolicyById(ctx *util2.RequestCtx, id int) (*bean.PromotionPolicy, error) {
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

func (impl *ArtifactPromotionDataReadServiceImpl) GetPromotionPolicyByName(ctx *util2.RequestCtx, name string) (*bean.PromotionPolicy, error) {
	globalPolicy, err := impl.globalPolicyDataManager.GetPolicyByName(name, bean2.GLOBAL_POLICY_TYPE_IMAGE_PROMOTION_POLICY)
	if err != nil {
		impl.logger.Errorw("error in fetching global policy by name", "name", name, "err", err)
		if errors.Is(err, pg.ErrNoRows) {
			return nil, &util.ApiError{
				HttpStatusCode:  http.StatusFound,
				InternalMessage: constants.PolicyNotFoundErr,
				UserMessage:     constants.PolicyNotFoundErr,
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

func (impl *ArtifactPromotionDataReadServiceImpl) GetPoliciesMetadata(ctx *util2.RequestCtx, policyMetadataRequest bean.PromotionPolicyMetaRequest) ([]*bean.PromotionPolicy, error) {

	promotionPolicies := make([]*bean.PromotionPolicy, 0)

	sortRequest := impl.parseSortByRequest(policyMetadataRequest)
	globalPolicies, err := impl.globalPolicyDataManager.GetPolicyByCriteria(policyMetadataRequest.Search, sortRequest)
	if err != nil {
		impl.logger.Errorw("error in fetching global policies by name search string", "policyMetadataRequest", policyMetadataRequest, "err", err)
		return nil, err
	}

	promotionPolicies, err = impl.parsePromotionPolicyFromGlobalPolicy(globalPolicies)
	if err != nil {
		impl.logger.Errorw("error in parsing global policy from promotion policy", "globalPolicy", globalPolicies, "err", err)
		return nil, err
	}

	policyIds := util2.Map(promotionPolicies, func(policy *bean.PromotionPolicy) int {
		return policy.Id
	})

	qualifierMappings, err := impl.resourceQualifierMappingService.GetResourceMappingsForResources(resourceQualifiers.ImagePromotionPolicy, policyIds, resourceQualifiers.ApplicationEnvironmentSelector)
	if err != nil {
		impl.logger.Errorw("error in finding the app env mappings using policy ids", "policyIds", policyIds, "err", err)
		return nil, err
	}
	uniquePolicyInAppCount := make(map[string]int, 0)
	policyIdToAppIdMapping := make(map[int]int, 0)
	for _, qualifierMapping := range qualifierMappings {
		uniqueKey := fmt.Sprintf("%d-%d", qualifierMapping.SelectionIdentifier.AppId, qualifierMapping.ResourceId)
		count := uniquePolicyInAppCount[uniqueKey] + 1
		uniquePolicyInAppCount[uniqueKey] = count
		policyIdToAppIdMapping[qualifierMapping.ResourceId] = qualifierMapping.SelectionIdentifier.AppId
	}

	for _, promotionPolicy := range promotionPolicies {
		identifierCount := 0
		if appId, ok := policyIdToAppIdMapping[promotionPolicy.Id]; ok {
			uniqueKey := fmt.Sprintf("%d-%d", appId, promotionPolicy.Id)
			identifierCount = uniquePolicyInAppCount[uniqueKey]
		}
		promotionPolicy.IdentifierCount = &identifierCount
	}

	return promotionPolicies, nil
}

func (impl *ArtifactPromotionDataReadServiceImpl) parseSortByRequest(policyMetadataRequest bean.PromotionPolicyMetaRequest) *bean2.SortByRequest {

	sortRequest := &bean2.SortByRequest{
		SortOrderDesc: policyMetadataRequest.SortOrder == constants.DESC,
	}
	switch policyMetadataRequest.SortBy {
	case constants.POLICY_NAME_SORT_KEY:
		sortRequest.SortByType = bean2.GlobalPolicyColumnField
	case constants.APPROVER_COUNT_SORT_KEY:
		sortRequest.SortByType = bean2.GlobalPolicySearchableField
		sortRequest.SearchableField = util2.SearchableField{
			FieldName: string(constants.APPROVER_COUNT_SEARCH_FIELD),
			FieldType: util2.NumericType,
		}
	}
	return sortRequest
}

func (impl *ArtifactPromotionDataReadServiceImpl) parsePromotionPolicyFromGlobalPolicy(globalPolicies []*bean2.GlobalPolicyBaseModel) ([]*bean.PromotionPolicy, error) {
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

func (impl *ArtifactPromotionDataReadServiceImpl) GetAllPoliciesNameForAutocomplete(ctx *util2.RequestCtx) ([]string, error) {
	policyNames := make([]string, 0)
	promotionPolicies, err := impl.globalPolicyDataManager.GetAllActivePoliciesByType(bean2.GLOBAL_POLICY_TYPE_IMAGE_PROMOTION_POLICY)
	if err != nil {
		impl.logger.Errorw("error in getting all global policies by type", "policyType", bean2.GLOBAL_POLICY_TYPE_IMAGE_PROMOTION_POLICY, "err", err)
		return policyNames, err
	}
	policyNames = util2.Map(promotionPolicies, func(policy *repository3.GlobalPolicy) string {
		return policy.Name
	})
	return policyNames, nil
}

func (impl *ArtifactPromotionDataReadServiceImpl) GetPromotionPendingRequestMapping(ctx *util2.RequestCtx, pipelineIds []int) (map[int]int, error) {
	pipelineIdToCountMap := make(map[int]int)
	requests, err := impl.artifactPromotionApprovalRequestRepository.FindPendingByDestinationPipelineIds(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error in finding pending requests on pipeline", "pipelineIds", pipelineIds, "err", err)
		return pipelineIdToCountMap, err
	}
	for _, request := range requests {
		pipelineIdToCountMap[request.DestinationPipelineId] = pipelineIdToCountMap[request.DestinationPipelineId] + 1
	}
	return pipelineIdToCountMap, nil
}

func (impl *ArtifactPromotionDataReadServiceImpl) GetPolicyHistoryIdsByPolicyIds(policyIds []int) ([]int, error) {
	return impl.globalPolicyDataManager.GetPolicyHistoryIdsByPolicyIds(policyIds)
}
