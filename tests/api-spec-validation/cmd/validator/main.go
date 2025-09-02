package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	api_spec_validation "github.com/devtron-labs/devtron/tests/api-spec-validation"
	"go.uber.org/zap"
)

func main() {
	var (
		serverURL = flag.String("server", "http://localhost:8080", "Server URL to test against")
		specsDir  = flag.String("specs", "../../../specs", "Directory containing API specs")
		outputDir = flag.String("output", "./reports", "Output directory for reports")
		verbose   = flag.Bool("verbose", false, "Enable verbose logging")
		token     = flag.String("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTQzOTM3MzQsImp0aSI6ImRiODUxYzg5LTg0YjUtNDhiOS1hYTcyLWQ0ZTA0MjRjN2U4MSIsImlhdCI6MTc1NDMwNzMzNCwiaXNzIjoiYXJnb2NkIiwibmJmIjoxNzU0MzA3MzM0LCJzdWIiOiJhZG1pbiJ9.wYGmONdpNUEjtAxXz_mViW44Rxh0YU3dax_SEuoAH5c", "Authentication token")
	)
	flag.Parse()

	// Setup logging
	var logger *zap.SugaredLogger
	if *verbose {
		zapLogger, _ := zap.NewDevelopment()
		logger = zapLogger.Sugar()
	} else {
		zapLogger, _ := zap.NewProduction()
		logger = zapLogger.Sugar()
	}
	defer logger.Sync()

	logger.Info("Starting API Spec Validation")

	// Create validator
	validator := api_spec_validation.NewAPISpecValidator(*serverURL, logger)

	// Set auth token if provided
	if *token != "" {
		validator.SetToken(*token)
	}

	validator.SetAdditionalCookie("test", "test")

	// Load specs
	logger.Infow("Loading specs from directory", "dir", *specsDir)
	if err := validator.LoadSpecs(*specsDir); err != nil {
		logger.Errorw("Failed to load specs", "error", err)
		os.Exit(1)
	}

	// Validate all specs
	logger.Info("Starting validation")
	results := validator.ValidateAllSpecs()

	// Generate report
	report := validator.GenerateReport()

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		logger.Errorw("Failed to create output directory", "error", err)
		os.Exit(1)
	}

	// Write report to file
	reportPath := filepath.Join(*outputDir, "api-spec-validation-report.md")
	if err := os.WriteFile(reportPath, []byte(report), 0644); err != nil {
		logger.Errorw("Failed to write report", "error", err)
		os.Exit(1)
	}

	// Print summary
	fmt.Println("\n" + report)
	fmt.Printf("\nDetailed report written to: %s\n", reportPath)

	// Exit with error code if there are failures
	failedCount := 0
	for _, result := range results {
		if result.Status == "FAIL" {
			failedCount++
		}
	}

	if failedCount > 0 {
		logger.Errorw("Validation completed with failures", "failed", failedCount)
		os.Exit(1)
	}

	logger.Info("Validation completed successfully")
}
