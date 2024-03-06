package expr

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/oliveagle/jsonpath"
)

func GetExprEnvFunctionMap() map[string]interface{} {
	return map[string]interface{}{
		"asInt":    AsInt,
		"asFloat":  AsFloat,
		"string":   AsStr,
		"jsonpath": JsonPath,
	}
}

func AsStr(val interface{}) interface{} {
	return fmt.Sprintf("%v", val)
}

func JsonPath(jsonStr string, path string) interface{} {
	var jsonMap interface{}
	err := json.Unmarshal([]byte(jsonStr), &jsonMap)
	if err != nil {
		panic(err)
	}
	value, err := jsonpath.JsonPathLookup(jsonMap, path)
	if err != nil {
		panic(err)
	}
	return value
}

func AsInt(in interface{}) int64 {
	switch i := in.(type) {
	case float64:
		return int64(i)
	case float32:
		return int64(i)
	case int64:
		return i
	case int32:
		return int64(i)
	case int16:
		return int64(i)
	case int8:
		return int64(i)
	case int:
		return int64(i)
	case uint64:
		return int64(i)
	case uint32:
		return int64(i)
	case uint16:
		return int64(i)
	case uint8:
		return int64(i)
	case uint:
		return int64(i)
	case string:
		inAsInt, err := strconv.ParseInt(i, 10, 64)
		if err == nil {
			return inAsInt
		}
		panic(err)
	}
	panic(fmt.Sprintf("asInt() not supported on %v %v", reflect.TypeOf(in), in))
}

func AsFloat(in interface{}) float64 {
	switch i := in.(type) {
	case float64:
		return i
	case float32:
		return float64(i)
	case int64:
		return float64(i)
	case int32:
		return float64(i)
	case int16:
		return float64(i)
	case int8:
		return float64(i)
	case int:
		return float64(i)
	case uint64:
		return float64(i)
	case uint32:
		return float64(i)
	case uint16:
		return float64(i)
	case uint8:
		return float64(i)
	case uint:
		return float64(i)
	case string:
		inAsFloat, err := strconv.ParseFloat(i, 64)
		if err == nil {
			return inAsFloat
		}
		panic(err)
	}
	panic(fmt.Sprintf("asFloat() not supported on %v %v", reflect.TypeOf(in), in))
}
