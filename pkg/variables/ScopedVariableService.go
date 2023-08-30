package variables

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	repository2 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"regexp"
	"sigs.k8s.io/yaml"
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
	CreateVariables(payload models.Payload) error
	GetScopedVariables(scope models.Scope, varNames []string) (scopedVariableDataObj []*ScopedVariableData, err error)
	GetJsonForVariables() (*models.Payload, error)
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
func getIdentifierKey(identifierType models.IdentifierType, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) int {
	switch identifierType {
	case models.ApplicationName:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID]
	case models.ClusterName:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]
	case models.EnvName:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID]
	default:
		return 0
	}
}
func (impl *ScopedVariableServiceImpl) getIdentifierType(searchableKeyId int) models.IdentifierType {
	SearchableKeyIdNameMap := impl.devtronResourceService.GetAllSearchableKeyIdNameMap()
	switch SearchableKeyIdNameMap[searchableKeyId] {
	case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID:
		return models.ApplicationName
	case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID:
		return models.EnvName
	case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID:
		return models.ClusterName
	default:
		return ""
	}
}

func getQualifierId(attributeType models.AttributeType) repository2.Qualifier {
	switch attributeType {
	case models.ApplicationEnv:
		return repository2.APP_AND_ENV_QUALIFIER
	case models.Application:
		return repository2.APP_QUALIFIER
	case models.Env:
		return repository2.ENV_QUALIFIER
	case models.Cluster:
		return repository2.CLUSTER_QUALIFIER
	case models.Global:
		return repository2.GLOBAL_QUALIFIER
	default:
		return 0
	}
}
func getAttributeType(qualifier repository2.Qualifier) models.AttributeType {
	switch qualifier {
	case repository2.APP_AND_ENV_QUALIFIER:
		return models.ApplicationEnv
	case repository2.APP_QUALIFIER:
		return models.Application
	case repository2.ENV_QUALIFIER:
		return models.Env
	case repository2.CLUSTER_QUALIFIER:
		return models.Cluster
	case repository2.GLOBAL_QUALIFIER:
		return models.Global
	default:
		return ""
	}
}

type ValueMapping struct {
	Attribute models.AttributeType
	Value     string
}

func complexTypeValidator(payload models.Payload) bool {
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
func getAuditLog(payload models.Payload) sql.AuditLog {
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
func (impl *ScopedVariableServiceImpl) CreateVariables(payload models.Payload) error {
	//validValue := complexTypeValidator(payload)
	//if !validValue {
	//	impl.logger.Errorw("variable value is not valid", validValue)
	//	return fmt.Errorf("invalid variable value")
	//}
	err := validateVariableScopeRequest(payload)
	if err != nil {
		return fmt.Errorf("custom validation err in CreateVariables")
	}
	err = impl.scopedVariableRepository.DeleteVariables()
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
				compositeString = value.AttributeParams[models.ApplicationName] + value.AttributeParams[models.EnvName]
			}
			if value.AttributeType == models.Global {
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
				if err != nil {
					return err
				}
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

func getAttributeNames(payload models.Payload) ([]string, []string, []string) {
	appNames := make([]string, 0)
	envNames := make([]string, 0)
	clusterNames := make([]string, 0)
	for _, variable := range payload.Variables {
		for _, value := range variable.AttributeValues {
			for identifierType, _ := range value.AttributeParams {
				if identifierType == models.ApplicationName {
					appNames = append(appNames, value.AttributeParams[identifierType])
				} else if identifierType == models.EnvName {
					envNames = append(envNames, value.AttributeParams[identifierType])
				} else if identifierType == models.ClusterName {
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

func getIdentifierValue(identifierType models.IdentifierType, appNameToIdMap map[string]int, identifierName string, envNameToIdMap map[string]int, clusterNameToIdMap map[string]int) (int, error) {
	var found bool
	var identifierValue int
	if identifierType == models.ApplicationName {
		identifierValue, found = appNameToIdMap[identifierName]
		if !found {
			return 0, fmt.Errorf("ApplicationName mapping not found")
		}
	} else if identifierType == models.EnvName {
		identifierValue, found = envNameToIdMap[identifierName]
		if !found {
			return 0, fmt.Errorf("EnvName mapping not found")
		}
	} else if identifierType == models.ClusterName {
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

func customComparator(a, b repository2.Qualifier) bool {
	return getPriority(a) < getPriority(b)
}
func findMinWithComparator(variableScope []*repository2.VariableScope, comparator func(a, b repository2.Qualifier) bool) *repository2.VariableScope {
	if len(variableScope) == 0 {
		return nil
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
	return variableScopeMapping
}
func (impl *ScopedVariableServiceImpl) GetScopedVariables(scope models.Scope, varNames []string) (scopedVariableDataObj []*ScopedVariableData, err error) {
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
func (impl *ScopedVariableServiceImpl) GetJsonForVariables() (*models.Payload, error) {
	dataForJson, err := impl.scopedVariableRepository.GetAllVariableScopeAndDefinition()
	if err != nil {
		return nil, err
	}
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
				if impl.getIdentifierType(scope.IdentifierKey) != "" {
					attribute.AttributeParams[impl.getIdentifierType(scope.IdentifierKey)] = scope.IdentifierValueString
				}
				if parentScopeId == scope.Id {
					var value interface{}
					value, err = variableDataTypeValidation(scope.VariableData.Data)
					if err != nil {
						return nil, err
					}
					attribute.VariableValue = models.VariableValue{
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
func getIdentifierType(attribute models.AttributeType) []models.IdentifierType {
	switch attribute {
	case models.ApplicationEnv:
		return []models.IdentifierType{models.ApplicationName, models.EnvName}
	case models.Application:
		return []models.IdentifierType{models.ApplicationName}
	case models.Env:
		return []models.IdentifierType{models.EnvName}
	case models.Cluster:
		return []models.IdentifierType{models.ClusterName}
	default:
		return nil
	}
}

func validateVariableScopeRequest(payload models.Payload) error {
	variableNamesList := make([]string, 0)
	for _, variable := range payload.Variables {
		if slices.Contains(variableNamesList, variable.Definition.VarName) {
			return fmt.Errorf("duplicate variable name")
		}
		exp := `^[a-zA-Z0-9_-]{1,64}$`
		rExp := regexp.MustCompile(exp)
		if !rExp.MatchString(variable.Definition.VarName) {
			return fmt.Errorf("invalid variable name %s", variable.Definition.VarName)
		}
		if variable.Definition.VarName[0] == '_' ||
			variable.Definition.VarName[0] == '-' ||
			variable.Definition.VarName[len(variable.Definition.VarName)-1] == '_' ||
			variable.Definition.VarName[len(variable.Definition.VarName)-1] == '-' {
			return fmt.Errorf("invalid variable name")
		}
		variableNamesList = append(variableNamesList, variable.Definition.VarName)
		uniqueVariableMap := make(map[string]interface{})
		for _, attributeValue := range variable.AttributeValues {
			validIdentifierTypeList := getIdentifierType(attributeValue.AttributeType)
			if len(validIdentifierTypeList) != len(attributeValue.AttributeParams) {
				return fmt.Errorf("length of AttributeParams is not valid")
			}
			for key, _ := range attributeValue.AttributeParams {
				if !slices.Contains(validIdentifierTypeList, key) {
					return fmt.Errorf("invalid IdentifierType %s for validIdentifierTypeList %s", key, validIdentifierTypeList)
				}
				match := false
				for _, identifier := range models.IdentifiersList {
					if identifier == key {
						match = true
					}
				}
				if !match {
					return fmt.Errorf("invalid identifier key %s for variable %s", key, variable.Definition.VarName)
				}
			}
			identifierString := fmt.Sprintf("%s-%s", variable.Definition.VarName, string(attributeValue.AttributeType))
			for _, key := range validIdentifierTypeList {
				identifierString = fmt.Sprintf("%s-%s", identifierString, attributeValue.AttributeParams[key])
			}
			if _, ok := uniqueVariableMap[identifierString]; ok {
				return fmt.Errorf("duplicate AttributeParams found for AttributeType %v", attributeValue.AttributeType)
			}
			uniqueVariableMap[identifierString] = attributeValue.VariableValue.Value
		}
	}
	return nil
}
