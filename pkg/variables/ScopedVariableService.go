package variables

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables/cache"
	"github.com/devtron-labs/devtron/pkg/variables/helper"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	repository2 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"sync"
	"time"
)

type ScopedVariableService interface {
	CreateVariables(payload models.Payload) error
	GetScopedVariables(scope resourceQualifiers.Scope, varNames []string, maskSensitiveData bool) (scopedVariableDataObj []*models.ScopedVariableData, err error)
	GetJsonForVariables() (*models.Payload, error)
}

type ScopedVariableServiceImpl struct {
	logger                   *zap.SugaredLogger
	scopedVariableRepository repository2.ScopedVariableRepository
	qualifierMappingService  resourceQualifiers.QualifierMappingService
	devtronResourceService   devtronResource.DevtronResourceService
	VariableNameConfig       *VariableConfig
	VariableCache            *cache.VariableCacheObj

	//Enterprise only
	appRepository         app.AppRepository
	environmentRepository repository.EnvironmentRepository
	clusterRepository     repository.ClusterRepository
}

func NewScopedVariableServiceImpl(logger *zap.SugaredLogger, scopedVariableRepository repository2.ScopedVariableRepository, appRepository app.AppRepository, environmentRepository repository.EnvironmentRepository, devtronResourceService devtronResource.DevtronResourceService, clusterRepository repository.ClusterRepository,
	qualifierMappingService resourceQualifiers.QualifierMappingService) (*ScopedVariableServiceImpl, error) {
	scopedVariableService := &ScopedVariableServiceImpl{
		logger:                   logger,
		scopedVariableRepository: scopedVariableRepository,
		qualifierMappingService:  qualifierMappingService,
		VariableCache:            &cache.VariableCacheObj{CacheLock: &sync.Mutex{}},

		//Enterprise only
		appRepository:          appRepository,
		environmentRepository:  environmentRepository,
		devtronResourceService: devtronResourceService,
		clusterRepository:      clusterRepository,
	}
	cfg, err := GetVariableNameConfig()
	if err != nil {
		return nil, err
	}
	loadVariableCache(cfg, scopedVariableService)
	scopedVariableService.VariableNameConfig = cfg
	return scopedVariableService, nil
}

type VariableConfig struct {
	VariableNameRegex    string `env:"SCOPED_VARIABLE_NAME_REGEX" envDefault:"^[a-zA-Z][a-zA-Z0-9_-]{0,62}[a-zA-Z0-9]$"`
	VariableCacheEnabled bool   `env:"VARIABLE_CACHE_ENABLED" envDefault:"true"`
}

func loadVariableCache(cfg *VariableConfig, service *ScopedVariableServiceImpl) {
	if cfg.VariableCacheEnabled {
		go service.loadVarCache()
	}
}
func GetVariableNameConfig() (*VariableConfig, error) {
	cfg := &VariableConfig{}
	err := env.Parse(cfg)
	return cfg, err
}
func (impl *ScopedVariableServiceImpl) SetVariableCache(cache *cache.VariableCacheObj) {
	impl.VariableCache = cache
}

func (impl *ScopedVariableServiceImpl) loadVarCache() {
	variableCache := impl.VariableCache
	variableCache.ResetCache()
	variableCache.TakeLock()
	defer variableCache.ReleaseLock()
	variableMetadata, err := impl.scopedVariableRepository.GetAllVariableMetadata()
	if err != nil {
		impl.logger.Errorw("error occurred while fetching variable metadata", "err", err)
		return
	}
	variableCache.SetData(variableMetadata)
	impl.logger.Info("variable cache loaded successfully")
}

func (impl *ScopedVariableServiceImpl) CreateVariables(payload models.Payload) error {
	err, _ := impl.isValidPayload(payload)
	if err != nil {
		impl.logger.Errorw("error in variable payload validation", "err", err)
		return err
	}

	auditLog := getAuditLog(payload)
	// Begin Transaction
	tx, err := impl.scopedVariableRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction of variable creation", "err", err)
		return err
	}
	// Rollback Transaction in case of any error
	defer func(scopedVariableRepository repository2.ScopedVariableRepository, tx *pg.Tx) {
		err = scopedVariableRepository.RollbackTx(tx)
		if err != nil {
			impl.logger.Errorw("error in rollback transaction of variable creation", "err", err)
			return
		}
	}(impl.scopedVariableRepository, tx)

	err = impl.qualifierMappingService.DeleteAllQualifierMappings(resourceQualifiers.Variable, auditLog, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting qualifier mappings", "err", err)
		return err
	}
	err = impl.scopedVariableRepository.DeleteVariables(auditLog, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting variables", "err", err)
		return err
	}
	if len(payload.Variables) != 0 {
		varNameIdMap, err := impl.storeVariableDefinitions(payload, auditLog, tx)
		if err != nil {
			return err
		}

		scopeIdToVarData, err := impl.createVariableScopes(payload, varNameIdMap, auditLog, tx)
		if err != nil {
			return err
		}
		err = impl.storeVariableData(scopeIdToVarData, auditLog, tx)
		if err != nil {
			return err
		}

	}
	err = impl.scopedVariableRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction of variable creation", "err", err)
		return err
	}
	loadVariableCache(impl.VariableNameConfig, impl)
	return nil
}

func (impl *ScopedVariableServiceImpl) storeVariableData(scopeIdToVarData map[int]string, auditLog sql.AuditLog, tx *pg.Tx) error {
	VariableDataList := make([]*repository2.VariableData, 0)
	for scopeId, data := range scopeIdToVarData {
		varData := &repository2.VariableData{
			VariableScopeId: scopeId,
			Data:            data,
			AuditLog:        auditLog,
		}
		VariableDataList = append(VariableDataList, varData)
	}
	if len(VariableDataList) > 0 {
		err := impl.scopedVariableRepository.CreateVariableData(VariableDataList, tx)
		if err != nil {
			impl.logger.Errorw("error in saving variable data", "err", err)
			return err
		}
	}
	return nil
}

func (impl *ScopedVariableServiceImpl) storeVariableDefinitions(payload models.Payload, auditLog sql.AuditLog, tx *pg.Tx) (map[string]int, error) {
	variableDefinitions := make([]*repository2.VariableDefinition, 0, len(payload.Variables))
	for _, variable := range payload.Variables {
		variableDefinition := repository2.CreateFromDefinition(variable.Definition, auditLog)
		variableDefinitions = append(variableDefinitions, variableDefinition)
	}
	varDef, err := impl.scopedVariableRepository.CreateVariableDefinition(variableDefinitions, tx)
	if err != nil {
		impl.logger.Errorw("error occurred while saving variable definition", "err", err)
		return nil, err
	}
	variableNameToId := make(map[string]int)
	for _, variable := range varDef {
		variableNameToId[variable.Name] = variable.Id
	}
	return variableNameToId, nil
}

func (impl *ScopedVariableServiceImpl) createVariableScopes(payload models.Payload, variableNameToId map[string]int, auditLog sql.AuditLog, tx *pg.Tx) (map[int]string, error) {
	appNameToIdMap, envNameToIdMap, clusterNameToIdMap, err := impl.getAttributesIdMapping(payload)
	if err != nil {
		impl.logger.Errorw("error in getting  variable AttributeNameToIdMappings", "err", err)
		return nil, err
	}
	searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
	variableScopes := make([]*models.VariableScope, 0)
	for _, variable := range payload.Variables {

		variableId := variableNameToId[variable.Definition.VarName]
		for _, value := range variable.AttributeValues {
			var varValue string
			varValue, err := utils.StringifyValue(value.VariableValue.Value)
			if err != nil {
				return nil, err
			}
			if value.AttributeType == models.Global {
				scope := &models.VariableScope{
					QualifierMapping: &resourceQualifiers.QualifierMapping{
						ResourceId:   variableId,
						ResourceType: resourceQualifiers.Variable,
						QualifierId:  int(helper.GetQualifierId(value.AttributeType)),
						Active:       true,
						AuditLog:     auditLog,
					},
					Data: varValue,
				}
				variableScopes = append(variableScopes, scope)
			} else {
				var compositeString string
				if value.AttributeType == models.ApplicationEnv {
					compositeString = fmt.Sprintf("%v-%s-%s", variableId, value.AttributeParams[models.ApplicationName], value.AttributeParams[models.EnvName])
				}
				for identifierType, IdentifierName := range value.AttributeParams {
					identifierValue, err := helper.GetIdentifierValue(identifierType, appNameToIdMap, IdentifierName, envNameToIdMap, clusterNameToIdMap)
					if err != nil {
						impl.logger.Errorw("error in getting identifierValue", "err", err)
						return nil, err
					}
					scope := &models.VariableScope{
						QualifierMapping: &resourceQualifiers.QualifierMapping{
							ResourceId:            variableId,
							QualifierId:           int(helper.GetQualifierId(value.AttributeType)),
							IdentifierKey:         helper.GetIdentifierKey(identifierType, searchableKeyNameIdMap),
							IdentifierValueInt:    identifierValue,
							Active:                true,
							CompositeKey:          compositeString,
							IdentifierValueString: IdentifierName,
							AuditLog:              auditLog,
						},
						Data: varValue,
					}
					variableScopes = append(variableScopes, scope)
				}
			}
		}
	}
	parentVariableScope := make([]*resourceQualifiers.QualifierMapping, 0)
	childrenVariableScope := make([]*resourceQualifiers.QualifierMapping, 0)
	parentScopesMap := make(map[string]*resourceQualifiers.QualifierMapping)

	for _, scope := range variableScopes {
		if scope.QualifierId == 1 && scope.IdentifierKey == searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID] {
			childrenVariableScope = append(childrenVariableScope, scope.QualifierMapping)
		} else {
			parentVariableScope = append(parentVariableScope, scope.QualifierMapping)
			if scope.QualifierId == 1 {
				parentScopesMap[scope.CompositeKey] = scope.QualifierMapping
			}
		}
	}
	var parentVarScope []*resourceQualifiers.QualifierMapping
	var childVarScope []*resourceQualifiers.QualifierMapping

	if len(parentVariableScope) > 0 {
		parentVarScope, err = impl.qualifierMappingService.CreateQualifierMappings(parentVariableScope, tx)
		if err != nil {
			impl.logger.Errorw("error in getting parentVarScope", "parentVarScope", parentVarScope, "err", err)
			return nil, err
		}
	}
	scopeIdToVarData := make(map[int]string)
	for _, parentVar := range variableScopes {
		scopeIdToVarData[parentVar.Id] = parentVar.Data
	}
	for _, childScope := range childrenVariableScope {
		parentScope, exists := parentScopesMap[childScope.CompositeKey]
		if exists {
			childScope.ParentIdentifier = parentScope.Id
		}
	}
	if len(childrenVariableScope) > 0 {
		childVarScope, err = impl.qualifierMappingService.CreateQualifierMappings(childrenVariableScope, tx)
		if err != nil {
			impl.logger.Errorw("error in getting childVarScope", err, childVarScope)
			return nil, err
		}
	}
	return scopeIdToVarData, nil
}

func (impl *ScopedVariableServiceImpl) getMatchedScopedVariables(varScope []*resourceQualifiers.QualifierMapping) map[int]int {
	variableIdToVariableScopes := make(map[int][]*resourceQualifiers.QualifierMapping)
	variableIdToSelectedScopeId := make(map[int]int)
	for _, vScope := range varScope {
		variableId := vScope.ResourceId
		variableIdToVariableScopes[variableId] = append(variableIdToVariableScopes[variableId], vScope)
	}
	// Filter out the unneeded scoped which were fetched from DB for the same variable and qualifier
	for variableId, scopes := range variableIdToVariableScopes {

		selectedScopes := make([]*resourceQualifiers.QualifierMapping, 0)
		compoundQualifierToScopes := make(map[resourceQualifiers.Qualifier][]*resourceQualifiers.QualifierMapping)

		for _, variableScope := range scopes {
			qualifier := resourceQualifiers.Qualifier(variableScope.QualifierId)
			if slices.Contains(resourceQualifiers.CompoundQualifiers, qualifier) {
				compoundQualifierToScopes[qualifier] = append(compoundQualifierToScopes[qualifier], variableScope)
			} else {
				selectedScopes = append(selectedScopes, variableScope)
			}
		}

		for _, qualifier := range resourceQualifiers.CompoundQualifiers {
			selectedScope := impl.selectScopeForCompoundQualifier(compoundQualifierToScopes[qualifier], resourceQualifiers.GetNumOfChildQualifiers(qualifier))
			if selectedScope != nil {
				selectedScopes = append(selectedScopes, selectedScope)
			}
		}
		variableIdToVariableScopes[variableId] = selectedScopes
	}

	var minScope *resourceQualifiers.QualifierMapping
	for variableId, scopes := range variableIdToVariableScopes {
		minScope = helper.FindMinWithComparator(scopes, helper.QualifierComparator)
		if minScope != nil {
			variableIdToSelectedScopeId[variableId] = minScope.Id
		}
	}
	return variableIdToSelectedScopeId
}

func (impl *ScopedVariableServiceImpl) selectScopeForCompoundQualifier(scopes []*resourceQualifiers.QualifierMapping, numQualifiers int) *resourceQualifiers.QualifierMapping {
	parentIdToChildScopes := make(map[int][]*resourceQualifiers.QualifierMapping)
	parentScopeIdToScope := make(map[int]*resourceQualifiers.QualifierMapping, 0)
	parentScopeIds := make([]int, 0)
	for _, scope := range scopes {
		// is not parent so append it to the list in the map with key as its parent scopeID
		if scope.ParentIdentifier > 0 {
			parentIdToChildScopes[scope.ParentIdentifier] = append(parentIdToChildScopes[scope.ParentIdentifier], scope)
		} else {
			//is parent so collect IDs and put it in a map for easy retrieval
			parentScopeIds = append(parentScopeIds, scope.Id)
			parentScopeIdToScope[scope.Id] = scope
		}
	}

	for parentScopeId, _ := range parentIdToChildScopes {
		// this deletes the keys in the map where the key does not exist in the collected IDs for parent
		if !slices.Contains(parentScopeIds, parentScopeId) {
			delete(parentIdToChildScopes, parentScopeId)
		}
	}

	// Now in the map only those will exist with all child matched or partial matches.
	// Because only one will entry exist with all matched we'll return that scope.
	var selectedParentScope *resourceQualifiers.QualifierMapping
	for parentScopeId, childScopes := range parentIdToChildScopes {
		if len(childScopes) == numQualifiers {
			selectedParentScope = parentScopeIdToScope[parentScopeId]
		}
	}
	return selectedParentScope
}

func (impl *ScopedVariableServiceImpl) GetScopedVariables(scope resourceQualifiers.Scope, varNames []string, maskSensitiveData bool) (scopedVariableDataObj []*models.ScopedVariableData, err error) {

	// getting all variables from cache
	allVariableDefinitions := impl.VariableCache.GetData()

	// cache is loaded and no active variables exist. Returns empty
	if allVariableDefinitions != nil && len(allVariableDefinitions) == 0 {
		return nil, nil
	}

	// Need to get from repo for isSensitive even if cache is loaded since cache only contains metadata
	if allVariableDefinitions == nil {
		allVariableDefinitions, err = impl.scopedVariableRepository.GetAllVariables()

		//Cache was not loaded and no active variables found
		if len(allVariableDefinitions) == 0 {
			return nil, nil
		}
	}

	// filtering out variables whose name is not present in varNames
	var variableDefinitions []*repository2.VariableDefinition
	for _, definition := range allVariableDefinitions {
		// we don't apply filter logic when provided varNames is nil
		if varNames == nil || slices.Contains(varNames, definition.Name) {
			variableDefinitions = append(variableDefinitions, definition)
		}
	}

	variableIds := make([]int, 0)
	variableIdToDefinition := make(map[int]*repository2.VariableDefinition)
	for _, definition := range variableDefinitions {
		variableIds = append(variableIds, definition.Id)
		variableIdToDefinition[definition.Id] = definition
	}

	// This to prevent corner case where no variables were found for the provided names
	if len(varNames) > 0 && len(variableIds) == 0 {
		return make([]*models.ScopedVariableData, 0), nil
	}

	searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
	varScope, err := impl.qualifierMappingService.GetQualifierMappings(resourceQualifiers.Variable, scope, searchableKeyNameIdMap, variableIds)
	if err != nil {
		impl.logger.Errorw("error in getting varScope", "err", err)
		return nil, err
	}

	variableIdToSelectedScopeId := impl.getMatchedScopedVariables(varScope)

	scopeIds := make([]int, 0)
	foundVarIds := make([]int, 0) // the variable IDs which have data
	for varId, scopeId := range variableIdToSelectedScopeId {
		scopeIds = append(scopeIds, scopeId)
		foundVarIds = append(foundVarIds, varId)
	}
	var variableData []*repository2.VariableData
	if len(scopeIds) != 0 {
		variableData, err = impl.scopedVariableRepository.GetDataForScopeIds(scopeIds)
		if err != nil {
			impl.logger.Errorw("error in getting variable data", "err", err)
			return nil, err
		}
	}

	scopeIdToVarData := make(map[int]*repository2.VariableData)
	for _, varData := range variableData {
		scopeIdToVarData[varData.VariableScopeId] = varData
	}

	for varId, scopeId := range variableIdToSelectedScopeId {
		var value interface{}
		value, err = utils.DestringifyValue(scopeIdToVarData[scopeId].Data)
		if err != nil {
			impl.logger.Errorw("error in validating value", "err", err)
			return nil, err
		}

		var varValue *models.VariableValue
		var isRedacted bool
		if !maskSensitiveData && variableIdToDefinition[varId].VarType == models.PRIVATE {
			varValue = &models.VariableValue{Value: ""}
			isRedacted = true
		} else {
			varValue = &models.VariableValue{Value: value}
		}
		scopedVariableData := &models.ScopedVariableData{
			VariableName:     variableIdToDefinition[varId].Name,
			ShortDescription: variableIdToDefinition[varId].ShortDescription,
			VariableValue:    varValue,
			IsRedacted:       isRedacted}

		scopedVariableDataObj = append(scopedVariableDataObj, scopedVariableData)
	}

	//adding variable def for variables which don't have any scoped data defined
	// This only happens when passed var names is null (called from UI to get all variables with or without data)
	if varNames == nil {
		for _, definition := range allVariableDefinitions {
			if !slices.Contains(foundVarIds, definition.Id) {
				scopedVariableDataObj = append(scopedVariableDataObj, &models.ScopedVariableData{
					VariableName:     definition.Name,
					ShortDescription: definition.ShortDescription,
				})
			}
		}
	}

	return scopedVariableDataObj, err
}

func (impl *ScopedVariableServiceImpl) GetJsonForVariables() (*models.Payload, error) {

	// getting all variables from cache, if empty then no variables exist
	allVariableDefinitions := impl.VariableCache.GetData()
	if allVariableDefinitions != nil && len(allVariableDefinitions) == 0 {
		return nil, nil
	}
	dataForJson, err := impl.scopedVariableRepository.GetAllVariableScopeAndDefinition()
	if err != nil {
		impl.logger.Errorw("error in getting data for json", "err", err)
		return nil, err
	}
	resourceKeyMap := impl.devtronResourceService.GetAllSearchableKeyIdNameMap()

	payload := &models.Payload{
		Variables: make([]*models.Variables, 0),
	}
	variables := make([]*models.Variables, 0)

	varIdVsScopeMappings, varScopeIds, err := impl.getVariableScopes(dataForJson)
	if err != nil {
		return nil, err
	}
	scopeIdVsDataMap, err := impl.getVariableScopeData(varScopeIds)
	if err != nil {
		return nil, err
	}

	for _, data := range dataForJson {
		definition := models.Definition{
			VarName:          data.Name,
			DataType:         data.DataType,
			VarType:          data.VarType,
			Description:      data.Description,
			ShortDescription: data.ShortDescription,
		}
		attributes := make([]models.AttributeValue, 0)

		scopedVariables := varIdVsScopeMappings[data.Id]
		scopeIdToVarScopes := make(map[int][]*resourceQualifiers.QualifierMapping)
		for _, scope := range scopedVariables {
			if scope.ParentIdentifier != 0 {
				scopeIdToVarScopes[scope.ParentIdentifier] = append(scopeIdToVarScopes[scope.ParentIdentifier], scope)
			} else {
				scopeIdToVarScopes[scope.Id] = []*resourceQualifiers.QualifierMapping{scope}
			}
		}
		for parentScopeId, scopes := range scopeIdToVarScopes {
			attribute := models.AttributeValue{
				AttributeParams: make(map[models.IdentifierType]string),
			}
			for _, scope := range scopes {
				if helper.GetIdentifierTypeFromResourceKey(scope.IdentifierKey, resourceKeyMap) != "" {
					attribute.AttributeParams[helper.GetIdentifierTypeFromResourceKey(scope.IdentifierKey, resourceKeyMap)] = scope.IdentifierValueString
				}
				scopeId := scope.Id
				if parentScopeId == scopeId {
					variableData := scopeIdVsDataMap[scopeId]
					var value interface{}
					value, err = utils.DestringifyValue(variableData.Data)
					if err != nil {
						return nil, err
					}
					attribute.VariableValue = models.VariableValue{
						Value: value,
					}
					attribute.AttributeType = helper.GetAttributeType(resourceQualifiers.Qualifier(scope.QualifierId))
				}
			}
			if len(attribute.AttributeParams) == 0 {
				attribute.AttributeParams = nil
			}
			attributes = append(attributes, attribute)
		}

		variable := &models.Variables{
			Definition:      definition,
			AttributeValues: attributes,
		}
		variables = append(variables, variable)
	}

	payload.Variables = variables
	if len(payload.Variables) == 0 {
		return nil, nil
	}
	return payload, nil
}

func (impl *ScopedVariableServiceImpl) getVariableScopes(dataForJson []*repository2.VariableDefinition) (map[int][]*resourceQualifiers.QualifierMapping, []int, error) {
	varIdVsScopeMappings := make(map[int][]*resourceQualifiers.QualifierMapping)
	var varScopeIds []int
	varDefnIds := make([]int, len(dataForJson))
	for _, variableDefinition := range dataForJson {
		varDefnIds = append(varDefnIds, variableDefinition.Id)
	}
	searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
	scope := resourceQualifiers.Scope{}
	scopedVariableMappings, err := impl.qualifierMappingService.GetQualifierMappings(resourceQualifiers.Variable, scope, searchableKeyNameIdMap, varDefnIds)
	if err != nil {
		//TODO KB: handle this
		return varIdVsScopeMappings, varScopeIds, err
	}

	for _, scopedVariableMapping := range scopedVariableMappings {
		varId := scopedVariableMapping.ResourceId
		varScopeIds = append(varScopeIds, scopedVariableMapping.Id)
		variableScopes := varIdVsScopeMappings[varId]
		variableScopes = append(variableScopes, scopedVariableMapping)
		varIdVsScopeMappings[varId] = variableScopes
	}
	return varIdVsScopeMappings, varScopeIds, nil
}

func (impl *ScopedVariableServiceImpl) getVariableScopeData(scopeIds []int) (map[int]*repository2.VariableData, error) {
	scopeIdVsVarDataMap := make(map[int]*repository2.VariableData, len(scopeIds))
	variableDataArray, err := impl.scopedVariableRepository.GetDataForScopeIds(scopeIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching data for scope ids", "err", err)
		return scopeIdVsVarDataMap, err
	}
	for _, variableData := range variableDataArray {
		variableScopeId := variableData.VariableScopeId
		scopeIdVsVarDataMap[variableScopeId] = variableData
	}
	return scopeIdVsVarDataMap, nil
}

func getAuditLog(payload models.Payload) sql.AuditLog {
	auditLog := sql.AuditLog{
		CreatedOn: time.Now(),
		CreatedBy: payload.UserId,
		UpdatedOn: time.Now(),
		UpdatedBy: payload.UserId,
	}
	return auditLog
}

func (impl *ScopedVariableServiceImpl) getAttributesIdMapping(payload models.Payload) (map[string]int, map[string]int, map[string]int, error) {
	appNames, envNames, clusterNames := helper.GetAttributeNames(payload)
	var appNameToId []*app.App
	var err error
	if len(appNames) != 0 {
		appNameToId, err = impl.appRepository.FindByNames(appNames)
		if err != nil {
			impl.logger.Errorw("error in getting appNameToId", err)
			return nil, nil, nil, err
		}
	}
	appNameToIdMap := make(map[string]int)
	envNameToIdMap := make(map[string]int)
	clusterNameToIdMap := make(map[string]int)
	for _, name := range appNameToId {
		appNameToIdMap[name.AppName] = name.Id
	}
	var envNameToId []*repository.Environment
	if len(envNames) != 0 {
		envNameToId, err = impl.environmentRepository.FindByNames(envNames)
		if err != nil {
			impl.logger.Errorw("error in getting envNameToId", err)
			return nil, nil, nil, err
		}
	}
	for _, val := range envNameToId {
		envNameToIdMap[val.Name] = val.Id

	}
	var clusterNameToId []*repository.Cluster
	if len(clusterNames) != 0 {
		clusterNameToId, err = impl.clusterRepository.FindByNames(clusterNames)
		if err != nil {
			impl.logger.Errorw("error in getting clusterNameToId", err)
			return nil, nil, nil, err
		}

	}

	for _, name := range clusterNameToId {
		clusterNameToIdMap[name.ClusterName] = name.Id
	}
	return appNameToIdMap, envNameToIdMap, clusterNameToIdMap, nil
}
