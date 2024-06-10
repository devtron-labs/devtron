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

package parsers

import (
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"golang.org/x/exp/slices"
)

const InvalidTemplate = "invalid-template"
const VariableParsingFailed = "variable-parsing-failed"
const UnknownVariableFound = "unknown-variable-found"

const UnknownVariableErrorMsg = "unknown variables found, %s"

type VariableTemplateType int

const (
	StringVariableTemplate VariableTemplateType = 0
	JsonVariableTemplate                        = 1
)

const DefaultVariableTemplate = JsonVariableTemplate

type VariableParserRequest struct {
	TemplateType           VariableTemplateType
	Template               string
	Variables              []*models.ScopedVariableData
	IgnoreUnknownVariables bool
}

func (request VariableParserRequest) GetEmptyResponse() VariableParserResponse {
	return VariableParserResponse{
		Request:          request,
		ResolvedTemplate: request.Template,
	}
}

func CreateParserRequest(template string, templateType VariableTemplateType, variables []*models.ScopedVariableData, ignoreUnknownVariables bool) VariableParserRequest {
	return VariableParserRequest{
		TemplateType:           templateType,
		Template:               template,
		Variables:              variables,
		IgnoreUnknownVariables: ignoreUnknownVariables,
	}
}

type VariableParserResponse struct {
	Request          VariableParserRequest
	ResolvedTemplate string
	Error            error
	DetailedError    string
}

func (request VariableParserRequest) GetValuesMap() map[string]string {
	variablesMap := make(map[string]string)
	variables := request.Variables
	for _, variable := range variables {
		variablesMap[variable.VariableName] = variable.VariableValue.StringValue()
	}
	return variablesMap
}

func (request VariableParserRequest) GetOriginalValuesMap() map[string]interface{} {
	var variableToValue = make(map[string]interface{}, 0)
	for _, variable := range request.Variables {
		variableToValue[variable.VariableName] = variable.VariableValue.Value
	}
	return variableToValue
}

func GetScopedVarData(varData map[string]string, nameToIsSensitive map[string]bool, isSuperAdmin bool) []*models.ScopedVariableData {
	scopedVarData := make([]*models.ScopedVariableData, 0)
	for key, value := range varData {

		finalValue := value
		if !isSuperAdmin && nameToIsSensitive[key] {
			finalValue = models.HiddenValue
		}
		scopedVarData = append(scopedVarData, &models.ScopedVariableData{VariableName: key, VariableValue: &models.VariableValue{Value: models.GetInterfacedValue(finalValue)}})
	}
	return scopedVarData
}

func GetVariableMapForUsedVariables(scopedVariables []*models.ScopedVariableData, usedVars []string) map[string]string {
	variableMap := make(map[string]string)
	for _, variable := range scopedVariables {
		if slices.Contains(usedVars, variable.VariableName) {
			variableMap[variable.VariableName] = variable.VariableValue.StringValue()
		}
	}
	return variableMap
}
