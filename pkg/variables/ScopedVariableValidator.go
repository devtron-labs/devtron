package variables

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/variables/helper"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	"golang.org/x/exp/slices"
	"regexp"
	"strings"
)

func (impl *ScopedVariableServiceImpl) isValidPayload(payload models.Payload) (error, bool) {
	variableNamesList := make([]string, 0)
	for _, variable := range payload.Variables {
		if slices.Contains(variableNamesList, variable.Definition.VarName) {
			return models.ValidationError{Err: fmt.Errorf("duplicate variable name %s", variable.Definition.VarName)}, false
		}

		if strings.HasPrefix(variable.Definition.VarName, impl.VariableNameConfig.SystemVariablePrefix) {
			return models.ValidationError{Err: fmt.Errorf("%s is not allowed. Prefix %s is reserved for system variables)", variable.Definition.VarName, impl.VariableNameConfig.SystemVariablePrefix)}, false
		}

		regex := impl.VariableNameConfig.VariableNameRegex

		regexExpression := regexp.MustCompile(regex)
		if !regexExpression.MatchString(variable.Definition.VarName) {
			return models.ValidationError{Err: fmt.Errorf("%s does not match the required format (Alphanumeric, 64 characters max, no hyphen/underscore at start/end)", variable.Definition.VarName)}, false
		}
		variableNamesList = append(variableNamesList, variable.Definition.VarName)
		uniqueVariableMap := make(map[string]interface{})
		for _, attributeValue := range variable.AttributeValues {
			validIdentifierTypeList := helper.GetIdentifierTypeFromAttributeType(attributeValue.AttributeType)
			if len(validIdentifierTypeList) != len(attributeValue.AttributeParams) {
				return models.ValidationError{Err: fmt.Errorf("attribute selectors are not valid for given category %s", attributeValue.AttributeType)}, false
			}
			for key, _ := range attributeValue.AttributeParams {
				if !slices.Contains(validIdentifierTypeList, key) {
					return models.ValidationError{Err: fmt.Errorf("invalid attribute selector key %s", key)}, false
				}
			}
			identifierString := fmt.Sprintf("%s-%s", variable.Definition.VarName, string(attributeValue.AttributeType))
			for _, key := range validIdentifierTypeList {
				identifierString = fmt.Sprintf("%s-%s", identifierString, attributeValue.AttributeParams[key])
			}
			if _, ok := uniqueVariableMap[identifierString]; ok {
				return models.ValidationError{Err: fmt.Errorf("duplicate Selectors  found for category %v", attributeValue.AttributeType)}, false
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
