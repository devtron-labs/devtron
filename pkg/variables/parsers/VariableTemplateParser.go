package parsers

import (
	"errors"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"go.uber.org/zap"
	"regexp"
	"strings"
	_ "github.com/hashicorp/hcl2/hcl/hclsyntax"
)

type VariableTemplateParser interface {
	ExtractVariables(template string) ([]string, error)
	ParseTemplate(template string, values map[string]string) string
}

type VariableTemplateParserImpl struct {
	logger *zap.SugaredLogger
}

func NewVariableTemplateParserImpl(logger *zap.SugaredLogger) *VariableTemplateParserImpl {
	return &VariableTemplateParserImpl{logger: logger}
}

func (impl *VariableTemplateParserImpl) ExtractVariables(template string) ([]string, error) {
	var variables []string
	// preprocess existing template to comment
	template = impl.convertToHclCompatible(template)
	hclExpression, diagnostics := hclsyntax.ParseExpression([]byte(template), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
	if !diagnostics.HasErrors() {
		hclVariables := hclExpression.Variables()
		for _, hclVariable := range hclVariables {
			variables = append(variables, hclVariable.RootName())
		}
	} else {
		impl.logger.Errorw("error occurred while extracting variables from template", "template", template, "error", diagnostics.Error())
		//TODO KB: handle this case
		return variables, errors.New("invalid-template")
	}
	return variables, nil
}

func (impl *VariableTemplateParserImpl) ParseTemplate(template string, values map[string]string) string {
	template = impl.convertToHclCompatible(template)
	//TODO KB: need to check for variables whose value is not present in values map, throw error or ignore variable
	hclExpression, diagnostics := hclsyntax.ParseExpression([]byte(template), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
	if !diagnostics.HasErrors() {
		hclVarValues := impl.getHclVarValues(values)
		_, valueDiagnostics := hclExpression.Value(&hcl.EvalContext{
			Variables: hclVarValues,
			Functions: map[string]function.Function{
				"upper": stdlib.UpperFunc,
			},
		})
		if !valueDiagnostics.HasErrors() {

		}
	} else {
		//TODO KB: handle this case
	}
	return template //TODO KB: fix this
}

func (impl *VariableTemplateParserImpl) convertToHclCompatible(template string) string {
	template = impl.diluteExistingHclVars(template)
	return impl.convertToHclExpression(template)
}

func (impl *VariableTemplateParserImpl) diluteExistingHclVars(template string) string {
	hclVarRegex := regexp.MustCompile(`\$\{`)
	indexesData := hclVarRegex.FindAllIndex([]byte(template), -1)
	var strBuilder strings.Builder
	strBuilder.Grow(len(template))
	currentIndex := 0
	for _, datum := range indexesData {
		startIndex := datum[0]
		endIndex := datum[1]
		strBuilder.WriteString(template[currentIndex:startIndex] + "$" + template[startIndex:endIndex])
		currentIndex = endIndex
	}
	if currentIndex <= len(template) {
		strBuilder.WriteString(template[currentIndex:])
	}
	output := strBuilder.String()
	return output
}

func (impl *VariableTemplateParserImpl) convertToHclExpression(deploymentTemplate string) string {
	var devtronRegexCompiledPattern = regexp.MustCompile(`@\{\{[a-zA-Z0-9-+/*%\s]+\}\}`) //TODO KB: add support of Braces () also
	indexesData := devtronRegexCompiledPattern.FindAllIndex([]byte(deploymentTemplate), -1)
	var strBuilder strings.Builder
	strBuilder.Grow(len(deploymentTemplate))
	currentIndex := 0
	for _, datum := range indexesData {
		startIndex := datum[0]
		endIndex := datum[1]
		strBuilder.WriteString(deploymentTemplate[currentIndex:startIndex] + "$" + deploymentTemplate[startIndex+2:endIndex-1])
		currentIndex = endIndex
	}
	if currentIndex <= len(deploymentTemplate) {
		strBuilder.WriteString(deploymentTemplate[currentIndex:])
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
