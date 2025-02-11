package errors

import (
	"fmt"
)

var (
	// ErrInvalidExpression returned when an expression is invalid
	ErrInvalidExpression error = fmt.Errorf("invalid expression")
	// ErrInvalidJSONPathSelector returned when the JSONPath selector is invalid
	ErrInvalidJSONPathSelector error = fmt.Errorf("invalid JSONPath selector")
	// ErrInvalidJSONData returned when the JSON data is invalid
	ErrInvalidJSONData error = fmt.Errorf("invalid data")
	// ErrInvalidToken returned when a token is invalid
	ErrInvalidToken error = fmt.Errorf("invalid token")
	// ErrInvalidTokenTarget returned when a token parses an invalid target
	ErrInvalidTokenTarget error = fmt.Errorf("%w target", ErrInvalidToken)
	// ErrUnexpectedExpressionResult returned when an expression unexpected result
	ErrUnexpectedExpressionResult error = fmt.Errorf("unexpected expression result")
	// ErrUnexpectedToken returned when an unexpected token string is parsed
	ErrUnexpectedToken error = fmt.Errorf("unexpected token")
)
