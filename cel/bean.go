package cel

type ParamValuesType string

const (
	ParamTypeString         ParamValuesType = "string"
	ParamTypeObject         ParamValuesType = "object"
	ParamTypeInteger        ParamValuesType = "integer"
	ParamTypeList           ParamValuesType = "list"
	ParamTypeBool           ParamValuesType = "bool"
	ParamTypeMapStringToAny ParamValuesType = "mapStringToAny"
)

type ParamName string

const AppName ParamName = "appName"
const ProjectName ParamName = "projectName"
const EnvName ParamName = "envName"
const CdPipelineName ParamName = "cdPipelineName"
const IsProdEnv ParamName = "isProdEnv"
const ClusterName ParamName = "clusterName"
const ChartRefId ParamName = "chartRefId"
const CdPipelineTriggerType ParamName = "cdPipelineTriggerType"
const ContainerRepo ParamName = "containerRepository"
const ContainerImage ParamName = "containerImage"
const ContainerImageTag ParamName = "containerImageTag"
const ImageLabels ParamName = "imageLabels"

type Request struct {
	Expression         string             `json:"expression"`
	ExpressionMetadata ExpressionMetadata `json:"params"`
}

type ExpressionMetadata struct {
	Params []ExpressionParam
}

type ExpressionParam struct {
	ParamName ParamName       `json:"paramName"`
	Value     interface{}     `json:"value"`
	Type      ParamValuesType `json:"type"`
}
