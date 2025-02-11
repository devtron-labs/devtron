package token

import (
	"reflect"
)

func newWildcardToken() *wildcardToken {
	return &wildcardToken{}
}

type wildcardToken struct {
}

func (token *wildcardToken) String() string {
	return "[*]"
}

func (token *wildcardToken) Type() string {
	return "wildcard"
}

func (token *wildcardToken) Apply(root, current interface{}, next []Token) (interface{}, error) {

	elements := make([]interface{}, 0)

	var nextToken Token
	var futureTokens []Token

	if len(next) > 0 {
		nextToken = next[0]
		futureTokens = next[1:]
	}

	objType, objVal := getTypeAndValue(current)
	if objType == nil {
		return nil, getInvalidTokenTargetNilError(
			token.Type(),
			reflect.Array, reflect.Map, reflect.Slice,
		)
	}

	switch objType.Kind() {
	case reflect.Map:
		keys := objVal.MapKeys()
		sortMapKeys(keys)
		for _, kv := range keys {
			value := objVal.MapIndex(kv).Interface()
			if item, add := token.handleNext(root, value, nextToken, futureTokens); add {
				elements = append(elements, item)
			}
		}
		break
	case reflect.Array, reflect.Slice:
		length := objVal.Len()
		for i := 0; i < length; i++ {
			value := objVal.Index(i).Interface()
			if item, add := token.handleNext(root, value, nextToken, futureTokens); add {
				elements = append(elements, item)
			}
		}
	case reflect.Struct:
		fields := getStructFields(objVal, true)
		for _, field := range fields {
			value := objVal.FieldByName(field.Name).Interface()
			if item, add := token.handleNext(root, value, nextToken, futureTokens); add {
				elements = append(elements, item)
			}
		}
		break
	default:
		return nil, getInvalidTokenTargetError(
			token.Type(),
			objType.Kind(),
			reflect.Array, reflect.Map, reflect.Slice,
		)
	}

	return elements, nil
}

func (token *wildcardToken) handleNext(root, item interface{}, nextToken Token, futureTokens []Token) (interface{}, bool) {
	if nextToken == nil {
		return item, true
	}
	result, _ := nextToken.Apply(root, item, futureTokens)
	if result == nil {
		return nil, false
	}
	return result, true
}
