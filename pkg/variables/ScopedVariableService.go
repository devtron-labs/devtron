package variables

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"go.uber.org/zap"
)

type ScopedVariables struct {
	CreateScopedVariableDTO []*CreateScopedVariableDTO `json:"createScopedVariableDTO"`
}
type CreateScopedVariableDTO struct {
	VariableDefinition
	ApplicationEnvironmentScope []*ApplicationEnvironmentScope `json:"applicationEnvironmentScope"`
	ApplicationScope            []*ApplicationScope            `json:"applicationScope"`
	EnvironmentScope            []*EnvironmentScope            `json:"environmentScope"`
	ClusterScope                []*ClusterScope                `json:"clusterScope"`
	GlobalScope                 *ScopedVariableValue           `json:"globalScope"`
}
type ScopedVariableValue struct {
	Value string `json:"value"`
}

type ApplicationEnvironmentScope struct {
	ScopedVariableValue
	ApplicationName string `json:"applicationName"`
	EnvironmentName string `json:"environmentName"`
}

type ApplicationScope struct {
	ScopedVariableValue
	ApplicationName string `json:"applicationName"`
}
type EnvironmentScope struct {
	ScopedVariableValue
	EnvironmentName string `json:"environmentName"`
}
type ClusterScope struct {
	ScopedVariableValue
	ClusterName string `json:"clusterName"`
}

type VariableDefinition struct {
	Id          int    `json:"id"`
	Name        string `json:"varName"`
	DataType    string `json:"dataType"`
	VarType     string `json:"varType"`
	Description string `json:"description"`
}

type VariableScope struct {
	Id                    int                                   `json:"id"`
	VariableId            int                                   `json:"variableId"`
	QualifierId           int                                   `json:"qualifierId"`
	IdentifierKey         bean.DevtronResourceSearchableKeyName `json:"identifierKey"`
	IdentifierValueInt    int                                   `json:"identifierValueInt"`
	Active                bool                                  `json:"active"`
	IdentifierValueString string                                `json:"identifierValueString"`
}

type VariableData struct {
	Id      int    `json:"id"`
	ScopeId int    `json:"scopeId"`
	Active  bool   `json:"active"`
	Data    string `json:"data"`
}
type ScopedVariableData struct {
	VariableName  string `json:"variableName"`
	VariableValue string `json:"variableValue"`
}

type ScopedVariableService interface {
	//SaveVariables(payload Payload) error
	GetScopedVariables(appId, envId, clusterId int, varIds []int) (scopedVariableDataObj []*ScopedVariableData, err error)
}

const (
	APP_AND_ENV_QUALIFIER = 1
	APP_QUALIFIER         = 2
	ENV_QUALIFIER         = 3
	CLUSTER_QUALIFIER     = 4
	GLOBAL_QUALIFIER      = 5
)

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

//	type Payload struct {
//		Variables []*Variables `json:"Variables"`
//	}
//
//	type Variables struct {
//		Definition      Definition       `json:"definition"`
//		AttributeValues []AttributeValue `json:"attributeValue"`
//	}
//
//	type AttributeValue struct {
//		VariableValue
//		AttributeType   AttributeType
//		AttributeParams map[IdentifierType]string
//	}
//
//	type Definition struct {
//		VarName     string `json:"varName"`
//		DataType    string `json:"dataType" validate:"oneof=json yaml primitive"`
//		VarType     string `json:"varType" validate:"oneof=private public"`
//		Description string `json:"description"`
//	}
//
// type AttributeType string
//
// const (
//
//	ApplicationEnv AttributeType = "ApplicationEnv"
//	Application    AttributeType = "Application"
//	Env            AttributeType = "Env"
//	Cluster        AttributeType = "Cluster"
//	Global         AttributeType = "Global"
//
// )
//
// type IdentifierType string
//
// const (
//
//	ApplicationName IdentifierType = "ApplicationName"
//	EnvName         IdentifierType = "EnvName"
//	ClusterName     IdentifierType = "ClusterName"
//
// )
//
//	type VariableValue struct {
//		Value string `json:"value"`
//	}
//
//	func (impl *ScopedVariableServiceImpl) SaveVariables(payload Payload) error {
//		searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
//		n := len(payload.Variables)
//		m := 0
//		if len(payload.Variables) == 0 {
//			return nil
//		}
//		tx, err := impl.scopedVariableRepository.StartTx()
//		if err != nil {
//			return err
//		}
//		variableDefinitions := make([]*scopedVariable.VariableDefinition, 0, n)
//		for _, variable := range payload.Variables {
//			variableDefinition := &scopedVariable.VariableDefinition{
//				Name:        variable.Definition.VarName,
//				DataType:    variable.Definition.DataType,
//				VarType:     variable.Definition.VarType,
//				Description: variable.Definition.Description,
//				Active:      true,
//				//AuditLog:&sql.AuditLog{
//				//	CreatedBy: use
//				//},
//			}
//			variableDefinitions = append(variableDefinitions, variableDefinition)
//			m += len(variable.ApplicationScope) + len(variable.ApplicationEnvironmentScope) + len(variable.EnvironmentScope) + len(variable.ClusterScope)
//			if variable.GlobalScope != nil {
//				m += 1
//			}
//		}
//
//		//assuming variables have unique names
//		variableNameToIdMap := make(map[string]int)
//		for _, variable := range variableDefinitions {
//			variableNameToIdMap[variable.Name] = variable.Id
//		}
//
//		variableScopes := make([]*scopedVariable.VariableScope, 0, m)
//		appNames := make([]string, 0)
//		err = impl.scopedVariableRepository.CreateVariableDefinition(variableDefinitions, tx)
//		for _, variable := range scopedVariables.CreateScopedVariableDTO {
//			variableId := variableNameToIdMap[variable.Name]
//			for _, appScopes := range variable.ApplicationScope {
//				appNames = append(appNames, appScopes.ApplicationName)
//			}
//			appNameToIdMap, err := impl.appRepository.FindIdsByNamesForScopedVariables(appNames)
//			if err != nil {
//				return err
//			}
//			for _, appScopes := range variable.ApplicationScope {
//				variableScopes = append(variableScopes, &scopedVariable.VariableScope{
//					VariableId:            variableId,
//					QualifierId:           APP_QUALIFIER,
//					IdentifierKey:         searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID],
//					IdentifierValueInt:    appNameToIdMap[appScopes.ApplicationName],
//					Active:                true,
//					IdentifierValueString: appScopes.ApplicationName,
//				})
//			}
//			appNames = make([]string, 0)
//			envNames := make([]string, 0)
//			var envNameToIdMap map[string]int
//			for _, appEnvScopes := range variable.ApplicationEnvironmentScope {
//				appNames = append(appNames, appEnvScopes.ApplicationName)
//				envNames = append(envNames, appEnvScopes.EnvironmentName)
//			}
//			appNameToIdMap, err = impl.appRepository.FindIdsByNamesForScopedVariables(appNames)
//			envNameToIdMap, err = impl.environmentRepository.FindIdsAndNamesByNames(envNames)
//			for _, appEnvScopes := range variable.ApplicationEnvironmentScope {
//				variableScopes = append(variableScopes, &scopedVariable.VariableScope{
//					VariableId:            variableId,
//					QualifierId:           APP_AND_ENV_QUALIFIER,
//					IdentifierKey:         searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID],
//					IdentifierValueInt:    appNameToIdMap[appEnvScopes.ApplicationName],
//					Active:                true,
//					IdentifierValueString: appEnvScopes.ApplicationName,
//				})
//			}
//			for _, appEnvScopes := range variable.ApplicationEnvironmentScope {
//				variableScopes = append(variableScopes, &scopedVariable.VariableScope{
//					VariableId:            variableId,
//					QualifierId:           APP_AND_ENV_QUALIFIER,
//					IdentifierKey:         searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID],
//					IdentifierValueInt:    envNameToIdMap[appEnvScopes.EnvironmentName],
//					Active:                true,
//					IdentifierValueString: appEnvScopes.EnvironmentName,
//				})
//			}
//			envNames = make([]string, 0)
//
//			for _, envScopes := range variable.EnvironmentScope {
//				envNames = append(envNames, envScopes.EnvironmentName)
//			}
//			envNameToIdMap, err = impl.environmentRepository.FindIdsAndNamesByNames(envNames)
//			for _, envScopes := range variable.EnvironmentScope {
//				variableScopes = append(variableScopes, &scopedVariable.VariableScope{
//					VariableId:            variableId,
//					QualifierId:           ENV_QUALIFIER,
//					IdentifierKey:         searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID],
//					IdentifierValueInt:    envNameToIdMap[envScopes.EnvironmentName],
//					Active:                true,
//					IdentifierValueString: envScopes.EnvironmentName,
//				})
//			}
//			clusterNames := make([]string, 0)
//
//			for _, clusterScopes := range variable.ClusterScope {
//				clusterNames = append(clusterNames, clusterScopes.ClusterName)
//			}
//			envNameToIdMap, err = impl.clusterRepository.FindIdsAndNamesByNames(clusterNames)
//			for _, clusterScopes := range variable.ClusterScope {
//				variableScopes = append(variableScopes, &scopedVariable.VariableScope{
//					VariableId:            variableId,
//					QualifierId:           CLUSTER_QUALIFIER,
//					IdentifierKey:         searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID],
//					IdentifierValueInt:    envNameToIdMap[clusterScopes.ClusterName],
//					Active:                true,
//					IdentifierValueString: clusterScopes.ClusterName,
//				})
//			}
//
//		}
//		//for _, globalScopes := range GlobalScope {
//		//	variableScopes = append(variableScopes, &scopedVariable.VariableScope{
//		//		VariableId:  variableId,
//		//		QualifierId: GLOBAL_QUALIFIER,
//		//		Active:      true,
//		//	})
//		//}
//
//		return nil
//	}
func getPriority(qualifier int) int {
	switch qualifier {
	case APP_AND_ENV_QUALIFIER:
		return 1
	case APP_QUALIFIER:
		return 2
	case ENV_QUALIFIER:
		return 3
	case CLUSTER_QUALIFIER:
		return 4
	case GLOBAL_QUALIFIER:
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
	if appId == 0 && envId == 0 && clusterId == 0 {
		return variablePriorityMap
	}
	for _, scope := range varScope {
		isMatch := false
		if appId != 0 && envId != 0 {
			isMatch = impl.filterMatch(scope, appId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID, 0)
			if isMatch && scope.IdentifierKey != 0 {
				for _, envScope := range varScope {
					if isMatch = impl.filterMatch(envScope, envId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID, scope.Id); isMatch {
						break
					}
				}
			}
		} else if appId != 0 {
			isMatch = impl.filterMatch(scope, appId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID, 0)
		} else if envId != 0 {
			isMatch = impl.filterMatch(scope, envId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID, 0)
		} else if clusterId != 0 {
			isMatch = impl.filterMatch(scope, clusterId, bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID, 0)
		}
		if isMatch {
			priority, ok := variablePriorityMap[scope.VariableId]
			currentPriority := getPriority(scope.QualifierId)
			if !ok || priority.Priority > currentPriority {
				variablePriorityMap[scope.VariableId] = &VariablePriorityMapping{
					ScopeId:  scope.Id,
					Priority: currentPriority,
				}
			}
		}
	}
	return variablePriorityMap
}
func (impl *ScopedVariableServiceImpl) GetScopedVariables(appId, envId, clusterId int, varIds []int) (scopedVariableDataObj []*ScopedVariableData, err error) {
	searchableKeyNameIdMap := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
	var varScope []*repository2.VariableScope
	var scopedVariableIds map[int]*VariablePriorityMapping
	scopeIdToVariableScope := make(map[int]*repository2.VariableScope)
	if varIds != nil {
		varScope, err = impl.scopedVariableRepository.GetScopedVariableDataForVarIds(appId, envId, clusterId, searchableKeyNameIdMap, varIds)
		if err != nil {
			return nil, err
		}
	} else {
		varScope, err = impl.scopedVariableRepository.GetScopedVariableData(appId, envId, clusterId, searchableKeyNameIdMap)
		if err != nil {
			return nil, err
		}
	}
	for _, scope := range varScope {
		scopeIdToVariableScope[scope.Id] = scope
	}
	scopedVariableIds = impl.getMatchedScopedVariable(varScope, appId, envId, clusterId)

	var scopeIds, scopedVarIds []int
	for varId, mapping := range scopedVariableIds {
		scopeIds = append(scopeIds, mapping.ScopeId)
		scopedVarIds = append(scopedVarIds, varId)
	}
	varDefs, err := impl.scopedVariableRepository.GetVariablesForVarIds(scopedVarIds)
	if err != nil {

	}
	varDatas, err := impl.scopedVariableRepository.GetDataForScopeIds(scopeIds)
	if err != nil {

	}
	for varId, mapping := range scopedVariableIds {
		scopedVariableData := &ScopedVariableData{}
		for _, varDef := range varDefs {
			if varDef.Id == varId {
				scopedVariableData.VariableName = varDef.Name
				break
			}
		}
		for _, varData := range varDatas {
			if varData.ScopeId == mapping.ScopeId {
				scopedVariableData.VariableValue = varData.Data
			}
		}
		scopedVariableDataObj = append(scopedVariableDataObj, scopedVariableData)
	}
	return scopedVariableDataObj, err

}

//jitna variable id hoga uska particular scope pe data , aur yadi var id nii hoga to jo scope hoga usme sare var ka data  data
//varibale name list output get all ids
