package standard

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type operator interface {
	Evaluate(parameters map[string]interface{}) (interface{}, error)
}

func getInteger(argument interface{}, parameters map[string]interface{}) (int64, error) {
	if argument == nil {
		return 0, errInvalidArgumentNil
	}
	if parameters == nil {
		parameters = make(map[string]interface{})
	}

	if sub, ok := argument.(operator); ok {
		arg, err := sub.Evaluate(parameters)
		if err != nil {
			return 0, err
		}
		argument = arg
	}

	if str, ok := argument.(string); ok {
		if arg, ok := parameters[str]; ok {
			argument = arg
		}

	}

	str := fmt.Sprintf("%v", argument)
	intVal, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, errInvalidArgumentExpectedInteger
	}
	return intVal, nil
}

func getNumber(argument interface{}, parameters map[string]interface{}) (float64, error) {
	if argument == nil {
		return 0, errInvalidArgumentNil
	}
	if parameters == nil {
		parameters = make(map[string]interface{})
	}

	if sub, ok := argument.(operator); ok {
		arg, err := sub.Evaluate(parameters)
		if err != nil {
			return 0, err
		}
		argument = arg
	}

	if str, ok := argument.(string); ok {
		if arg, ok := parameters[str]; ok {
			argument = arg
		}
	}

	str := fmt.Sprintf("%v", argument)
	floatVal, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, errInvalidArgumentExpectedNumber
	}
	return floatVal, nil
}

func getBoolean(argument interface{}, parameters map[string]interface{}) (bool, error) {
	if argument == nil {
		return false, nil
	}
	if parameters == nil {
		parameters = make(map[string]interface{})
	}

	if sub, ok := argument.(operator); ok {
		arg, err := sub.Evaluate(parameters)
		if err != nil {
			return false, err
		}
		argument = arg
	}

	if str, ok := argument.(string); ok {
		if arg, ok := parameters[str]; ok {
			argument = arg
		}
	}

	if argument == nil {
		return false, nil
	}

	str := fmt.Sprintf("%v", argument)
	boolValue, err := strconv.ParseBool(str)
	if err != nil {
		return false, errInvalidArgumentExpectedBoolean
	}
	return boolValue, nil
}

func getString(argument interface{}, parameters map[string]interface{}) (string, error) {
	if argument == nil {
		return "", errInvalidArgumentNil
	}
	if parameters == nil {
		parameters = make(map[string]interface{})
	}

	if sub, ok := argument.(operator); ok {
		arg, err := sub.Evaluate(parameters)
		if err != nil {
			return "", err
		}
		argument = arg
	}

	if str, ok := argument.(string); ok {
		if arg, ok := parameters[str]; ok {
			argument = arg
			if parsed, ok := arg.(string); ok {
				str = parsed

				if len(str) > 1 {
					if strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'") {
						return str, nil
					} else if strings.HasPrefix(str, `"`) && strings.HasSuffix(str, `"`) {
						str = str[1 : len(str)-1]
					}
				}
				return fmt.Sprintf("'%s'", str), nil
			}
		} else {
			if len(str) > 1 {
				if strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'") {
					return str, nil
				} else if strings.HasPrefix(str, `"`) && strings.HasSuffix(str, `"`) {
					str = str[1 : len(str)-1]
					return fmt.Sprintf("'%s'", str), nil
				}
			}
			return str, nil
		}
	}

	return fmt.Sprintf("%v", argument), nil
}

func getElements(argument interface{}, parameters map[string]interface{}) ([]interface{}, error) {
	if argument == nil {
		return nil, errInvalidArgumentNil
	}

	if sub, ok := argument.(operator); ok {
		arg, err := sub.Evaluate(parameters)
		if err != nil {
			return nil, err
		}
		argument = arg
	}

	if strValue, ok := argument.(string); ok {
		if param, ok := parameters[strValue]; ok {
			argument = param
		} else {
			if strings.HasPrefix(strValue, "{") && strings.HasSuffix(strValue, "}") {
				// object
				root := make(map[string]interface{})
				if err := json.Unmarshal([]byte(strValue), &root); err != nil {
					return nil, errInvalidArgument
				}
				argument = root
			} else if strings.HasPrefix(strValue, "[") && strings.HasSuffix(strValue, "]") {
				// array
				root := make([]interface{}, 0)
				if err := json.Unmarshal([]byte(strValue), &root); err != nil {
					return nil, errInvalidArgument
				}
				argument = root
			}
		}
	}

	objType := reflect.TypeOf(argument)
	if objType == nil {
		return nil, errInvalidArgumentNil
	}
	objValue := reflect.ValueOf(argument)

	if objType.Kind() == reflect.Ptr {
		if objValue.IsNil() {
			return nil, errInvalidArgumentNil
		}
		objType = objType.Elem()
		objValue = objValue.Elem()
	}

	elements := make([]interface{}, 0)

	switch objType.Kind() {
	case reflect.Array, reflect.Slice:
		length := objValue.Len()
		for i := 0; i < length; i++ {
			elements = append(elements, objValue.Index(i).Interface())
		}
		break
	case reflect.Map:
		keys := objValue.MapKeys()
		for _, key := range keys {
			elements = append(elements, objValue.MapIndex(key).Interface())
		}
		break
	default:
		return nil, errInvalidArgumentExpectedCollection
	}

	return elements, nil
}
