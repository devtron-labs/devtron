package variables

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
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
	GetScopedVariables(scope models.Scope, varNames []string) (scopedVariableDataObj []*models.ScopedVariableData, err error)
	GetJsonForVariables() (*models.Payload, error)
}

type ScopedVariableServiceImpl struct {
	logger                   *zap.SugaredLogger
	scopedVariableRepository repository2.ScopedVariableRepository
	appRepository            app.AppRepository
	environmentRepository    repository.EnvironmentRepository
	devtronResourceService   devtronResource.DevtronResourceService
	clusterRepository        repository.ClusterRepository
	variableNameConfig       *VariableConfig
	VariableCache            *cache.VariableCacheObj
}

func NewScopedVariableServiceImpl(logger *zap.SugaredLogger, scopedVariableRepository repository2.ScopedVariableRepository, appRepository app.AppRepository, environmentRepository repository.EnvironmentRepository, devtronResourceService devtronResource.DevtronResourceService, clusterRepository repository.ClusterRepository) (*ScopedVariableServiceImpl, error) {
	scopedVariableService := &ScopedVariableServiceImpl{
		logger:                   logger,
		scopedVariableRepository: scopedVariableRepository,
		appRepository:            appRepository,
		environmentRepository:    environmentRepository,
		devtronResourceService:   devtronResourceService,
		clusterRepository:        clusterRepository,
		VariableCache:            &cache.VariableCacheObj{CacheLock: &sync.Mutex{}},
	}
	cfg, err := GetVariableNameConfig()
	if err != nil {
		return nil, err
	}
	scopedVariableService.variableNameConfig = cfg
	go scopedVariableService.loadVarCache()
	return scopedVariableService, nil
}

type VariableConfig struct {
	VariableNameRegex string `env:"SCOPED_VARIABLE_NAME_REGEX" envDefault:"^[a-zA-Z][a-zA-Z0-9_-]{0,62}[a-zA-Z0-9]$"`
}

func GetVariableNameConfig() (*VariableConfig, error) {
	cfg := &VariableConfig{}
	err := env.Parse(cfg)
	return cfg, err
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
	if len(payload.Variables) == 0 {
		return nil
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

	err = impl.scopedVariableRepository.DeleteVariables(auditLog, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting variables", "err", err)
		return err
	}

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
	err = impl.scopedVariableRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction of variable creation", "err", err)
		return err
	}
	go impl.loadVarCache()
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
	variableScopes := make([]*repository2.VariableScope, 0)
	for _, variable := range payload.Variables {

		variableId := variableNameToId[variable.Definition.VarName]
		for _, value := range variable.AttributeValues {
			var varValue string
			varValue, err = utils.StringifyValue(value.VariableValue.Value)
			if err != nil {
				impl.logger.Errorw("error in validating dataType", "err", err)
				return nil, err
			}
			if value.AttributeType == models.Global {
				scope := &repository2.VariableScope{
					VariableDefinitionId: variableId,
					QualifierId:          int(utils.GetQualifierId(value.AttributeType)),
					Active:               true,
					Data:                 varValue,
					AuditLog:             auditLog,
				}
				variableScopes = append(variableScopes, scope)
			} else {
				var compositeString string
				if value.AttributeType == models.ApplicationEnv {
					compositeString = fmt.Sprintf("%v-%s-%s", variableId, value.AttributeParams[models.ApplicationName], value.AttributeParams[models.EnvName])
				}
				for identifierType, IdentifierName := range value.AttributeParams {
					var identifierValue int
					identifierValue, err = helper.GetIdentifierValue(identifierType, appNameToIdMap, IdentifierName, envNameToIdMap, clusterNameToIdMap)
					if err != nil {
						impl.logger.Errorw("error in getting identifierValue", "err", err)
						return nil, err
					}
					scope := &repository2.VariableScope{
						VariableDefinitionId:  variableId,
						QualifierId:           int(utils.GetQualifierId(value.AttributeType)),
						IdentifierKey:         utils.GetIdentifierKey(identifierType, searchableKeyNameIdMap),
						IdentifierValueInt:    identifierValue,
						Active:                true,
						CompositeKey:          compositeString,
						IdentifierValueString: IdentifierName,
						Data:                  varValue,
						AuditLog:              auditLog,
					}
					variableScopes = append(variableScopes, scope)
				}
			}
		}
	}
	parentVariableScope := make([]*repository2.VariableScope, 0)
	childrenVariableScope := make([]*repository2.VariableScope, 0)
	parentScopesMap := make(map[string]*repository2.VariableScope)

	for _, scope := range variableScopes {
		if scope.QualifierId == 1 && scope.IdentifierKey == searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID] {
			childrenVariableScope = append(childrenVariableScope, scope)
		} else {
			parentVariableScope = append(parentVariableScope, scope)
			if scope.QualifierId == 1 {
				parentScopesMap[scope.CompositeKey] = scope
			}
		}
	}
	var parentVarScope []*repository2.VariableScope
	var childVarScope []*repository2.VariableScope
	if len(parentVariableScope) > 0 {
		parentVarScope, err = impl.scopedVariableRepository.CreateVariableScope(parentVariableScope, tx)
		if err != nil {
			impl.logger.Errorw("error in getting parentVarScope", parentVarScope)
			return nil, err
		}
	}
	scopeIdToVarData := make(map[int]string)
	for _, parentVar := range parentVarScope {
		scopeIdToVarData[parentVar.Id] = parentVar.Data
	}
	for _, childScope := range childrenVariableScope {
		parentScope, exists := parentScopesMap[childScope.CompositeKey]
		if exists {
			childScope.ParentIdentifier = parentScope.Id
		}
	}
	if len(childrenVariableScope) > 0 {
		childVarScope, err = impl.scopedVariableRepository.CreateVariableScope(childrenVariableScope, tx)
		if err != nil {
			impl.logger.Errorw("error in getting childVarScope", err, childVarScope)
			return nil, err
		}
	}
	return scopeIdToVarData, nil
}

func (impl *ScopedVariableServiceImpl) getMatchedScopedVariables(varScope []*repository2.VariableScope) map[int]int {
	variableIdToVariableScopes := make(map[int][]*repository2.VariableScope)
	variableIdToSelectedScopeId := make(map[int]int)
	for _, vScope := range varScope {
		variableId := vScope.VariableDefinitionId
		variableIdToVariableScopes[variableId] = append(variableIdToVariableScopes[variableId], vScope)
	}
	// Filter out the unneeded scoped which were fetched from DB for the same variable and qualifier
	for variableId, scopes := range variableIdToVariableScopes {
		scopeIdToScope := make(map[int]*repository2.VariableScope)
		var matchedScope *repository2.VariableScope
		selectedScopes := make([]*repository2.VariableScope, 0)
		for _, variableScope := range scopes {
			if slices.Contains(repository2.CompoundQualifiers, repository2.Qualifier(variableScope.QualifierId)) && matchedScope == nil {
				if _, ok := scopeIdToScope[variableScope.Id]; ok {
					// when child was found first, it would be present in map
					// we'll select the scope of parent which is the current scope
					matchedScope = variableScope
				} else {
					scopeIdToScope[variableScope.Id] = variableScope
				}
				if variableScope.ParentIdentifier > 0 {
					if foundScope, ok := scopeIdToScope[variableScope.ParentIdentifier]; ok {
						// when parent was found first, it would be present in map
						// we'll select the scope of parent which is found in the map
						matchedScope = foundScope
					} else {
						scopeIdToScope[variableScope.ParentIdentifier] = variableScope
					}
				}
			} else {
				selectedScopes = append(selectedScopes, variableScope)
			}
		}
		if matchedScope != nil {
			selectedScopes = append(selectedScopes, matchedScope)
		}
		variableIdToVariableScopes[variableId] = selectedScopes
	}

	var minScope *repository2.VariableScope
	for variableId, scopes := range variableIdToVariableScopes {
		minScope = helper.FindMinWithComparator(scopes, helper.QualifierComparator)
		variableIdToSelectedScopeId[variableId] = minScope.Id
	}
	return variableIdToSelectedScopeId
}
func (impl *ScopedVariableServiceImpl) GetScopedVariables(scope models.Scope, varNames []string) (scopedVariableDataObj []*models.ScopedVariableData, err error) {

	// getting all variables from cache, if nil get from repo
	allVariableDefinitions := impl.VariableCache.GetData()
	if allVariableDefinitions == nil {
		allVariableDefinitions, err = impl.scopedVariableRepository.GetAllVariables()
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

	varScope, err := impl.scopedVariableRepository.GetScopedVariableData(scope, searchableKeyNameIdMap, variableIds)
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
		scopedVariableData := &models.ScopedVariableData{
			VariableName:  variableIdToDefinition[varId].Name,
			VariableValue: models.VariableValue{Value: value}}

		scopedVariableDataObj = append(scopedVariableDataObj, scopedVariableData)
	}

	//adding variable def for variables which don't have any scoped data defined
	// This only happens when passed var names is null (called from UI to get all variables with or without data)
	if varNames == nil {
		for _, definition := range allVariableDefinitions {
			if !slices.Contains(foundVarIds, definition.Id) {
				scopedVariableDataObj = append(scopedVariableDataObj, &models.ScopedVariableData{
					VariableName: definition.Name,
				})
			}
		}
	}

	return scopedVariableDataObj, err
}

func (impl *ScopedVariableServiceImpl) GetJsonForVariables() (*models.Payload, error) {

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

	for _, data := range dataForJson {
		definition := models.Definition{
			VarName:     data.Name,
			DataType:    data.DataType,
			VarType:     data.VarType,
			Description: data.Description,
		}
		attributes := make([]models.AttributeValue, 0)

		scopeIdToVarScopes := make(map[int][]*repository2.VariableScope)
		for _, scope := range data.VariableScope {
			if scope.ParentIdentifier != 0 {
				scopeIdToVarScopes[scope.ParentIdentifier] = append(scopeIdToVarScopes[scope.ParentIdentifier], scope)
			} else {
				scopeIdToVarScopes[scope.Id] = []*repository2.VariableScope{scope}
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
				if parentScopeId == scope.Id {
					var value interface{}
					value, err = utils.DestringifyValue(scope.VariableData.Data)
					if err != nil {
						return nil, err
					}
					attribute.VariableValue = models.VariableValue{
						Value: value,
					}
					attribute.AttributeType = utils.GetAttributeType(repository2.Qualifier(scope.QualifierId))
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
