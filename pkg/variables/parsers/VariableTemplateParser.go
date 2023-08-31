package parsers

import (
	"errors"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	_ "github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"github.com/zclconf/go-cty/cty/json"
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"strings"
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
		variables = impl.extractVarNames(hclVariables)
	} else {
		impl.logger.Errorw("error occurred while extracting variables from template", "template", template, "error", diagnostics.Error())
		//TODO KB: handle this case
		return variables, errors.New("invalid-template")
	}
	impl.logger.Info("variables:", variables)
	return variables, nil
}

func (impl *VariableTemplateParserImpl) extractVarNames(hclVariables []hcl.Traversal) []string {
	var variables []string
	for _, hclVariable := range hclVariables {
		variables = append(variables, hclVariable.RootName())
	}
	return variables
}

func (impl *VariableTemplateParserImpl) ParseTemplate(template string, values map[string]string) string {
	//TODO KB: in case of yaml, need to convert it into JSON structure
	output := template
	template = impl.convertToHclCompatible(template)
	ignoreDefaultedVariables := true
	impl.logger.Debug("valueMap", values)
	//TODO KB: need to check for variables whose value is not present in values map, throw error or ignore variable
	hclExpression, diagnostics := hclsyntax.ParseExpression([]byte(template), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
	if !diagnostics.HasErrors() {
		hclVariables := hclExpression.Variables()
		//variables := impl.extractVarNames(hclVariables)
		hclVarValues := impl.getHclVarValues(values)
		defaultedVars := impl.getDefaultedVariables(hclVariables, values)
		if len(defaultedVars) > 0 && ignoreDefaultedVariables {
			template = impl.ignoreDefaultedVars(template, defaultedVars)
			hclExpression, diagnostics = hclsyntax.ParseExpression([]byte(template), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
			if diagnostics.HasErrors() {
				//TODO KB: throw error with proper variable names creating problem
			}
		}
		opValue, valueDiagnostics := hclExpression.Value(&hcl.EvalContext{
			Variables: hclVarValues,
			Functions: impl.getDefaultMappedFunc(),
		})
		if !valueDiagnostics.HasErrors() {
			//opValueMap := opValue.AsValueMap()
			//rootValue := opValueMap["root"]
			//output = rootValue.AsString()
			simpleJSONValue := json.SimpleJSONValue{Value: opValue}
			marshalJSON, err := simpleJSONValue.MarshalJSON()
			if err == nil {
				output = string(marshalJSON)
			}
		}
	} else {
		//TODO KB: handle this case
	}
	return output
}

func (impl *VariableTemplateParserImpl) getDefaultMappedFunc() map[string]function.Function {
	return map[string]function.Function{
		"upper":  stdlib.UpperFunc,
		"toInt":  stdlib.IntFunc,
		"toBool": ParseBoolFunc,
	}
}

func (impl *VariableTemplateParserImpl) convertToHclCompatible(template string) string {
	//jsonStringify, err := json.Marshal(template)
	//if err != nil {
	//	impl.logger.Errorw("error occurred while marshalling template", "err", err, "template", template)
	//} else {
	//	template = string(jsonStringify)
	//}
	//template = fmt.Sprintf(`{"root":%s}`, template)
	//fmt.Println("template", template)
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

func (impl *VariableTemplateParserImpl) convertToHclExpression(template string) string {
	var devtronRegexCompiledPattern = regexp.MustCompile(`@\{\{[a-zA-Z0-9-+/*%_\s]+\}\}`) //TODO KB: add support of Braces () also
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
	//impl.logger.Info("template for ignore case", template)
	maxSize := len(template) + len(hclVars)
	processedDtBuilder.Grow(maxSize)
	currentIndex := 0
	for _, hclVar := range hclVars {
		startIndex := hclVar.SourceRange().Start.Column
		endIndex := hclVar.SourceRange().End.Column
		//fmt.Println("variable: ", template[startIndex:endIndex])
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
