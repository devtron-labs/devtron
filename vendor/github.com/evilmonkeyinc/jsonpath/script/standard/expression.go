package standard

import "github.com/evilmonkeyinc/jsonpath/option"

type compiledExpression struct {
	expression   string
	rootOperator operator
	engine       *ScriptEngine
	options      *option.QueryOptions
}

func (compiled *compiledExpression) Evaluate(root, current interface{}) (interface{}, error) {
	expression := compiled.expression
	if expression == "" {
		return nil, getInvalidExpressionEmptyError()
	}
	parameters := map[string]interface{}{
		"$":    root,
		"@":    current,
		"nil":  nil,
		"null": nil,
	}

	if compiled.rootOperator == nil {
		if val, ok := parameters[expression]; ok {
			return val, nil
		}

		if number, err := getNumber(expression, parameters); err == nil {
			return number, nil
		} else if boolean, err := getBoolean(expression, parameters); err == nil {
			return boolean, nil
		}

		return expression, nil
	}

	value, err := compiled.rootOperator.Evaluate(parameters)
	if err != nil {
		return nil, err
	}

	return value, nil
}
