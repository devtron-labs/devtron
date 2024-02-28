package policyGovernance

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"net/http"
)

type CommonPolicyActionsService interface {
	ListAppEnvPolicies(listFilter *AppEnvPolicyMappingsListFilter) ([]AppEnvPolicyContainer, int, error)
	ApplyPolicyToIdentifiers(userId int32, applyIdentifiersRequest BulkPromotionPolicyApplyRequest) error
}
type CommonPolicyActionsServiceImpl struct {
	globalPolicyDataManager         globalPolicy.GlobalPolicyDataManager
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService
	pipelineService                 pipeline.CdPipelineConfigService
	appService                      app.AppService
	environmentService              cluster.EnvironmentService
	logger                          *zap.SugaredLogger
}

func NewCommonPolicyActionsService(globalPolicyDataManager globalPolicy.GlobalPolicyDataManager,
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService,
	pipelineService pipeline.CdPipelineConfigService,
	appService app.AppService,
	environmentService cluster.EnvironmentService,
	logger *zap.SugaredLogger,
) *CommonPolicyActionsServiceImpl {
	return &CommonPolicyActionsServiceImpl{
		globalPolicyDataManager:         globalPolicyDataManager,
		resourceQualifierMappingService: resourceQualifierMappingService,
		pipelineService:                 pipelineService,
		logger:                          logger,
		appService:                      appService,
		environmentService:              environmentService,
	}
}

func (impl CommonPolicyActionsServiceImpl) ListAppEnvPolicies(listFilter *AppEnvPolicyMappingsListFilter) ([]AppEnvPolicyContainer, int, error) {
	if len(listFilter.PolicyNames) > 0 {
		return impl.listAppEnvPoliciesByPolicyFilter(listFilter)
	} else {
		return impl.listAppEnvPoliciesByEmptyPolicyFilter(listFilter)
	}

}

func (impl CommonPolicyActionsServiceImpl) listAppEnvPoliciesByPolicyFilter(listFilter *AppEnvPolicyMappingsListFilter) ([]AppEnvPolicyContainer, int, error) {
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

func (impl CommonPolicyActionsServiceImpl) listAppEnvPoliciesByEmptyPolicyFilter(listFilter *AppEnvPolicyMappingsListFilter) ([]AppEnvPolicyContainer, int, error) {
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

func (impl CommonPolicyActionsServiceImpl) getPolicies(policyNames []string, policyType bean2.GlobalPolicyType) ([]*bean2.GlobalPolicyBaseModel, error) {
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

func (impl CommonPolicyActionsServiceImpl) ApplyPolicyToIdentifiers(userId int32, applyIdentifiersRequest BulkPromotionPolicyApplyRequest) error {
	updateToPolicy, err := impl.globalPolicyDataManager.GetPolicyByName(applyIdentifiersRequest.ApplyToPolicyName)
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

	var scopes []*resourceQualifiers.Scope
	if len(applyIdentifiersRequest.ApplicationEnvironments) > 0 {
		scopes, err = impl.fetchScopesByAppEnvNames(applyIdentifiersRequest.ApplicationEnvironments)
		if err != nil {
			impl.logger.Errorw("error in fetching scope objects using appEnv names", "appEnvNames", applyIdentifiersRequest.ApplicationEnvironments, "err", err)
			return err
		}
	}

	if applyIdentifiersRequest.AppEnvPolicyListFilter != nil {
		appEnvPolicyContainers, _, err := impl.ListAppEnvPolicies(applyIdentifiersRequest.AppEnvPolicyListFilter)
		if err != nil {
			impl.logger.Errorw("error in listing application environment policies list using listing filter", "appEnvNames", applyIdentifiersRequest.ApplicationEnvironments, "err", err)
			return err
		}
		scopes = make([]*resourceQualifiers.Scope, 0, len(appEnvPolicyContainers))
		for _, appEnvPolicyContainer := range appEnvPolicyContainers {
			scopes = append(scopes, &resourceQualifiers.Scope{
				AppId: appEnvPolicyContainer.AppId,
				EnvId: appEnvPolicyContainer.EnvId,
				SystemMetadata: &resourceQualifiers.SystemMetadata{
					AppName:         appEnvPolicyContainer.AppName,
					EnvironmentName: appEnvPolicyContainer.EnvName,
				},
			})
		}
	}

	tx, err := impl.resourceQualifierMappingService.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction while bulk applying policies to selected app env entities", "requestPayload", applyIdentifiersRequest, "err", err)
		return err
	}
	defer impl.resourceQualifierMappingService.RollbackTx(tx)
	// delete all the existing mappings for the updateToProfile.Id resource
	err = impl.resourceQualifierMappingService.DeleteAllQualifierMappingsByResourceTypeAndId(resourceQualifiers.ImagePromotionPolicy, updateToPolicy.Id, sql.NewDefaultAuditLog(userId), nil)
	if err != nil {
		impl.logger.Errorw("error in deleting old qualifier mappings for a policy", "policyId", updateToPolicy.Id, "policyType", resourceQualifiers.ImagePromotionPolicy, "err", err)
		return err
	}
	// create new mappings using resourceQualifierMapping
	err = impl.resourceQualifierMappingService.CreateMappings(nil, userId, resourceQualifiers.ImagePromotionPolicy, []int{updateToPolicy.Id}, resourceQualifiers.ApplicationEnvironmentSelector, scopes)
	if err != nil {
		impl.logger.Errorw("error in creating new qualifier mappings for a policy", "policyId", updateToPolicy.Id, "policyType", resourceQualifiers.ImagePromotionPolicy, "err", err)
		return err
	}
	err = impl.resourceQualifierMappingService.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction while bulk applying policies to selected app env entities", "requestPayload", applyIdentifiersRequest, "err", err)
		return err
	}
	return nil
}

func (impl CommonPolicyActionsServiceImpl) fetchScopesByAppEnvNames(applicationEnvironments []AppEnvPolicyContainer) ([]*resourceQualifiers.Scope, error) {
	appNames := make([]string, 0, len(applicationEnvironments))
	envNames := make([]string, 0, len(applicationEnvironments))
	for _, appEnv := range applicationEnvironments {
		appNames = append(appNames, appEnv.AppName)
		envNames = append(envNames, appEnv.EnvName)
	}

	apps, err := impl.appService.FindAppByNames(appNames)
	if err != nil {
		impl.logger.Errorw("error in finding the apps with names", "appNames", appNames, "err", err)
		return nil, err
	}
	envs, err := impl.environmentService.FindByNames(envNames)
	if err != nil {
		impl.logger.Errorw("error in finding the environments with names", "envNames", envNames, "err", err)
		return nil, err
	}

	appNameIdMap := make(map[string]int)
	envNameIdMap := make(map[string]int)

	for _, app := range apps {
		appNameIdMap[app.AppName] = app.Id
	}

	for _, env := range envs {
		envNameIdMap[env.Environment] = env.Id
	}

	scopes := make([]*resourceQualifiers.Scope, 0, len(applicationEnvironments))
	uniqueScopes := make(map[string]bool)
	for _, appEnv := range applicationEnvironments {
		key := fmt.Sprintf("%s,%s", appEnv.AppName, appEnv.EnvName)
		if _, ok := uniqueScopes[key]; !ok {
			scopes = append(scopes, &resourceQualifiers.Scope{
				AppId: appNameIdMap[appEnv.AppName],
				EnvId: envNameIdMap[appEnv.EnvName],
				SystemMetadata: &resourceQualifiers.SystemMetadata{
					AppName:         appEnv.AppName,
					EnvironmentName: appEnv.EnvName,
				},
			})
		}
	}

	return scopes, nil
}
