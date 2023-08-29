package variables

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	repository2 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/go-pg/pg"
	"github.com/invopop/jsonschema"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
	"slices"
	"strconv"
	"strings"
	"time"
)

type ScopedVariableData struct {
	VariableName  string      `json:"variableName"`
	VariableValue interface{} `json:"variableValue,omitempty"`
}

const (
	YAML_TYPE = "yaml"
	JSON_TYPE = "json"
)

type ScopedVariableService interface {
	CreateVariables(payload repository2.Payload) error
	GetScopedVariables(scope repository2.Scope, varNames []string) (scopedVariableDataObj []*ScopedVariableData, err error)
	GetJsonForVariables() (*repository2.Payload, string, error)
}

type ScopedVariableServiceImpl struct {
	logger                   *zap.SugaredLogger
	scopedVariableRepository repository2.ScopedVariableRepository
	appRepository            app.AppRepository
	environmentRepository    repository.EnvironmentRepository
	devtronResourceService   devtronResource.DevtronResourceService
	clusterRepository        repository.ClusterRepository
}

func NewScopedVariableServiceImpl(logger *zap.SugaredLogger, scopedVariableRepository repository2.ScopedVariableRepository, appRepository app.AppRepository, environmentRepository repository.EnvironmentRepository, devtronResourceService devtronResource.DevtronResourceService, clusterRepository repository.ClusterRepository) (*ScopedVariableServiceImpl, error) {
	scopedVariableService := &ScopedVariableServiceImpl{
		logger:                   logger,
		scopedVariableRepository: scopedVariableRepository,
		appRepository:            appRepository,
		environmentRepository:    environmentRepository,
		devtronResourceService:   devtronResourceService,
		clusterRepository:        clusterRepository,
	}

	return scopedVariableService, nil
}
func getIdentifierKey(identifierType repository2.IdentifierType, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) int {
	switch identifierType {
	case repository2.ApplicationName:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID]
	case repository2.ClusterName:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]
	case repository2.EnvName:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID]
	default:
		return 0
	}
}
func (impl *ScopedVariableServiceImpl) getIdentifierType(searchableKeyId int) repository2.IdentifierType {
	SearchableKeyIdNameMap := impl.devtronResourceService.GetAllSearchableKeyIdNameMap()
	switch SearchableKeyIdNameMap[searchableKeyId] {
	case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID:
		return repository2.ApplicationName
	case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID:
		return repository2.EnvName
	case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID:
		return repository2.ClusterName
	default:
		return ""
	}
}

func getQualifierId(attributeType repository2.AttributeType) repository2.Qualifier {
	switch attributeType {
	case repository2.ApplicationEnv:
		return repository2.APP_AND_ENV_QUALIFIER
	case repository2.Application:
		return repository2.APP_QUALIFIER
	case repository2.Env:
		return repository2.ENV_QUALIFIER
	case repository2.Cluster:
		return repository2.CLUSTER_QUALIFIER
	case repository2.Global:
		return repository2.GLOBAL_QUALIFIER
	default:
		return 0
	}
}
func getAttributeType(qualifier repository2.Qualifier) repository2.AttributeType {
	switch qualifier {
	case repository2.APP_AND_ENV_QUALIFIER:
		return repository2.ApplicationEnv
	case repository2.APP_QUALIFIER:
		return repository2.Application
	case repository2.ENV_QUALIFIER:
		return repository2.Env
	case repository2.CLUSTER_QUALIFIER:
		return repository2.Cluster
	case repository2.GLOBAL_QUALIFIER:
		return repository2.Global
	default:
		return ""
	}
}

type ValueMapping struct {
	Attribute repository2.AttributeType
	Value     string
}

func complexTypeValidator(payload repository2.Payload) bool {
	for _, variable := range payload.Variables {
		variableType := variable.Definition.DataType
		if variableType == YAML_TYPE || variableType == JSON_TYPE {
			for _, attributeValue := range variable.AttributeValues {
				if attributeValue.VariableValue.Value != "" {
					if variable.Definition.DataType == YAML_TYPE {
						if !isValidYAML(attributeValue.VariableValue.Value.(string)) {
							return false
						}
					} else if variable.Definition.DataType == JSON_TYPE {
						if !isValidJSON(attributeValue.VariableValue.Value.(string)) {
							return false
						}
					}
				} else {
					return false
				}
			}
		}
	}
	return true
}

func isValidYAML(input string) bool {
	jsonInput, err := yaml.YAMLToJSONStrict([]byte(input))
	if err != nil {
		return false
	}
	validJson := isValidJSON(string(jsonInput))
	return validJson
}
func isValidJSON(input string) bool {
	data := make(map[string]interface{})
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return false
	}
	return true
}
func getAuditLog(payload repository2.Payload) sql.AuditLog {
	var auditLog sql.AuditLog
	auditLog = sql.AuditLog{
		CreatedOn: time.Now(),
		CreatedBy: payload.UserId,
		UpdatedOn: time.Now(),
		UpdatedBy: payload.UserId,
	}
	return auditLog
}
func validateVariableToSaveData(data interface{}) (string, error) {
	var value string
	switch data.(type) {
	case json.Number:
		marshal, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		value = string(marshal)
	case string:
		value = data.(string)
		value = "\"" + value + "\""
	case bool:
		value = strconv.FormatBool(data.(bool))
	}
	return value, nil
}
func (impl *ScopedVariableServiceImpl) CreateVariables(payload repository2.Payload) error {
	validValue := complexTypeValidator(payload)
	if !validValue {
		impl.logger.Errorw("variable value is not valid", validValue)
		return fmt.Errorf("invalid variable value")
	}
	err := impl.scopedVariableRepository.DeleteVariables()
	if err != nil {
		return err
	}
	searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
	n := len(payload.Variables)
	if len(payload.Variables) == 0 {
		return nil
	}
	tx, err := impl.scopedVariableRepository.StartTx()
	if err != nil {
		return err
	}
	variableDefinitions := make([]*repository2.VariableDefinition, 0, n)
	for _, variable := range payload.Variables {
		variableDefinition := &repository2.VariableDefinition{
			Name:        variable.Definition.VarName,
			DataType:    variable.Definition.DataType,
			VarType:     variable.Definition.VarType,
			Description: variable.Definition.Description,
			Active:      true,
			AuditLog:    getAuditLog(payload),
		}
		variableDefinitions = append(variableDefinitions, variableDefinition)
	}
	variableScopes := make([]*repository2.VariableScope, 0)
	variableNameToId := make(map[string]int)
	var varDef []*repository2.VariableDefinition
	varDef, err = impl.scopedVariableRepository.CreateVariableDefinition(variableDefinitions, tx)
	for _, variable := range varDef {
		variableNameToId[variable.Name] = variable.Id
	}

	envNameToIdMap := make(map[string]int)
	clusterNameToIdMap := make(map[string]int)
	appNameToIdMap := make(map[string]int)
	appNames, envNames, clusterNames := getAttributeNames(payload)
	appNameToIdMap, envNameToIdMap, clusterNameToIdMap, err = impl.getAttributeNameToIdMappings(appNames, envNames, clusterNames)
	if err != nil {
		return err
	}

	for _, variable := range payload.Variables {

		variableId := variableNameToId[variable.Definition.VarName]
		for _, value := range variable.AttributeValues {
			var compositeString string
			if getQualifierId(value.AttributeType) == 1 {
				compositeString = value.AttributeParams[repository2.ApplicationName] + value.AttributeParams[repository2.EnvName]
			}
			if value.AttributeType == repository2.Global {
				scope := &repository2.VariableScope{
					VariableDefinitionId: variableId,
					QualifierId:          int(getQualifierId(value.AttributeType)),
					Active:               true,
					AuditLog:             getAuditLog(payload),
				}
				variableScopes = append(variableScopes, scope)
			} else {
				for identifierType, IdentifierName := range value.AttributeParams {
					var identifierValue int
					identifierValue, err = getIdentifierValue(identifierType, appNameToIdMap, IdentifierName, envNameToIdMap, clusterNameToIdMap)
					if err != nil {
						return err
					}
					scope := &repository2.VariableScope{
						VariableDefinitionId:  variableId,
						QualifierId:           int(getQualifierId(value.AttributeType)),
						IdentifierKey:         getIdentifierKey(identifierType, searchableKeyNameIdMap),
						IdentifierValueInt:    identifierValue,
						Active:                true,
						CompositeKey:          compositeString,
						IdentifierValueString: IdentifierName,
						AuditLog:              getAuditLog(payload),
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
	variableIdToValueMappings := make(map[int][]*ValueMapping)
	var parentVarScope []*repository2.VariableScope
	var childVarScope []*repository2.VariableScope
	for _, variable := range payload.Variables {
		variableName := variable.Definition.VarName
		if id, ok := variableNameToId[variableName]; ok {
			for _, attrValue := range variable.AttributeValues {
				var value string
				value, err = validateVariableToSaveData(attrValue.VariableValue.Value)
				//switch attrValue.VariableValue.Value.(type) {
				//case json.Number:
				//	marshal, err := json.Marshal(attrValue.VariableValue.Value)
				//	if err != nil {
				//		return err
				//	}
				//	value = string(marshal)
				//case string:
				//	value = attrValue.VariableValue.Value.(string)
				//	value = "\"" + value + "\""
				//case bool:
				//	value = strconv.FormatBool(attrValue.VariableValue.Value.(bool))
				//}

				variableIdToValueMappings[id] = append(variableIdToValueMappings[id], &ValueMapping{
					Attribute: attrValue.AttributeType,
					Value:     value,
				})
			}
		}
	}
	if len(parentVariableScope) > 0 {
		parentVarScope, err = impl.scopedVariableRepository.CreateVariableScope(parentVariableScope, tx)
		if err != nil {
			impl.logger.Errorw("error in getting parentVarScope", parentVarScope)
			return err
		}
	}

	scopeIdToVarData := make(map[int]string)
	for _, parentVar := range parentVarScope {
		if variables, exists := variableIdToValueMappings[parentVar.VariableDefinitionId]; exists {
			for _, varRange := range variables {
				if int(getQualifierId(varRange.Attribute)) == parentVar.QualifierId {
					scopeIdToVarData[parentVar.Id] = varRange.Value
				}
			}

		}
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
			impl.logger.Errorw("error in getting childVarScope", childVarScope)
			return err
		}
	}

	VariableDataList := make([]*repository2.VariableData, 0)

	for scopeId, data := range scopeIdToVarData {
		varData := &repository2.VariableData{

			VariableScopeId: scopeId,
			Data:            data,
			AuditLog:        getAuditLog(payload),
		}
		VariableDataList = append(VariableDataList, varData)
	}
	if len(VariableDataList) > 0 {
		err = impl.scopedVariableRepository.CreateVariableData(VariableDataList, tx)
		if err != nil {
			impl.logger.Errorw("error in saving variable data", parentVarScope)
			return err
		}
	}

	defer func(scopedVariableRepository repository2.ScopedVariableRepository, tx *pg.Tx) {
		err = scopedVariableRepository.CommitTx(tx)
		if err != nil {
			return
		}
	}(impl.scopedVariableRepository, tx)
	return nil
}

func getAttributeNames(payload repository2.Payload) ([]string, []string, []string) {
	appNames := make([]string, 0)
	envNames := make([]string, 0)
	clusterNames := make([]string, 0)
	for _, variable := range payload.Variables {
		for _, value := range variable.AttributeValues {
			for identifierType, _ := range value.AttributeParams {
				if identifierType == repository2.ApplicationName {
					appNames = append(appNames, value.AttributeParams[identifierType])
				} else if identifierType == repository2.EnvName {
					envNames = append(envNames, value.AttributeParams[identifierType])
				} else if identifierType == repository2.ClusterName {
					clusterNames = append(clusterNames, value.AttributeParams[identifierType])
				}
			}

		}
	}
	return appNames, envNames, clusterNames
}

func (impl *ScopedVariableServiceImpl) getAttributeNameToIdMappings(appNames []string, envNames []string, clusterNames []string) (map[string]int, map[string]int, map[string]int, error) {
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

func getIdentifierValue(identifierType repository2.IdentifierType, appNameToIdMap map[string]int, identifierName string, envNameToIdMap map[string]int, clusterNameToIdMap map[string]int) (int, error) {
	var found bool
	var identifierValue int
	if identifierType == repository2.ApplicationName {
		identifierValue, found = appNameToIdMap[identifierName]
		if !found {
			return 0, fmt.Errorf("ApplicationName mapping not found")
		}
	} else if identifierType == repository2.EnvName {
		identifierValue, found = envNameToIdMap[identifierName]
		if !found {
			return 0, fmt.Errorf("EnvName mapping not found")
		}
	} else if identifierType == repository2.ClusterName {
		identifierValue, found = clusterNameToIdMap[identifierName]
		if !found {
			return 0, fmt.Errorf("ClusterName mapping not found")
		}
	} else {
		return 0, fmt.Errorf("invalid identifierType")
	}
	return identifierValue, nil
}
func getPriority(qualifier repository2.Qualifier) int {
	switch qualifier {
	case repository2.APP_AND_ENV_QUALIFIER:
		return 1
	case repository2.APP_QUALIFIER:
		return 2
	case repository2.ENV_QUALIFIER:
		return 3
	case repository2.CLUSTER_QUALIFIER:
		return 4
	case repository2.GLOBAL_QUALIFIER:
		return 5
	default:
		return 0
	}
}

type VariableScopeMapping struct {
	ScopeId int
}

//	func (impl *ScopedVariableServiceImpl) filterMatch(scope *repository2.VariableScope, identifierId int, searchableKeyName bean.DevtronResourceSearchableKeyName, parentRefId int) bool {
//		searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
//		if expectedIdentifierKey, ok := searchableKeyNameIdMap[searchableKeyName]; ok {
//			if scope.IdentifierKey == expectedIdentifierKey && scope.IdentifierValueInt == identifierId {
//				if parentRefId != 0 && scope.ParentIdentifier != parentRefId {
//					return false
//				}
//				return true
//			}
//		} else {
//			if scope.IdentifierKey == 0 && scope.IdentifierValueInt == 0 && int(getQualifierId(repository2.Global)) == scope.QualifierId {
//				return true
//			}
//		}
//		return false
//	}
func customComparator(a, b repository2.Qualifier) bool {
	return getPriority(a) < getPriority(b)
}
func findMinWithComparator(variableScope []*repository2.VariableScope, comparator func(a, b repository2.Qualifier) bool) *repository2.VariableScope {
	if len(variableScope) == 0 {
		panic("variableScope is empty")
	}
	min := variableScope[0]
	for _, val := range variableScope {
		if comparator(repository2.Qualifier(val.QualifierId), repository2.Qualifier(min.QualifierId)) {
			min = val
		}
	}
	return min
}

func (impl *ScopedVariableServiceImpl) getMatchedScopedVariables(varScope []*repository2.VariableScope) map[int]*VariableScopeMapping {
	variableIdToVariableScopes := make(map[int][]*repository2.VariableScope)
	variableScopeMapping := make(map[int]*VariableScopeMapping)
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
		minScope = findMinWithComparator(scopes, customComparator)
		variableScopeMapping[variableId] = &VariableScopeMapping{
			ScopeId: minScope.Id,
		}
	}
	return nil
}

//func (impl *ScopedVariableServiceImpl) getMatchedScopedVariable(varScope []*repository2.VariableScope, scope repository2.Scope) map[int]*VariablePriorityMapping {
//	variablePriorityMap := make(map[int]*VariablePriorityMapping)
//
//	searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
//	if scope.AppId == 0 && scope.EnvId == 0 && scope.ClusterId == 0 {
//		return variablePriorityMap //todo in this case have to return global variable scope id
//	}
//
//	var expectedQualifier int
//	if scope.AppId != 0 && scope.EnvId != 0 {
//		expectedQualifier = repository2.APP_AND_ENV_QUALIFIER
//	} else if scope.AppId != 0 {
//		expectedQualifier = repository2.APP_QUALIFIER
//	} else if scope.EnvId != 0 {
//		expectedQualifier = repository2.ENV_QUALIFIER
//	} else if scope.ClusterId != 0 {
//		expectedQualifier = repository2.CLUSTER_QUALIFIER
//	} else {
//		expectedQualifier = repository2.GLOBAL_QUALIFIER
//	}
//
//	for _, vScope := range varScope {
//		isMatch := false
//		if scope.AppId != 0 && scope.EnvId != 0 && vScope.QualifierId == repository2.APP_AND_ENV_QUALIFIER {
//			isMatch = impl.filterMatch(vScope, scope.AppId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID, 0)
//			if isMatch && (vScope.IdentifierKey != 0 || vScope.IdentifierKey == searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]) {
//				for _, envScope := range varScope {
//					if isMatch = impl.filterMatch(envScope, scope.EnvId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID, vScope.Id); isMatch {
//						break
//					}
//				}
//			}
//		}
//		if !isMatch && scope.AppId != 0 && vScope.QualifierId == repository2.APP_QUALIFIER {
//			isMatch = impl.filterMatch(vScope, scope.AppId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID, 0)
//		}
//		if !isMatch && scope.EnvId != 0 && vScope.QualifierId == repository2.ENV_QUALIFIER {
//			isMatch = impl.filterMatch(vScope, scope.EnvId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID, 0)
//		}
//		if !isMatch && scope.ClusterId != 0 && vScope.QualifierId == repository2.CLUSTER_QUALIFIER {
//			isMatch = impl.filterMatch(vScope, scope.ClusterId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID, 0)
//		}
//		if !isMatch && vScope.QualifierId == repository2.GLOBAL_QUALIFIER {
//			isMatch = impl.filterMatch(vScope, 0, "", 0)
//		}
//		if isMatch {
//			priority, ok := variablePriorityMap[vScope.VariableDefinitionId]
//			currentPriority := getPriority(vScope.QualifierId)
//			if !ok || (priority.Priority > currentPriority && currentPriority >= getPriority(expectedQualifier)) {
//				variablePriorityMap[vScope.VariableDefinitionId] = &VariablePriorityMapping{
//					ScopeId:  vScope.Id,
//					Priority: currentPriority,
//				}
//			}
//		}
//	}
//	return variablePriorityMap
//}

func (impl *ScopedVariableServiceImpl) GetScopedVariables(scope repository2.Scope, varNames []string) (scopedVariableDataObj []*ScopedVariableData, err error) {
	var vDef []*repository2.VariableDefinition
	var varIds []int
	vDef, err = impl.scopedVariableRepository.GetVariablesByNames(varNames)
	for _, def := range vDef {
		varIds = append(varIds, def.Id)
	}
	var env *repository.Environment
	if scope.EnvId != 0 && scope.ClusterId == 0 {
		env, err = impl.environmentRepository.FindById(scope.EnvId)
		if err != nil {
			return nil, err
		}
		scope.ClusterId = env.ClusterId
	}

	searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
	var varScope []*repository2.VariableScope
	var scopedVariableIds map[int]*VariableScopeMapping
	scopeIdToVariableScope := make(map[int]*repository2.VariableScope)
	varScope, err = impl.scopedVariableRepository.GetScopedVariableData(scope, searchableKeyNameIdMap, varIds)
	if err != nil {
		return nil, err
	}
	for _, vScope := range varScope {
		scopeIdToVariableScope[vScope.Id] = vScope
	}
	scopedVariableIds = impl.getMatchedScopedVariables(varScope)

	var scopeIds, scopedVarIds []int
	for varId, mapping := range scopedVariableIds {
		scopeIds = append(scopeIds, mapping.ScopeId)
		scopedVarIds = append(scopedVarIds, varId)
	}
	var varDefs []*repository2.VariableDefinition
	if scopedVarIds != nil {
		varDefs, err = impl.scopedVariableRepository.GetVariablesForVarIds(scopedVarIds)
		if err != nil {
			return nil, err
		}
	}
	var vData []*repository2.VariableData
	if scopeIds != nil {
		vData, err = impl.scopedVariableRepository.GetDataForScopeIds(scopeIds)
		if err != nil {
			return nil, err
		}
	}

	for varId, mapping := range scopedVariableIds {
		scopedVariableData := &ScopedVariableData{}
		for _, varDef := range varDefs {
			if varDef.Id == varId {
				scopedVariableData.VariableName = varDef.Name
				break
			}
		}
		for _, varData := range vData {
			if varData.VariableScopeId == mapping.ScopeId {
				var value interface{}
				value, err = variableDataTypeValidation(varData.Data)
				if err != nil {
					return nil, err
				}
				scopedVariableData.VariableValue = value
			}
		}
		scopedVariableDataObj = append(scopedVariableDataObj, scopedVariableData)
	}
	var variableList []*repository2.VariableDefinition
	if varNames == nil {
		variableList, err = impl.scopedVariableRepository.GetAllVariables()
		if err != nil {
			return nil, err
		}
		for _, existing := range variableList {
			found := false
			for _, variable := range scopedVariableDataObj {
				if variable.VariableName == existing.Name {
					found = true
					break
				}
			}
			if !found {
				newData := &ScopedVariableData{
					VariableName: existing.Name,
				}
				scopedVariableDataObj = append(scopedVariableDataObj, newData)
			}
		}
	}
	return scopedVariableDataObj, err
}
func getSchema() (string, error) {
	schema := jsonschema.Reflect(repository2.Payload{})
	schemaData, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(schemaData), nil
}
func variableDataTypeValidation(Data string) (interface{}, error) {
	var value interface{}
	if intValue, err := strconv.Atoi(Data); err == nil {
		value = intValue
	} else if floatValue, err := strconv.ParseFloat(Data, 64); err == nil {
		value = floatValue
	} else if boolValue, err := strconv.ParseBool(Data); err == nil {
		value = boolValue
	} else {
		value = strings.Trim(Data, "\"")
	}
	return value, nil
}
func (impl *ScopedVariableServiceImpl) GetJsonForVariables() (*repository2.Payload, string, error) {
	dataForJson, err := impl.scopedVariableRepository.GetAllVariableScopeAndDefinition()
	if err != nil {
		return nil, "", err
	}
	payload := &repository2.Payload{
		Variables: make([]*repository2.Variables, 0),
	}

	variables := make([]*repository2.Variables, 0)
	for _, data := range dataForJson {
		definition := repository2.Definition{
			VarName:     data.Name,
			DataType:    data.DataType,
			VarType:     data.VarType,
			Description: data.Description,
		}
		attributes := make([]repository2.AttributeValue, 0)

		scopeIdToVarScopes := make(map[int][]*repository2.VariableScope)
		for _, scope := range data.VariableScope {
			if scope.ParentIdentifier != 0 {
				scopeIdToVarScopes[scope.ParentIdentifier] = append(scopeIdToVarScopes[scope.ParentIdentifier], scope)
			} else {
				scopeIdToVarScopes[scope.Id] = []*repository2.VariableScope{scope}
			}
		}
		for parentScopeId, scopes := range scopeIdToVarScopes {
			attribute := repository2.AttributeValue{
				AttributeParams: make(map[repository2.IdentifierType]string),
			}
			for _, scope := range scopes {
				if impl.getIdentifierType(scope.IdentifierKey) != "" {
					attribute.AttributeParams[impl.getIdentifierType(scope.IdentifierKey)] = scope.IdentifierValueString
				}
				if parentScopeId == scope.Id {
					var value interface{}
					value, err = variableDataTypeValidation(scope.VariableData.Data)
					if err != nil {
						return nil, "", err
					}
					//if intValue, err := strconv.Atoi(scope.VariableData.Data); err == nil {
					//	value = intValue
					//} else if floatValue, err := strconv.ParseFloat(scope.VariableData.Data, 64); err == nil {
					//	value = floatValue
					//} else if boolValue, err := strconv.ParseBool(scope.VariableData.Data); err == nil {
					//	value = boolValue
					//} else {
					//	value = strings.Trim(scope.VariableData.Data, "\"")
					//}
					attribute.VariableValue = repository2.VariableValue{
						Value: value,
					}
					attribute.AttributeType = getAttributeType(repository2.Qualifier(scope.QualifierId))
				}
			}
			if len(attribute.AttributeParams) == 0 {
				attribute.AttributeParams = nil
			}
			attributes = append(attributes, attribute)
		}

		variable := &repository2.Variables{
			Definition:      definition,
			AttributeValues: attributes,
		}
		variables = append(variables, variable)
	}
	jsonSchema, err := getSchema()
	if err != nil {
		return nil, "", nil
	}
	payload.Variables = variables
	payload.SpecVersion = "v1"
	if len(payload.Variables) == 0 {
		return nil, jsonSchema, nil
	}
	return payload, jsonSchema, nil
}
