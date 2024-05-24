package cel

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"go.uber.org/zap"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type EvaluatorService interface {
	EvaluateCELRequest(request Request) (bool, error)
	Validate(request Request) (*cel.Ast, *cel.Env, error)
}

type EvaluatorServiceImpl struct {
	logger *zap.SugaredLogger
}

func NewCELServiceImpl(logger *zap.SugaredLogger) *EvaluatorServiceImpl {
	return &EvaluatorServiceImpl{
		logger: logger,
	}
}

func (impl *EvaluatorServiceImpl) EvaluateCELRequest(request Request) (bool, error) {

	ast, env, err := impl.Validate(request)
	if err != nil {
		impl.logger.Errorw("error occurred while validating CEL request", "request", request, "err", err)
		return false, err
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("program construction error: %s", err)
	}

	expressionMetadata := request.ExpressionMetadata
	valuesMap := make(map[string]interface{})
	for _, param := range expressionMetadata.Params {
		valuesMap[string(param.ParamName)] = param.Value
	}

	out, _, err := prg.Eval(valuesMap)
	if err != nil {
		return false, err
	}

	if boolValue, ok := out.Value().(bool); ok {
		return boolValue, nil
	}

	return false, fmt.Errorf("expression did not evaluate to a boolean")

}

func (impl *EvaluatorServiceImpl) Validate(request Request) (*cel.Ast, *cel.Env, error) {

	var declarations []*expr.Decl

	expressionMetadata := request.ExpressionMetadata
	for _, param := range expressionMetadata.Params {
		declsType, err := getDeclarationType(param.Type)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid parameter type '%s' for '%s': %v", param.Type, param.Type, err)
		}
		declaration := decls.NewVar(string(param.ParamName), declsType)
		declarations = append(declarations, declaration)
	}

	env, err := cel.NewEnv(
		cel.Declarations(declarations...),
	)

	if err != nil {
		return nil, nil, err
	}

	ast, issues := env.Compile(request.Expression)
	if issues != nil && issues.Err() != nil {
		return nil, nil, fmt.Errorf("type-check error: %s", issues.Err())
	}

	return ast, env, nil
}

func getDeclarationType(paramType ParamValuesType) (*expr.Type, error) {
	switch paramType {
	case ParamTypeString:
		return decls.String, nil
	case ParamTypeObject:
		return decls.Dyn, nil
	case ParamTypeInteger:
		return decls.Int, nil
	case ParamTypeBool:
		return decls.Bool, nil
	case ParamTypeList:
		return decls.NewListType(decls.String), nil
	case ParamTypeMapStringToAny:
		return decls.NewMapType(decls.String, decls.Dyn), nil
	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", paramType)
	}
}
