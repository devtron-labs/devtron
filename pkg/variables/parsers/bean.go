package parsers

import "github.com/devtron-labs/devtron/pkg/variables/models"

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
			finalValue = "*******"
		}
		scopedVarData = append(scopedVarData, &models.ScopedVariableData{VariableName: key, VariableValue: &models.VariableValue{Value: models.GetInterfacedValue(finalValue)}})
	}
	return scopedVarData
}
