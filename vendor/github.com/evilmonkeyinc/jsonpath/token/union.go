package token

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/evilmonkeyinc/jsonpath/option"
)

func newUnionToken(arguments []interface{}, options *option.QueryOptions) *unionToken {
	allowMap := false
	allowString := false
	failUnionOnInvalidIdentifier := false

	if options != nil {
		allowMap = options.AllowMapReferenceByIndex || options.AllowMapReferenceByIndexInUnion
		allowString = options.AllowStringReferenceByIndex || options.AllowStringReferenceByIndexInUnion

		failUnionOnInvalidIdentifier = options.FailUnionOnInvalidIdentifier
	}

	return &unionToken{
		arguments:                    arguments,
		allowMap:                     allowMap,
		allowString:                  allowString,
		failUnionOnInvalidIdentifier: failUnionOnInvalidIdentifier,
	}
}

type unionToken struct {
	arguments                    []interface{}
	allowMap                     bool
	allowString                  bool
	failUnionOnInvalidIdentifier bool
}

func (token *unionToken) String() string {
	args := ""
	for _, arg := range token.arguments {
		if strArg, ok := arg.(string); ok {
			args += fmt.Sprintf("'%s',", strArg)
		} else if intArg, ok := isInteger(arg); ok {
			args += fmt.Sprintf("%d,", intArg)
		} else {
			args += fmt.Sprintf("%s,", arg)
		}
	}
	args = strings.Trim(args, ",")
	return fmt.Sprintf("[%s]", args)
}

func (token *unionToken) Type() string {
	return "union"
}

func (token *unionToken) Apply(root, current interface{}, next []Token) (interface{}, error) {
	arguments := token.arguments
	if len(arguments) == 0 {
		return nil, getInvalidTokenArgumentNilError(token.Type(), reflect.Array, reflect.Slice)
	}

	keys := make([]string, 0)
	indices := make([]int64, 0)

	for _, arg := range arguments {
		argument, kind, err := token.parseArgument(root, current, arg)
		if err != nil {
			return nil, err
		}

		switch kind {
		case reflect.String:
			keys = append(keys, argument.(string))
			if len(indices) > 0 {
				return nil, getInvalidTokenArgumentError(token.Type(), reflect.String, reflect.Int)
			}
			break
		case reflect.Int64:
			indices = append(indices, argument.(int64))
			if len(keys) > 0 {
				return nil, getInvalidTokenArgumentError(token.Type(), reflect.Int, reflect.String)
			}
			break
		}
	}

	if len(keys) > 0 {
		return token.getUnionByKey(root, current, keys, next)
	}
	return token.getUnionByIndex(root, current, indices, next)
}

func (token *unionToken) parseArgument(root, current, argument interface{}) (interface{}, reflect.Kind, error) {
	if argToken, ok := argument.(Token); ok {
		result, err := argToken.Apply(root, current, nil)
		if err != nil {
			return nil, reflect.Invalid, getInvalidTokenError(token.Type(), err)
		}
		argument = result
	}

	if argument == nil {
		return nil, reflect.Invalid, getInvalidTokenArgumentNilError(token.Type(), reflect.Int, reflect.String)
	}

	if strArg, ok := argument.(string); ok {
		return strArg, reflect.String, nil
	} else if intArg, ok := isInteger(argument); ok {
		return intArg, reflect.Int64, nil
	}
	argType := reflect.TypeOf(argument)
	return nil, reflect.Invalid, getInvalidTokenArgumentError(token.Type(), argType.Kind(), reflect.Int, reflect.String)
}

func (token *unionToken) getUnionByKey(root, current interface{}, keys []string, next []Token) (interface{}, error) {
	objType, objVal := getTypeAndValue(current)
	if objType == nil {
		return nil, getInvalidTokenTargetNilError(token.Type(), reflect.Map)
	}

	var nextToken Token
	var futureTokens []Token
	forEach := false

	if len(next) > 0 {
		nextToken = next[0]
		futureTokens = next[1:]

		if _, ok := nextToken.(*indexToken); !ok {
			forEach = true
		}
	}

	elements := make([]interface{}, 0)

	switch objType.Kind() {
	case reflect.Map:
		mapKeys := objVal.MapKeys()
		sortMapKeys(mapKeys)

		keysMap := make(map[string]reflect.Value)
		for _, key := range mapKeys {
			keysMap[key.String()] = key
		}

		missingKeys := make([]string, 0)

		for _, requestedKey := range keys {
			if key, ok := keysMap[requestedKey]; ok {
				val := objVal.MapIndex(key).Interface()
				if item, add := token.handleNext(root, val, forEach, nextToken, futureTokens); add {
					elements = append(elements, item)
				}
			} else {
				missingKeys = append(missingKeys, requestedKey)
			}
		}

		if token.failUnionOnInvalidIdentifier && len(missingKeys) > 0 {
			sort.Strings(missingKeys)
			return nil, getInvalidTokenKeyNotFoundError(token.Type(), strings.Join(missingKeys, ","))
		}
	case reflect.Struct:
		keysMap := getStructFields(objVal, false)
		missingKeys := make([]string, 0)

		for _, requestedKey := range keys {
			if field, ok := keysMap[requestedKey]; ok {
				val := objVal.FieldByName(field.Name).Interface()
				if item, add := token.handleNext(root, val, forEach, nextToken, futureTokens); add {
					elements = append(elements, item)
				}
			} else {
				missingKeys = append(missingKeys, requestedKey)
			}
		}

		if token.failUnionOnInvalidIdentifier && len(missingKeys) > 0 {
			sort.Strings(missingKeys)
			return nil, getInvalidTokenKeyNotFoundError(token.Type(), strings.Join(missingKeys, ","))
		}
	default:
		return nil, getInvalidTokenTargetError(
			token.Type(),
			objType.Kind(),
			reflect.Map,
		)
	}

	if !forEach && nextToken != nil {
		return nextToken.Apply(root, elements, futureTokens)
	}

	return elements, nil
}

func (token *unionToken) getUnionByIndex(root, current interface{}, indices []int64, next []Token) (interface{}, error) {
	allowedType := []reflect.Kind{
		reflect.Array,
		reflect.Slice,
	}
	if token.allowMap {
		allowedType = append(allowedType, reflect.Map)
	}
	if token.allowString {
		allowedType = append(allowedType, reflect.String)
	}

	objType, objVal := getTypeAndValue(current)
	if objType == nil {
		return nil, getInvalidTokenTargetNilError(
			token.Type(),
			allowedType...,
		)
	}

	var nextToken Token
	var futureTokens []Token
	forEach := false

	if len(next) > 0 {
		nextToken = next[0]
		futureTokens = next[1:]

		if _, ok := nextToken.(*indexToken); !ok {
			forEach = true
		}
	}

	var length int64
	var mapKeys []reflect.Value
	isString := false

	switch objType.Kind() {
	case reflect.Map:
		if !token.allowMap {
			return nil, getInvalidTokenTargetError(
				token.Type(),
				objType.Kind(),
				allowedType...,
			)
		}
		length = int64(objVal.Len())
		mapKeys = objVal.MapKeys()
		sortMapKeys(mapKeys)
		break
	case reflect.String:
		if !token.allowString {
			return nil, getInvalidTokenTargetError(
				token.Type(),
				objType.Kind(),
				allowedType...,
			)
		}
		isString = true
		fallthrough
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		length = int64(objVal.Len())
		mapKeys = nil
		break
	default:
		return nil, getInvalidTokenTargetError(
			token.Type(),
			objType.Kind(),
			allowedType...,
		)
	}

	values := make([]interface{}, 0)
	substring := ""

	for _, idx := range indices {
		if idx < 0 {
			idx = length + idx
		}
		if idx < 0 || idx >= length {
			if token.failUnionOnInvalidIdentifier {
				return nil, getInvalidTokenOutOfRangeError(token.Type())
			}
			continue
		}

		if mapKeys != nil {
			key := mapKeys[idx]
			val := objVal.MapIndex(key).Interface()
			if item, add := token.handleNext(root, val, forEach, nextToken, futureTokens); add {
				values = append(values, item)
			}
		} else if isString {
			value := objVal.Index(int(idx)).Interface()
			if u, ok := value.(uint8); ok {
				substring += fmt.Sprintf("%c", u)
			}
		} else {
			val := objVal.Index(int(idx)).Interface()
			if item, add := token.handleNext(root, val, forEach, nextToken, futureTokens); add {
				values = append(values, item)
			}
		}
	}

	if isString {
		if nextToken != nil {
			return nextToken.Apply(root, substring, futureTokens)
		}
		return substring, nil
	}

	if !forEach && nextToken != nil {
		return nextToken.Apply(root, values, futureTokens)
	}

	return values, nil
}

func (token *unionToken) handleNext(root, item interface{}, forEach bool, nextToken Token, futureTokens []Token) (interface{}, bool) {
	if !forEach {
		return item, true
	}
	val, err := nextToken.Apply(root, item, futureTokens)
	if err != nil {
		return nil, false
	}
	if val == nil {
		return nil, false
	}
	return val, true
}
