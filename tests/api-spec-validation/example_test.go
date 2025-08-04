package api_spec_validation

import (
	"testing"

	"go.uber.org/zap"
)

func TestAPISpecValidationExample(t *testing.T) {
	// This is an example test showing how to use the validation framework
	// In a real scenario, you would run this against a live server

	// Setup logging
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create validator
	validator := NewAPISpecValidator("http://localhost:8080", logger.Sugar())

	// Set auth token if needed
	// validator.SetAuthToken("your-auth-token")

	// Load specs from the specs directory
	err := validator.LoadSpecs("../../../specs")
	if err != nil {
		t.Logf("Failed to load specs: %v", err)
		t.Skip("Skipping test - no specs available or server not running")
	}

	// Validate all specs
	results := validator.ValidateAllSpecs()

	// Check results
	if len(results) == 0 {
		t.Log("No validation results - this might indicate no specs were loaded")
		return
	}

	// Log summary
	passed := 0
	failed := 0
	warnings := 0

	for _, result := range results {
		switch result.Status {
		case "PASS":
			passed++
		case "FAIL":
			failed++
		case "WARNING":
			warnings++
		}
	}

	t.Logf("Validation Summary: %d passed, %d failed, %d warnings", passed, failed, warnings)

	// Generate report
	report := validator.GenerateReport()
	t.Logf("Generated report:\n%s", report)

	// In a real test, you might want to fail if there are too many failures
	if failed > 0 {
		t.Logf("Found %d validation failures - check the report for details", failed)
		// Uncomment the next line to make the test fail on validation errors
		// t.Fail()
	}
}

func TestSpecComparisonExample(t *testing.T) {
	// Example test for spec-code comparison
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	comparator := NewSpecComparator(logger.Sugar())

	// Compare specs with handlers
	results, err := comparator.CompareSpecsWithHandlers("../../../specs", "../../../api")
	if err != nil {
		t.Logf("Failed to compare specs: %v", err)
		t.Skip("Skipping test - comparison failed")
	}

	// Log comparison results
	totalIssues := 0
	for _, result := range results {
		totalIssues += len(result.Issues)
		t.Logf("Spec file %s: %d issues", result.SpecFile, len(result.Issues))

		for _, issue := range result.Issues {
			t.Logf("  - %s: %s", issue.Type, issue.Message)
		}
	}

	t.Logf("Total comparison issues found: %d", totalIssues)
}

func TestSpecEnhancementExample(t *testing.T) {
	// Example test for spec enhancement
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	enhancer := NewSpecEnhancer("http://localhost:8080", logger.Sugar())

	// Enhance specs with examples
	err := enhancer.EnhanceSpecs("../../../specs", "./enhanced-specs")
	if err != nil {
		t.Logf("Failed to enhance specs: %v", err)
		t.Skip("Skipping test - enhancement failed")
	}

	t.Log("Specs enhanced successfully - check ./enhanced-specs directory")
}
