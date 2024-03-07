package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"
)

func CheckIfNilInWire() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test_wire_gen.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Error parsing file: %v", err)
	}

	funcName := "InitializeAppTest"
	var targetFunc *ast.FuncDecl
	for _, decl := range node.Decls {
		if fdecl, ok := decl.(*ast.FuncDecl); ok && fdecl.Name.Name == funcName {
			targetFunc = fdecl
			break
		}
	}
	if targetFunc == nil {
		log.Fatalf("Function %s not found in wire_gen.go", funcName)
	}

	ast.Inspect(targetFunc.Body, func(n ast.Node) bool {
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if ident, ok := callExpr.Fun.(*ast.Ident); ok {
				if strings.HasPrefix(ident.Name, "Provider") {
					fmt.Println("found call to provider function", ident.Name)
					response := Provider()
					if hasNilFields(response) {
						fmt.Println("error, Response object has nil fields")
					}
				}
			}
		}
		return true
	})
}

func hasNilFields(response *Response) bool {
	if response == nil {
		return true
	}
	//to check for interface and add reflect support
	if response.Field1 == nil || response.Field2 == nil {
		return true
	}
	return false
}
