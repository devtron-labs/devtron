package expressionEvaluators

type ExpressionMetadata struct {
	Params []ExpressionParam
}

type ExpressionParam struct {
	ParamName ParamName       `json:"paramName"`
	Value     interface{}     `json:"value"`
	Type      ParamValuesType `json:"type"`
}

type ParamValuesType string

const (
	ParamTypeString           ParamValuesType = "string"
	ParamTypeObject           ParamValuesType = "object"
	ParamTypeInteger          ParamValuesType = "integer"
	ParamTypeList             ParamValuesType = "list"
	ParamTypeBool             ParamValuesType = "bool"
	ParamTypeCommitDetails    ParamValuesType = "CommitDetails"
	ParamTypeCommitDetailsMap ParamValuesType = "commitDetailsMap"
)

type ParamName string

const ContainerRepo ParamName = "containerRepository"
const ContainerImage ParamName = "containerImage"
const ContainerImageTag ParamName = "containerImageTag"
const ImageLabels ParamName = "imageLabels"
const GitCommitDetails ParamName = "gitCommitDetails"
const Severity ParamName = "severity"
const PolicyPermission ParamName = "policyPermission"
