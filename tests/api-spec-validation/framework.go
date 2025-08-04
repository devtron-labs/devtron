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
	serverURL         string
	token             string
	httpClient        *http.Client
	specs             map[string]*openapi3.T
	additionalCookies map[string]string // For storing additional cookies
	logger            *zap.SugaredLogger
	results           []ValidationResult
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
		specs:             make(map[string]*openapi3.T),
		additionalCookies: make(map[string]string),
		logger:            logger,
		results:           make([]ValidationResult, 0),
	}
}

func (v *APISpecValidator) buildCurlCommand(req *http.Request) string {
	var cmd strings.Builder

	cmd.WriteString("curl --location '")
	cmd.WriteString(req.URL.String())
	cmd.WriteString("' \\\n")

	// Add headers
	for key, values := range req.Header {
		for _, value := range values {
			cmd.WriteString(fmt.Sprintf("  --header '%s: %s' \\\n", key, value))
		}
	}

	// Add cookies
	if len(v.additionalCookies) > 0 {
		cmd.WriteString("  --cookie '")
		for name, value := range v.additionalCookies {
			cmd.WriteString(name)
			cmd.WriteString("=")
			cmd.WriteString(value)
			cmd.WriteString("; ")
		}
		cmd.WriteString("' \\\n")
	}

	return cmd.String()
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
	// Process path parameters and build the full URL
	processedPath, err := v.processPathParameters(path, operation)
	if err != nil {
		return fmt.Errorf("failed to process path parameters: %w", err)
	}
	url := v.serverURL + processedPath

	// Create request
	var req *http.Request
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

	// Add headers similar to the curl request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,hi;q=0.8")

	// Set authentication - cookie-based only as per curl request pattern
	v.setAuthentication(req)

	// Add query parameters if specified in the spec
	if operation.Parameters != nil {
		for _, param := range operation.Parameters {
			if param.Value != nil && param.Value.In == "query" {
				// Add sample query parameters
				req.URL.Query().Set(param.Value.Name, v.generateSampleValue(param.Value.Schema))
			}
		}
	}

	// print the curl command for debugging
	v.logger.Debugw("Curl command", "command", v.buildCurlCommand(req))

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

// generateSampleValue generates a sample value for a parameter based on its schema
func (v *APISpecValidator) generateSampleValue(schema *openapi3.SchemaRef) string {
	if schema == nil || schema.Value == nil {
		return "sample_value"
	}

	schemaValue := schema.Value

	// Use example if provided
	if schemaValue.Example != nil {
		return fmt.Sprintf("%v", schemaValue.Example)
	}

	// Generate based on type
	if schemaValue.Type != nil && schemaValue.Type.Is(openapi3.TypeInteger) {
		// Use minimum if specified, otherwise default to 1
		if schemaValue.Min != nil && *schemaValue.Min >= 1 {
			return fmt.Sprintf("%.0f", *schemaValue.Min)
		}
		return "1"
	} else if schemaValue.Type != nil && schemaValue.Type.Is(openapi3.TypeNumber) {
		if schemaValue.Min != nil {
			return fmt.Sprintf("%g", *schemaValue.Min)
		}
		return "1.0"
	} else if schemaValue.Type != nil && schemaValue.Type.Is(openapi3.TypeBoolean) {
		return "true"
	} else if schemaValue.Type != nil && schemaValue.Type.Is(openapi3.TypeString) {
		// Handle specific string formats
		switch schemaValue.Format {
		case "uuid":
			return "123e4567-e89b-12d3-a456-426614174000"
		case "email":
			return "test@example.com"
		case "date":
			return "2023-01-01"
		case "date-time":
			return "2023-01-01T00:00:00Z"
		default:
			// Check for pattern-based strings (like git hash)
			if schemaValue.Pattern != "" {
				return v.generatePatternValue(schemaValue.Pattern)
			}
			return "sample_string"
		}
	} else if schemaValue.Type != nil && schemaValue.Type.Is(openapi3.TypeArray) {
		return "[]"
	} else if schemaValue.Type != nil && schemaValue.Type.Is(openapi3.TypeObject) {
		return "{}"
	} else {
		return "sample_value"
	}
}

// generatePatternValue generates a sample value based on a regex pattern
func (v *APISpecValidator) generatePatternValue(pattern string) string {
	// Handle common patterns
	switch pattern {
	case "^[a-f0-9]{7,40}$": // Git hash pattern
		return "a1b2c3d4e5f6789"
	case "^[0-9]+$": // Numeric string
		return "123"
	case "^[a-zA-Z0-9]+$": // Alphanumeric
		return "abc123"
	default:
		// For unknown patterns, return a generic string
		return "sample_pattern_value"
	}
}

// processPathParameters replaces path parameters in the URL with sample values
func (v *APISpecValidator) processPathParameters(path string, operation *openapi3.Operation) (string, error) {
	if operation.Parameters == nil {
		return path, nil
	}

	processedPath := path
	for _, param := range operation.Parameters {
		if param.Value != nil && param.Value.In == "path" {
			placeholder := "{" + param.Value.Name + "}"
			sampleValue := v.generateSampleValue(param.Value.Schema)
			processedPath = strings.ReplaceAll(processedPath, placeholder, sampleValue)

			v.logger.Debugw("Replaced path parameter",
				"parameter", param.Value.Name,
				"placeholder", placeholder,
				"value", sampleValue)
		}
	}

	return processedPath, nil
}

// SetToken sets the authentication token for requests
func (v *APISpecValidator) SetToken(token string) {
	v.token = token
}

// SetAdditionalCookie sets an additional cookie for authentication
func (v *APISpecValidator) SetAdditionalCookie(name, value string) {
	v.additionalCookies[name] = value
}

// setAuthentication sets cookie-based authentication only
func (v *APISpecValidator) setAuthentication(req *http.Request) {
	// Set cookies as per the curl request pattern
	v.BuildAndSetCookies(req)
}

// BuildAndSetCookies builds and sets the Cookie header similar to the curl request
func (v *APISpecValidator) BuildAndSetCookies(req *http.Request) {
	var cookies []string

	// Add auth token as cookie by default if available
	if v.token != "" {
		cookies = append(cookies, "argocd.token="+v.token)
	}

	// Add additional cookies
	for name, value := range v.additionalCookies {
		cookies = append(cookies, name+"="+value)
	}

	// Set the Cookie header if we have any cookies
	if len(cookies) > 0 {
		req.Header.Set("Cookie", strings.Join(cookies, "; "))
	}
}

// SetCookiesFromString parses a cookie string (like from curl -b flag) and sets the cookies
// Example: "_ga=GA1.1.654831891.1739442610; _ga_5WWMF8TQVE=GS1.1.1742452726.1.1.1742452747.0.0.0"
func (v *APISpecValidator) SetCookiesFromString(cookieString string) {
	if cookieString == "" {
		return
	}

	// Split by semicolon and parse each cookie
	cookiePairs := strings.Split(cookieString, ";")
	for _, pair := range cookiePairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// Split by first equals sign
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			v.SetAdditionalCookie(name, value)
		}
	}
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
