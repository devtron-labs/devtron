package token

import (
	"fmt"
	"reflect"
	"strings"
)

func newKeyToken(key string) *keyToken {
	return &keyToken{key: key}
}

type keyToken struct {
	key string
}

func (token *keyToken) String() string {
	key := token.key
	key = strings.ReplaceAll(key, "'", "\\'")
	return fmt.Sprintf("['%s']", key)
}

func (token *keyToken) Type() string {
	return "key"
}

func (token *keyToken) Apply(root, current interface{}, next []Token) (interface{}, error) {
	objType, objVal := getTypeAndValue(current)
	if objType == nil {
		return nil, getInvalidTokenTargetNilError(
			token.Type(),
			reflect.Map,
		)
	}

	switch objType.Kind() {
	case reflect.Map:
		keys := objVal.MapKeys()
		for _, kv := range keys {
			if kv.String() == token.key {
				value := objVal.MapIndex(kv).Interface()

				if len(next) > 0 {
					return next[0].Apply(root, value, next[1:])
				}
				return value, nil
			}
		}
		return nil, getInvalidTokenKeyNotFoundError(token.Type(), token.key)
	case reflect.Struct:
		fields := getStructFields(objVal, false)
		if field, ok := fields[token.key]; ok {
			value := objVal.FieldByName(field.Name).Interface()

			if len(next) > 0 {
				return next[0].Apply(root, value, next[1:])
			}
			return value, nil
		}
		return nil, getInvalidTokenKeyNotFoundError(token.Type(), token.key)
	default:
		return nil, getInvalidTokenTargetError(
			token.Type(),
			objType.Kind(),
			reflect.Map)
	}
}
