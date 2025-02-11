package token

import (
	"math"
	"reflect"
	"sort"
	"strings"
)

func isInteger(obj interface{}) (int64, bool) {
	objType, objVal := getTypeAndValue(obj)
	if objType == nil {
		return 0, false
	}

	switch objType.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		return objVal.Int(), true
	case reflect.Float32, reflect.Float64:
		float := objVal.Float()
		if trunc := math.Trunc(float); trunc == float {
			return int64(trunc), true
		}
	}

	return 0, false
}

func getStructFields(obj reflect.Value, omitempty bool) map[string]reflect.StructField {
	objType := obj.Type()
	if objType.Kind() != reflect.Struct {
		return nil
	}

	fields := make(map[string]reflect.StructField)

	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		fieldName := field.Name

		switch jsonTag := field.Tag.Get("json"); jsonTag {
		case "-":
			// explicitly told to skip
			break
		case "":
			if _, exists := fields[fieldName]; !exists {
				// Do not want to override one set with json tag
				fields[fieldName] = field
			}
			break
		default:
			parts := strings.Split(jsonTag, ",")
			name := parts[0]
			if name == "" {
				name = fieldName
			}

			if omitempty && len(parts) > 1 {
				if parts[1] == "omitempty" {
					// if we are to omit empty we check if it has 0 value
					value := obj.FieldByName(field.Name)
					if value.IsZero() {
						continue
					}
				}
			}

			fields[name] = field
			break
		}
	}

	return fields
}

func getTypeAndValue(obj interface{}) (reflect.Type, reflect.Value) {
	objType := reflect.TypeOf(obj)
	if objType == nil {
		return nil, reflect.ValueOf(nil)
	}

	objVal := reflect.ValueOf(obj)

	if objType.Kind() == reflect.Ptr {
		if objVal.IsNil() {
			return nil, reflect.ValueOf(nil)
		}
		objType = objType.Elem()
		objVal = objVal.Elem()
	}

	return objType, objVal
}

func sortMapKeys(mapKeys []reflect.Value) {
	sort.SliceStable(mapKeys, func(i, j int) bool {
		one := mapKeys[i]
		two := mapKeys[j]

		return one.String() < two.String()
	})
}
