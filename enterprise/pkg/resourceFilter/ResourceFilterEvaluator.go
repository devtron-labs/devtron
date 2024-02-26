package resourceFilter

import (
	"go.uber.org/zap"
)

type ResourceFilterEvaluator interface {
	EvaluateFilter(filterConditions []ResourceCondition, expressionMetadata ExpressionMetadata) (bool, error)
}

type ResourceFilterEvaluatorImpl struct {
	logger       *zap.SugaredLogger
	celEvaluator CELEvaluatorService
}

func NewResourceFilterEvaluatorImpl(logger *zap.SugaredLogger, celEvaluator CELEvaluatorService) (*ResourceFilterEvaluatorImpl, error) {
	return &ResourceFilterEvaluatorImpl{
		logger:       logger,
		celEvaluator: celEvaluator,
	}, nil
}

func (impl *ResourceFilterEvaluatorImpl) EvaluateFilter(filterConditions []ResourceCondition, expressionMetadata ExpressionMetadata) (bool, error) {
	exprResponse := expressionResponse{}
	for _, resourceCondition := range filterConditions {
		expression := resourceCondition.Expression
		celRequest := CELRequest{
			Expression:         expression,
			ExpressionMetadata: expressionMetadata,
		}
		response, err := impl.celEvaluator.EvaluateCELRequest(celRequest)
		if err != nil {
			return false, err
		}
		if resourceCondition.IsFailCondition() {
			exprResponse.blockConditionAvail = true
			exprResponse.blockResponse = response
		} else {
			exprResponse.allowConditionAvail = true
			exprResponse.allowResponse = response
		}
	}
	return exprResponse.getFinalResponse(), nil
}
