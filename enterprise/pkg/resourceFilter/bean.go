package resourceFilter

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"strings"
	"time"
)

type IdentifierType int

const (
	ProjectIdentifier     IdentifierType = 0
	AppIdentifier         IdentifierType = 1
	ClusterIdentifier     IdentifierType = 2
	EnvironmentIdentifier IdentifierType = 3
)

type FilterMetaDataBean struct {
	Id           int                 `json:"id"`
	TargetObject *FilterTargetObject `json:"targetObject" validate:"required,min=0,max=1"`
	Description  string              `json:"description" `
	Name         string              `json:"name" validate:"required,max=300"`
	Conditions   []ResourceCondition `json:"conditions" validate:"required,dive"`
}

type FilterRequestResponseBean struct {
	*FilterMetaDataBean
	QualifierSelector QualifierSelector `json:"qualifierSelector" validate:"dive"`
}

type ResourceCondition struct {
	ConditionType ResourceConditionType `json:"conditionType" validate:"min=0,max=1"`
	Expression    string                `json:"expression" validate:"required,min=1"`
	ErrorMsg      string                `json:"errorMsg,omitempty"`
}

func (condition ResourceCondition) IsFailCondition() bool {
	return condition.ConditionType == FAIL
}

type ApplicationSelector struct {
	ProjectName  string   `json:"projectName" validate:"required,min=1"`
	Applications []string `json:"applications"`
}

type EnvironmentSelector struct {
	ClusterName  string   `json:"clusterName" validate:"min=1"`
	Environments []string `json:"environments"`
}

type QualifierSelector struct {
	ApplicationSelectors []ApplicationSelector `json:"applicationSelectors" validate:"dive"`
	EnvironmentSelectors []EnvironmentSelector `json:"environmentSelectors" validate:"dive"`
}

func (o QualifierSelector) BuildQualifierMappings(resourceFilterId int, projectNameToIdMap, appNameToIdMap, clusterNameToIdMap, envNameToIdMap map[string]int, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int, userId int32) ([]*resourceQualifiers.QualifierMapping, error) {
	currentTime := time.Now()
	auditLog := sql.AuditLog{
		CreatedOn: currentTime,
		UpdatedOn: currentTime,
		CreatedBy: userId,
		UpdatedBy: userId,
	}
	appQualifierMappings := o.buildApplicationQualifierMappings(resourceFilterId, projectNameToIdMap, appNameToIdMap, searchableKeyNameIdMap, auditLog)
	envQualifierMappings, err := o.buildEnvironmentQualifierMappings(resourceFilterId, clusterNameToIdMap, envNameToIdMap, searchableKeyNameIdMap, auditLog)
	if err != nil {
		return nil, err
	}
	qualifierMappings := append(appQualifierMappings, envQualifierMappings...)
	return qualifierMappings, nil
}

func (o QualifierSelector) buildApplicationQualifierMappings(resourceFilterId int, projectNameToIdMap, appNameToIdMap map[string]int, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int, auditLog sql.AuditLog) []*resourceQualifiers.QualifierMapping {
	qualifierMappings := make([]*resourceQualifiers.QualifierMapping, 0)
	applicationSelectors := o.ApplicationSelectors
	//case-1) all existing and future applications -> will get empty ApplicationSelector , db entry (proj,0,"0")
	if len(applicationSelectors) == 1 && applicationSelectors[0].ProjectName == resourceQualifiers.AllProjectsValue {
		allExistingAndFutureAppsQualifierMapping := &resourceQualifiers.QualifierMapping{
			ResourceId:            resourceFilterId,
			ResourceType:          resourceQualifiers.Filter,
			QualifierId:           int(resourceQualifiers.APP_AND_ENV_QUALIFIER),
			IdentifierKey:         GetIdentifierKey(ProjectIdentifier, searchableKeyNameIdMap),
			Active:                true,
			IdentifierValueInt:    resourceQualifiers.AllProjectsInt,
			IdentifierValueString: resourceQualifiers.AllProjectsValue,
			AuditLog:              auditLog,
		}
		qualifierMappings = append(qualifierMappings, allExistingAndFutureAppsQualifierMapping)
	} else {

		for _, appSelector := range applicationSelectors {
			//case-2) all existing and future apps in a project ->  will get projectName and empty applications array
			if len(appSelector.Applications) == 0 {
				allExistingAppsQualifierMapping := &resourceQualifiers.QualifierMapping{
					ResourceId:            resourceFilterId,
					QualifierId:           int(resourceQualifiers.APP_AND_ENV_QUALIFIER),
					ResourceType:          resourceQualifiers.Filter,
					IdentifierKey:         GetIdentifierKey(ProjectIdentifier, searchableKeyNameIdMap),
					Active:                true,
					IdentifierValueInt:    projectNameToIdMap[appSelector.ProjectName],
					IdentifierValueString: appSelector.ProjectName,
					AuditLog:              auditLog,
				}
				qualifierMappings = append(qualifierMappings, allExistingAppsQualifierMapping)
			}
			//case-3) all existing applications -> will get all apps in payload
			//case-4) particular apps -> will get ApplicationSelectors array
			//case-5) all existing apps in a project -> will get projectName and all applications array
			for _, appName := range appSelector.Applications {
				appQualifierMapping := &resourceQualifiers.QualifierMapping{
					ResourceId:            resourceFilterId,
					QualifierId:           int(resourceQualifiers.APP_AND_ENV_QUALIFIER),
					ResourceType:          resourceQualifiers.Filter,
					IdentifierKey:         GetIdentifierKey(AppIdentifier, searchableKeyNameIdMap),
					Active:                true,
					IdentifierValueInt:    appNameToIdMap[appName],
					IdentifierValueString: appName,
					AuditLog:              auditLog,
				}
				qualifierMappings = append(qualifierMappings, appQualifierMapping)
			}
		}
	}
	return qualifierMappings
}

func (o QualifierSelector) buildEnvironmentQualifierMappings(resourceFilterId int, clusterNameToIdMap, envNameToIdMap map[string]int, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int, auditLog sql.AuditLog) ([]*resourceQualifiers.QualifierMapping, error) {
	qualifierMappings := make([]*resourceQualifiers.QualifierMapping, 0)
	allClusterEnvSelectors, otherEnvSelectors, err := o.validateAndSplitEnvSelectors()
	if err != nil {
		return qualifierMappings, err
	}

	//1) all existing and future prod envs -> get single EnvironmentSelector with clusterName as "0"(prod) (cluster,0,"0")
	//2) all existing and future non-prod envs -> get single EnvironmentSelector with clusterName as "-1"(non-prod) (cluster,-1,"-1")
	for _, envSelector := range allClusterEnvSelectors {
		allExistingAndFutureEnvQualifierMapping := &resourceQualifiers.QualifierMapping{
			ResourceId:    resourceFilterId,
			QualifierId:   int(resourceQualifiers.APP_AND_ENV_QUALIFIER),
			ResourceType:  resourceQualifiers.Filter,
			IdentifierKey: GetIdentifierKey(ClusterIdentifier, searchableKeyNameIdMap),
			Active:        true,
			AuditLog:      auditLog,
		}
		if envSelector.ClusterName == resourceQualifiers.AllExistingAndFutureProdEnvsValue {
			allExistingAndFutureEnvQualifierMapping.IdentifierValueInt = resourceQualifiers.AllExistingAndFutureProdEnvsInt
			allExistingAndFutureEnvQualifierMapping.IdentifierValueString = resourceQualifiers.AllExistingAndFutureProdEnvsValue
		} else {
			allExistingAndFutureEnvQualifierMapping.IdentifierValueInt = resourceQualifiers.AllExistingAndFutureNonProdEnvsInt
			allExistingAndFutureEnvQualifierMapping.IdentifierValueString = resourceQualifiers.AllExistingAndFutureNonProdEnvsValue
		}
		qualifierMappings = append(qualifierMappings, allExistingAndFutureEnvQualifierMapping)
	}

	for _, envSelector := range otherEnvSelectors {
		//3) all existing and future envs of a cluster ->  get clusterName and empty environments list (cluster,clusterId,clusterName)
		if len(envSelector.Environments) == 0 {
			allCurrentAndFutureEnvsInClusterQualifierMapping := &resourceQualifiers.QualifierMapping{
				ResourceId:            resourceFilterId,
				QualifierId:           int(resourceQualifiers.APP_AND_ENV_QUALIFIER),
				ResourceType:          resourceQualifiers.Filter,
				IdentifierKey:         GetIdentifierKey(ClusterIdentifier, searchableKeyNameIdMap),
				IdentifierValueInt:    clusterNameToIdMap[envSelector.ClusterName],
				IdentifierValueString: envSelector.ClusterName,
				Active:                true,
				AuditLog:              auditLog,
			}
			qualifierMappings = append(qualifierMappings, allCurrentAndFutureEnvsInClusterQualifierMapping)
		}
		//4) all existing envs of a cluster -> get clusterName and all the envs list
		//5) particular envs , will get EnvironmentSelector array
		for _, env := range envSelector.Environments {
			envQualifierMapping := &resourceQualifiers.QualifierMapping{
				ResourceId:            resourceFilterId,
				QualifierId:           int(resourceQualifiers.APP_AND_ENV_QUALIFIER),
				ResourceType:          resourceQualifiers.Filter,
				IdentifierKey:         GetIdentifierKey(EnvironmentIdentifier, searchableKeyNameIdMap),
				IdentifierValueInt:    envNameToIdMap[env],
				IdentifierValueString: env,
				Active:                true,
				AuditLog:              auditLog,
			}
			qualifierMappings = append(qualifierMappings, envQualifierMapping)
		}
	}
	return qualifierMappings, nil
}

func (o QualifierSelector) validateAndSplitEnvSelectors() ([]EnvironmentSelector, []EnvironmentSelector, error) {
	//type1: allExistingFutureProdEnvs
	//type2: allExistingFutureNonProdEnvs
	//type3: allExistingFutureEnvsOfACluster
	//type4: remaining types
	envSelectors := o.EnvironmentSelectors
	allExistingFutureProdEnvSelectors := make([]EnvironmentSelector, 0)
	allExistingFutureNonProdEnvSelectors := make([]EnvironmentSelector, 0)
	allExistingFutureEnvsOfACluster := make([]EnvironmentSelector, 0)
	otherEnvSelectors := make([]EnvironmentSelector, 0)

	//ValidCases:
	//   case1 : type1 + type4(nonProdEnvs),
	//   case2 : type2 + type4(prodEnvs),
	//   case3 : type1 + type2
	//   case4 : (type1 or type2) + type3

	for _, envSelector := range envSelectors {
		//order of these cases are **IMPORTANT**
		if envSelector.ClusterName == resourceQualifiers.AllExistingAndFutureProdEnvsValue {
			allExistingFutureProdEnvSelectors = append(allExistingFutureProdEnvSelectors, envSelector)
		} else if envSelector.ClusterName == resourceQualifiers.AllExistingAndFutureNonProdEnvsValue {
			allExistingFutureNonProdEnvSelectors = append(allExistingFutureNonProdEnvSelectors, envSelector)
		} else if len(envSelector.Environments) == 0 {
			allExistingFutureEnvsOfACluster = append(allExistingFutureEnvsOfACluster, envSelector)
		} else {
			otherEnvSelectors = append(otherEnvSelectors, envSelector)
		}
	}

	//InValidCases:
	//   case1: multiple type1 or multiple type2
	if len(allExistingFutureProdEnvSelectors) > 1 || len(allExistingFutureNonProdEnvSelectors) > 1 {
		return nil, nil, errors.New("multiple selectors of type allExistingFutureProdEnvSelector or allExistingFutureNonProdEnvSelector found, invalid selectors request")
	}

	//   case2: type1 + type2 + (type4 or type3)
	if len(allExistingFutureProdEnvSelectors) != 0 && len(allExistingFutureNonProdEnvSelectors) != 0 && (len(otherEnvSelectors) != 0 || len(allExistingFutureEnvsOfACluster) != 0) {
		return nil, nil, errors.New("some other selectors found along with allExistingFutureProdEnvSelector and allExistingFutureNonProdEnvSelector found, invalid selectors request")
	}

	//TODO: handle(requires db call and then validate)
	//   case3: type1 + type4(prodEnvs)
	//   case4: type2 + type4(nonProdEnvs)

	allClusterEnvSelectors := append(allExistingFutureProdEnvSelectors, allExistingFutureNonProdEnvSelectors...)
	otherEnvSelectors = append(otherEnvSelectors, allExistingFutureEnvsOfACluster...)
	return allClusterEnvSelectors, otherEnvSelectors, nil
}

type ParamValuesType string

const (
	ParamTypeString  ParamValuesType = "string"
	ParamTypeObject  ParamValuesType = "object"
	ParamTypeInteger ParamValuesType = "integer"
	ParamTypeList    ParamValuesType = "list"
)

type ExpressionParam struct {
	ParamName string          `json:"paramName"`
	Value     interface{}     `json:"value"`
	Type      ParamValuesType `json:"type"`
}

func getParamsFromArtifact(artifact string, imageLabels []string) []ExpressionParam {

	lastColonIndex := strings.LastIndex(artifact, ":")

	containerRepository := artifact[:lastColonIndex]
	containerImageTag := artifact[lastColonIndex+1:]
	containerImage := artifact
	params := []ExpressionParam{
		{
			ParamName: "containerRepository",
			Value:     containerRepository,
			Type:      ParamTypeString,
		},
		{
			ParamName: "containerImage",
			Value:     containerImage,
			Type:      ParamTypeString,
		},
		{
			ParamName: "containerImageTag",
			Value:     containerImageTag,
			Type:      ParamTypeString,
		},
		{
			ParamName: "imageLabels",
			Value:     imageLabels,
			Type:      ParamTypeList,
		},
	}

	return params
}

type ExpressionMetadata struct {
	Params []ExpressionParam
}

type expressionResponse struct {
	allowConditionAvail bool
	allowResponse       bool
	blockConditionAvail bool
	blockResponse       bool
}

func (response expressionResponse) getFinalResponse() bool {
	if response.blockConditionAvail && response.blockResponse {
		return false
	}

	if response.allowConditionAvail && !response.allowResponse {
		return false
	}
	return true
}
