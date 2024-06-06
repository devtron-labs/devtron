/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package variables

import (
	"fmt"
	"github.com/argoproj/argo-workflows/v3/errors"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource/read"
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
	"regexp"
	"strings"
	"sync"
	"time"
)

type ScopedVariableService interface {
	CreateVariables(payload models.Payload) error
	GetScopedVariables(scope resourceQualifiers.Scope, varNames []string, unmaskSensitiveData bool) (scopedVariableDataObj []*models.ScopedVariableData, err error)
	GetJsonForVariables() (*models.Payload, error)
	CheckForSensitiveVariables(variableNames []string) (map[string]bool, error)
	GetFormattedVariableForName(name string) string
}

type ScopedVariableServiceImpl struct {
	logger                   *zap.SugaredLogger
	scopedVariableRepository repository2.ScopedVariableRepository
	qualifierMappingService  resourceQualifiers.QualifierMappingService
	VariableNameConfig       *VariableConfig
	VariableCache            *cache.VariableCacheObj
}

func NewScopedVariableServiceImpl(logger *zap.SugaredLogger, scopedVariableRepository repository2.ScopedVariableRepository, appRepository app.AppRepository, environmentRepository repository.EnvironmentRepository, devtronResourceSearchableKeyService read.DevtronResourceSearchableKeyService, clusterRepository repository.ClusterRepository,
	qualifierMappingService resourceQualifiers.QualifierMappingService) (*ScopedVariableServiceImpl, error) {
	scopedVariableService := &ScopedVariableServiceImpl{
		logger:                   logger,
		scopedVariableRepository: scopedVariableRepository,
		qualifierMappingService:  qualifierMappingService,
		VariableCache:            &cache.VariableCacheObj{CacheLock: &sync.Mutex{}},
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
	SystemVariablePrefix string `env:"SYSTEM_VAR_PREFIX" envDefault:"DEVTRON_"`
	ScopedVariableFormat string `env:"SCOPED_VARIABLE_FORMAT" envDefault:"@{{%s}}"`
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

func (impl *ScopedVariableServiceImpl) GetFormattedVariableForName(name string) string {
	return fmt.Sprintf(impl.VariableNameConfig.ScopedVariableFormat, name)
}

func (impl *ScopedVariableServiceImpl) CheckForSensitiveVariables(variableNames []string) (map[string]bool, error) {

	// getting all variables from cache
	allVariableDefinitions := impl.VariableCache.GetData()

	var err error
	// cache is not loaded get from repo
	if allVariableDefinitions == nil {
		allVariableDefinitions, err = impl.scopedVariableRepository.GetVariableTypeForVariableNames(variableNames)
		if err != nil {
			return nil, errors.Wrap(err, "400", "error in fetching variable type")
		}
	}

	variableNameToType := make(map[string]models.VariableType)
	for _, definition := range allVariableDefinitions {
		variableNameToType[definition.Name] = definition.VarType
	}

	varNameToIsSensitive := make(map[string]bool)
	for _, name := range variableNames {

		// by default all variables are marked sensitive to handle deleted variables
		// only super admin will be able to see the values once variable is deleted from system
		if varType, ok := variableNameToType[name]; ok {
			varNameToIsSensitive[name] = varType.IsTypeSensitive()
		} else {
			varNameToIsSensitive[name] = true
		}
	}
	return varNameToIsSensitive, nil
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

		scopeIdToVarData, err := impl.createVariableScopes(payload, varNameIdMap, auditLog.CreatedBy, tx)
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

func (impl *ScopedVariableServiceImpl) createVariableScopes(payload models.Payload, variableNameToId map[string]int, userId int32, tx *pg.Tx) (map[int]string, error) {

	variableScopes := make([]*models.VariableScope, 0)
	for _, variable := range payload.Variables {
		variableId := variableNameToId[variable.Definition.VarName]
		for _, value := range variable.AttributeValues {
			varValue, err := utils.StringifyValue(value.VariableValue.Value)
			if err != nil {
				return nil, err
			}
			selector := resourceQualifiers.GlobalSelector
			varScope := &models.VariableScope{
				Data: varValue,
				ResourceMappingSelection: &resourceQualifiers.ResourceMappingSelection{
					ResourceType:      resourceQualifiers.Variable,
					ResourceId:        variableId,
					QualifierSelector: selector,
				},
			}
			variableScopes = append(variableScopes, varScope)
		}
	}

	varScopeToSelection := make(map[*resourceQualifiers.ResourceMappingSelection]*models.VariableScope)
	selections := make([]*resourceQualifiers.ResourceMappingSelection, 0)
	for _, scope := range variableScopes {
		varScopeToSelection[scope.ResourceMappingSelection] = scope
		selections = append(selections, scope.ResourceMappingSelection)
	}

	savedSelections, err := impl.qualifierMappingService.CreateMappingsForSelections(tx, userId, selections)
	if err != nil {
		return nil, err
	}
	scopeIdToVarData := make(map[int]string)
	for _, savedSelection := range savedSelections {
		scopeIdToVarData[savedSelection.Id] = varScopeToSelection[savedSelection].Data //parentVar.Data
	}
	return scopeIdToVarData, nil
}

func (impl *ScopedVariableServiceImpl) getMatchedScopedVariables(varScope []*resourceQualifiers.QualifierMapping) map[int][]*resourceQualifiers.QualifierMapping {
	variableIdToVariableScopes := make(map[int][]*resourceQualifiers.QualifierMapping)
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
			selectedScope := impl.selectScopeForCompoundQualifier(compoundQualifierToScopes[qualifier], qualifier)
			if selectedScope != nil {
				selectedScopes = append(selectedScopes, selectedScope)
			}
		}
		variableIdToVariableScopes[variableId] = selectedScopes
	}

	return variableIdToVariableScopes

}

func (impl *ScopedVariableServiceImpl) getScopeWithPriority(variableIdToVariableScopes map[int][]*resourceQualifiers.QualifierMapping) map[int]int {
	variableIdToSelectedScopeId := make(map[int]int)
	var minScope *resourceQualifiers.QualifierMapping
	for variableId, scopes := range variableIdToVariableScopes {
		minScope = helper.FindMinWithComparator(scopes, helper.QualifierComparator)
		if minScope != nil {
			variableIdToSelectedScopeId[variableId] = minScope.Id
		}
	}
	return variableIdToSelectedScopeId
}

func (impl *ScopedVariableServiceImpl) selectScopeForCompoundQualifier(scopes []*resourceQualifiers.QualifierMapping, qualifier resourceQualifiers.Qualifier) *resourceQualifiers.QualifierMapping {
	numQualifiers := resourceQualifiers.GetNumOfChildQualifiers(qualifier)
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

func (impl *ScopedVariableServiceImpl) GetScopedVariables(scope resourceQualifiers.Scope, varNames []string, unmaskSensitiveData bool) (scopedVariableDataObj []*models.ScopedVariableData, err error) {

	//populating system variables from system metadata
	var systemVariableData, allSystemVariables []*models.ScopedVariableData
	if scope.SystemMetadata != nil {
		systemVariableData, allSystemVariables = impl.getSystemVariablesData(scope.SystemMetadata, varNames)
		scopedVariableDataObj = append(scopedVariableDataObj, systemVariableData...)
	}

	// getting all variables from cache
	allVariableDefinitions := impl.VariableCache.GetData()

	// cache is loaded and no active variables exist. Returns empty
	if allVariableDefinitions != nil && len(allVariableDefinitions) == 0 {
		return scopedVariableDataObj, nil
	}

	// Cache is not loaded
	if allVariableDefinitions == nil {
		allVariableDefinitions, err = impl.scopedVariableRepository.GetAllVariables()

		//Cache was not loaded and no active variables found
		if len(allVariableDefinitions) == 0 {
			return scopedVariableDataObj, nil
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
	for _, definition := range variableDefinitions {
		variableIds = append(variableIds, definition.Id)
	}
	// This to prevent corner case where no variables were found for the provided names
	if len(varNames) > 0 && len(variableIds) == 0 {
		return scopedVariableDataObj, nil
	}

	allVariableIds := make([]int, 0)
	variableIdToDefinition := make(map[int]*repository2.VariableDefinition)
	for _, definition := range allVariableDefinitions {
		allVariableIds = append(allVariableIds, definition.Id)
		variableIdToDefinition[definition.Id] = definition
	}

	varScope, err := impl.qualifierMappingService.GetQualifierMappings(resourceQualifiers.Variable, &scope, allVariableIds)
	if err != nil {
		impl.logger.Errorw("error in getting varScope", "err", err)
		return nil, err
	}

	matchedScopes := impl.getMatchedScopedVariables(varScope)
	variableIdToSelectedScopeId := impl.getScopeWithPriority(matchedScopes)

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
		if !unmaskSensitiveData && variableIdToDefinition[varId].VarType == models.PRIVATE {
			varValue = &models.VariableValue{Value: models.HiddenValue}
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

	allScopedVariableDataObj := scopedVariableDataObj
	usedScopedVariableDataObj := make([]*models.ScopedVariableData, 0)
	for _, data := range scopedVariableDataObj {
		if varNames == nil || slices.Contains(varNames, data.VariableName) {
			usedScopedVariableDataObj = append(usedScopedVariableDataObj, data)
		}
	}

	//adding variable def for variables which don't have any scoped data defined
	// This only happens when passed var names is null (called from UI to get all variables with or without data)
	if varNames == nil {
		for _, definition := range allVariableDefinitions {
			if !slices.Contains(foundVarIds, definition.Id) {
				usedScopedVariableDataObj = append(usedScopedVariableDataObj, &models.ScopedVariableData{
					VariableName:     definition.Name,
					ShortDescription: definition.ShortDescription,
				})
			}
		}
	}

	allScopedVariableDataObj = append(allScopedVariableDataObj, allSystemVariables...)
	impl.deduceVariables(usedScopedVariableDataObj, allScopedVariableDataObj)
	return usedScopedVariableDataObj, err
}

func resolveExpressionWithVariableValues(expr string, varNameToData map[string]*models.ScopedVariableData) (string, error) {
	// regex to find  variable placeholder and extracts a variable name which is alphanumeric
	// and can contain hyphen, underscore and whitespaces. white spaces will be trimmed on lookup
	variableRegex := `@{{([a-zA-Z0-9-_\s]+)}}`

	re := regexp.MustCompile(variableRegex)
	matches := re.FindAllStringSubmatch(expr, -1)

	for _, match := range matches {
		if len(match) == 2 {
			originalMatch := match[0]
			innerContent := match[1]

			variableName := strings.TrimSpace(innerContent)

			if data, ok := varNameToData[variableName]; ok {
				value := data.VariableValue.StringValue()
				expr = strings.Replace(expr, originalMatch, value, 1)
			} else {
				return expr, fmt.Errorf("variable not found %s", variableName)
			}
		}
	}
	return expr, nil
}

func (impl *ScopedVariableServiceImpl) deduceVariables(scopedVariableDataList []*models.ScopedVariableData, systemVariables []*models.ScopedVariableData) {
	varNameToData := make(map[string]*models.ScopedVariableData)
	for _, variable := range systemVariables {
		varNameToData[variable.VariableName] = variable
	}

	for _, data := range scopedVariableDataList {
		if data.VariableValue == nil {
			continue
		}
		value := data.VariableValue.Value
		if utils.IsStringType(value) {
			resolvedValue, err := resolveExpressionWithVariableValues(value.(string), varNameToData)
			if err != nil {
				impl.logger.Warnw("variables not resolved", "err", err, "value", value)
				continue
			}
			data.VariableValue.Value = resolvedValue
		}
	}
}

func (impl *ScopedVariableServiceImpl) getSystemVariablesData(metadata *resourceQualifiers.SystemMetadata, varNames []string) ([]*models.ScopedVariableData, []*models.ScopedVariableData) {
	systemVariables := make([]*models.ScopedVariableData, 0)
	allSystemVariables := make([]*models.ScopedVariableData, 0)
	for _, variable := range resourceQualifiers.SystemVariables {
		if len(metadata.GetDataFromSystemVariable(variable)) > 0 {
			systemVariable := &models.ScopedVariableData{
				VariableName:  string(variable),
				VariableValue: &models.VariableValue{Value: metadata.GetDataFromSystemVariable(variable)},
			}
			allSystemVariables = append(allSystemVariables, systemVariable)
			if slices.Contains(varNames, string(variable)) {
				systemVariables = append(systemVariables, systemVariable)
			}
		}
	}
	return systemVariables, allSystemVariables
}

func (impl *ScopedVariableServiceImpl) GetJsonForVariables() (*models.Payload, error) {

	// getting all variables from cache, if empty then no variables exist
	allVariableDefinitions := impl.VariableCache.GetData()
	if allVariableDefinitions != nil && len(allVariableDefinitions) == 0 {
		return nil, nil
	}
	dataForJson, err := impl.scopedVariableRepository.GetAllVariableDefinition()
	if err != nil {
		impl.logger.Errorw("error in getting data for json", "err", err)
		return nil, err
	}

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
	varDefnIds := make([]int, 0, len(dataForJson))
	for _, variableDefinition := range dataForJson {
		varDefnIds = append(varDefnIds, variableDefinition.Id)
	}
	if len(varDefnIds) == 0 {
		return varIdVsScopeMappings, varScopeIds, nil
	}
	scopedVariableMappings, err := impl.qualifierMappingService.GetQualifierMappings(resourceQualifiers.Variable, nil, varDefnIds)
	if err != nil {
		//TODO KB: handle this
		return varIdVsScopeMappings, varScopeIds, err
	}

	for _, scopedVariableMapping := range scopedVariableMappings {
		varId := scopedVariableMapping.ResourceId
		varScopeIds = append(varScopeIds, scopedVariableMapping.Id)
		varIdVsScopeMappings[varId] = append(varIdVsScopeMappings[varId], scopedVariableMapping)
	}
	return varIdVsScopeMappings, varScopeIds, nil
}

func (impl *ScopedVariableServiceImpl) getVariableScopeData(scopeIds []int) (map[int]*repository2.VariableData, error) {
	scopeIdVsVarDataMap := make(map[int]*repository2.VariableData, len(scopeIds))
	if len(scopeIds) == 0 {
		return scopeIdVsVarDataMap, nil
	}
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
