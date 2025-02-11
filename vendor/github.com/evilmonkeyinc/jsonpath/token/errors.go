package token

import (
	goErr "errors"
	"fmt"
	"reflect"

	"github.com/evilmonkeyinc/jsonpath/errors"
)

// TODO : after refactor ensure all these are still being used

func isInvalidTokenTargetError(err error) bool {
	return goErr.Is(err, errors.ErrInvalidTokenTarget)
}

func getInvalidExpressionEmptyError() error {
	return fmt.Errorf("%w. is empty", errors.ErrInvalidExpression)
}

func getInvalidExpressionError(reason error) error {
	if goErr.Is(reason, errors.ErrInvalidExpression) {
		return reason
	}
	return fmt.Errorf("%w. %s", errors.ErrInvalidExpression, reason.Error())
}

func getInvalidExpressionFormatError(format string) error {
	return fmt.Errorf("%w. invalid format '%s'", errors.ErrInvalidExpression, format)
}

func getInvalidTokenArgumentError(tokenType string, got reflect.Kind, expected ...reflect.Kind) error {
	return fmt.Errorf("%s: %w argument. expected %v got [%v]", tokenType, errors.ErrInvalidToken, expected, got)
}

func getInvalidTokenArgumentNilError(tokenType string, expected ...reflect.Kind) error {
	return fmt.Errorf("%s: %w argument. expected %v got [nil]", tokenType, errors.ErrInvalidToken, expected)
}

func getInvalidTokenEmpty() error {
	return fmt.Errorf("%w. token string is empty", errors.ErrInvalidToken)
}

func getInvalidTokenError(tokenType string, reason error) error {
	if goErr.Is(reason, errors.ErrInvalidToken) {
		return reason
	}
	return fmt.Errorf("%s: %w %s", tokenType, errors.ErrInvalidToken, reason.Error())
}

func getInvalidTokenFormatError(tokenString string) error {
	return fmt.Errorf("%w. '%s' does not match any token format", errors.ErrInvalidToken, tokenString)
}

func getInvalidTokenKeyNotFoundError(tokenType, key string) error {
	return fmt.Errorf("%s: %w key '%s' not found", tokenType, errors.ErrInvalidToken, key)
}

func getInvalidTokenOutOfRangeError(tokenType string) error {
	return fmt.Errorf("%s: %w out of range", tokenType, errors.ErrInvalidToken)
}

func getInvalidTokenTargetError(tokenType string, got reflect.Kind, expected ...reflect.Kind) error {
	return fmt.Errorf("%s: %w. expected %v got [%v]", tokenType, errors.ErrInvalidTokenTarget, expected, got)
}

func getInvalidTokenTargetNilError(tokenType string, expected ...reflect.Kind) error {
	return fmt.Errorf("%s: %w. expected %v got [nil]", tokenType, errors.ErrInvalidTokenTarget, expected)
}

func getUnexpectedExpressionResultError(got reflect.Kind, expected ...reflect.Kind) error {
	return fmt.Errorf("%w. expected %v got [%v]", errors.ErrUnexpectedExpressionResult, expected, got)
}

func getUnexpectedExpressionResultNilError(expected ...reflect.Kind) error {
	return fmt.Errorf("%w. expected %v got [nil]", errors.ErrUnexpectedExpressionResult, expected)
}

func getUnexpectedTokenError(tokenType string, index int) error {
	return fmt.Errorf("%w '%s' at index %d", errors.ErrUnexpectedToken, tokenType, index)
}
