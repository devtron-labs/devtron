package expressionEvaluators

import (
	"fmt"
	"github.com/devtron-labs/devtron/util"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types/ref"
	"go.uber.org/zap"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

type CELEvaluatorService interface {
	EvaluateCELForBool(request CELRequest) (bool, error)
	EvaluateCELForObject(request CELRequest) (interface{}, error)
	Validate(request CELRequest) (*cel.Ast, *cel.Env, error)
	ValidateCELRequest(request ValidateRequestResponse) (ValidateRequestResponse, bool)
}

type CELServiceImpl struct {
	Logger *zap.SugaredLogger
}

func NewCELServiceImpl(logger *zap.SugaredLogger) *CELServiceImpl {
	return &CELServiceImpl{
		Logger: logger,
	}
}

type FilterState int

const (
	ALLOW FilterState = 0
	BLOCK FilterState = 1
	ERROR FilterState = 2
)

type ValidateRequestResponse struct {
	Conditions []util.ResourceCondition `json:"conditions"`
}

type CELRequest struct {
	Expression         string             `json:"expression"`
	ExpressionMetadata ExpressionMetadata `json:"params"`
}

func (impl *CELServiceImpl) EvaluateCELForObject(request CELRequest) (interface{}, error) {
	outValue, err := impl.evaluateCEL(request)
	if err != nil {
		return false, err
	}
	return outValue.Value(), nil
}

func (impl *CELServiceImpl) EvaluateCELForBool(request CELRequest) (bool, error) {

	out, err := impl.evaluateCEL(request)
	if err != nil {
		return false, err
	}
	outValue := out.Value()
	if boolValue, ok := outValue.(bool); ok {
		return boolValue, nil
	}

	return false, fmt.Errorf("expression did not evaluate to a boolean")

}

func (impl *CELServiceImpl) evaluateCEL(request CELRequest) (ref.Val, error) {
	ast, env, err := impl.Validate(request)
	if err != nil {
		impl.Logger.Errorw("error occurred while validating CEL request", "request", request, "err", err)
		return nil, err
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("program construction error: %s", err)
	}

	expressionMetadata := request.ExpressionMetadata
	valuesMap := make(map[string]interface{})
	for _, param := range expressionMetadata.Params {
		valuesMap[string(param.ParamName)] = param.Value
	}

	out, _, err := prg.Eval(valuesMap)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (impl *CELServiceImpl) Validate(request CELRequest) (*cel.Ast, *cel.Env, error) {

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

func (impl *CELServiceImpl) ValidateCELRequest(request ValidateRequestResponse) (ValidateRequestResponse, bool) {
	errored := false
	params := []ExpressionParam{
		{
			ParamName: ContainerRepo,
			Type:      ParamTypeString,
		},
		{
			ParamName: ContainerImage,
			Type:      ParamTypeString,
		},
		{
			ParamName: ContainerImageTag,
			Type:      ParamTypeString,
		},
		{
			ParamName: ImageLabels,
			Type:      ParamTypeList,
		},
		{
			ParamName: GitCommitDetails,
			Type:      ParamTypeCommitDetailsMap,
		},
	}

	for i, e := range request.Conditions {
		validateExpression := CELRequest{
			Expression:         e.Expression,
			ExpressionMetadata: ExpressionMetadata{Params: params},
		}
		_, _, err := impl.Validate(validateExpression)
		if err != nil {
			errored = true
			e.ErrorMsg = err.Error()
		}
		request.Conditions[i] = e
	}

	return request, errored
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
	case ParamTypeCommitDetailsMap:
		return decls.NewMapType(decls.String, decls.Dyn), nil
	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", paramType)
	}
}
