package jsonpath

import (
	goErr "errors"
	"fmt"

	"github.com/evilmonkeyinc/jsonpath/errors"
)

var (
	errDataIsUnexpectedTypeOrNil error = fmt.Errorf("unexpected type or nil")
	errOptionAlreadySet          error = fmt.Errorf("option already set")
)

func getInvalidJSONData(reason error) error {
	return fmt.Errorf("%w. %s", errors.ErrInvalidJSONData, reason.Error())
}

func getInvalidJSONPathSelector(selector string) error {
	return fmt.Errorf("%w '%s'", errors.ErrInvalidJSONPathSelector, selector)
}

func getInvalidJSONPathSelectorWithReason(selector string, reason error) error {
	if goErr.Is(reason, errors.ErrInvalidJSONPathSelector) {
		return reason
	}
	return fmt.Errorf("%w '%s' %s", errors.ErrInvalidJSONPathSelector, selector, reason.Error())
}
