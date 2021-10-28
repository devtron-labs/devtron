/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package util

import (
	"gopkg.in/go-playground/validator.v9"
	"regexp"
	"strings"
)

func ValidateName(fl validator.FieldLevel) bool {
	hostnameRegexString := `^[a-z]+[a-z0-9\-\?]*[a-z0-9]+$`
	hostnameRegexRFC952 := regexp.MustCompile(hostnameRegexString)
	return hostnameRegexRFC952.MatchString(fl.Field().String())
}

func ValidateCheckoutPath(fl validator.FieldLevel) bool {
	checkoutPath := fl.Field().String()
	if checkoutPath != "" && (!strings.HasPrefix(checkoutPath, "./")) {
		return false
	}
	return true
}

func validateAppLabel(fl validator.FieldLevel) bool {
	label := fl.Field().String()
	if len(label) == 0 {
		return false
	}
	index := strings.Index(label, ":")
	if index == -1 || index == 0 || index == len(label)-1 {
		return false
	}
	/*kv := strings.Split(label, ":")
	if len(kv) != 2 {
		return false
	}*/
	return true
}

func validateNonEmptyUrl(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if len(value) == 0 {
		return true
	}
	return IsValidUrl(value)
}

func IntValidator() (*validator.Validate, error) {
	v := validator.New()
	err := v.RegisterValidation("name-component", ValidateName)
	if err != nil {
		return v, err
	}
	err = v.RegisterValidation("checkout-path-component", ValidateCheckoutPath)
	if err != nil {
		return v, err
	}
	err = v.RegisterValidation("app-label-component", validateAppLabel)
	if err != nil {
		return v, err
	}
	err = v.RegisterValidation("validate-non-empty-url", validateNonEmptyUrl)
	if err != nil {
		return v, err
	}
	return v, err
}
