package resourceFilter

import (
	"go.uber.org/zap"
)

type ResourceFilterEvaluator interface {
	EvaluateFilter(filter *FilterMetaDataBean, expressionMetadata ExpressionMetadata) (bool, error)
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

func (impl *ResourceFilterEvaluatorImpl) EvaluateFilter(filter *FilterMetaDataBean, expressionMetadata ExpressionMetadata) (bool, error) {
	resourceConditions := filter.Conditions
	exprResponse := expressionResponse{}
	for _, resourceCondition := range resourceConditions {
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
