package artifactPromotion

import (
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/artifactPromotion/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"go.uber.org/zap"
)

type PromotionPolicy interface {
	PromotionPolicyReadService
	PromotionPolicyCUDService
}

type PromotionPolicyReadService interface {
	GetByAppAndEnvId(appId, envId int) (*bean.PromotionPolicy, error)
	GetByAppIdAndEnvIds(appId int, envIds []int) (map[string]*bean.PromotionPolicy, error)
	GetByIds(ids []int) ([]*bean.PromotionPolicy, error)
}

type PromotionPolicyCUDService interface {
	UpdatePolicy(userId int32, policyName string, policyBean *bean.PromotionPolicy) error
	CreatePolicy(userId int32, policyBean *bean.PromotionPolicy) error
	DeletePolicy(userId int32, profileName string) error
	ApplyPolicyToIdentifiers(userId int32, applyIdentifiersRequest interface{}) error
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

func (impl PromotionPolicyServiceImpl) GetByIds(ids []int) ([]*bean.PromotionPolicy, error) {
	return nil, nil
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

func (impl PromotionPolicyServiceImpl) UpdatePolicy(userId int32, policyName string, policyBean *bean.PromotionPolicy) error {

}
func (impl PromotionPolicyServiceImpl) CreatePolicy(userId int32, policyBean *bean.PromotionPolicy) error {

}
func (impl PromotionPolicyServiceImpl) DeletePolicy(userId int32, policyName string) error {

}
func (impl PromotionPolicyServiceImpl) ApplyPolicyToIdentifiers(userId int32, applyIdentifiersRequest interface{}) error {

}
