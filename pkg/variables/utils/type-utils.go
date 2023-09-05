package utils

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"strconv"
	"strings"
)

func StringifyValue(data interface{}) (string, error) {
	var value string
	switch data.(type) {
	case json.Number:
		marshal, _ := json.Marshal(data)
		value = string(marshal)
	case string:
		value = data.(string)
		value = "\"" + value + "\""
	case bool:
		value = strconv.FormatBool(data.(bool))
	default:
		return "", fmt.Errorf("complex values are not allowed. %v neeeds to be stringified", data)
	}
	return value, nil
}
func DestringifyValue(Data string) (interface{}, error) {
	var value interface{}
	if intValue, err := strconv.Atoi(Data); err == nil {
		value = intValue
	} else if floatValue, err := strconv.ParseFloat(Data, 64); err == nil {
		value = floatValue
	} else if boolValue, err := strconv.ParseBool(Data); err == nil {
		value = boolValue
	} else {
		value = strings.Trim(Data, "\"")
	}
	return value, nil
}

func IsValidYAML(input string) bool {
	jsonInput, err := yaml.YAMLToJSONStrict([]byte(input))
	if err != nil {
		return false
	}
	validJson := IsValidJSON(string(jsonInput))
	return validJson
}
func IsValidJSON(input string) bool {
	data := make(map[string]interface{})
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return false
	}
	return true
}
