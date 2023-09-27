package resourceFilter

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
)

const (
	NoResourceFiltersFound               = "no active resource filters found"
	AppAndEnvSelectorRequiredMessage     = "both application and environment selectors are required"
	InvalidExpressions                   = "one or more expressions are invalid"
	AllProjectsValue                     = "0"
	AllProjectsInt                       = 0
	AllExistingAndFutureProdEnvsValue    = "0"
	AllExistingAndFutureProdEnvsInt      = 0
	AllExistingAndFutureNonProdEnvsValue = "-1"
	AllExistingAndFutureNonProdEnvsInt   = -1
)

type IdentifierType int

const (
	ProjectIdentifier     = 0
	AppIdentifier         = 1
	ClusterIdentifier     = 2
	EnvironmentIdentifier = 3
)

type FilterMetaDataBean struct {
	Id           int                 `json:"id"`
	TargetObject *FilterTargetObject `json:"targetObject" validate:"required,min=0,max=1"`
	Description  string              `json:"description" `
	Name         string              `json:"name" validate:"required,max=300"`
}

type FilterRequestResponseBean struct {
	*FilterMetaDataBean
	Conditions        []ResourceCondition `json:"conditions" validate:"required,dive"`
	QualifierSelector QualifierSelector   `json:"qualifierSelector" validate:"required,dive"`
}

type ResourceCondition struct {
	ConditionType ResourceConditionType `json:"conditionType" validate:"min=0,max=1"`
	Expression    string                `json:"expression" validate:"required,min=1""`
	ErrorMsg      string                `json:"errorMsg,omitempty"`
}

func (condition ResourceCondition) IsFailCondition() bool {
	return condition.ConditionType == FAIL
}

type QualifierSelector struct {
	ApplicationSelectors []ApplicationSelector `json:"applicationSelectors" validate:"required,dive"`
	EnvironmentSelectors []EnvironmentSelector `json:"environmentSelectors" validate:"required,dive"`
}

type ApplicationSelector struct {
	ProjectName  string   `json:"projectName" validate:"required,min=1"`
	Applications []string `json:"applications"`
}

type EnvironmentSelector struct {
	ClusterName  string   `json:"clusterName" validate:"min=1"`
	Environments []string `json:"environments"`
}

type ExpressionMetadata struct {
	Params []ExpressionParam
}

type ExpressionParam struct {
	ParamName string          `json:"paramName"`
	Value     interface{}     `json:"value"`
	Type      ParamValuesType `json:"type"`
}

type ParamValuesType string

const (
	ParamTypeString  ParamValuesType = "string"
	ParamTypeObject  ParamValuesType = "object"
	ParamTypeInteger ParamValuesType = "integer"
)

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

func getJsonStringFromResourceCondition(resourceConditions []ResourceCondition) (string, error) {

	jsonBytes, err := json.Marshal(resourceConditions)
	return string(jsonBytes), err
}

func extractResourceConditions(resourceConditionJson string) ([]ResourceCondition, error) {
	var resourceConditions []ResourceCondition
	err := json.Unmarshal([]byte(resourceConditionJson), &resourceConditions)
	return resourceConditions, err
}

func convertToResponseBeans(resourceFilters []*ResourceFilter) ([]*FilterRequestResponseBean, error) {
	var filterResponseBeans []*FilterRequestResponseBean
	for _, resourceFilter := range resourceFilters {
		filterResponseBean, err := convertToFilterBean(resourceFilter)
		if err != nil {
			return filterResponseBeans, err
		}
		filterResponseBeans = append(filterResponseBeans, filterResponseBean)
	}
	return filterResponseBeans, nil
}

func convertToFilterBean(resourceFilter *ResourceFilter) (*FilterRequestResponseBean, error) {
	var err error
	resourceConditions, err := extractResourceConditions(resourceFilter.ConditionExpression)
	if err != nil {
		return nil, err
	}
	filterResponseBean := &FilterRequestResponseBean{
		FilterMetaDataBean: &FilterMetaDataBean{
			Id:           resourceFilter.Id,
			TargetObject: resourceFilter.TargetObject,
			Description:  resourceFilter.Description,
			Name:         resourceFilter.Name,
		},
		Conditions: resourceConditions,
	}
	return filterResponseBean, nil
}

func GetIdentifierKey(identifierType IdentifierType, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) int {
	switch identifierType {
	case AppIdentifier:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID]
	case ClusterIdentifier:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_CLUSTER_ID]
	case EnvironmentIdentifier:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID]
	case ProjectIdentifier:
		return searchableKeyNameIdMap[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_PROJECT_ID]
	default:
		//TODO: revisit
		return -1
	}
}
