package token

import (
	"fmt"
	"reflect"

	"github.com/evilmonkeyinc/jsonpath/option"
)

func newRangeToken(from, to, step interface{}, options *option.QueryOptions) *rangeToken {
	allowMap := false
	allowString := false

	if options != nil {
		allowMap = options.AllowMapReferenceByIndex || options.AllowMapReferenceByIndexInRange
		allowString = options.AllowStringReferenceByIndex || options.AllowStringReferenceByIndexInRange
	}

	return &rangeToken{
		from:        from,
		to:          to,
		step:        step,
		allowMap:    allowMap,
		allowString: allowString,
	}
}

type rangeToken struct {
	from, to, step interface{}
	allowMap       bool
	allowString    bool
}

func (token *rangeToken) String() string {
	fString := ""
	if token.from != nil {
		fString = fmt.Sprint(token.from)
	}
	tString := ""
	if token.to != nil {
		tString = fmt.Sprint(token.to)
	}
	if token.step == nil {
		return fmt.Sprintf("[%s:%s]", fString, tString)
	}

	sString := fmt.Sprint(token.step)
	return fmt.Sprintf("[%s:%s:%s]", fString, tString, sString)
}

func (token *rangeToken) Type() string {
	return "range"
}

func (token *rangeToken) Apply(root, current interface{}, next []Token) (interface{}, error) {

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

	var from int64 = 0
	if token.from != nil {
		var err error
		from, err = token.parseArgument(root, current, token.from)
		if err != nil {
			return nil, err
		}
		if from < 0 {
			from = length + from
		}
		if from < 0 {
			from = 0
		}
		if from > length {
			from = length
		}
	}

	to := length
	if token.to != nil {
		var err error
		to, err = token.parseArgument(root, current, token.to)
		if err != nil {
			return nil, err
		}
		if to < 0 {
			to = length + to
		}

		if to < 0 {
			to = 0
		}
		if to > length {
			to = length
		}
	}

	var step int64 = 1
	if token.step != nil {
		var err error
		step, err = token.parseArgument(root, current, token.step)
		if err != nil {
			return nil, err
		}
		if step == 0 {
			return nil, getInvalidTokenOutOfRangeError(token.Type())
		}
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

	if mapKeys != nil {
		if step < 0 {
			for i := to - 1; i >= from; i += step {
				key := mapKeys[i]
				item, add := token.handleNext(root, objVal.MapIndex(key).Interface(), forEach, nextToken, futureTokens)
				if add {
					elements = append(elements, item)
				}
			}
		} else {
			for i := from; i < to; i += step {
				key := mapKeys[i]
				item, add := token.handleNext(root, objVal.MapIndex(key).Interface(), forEach, nextToken, futureTokens)
				if add {
					elements = append(elements, item)
				}
			}
		}
	} else if isString {
		substring := ""
		if step < 0 {
			for i := to - 1; i >= from; i += step {
				value := objVal.Index(int(i)).Uint()
				substring += fmt.Sprintf("%c", value)
			}
		} else {
			for i := from; i < to; i += step {
				value := objVal.Index(int(i)).Uint()
				substring += fmt.Sprintf("%c", value)
			}
		}

		if len(next) > 0 {
			return next[0].Apply(root, substring, next[1:])
		}

		return substring, nil
	} else {
		if step < 0 {
			for i := to - 1; i >= from; i += step {
				item, add := token.handleNext(root, objVal.Index(int(i)).Interface(), forEach, nextToken, futureTokens)
				if add {
					elements = append(elements, item)
				}
			}
		} else {
			for i := from; i < to; i += step {
				item, add := token.handleNext(root, objVal.Index(int(i)).Interface(), forEach, nextToken, futureTokens)
				if add {
					elements = append(elements, item)
				}
			}
		}
	}

	if !forEach && nextToken != nil {
		return nextToken.Apply(root, elements, futureTokens)
	}

	return elements, nil
}

func (token *rangeToken) handleNext(root, item interface{}, forEach bool, nextToken Token, futureTokens []Token) (interface{}, bool) {
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

func (token *rangeToken) parseArgument(root, current interface{}, argument interface{}) (int64, error) {
	if script, ok := argument.(Token); ok {
		result, err := script.Apply(root, current, nil)
		if err != nil {
			return 0, getInvalidTokenError(token.Type(), err)
		}

		if result == nil {
			err := getUnexpectedExpressionResultNilError(reflect.Int)
			return 0, getInvalidTokenError(token.Type(), err)
		}
		if intVal, ok := isInteger(result); ok {
			return intVal, nil
		}

		kind := reflect.TypeOf(result).Kind()
		err = getUnexpectedExpressionResultError(kind, reflect.Int)
		return 0, getInvalidTokenError(token.Type(), err)
	} else if intVal, ok := isInteger(argument); ok {
		return intVal, nil
	}

	kind := reflect.TypeOf(argument).Kind()
	return 0, getInvalidTokenArgumentError(token.Type(), kind, reflect.Int)
}
