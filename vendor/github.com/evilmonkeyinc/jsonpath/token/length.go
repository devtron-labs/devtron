package token

import (
	"reflect"
)

func newLengthToken() *lengthToken {
	return &lengthToken{}
}

type lengthToken struct {
}

func (token *lengthToken) String() string {
	return ".length"
}

func (token *lengthToken) Type() string {
	return "length"
}

func (token *lengthToken) Apply(root, current interface{}, next []Token) (interface{}, error) {

	objType, objVal := getTypeAndValue(current)
	if objType == nil {
		return nil, getInvalidTokenTargetNilError(
			token.Type(),
			reflect.Array,
			reflect.Map,
			reflect.Slice,
			reflect.String,
		)
	}

	switch objType.Kind() {
	case reflect.Map:
		current = int64(objVal.Len())

		keys := objVal.MapKeys()
		for _, kv := range keys {
			if kv.String() == "length" {
				current = objVal.MapIndex(kv).Interface()
			}
		}
		break
	case reflect.Array, reflect.Slice, reflect.String:
		current = int64(objVal.Len())
		break
	default:
		return nil, getInvalidTokenTargetError(
			token.Type(),
			objType.Kind(),
			reflect.Array, reflect.Map, reflect.Slice, reflect.String,
		)
	}

	if len(next) > 0 {
		return next[0].Apply(root, current, next[1:])
	}
	return current, nil
}
