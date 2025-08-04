package api_spec_validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"go.uber.org/zap"
)

// APISpecValidator handles validation of API specs against actual server implementation
type APISpecValidator struct {
	serverURL   string
	httpClient  *http.Client
	specs       map[string]*openapi3.T
	authToken   string
	argoCDToken string
	logger      *zap.SugaredLogger
	results     []ValidationResult
}

// ValidationResult represents the result of validating a single endpoint
type ValidationResult struct {
	Endpoint   string
	Method     string
	Status     string // PASS, FAIL, SKIP
	Issues     []ValidationIssue
	Request    interface{}
	Response   interface{}
	StatusCode int
	Duration   time.Duration
	SpecFile   string // Added for reporting
}

// ValidationIssue represents a specific validation problem
type ValidationIssue struct {
	Type     string // PARAMETER_MISMATCH, RESPONSE_MISMATCH, AUTH_ERROR, SCHEMA_MISMATCH
	Field    string
	Expected interface{}
	Actual   interface{}
	Message  string
	Severity string // ERROR, WARNING, INFO
}

// NewAPISpecValidator creates a new validator instance
func NewAPISpecValidator(serverURL string, logger *zap.SugaredLogger) *APISpecValidator {
	return &APISpecValidator{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		specs:   make(map[string]*openapi3.T),
		logger:  logger,
		results: make([]ValidationResult, 0),
	}
}

// LoadSpecs loads all OpenAPI specs from the specs directory
func (v *APISpecValidator) LoadSpecs(specsDir string) error {
	v.logger.Infow("Loading API specs from directory", "dir", specsDir)

	return filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		v.logger.Infow("Loading spec file", "file", path)

		loader := openapi3.NewLoader()
		spec, err := loader.LoadFromFile(path)
		if err != nil {
			v.logger.Errorw("Failed to load spec file", "file", path, "error", err)
			// Continue loading other specs instead of failing completely
			return nil
		}

		// Validate the spec
		if err := spec.Validate(loader.Context); err != nil {
			v.logger.Errorw("Invalid OpenAPI spec", "file", path, "error", err)
			// Continue loading other specs instead of failing completely
			return nil
		}

		v.specs[path] = spec
		endpointCount := 0
		if spec.Paths != nil {
			for range spec.Paths.Map() {
				endpointCount++
			}
		}
		v.logger.Infow("Successfully loaded spec", "file", path, "endpoints", endpointCount)

		return nil
	})
}

// ValidateAllSpecs validates all loaded specs against the server
func (v *APISpecValidator) ValidateAllSpecs() []ValidationResult {
	v.logger.Info("Starting validation of all API specs")

	for specPath, spec := range v.specs {
		v.logger.Infow("Validating spec", "file", specPath)
		v.validateSpec(spec, specPath)
	}

	return v.results
}

// validateSpec validates a single OpenAPI spec
func (v *APISpecValidator) validateSpec(spec *openapi3.T, specPath string) {
	if spec.Paths == nil {
		return
	}

	for path, pathItem := range spec.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			result := v.validateEndpoint(path, method, operation, spec)
			result.SpecFile = specPath
			v.results = append(v.results, result)
		}
	}
}

// validateEndpoint validates a single endpoint
func (v *APISpecValidator) validateEndpoint(path, method string, operation *openapi3.Operation, spec *openapi3.T) ValidationResult {
	start := time.Now()

	result := ValidationResult{
		Endpoint: path,
		Method:   method,
		Status:   "SKIP",
		Issues:   make([]ValidationIssue, 0),
	}

	v.logger.Infow("Validating endpoint", "method", method, "path", path)

	// Skip if operation is nil
	if operation == nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Type:     "SKIP",
			Message:  "Operation is nil",
			Severity: "INFO",
		})
		return result
	}

	// Test the endpoint
	if err := v.testEndpoint(&result, path, method, operation); err != nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Type:     "REQUEST_ERROR",
			Message:  err.Error(),
			Severity: "ERROR",
		})
		result.Status = "FAIL"
	}

	result.Duration = time.Since(start)

	// Determine overall status
	if len(result.Issues) == 0 {
		result.Status = "PASS"
	} else {
		hasErrors := false
		for _, issue := range result.Issues {
			if issue.Severity == "ERROR" {
				hasErrors = true
				break
			}
		}
		if hasErrors {
			result.Status = "FAIL"
		} else {
			result.Status = "WARNING"
		}
	}

	return result
}

// testEndpoint makes an actual HTTP request to test the endpoint
func (v *APISpecValidator) testEndpoint(result *ValidationResult, path, method string, operation *openapi3.Operation) error {
	// Build the full URL
	url := v.serverURL + path

	// Create request
	var req *http.Request
	var err error

	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		// For other methods, we'll need to generate appropriate request bodies
		// This is a simplified version - in practice, you'd generate proper test data
		req, err = http.NewRequest(method, url, strings.NewReader("{}"))
	}

	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	if v.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+v.authToken)
	}

	// Add ArgoCD token as cookie if available
	if v.argoCDToken != "" {
		req.Header.Set("Cookie", "argocd.token="+v.argoCDToken)
	}

	// Add query parameters if specified in the spec
	if operation.Parameters != nil {
		for _, param := range operation.Parameters {
			if param.Value != nil && param.Value.In == "query" {
				// Add sample query parameters
				req.URL.Query().Set(param.Value.Name, v.generateSampleValue(param.Value.Schema))
			}
		}
	}

	// Make the request
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	result.StatusCode = resp.StatusCode
	result.Response = string(body)

	// Validate response against spec
	v.validateResponse(result, operation, resp.StatusCode, body)

	return nil
}

// validateResponse validates the actual response against the spec
func (v *APISpecValidator) validateResponse(result *ValidationResult, operation *openapi3.Operation, statusCode int, body []byte) {
	// Check if the status code is expected
	expectedStatus := "200" // Default
	if operation.Responses != nil {
		for status := range operation.Responses.Map() {
			if status == "200" || status == "201" {
				expectedStatus = status
				break
			}
		}
	}

	if fmt.Sprintf("%d", statusCode) != expectedStatus {
		result.Issues = append(result.Issues, ValidationIssue{
			Type:     "STATUS_CODE_MISMATCH",
			Expected: expectedStatus,
			Actual:   statusCode,
			Message:  fmt.Sprintf("Expected status %s, got %d", expectedStatus, statusCode),
			Severity: "ERROR",
		})
	}

	// Validate response body structure if we have a schema
	if operation.Responses != nil {
		responsesMap := operation.Responses.Map()
		if response, exists := responsesMap[fmt.Sprintf("%d", statusCode)]; exists && response.Value != nil {
			if response.Value.Content != nil {
				if jsonContent, exists := response.Value.Content["application/json"]; exists && jsonContent.Schema != nil {
					// Parse response body as JSON
					var responseData interface{}
					if err := json.Unmarshal(body, &responseData); err != nil {
						result.Issues = append(result.Issues, ValidationIssue{
							Type:     "RESPONSE_FORMAT_ERROR",
							Message:  fmt.Sprintf("Response is not valid JSON: %v", err),
							Severity: "ERROR",
						})
					} else {
						// Here you would validate the response data against the schema
						// This is a simplified version - in practice, you'd use a JSON schema validator
						v.logger.Debugw("Response validation", "data", responseData)
					}
				}
			}
		}
	}
}

// generateSampleValue generates a sample value for a parameter
func (v *APISpecValidator) generateSampleValue(schema *openapi3.SchemaRef) string {
	if schema == nil || schema.Value == nil {
		return ""
	}

	// For now, return a simple string value
	// In a full implementation, you would properly check the schema type
	return "sample_value"
}

// SetAuthToken sets the authentication token for requests
func (v *APISpecValidator) SetAuthToken(token string) {
	v.authToken = token
}

// SetArgoCDToken sets the ArgoCD token for cookie-based authentication
func (v *APISpecValidator) SetArgoCDToken(token string) {
	v.argoCDToken = token
}

// GetResults returns all validation results
func (v *APISpecValidator) GetResults() []ValidationResult {
	return v.results
}

// GenerateReport generates a comprehensive validation report
func (v *APISpecValidator) GenerateReport() string {
	var report strings.Builder

	report.WriteString("# API Spec Validation Report\n\n")
	report.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))

	// Summary
	total := len(v.results)
	passed := 0
	failed := 0
	warnings := 0

	for _, result := range v.results {
		switch result.Status {
		case "PASS":
			passed++
		case "FAIL":
			failed++
		case "WARNING":
			warnings++
		}
	}

	report.WriteString("## Summary\n\n")
	report.WriteString(fmt.Sprintf("- Total Endpoints: %d\n", total))
	report.WriteString(fmt.Sprintf("- Passed: %d\n", passed))
	report.WriteString(fmt.Sprintf("- Failed: %d\n", failed))
	report.WriteString(fmt.Sprintf("- Warnings: %d\n", warnings))
	report.WriteString(fmt.Sprintf("- Success Rate: %.2f%%\n\n", float64(passed)/float64(total)*100))

	// Detailed results
	report.WriteString("## Detailed Results\n\n")

	for _, result := range v.results {
		statusIcon := "✅"
		if result.Status == "FAIL" {
			statusIcon = "❌"
		} else if result.Status == "WARNING" {
			statusIcon = "⚠️"
		}

		report.WriteString(fmt.Sprintf("### %s %s %s\n\n", statusIcon, result.Method, result.Endpoint))
		report.WriteString(fmt.Sprintf("- **Status**: %s\n", result.Status))
		report.WriteString(fmt.Sprintf("- **Duration**: %s\n", result.Duration))
		report.WriteString(fmt.Sprintf("- **Spec File**: %s\n", result.SpecFile))

		if result.StatusCode != 0 {
			report.WriteString(fmt.Sprintf("- **Response Code**: %d\n", result.StatusCode))
		}

		if len(result.Issues) > 0 {
			report.WriteString("\n**Issues:**\n")
			for _, issue := range result.Issues {
				report.WriteString(fmt.Sprintf("- **%s**: %s\n", issue.Type, issue.Message))
			}
		}

		report.WriteString("\n---\n\n")
	}

	return report.String()
}
