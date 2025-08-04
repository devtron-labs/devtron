.PHONY: build test clean help

# Default target
help:
	@echo "Available targets:"
	@echo "  build    - Build the API spec validator"
	@echo "  test     - Run API spec validation tests"
	@echo "  compare  - Compare specs with REST handlers"
	@echo "  clean    - Clean build artifacts"
	@echo "  help     - Show this help message"

# Build the validator
build:
	@echo "Building API spec validator..."
	go mod vendor
	go build -o bin/validator cmd/validator/main.go
	go build -o bin/live-test cmd/live-test/main.go

# Run API spec validation tests
test: build
	@echo "Running API spec validation tests..."
	@mkdir -p reports
	./bin/validator \
		--server=http://localhost:8080 \
		--specs=../../../specs \
		--output=./reports \
		--verbose

# Run with custom server URL
test-server: build
	@echo "Running API spec validation tests against custom server..."
	@mkdir -p reports
	./bin/validator \
		--server=$(SERVER_URL) \
		--specs=../../../specs \
		--output=./reports \
		--verbose

# Test against live Devtron server
test-live: build
	@echo "Running API spec validation tests against live Devtron server..."
	@mkdir -p reports
	./bin/live-test \
		--server=https://devtron-ent-2.devtron.info \
		--specs=../../specs \
		--output=./reports \
		--token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjphZG1pbiIsInZlcnNpb24iOiIxIiwiaXNzIjoiYXBpVG9rZW5Jc3N1ZXIiLCJleHAiOjE3NTY1NTc1NzR9.ZHhQdhXpGygCOiO7rDah0mBB7zZYZ3y9WlJL9egRfq4 \
		--verbose

# Compare specs with REST handlers
compare:
	@echo "Comparing API specs with REST handlers..."
	go run cmd/compare/main.go \
		--specs=../../../specs \
		--handlers=../../../api \
		--output=./reports

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf reports/

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go get github.com/getkin/kin-openapi/openapi3
	go get go.uber.org/zap

# Run all validations
all: deps build test compare
	@echo "All validations completed. Check reports/ directory for results." 