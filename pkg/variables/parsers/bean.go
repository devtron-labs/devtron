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

func GetScopedVarData(varData map[string]string) []*models.ScopedVariableData {
	scopedVarData := make([]*models.ScopedVariableData, len(varData))
	for key, value := range varData {
		scopedVarData = append(scopedVarData, &models.ScopedVariableData{VariableName: key, VariableValue: &models.VariableValue{Value: value}})
	}
	return scopedVarData
}
