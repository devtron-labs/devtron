package token

import (
	"reflect"
)

func newRecursiveToken() *recursiveToken {
	return &recursiveToken{}
}

type recursiveToken struct {
}

func (token *recursiveToken) String() string {
	return ".."
}

func (token *recursiveToken) Type() string {
	return "recursive"
}

func (token *recursiveToken) Apply(root, current interface{}, next []Token) (interface{}, error) {
	return token.recursiveApply(root, current, next), nil
}

func (token *recursiveToken) recursiveApply(root, current interface{}, next []Token) []interface{} {

	slice := make([]interface{}, 0)

	objType, objVal := getTypeAndValue(current)
	if objType == nil {
		return slice
	}

	if len(next) > 0 {
		result, _ := next[0].Apply(root, objVal.Interface(), next[1:])
		objType, objVal := getTypeAndValue(result)
		if objType != nil {
			switch objType.Kind() {
			case reflect.Array, reflect.Slice:
				length := objVal.Len()
				for i := 0; i < length; i++ {
					slice = append(slice, objVal.Index(i).Interface())
				}
				break
			default:
				slice = append(slice, objVal.Interface())
				break
			}
		}
	} else {
		slice = append(slice, objVal.Interface())
	}

	switch objType.Kind() {
	case reflect.Map:
		keys := objVal.MapKeys()
		sortMapKeys(keys)
		for _, kv := range keys {
			value := objVal.MapIndex(kv).Interface()
			result := token.recursiveApply(root, value, next)
			slice = append(slice, result...)
		}
	case reflect.Array, reflect.Slice:
		length := objVal.Len()
		for i := 0; i < length; i++ {
			value := objVal.Index(i).Interface()
			result := token.recursiveApply(root, value, next)
			slice = append(slice, result...)
		}
	case reflect.Struct:
		fields := getStructFields(objVal, true)
		for _, field := range fields {
			value := objVal.FieldByName(field.Name).Interface()
			result := token.recursiveApply(root, value, next)
			slice = append(slice, result...)

		}
	default:
		break
	}

	return slice
}
