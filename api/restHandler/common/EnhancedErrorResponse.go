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

package common

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/middleware"
	"github.com/devtron-labs/devtron/internal/util"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

// ErrorResponseBuilder provides a fluent interface for building error responses
type ErrorResponseBuilder struct {
	request      *http.Request
	writer       http.ResponseWriter
	operation    string
	resourceType string
	resourceID   string
}

// NewErrorResponseBuilder creates a new error response builder
func NewErrorResponseBuilder(w http.ResponseWriter, r *http.Request) *ErrorResponseBuilder {
	// Try to extract resource context from request
	reqCtx := middleware.GetRequestContext(r)
	resourceType := ""
	resourceID := ""

	if reqCtx != nil {
		resourceType = reqCtx.ResourceType
		resourceID = reqCtx.ResourceID
	}

	return &ErrorResponseBuilder{
		request:      r,
		writer:       w,
		resourceType: resourceType,
		resourceID:   resourceID,
	}
}

// WithOperation sets the operation context for better error messages
func (erb *ErrorResponseBuilder) WithOperation(operation string) *ErrorResponseBuilder {
	erb.operation = operation
	return erb
}

// WithResource sets the resource context for better error messages
func (erb *ErrorResponseBuilder) WithResource(resourceType, resourceID string) *ErrorResponseBuilder {
	erb.resourceType = resourceType
	erb.resourceID = resourceID
	return erb
}

// WithResourceFromId sets the resource context with an integer ID
func (erb *ErrorResponseBuilder) WithResourceFromId(resourceType string, resourceID int) *ErrorResponseBuilder {
	erb.resourceType = resourceType
	erb.resourceID = strconv.Itoa(resourceID)
	return erb
}

// HandleError processes an error and writes an appropriate response
func (erb *ErrorResponseBuilder) HandleError(err error) {
	if err == nil {
		return
	}

	// If it's already an ApiError, use it directly
	if apiErr, ok := err.(*util.ApiError); ok {
		WriteJsonResp(erb.writer, apiErr, nil, apiErr.HttpStatusCode)
		return
	}

	// Handle database errors
	if util.IsErrNoRows(err) {
		if erb.resourceType != "" && erb.resourceID != "" {
			apiErr := util.NewResourceNotFoundError(erb.resourceType, erb.resourceID)
			WriteJsonResp(erb.writer, apiErr, nil, apiErr.HttpStatusCode)
		} else {
			apiErr := util.NewGenericResourceNotFoundError()
			WriteJsonResp(erb.writer, apiErr, nil, apiErr.HttpStatusCode)
		}
		return
	}

	// Handle validation errors
	if isValidationError(err) {
		apiErr := util.NewApiError(http.StatusBadRequest, "Validation failed", err.Error()).
			WithCode("11004")
		WriteJsonResp(erb.writer, apiErr, nil, apiErr.HttpStatusCode)
		return
	}

	// Handle business logic errors (check for common patterns)
	if isBusinessLogicError(err) {
		apiErr := util.NewApiError(http.StatusConflict, "Operation failed", err.Error()).
			WithCode("11008")
		WriteJsonResp(erb.writer, apiErr, nil, apiErr.HttpStatusCode)
		return
	}

	// Default to internal server error
	operation := erb.operation
	if operation == "" {
		operation = "operation"
	}

	apiErr := util.NewApiError(http.StatusInternalServerError,
		fmt.Sprintf("Internal server error during %s", operation),
		err.Error()).WithCode("11009")
	WriteJsonResp(erb.writer, apiErr, nil, apiErr.HttpStatusCode)
}

// HandleSuccess writes a successful response
func (erb *ErrorResponseBuilder) HandleSuccess(data interface{}) {
	WriteJsonResp(erb.writer, nil, data, http.StatusOK)
}

// isValidationError checks if the error is a validation error
func isValidationError(err error) bool {
	errMsg := err.Error()
	// Common validation error patterns
	validationPatterns := []string{
		"validation failed",
		"invalid input",
		"required field",
		"invalid format",
		"constraint violation",
	}

	for _, pattern := range validationPatterns {
		if contains(errMsg, pattern) {
			return true
		}
	}
	return false
}

// isBusinessLogicError checks if the error is a business logic error
func isBusinessLogicError(err error) bool {
	errMsg := err.Error()
	// Common business logic error patterns
	businessPatterns := []string{
		"already exists",
		"duplicate",
		"conflict",
		"not allowed",
		"permission denied",
		"unauthorized",
		"forbidden",
	}

	for _, pattern := range businessPatterns {
		if contains(errMsg, pattern) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

// containsSubstring checks if a string contains a substring anywhere
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Convenience functions for common error scenarios

// HandleParameterError handles path parameter validation errors
func HandleParameterError(w http.ResponseWriter, r *http.Request, paramName, paramValue string) {
	apiErr := util.NewInvalidPathParameterError(paramName, paramValue)
	WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
}

// HandleResourceNotFound handles resource not found errors
func HandleResourceNotFound(w http.ResponseWriter, r *http.Request, resourceType, resourceID string) {
	apiErr := util.NewResourceNotFoundError(resourceType, resourceID)
	WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
}

// HandleUnauthorized handles unauthorized access errors
func HandleUnauthorized(w http.ResponseWriter, r *http.Request) {
	apiErr := util.NewApiError(http.StatusUnauthorized, "Unauthorized access", "unauthorized").
		WithCode("11010")
	WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
}

// HandleForbidden handles forbidden access errors
func HandleForbidden(w http.ResponseWriter, r *http.Request, resource string) {
	apiErr := util.NewApiError(http.StatusForbidden,
		fmt.Sprintf("Access denied for %s", resource),
		"forbidden").WithCode("11011")
	WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
}

// HandleValidationError handles validation errors
func HandleValidationError(w http.ResponseWriter, r *http.Request, fieldName, message string) {
	apiErr := util.NewValidationErrorForField(fieldName, message)
	WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
}

// HandleValidationErrors handles multiple validation errors
func HandleValidationErrors(w http.ResponseWriter, r *http.Request, err error) {
	// validator.ValidationErrors is a slice
	var vErrs validator.ValidationErrors
	if errors.As(err, &vErrs) {
		for _, fe := range vErrs {
			field := fe.Field()
			message := validationMessage(fe)
			HandleValidationError(w, r, field, message)
			return
		}
	}

	// fallback: generic
	HandleValidationError(w, r, "request", "invalid request payload")
}
func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	// validation tag for api token name
	case "validate-api-token-name":
		return fmt.Sprintf(
			"%s must start and end with a lowercase letter or digit; may only contain lowercase letters, digits, '_' or '-' (no spaces or commas)",
			fe.Field(),
		)
		// validation tag for sso config name
	case "validate-sso-config-name":
		return fmt.Sprintf(
			"%s must be one of [google, github, gitlab, microsoft, ldap, oidc, openshift]",
			fe.Field(),
		)
	// if a certain validator tag is not included in switch case then,
	// we will parse the error as generic validator error,
	// and further divide them on basis of parametric and non-parametric validation tags
	default:
		if fe.Param() != "" {
			// generic parametric fallback (e.g., min=3, max=50)
			return fmt.Sprintf("%s failed validation rule '%s=%s'", fe.Field(), fe.Tag(), fe.Param())
		}
		// generic non-parametric fallback (e.g., required, email, uuid)
		return fmt.Sprintf("%s failed validation rule '%s'", fe.Field(), fe.Tag())
	}
}
