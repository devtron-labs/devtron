package pipeline

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"go.uber.org/zap"
	expr "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"strings"
)

type CELService interface {
	EvaluateCELRequest(request CELRequest) (bool, error)
	ValidateCELRequest(request CELRequest) (*cel.Ast, *cel.Env, error)
	GetParamsFromArtifact(artifact string) []CELParam
}

type CELServiceImpl struct {
	Logger *zap.SugaredLogger
}

func NewCELServiceImpl(logger *zap.SugaredLogger) *CELServiceImpl {
	return &CELServiceImpl{
		Logger: logger,
	}
}

type ParamValuesType string

const (
	ParamTypeString  ParamValuesType = "string"
	ParamTypeObject  ParamValuesType = "object"
	ParamTypeInteger ParamValuesType = "integer"
)

type CELRequest struct {
	Expression string     `json:"expression"`
	Params     []CELParam `json:"params"`
}

//type CELResponse struct {
//	Request           CELRequest
//	IsExpressionValid bool   `json:"isExpressionValid"`
//	err               string `json:"err"`
//}

type CELParam struct {
	ParamName string          `json:"paramName"`
	Value     interface{}     `json:"value"`
	Type      ParamValuesType `json:"type"`
}

func (impl *CELServiceImpl) EvaluateCELRequest(request CELRequest) (bool, error) {

	ast, env, err := impl.ValidateCELRequest(request)
	if err != nil {
		impl.Logger.Errorw("error occurred while validating CEL request", "request", request)
		return false, err
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("program construction error: %s", err)
	}

	valuesMap := make(map[string]interface{})
	for _, param := range request.Params {
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

func (impl *CELServiceImpl) ValidateCELRequest(request CELRequest) (*cel.Ast, *cel.Env, error) {

	var declarations []*expr.Decl

	for _, param := range request.Params {
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

func (impl *CELServiceImpl) GetParamsFromArtifact(artifact string) []CELParam {

	lastColonIndex := strings.LastIndex(artifact, ":")

	containerName := artifact[:lastColonIndex]
	image := artifact[lastColonIndex+1:]
	params := []CELParam{
		{
			ParamName: "containerName",
			Value:     containerName,
			Type:      ParamTypeString,
		},
		{
			ParamName: "image",
			Value:     image,
			Type:      ParamTypeString,
		},
	}
	return params
}

//func Evaluate(request CELRequest) (bool, error) {
//
//	ast, env, err := Validate(request)
//	if err != nil {
//		return false, err
//	}
//
//	prg, err := env.Program(ast)
//	if err != nil {
//		return false, fmt.Errorf("program construction error: %s", err)
//	}
//
//	valuesMap := make(map[string]interface{})
//	for _, param := range request.Params {
//		valuesMap[param.ParamName] = param.Value
//	}
//
//	out, _, err := prg.Eval(valuesMap)
//	if err != nil {
//		return false, err
//	}
//
//	if boolValue, ok := out.Value().(bool); ok {
//		return boolValue, nil
//	}
//
//	return false, fmt.Errorf("expression did not evaluate to a boolean")
//
//}
//
//func Validate(request CELRequest) (*cel.Ast, *cel.Env, error) {
//
//	var declarations []*expr.Decl
//
//	for _, param := range request.Params {
//		declsType, err := getDeclarationType(param.Type)
//		if err != nil {
//			return nil, nil, fmt.Errorf("invalid parameter type '%s' for '%s': %v", param.Type, param.Type, err)
//		}
//		declaration := decls.NewVar(param.ParamName, declsType)
//		declarations = append(declarations, declaration)
//	}
//
//	env, err := cel.NewEnv(
//		cel.Declarations(declarations...),
//	)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	ast, issues := env.Compile(request.Expression)
//	if issues != nil && issues.Err() != nil {
//		return nil, nil, fmt.Errorf("type-check error: %s", issues.Err())
//	}
//
//	return ast, env, nil
//}
//
