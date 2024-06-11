/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
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
