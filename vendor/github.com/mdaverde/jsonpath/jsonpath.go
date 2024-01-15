package jsonpath

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// DoesNotExist is returned by jsonpath.Get on nonexistent paths
type DoesNotExist struct{}

func (d DoesNotExist) Error() string {
	return "path not found"
}

var errInvalidObj = errors.New("invalid object")
var pathDelimiter = "."

func tokenizePath(path string) ([]string, error) {
	var tokens []string
	for _, stem := range strings.Split(path, pathDelimiter) {
		if !strings.Contains(stem, "[") {
			tokens = append(tokens, stem)
			continue
		}
		firstBracketIndex := strings.Index(stem, "[")
		lastBracketIndex := strings.LastIndex(stem, "]")
		if lastBracketIndex < 0 {
			return nil, fmt.Errorf("invalid path: %v", path)
		}
		tokens = append(tokens, stem[0:firstBracketIndex])
		innerText := stem[firstBracketIndex+1 : lastBracketIndex]
		tokens = append(tokens, innerText)
	}
	return tokens, nil
}

func getToken(obj interface{}, token string) (interface{}, error) {
	if reflect.TypeOf(obj) == nil {
		return nil, errInvalidObj
	}

	switch reflect.ValueOf(obj).Kind() {
	case reflect.Map:
		for _, kv := range reflect.ValueOf(obj).MapKeys() {
			if kv.String() == token {
				return reflect.ValueOf(obj).MapIndex(kv).Interface(), nil
			}
		}
		return nil, DoesNotExist{}
	case reflect.Slice:
		idx, err := strconv.Atoi(token)
		if err != nil {
			return nil, err
		}
		length := reflect.ValueOf(obj).Len()
		if idx > -1 {
			if idx >= length {
				return nil, DoesNotExist{}
			}
			return reflect.ValueOf(obj).Index(idx).Interface(), nil
		}
		return nil, DoesNotExist{}
	default:
		return nil, fmt.Errorf("object is not a map or a slice: %v", reflect.ValueOf(obj).Kind())
	}
}

func getByTokens(data interface{}, tokens []string) (interface{}, error) {
	var err error

	child := data
	for _, token := range tokens {
		child, err = getToken(child, token)
		if err != nil {
			return nil, err
		}
	}

	if child != nil {
		return child, nil
	}

	return nil, errors.New("could not get value at path")
}

func followPtr(data interface{}) interface{} {
	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return rv.Interface()
}

// Get returns the value at the json path or if an error occurred
func Get(data interface{}, path string) (interface{}, error) {
	var err error
	tokens, err := tokenizePath(path)
	if err != nil {
		return nil, err
	}

	data = followPtr(data)

	return getByTokens(data, tokens)
}

// Set value on data at that json path
func Set(data interface{}, path string, value interface{}) error {
	tokens, err := tokenizePath(path)
	if err != nil {
		return nil
	}

	head := tokens[:len(tokens)-1]
	last := tokens[len(tokens)-1]

	data = followPtr(data)
	value = followPtr(value)

	child := data
	parent := data

	for tokenIdx, token := range head {
		child, err = getToken(parent, token)
		if err != nil {
			if _, ok := err.(DoesNotExist); !ok && err != errInvalidObj {
				return err
			}

			child = map[string]interface{}{}

			if tokenIdx+1 < len(tokens) {
				nextToken := tokens[tokenIdx+1]
				if idx, err := strconv.Atoi(nextToken); err == nil {
					var childSlice []interface{}
					for i := 0; i < idx; i++ {
						childSlice = append(childSlice, nil)
					}
					child = append(childSlice, map[string]interface{}{})
				}
			}

			switch reflect.ValueOf(parent).Kind() {
			case reflect.Map:
				reflect.ValueOf(parent).SetMapIndex(reflect.ValueOf(token), reflect.ValueOf(child))
			case reflect.Slice:
				sliceValue := reflect.ValueOf(parent)
				idx, err := strconv.Atoi(token)
				if err != nil {
					return err
				}
				sliceValue.Index(idx).Set(reflect.ValueOf(child))
			default:
				return errors.New("path contains items that are not maps nor structs")
			}
		}

		parent = child
	}

	switch reflect.ValueOf(child).Kind() {
	case reflect.Map:
		reflect.ValueOf(child).SetMapIndex(reflect.ValueOf(last), reflect.ValueOf(value))
		return nil
	case reflect.Slice:
		sliceValue := reflect.ValueOf(child)
		idx, err := strconv.Atoi(last)
		if err != nil {
			return err
		}

		sliceValue.Index(idx).Set(reflect.ValueOf(value))
		return nil
	}

	return errors.New("could not set value at path")
}
