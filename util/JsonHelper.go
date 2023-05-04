package util

import (
	"encoding/json"
	"reflect"
)

// IsJSONStringEqual compares if two json strings are equal(on object level) or not
func IsAJSONStringAndAnInterfaceEqual(json1 string, jIt2 interface{}) (bool, error) {
	var jIt1 interface{}
	if err := json.Unmarshal([]byte(json1), &jIt1); err != nil {
		return false, err
	}
	return reflect.DeepEqual(jIt2, jIt1), nil
}
