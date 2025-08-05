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
		serverURL = flag.String("server", "https://devtron-ent-2.devtron.info", "Server URL to test against")
		specsDir  = flag.String("specs", "../../../specs", "Directory containing API specs")
		outputDir = flag.String("output", "./reports", "Output directory for reports")
		token     = flag.String("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTQ0OTE5NzMsImp0aSI6IjI3YWI0ZDE5LTdmNGYtNDg5YS1iZDg5LTVmYjU5MjNjNWE4YSIsImlhdCI6MTc1NDQwNTU3MywiaXNzIjoiYXJnb2NkIiwibmJmIjoxNzU0NDA1NTczLCJzdWIiOiJhZG1pbiJ9.ckHWRP_Lqr-RDzsM8H7Fcs7nICho-PZcgIOTppH1tLE", "Authentication token")
		verbose   = flag.Bool("verbose", true, "Enable verbose logging")
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

	logger.Info("Starting API Spec Validation against Live Devtron Server")

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
	logger.Info("Starting validation against live server")
	results := validator.ValidateAllSpecs()

	// Generate report
	report := validator.GenerateReport()

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		logger.Errorw("Failed to create output directory", "error", err)
		os.Exit(1)
	}

	// Write report to file
	reportPath := filepath.Join(*outputDir, "live-server-validation-report.md")
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
