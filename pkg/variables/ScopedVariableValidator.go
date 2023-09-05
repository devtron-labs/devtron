package variables

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/variables/helper"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	"golang.org/x/exp/slices"
	"regexp"
)

func (impl *ScopedVariableServiceImpl) isValidPayload(payload models.Payload) (error, bool) {
	variableNamesList := make([]string, 0)
	for _, variable := range payload.Variables {
		if slices.Contains(variableNamesList, variable.Definition.VarName) {
			return fmt.Errorf("duplicate variable name"), false
		}
		regex := impl.VariableNameConfig.VariableNameRegex

		regexExpression := regexp.MustCompile(regex)
		if !regexExpression.MatchString(variable.Definition.VarName) {
			return fmt.Errorf("variable name %s doesnot match regex %s", variable.Definition.VarName, regex), false
		}
		variableNamesList = append(variableNamesList, variable.Definition.VarName)
		uniqueVariableMap := make(map[string]interface{})
		for _, attributeValue := range variable.AttributeValues {
			validIdentifierTypeList := helper.GetIdentifierTypeFromAttributeType(attributeValue.AttributeType)
			if len(validIdentifierTypeList) != len(attributeValue.AttributeParams) {
				return fmt.Errorf("length of AttributeParams is not valid"), false
			}
			for key, _ := range attributeValue.AttributeParams {
				if !slices.Contains(validIdentifierTypeList, key) {
					return fmt.Errorf("invalid IdentifierType %s for validIdentifierTypeList %s", key, validIdentifierTypeList), false
				}
			}
			identifierString := fmt.Sprintf("%s-%s", variable.Definition.VarName, string(attributeValue.AttributeType))
			for _, key := range validIdentifierTypeList {
				identifierString = fmt.Sprintf("%s-%s", identifierString, attributeValue.AttributeParams[key])
			}
			if _, ok := uniqueVariableMap[identifierString]; ok {
				return fmt.Errorf("duplicate AttributeParams found for AttributeType %v", attributeValue.AttributeType), false
			}
			uniqueVariableMap[identifierString] = attributeValue.VariableValue.Value
		}
	}
	return nil, true
}

func complexTypeValidator(payload models.Payload) bool {
	for _, variable := range payload.Variables {
		variableType := variable.Definition.DataType
		if variableType == models.YAML_TYPE || variableType == models.JSON_TYPE {
			for _, attributeValue := range variable.AttributeValues {
				if attributeValue.VariableValue.Value != "" {
					if variable.Definition.DataType == models.YAML_TYPE {
						if !utils.IsValidYAML(attributeValue.VariableValue.Value.(string)) {
							return false
						}
					} else if variable.Definition.DataType == models.JSON_TYPE {
						if !utils.IsValidJSON(attributeValue.VariableValue.Value.(string)) {
							return false
						}
					}
				} else {
					return false
				}
			}
		}
	}
	return true
}
