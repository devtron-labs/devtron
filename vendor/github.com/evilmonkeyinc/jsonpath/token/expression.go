package token

import (
	"fmt"

	"github.com/evilmonkeyinc/jsonpath/option"
	"github.com/evilmonkeyinc/jsonpath/script"
)

func newExpressionToken(expression string, engine script.Engine, options *option.QueryOptions) (*expressionToken, error) {
	compiledExpression, err := engine.Compile(expression, options)
	if err != nil {
		return nil, err
	}

	return &expressionToken{
		expression:         expression,
		compiledExpression: compiledExpression,
		options:            options,
	}, nil
}

type expressionToken struct {
	expression         string
	compiledExpression script.CompiledExpression
	options            *option.QueryOptions
}

func (token *expressionToken) String() string {
	return fmt.Sprintf("(%s)", token.expression)
}

func (token *expressionToken) Type() string {
	return "expression"
}

func (token *expressionToken) Apply(root, current interface{}, next []Token) (interface{}, error) {
	if token.expression == "" {
		return nil, getInvalidExpressionEmptyError()
	}

	value, err := token.compiledExpression.Evaluate(root, current)
	if err != nil {
		return nil, getInvalidExpressionError(err)
	}

	if len(next) > 0 {
		return next[0].Apply(root, value, next[1:])
	}

	return value, nil
}
