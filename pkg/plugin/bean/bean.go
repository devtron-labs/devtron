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

package bean

import (
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

const (
	CREATEPLUGIN      = 0
	UPDATEPLUGIN      = 1
	DELETEPLUGIN      = 2
	CI_TYPE_PLUGIN    = "CI"
	CD_TYPE_PLUGIN    = "CD"
	CI_CD_TYPE_PLUGIN = "CI_CD"
)

type PluginDetailDto struct {
	Metadata        *PluginMetadataDto   `json:"metadata"`
	InputVariables  []*PluginVariableDto `json:"inputVariables"`
	OutputVariables []*PluginVariableDto `json:"outputVariables"`
}

type PluginListComponentDto struct { //created new struct for backward compatibility (needed to add input and output Vars along with metadata fields)
	*PluginMetadataDto
	InputVariables  []*PluginVariableDto `json:"inputVariables"`
	OutputVariables []*PluginVariableDto `json:"outputVariables"`
}

type PluginMetadataDto struct {
	Id          int               `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type,omitempty" validate:"oneof=SHARED PRESET"` // SHARED, PRESET etc
	Icon        string            `json:"icon,omitempty"`
	Tags        []string          `json:"tags"`
	Action      int               `json:"action,omitempty"`
	PluginStage string            `json:"pluginStage,omitempty"`
	PluginSteps []*PluginStepsDto `json:"pluginSteps,omitempty"`
}

type PluginsDto struct {
	ParentPlugins []*PluginParentMetadataDto `json:"parentPlugins"`
	TotalCount    int                        `json:"totalCount"`
}

func NewPluginsDto() *PluginsDto {
	return &PluginsDto{}
}

func (r *PluginsDto) WithParentPlugins(parentPlugins []*PluginParentMetadataDto) *PluginsDto {
	r.ParentPlugins = parentPlugins
	return r
}

func (r *PluginsDto) WithTotalCount(count int) *PluginsDto {
	r.TotalCount = count
	return r
}

type PluginParentMetadataDto struct {
	Id               int             `json:"id"`
	Name             string          `json:"name"`
	PluginIdentifier string          `json:"pluginIdentifier"`
	Description      string          `json:"description"`
	Type             string          `json:"type,omitempty" validate:"oneof=SHARED PRESET"`
	Icon             string          `json:"icon,omitempty"`
	Versions         *PluginVersions `json:"pluginVersions"`
}

func NewPluginParentMetadataDto() *PluginParentMetadataDto {
	return &PluginParentMetadataDto{}
}

func (r *PluginParentMetadataDto) WithNameAndId(name string, id int) *PluginParentMetadataDto {
	r.Id = id
	r.Name = name
	return r
}

func (r *PluginParentMetadataDto) WithPluginIdentifier(identifier string) *PluginParentMetadataDto {
	r.PluginIdentifier = identifier
	return r
}

func (r *PluginParentMetadataDto) WithDescription(desc string) *PluginParentMetadataDto {
	r.Description = desc
	return r
}

func (r *PluginParentMetadataDto) WithIcon(icon string) *PluginParentMetadataDto {
	r.Icon = icon
	return r
}

func (r *PluginParentMetadataDto) WithType(pluginType string) *PluginParentMetadataDto {
	r.Type = pluginType
	return r
}

func (r *PluginParentMetadataDto) WithVersions(versions *PluginVersions) *PluginParentMetadataDto {
	r.Versions = versions
	return r
}

type PluginVersions struct {
	DetailedPluginVersionData []*PluginsVersionDetail `json:"detailedPluginVersionData"` // contains detailed data with all input and output variables
	MinimalPluginVersionData  []*PluginsVersionDetail `json:"minimalPluginVersionData"`  // contains only few metadata
}

type PluginMinDto struct {
	PluginName     string                  `json:"pluginName"`
	PluginVersions []*PluginVersionsMinDto `json:"pluginVersions"`
	Icon           string                  `json:"icon"`
}

type PluginVersionsMinDto struct {
	Id      int    `json:"id"`
	Version string `json:"version"`
}

func NewPluginVersions() *PluginVersions {
	return &PluginVersions{}
}

func (r *PluginVersions) WithDetailedPluginVersionData(detailedPluginVersionData []*PluginsVersionDetail) *PluginVersions {
	r.DetailedPluginVersionData = detailedPluginVersionData
	return r
}

func (r *PluginVersions) WithMinimalPluginVersionData(minimalPluginVersionData []*PluginsVersionDetail) *PluginVersions {
	r.MinimalPluginVersionData = minimalPluginVersionData
	return r
}

type PluginsVersionDetail struct {
	*PluginMetadataDto
	InputVariables  []*PluginVariableDto `json:"inputVariables"`
	OutputVariables []*PluginVariableDto `json:"outputVariables"`
	DocLink         string               `json:"docLink"`
	Version         string               `json:"pluginVersion"`
	IsLatest        bool                 `json:"isLatest"`
	UpdatedBy       string               `json:"updatedBy"`
	CreatedOn       time.Time            `json:"-"`
}

func NewPluginsVersionDetail() *PluginsVersionDetail {
	return &PluginsVersionDetail{PluginMetadataDto: &PluginMetadataDto{}}
}

// SetMinimalPluginsVersionDetail sets and return PluginsVersionDetail obj, returns lightweight obj e.g. excluding input and output variables
func (r *PluginsVersionDetail) SetMinimalPluginsVersionDetail(pluginVersionMetadata *repository.PluginMetadata) *PluginsVersionDetail {
	r.Id = pluginVersionMetadata.Id
	r.Name = pluginVersionMetadata.Name
	r.Description = pluginVersionMetadata.Description
	r.Version = pluginVersionMetadata.PluginVersion
	r.IsLatest = pluginVersionMetadata.IsLatest
	return r
}

func (r *PluginsVersionDetail) WithLastUpdatedEmail(email string) *PluginsVersionDetail {
	r.UpdatedBy = email
	return r
}

func (r *PluginsVersionDetail) WithCreatedOn(createdOn time.Time) *PluginsVersionDetail {
	r.CreatedOn = createdOn
	return r
}

func (r *PluginsVersionDetail) WithInputVariables(inputVariables []*PluginVariableDto) *PluginsVersionDetail {
	r.InputVariables = inputVariables
	return r
}

func (r *PluginsVersionDetail) WithOutputVariables(outputVariables []*PluginVariableDto) *PluginsVersionDetail {
	r.OutputVariables = outputVariables
	return r
}

func (r *PluginsVersionDetail) WithTags(tags []string) *PluginsVersionDetail {
	r.Tags = tags
	return r
}

type PluginsListFilter struct {
	Offset                 int
	Limit                  int
	SearchKey              string
	Tags                   []string
	FetchAllVersionDetails bool
}

func NewPluginsListFilter() *PluginsListFilter {
	return &PluginsListFilter{}
}

func (r *PluginsListFilter) WithOffset(offset int) *PluginsListFilter {
	r.Offset = offset
	return r
}

func (r *PluginsListFilter) WithLimit(limit int) *PluginsListFilter {
	r.Limit = limit
	return r
}

func (r *PluginsListFilter) WithSearchKey(searchKey string) *PluginsListFilter {
	r.SearchKey = searchKey
	return r
}

func (r *PluginsListFilter) WithTags(tags []string) *PluginsListFilter {
	r.Tags = tags
	return r
}

type PluginTagsDto struct {
	TagNames []string `json:"tagNames"`
}

func NewPluginTagsDto() *PluginTagsDto {
	return &PluginTagsDto{}
}

func (r *PluginTagsDto) WithTagNames(tags []string) *PluginTagsDto {
	r.TagNames = tags
	return r
}

func (r *PluginMetadataDto) GetPluginMetadataSqlObj(userId int32) *repository.PluginMetadata {
	return &repository.PluginMetadata{
		Name:        r.Name,
		Description: r.Description,
		Type:        repository.PluginType(r.Type),
		Icon:        r.Icon,
		AuditLog:    sql.NewDefaultAuditLog(userId),
	}
}

type PluginStepsDto struct {
	Id                   int                       `json:"id,pk"`
	Name                 string                    `json:"name"`
	Description          string                    `json:"description"`
	Index                int                       `json:"index"`
	StepType             repository.PluginStepType `json:"stepType"`
	RefPluginId          int                       `json:"refPluginId"` //id of plugin used as reference
	OutputDirectoryPath  []string                  `json:"outputDirectoryPath"`
	DependentOnStep      string                    `json:"dependentOnStep"`
	PluginStepVariable   []*PluginVariableDto      `json:"pluginStepVariable,omitempty"`
	PluginPipelineScript *PluginPipelineScript     `json:"pluginPipelineScript,omitempty"`
}

type PluginVariableDto struct {
	Id                        int                                     `json:"id,omitempty"`
	Name                      string                                  `json:"name"`
	Format                    repository.PluginStepVariableFormatType `json:"format"`
	Description               string                                  `json:"description"`
	IsExposed                 bool                                    `json:"isExposed"`
	AllowEmptyValue           bool                                    `json:"allowEmptyValue"`
	DefaultValue              string                                  `json:"defaultValue"`
	Value                     string                                  `json:"value,omitempty"`
	VariableType              repository.PluginStepVariableType       `json:"variableType"`
	ValueType                 repository.PluginStepVariableValueType  `json:"valueType,omitempty"`
	PreviousStepIndex         int                                     `json:"previousStepIndex,omitempty"`
	VariableStepIndex         int                                     `json:"variableStepIndex"`
	VariableStepIndexInPlugin int                                     `json:"variableStepIndexInPlugin"`
	ReferenceVariableName     string                                  `json:"referenceVariableName,omitempty"`
	PluginStepCondition       []*PluginStepCondition                  `json:"pluginStepCondition,omitempty"`
}

type PluginPipelineScript struct {
	Id                       int                                  `json:"id"`
	Script                   string                               `json:"script"`
	StoreScriptAt            string                               `json:"storeScriptAt"`
	Type                     repository.ScriptType                `json:"type"`
	DockerfileExists         bool                                 `json:"dockerfileExists"`
	MountPath                string                               `json:"mountPath"`
	MountCodeToContainer     bool                                 `json:"mountCodeToContainer"`
	MountCodeToContainerPath string                               `json:"mountCodeToContainerPath"`
	MountDirectoryFromHost   bool                                 `json:"mountDirectoryFromHost"`
	ContainerImagePath       string                               `json:"containerImagePath"`
	ImagePullSecretType      repository.ScriptImagePullSecretType `json:"imagePullSecretType"`
	ImagePullSecret          string                               `json:"imagePullSecret"`
	Deleted                  bool                                 `json:"deleted"`
	PathArgPortMapping       []*ScriptPathArgPortMapping          `json:"pathArgPortMapping"`
}

type PluginStepCondition struct {
	Id                  int                                `json:"id"`
	PluginStepId        int                                `json:"pluginStepId"`
	ConditionVariableId int                                `json:"conditionVariableId"` //id of variable on which condition is written
	ConditionType       repository.PluginStepConditionType `json:"conditionType"`
	ConditionalOperator string                             `json:"conditionalOperator"`
	ConditionalValue    string                             `json:"conditionalValue"`
	Deleted             bool                               `json:"deleted"`
}

type ScriptPathArgPortMapping struct {
	Id                  int                          `json:"id"`
	TypeOfMapping       repository.ScriptMappingType `json:"typeOfMapping"`
	FilePathOnDisk      string                       `json:"filePathOnDisk"`
	FilePathOnContainer string                       `json:"filePathOnContainer"`
	Command             string                       `json:"command"`
	Args                []string                     `json:"args"`
	PortOnLocal         int                          `json:"portOnLocal"`
	PortOnContainer     int                          `json:"portOnContainer"`
	ScriptId            int                          `json:"scriptId"`
}

type RegistryCredentials struct {
	RegistryType       string `json:"registryType" validate:"required"`
	RegistryURL        string `json:"registryURL"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	AWSAccessKeyId     string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
	AWSRegion          string `json:"awsRegion,omitempty"`
}

const (
	NoPluginOrParentIdProvidedErr      = "no value for pluginVersionIds and parentPluginIds provided in query param"
	NoPluginFoundForThisSearchQueryErr = "unable to find desired plugin for the query filter"
)

const (
	SpecialCharsRegex = ` !"#$%&'()*+,./:;<=>?@[\]^_{|}~` + "`"
)
