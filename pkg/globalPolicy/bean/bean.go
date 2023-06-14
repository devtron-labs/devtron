package bean

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"time"
)

const (
	POLICY_ALL_OBJECTS_PLACEHOLDER = "*"
	TRUE_STRING                    = "true"
	FALSE_STRING                   = "false"
)

var GlobalPluginPolicyDefinitionAttributes = []bean.DevtronResourceAttributeName{
	bean.DEVTRON_RESOURCE_ATTRIBUTE_CI_PIPELINE_STAGE,
}

var GlobalPluginPolicySelectorAttributes = []bean.DevtronResourceAttributeName{
	bean.DEVTRON_RESOURCE_ATTRIBUTE_APP_NAME,
	bean.DEVTRON_RESOURCE_ATTRIBUTE_ENVIRONMENT_NAME,
	bean.DEVTRON_RESOURCE_ATTRIBUTE_CI_PIPELINE_BRANCH_VALUE,
	bean.DEVTRON_RESOURCE_ATTRIBUTE_ENVIRONMENT_IS_PRODUCTION,
}

type GlobalPolicyComponent int

const (
	GLOBAL_POLICY_COMPONENT_DEFINITION  = 0
	GLOBAL_POLICY_COMPONENT_SELECTOR    = 1
	GLOBAL_POLICY_COMPONENT_CONSEQUENCE = 2
)

type GlobalPolicyType string

const (
	GLOBAL_POLICY_TYPE_PLUGIN GlobalPolicyType = "PLUGIN"
)

func (t GlobalPolicyType) ToString() string {
	return string(t)
}

type GlobalPolicyVersion string

const (
	GLOBAL_POLICY_VERSION_V1 GlobalPolicyVersion = "V1"
)

func (v GlobalPolicyVersion) ToString() string {
	return string(v)
}

type ConsequenceAction string

const (
	CONSEQUENCE_ACTION_BLOCK            ConsequenceAction = "BLOCK"
	CONSEQUENCE_ACTION_ALLOW_FOREVER    ConsequenceAction = "ALLOW_FOREVER"
	CONSEQUENCE_ACTION_ALLOW_UNTIL_TIME ConsequenceAction = "ALLOW_UNTIL_TIME"
)

func (a ConsequenceAction) ToString() string {
	return string(a)
}

func (a ConsequenceAction) ToPriorityInt() int { //greater value, more priority
	switch a {
	case CONSEQUENCE_ACTION_ALLOW_FOREVER:
		return 0
	case CONSEQUENCE_ACTION_ALLOW_UNTIL_TIME:
		return 1
	case CONSEQUENCE_ACTION_BLOCK:
		return 2
	default:
		return 0
	}
}

type PluginApplyStage string

const (
	PLUGIN_APPLY_STAGE_PRE_CI         PluginApplyStage = "PRE_CI"
	PLUGIN_APPLY_STAGE_POST_CI        PluginApplyStage = "POST_CI"
	PLUGIN_APPLY_STAGE_PRE_OR_POST_CI PluginApplyStage = "PRE_OR_POST_CI"
)

func (p PluginApplyStage) ToString() string {
	return string(p)
}

type GlobalPolicyDto struct {
	Id          int    `json:"id,omitempty"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	*GlobalPolicyDetailDto
	PolicyOf      GlobalPolicyType    `json:"-"  validate:"oneof=PLUGIN"`
	PolicyVersion GlobalPolicyVersion `json:"-"  validate:"oneof=V1"`
	UserId        int32               `json:"-"`
}

type GlobalPolicyDetailDto struct {
	Definitions  []*DefinitionDto  `json:"definitions" validate:"required,min=1,dive"`
	Selectors    *SelectorDto      `json:"selectors" validate:"required,dive"`
	Consequences []*ConsequenceDto `json:"consequences" validate:"required,min=1,dive"`
}

type ConsequenceDto struct {
	Action        ConsequenceAction `json:"action" validate:"oneof=BLOCK ALLOW_FOREVER ALLOW_UNTIL_TIME"`
	MetadataField time.Time         `json:"metadataField"`
}

type SelectorDto struct {
	ApplicationSelector []*ProjectAppDto        `json:"application" validate:"omitempty,dive"`
	EnvironmentSelector *EnvironmentSelectorDto `json:"environment" validate:"omitempty,dive"`
	BranchList          []*BranchDto            `json:"branch" validate:"omitempty,dive"`
}

type EnvironmentSelectorDto struct {
	AllProductionEnvironments bool             `json:"allProductionEnvironments"`
	ClusterEnvList            []*ClusterEnvDto `json:"clusterEnv" validate:"omitempty,dive"`
}

type DefinitionDto struct {
	AttributeType bean.DevtronResourceAttributeType `json:"attributeType" validate:"oneof=PLUGIN"`
	Data          DefinitionDataDto                 `json:"data" validate:"required,dive"`
}

type DefinitionDataDto struct {
	PluginId     int              `json:"pluginId" validate:"min=1"`
	ApplyToStage PluginApplyStage `json:"applyToStage" validate:"oneof=PRE_CI POST_CI PRE_OR_POST_CI"`
}

type ProjectAppDto struct {
	ProjectName string   `json:"projectName" validate:"required"`
	AppNames    []string `json:"appNames,omitempty"` //if all applications(existing and future) then this array will be empty
}

type ClusterEnvDto struct {
	ClusterName string   `json:"clusterName" validate:"required"` // for all prod environments expecting this to be 0 and below array should contain -1
	EnvNames    []string `json:"envNames,omitempty"`              //if all environments(existing and future) then this array will be empty
}

type BranchDto struct {
	BranchValueType bean.ValueType `json:"branchValueType" validate:"oneof=REGEX FIXED"`
	Value           string         `json:"value" validate:"required,min=1"`
}

type MandatoryPluginDto struct {
	Definitions []*MandatoryPluginDefinitionDto `json:"definitions"`
}

type MandatoryPluginDefinitionDto struct {
	*DefinitionDto
	DefinitionSources []*DefinitionSourceDto `json:"definitionSources"`
}

type DefinitionSourceDto struct {
	ProjectName                  string   `json:"projectName,omitempty"`
	AppName                      string   `json:"appName,omitempty"`
	ClusterName                  string   `json:"clusterName,omitempty"`
	EnvironmentName              string   `json:"environmentName,omitempty"`
	BranchNames                  []string `json:"branchNames,omitempty"`
	IsDueToProductionEnvironment bool     `json:"isDueToProductionEnvironment"`
	IsDueToLinkedPipeline        bool     `json:"isDueToLinkedPipeline"`
	CiPipelineName               string   `json:"ciPipelineName"`
	PolicyName                   string   `json:"policyName"`
}

type PolicyOffendingPipelineWfTreeObject struct {
	PolicyId  int                         `json:"policyId"`
	Workflows []*WorkflowTreeComponentDto `json:"workflows"`
}

type WorkflowTreeComponentDto struct {
	Id             int                                  `json:"id"`
	Name           string                               `json:"name"`
	AppId          int                                  `json:"appId"`
	CiPipelineId   int                                  `json:"ciPipelineId"`
	CiMaterials    []*pipelineConfig.CiPipelineMaterial `json:"ciMaterials"`
	CiPipelineName string                               `json:"ciPipelineName"`
	CdPipelines    []string                             `json:"cdPipelines"`
	GitMaterials   []*Material                          `json:"gitMaterials"`
}

type Material struct {
	GitMaterialId int    `json:"gitMaterialId"`
	MaterialName  string `json:"materialName"`
}

type PluginSourceCiPipelineAppDetailDto struct {
	ProjectName string
	AppName     string
}

type PluginSourceCiPipelineEnvDetailDto struct {
	ClusterName string
	EnvName     string
}

type Severity int

const (
	SEVERITY_MORE_SEVERE Severity = 1
	SEVERITY_SAME_SEVERE          = 2
	SEVERITY_LESS_SEVERE          = 3
)

func (consequence1 *ConsequenceDto) GetSeverity(consequence2 *ConsequenceDto) Severity {
	consequence1ActionInt := consequence1.Action.ToPriorityInt()
	consequence2ActionInt := consequence2.Action.ToPriorityInt()
	if consequence2ActionInt > consequence1ActionInt {
		return SEVERITY_MORE_SEVERE
	} else if consequence2ActionInt == consequence1ActionInt {
		if consequence1.Action == CONSEQUENCE_ACTION_ALLOW_UNTIL_TIME {
			if consequence2.MetadataField.Before(consequence1.MetadataField) {
				return SEVERITY_MORE_SEVERE
			} else if consequence2.MetadataField.After(consequence1.MetadataField) {
				return SEVERITY_LESS_SEVERE
			}
		}
		return SEVERITY_SAME_SEVERE
	} else if consequence2ActionInt < consequence1ActionInt {
		return SEVERITY_LESS_SEVERE
	}
	return SEVERITY_LESS_SEVERE
}
