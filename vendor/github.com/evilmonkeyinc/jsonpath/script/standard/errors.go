package standard

import (
	"fmt"

	"github.com/evilmonkeyinc/jsonpath/errors"
)

var (
	errUnsupportedOperator               error = fmt.Errorf("unsupported operator")
	errInvalidArgument                   error = fmt.Errorf("invalid argument")
	errInvalidArgumentNil                error = fmt.Errorf("%w. is nil", errInvalidArgument)
	errInvalidArgumentExpectedInteger    error = fmt.Errorf("%w. expected integer", errInvalidArgument)
	errInvalidArgumentExpectedNumber     error = fmt.Errorf("%w. expected number", errInvalidArgument)
	errInvalidArgumentExpectedBoolean    error = fmt.Errorf("%w. expected boolean", errInvalidArgument)
	errInvalidArgumentExpectedRegex      error = fmt.Errorf("%w. expected a valid regexp", errInvalidArgument)
	errInvalidArgumentExpectedCollection error = fmt.Errorf("%w. expected array, map, or slice", errInvalidArgument)
)

func getInvalidExpressionEmptyError() error {
	return fmt.Errorf("%w. is empty", errors.ErrInvalidExpression)
}
