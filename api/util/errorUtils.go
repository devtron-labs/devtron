/*
 * Copyright (c) 2024. Devtron Inc.
 */

package util

import (
	"errors"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
)

func CustomizeValidationError(err error) error {
	if validatorErr := (validator.ValidationErrors{}); errors.As(err, &validatorErr) {
		fieldErr := validatorErr[0]
		switch fieldErr.Tag() {
		case "required":
			return fmt.Errorf("field %s is required %v", fieldErr.Field(), err)
		case "oneof":
			return fmt.Errorf("value %s is unsupported  %v", fieldErr.Value(), err)
		}
	}
	return err
}
