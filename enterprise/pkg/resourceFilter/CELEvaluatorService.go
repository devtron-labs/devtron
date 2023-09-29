package resourceFilter

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"go.uber.org/zap"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"strings"
)

type CELEvaluatorService interface {
	EvaluateCELRequest(request CELRequest) (bool, error)
	ValidateCELRequest(request ValidateRequestResponse) (ValidateRequestResponse, bool)
	GetParamsFromArtifact(artifact string) []ExpressionParam
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
	BLOCK FilterState = 0
	ALLOW FilterState = 1
	ERROR FilterState = 2
)

type ValidateRequestResponse struct {
	Conditions []ResourceCondition `json:"conditions"`
}

type CELRequest struct {
	Expression         string             `json:"expression"`
	ExpressionMetadata ExpressionMetadata `json:"params"`
}

func (impl *CELServiceImpl) EvaluateCELRequest(request CELRequest) (bool, error) {

	ast, env, err := impl.validate(request)
	if err != nil {
		impl.Logger.Errorw("error occurred while validating CEL request", "request", request, "err", err)
		return false, err
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("program construction error: %s", err)
	}

	expressionMetadata := request.ExpressionMetadata
	valuesMap := make(map[string]interface{})
	for _, param := range expressionMetadata.Params {
		valuesMap[param.ParamName] = param.Value
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

func (impl *CELServiceImpl) validate(request CELRequest) (*cel.Ast, *cel.Env, error) {

	var declarations []*expr.Decl

	expressionMetadata := request.ExpressionMetadata
	for _, param := range expressionMetadata.Params {
		declsType, err := getDeclarationType(param.Type)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid parameter type '%s' for '%s': %v", param.Type, param.Type, err)
		}
		declaration := decls.NewVar(param.ParamName, declsType)
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
			ParamName: "containerRepository",
			Type:      ParamTypeString,
		},
		{
			ParamName: "containerImage",
			Type:      ParamTypeString,
		},
		{
			ParamName: "containerImageTag",
			Type:      ParamTypeString,
		},
	}

	for i, e := range request.Conditions {
		validateExpression := CELRequest{
			Expression:         e.Expression,
			ExpressionMetadata: ExpressionMetadata{Params: params},
		}
		_, _, err := impl.validate(validateExpression)
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
	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", paramType)
	}
}

func (impl *CELServiceImpl) GetParamsFromArtifact(artifact string) []ExpressionParam {

	lastColonIndex := strings.LastIndex(artifact, ":")

	containerRepository := artifact[:lastColonIndex]
	containerImageTag := artifact[lastColonIndex+1:]
	containerImage := artifact
	params := []ExpressionParam{
		{
			ParamName: "containerRepository",
			Value:     containerRepository,
			Type:      ParamTypeString,
		},
		{
			ParamName: "containerImage",
			Value:     containerImage,
			Type:      ParamTypeString,
		},
		{
			ParamName: "containerImageTag",
			Value:     containerImageTag,
			Type:      ParamTypeString,
		},
	}

	return params
}
