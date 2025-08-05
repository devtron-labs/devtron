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
	"fmt"
	"github.com/devtron-labs/devtron/internal/constants"
	"net/http"
)

// ResourceContext holds information about the resource for better error messages
type ResourceContext struct {
	ResourceType string
	ResourceId   string
	Operation    string
}

// NewResourceNotFoundError creates a user-friendly error for resource not found scenarios
// Leverages existing util.NewApiError() function
func NewResourceNotFoundError(resourceType, resourceId string) *ApiError {
	return NewApiError(
		http.StatusNotFound,
		fmt.Sprintf("%s with ID '%s' not found", resourceType, resourceId),
		fmt.Sprintf("%s not found: %s", resourceType, resourceId),
	).WithCode(constants.ResourceNotFound).
		WithUserDetailMessage(fmt.Sprintf("The requested %s does not exist or has been deleted.", resourceType))
}

// NewDuplicateResourceError creates a user-friendly error for duplicate resource scenarios
func NewDuplicateResourceError(resourceType, resourceName string) *ApiError {
	return NewApiError(
		http.StatusConflict,
		fmt.Sprintf("%s with name '%s' already exists", resourceType, resourceName),
		fmt.Sprintf("duplicate %s: %s", resourceType, resourceName),
	).WithCode(constants.DuplicateResource).
		WithUserDetailMessage(fmt.Sprintf("A %s with this name already exists. Please choose a different name.", resourceType))
}

// NewValidationErrorForField creates a user-friendly error for field validation failures
func NewValidationErrorForField(fieldName, reason string) *ApiError {
	return NewApiError(
		http.StatusBadRequest,
		fmt.Sprintf("Validation failed for field '%s': %s", fieldName, reason),
		fmt.Sprintf("validation failed for %s: %s", fieldName, reason),
	).WithCode(constants.ValidationFailed).
		WithUserDetailMessage("Please check the field value and try again.")
}

// NewInvalidPathParameterError creates a user-friendly error for invalid path parameters
func NewInvalidPathParameterError(paramName, paramValue string) *ApiError {
	return NewApiError(
		http.StatusBadRequest,
		fmt.Sprintf("Invalid path parameter '%s'", paramName),
		fmt.Sprintf("invalid path parameter %s: %s", paramName, paramValue),
	).WithCode(constants.InvalidPathParameter).
		WithUserDetailMessage("Please check the parameter format and try again.")
}

// NewMissingRequiredFieldError creates a user-friendly error for missing required fields
func NewMissingRequiredFieldError(fieldName string) *ApiError {
	return NewApiError(
		http.StatusBadRequest,
		fmt.Sprintf("Required field '%s' is missing", fieldName),
		fmt.Sprintf("missing required field: %s", fieldName),
	).WithCode(constants.MissingRequiredField).
		WithUserDetailMessage("Please provide all required fields and try again.")
}

// NewGenericResourceNotFoundError creates a generic not found error when resource context is unknown
func NewGenericResourceNotFoundError() *ApiError {
	return NewApiError(
		http.StatusNotFound,
		"Requested resource not found",
		"resource not found",
	).WithCode(constants.ResourceNotFound).
		WithUserDetailMessage("The requested resource does not exist or has been deleted.")
}
