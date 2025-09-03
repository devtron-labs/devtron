package api_spec_validation

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"go.uber.org/zap"
)

// SpecComparator compares API specs with REST handler implementations
type SpecComparator struct {
	logger *zap.SugaredLogger
}

// HandlerInfo represents information about a REST handler
type HandlerInfo struct {
	Package     string
	FileName    string
	HandlerName string
	Method      string
	Path        string
	Parameters  []ParameterInfo
	RequestBody *RequestBodyInfo
	Response    *ResponseInfo
}

// ParameterInfo represents a handler parameter
type ParameterInfo struct {
	Name     string
	Type     string
	Location string // "query", "path", "header"
	Required bool
}

// RequestBodyInfo represents request body information
type RequestBodyInfo struct {
	Type   string
	Fields []FieldInfo
}

// ResponseInfo represents response information
type ResponseInfo struct {
	StatusCode int
	Type       string
	Fields     []FieldInfo
}

// FieldInfo represents a field in a struct
type FieldInfo struct {
	Name     string
	Type     string
	Tag      string
	Required bool
}

// NewSpecComparator creates a new spec comparator
func NewSpecComparator(logger *zap.SugaredLogger) *SpecComparator {
	return &SpecComparator{
		logger: logger,
	}
}

// CompareSpecsWithHandlers compares API specs with REST handler implementations
func (sc *SpecComparator) CompareSpecsWithHandlers(specsDir, handlersDir string) ([]ComparisonResult, error) {
	var results []ComparisonResult

	// Load all specs
	specs, err := sc.loadAllSpecs(specsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load specs: %w", err)
	}

	// Load all handlers
	handlers, err := sc.loadAllHandlers(handlersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load handlers: %w", err)
	}

	// Compare each spec with corresponding handlers
	for specPath, spec := range specs {
		result := sc.compareSpecWithHandlers(specPath, spec, handlers)
		results = append(results, result)
	}

	return results, nil
}

// loadAllSpecs loads all OpenAPI specs from the given directory
func (sc *SpecComparator) loadAllSpecs(specsDir string) (map[string]*openapi3.T, error) {
	specs := make(map[string]*openapi3.T)

	err := filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		loader := openapi3.NewLoader()
		spec, err := loader.LoadFromFile(path)
		if err != nil {
			sc.logger.Errorw("Failed to load spec file", "file", path, "error", err)
			return err
		}

		specs[path] = spec
		return nil
	})

	return specs, err
}

// loadAllHandlers loads all REST handlers from the given directory
func (sc *SpecComparator) loadAllHandlers(handlersDir string) (map[string]*HandlerInfo, error) {
	handlers := make(map[string]*HandlerInfo)

	err := filepath.Walk(handlersDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// Skip test files
		if strings.Contains(info.Name(), "_test.go") {
			return nil
		}

		fileHandlers, err := sc.parseHandlerFile(path)
		if err != nil {
			sc.logger.Errorw("Failed to parse handler file", "file", path, "error", err)
			return err
		}

		for _, handler := range fileHandlers {
			key := fmt.Sprintf("%s:%s", handler.Method, handler.Path)
			handlers[key] = handler
		}

		return nil
	})

	return handlers, err
}

// parseHandlerFile parses a Go file to extract handler information
func (sc *SpecComparator) parseHandlerFile(filePath string) ([]*HandlerInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var handlers []*HandlerInfo

	// Extract package name
	packageName := node.Name.Name

	// Look for handler functions
	ast.Inspect(node, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			handler := sc.extractHandlerInfo(funcDecl, packageName, filePath)
			if handler != nil {
				handlers = append(handlers, handler)
			}
		}
		return true
	})

	return handlers, nil
}

// extractHandlerInfo extracts handler information from a function declaration
func (sc *SpecComparator) extractHandlerInfo(funcDecl *ast.FuncDecl, packageName, filePath string) *HandlerInfo {
	// Look for common handler patterns
	handlerName := funcDecl.Name.Name

	// Check if this looks like a REST handler
	if !sc.isRestHandler(handlerName) {
		return nil
	}

	handler := &HandlerInfo{
		Package:     packageName,
		FileName:    filePath,
		HandlerName: handlerName,
		Parameters:  make([]ParameterInfo, 0),
	}

	// Extract parameters
	if funcDecl.Type != nil && funcDecl.Type.Params != nil {
		for _, param := range funcDecl.Type.Params.List {
			paramInfo := sc.extractParameterInfo(param)
			handler.Parameters = append(handler.Parameters, paramInfo)
		}
	}

	// Try to extract path and method from comments or function name
	handler.Path, handler.Method = sc.extractPathAndMethod(funcDecl)

	return handler
}

// isRestHandler checks if a function looks like a REST handler
func (sc *SpecComparator) isRestHandler(funcName string) bool {
	// Common REST handler patterns
	patterns := []string{
		"Create", "Get", "Update", "Delete", "List", "Find",
		"Create", "Get", "Update", "Delete", "List", "Find",
		"Handler", "RestHandler", "APIHandler",
	}

	for _, pattern := range patterns {
		if strings.Contains(funcName, pattern) {
			return true
		}
	}

	return false
}

// extractParameterInfo extracts parameter information from an AST field
func (sc *SpecComparator) extractParameterInfo(field *ast.Field) ParameterInfo {
	param := ParameterInfo{}

	// Extract parameter name
	if len(field.Names) > 0 {
		param.Name = field.Names[0].Name
	}

	// Extract parameter type
	if field.Type != nil {
		param.Type = sc.getTypeString(field.Type)
	}

	// Try to determine if it's required based on type
	param.Required = !strings.Contains(param.Type, "*") && !strings.Contains(param.Type, "interface{}")

	return param
}

// getTypeString converts an AST type to a string representation
func (sc *SpecComparator) getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + sc.getTypeString(t.X)
	case *ast.SelectorExpr:
		return sc.getTypeString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + sc.getTypeString(t.Elt)
	default:
		return "unknown"
	}
}

// extractPathAndMethod extracts path and method from function comments or name
func (sc *SpecComparator) extractPathAndMethod(funcDecl *ast.FuncDecl) (string, string) {
	// Look for path and method in comments
	if funcDecl.Doc != nil {
		for _, comment := range funcDecl.Doc.List {
			text := comment.Text
			if strings.Contains(text, "Path:") {
				// Extract path from comment
				// This is a simplified version - in practice, you'd use regex
			}
			if strings.Contains(text, "Method:") {
				// Extract method from comment
				// This is a simplified version - in practice, you'd use regex
			}
		}
	}

	// Try to infer from function name
	funcName := funcDecl.Name.Name
	if strings.Contains(funcName, "Get") {
		return "/unknown", "GET"
	}
	if strings.Contains(funcName, "Create") || strings.Contains(funcName, "Post") {
		return "/unknown", "POST"
	}
	if strings.Contains(funcName, "Update") || strings.Contains(funcName, "Put") {
		return "/unknown", "PUT"
	}
	if strings.Contains(funcName, "Delete") {
		return "/unknown", "DELETE"
	}

	return "/unknown", "GET"
}

// compareSpecWithHandlers compares a spec with handler implementations
func (sc *SpecComparator) compareSpecWithHandlers(specPath string, spec *openapi3.T, handlers map[string]*HandlerInfo) ComparisonResult {
	result := ComparisonResult{
		SpecFile: specPath,
		Issues:   make([]ComparisonIssue, 0),
	}

	if spec.Paths == nil {
		result.Issues = append(result.Issues, ComparisonIssue{
			Type:    "NO_PATHS",
			Message: "Spec has no paths defined",
		})
		return result
	}

	// Compare each path in the spec
	for path, pathItem := range spec.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			handlerKey := fmt.Sprintf("%s:%s", method, path)
			handler, exists := handlers[handlerKey]

			if !exists {
				result.Issues = append(result.Issues, ComparisonIssue{
					Type:    "MISSING_HANDLER",
					Path:    path,
					Method:  method,
					Message: fmt.Sprintf("No handler found for %s %s", method, path),
				})
				continue
			}

			// Compare parameters
			sc.compareParameters(&result, operation, handler, path, method)

			// Compare request body
			sc.compareRequestBody(&result, operation, handler, path, method)

			// Compare response
			sc.compareResponse(&result, operation, handler, path, method)
		}
	}

	return result
}

// ComparisonResult represents the result of comparing a spec with handlers
type ComparisonResult struct {
	SpecFile string
	Issues   []ComparisonIssue
}

// ComparisonIssue represents a specific comparison issue
type ComparisonIssue struct {
	Type    string
	Path    string
	Method  string
	Field   string
	Message string
}

// compareParameters compares spec parameters with handler parameters
func (sc *SpecComparator) compareParameters(result *ComparisonResult, operation *openapi3.Operation, handler *HandlerInfo, path, method string) {
	if operation.Parameters == nil {
		return
	}

	// Create a map of handler parameters for easy lookup
	handlerParams := make(map[string]ParameterInfo)
	for _, param := range handler.Parameters {
		handlerParams[param.Name] = param
	}

	// Check each spec parameter
	for _, param := range operation.Parameters {
		if param.Value == nil {
			continue
		}

		specParam := param.Value
		handlerParam, exists := handlerParams[specParam.Name]

		if !exists {
			result.Issues = append(result.Issues, ComparisonIssue{
				Type:    "MISSING_PARAMETER",
				Path:    path,
				Method:  method,
				Field:   specParam.Name,
				Message: fmt.Sprintf("Parameter '%s' defined in spec but not found in handler", specParam.Name),
			})
			continue
		}

		// Compare parameter types and requirements
		if specParam.Required && !handlerParam.Required {
			result.Issues = append(result.Issues, ComparisonIssue{
				Type:    "PARAMETER_REQUIRED_MISMATCH",
				Path:    path,
				Method:  method,
				Field:   specParam.Name,
				Message: fmt.Sprintf("Parameter '%s' is required in spec but optional in handler", specParam.Name),
			})
		}
	}
}

// compareRequestBody compares spec request body with handler
func (sc *SpecComparator) compareRequestBody(result *ComparisonResult, operation *openapi3.Operation, handler *HandlerInfo, path, method string) {
	// This is a simplified comparison - in practice, you'd do more detailed analysis
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		// Check if handler has appropriate request body handling
		hasRequestBody := false
		for _, param := range handler.Parameters {
			if strings.Contains(param.Type, "Request") || strings.Contains(param.Type, "Body") {
				hasRequestBody = true
				break
			}
		}

		if !hasRequestBody {
			result.Issues = append(result.Issues, ComparisonIssue{
				Type:    "MISSING_REQUEST_BODY",
				Path:    path,
				Method:  method,
				Message: "Request body defined in spec but no corresponding handler parameter found",
			})
		}
	}
}

// compareResponse compares spec response with handler
func (sc *SpecComparator) compareResponse(result *ComparisonResult, operation *openapi3.Operation, handler *HandlerInfo, path, method string) {
	// This is a simplified comparison - in practice, you'd do more detailed analysis
	if operation.Responses != nil {
		// Check if handler has appropriate response handling
		hasResponse := false
		for _, param := range handler.Parameters {
			if strings.Contains(param.Type, "Response") || strings.Contains(param.Type, "Writer") {
				hasResponse = true
				break
			}
		}

		if !hasResponse {
			result.Issues = append(result.Issues, ComparisonIssue{
				Type:    "MISSING_RESPONSE_HANDLING",
				Path:    path,
				Method:  method,
				Message: "Response defined in spec but no corresponding handler response handling found",
			})
		}
	}
}
