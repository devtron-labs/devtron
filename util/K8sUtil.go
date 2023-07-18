package util

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/util/validation"
	"strings"
)

func CheckIfValidLabel(labelKey string, labelValue string) error {
	labelKey = strings.TrimSpace(labelKey)
	labelValue = strings.TrimSpace(labelValue)

	errs := validation.IsQualifiedName(labelKey)
	if len(labelKey) == 0 || len(errs) > 0 {
		return errors.New(fmt.Sprintf("Validation error - label key - %s is not satisfying the label key criteria", labelKey))
	}

	errs = validation.IsValidLabelValue(labelValue)
	if len(labelValue) == 0 || len(errs) > 0 {
		return errors.New(fmt.Sprintf("Validation error - label value - %s is not satisfying the label value criteria for label key - %s", labelValue, labelKey))
	}
	return nil
}
