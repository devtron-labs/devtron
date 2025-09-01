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

// SpecEnhancer enhances OpenAPI specs with realistic examples
type SpecEnhancer struct {
	serverURL  string
	httpClient *http.Client
	authToken  string
	logger     *zap.SugaredLogger
}

// ExampleData represents example data for an endpoint
type ExampleData struct {
	Request  interface{} `json:"request,omitempty"`
	Response interface{} `json:"response,omitempty"`
	Status   int         `json:"status"`
}

// NewSpecEnhancer creates a new spec enhancer
func NewSpecEnhancer(serverURL string, logger *zap.SugaredLogger) *SpecEnhancer {
	return &SpecEnhancer{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// EnhanceSpecs enhances all specs in the given directory with examples
func (se *SpecEnhancer) EnhanceSpecs(specsDir, outputDir string) error {
	se.logger.Infow("Enhancing specs with examples", "specsDir", specsDir, "outputDir", outputDir)

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Walk through all spec files
	return filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		se.logger.Infow("Enhancing spec file", "file", path)
		return se.enhanceSpecFile(path, outputDir)
	})
}

// enhanceSpecFile enhances a single spec file with examples
func (se *SpecEnhancer) enhanceSpecFile(specPath, outputDir string) error {
	// Load the spec
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to load spec %s: %w", specPath, err)
	}

	// Enhance the spec with examples
	se.enhanceSpec(spec)

	// Determine output path
	relPath, err := filepath.Rel(specPath, specPath)
	if err != nil {
		relPath = filepath.Base(specPath)
	}
	outputPath := filepath.Join(outputDir, relPath)

	// Create output directory if needed
	outputDirPath := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDirPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write enhanced spec
	if err := se.writeEnhancedSpec(spec, outputPath); err != nil {
		return fmt.Errorf("failed to write enhanced spec: %w", err)
	}

	se.logger.Infow("Enhanced spec written", "output", outputPath)
	return nil
}

// enhanceSpec enhances a spec with examples
func (se *SpecEnhancer) enhanceSpec(spec *openapi3.T) {
	if spec.Paths == nil {
		return
	}

	for path, pathItem := range spec.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			se.enhanceOperation(operation, path, method)
		}
	}
}

// enhanceOperation enhances an operation with examples
func (se *SpecEnhancer) enhanceOperation(operation *openapi3.Operation, path, method string) {
	if operation == nil {
		return
	}

	// Generate example data
	exampleData := se.generateExampleData(path, method, operation)

	// Add request body example
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		if exampleData.Request != nil {
			se.addRequestBodyExample(operation, exampleData.Request)
		}
	}

	// Add response examples
	if operation.Responses != nil {
		responsesMap := operation.Responses.Map()
		for statusCode, response := range responsesMap {
			if response.Value != nil && response.Value.Content != nil {
				if jsonContent, exists := response.Value.Content["application/json"]; exists {
					se.addResponseExample(jsonContent, statusCode, exampleData)
				}
			}
		}
	}

	// Add parameter examples
	if operation.Parameters != nil {
		for _, param := range operation.Parameters {
			if param.Value != nil {
				se.addParameterExample(param.Value)
			}
		}
	}
}

// generateExampleData generates example data for an endpoint
func (se *SpecEnhancer) generateExampleData(path, method string, operation *openapi3.Operation) *ExampleData {
	exampleData := &ExampleData{}

	// Try to get real data from the server
	if realData := se.getRealExampleData(path, method, operation); realData != nil {
		return realData
	}

	// Generate synthetic examples if real data is not available
	exampleData.Request = se.generateSyntheticRequest(operation)
	exampleData.Response = se.generateSyntheticResponse(operation)
	exampleData.Status = 200

	return exampleData
}

// getRealExampleData attempts to get real example data from the server
func (se *SpecEnhancer) getRealExampleData(path, method string, operation *openapi3.Operation) *ExampleData {
	// Skip if server is not available
	if se.serverURL == "" {
		return nil
	}

	url := se.serverURL + path
	var req *http.Request
	var err error

	// Create request with appropriate body for non-GET methods
	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		// Generate a minimal request body for other methods
		requestBody := se.generateMinimalRequestBody(operation)
		bodyBytes, _ := json.Marshal(requestBody)
		req, err = http.NewRequest(method, url, strings.NewReader(string(bodyBytes)))
	}

	if err != nil {
		se.logger.Debugw("Failed to create request for example data", "error", err)
		return nil
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	if se.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+se.authToken)
	}

	// Add query parameters
	if operation.Parameters != nil {
		for _, param := range operation.Parameters {
			if param.Value != nil && param.Value.In == "query" {
				sampleValue := se.generateSampleValue(param.Value.Schema)
				if strValue, ok := sampleValue.(string); ok {
					req.URL.Query().Set(param.Value.Name, strValue)
				}
			}
		}
	}

	// Make the request
	resp, err := se.httpClient.Do(req)
	if err != nil {
		se.logger.Debugw("Failed to get example data from server", "error", err)
		return nil
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		se.logger.Debugw("Failed to read response body", "error", err)
		return nil
	}

	// Parse response
	var responseData interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		se.logger.Debugw("Failed to parse response as JSON", "error", err)
		return nil
	}

	return &ExampleData{
		Response: responseData,
		Status:   resp.StatusCode,
	}
}

// generateSyntheticRequest generates a synthetic request example
func (se *SpecEnhancer) generateSyntheticRequest(operation *openapi3.Operation) interface{} {
	if operation.RequestBody == nil || operation.RequestBody.Value == nil {
		return nil
	}

	// This is a simplified version - in practice, you'd generate based on the schema
	return map[string]interface{}{
		"example_field": "example_value",
		"example_id":    123,
		"example_flag":  true,
	}
}

// generateSyntheticResponse generates a synthetic response example
func (se *SpecEnhancer) generateSyntheticResponse(operation *openapi3.Operation) interface{} {
	// This is a simplified version - in practice, you'd generate based on the schema
	return map[string]interface{}{
		"id":      1,
		"name":    "Example Resource",
		"status":  "active",
		"created": "2024-01-15T10:30:00Z",
		"updated": "2024-01-15T10:30:00Z",
		"metadata": map[string]interface{}{
			"version": "1.0.0",
			"tags":    []string{"example", "test"},
		},
	}
}

// generateMinimalRequestBody generates a minimal request body for testing
func (se *SpecEnhancer) generateMinimalRequestBody(operation *openapi3.Operation) interface{} {
	if operation.RequestBody == nil || operation.RequestBody.Value == nil {
		return map[string]interface{}{}
	}

	// Generate minimal required fields
	body := make(map[string]interface{})

	// This is a simplified version - in practice, you'd analyze the schema
	// and generate appropriate minimal data
	body["name"] = "test"
	body["description"] = "test description"

	return body
}

// addRequestBodyExample adds an example to the request body
func (se *SpecEnhancer) addRequestBodyExample(operation *openapi3.Operation, example interface{}) {
	if operation.RequestBody.Value.Content == nil {
		operation.RequestBody.Value.Content = make(map[string]*openapi3.MediaType)
	}

	if jsonContent, exists := operation.RequestBody.Value.Content["application/json"]; exists {
		if jsonContent.Example == nil {
			jsonContent.Example = example
		}
	} else {
		operation.RequestBody.Value.Content["application/json"] = &openapi3.MediaType{
			Example: example,
		}
	}
}

// addResponseExample adds an example to a response
func (se *SpecEnhancer) addResponseExample(content *openapi3.MediaType, statusCode string, exampleData *ExampleData) {
	if content.Example == nil {
		content.Example = exampleData.Response
	}

	// Add multiple examples if they don't exist
	if content.Examples == nil {
		content.Examples = make(map[string]*openapi3.ExampleRef)
	}

	// Add success example
	if _, exists := content.Examples["success"]; !exists {
		content.Examples["success"] = &openapi3.ExampleRef{
			Value: &openapi3.Example{
				Summary: "Successful response",
				Value:   exampleData.Response,
			},
		}
	}

	// Add error example
	if _, exists := content.Examples["error"]; !exists {
		content.Examples["error"] = &openapi3.ExampleRef{
			Value: &openapi3.Example{
				Summary: "Error response",
				Value: map[string]interface{}{
					"error":   "Bad Request",
					"message": "Invalid request parameters",
					"code":    400,
				},
			},
		}
	}
}

// addParameterExample adds an example to a parameter
func (se *SpecEnhancer) addParameterExample(param *openapi3.Parameter) {
	if param.Example == nil {
		param.Example = se.generateSampleValue(param.Schema)
	}
}

// generateSampleValue generates a sample value for a parameter
func (se *SpecEnhancer) generateSampleValue(schema *openapi3.SchemaRef) interface{} {
	if schema == nil || schema.Value == nil {
		return "sample_value"
	}

	// This is a simplified version - in practice, you'd check the schema type
	// and generate appropriate sample values
	return "sample_value"
}

// writeEnhancedSpec writes the enhanced spec to a file
func (se *SpecEnhancer) writeEnhancedSpec(spec *openapi3.T, outputPath string) error {
	// Convert spec to YAML
	data, err := spec.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal spec: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write spec file: %w", err)
	}

	return nil
}

// SetAuthToken sets the authentication token for requests
func (se *SpecEnhancer) SetAuthToken(token string) {
	se.authToken = token
}
