package token

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/evilmonkeyinc/jsonpath/option"
	"github.com/evilmonkeyinc/jsonpath/script"
)

func newFilterToken(expression string, engine script.Engine, options *option.QueryOptions) (*filterToken, error) {
	compiledExpression, err := engine.Compile(expression, options)
	if err != nil {
		return nil, err
	}
	return &filterToken{
		expression:         expression,
		compiledExpression: compiledExpression,
		options:            options,
	}, nil
}

type filterToken struct {
	expression         string
	compiledExpression script.CompiledExpression
	options            *option.QueryOptions
}

func (token *filterToken) String() string {
	return fmt.Sprintf("[?(%s)]", token.expression)
}

func (token *filterToken) Type() string {
	return "filter"
}

func (token *filterToken) Apply(root, current interface{}, next []Token) (interface{}, error) {
	if token.expression == "" {
		return nil, getInvalidExpressionEmptyError()
	}

	shouldInclude := func(evaluation interface{}) bool {
		if evaluation == nil {
			return false
		}

		objType, objValue := getTypeAndValue(evaluation)
		if objType == nil {
			return false
		}

		switch objType.Kind() {
		case reflect.Bool:
			return objValue.Bool()
		case reflect.Array, reflect.Slice, reflect.Map:
			return objValue.Len() > 0
		case reflect.String:
			strValue := objValue.String()
			if len(strValue) > 1 {
				if strings.HasPrefix(strValue, "'") && strings.HasSuffix(strValue, "'") {
					strValue = strValue[1 : len(strValue)-1]
				} else if strings.HasPrefix(strValue, `"`) && strings.HasSuffix(strValue, `"`) {
					strValue = strValue[1 : len(strValue)-1]
				}
			}
			return strValue != ""
		default:
			return !objValue.IsZero()
		}
	}

	elements := make([]interface{}, 0)

	objType, objVal := getTypeAndValue(current)
	if objType == nil {
		return nil, getInvalidTokenTargetNilError(token.Type(), reflect.Array, reflect.Map, reflect.Slice)
	}

	switch objType.Kind() {
	case reflect.Map:
		keys := objVal.MapKeys()
		sortMapKeys(keys)

		for _, kv := range keys {
			element := objVal.MapIndex(kv).Interface()

			evaluation, err := token.compiledExpression.Evaluate(root, element)
			if err != nil {
				// we ignore errors, it has failed evaluation
				evaluation = nil
			}

			if shouldInclude(evaluation) {
				elements = append(elements, element)
			}
		}
	case reflect.Array, reflect.Slice:
		length := objVal.Len()

		for i := 0; i < length; i++ {
			element := objVal.Index(i).Interface()

			evaluation, err := token.compiledExpression.Evaluate(root, element)
			if err != nil {
				// we ignore errors, it has failed evaluation
				evaluation = nil
			}

			if shouldInclude(evaluation) {
				elements = append(elements, element)
			}
		}
	default:
		return nil, getInvalidTokenTargetError(
			token.Type(),
			objType.Kind(),
			reflect.Array, reflect.Map, reflect.Slice,
		)
	}

	if len(next) > 0 {
		nextToken := next[0]
		futureTokens := next[1:]

		if indexToken, ok := nextToken.(*indexToken); ok {
			// if next is asking for specific index
			return indexToken.Apply(current, elements, futureTokens)
		}
		// any other token type
		results := make([]interface{}, 0)
		for _, element := range elements {
			result, _ := nextToken.Apply(root, element, futureTokens)
			if result != nil {
				results = append(results, result)
			}
		}
		return results, nil
	}
	return elements, nil
}
