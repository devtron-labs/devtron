package artifactPromotion

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type PromotionPolicy interface {
	PromotionPolicyReadService
	PromotionPolicyCUDService
}

type PromotionPolicyReadService interface {
	GetByAppAndEnvId(appId, envId int) (*bean.PromotionPolicy, error)
	GetByAppIdAndEnvIds(appId int, envIds []int) (map[string]*bean.PromotionPolicy, error)
	GetByIds(ids []int) ([]*bean.PromotionPolicy, error)
	GetPoliciesMetadata(policyMetadataRequest bean.PromotionPolicyMetaRequest) ([]*bean.PromotionPolicy, error)
}

type PromotionPolicyCUDService interface {
	UpdatePolicy(userId int32, policyName string, policyBean *bean.PromotionPolicy) error
	CreatePolicy(userId int32, policyBean *bean.PromotionPolicy) error
	DeletePolicy(userId int32, profileName string) error
	ApplyPolicyToIdentifiers(userId int32, applyIdentifiersRequest bean.BulkPromotionPolicyApplyRequest) error
}

type PromotionPolicyServiceImpl struct {
	globalPolicyDataManager         globalPolicy.GlobalPolicyDataManager
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService
	logger                          *zap.SugaredLogger
}

func NewPromotionPolicyServiceImpl(globalPolicyDataManager globalPolicy.GlobalPolicyDataManager,
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService,
	logger *zap.SugaredLogger,
) *PromotionPolicyServiceImpl {
	return &PromotionPolicyServiceImpl{
		globalPolicyDataManager:         globalPolicyDataManager,
		resourceQualifierMappingService: resourceQualifierMappingService,
		logger:                          logger,
	}
}

func (impl PromotionPolicyServiceImpl) GetByAppAndEnvId(appId, envId int) (*bean.PromotionPolicy, error) {

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

func (impl PromotionPolicyServiceImpl) GetByAppIdAndEnvIds(appId int, envIds []int) (map[string]*bean.PromotionPolicy, error) {
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

func (impl PromotionPolicyServiceImpl) GetByIds(ids []int) ([]*bean.PromotionPolicy, error) {
	return nil, nil
}

func (impl PromotionPolicyServiceImpl) GetPoliciesMetadata(policyMetadataRequest bean.PromotionPolicyMetaRequest) ([]*bean.PromotionPolicy, error) {
	return nil, nil
}

func (impl PromotionPolicyServiceImpl) UpdatePolicy(userId int32, policyName string, policyBean *bean.PromotionPolicy) error {
	globalPolicyDataModel, err := policyBean.ConvertToGlobalPolicyDataModel(userId)
	if err != nil {
		impl.logger.Errorw("error in create policy, not able to convert promotion policy object to global policy data model", "policyBean", policyBean, "err", err)
		return err
	}

	_, err = impl.globalPolicyDataManager.UpdatePolicyByName(policyName, globalPolicyDataModel, nil)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), bean2.UniqueActiveNameConstraint) {
			err = errors.New("policy name already exists, err: duplicate name")
			statusCode = http.StatusConflict
		}
		return &util.ApiError{
			HttpStatusCode:  statusCode,
			InternalMessage: err.Error(),
			UserMessage:     err.Error(),
		}
	}
	return nil
}

func (impl PromotionPolicyServiceImpl) CreatePolicy(userId int32, policyBean *bean.PromotionPolicy) error {
	globalPolicyDataModel, err := policyBean.ConvertToGlobalPolicyDataModel(userId)
	if err != nil {
		impl.logger.Errorw("error in create policy, not able to convert promotion policy object to global policy data model", "policyBean", policyBean, "err", err)
		return err
	}

	_, err = impl.globalPolicyDataManager.CreatePolicy(globalPolicyDataModel, nil)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), bean2.UniqueActiveNameConstraint) {
			err = errors.New("policy name already exists, err: duplicate name")
			statusCode = http.StatusConflict
		}
		return &util.ApiError{
			HttpStatusCode:  statusCode,
			InternalMessage: err.Error(),
			UserMessage:     err.Error(),
		}
	}
	return nil
}

func (impl PromotionPolicyServiceImpl) DeletePolicy(userId int32, policyName string) error {
	err := impl.globalPolicyDataManager.DeletePolicyByName(policyName, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting the promotion policy using name", "policyName", policyName, "userId", userId, "err", err)
	}
	return err
}

func (impl PromotionPolicyServiceImpl) ApplyPolicyToIdentifiers(userId int32, applyIdentifiersRequest bean.BulkPromotionPolicyApplyRequest) error {
	updateToProfile, err := impl.globalPolicyDataManager.GetPolicyByName(applyIdentifiersRequest.ApplyToPolicyName)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, pg.ErrNoRows) {
			err = errors.New(fmt.Sprintf("promotion policy with name '%s' does not exist", applyIdentifiersRequest.ApplyToPolicyName))
			statusCode = http.StatusConflict
		}
		return &util.ApiError{
			HttpStatusCode:  statusCode,
			InternalMessage: err.Error(),
			UserMessage:     err.Error(),
		}
	}
	if len(applyIdentifiersRequest.ApplicationEnvironments) > 0 {

	}
	if applyIdentifiersRequest.AppEnvPolicyListFilter != nil {

	}
}
