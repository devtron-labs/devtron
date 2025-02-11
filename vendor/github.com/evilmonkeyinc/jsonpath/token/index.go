package token

import (
	"fmt"
	"reflect"

	"github.com/evilmonkeyinc/jsonpath/option"
)

func newIndexToken(index int64, options *option.QueryOptions) *indexToken {
	allowMap := false
	allowString := false

	if options != nil {
		allowMap = options.AllowMapReferenceByIndex || options.AllowMapReferenceByIndexInSubscript
		allowString = options.AllowStringReferenceByIndex || options.AllowStringReferenceByIndexInSubscript
	}

	return &indexToken{
		index:       index,
		allowMap:    allowMap,
		allowString: allowString,
	}
}

type indexToken struct {
	index       int64
	allowMap    bool
	allowString bool
}

func (token *indexToken) String() string {
	return fmt.Sprintf("[%d]", token.index)
}

func (token *indexToken) Type() string {
	return "index"
}

func (token *indexToken) Apply(root, current interface{}, next []Token) (interface{}, error) {
	idx := token.index

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
	case reflect.Array, reflect.Slice:
		length = int64(objVal.Len())
		mapKeys = nil
	default:
		return nil, getInvalidTokenTargetError(
			token.Type(),
			objType.Kind(),
			allowedType...,
		)
	}

	if idx < 0 {
		idx = length + idx
	}

	if idx < 0 || idx >= length {
		return nil, getInvalidTokenOutOfRangeError(token.Type())
	}

	var value interface{}

	if mapKeys != nil {
		key := mapKeys[idx]
		value = objVal.MapIndex(key).Interface()

	} else if isString {
		value = objVal.Index(int(idx)).Interface()

		if u, ok := value.(uint8); ok {
			value = fmt.Sprintf("%c", u)
		}
	} else {
		value = objVal.Index(int(idx)).Interface()
	}

	if len(next) > 0 {
		return next[0].Apply(root, value, next[1:])
	}
	return value, nil
}
