package util

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/util/validation"
)

func CheckIfValidLabel(labelKey string, labelValue string) error {
	errs := validation.IsQualifiedName(labelKey)
	if len(errs) > 0 {
		return errors.New(fmt.Sprintf("Validation error - label key - %s is not satisfying the label key criteria", labelKey))
	}

	errs = validation.IsValidLabelValue(labelValue)
	if len(errs) > 0 {
		return errors.New(fmt.Sprintf("Validation error - label value - %s is not satisfying the label value criteria for label key - %s", labelValue, labelKey))
	}
	return nil
}
