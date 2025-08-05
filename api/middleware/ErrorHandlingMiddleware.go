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

package middleware

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

// ErrorHandlingMiddleware provides enhanced error handling and logging for REST handlers
type ErrorHandlingMiddleware struct {
	logger *zap.SugaredLogger
}

// NewErrorHandlingMiddleware creates a new error handling middleware
func NewErrorHandlingMiddleware(logger *zap.SugaredLogger) *ErrorHandlingMiddleware {
	return &ErrorHandlingMiddleware{
		logger: logger,
	}
}

// RequestContext holds information about the current request for better error handling
type RequestContext struct {
	RequestID    string
	StartTime    time.Time
	Method       string
	Path         string
	ResourceType string
	ResourceID   string
	UserID       int
}

// ContextKey is used for storing request context in the request context
type ContextKey string

const RequestContextKey ContextKey = "request_context"

// WithRequestContext middleware adds request context for better error handling
func (m *ErrorHandlingMiddleware) WithRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate request ID for correlation
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())

		// Extract resource information from path
		vars := mux.Vars(r)
		resourceType, resourceID := extractResourceFromPath(r.URL.Path, vars)

		// Create request context
		reqCtx := &RequestContext{
			RequestID:    requestID,
			StartTime:    time.Now(),
			Method:       r.Method,
			Path:         r.URL.Path,
			ResourceType: resourceType,
			ResourceID:   resourceID,
		}

		// Add to request context
		ctx := context.WithValue(r.Context(), RequestContextKey, reqCtx)
		r = r.WithContext(ctx)

		// Add request ID to response headers for debugging
		w.Header().Set("X-Request-ID", requestID)

		// Log request start
		m.logger.Infow("Request started",
			"requestId", requestID,
			"method", r.Method,
			"path", r.URL.Path,
			"resourceType", resourceType,
			"resourceId", resourceID,
		)

		next.ServeHTTP(w, r)

		// Log request completion
		duration := time.Since(reqCtx.StartTime)
		m.logger.Infow("Request completed",
			"requestId", requestID,
			"duration", duration,
			"method", r.Method,
			"path", r.URL.Path,
		)
	})
}

// extractResourceFromPath attempts to extract resource type and ID from the request path
func extractResourceFromPath(path string, vars map[string]string) (resourceType, resourceID string) {
	// Common resource type mappings based on path patterns
	resourceMappings := map[string]string{
		"/api/v1/team/":       "team",
		"/api/v1/cluster/":    "cluster",
		"/api/v1/env/":        "environment",
		"/api/v1/app/":        "application",
		"/api/v1/docker/":     "docker registry",
		"/api/v1/git/":        "git provider",
		"/api/v1/pipeline/":   "pipeline",
		"/api/v1/webhook/":    "webhook",
		"/orchestrator/team/": "team",
		"/orchestrator/app/":  "application",
		"/orchestrator/env/":  "environment",
	}

	// Try to match path patterns
	for pathPrefix, resType := range resourceMappings {
		if len(path) > len(pathPrefix) && path[:len(pathPrefix)] == pathPrefix {
			resourceType = resType
			break
		}
	}

	// Try to extract ID from common parameter names
	if id, exists := vars["id"]; exists {
		resourceID = id
	} else if id, exists := vars["teamId"]; exists {
		resourceID = id
	} else if id, exists := vars["appId"]; exists {
		resourceID = id
	} else if id, exists := vars["clusterId"]; exists {
		resourceID = id
	} else if id, exists := vars["envId"]; exists {
		resourceID = id
	} else if id, exists := vars["gitHostId"]; exists {
		resourceID = id
	}

	return resourceType, resourceID
}

// GetRequestContext retrieves the request context from the HTTP request
func GetRequestContext(r *http.Request) *RequestContext {
	if ctx := r.Context().Value(RequestContextKey); ctx != nil {
		if reqCtx, ok := ctx.(*RequestContext); ok {
			return reqCtx
		}
	}
	return nil
}

// LogError logs an error with request context for better debugging
func (m *ErrorHandlingMiddleware) LogError(r *http.Request, err error, operation string) {
	reqCtx := GetRequestContext(r)
	if reqCtx != nil {
		m.logger.Errorw("Request error",
			"requestId", reqCtx.RequestID,
			"operation", operation,
			"error", err,
			"method", reqCtx.Method,
			"path", reqCtx.Path,
			"resourceType", reqCtx.ResourceType,
			"resourceId", reqCtx.ResourceID,
		)
	} else {
		m.logger.Errorw("Request error",
			"operation", operation,
			"error", err,
			"method", r.Method,
			"path", r.URL.Path,
		)
	}
}

// ValidateIntPathParam validates and extracts an integer path parameter with enhanced error handling
func ValidateIntPathParam(r *http.Request, paramName string) (int, *util.ApiError) {
	vars := mux.Vars(r)
	paramValue := vars[paramName]

	if paramValue == "" {
		return 0, util.NewMissingRequiredFieldError(paramName)
	}

	id, err := strconv.Atoi(paramValue)
	if err != nil {
		return 0, util.NewInvalidPathParameterError(paramName, paramValue)
	}

	if id <= 0 {
		return 0, util.NewValidationErrorForField(paramName, "must be a positive integer")
	}

	return id, nil
}
