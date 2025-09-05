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
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"net/http"
	"strconv"
	"strings"
)

// extractResourceContext tries to extract resource type and ID from response body
func extractResourceContext(respBody interface{}) (resourceType, resourceId string) {
	// Try to extract from response body if it contains context
	if respBody == nil {
		return "", ""
	}

	// Check if respBody is a map with resource context
	if contextMap, ok := respBody.(map[string]interface{}); ok {
		if rt, exists := contextMap["resourceType"]; exists {
			resourceType = fmt.Sprintf("%v", rt)
		}
		if ri, exists := contextMap["resourceId"]; exists {
			resourceId = fmt.Sprintf("%v", ri)
		}
	}

	return resourceType, resourceId
}

// WriteJsonRespWithResourceContext enhances error response with resource context
// This function provides better error messages for database errors by including resource context
func WriteJsonRespWithResourceContext(w http.ResponseWriter, err error, respBody interface{},
	status int, resourceType, resourceId string) {

	if err != nil && util.IsErrNoRows(err) {
		// Override respBody with resource context for better error handling
		respBody = map[string]interface{}{
			"resourceType": resourceType,
			"resourceId":   resourceId,
		}
	}
	WriteJsonResp(w, err, respBody, status)
}

// WriteJsonRespWithResourceContextFromId is a convenience function for integer IDs
func WriteJsonRespWithResourceContextFromId(w http.ResponseWriter, err error, respBody interface{},
	status int, resourceType string, resourceId int) {
	WriteJsonRespWithResourceContext(w, err, respBody, status, resourceType, strconv.Itoa(resourceId))
}

// Convenience functions for common error scenarios (to fix build errors)
func WriteMissingRequiredFieldError(w http.ResponseWriter, fieldName string) {
	apiErr := util.NewMissingRequiredFieldError(fieldName)
	WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
}

func WriteForbiddenError(w http.ResponseWriter, operation string, resource string) {
	apiErr := util.NewApiError(http.StatusForbidden,
		fmt.Sprintf("Access denied for %s operation on %s", operation, resource),
		"forbidden").WithCode("11008")
	WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
}

func WriteValidationError(w http.ResponseWriter, fieldName string, message string) {
	apiErr := util.NewValidationErrorForField(fieldName, message)
	WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
}

func WriteSpecificErrorResponse(w http.ResponseWriter, errorCode string, message string, details []string, statusCode int) {
	apiErr := util.NewApiError(statusCode, message, fmt.Sprintf("Error: %s", errorCode)).
		WithCode(errorCode).
		WithUserDetailMessage(fmt.Sprintf("Details: %s", strings.Join(details, "; ")))
	WriteJsonResp(w, apiErr, nil, statusCode)
}

func WritePipelineNotFoundError(w http.ResponseWriter, pipelineId int) {
	apiErr := util.NewResourceNotFoundError("pipeline", strconv.Itoa(pipelineId))
	WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
}

func WriteDatabaseError(w http.ResponseWriter, operation string, err error) {
	apiErr := util.NewApiError(http.StatusInternalServerError,
		fmt.Sprintf("Database operation failed: %s", operation),
		err.Error()).WithCode("11009")
	WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
}

func WriteUnauthorizedError(w http.ResponseWriter) {
	apiErr := util.NewApiError(http.StatusUnauthorized, "Unauthorized access", "unauthorized").
		WithCode("11010")
	WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
}

func WriteInvalidAppIdError(w http.ResponseWriter, appId string) {
	apiErr := util.NewInvalidPathParameterError("appId", appId)
	WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
}
