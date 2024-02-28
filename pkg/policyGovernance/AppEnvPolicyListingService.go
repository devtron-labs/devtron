package policyGovernance

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type AppEnvPolicyListingServiceImpl struct {
	globalPolicyDataManager         globalPolicy.GlobalPolicyDataManager
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService
	pipelineService                 pipeline.CdPipelineConfigService
	logger                          *zap.SugaredLogger
}

func NewAppEnvPolicyListingService(globalPolicyDataManager globalPolicy.GlobalPolicyDataManager,
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService,
	pipelineService pipeline.CdPipelineConfigService,
	logger *zap.SugaredLogger,
) *AppEnvPolicyListingServiceImpl {
	return &AppEnvPolicyListingServiceImpl{
		globalPolicyDataManager:         globalPolicyDataManager,
		resourceQualifierMappingService: resourceQualifierMappingService,
		pipelineService:                 pipelineService,
		logger:                          logger,
	}
}

func (impl AppEnvPolicyListingServiceImpl) ListAppEnvPolicies(listFilter *AppEnvPolicyMappingsListFilter) ([]AppEnvPolicyContainer, int, error) {
	if len(listFilter.PolicyNames) > 0 {
		return impl.listAppEnvPoliciesByPolicyFilter(listFilter)
	} else {
		return impl.listAppEnvPoliciesByEmptyPolicyFilter(listFilter)
	}

}

func (impl AppEnvPolicyListingServiceImpl) listAppEnvPoliciesByPolicyFilter(listFilter *AppEnvPolicyMappingsListFilter) ([]AppEnvPolicyContainer, int, error) {
	noPolicyFilter := false
	validPolicyNames := make([]string, 0, len(listFilter.PolicyNames))
	validPolicyNameMap := make(map[string]bool)
	var policyNames []string
	for _, policyName := range listFilter.PolicyNames {
		if policyName != NO_POLICY {
			validPolicyNames = append(validPolicyNames, policyName)
			validPolicyNameMap[policyName] = true
		} else {
			noPolicyFilter = true
		}
	}
	if !noPolicyFilter {
		policyNames = validPolicyNames
	}
	policies, err := impl.getPolicies(policyNames, listFilter.PolicyType)
	if err != nil {
		return nil, 0, err
	}
	includePolicyIds := make([]int, 0, len(policies))
	includedPoliciesMap := make(map[int]*bean2.GlobalPolicyBaseModel)
	excludePolicyIds := make([]int, 0, len(policies))
	for _, policy := range policies {
		if validPolicyNameMap[policy.Name] {
			includePolicyIds = append(includePolicyIds, policy.Id)
			includedPoliciesMap[policy.Id] = policy
		} else {
			excludePolicyIds = append(excludePolicyIds, policy.Id)
		}
	}

	includeQualifierMappings, err := impl.resourceQualifierMappingService.GetResourceMappingsForResources(resourceQualifiers.ImagePromotionPolicy, includePolicyIds, resourceQualifiers.ApplicationEnvironmentSelector)
	if err != nil {
		impl.logger.Errorw("error in finding the app env mappings using policy ids", "policyIds", includePolicyIds, "err", err)
		return nil, 0, err
	}

	excludeQualifierMappings, err := impl.resourceQualifierMappingService.GetResourceMappingsForResources(resourceQualifiers.ImagePromotionPolicy, excludePolicyIds, resourceQualifiers.ApplicationEnvironmentSelector)
	if err != nil {
		impl.logger.Errorw("error in finding the app env mappings using policy ids", "policyIds", excludePolicyIds, "err", err)
		return nil, 0, err
	}

	appIdEnvIdPolicyMap := make(map[string]*bean2.GlobalPolicyBaseModel)
	includeAppEnvIds := make([]string, len(includeQualifierMappings))
	excludeAppEnvIds := make([]string, len(includeQualifierMappings))
	for _, includeQualifierMapping := range includeQualifierMappings {
		key := fmt.Sprintf("%d,%d", includeQualifierMapping.Scope.AppId, includeQualifierMapping.Scope.AppId)
		appIdEnvIdPolicyMap[key] = includedPoliciesMap[includeQualifierMapping.ResourceId]
		includeAppEnvIds = append(includeAppEnvIds, key)
	}

	for _, excludeQualifierMapping := range excludeQualifierMappings {
		key := fmt.Sprintf("%d,%d", excludeQualifierMapping.Scope.AppId, excludeQualifierMapping.Scope.AppId)
		excludeAppEnvIds = append(excludeAppEnvIds, key)
	}
	filter := bean3.CdPipelineListFilter{
		SortOrder:        listFilter.SortOrder,
		SortBy:           listFilter.SortBy,
		Limit:            listFilter.Size,
		Offset:           listFilter.Offset,
		IncludeAppEnvIds: includeAppEnvIds,
		ExcludeAppEnvIds: excludeAppEnvIds,
		EnvNames:         listFilter.EnvNames,
		AppNames:         listFilter.AppNames,
	}
	totalCount := 0
	paginatedAppEnvData, err := impl.pipelineService.FindAppAndEnvDetailsByListFilter(filter)
	if err != nil {
		impl.logger.Errorw("error in fetching the paginated app environment list using filter", "filter", filter, "err", err)
		return nil, 0, err
	}
	result := lo.Map(paginatedAppEnvData, func(cdPipMeta bean3.CdPipelineMetaData, i int) AppEnvPolicyContainer {
		totalCount = cdPipMeta.TotalCount
		key := fmt.Sprintf("%d,%d", cdPipMeta.AppId, cdPipMeta.EnvId)
		policyName := appIdEnvIdPolicyMap[key].Name
		return AppEnvPolicyContainer{
			AppId:      cdPipMeta.AppId,
			EnvId:      cdPipMeta.EnvId,
			AppName:    cdPipMeta.AppName,
			EnvName:    cdPipMeta.EnvironmentName,
			PolicyName: policyName,
		}
	})

	return result, totalCount, nil
}

func (impl AppEnvPolicyListingServiceImpl) listAppEnvPoliciesByEmptyPolicyFilter(listFilter *AppEnvPolicyMappingsListFilter) ([]AppEnvPolicyContainer, int, error) {
	filter := bean3.CdPipelineListFilter{
		SortOrder: listFilter.SortOrder,
		SortBy:    listFilter.SortBy,
		Limit:     listFilter.Size,
		Offset:    listFilter.Offset,
		EnvNames:  listFilter.EnvNames,
		AppNames:  listFilter.AppNames,
	}
	paginatedAppEnvData, err := impl.pipelineService.FindAppAndEnvDetailsByListFilter(filter)
	if err != nil {
		impl.logger.Errorw("error in fetching the paginated app environment list using filter", "filter", filter, "err", err)
		return nil, 0, err
	}
	scopes := make([]*resourceQualifiers.Scope, 0, len(paginatedAppEnvData))
	for _, appEnv := range paginatedAppEnvData {
		scopes = append(scopes, &resourceQualifiers.Scope{
			AppId: appEnv.AppId,
			EnvId: appEnv.EnvId,
		})
	}

	qualifierMappings, err := impl.resourceQualifierMappingService.GetResourceMappingsForScopes(resourceQualifiers.ImagePromotionPolicy, resourceQualifiers.ApplicationEnvironmentSelector, scopes)
	if err != nil {
		impl.logger.Errorw("error in finding the app env mappings using scopes", "scopes", scopes, "policyType", resourceQualifiers.ImagePromotionPolicy, "qualifierSelector", resourceQualifiers.ApplicationEnvironmentSelector, "err", err)
		return nil, 0, err
	}

	appEnvPolicyMap := make(map[string]int)
	policyIds := make([]int, 0, len(qualifierMappings))
	for _, qualifierMapping := range qualifierMappings {
		policyIds = append(policyIds, qualifierMapping.ResourceId)
		appEnvKey := fmt.Sprintf("%d,%d", qualifierMapping.Scope.AppId, qualifierMapping.Scope.EnvId)
		appEnvPolicyMap[appEnvKey] = qualifierMapping.ResourceId
	}

	policies, err := impl.globalPolicyDataManager.GetPolicyByIds(policyIds)
	if err != nil {
		impl.logger.Errorw("error in finding the profiles with ids", "policyIds", policyIds, "err", err)
		return nil, 0, err
	}

	policyMap := make(map[int]string)
	for _, policy := range policies {
		policyMap[policy.Id] = policy.Name
	}
	totalCount := 0
	result := lo.Map(paginatedAppEnvData, func(cdPipMeta bean3.CdPipelineMetaData, i int) AppEnvPolicyContainer {
		totalCount = cdPipMeta.TotalCount
		key := fmt.Sprintf("%d,%d", cdPipMeta.AppId, cdPipMeta.EnvId)
		policyId := appEnvPolicyMap[key]
		policyName := policyMap[policyId]
		return AppEnvPolicyContainer{
			AppId:      cdPipMeta.AppId,
			EnvId:      cdPipMeta.EnvId,
			AppName:    cdPipMeta.AppName,
			EnvName:    cdPipMeta.EnvironmentName,
			PolicyName: policyName,
		}
	})

	return result, totalCount, nil
}

func (impl AppEnvPolicyListingServiceImpl) getPolicies(policyNames []string, policyType bean2.GlobalPolicyType) ([]*bean2.GlobalPolicyBaseModel, error) {
	if len(policyNames) == 0 {
		policies, err := impl.globalPolicyDataManager.GetAllActiveByType(policyType)
		if err != nil {
			impl.logger.Errorw("error in finding the all the active promotion policies with names", "err", err)
			return policies, err
		}
	}
	policies, err := impl.globalPolicyDataManager.GetPolicyByNames(policyNames)
	if err != nil {
		impl.logger.Errorw("error in finding the profiles with names", "profileNames", policyNames, "err", err)
		return policies, err
	}
	return policies, err
}
