package api_spec_validation

import (
	"net/http"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"go.uber.org/zap"
)

func TestAuthenticationTokenInCookies(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	validator := NewAPISpecValidator("http://localhost:8080", logger.Sugar())

	// Set auth token
	authToken := "test-auth-token"
	validator.SetToken(authToken)

	// Create a test request
	req, err := http.NewRequest("GET", "http://localhost:8080/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Apply authentication
	validator.setAuthentication(req)

	// Check that no Authorization header is set (cookie-only authentication)
	authHeader := req.Header.Get("Authorization")
	if authHeader != "" {
		t.Errorf("Expected no Authorization header in cookie-only mode, got '%s'", authHeader)
	}

	cookieHeader := req.Header.Get("Cookie")
	if cookieHeader == "" {
		t.Error("Expected Cookie header to be set, but it was empty")
	}

	// Check that auth token is in cookies
	if !strings.Contains(cookieHeader, "token="+authToken) {
		t.Errorf("Expected cookie to contain 'token=%s', got '%s'", authToken, cookieHeader)
	}
}

func TestSetCookiesFromString(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	validator := NewAPISpecValidator("http://localhost:8080", logger.Sugar())

	// Test cookie string similar to the curl request
	cookieString := "_ga=GA1.1.654831891.1739442610; _ga_5WWMF8TQVE=GS1.1.1742452726.1.1.1742452747.0.0.0; session_id=abc123"
	validator.SetCookiesFromString(cookieString)
	// Create a test request
	req, err := http.NewRequest("GET", "http://localhost:8080/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Apply cookies directly to test cookie parsing
	validator.BuildAndSetCookies(req)

	cookieHeader := req.Header.Get("Cookie")
	if cookieHeader == "" {
		t.Error("Expected Cookie header to be set, but it was empty")
	}

	// Check that parsed cookies are included
	if !strings.Contains(cookieHeader, "_ga=GA1.1.654831891.1739442610") {
		t.Errorf("Expected cookie to contain '_ga=GA1.1.654831891.1739442610', got '%s'", cookieHeader)
	}

	if !strings.Contains(cookieHeader, "session_id=abc123") {
		t.Errorf("Expected cookie to contain 'session_id=abc123', got '%s'", cookieHeader)
	}
}

func TestCookieOnlyAuthentication(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	validator := NewAPISpecValidator("http://localhost:8080", logger.Sugar())

	authToken := "test-auth-token"
	validator.SetToken(authToken)

	// Test that only cookies are set, no Authorization header
	req, _ := http.NewRequest("GET", "http://localhost:8080/test", nil)
	validator.setAuthentication(req)

	authHeader := req.Header.Get("Authorization")
	if authHeader != "" {
		t.Errorf("Expected no Authorization header in cookie-only mode, got '%s'", authHeader)
	}

	cookieHeader := req.Header.Get("Cookie")
	if !strings.Contains(cookieHeader, "token="+authToken) {
		t.Errorf("Expected cookie to contain 'token=%s', got '%s'", authToken, cookieHeader)
	}
}

func TestPathParameterReplacement(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	validator := NewAPISpecValidator("http://localhost:8080", logger.Sugar())

	// Test integer path parameter
	integerTypes := openapi3.Types{openapi3.TypeInteger}
	integerParam := &openapi3.Parameter{
		Name: "pipelineId",
		In:   "path",
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:    &integerTypes,
				Example: 456,
			},
		},
	}

	// Test string path parameter with pattern
	stringTypes := openapi3.Types{openapi3.TypeString}
	stringParam := &openapi3.Parameter{
		Name: "gitHash",
		In:   "path",
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:    &stringTypes,
				Pattern: "^[a-f0-9]{7,40}$",
			},
		},
	}

	operation := &openapi3.Operation{
		Parameters: []*openapi3.ParameterRef{
			{Value: integerParam},
			{Value: stringParam},
		},
	}

	// Test path processing
	originalPath := "/app/workflow/trigger/{pipelineId}/commit/{gitHash}"
	processedPath, err := validator.processPathParameters(originalPath, operation)
	if err != nil {
		t.Fatalf("Failed to process path parameters: %v", err)
	}

	// Check that placeholders were replaced
	if strings.Contains(processedPath, "{pipelineId}") {
		t.Error("pipelineId placeholder was not replaced")
	}
	if strings.Contains(processedPath, "{gitHash}") {
		t.Error("gitHash placeholder was not replaced")
	}

	// Check that the path contains expected values
	if !strings.Contains(processedPath, "456") { // from example
		t.Error("Expected pipelineId value '456' not found in processed path")
	}
	if !strings.Contains(processedPath, "a1b2c3d4e5f6789") { // from pattern
		t.Error("Expected gitHash pattern value not found in processed path")
	}

	t.Logf("Original path: %s", originalPath)
	t.Logf("Processed path: %s", processedPath)
}

func TestGenerateSampleValue(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	validator := NewAPISpecValidator("http://localhost:8080", logger.Sugar())

	// Test integer with example
	intTypes := openapi3.Types{openapi3.TypeInteger}
	intSchema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:    &intTypes,
			Example: 123,
		},
	}
	intValue := validator.generateSampleValue(intSchema)
	if intValue != "123" {
		t.Errorf("Expected '123', got '%s'", intValue)
	}

	// Test string with pattern
	stringTypes2 := openapi3.Types{openapi3.TypeString}
	stringSchema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:    &stringTypes2,
			Pattern: "^[a-f0-9]{7,40}$",
		},
	}
	stringValue := validator.generateSampleValue(stringSchema)
	if stringValue != "a1b2c3d4e5f6789" {
		t.Errorf("Expected 'a1b2c3d4e5f6789', got '%s'", stringValue)
	}

	// Test boolean
	boolTypes := openapi3.Types{openapi3.TypeBoolean}
	boolSchema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &boolTypes,
		},
	}
	boolValue := validator.generateSampleValue(boolSchema)
	if boolValue != "true" {
		t.Errorf("Expected 'true', got '%s'", boolValue)
	}
}
