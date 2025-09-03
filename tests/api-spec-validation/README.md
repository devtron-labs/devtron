# API Spec Validation Framework

This framework provides comprehensive testing and validation of OpenAPI specifications against the actual Devtron server implementation. It helps ensure that API specs are accurate, up-to-date, and match the real server behavior.

## Features

- **Live Server Testing**: Tests API specs against a running Devtron server
- **Spec-Handler Comparison**: Compares OpenAPI specs with REST handler implementations
- **Parameter Validation**: Validates request parameters, types, and requirements
- **Response Validation**: Validates response structures and status codes
- **Comprehensive Reporting**: Generates detailed reports with issues and recommendations
- **Example Generation**: Enhances specs with realistic examples from actual server responses

## Quick Start

### Prerequisites

1. Go 1.19 or later
2. A running Devtron server (default: http://localhost:8080)
3. Access to the Devtron codebase

### Installation

```bash
# Navigate to the validation framework directory
cd tests/api-spec-validation

# Install dependencies
make deps

# Build the validator
make build
```

### Basic Usage

```bash
# Run validation against local server
make test

# Run validation against custom server
make test-server SERVER_URL=http://your-server:8080

# Compare specs with REST handlers
make compare

# Run all validations
make all
```

## Detailed Usage

### Command Line Options

The validator supports the following command line options:

```bash
./bin/validator [options]

Options:
  --server string     Server URL to test against (default "http://localhost:8080")
  --specs string      Directory containing API specs (default "../../../specs")
  --output string     Output directory for reports (default "./reports")
  --token string      Authentication token (optional)
  --verbose           Enable verbose logging
```

### Examples

```bash
# Test with authentication
./bin/validator --server=http://localhost:8080 --token=your-auth-token

# Test specific specs directory
./bin/validator --specs=/path/to/specs --output=/path/to/reports

# Verbose output for debugging
./bin/validator --verbose
```

## Framework Architecture

### Core Components

1. **APISpecValidator**: Main validation engine that tests specs against live server
2. **SpecComparator**: Compares OpenAPI specs with REST handler implementations
3. **Test Runner**: Orchestrates the validation process and generates reports

### Validation Process

1. **Spec Loading**: Loads all OpenAPI 3.0 specs from the specified directory
2. **Endpoint Discovery**: Extracts all endpoints and operations from specs
3. **Live Testing**: Makes actual HTTP requests to the running server
4. **Response Validation**: Compares actual responses with spec expectations
5. **Handler Comparison**: Analyzes REST handler implementations for consistency
6. **Report Generation**: Creates comprehensive validation reports

### Validation Types

#### 1. Parameter Validation
- Required vs optional parameters
- Parameter types and formats
- Query, path, and header parameters
- Request body validation

#### 2. Response Validation
- Status code validation
- Response body structure validation
- Content-Type validation
- Error response validation

#### 3. Handler Comparison
- Missing handler implementations
- Parameter mismatches
- Request/response body handling
- Authentication requirements

## Report Format

The framework generates detailed reports in Markdown format:

```markdown
# API Spec Validation Report

Generated: 2024-01-15T10:30:00Z

## Summary
- Total Endpoints: 150
- Passed: 120
- Failed: 20
- Warnings: 10
- Success Rate: 80.00%

## Detailed Results

### ✅ GET /app-store/discover
- **Status**: PASS
- **Duration**: 150ms
- **Spec File**: specs/app-store.yaml
- **Response Code**: 200

### ❌ POST /app-store/install
- **Status**: FAIL
- **Duration**: 200ms
- **Spec File**: specs/app-store.yaml
- **Response Code**: 400

**Issues:**
- **PARAMETER_MISMATCH**: Required parameter 'appId' missing in request
- **RESPONSE_FORMAT_ERROR**: Response is not valid JSON
```

## Configuration

### Environment Variables

```bash
# Server configuration
DEVRON_SERVER_URL=http://localhost:8080
DEVRON_AUTH_TOKEN=your-token

# Test configuration
VALIDATION_TIMEOUT=30s
MAX_CONCURRENT_REQUESTS=10
```

### Custom Test Data

Create test data files in `testdata/` directory:

```yaml
# testdata/app-store.yaml
test_apps:
  - name: "sample-app-1"
    id: 1
    version: "1.0.0"
  - name: "sample-app-2"
    id: 2
    version: "2.0.0"
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: API Spec Validation

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  validate-specs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v3
        with:
          go-version: 1.19
      
      - name: Start Devtron Server
        run: |
          # Start your Devtron server here
          docker-compose up -d
      
      - name: Run API Spec Validation
        run: |
          cd tests/api-spec-validation
          make test
      
      - name: Upload Reports
        uses: actions/upload-artifact@v3
        with:
          name: api-spec-reports
          path: tests/api-spec-validation/reports/
```

### Jenkins Pipeline Example

```groovy
pipeline {
    agent any
    
    stages {
        stage('Validate API Specs') {
            steps {
                script {
                    // Start Devtron server
                    sh 'docker-compose up -d'
                    
                    // Run validation
                    dir('tests/api-spec-validation') {
                        sh 'make test'
                    }
                    
                    // Archive reports
                    archiveArtifacts artifacts: 'tests/api-spec-validation/reports/**/*'
                }
            }
        }
    }
    
    post {
        always {
            // Cleanup
            sh 'docker-compose down'
        }
    }
}
```

## Troubleshooting

### Common Issues

1. **Server Connection Failed**
   ```
   Error: request failed: dial tcp localhost:8080: connect: connection refused
   ```
   **Solution**: Ensure Devtron server is running on the specified port

2. **Authentication Errors**
   ```
   Error: 401 Unauthorized
   ```
   **Solution**: Provide valid authentication token using `--token` flag

3. **Spec Loading Errors**
   ```
   Error: Invalid OpenAPI spec
   ```
   **Solution**: Validate your OpenAPI 3.0 specs using online validators

4. **Timeout Errors**
   ```
   Error: request failed: context deadline exceeded
   ```
   **Solution**: Increase timeout or check server performance

### Debug Mode

Enable verbose logging for detailed debugging:

```bash
./bin/validator --verbose
```

This will show:
- Detailed request/response logs
- Spec parsing information
- Handler comparison details
- Performance metrics

## Contributing

### Adding New Validation Rules

1. Extend the `ValidationIssue` struct in `framework.go`
2. Add validation logic in the appropriate validation method
3. Update the report generation to include new issue types
4. Add tests for the new validation rules

### Example: Adding Custom Validation

```go
// In framework.go
func (v *APISpecValidator) validateCustomRule(result *ValidationResult, operation *openapi3.Operation) {
    // Your custom validation logic here
    if someCondition {
        result.Issues = append(result.Issues, ValidationIssue{
            Type:     "CUSTOM_RULE_VIOLATION",
            Message:  "Custom validation failed",
            Severity: "ERROR",
        })
    }
}
```

## Best Practices

1. **Regular Validation**: Run validation tests regularly (daily/weekly)
2. **CI/CD Integration**: Include validation in your CI/CD pipeline
3. **Spec Maintenance**: Keep specs updated with code changes
4. **Test Data**: Use realistic test data for comprehensive validation
5. **Documentation**: Document any custom validation rules or configurations

## Support

For issues and questions:

1. Check the troubleshooting section above
2. Review the generated reports for specific error details
3. Enable verbose logging for debugging
4. Create an issue in the Devtron repository

## License

This framework is part of the Devtron project and follows the same license terms. 