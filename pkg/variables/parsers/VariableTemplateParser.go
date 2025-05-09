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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	_ "github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	ctyJson "github.com/zclconf/go-cty/cty/json"
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"strings"
)

type VariableTemplateParser interface {
	ExtractVariables(template string, templateType VariableTemplateType) ([]string, error)
	//ParseTemplate(template string, values map[string]string) string
	ParseTemplate(parserRequest VariableParserRequest) VariableParserResponse
}

type VariableTemplateParserImpl struct {
	logger                       *zap.SugaredLogger
	variableTemplateParserConfig *VariableTemplateParserConfig
}

func NewVariableTemplateParserImpl(logger *zap.SugaredLogger) (*VariableTemplateParserImpl, error) {
	impl := &VariableTemplateParserImpl{logger: logger}
	cfg, err := getVariableTemplateParserConfig()
	if err != nil {
		return nil, err
	}
	impl.variableTemplateParserConfig = cfg
	return impl, nil
}

type VariableTemplateParserConfig struct {
	ScopedVariableEnabled          bool   `env:"SCOPED_VARIABLE_ENABLED" envDefault:"false" description:"To enable scoped variable option"`
	ScopedVariableHandlePrimitives bool   `env:"SCOPED_VARIABLE_HANDLE_PRIMITIVES" envDefault:"false" description:"This describe should we handle primitives or not in scoped variable template parsing."`
	VariableExpressionRegex        string `env:"VARIABLE_EXPRESSION_REGEX" envDefault:"@{{([^}]+)}}" description:"Scoped variable expression regex"`
}

func (cfg VariableTemplateParserConfig) isScopedVariablesDisabled() bool {
	return !cfg.ScopedVariableEnabled
}

func getVariableTemplateParserConfig() (*VariableTemplateParserConfig, error) {
	cfg := &VariableTemplateParserConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

func getRegexSubMatches(regex string, input string) [][]string {
	re := regexp.MustCompile(regex)
	matches := re.FindAllStringSubmatch(input, -1)
	return matches
}

const quote = "\""
const escapedQuote = `\\"`

func (impl *VariableTemplateParserImpl) preProcessPlaceholder(template string, variableValueMap map[string]interface{}) string {

	variableSubRegexWithQuotes := quote + impl.variableTemplateParserConfig.VariableExpressionRegex + quote
	variableSubRegexWithEscapedQuotes := escapedQuote + impl.variableTemplateParserConfig.VariableExpressionRegex + escapedQuote

	matches := getRegexSubMatches(variableSubRegexWithQuotes, template)
	matches = append(matches, getRegexSubMatches(variableSubRegexWithEscapedQuotes, template)...)

	// Replace the surrounding quotes for variables whose value is known
	// and type is primitive
	for _, match := range matches {
		if len(match) == 2 {
			originalMatch := match[0]
			innerContent := match[1]
			if val, ok := variableValueMap[innerContent]; ok && utils.IsPrimitiveType(val) {
				replacement := fmt.Sprintf("@{{%s}}", innerContent)
				template = strings.Replace(template, originalMatch, replacement, 1)
			}
		}
	}
	return template
}

func (impl *VariableTemplateParserImpl) ParseTemplate(parserRequest VariableParserRequest) VariableParserResponse {

	if impl.variableTemplateParserConfig.isScopedVariablesDisabled() {
		return parserRequest.GetEmptyResponse()
	}
	request := parserRequest
	if impl.handlePrimitivesForJson(parserRequest) {
		variableToValue := parserRequest.GetOriginalValuesMap()
		template := impl.preProcessPlaceholder(parserRequest.Template, variableToValue)

		//overriding request to handle primitives in json request
		request.TemplateType = StringVariableTemplate
		request.Template = template
	}
	return impl.parseTemplate(request)
}

func (impl *VariableTemplateParserImpl) handlePrimitivesForJson(parserRequest VariableParserRequest) bool {
	return impl.variableTemplateParserConfig.ScopedVariableHandlePrimitives && parserRequest.TemplateType == JsonVariableTemplate
}

func (impl *VariableTemplateParserImpl) ExtractVariables(template string, templateType VariableTemplateType) ([]string, error) {
	var variables []string

	if template == "" {
		return variables, nil
	}

	if !impl.variableTemplateParserConfig.ScopedVariableEnabled {
		return variables, nil
	}

	// preprocess existing template to comment
	template, err := impl.convertToHclCompatible(templateType, template)
	if err != nil {
		return variables, err
	}
	hclExpression, diagnostics := hclsyntax.ParseExpression([]byte(template), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
	if diagnostics.HasErrors() {
		impl.logger.Errorw("error occurred while extracting variables from template", "template", template, "error", diagnostics.Error())
		return variables, errors.New(InvalidTemplate)
	} else {
		hclVariables := hclExpression.Variables()
		variables = impl.extractVarNames(hclVariables)
	}
	impl.logger.Info("extracted variables from template", variables)
	return variables, nil
}

func (impl *VariableTemplateParserImpl) extractVarNames(hclVariables []hcl.Traversal) []string {
	var variables []string
	for _, hclVariable := range hclVariables {
		variables = append(variables, hclVariable.RootName())
	}
	return variables
}

func (impl *VariableTemplateParserImpl) parseTemplate(parserRequest VariableParserRequest) VariableParserResponse {
	template := parserRequest.Template
	response := VariableParserResponse{Request: parserRequest, ResolvedTemplate: template}
	values := parserRequest.GetValuesMap()
	templateType := parserRequest.TemplateType
	template, err := impl.convertToHclCompatible(templateType, template)
	if err != nil {
		response.Error = err
		return response
	}
	impl.logger.Debug("variable hcl template valueMap", values)
	hclExpression, diagnostics := hclsyntax.ParseExpression([]byte(template), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
	containsError := impl.checkAndUpdateDiagnosticError(diagnostics, template, &response, InvalidTemplate)
	if containsError {
		return response
	}
	updatedHclExpression, template, containsError := impl.checkForDefaultedVariables(parserRequest, hclExpression.Variables(), template, &response)
	if containsError {
		return response
	}
	if updatedHclExpression != nil {
		hclExpression = updatedHclExpression
	}

	hclVarValues := impl.getHclVarValues(values)
	opValue, diagnostics := hclExpression.Value(&hcl.EvalContext{
		Variables: hclVarValues,
		Functions: impl.getDefaultMappedFunc(),
	})
	containsError = impl.checkAndUpdateDiagnosticError(diagnostics, template, &response, VariableParsingFailed)
	if containsError {
		return response
	}
	output, err := impl.extractResolvedTemplate(templateType, opValue)
	if err != nil {
		output = template
	}
	response.ResolvedTemplate = output
	return response
}

func (impl *VariableTemplateParserImpl) checkForDefaultedVariables(parserRequest VariableParserRequest, variables []hcl.Traversal, template string, response *VariableParserResponse) (hclsyntax.Expression, string, bool) {
	var hclExpression hclsyntax.Expression
	var diagnostics hcl.Diagnostics
	valuesMap := parserRequest.GetValuesMap()
	defaultedVars := impl.getDefaultedVariables(variables, valuesMap)
	ignoreDefaultedVariables := parserRequest.IgnoreUnknownVariables
	if len(defaultedVars) > 0 {
		if ignoreDefaultedVariables {
			template = impl.ignoreDefaultedVars(template, defaultedVars)
			hclExpression, diagnostics = hclsyntax.ParseExpression([]byte(template), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
			hasErrors := impl.checkAndUpdateDiagnosticError(diagnostics, template, response, InvalidTemplate)
			if hasErrors {
				return nil, template, true
			}
		} else {
			impl.logger.Errorw("error occurred while parsing template, unknown variables found", "defaultedVars", defaultedVars)
			response.Error = errors.New(UnknownVariableFound)
			defaultsVarNames := impl.extractVarNames(defaultedVars)
			response.DetailedError = fmt.Sprintf(UnknownVariableErrorMsg, strings.Join(defaultsVarNames, ","))
			return nil, template, true
		}
	}
	return hclExpression, template, false
}

func (impl *VariableTemplateParserImpl) extractResolvedTemplate(templateType VariableTemplateType, opValue cty.Value) (string, error) {
	var output string
	if templateType == StringVariableTemplate {
		opValueMap := opValue.AsValueMap()
		rootValue := opValueMap["root"]
		output = rootValue.AsString()
	} else {
		simpleJSONValue := ctyJson.SimpleJSONValue{Value: opValue}
		marshalJSON, err := simpleJSONValue.MarshalJSON()
		if err == nil {
			output = string(marshalJSON)
		} else {
			impl.logger.Errorw("error occurred while marshalling json value of parsed template", "err", err)
			return "", err
		}
	}
	return output, nil
}

func (impl *VariableTemplateParserImpl) checkAndUpdateDiagnosticError(diagnostics hcl.Diagnostics, template string, response *VariableParserResponse, errMsg string) bool {
	if !diagnostics.HasErrors() {
		return false
	}
	detailedError := diagnostics.Error()
	impl.logger.Errorw("error occurred while extracting variables from template", "template", template, "error", detailedError)
	response.Error = errors.New(errMsg)
	response.DetailedError = detailedError
	return true
}

//func (impl *VariableTemplateParserImpl) ParseTemplate(template string, values map[string]string) string {
//	//TODO KB: in case of yaml, need to convert it into JSON structure
//	output := template
//	template, _ = impl.convertToHclCompatible(JsonVariableTemplate, template)
//	ignoreDefaultedVariables := true
//	impl.logger.Debug("variable hcl template valueMap", values)
//	//TODO KB: need to check for variables whose value is not present in values map, throw error or ignore variable
//	hclExpression, diagnostics := hclsyntax.ParseExpression([]byte(template), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
//	if !diagnostics.HasErrors() {
//		hclVariables := hclExpression.Variables()
//		//variables := impl.extractVarNames(hclVariables)
//		hclVarValues := impl.getHclVarValues(values)
//		defaultedVars := impl.getDefaultedVariables(hclVariables, values)
//		if len(defaultedVars) > 0 && ignoreDefaultedVariables {
//			template = impl.ignoreDefaultedVars(template, defaultedVars)
//			hclExpression, diagnostics = hclsyntax.ParseExpression([]byte(template), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
//			if diagnostics.HasErrors() {
//				//TODO KB: throw error with proper variable names creating problem
//			}
//		}
//		opValue, valueDiagnostics := hclExpression.Value(&hcl.EvalContext{
//			Variables: hclVarValues,
//			Functions: impl.getDefaultMappedFunc(),
//		})
//		if !valueDiagnostics.HasErrors() {
//			//opValueMap := opValue.AsValueMap()
//			//rootValue := opValueMap["root"]
//			//output = rootValue.AsString()
//			simpleJSONValue := ctyJson.SimpleJSONValue{Value: opValue}
//			marshalJSON, err := simpleJSONValue.MarshalJSON()
//			if err == nil {
//				output = string(marshalJSON)
//			}
//		}
//	} else {
//		//TODO KB: handle this case
//	}
//	return output
//}

func (impl *VariableTemplateParserImpl) getDefaultMappedFunc() map[string]function.Function {
	return map[string]function.Function{
		"upper":  stdlib.UpperFunc,
		"toInt":  stdlib.IntFunc,
		"toBool": ParseBoolFunc,
		"split":  stdlib.SplitFunc,
	}
}

func (impl *VariableTemplateParserImpl) convertToHclCompatible(templateType VariableTemplateType, template string) (string, error) {
	if templateType == StringVariableTemplate {
		jsonStringify, err := json.Marshal(template)
		if err != nil {
			impl.logger.Errorw("error occurred while marshalling template, but continuing with the template", "err", err, "templateType", templateType)
			//return "", errors.New(InvalidTemplate)
		} else {
			template = string(jsonStringify)
		}
		template = fmt.Sprintf(`{"root":%s}`, template)
	}
	template = impl.diluteExistingHclVars(template, "\\$", "$")
	template = impl.diluteExistingHclVars(template, "%", "%")
	return impl.convertToHclExpression(template), nil
}

func (impl *VariableTemplateParserImpl) diluteExistingHclVars(template string, templateControlKeyword string, replaceKeyword string) string {
	hclVarRegex := regexp.MustCompile(templateControlKeyword + `\{`)
	indexesData := hclVarRegex.FindAllIndex([]byte(template), -1)
	var strBuilder strings.Builder
	strBuilder.Grow(len(template))
	currentIndex := 0
	for _, datum := range indexesData {
		startIndex := datum[0]
		endIndex := datum[1]
		strBuilder.WriteString(template[currentIndex:startIndex] + replaceKeyword + template[startIndex:endIndex])
		currentIndex = endIndex
	}
	if currentIndex <= len(template) {
		strBuilder.WriteString(template[currentIndex:])
	}
	output := strBuilder.String()
	return output
}

func (impl *VariableTemplateParserImpl) convertToHclExpression(template string) string {

	var devtronRegexCompiledPattern = regexp.MustCompile(impl.variableTemplateParserConfig.VariableExpressionRegex)
	indexesData := devtronRegexCompiledPattern.FindAllIndex([]byte(template), -1)
	var strBuilder strings.Builder
	strBuilder.Grow(len(template))
	currentIndex := 0
	for _, datum := range indexesData {
		startIndex := datum[0]
		endIndex := datum[1]
		strBuilder.WriteString(template[currentIndex:startIndex])
		initQuoteAdded := false
		if startIndex > 0 && template[startIndex-1] == '"' { // if quotes are already present then ignore
			strBuilder.WriteString("$")
		} else {
			initQuoteAdded = true
			strBuilder.WriteString("$")
			//strBuilder.WriteString("\"$")
		}
		strBuilder.WriteString(template[startIndex+2 : endIndex-1])
		if initQuoteAdded { // adding closing quote
			//strBuilder.WriteString("\"")
		}
		currentIndex = endIndex
	}
	if currentIndex <= len(template) {
		strBuilder.WriteString(template[currentIndex:])
	}
	output := strBuilder.String()
	return output
}

func (impl *VariableTemplateParserImpl) getHclVarValues(values map[string]string) map[string]cty.Value {
	variables := map[string]cty.Value{}
	for varName, varValue := range values {
		variables[varName] = cty.StringVal(varValue)
	}
	return variables
}

func (impl *VariableTemplateParserImpl) getDefaultedVariables(variables []hcl.Traversal, varValues map[string]string) []hcl.Traversal {
	var defaultedVars []hcl.Traversal
	for _, traversal := range variables {
		if _, ok := varValues[traversal.RootName()]; !ok {
			defaultedVars = append(defaultedVars, traversal)
		}
	}
	return defaultedVars
}

func (impl *VariableTemplateParserImpl) ignoreDefaultedVars(template string, hclVars []hcl.Traversal) string {
	var processedDtBuilder strings.Builder
	impl.logger.Info("ignoring defaulted vars", "vars", hclVars)
	maxSize := len(template) + len(hclVars)
	processedDtBuilder.Grow(maxSize)
	currentIndex := 0
	for _, hclVar := range hclVars {
		startIndex := hclVar.SourceRange().Start.Column
		endIndex := hclVar.SourceRange().End.Column
		startIndex = impl.getVarStartIndex(template, startIndex-1)
		if startIndex == -1 {
			continue
		}
		endIndex = impl.getVarEndIndex(template, endIndex-1)
		if endIndex == -1 {
			continue
		}
		processedDtBuilder.WriteString(template[currentIndex:startIndex] + "@{" + template[startIndex+1:endIndex] + "}")
		currentIndex = endIndex
	}
	if currentIndex <= maxSize {
		processedDtBuilder.WriteString(template[currentIndex:])
	}
	return processedDtBuilder.String()
}

func (impl *VariableTemplateParserImpl) getVarStartIndex(template string, startIndex int) int {
	currentIndex := startIndex - 1
	for ; currentIndex > 0; currentIndex-- {
		//fmt.Println("value", string(template[currentIndex]))
		if template[currentIndex] == '{' && template[currentIndex-1] == '$' {
			return currentIndex - 1
		}
	}
	return -1
}

func (impl *VariableTemplateParserImpl) getVarEndIndex(template string, endIndex int) int {
	currentIndex := endIndex
	for ; currentIndex < len(template); currentIndex++ {
		//fmt.Println("value", string(template[currentIndex]))
		if template[currentIndex] == '}' {
			return currentIndex
		}
	}
	return -1
}

var ParseBoolFunc = function.New(&function.Spec{
	Description: `convert to bool value`,
	Params: []function.Parameter{
		{
			Name:             "val",
			Type:             cty.String,
			AllowDynamicType: true,
			AllowMarked:      true,
		},
	},
	Type:         function.StaticReturnType(cty.Bool),
	RefineResult: refineNonNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		boolVal, err := strconv.ParseBool(args[0].AsString())
		return cty.BoolVal(boolVal), err
	},
})

func refineNonNull(b *cty.RefinementBuilder) *cty.RefinementBuilder {
	return b.NotNull()
}
