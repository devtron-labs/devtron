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

		// Skip directories and files that are not OpenAPI specs
		if info.IsDir() || (!strings.HasSuffix(info.Name(), ".yaml") && !strings.HasSuffix(info.Name(), ".yml") && !strings.HasSuffix(info.Name(), ".json")) {
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

	// Create request with proper body
	var req *http.Request
	var requestPayload string

	if method == "GET" || method == "DELETE" {
		req, err = http.NewRequest(method, url, nil)
		requestPayload = "" // No body for GET/DELETE
	} else {
		// Generate appropriate request body based on OpenAPI schema
		requestBody, err := v.generateRequestBody(operation, path, method)
		if err != nil {
			v.logger.Warnw("Failed to generate request body, using empty object", "error", err)
			requestPayload = "{}"
			req, err = http.NewRequest(method, url, strings.NewReader("{}"))
		} else if requestBody != nil {
			// Read the request body to capture it for reporting
			bodyBytes, readErr := io.ReadAll(requestBody)
			if readErr != nil {
				v.logger.Warnw("Failed to read request body for reporting", "error", readErr)
				requestPayload = "{unknown}"
			} else {
				requestPayload = string(bodyBytes)
			}
			// Create a new reader for the actual request
			req, err = http.NewRequest(method, url, strings.NewReader(requestPayload))
		} else {
			requestPayload = ""
			req, err = http.NewRequest(method, url, nil)
		}
	}

	// Store the request payload in the result
	result.Request = requestPayload

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

		// Add request payload information
		if result.Request != nil && result.Request != "" {
			requestStr, ok := result.Request.(string)
			if ok && requestStr != "" {
				// Pretty print JSON if possible
				var requestData interface{}
				if err := json.Unmarshal([]byte(requestStr), &requestData); err == nil {
					if prettyJSON, err := json.MarshalIndent(requestData, "", "  "); err == nil {
						report.WriteString(fmt.Sprintf("- **Request Payload**:\n```json\n%s\n```\n", string(prettyJSON)))
					} else {
						report.WriteString(fmt.Sprintf("- **Request Payload**: %s\n", requestStr))
					}
				} else {
					// Not JSON, display as-is
					report.WriteString(fmt.Sprintf("- **Request Payload**: %s\n", requestStr))
				}
			} else {
				report.WriteString("- **Request Payload**: (empty)\n")
			}
		} else {
			report.WriteString("- **Request Payload**: (none)\n")
		}

		if result.StatusCode != 0 {
			report.WriteString(fmt.Sprintf("- **Response Code**: %d\n", result.StatusCode))

			// Add error message if response contains error information
			if result.Response != nil {
				responseStr, ok := result.Response.(string)
				if ok && responseStr != "" {
					// Try to extract error message from JSON response
					var responseData map[string]interface{}
					if err := json.Unmarshal([]byte(responseStr), &responseData); err == nil {
						// Look for common error message fields
						if msg, exists := responseData["message"]; exists {
							report.WriteString(fmt.Sprintf("- **Error/Msg**: %v\n", msg))
						} else if msg, exists := responseData["error"]; exists {
							report.WriteString(fmt.Sprintf("- **Error/Msg**: %v\n", msg))
						} else if msg, exists := responseData["msg"]; exists {
							report.WriteString(fmt.Sprintf("- **Error/Msg**: %v\n", msg))
						} else {
							// If no structured error found, show the raw response (truncated if too long)
							if len(responseStr) > 200 {
								report.WriteString(fmt.Sprintf("- **Error/Msg**: %s...\n", responseStr[:200]))
							} else {
								report.WriteString(fmt.Sprintf("- **Error/Msg**: %s\n", responseStr))
							}
						}
					} else {
						// If not JSON, show the raw response (truncated if too long)
						if len(responseStr) > 200 {
							report.WriteString(fmt.Sprintf("- **Error/Msg**: %s...\n", responseStr[:200]))
						} else {
							report.WriteString(fmt.Sprintf("- **Error/Msg**: %s\n", responseStr))
						}
					}
				} else {
					report.WriteString("- **Error/Msg**: {}\n")
				}
			} else {
				report.WriteString("- **Error/Msg**: {}\n")
			}
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

// generateRequestBody generates a proper request body based on OpenAPI operation schema
func (v *APISpecValidator) generateRequestBody(operation *openapi3.Operation, path, method string) (io.Reader, error) {
	if operation.RequestBody == nil || operation.RequestBody.Value == nil {
		return nil, nil
	}

	// First try specialized payloads for known problematic endpoints
	if specialPayload := v.generateSpecializedPayload(path, method); specialPayload != nil {
		jsonBytes, err := json.MarshalIndent(specialPayload, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal specialized payload: %w", err)
		}
		v.logger.Debugw("Generated specialized request body", "path", path, "method", method, "payload", string(jsonBytes))
		return strings.NewReader(string(jsonBytes)), nil
	}

	// Look for JSON content type
	content := operation.RequestBody.Value.Content
	if content == nil {
		return nil, nil
	}

	var schema *openapi3.SchemaRef

	// Try different content types
	if jsonContent, exists := content["application/json"]; exists && jsonContent.Schema != nil {
		schema = jsonContent.Schema
	} else if formContent, exists := content["multipart/form-data"]; exists && formContent.Schema != nil {
		// For multipart/form-data, we'll skip for now as it requires special handling
		return strings.NewReader("{}"), nil
	} else {
		// No suitable content type found
		return strings.NewReader("{}"), nil
	}

	// Generate JSON payload from schema
	payload := v.generateJSONFromSchema(schema)
	if payload == nil {
		return strings.NewReader("{}"), nil
	}

	jsonBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal generated payload: %w", err)
	}

	v.logger.Debugw("Generated request body", "payload", string(jsonBytes))
	return strings.NewReader(string(jsonBytes)), nil
}

// generateJSONFromSchema generates a JSON object/value from an OpenAPI schema
func (v *APISpecValidator) generateJSONFromSchema(schema *openapi3.SchemaRef) interface{} {
	if schema == nil || schema.Value == nil {
		return nil
	}

	schemaValue := schema.Value

	// Use example if provided
	if schemaValue.Example != nil {
		return schemaValue.Example
	}

	// Handle different schema types
	if schemaValue.Type != nil {
		switch {
		case schemaValue.Type.Is(openapi3.TypeObject):
			return v.generateObjectFromSchema(schemaValue)
		case schemaValue.Type.Is(openapi3.TypeArray):
			return v.generateArrayFromSchema(schemaValue)
		case schemaValue.Type.Is(openapi3.TypeString):
			return v.generateStringFromSchema(schemaValue)
		case schemaValue.Type.Is(openapi3.TypeInteger):
			return v.generateIntegerFromSchema(schemaValue)
		case schemaValue.Type.Is(openapi3.TypeNumber):
			return v.generateNumberFromSchema(schemaValue)
		case schemaValue.Type.Is(openapi3.TypeBoolean):
			return true
		}
	}

	// Handle references to other schemas
	if schemaValue.AllOf != nil && len(schemaValue.AllOf) > 0 {
		// For allOf, merge all schemas (simplified approach)
		result := make(map[string]interface{})
		for _, subSchema := range schemaValue.AllOf {
			if subObj := v.generateJSONFromSchema(subSchema); subObj != nil {
				if objMap, ok := subObj.(map[string]interface{}); ok {
					for k, v := range objMap {
						result[k] = v
					}
				}
			}
		}
		return result
	}

	// Default fallback
	return "sample_value"
}

// generateObjectFromSchema generates a JSON object from an OpenAPI object schema
func (v *APISpecValidator) generateObjectFromSchema(schema *openapi3.Schema) map[string]interface{} {
	result := make(map[string]interface{})

	// Handle required fields first
	for _, requiredField := range schema.Required {
		if propSchema, exists := schema.Properties[requiredField]; exists {
			result[requiredField] = v.generateJSONFromSchema(propSchema)
		} else {
			// If required field doesn't have a schema, provide a default
			result[requiredField] = v.getDefaultValueForRequiredField(requiredField)
		}
	}

	// Add some optional fields for completeness (but limit to avoid huge payloads)
	optionalCount := 0
	maxOptional := 3
	for propName, propSchema := range schema.Properties {
		if _, exists := result[propName]; !exists && optionalCount < maxOptional {
			result[propName] = v.generateJSONFromSchema(propSchema)
			optionalCount++
		}
	}

	return result
}

// generateArrayFromSchema generates a JSON array from an OpenAPI array schema
func (v *APISpecValidator) generateArrayFromSchema(schema *openapi3.Schema) []interface{} {
	if schema.Items == nil {
		return []interface{}{}
	}

	// Generate 1-2 items for the array
	result := make([]interface{}, 0, 2)

	// Generate first item
	item := v.generateJSONFromSchema(schema.Items)
	if item != nil {
		result = append(result, item)
	}

	// For some cases, add a second item with slight variation
	if len(result) > 0 {
		if itemMap, ok := item.(map[string]interface{}); ok {
			// Create a variant of the first item
			secondItem := make(map[string]interface{})
			for k, v := range itemMap {
				secondItem[k] = v
			}
			// Modify some fields to create variation
			if _, exists := secondItem["id"]; exists {
				secondItem["id"] = 2
			}
			if _, exists := secondItem["name"]; exists {
				secondItem["name"] = "sample-item-2"
			}
			result = append(result, secondItem)
		}
	}

	return result
}

// generateStringFromSchema generates a string value from an OpenAPI string schema
func (v *APISpecValidator) generateStringFromSchema(schema *openapi3.Schema) string {
	// Use example if provided
	if schema.Example != nil {
		if str, ok := schema.Example.(string); ok {
			return str
		}
	}

	// Handle specific string formats
	switch schema.Format {
	case "uuid":
		return "123e4567-e89b-12d3-a456-426614174000"
	case "email":
		return "test@example.com"
	case "date":
		return "2023-01-01"
	case "date-time":
		return "2023-01-01T00:00:00Z"
	case "uri":
		return "https://example.com"
	case "password":
		return "password123"
	default:
		// Handle enum values
		if len(schema.Enum) > 0 {
			if str, ok := schema.Enum[0].(string); ok {
				return str
			}
		}

		// Check for pattern-based strings
		if schema.Pattern != "" {
			return v.generatePatternValue(schema.Pattern)
		}

		// Default string value
		return "sample_string"
	}
}

// generateIntegerFromSchema generates an integer value from an OpenAPI integer schema
func (v *APISpecValidator) generateIntegerFromSchema(schema *openapi3.Schema) int64 {
	// Use example if provided
	if schema.Example != nil {
		if num, ok := schema.Example.(float64); ok {
			return int64(num)
		}
		if num, ok := schema.Example.(int64); ok {
			return num
		}
		if num, ok := schema.Example.(int); ok {
			return int64(num)
		}
	}

	// Use minimum if specified and >= 1
	if schema.Min != nil && *schema.Min >= 1 {
		return int64(*schema.Min)
	}

	// Use maximum if specified and reasonable
	if schema.Max != nil && *schema.Max > 0 && *schema.Max <= 1000 {
		return int64(*schema.Max)
	}

	// Default to 1
	return 1
}

// generateNumberFromSchema generates a number value from an OpenAPI number schema
func (v *APISpecValidator) generateNumberFromSchema(schema *openapi3.Schema) float64 {
	// Use example if provided
	if schema.Example != nil {
		if num, ok := schema.Example.(float64); ok {
			return num
		}
		if num, ok := schema.Example.(int64); ok {
			return float64(num)
		}
		if num, ok := schema.Example.(int); ok {
			return float64(num)
		}
	}

	// Use minimum if specified
	if schema.Min != nil {
		return *schema.Min
	}

	// Default to 1.0
	return 1.0
}

// getDefaultValueForRequiredField provides sensible defaults for common required field names
func (v *APISpecValidator) getDefaultValueForRequiredField(fieldName string) interface{} {
	switch strings.ToLower(fieldName) {
	case "id":
		return 1
	case "name":
		return "sample-name"
	case "email", "emailid":
		return "test@example.com"
	case "url":
		return "https://example.com"
	case "description":
		return "Sample description"
	case "type":
		return "sample-type"
	case "provider":
		return "GITHUB"
	case "authmode":
		return "ANONYMOUS"
	case "active":
		return true
	case "monitoringtoolid":
		return 1
	case "clusterid":
		return 1
	case "appid":
		return 1
	case "envid":
		return 1
	case "pipelineid":
		return 1
	case "action":
		return "CREATE"
	case "username":
		return "sample-user"
	case "password":
		return "password123"
	case "token":
		return "sample-token"
	case "host":
		return "https://github.com"
	case "port":
		return "443"
	case "region":
		return "us-east-1"
	case "accesskey":
		return "sample-access-key"
	case "secretkey":
		return "sample-secret-key"
	case "fromemail":
		return "noreply@example.com"
	case "webhookurl":
		return "https://hooks.slack.com/services/sample"
	case "configname":
		return "sample-config"
	case "identifier":
		return "1"
	case "clustername":
		return "sample-cluster"
	case "expireatinms":
		return 1735689600000 // Future timestamp
	default:
		return "sample-value"
	}
}

// generateSpecializedPayload creates specialized payloads for known problematic endpoints
func (v *APISpecValidator) generateSpecializedPayload(path, method string) interface{} {
	// External Links endpoints
	if strings.Contains(path, "/external-links") && method == "POST" {
		return []map[string]interface{}{
			{
				"id":               0,
				"monitoringToolId": 1,
				"name":             "sample-external-link",
				"url":              "https://grafana.example.com",
				"type":             "appLevel",
				"identifiers": []map[string]interface{}{
					{
						"type":       "devtron-app",
						"identifier": "1",
						"clusterId":  1,
					},
				},
				"description": "Sample external link for testing",
				"isEditable":  true,
			},
		}
	}

	// External Links PUT endpoint
	if strings.Contains(path, "/external-links") && method == "PUT" {
		return map[string]interface{}{
			"id":               1,
			"monitoringToolId": 1,
			"name":             "updated-external-link",
			"url":              "https://grafana-updated.example.com",
			"type":             "appLevel",
			"identifiers": []map[string]interface{}{
				{
					"type":       "devtron-app",
					"identifier": "1",
					"clusterId":  1,
				},
			},
			"description": "Updated external link for testing",
			"isEditable":  true,
		}
	}

	// User management endpoints
	if strings.Contains(path, "/user") && (method == "POST" || method == "PUT") {
		return map[string]interface{}{
			"emailId":     "test@example.com",
			"groups":      []interface{}{},
			"roleFilters": []interface{}{},
			"superAdmin":  false,
		}
	}

	// API Token endpoints
	if strings.Contains(path, "/api-token") && method == "POST" {
		return map[string]interface{}{
			"name":         "sample-api-token",
			"description":  "Sample API token for testing",
			"expireAtInMs": 1735689600000,
		}
	}

	if strings.Contains(path, "/api-token") && method == "PUT" {
		return map[string]interface{}{
			"id":          1,
			"name":        "updated-api-token",
			"description": "Updated API token for testing",
		}
	}

	// Chart Repository endpoints
	if strings.Contains(path, "/chart-repo") && (method == "POST" || method == "PUT") {
		return map[string]interface{}{
			"name":                      "sample-chart-repo",
			"url":                       "https://charts.example.com",
			"authMode":                  "ANONYMOUS",
			"active":                    true,
			"default":                   false,
			"allow_insecure_connection": false,
		}
	}

	// GitOps Config endpoints
	if strings.Contains(path, "/gitops/config") && method == "POST" {
		return map[string]interface{}{
			"provider":       "GITHUB",
			"username":       "sample-user",
			"token":          "sample-token",
			"gitLabGroupId":  "",
			"gitHubOrgId":    "sample-org",
			"host":           "https://github.com",
			"active":         true,
			"azureProjectId": "",
		}
	}

	// Deployment App Type Change
	if strings.Contains(path, "/deployment") && strings.Contains(path, "/patch") {
		return map[string]interface{}{
			"appId":            1,
			"envId":            1,
			"targetChartRefId": 1,
		}
	}

	// Server Action endpoints
	if strings.Contains(path, "/server") && method == "POST" {
		return map[string]interface{}{
			"action": "RESTART",
		}
	}

	// Module endpoints
	if strings.Contains(path, "/module") && method == "POST" {
		return map[string]interface{}{
			"name":   "sample-module",
			"action": "INSTALL",
		}
	}

	// Notification endpoints
	if strings.Contains(path, "/notification") && (method == "POST" || method == "PUT") {
		return map[string]interface{}{
			"notificationConfigRequest": map[string]interface{}{
				"teamId":       1,
				"appId":        1,
				"envId":        1,
				"pipelineId":   1,
				"pipelineType": "CI",
				"eventTypeIds": []int{1},
				"providers": []map[string]interface{}{
					{
						"dest":     "test@example.com",
						"configId": 1,
					},
				},
			},
		}
	}

	return nil
}
