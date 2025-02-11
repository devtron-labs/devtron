package script

import "github.com/evilmonkeyinc/jsonpath/option"

// Engine represents a script engine used by the JSONPath query parser
type Engine interface {
	// Compile returns a compiled expression that can be evaluated multiple times
	Compile(expression string, options *option.QueryOptions) (CompiledExpression, error)
	// Evaluate return the result of the expression evaluation
	Evaluate(root, current interface{}, expression string, options *option.QueryOptions) (interface{}, error)
}

// CompiledExpression represents a compile expression that can be evaluated multiple times
type CompiledExpression interface {
	// Evaluate return the result of the expression evaluation
	Evaluate(root, current interface{}) (interface{}, error)
}
