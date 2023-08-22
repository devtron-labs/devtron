package variables

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/sql"
	repository2 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"log"
	"time"
)

type ScopedVariableData struct {
	VariableName  string `json:"variableName"`
	VariableValue string `json:"variableValue"`
}

type ScopedVariableService interface {
	CreateVariables(payload repository2.Payload) error
	GetScopedVariables(scope Scope, varNames []string) (scopedVariableDataObj []*ScopedVariableData, err error)
}

type ScopedVariableServiceImpl struct {
	logger                   *zap.SugaredLogger
	scopedVariableRepository repository2.ScopedVariableRepository
	appRepository            app.AppRepository
	environmentRepository    repository.EnvironmentRepository
	devtronResourceService   devtronResource.DevtronResourceService
	clusterRepository        repository.ClusterRepository
	pipelineBuilder          pipeline.PipelineBuilder
}

func NewScopedVariableServiceImpl(logger *zap.SugaredLogger, scopedVariableRepository repository2.ScopedVariableRepository, appRepository app.AppRepository, environmentRepository repository.EnvironmentRepository, devtronResourceService devtronResource.DevtronResourceService, clusterRepository repository.ClusterRepository, pipelineBuilder pipeline.PipelineBuilder) (*ScopedVariableServiceImpl, error) {
	scopedVariableService := &ScopedVariableServiceImpl{
		logger:                   logger,
		scopedVariableRepository: scopedVariableRepository,
		appRepository:            appRepository,
		environmentRepository:    environmentRepository,
		devtronResourceService:   devtronResourceService,
		clusterRepository:        clusterRepository,
		pipelineBuilder:          pipelineBuilder,
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

func getQualifierId(attributeType repository2.AttributeType) int {
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

type ValueMapping struct {
	Attribute repository2.AttributeType
	Value     string
}

func (impl *ScopedVariableServiceImpl) CreateVariables(payload repository2.Payload) error {
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
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: payload.UserId,
				UpdatedOn: time.Now(),
				UpdatedBy: payload.UserId,
			},
		}
		variableDefinitions = append(variableDefinitions, variableDefinition)
	}
	variableScopes := make([]*repository2.VariableScope, 0)
	variableNameToIdMap := make(map[string]int)
	var vardef []*repository2.VariableDefinition
	vardef, err = impl.scopedVariableRepository.CreateVariableDefinition(variableDefinitions, tx)
	for _, variable := range vardef {
		variableNameToIdMap[variable.Name] = variable.Id
	}
	appNames := make([]string, 0)
	envNames := make([]string, 0)
	clusterNames := make([]string, 0)
	envNameToIdMap := make(map[string]int)
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
	var appNameToId []*app.App
	if len(appNames) != 0 {
		appNameToId, err = impl.appRepository.FindIdsByNamesForScopedVariables(appNames)
		if err != nil {
			impl.logger.Errorw("error in getting appNameToId", err)
			return err
		}
	}
	appNameToIdMap := make(map[string]int)
	for _, name := range appNameToId {
		appNameToIdMap[name.AppName] = name.Id
	}
	var envNameToId []*repository.Environment
	if len(envNames) != 0 {
		envNameToId, err = impl.environmentRepository.FindIdsAndNamesByNames(envNames)
		if err != nil {
			impl.logger.Errorw("error in getting envNameToIdMap", err)
			return err
		}
	}
	for _, name := range envNameToId {
		envNameToIdMap[name.Name] = name.Id
	}
	var clusterNameToId []*repository.Cluster
	if len(clusterNames) != 0 {
		clusterNameToId, err = impl.clusterRepository.FindIdsAndNamesByNames(clusterNames)
		if err != nil {
			impl.logger.Errorw("error in getting clusterNameToId", err)
			return err
		}

	}

	clusterNameToIdMap := make(map[string]int)
	for _, name := range clusterNameToId {
		clusterNameToIdMap[name.ClusterName] = name.Id
	}

	for _, variable := range payload.Variables {
		variableId := variableNameToIdMap[variable.Definition.VarName]
		for _, value := range variable.AttributeValues {
			var compositeKey string
			if getQualifierId(value.AttributeType) == 1 {
				compositeKey = value.AttributeParams[repository2.ApplicationName] + value.AttributeParams[repository2.EnvName]
			}
			for identifierType, s := range value.AttributeParams {
				var identifierValue int
				if identifierType == repository2.ApplicationName {
					identifierValue = appNameToIdMap[s]
				} else if identifierType == repository2.EnvName {
					identifierValue = envNameToIdMap[s]
				} else if identifierType == repository2.ClusterName {
					identifierValue = clusterNameToIdMap[s]
				}
				scope := &repository2.VariableScope{
					VariableDefinitionId:  variableId,
					QualifierId:           getQualifierId(value.AttributeType),
					IdentifierKey:         getIdentifierKey(identifierType, searchableKeyNameIdMap),
					IdentifierValueInt:    identifierValue,
					Active:                true,
					CompositeKey:          compositeKey,
					IdentifierValueString: s,
					AuditLog: sql.AuditLog{
						CreatedOn: time.Now(),
						CreatedBy: payload.UserId,
						UpdatedOn: time.Now(),
						UpdatedBy: payload.UserId,
					},
				}
				variableScopes = append(variableScopes, scope)
			}
		}
	}
	parentVariableDefinition := make([]*repository2.VariableScope, 0)
	childrenVariableDefinition := make([]*repository2.VariableScope, 0)
	parentScopesMap := make(map[string]*repository2.VariableScope)

	for _, scope := range variableScopes {
		if scope.QualifierId == 1 && scope.IdentifierKey == searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID] {
			childrenVariableDefinition = append(childrenVariableDefinition, scope)
		} else {
			parentVariableDefinition = append(parentVariableDefinition, scope)
			if scope.QualifierId == 1 {
				parentScopesMap[scope.CompositeKey] = scope
			}

		}
	}
	variableNameToVariableValueMap := make(map[int][]*ValueMapping)
	var parentVarScope []*repository2.VariableScope
	var childVarScope []*repository2.VariableScope
	for _, variable := range payload.Variables {
		variableName := variable.Definition.VarName
		if id, exists := variableNameToIdMap[variableName]; exists {
			for _, attrValue := range variable.AttributeValues {
				variableNameToVariableValueMap[id] = append(variableNameToVariableValueMap[id], &ValueMapping{
					Attribute: attrValue.AttributeType,
					Value:     attrValue.VariableValue.Value,
				})
			}
		}
	}
	parentVarScope, err = impl.scopedVariableRepository.CreateVariableScope(parentVariableDefinition, tx)
	if err != nil {
		impl.logger.Errorw("error in getting parentVarScope", parentVarScope)
		return err
	}
	scopeIdToVarData := make(map[int]string)
	for _, parentvar := range parentVarScope {
		if variables, exists := variableNameToVariableValueMap[parentvar.VariableDefinitionId]; exists {
			for _, varRange := range variables {
				tt := getQualifierId(varRange.Attribute) == parentvar.QualifierId
				log.Print(tt)
				if getQualifierId(varRange.Attribute) == parentvar.QualifierId {
					scopeIdToVarData[parentvar.Id] = varRange.Value
				}
			}

		}
	}
	for _, childScope := range childrenVariableDefinition {
		parentScope, exists := parentScopesMap[childScope.CompositeKey]
		if exists {
			childScope.ParentIdentifier = parentScope.Id
		}
	}
	if childVarScope != nil {
		childVarScope, err = impl.scopedVariableRepository.CreateVariableScope(childrenVariableDefinition, tx)

		if err != nil {
			impl.logger.Errorw("error in getting childVarScope", childVarScope)
			return err
		}
	}

	variableDatas := make([]*repository2.VariableData, 0)

	for scopeId, data := range scopeIdToVarData {
		varData := &repository2.VariableData{
			VariableScopeId: scopeId,
			Data:            data,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: payload.UserId,
				UpdatedOn: time.Now(),
				UpdatedBy: payload.UserId,
			},
		}
		variableDatas = append(variableDatas, varData)
	}
	err = impl.scopedVariableRepository.CreateVariableData(variableDatas, tx)
	defer func(scopedVariableRepository repository2.ScopedVariableRepository, tx *pg.Tx) {
		err = scopedVariableRepository.CommitTx(tx)
		if err != nil {
			return
		}
	}(impl.scopedVariableRepository, tx)
	return nil
}
func getPriority(qualifier int) int {
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

type VariablePriorityMapping struct {
	ScopeId  int
	Priority int
}

func (impl *ScopedVariableServiceImpl) filterMatch(scope *repository2.VariableScope, identifierId int, searchableKeyName bean.DevtronResourceSearchableKeyName, parentRefId int) bool {
	searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
	if identifierId != 0 {
		if (scope.IdentifierKey == searchableKeyNameIdMap[searchableKeyName] && scope.IdentifierValueInt == identifierId) || scope.IdentifierKey == 0 {
			if parentRefId != 0 && scope.ParentIdentifier != parentRefId {
				return false
			}
			return true
		} else {
			return false
		}
	}
	return false
}
func (impl *ScopedVariableServiceImpl) getMatchedScopedVariable(varScope []*repository2.VariableScope, appId, envId, clusterId int) map[int]*VariablePriorityMapping {
	variablePriorityMap := make(map[int]*VariablePriorityMapping)
	searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()

	if appId == 0 && envId == 0 && clusterId == 0 {
		return variablePriorityMap
	}
	for _, scope := range varScope {
		isMatch := false
		if appId != 0 && envId != 0 {
			isMatch = impl.filterMatch(scope, appId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID, 0)
			if isMatch && (scope.IdentifierKey != 0 || scope.IdentifierKey == searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]) {
				for _, envScope := range varScope {
					if isMatch = impl.filterMatch(envScope, envId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID, scope.Id); isMatch {
						break
					}
				}
			}
		}
		if !isMatch && appId != 0 {
			isMatch = impl.filterMatch(scope, appId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID, 0)
		}
		if !isMatch && envId != 0 {
			isMatch = impl.filterMatch(scope, envId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID, 0)
		}
		if !isMatch && clusterId != 0 {
			isMatch = impl.filterMatch(scope, clusterId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID, 0)
		}
		if isMatch {
			priority, ok := variablePriorityMap[scope.VariableDefinitionId]
			currentPriority := getPriority(scope.QualifierId)
			if !ok || priority.Priority > currentPriority {
				variablePriorityMap[scope.VariableDefinitionId] = &VariablePriorityMapping{
					ScopeId:  scope.Id,
					Priority: currentPriority,
				}
			}
		}
	}
	return variablePriorityMap
}

type Scope struct {
	AppId     int `json:"appId"`
	EnvId     int `json:"env_id"`
	ClusterId int `json:"clusterId"`
}

func (impl *ScopedVariableServiceImpl) GetScopedVariables(scope Scope, varNames []string) (scopedVariableDataObj []*ScopedVariableData, err error) {
	var varDef []*repository2.VariableDefinition
	var varIds []int
	varDef, err = impl.scopedVariableRepository.GetVariablesByNames(varNames)
	for _, def := range varDef {
		varIds = append(varIds, def.Id)
	}
	searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
	var varScope []*repository2.VariableScope
	var scopedVariableIds map[int]*VariablePriorityMapping
	scopeIdToVariableScope := make(map[int]*repository2.VariableScope)
	if varIds != nil {
		varScope, err = impl.scopedVariableRepository.GetScopedVariableDataForVarIds(scope.AppId, scope.EnvId, scope.ClusterId, searchableKeyNameIdMap, varIds)
		if err != nil {
			return nil, err
		}
	} else {
		varScope, err = impl.scopedVariableRepository.GetScopedVariableData(scope.AppId, scope.EnvId, scope.ClusterId, searchableKeyNameIdMap)
		if err != nil {
			return nil, err
		}
	}
	for _, scope := range varScope {
		scopeIdToVariableScope[scope.Id] = scope
	}
	scopedVariableIds = impl.getMatchedScopedVariable(varScope, scope.AppId, scope.EnvId, scope.ClusterId)

	var scopeIds, scopedVarIds []int
	for varId, mapping := range scopedVariableIds {
		scopeIds = append(scopeIds, mapping.ScopeId)
		scopedVarIds = append(scopedVarIds, varId)
	}
	varDefs, err := impl.scopedVariableRepository.GetVariablesForVarIds(scopedVarIds)
	if err != nil {
		return nil, err
	}
	vData, err := impl.scopedVariableRepository.GetDataForScopeIds(scopeIds)
	if err != nil {
		return nil, err
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
				scopedVariableData.VariableValue = varData.Data
			}
		}
		scopedVariableDataObj = append(scopedVariableDataObj, scopedVariableData)
	}
	return scopedVariableDataObj, err

}
